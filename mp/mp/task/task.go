package task

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/mp/database"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
	"go.uber.org/zap"
)

var (
	p          *proxy.Proxy
	logger     *zap.SugaredLogger
	proxy_name string
)

func InitializeTask(proxy *proxy.Proxy, log *zap.SugaredLogger, pn string) {
	p = proxy
	logger = log
	proxy_name = pn

	startTaskPerformListener()

	logger.Info("Successfully initialized task.")
}

type Task interface {
	CreateTask(targetProxy string) error
	PerformTask(responseChannel string)
	SendResponse(errorString, responseChannel string)
}

// error returns
var (
	ErrPlayerNotFound = errors.New("player not found")
	ErrSuccessful     = errors.New("successful")
)

func send(targetProxy string, task Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		logger.Warn("could not marshal task before sending", err)
		return err
	}

	var taskMap map[string]any
	err = json.Unmarshal(data, &taskMap)
	if err != nil {
		logger.Warn("could not unmarshal task data", err)
		return err
	}

	responseChannel := "task_response_" + uuid.New().String()

	taskMap["target_proxy"] = targetProxy
	taskMap["response_channel"] = responseChannel
	data, err = json.Marshal(taskMap)
	if err != nil {
		logger.Warn("could not marshal task data with target_proxy and response_channel", err)
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
	case ErrPlayerNotFound.Error():
		return ErrPlayerNotFound

	case ErrSuccessful.Error():
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

		targetProxy, ok := taskMap["target_proxy"].(string)
		if !ok || targetProxy != proxy_name {
			// Task is not for this proxy, ignore it
			return
		}

		var task Task
		switch taskType {
		case "MessageTask":
			task = &MessageTask{}
		case "BanTask":
			task = &BanTask{}
		default:
			logger.Warn("Unknown task type:", taskType)
			return
		}

		err = json.Unmarshal([]byte(msg.Payload), task)
		if err != nil {
			logger.Warn("Error unmarshalling task:", err)
			return
		}

		responseChannel, ok := taskMap["response_channel"].(string)
		if !ok {
			logger.Warn("Invalid response channel")
			return
		}
		task.PerformTask(responseChannel)
	})
}
