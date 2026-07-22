package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/code-practice-archives/api-demo/internal/model"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/code-practice-archives/api-demo/internal/pkg/oauth"
	"github.com/code-practice-archives/api-demo/internal/repository"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	responseTypeCode       = "code"
	challengeMethodS256    = "S256"
	grantTypeAuthCode      = "authorization_code"
	grantTypeRefreshToken  = "refresh_token"
)

type OAuthService struct {
	repos   *repository.Repositories
	jwt     *jwtx.Manager
	codeTTL time.Duration
	log     *logger.Logger
}

func NewOAuthService(repos *repository.Repositories, jwt *jwtx.Manager, cfg oauth.Config, log *logger.Logger) *OAuthService {
	return &OAuthService{
		repos:   repos,
		jwt:     jwt,
		codeTTL: cfg.CodeTTL(),
		log:     log.Named("oauth"),
	}
}

type AuthorizeInput struct {
	UserID              int64
	ClientID            string
	RedirectURI         string
	ResponseType        string
	State               string
	Scope               string
	CodeChallenge       string
	CodeChallengeMethod string
}

type AuthorizeResult struct {
	Code        string
	State       string
	RedirectURI string
}

// Authorize 校验客户端与 PKCE 参数，签发一次性授权码（仅存哈希）。
func (s *OAuthService) Authorize(ctx context.Context, in AuthorizeInput) (*AuthorizeResult, error) {
	log := s.log.WithContext(ctx).With(
		zap.Int64("user_id", in.UserID),
		zap.String("client_id", in.ClientID),
	)

	in.ClientID = strings.TrimSpace(in.ClientID)
	in.RedirectURI = strings.TrimSpace(in.RedirectURI)
	in.ResponseType = strings.TrimSpace(in.ResponseType)
	in.CodeChallenge = strings.TrimSpace(in.CodeChallenge)
	in.CodeChallengeMethod = strings.TrimSpace(in.CodeChallengeMethod)
	in.Scope = strings.TrimSpace(in.Scope)

	if in.ClientID == "" {
		return nil, oauth.ErrInvalidRequest.WithDescription("client_id is required")
	}
	if in.RedirectURI == "" {
		return nil, oauth.ErrInvalidRequest.WithDescription("redirect_uri is required")
	}
	if in.ResponseType == "" {
		return nil, oauth.ErrInvalidRequest.WithDescription("response_type is required")
	}
	if in.ResponseType != responseTypeCode {
		return nil, oauth.ErrUnsupportedResponseType.WithDescription("only response_type=code is supported")
	}
	if in.CodeChallenge == "" {
		return nil, oauth.ErrInvalidRequest.WithDescription("code_challenge is required")
	}
	if in.CodeChallengeMethod == "" {
		return nil, oauth.ErrInvalidRequest.WithDescription("code_challenge_method is required")
	}
	if !strings.EqualFold(in.CodeChallengeMethod, challengeMethodS256) {
		return nil, oauth.ErrInvalidRequest.WithDescription("only code_challenge_method=S256 is supported")
	}

	client, err := s.repos.OAuthClient.FindByClientID(ctx, in.ClientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, oauth.ErrInvalidClient.WithDescription("unknown client_id")
		}
		log.Error("authorize find client failed", zap.Error(err))
		return nil, oauth.ErrServerError
	}

	if !redirectURIAllowed(client.RedirectURIs, in.RedirectURI) {
		return nil, oauth.ErrInvalidRequest.WithDescription("redirect_uri is not registered")
	}

	plain, hash, err := newRefreshToken()
	if err != nil {
		log.Error("authorize generate code failed", zap.Error(err))
		return nil, oauth.ErrServerError
	}

	now := time.Now()
	ts := now.Unix()
	row := &model.OAuthAuthorizationCode{
		Model: model.Model{
			CreatedAt: ts,
			UpdatedAt: ts,
		},
		CodeHash:            hash,
		ClientID:            client.ClientID,
		UserID:              in.UserID,
		RedirectURI:         in.RedirectURI,
		Scope:               in.Scope,
		CodeChallenge:       in.CodeChallenge,
		CodeChallengeMethod: challengeMethodS256,
		ExpiresAt:           now.Add(s.codeTTL).Unix(),
	}
	if err := s.repos.OAuthAuthCode.Create(ctx, row); err != nil {
		log.Error("authorize persist code failed", zap.Error(err))
		return nil, oauth.ErrServerError
	}

	log.Info("authorize success")
	return &AuthorizeResult{
		Code:        plain,
		State:       in.State,
		RedirectURI: in.RedirectURI,
	}, nil
}

