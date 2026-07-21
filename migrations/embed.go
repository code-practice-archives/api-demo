package migrations

import "embed"

// FS 内嵌 goose 迁移文件。
//
//go:embed mysql/*.sql
var FS embed.FS
