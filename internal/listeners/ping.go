package listeners

import (
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/mp/datasync"
	"github.com/team-vesperis/vesperis-mp/internal/utils"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/ping"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/favicon"
	"go.minekube.com/gate/pkg/util/uuid"
)

var fav favicon.Favicon

func initPing() {
	var err error
	fav, err = favicon.FromFile("../../logo.png")
	if err != nil {
		logger.Error("Error loading logo: ", err)
	}
}

func onPing(event *proxy.PingEvent) {
	playerCount, err := datasync.GetTotalPlayerCount()
	if err != nil {
		logger.Error("Error getting total player count for ping: ", err)
	}

	ping := &ping.ServerPing{
		Version: ping.Version{
			Name:     "Vesperis",
			Protocol: event.Ping().Version.Protocol,
		},

		Players: &ping.Players{
			Online: playerCount,
			Max:    playerCount + 1,
			Sample: []ping.SamplePlayer{
				{
					Name: "§eSupported version: §a1.21.5",
					ID:   uuid.New(),
				},
			},
		},

		Description: &component.Text{
			Content: "Vesperis",
			S:       component.Style{Color: utils.GetColorOrange()},
			Extra: []component.Component{
				&component.Text{
					Content: " - ",
					S:       component.Style{Color: color.Aqua},
				},
				&component.Text{
					Content: config.GetProxyName(),
					S:       component.Style{Color: color.Green},
				},
			},
		},

		Favicon: fav,
	}
	event.SetPing(ping)

}
