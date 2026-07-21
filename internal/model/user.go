package model

import (
	"gorm.io/plugin/soft_delete"
)

// User 用户账号。
// 表结构约束见 migrations/；此处只保留序列化与软删除行为所需信息。
type User struct {
	Id        int64                 `json:"id"`
	CreatedAt int64                 `json:"created_at"`
	UpdatedAt int64                 `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"-"` // 删除时间，0 表示未删除

	Username string `json:"username"`
	Password string `json:"-"` // 仅存 bcrypt 哈希
}

func (User) TableName() string {
	return "users"
}
