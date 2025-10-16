package tasks

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/common/minecraft/color"
	c "go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MessageTask struct {
	OriginName      string    `json:"originName"`
	TargetPlayerId  uuid.UUID `json:"targetPlayerId"`
	Message         string    `json:"message"`
	TargetProxyId   uuid.UUID `json:"targetProxyId"`
	ResponseChannel string    `json:"responseChannel"`
}

func NewMessageTask(originName string, targetPlayerId, targetProxyId uuid.UUID, message string) *MessageTask {
	return &MessageTask{
		OriginName:     originName,
		TargetPlayerId: targetPlayerId,
		Message:        message,
		TargetProxyId:  targetProxyId,
	}
}

func (mt *MessageTask) PerformTask(tm *task.TaskManager) *task.TaskResponse {
	t := tm.GetOwnerGate().Player(mt.TargetPlayerId)
	if t == nil {
		return task.NewTaskResponse(false, "target not found")
	}

	t.SendMessage(&c.Text{
		Content: "[<- " + mt.OriginName + "]",
		S: c.Style{
			Color: util.ColorCyan,
			HoverEvent: c.ShowText(&c.Text{
				Content: "Click to reply",
				S:       util.StyleMysterious,
			}),
			ClickEvent: c.NewClickEvent(c.SuggestCommandAction, "/message "+mt.OriginName+" "),
		},
		Extra: []c.Component{
			&c.Text{
				Content: ": " + mt.Message,
				S: c.Style{
					Color: color.White,
				},
			},
		},
	})

	return task.NewTaskResponse(true, "")
}

func (mt *MessageTask) GetTargetProxyId() uuid.UUID {
	return mt.TargetProxyId
}

func (mt *MessageTask) GetResponseChannel() string {
	return mt.ResponseChannel
}

func (mt *MessageTask) SetResponseChannel(ch string) {
	mt.ResponseChannel = ch
}

func (mt *MessageTask) GetTaskType() string {
	return messageTask
}
