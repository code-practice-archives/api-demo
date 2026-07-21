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

type Repositories struct {
	User         UserStore
	RefreshToken RefreshTokenStore
}

func New(db *gorm.DB) *Repositories {
	return &Repositories{
		User:         NewUserRepository(db),
		RefreshToken: NewRefreshTokenRepository(db),
	}
}
