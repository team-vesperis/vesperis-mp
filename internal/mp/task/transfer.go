package task

import (
	"context"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/mp"
	"github.com/team-vesperis/vesperis-mp/internal/transfer"
)

type TransferTask struct {
	TaskType         string `json:"task_type"`
	TargetPlayerName string `json:"target_player_name"`
	TargetProxyName  string `json:"target_proxy_name"`
	TargetServerName string `json:"target_server_name"` // not always needed
}

func (t *TransferTask) CreateTask(target_proxy string) error {
	t.TaskType = "TransferTask"
	return send(target_proxy, t)
}

func (t *TransferTask) PerformTask(responseChannel string) {
	targetPlayer := p.PlayerByName(t.TargetPlayerName)
	if targetPlayer == nil {
		t.SendResponse(mp.ErrPlayerNotFound.Error(), responseChannel)
	} else {
		err := transfer.TransferPlayerToServerOnOtherProxy(targetPlayer, t.TargetProxyName, t.TargetServerName)
		if err == nil {
			t.SendResponse(mp.ErrSuccessful.Error(), responseChannel)
		} else {
			t.SendResponse(err.Error(), responseChannel)
		}
	}

}

func (t *TransferTask) SendResponse(errorString, responseChannel string) {
	database.Publish(context.Background(), responseChannel, errorString)
}
