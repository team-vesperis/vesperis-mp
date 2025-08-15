package multiproxy

import (
	"github.com/robinbraemer/event"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/proxy/commands"
	"github.com/team-vesperis/vesperis-mp/internal/proxy/listeners"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/gate"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiProxy struct {
	// The id of the mp.
	// Used to differentiate the proxy from others.
	id string

	// The gate proxy used in the mp.
	p *proxy.Proxy

	// The command manager used in the mp.
	cm *commands.CommandManager

	// The listener manager in the mp.
	lm *listeners.ListenerManager

	mpm *MultiProxyManager
}

func New(mpm *MultiProxyManager) (MultiProxy, error) {
	// if id == "" || db.CheckIfProxyIdIsAvailable(id) == false {
	// 	// set to a unique id
	// 	// TODO: create standalone function
	// 	// for creating unique id to check if the new id is not used
	// 	id = "proxy_" + uuid.New().Undashed()
	// }

	id := "proxy_" + uuid.New().Undashed()

	mp := MultiProxy{
		id:  id,
		mpm: mpm,
	}

	cfg, err := gate.LoadConfig(mp.mpm.c.GetViper())
	gate, err := gate.New(gate.Options{
		Config:   cfg,
		EventMgr: event.New(),
	})
	if err != nil {
		mp.mpm.l.Error("creating gate instance error", "error", err)
		return mp, err
	}

	mp.p = gate.Java()
	event.Subscribe(mp.p.Event(), 0, mp.onShutdown)

	mp.cm = commands.Init(mp.p, mp.mpm.l, mp.mpm.db, mp.mpm.mpm)
	mp.lm = listeners.Init(mp.p.Event(), mp.mpm.l, mp.mpm.db, mp.mpm.mpm, mp.id)

	mpm.multiProxyMap.Store(id, mp)

	return mp, nil
}

func (mp *MultiProxy) Start() {
	err := mp.p.Start(mp.mpm.ctx)
	if err != nil {
		mp.mpm.l.Error("error starting proxy", "error", err)
	}
}

func (mp *MultiProxy) Shutdown(reason component.Text) {
	mp.p.Shutdown(&reason)
}

func (mp *MultiProxy) GetLogger() *logger.Logger {
	return mp.mpm.l
}

func (mp *MultiProxy) onShutdown(event *proxy.ShutdownEvent) {
	mp.mpm.Close()
}
