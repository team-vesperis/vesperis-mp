package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Database interface {
	Get(key string) (any, error)
	Set(key string, value any) error

	SetPlayerDataField(playerId, field string, value any) error
	GetPlayerDataField(playerId, field string) (any, error)

	Publish(channel string, message any) error
	Subscribe(channel string) *redis.PubSub
	SubscribeWithTimeout(channel string, timeout time.Duration) (*redis.Message, error)

	// Combination of Publish & Subscribe. Publish message in a channel, wait for a return message with a time limit.
	SendAndReturn(channel string, message any, timeout time.Duration) (*redis.Message, error)
	// Create a listener to listen for incoming calls. Is basically the same as the Subscribe function but it handles it for you. The listener can be stopped by using DeleteListener()
	CreateListener(channel string, handler func(msg *redis.Message))
	DeleteListener(channel string) error
	DeleteAllListeners() error

	// Close the database. Closes the connection with Redis and PostgreSQL
	Close()
}

func Init(ctx context.Context, c *viper.Viper, l *zap.SugaredLogger) (Database, error) {
	r, err := initializeRedis(ctx, l)
	if err != nil {
		return nil, err
	}

	p, err := initializePostgres(ctx, l)
	if err != nil {
		return nil, err
	}

	database := new(ctx, r, l, p)
	return database, nil
}
