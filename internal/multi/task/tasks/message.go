package tasks

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MessageTask struct {
	originName      string
	targetPlayerId  uuid.UUID
	message         string
	targetProxyId   uuid.UUID
	responseChannel string
}

func (mt *MessageTask) PerformTask(tm *task.TaskManager) *task.TaskResponse {
	// checks if the targetPlayerName is on this proxy. only happens if player just went offline
	t := tm.GetOwnerGate().Player(mt.targetPlayerId)
	if t == nil {
		return task.NewTaskResponse(false, "target not found")
	}

	t.SendMessage(&c.Text{
		Content: "[<- " + mt.originName + "]",
		S:       util.StyleColorCyan,
		Extra: []c.Component{
			&c.Text{
				Content: ": " + mt.message,
				S: c.Style{
					Color: color.White,
				},
			},
		},
	})

	return task.NewTaskResponse(true, "")
}

func (mt *MessageTask) GetTargetProxyId() uuid.UUID {
	return mt.targetProxyId
}

func (mt *MessageTask) GetResponseChannel() string {
	return mt.responseChannel
}
