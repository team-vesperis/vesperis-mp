package multiplayer

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.minekube.com/gate/pkg/util/uuid"
)

type Task interface {
	PerformTask(mpm *MultiPlayerManager) *TaskResponse
	GetTargetProxyId() uuid.UUID
	GetResponseChannel() string
}

const multiPlayerTaskChannel = "task_mp"

type TaskResponse struct {
	s bool
	r string
}

func NewTaskResponse(successful bool, reason string) *TaskResponse {
	return &TaskResponse{
		s: successful,
		r: reason,
	}
}

func (r *TaskResponse) IsSuccessful() bool {
	return r.s
}

func (r *TaskResponse) GetReason() string {
	return r.r
}

func (mpm *MultiPlayerManager) SendTask(targetProxyId uuid.UUID, responseChannel string, t Task) *TaskResponse {
	d, err := json.Marshal(t)
	if err != nil {
		return &TaskResponse{
			s: false,
			r: "task confirmation could not marshal task",
		}
	}

	msg, err := mpm.db.SendAndReturn(multiPlayerTaskChannel, t.GetResponseChannel(), d, 2*time.Second)
	if err != nil {
		return &TaskResponse{
			s: false,
			r: err.Error(),
		}
	}

	l := strings.Split(msg.Payload, "_")
	if len(l) != 2 {
		return &TaskResponse{
			s: false,
			r: "task confirmation returned an incorrect length",
		}
	}

	s, err := strconv.ParseBool(l[0])
	if err != nil {
		return &TaskResponse{
			s: false,
			r: "task confirmation returned not a bool",
		}
	}

	return &TaskResponse{
		s: s,
		r: l[1],
	}
}

func (mpm *MultiPlayerManager) createTaskListener() func(msg *redis.Message) {
	return func(msg *redis.Message) {
		var t Task
		err := json.Unmarshal([]byte(msg.Payload), &t)
		if err != nil {
			return
		}

		if mpm.ownerProxyId == t.GetTargetProxyId() {
			tr := t.PerformTask(mpm)
			m := strconv.FormatBool(tr.s) + "_" + tr.r

			err := mpm.db.Publish(t.GetResponseChannel(), m)
			if err != nil {
				return
			}
		}
	}
}
