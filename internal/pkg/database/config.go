package database

import "fmt"

const (
	DriverMySQL  = "mysql"
	DriverSQLite = "sqlite"
)

// Config 数据库连接配置。
type Config struct {
	Driver string `mapstructure:"driver" default:"mysql"` // mysql | sqlite
	DSN    string `mapstructure:"dsn"`
}

func (c Config) Validate() error {
	switch c.Driver {
	case DriverMySQL, DriverSQLite:
	default:
		return fmt.Errorf("db.driver must be %q or %q", DriverMySQL, DriverSQLite)
	}
	if c.DSN == "" {
		return fmt.Errorf("db.dsn is required")
	}
	return nil
}
