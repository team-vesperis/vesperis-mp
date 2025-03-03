package listeners

import (
	"github.com/team-vesperis/vesperis-mp/mp/share"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func onPlayerCountJoin() func(*proxy.PostLoginEvent) {
	return func(event *proxy.PostLoginEvent) {
		share.AddPlayerToPlayerCount()
		share.AddPlayer(event.Player().Username())
	}
}

func onPlayerCountLeave() func(*proxy.DisconnectEvent) {
	return func(event *proxy.DisconnectEvent) {
		share.RemovePlayerFromPlayerCount()
		share.RemovePlayer(event.Player().Username())
	}
}
