package ratelimit

import "time"

// Config 请求限流配置。根配置 YAML 键为 rate_limit。存储固定为 Redis。
// Enabled 不用 default 标签：bool 零值即 false，与 creasty/defaults 冲突。
type Config struct {
	Enabled       bool `mapstructure:"enabled"` // 默认关闭，需在 YAML 中显式打开
	Limit         int  `mapstructure:"limit" default:"120"`
	WindowSeconds int  `mapstructure:"window_seconds" default:"60"`
}

func (c Config) Window() time.Duration {
	sec := c.WindowSeconds
	if sec <= 0 {
		sec = 60
	}
	return time.Duration(sec) * time.Second
}
