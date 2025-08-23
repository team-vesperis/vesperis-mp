package listeners

import (
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/multiplayer"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

var loginDenyComponent = &component.Text{
	Content: "There was an error with the login. Please try again later.",
	S: component.Style{
		Color: color.Red,
	},
}

func (lm *ListenerManager) onLogin(e *proxy.LoginEvent) {
	p := e.Player()
	id := p.ID()

	mp, err := lm.mpm.GetMultiPlayer(id)
	// player hasn't joined before -> creating default mp
	if err == database.ErrDataNotFound {
		var err error
		mp, err = multiplayer.New(p, lm.db, lm.mpm)
		if err != nil {
			lm.l.Error("player login create new multiplayer error", "playerId", id, "error", err)
			e.Deny(loginDenyComponent)
			return
		}
	} else if err != nil {
		lm.l.Error("player login get multiplayer error", "playerId", id, "error", err)
		e.Deny(loginDenyComponent)
		return
	}

	err = mp.SetOnline(true)
	if err != nil {
		lm.l.Error("player login set online error", "playerId", id, "error", err)
		e.Deny(loginDenyComponent)
	}

	err = mp.SetProxyId(lm.id)
	if err != nil {
		lm.l.Error("player post login set proxy id error", "playerId", id, "error", err)
		e.Deny(loginDenyComponent)
		return
	}
}

func (lm *ListenerManager) onDisconnect(e *proxy.DisconnectEvent) {
	id := e.Player().ID()
	mp, err := lm.mpm.GetMultiPlayer(id)
	if err != nil {
		lm.l.Error("player disconnect get multiplayer error", "playerId", id, "error", err)
		return
	}

	err = mp.SetOnline(false)
	if err != nil {
		lm.l.Error("player disconnect set online error", "playerId", id, "error", err)
	}
}
