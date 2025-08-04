package multiplayer

import (
	"sync"
	"time"

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

	vanished bool

	// List of friends.
	friends []*MultiPlayer

	mu sync.RWMutex

	mpm *MultiPlayerManager

	ppm *PlayerPermissionManager
}

// New returns a new MultiPlayer
func New(p proxy.Player, proxyId string, mpm *MultiPlayerManager, ppm *PlayerPermissionManager) (*MultiPlayer, error) {
	now := time.Now()

	mp := &MultiPlayer{
		p:    proxyId,
		b:    p.CurrentServer().Server().ServerInfo().Name(),
		id:   p.ID().String(),
		name: p.Username(),
		mpm:  mpm,
		ppm:  ppm,
	}

	mpm.multiPlayerMap.Store(mp.id, mp)

	err := mp.SetRole(RoleDefault)
	if err != nil {
		return nil, err
	}

	err = mp.SetRank(RankDefault)
	if err != nil {
		return nil, err
	}

	err = mp.SaveAll()
	if err != nil {
		return nil, err
	}

	mpm.l.Info("created new multiplayer", "mp", mp, "playerId", p.ID().String(), "duration", time.Since(now))
	return mp, nil
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
	err := mp.mpm.db.SetPlayerDataField(mp.id, key, value)
	if err != nil {
		return err
	}

	m := mp.id + "_" + key
	return mp.mpm.db.Publish(multiPlayerUpdateChannel, m)
}

func (mp *MultiPlayer) GetProxyId() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.p
}

func (mp *MultiPlayer) SetProxyId(id string, notify bool) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.p = id
	if notify {
		mp.Save("p", id)
	}
}

func (mp *MultiPlayer) GetBackendId() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.b
}

func (mp *MultiPlayer) SetBackendId(id string, notify bool) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.b = id
	if notify {
		mp.Save("b", id)
	}
}

func (mp *MultiPlayer) GetId() string {
	return mp.id
}

func (mp *MultiPlayer) GetName() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.name
}

func (mp *MultiPlayer) SetName(name string, notify bool) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.name = name
	if notify {
		mp.Save("name", name)
	}
}

func (mp *MultiPlayer) GetRole() string {
	role, _ := mp.ppm.GetRoleFromId(mp.id)
	return role
}

func (mp *MultiPlayer) SetRole(role string) error {
	return mp.ppm.SetRoleWithId(mp.id, role)
}

func (mp *MultiPlayer) GetRank() string {
	rank, _ := mp.ppm.GetRankFromId(mp.id)
	return rank
}

func (mp *MultiPlayer) SetRank(rank string) error {
	return mp.ppm.SetRankWithId(mp.id, rank)
}

func (mp *MultiPlayer) IsOnline() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.online
}

func (mp *MultiPlayer) SetOnline(online bool, notify bool) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.online = online
	if notify {
		mp.Save("online", online)
	}
}

func (mp *MultiPlayer) IsVanished() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.vanished
}

func (mp *MultiPlayer) SetVanished(vanished bool, notify bool) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.vanished = vanished
	if notify {
		mp.Save("vanished", vanished)
	}
}

func (mp *MultiPlayer) GetFriends() []*MultiPlayer {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.friends
}

func (mp *MultiPlayer) SetFriends(friends []*MultiPlayer, notify bool) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.friends = friends
	if notify {
		mp.Save("friends", friends)
	}
}

func (mp *MultiPlayer) AddFriend(friend *MultiPlayer, notify bool) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.friends = append(mp.friends, friend)
	var ids []string

	for _, friend := range mp.friends {
		ids = append(ids, friend.id)
	}

	if notify {
		mp.Save("friends", ids)
	}
}
