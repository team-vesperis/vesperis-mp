package database

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"go.uber.org/zap"
)

func initializeRedis(ctx context.Context, l *zap.SugaredLogger) (*redis.Client, error) {
	opt, urlErr := redis.ParseURL(config.GetRedisUrl())
	if urlErr != nil {
		l.Error("Error parsing url in the Redis Database. - ", urlErr)
		return nil, urlErr
	}

	c := redis.NewClient(opt)

	setError := c.Set(ctx, "ping", "pong", 0).Err()
	if setError != nil {
		l.Error("Error sending value to the Redis Database. - ", setError)
		return nil, setError
	}

	value, getError := c.Get(ctx, "ping").Result()
	if getError != nil {
		l.Error("Error retrieving value from the Redis Database. - ", getError)
		return nil, getError
	}

	if value != "pong" {
		err := errors.New("Incorrect value return with Redis Database.")
		l.Error(err)
		return nil, err
	}

	l.Info("Successfully initialized the Redis Database.")
	return c, nil
}
