package multiproxy

import (
	"sync"

	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multiplayer"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiProxy struct {
	mu sync.RWMutex

	id uuid.UUID

	mpm *MultiProxyManager

	maintenance bool

	address string

	connectedPlayers []*multiplayer.MultiPlayer
}

func NewMultiProxy(address string, id uuid.UUID, mpm *MultiProxyManager) (*MultiProxy, error) {
	mp := &MultiProxy{
		id:      id,
		mpm:     mpm,
		address: address,
	}

	mpm.multiProxyMap.Store(id, mp)

	return mp, nil
}

func (mp *MultiProxy) GetLogger() *logger.Logger {
	return mp.mpm.l
}

func (mp *MultiProxy) GetAddress() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.address
}

func (mp *MultiProxy) IsInMaintenance() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.maintenance
}

func (mp *MultiProxy) SetInMaintenance(maintenance bool) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.maintenance = maintenance
}

func (mp *MultiProxy) GetConnectedPlayers() []*multiplayer.MultiPlayer {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.connectedPlayers
}

// creates id
func (mpm *MultiProxyManager) createNewProxyId() uuid.UUID {
	for {
		id := uuid.New()
		mp, _ := mpm.GetMultiProxy(id)
		if mp == nil {
			return id
		}
	}
}
