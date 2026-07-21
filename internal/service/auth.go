package service

import (
	"errors"
	"strings"

	"github.com/code-practice-archives/api-demo/internal/model"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("username or password is incorrect")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidInput       = errors.New("invalid input")
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

func (s *AuthService) Register(in RegisterInput) (*AuthResult, error) {
	username := strings.TrimSpace(in.Username)
	password := in.Password

	if err := validateCredentials(username, password); err != nil {
		return nil, err
	}

	exists, err := s.users.ExistsByUsername(username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: username,
		Password: string(hash),
	}
	if err := s.users.Create(user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrUsernameTaken
		}
		return nil, err
	}

	token, err := s.jwt.Sign(user.Id, user.Username)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) Login(in LoginInput) (*AuthResult, error) {
	username := strings.TrimSpace(in.Username)
	password := in.Password

	if err := validateCredentials(username, password); err != nil {
		return nil, err
	}

	user, err := s.users.FindByUsername(username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.jwt.Sign(user.Id, user.Username)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}

func validateCredentials(username, password string) error {
	if len(username) < 3 || len(username) > 64 {
		return ErrInvalidInput
	}
	if len(password) < 6 || len(password) > 72 {
		return ErrInvalidInput
	}
	return nil
}
