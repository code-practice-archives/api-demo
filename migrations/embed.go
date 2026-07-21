package migrations

import "embed"

// FS 内嵌各数据库方言的 goose 迁移文件。
//
//go:embed mysql/*.sql sqlite/*.sql
var FS embed.FS
