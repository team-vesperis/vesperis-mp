package tasks

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"go.minekube.com/common/minecraft/key"
	"go.minekube.com/gate/pkg/edition/java/cookie"
	"go.minekube.com/gate/pkg/util/uuid"
)

type TransferTask struct {
	TargetPlayerId uuid.UUID `json:"targetPlayerId"`
	TargetProxyId  uuid.UUID `json:"targetProxyId"`

	TransferProxyId   uuid.UUID `json:"transferProxyId"`
	TransferBackendId uuid.UUID `json:"transferBackendId"`

	ResponseChannel string `json:"responseChannel"`
}

var TransferKey = key.New("vesperis", "transfer")

func NewTransferTask(targetPlayerId, targetProxyId, transferProxyId, transferBackendId uuid.UUID) *TransferTask {
	return &TransferTask{
		TargetPlayerId:    targetPlayerId,
		TargetProxyId:     targetProxyId,
		TransferProxyId:   transferProxyId,
		TransferBackendId: transferBackendId,
	}
}

func (tt *TransferTask) PerformTask(tm *task.TaskManager) *task.TaskResponse {
	t := tm.GetOwnerGate().Player(tt.TargetPlayerId)
	if t == nil {
		return task.NewTaskResponse(false, ErrStringTargetNotFound)
	}

	mp, err := tm.GetMultiManager().GetMultiProxy(tt.TransferProxyId)
	if err != nil {
		return task.NewTaskResponse(false, err.Error())
	}

	tr := tm.BuildTask(NewTransferRequestTask(mp.GetId(), tt.TransferBackendId))
	if !tr.IsSuccessful() {
		return tr
	}

	// tr.GetInfo() will be one of four things:
	// 0, given server is not available
	// 1, given server is found but not responding
	// 2, given server is available
	// 3, none server is specified

	if tr.GetInfo() == "0" {
		tm.GetLogger().Warn("transfer specified server not found error", "playerId", t.ID(), "targetBackendId", tt.TransferBackendId)
		return task.NewTaskResponse(false, "specified server was not found")
	}

	if tr.GetInfo() == "1" {
		tm.GetLogger().Warn("transfer specified server found but not responding error", "playerId", t.ID(), "targetBackendId", tt.TransferBackendId)
		return task.NewTaskResponse(false, "specified server was found but not responding")
	}

	if tr.GetInfo() == "2" {
		c := &cookie.Cookie{
			Key:     TransferKey,
			Payload: []byte(tt.TransferBackendId.String()),
		}

		err := cookie.Store(t, c)
		if err != nil {
			tm.GetLogger().Warn("transfer manager could not store cookie on player", "playerId", t.ID(), "cookie", c)
			return task.NewTaskResponse(false, err.Error())
		}
	}

	err = t.TransferToHost(mp.GetAddress())
	if err != nil {
		return task.NewTaskResponse(false, err.Error())
	}

	tm.GetLogger().Info("player transfer successful", "playerId", t.ID(), "proxyId", mp.GetId())

	return task.NewTaskResponse(true, "")
}

func (tt *TransferTask) GetTargetProxyId() uuid.UUID {
	return tt.TargetProxyId
}

func (tt *TransferTask) GetResponseChannel() string {
	return tt.ResponseChannel
}

func (tt *TransferTask) SetResponseChannel(channel string) {
	tt.ResponseChannel = channel
}

func (tt *TransferTask) GetTaskType() string {
	return transferTask
}
