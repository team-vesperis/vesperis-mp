package manager

import (
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/data"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util/key"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

func (mm *MultiManager) createPlayerUpdateListener() func(msg *redis.Message) {
	return func(msg *redis.Message) {
		m := msg.Payload
		s := strings.Split(m, "_")
		if len(s) != 3 {
			mm.l.Warn("multiplayer update channel received message with incorrect length", "message", m)
			return
		}

		originProxy := s[0]
		// from own proxy, no update needed
		if mm.ownerMP.GetId().String() == originProxy {
			return
		}

		id, err := uuid.Parse(s[1])
		if err != nil {
			mm.l.Error("multiplayer update channel parse uuid error", "parsed uuid", s[1], "error", err)
			return
		}

		k := s[2]

		mp, err := mm.GetMultiPlayer(id)
		if err != nil {
			mm.l.Error("multiplayer update channel get multiplayer error", "playerId", id, "error", err)
			return
		}

		mm.l.Debug("received player update request", "originProxyId", originProxy, "playerId", id, "key", k)

		if k == "new" {
			return
		}

		dataKey, err := key.GetPlayerKey(k)
		if err != nil {
			mm.l.Error("multiplayer update channel get data key error", "playerId", id, "key", k, "error", err)
			return
		}

		mp.Update(dataKey)
	}
}

func (mm *MultiManager) NewMultiPlayer(p proxy.Player) (*multi.Player, error) {
	now := time.Now()
	id := p.ID()

	data := &data.PlayerData{
		Proxy:    uuid.Nil,
		Backend:  uuid.Nil,
		Username: p.Username(),
		Nickname: p.Username(),
		Permission: &data.PermissionData{
			Role: multi.RoleDefault.String(),
			Rank: multi.RankDefault.String(),
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
		Friend: &data.FriendData{
			Friends:               make([]uuid.UUID, 0),
			FriendRequests:        make([]uuid.UUID, 0),
			FriendPendingRequests: make([]uuid.UUID, 0),
		},
		PartyId:          uuid.Nil,
		PartyInvitations: make([]uuid.UUID, 0),
	}

	err := mm.db.SetPlayerData(id, data)
	if err != nil {
		return nil, err
	}

	mp, err := mm.CreateMultiPlayerFromDatabase(id)
	if err != nil {
		return nil, err
	}

	// update every proxies' map
	m := mm.ownerMP.GetId().String() + "_" + id.String() + "_new"
	err = mm.db.Publish(multi.UpdateMultiPlayerChannel, m)
	if err != nil {
		return nil, err
	}

	mm.l.Info("created new multiplayer", "playerId", id, "duration", time.Since(now))
	return mp, nil
}

/*
Gets a multiplayer in two ways:

1. Use the map with all multiplayers.
This method will be used the most since all existing players are in the map located.

2. Create a new multiplayer based on the player data from the database.
*/
func (mm *MultiManager) GetMultiPlayer(id uuid.UUID) (*multi.Player, error) {
	mm.mu.RLock()
	mp, ok := mm.playerMap[id]
	mm.mu.RUnlock()

	if ok {
		return mp, nil
	}

	return mm.CreateMultiPlayerFromDatabase(id)
}

// if player has never joined before, this function will return database.ErrDataNotFound
func (mm *MultiManager) CreateMultiPlayerFromDatabase(id uuid.UUID) (*multi.Player, error) {
	data, err := mm.db.GetPlayerData(id)
	if err != nil {
		return nil, err
	}

	mp := multi.NewPlayer(id, mm.ownerMP.GetId(), mm.l, mm.db, data)

	mm.mu.Lock()
	mm.playerMap[id] = mp
	mm.mu.Unlock()

	return mp, nil
}

func (mm *MultiManager) GetAllMultiPlayers(includeVanished bool) []*multi.Player {
	var l []*multi.Player

	mm.mu.RLock()
	for _, mp := range mm.playerMap {
		if !includeVanished {
			if !mp.IsVanished() {
				l = append(l, mp)
			}
		} else {
			l = append(l, mp)
		}
	}
	mm.mu.RUnlock()

	return l
}

func (mm *MultiManager) GetAllMultiPlayersFromDatabase() ([]*multi.Player, error) {
	var l []*multi.Player

	i, err := mm.db.GetAllPlayerIds()
	if err != nil {
		return nil, err
	}

	for _, id := range i {
		mp, err := mm.GetMultiPlayer(id)
		if err != nil {
			return nil, err
		}

		l = append(l, mp)
	}

	return l, nil
}

func (mm *MultiManager) GetAllOnlinePlayers(includeVanished bool) []*multi.Player {
	var l []*multi.Player

	for _, mp := range mm.GetAllMultiPlayers(includeVanished) {
		if mp.IsOnline() {
			l = append(l, mp)
		}
	}

	return l
}
