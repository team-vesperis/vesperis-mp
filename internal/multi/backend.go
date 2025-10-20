package multi

import (
	"sync"

	"go.minekube.com/gate/pkg/util/uuid"
)

type Backend struct {
	id uuid.UUID

	address string
	mp      *Proxy

	maintenance bool

	players map[*Player]bool

	mu *sync.RWMutex
}

func (mb *Backend) GetAddress() string {
	return mb.address
}

func (mb *Backend) GetId() uuid.UUID {
	return mb.id
}

// return the multiproxy the multibackend is located under
func (mb *Backend) GetMultiProxy() *Proxy {
	return mb.mp
}

func (mb *Backend) IsInMaintenance() bool {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return mb.maintenance
}

func (mb *Backend) SetInMaintenance(maintenance bool) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.maintenance = maintenance
}

func (mb *Backend) GetAllPlayers() []*Player {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	l := make([]*Player, 0, len(mb.players))
	i := 0
	for p := range mb.players {
		l[i] = p
		i++
	}

	return l
}

func (mb *Backend) AddPlayer(p *Player) {
	mb.mu.Lock()
	mb.players[p] = true
	mb.mu.Unlock()
}

func (mb *Backend) RemovePlayer(p *Player) {
	mb.mu.Lock()
	delete(mb.players, p)
	mb.mu.Unlock()
}

func (mb *Backend) IsPlayerOnBackend(p *Player) bool {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	_, exists := mb.players[p]
	return exists
}
