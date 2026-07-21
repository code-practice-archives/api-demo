package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	log        *logger.Logger
}

func New(httpServer *http.Server, log *logger.Logger) *Server {
	return &Server{
		httpServer: httpServer,
		log:        log,
	}
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
	return s.httpServer.Shutdown(ctx)
}

// Log 暴露 Logger，供 main 等上层记录启动/关闭日志。
func (s *Server) Log() *logger.Logger {
	return s.log
}
