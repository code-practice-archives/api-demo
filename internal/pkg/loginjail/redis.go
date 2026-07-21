package loginjail

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisJail struct {
	rdb        redis.Cmdable
	maxRetries int
	lockFor    time.Duration
	keyPrefix  string
}

// NewRedis 使用已建立的 Redis 客户端创建 Jail。
func NewRedis(rdb redis.Cmdable, maxRetries int, lockFor time.Duration) Jail {
	if maxRetries <= 0 {
		maxRetries = 5
	}
	if lockFor <= 0 {
		lockFor = 15 * time.Minute
	}
	return &redisJail{
		rdb:        rdb,
		maxRetries: maxRetries,
		lockFor:    lockFor,
		keyPrefix:  "loginjail",
	}
}

func (j *redisJail) failKey(username string) string {
	return fmt.Sprintf("%s:fail:%s", j.keyPrefix, username)
}

func (j *redisJail) lockKey(username string) string {
	return fmt.Sprintf("%s:lock:%s", j.keyPrefix, username)
}

func (j *redisJail) Locked(ctx context.Context, username string) (bool, error) {
	n, err := j.rdb.Exists(ctx, j.lockKey(username)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (j *redisJail) Fail(ctx context.Context, username string) (bool, error) {
	lockKey := j.lockKey(username)
	failKey := j.failKey(username)

	locked, err := j.rdb.Exists(ctx, lockKey).Result()
	if err != nil {
		return false, err
	}
	if locked > 0 {
		return true, nil
	}

	count, err := j.rdb.Incr(ctx, failKey).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		if err := j.rdb.Expire(ctx, failKey, j.lockFor).Err(); err != nil {
			return false, err
		}
	}

	if count < int64(j.maxRetries) {
		return false, nil
	}

	pipe := j.rdb.TxPipeline()
	pipe.Set(ctx, lockKey, "1", j.lockFor)
	pipe.Del(ctx, failKey)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (j *redisJail) Reset(ctx context.Context, username string) error {
	return j.rdb.Del(ctx, j.failKey(username), j.lockKey(username)).Err()
}
