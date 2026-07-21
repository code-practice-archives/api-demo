package service

import (
	"context"
	"errors"
	"strings"

	"github.com/code-practice-archives/api-demo/internal/model"
	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/code-practice-archives/api-demo/internal/pkg/validator"
	"github.com/code-practice-archives/api-demo/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	repos *repository.Repositories
	jwt   *jwtx.Manager
	log   *logger.Logger
}

func NewAuthService(repos *repository.Repositories, jwt *jwtx.Manager, log *logger.Logger) *AuthService {
	return &AuthService{repos: repos, jwt: jwt, log: log.Named("auth")}
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
	log := s.log.WithContext(ctx).With(zap.String("username", in.Username))

	if err := validator.Struct(&in); err != nil {
		log.Warn("register validation failed", zap.Error(err))
		return nil, err
	}

	exists, err := s.repos.User.ExistsByUsername(ctx, in.Username)
	if err != nil {
		log.Error("register check username failed", zap.Error(err))
		return nil, err
	}
	if exists {
		log.Warn("register username taken")
		return nil, errcode.ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("register hash password failed", zap.Error(err))
		return nil, err
	}

	user := &model.User{
		Username: in.Username,
		Password: string(hash),
	}
	if err := s.repos.User.Create(ctx, user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			log.Warn("register username taken")
			return nil, errcode.ErrUsernameTaken
		}
		log.Error("register create user failed", zap.Error(err))
		return nil, err
	}

	token, err := s.jwt.Sign(user.Id, user.Username)
	if err != nil {
		log.Error("register sign token failed", zap.Error(err), zap.Int64("user_id", user.Id))
		return nil, err
	}

	log.Info("register success", zap.Int64("user_id", user.Id))
	return &AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	in.Username = strings.TrimSpace(in.Username)
	log := s.log.WithContext(ctx).With(zap.String("username", in.Username))

	if err := validator.Struct(&in); err != nil {
		log.Warn("login validation failed", zap.Error(err))
		return nil, err
	}

	user, err := s.repos.User.FindByUsername(ctx, in.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("login invalid credentials")
			return nil, errcode.ErrInvalidCredentials
		}
		log.Error("login find user failed", zap.Error(err))
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		log.Warn("login invalid credentials", zap.Int64("user_id", user.Id))
		return nil, errcode.ErrInvalidCredentials
	}

	token, err := s.jwt.Sign(user.Id, user.Username)
	if err != nil {
		log.Error("login sign token failed", zap.Error(err), zap.Int64("user_id", user.Id))
		return nil, err
	}

	log.Info("login success", zap.Int64("user_id", user.Id))
	return &AuthResult{Token: token, User: user}, nil
}
