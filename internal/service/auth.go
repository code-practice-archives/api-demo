package service

import (
	"context"
	"errors"
	"strings"

	"github.com/code-practice-archives/api-demo/internal/model"
	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/validator"
	"github.com/code-practice-archives/api-demo/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	repos *repository.Repositories
	jwt   *jwtx.Manager
}

func NewAuthService(repos *repository.Repositories, jwt *jwtx.Manager) *AuthService {
	return &AuthService{repos: repos, jwt: jwt}
}

type RegisterInput struct {
	Username string `validate:"required,min=3,max=64"`
	Password string `validate:"required,min=6,max=72"`
}

type LoginInput struct {
	Username string `validate:"required,min=3,max=64"`
	Password string `validate:"required,min=6,max=72"`
}

type AuthResult struct {
	Token string
	User  *model.User
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	in.Username = strings.TrimSpace(in.Username)
	if err := validator.Struct(&in); err != nil {
		return nil, err
	}

	exists, err := s.repos.User.ExistsByUsername(ctx, in.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errcode.ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: in.Username,
		Password: string(hash),
	}
	if err := s.repos.User.Create(ctx, user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errcode.ErrUsernameTaken
		}
		return nil, err
	}

	token, err := s.jwt.Sign(user.Id, user.Username)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	in.Username = strings.TrimSpace(in.Username)
	if err := validator.Struct(&in); err != nil {
		return nil, err
	}

	user, err := s.repos.User.FindByUsername(ctx, in.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return nil, errcode.ErrInvalidCredentials
	}

	token, err := s.jwt.Sign(user.Id, user.Username)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}
