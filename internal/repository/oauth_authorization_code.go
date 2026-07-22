package repository

import (
	"context"

	"github.com/code-practice-archives/api-demo/internal/model"
	"github.com/code-practice-archives/api-demo/internal/repository/query"
	"gorm.io/gorm"
)

type OAuthAuthorizationCodeRepository struct {
	q *query.Query
}

func NewOAuthAuthorizationCodeRepository(db *gorm.DB) *OAuthAuthorizationCodeRepository {
	return &OAuthAuthorizationCodeRepository{q: query.Use(db)}
}

func (r *OAuthAuthorizationCodeRepository) Create(ctx context.Context, code *model.OAuthAuthorizationCode) error {
	return r.q.OAuthAuthorizationCode.WithContext(ctx).Create(code)
}

func (r *OAuthAuthorizationCodeRepository) FindByCodeHash(ctx context.Context, hash string) (*model.OAuthAuthorizationCode, error) {
	c := r.q.OAuthAuthorizationCode
	return c.WithContext(ctx).Where(c.CodeHash.Eq(hash)).First()
}

func (r *OAuthAuthorizationCodeRepository) MarkUsed(ctx context.Context, id int64, usedAt int64) error {
	c := r.q.OAuthAuthorizationCode
	_, err := c.WithContext(ctx).Where(c.Id.Eq(id), c.UsedAt.Eq(0)).Update(c.UsedAt, usedAt)
	return err
}
