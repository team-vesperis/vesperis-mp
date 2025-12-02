package tasks

import (
	"context"

	"github.com/team-vesperis/vesperis-mp/internal/database"
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
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
		if err == database.ErrDataNotFound {
			return task.NewTaskResponse(false, ErrStringProxyNotFound)
		}

		return task.NewTaskResponse(false, err.Error())
	}

	tr := tm.BuildTask(NewTransferRequestTask(mp.GetId(), tt.TransferBackendId))
	if !tr.IsSuccessful() {
		return tr
	}

	// tr.GetInfo() will be one of four things:
	// 0, given backend is not found
	// 1, given backend is found but not responding
	// 2, given backend is found and responding
	// 3, no backend is specified

	if tr.GetInfo() == "0" {
		return task.NewTaskResponse(false, ErrStringBackendNotFound)
	}

	if tr.GetInfo() == "1" {
		tm.GetLogger().Warn("transfer backend found but not responding error", "playerId", t.ID(), "targetBackendId", tt.TransferBackendId)
		return task.NewTaskResponse(false, util.ErrStringBackendNotResponding)
	}

	if tr.GetInfo() == "2" {
		// target is already on this proxy and can be send to backend internally
		if tt.TargetProxyId == tt.TransferProxyId {
			mb, err := tm.GetMultiManager().GetMultiBackend(tt.TransferBackendId)
			if err == nil {
				_, err := t.CreateConnectionRequest(tm.GetOwnerGate().Server(mb.GetName())).Connect(context.Background())
				if err == nil {
					tm.GetLogger().Info("player internal transfer successful", "playerId", t.ID(), "backendId", mb.GetId())
					return task.NewTaskResponse(true, "")
				} else {
					tm.GetLogger().Warn("transfer create connection request error", "playerId", t.ID(), "targetBackendId", mb.GetId(), "error", err)
				}
			} else {
				tm.GetLogger().Warn("transfer get multibackend error", "playerId", t.ID(), "targetBackendId", mb.GetId(), "error", err)
			}
		}

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
