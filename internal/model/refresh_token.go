package model

// RefreshToken 服务端持久化的 opaque refresh token（库中只存哈希）。
// 表结构约束见 migrations/。
type RefreshToken struct {
	Id        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	TokenHash string `json:"-"` // SHA-256 hex，不回传客户端
	ExpiresAt int64  `json:"expires_at"`
	RevokedAt int64  `json:"-"` // 0 表示未吊销
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
