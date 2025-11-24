package listeners

import "go.minekube.com/gate/pkg/edition/java/proxy"

// also works on startup
func (lm *ListenerManager) onRegister(e *proxy.ServerRegisteredEvent) {
	lm.l.Info("registering server ", "name", e.Server().ServerInfo().Name())

	id, err := lm.mm.CreateNewBackendId()
	if err != nil {
		lm.l.Error("error", "error", err)
		return
	}

	lm.l.Debug("found a backend id to use", "id", id)

	si := e.Server().ServerInfo()
	_, err = lm.mm.NewMultiBackend(si.Name(), si.Addr().String(), id)
	if err != nil {
		lm.l.Error("error", "error", err)
		return
	}
}

func (lm *ListenerManager) onUnRegister(e *proxy.ServerUnregisteredEvent) {
	lm.l.Info("unregister ", "name", e.ServerInfo().Name())

	addr := e.ServerInfo().Addr().String()
	mb, err := lm.mm.GetMultiBackendUsingAddress(addr)
	if err != nil {
		lm.l.Error("error", "error", err)
		return
	}

	err = lm.mm.DeleteMultiBackend(mb.GetId())
	if err != nil {
		lm.l.Error("error", "error", err)
		return
	}
}
