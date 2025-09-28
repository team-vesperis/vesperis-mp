package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/multi/proxymanager"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/common/minecraft/component"
)

func main() {
	ctx, canc := context.WithCancel(context.Background())
	defer canc()

	now := time.Now()

	mpm, err := proxymanager.Init(ctx)
	if err != nil {
		return
	}

	mpm.GetLogger().Info("initialized MultiProxy manager", "duration", time.Since(now))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		now = time.Now()
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
