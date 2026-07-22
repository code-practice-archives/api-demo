package model

import (
	"gorm.io/plugin/soft_delete"
)

// Model 各表通用主键、时间戳与软删除字段（Unix 秒；DeletedAt=0 表示未删除）。
type Model struct {
	Id        int64                 `json:"id"`
	CreatedAt int64                 `json:"created_at"`
	UpdatedAt int64                 `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"-"`
}
