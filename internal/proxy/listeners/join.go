package listeners

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi"
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

	err = mp.SetProxy(lm.ownerMultiProxy)
	if err != nil {
		lm.l.Error("player post login set proxy error", "playerId", p.ID(), "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}

	err = lm.ownerMultiProxy.AddPlayerId(p.ID())
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

	var mb *multi.Backend
	for _, b := range lm.mm.GetAllMultiBackends() {
		if si.Addr().String() == b.GetAddress() {
			mb = b
		}
	}

	if mb == nil {
		l, err := lm.mm.GetAllMultiBackendsFromDatabase()
		if err != nil {
			lm.l.Error("player server post connect get all multibackends from database error", "playerId", p.ID(), "error", err)
			p.Disconnect(loginDenyComponent)
			return
		}

		for _, b := range l {
			if si.Addr().String() == b.GetAddress() {
				mb = b
			}
		}

		if mb == nil {
			lm.l.Error("player server post connect backend is not registered in the database", "backendName", si.Name(), "backendAddress", si.Addr().String())
			p.Disconnect(loginDenyComponent)
			return
		}
	}

	mp, err := lm.mm.GetMultiPlayer(p.ID())
	if err != nil {
		lm.l.Error("player server post connect get multiplayer error", "playerId", p.ID(), "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}

	err = mb.AddPlayerId(p.ID())
	if err != nil {
		lm.l.Error("player server post connect add playerId to multibackend error", "playerId", p.ID(), "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}

	err = mp.SetBackend(mb)
	if err != nil {
		lm.l.Error("player server post connect set multibackend error", "playerId", p.ID(), "backendId", mb.GetId(), "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}
}
