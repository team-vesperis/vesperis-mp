package database

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/config"
)

var client *redis.Client

func initializeRedis() error {
	opt, urlError := redis.ParseURL(config.GetRedisUrl())
	if urlError != nil {
		logger.Error("Error parsing url in the Redis Database. - ", urlError)
		return urlError
	}

	client = redis.NewClient(opt)

	setError := client.Set(context.Background(), "ping", "pong", 0).Err()
	if setError != nil {
		logger.Error("Error sending value to the Redis Database. - ", setError)
		return setError
	}

	value, getError := client.Get(context.Background(), "ping").Result()
	if getError != nil {
		logger.Error("Error retrieving value from the Redis Database. - ", getError)
		return getError
	}

	if value != "pong" {
		err := errors.New("Incorrect value return with Redis Database.")
		logger.Error(err)
		return err
	}

	logger.Info("Successfully initialized the Redis Database.")
	return nil
}

func GetRedisClient() *redis.Client {
	if client == nil {
		logger.Error("Redis client not found.")
	}
	return client
}

func closeRedis() {
	if client != nil {
		client.Close()
	}
}
