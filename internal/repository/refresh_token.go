package repository

import (
	"context"

	"github.com/code-practice-archives/api-demo/internal/model"
	"github.com/code-practice-archives/api-demo/internal/repository/query"
	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	q *query.Query
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{q: query.Use(db)}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	return r.q.RefreshToken.WithContext(ctx).Create(token)
}

func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	t := r.q.RefreshToken
	return t.WithContext(ctx).Where(t.TokenHash.Eq(hash)).First()
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id int64, revokedAt int64) error {
	t := r.q.RefreshToken
	_, err := t.WithContext(ctx).Where(t.Id.Eq(id), t.RevokedAt.Eq(0)).UpdateSimple(
		t.RevokedAt.Value(revokedAt),
		t.UpdatedAt.Value(revokedAt),
	)
	return err
}
