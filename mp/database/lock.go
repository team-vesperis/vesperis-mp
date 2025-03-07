package database

import (
	"context"
	"time"
)

func AcquireLockWithTimeout(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	result, err := client.SetNX(ctx, key, "locked", expiration).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

func AcquireLock(ctx context.Context, key string) (bool, error) {
	return AcquireLockWithTimeout(ctx, key, 0)
}

func ReleaseLock(ctx context.Context, key string) error {
	return client.Del(ctx, key).Err()
}
