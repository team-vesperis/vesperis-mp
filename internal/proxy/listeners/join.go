package listeners

import "go.minekube.com/gate/pkg/edition/java/proxy"

func (lm *ListenerManager) onProxyJoin(e *proxy.PostLoginEvent) {
	p := e.Player()
	mp, _ := lm.mpm.GetMultiPlayer(p.ID())
	if mp != nil {
		mp.SetProxyId(lm.id, true)
	}
}

func (lm *ListenerManager) onServerJoin(e *proxy.ServerPostConnectEvent) {
	p := e.Player()
	mp, _ := lm.mpm.GetMultiPlayer(p.ID())
	if mp != nil {
		mp.SetBackendId(p.CurrentServer().Server().ServerInfo().Name(), true)
	}
}