type TokenInput struct {
	GrantType    string
	Code         string
	RedirectURI  string
	ClientID     string
	ClientSecret string
	CodeVerifier string
	RefreshToken string
}

type TokenResult struct {
	AccessToken  string
	TokenType    string
	ExpiresIn    int64
	RefreshToken string
	Scope        string
}

// Token 处理 authorization_code / refresh_token 授权。
func (s *OAuthService) Token(ctx context.Context, in TokenInput) (*TokenResult, error) {
	in.GrantType = strings.TrimSpace(in.GrantType)
	in.ClientID = strings.TrimSpace(in.ClientID)

	switch in.GrantType {
	case grantTypeAuthCode:
		return s.exchangeCode(ctx, in)
	case grantTypeRefreshToken:
		return s.refresh(ctx, in)
	case "":
		return nil, oauth.ErrInvalidRequest.WithDescription("grant_type is required")
	default:
		return nil, oauth.ErrUnsupportedGrantType
	}
}

func (s *OAuthService) exchangeCode(ctx context.Context, in TokenInput) (*TokenResult, error) {
	log := s.log.WithContext(ctx).With(zap.String("client_id", in.ClientID))

	in.Code = strings.TrimSpace(in.Code)
	in.RedirectURI = strings.TrimSpace(in.RedirectURI)
	in.CodeVerifier = strings.TrimSpace(in.CodeVerifier)
	in.ClientSecret = strings.TrimSpace(in.ClientSecret)

	if in.ClientID == "" {
		return nil, oauth.ErrInvalidRequest.WithDescription("client_id is required")
	}
	if in.Code == "" {
		return nil, oauth.ErrInvalidRequest.WithDescription("code is required")
	}
	if in.RedirectURI == "" {
		return nil, oauth.ErrInvalidRequest.WithDescription("redirect_uri is required")
	}
	if in.CodeVerifier == "" {
		return nil, oauth.ErrInvalidRequest.WithDescription("code_verifier is required")
	}

	client, err := s.authenticateClient(ctx, in.ClientID, in.ClientSecret)
	if err != nil {
		return nil, err
	}

	stored, err := s.repos.OAuthAuthCode.FindByCodeHash(ctx, hashRefreshToken(in.Code))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, oauth.ErrInvalidGrant.WithDescription("invalid authorization code")
		}
		log.Error("token find code failed", zap.Error(err))
		return nil, oauth.ErrServerError
	}

	now := time.Now().Unix()
	if stored.UsedAt != 0 || stored.ExpiresAt <= now {
		return nil, oauth.ErrInvalidGrant.WithDescription("authorization code expired or already used")
	}
	if stored.ClientID != client.ClientID {
		return nil, oauth.ErrInvalidGrant.WithDescription("authorization code was not issued to this client")
	}
	if stored.RedirectURI != in.RedirectURI {
		return nil, oauth.ErrInvalidGrant.WithDescription("redirect_uri mismatch")
	}
	if !verifyPKCE(in.CodeVerifier, stored.CodeChallenge) {
		return nil, oauth.ErrInvalidGrant.WithDescription("invalid code_verifier")
	}

	if err := s.repos.OAuthAuthCode.MarkUsed(ctx, stored.Id, now); err != nil {
		log.Error("token mark code used failed", zap.Error(err))
		return nil, oauth.ErrServerError
	}

	user, err := s.repos.User.FindByID(ctx, stored.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, oauth.ErrInvalidGrant.WithDescription("user not found")
		}
		log.Error("token find user failed", zap.Error(err), zap.Int64("user_id", stored.UserID))
		return nil, oauth.ErrServerError
	}

	result, err := s.issueOAuthTokens(ctx, user, client.ClientID, stored.Scope)
	if err != nil {
		log.Error("token issue failed", zap.Error(err), zap.Int64("user_id", user.Id))
		return nil, oauth.ErrServerError
	}

	log.Info("token authorization_code success", zap.Int64("user_id", user.Id))
	return result, nil
}

