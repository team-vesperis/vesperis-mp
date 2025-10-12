package multi

import (
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
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
	friendIds []uuid.UUID

	mu sync.RWMutex
}

func NewPlayer(id uuid.UUID, data *util.PlayerData) *Player {
	mp := &Player{
		id: id,
	}

	mp.pi = newPermissionInfo(mp, data)
	mp.bi = newBanInfo(mp, data)

	mp.username = data.Username
	mp.nickname = data.Nickname
	mp.online = data.Online
	mp.vanished = data.Vanished
	mp.lastSeen = data.LastSeen
	mp.friendIds = data.Friends

	return mp
}

type Proxymanager interface {
	GetMultiProxy(id uuid.UUID) (*Proxy, error)
	GetMultiBackend(id uuid.UUID) (*Backend, error)
}

var ErrProxyManagerNotSet = errors.New("proxy manager not set")
var proxyManagerInstance Proxymanager

func SetProxyManager(pm Proxymanager) {
	proxyManagerInstance = pm
}

type PlayerManager interface {
	Save(id uuid.UUID, key util.PlayerKey, val any) error
}

var ErrPlayerManagerNotSet = errors.New("player manager not set")
var playerManagerInstance PlayerManager

func SetPlayerManager(pm PlayerManager) {
	playerManagerInstance = pm
}

// Update specific value of the multi player into the database
// Notifies other proxies to update that value
func (mp *Player) save(key util.PlayerKey, val any) error {
	if playerManagerInstance == nil {
		return ErrPlayerManagerNotSet
	}

	if !slices.Contains(util.AllowedPlayerKeys, key) {
		return util.ErrIncorrectPlayerKey
	}

	return playerManagerInstance.Save(mp.id, key, val)
}

func (mp *Player) Update(key util.PlayerKey, db *database.Database) {
	if proxyManagerInstance == nil {
		return
	}

	if !slices.Contains(util.AllowedPlayerKeys, key) {
		return
	}

	switch key {
	case util.PlayerKey_ProxyId:
		var proxyId uuid.UUID
		db.GetPlayerDataField(mp.id, util.PlayerKey_ProxyId, &proxyId)
		p, err := proxyManagerInstance.GetMultiProxy(proxyId)
		if err == nil {
			mp.setProxy(p, false)
		}

	case util.PlayerKey_BackendId:
		var backendId uuid.UUID
		db.GetPlayerDataField(mp.id, util.PlayerKey_BackendId, &backendId)
		b, err := proxyManagerInstance.GetMultiBackend(backendId)
		if err == nil {
			mp.setBackend(b, false)
		}

	case util.PlayerKey_Username:
		var username string
		db.GetPlayerDataField(mp.id, util.PlayerKey_Username, &username)
		mp.setUsername(username, false)

	case util.PlayerKey_Nickname:
		var nickname string
		db.GetPlayerDataField(mp.id, util.PlayerKey_Nickname, &nickname)
		mp.setNickname(nickname, false)

	case util.PlayerKey_Permission_Role:
		var role string
		db.GetPlayerDataField(mp.id, util.PlayerKey_Permission_Role, &role)
		mp.pi.setRole(role, false)

	case util.PlayerKey_Permission_Rank:
		var rank string
		db.GetPlayerDataField(mp.id, util.PlayerKey_Permission_Rank, &rank)
		mp.pi.setRank(rank, false)

	case util.PlayerKey_Ban_Banned:
		var banned bool
		db.GetPlayerDataField(mp.id, util.PlayerKey_Ban_Banned, &banned)
		mp.bi.setBanned(banned, false)

	case util.PlayerKey_Ban_Reason:
		var reason string
		db.GetPlayerDataField(mp.id, util.PlayerKey_Ban_Reason, &reason)
		mp.bi.setReason(reason, false)

	case util.PlayerKey_Ban_Permanently:
		var permanently bool
		db.GetPlayerDataField(mp.id, util.PlayerKey_Ban_Permanently, &permanently)
		mp.bi.setPermanently(permanently, false)

	case util.PlayerKey_Ban_Expiration:
		var expiration time.Time
		db.GetPlayerDataField(mp.id, util.PlayerKey_Ban_Expiration, &expiration)
		mp.bi.setExpiration(expiration, false)

	case util.PlayerKey_Online:
		var online bool
		db.GetPlayerDataField(mp.id, util.PlayerKey_Online, &online)
		mp.setOnline(online, false)

	case util.PlayerKey_Vanished:
		var vanished bool
		db.GetPlayerDataField(mp.id, util.PlayerKey_Vanished, &vanished)
		mp.setVanished(vanished, false)

	case util.PlayerKey_LastSeen:
		var lastSeen *time.Time
		db.GetPlayerDataField(mp.id, util.PlayerKey_LastSeen, &lastSeen)
		mp.setLastSeen(lastSeen, false)

	case util.PlayerKey_Friends:
		var friends []uuid.UUID
		db.GetPlayerDataField(mp.id, util.PlayerKey_Friends, &friends)
		mp.setFriendsIds(friends, false)
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
		return mp.save(util.PlayerKey_ProxyId, mproxy.id)
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
	mp.b = mb
	mp.mu.Unlock()

	if notify {
		return mp.save(util.PlayerKey_BackendId, mb.id)
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
	mp.username = name
	mp.mu.Unlock()

	if notify {
		return mp.save(util.PlayerKey_Username, name)
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
	mp.username = name
	mp.mu.Unlock()

	if notify {
		return mp.save(util.PlayerKey_Nickname, name)
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
	mp.online = online
	mp.mu.Unlock()

	if notify {
		return mp.save(util.PlayerKey_Online, online)
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
	mp.vanished = vanished
	mp.mu.Unlock()

	if notify {
		return mp.save(util.PlayerKey_Vanished, vanished)
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
	mp.lastSeen = lastSeen
	mp.mu.Unlock()

	if notify {
		return mp.save(util.PlayerKey_LastSeen, lastSeen)
	}

	return nil
}

func (mp *Player) GetFriendsIds() []uuid.UUID {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return slices.Clone(mp.friendIds)
}

func (mp *Player) SetFriendsIds(ids []uuid.UUID) error {
	return mp.setFriendsIds(ids, true)
}

func (mp *Player) setFriendsIds(ids []uuid.UUID, notify bool) error {
	mp.mu.Lock()
	mp.friendIds = ids
	mp.mu.Unlock()

	if notify {
		return mp.save(util.PlayerKey_Friends, ids)
	}

	return nil
}

func (mp *Player) AddFriendId(id uuid.UUID) error {
	return mp.addFriendId(id, true)
}

func (mp *Player) addFriendId(id uuid.UUID, notify bool) error {
	mp.mu.Lock()
	if !slices.Contains(mp.friendIds, id) {
		mp.friendIds = append(mp.friendIds, id)
	}
	mp.mu.Unlock()

	if notify {
		return mp.save(util.PlayerKey_Friends, mp.GetFriendsIds())
	}

	return nil
}

var ErrFriendNotFound = errors.New("friend not found")

func (mp *Player) RemoveFriendId(id uuid.UUID) error {
	return mp.removeFriendId(id, true)
}

func (mp *Player) removeFriendId(id uuid.UUID, notify bool) error {
	mp.mu.Lock()
	i := slices.Index(mp.friendIds, id)
	if i == -1 {
		mp.mu.Unlock()
		return ErrFriendNotFound
	}
	mp.friendIds = slices.Delete(mp.friendIds, i, i+1)
	mp.mu.Unlock()

	if notify {
		return mp.save(util.PlayerKey_Friends, mp.GetFriendsIds())
	}

	return nil
}
