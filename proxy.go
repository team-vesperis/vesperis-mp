package main

import (
	"context"
	"fmt"

	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/mp/ban"
	"github.com/team-vesperis/vesperis-mp/mp/commands"
	"github.com/team-vesperis/vesperis-mp/mp/config"
	"github.com/team-vesperis/vesperis-mp/mp/database"
	"github.com/team-vesperis/vesperis-mp/mp/listeners"
	log "github.com/team-vesperis/vesperis-mp/mp/logger"
	"github.com/team-vesperis/vesperis-mp/mp/register"
	"github.com/team-vesperis/vesperis-mp/mp/share"
	"github.com/team-vesperis/vesperis-mp/mp/terminal"
	"github.com/team-vesperis/vesperis-mp/mp/transfer"
	"github.com/team-vesperis/vesperis-mp/mp/utils"
	"go.minekube.com/gate/cmd/gate"
	"go.minekube.com/gate/pkg/edition/java/proxy"

	"go.uber.org/zap"
)

var (
	logger     *zap.SugaredLogger
	proxy_name string
	p          *proxy.Proxy
)

func main() {
	logger = log.InitializeLogger()
	config.LoadConfig(logger)
	proxy_name = config.GetProxyName()

	logger.Info("Starting " + proxy_name + "...")
	database.InitializeDatabases(logger)

	proxy.Plugins = append(proxy.Plugins, proxy.Plugin{
		Name: "VesperisMP-" + proxy_name,
		Init: func(ctx context.Context, proxy *proxy.Proxy) error {
			p = proxy
			logger.Info("Creating plugin...")

			event.Subscribe(p.Event(), 0, onShutdown())

			transfer.InitializeTransfer(p, logger, proxy_name)
			commands.InitializeCommands(p, logger)
			listeners.InitializeListeners(p, logger)
			utils.InitializeUtils(p, logger)
			register.InitializeRegister(p, logger)
			ban.InitializeBanManager(logger)

			go share.InitializeShare(logger, p, proxy_name)
			go terminal.HandleTerminalInput(p, logger)

			logger.Info("Successfully created plugin.")
			return nil
		},
	})

	gate.Execute()
}

func shutdown() {
	logger.Info("Stopping " + proxy_name + "...")

	transfer.CloseTransfer()
	share.CloseShare()
	database.CloseDatabases()

	defer func() {
		fmt.Println("Press 'Enter' to exit...")
		fmt.Scanln()
	}()
}

func onShutdown() func(*proxy.ShutdownEvent) {
	return func(event *proxy.ShutdownEvent) {
		shutdown()
	}
}
