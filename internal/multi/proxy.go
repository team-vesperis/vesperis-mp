package multi

import (
	"sync"

	"go.minekube.com/gate/pkg/util/uuid"
)

type Proxy struct {
	mu sync.RWMutex

	id uuid.UUID

	maintenance bool

	address string

	backends map[*Backend]bool
	players  map[*Player]bool
}

func NewMultiProxy(address string, id uuid.UUID) *Proxy {
	mp := &Proxy{
		id:      id,
		address: address,
	}

	return mp
}

func (mp *Proxy) GetId() uuid.UUID {
	return mp.id
}

func (mp *Proxy) GetAddress() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.address
}

func (mp *Proxy) IsInMaintenance() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.maintenance
}

func (mp *Proxy) SetInMaintenance(maintenance bool) {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.maintenance = maintenance
}

func (mp *Proxy) GetAllPlayers() []*Player {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	l := make([]*Player, 0, len(mp.players))
	i := 0
	for p := range mp.players {
		l[i] = p
		i++
	}

	return l
}

func (mp *Proxy) AddPlayer(p *Player) {
	mp.mu.Lock()
	mp.players[p] = true
	mp.mu.Unlock()
}

func (mp *Proxy) RemovePlayer(p *Player) {
	mp.mu.Lock()
	delete(mp.players, p)
	mp.mu.Unlock()
}

func (mp *Proxy) IsPlayerOnProxy(p *Player) bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	_, exists := mp.players[p]
	return exists
}
