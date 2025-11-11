package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/team-vesperis/vesperis-mp/internal/config"
	"github.com/team-vesperis/vesperis-mp/internal/logger"
)

func initRedis(ctx context.Context, l *logger.Logger, c *config.Config) (*redis.Client, error) {
	now := time.Now()
	opt, urlErr := redis.ParseURL(c.GetRedisUrl())
	if urlErr != nil {
		l.Error("redis parsing url error", "options", opt, "error", urlErr)
		return nil, urlErr
	}

	r := redis.NewClient(opt)

	pingErr := r.Ping(ctx).Err()
	if pingErr != nil {
		l.Error("redis ping error", "error", pingErr)
		return nil, pingErr
	}

	l.Debug("initialized redis", "duration", time.Since(now))
	return r, nil
}
