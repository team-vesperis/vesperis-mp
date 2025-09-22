package multi

import (
	"errors"
	"slices"
	"sync"
	"time"

	"go.minekube.com/gate/pkg/util/uuid"
)

type Player struct {
	// The MultiProxy the player is located on.
	// can be nil!
	mp *Proxy

	// The MultiBackend the player is located on.
	// can be nil!
	mb *Backend

	// The id of the underlying player
	id uuid.UUID

	// The username of the underlying player
	username string

	nickname string

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

func NewPlayer(id uuid.UUID, data map[string]any) *Player {
	mp := &Player{
		id: id,
	}

	mp.pi = newPermissionInfo(mp)
	mp.bi = newBanInfo(mp)
	username, ok := data["username"].(string)
	if ok {
		mp.username = username
	}

	nickname, ok := data["nickname"].(string)
	if ok {
		mp.nickname = nickname
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
func (mp *Player) save(key string, val any) error {
	if playerManagerInstance == nil {
		return errors.New("player manager not set")
	}
	return playerManagerInstance.Save(mp.id, key, val)
}

func (mp *Player) Update(key string, val any) {
	switch key {
	case "username":
		name, ok := val.(string)
		if ok {
			mp.mu.Lock()
			mp.username = name
			mp.mu.Unlock()
		}
	case "nickname":
		name, ok := val.(string)
		if ok {
			mp.mu.Lock()
			mp.nickname = name
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

var ErrProxyNilWhileOnline = errors.New("proxy is nil but player is online")

// can return nil!
func (mp *Player) GetProxy() *Proxy {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.mp
}

func (mp *Player) SetProxy(mproxy *Proxy) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.mp = mproxy

	return mp.save("mp", mproxy)
}

var ErrBackendNilWhileOnline = errors.New("backend is nil but player is online")

// can return nil!
func (mp *Player) GetBackend() *Backend {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.mb
}

func (mp *Player) SetBackend(mb *Backend) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.mb = mb

	return mp.save("mb", mb)
}

func (mp *Player) GetId() uuid.UUID {
	return mp.id
}

func (mp *Player) GetUsername() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.username
}

func (mp *Player) SetUsername(name string) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.username = name

	return mp.save("username", name)
}

func (mp *Player) GetNickname() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.nickname
}

func (mp *Player) SetNickname(name string) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.nickname = name

	return mp.save("nickname", name)
}

func (mp *Player) GetPermissionInfo() *permissionInfo {
	return mp.pi
}

func (mp *Player) GetBanInfo() *banInfo {
	return mp.bi
}

func (mp *Player) IsOnline() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.online
}

func (mp *Player) SetOnline(online bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.online = online

	return mp.save("online", online)
}

func (mp *Player) IsVanished() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.vanished
}

func (mp *Player) SetVanished(vanished bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.vanished = vanished

	return mp.save("vanished", vanished)
}

func (mp *Player) GetLastSeen() time.Time {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.lastSeen
}

func (mp *Player) SetLastSeen(time time.Time) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.lastSeen = time

	return mp.save("last_seen", time)
}

func (mp *Player) GetFriendsIds() []uuid.UUID {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return slices.Clone(mp.friendIds)
}

func (mp *Player) SetFriendsIds(ids []uuid.UUID) error {
	mp.mu.Lock()
	mp.friendIds = slices.Clone(ids)
	mp.mu.Unlock()

	return mp.save("friends", ids)
}

func (mp *Player) AddFriendId(id uuid.UUID) error {
	mp.mu.Lock()
	if !slices.Contains(mp.friendIds, id) {
		mp.friendIds = append(mp.friendIds, id)
	}
	mp.mu.Unlock()

	return mp.save("friends", mp.GetFriendsIds())
}

var ErrFriendNotFound = errors.New("friend not found")

func (mp *Player) RemoveFriendId(id uuid.UUID) error {
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
