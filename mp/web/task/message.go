package task

import (
	"context"

	"github.com/team-vesperis/vesperis-mp/mp/database"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/util/uuid"
)

type MessageTask struct {
	TaskType         string `json:"task_type"`
	OriginPlayerName string `json:"origin_player_name"`
	TargetPlayerName string `json:"target_player_name"`
	Message          string `json:"message"`
	ResponseChannel  string `json:"response_channel"`
}

func (t *MessageTask) CreateTask() error {
	t.TaskType = "MessageTask"
	responseChannel := uuid.New().String() + "__message_task"
	t.ResponseChannel = responseChannel
	return send(t, responseChannel)
}

func (t *MessageTask) PerformTask() {
	targetPlayer := p.PlayerByName(t.TargetPlayerName)
	if targetPlayer == nil {
		t.SendResponse(playerNotFound)
	} else {
		targetPlayer.SendMessage(&component.Text{
			Content: "[" + t.OriginPlayerName + " ->]",
			S: component.Style{
				Color: color.Aqua,
			},
			Extra: []component.Component{
				&component.Text{
					Content: ": " + t.Message,
					S: component.Style{
						Color: color.White,
					},
				},
			},
		})

		t.SendResponse(successful)
	}
}

func (t *MessageTask) SendResponse(errorString string) {
	database.Publish(context.Background(), t.ResponseChannel, errorString)
}
