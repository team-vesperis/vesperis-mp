package multiplayer

import (
	"strings"
	"sync"

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
			mpm.l.Error("multiplayer update channel parse uuid error", "error", err)
			return
		}

		key := s[1]

		mp, _ := mpm.GetMultiPlayer(id)
		val, err := mpm.db.GetPlayerDataField(id, key)
		if err != nil {
			mpm.l.Error("multiplayer update channel get player data field error", "error", err)
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
				mp.SetRole(role, false)
			}
		case "permission.rank":
			rank, ok := val.(string)
			if ok {
				mp.SetRank(rank, false)
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
			// database values that are saved as []string will return as []any
			list, ok := val.([]any)
			if ok {
				var friends []*MultiPlayer
				for _, f := range list {
					id, ok := f.(uuid.UUID)
					if ok {
						friend, _ := mpm.GetMultiPlayer(id)
						if friend != nil {
							friends = append(friends, friend)
						}
					}
				}

				mp.SetFriends(friends, false)
			}
		}
	})

	// fill map
	mpm.GetAllMultiPlayers()

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

	role, ok := data["role"].(string)
	if ok {
		mp.role = role
	}

	rank, ok := data["rank"].(string)
	if ok {
		mp.rank = rank
	}

	online, ok := data["online"].(bool)
	if ok {
		mp.online = online
	}

	mpm.multiPlayerMap.Store(id, mp)
	return mp, nil
}

func (mpm *MultiPlayerManager) GetAllMultiPlayers() ([]*MultiPlayer, error) {
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