func (s *OAuthService) refresh(ctx context.Context, in TokenInput) (*TokenResult, error) {
	log := s.log.WithContext(ctx).With(zap.String("client_id", in.ClientID))

	in.RefreshToken = strings.TrimSpace(in.RefreshToken)
	in.ClientSecret = strings.TrimSpace(in.ClientSecret)

	if in.ClientID == "" {
		return nil, oauth.ErrInvalidRequest.WithDescription("client_id is required")
	}
	if in.RefreshToken == "" {
		return nil, oauth.ErrInvalidRequest.WithDescription("refresh_token is required")
	}

	client, err := s.authenticateClient(ctx, in.ClientID, in.ClientSecret)
	if err != nil {
		return nil, err
	}

	stored, err := s.repos.RefreshToken.FindByTokenHash(ctx, hashRefreshToken(in.RefreshToken))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, oauth.ErrInvalidGrant.WithDescription("invalid refresh token")
		}
		log.Error("refresh find token failed", zap.Error(err))
		return nil, oauth.ErrServerError
	}

	now := time.Now().Unix()
	if stored.RevokedAt != 0 || stored.ExpiresAt <= now {
		return nil, oauth.ErrInvalidGrant.WithDescription("refresh token expired or revoked")
	}
	if stored.ClientID == "" || stored.ClientID != client.ClientID {
		return nil, oauth.ErrInvalidGrant.WithDescription("refresh token was not issued to this client")
	}

	user, err := s.repos.User.FindByID(ctx, stored.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, oauth.ErrInvalidGrant.WithDescription("user not found")
		}
		log.Error("refresh find user failed", zap.Error(err), zap.Int64("user_id", stored.UserID))
		return nil, oauth.ErrServerError
	}

	if err := s.repos.RefreshToken.Revoke(ctx, stored.Id, now); err != nil {
		log.Error("refresh revoke old token failed", zap.Error(err))
		return nil, oauth.ErrServerError
	}

	result, err := s.issueOAuthTokens(ctx, user, client.ClientID, "")
	if err != nil {
		log.Error("refresh issue tokens failed", zap.Error(err), zap.Int64("user_id", user.Id))
		return nil, oauth.ErrServerError
	}

	log.Info("token refresh_token success", zap.Int64("user_id", user.Id))
	return result, nil
}

func (s *OAuthService) authenticateClient(ctx context.Context, clientID, clientSecret string) (*model.OAuthClient, error) {
	client, err := s.repos.OAuthClient.FindByClientID(ctx, clientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, oauth.ErrInvalidClient.WithDescription("unknown client_id")
		}
		s.log.WithContext(ctx).Error("authenticate client find failed", zap.Error(err), zap.String("client_id", clientID))
		return nil, oauth.ErrServerError
	}

	if client.IsPublic() {
		if clientSecret != "" {
			return nil, oauth.ErrInvalidClient.WithDescription("public client must not send client_secret")
		}
		return client, nil
	}

	if clientSecret == "" {
		return nil, oauth.ErrInvalidClient.WithDescription("client_secret is required")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(client.ClientSecretHash), []byte(clientSecret)); err != nil {
		return nil, oauth.ErrInvalidClient.WithDescription("invalid client credentials")
	}
	return client, nil
}

func (s *OAuthService) issueOAuthTokens(ctx context.Context, user *model.User, clientID, scope string) (*TokenResult, error) {
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
		ClientID:  clientID,
		TokenHash: hash,
		ExpiresAt: now.Add(s.jwt.RefreshExpire()).Unix(),
	}
	if err := s.repos.RefreshToken.Create(ctx, rt); err != nil {
		return nil, err
	}

	return &TokenResult{
		AccessToken:  access,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwt.AccessExpire() / time.Second),
		RefreshToken: plain,
		Scope:        scope,
	}, nil
}

func redirectURIAllowed(rawJSON, redirectURI string) bool {
	var uris []string
	if err := json.Unmarshal([]byte(rawJSON), &uris); err != nil {
		return false
	}
	for _, u := range uris {
		if u == redirectURI {
			return true
		}
	}
	return false
}

// verifyPKCE 校验 S256：BASE64URL(SHA256(verifier)) == challenge。
func verifyPKCE(verifier, challenge string) bool {
	if verifier == "" || challenge == "" {
		return false
	}
	sum := sha256.Sum256([]byte(verifier))
	encoded := base64.RawURLEncoding.EncodeToString(sum[:])
	return encoded == challenge
}
