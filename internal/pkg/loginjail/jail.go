package loginjail

import "context"

// Jail 登录失败锁定存储（生产使用 Redis；单测可用 miniredis）。
type Jail interface {
	Locked(ctx context.Context, username string) (bool, error)
	Fail(ctx context.Context, username string) (locked bool, err error)
	Reset(ctx context.Context, username string) error
}
