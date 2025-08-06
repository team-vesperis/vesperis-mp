package multiplayer

import (
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiPlayer struct {
	// The proxy id on which the underlying player is/was located
	p string

	// The backend id on which the underlying player is/was located
	b string

	// The id of the underlying player
	id uuid.UUID

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

// New returns a new multiplayer
func New(p proxy.Player, db *database.Database, mpm *MultiPlayerManager) (*MultiPlayer, error) {
	now := time.Now()
	id := p.ID()

	defaultPlayerData := map[string]any{
		"name":            p.Username(),
		"permission.role": RoleDefault,
		"permission.rank": RankDefault,
		"online":          false,
		"vanished":        false,
	}

	err := db.SetPlayerData(id, defaultPlayerData)
	if err != nil {
		return nil, err
	}

	mp, err := mpm.CreateMultiPlayerFromDatabase(id)
	if err != nil {
		return nil, err
	}

	mpm.l.Info("created new multiplayer", "mp", mp, "playerId", id, "duration", time.Since(now))
	return mp, nil
}

const multiPlayerUpdateChannel = "update_mp"

// Update specific value of the multi player into the database
// Notifies other proxies to update that value
func (mp *MultiPlayer) save(key string, value any) error {
	err := mp.mpm.db.SetPlayerDataField(mp.id, key, value)
	if err != nil {
		return err
	}

	m := mp.id.String() + "_" + key
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
		err = mp.save("p", id)
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
		err = mp.save("b", id)
	}

	return err
}

func (mp *MultiPlayer) GetId() uuid.UUID {
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
		err = mp.save("name", name)
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
		err = mp.save("permission.role", role)
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
		err = mp.save("permission.rank", rank)
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
		err = mp.save("online", online)
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
		err = mp.save("vanished", vanished)
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
		err = mp.save("friends", friends)
	}

	return err
}

func (mp *MultiPlayer) AddFriend(friend *MultiPlayer, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.friends = append(mp.friends, friend)
	var ids []uuid.UUID

	for _, friend := range mp.friends {
		ids = append(ids, friend.id)
	}

	var err error
	if notify {
		err = mp.save("friends", ids)
	}

	return err
}

var ErrFriendNotFound = errors.New("friend not found")

func (mp *MultiPlayer) RemoveFriend(friend *MultiPlayer, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if !slices.Contains(mp.friends, friend) {
		return ErrFriendNotFound
	}

	for i, f := range mp.friends {
		if f == friend {
			mp.friends = slices.Delete(mp.friends, i, i+1)
		}
	}

	var err error
	if notify {
		err = mp.save("friends", mp.friends)
	}

	return err
}
