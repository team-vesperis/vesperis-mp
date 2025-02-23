package main

import (
	"github.com/team-vesperis/vesperis-mp/mp/config"
	"github.com/team-vesperis/vesperis-mp/mp/database"
	log "github.com/team-vesperis/vesperis-mp/mp/logger"
	"go.uber.org/zap"
)

var (
	logger     *zap.SugaredLogger
	proxy_name string
)

func main() {
	logger = log.InitializeLogger()
	config.LoadConfig(logger)
	proxy_name = config.GetProxyName()

	logger.Info("Starting " + proxy_name + "...")
	database.InitializeDatabases(logger)
}
