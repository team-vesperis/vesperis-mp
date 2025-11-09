package tasks

import (
	"github.com/team-vesperis/vesperis-mp/internal/multi/task"
)

func Init() {
	task.RegisterTaskType(messageTask, func() task.Task { return &MessageTask{} })
	task.RegisterTaskType(kickTask, func() task.Task { return &KickTask{} })
	task.RegisterTaskType(transferRequestTask, func() task.Task { return &TransferRequestTask{} })
	task.RegisterTaskType(transferTask, func() task.Task { return &TransferTask{} })
}

// task types
const (
	messageTask         = "message"
	kickTask            = "kick"
	transferRequestTask = "transfer_request"
	transferTask        = "transfer"
)

const ErrStringTargetNotFound = "target not found"
