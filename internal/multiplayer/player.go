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
	p uuid.UUID

	// The backend id on which the underlying player is/was located
	b uuid.UUID

	// The id of the underlying player
	id uuid.UUID

	// The username of the underlying player
	name string

	// The permission info of the multiplayer.
	pi *permissionInfo

	// The ban info of the multiplayer.
	bi *banInfo

	online bool

	vanished bool

	// List of friends.
	friends []*MultiPlayer

	mu sync.RWMutex

	mpm *MultiPlayerManager
}

var defaultPlayerData = map[string]any{
	"permission.role": RoleDefault,
	"permission.rank": RankDefault,
	"online":          false,
	"vanished":        false,
}

// New returns a new multiplayer
func New(p proxy.Player, db *database.Database, mpm *MultiPlayerManager) (*MultiPlayer, error) {
	now := time.Now()
	id := p.ID()

	defaultPlayerData["name"] = p.Username()
	err := db.SetPlayerData(id, defaultPlayerData)
	if err != nil {
		return nil, err
	}

	mp, err := mpm.CreateMultiPlayerFromDatabase(id)
	if err != nil {
		return nil, err
	}

	// update every proxies' map
	m := id.String() + "_new"
	err = mpm.db.Publish(multiPlayerUpdateChannel, m)
	if err != nil {
		return nil, err
	}

	mpm.l.Info("created new multiplayer", "playerId", id, "duration", time.Since(now))
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

	m := mp.mpm.ownerProxyId.String() + "_" + mp.id.String() + "_" + key
	return mp.mpm.db.Publish(multiPlayerUpdateChannel, m)
}

func (mp *MultiPlayer) GetProxyId() uuid.UUID {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.p
}

func (mp *MultiPlayer) SetProxyId(id uuid.UUID) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.p = id

	return mp.save("p", id)
}

func (mp *MultiPlayer) GetBackendId() uuid.UUID {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.b
}

func (mp *MultiPlayer) SetBackendId(id uuid.UUID) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.b = id

	return mp.save("b", id)
}

func (mp *MultiPlayer) GetId() uuid.UUID {
	return mp.id
}

func (mp *MultiPlayer) GetName() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.name
}

func (mp *MultiPlayer) SetName(name string) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.name = name

	return mp.save("name", name)
}

func (mp *MultiPlayer) GetPermissionInfo() *permissionInfo {
	return mp.pi
}

func (mp *MultiPlayer) GetBanInfo() *banInfo {
	return mp.bi
}

func (mp *MultiPlayer) IsOnline() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.online
}

func (mp *MultiPlayer) SetOnline(online bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.online = online

	return mp.save("online", online)
}

func (mp *MultiPlayer) IsVanished() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.vanished
}

func (mp *MultiPlayer) SetVanished(vanished bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.vanished = vanished

	return mp.save("vanished", vanished)
}

func (mp *MultiPlayer) GetFriends() []*MultiPlayer {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.friends
}

func (mp *MultiPlayer) GetFriendsIds() []uuid.UUID {
	var ids []uuid.UUID
	for _, f := range mp.GetFriends() {
		ids = append(ids, f.id)
	}

	return ids
}

func (mp *MultiPlayer) SetFriends(friends []*MultiPlayer) error {
	mp.mu.Lock()
	mp.friends = friends
	mp.mu.Unlock()

	return mp.save("friends", mp.GetFriendsIds())
}

func (mp *MultiPlayer) AddFriend(friend *MultiPlayer) error {
	mp.mu.Lock()
	mp.friends = append(mp.friends, friend)
	mp.mu.Unlock()

	return mp.save("friends", mp.GetFriendsIds())
}

var ErrFriendNotFound = errors.New("friend not found")

func (mp *MultiPlayer) RemoveFriend(friend *MultiPlayer) error {
	mp.mu.Lock()
	if !slices.Contains(mp.friends, friend) {
		return ErrFriendNotFound
	}

	for i, f := range mp.friends {
		if f == friend {
			mp.friends = slices.Delete(mp.friends, i, i+1)
			break
		}
	}
	mp.mu.Unlock()

	return mp.save("friends", mp.GetFriendsIds())
}
