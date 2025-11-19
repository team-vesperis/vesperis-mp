package multi

import (
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
	"go.minekube.com/gate/pkg/util/uuid"
)

type Player struct {
	// The MultiProxy the player is located on.
	// can be nil!
	p *Proxy

	// The MultiBackend the player is located on.
	// can be nil!
	b *Backend

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

	lastSeen *time.Time

	// List of friend UUIDs.
	friends []uuid.UUID

	managerId uuid.UUID
	l         *logger.Logger
	db        *database.Database
	mu        sync.RWMutex
}

func NewPlayer(id, mId uuid.UUID, l *logger.Logger, db *database.Database, data *data.PlayerData) *Player {
	mp := &Player{
		id:        id,
		managerId: mId,
		l:         l,
		db:        db,
		mu:        sync.RWMutex{},
	}

	mp.pi = newPermissionInfo(mp, data)
	mp.bi = newBanInfo(mp, data)

	mp.username = data.Username
	mp.nickname = data.Nickname
	mp.online = data.Online
	mp.vanished = data.Vanished
	mp.lastSeen = data.LastSeen
	mp.friends = data.Friends

	return mp
}

var ErrPlayerNotFound = errors.New("player not found")

type MultiManager interface {
	GetMultiProxy(id uuid.UUID) (*Proxy, error)
	GetMultiBackend(id uuid.UUID) (*Backend, error)
}

var ErrMultiManagerNotSet = errors.New("multi manager not set")
var proxyManagerInstance MultiManager

func SetMultiManager(pm MultiManager) {
	proxyManagerInstance = pm
}

const UpdateMultiPlayerChannel = "update_multiplayer"

// Update specific value of the multi player into the database.
// Notifies other proxies to update that value for themselves.
func (mp *Player) save(k key.PlayerKey, val any) error {
	err := mp.db.SetPlayerDataField(mp.id, k, val)
	if err != nil {
		return err
	}

	m := mp.managerId.String() + "_" + mp.id.String() + "_" + k.String()
	return mp.db.Publish(UpdateMultiPlayerChannel, m)
}

func (mp *Player) Update(k key.PlayerKey) {
	if proxyManagerInstance == nil {
		return
	}

	var err error

	switch k {
	case key.PlayerKey_ProxyId:
		var proxyId uuid.UUID
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_ProxyId, &proxyId)
		if proxyId == uuid.Nil {
			mp.setProxy(nil, false)
			break
		}
		p, err := proxyManagerInstance.GetMultiProxy(proxyId)
		if err == nil {
			mp.setProxy(p, false)
		}
	case key.PlayerKey_BackendId:
		var backendId uuid.UUID
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_BackendId, &backendId)
		if backendId == uuid.Nil {
			mp.setBackend(nil, false)
		}

		b, err := proxyManagerInstance.GetMultiBackend(backendId)
		if err == nil {
			mp.setBackend(b, false)
		}

	case key.PlayerKey_Username:
		var username string
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Username, &username)
		mp.setUsername(username, false)

	case key.PlayerKey_Nickname:
		var nickname string
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Nickname, &nickname)
		mp.setNickname(nickname, false)

	case key.PlayerKey_Permission_Role:
		var role string
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Permission_Role, &role)
		mp.pi.setRole(Role(role), false)

	case key.PlayerKey_Permission_Rank:
		var rank string
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Permission_Rank, &rank)
		mp.pi.setRank(Rank(rank), false)

	case key.PlayerKey_Ban_Banned:
		var banned bool
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Ban_Banned, &banned)
		mp.bi.setBanned(banned, false)

	case key.PlayerKey_Ban_Reason:
		var reason string
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Ban_Reason, &reason)
		mp.bi.setReason(reason, false)

	case key.PlayerKey_Ban_Permanently:
		var permanently bool
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Ban_Permanently, &permanently)
		mp.bi.setPermanently(permanently, false)

	case key.PlayerKey_Ban_Expiration:
		var expiration time.Time
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Ban_Expiration, &expiration)
		mp.bi.setExpiration(expiration, false)

	case key.PlayerKey_Online:
		var online bool
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Online, &online)
		mp.setOnline(online, false)

	case key.PlayerKey_Vanished:
		var vanished bool
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Vanished, &vanished)
		mp.setVanished(vanished, false)

	case key.PlayerKey_LastSeen:
		var lastSeen *time.Time
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_LastSeen, &lastSeen)
		mp.setLastSeen(lastSeen, false)

	case key.PlayerKey_Friends:
		var friends []uuid.UUID
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Friends, &friends)
		mp.setFriendsIds(friends, false)
	}

	if err != nil {
		mp.l.Error("multiplayer update playerkey get field from database error", "error", err)
	}
}

