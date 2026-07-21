package loginjail

import (
	"fmt"
	"time"
)

// Config 登录失败锁定相关配置。根配置 YAML 键为 auth。
type Config struct {
	MaxRetries  int    `mapstructure:"max_retries" default:"5"`
	LockMinutes int    `mapstructure:"lock_minutes" default:"15"`
	Store       string `mapstructure:"store" default:"memory"` // memory | redis
}

func (c Config) LockDuration() time.Duration {
	minutes := c.LockMinutes
	if minutes <= 0 {
		minutes = 15
	}
	return time.Duration(minutes) * time.Minute
}

// Validate 校验 store；选用 redis 时要求 redisAddr 非空。
func (c Config) Validate(redisAddr string) error {
	switch c.Store {
	case "", StoreMemory, StoreRedis:
	default:
		return fmt.Errorf("auth.store must be %q or %q", StoreMemory, StoreRedis)
	}
	if c.Store == StoreRedis && redisAddr == "" {
		return fmt.Errorf("redis.addr is required when auth.store is redis")
	}
	return nil
}
