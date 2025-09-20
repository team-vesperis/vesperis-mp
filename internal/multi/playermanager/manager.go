package playermanager

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
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

	mpm.db.CreateListener(task.TaskChannel, mpm.createTaskListener())

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

func (mpm *MultiPlayerManager) SendTask(targetProxyId uuid.UUID, responseChannel string, t task.Task) *task.TaskResponse {
	d, err := json.Marshal(t)
	if err != nil {
		return task.NewTaskResponse(false, "task confirmation could not marshal task")
	}

	msg, err := mpm.db.SendAndReturn(task.TaskChannel, t.GetResponseChannel(), d, 2*time.Second)
	if err != nil {
		return task.NewTaskResponse(false, err.Error())
	}

	l := strings.Split(msg.Payload, "_")
	if len(l) != 2 {
		return task.NewTaskResponse(false, "task confirmation returned an incorrect length")
	}

	s, err := strconv.ParseBool(l[0])
	if err != nil {
		return task.NewTaskResponse(false, "task confirmation returned not a bool")
	}

	return task.NewTaskResponse(s, l[1])
}

func (mpm *MultiPlayerManager) createTaskListener() func(msg *redis.Message) {
	return func(msg *redis.Message) {
		var t task.Task
		err := json.Unmarshal([]byte(msg.Payload), &t)
		if err != nil {
			return
		}

		if mpm.ownerProxyId == t.GetTargetProxyId() {
			tr := t.PerformTask(mpm)
			m := strconv.FormatBool(tr.IsSuccessful()) + "_" + tr.GetReason()

			err := mpm.db.Publish(t.GetResponseChannel(), m)
			if err != nil {
				return
			}
		}
	}
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

func (mpm *MultiPlayerManager) NewMultiPlayer(p proxy.Player) (*multi.MultiPlayer, error) {
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
func (mpm *MultiPlayerManager) GetMultiPlayer(id uuid.UUID) (*multi.MultiPlayer, error) {
	val, ok := mpm.multiPlayerMap.Load(id)
	if ok {
		mp, ok := val.(*multi.MultiPlayer)
		if ok {
			return mp, nil
		} else {
			mpm.multiPlayerMap.Delete(id)
		}
	}

	return mpm.CreateMultiPlayerFromDatabase(id)
}

func (mpm *MultiPlayerManager) CreateMultiPlayerFromDatabase(id uuid.UUID) (*multi.MultiPlayer, error) {
	data, err := mpm.db.GetPlayerData(id)
	if err != nil {
		return nil, err
	}

	mp := multi.NewMultiPlayerWithData(id, data)

	mpm.multiPlayerMap.Store(id, mp)
	return mp, nil
}

func (mpm *MultiPlayerManager) GetAllMultiPlayers() []*multi.MultiPlayer {
	var l []*multi.MultiPlayer

	mpm.multiPlayerMap.Range(func(key, value any) bool {
		mp, ok := value.(*multi.MultiPlayer)
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

func (mpm *MultiPlayerManager) GetAllMultiPlayersFromDatabase() ([]*multi.MultiPlayer, error) {
	var l []*multi.MultiPlayer

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

func (mpm *MultiPlayerManager) GetAllOnlinePlayers() []*multi.MultiPlayer {
	var l []*multi.MultiPlayer

	all := mpm.GetAllMultiPlayers()

	for _, mp := range all {
		if mp.IsOnline() {
			l = append(l, mp)
		}
	}

	return l
}
