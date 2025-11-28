package tasks

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MessageTask struct {
	TargetPlayerId  uuid.UUID `json:"targetPlayerId"`
	Message         string    `json:"message"`
	TargetProxyId   uuid.UUID `json:"targetProxyId"`
	ResponseChannel string    `json:"responseChannel"`
}

func NewMessageTask(targetPlayerId, targetProxyId uuid.UUID, message string) *MessageTask {
	return &MessageTask{
		TargetPlayerId: targetPlayerId,
		Message:        message,
		TargetProxyId:  targetProxyId,
	}
}

func (mt *MessageTask) PerformTask(tm *task.TaskManager) *task.TaskResponse {
	t := tm.GetOwnerGate().Player(mt.TargetPlayerId)
	if t == nil {
		return task.NewTaskResponse(false, ErrStringTargetNotFound)
	}

	t.SendMessage(util.StringToComponent(mt.Message))
	return task.NewTaskResponse(true, "")
}

func (mt *MessageTask) GetTargetProxyId() uuid.UUID {
	return mt.TargetProxyId
}

func (mt *MessageTask) GetResponseChannel() string {
	return mt.ResponseChannel
}

func (mt *MessageTask) SetResponseChannel(channel string) {
	mt.ResponseChannel = channel
}

func (mt *MessageTask) GetTaskType() string {
	return messageTask
}
