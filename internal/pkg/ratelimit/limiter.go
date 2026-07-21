package ratelimit

import "context"

// Limiter 固定窗口计数限流（生产使用 Redis；单测可用 Memory）。
type Limiter interface {
	// Allow 在窗口内对 key 计数；未超限返回 true。
	Allow(ctx context.Context, key string) (allowed bool, err error)
}

// Noop 始终放行，用于关闭限流。
type Noop struct{}

func (Noop) Allow(context.Context, string) (bool, error) { return true, nil }
