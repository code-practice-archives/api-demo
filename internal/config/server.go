package config

type ServerConfig struct {
	Addr string `mapstructure:"addr" default:":8080"`
}
