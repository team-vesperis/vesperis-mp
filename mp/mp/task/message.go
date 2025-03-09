package task

import (
	"context"

	"github.com/team-vesperis/vesperis-mp/mp/database"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
)

type MessageTask struct {
	TaskType         string `json:"task_type"`
	OriginPlayerName string `json:"origin_player_name"`
	TargetPlayerName string `json:"target_player_name"`
	Message          string `json:"message"`
}

func (t *MessageTask) CreateTask(target_proxy string) error {
	t.TaskType = "MessageTask"
	return send(target_proxy, t)
}

func (t *MessageTask) PerformTask(responseChannel string) {
	targetPlayer := p.PlayerByName(t.TargetPlayerName)
	if targetPlayer == nil {
		t.SendResponse(Player_Not_Found, responseChannel)
	} else {
		targetPlayer.SendMessage(&component.Text{
			Content: "[<- " + t.OriginPlayerName + "]",
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

		t.SendResponse(Successful, responseChannel)
	}
}

func (t *MessageTask) SendResponse(errorString, responseChannel string) {
	database.Publish(context.Background(), responseChannel, errorString)
}
