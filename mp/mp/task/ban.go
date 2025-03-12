package task

import (
	"context"

	"github.com/team-vesperis/vesperis-mp/mp/ban"
	"github.com/team-vesperis/vesperis-mp/mp/database"
)

type BanTask struct {
	TaskType         string `json:"task_type"`
	TargetPlayerName string `json:"target_player_name"`
	Reason           string `json:"message"`
}

func (t *BanTask) CreateTask(target_proxy string) error {
	t.TaskType = "BanTask"
	return send(target_proxy, t)
}

func (t *BanTask) PerformTask(responseChannel string) {
	targetPlayer := p.PlayerByName(t.TargetPlayerName)
	if targetPlayer == nil {
		t.SendResponse(Player_Not_Found, responseChannel)
	} else {
		ban.BanPlayer(targetPlayer, t.Reason)
		t.SendResponse(Successful, responseChannel)
	}
}

func (t *BanTask) SendResponse(errorString, responseChannel string) {
	database.Publish(context.Background(), responseChannel, errorString)
}
