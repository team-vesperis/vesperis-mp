package database

import "go.uber.org/zap"

var logger *zap.SugaredLogger

func InitializeDatabases(log *zap.SugaredLogger) error {
	logger = log
	logger.Info("Initializing databases...")

	err := initializeRedis()
	if err != nil {
		return err
	}

	err = initializeMysql()
	if err != nil {
		return err
	}

	logger.Info("Successfully initialized databases.")
	return nil
}

func CloseDatabases() {
	logger.Info("Closing databases...")

	closeListeners()

	closeRedis()
	closeMySQL()

	logger.Info("Successfully closed databases.")
}