var ErrProxyNilWhileOnline = errors.New("proxy is nil but player is online")

// can return nil!
func (mp *Player) GetProxy() *Proxy {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.p
}

func (mp *Player) SetProxy(mproxy *Proxy) error {
	return mp.setProxy(mproxy, true)
}

func (mp *Player) setProxy(mproxy *Proxy, notify bool) error {
	mp.mu.Lock()
	mp.p = mproxy
	mp.mu.Unlock()

	if notify {
		if mproxy == nil {
			return mp.save(key.PlayerKey_ProxyId, uuid.Nil)
		}

		return mp.save(key.PlayerKey_ProxyId, mproxy.id)
	}

	return nil
}

var ErrBackendNilWhileOnline = errors.New("backend is nil but player is online")

// can return nil!
func (mp *Player) GetBackend() *Backend {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.b
}

func (mp *Player) SetBackend(mb *Backend) error {
	return mp.setBackend(mb, true)
}

func (mp *Player) setBackend(mb *Backend, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.b = mb

	if notify {
		if mb == nil {
			return mp.save(key.PlayerKey_BackendId, uuid.Nil)
		}

		return mp.save(key.PlayerKey_BackendId, mb.id)
	}

	return nil
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
	return mp.setUsername(name, true)
}

func (mp *Player) setUsername(name string, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.username = name

	if notify {
		return mp.save(key.PlayerKey_Username, name)
	}

	return nil
}

func (mp *Player) GetNickname() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.nickname
}

func (mp *Player) SetNickname(name string) error {
	return mp.setNickname(name, true)
}

func (mp *Player) setNickname(name string, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.username = name

	if notify {
		return mp.save(key.PlayerKey_Nickname, name)
	}

	return nil
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
	return mp.setOnline(online, true)
}

func (mp *Player) setOnline(online bool, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.online = online

	if notify {
		return mp.save(key.PlayerKey_Online, online)
	}

	return nil
}

func (mp *Player) IsVanished() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.vanished
}

func (mp *Player) SetVanished(vanished bool) error {
	return mp.setVanished(vanished, true)
}

func (mp *Player) setVanished(vanished bool, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.vanished = vanished

	if notify {
		return mp.save(key.PlayerKey_Vanished, vanished)
	}

	return nil
}

func (mp *Player) GetLastSeen() *time.Time {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.lastSeen
}

func (mp *Player) SetLastSeen(lastSeen *time.Time) error {
	return mp.setLastSeen(lastSeen, true)
}

func (mp *Player) setLastSeen(lastSeen *time.Time, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.lastSeen = lastSeen

	if notify {
		return mp.save(key.PlayerKey_LastSeen, lastSeen)
	}

	return nil
}

func (mp *Player) GetFriendsIds() []uuid.UUID {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return slices.Clone(mp.friends)
}

func (mp *Player) SetFriendsIds(ids []uuid.UUID) error {
	return mp.setFriendsIds(ids, true)
}

func (mp *Player) setFriendsIds(ids []uuid.UUID, notify bool) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.friends = ids

	if notify {
		return mp.save(key.PlayerKey_Friends, ids)
	}

	return nil
}

func (mp *Player) AddFriendId(id uuid.UUID) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if !slices.Contains(mp.friends, id) {
		mp.friends = append(mp.friends, id)
	}

	return mp.save(key.PlayerKey_Friends, mp.friends)
}

var ErrFriendNotFound = errors.New("friend not found")

func (mp *Player) RemoveFriendId(id uuid.UUID) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	i := slices.Index(mp.friends, id)
	if i == -1 {
		return ErrFriendNotFound
	}
	mp.friends = slices.Delete(mp.friends, i, i+1)

	return mp.save(key.PlayerKey_Friends, mp.friends)

}
