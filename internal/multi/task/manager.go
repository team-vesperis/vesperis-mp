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
	ownerMP   *multi.Proxy
	ownerGate *proxy.Proxy
	mpm       *playermanager.MultiPlayerManager
}

func InitTaskManager(db *database.Database, l *logger.Logger, mp *multi.Proxy, proxy *proxy.Proxy, mpm *playermanager.MultiPlayerManager) *TaskManager {
	tm := &TaskManager{
		db:        db,
		l:         l,
		ownerMP:   mp,
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
	return tm.ownerMP.GetId()
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

		t.SetResponseChannel("task_response-" + uuid.New().Undashed())

		if tm.GetOwnerId() == t.GetTargetProxyId() {
			tr := t.PerformTask(tm)
			m := strconv.FormatBool(tr.IsSuccessful()) + "_" + tr.GetReason()

			err := tm.db.Publish(t.GetResponseChannel(), m)
			if err != nil {
				return
			}
		}
	}
}

// BuildTask handles conversation between multiple proxies.
//
// It will first check if it can be handled on this proxy. If so, no need for the database.
// Otherwise it will send out a message using Redis PubSub. The other proxies will check if its a task that they need to handle.
// If they do need to handle it, the proxy will perform the task and send feedback back to the original proxy.
//
// Returns TaskResponse. Use .IsSuccessful() to check if everything went accordingly. If not, use .GetReason() to find out what happened.
func (tm *TaskManager) BuildTask(targetMultiProxy *multi.Proxy, t Task) *TaskResponse {
	if t.GetTargetProxyId() == tm.GetOwnerId() {
		return t.PerformTask(tm)
	}

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
