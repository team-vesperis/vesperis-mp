package listeners

import (
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

var loginDenyComponent = util.TextError("There was an error logging in. Please try again later.")

func (lm *ListenerManager) onLogin(e *proxy.LoginEvent) {
	p := e.Player()
	id := p.ID()

	mp, err := lm.mm.GetMultiPlayer(id)
	if err != nil {
		if err != database.ErrDataNotFound {
			lm.l.Error("player login get multiplayer error", "playerId", id, "error", err)
			e.Deny(loginDenyComponent)
			return
		}

		// player hasn't joined before -> creating default mp
		mp, err = lm.mm.NewMultiPlayer(p)
		if err != nil {
			lm.l.Error("player login create new multiplayer error", "playerId", id, "error", err)
			e.Deny(loginDenyComponent)
			return
		}
	}

	if mp.GetBanInfo().IsBanned() {
		if mp.GetBanInfo().IsPermanently() {
			e.Deny(&component.Text{
				Content: "You are permanently banned.",
				S:       util.StyleColorRed,
				Extra: []component.Component{
					&component.Text{
						Content: "\n\nReason: ",
						S:       util.StyleColorRed,
					},
					&component.Text{
						Content: mp.GetBanInfo().GetReason(),
						S:       util.StyleColorLightBlue,
					},
				},
			})
		} else {
			// unban
			if time.Now().After(mp.GetBanInfo().GetExpiration()) {
				err := mp.GetBanInfo().UnBan()
				if err != nil {
					lm.l.Error("player login unban error", "playerId", id, "error", err)
					e.Deny(loginDenyComponent)
					return
				}

				e.Allow()
			} else {
				e.Deny(&component.Text{
					Content: "You are temporarily banned.",
					S:       util.StyleColorRed,
					Extra: []component.Component{
						&component.Text{
							Content: "\n\nReason: ",
							S:       util.StyleColorRed,
						},
						&component.Text{
							Content: mp.GetBanInfo().GetReason(),
							S:       util.StyleColorLightBlue,
						},
						&component.Text{
							Content: "\nExpiration: ",
							S:       util.StyleColorRed,
						},
						&component.Text{
							Content: util.FormatTimeUntil(mp.GetBanInfo().GetExpiration()),
							S:       util.StyleColorLightBlue,
						},
					},
				})
			}
		}
	}
}

func (lm *ListenerManager) onDisconnect(e *proxy.DisconnectEvent) {
	id := e.Player().ID()
	mp, err := lm.mm.GetMultiPlayer(id)
	if err != nil {
		lm.l.Error("player disconnect get multiplayer error", "playerId", id, "error", err)
		return
	}

	if !mp.IsOnline() {
		return
	}

	now := time.Now()
	err = mp.SetLastSeen(&now)
	if err != nil {
		lm.l.Error("player disconnect set last seen error", "playerId", id, "error", err)
		return
	}

	err = mp.SetOnline(false)
	if err != nil {
		lm.l.Error("player disconnect set online error", "playerId", id, "error", err)
		return
	}

	if mp.GetProxy() == nil {
		return
	}

	err = lm.mm.GetOwnerMultiProxy().RemovePlayerId(id)
	if err != nil {
		lm.l.Error("player disconnect remove playerId from proxy error", "playerId", id, "error", err)
		return
	}

	err = mp.SetProxy(nil)
	if err != nil {
		lm.l.Error("player disconnect set proxy error", "playerId", id, "error", err)
		return
	}

	if mp.GetBackend() == nil {
		return
	}

	err = mp.GetBackend().RemovePlayerId(id)
	if err != nil {
		lm.l.Error("player disconnect remove playerId from backend error", "playerId", id, "error", err)
		return
	}

	err = mp.SetBackend(nil)
	if err != nil {
		lm.l.Error("player disconnect set backend error", "playerId", id, "error", err)
		return
	}
}
