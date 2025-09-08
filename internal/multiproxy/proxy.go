package multiproxy

import (
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multiplayer"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiProxy struct {
	// The id of the mp.
	// Used to differentiate the proxy from others.
	id uuid.UUID

	mpm *MultiProxyManager

	maintenance bool

	address string

	connectedPlayers []*multiplayer.MultiPlayer
}

func New(id uuid.UUID, mpm *MultiProxyManager) (MultiProxy, error) {
	mp := MultiProxy{
		id:  id,
		mpm: mpm,
	}

	mpm.multiProxyMap.Store(id, mp)

	return mp, nil
}

func (mp *MultiProxy) GetLogger() *logger.Logger {
	return mp.mpm.l
}

func (mp *MultiProxy) GetAddress() string {
	return mp.address
}

func (mp *MultiProxy) IsInMaintenance() bool {
	return mp.maintenance
}

func (mp *MultiProxy) SetInMaintenance(maintenance bool) {
	mp.maintenance = maintenance
}

func (mp *MultiProxy) GetConnectedPlayers() []*multiplayer.MultiPlayer {
	return mp.connectedPlayers
}

// creates id
func (mpm *MultiProxyManager) createNewProxyId() uuid.UUID {
	id := uuid.New()
	mp, _ := mpm.GetMultiProxy(id)
	if mp == nil {
		return id
	}

	// loops until an id that is not used
	return mpm.createNewProxyId()
}
