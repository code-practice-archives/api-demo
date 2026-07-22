package config

import (
	"fmt"

	"github.com/code-practice-archives/api-demo/internal/pkg/database"
	"github.com/code-practice-archives/api-demo/internal/pkg/ipallowlist"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/code-practice-archives/api-demo/internal/pkg/loginjail"
	"github.com/code-practice-archives/api-demo/internal/pkg/ratelimit"
	"github.com/code-practice-archives/api-demo/internal/pkg/redisx"
	"github.com/creasty/defaults"
	"github.com/spf13/viper"
)

const DefaultConfigFile = "configs/config.yaml"

// Config 聚合各子系统配置；字段类型定义在对应包内，此处只做装配。
type Config struct {
	Server    ServerConfig     `mapstructure:"server"`     // HTTP 监听地址等，留在本包以避免与 server 循环依赖
	DB        database.Config  `mapstructure:"db"`         // 数据库连接
	Redis     redisx.Config    `mapstructure:"redis"`      // Redis 连接
	JWT       jwtx.Config      `mapstructure:"jwt"`        // JWT 签发密钥与过期时间
	Jail      loginjail.Config `mapstructure:"auth"`       // 登录失败锁定；YAML 键为 auth
	RateLimit   ratelimit.Config   `mapstructure:"rate_limit"`   // 请求限流
	IPAllowlist ipallowlist.Config `mapstructure:"ip_allowlist"` // 私有接口 IP 白名单
	Log         logger.Config      `mapstructure:"log"`          // 日志级别与文件轮转
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
	if err := c.Server.Validate(); err != nil {
		return err
	}
	if err := c.DB.Validate(); err != nil {
		return err
	}
	return c.JWT.Validate()
}
