package listeners

import (
	"go.minekube.com/gate/pkg/edition/java/ping"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/favicon"
)

var fav favicon.Favicon

func (lm *ListenerManager) initFavicon() error {
	var f string
	err := lm.db.GetData("favicon", &f)
	if err != nil {
		lm.l.Error("get favicon string from database error", "error", err)
		return err
	}

	fav, err = favicon.Parse(f)
	if err != nil {
		lm.l.Error("init favicon error", "error", err)
		return err
	}

	return nil
}

func (lm *ListenerManager) onPing(event *proxy.PingEvent) {
	ping := &ping.ServerPing{
		Version: ping.Version{
			Name:     "Vesperis",
			Protocol: event.Connection().Protocol(),
		},

		Favicon: fav,
	}

	event.SetPing(ping)
}

// func onPing(event *proxy.PingEvent) {
// 	playerCount, err := datasync.GetTotalPlayerCount()
// 	if err != nil {
// 		logger.Error("Error getting total player count for ping: ", err)
// 	}

// 	ping := &ping.ServerPing{
// 		Version: ping.Version{
// 			Name:     "Vesperis",
// 			Protocol: event.Ping().Version.Protocol,
// 		},

// 		Players: &ping.Players{
// 			Online: playerCount,
// 			Max:    playerCount + 1,
// 			Sample: []ping.SamplePlayer{
// 				{
// 					Name: "§eSupported version: §a1.21.5",
// 					ID:   uuid.New(),
// 				},
// 			},
// 		},

// 		Description: &component.Text{
// 			Content: "Vesperis",
// 			S:       component.Style{Color: utils.GetColorOrange()},
// 			Extra: []component.Component{
// 				&component.Text{
// 					Content: " - ",
// 					S:       component.Style{Color: color.Aqua},
// 				},
// 				&component.Text{
// 					Content: config.GetProxyName(),
// 					S:       component.Style{Color: color.Green},
// 				},
// 			},
// 		},

// 		Favicon: fav,
// 	}
// 	event.SetPing(ping)

// }
