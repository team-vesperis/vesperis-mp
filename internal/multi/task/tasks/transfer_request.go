package tasks

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
	"github.com/team-vesperis/vesperis-mp/internal/multi/util"
	"go.minekube.com/gate/pkg/util/uuid"
)

type TransferRequestTask struct {
	TargetProxyId   uuid.UUID `json:"targetProxyId"`
	TargetBackendId uuid.UUID `json:"targetBackendId"`
	ResponseChannel string    `json:"responseChannel"`
}

func NewTransferRequestTask(targetProxyId, targetBackendId uuid.UUID) *TransferRequestTask {
	return &TransferRequestTask{
		TargetProxyId:   targetProxyId,
		TargetBackendId: targetBackendId,
	}
}

func (trt *TransferRequestTask) PerformTask(tm *task.TaskManager) *task.TaskResponse {
	if trt.TargetBackendId == uuid.Nil {
		return task.NewTaskResponse(true, "3")
	}

	b, err := tm.GetMultiManager().GetMultiBackend(trt.TargetBackendId)
	if err != nil {
		return task.NewTaskResponse(true, "0")
	}

	if !util.IsBackendResponding(b.GetAddress()) {
		return task.NewTaskResponse(true, "1")
	}

	return task.NewTaskResponse(true, "2")
}

func (trt *TransferRequestTask) GetTargetProxyId() uuid.UUID {
	return trt.TargetProxyId
}

func (trt *TransferRequestTask) GetResponseChannel() string {
	return trt.ResponseChannel
}

func (trt *TransferRequestTask) SetResponseChannel(channel string) {
	trt.ResponseChannel = channel
}

func (trt *TransferRequestTask) GetTaskType() string {
	return transferRequestTask
}
