package listeners

import (
	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/mp/transfer"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.uber.org/zap"
)

var (
	p      *proxy.Proxy
	logger *zap.SugaredLogger
)

func InitializeListeners(proxy *proxy.Proxy, log *zap.SugaredLogger) {
	p = proxy
	logger = log

	event.Subscribe(p.Event(), 0, onPing())
	event.Subscribe(p.Event(), 1, onPlayerCountJoin())
	event.Subscribe(p.Event(), 1, onPlayerCountLeave())
	event.Subscribe(p.Event(), 1, transfer.OnPreShutdown())
	event.Subscribe(p.Event(), 0, transfer.OnChooseInitialServer())
	event.Subscribe(p.Event(), 0, onLogin())
	event.Subscribe(p.Event(), 0, onPluginMessage())

	logger.Info("Successfully registered all listeners.")
}

func onPluginMessage() func(*proxy.PluginMessageEvent) {
	return func(event *proxy.PluginMessageEvent) {
		logger.Info(event)
	}
}
