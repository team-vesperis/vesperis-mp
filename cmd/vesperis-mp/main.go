package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi/proxymanager"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/common/minecraft/component"
)

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

	mpm, err := proxymanager.Init(ctx, l, cf, db)
	if err != nil {
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		now := time.Now()
		mpm.GetOwnerGate().Shutdown(&component.Text{
			Content: "This proxy has been manually shut using the terminal.",
			S:       util.StyleColorRed,
		})

		mpm.GetLogger().Info("stopped MultiProxy", "duration", time.Since(now))
		defer os.Exit(0)
	}()

	mpm.GetLogger().GetGateLogger().Info("starting internal gate proxy")
	mpm.Start()
}
