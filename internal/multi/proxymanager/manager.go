package proxymanager

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/playermanager"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/proxy/commands"
	"github.com/team-vesperis/vesperis-mp/internal/proxy/listeners"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/gate"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiProxyManager struct {
	multiProxyMap sync.Map

	// the id of the mp that has this manager
	ownerId uuid.UUID

	ownerMultiProxy *multi.Proxy

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

func InitMultiProxyManager(ctx context.Context) (*MultiProxyManager, error) {
	l, logErr := logger.Init()
	if logErr != nil {
		return &MultiProxyManager{}, logErr
	}

	c, cfErr := config.Init(l)
	if cfErr != nil {
		l.Error("config initialization error")
		return &MultiProxyManager{}, cfErr
	}

	l.Info("initializing database")
	db, dbErr := database.Init(ctx, c, l)
	if dbErr != nil {
		l.Error("database initialization error")
		return &MultiProxyManager{}, dbErr
	}

	mpm := &MultiProxyManager{
		multiProxyMap: sync.Map{},
		l:             l,
		c:             c,
		db:            db,
		ctx:           logr.NewContext(ctx, zapr.NewLogger(l.GetGateLogger())),
	}

	mpm.ownerId = mpm.createNewProxyId()

	mpm.mpm = playermanager.InitMultiPlayerManager(l, db, mpm.ownerId)

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

	mpm.cm, err = commands.Init(mpm.ownerGate, mpm.l, mpm.db, mpm.mpm, mpm.tm)
	if err != nil {
		return mpm, nil
	}

	mpm.lm, err = listeners.Init(mpm.ownerGate.Event(), mpm.l, mpm.db, mpm.mpm, mpm.mpm.GetOwnerProxyId(), mpm.ownerMultiProxy)
	if err != nil {
		return mpm, err
	}

	address := mpm.ownerGate.Config().Bind
	mpm.ownerMultiProxy = mpm.NewMultiProxy(address, mpm.ownerId)

	return mpm, nil
}

func (mpm *MultiProxyManager) NewMultiProxy(address string, id uuid.UUID) *multi.Proxy {
	mp := multi.NewMultiProxy(address, id)
	mpm.multiProxyMap.Store(id, mp)
	return mp
}

func (mpm *MultiProxyManager) GetOwnerGate() *proxy.Proxy {
	return mpm.ownerGate
}

func (mpm *MultiProxyManager) GetOwnerMultiProxy() *multi.Proxy {
	return mpm.ownerMultiProxy
}

func (mpm *MultiProxyManager) Start() {
	go func() {
		time.Sleep(5 * time.Second)
		mpm.tm = task.InitTaskManager(mpm.db, mpm.l, mpm.ownerId, mpm.ownerGate, mpm.mpm)
	}()

	err := mpm.ownerGate.Start(mpm.ctx)
	if err != nil {
		mpm.l.Error("error starting proxy", "error", err)
	}
}

func (mpm *MultiProxyManager) onShutdown(event *proxy.ShutdownEvent) {
	mpm.Close()
}

func (mpm *MultiProxyManager) GetMultiProxy(id uuid.UUID) (*multi.Proxy, error) {
	val, ok := mpm.multiProxyMap.Load(id)
	if ok {
		mp, ok := val.(*multi.Proxy)
		if ok {
			return mp, nil
		}

		mpm.multiProxyMap.Delete(id)
	}

	return nil, nil
}

func (mpm *MultiProxyManager) GetLogger() *logger.Logger {
	return mpm.l
}

// creates id
func (mpm *MultiProxyManager) createNewProxyId() uuid.UUID {
	for {
		id := uuid.New()
		mp, _ := mpm.GetMultiProxy(id)
		if mp == nil {
			return id
		}
	}
}

func (mpm *MultiProxyManager) Close() {
	mpm.l.Info("stopping mp")
	mpm.db.Close()
	mpm.l.Close()
}
