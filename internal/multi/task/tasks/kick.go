package tasks

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/util/uuid"
)

type KickTask struct {
	TargetPlayerId  uuid.UUID `json:"targetPlayerId"`
	Reason          string    `json:"reason"`
	TargetProxyId   uuid.UUID `json:"targetProxyId"`
	ResponseChannel string    `json:"responseChannel"`
}

func NewKickTask(targetPlayerId, targetProxyId uuid.UUID, reason string) *KickTask {
	return &KickTask{
		TargetPlayerId: targetPlayerId,
		TargetProxyId:  targetProxyId,
		Reason:         reason,
	}
}

func (kt *KickTask) PerformTask(tm *task.TaskManager) *task.TaskResponse {
	t := tm.GetOwnerGate().Player(kt.TargetPlayerId)
	if t == nil {
		return task.NewTaskResponse(false, "target not found")
	}

	t.Disconnect(&component.Text{
		Content: "You have been kicked for the following reason:",
		S:       util.StyleColorRed,
		Extra: []component.Component{
			&component.Text{
				Content: "\n\n" + kt.Reason,
				S:       util.StyleColorCyan,
			},
		},
	})

	return task.NewTaskResponse(true, "")
}

func (kt *KickTask) GetTargetProxyId() uuid.UUID {
	return kt.TargetProxyId
}

func (kt *KickTask) GetResponseChannel() string {
	return kt.ResponseChannel
}

func (kt *KickTask) SetResponseChannel(ch string) {
	kt.ResponseChannel = ch
}

func (mt *KickTask) GetTaskType() string {
	return kickTask
}
