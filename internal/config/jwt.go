package config

import "time"

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours" default:"24"`
}

func (c JWTConfig) Expire() time.Duration {
	hours := c.ExpireHours
	if hours <= 0 {
		hours = 24
	}
	return time.Duration(hours) * time.Hour
}
