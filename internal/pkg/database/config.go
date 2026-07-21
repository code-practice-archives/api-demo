package database

import "fmt"

// Config 数据库连接配置。
type Config struct {
	DSN string `mapstructure:"dsn"`
}

func (c Config) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("db.dsn is required")
	}
	return nil
}
