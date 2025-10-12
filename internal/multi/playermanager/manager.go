package playermanager

import (
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiPlayerManager struct {
	multiPlayerMap map[uuid.UUID]*multi.Player
	mu             sync.RWMutex

	ownerProxyId uuid.UUID

	l  *logger.Logger
	db *database.Database
}

const UpdateChannel = "update_mp"

func Init(l *logger.Logger, db *database.Database, id uuid.UUID) *MultiPlayerManager {
	now := time.Now()

	mpm := &MultiPlayerManager{
		multiPlayerMap: map[uuid.UUID]*multi.Player{},
		l:              l,
		db:             db,
		ownerProxyId:   id,
	}

	multi.SetPlayerManager(mpm)

	// start update listener
	mpm.db.CreateListener(UpdateChannel, mpm.createUpdateListener())

	// fill map
	_, err := mpm.GetAllMultiPlayersFromDatabase()
	if err != nil {
		mpm.l.Error("filling up multiplayer map error", "error", err)
	}

	mpm.l.Info("initialized multiplayer manager", "duration", time.Since(now))
	return mpm
}

func (mpm *MultiPlayerManager) GetDatabase() *database.Database {
	return mpm.db
}

func (mpm *MultiPlayerManager) GetLogger() *logger.Logger {
	return mpm.l
}

func (mpm *MultiPlayerManager) GetOwnerProxyId() uuid.UUID {
	return mpm.ownerProxyId
}

func (mpm *MultiPlayerManager) Save(id uuid.UUID, key util.PlayerKey, val any) error {
	err := mpm.db.SetPlayerDataField(id, key, val)
	if err != nil {
		return err
	}

	m := mpm.ownerProxyId.String() + "_" + id.String() + "_" + key.String()
	return mpm.db.Publish(UpdateChannel, m)
}

func (mpm *MultiPlayerManager) createUpdateListener() func(msg *redis.Message) {
	return func(msg *redis.Message) {
		m := msg.Payload
		s := strings.Split(m, "_")

		originProxy := s[0]
		if mpm.ownerProxyId.String() == originProxy {
			return
		}

		id, err := uuid.Parse(s[1])
		if err != nil {
			mpm.l.Error("multiplayer update channel parse uuid error", "parsed uuid", s[1], "error", err)
			return
		}

		key := s[2]

		mp, err := mpm.GetMultiPlayer(id)
		if err != nil {
			mpm.l.Error("multiplayer update channel get multiplayer error", "playerId", id, "error", err)
		}

		if key == "new" {
			return
		}

		datakey, err := util.GetPlayerKey(key)
		if err != nil {
			mpm.l.Error("multiplayer update channel get data key error", "playerId", id, "key", key, "error", err)
		}

		mp.Update(datakey, mpm.db)
	}
}

func (mpm *MultiPlayerManager) NewMultiPlayer(p proxy.Player) (*multi.Player, error) {
	now := time.Now()
	id := p.ID()

	data := &util.PlayerData{
		ProxyId:   uuid.Nil,
		BackendId: uuid.Nil,
		Username:  p.Username(),
		Nickname:  p.Username(),
		Permission: &util.PermissionData{
			Role: multi.RoleDefault,
			Rank: multi.RankDefault,
		},
		Ban: &util.BanData{
			Banned:      false,
			Reason:      "",
			Permanently: false,
			Expiration:  time.Time{},
		},
		Online:   false,
		Vanished: false,
		LastSeen: &time.Time{},
		Friends:  make([]uuid.UUID, 0),
	}

	err := mpm.db.SetPlayerData(id, data)
	if err != nil {
		return nil, err
	}

	mp, err := mpm.CreateMultiPlayerFromDatabase(id)
	if err != nil {
		return nil, err
	}

	// update every proxies' map
	m := mpm.ownerProxyId.String() + "_" + id.String() + "_new"
	err = mpm.db.Publish(UpdateChannel, m)
	if err != nil {
		return nil, err
	}

	mpm.GetLogger().Info("created new multiplayer", "playerId", id, "duration", time.Since(now))
	return mp, nil
}

func (mpm *MultiPlayerManager) SavePlayer(player *multi.Player) error {
	data := &util.PlayerData{
		Username: player.GetUsername(),
		Nickname: player.GetNickname(),
		Permission: &util.PermissionData{
			Role: player.GetPermissionInfo().GetRole(),
			Rank: player.GetPermissionInfo().GetRank(),
		},
		Ban: &util.BanData{
			Banned:      player.GetBanInfo().IsBanned(),
			Reason:      player.GetBanInfo().GetReason(),
			Permanently: player.GetBanInfo().IsPermanently(),
			Expiration:  player.GetBanInfo().GetExpiration(),
		},
		Online:   player.IsOnline(),
		Vanished: player.IsVanished(),
		LastSeen: player.GetLastSeen(),
		Friends:  player.GetFriendsIds(),
	}

	err := mpm.db.SetPlayerData(player.GetId(), data)
	return err
}

/*
Gets a multiplayer in two ways:

1. Use the map with all multiplayers.
This method will be used the most since all existing players are in the map located.

2. Create a new multiplayer based on the player data from the database.
*/
func (mpm *MultiPlayerManager) GetMultiPlayer(id uuid.UUID) (*multi.Player, error) {
	mpm.mu.RLock()
	mp, ok := mpm.multiPlayerMap[id]
	mpm.mu.RUnlock()

	if ok {
		return mp, nil
	}

	return mpm.CreateMultiPlayerFromDatabase(id)
}

// if player has never joined before, this function will return database.ErrDataNotFound
func (mpm *MultiPlayerManager) CreateMultiPlayerFromDatabase(id uuid.UUID) (*multi.Player, error) {
	data, err := mpm.db.GetPlayerData(id)
	if err != nil {
		return nil, err
	}

	mp := multi.NewPlayer(id, data)

	mpm.mu.Lock()
	mpm.multiPlayerMap[id] = mp
	mpm.mu.Unlock()

	return mp, nil
}

func (mpm *MultiPlayerManager) GetAllMultiPlayers() []*multi.Player {
	var l []*multi.Player

	mpm.mu.RLock()
	for _, mp := range mpm.multiPlayerMap {
		l = append(l, mp)
	}
	mpm.mu.RUnlock()

	return l
}

func (mpm *MultiPlayerManager) GetAllMultiPlayersFromDatabase() ([]*multi.Player, error) {
	var l []*multi.Player

	i, err := mpm.db.GetAllPlayerIds()
	if err != nil {
		return nil, err
	}

	for _, id := range i {
		mp, err := mpm.GetMultiPlayer(id)
		// this is not possible... probably
		if err != nil {
			return nil, err
		}

		l = append(l, mp)
	}

	return l, nil
}

func (mpm *MultiPlayerManager) GetAllOnlinePlayers() []*multi.Player {
	var l []*multi.Player

	all := mpm.GetAllMultiPlayers()

	for _, mp := range all {
		if mp.IsOnline() {
			l = append(l, mp)
		}
	}

	return l
}
