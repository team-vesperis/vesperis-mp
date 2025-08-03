package multiplayer

import (
	"sync"

	"go.minekube.com/gate/pkg/edition/java/proxy"
)

type MultiPlayer struct {
	// The proxy id on which the underlying player is located.
	// Can be nil if not online!
	p string

	// The backend id on which the underlying player is located.
	// Can be nil if not online!
	b string

	// The id of the underlying player
	id string

	// The username of the underlying player
	name string

	online bool

	// List of friends.
	// Holds the id of the friend
	friends []string

	mu sync.RWMutex

	m *MultiPlayerManager
}

// New returns a new MultiPlayer
func New(p proxy.Player, proxyId string, m *MultiPlayerManager) (*MultiPlayer, error) {
	mp := &MultiPlayer{
		p:    proxyId,
		b:    p.CurrentServer().Server().ServerInfo().Name(),
		id:   p.ID().String(),
		name: p.Username(),
		m:    m,
	}

	m.multiPlayerMap.Store(mp.id, mp)

	err := mp.SaveAll()
	if err != nil {
		return nil, err
	}

	m.l.Info("created new multiplayer", "mp", mp, "playerId", p.ID().String())
	return mp, nil
}

func Get(id string, m *MultiPlayerManager) *MultiPlayer {
	val, ok := m.multiPlayerMap.Load(id)
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

	m.multiPlayerMap.Store(id, mp)
	return mp
}

const multiPlayerUpdateChannel = "update_mp"

// Updates the multi player into the database
// Notifies other proxies to update
func (mp *MultiPlayer) SaveAll() error {
	err := mp.Save("p", mp.GetProxyId())
	if err != nil {
		return err
	}

	err = mp.Save("b", mp.GetBackendId())
	if err != nil {
		return err
	}

	err = mp.Save("name", mp.GetName())
	if err != nil {
		return err
	}

	err = mp.Save("online", mp.IsOnline())

	return nil
}

// Update specific value of the multi player into the database
// Notifies other proxies to update that value
func (mp *MultiPlayer) Save(key string, value any) error {
	err := mp.m.db.SetPlayerDataField(mp.id, key, value)
	if err != nil {
		return err
	}

	m := mp.id + "_" + key
	return mp.m.db.Publish(multiPlayerUpdateChannel, m)
}

func (mp *MultiPlayer) GetProxyId() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.p
}

func (mp *MultiPlayer) GetBackendId() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.b
}

func (mp *MultiPlayer) SetBackendId(b string) {
	mp.mu.Lock()
	mp.b = b
	mp.Save("b", b)
	mp.mu.Unlock()
}

func (mp *MultiPlayer) GetId() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.id
}

func (mp *MultiPlayer) GetName() string {
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
	mp.Save("online", online)
	mp.mu.Unlock()
}
