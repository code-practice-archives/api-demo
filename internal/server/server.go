package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/code-practice-archives/api-demo/internal/handler"
)

type Server struct {
	httpServer *http.Server
}

func New() *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    ":8080",
			Handler: handler.NewRouter(),
		},
	}
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
