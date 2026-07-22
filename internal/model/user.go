package model

// User 用户账号。
// 表结构约束见 migrations/；此处只保留序列化与软删除行为所需信息。
type User struct {
	Model

	Username string `json:"username"`
	Password string `json:"-"` // 仅存 bcrypt 哈希
}

func (User) TableName() string {
	return "users"
}
