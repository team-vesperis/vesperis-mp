package multiproxy

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multiplayer"
	"github.com/team-vesperis/vesperis-mp/internal/proxy/commands"
	"github.com/team-vesperis/vesperis-mp/internal/proxy/listeners"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/gate"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiProxy struct {
	// The database used in the mp.
	// Contains a connection with Redis and Postgres.
	// Combines both in functions for fast and safe usage.
	db *database.Database

	// The id of the mp.
	// Used to differentiate the proxy from others.
	id string

	// The logger used in the mp.
	l *logger.Logger

	// The gate proxy used in the mp.
	p *proxy.Proxy

	// The config used in the mp.
	// Determines the database connection variables, proxy id, etc.
	c *config.Config

	// The context used in the mp.
	// Contains a cancel and logger.
	ctx context.Context

	// The command manager used in the mp.
	cm *commands.CommandManager

	// The listener manager in the mp.
	lm *listeners.ListenerManager

	// The multiplayer manager used in the mp.
	mpm *multiplayer.MultiPlayerManager
}

func New(ctx context.Context) (MultiProxy, error) {
	l, logErr := logger.Init()
	if logErr != nil {
		return MultiProxy{}, logErr
	}

	c, cfErr := config.Init(l)
	if cfErr != nil {
		l.Error("config initialization error")
		return MultiProxy{}, cfErr
	}

	db, dbErr := database.Init(ctx, c, l)

	if dbErr != nil {
		l.Error("database initialization error")
		return MultiProxy{}, dbErr
	}

	// if id == "" || db.CheckIfProxyIdIsAvailable(id) == false {
	// 	// set to a unique id
	// 	// TODO: create standalone function
	// 	// for creating unique id to check if the new id is not used
	// 	id = "proxy_" + uuid.New().Undashed()
	// }

	id := "proxy_" + uuid.New().Undashed()

	lr := zapr.NewLogger(l.GetLogger())
	ctx = logr.NewContext(ctx, lr)

	mpm := multiplayer.InitMultiPlayerManager(l, db)

	now := time.Now()
	for range 100000 {
		mpm.GetAllMultiPlayersWay1()
	}
	l.Info("way 1", "duration", time.Since(now))

	now = time.Now()
	for range 100000 {
		mpm.GetAllMultiPlayersWay2()
	}
	l.Info("way 2", "duration", time.Since(now))

	mp := MultiProxy{
		db:  db,
		id:  id,
		l:   l,
		c:   c,
		ctx: ctx,
		mpm: mpm,
	}

	cfg, err := gate.LoadConfig(mp.c.GetViper())
	gate, err := gate.New(gate.Options{
		Config:   cfg,
		EventMgr: event.New(),
	})
	if err != nil {
		mp.l.Error("creating gate instance error", "error", err)
		return mp, err
	}

	mp.p = gate.Java()
	event.Subscribe(mp.p.Event(), 0, mp.onShutdown)

	mp.cm = commands.Init(mp.p, mp.l, mp.db, mp.mpm)
	mp.lm = listeners.Init(mp.p.Event(), mp.l, mp.db, mp.mpm, mp.id)

	return mp, nil
}

func (mp *MultiProxy) Start() {
	err := mp.p.Start(mp.ctx)
	if err != nil {
		mp.l.Error("error starting proxy", "error", err)
	}
}

func (mp *MultiProxy) Shutdown(reason component.Text) {
	mp.p.Shutdown(&reason)
}

func (mp *MultiProxy) GetLogger() *logger.Logger {
	return mp.l
}

func (mp *MultiProxy) onShutdown(event *proxy.ShutdownEvent) {
	mp.l.Info("Stopping...")
	mp.db.Close()
	mp.l.Close()
}
