package tasks

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/playermanager"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MessageTask struct {
	originPlayer    string
	targetPlayer    string
	m               string
	targetProxyId   uuid.UUID
	responseChannel string
}

func (mt *MessageTask) PerformTask(mpm *playermanager.MultiPlayerManager) *task.TaskResponse {
	return &task.TaskResponse{}
}

func (mt *MessageTask) GetTargetProxyId() uuid.UUID {
	return mt.targetProxyId
}

func (mt *MessageTask) GetResponseChannel() string {
	return mt.responseChannel
}
