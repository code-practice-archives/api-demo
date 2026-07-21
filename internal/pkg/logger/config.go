package logger

// Config 日志输出与轮转配置。
type Config struct {
	Level      string `mapstructure:"level" default:"info"` // debug | info | warn | error
	Filename   string `mapstructure:"filename" default:"logs/app.log"`
	MaxSize    int    `mapstructure:"max_size" default:"100"` // MB
	MaxBackups int    `mapstructure:"max_backups" default:"7"`
	MaxAge     int    `mapstructure:"max_age" default:"30"` // days
	Compress   bool   `mapstructure:"compress" default:"true"`
}
