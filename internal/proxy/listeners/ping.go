package listeners

// var fav favicon.Favicon

// func initPing() {
// 	var err error
// 	fav, err = favicon.Parse("data:image/png;base64,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")
// 	if err != nil {
// 		logger.Error("Error loading logo: ", err)
// 	}
// }

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
