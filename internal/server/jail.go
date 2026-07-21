package server

import (
	"fmt"

	"github.com/code-practice-archives/api-demo/internal/pkg/loginjail"
	"github.com/redis/go-redis/v9"
)

func newLoginJail(cfg loginjail.Config, rdb *redis.Client) (loginjail.Jail, error) {
	maxRetries := cfg.MaxRetries
	lockFor := cfg.LockDuration()

	switch cfg.Store {
	case "", loginjail.StoreMemory:
		return loginjail.NewMemory(maxRetries, lockFor), nil
	case loginjail.StoreRedis:
		if rdb == nil {
			return nil, fmt.Errorf("redis client is required when auth.store is redis")
		}
		return loginjail.NewRedis(rdb, maxRetries, lockFor), nil
	default:
		return nil, fmt.Errorf("auth.store must be %q or %q", loginjail.StoreMemory, loginjail.StoreRedis)
	}
}
