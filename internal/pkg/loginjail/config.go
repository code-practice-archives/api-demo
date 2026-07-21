package loginjail

import "time"

// Config 登录失败锁定相关配置。根配置 YAML 键为 auth。存储固定为 Redis。
type Config struct {
	MaxRetries  int `mapstructure:"max_retries" default:"5"`
	LockMinutes int `mapstructure:"lock_minutes" default:"15"`
}

func (c Config) LockDuration() time.Duration {
	minutes := c.LockMinutes
	if minutes <= 0 {
		minutes = 15
	}
	return time.Duration(minutes) * time.Minute
}
