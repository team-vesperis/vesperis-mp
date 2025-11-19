package listeners

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

func (lm *ListenerManager) onProxyJoin(e *proxy.PostLoginEvent) {
	p := e.Player()

	mp, err := lm.mm.GetMultiPlayer(p.ID())
	if err != nil {
		lm.l.Error("player post login get multiplayer error", "playerId", p.ID(), "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}

	err = mp.SetOnline(true)
	if err != nil {
		lm.l.Error("player login set online error", "playerId", p.ID(), "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}

	err = mp.SetProxy(lm.mm.GetOwnerMultiProxy())
	if err != nil {
		lm.l.Error("player post login set proxy error", "playerId", p.ID(), "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}

	err = lm.mm.GetOwnerMultiProxy().AddPlayerId(p.ID())
	if err != nil {
		lm.l.Error("player post login add playerId to multiproxy error", "playerId", p.ID(), "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}

	if p.Username() != mp.GetUsername() {
		err := mp.SetUsername(p.Username())
		if err != nil {
			lm.l.Error("player post login set name error", "playerId", p.ID(), "error", err)
		}
	}
}

func (lm *ListenerManager) onServerJoin(e *proxy.ServerPostConnectEvent) {
	p := e.Player()
	si := p.CurrentServer().Server().ServerInfo()

	util.PlayLevelUpSound(p)

	mb, err := lm.mm.GetMultiBackendUsingAddress(si.Addr().String())
	if err != nil {
		lm.l.Error("player server post connect get multibackend error", "playerId", p.ID(), "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}

	mp, err := lm.mm.GetMultiPlayer(p.ID())
	if err != nil {
		lm.l.Error("player server post connect get multiplayer error", "playerId", p.ID(), "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}

	err = mp.SetBackend(mb)
	if err != nil {
		lm.l.Error("player server post connect set multibackend error", "playerId", p.ID(), "backendId", mb.GetId(), "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}

	err = mb.AddPlayerId(p.ID())
	if err != nil {
		lm.l.Error("player server post connect add playerId to multibackend error", "playerId", p.ID(), "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}

}
