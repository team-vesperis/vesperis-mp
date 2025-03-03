package share

import (
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.uber.org/zap"
)

var (
	logger     *zap.SugaredLogger
	p          *proxy.Proxy
	proxy_name string
)

func InitializeShare(log *zap.SugaredLogger, proxy *proxy.Proxy, p_name string) {
	logger = log
	p = proxy
	proxy_name = p_name

	registerProxy()
	registerServers()
}

func CloseShare() {
	unregisterProxy()
	unregisterServers()
}
