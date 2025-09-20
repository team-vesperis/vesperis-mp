package multi

import (
	"sync"

	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiProxy struct {
	mu sync.RWMutex

	id uuid.UUID

	maintenance bool

	address string

	connectedBackends []*MultiBackend
	connectedPlayers  []*MultiPlayer
}

func NewMultiProxy(address string, id uuid.UUID) *MultiProxy {
	mp := &MultiProxy{
		id:      id,
		address: address,
	}

	return mp
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

func (mp *MultiProxy) GetConnectedPlayers() []*MultiPlayer {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.connectedPlayers
}

func (mp *MultiProxy) AddMultiPlayerToMultiProxy(mplayer *MultiPlayer) {
	mp.connectedPlayers = append(mp.connectedPlayers, mplayer)

}
