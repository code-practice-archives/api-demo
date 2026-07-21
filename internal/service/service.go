package service

import (
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/code-practice-archives/api-demo/internal/repository"
)

type Services struct {
	Auth *AuthService
}

func New(repos *repository.Repositories, jwt *jwtx.Manager, log *logger.Logger) *Services {
	return &Services{
		Auth: NewAuthService(repos, jwt, log),
	}
}
