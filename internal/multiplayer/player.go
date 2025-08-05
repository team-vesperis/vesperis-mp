package multiplayer

import (
	"sync"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/database"
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

	role string

	rank string

	online bool

	vanished bool

	// List of friends.
	friends []*MultiPlayer

	mu sync.RWMutex

	mpm *MultiPlayerManager
}

// New returns a new MultiPlayer
func New(p proxy.Player, db *database.Database, mpm *MultiPlayerManager) (*MultiPlayer, error) {
	now := time.Now()
	id := p.ID().String()

	mp := &MultiPlayer{
		id:   id,
		name: p.Username(),
		mpm:  mpm,
	}

	mpm.multiPlayerMap.Store(mp.id, mp)

	defaultPlayerData := map[string]any{
		"name":            p.Username(),
		"permission.role": "default",
		"permission.rank": "default",
	}

	err := db.SetPlayerData(id, defaultPlayerData)
	if err != nil {
		return nil, err
	}

	mpm.l.Info("created new multiplayer", "mp", mp, "playerId", id, "duration", time.Since(now))
	return mp, nil
}

const multiPlayerUpdateChannel = "update_mp"

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

func (mp *MultiPlayer) SetProxyId(id string, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.p = id

	var err error
	if notify {
		err = mp.Save("p", id)
	}

	return err
}

func (mp *MultiPlayer) GetBackendId() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.b
}

func (mp *MultiPlayer) SetBackendId(id string, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.b = id

	var err error
	if notify {
		err = mp.Save("b", id)
	}

	return err
}

func (mp *MultiPlayer) GetId() string {
	return mp.id
}

func (mp *MultiPlayer) GetName() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.name
}

func (mp *MultiPlayer) SetName(name string, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.name = name

	var err error
	if notify {
		err = mp.Save("name", name)
	}

	return err
}

func (mp *MultiPlayer) GetRole() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.role
}

func (mp *MultiPlayer) SetRole(role string, notify bool) error {
	if !IsValidRole(role) {
		return ErrIncorrectRole
	}

	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.role = role

	var err error
	if notify {
		err = mp.Save("permission.role", role)
	}

	return err
}

// Check if multiplayer has one of the following roles: admin, builder or moderator.
func (mp *MultiPlayer) IsPrivileged() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.role == RoleAdmin || mp.role == RoleBuilder || mp.role == RoleModerator
}

func (mp *MultiPlayer) GetRank() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.rank
}

func (mp *MultiPlayer) SetRank(rank string, notify bool) error {
	if !IsValidRank(rank) {
		return ErrIncorrectRank
	}

	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.rank = rank

	var err error
	if notify {
		err = mp.Save("permission.rank", rank)
	}

	return err
}

func (mp *MultiPlayer) IsOnline() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.online
}

func (mp *MultiPlayer) SetOnline(online bool, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.online = online

	var err error
	if notify {
		err = mp.Save("online", online)
	}

	return err
}

func (mp *MultiPlayer) IsVanished() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.vanished
}

func (mp *MultiPlayer) SetVanished(vanished bool, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.vanished = vanished

	var err error
	if notify {
		err = mp.Save("vanished", vanished)
	}

	return err
}

func (mp *MultiPlayer) GetFriends() []*MultiPlayer {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.friends
}

func (mp *MultiPlayer) SetFriends(friends []*MultiPlayer, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.friends = friends

	var err error
	if notify {
		err = mp.Save("friends", friends)
	}

	return err
}

func (mp *MultiPlayer) AddFriend(friend *MultiPlayer, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.friends = append(mp.friends, friend)
	var ids []string

	for _, friend := range mp.friends {
		ids = append(ids, friend.id)
	}

	var err error
	if notify {
		err = mp.Save("friends", ids)
	}

	return err
}
