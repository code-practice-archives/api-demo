package loginjail

import (
	"context"
	"testing"
	"time"
)

func TestRedis_LockAfterMaxRetries(t *testing.T) {
	j, _ := newTestRedisJail(t, 3, time.Minute)
	ctx := context.Background()

	locked, err := j.Locked(ctx, "alice")
	if err != nil || locked {
		t.Fatalf("should not be locked initially: locked=%v err=%v", locked, err)
	}

	for i, wantLocked := range []bool{false, false, true} {
		got, err := j.Fail(ctx, "alice")
		if err != nil {
			t.Fatalf("fail %d: %v", i+1, err)
		}
		if got != wantLocked {
			t.Fatalf("fail %d: locked=%v, want %v", i+1, got, wantLocked)
		}
	}

	locked, err = j.Locked(ctx, "alice")
	if err != nil || !locked {
		t.Fatalf("expected locked: locked=%v err=%v", locked, err)
	}
}

func TestRedis_ResetClearsLock(t *testing.T) {
	j, _ := newTestRedisJail(t, 2, time.Minute)
	ctx := context.Background()

	_, _ = j.Fail(ctx, "bob")
	if err := j.Reset(ctx, "bob"); err != nil {
		t.Fatalf("reset: %v", err)
	}

	locked, err := j.Locked(ctx, "bob")
	if err != nil || locked {
		t.Fatalf("should not be locked after reset: locked=%v err=%v", locked, err)
	}
}

func TestRedis_Expires(t *testing.T) {
	j, mr := newTestRedisJail(t, 1, time.Minute)
	ctx := context.Background()

	locked, err := j.Fail(ctx, "carol")
	if err != nil || !locked {
		t.Fatalf("should lock on first fail: locked=%v err=%v", locked, err)
	}

	mr.FastForward(time.Minute + time.Second)

	locked, err = j.Locked(ctx, "carol")
	if err != nil || locked {
		t.Fatalf("lock should have expired: locked=%v err=%v", locked, err)
	}
}
