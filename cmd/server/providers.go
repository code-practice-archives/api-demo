package main

import (
	"fmt"
	"net/http"

	"github.com/code-practice-archives/api-demo/internal/config"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/code-practice-archives/api-demo/internal/pkg/loginjail"
	"github.com/code-practice-archives/api-demo/internal/pkg/redisx"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func provideLogger(cfg logger.Config) (*logger.Logger, func(), error) {
	log, err := logger.New(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("init logger: %w", err)
	}
	return log, func() { _ = log.Sync() }, nil
}

func provideRedis(cfg redisx.Config) (*redis.Client, func(), error) {
	rdb, err := redisx.Open(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("open redis: %w", err)
	}
	return rdb, func() { _ = rdb.Close() }, nil
}

func provideJWTManager(cfg jwtx.Config) *jwtx.Manager {
	return jwtx.NewManager(cfg.Secret, cfg.Expire())
}

func provideLoginJail(cfg loginjail.Config, rdb *redis.Client) (loginjail.Jail, error) {
	maxRetries := cfg.MaxRetries
	lockFor := cfg.LockDuration()

	switch cfg.Store {
	case "", loginjail.StoreMemory:
		return loginjail.NewMemory(maxRetries, lockFor), nil
	case loginjail.StoreRedis:
		return loginjail.NewRedis(rdb, maxRetries, lockFor), nil
	default:
		return nil, fmt.Errorf("init login jail: auth.store must be %q or %q", loginjail.StoreMemory, loginjail.StoreRedis)
	}
}

func provideHTTPServer(cfg config.ServerConfig, engine *gin.Engine) *http.Server {
	return &http.Server{
		Addr:    cfg.Addr,
		Handler: engine,
	}
}
