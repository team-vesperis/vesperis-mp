package multi

import (
	"sync"

	"go.minekube.com/gate/pkg/util/uuid"
)

type Proxy struct {
	variable_mu sync.RWMutex
	id          uuid.UUID
	maintenance bool
	address     string

	b_mu sync.RWMutex
	b    map[*Backend]bool

	p_mu sync.RWMutex
	p    map[*Player]bool
}

func NewProxy(address string, id uuid.UUID) *Proxy {
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
	mp.variable_mu.RLock()
	defer mp.variable_mu.RUnlock()
	return mp.address
}

func (mp *Proxy) IsInMaintenance() bool {
	mp.variable_mu.RLock()
	defer mp.variable_mu.RUnlock()
	return mp.maintenance
}

func (mp *Proxy) SetInMaintenance(maintenance bool) {
	mp.variable_mu.Lock()
	defer mp.variable_mu.Unlock()
	mp.maintenance = maintenance
}

func (mp *Proxy) GetAllPlayers() []*Player {
	mp.p_mu.RLock()
	defer mp.p_mu.RUnlock()

	l := make([]*Player, 0, len(mp.p))
	i := 0
	for p := range mp.p {
		l[i] = p
		i++
	}

	return l
}

func (mp *Proxy) AddPlayer(p *Player) {
	mp.p_mu.Lock()
	mp.p[p] = true
	mp.p_mu.Unlock()
}

func (mp *Proxy) RemovePlayer(p *Player) {
	mp.p_mu.Lock()
	delete(mp.p, p)
	mp.p_mu.Unlock()
}

func (mp *Proxy) IsPlayerOnProxy(p *Player) bool {
	mp.p_mu.RLock()
	defer mp.p_mu.RUnlock()

	_, exists := mp.p[p]
	return exists
}

func (mp *Proxy) GetAllBackends() []*Backend {
	mp.b_mu.RLock()
	defer mp.b_mu.RUnlock()

	l := make([]*Backend, 0, len(mp.b))
	i := 0
	for b := range mp.b {
		l[i] = b
		i++
	}

	return l
}

func (mp *Proxy) AddBackend(b *Backend) {
	mp.b_mu.Lock()
	mp.b[b] = true
	mp.b_mu.Unlock()
}

func (mp *Proxy) RemoveBackend(b *Backend) {
	mp.b_mu.Lock()
	delete(mp.b, b)
	mp.b_mu.Unlock()
}
