package jwtx

import (
	"fmt"
	"time"
)

// Config JWT 签发配置。
type Config struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours" default:"24"`
}

func (c Config) Expire() time.Duration {
	hours := c.ExpireHours
	if hours <= 0 {
		hours = 24
	}
	return time.Duration(hours) * time.Hour
}

func (c Config) Validate() error {
	if c.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	return nil
}
