package multiplayer

import (
	"sync"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

type MultiPlayerManager struct {
	MultiPlayerMap sync.Map

	l  *logger.Logger
	db *database.Database
}

func InitMultiPlayerManager() {

}

type MultiPlayer struct {
	// The proxy id on which the underlying player is located
	// Can be nil if not online!
	p string

	// The backend id on which the underlying player is located
	// Can be nil if not online!
	b string

	// The id of the underlying player
	id string

	// The username of the underlying player
	name string

	online bool

	// List of friends
	// Holds the id of the friend
	friends []string

	mu sync.RWMutex
}

// New returns a new MultiPlayer
func New(p proxy.Player, p_id, b_id string, m *MultiPlayerManager) *MultiPlayer {
	mp := &MultiPlayer{
		p:    p_id,
		b:    b_id,
		id:   p.ID().String(),
		name: p.Username(),
	}

	m.MultiPlayerMap.Store(mp.id, mp)
	m.l.Info("created new multiplayer", "mp", mp, "playerId", p.ID().String())
	return mp
}

func Get(id string, m *MultiPlayerManager) *MultiPlayer {
	val, ok := m.MultiPlayerMap.Load(id)
	if ok {
		mp, ok := val.(*MultiPlayer)
		if ok {
			return mp
		}
	}

	data, err := m.db.GetPlayerData(id)
	if err != nil || data == nil {
		return nil
	}
	mp := &MultiPlayer{
		id: id,
	}
	if v, ok := data["p"].(string); ok {
		mp.p = v
	}
	if v, ok := data["b"].(string); ok {
		mp.b = v
	}
	if v, ok := data["name"].(string); ok {
		mp.name = v
	}
	if v, ok := data["online"].(bool); ok {
		mp.online = v
	}
	m.MultiPlayerMap.Store(id, mp)
	return mp
}

const multiPlayerUpdateChannel = "update_mp"

// Updates the multi player into the database
// Notifies other proxies to update
func (mp *MultiPlayer) SaveAll(m *MultiPlayerManager) error {
	data := map[string]any{
		"p":       mp.p,
		"b":       mp.b,
		"id":      mp.id,
		"name":    mp.name,
		"online":  mp.online,
		"friends": mp.friends,
	}
	for k, v := range data {
		if err := m.db.SetPlayerDataField(mp.id, k, v); err != nil {
			return err
		}
	}

	return m.db.Publish(multiPlayerUpdateChannel, mp.id)
}

func (mp *MultiPlayer) Save() {

}

func (mp *MultiPlayer) ProxyId() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.p
}

func (mp *MultiPlayer) BackendId() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.b
}

func (mp *MultiPlayer) SetBackendId(b string) {
	mp.mu.Lock()
	mp.b = b
	mp.mu.Unlock()
	mp.Save()
}

func (mp *MultiPlayer) Id() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.id
}

func (mp *MultiPlayer) Name() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.name
}

func (mp *MultiPlayer) IsOnline() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.online
}

func (mp *MultiPlayer) SetOnline(online bool) {
	mp.mu.Lock()
	mp.online = online
	mp.mu.Unlock()
	mp.Save()
}
