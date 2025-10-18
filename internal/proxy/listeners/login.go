package listeners

import (
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

var loginDenyComponent = &c.Text{
	Content: "There was an error while login in. Please try again later.",
	S: c.Style{
		Color: color.Red,
	},
}

func (lm *ListenerManager) onLogin(e *proxy.LoginEvent) {
	p := e.Player()
	id := p.ID()

	_, err := lm.mm.GetMultiPlayer(id)
	// player hasn't joined before -> creating default mp
	if err == database.ErrDataNotFound {
		_, err := lm.mm.NewMultiPlayer(p)
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
}

func (lm *ListenerManager) onDisconnect(e *proxy.DisconnectEvent) {
	id := e.Player().ID()
	mp, err := lm.mm.GetMultiPlayer(id)
	if err != nil {
		lm.l.Error("player disconnect get multiplayer error", "playerId", id, "error", err)
		return
	}

	err = mp.SetOnline(false)
	if err != nil {
		lm.l.Error("player disconnect set online error", "playerId", id, "error", err)
		return
	}

	now := time.Now()
	err = mp.SetLastSeen(&now)
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
