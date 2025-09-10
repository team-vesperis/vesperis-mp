package multiproxy

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multiplayer"
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
	mpm *multiplayer.MultiPlayerManager
}

func InitManager(ctx context.Context) (*MultiProxyManager, uuid.UUID, error) {

	l, logErr := logger.Init()
	if logErr != nil {
		return &MultiProxyManager{}, uuid.Nil, logErr
	}

	c, cfErr := config.Init(l)
	if cfErr != nil {
		l.Error("config initialization error")
		return &MultiProxyManager{}, uuid.Nil, cfErr
	}

	db, dbErr := database.Init(ctx, c, l)
	if dbErr != nil {
		l.Error("database initialization error")
		return &MultiProxyManager{}, uuid.Nil, dbErr
	}

	mproxym := &MultiProxyManager{
		multiProxyMap: sync.Map{},
		l:             l,
		c:             c,
		db:            db,
		ctx:           logr.NewContext(ctx, zapr.NewLogger(l.GetGateLogger())),
	}

	mproxym.ownerId = mproxym.createNewProxyId()

	mproxym.mpm = multiplayer.InitManager(l, db, mproxym.ownerId)

	cfg, err := gate.LoadConfig(mproxym.c.GetViper())
	if err != nil {
		mproxym.l.Error("load config for gate error", "error", err)
		return mproxym, uuid.Nil, err
	}

	gate, err := gate.New(gate.Options{
		Config:   cfg,
		EventMgr: event.New(),
	})
	if err != nil {
		mproxym.l.Error("creating gate instance error", "error", err)
		return mproxym, uuid.Nil, err
	}

	mproxym.ownerGate = gate.Java()
	event.Subscribe(mproxym.ownerGate.Event(), 0, mproxym.onShutdown)

	mproxym.cm, err = commands.Init(mproxym.ownerGate, mproxym.l, mproxym.db, mproxym.mpm)
	if err != nil {
		return mproxym, uuid.Nil, nil
	}

	mproxym.lm, err = listeners.Init(mproxym.ownerGate.Event(), mproxym.l, mproxym.db, mproxym.mpm, mproxym.ownerId)
	if err != nil {
		return mproxym, uuid.Nil, err
	}

	return mproxym, mproxym.ownerId, nil
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

func (mpm *MultiProxyManager) GetMultiProxy(id uuid.UUID) (*MultiProxy, error) {
	val, ok := mpm.multiProxyMap.Load(id)
	if ok {
		mp, ok := val.(*MultiProxy)
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

func (mpm *MultiProxyManager) Close() {
	mpm.l.Info("Stopping...")
	mpm.db.Close()
	mpm.l.Close()
}
