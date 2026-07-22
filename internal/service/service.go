package service

import (
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/code-practice-archives/api-demo/internal/pkg/loginjail"
	"github.com/code-practice-archives/api-demo/internal/pkg/oauth"
	"github.com/code-practice-archives/api-demo/internal/repository"
)

type Services struct {
	Auth  *AuthService
	OAuth *OAuthService
}

func New(repos *repository.Repositories, jwt *jwtx.Manager, jail loginjail.Jail, oauthCfg oauth.Config, log *logger.Logger) *Services {
	return &Services{
		Auth:  NewAuthService(repos, jwt, jail, log),
		OAuth: NewOAuthService(repos, jwt, oauthCfg, log),
	}
}
