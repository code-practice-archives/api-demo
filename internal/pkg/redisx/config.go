package redisx

// Config Redis 连接配置。
type Config struct {
	Addr     string `mapstructure:"addr" default:"127.0.0.1:6379"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db" default:"0"`
}
