package main

import (
	"context"
	"os"

	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/internal/ban"
	"github.com/team-vesperis/vesperis-mp/internal/commands"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/listeners"
	log "github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/mp/datasync"
	"github.com/team-vesperis/vesperis-mp/internal/mp/register"
	"github.com/team-vesperis/vesperis-mp/internal/mp/task"
	"github.com/team-vesperis/vesperis-mp/internal/playerdata"
	"github.com/team-vesperis/vesperis-mp/internal/terminal"
	"github.com/team-vesperis/vesperis-mp/internal/transfer"
	"github.com/team-vesperis/vesperis-mp/internal/utils"

	"go.minekube.com/gate/cmd/gate"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
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
	err := database.InitializeDatabases(logger)
	if err != nil {
		shutdown(false)
	}

	used := datasync.IsProxyAvailable(proxy_name)
	if used {
		logger.Warn("Proxy name is already used! Changing name to new one.")
		proxy_name = "proxy_" + uuid.New().String()
		config.SetProxyName(proxy_name)
	}

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

	gate.Execute()
}

func shutdown(alreadyRunning bool) {
	logger.Info("Stopping " + proxy_name + "...")

	if alreadyRunning {
		ban.CloseBanManager()
		datasync.CloseDataSync()
		database.CloseDatabases()
	}

	logger.Info("Successfully stopped " + proxy_name + ".")

	if !alreadyRunning {
		os.Exit(0)
	}
}

func onShutdown(event *proxy.ShutdownEvent) {
	shutdown(true)
}
