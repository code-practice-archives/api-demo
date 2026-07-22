package database

import (
	"fmt"

	mysqldriver "github.com/go-sql-driver/mysql"
)

// Config 数据库连接配置。
type Config struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
}

func (c Config) Validate() error {
	if c.User == "" {
		return fmt.Errorf("db.user is required")
	}
	if c.Host == "" {
		return fmt.Errorf("db.host is required")
	}
	if c.Port <= 0 {
		return fmt.Errorf("db.port is required")
	}
	if c.Database == "" {
		return fmt.Errorf("db.database is required")
	}
	return nil
}

// DSN 由账号密码等字段组装 MySQL DSN，密码中的特殊字符无需手动 URL 编码。
func (c Config) DSN() string {
	cfg := mysqldriver.Config{
		User:                 c.User,
		Passwd:               c.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", c.Host, c.Port),
		DBName:               c.Database,
		Params:               map[string]string{"charset": "utf8mb4", "loc": "Local"},
		ParseTime:            true,
		AllowNativePasswords: true,
	}
	return cfg.FormatDSN()
}
