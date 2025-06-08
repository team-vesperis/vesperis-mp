package main

import (
	"context"

	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/mp/ban"
	"github.com/team-vesperis/vesperis-mp/mp/commands"
	"github.com/team-vesperis/vesperis-mp/mp/config"
	"github.com/team-vesperis/vesperis-mp/mp/database"
	"github.com/team-vesperis/vesperis-mp/mp/listeners"
	log "github.com/team-vesperis/vesperis-mp/mp/logger"
	"github.com/team-vesperis/vesperis-mp/mp/mp/datasync"
	"github.com/team-vesperis/vesperis-mp/mp/mp/register"
	"github.com/team-vesperis/vesperis-mp/mp/mp/task"
	"github.com/team-vesperis/vesperis-mp/mp/mp/transfer"
	"github.com/team-vesperis/vesperis-mp/mp/playerdata"
	"github.com/team-vesperis/vesperis-mp/mp/terminal"
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

			event.Subscribe(p.Event(), 0, onShutdown)

			transfer.InitializeTransfer(p, logger, proxy_name)
			commands.InitializeCommands(p, logger, proxy_name)
			listeners.InitializeListeners(p, logger, proxy_name)
			utils.InitializeUtils(p, logger)
			register.InitializeRegister(p, logger, proxy_name)
			ban.InitializeBanManager(logger)
			datasync.InitializeDataSync(proxy, logger, proxy_name)
			task.InitializeTask(proxy, logger, proxy_name)
			playerdata.InitializePlayerData(logger)

			go terminal.HandleTerminalInput(p, logger)

			logger.Info("Successfully created plugin.")
			return nil
		},
	})

	logger.Info("Successfully started " + proxy_name + ".")
	gate.Execute()
}

func shutdown() {
	logger.Info("Stopping " + proxy_name + "...")

	ban.CloseBanManager()
	datasync.CloseDataSync()
	database.CloseDatabases()

	logger.Info("Successfully stopped " + proxy_name + ".")
}

func onShutdown(event *proxy.ShutdownEvent) {
	shutdown()
}
