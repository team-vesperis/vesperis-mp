package listeners

import (
	"github.com/team-vesperis/vesperis-mp/mp/config"
	"github.com/team-vesperis/vesperis-mp/mp/share"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/ping"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func onPing() func(*proxy.PingEvent) {
	return func(event *proxy.PingEvent) {
		ping := &ping.ServerPing{
			Version: ping.Version{
				Name:     "1.21.4",
				Protocol: 769,
			},
			Players: &ping.Players{
				Online: share.GetPlayerCount(),
				Max:    200,
				Sample: []ping.SamplePlayer{},
			},
			Description: &component.Text{
				Content: config.GetProxyName(),
			},
		}
		event.SetPing(ping)
	}
}
