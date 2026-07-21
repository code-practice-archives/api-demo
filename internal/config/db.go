package config

const (
	DBDriverMySQL  = "mysql"
	DBDriverSQLite = "sqlite"
)

type DBConfig struct {
	Driver string `mapstructure:"driver" default:"mysql"` // mysql | sqlite
	DSN    string `mapstructure:"dsn"`
}
