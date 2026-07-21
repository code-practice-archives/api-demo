package service

import (
	"context"
	"errors"
	"strings"

	"github.com/code-practice-archives/api-demo/internal/model"
	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/code-practice-archives/api-demo/internal/pkg/loginjail"
	"github.com/code-practice-archives/api-demo/internal/pkg/validator"
	"github.com/code-practice-archives/api-demo/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// dummyPasswordHash 用于用户不存在时的 bcrypt 比对，降低时序侧信道。
const dummyPasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

type AuthService struct {
	repos *repository.Repositories
	jwt   *jwtx.Manager
	jail  loginjail.Jail
	log   *logger.Logger
}

func NewAuthService(repos *repository.Repositories, jwt *jwtx.Manager, jail loginjail.Jail, log *logger.Logger) *AuthService {
	return &AuthService{repos: repos, jwt: jwt, jail: jail, log: log.Named("auth")}
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

// Register 注册新用户并签发 JWT。密码仅存 bcrypt 哈希；并发下依赖唯一约束兜底重复用户名。
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	in.Username = strings.TrimSpace(in.Username)
	log := s.log.WithContext(ctx).With(zap.String("username", in.Username))

	// 1. 校验输入
	if err := validator.Struct(&in); err != nil {
		log.Warn("register validation failed", zap.Error(err))
		return nil, err
	}

	// 2. 预检用户名（快速失败；真正防重仍靠 DB 唯一索引）
	exists, err := s.repos.User.ExistsByUsername(ctx, in.Username)
	if err != nil {
		log.Error("register check username failed", zap.Error(err))
		return nil, err
	}
	if exists {
		log.Warn("register username taken")
		return nil, errcode.ErrUsernameTaken
	}

	// 3. 哈希密码后落库
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
		// Exists 与 Create 之间可能被并发插入，映射为业务错误
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			log.Warn("register username taken")
			return nil, errcode.ErrUsernameTaken
		}
		log.Error("register create user failed", zap.Error(err))
		return nil, err
	}

	// 4. 签发 Token，注册成功即视为已登录
	token, err := s.jwt.Sign(user.Id, user.Username)
	if err != nil {
		log.Error("register sign token failed", zap.Error(err), zap.Int64("user_id", user.Id))
		return nil, err
	}

	log.Info("register success", zap.Int64("user_id", user.Id))
	return &AuthResult{Token: token, User: user}, nil
}

// Login 校验凭证并签发 JWT。失败统一返回凭证错误；连续失败由 jail 锁定以防爆破。
func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	in.Username = strings.TrimSpace(in.Username)
	log := s.log.WithContext(ctx).With(zap.String("username", in.Username))

	// 1. 校验输入
	if err := validator.Struct(&in); err != nil {
		log.Warn("login validation failed", zap.Error(err))
		return nil, err
	}

	// 2. 已锁定则直接拒绝，避免无意义的验密与计次
	locked, err := s.jail.Locked(ctx, in.Username)
	if err != nil {
		log.Error("login check jail failed", zap.Error(err))
		return nil, err
	}
	if locked {
		log.Warn("login account locked")
		return nil, errcode.ErrAccountLocked
	}

	// 3. 查用户并验密；用户不存在时仍走 dummy bcrypt + Fail，降低枚举与时序差异
	user, err := s.repos.User.FindByUsername(ctx, in.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(in.Password))
			locked, jerr := s.jail.Fail(ctx, in.Username)
			if jerr != nil {
				log.Error("login jail fail failed", zap.Error(jerr))
				return nil, jerr
			}
			if locked {
				log.Warn("login account locked after failures")
				return nil, errcode.ErrAccountLocked
			}
			log.Warn("login invalid credentials")
			return nil, errcode.ErrInvalidCredentials
		}
		log.Error("login find user failed", zap.Error(err))
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		locked, jerr := s.jail.Fail(ctx, in.Username)
		if jerr != nil {
			log.Error("login jail fail failed", zap.Error(jerr), zap.Int64("user_id", user.Id))
			return nil, jerr
		}
		if locked {
			log.Warn("login account locked after failures", zap.Int64("user_id", user.Id))
			return nil, errcode.ErrAccountLocked
		}
		log.Warn("login invalid credentials", zap.Int64("user_id", user.Id))
		return nil, errcode.ErrInvalidCredentials
	}

	// 4. 成功后清零失败计数并签发 Token
	if err := s.jail.Reset(ctx, in.Username); err != nil {
		log.Error("login jail reset failed", zap.Error(err), zap.Int64("user_id", user.Id))
		return nil, err
	}

	token, err := s.jwt.Sign(user.Id, user.Username)
	if err != nil {
		log.Error("login sign token failed", zap.Error(err), zap.Int64("user_id", user.Id))
		return nil, err
	}

	log.Info("login success", zap.Int64("user_id", user.Id))
	return &AuthResult{Token: token, User: user}, nil
}

// Me 按 userID 返回当前用户；记录不存在时映射为 Unauthorized，避免暴露「用户已删」。
func (s *AuthService) Me(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.repos.User.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrUnauthorized
		}
		return nil, err
	}
	return user, nil
}
