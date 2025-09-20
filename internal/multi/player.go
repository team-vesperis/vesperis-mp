package multi

import (
	"errors"
	"slices"
	"sync"
	"time"

	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiPlayer struct {
	// The MultiProxy the player is located on.
	// can be nil!
	mp *MultiProxy

	// The MultiBackend the player is located on.
	// can be nil!
	mb *MultiBackend

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

	lastSeen time.Time

	// List of friend UUIDs.
	friendIds []uuid.UUID

	mu sync.RWMutex
}

func NewMultiPlayer(id uuid.UUID) *MultiPlayer {
	mp := &MultiPlayer{
		id: id,
	}

	mp.pi = NewPermissionInfo(mp)
	mp.bi = NewBanInfo(mp)

	return mp
}

func NewMultiPlayerWithData(id uuid.UUID, data map[string]any) *MultiPlayer {
	// create an empty multiplayer, then fill in using the data
	mp := NewMultiPlayer(id)

	name, ok := data["name"].(string)
	if ok {
		mp.name = name
	}

	permission, ok := data["permission"].(map[string]any)
	if ok {
		role, ok := permission["role"].(string)
		if ok {
			mp.pi.role = role
		}

		rank, ok := permission["rank"].(string)
		if ok {
			mp.pi.rank = rank
		}
	}

	online, ok := data["online"].(bool)
	if ok {
		mp.online = online
	}

	vanished, ok := data["vanished"].(bool)
	if ok {
		mp.vanished = vanished
	}

	return mp
}

type PlayerManager interface {
	Save(id uuid.UUID, key string, val any) error
}

var playerManagerInstance PlayerManager

func SetPlayerManager(pm PlayerManager) {
	playerManagerInstance = pm
}

// Update specific value of the multi player into the database
// Notifies other proxies to update that value
func (mp *MultiPlayer) save(key string, val any) error {
	if playerManagerInstance == nil {
		return errors.New("player manager not set")
	}
	return playerManagerInstance.Save(mp.id, key, val)
}

func (mp *MultiPlayer) Update(key string, val any) {
	switch key {
	case "name":
		name, ok := val.(string)
		if ok {
			mp.mu.Lock()
			mp.name = name
			mp.mu.Unlock()
		}
	case "permission.role":
		role, ok := val.(string)
		if ok {
			mp.pi.mu.Lock()
			mp.pi.role = role
			mp.pi.mu.Unlock()
		}
	case "permission.rank":
		rank, ok := val.(string)
		if ok {
			mp.pi.mu.Lock()
			mp.pi.rank = rank
			mp.pi.mu.Unlock()
		}
	case "online":
		online, ok := val.(bool)
		if ok {
			mp.mu.Lock()
			mp.online = online
			mp.mu.Unlock()
		}
	case "vanished":
		vanished, ok := val.(bool)
		if ok {
			mp.mu.Lock()
			mp.vanished = vanished
			mp.mu.Unlock()
		}
	case "friends":
		list, ok := val.([]any)
		if ok {
			var mp_list []uuid.UUID
			for _, l := range list {
				id, ok := l.(uuid.UUID)
				if ok {
					mp_list = append(mp_list, id)
				}
			}

			mp.mu.Lock()
			mp.friendIds = mp_list
			mp.mu.Unlock()
		}

	case "last_seen":
		time, ok := val.(time.Time)
		if ok {
			mp.mu.Lock()
			mp.lastSeen = time
			mp.mu.Unlock()
		}
	}
}

func (mp *MultiPlayer) GetMultiProxy() *MultiProxy {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.mp
}

func (mp *MultiPlayer) SetMultiProxy(mproxy *MultiProxy) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.mp = mproxy

	return mp.save("mp", mproxy)
}

func (mp *MultiPlayer) GetMultiBackend() *MultiBackend {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.mb
}

func (mp *MultiPlayer) SetMultiBackend(mb *MultiBackend) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.mb = mb

	return mp.save("mb", mb)
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

func (mp *MultiPlayer) GetLastSeen() time.Time {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.lastSeen
}

func (mp *MultiPlayer) SetLastSeen(time time.Time) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.lastSeen = time

	return mp.save("last_seen", time)
}

func (mp *MultiPlayer) GetFriendsIds() []uuid.UUID {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return slices.Clone(mp.friendIds)
}

func (mp *MultiPlayer) SetFriendsIds(ids []uuid.UUID) error {
	mp.mu.Lock()
	mp.friendIds = slices.Clone(ids)
	mp.mu.Unlock()

	return mp.save("friends", ids)
}

func (mp *MultiPlayer) AddFriendId(id uuid.UUID) error {
	mp.mu.Lock()
	if !slices.Contains(mp.friendIds, id) {
		mp.friendIds = append(mp.friendIds, id)
	}
	mp.mu.Unlock()

	return mp.save("friends", mp.GetFriendsIds())
}

var ErrFriendNotFound = errors.New("friend not found")

func (mp *MultiPlayer) RemoveFriendId(id uuid.UUID) error {
	mp.mu.Lock()
	i := slices.Index(mp.friendIds, id)
	if i == -1 {
		mp.mu.Unlock()
		return ErrFriendNotFound
	}
	mp.friendIds = slices.Delete(mp.friendIds, i, i+1)
	mp.mu.Unlock()

	return mp.save("friends", mp.GetFriendsIds())
}
