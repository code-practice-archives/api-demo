package loginjail

import (
	"context"
	"sync"
	"time"
)

// Memory 进程内实现，仅适合单机或测试。
type Memory struct {
	mu         sync.Mutex
	entries    map[string]*entry
	maxRetries int
	lockFor    time.Duration
	now        func() time.Time
}

type entry struct {
	fails       int
	lockedUntil time.Time
}

// NewMemory 创建内存 Jail。maxRetries <= 0 时默认 5；lockFor <= 0 时默认 15 分钟。
func NewMemory(maxRetries int, lockFor time.Duration) *Memory {
	if maxRetries <= 0 {
		maxRetries = 5
	}
	if lockFor <= 0 {
		lockFor = 15 * time.Minute
	}
	return &Memory{
		entries:    make(map[string]*entry),
		maxRetries: maxRetries,
		lockFor:    lockFor,
		now:        time.Now,
	}
}

func (j *Memory) Locked(_ context.Context, username string) (bool, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	e, ok := j.entries[username]
	if !ok {
		return false, nil
	}
	if e.lockedUntil.IsZero() {
		return false, nil
	}
	if j.now().Before(e.lockedUntil) {
		return true, nil
	}
	delete(j.entries, username)
	return false, nil
}

func (j *Memory) Fail(_ context.Context, username string) (bool, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	e, ok := j.entries[username]
	if !ok {
		e = &entry{}
		j.entries[username] = e
	}

	if !e.lockedUntil.IsZero() && j.now().Before(e.lockedUntil) {
		return true, nil
	}

	e.fails++
	if e.fails >= j.maxRetries {
		e.lockedUntil = j.now().Add(j.lockFor)
		e.fails = 0
		return true, nil
	}
	return false, nil
}

func (j *Memory) Reset(_ context.Context, username string) error {
	j.mu.Lock()
	defer j.mu.Unlock()
	delete(j.entries, username)
	return nil
}
