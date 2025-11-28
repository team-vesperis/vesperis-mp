package listeners

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/ping"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/favicon"
	"go.minekube.com/gate/pkg/util/uuid"
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

func (lm *ListenerManager) onPing(e *proxy.PingEvent) {
	playerCount := len(lm.mm.GetAllOnlinePlayers(false))

	ping := &ping.ServerPing{
		Description: &component.Text{
			Content: "Vesperis",
			S:       util.StyleColorLightBlue,
		},

		Version: ping.Version{
			Name:     "Vesperis",
			Protocol: e.Connection().Protocol(),
		},

		Players: &ping.Players{
			Online: playerCount,
			Max:    playerCount + 1,
			Sample: []ping.SamplePlayer{
				{
					Name: "§eSupported versions: §a1.21.9 §e& §a1.21.10",
					ID:   uuid.New(),
				},
				{
					Name: "§aVesperis-Proxy-" + lm.mm.GetOwnerMultiProxy().GetId().String(),
					ID:   uuid.New(),
				},
			},
		},

		Favicon: fav,
	}

	e.SetPing(ping)
}
