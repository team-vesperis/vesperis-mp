package listeners

import (
	"github.com/team-vesperis/vesperis-mp/internal/multiplayer"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func (lm *ListenerManager) onLogin(e *proxy.LoginEvent) {
	p := e.Player()
	id := p.ID().String()

	mp := lm.mpm.GetMultiPlayer(id)
	// player hasn't joined before -> creating default mp
	if mp == nil {
		var err error
		mp, err = multiplayer.New(p, lm.db, lm.mpm)
		if err != nil {
			lm.l.Error("error creating multiplayer", "playerId", id, "error", err)
			e.Deny(&component.Text{
				Content: "There was an error with the login. Please try again later.",
				S: component.Style{
					Color: color.Red,
				},
			})
			return
		}
	}

	mp.SetOnline(true, true)
}

func (lm *ListenerManager) onDisconnect(e *proxy.DisconnectEvent) {
	id := e.Player().ID().String()
	mp := lm.mpm.GetMultiPlayer(id)
	if mp != nil {
		mp.SetOnline(false, true)
	}
}
