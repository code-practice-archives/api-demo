package model

// OAuthAuthorizationCode Authorization Code + PKCE 流程中的一次性授权码。
// 表结构约束见 migrations/；库中只存哈希，明文 code 仅在 authorize 响应中返回一次。
type OAuthAuthorizationCode struct {
	Model

	// CodeHash 授权码明文的 SHA-256 hex，用于 token 兑换时查找；不回传客户端。
	CodeHash string `json:"-"`

	// 授予访问权限的资源所有者（已登录用户）
	UserID int64 `json:"user_id"`

	// 申请授权的客户端；兑换时必须与请求方一致
	ClientID string `json:"client_id"`

	// 授权时登记的回调；兑换时须原样回传并与登记值一致
	RedirectURI string `json:"redirect_uri"`

	// 授权范围，原样回传至 token 响应（当前不做细粒度校验）
	Scope string `json:"scope"`

	// CodeChallenge PKCE S256 挑战值：BASE64URL(SHA256(code_verifier))。
	// 兑换 token 时用客户端提交的 code_verifier 重算并比对，防止授权码拦截。
	CodeChallenge string `json:"-"`

	// CodeChallengeMethod 挑战算法，目前仅支持 S256（拒绝 plain）。
	CodeChallengeMethod string `json:"code_challenge_method"`

	// 授权码过期时间（Unix 秒），通常远短于 refresh
	ExpiresAt int64 `json:"expires_at"`

	// 首次兑换时间（Unix 秒）；0 表示未使用，用于防重放
	UsedAt int64 `json:"-"`
}

func (OAuthAuthorizationCode) TableName() string {
	return "oauth_authorization_codes"
}
