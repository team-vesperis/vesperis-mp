package tasks

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/gate/pkg/util/uuid"
)

type FriendResponseTask struct {
	TargetPlayerId   uuid.UUID `json:"targetPlayerId"`
	OriginPlayerName string    `json:"originPlayerId"`

	Successful bool `json:"successful"`

	TargetProxyId   uuid.UUID `json:"targetProxyId"`
	ResponseChannel string    `json:"responseChannel"`
}

func NewFriendResponseTask(targetPlayerId, targetProxyId uuid.UUID, originPlayerName string, successful bool) *FriendResponseTask {
	return &FriendResponseTask{
		TargetPlayerId:   targetPlayerId,
		TargetProxyId:    targetProxyId,
		OriginPlayerName: originPlayerName,
		Successful:       successful,
	}
}

func (frt *FriendResponseTask) PerformTask(tm *task.TaskManager) *task.TaskResponse {
	t := tm.GetOwnerGate().Player(frt.TargetPlayerId)
	if t == nil {
		return task.NewTaskResponse(false, ErrStringTargetNotFound)
	}

	if frt.Successful {
		t.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorCyan, util.ColorLightGreen), frt.OriginPlayerName, " has accepted your friend request!"))
	} else {
		t.SendMessage(util.TextAlternatingColors(util.ColorList(util.ColorCyan, util.ColorOrange), frt.OriginPlayerName, " has declined your friend request."))
	}

	return task.NewTaskResponse(true, "")
}

func (frt *FriendResponseTask) GetTargetProxyId() uuid.UUID {
	return frt.TargetProxyId
}

func (frt *FriendResponseTask) GetResponseChannel() string {
	return frt.ResponseChannel
}

func (frt *FriendResponseTask) SetResponseChannel(channel string) {
	frt.ResponseChannel = channel
}

func (frt *FriendResponseTask) GetTaskType() string {
	return friendResponse
}
