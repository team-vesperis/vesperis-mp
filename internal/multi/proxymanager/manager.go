package proxymanager

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/redis/go-redis/v9"
	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/playermanager"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
	"github.com/team-vesperis/vesperis-mp/internal/proxy/commands"
	"github.com/team-vesperis/vesperis-mp/internal/proxy/listeners"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/gate"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiProxyManager struct {
	multiProxyMap map[uuid.UUID]*multi.Proxy
	mu            sync.RWMutex

	ownerMP *multi.Proxy

	ownerGate *proxy.Proxy

	// The command manager used in the mp.
	cm *commands.CommandManager

	// The listener manager in the mp.
	lm *listeners.ListenerManager

	// The context used in the mp.
	// Contains a cancel and logger.
	ctx context.Context

	// The logger used in the mp.
	l *logger.Logger

	// The database used in the multiproxy manager.
	// Contains a connection with Redis and Postgres.
	// Combines both in functions for fast and safe usage.
	db *database.Database

	// The config used in the mp.
	// Determines the database connection variables, proxy id, etc.
	c *config.Config

	// The multiplayer manager used in the mp.
	mpm *playermanager.MultiPlayerManager

	tm *task.TaskManager
}

func Init(ctx context.Context, l *logger.Logger, c *config.Config, db *database.Database) (*MultiProxyManager, error) {
	now := time.Now()

	mpm := &MultiProxyManager{
		multiProxyMap: make(map[uuid.UUID]*multi.Proxy),
		l:             l,
		c:             c,
		db:            db,
		ctx:           logr.NewContext(ctx, zapr.NewLogger(l.GetGateLogger())),
	}

	id, err := mpm.createNewProxyId()
	if err != nil {
		return mpm, err
	}
	mpm.l.Info("Found a id to use.", "id", id)

	mpm.ownerMP, err = mpm.NewMultiProxy(id)
	if err != nil {
		return &MultiProxyManager{}, err
	}

	tasks.Init()

	mpm.mpm = playermanager.Init(l, db, mpm.ownerMP)

	cfg, err := gate.LoadConfig(mpm.c.GetViper())
	if err != nil {
		mpm.l.Error("load config for gate error", "error", err)
		return mpm, err
	}

	gate, err := gate.New(gate.Options{
		Config:   cfg,
		EventMgr: event.New(),
	})
	if err != nil {
		mpm.l.Error("creating gate instance error", "error", err)
		return mpm, err
	}

	mpm.ownerGate = gate.Java()
	event.Subscribe(mpm.ownerGate.Event(), 0, mpm.onShutdown)

	mpm.tm = task.InitTaskManager(mpm.db, mpm.l, mpm.ownerMP, mpm.ownerGate, mpm.mpm)

	mpm.cm, err = commands.Init(mpm.ownerGate, mpm.l, mpm.db, mpm.mpm, mpm.tm)
	if err != nil {
		return mpm, nil
	}

	mpm.lm, err = listeners.Init(mpm.ownerGate.Event(), mpm.l, mpm.db, mpm.mpm, mpm.ownerMP)
	if err != nil {
		return mpm, err
	}

	multi.SetProxyManager(mpm)

	// start update listener
	mpm.db.CreateListener(multi.UpdateMultiProxyChannel, mpm.createUpdateListener())

	// fill map
	_, err = mpm.GetAllMultiProxiesFromDatabase()
	if err != nil {
		mpm.l.Error("filling up multiproxy map error", "error", err)
	}

	mpm.GetLogger().Info("initialized multiproxy manager", "duration", time.Since(now))
	return mpm, nil
}

func (mpm *MultiProxyManager) createUpdateListener() func(msg *redis.Message) {
	return func(msg *redis.Message) {
		mpm.l.Info(msg.Payload)
		m := msg.Payload
		s := strings.Split(m, "_")

		originProxy := s[0]
		// from own proxy, no update needed
		if mpm.ownerMP.GetId().String() == originProxy {
			return
		}

		id, err := uuid.Parse(s[1])
		if err != nil {
			mpm.l.Error("multiproxy update channel parse uuid error", "parsed uuid", s[1], "error", err)
			return
		}

		k := s[2]

		mp, err := mpm.GetMultiProxy(id)
		if err != nil {
			mpm.l.Error("multiproxy update channel get multiproxy error", "proxyId", id, "error", err)
			return
		}

		// already created
		if k == "new" {
			return
		}

		if k == "delete" {
			err := mpm.DeleteMultiProxy(id)
			if err != nil {
				mpm.l.Error("multiproxy update channel delete multiproxy error", "proxyId", id, "error", err)
			}
			return
		}

		dataKey, err := key.GetProxyKey(k)
		if err != nil {
			mpm.l.Error("multiproxy update channel get data key error", "proxyId", id, "key", k, "error", err)
			return
		}

		mp.Update(dataKey)
	}
}

