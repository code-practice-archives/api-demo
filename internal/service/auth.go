package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/code-practice-archives/api-demo/internal/model"
	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/code-practice-archives/api-demo/internal/pkg/loginjail"
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

type AuthResult struct {
	Token        string
	RefreshToken string
	ExpiresIn    int64 // access token 剩余秒数
	User         *model.User
}

type RegisterInput struct {
	Username string
	Password string
}

const (
	usernameMinLen = 3
	usernameMaxLen = 64
	passwordMinLen = 6
	passwordMaxLen = 72 // bcrypt 上限
)

// validateRegisterInput 注册策略：用户名/密码长度约束。登录不走这套规则。
func validateRegisterInput(username, password string) error {
	switch {
	case username == "":
		return errcode.ErrInvalidArgument.WithMessage("username is required")
	case len(username) < usernameMinLen:
		return errcode.ErrInvalidArgument.WithMessage(fmt.Sprintf("username must be at least %d characters", usernameMinLen))
	case len(username) > usernameMaxLen:
		return errcode.ErrInvalidArgument.WithMessage(fmt.Sprintf("username must be at most %d characters", usernameMaxLen))
	case password == "":
		return errcode.ErrInvalidArgument.WithMessage("password is required")
	case len(password) < passwordMinLen:
		return errcode.ErrInvalidArgument.WithMessage(fmt.Sprintf("password must be at least %d characters", passwordMinLen))
	case len(password) > passwordMaxLen:
		return errcode.ErrInvalidArgument.WithMessage(fmt.Sprintf("password must be at most %d characters", passwordMaxLen))
	default:
		return nil
	}
}

// Register 注册新用户并签发 access + refresh。密码仅存 bcrypt 哈希；并发下依赖唯一约束兜底重复用户名。
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	in.Username = strings.TrimSpace(in.Username)
	log := s.log.WithContext(ctx).With(zap.String("username", in.Username))

	// 1. 校验输入
	if err := validateRegisterInput(in.Username, in.Password); err != nil {
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
	result, err := s.issueTokens(ctx, user)
	if err != nil {
		log.Error("register issue tokens failed", zap.Error(err), zap.Int64("user_id", user.Id))
		return nil, err
	}

	log.Info("register success", zap.Int64("user_id", user.Id))
	return result, nil
}

type LoginInput struct {
	Username string
	Password string
}

// validateLoginInput 登录只要求凭证非空；长短是否合法交给验密结果，避免泄露注册策略。
func validateLoginInput(username, password string) error {
	switch {
	case username == "":
		return errcode.ErrInvalidArgument.WithMessage("username is required")
	case password == "":
		return errcode.ErrInvalidArgument.WithMessage("password is required")
	default:
		return nil
	}
}

// Login 校验凭证并签发 access + refresh。失败统一返回凭证错误；连续失败由 jail 锁定以防爆破。
func (s *AuthService) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	in.Username = strings.TrimSpace(in.Username)
	log := s.log.WithContext(ctx).With(zap.String("username", in.Username))

	// 1. 校验输入
	if err := validateLoginInput(in.Username, in.Password); err != nil {
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

	result, err := s.issueTokens(ctx, user)
	if err != nil {
		log.Error("login issue tokens failed", zap.Error(err), zap.Int64("user_id", user.Id))
		return nil, err
	}

	log.Info("login success", zap.Int64("user_id", user.Id))
	return result, nil
}

type RefreshInput struct {
	RefreshToken string
}

