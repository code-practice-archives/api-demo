package model

// RefreshToken 服务端持久化的 opaque refresh token（库中只存哈希）。
// 表结构约束见 migrations/。
type RefreshToken struct {
	Model

	UserID int64 `json:"user_id"` // 令牌所属用户

	// ClientID 签发方标识：空字符串表示第一方 /auth/*；非空表示 OAuth 客户端。
	// refresh 时必须与请求方 client_id 一致，防止跨客户端盗用。
	ClientID string `json:"client_id"`

	// TokenHash 明文 refresh token 的 SHA-256 hex；库中永不存明文，接口也不回传。
	TokenHash string `json:"-"`

	// 过期时间（Unix 秒），超时后一律视为无效
	ExpiresAt int64 `json:"expires_at"`

	// 吊销时间（Unix 秒）；0 表示未吊销，轮换/登出时写入
	RevokedAt int64 `json:"-"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
