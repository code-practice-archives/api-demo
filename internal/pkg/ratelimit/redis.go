package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisLimiter struct {
	rdb       redis.Cmdable
	limit     int
	window    time.Duration
	keyPrefix string
}

// NewRedis 使用已建立的 Redis 客户端创建 Limiter。
func NewRedis(rdb redis.Cmdable, limit int, window time.Duration) Limiter {
	if limit <= 0 {
		limit = 120
	}
	if window <= 0 {
		window = time.Minute
	}
	return &redisLimiter{
		rdb:       rdb,
		limit:     limit,
		window:    window,
		keyPrefix: "ratelimit",
	}
}

func (l *redisLimiter) Allow(ctx context.Context, key string) (bool, error) {
	k := fmt.Sprintf("%s:%s", l.keyPrefix, key)

	n, err := l.rdb.Incr(ctx, k).Result()
	if err != nil {
		return false, err
	}
	if n == 1 {
		if err := l.rdb.Expire(ctx, k, l.window).Err(); err != nil {
			return false, err
		}
	}
	return n <= int64(l.limit), nil
}
