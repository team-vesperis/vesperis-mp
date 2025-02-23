package database

import "go.uber.org/zap"

var logger *zap.SugaredLogger

func InitializeDatabases(log *zap.SugaredLogger) {
	logger = log
	logger.Info("Initializing databases...")

	initializeRedis()
	initializeMysql()

	logger.Info("Successfully initialized databases.")
}

func CloseDatabases() {
	logger.Info("Closing databases...")

	closeRedis()
	closeMySQL()

	logger.Info("Successfully closed databases.")
}
