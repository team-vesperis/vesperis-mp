package listeners

import (
	"github.com/team-vesperis/vesperis-mp/mp/web/datasync"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func onServerConnect(event *proxy.ServerConnectedEvent) {
	datasync.UnregisterPlayer(proxy_name, event.PreviousServer().ServerInfo().Name(), event.Player())
	datasync.RegisterPlayer(proxy_name, event.Server().ServerInfo().Name(), event.Player())
}
