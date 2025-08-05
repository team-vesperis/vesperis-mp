package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
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

	"go.minekube.com/common/minecraft/color"
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

	// The player permission manager used in the mp.
	ppm *multiplayer.PlayerPermissionManager
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

	id := c.GetProxyId()
	// if id == "" || db.CheckIfProxyIdIsAvailable(id) == false {
	// 	// set to a unique id
	// 	// TODO: create standalone function
	// 	// for creating unique id to check if the new id is not used
	// 	id = "proxy_" + uuid.New().Undashed()
	// }

	id = "proxy_" + uuid.New().Undashed()

	lr := zapr.NewLogger(l.GetLogger())
	ctx = logr.NewContext(ctx, lr)

	ppm := multiplayer.InitPlayerPermissionManager(db, l)

	mpm := multiplayer.InitMultiPlayerManager(l, db)

	return MultiProxy{
		db:  db,
		id:  id,
		l:   l,
		c:   c,
		ctx: ctx,
		ppm: ppm,
		mpm: mpm,
	}, nil
}

func main() {
	ctx, canc := context.WithCancel(context.Background())
	defer canc()

	now := time.Now()

	// Create the MultiProxy structure and initialize its values
	mp, err := New(ctx)
	if err != nil {
		return
	}
	mp.l.Info("created MultiProxy", "duration", time.Since(now))

	cfg, err := gate.LoadConfig(mp.c.V)
	gate, err := gate.New(gate.Options{
		Config:   cfg,
		EventMgr: event.New(),
	})
	if err != nil {
		mp.l.Error("creating gate instance error", "error", err)
		return
	}

	mp.p = gate.Java()
	event.Subscribe(mp.p.Event(), 0, mp.onShutdown)

	mp.cm = commands.Init(mp.p, mp.l, mp.db, mp.ppm, mp.mpm)
	mp.lm = listeners.Init(mp.p.Event(), mp.l, mp.db, mp.ppm, mp.mpm, mp.id)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		now = time.Now()
		mp.p.Shutdown(&component.Text{
			Content: "This proxy has been manually shut using the terminal.",
			S:       component.Style{Color: color.Red},
		})
		mp.l.Info("stopped MultiProxy", "duration", time.Since(now))
		os.Exit(0)
	}()

	// blocks
	mp.p.Start(mp.ctx)
}

func (mp *MultiProxy) onShutdown(event *proxy.ShutdownEvent) {
	mp.l.Info("Stopping...")
	mp.db.Close()
	mp.l.Close()
}
