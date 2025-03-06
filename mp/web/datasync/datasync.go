package datasync

import (
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.uber.org/zap"
)

var (
	p      *proxy.Proxy
	logger *zap.SugaredLogger
)

func InitializeDataSync(proxy *proxy.Proxy, log *zap.SugaredLogger) {
	p = proxy
	logger = log

	logger.Info("Successfully initialized sync.")
}

func CloseDataSync() {
}

func sync(data []string) error {

	return nil
}
