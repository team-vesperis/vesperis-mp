package multi

import (
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
	"go.minekube.com/gate/pkg/util/uuid"
)

type Proxy struct {
	id          uuid.UUID
	maintenance bool
	address     string

	backends []uuid.UUID
	players  []uuid.UUID

	mu        sync.RWMutex
	managerId uuid.UUID

	l  *logger.Logger
	db *database.Database
	cf *config.Config

	lastHeartBeat *time.Time
}

func NewProxy(id, managerId uuid.UUID, l *logger.Logger, db *database.Database, cf *config.Config, data *data.ProxyData) *Proxy {
	mp := &Proxy{
		id:        id,
		managerId: managerId,
		l:         l,
		db:        db,
		cf:        cf,
	}

	mp.address = data.Address
	mp.maintenance = data.Maintenance
	mp.backends = data.Backends
	mp.players = data.Players

	return mp
}

const UpdateMultiProxyChannel = "update_multiproxy"

func (mp *Proxy) save(k key.ProxyKey, val any) error {
	err := mp.db.SetProxyDataField(mp.id, k, val)
	if err != nil {
		return err
	}

	m := mp.managerId.String() + "_" + mp.id.String() + "_" + k.String()
	return mp.db.Publish(UpdateMultiProxyChannel, m)
}

func (mp *Proxy) Update(k key.ProxyKey) {
	var err error

	switch k {
	case key.ProxyKey_Maintenance:
		var maintenance bool
		err = mp.db.GetProxyDataField(mp.id, key.ProxyKey_Maintenance, &maintenance)
		mp.setInMaintenance(maintenance, false)
	case key.ProxyKey_BackendList:
		var backends []uuid.UUID
		err = mp.db.GetProxyDataField(mp.id, key.ProxyKey_BackendList, &backends)
		mp.setBackendsIds(backends, false)
	case key.ProxyKey_PlayerList:
		var players []uuid.UUID
		err = mp.db.GetProxyDataField(mp.id, key.ProxyKey_PlayerList, &players)
		mp.setPlayerIds(players, false)
	case key.ProxyKey_LastHeartBeat:
		var time time.Time
		err = mp.db.GetProxyDataField(mp.id, key.ProxyKey_LastHeartBeat, &time)
		mp.setLastHeartBeat(&time, false)
	}

	if err != nil {
		mp.l.Error("multiproxy update proxykey get field from database error", "error", err)
	}
}

var ErrProxyIdNilWhileOnline = errors.New("proxyId is nil but player is online")

func (mp *Proxy) GetId() uuid.UUID {
	return mp.id
}

func (mp *Proxy) GetAddress() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.address
}

func (mp *Proxy) GetLastHeartBeat() *time.Time {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.lastHeartBeat
}

func (mp *Proxy) SetLastHeartBeat(t *time.Time) error {
	return mp.setLastHeartBeat(t, true)
}

func (mp *Proxy) setLastHeartBeat(t *time.Time, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.lastHeartBeat = t

	if notify {
		return mp.save(key.ProxyKey_LastHeartBeat, t)
	}

	return nil
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
	defer mp.mu.Unlock()

	mp.maintenance = maintenance

	if notify {
		return mp.save(key.ProxyKey_Maintenance, maintenance)
	}

	return nil
}

func (mp *Proxy) GetPlayerIds() []uuid.UUID {
	mp.mu.RLock()
	c := append([]uuid.UUID{}, mp.players...)
	mp.mu.RUnlock()

	return c
}

func (mp *Proxy) SetPlayerIds(ids []uuid.UUID) error {
	return mp.setPlayerIds(ids, true)
}

func (mp *Proxy) setPlayerIds(ids []uuid.UUID, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.players = ids

	if notify {
		return mp.save(key.ProxyKey_PlayerList, ids)
	}

	return nil
}

func (mp *Proxy) AddPlayerId(id uuid.UUID) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if !slices.Contains(mp.players, id) {
		mp.players = append(mp.players, id)
	}

	return mp.save(key.ProxyKey_PlayerList, mp.players)
}

func (mp *Proxy) RemovePlayerId(id uuid.UUID) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	i := slices.Index(mp.players, id)
	if i == -1 {
		return ErrPlayerNotFound
	}
	mp.players = slices.Delete(mp.players, i, i+1)

	return mp.save(key.ProxyKey_PlayerList, mp.players)
}

func (mp *Proxy) IsPlayerIdOnProxy(id uuid.UUID) bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return slices.Contains(mp.players, id)
}

func (mp *Proxy) GetBackendsIds() []uuid.UUID {
	mp.mu.RLock()
	c := append([]uuid.UUID(nil), mp.backends...)
	mp.mu.RUnlock()

	return c
}

func (mp *Proxy) SetBackendsIds(ids []uuid.UUID) error {
	return mp.setBackendsIds(ids, true)
}

func (mp *Proxy) setBackendsIds(ids []uuid.UUID, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.backends = ids

	if notify {
		return mp.save(key.ProxyKey_BackendList, ids)
	}

	return nil
}

func (mp *Proxy) AddBackend(id uuid.UUID) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if !slices.Contains(mp.backends, id) {
		mp.backends = append(mp.backends, id)
	}

	return mp.save(key.ProxyKey_BackendList, mp.backends)
}

func (mp *Proxy) RemoveBackendId(id uuid.UUID) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	i := slices.Index(mp.backends, id)
	if i == -1 {
		return ErrBackendNotFound
	}
	mp.backends = slices.Delete(mp.backends, i, i+1)

	return mp.save(key.ProxyKey_BackendList, mp.backends)
}
