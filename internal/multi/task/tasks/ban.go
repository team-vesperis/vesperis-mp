package tasks

import (
	"time"

	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/util/uuid"
)

type BanTask struct {
	TargetPlayerId uuid.UUID `json:"targetPlayerId"`
	Reason         string    `json:"reason"`
	Permanently    bool      `json:"permanently"`
	Expiration     time.Time `json:"expiration"`

	TargetProxyId   uuid.UUID `json:"targetProxyId"`
	ResponseChannel string    `json:"responseChannel"`
}

func NewBanTask(targetPlayerId, targetProxyId uuid.UUID, reason string, permanently bool, expiration time.Time) *BanTask {
	return &BanTask{
		TargetPlayerId: targetPlayerId,
		Reason:         reason,
		Permanently:    permanently,
		Expiration:     expiration,
		TargetProxyId:  targetProxyId,
	}
}

func (bt *BanTask) PerformTask(tm *task.TaskManager) *task.TaskResponse {
	t := tm.GetOwnerGate().Player(bt.TargetPlayerId)
	if t == nil {
		return task.NewTaskResponse(false, ErrStringTargetNotFound)
	}

	mp, err := tm.GetMultiManager().GetMultiPlayer(t.ID())
	if err != nil {

	}

	if bt.Permanently {
		err = mp.GetBanInfo().Ban(bt.Reason)
		if err != nil {
			return task.NewTaskResponse(false, err.Error())
		}

		t.Disconnect(&component.Text{
			Content: "You have been permanently banned.",
			S:       util.StyleColorRed,
			Extra: []component.Component{
				&component.Text{
					Content: "\n\nReason: ",
					S:       util.StyleColorRed,
				},
				&component.Text{
					Content: bt.Reason,
					S:       util.StyleColorLightBlue,
				},
			},
		})

	} else {
		err = mp.GetBanInfo().TempBan(bt.Reason, bt.Expiration)
		if err != nil {
			return task.NewTaskResponse(false, err.Error())
		}

		t.Disconnect(&component.Text{
			Content: "You have been temporarily banned.",
			S:       util.StyleColorRed,
			Extra: []component.Component{
				&component.Text{
					Content: "\n\nReason: ",
					S:       util.StyleColorRed,
				},
				&component.Text{
					Content: bt.Reason,
					S:       util.StyleColorLightBlue,
				},
				&component.Text{
					Content: "\nExpiration: ",
					S:       util.StyleColorRed,
				},
				&component.Text{
					Content: util.FormatTimeUntil(bt.Expiration),
					S:       util.StyleColorLightBlue,
				},
			},
		})
	}

	return task.NewTaskResponse(true, "")
}
func (bt *BanTask) GetTargetProxyId() uuid.UUID {
	return bt.TargetProxyId
}

func (bt *BanTask) GetResponseChannel() string {
	return bt.ResponseChannel
}

func (bt *BanTask) SetResponseChannel(channel string) {
	bt.ResponseChannel = channel
}

func (bt *BanTask) GetTaskType() string {
	return banTask
}
