package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/code-practice-archives/api-demo/internal/config"
	"github.com/code-practice-archives/api-demo/internal/handler"
	"github.com/code-practice-archives/api-demo/internal/pkg/database"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/code-practice-archives/api-demo/internal/pkg/loginjail"
	"github.com/code-practice-archives/api-demo/internal/pkg/redisx"
	"github.com/code-practice-archives/api-demo/internal/repository"
	"github.com/code-practice-archives/api-demo/internal/router"
	"github.com/code-practice-archives/api-demo/internal/service"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	log        *logger.Logger
	rdb        *redis.Client
}

func New(cfg *config.Config) (*Server, error) {
	log, err := logger.New(cfg.Log)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	db, err := database.Open(cfg.DB)
	if err != nil {
		_ = log.Sync()
		return nil, err
	}

	var rdb *redis.Client
	if cfg.Jail.Store == loginjail.StoreRedis {
		rdb, err = redisx.Open(cfg.Redis)
		if err != nil {
			_ = log.Sync()
			return nil, fmt.Errorf("open redis: %w", err)
		}
	}

	repos := repository.New(db)
	jwtMgr := jwtx.NewManager(cfg.JWT.Secret, cfg.JWT.Expire())
	jail, err := newLoginJail(cfg.Jail, rdb)
	if err != nil {
		if rdb != nil {
			_ = rdb.Close()
		}
		_ = log.Sync()
		return nil, fmt.Errorf("init login jail: %w", err)
	}
	svc := service.New(repos, jwtMgr, jail, log)

	return &Server{
		log: log,
		rdb: rdb,
		httpServer: &http.Server{
			Addr: cfg.Server.Addr,
			Handler: router.New(handler.Handlers{
				Auth: handler.NewAuthHandler(svc),
			}, jwtMgr, log),
		},
	}, nil
}

func (s *Server) Start() error {
	s.log.WithContext(context.Background()).Info("server listening", zap.String("addr", s.httpServer.Addr))
	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.httpServer.Shutdown(ctx)
	if s.rdb != nil {
		_ = s.rdb.Close()
	}
	_ = s.log.Sync()
	return err
}

// Log 暴露 Logger，供 main 等上层记录启动/关闭日志。
func (s *Server) Log() *logger.Logger {
	return s.log
}
