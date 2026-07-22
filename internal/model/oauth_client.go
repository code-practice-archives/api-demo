package model

// OAuthClient OAuth 2.0 客户端配置。
// 表结构约束见 migrations/。
type OAuthClient struct {
	Model

	// 展示用名称，不参与鉴权逻辑
	Name string `json:"name"`

	// 对外公开的客户端标识，authorize/token 请求必填
	ClientID string `json:"client_id"`

	// ClientSecretHash 客户端密钥的 bcrypt 哈希；空字符串表示公开客户端。
	// 公开客户端不得提交 client_secret，且必须使用 PKCE；机密客户端换 token 时校验密钥。
	ClientSecretHash string `json:"-"`

	// RedirectURIs 允许的回调地址列表，JSON 字符串数组。
	// authorize / token 的 redirect_uri 必须与其中某一项精确匹配，防止开放重定向。
	RedirectURIs string `json:"redirect_uris"`
}

func (OAuthClient) TableName() string {
	return "oauth_clients"
}

// IsPublic 公开客户端无 client_secret，须走 PKCE。
func (c *OAuthClient) IsPublic() bool {
	return c.ClientSecretHash == ""
}
