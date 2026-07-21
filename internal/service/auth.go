package service

import (
	"context"
	"errors"
	"strings"

	"github.com/code-practice-archives/api-demo/internal/model"
	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	users *repository.UserRepository
	jwt   *jwtx.Manager
}

func NewAuthService(users *repository.UserRepository, jwt *jwtx.Manager) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}

type RegisterInput struct {
	Username string
	Password string
}

type LoginInput struct {
	Username string
	Password string
}

type AuthResult struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	username := strings.TrimSpace(in.Username)
	password := in.Password

	if err := validateCredentials(username, password); err != nil {
		return nil, err
	}

	exists, err := s.users.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errcode.ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: username,
		Password: string(hash),
	}
	if err := s.users.Create(ctx, user); err != nil {
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
	username := strings.TrimSpace(in.Username)
	password := in.Password

	if err := validateCredentials(username, password); err != nil {
		return nil, err
	}

	user, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, errcode.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errcode.ErrInvalidCredentials
	}

	token, err := s.jwt.Sign(user.Id, user.Username)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}

func validateCredentials(username, password string) error {
	if len(username) < 3 || len(username) > 64 {
		return errcode.ErrInvalidArgument.WithMessage("username must be 3-64 chars, password must be 6-72 chars")
	}
	if len(password) < 6 || len(password) > 72 {
		return errcode.ErrInvalidArgument.WithMessage("username must be 3-64 chars, password must be 6-72 chars")
	}
	return nil
}
