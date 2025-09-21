package playermanager

import (
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiPlayerManager struct {
	multiPlayerMap sync.Map

	ownerProxyId uuid.UUID

	l  *logger.Logger
	db *database.Database
}

const UpdateChannel = "update_mp"

func InitMultiPlayerManager(l *logger.Logger, db *database.Database, id uuid.UUID) *MultiPlayerManager {
	now := time.Now()

	mpm := &MultiPlayerManager{
		multiPlayerMap: sync.Map{},
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

func (mpm *MultiPlayerManager) Save(id uuid.UUID, key string, val any) error {
	err := mpm.db.SetPlayerDataField(id, key, val)
	if err != nil {
		return err
	}

	m := mpm.ownerProxyId.String() + "_" + id.String() + "_" + key
	return mpm.db.Publish(UpdateChannel, m)
}

func (mpm *MultiPlayerManager) createUpdateListener() func(msg *redis.Message) {
	return func(msg *redis.Message) {
		m := msg.Payload

		mpm.l.Info("received update message ")
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

		val, err := mpm.db.GetPlayerDataField(id, key)
		if err != nil {
			mpm.l.Error("multiplayer update channel get player data field error", "playerId", id, "error", err)
			return
		}

		mp.Update(key, val)
	}
}

func (mpm *MultiPlayerManager) NewMultiPlayer(p proxy.Player) (*multi.Player, error) {
	now := time.Now()
	id := p.ID()

	defaultPlayerData := map[string]any{
		"name":            p.Username(),
		"permission.role": multi.RoleDefault,
		"permission.rank": multi.RankDefault,
		"online":          false,
		"vanished":        false,
	}

	err := mpm.db.SetPlayerData(id, defaultPlayerData)
	if err != nil {
		return nil, err
	}

	mp, err := mpm.CreateMultiPlayerFromDatabase(id)
	if err != nil {
		return nil, err
	}

	// update every proxies' map
	m := id.String() + "_new"
	err = mpm.db.Publish(UpdateChannel, m)
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
	val, ok := mpm.multiPlayerMap.Load(id)
	if ok {
		mp, ok := val.(*multi.Player)
		if ok {
			return mp, nil
		} else {
			mpm.multiPlayerMap.Delete(id)
		}
	}

	return mpm.CreateMultiPlayerFromDatabase(id)
}

func (mpm *MultiPlayerManager) CreateMultiPlayerFromDatabase(id uuid.UUID) (*multi.Player, error) {
	data, err := mpm.db.GetPlayerData(id)
	if err != nil {
		return nil, err
	}

	mp := multi.NewMultiPlayerWithData(id, data)

	mpm.multiPlayerMap.Store(id, mp)
	return mp, nil
}

func (mpm *MultiPlayerManager) GetAllMultiPlayers() []*multi.Player {
	var l []*multi.Player

	mpm.multiPlayerMap.Range(func(key, value any) bool {
		mp, ok := value.(*multi.Player)
		if !ok {
			mpm.l.Info("detected incorrect value saved in the multiplayer map", "key", key, "value", value)
			mpm.multiPlayerMap.Delete(key)
		} else {
			l = append(l, mp)
		}

		return true
	})

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
