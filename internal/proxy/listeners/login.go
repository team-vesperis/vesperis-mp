package listeners

import (
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"go.minekube.com/common/minecraft/color"
	. "go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

var loginDenyComponent = &Text{
	Content: "There was an error with the login. Please try again later.",
	S: Style{
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
		mp, err = lm.mpm.NewMultiPlayer(p)
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
		return
	}

	err = mp.SetProxy(lm.ownerMultiProxy)
	if err != nil {
		lm.l.Error("player post login set proxy error", "playerId", id, "error", err)
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
		return
	}

	err = mp.SetLastSeen(time.Now())
	if err != nil {
		lm.l.Error("player disconnect set last seen error", "playerId", id, "error", err)
		return
	}

	err = mp.SetProxy(nil)
	if err != nil {
		lm.l.Error("player disconnect set proxy error", "playerId", id, "error", err)
		return
	}
}