// Refresh 用 opaque refresh token 轮换签发新的 access + refresh；旧 refresh 立即吊销。
func (s *AuthService) Refresh(ctx context.Context, in RefreshInput) (*AuthResult, error) {
	plain := strings.TrimSpace(in.RefreshToken)
	log := s.log.WithContext(ctx)

	if plain == "" {
		return nil, errcode.ErrInvalidArgument.WithMessage("refresh_token is required")
	}

	stored, err := s.repos.RefreshToken.FindByTokenHash(ctx, hashRefreshToken(plain))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("refresh token not found")
			return nil, errcode.ErrInvalidRefreshToken
		}
		log.Error("refresh find token failed", zap.Error(err))
		return nil, err
	}

	now := time.Now().Unix()
	if stored.RevokedAt != 0 || stored.ExpiresAt <= now {
		log.Warn("refresh token revoked or expired", zap.Int64("token_id", stored.Id), zap.Int64("user_id", stored.UserID))
		return nil, errcode.ErrInvalidRefreshToken
	}
	// OAuth 签发的 refresh 必须走 /oauth/token，避免被第一方端点轮换成无 client 绑定的令牌。
	if stored.ClientID != "" {
		log.Warn("refresh rejected oauth token via first-party endpoint", zap.Int64("token_id", stored.Id), zap.String("client_id", stored.ClientID))
		return nil, errcode.ErrInvalidRefreshToken
	}

	user, err := s.repos.User.FindByID(ctx, stored.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrInvalidRefreshToken
		}
		log.Error("refresh find user failed", zap.Error(err), zap.Int64("user_id", stored.UserID))
		return nil, err
	}

	if err := s.repos.RefreshToken.Revoke(ctx, stored.Id, now); err != nil {
		log.Error("refresh revoke old token failed", zap.Error(err), zap.Int64("token_id", stored.Id))
		return nil, err
	}

	result, err := s.issueTokens(ctx, user)
	if err != nil {
		log.Error("refresh issue tokens failed", zap.Error(err), zap.Int64("user_id", user.Id))
		return nil, err
	}

	log.Info("refresh success", zap.Int64("user_id", user.Id))
	return result, nil
}

type LogoutInput struct {
	RefreshToken string
}

// Logout 吊销给定的 refresh token；幂等（已失效也视为成功）。
func (s *AuthService) Logout(ctx context.Context, in LogoutInput) error {
	plain := strings.TrimSpace(in.RefreshToken)
	log := s.log.WithContext(ctx)

	if plain == "" {
		return errcode.ErrInvalidArgument.WithMessage("refresh_token is required")
	}

	stored, err := s.repos.RefreshToken.FindByTokenHash(ctx, hashRefreshToken(plain))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		log.Error("logout find token failed", zap.Error(err))
		return err
	}

	if stored.RevokedAt != 0 {
		return nil
	}

	if err := s.repos.RefreshToken.Revoke(ctx, stored.Id, time.Now().Unix()); err != nil {
		log.Error("logout revoke token failed", zap.Error(err), zap.Int64("token_id", stored.Id))
		return err
	}

	log.Info("logout success", zap.Int64("user_id", stored.UserID))
	return nil
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

// issueTokens 签发短寿命 JWT access，并持久化 opaque refresh（仅存哈希）。
func (s *AuthService) issueTokens(ctx context.Context, user *model.User) (*AuthResult, error) {
	access, err := s.jwt.Sign(user.Id, user.Username)
	if err != nil {
		return nil, err
	}

	plain, hash, err := newRefreshToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	ts := now.Unix()
	rt := &model.RefreshToken{
		Model: model.Model{
			CreatedAt: ts,
			UpdatedAt: ts,
		},
		UserID:    user.Id,
		TokenHash: hash,
		ExpiresAt: now.Add(s.jwt.RefreshExpire()).Unix(),
	}
	if err := s.repos.RefreshToken.Create(ctx, rt); err != nil {
		return nil, err
	}

	return &AuthResult{
		Token:        access,
		RefreshToken: plain,
		ExpiresIn:    int64(s.jwt.AccessExpire() / time.Second),
		User:         user,
	}, nil
}

func newRefreshToken() (plain, hash string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	plain = hex.EncodeToString(raw)
	return plain, hashRefreshToken(plain), nil
}

func hashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
