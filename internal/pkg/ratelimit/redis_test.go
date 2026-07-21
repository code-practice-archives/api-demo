package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedisLimiter(t *testing.T, limit int, window time.Duration) (Limiter, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	return NewRedis(rdb, limit, window), mr
}

func TestRedis_Allow(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		requests int
		wantLast bool
	}{
		{name: "未超限放行", limit: 3, requests: 3, wantLast: true},
		{name: "超限拒绝", limit: 3, requests: 4, wantLast: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, _ := newTestRedisLimiter(t, tt.limit, time.Minute)
			ctx := context.Background()

			var got bool
			var err error
			for i := 0; i < tt.requests; i++ {
				got, err = l.Allow(ctx, "login:1.2.3.4")
				if err != nil {
					t.Fatalf("Allow #%d: %v", i+1, err)
				}
			}
			if got != tt.wantLast {
				t.Fatalf("last Allow = %v, want %v", got, tt.wantLast)
			}
		})
	}
}

func TestRedis_WindowExpires(t *testing.T) {
	l, mr := newTestRedisLimiter(t, 1, time.Minute)
	ctx := context.Background()

	ok, err := l.Allow(ctx, "ip")
	if err != nil || !ok {
		t.Fatalf("first: allowed=%v err=%v", ok, err)
	}
	ok, err = l.Allow(ctx, "ip")
	if err != nil || ok {
		t.Fatalf("second should deny: allowed=%v err=%v", ok, err)
	}

	mr.FastForward(time.Minute + time.Second)

	ok, err = l.Allow(ctx, "ip")
	if err != nil || !ok {
		t.Fatalf("after expire should allow: allowed=%v err=%v", ok, err)
	}
}
