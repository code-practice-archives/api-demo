package loginjail

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// NewTestJail 创建基于 miniredis 的 Jail，供单测使用。
func NewTestJail(t testing.TB, maxRetries int, lockFor time.Duration) Jail {
	t.Helper()
	j, _ := newTestRedisJail(t, maxRetries, lockFor)
	return j
}

func newTestRedisJail(t testing.TB, maxRetries int, lockFor time.Duration) (Jail, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	return NewRedis(rdb, maxRetries, lockFor), mr
}
