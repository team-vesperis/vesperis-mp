package playermanager

import (
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiPlayerManager struct {
	multiPlayerMap map[uuid.UUID]*multi.Player
	mu             sync.RWMutex

	ownerMP *multi.Proxy

	l  *logger.Logger
	db *database.Database
}

func (mpm *MultiPlayerManager) GetOwnerMultiProxy() *multi.Proxy {
	return mpm.ownerMP
}

func Init(l *logger.Logger, db *database.Database, p *multi.Proxy) *MultiPlayerManager {
	now := time.Now()

	mpm := &MultiPlayerManager{
		multiPlayerMap: map[uuid.UUID]*multi.Player{},
		l:              l,
		db:             db,
		ownerMP:        p,
	}

	// start update listener
	mpm.db.CreateListener(multi.UpdateMultiPlayerChannel, mpm.createUpdateListener())

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

func (mpm *MultiPlayerManager) createUpdateListener() func(msg *redis.Message) {
	return func(msg *redis.Message) {
		m := msg.Payload
		s := strings.Split(m, "_")

		originProxy := s[0]
		// from own proxy, no update needed
		if mpm.ownerMP.GetId().String() == originProxy {
			return
		}

		id, err := uuid.Parse(s[1])
		if err != nil {
			mpm.l.Error("multiplayer update channel parse uuid error", "parsed uuid", s[1], "error", err)
			return
		}

		k := s[2]

		mp, err := mpm.GetMultiPlayer(id)
		if err != nil {
			mpm.l.Error("multiplayer update channel get multiplayer error", "playerId", id, "error", err)
			return
		}

		if k == "new" {
			return
		}

		dataKey, err := key.GetPlayerKey(k)
		if err != nil {
			mpm.l.Error("multiplayer update channel get data key error", "playerId", id, "key", k, "error", err)
			return
		}

		mp.Update(dataKey)
	}
}

func (mpm *MultiPlayerManager) NewMultiPlayer(p proxy.Player) (*multi.Player, error) {
	now := time.Now()
	id := p.ID()

	data := &data.PlayerData{
		ProxyId:   uuid.Nil,
		BackendId: uuid.Nil,
		Username:  p.Username(),
		Nickname:  p.Username(),
		Permission: &data.PermissionData{
			Role: multi.RoleDefault,
			Rank: multi.RankDefault,
		},
		Ban: &data.BanData{
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
	m := mpm.ownerMP.GetId().String() + "_" + id.String() + "_new"
	err = mpm.db.Publish(multi.UpdateMultiPlayerChannel, m)
	if err != nil {
		return nil, err
	}

	mpm.GetLogger().Info("created new multiplayer", "playerId", id, "duration", time.Since(now))
	return mp, nil
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

	mp := multi.NewPlayer(id, mpm.ownerMP.GetId(), mpm.db, data)

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
