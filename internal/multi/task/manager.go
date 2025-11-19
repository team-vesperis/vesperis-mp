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
	"github.com/team-vesperis/vesperis-mp/internal/multi/manager"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
)

type TaskManager struct {
	db           *database.Database
	l            *logger.Logger
	ownerGate    *proxy.Proxy
	multiManager *manager.MultiManager
}

func InitTaskManager(db *database.Database, l *logger.Logger, mp *multi.Proxy, proxy *proxy.Proxy, mm *manager.MultiManager) *TaskManager {
	tm := &TaskManager{
		db:           db,
		l:            l,
		ownerGate:    proxy,
		multiManager: mm,
	}

	tm.db.CreateListener(taskChannel, tm.createTaskListener())

	return tm
}

type TaskType struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

var taskRegistry = map[string]func() Task{}

func RegisterTaskType(name string, constructor func() Task) {
	taskRegistry[name] = constructor
}

func (tm *TaskManager) GetDatabase() *database.Database {
	return tm.db
}

func (tm *TaskManager) GetLogger() *logger.Logger {
	return tm.l
}

func (tm *TaskManager) GetOwnerGate() *proxy.Proxy {
	return tm.ownerGate
}

func (tm *TaskManager) GetMultiManager() *manager.MultiManager {
	return tm.multiManager
}

func (tm *TaskManager) createTaskListener() func(msg *redis.Message) {
	return func(msg *redis.Message) {
		var tt TaskType
		err := json.Unmarshal([]byte(msg.Payload), &tt)
		if err != nil {
			tm.l.Warn("task listener unmarshal task type error", "error", err)
			return
		}

		constructor, ok := taskRegistry[tt.Type]
		if !ok {
			tm.l.Warn("task listener unknown task type", "type", tt.Type)
			return
		}

		// unmarshal based on task type
		t := constructor()
		err = json.Unmarshal(tt.Data, t)
		if err != nil {
			tm.l.Warn("task listener unmarshal task error", "error", err)
			return
		}

		if tm.multiManager.GetOwnerMultiProxy().GetId() == t.GetTargetProxyId() {
			tr := t.PerformTask(tm)
			m := strconv.FormatBool(tr.IsSuccessful()) + "_" + tr.GetInfo()

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
// Returns TaskResponse.
// Use .IsSuccessful() to check if everything went accordingly.
// Use .GetInfo() to get information, like error messages.
func (tm *TaskManager) BuildTask(t Task) *TaskResponse {
	if t.GetTargetProxyId() == tm.GetMultiManager().GetOwnerMultiProxy().GetId() {
		return t.PerformTask(tm)
	}

	t.SetResponseChannel("task_response-" + uuid.New().Undashed())

	data, err := json.Marshal(t)
	if err != nil {
		return NewTaskResponse(false, "task confirmation could not marshal data task")
	}

	tt := TaskType{
		Type: t.GetTaskType(),
		Data: data,
	}

	d, err := json.Marshal(tt)
	if err != nil {
		return NewTaskResponse(false, "task confirmation could not marshal task type")
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
