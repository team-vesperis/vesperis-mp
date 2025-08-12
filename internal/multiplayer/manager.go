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

	l  *logger.Logger
	db *database.Database
}

func InitMultiPlayerManager(l *logger.Logger, db *database.Database) *MultiPlayerManager {
	now := time.Now()

	mpm := &MultiPlayerManager{
		multiPlayerMap: sync.Map{},
		l:              l,
		db:             db,
	}

	// start update listener
	mpm.db.CreateListener(multiPlayerUpdateChannel, func(msg *redis.Message) {
		m := msg.Payload
		mpm.l.Info(m)

		s := strings.Split(m, "_")

		id, err := uuid.Parse(s[0])
		if err != nil {
			mpm.l.Error("multiplayer update channel parse uuid error", "parsed uuid", s[0], "error", err)
			return
		}

		key := s[1]

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

		// don't notify so there will be no loop created
		switch key {
		case "p":
			p, ok := val.(string)
			if ok {
				mp.SetProxyId(p, false)
			}
		case "b":
			b, ok := val.(string)
			if ok {
				mp.SetBackendId(b, false)
			}
		case "name":
			name, ok := val.(string)
			if ok {
				mp.SetName(name, false)
			}
		case "permission.role":
			role, ok := val.(string)
			if ok {
				mp.pi.SetRole(role, false)
			}
		case "permission.rank":
			rank, ok := val.(string)
			if ok {
				mp.pi.SetRank(rank, false)
			}
		case "online":
			online, ok := val.(bool)
			if ok {
				mp.SetOnline(online, false)
			}
		case "vanished":
			vanished, ok := val.(bool)
			if ok {
				mp.SetVanished(vanished, false)
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

				mp.SetFriends(mp_list, false)
			}
		}
	})

	// fill map
	_, err := mpm.GetAllMultiPlayersFromDatabase()
	if err != nil {
		mpm.l.Error("filling up multiplayer map error", "error", err)
	}

	mpm.l.Info("initialized multiplayer manager", "duration", time.Since(now))
	return mpm
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

	p, ok := data["p"].(string)
	if ok {
		mp.p = p
	}

	b, ok := data["b"].(string)
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

func (mpm *MultiPlayerManager) GetAllMultiPlayers() ([]*MultiPlayer, error) {
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

	return l, nil
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

func (mpm *MultiPlayerManager) GetAllOnlinePlayers() ([]*MultiPlayer, error) {
	var l []*MultiPlayer

	all, err := mpm.GetAllMultiPlayers()
	if err != nil {
		return nil, err
	}

	for _, mp := range all {
		if mp.IsOnline() {
			l = append(l, mp)
		}
	}

	return l, nil
}
