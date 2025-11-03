package main

import (
	"context"

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
	"github.com/team-vesperis/vesperis-mp/internal/proxy/commands"
	"github.com/team-vesperis/vesperis-mp/internal/proxy/listeners"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/gate"
)

type Manager struct {
	ownerMP *multi.Proxy

	ownerGate *proxy.Proxy

	// The command manager used in the mp.
	command *commands.CommandManager

	// The listener manager in the mp.
	listener *listeners.ListenerManager

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

func Init(ctx context.Context, c *config.Config, l *logger.Logger, db *database.Database) (*Manager, error) {
	m := &Manager{
		ctx: logr.NewContext(ctx, zapr.NewLogger(l.GetGateLogger())),
		c:   c,
		l:   l,
		db:  db,
	}

	m.multi = manager.Init(db, l)

	id, err := m.multi.CreateNewProxyId()
	if err != nil {
		return m, err
	}
	m.l.Info("Found a id to use.", "id", id)

	m.ownerMP, err = m.multi.NewMultiProxy(id)
	if err != nil {
		return &Manager{}, err
	}

	m.multi.StartProxy()
	m.multi.StartBackend()
	tasks.Init()
	m.multi.StartPlayer()

	cfg, err := gate.LoadConfig(m.c.GetViper())
	if err != nil {
		m.l.Error("load config for gate error", "error", err)
		return m, err
	}

	gate, err := gate.New(gate.Options{
		Config:   cfg,
		EventMgr: event.New(),
	})
	if err != nil {
		m.l.Error("creating gate instance error", "error", err)
		return m, err
	}

	m.ownerGate = gate.Java()
	event.Subscribe(m.ownerGate.Event(), 0, m.onShutdown)

	m.task = task.InitTaskManager(m.db, m.l, m.ownerMP, m.ownerGate, m.multi)

	m.command, err = commands.Init(m.ownerGate, m.l, m.db, m.multi, m.task)
	if err != nil {
		return m, nil
	}

	m.listener, err = listeners.Init(m.ownerGate.Event(), m.l, m.db, m.multi, m.ownerMP)
	if err != nil {
		return m, err
	}

	m.transfer = transfer.Init(l, db, m.task, m.ownerGate, m.multi)
	event.Subscribe(m.ownerGate.Event(), 0, m.transfer.OnChooseInitialServer)
	event.Subscribe(m.ownerGate.Event(), 0, m.transfer.OnPreShutdown)

	m.multi.InitHeartBeatManager()

	return m, nil
}

func (m *Manager) start() {
	err := m.ownerGate.Start(m.ctx)
	if err != nil {
		m.l.Error("error starting proxy", "error", err)
	}
}

func (m *Manager) onShutdown(event *proxy.ShutdownEvent) {
	m.close()
}

func (m *Manager) close() {
	m.l.Info("stopping mp")

	err := m.multi.Close()
	if err != nil {
		m.l.Error("", "error", err)
	}

	err = m.db.Close()
	if err != nil {

	}

	err = m.l.Close()
	if err != nil {

	}
}
