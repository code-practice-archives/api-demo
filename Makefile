.PHONY: run build tidy test init-config

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

tidy:
	go mod tidy

test:
	go test ./...
