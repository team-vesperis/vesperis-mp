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
	r string
}

func NewTaskResponse(successful bool, reason string) *TaskResponse {
	return &TaskResponse{
		s: successful,
		r: reason,
	}
}

func (tr *TaskResponse) IsSuccessful() bool {
	return tr.s
}

func (tr *TaskResponse) GetReason() string {
	return tr.r
}
