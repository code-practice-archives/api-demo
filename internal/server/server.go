package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/code-practice-archives/api-demo/internal/config"
	"github.com/code-practice-archives/api-demo/internal/handler"
	"github.com/code-practice-archives/api-demo/internal/pkg/database"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/repository"
	"github.com/code-practice-archives/api-demo/internal/router"
	"github.com/code-practice-archives/api-demo/internal/service"
)

type Server struct {
	httpServer *http.Server
}

func New(cfg *config.Config) (*Server, error) {
	db, err := database.Open(cfg.DB.Driver, cfg.DB.DSN)
	if err != nil {
		return nil, err
	}

	repos := repository.New(db)
	jwtMgr := jwtx.NewManager(cfg.JWT.Secret, cfg.JWT.Expire())
	svc := service.New(repos, jwtMgr)

	return &Server{
		httpServer: &http.Server{
			Addr: cfg.Server.Addr,
			Handler: router.New(handler.Handlers{
				Auth: handler.NewAuthHandler(svc),
			}),
		},
	}, nil
}

func (s *Server) Start() error {
	log.Printf("server listening on %s", s.httpServer.Addr)
	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
