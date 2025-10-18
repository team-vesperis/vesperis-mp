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
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/manager"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/multi/transfer"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"github.com/team-vesperis/vesperis-mp/internal/proxy/commands"
	"github.com/team-vesperis/vesperis-mp/internal/proxy/listeners"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/gate"
)

type MultiManager struct {
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

	multi *manager.MultiManager

	task *task.TaskManager

	transfer *transfer.TransferManager
}

func main() {
	ctx, canc := context.WithCancel(context.Background())
	defer canc()

	l, logErr := logger.Init()
	if logErr != nil {
		return
	}

	cf, cfErr := config.Init(l)
	if cfErr != nil {
		l.Error("config initialization error")
		return
	}

	db, dbErr := database.Init(ctx, cf, l)
	if dbErr != nil {
		l.Error("database initialization error")
		return
	}

	mm, err := Init(ctx, cf, l, db)
	if err != nil {
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		now := time.Now()
		mm.ownerGate.Shutdown(&component.Text{
			Content: "This proxy has been manually shut using the terminal.",
			S:       util.StyleColorRed,
		})

		l.Info("stopped MultiProxy", "duration", time.Since(now))
		defer os.Exit(0)
	}()

	l.GetGateLogger().Info("starting internal gate proxy")
	mm.start()
}

func Init(ctx context.Context, c *config.Config, l *logger.Logger, db *database.Database) (*MultiManager, error) {
	mm := &MultiManager{
		ctx: logr.NewContext(ctx, zapr.NewLogger(l.GetGateLogger())),
		c:   c,
		l:   l,
		db:  db,
	}

	mm.multi = manager.Init(db, l)

	id, err := mm.multi.CreateNewProxyId()
	if err != nil {
		return mm, err
	}
	mm.l.Info("Found a id to use.", "id", id)

	mm.ownerMP, err = mm.multi.NewMultiProxy(id)
	if err != nil {
		return &MultiManager{}, err
	}

	mm.multi.StartProxy()
	tasks.Init()
	mm.multi.StartPlayer()

	cfg, err := gate.LoadConfig(mm.c.GetViper())
	if err != nil {
		mm.l.Error("load config for gate error", "error", err)
		return mm, err
	}

	gate, err := gate.New(gate.Options{
		Config:   cfg,
		EventMgr: event.New(),
	})
	if err != nil {
		mm.l.Error("creating gate instance error", "error", err)
		return mm, err
	}

	mm.ownerGate = gate.Java()
	event.Subscribe(mm.ownerGate.Event(), 0, mm.onShutdown)

	mm.task = task.InitTaskManager(mm.db, mm.l, mm.ownerMP, mm.ownerGate, mm.multi)

	mm.cm, err = commands.Init(mm.ownerGate, mm.l, mm.db, mm.multi, mm.task)
	if err != nil {
		return mm, nil
	}

	mm.lm, err = listeners.Init(mm.ownerGate.Event(), mm.l, mm.db, mm.multi, mm.ownerMP)
	if err != nil {
		return mm, err
	}

	mm.transfer = transfer.Init(l, db, mm.ownerGate, mm.multi)
	event.Subscribe(mm.ownerGate.Event(), 0, mm.transfer.OnChooseInitialServer)
	event.Subscribe(mm.ownerGate.Event(), 0, mm.transfer.OnPreShutdown)

	return mm, nil
}

func (mm *MultiManager) start() {
	err := mm.ownerGate.Start(mm.ctx)
	if err != nil {
		mm.l.Error("error starting proxy", "error", err)
	}
}

func (mm *MultiManager) onShutdown(event *proxy.ShutdownEvent) {
	mm.close()
}

func (mm *MultiManager) close() {
	mm.l.Info("stopping mp")
	err := mm.multi.DeleteMultiProxy(mm.ownerMP.GetId())
	if err != nil {

	}

	err = mm.db.Close()
	if err != nil {

	}

	err = mm.l.Close()
	if err != nil {

	}
}
