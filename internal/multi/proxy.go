package multi

import (
	"slices"
	"sync"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/gate/pkg/util/uuid"
)

type Proxy struct {
	id          uuid.UUID
	maintenance bool
	address     string

	b map[*Backend]bool
	p map[*Player]bool

	mu        sync.RWMutex
	managerId uuid.UUID
	db        *database.Database
}

func NewProxy(id, managerId uuid.UUID, db *database.Database, data *util.ProxyData) *Proxy {
	mp := &Proxy{
		id:        id,
		managerId: id,
		db:        db,
	}

	mp.address = data.Address
	mp.maintenance = data.Maintenance

	return mp
}

const UpdateMultiProxyChannel = "update_multiproxy"

func (mp *Proxy) save(key util.ProxyKey, val any) error {
	if !slices.Contains(util.AllowedProxyKeys, key) {
		return util.ErrIncorrectProxyKey
	}

	err := mp.db.SetProxyDataField(mp.id, key, val)
	if err != nil {
		return err
	}

	m := mp.managerId.String() + "_" + mp.id.String() + "_" + key.String()
	return mp.db.Publish(UpdateMultiProxyChannel, m)
}

func (mp *Proxy) Update(key util.ProxyKey) {

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

func (mp *Proxy) SetInMaintenance(maintenance bool) error {
	return mp.setInMaintenance(maintenance, true)
}

func (mp *Proxy) setInMaintenance(maintenance, notify bool) error {
	mp.mu.Lock()
	mp.maintenance = maintenance
	mp.mu.Unlock()

	if notify {
		return mp.save(util.ProxyKey_Maintenance, maintenance)
	}

	return nil
}

func (mp *Proxy) GetAllPlayers() []*Player {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	l := make([]*Player, 0, len(mp.p))
	i := 0
	for p := range mp.p {
		l[i] = p
		i++
	}

	return l
}

func (mp *Proxy) AddPlayer(p *Player) {
	mp.mu.Lock()
	mp.p[p] = true
	mp.mu.Unlock()
}

func (mp *Proxy) RemovePlayer(p *Player) {
	mp.mu.Lock()
	delete(mp.p, p)
	mp.mu.Unlock()
}

func (mp *Proxy) IsPlayerOnProxy(p *Player) bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	_, exists := mp.p[p]
	return exists
}

func (mp *Proxy) GetAllBackends() []*Backend {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	l := make([]*Backend, 0, len(mp.b))
	i := 0
	for b := range mp.b {
		l[i] = b
		i++
	}

	return l
}

func (mp *Proxy) AddBackend(b *Backend) {
	mp.mu.Lock()
	mp.b[b] = true
	mp.mu.Unlock()
}

func (mp *Proxy) RemoveBackend(b *Backend) {
	mp.mu.Lock()
	delete(mp.b, b)
	mp.mu.Unlock()
}
