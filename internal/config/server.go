package config

import "fmt"

type ServerConfig struct {
	Addr string `mapstructure:"addr" default:":8080"`
}

func (c ServerConfig) Validate() error {
	if c.Addr == "" {
		return fmt.Errorf("server.addr is required")
	}
	return nil
}
