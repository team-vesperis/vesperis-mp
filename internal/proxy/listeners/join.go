package listeners

import "go.minekube.com/gate/pkg/edition/java/proxy"

func (lm *ListenerManager) onProxyJoin(e *proxy.PostLoginEvent) {
	p := e.Player()
	id := p.ID()

	mp, err := lm.mpm.GetMultiPlayer(id)
	if err != nil {
		lm.l.Error("player post login get multiplayer error", "playerId", id, "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}

	err = mp.SetProxyId(lm.id, true)
	if err != nil {
		lm.l.Error("player post login set proxy id error", "playerId", id, "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}
}

func (lm *ListenerManager) onServerJoin(e *proxy.ServerPostConnectEvent) {
	p := e.Player()
	id := p.ID()

	mp, err := lm.mpm.GetMultiPlayer(id)
	if err != nil {
		lm.l.Error("player server post connect get multiplayer error", "playerId", id, "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}

	backendId := p.CurrentServer().Server().ServerInfo().Name()
	err = mp.SetBackendId(backendId, true)
	if err != nil {
		lm.l.Error("player server post connect set backend id error", "playerId", id, "backendId", backendId, "error", err)
		p.Disconnect(loginDenyComponent)
		return
	}
}
