package main

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi/manager"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task/tasks"
	"github.com/team-vesperis/vesperis-mp/internal/proxy/commands"
	"github.com/team-vesperis/vesperis-mp/internal/proxy/listeners"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/gate"
	"go.uber.org/zap/zapcore"
)

type Manager struct {
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
	cf *config.Config

	multi *manager.MultiManager

	task *task.TaskManager
}

func Init(ctx context.Context, cf *config.Config, l *logger.Logger, db *database.Database) (*Manager, error) {
	m := &Manager{
		ctx: logr.NewContext(ctx, zapr.NewLogger(l.GetGateLogger())),
		cf:  cf,
		l:   l,
		db:  db,
	}

	var err error
	m.multi, err = manager.Init(cf, db, l)
	if err != nil {
		return m, err
	}

	tasks.Init()

	cfg, err := gate.LoadConfig(m.cf.GetViper())
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

	m.task = task.InitTaskManager(m.db, m.l, m.multi.GetOwnerMultiProxy(), m.ownerGate, m.multi)

	m.command, err = commands.Init(m.ownerGate, m.l, m.db, m.multi, m.task)
	if err != nil {
		return m, err
	}

	m.listener, err = listeners.Init(m.ownerGate.Event(), m.l, m.db, m.multi, m.ownerGate, m.task)
	if err != nil {
		return m, err
	}

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
		m.l.Error("multimanager close error", "error", err)
	}

	err = m.db.Close()
	if err != nil {
		m.l.Error("database close error", "error", err)
	}
}

func (m *Manager) SetDebug(debug bool) {
	if debug {
		m.l.SetLevel(zapcore.DebugLevel)
	} else {
		m.l.SetLevel(zapcore.InfoLevel)
	}
}