func (mpm *MultiProxyManager) NewMultiProxy(id uuid.UUID) (*multi.Proxy, error) {
	now := time.Now()

	data := &data.ProxyData{
		Address:     "localhost:25565",
		Maintenance: false,
		Backends:    make([]uuid.UUID, 0),
		Players:     make([]uuid.UUID, 0),
	}

	err := mpm.db.SetProxyData(id, data)
	if err != nil {
		return nil, err
	}

	mp, err := mpm.CreateMultiProxyFromDatabase(id)
	if err != nil {
		return nil, err
	}

	// update every proxies' map
	var m string
	if mpm.ownerMP != nil {
		m = mpm.ownerMP.GetId().String() + "_" + id.String() + "_new"
	} else {
		m = id.String() + "_" + id.String() + "_new"
	}

	err = mpm.db.Publish(multi.UpdateMultiProxyChannel, m)
	if err != nil {
		return nil, err
	}

	mpm.GetLogger().Info("created new multiproxy", "proxyId", id, "duration", time.Since(now))
	return mp, nil
}

func (mpm *MultiProxyManager) DeleteMultiProxy(id uuid.UUID) error {
	now := time.Now()
	for key := range mpm.multiProxyMap {
		if key == id {
			mpm.multiProxyMap[key] = nil
		}
	}

	_, err := mpm.db.GetProxyData(id)
	if err != nil {
		if err == database.ErrDataNotFound {
			return nil
		}

		mpm.l.Error("could not get proxy data")
		return err
	}

	err = mpm.db.DeleteProxyData(id)
	if err != nil {
		return err
	}

	m := mpm.ownerMP.GetId().String() + "_" + id.String() + "_delete"
	err = mpm.db.Publish(multi.UpdateMultiProxyChannel, m)
	if err != nil {
		return err
	}

	mpm.GetLogger().Info("deleted multiproxy", "proxyId", id, "duration", time.Since(now))
	return nil
}

func (mpm *MultiProxyManager) GetOwnerGate() *proxy.Proxy {
	return mpm.ownerGate
}

func (mpm *MultiProxyManager) Start() {
	err := mpm.ownerGate.Start(mpm.ctx)
	if err != nil {
		mpm.l.Error("error starting proxy", "error", err)
	}
}

func (mpm *MultiProxyManager) onShutdown(event *proxy.ShutdownEvent) {
	mpm.Close()
}

func (mpm *MultiProxyManager) GetMultiProxy(id uuid.UUID) (*multi.Proxy, error) {
	mpm.mu.RLock()
	mp, ok := mpm.multiProxyMap[id]
	mpm.mu.RUnlock()

	if ok {
		return mp, nil
	}

	return mpm.CreateMultiProxyFromDatabase(id)
}

func (mpm *MultiProxyManager) CreateMultiProxyFromDatabase(id uuid.UUID) (*multi.Proxy, error) {
	data, err := mpm.db.GetProxyData(id)
	if err != nil {
		return nil, err
	}

	var managerId uuid.UUID
	if mpm.ownerMP == nil {
		managerId = id
	} else {
		managerId = mpm.ownerMP.GetId()
	}

	mp := multi.NewProxy(id, managerId, mpm.db, data)

	mpm.mu.Lock()
	mpm.multiProxyMap[id] = mp
	mpm.mu.Unlock()

	return mp, nil
}

func (mpm *MultiProxyManager) GetAllMultiProxies() []*multi.Proxy {
	var l []*multi.Proxy

	mpm.mu.RLock()
	for _, mp := range mpm.multiProxyMap {
		l = append(l, mp)
	}
	mpm.mu.RUnlock()

	return l
}

func (mpm *MultiProxyManager) GetAllMultiProxiesFromDatabase() ([]*multi.Proxy, error) {
	var l []*multi.Proxy

	i, err := mpm.db.GetAllProxyIds()
	if err != nil {
		return nil, err
	}

	for _, id := range i {
		mp, err := mpm.GetMultiProxy(id)
		if err != nil {
			return nil, err
		}

		l = append(l, mp)
	}

	return l, nil
}

func (mpm *MultiProxyManager) GetMultiBackend(id uuid.UUID) (*multi.Backend, error) {
	return nil, nil
}

func (mpm *MultiProxyManager) GetLogger() *logger.Logger {
	return mpm.l
}

// creates id
func (mpm *MultiProxyManager) createNewProxyId() (uuid.UUID, error) {
	var break_err error

	for {
		id := uuid.New()
		_, err := mpm.GetMultiProxy(id)
		if err == database.ErrDataNotFound {
			return id, nil
		}

		if err != nil {
			break_err = err
			break
		}
	}

	mpm.l.Error("create new proxy id error", "error", break_err)
	return uuid.Nil, break_err
}

func (mpm *MultiProxyManager) Close() {
	mpm.l.Info("stopping mp")
	err := mpm.DeleteMultiProxy(mpm.ownerMP.GetId())
	if err != nil {

	}

	err = mpm.db.Close()
	if err != nil {

	}

	err = mpm.l.Close()
	if err != nil {

	}
}
