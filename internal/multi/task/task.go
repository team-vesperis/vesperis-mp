package task

import (
	"go.minekube.com/gate/pkg/util/uuid"
)

type Task interface {
	PerformTask(tm *TaskManager) *TaskResponse
	GetTargetProxyId() uuid.UUID
	GetResponseChannel() string
	SetResponseChannel(ch string)
	GetTaskType() string
}

const taskChannel = "task_mp"

type TaskResponse struct {
	s bool
	i string
}

func NewTaskResponse(successful bool, info string) *TaskResponse {
	return &TaskResponse{
		s: successful,
		i: info,
	}
}

func (tr *TaskResponse) IsSuccessful() bool {
	return tr.s
}

func (tr *TaskResponse) GetInfo() string {
	return tr.i
}
