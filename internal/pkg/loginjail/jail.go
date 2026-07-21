package loginjail

import "context"

// Jail 登录失败锁定存储。内存实现仅适合单机；分布式请用 Redis。
type Jail interface {
	Locked(ctx context.Context, username string) (bool, error)
	Fail(ctx context.Context, username string) (locked bool, err error)
	Reset(ctx context.Context, username string) error
}

const (
	StoreMemory = "memory"
	StoreRedis  = "redis"
)
