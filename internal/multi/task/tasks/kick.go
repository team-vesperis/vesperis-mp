package tasks

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/util/uuid"
)

type KickTask struct {
	targetPlayerId  uuid.UUID
	reason          string
	targetProxyId   uuid.UUID
	responseChannel string
}

func NewKickTask(targetPlayerId, targetProxyId uuid.UUID, reason string) *KickTask {
	return &KickTask{
		targetPlayerId: targetPlayerId,
		targetProxyId:  targetProxyId,
		reason:         reason,
	}
}

func (kt *KickTask) PerformTask(tm *task.TaskManager) *task.TaskResponse {
	t := tm.GetOwnerGate().Player(kt.targetPlayerId)
	if t == nil {
		return task.NewTaskResponse(false, "target not found")
	}

	t.Disconnect(&component.Text{
		Content: "You have been kicked for the following reason:",
		S:       util.StyleColorRed,
		Extra: []component.Component{
			&component.Text{
				Content: "\n\n" + kt.reason,
				S:       util.StyleColorCyan,
			},
		},
	})

	return task.NewTaskResponse(true, "")
}

func (kt *KickTask) GetTargetProxyId() uuid.UUID {
	return kt.targetProxyId
}

func (kt *KickTask) GetResponseChannel() string {
	return kt.responseChannel
}

func (kt *KickTask) SetResponseChannel(ch string) {
	kt.responseChannel = ch
}
