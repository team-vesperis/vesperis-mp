package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/proxy/commands"

	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/gate"
)

type MultiProxy struct {
	// The database used in the mp. Contains a connection with Redis and MySQL. Combines both in functions for fast and safe usage.
	db *database.Database

	// The ID of the mp. Used to differentiate the proxy from others.
	id string

	// The logger used in the mp
	l *logger.Logger

	// gate proxy used in the mp
	p *proxy.Proxy

	// config used in the mp. Determines the database connection variables, proxy id, etc.
	c *config.Config

	ctx context.Context

	cm commands.CommandManager
}

func New(ctx context.Context) (MultiProxy, error) {
	l, logErr := logger.Init()
	if logErr != nil {
		return MultiProxy{}, logErr
	}

	c := config.Init(l)
	db, dbErr := database.Init(ctx, c, l)

	if dbErr != nil {
		l.Error("database initialization error")
		return MultiProxy{}, dbErr
	}

	lr := zapr.NewLogger(l.GetLogger())
	ctx = logr.NewContext(ctx, lr)

	return MultiProxy{
		db:  db,
		id:  c.GetProxyId(),
		l:   l,
		c:   c,
		ctx: ctx,
	}, nil
}

func main() {
	ctx, canc := context.WithCancel(context.Background())
	defer canc()

	// Create the MultiProxy structure and initialize its values
	mp, err := New(ctx)
	if err != nil {
		return
	}
	mp.l.Info("created MultiProxy")

	// Set correct bind.
	bind := "0.0.0.0:25565"
	mp.c.SetBind(bind)

	cfg, err := gate.LoadConfig(mp.c.GetViper())
	mp.l.Info(cfg.Config.Bind)
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

	mp.cm = commands.Init(mp.p, mp.l, mp.db)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		mp.p.Shutdown(&component.Text{
			Content: "This proxy has been manually shut using the terminal.",
			S:       component.Style{Color: color.Red},
		})
		mp.l.Info("stopped MultiProxy")
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
