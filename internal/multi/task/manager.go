package task

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
	"github.com/team-vesperis/vesperis-mp/internal/multi"
	"github.com/team-vesperis/vesperis-mp/internal/multi/playermanager"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

type TaskManager struct {
	db        *database.Database
	l         *logger.Logger
	ownerId   uuid.UUID
	ownerGate *proxy.Proxy
	mpm       *playermanager.MultiPlayerManager
}

func InitTaskManager(db *database.Database, l *logger.Logger, id uuid.UUID, proxy *proxy.Proxy, mpm *playermanager.MultiPlayerManager) *TaskManager {
	tm := &TaskManager{
		db:        db,
		l:         l,
		ownerId:   id,
		ownerGate: proxy,
		mpm:       mpm,
	}

	tm.db.CreateListener(taskChannel, tm.createTaskListener())

	return tm
}

func (tm *TaskManager) GetDatabase() *database.Database {
	return tm.db
}

func (tm *TaskManager) GetLogger() *logger.Logger {
	return tm.l
}

func (tm *TaskManager) GetOwnerId() uuid.UUID {
	return tm.ownerId
}

func (tm *TaskManager) GetOwnerGate() *proxy.Proxy {
	return tm.ownerGate
}

func (tm *TaskManager) GetMultiPlayerManager() *playermanager.MultiPlayerManager {
	return tm.mpm
}

func (tm *TaskManager) createTaskListener() func(msg *redis.Message) {
	return func(msg *redis.Message) {
		var t Task
		err := json.Unmarshal([]byte(msg.Payload), &t)
		if err != nil {
			tm.l.Warn("")
			return
		}

		if tm.ownerId == t.GetTargetProxyId() {
			tr := t.PerformTask(tm)
			m := strconv.FormatBool(tr.IsSuccessful()) + "_" + tr.GetReason()

			err := tm.db.Publish(t.GetResponseChannel(), m)
			if err != nil {
				return
			}
		}
	}
}

func (tm *TaskManager) SendTask(targetMultiProxy *multi.Proxy, t Task) *TaskResponse {
	d, err := json.Marshal(t)
	if err != nil {
		return NewTaskResponse(false, "task confirmation could not marshal task")
	}

	msg, err := tm.db.SendAndReturn(taskChannel, t.GetResponseChannel(), d, 2*time.Second)
	if err != nil {
		return NewTaskResponse(false, err.Error())
	}

	l := strings.Split(msg.Payload, "_")
	if len(l) != 2 {
		return NewTaskResponse(false, "task confirmation returned an incorrect length")
	}

	s, err := strconv.ParseBool(l[0])
	if err != nil {
		return NewTaskResponse(false, "task confirmation returned not a bool")
	}

	return NewTaskResponse(s, l[1])
}
