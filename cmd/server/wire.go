//go:build wireinject

package main

import (
	"github.com/code-practice-archives/api-demo/internal/config"
	"github.com/code-practice-archives/api-demo/internal/handler"
	"github.com/code-practice-archives/api-demo/internal/pkg/database"
	"github.com/code-practice-archives/api-demo/internal/repository"
	"github.com/code-practice-archives/api-demo/internal/router"
	"github.com/code-practice-archives/api-demo/internal/server"
	"github.com/code-practice-archives/api-demo/internal/service"
	"github.com/google/wire"
)

//go:generate go run -mod=mod github.com/google/wire/cmd/wire

func initializeServer(cfg *config.Config) (*server.Server, func(), error) {
	wire.Build(
		wire.FieldsOf(new(*config.Config), "Log", "DB", "Jail", "Redis", "JWT", "Server"),
		provideLogger,
		database.Open,
		provideRedis,
		provideJWTManager,
		provideLoginJail,
		repository.New,
		service.New,
		handler.NewAuthHandler,
		handler.New,
		router.New,
		provideHTTPServer,
		server.New,
	)
	return nil, nil, nil
}
