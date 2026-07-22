package repository

import (
	"context"

	"github.com/code-practice-archives/api-demo/internal/model"
	"gorm.io/gorm"
)

// UserStore 用户持久化接口；生产用 UserRepository，单测用 MockUserStore。
type UserStore interface {
	Create(ctx context.Context, user *model.User) error
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByID(ctx context.Context, id int64) (*model.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}

// RefreshTokenStore opaque refresh token 持久化接口。
type RefreshTokenStore interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	FindByTokenHash(ctx context.Context, hash string) (*model.RefreshToken, error)
	Revoke(ctx context.Context, id int64, revokedAt int64) error
}

// OAuthClientStore OAuth 客户端持久化接口。
type OAuthClientStore interface {
	FindByClientID(ctx context.Context, clientID string) (*model.OAuthClient, error)
}

// OAuthAuthorizationCodeStore 授权码持久化接口。
type OAuthAuthorizationCodeStore interface {
	Create(ctx context.Context, code *model.OAuthAuthorizationCode) error
	FindByCodeHash(ctx context.Context, hash string) (*model.OAuthAuthorizationCode, error)
	MarkUsed(ctx context.Context, id int64, usedAt int64) error
}

type Repositories struct {
	User             UserStore
	RefreshToken     RefreshTokenStore
	OAuthClient      OAuthClientStore
	OAuthAuthCode    OAuthAuthorizationCodeStore
}

func New(db *gorm.DB) *Repositories {
	return &Repositories{
		User:          NewUserRepository(db),
		RefreshToken:  NewRefreshTokenRepository(db),
		OAuthClient:   NewOAuthClientRepository(db),
		OAuthAuthCode: NewOAuthAuthorizationCodeRepository(db),
	}
}
