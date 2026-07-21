package repository

import (
	"context"
	"errors"

	"github.com/code-practice-archives/api-demo/internal/model"
	"github.com/code-practice-archives/api-demo/internal/repository/query"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

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
	user, err := u.WithContext(ctx).Where(u.Username.Eq(username)).First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	u := r.q.User
	count, err := u.WithContext(ctx).Where(u.Username.Eq(username)).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
