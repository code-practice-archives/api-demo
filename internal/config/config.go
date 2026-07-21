package config

import (
	"fmt"

	"github.com/creasty/defaults"
	"github.com/spf13/viper"
)

const DefaultConfigFile = "configs/config.yaml"

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"db"`
	JWT    JWTConfig    `mapstructure:"jwt"`
}

// Load 使用 viper 读取 YAML 配置。path 为空时使用 DefaultConfigFile。
func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigFile
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config %s: %w (copy configs/config.example.yaml to configs/config.yaml)", path, err)
	}

	// 必须先 Unmarshal 再 defaults.Set：后者只填充零值字段。
	// 若顺序颠倒，文件中未出现的键在 Unmarshal 时会被写成零值，冲掉已设的默认值。
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := defaults.Set(&cfg); err != nil {
		return nil, fmt.Errorf("apply config defaults: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Server.Addr == "" {
		return fmt.Errorf("server.addr is required")
	}
	switch c.DB.Driver {
	case DBDriverMySQL, DBDriverSQLite:
	default:
		return fmt.Errorf("db.driver must be %q or %q", DBDriverMySQL, DBDriverSQLite)
	}
	if c.DB.DSN == "" {
		return fmt.Errorf("db.dsn is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	return nil
}
