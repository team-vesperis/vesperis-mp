package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/multiproxy"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
)

func main() {
	ctx, canc := context.WithCancel(context.Background())
	defer canc()

	now := time.Now()
	mp, err := multiproxy.New(ctx)
	if err != nil {
		return
	}

	mp.GetLogger().Info("created MultiProxy", "duration", time.Since(now))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		now = time.Now()
		mp.Shutdown(component.Text{
			Content: "This proxy has been manually shut using the terminal.",
			S:       component.Style{Color: color.Red},
		})

		mp.GetLogger().Info("stopped MultiProxy", "duration", time.Since(now))
		defer os.Exit(0)
	}()

	// blocks
	mp.Start()
}
