.PHONY: run build tidy test gen init-config

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

tidy:
	go mod tidy

test:
	go test ./...

# 基于 internal/model 重新生成 internal/repository/query（gorm gen）
gen:
	go run ./cmd/gen
