package tasks

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/util/uuid"
)

type FriendRequestTask struct {
	TargetPlayerId   uuid.UUID `json:"targetPlayerId"`
	OriginPlayerName string    `json:"originPlayerId"`

	TargetProxyId   uuid.UUID `json:"targetProxyId"`
	ResponseChannel string    `json:"responseChannel"`
}

func NewFriendRequestTask(targetPlayerId, targetProxyId uuid.UUID, originPlayerName string) *FriendRequestTask {
	return &FriendRequestTask{
		TargetPlayerId:   targetPlayerId,
		TargetProxyId:    targetProxyId,
		OriginPlayerName: originPlayerName,
	}
}

func (frt *FriendRequestTask) PerformTask(tm *task.TaskManager) *task.TaskResponse {
	t := tm.GetOwnerGate().Player(frt.TargetPlayerId)
	if t == nil {
		return task.NewTaskResponse(false, ErrStringTargetNotFound)
	}

	t.SendMessage(&component.Text{
		Content: "Received friend request from ",
		S:       util.StyleColorLightGreen,
		Extra: []component.Component{
			&component.Text{
				Content: frt.OriginPlayerName,
				S:       util.StyleColorCyan,
			},
			&component.Text{
				Content: ". ",
				S:       util.StyleColorLightGreen,
			},
			&component.Text{
				Content: "[ACCEPT]",
				S: component.Style{
					Color: util.ColorGreen,
					Bold:  component.True,
					HoverEvent: component.ShowText(&component.Text{
						Content: "Click to accept friend request.",
						S:       util.StyleColorGray,
					}),
					ClickEvent: component.SuggestCommand("/friends accept " + frt.OriginPlayerName),
				},
			},
			&component.Text{
				Content: " - ",
				S:       util.StyleColorGray,
			},
			&component.Text{
				Content: "[DECLINE]",
				S: component.Style{
					Color: util.ColorRed,
					Bold:  component.True,
					HoverEvent: component.ShowText(&component.Text{
						Content: "Click to decline friend request.",
						S:       util.StyleColorGray,
					}),
					ClickEvent: component.SuggestCommand("/friends decline " + frt.OriginPlayerName),
				},
			},
		},
	})

	return task.NewTaskResponse(true, "")
}

func (frt *FriendRequestTask) GetTargetProxyId() uuid.UUID {
	return frt.TargetProxyId
}

func (frt *FriendRequestTask) GetResponseChannel() string {
	return frt.ResponseChannel
}

func (frt *FriendRequestTask) SetResponseChannel(channel string) {
	frt.ResponseChannel = channel
}

func (frt *FriendRequestTask) GetTaskType() string {
	return friendRequest
}
