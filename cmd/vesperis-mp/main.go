package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/common/minecraft/component"
	"go.uber.org/zap/zapcore"
)

func main() {
	ctx, canc := context.WithCancel(context.Background())
	defer canc()

	l, err := logger.Init()
	if err != nil {
		return
	}

	cf, err := config.Init(l)
	if err != nil {
		l.Error("config initialization error", "error", err)
		return
	}

	if cf.IsInDebug() {
		l.SetLevel(zapcore.DebugLevel)
		l.Debug("debug mode active")
	}

	db, err := database.Init(ctx, cf, l)
	if err != nil {
		l.Error("database initialization error", "error", err)
		return
	}

	m, err := Init(ctx, cf, l, db)
	if err != nil {
		l.Error("manager initialization error", "error", err)
		return
	}

	cf.GetViper().OnConfigChange(func(in fsnotify.Event) {
		m.l.Debug("config changed")
		m.SetDebug(cf.IsInDebug())
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		now := time.Now()
		m.ownerGate.Shutdown(&component.Text{
			Content: "This proxy has been manually shut using the terminal.",
			S:       util.StyleColorRed,
		})

		l.Info("stopped MultiProxy", "duration", time.Since(now))
		l.Close()
		defer os.Exit(0)
	}()

	go func() {
		time.Sleep(30 * time.Second)
		panic("")
	}()

	l.GetGateLogger().Info("starting internal gate proxy")
	m.start()
}
