package listeners

import (
	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/mp/mp/transfer"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.uber.org/zap"
)

var (
	p          *proxy.Proxy
	logger     *zap.SugaredLogger
	proxy_name string
)

func InitializeListeners(proxy *proxy.Proxy, log *zap.SugaredLogger, pn string) {
	p = proxy
	logger = log
	proxy_name = pn

	event.Subscribe(p.Event(), 0, onPing)
	event.Subscribe(p.Event(), 1, onServerConnect)
	event.Subscribe(p.Event(), 1, onDisconnect)
	event.Subscribe(p.Event(), 0, transfer.OnPreShutdown)
	event.Subscribe(p.Event(), 0, transfer.OnChooseInitialServer)
	event.Subscribe(p.Event(), 0, onLogin)

	logger.Info("Successfully registered all listeners.")
}
