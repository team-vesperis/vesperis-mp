package listeners

import "go.minekube.com/gate/pkg/edition/java/proxy"

// also works on startup
func (lm *ListenerManager) onRegister(e *proxy.ServerRegisteredEvent) {
	si := e.Server().ServerInfo()
	lm.l.Info("registering multibackend ", "name", si.Name())

	_, err := lm.mm.NewMultiBackend(si.Name(), si.Addr().String())
	if err != nil {
		lm.l.Error("register server error", "error", err)
		return
	}
}

func (lm *ListenerManager) onUnRegister(e *proxy.ServerUnregisteredEvent) {
	si := e.ServerInfo()
	lm.l.Info("unregister multibackend", "name", si.Name())

	addr := e.ServerInfo().Addr().String()
	mb, err := lm.mm.GetMultiBackendUsingAddress(addr)
	if err != nil {
		lm.l.Error("unregister server get multibackend error", "error", err)
		return
	}

	err = lm.mm.DeleteMultiBackend(mb.GetId())
	if err != nil {
		lm.l.Error("unregister server delete multibackend error", "error", err)
		return
	}
}
