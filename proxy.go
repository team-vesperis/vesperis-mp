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
	"github.com/team-vesperis/vesperis-mp/mp/share"
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

			event.Subscribe(p.Event(), 0, onShutdown())

			transfer.InitializeTransfer(p, logger, proxy_name)
			commands.InitializeCommands(p, logger)
			listeners.InitializeListeners(p, logger)
			utils.InitializeUtils(p, logger)
			register.InitializeRegister(p, logger)
			ban.InitializeBanManager(logger)
			datasync.InitializeDataSync(proxy, logger)
			task.InitializeTask(proxy, logger)
			playerdata.InitializePlayerData(logger)

			go share.InitializeShare(logger, p, proxy_name)
			go terminal.HandleTerminalInput(p, logger)

			logger.Info("Successfully created plugin.")
			return nil
		},
	})

	go testTask()
	gate.Execute()
}

func testTask() {
	time.Sleep(30 * time.Second)

	messageTask := &task.MessageTask{
		OriginPlayerName: "Bores",
		TargetPlayerName: "BorisP",
		Message:          "hello!",
	}

	err := messageTask.CreateTask()
	if err != nil {
		logger.Info(err)
	}
}

func shutdown() {
	logger.Info("Stopping " + proxy_name + "...")

	close(ban.Quit) // close unban checker
	datasync.CloseDataSync()
	share.CloseShare()
	database.CloseDatabases()
}

func onShutdown() func(*proxy.ShutdownEvent) {
	return func(event *proxy.ShutdownEvent) {
		shutdown()
	}
}
