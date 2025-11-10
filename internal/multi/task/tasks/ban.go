package tasks

import (
	"fmt"
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

func NewBanTask(targetPlayerId uuid.UUID, reason string, permanently bool, expiration time.Time) *BanTask {
	return &BanTask{
		TargetPlayerId: targetPlayerId,
		Reason:         reason,
		Permanently:    permanently,
		Expiration:     expiration,
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
					S:       util.StyleColorCyan,
				},
			},
		})

	} else {
		err = mp.GetBanInfo().TempBan(bt.Reason, bt.Expiration)
		if err != nil {
			return task.NewTaskResponse(false, err.Error())
		}

		duration := time.Until(bt.Expiration)
		durationStr := formatDuration(duration)

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
					S:       util.StyleColorCyan,
				},
				&component.Text{
					Content: "\nExpiration: ",
					S:       util.StyleColorRed,
				},
				&component.Text{
					Content: durationStr,
					S:       util.StyleColorCyan,
				},
			},
		})
	}

	return task.NewTaskResponse(true, "")
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	return fmt.Sprintf("%d days, %d hours, %d minutes and %d seconds", days, hours, minutes, seconds)
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
