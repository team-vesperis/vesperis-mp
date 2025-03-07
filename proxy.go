package main

import (
	"context"
	"time"

	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/mp/ban"
	"github.com/team-vesperis/vesperis-mp/mp/commands"
	"github.com/team-vesperis/vesperis-mp/mp/config"
	"github.com/team-vesperis/vesperis-mp/mp/database"
	"github.com/team-vesperis/vesperis-mp/mp/listeners"
	log "github.com/team-vesperis/vesperis-mp/mp/logger"
	"github.com/team-vesperis/vesperis-mp/mp/playerdata"
	"github.com/team-vesperis/vesperis-mp/mp/register"
	"github.com/team-vesperis/vesperis-mp/mp/terminal"
	"github.com/team-vesperis/vesperis-mp/mp/transfer"
	"github.com/team-vesperis/vesperis-mp/mp/utils"
	"github.com/team-vesperis/vesperis-mp/mp/web/datasync"
	"github.com/team-vesperis/vesperis-mp/mp/web/task"
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
			commands.InitializeCommands(p, logger)
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

	go testTask()
	logger.Info("Successfully started " + proxy_name + ".")
	gate.Execute()
}

func testTask() {
	time.Sleep(25 * time.Second)

	messageTask := &task.MessageTask{
		OriginPlayerName: "Bores",
		TargetPlayerName: "BorisP",
		Message:          "hello!",
	}

	err := messageTask.CreateTask("proxy_1")
	if err != nil {
		if err.Error() == task.Player_Not_Found {
			logger.Info("player not found!")
		} else {
			logger.Info(err)
		}
	}

	err = messageTask.CreateTask("sdkfj")
	if err != nil {
		if err.Error() == task.Player_Not_Found {
			logger.Info("player not found!")
		} else {
			logger.Info(err)
		}
	}
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
