package listeners

import (
	"github.com/team-vesperis/vesperis-mp/mp/mp/datasync"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func onServerConnect(event *proxy.ServerConnectedEvent) {
	player := event.Player()
	server := event.Server().ServerInfo().Name()

	err := datasync.RegisterPlayer(proxy_name, server, player)
	if err != nil {
		logger.Error("Failed to register player: ", err)
	}
}

func onDisconnect(event *proxy.DisconnectEvent) {
	player := event.Player()
	if player.CurrentServer() == nil {
		return
	}

	server := player.CurrentServer().Server().ServerInfo().Name()
	err := datasync.UnregisterPlayer(proxy_name, server, player)
	if err != nil {
		logger.Error("Failed to unregister player: ", err)
	}
}
