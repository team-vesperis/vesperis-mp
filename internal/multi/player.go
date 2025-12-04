package multi

import (
	"errors"
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

	id uuid.UUID

	username string
	nickname string

	pmi *permissionInfo
	bi  *banInfo
	fi  *friendInfo
	pi  *partyInfo

	online   bool
	vanished bool

	lastSeen *time.Time

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

	mp.pmi = newPermissionInfo(mp, data)
	mp.bi = newBanInfo(mp, data)
	mp.fi = newFriendInfo(mp, data)
	mp.pi = newPartyInfo(mp, data)

	mp.username = data.Username
	mp.nickname = data.Nickname
	mp.online = data.Online
	mp.vanished = data.Vanished
	mp.lastSeen = data.LastSeen

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
	case key.PlayerKey_Proxy:
		var proxyId uuid.UUID
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Proxy, &proxyId)
		if proxyId == uuid.Nil {
			mp.setProxy(nil, false)
			break
		}
		p, err := proxyManagerInstance.GetMultiProxy(proxyId)
		if err == nil {
			mp.setProxy(p, false)
		}
	case key.PlayerKey_Backend:
		var backendId uuid.UUID
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Backend, &backendId)
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
		mp.pmi.setRole(Role(role), false)

	case key.PlayerKey_Permission_Rank:
		var rank string
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Permission_Rank, &rank)
		mp.pmi.setRank(Rank(rank), false)

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

	case key.PlayerKey_Friend_Friends:
		var friends []uuid.UUID
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Friend_Friends, &friends)
		mp.fi.setFriendsIds(friends, false)

	case key.PlayerKey_Friend_FriendPendingRequests:
		var friends []uuid.UUID
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Friend_FriendPendingRequests, &friends)
		mp.fi.setPendingFriendIds(friends, false)

	case key.PlayerKey_Friend_FriendRequests:
		var friends []uuid.UUID
		err = mp.db.GetPlayerDataField(mp.id, key.PlayerKey_Friend_FriendRequests, &friends)
		mp.fi.setFriendRequestIds(friends, false)
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
			return mp.save(key.PlayerKey_Proxy, uuid.Nil)
		}

		return mp.save(key.PlayerKey_Proxy, mproxy.id)
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
			return mp.save(key.PlayerKey_Backend, uuid.Nil)
		}

		return mp.save(key.PlayerKey_Backend, mb.id)
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
	return mp.pmi
}

func (mp *Player) GetBanInfo() *banInfo {
	return mp.bi
}

func (mp *Player) GetFriendInfo() *friendInfo {
	return mp.fi
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
