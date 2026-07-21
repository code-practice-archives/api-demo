package repository

import (
	"context"

	"github.com/code-practice-archives/api-demo/internal/model"
	"github.com/code-practice-archives/api-demo/internal/repository/query"
	"gorm.io/gorm"
)

type UserRepository struct {
	q *query.Query
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{q: query.Use(db)}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.q.User.WithContext(ctx).Create(user)
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	u := r.q.User
	return u.WithContext(ctx).Where(u.Username.Eq(username)).First()
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	u := r.q.User
	return u.WithContext(ctx).Where(u.Id.Eq(id)).First()
}

func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	u := r.q.User
	count, err := u.WithContext(ctx).Where(u.Username.Eq(username)).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
