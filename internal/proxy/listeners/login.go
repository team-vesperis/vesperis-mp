package listeners

import (
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/multiplayer"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func (lm *ListenerManager) onLogin(e *proxy.LoginEvent) {
	p := e.Player()
	id := p.ID().String()
	_, err := lm.db.GetPlayerData(id)
	if err == database.ErrDataNotFound {
		mp, err := multiplayer.New(p, lm.id, lm.mpm, lm.ppm)
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

		lm.l.Info("heello?")
		mp.SetOnline(true, true)
	}
}

func (lm *ListenerManager) onDisconnect(e *proxy.DisconnectEvent) {
	id := e.Player().ID().String()
	mp := lm.mpm.GetMultiPlayer(id)
	if mp != nil {
		mp.SetOnline(false, true)
	}
}
