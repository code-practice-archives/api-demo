package ipallowlist

import "net"

// Allowed 判断 clientIP 是否命中白名单条目（精确 IP 或 CIDR）。
// 非法 clientIP 或不在名单内返回 false。
func Allowed(clientIP string, entries []string) bool {
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	for _, entry := range entries {
		if entry == "" {
			continue
		}
		if parsed := net.ParseIP(entry); parsed != nil {
			if ip.Equal(parsed) {
				return true
			}
			continue
		}
		if _, network, err := net.ParseCIDR(entry); err == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}
