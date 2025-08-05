package multiplayer

import (
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
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
		id := s[0]
		key := s[1]

		mp := mpm.GetMultiPlayer(id)
		val, err := mpm.db.GetPlayerDataField(id, key)
		if err != nil {
			mpm.l.Warn("multiplayer update channel get player data field error", "error", err)
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
					id, ok := f.(string)
					if ok {
						friend := mpm.GetMultiPlayer(id)
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

func (mpm *MultiPlayerManager) GetAllMultiPlayers() []*MultiPlayer {
	var l []*MultiPlayer

	i, err := mpm.db.GetAllPlayerIds()
	if err != nil {
		return l
	}

	for _, id := range i {
		mp := mpm.GetMultiPlayer(id)
		l = append(l, mp)
	}

	return l
}

func (mpm *MultiPlayerManager) GetMultiPlayer(id string) *MultiPlayer {
	val, ok := mpm.multiPlayerMap.Load(id)
	if ok {
		mp, ok := val.(*MultiPlayer)
		if ok {
			return mp
		} else {
			mpm.multiPlayerMap.Delete(id)
		}
	}

	data, err := mpm.db.GetPlayerData(id)
	if err != nil {
		return nil
	}

	mp := &MultiPlayer{
		id:  id,
		mpm: mpm,
	}

	if v, ok := data["p"].(string); ok {
		mp.p = v
	}
	if v, ok := data["b"].(string); ok {
		mp.b = v
	}
	if v, ok := data["name"].(string); ok {
		mp.name = v
	}
	if v, ok := data["online"].(bool); ok {
		mp.online = v
	}

	mpm.multiPlayerMap.Store(id, mp)
	return mp
}
