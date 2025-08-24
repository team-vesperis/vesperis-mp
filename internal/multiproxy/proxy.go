package multiproxy

import (
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiProxy struct {
	// The id of the mp.
	// Used to differentiate the proxy from others.
	id uuid.UUID

	mpm *MultiProxyManager

	maintenance bool

	address string

	connectedPlayers int
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

func (mp *MultiProxy) SetAddress(addr string) {
	mp.address = addr
}

func (mp *MultiProxy) GetConnectedPlayers() int {
	return mp.connectedPlayers
}

func (mp *MultiProxy) SetConnectedPlayers(count int) {
	mp.connectedPlayers = count
}

// creates id and checks if available
func (mpm *MultiProxyManager) createNewProxyId() uuid.UUID {
	id := uuid.New()
	mp, _ := mpm.GetMultiProxy(id)
	if mp == nil {
		return id
	}

	return mpm.createNewProxyId()
}
