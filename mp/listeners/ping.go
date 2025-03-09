package listeners

import (
	"github.com/team-vesperis/vesperis-mp/mp/config"
	"github.com/team-vesperis/vesperis-mp/mp/mp/datasync"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/ping"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func onPing(event *proxy.PingEvent) {
	playerCount, err := datasync.GetTotalPlayerCount()
	if err != nil {
		logger.Error("Error getting total player count for ping: ", err)
	}

	ping := &ping.ServerPing{
		Version: ping.Version{
			Name:     "1.21.4",
			Protocol: 769,
		},
		Players: &ping.Players{
			Online: playerCount,
			Max:    200,
			Sample: []ping.SamplePlayer{},
		},
		Description: &component.Text{
			Content: config.GetProxyName(),
		},
	}
	event.SetPing(ping)

}
