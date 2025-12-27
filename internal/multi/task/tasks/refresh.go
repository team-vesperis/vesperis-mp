package tasks

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"go.minekube.com/gate/pkg/util/uuid"
)

type RefreshTask struct {
	TargetProxyId   uuid.UUID `json:"targetProxyId"`
	ResponseChannel string    `json:"responseChannel"`
}

func NewRefreshTask(targetProxyId uuid.UUID) *RefreshTask {
	return &RefreshTask{
		TargetProxyId: targetProxyId,
	}
}

func (rt *RefreshTask) PerformTask(tm *task.TaskManager) *task.TaskResponse {
	duration := tm.GetMultiManager().Refresh()
	return task.NewTaskResponse(true, duration.String())
}

func (rt *RefreshTask) GetTargetProxyId() uuid.UUID {
	return rt.TargetProxyId
}

func (rt *RefreshTask) GetResponseChannel() string {
	return rt.ResponseChannel
}

func (rt *RefreshTask) SetResponseChannel(channel string) {
	rt.ResponseChannel = channel
}

func (rt *RefreshTask) GetTaskType() string {
	return refreshTask
}
