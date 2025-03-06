package task

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/mp/database"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.uber.org/zap"
)

var (
	p             *proxy.Proxy
	logger        *zap.SugaredLogger
	proxy_name    string
	performPubSub *redis.PubSub
)

func InitializeTask(proxy *proxy.Proxy, log *zap.SugaredLogger) {
	p = proxy
	logger = log

	startTaskPerformListener()

	logger.Info("Successfully initialized sync.")
}

type Task interface {
	CreateTask() error
	PerformTask()
	SendResponse(errorString string)
}

// error returns
var (
	playerNotFound = "playerNotFound"
	successful     = "successful"
)

func send(task Task, responseChannel string) error {
	data, err := json.Marshal(task)
	if err != nil {
		logger.Warn("could not marshal task before sending", err)
		return err
	}

	msg, err := database.SendAndReturn(context.Background(), "task_listener", responseChannel, data, 2*time.Second)
	if err != nil {
		if err == context.DeadlineExceeded {
			return errors.New("timeout waiting for task confirmation")
		}
		return err
	}

	switch msg.Payload {
	case playerNotFound:
		return errors.New(playerNotFound)

	case successful:
		return nil

	default:
		return errors.New("unknown task confirmation")
	}
}

func startTaskPerformListener() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database.StartListener(ctx, "task_listener", func(msg *redis.Message) {
		var taskMap map[string]interface{}
		err := json.Unmarshal([]byte(msg.Payload), &taskMap)
		if err != nil {
			logger.Warn("Error unmarshalling task:", err)
			return
		}

		taskType, ok := taskMap["task_type"].(string)
		if !ok {
			logger.Warn("Invalid task type")
			return
		}

		var task Task
		switch taskType {
		case "MessageTask":
			task = &MessageTask{}
		default:
			logger.Warn("Unknown task type:", taskType)
			return
		}

		err = json.Unmarshal([]byte(msg.Payload), task)
		if err != nil {
			logger.Warn("Error unmarshalling task:", err)
			return
		}

		task.PerformTask()
	})
}
