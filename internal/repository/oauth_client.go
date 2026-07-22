package repository

import (
	"context"

	"github.com/code-practice-archives/api-demo/internal/model"
	"github.com/code-practice-archives/api-demo/internal/repository/query"
	"gorm.io/gorm"
)

type OAuthClientRepository struct {
	q *query.Query
}

func NewOAuthClientRepository(db *gorm.DB) *OAuthClientRepository {
	return &OAuthClientRepository{q: query.Use(db)}
}

func (r *OAuthClientRepository) FindByClientID(ctx context.Context, clientID string) (*model.OAuthClient, error) {
	c := r.q.OAuthClient
	return c.WithContext(ctx).Where(c.ClientID.Eq(clientID)).First()
}
