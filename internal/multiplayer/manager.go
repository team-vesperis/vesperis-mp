package multiplayer

import (
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MultiPlayerManager struct {
	multiPlayerMap sync.Map

	ownerProxyId uuid.UUID

	l  *logger.Logger
	db *database.Database
}

func InitManager(l *logger.Logger, db *database.Database, id uuid.UUID) *MultiPlayerManager {
	now := time.Now()

	mpm := &MultiPlayerManager{
		multiPlayerMap: sync.Map{},
		l:              l,
		db:             db,
		ownerProxyId:   id,
	}

	// start update listener
	mpm.db.CreateListener(multiPlayerUpdateChannel, mpm.createUpdateListener())

	// fill map
	_, err := mpm.GetAllMultiPlayersFromDatabase()
	if err != nil {
		mpm.l.Error("filling up multiplayer map error", "error", err)
	}

	mpm.l.Info("initialized multiplayer manager", "duration", time.Since(now))
	return mpm
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

		mpm.l.Info(m)

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

		switch key {
		case "p":
			p, ok := val.(uuid.UUID)
			if ok {
				mp.mu.Lock()
				mp.p = p
				mp.mu.Unlock()
			}
		case "b":
			b, ok := val.(uuid.UUID)
			if ok {
				mp.mu.Lock()
				mp.b = b
				mp.mu.Unlock()
			}
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
				var mp_list []*MultiPlayer
				for _, l := range list {
					id, ok := l.(uuid.UUID)
					if ok {
						mp, err := mpm.GetMultiPlayer(id)
						if err != nil {
							mp_list = append(mp_list, mp)
						}
					}
				}

				mp.mu.Lock()
				mp.friends = mp_list
				mp.mu.Unlock()
			}
		}
	}
}

/*
Gets a multiplayer in two ways:

1. Use the map with all multiplayers.
This method will be used the most since all existing players are in the map located.

2. Create a new multiplayer based on the player data from the database.
*/
func (mpm *MultiPlayerManager) GetMultiPlayer(id uuid.UUID) (*MultiPlayer, error) {
	val, ok := mpm.multiPlayerMap.Load(id)
	if ok {
		mp, ok := val.(*MultiPlayer)
		if ok {
			return mp, nil
		} else {
			mpm.multiPlayerMap.Delete(id)
		}
	}

	return mpm.CreateMultiPlayerFromDatabase(id)
}

func (mpm *MultiPlayerManager) CreateMultiPlayerFromDatabase(id uuid.UUID) (*MultiPlayer, error) {
	data, err := mpm.db.GetPlayerData(id)
	if err != nil {
		return nil, err
	}

	mp := &MultiPlayer{
		id:  id,
		mpm: mpm,
	}

	mp.pi = newPermissionInfo(mp)
	mp.bi = newBanInfo()

	p, ok := data["p"].(uuid.UUID)
	if ok {
		mp.p = p
	}

	b, ok := data["b"].(uuid.UUID)
	if ok {
		mp.b = b
	}

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

	mpm.multiPlayerMap.Store(id, mp)
	return mp, nil
}

func (mpm *MultiPlayerManager) GetAllMultiPlayers() []*MultiPlayer {
	var l []*MultiPlayer

	mpm.multiPlayerMap.Range(func(key, value any) bool {
		mp, ok := value.(*MultiPlayer)
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

func (mpm *MultiPlayerManager) GetAllMultiPlayersFromDatabase() ([]*MultiPlayer, error) {
	var l []*MultiPlayer

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

func (mpm *MultiPlayerManager) GetAllOnlinePlayers() []*MultiPlayer {
	var l []*MultiPlayer

	all := mpm.GetAllMultiPlayers()

	for _, mp := range all {
		if mp.IsOnline() {
			l = append(l, mp)
		}
	}

	return l
}
