package jwtx

import (
	"fmt"
	"time"
)

// Config JWT / refresh 签发配置。
type Config struct {
	Secret            string `mapstructure:"secret"`
	ExpireHours       int    `mapstructure:"expire_hours" default:"1"`        // access token TTL（小时）
	RefreshExpireDays int    `mapstructure:"refresh_expire_days" default:"7"` // refresh token RTTL（天）
}

func (c Config) Expire() time.Duration {
	hours := c.ExpireHours
	if hours <= 0 {
		hours = 1
	}
	return time.Duration(hours) * time.Hour
}

func (c Config) RefreshExpire() time.Duration {
	days := c.RefreshExpireDays
	if days <= 0 {
		days = 7
	}
	return time.Duration(days) * 24 * time.Hour
}

func (c Config) Validate() error {
	if c.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	return nil
}
