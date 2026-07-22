package ipallowlist

// Config IP 白名单配置。根配置 YAML 键为 ip_allowlist。
// Enabled 不用 default 标签：bool 零值即 false，与 creasty/defaults 冲突。
type Config struct {
	Enabled bool     `mapstructure:"enabled"` // 默认关闭，需在 YAML 中显式打开
	IPs     []string `mapstructure:"ips"`     // 允许的 IP 或 CIDR，如 127.0.0.1、10.0.0.0/8
}
