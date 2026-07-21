// Package main 基于 internal/model 生成类型安全的 DAO（internal/repository/query）。
//
//	go run ./cmd/gen
//	make gen
package main

import (
	"github.com/code-practice-archives/api-demo/internal/model"
	"gorm.io/gen"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:       "./internal/repository/query",
		Mode:          gen.WithDefaultQuery,
		FieldNullable: false,
	})

	g.ApplyBasic(model.User{})
	g.Execute()
}
