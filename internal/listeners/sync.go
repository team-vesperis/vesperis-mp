package listeners

import (
	"github.com/team-vesperis/vesperis-mp/internal/mp/datasync"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func onServerConnect(event *proxy.ServerConnectedEvent) {
	player := event.Player()
	newServer := event.Server()
	oldServer := event.PreviousServer()

	if oldServer != nil {
		logger.Info("Player is moving server from ", oldServer, " to ", newServer)
		err := datasync.UnregisterPlayer(proxy_name, oldServer.ServerInfo().Name(), player)
		if err != nil {
			logger.Error("Failed to unregister player: ", err)
		}
	}

	err := datasync.RegisterPlayer(proxy_name, newServer.ServerInfo().Name(), player)
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
