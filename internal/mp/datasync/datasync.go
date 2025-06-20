package datasync

import (
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.uber.org/zap"
)

var (
	p          *proxy.Proxy
	logger     *zap.SugaredLogger
	proxy_name string
)

func InitializeDataSync(proxy *proxy.Proxy, log *zap.SugaredLogger, pn string) {
	p = proxy
	logger = log
	proxy_name = pn

	registerProxy(proxy_name)
	logger.Info("Successfully initialized data sync.")
}

func CloseDataSync() {
	unregisterProxy(proxy_name)
}
