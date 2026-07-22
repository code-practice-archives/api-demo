package ipallowlist

import "testing"

func TestAllowed(t *testing.T) {
	tests := []struct {
		name     string
		clientIP string
		entries  []string
		want     bool
	}{
		{
			name:     "精确匹配 IPv4",
			clientIP: "127.0.0.1",
			entries:  []string{"127.0.0.1"},
			want:     true,
		},
		{
			name:     "精确匹配 IPv6",
			clientIP: "::1",
			entries:  []string{"::1"},
			want:     true,
		},
		{
			name:     "CIDR 命中",
			clientIP: "10.0.0.8",
			entries:  []string{"10.0.0.0/8"},
			want:     true,
		},
		{
			name:     "CIDR 未命中",
			clientIP: "192.168.1.1",
			entries:  []string{"10.0.0.0/8"},
			want:     false,
		},
		{
			name:     "名单外 IP",
			clientIP: "8.8.8.8",
			entries:  []string{"127.0.0.1", "::1"},
			want:     false,
		},
		{
			name:     "非法 clientIP",
			clientIP: "not-an-ip",
			entries:  []string{"127.0.0.1"},
			want:     false,
		},
		{
			name:     "空名单拒绝",
			clientIP: "127.0.0.1",
			entries:  nil,
			want:     false,
		},
		{
			name:     "跳过空条目",
			clientIP: "127.0.0.1",
			entries:  []string{"", "127.0.0.1"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Allowed(tt.clientIP, tt.entries); got != tt.want {
				t.Fatalf("Allowed(%q, %v) = %v, want %v", tt.clientIP, tt.entries, got, tt.want)
			}
		})
	}
}
