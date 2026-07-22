package oauth

import "time"

// Config OAuth 授权服务器配置。根配置 YAML 键为 oauth。
type Config struct {
	CodeTTLMinutes int `mapstructure:"code_ttl_minutes" default:"10"` // 授权码有效分钟数
}

func (c Config) CodeTTL() time.Duration {
	minutes := c.CodeTTLMinutes
	if minutes <= 0 {
		minutes = 10
	}
	return time.Duration(minutes) * time.Minute
}
