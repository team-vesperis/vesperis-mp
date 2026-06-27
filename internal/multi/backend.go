package multi

import (
	"errors"
	"slices"
	"sync"

	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
	"go.minekube.com/gate/pkg/util/uuid"
)

type Backend struct {
	id uuid.UUID
	mp *Proxy

	name    string
	address string
	players []uuid.UUID

	mi *maintenanceInfo

	mu        sync.RWMutex
	managerId uuid.UUID

	l  *logger.Logger
	db *database.Database
	cf *config.Config
}

func NewBackend(id, managerId uuid.UUID, ownerMP *Proxy, l *logger.Logger, db *database.Database, cf *config.Config, data *data.BackendData) *Backend {
	mb := &Backend{
		id:        id,
		managerId: managerId,
		mp:        ownerMP,
		l:         l,
		db:        db,
		cf:        cf,
		mu:        sync.RWMutex{},
	}

	mb.name = data.Name
	mb.address = data.Address
	mb.players = data.Players
	mb.mi = newBackendMaintenanceInfo(mb, data)

	return mb
}

var ErrBackendNotFound = errors.New("backend not found")

const UpdateMultiBackendChannel = "update_multibackend"

func (mb *Backend) save(k key.Key, val any) error {
	bk, ok := k.(key.BackendKey)
	if !ok {
		return key.ErrIncorrectBackendKey
	}

	err := mb.db.SetBackendDataField(mb.id, bk, val)
	if err != nil {
		return err
	}

	m := mb.managerId.String() + "_" + mb.id.String() + "_" + k.String()
	return mb.db.Publish(UpdateMultiBackendChannel, m)
}

func (mb *Backend) Update(k key.BackendKey) {
	var err error

	switch k {
	case key.BackendKey_PlayerList:
		var playerList []uuid.UUID
		err = mb.db.GetBackendDataField(mb.id, key.BackendKey_PlayerList, &playerList)
		mb.setPlayerIds(playerList, false)
	}

	if err != nil {
		mb.l.Error("multibackend update backendkey get field from database error", "error", err)
	}
}

func (mb *Backend) GetName() string {
	return mb.name
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

func (mb *Backend) GetMaintenanceInfo() *maintenanceInfo {
	return mb.mi
}

func (mb *Backend) GetPlayerIds() []uuid.UUID {
	mb.mu.RLock()
	c := make([]uuid.UUID, len(mb.players))
	copy(c, mb.players)
	mb.mu.RUnlock()

	return c
}

func (mb *Backend) SetPlayerIds(ids []uuid.UUID) error {
	return mb.setPlayerIds(ids, true)
}

func (mb *Backend) setPlayerIds(ids []uuid.UUID, notify bool) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.players = ids

	if notify {
		return mb.save(key.BackendKey_PlayerList, ids)
	}

	return nil
}

func (mb *Backend) AddPlayerId(id uuid.UUID) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if !slices.Contains(mb.players, id) {
		mb.players = append(mb.players, id)
	}

	return mb.save(key.BackendKey_PlayerList, mb.players)
}

func (mb *Backend) RemovePlayerId(id uuid.UUID) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	i := slices.Index(mb.players, id)
	if i == -1 {
		return ErrPlayerNotFound
	}
	mb.players = slices.Delete(mb.players, i, i+1)

	return mb.save(key.BackendKey_PlayerList, mb.players)
}

func (mb *Backend) IsPlayerIdOnProxy(id uuid.UUID) bool {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	return slices.Contains(mb.players, id)
}
