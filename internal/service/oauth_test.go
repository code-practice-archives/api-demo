package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/code-practice-archives/api-demo/internal/model"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/code-practice-archives/api-demo/internal/pkg/loginjail"
	"github.com/code-practice-archives/api-demo/internal/pkg/oauth"
	"github.com/code-practice-archives/api-demo/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const (
	testRedirectURI = "http://localhost:3000/callback"
	testPublicClient = "demo-public"
	testConfClient   = "demo-confidential"
	testConfSecret   = "super-secret"
)

func pkcePair(t *testing.T) (verifier, challenge string) {
	t.Helper()
	verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge
}

func newTestOAuthService(t *testing.T) (*OAuthService, *AuthService) {
	t.Helper()

	clients := repository.NewMockOAuthClientStore()
	clients.Seed(&model.OAuthClient{
		ClientID:     testPublicClient,
		Name:         "Demo Public",
		RedirectURIs: `["http://localhost:3000/callback"]`,
	})

	hash, err := bcrypt.GenerateFromPassword([]byte(testConfSecret), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash secret: %v", err)
	}
	clients.Seed(&model.OAuthClient{
		ClientID:         testConfClient,
		ClientSecretHash: string(hash),
		Name:             "Demo Confidential",
		RedirectURIs:     `["http://localhost:3000/callback"]`,
	})

	repos := &repository.Repositories{
		User:          repository.NewMockUserStore(),
		RefreshToken:  repository.NewMockRefreshTokenStore(),
		OAuthClient:   clients,
		OAuthAuthCode: repository.NewMockOAuthAuthorizationCodeStore(),
	}
	jwtMgr := jwtx.NewManager("test-secret", time.Hour, 7*24*time.Hour)
	jail := loginjail.NewTestJail(t, 5, 15*time.Minute)
	auth := NewAuthService(repos, jwtMgr, jail, logger.Nop())
	oauthSvc := NewOAuthService(repos, jwtMgr, oauth.Config{CodeTTLMinutes: 10}, logger.Nop())
	return oauthSvc, auth
}

func seedUser(t *testing.T, auth *AuthService) *model.User {
	t.Helper()
	reg, err := auth.Register(context.Background(), RegisterInput{Username: "alice", Password: "secret123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	return reg.User
}

func authorizeOK(t *testing.T, svc *OAuthService, userID int64, clientID, challenge string) *AuthorizeResult {
	t.Helper()
	res, err := svc.Authorize(context.Background(), AuthorizeInput{
		UserID:              userID,
		ClientID:            clientID,
		RedirectURI:         testRedirectURI,
		ResponseType:        "code",
		State:               "xyz",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
		Scope:               "read",
	})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	return res
}

func TestOAuthService_AuthorizeAndExchange_PublicClient(t *testing.T) {
	svc, auth := newTestOAuthService(t)
	user := seedUser(t, auth)
	verifier, challenge := pkcePair(t)

	authz := authorizeOK(t, svc, user.Id, testPublicClient, challenge)

	tok, err := svc.Token(context.Background(), TokenInput{
		GrantType:    "authorization_code",
		Code:         authz.Code,
		RedirectURI:  testRedirectURI,
		ClientID:     testPublicClient,
		CodeVerifier: verifier,
	})
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	if tok.AccessToken == "" || tok.RefreshToken == "" || tok.TokenType != "Bearer" {
		t.Fatalf("unexpected token result: %+v", tok)
	}
	if tok.Scope != "read" {
		t.Fatalf("scope = %q, want read", tok.Scope)
	}
	if tok.ExpiresIn != int64(time.Hour/time.Second) {
		t.Fatalf("expires_in = %d", tok.ExpiresIn)
	}
}

func TestOAuthService_Authorize(t *testing.T) {
	_, challenge := pkcePair(t)

	tests := []struct {
		name    string
		mutate  func(in *AuthorizeInput)
		wantErr string
	}{
		{
			name: "success",
		},
		{
			name: "unknown client",
			mutate: func(in *AuthorizeInput) {
				in.ClientID = "no-such-client"
			},
			wantErr: "invalid_client",
		},
		{
			name: "redirect mismatch",
			mutate: func(in *AuthorizeInput) {
				in.RedirectURI = "http://evil.example/callback"
			},
			wantErr: "invalid_request",
		},
		{
			name: "unsupported response_type",
			mutate: func(in *AuthorizeInput) {
				in.ResponseType = "token"
			},
			wantErr: "unsupported_response_type",
		},
		{
			name: "plain challenge method rejected",
			mutate: func(in *AuthorizeInput) {
				in.CodeChallengeMethod = "plain"
			},
			wantErr: "invalid_request",
		},
		{
			name: "missing code_challenge",
			mutate: func(in *AuthorizeInput) {
				in.CodeChallenge = ""
			},
			wantErr: "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, auth := newTestOAuthService(t)
			user := seedUser(t, auth)

			in := AuthorizeInput{
				UserID:              user.Id,
				ClientID:            testPublicClient,
				RedirectURI:         testRedirectURI,
				ResponseType:        "code",
				State:               "s",
				CodeChallenge:       challenge,
				CodeChallengeMethod: "S256",
			}
			if tt.mutate != nil {
				tt.mutate(&in)
			}

			_, err := svc.Authorize(context.Background(), in)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				return
			}
			assertOAuthErr(t, err, tt.wantErr)
		})
	}
}

func TestOAuthService_ExchangeCode(t *testing.T) {
	verifier, challenge := pkcePair(t)

	tests := []struct {
		name    string
		setup   func(t *testing.T, svc *OAuthService, userID int64, code string) TokenInput
		wantErr string
	}{
		{
			name: "success public",
			setup: func(t *testing.T, svc *OAuthService, userID int64, code string) TokenInput {
				return TokenInput{
					GrantType:    "authorization_code",
					Code:         code,
					RedirectURI:  testRedirectURI,
					ClientID:     testPublicClient,
					CodeVerifier: verifier,
				}
			},
		},
		{
			name: "success confidential",
			setup: func(t *testing.T, svc *OAuthService, userID int64, code string) TokenInput {
				authz := authorizeOK(t, svc, userID, testConfClient, challenge)
				return TokenInput{
					GrantType:    "authorization_code",
					Code:         authz.Code,
					RedirectURI:  testRedirectURI,
					ClientID:     testConfClient,
					ClientSecret: testConfSecret,
					CodeVerifier: verifier,
				}
			},
		},
		{
			name: "bad code_verifier",
			setup: func(t *testing.T, svc *OAuthService, userID int64, code string) TokenInput {
				return TokenInput{
					GrantType:    "authorization_code",
					Code:         code,
					RedirectURI:  testRedirectURI,
					ClientID:     testPublicClient,
					CodeVerifier: "wrong-verifier-value-xxxxxxxxxxxxxxxxxxx",
				}
			},
			wantErr: "invalid_grant",
		},
		{
			name: "redirect mismatch",
			setup: func(t *testing.T, svc *OAuthService, userID int64, code string) TokenInput {
				return TokenInput{
					GrantType:    "authorization_code",
					Code:         code,
					RedirectURI:  "http://localhost:3000/other",
					ClientID:     testPublicClient,
					CodeVerifier: verifier,
				}
			},
			wantErr: "invalid_grant",
		},
		{
			name: "code replay",
			setup: func(t *testing.T, svc *OAuthService, userID int64, code string) TokenInput {
				in := TokenInput{
					GrantType:    "authorization_code",
					Code:         code,
					RedirectURI:  testRedirectURI,
					ClientID:     testPublicClient,
					CodeVerifier: verifier,
				}
				if _, err := svc.Token(context.Background(), in); err != nil {
					t.Fatalf("first exchange: %v", err)
				}
				return in
			},
			wantErr: "invalid_grant",
		},
		{
			name: "confidential missing secret",
			setup: func(t *testing.T, svc *OAuthService, userID int64, code string) TokenInput {
				authz := authorizeOK(t, svc, userID, testConfClient, challenge)
				return TokenInput{
					GrantType:    "authorization_code",
					Code:         authz.Code,
					RedirectURI:  testRedirectURI,
					ClientID:     testConfClient,
					CodeVerifier: verifier,
				}
			},
			wantErr: "invalid_client",
		},
		{
			name: "public client must not send secret",
			setup: func(t *testing.T, svc *OAuthService, userID int64, code string) TokenInput {
				return TokenInput{
					GrantType:    "authorization_code",
					Code:         code,
					RedirectURI:  testRedirectURI,
					ClientID:     testPublicClient,
					ClientSecret: "should-not-send",
					CodeVerifier: verifier,
				}
			},
			wantErr: "invalid_client",
		},
		{
			name: "unsupported grant",
			setup: func(t *testing.T, svc *OAuthService, userID int64, code string) TokenInput {
				return TokenInput{GrantType: "password", ClientID: testPublicClient}
			},
			wantErr: "unsupported_grant_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, auth := newTestOAuthService(t)
			user := seedUser(t, auth)
			authz := authorizeOK(t, svc, user.Id, testPublicClient, challenge)

			in := tt.setup(t, svc, user.Id, authz.Code)
			_, err := svc.Token(context.Background(), in)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				return
			}
			assertOAuthErr(t, err, tt.wantErr)
		})
	}
}

func TestOAuthService_RefreshBoundToClient(t *testing.T) {
	svc, auth := newTestOAuthService(t)
	user := seedUser(t, auth)
	verifier, challenge := pkcePair(t)

	authz := authorizeOK(t, svc, user.Id, testPublicClient, challenge)
	tok, err := svc.Token(context.Background(), TokenInput{
		GrantType:    "authorization_code",
		Code:         authz.Code,
		RedirectURI:  testRedirectURI,
		ClientID:     testPublicClient,
		CodeVerifier: verifier,
	})
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}

	// 正确 client 可 refresh
	refreshed, err := svc.Token(context.Background(), TokenInput{
		GrantType:    "refresh_token",
		RefreshToken: tok.RefreshToken,
		ClientID:     testPublicClient,
	})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Fatal("expected new tokens")
	}
	if refreshed.RefreshToken == tok.RefreshToken {
		t.Fatal("refresh token should rotate")
	}

	// 旧 refresh 不可再用
	_, err = svc.Token(context.Background(), TokenInput{
		GrantType:    "refresh_token",
		RefreshToken: tok.RefreshToken,
		ClientID:     testPublicClient,
	})
	assertOAuthErr(t, err, "invalid_grant")

	// 错误 client 不能用
	authz2 := authorizeOK(t, svc, user.Id, testPublicClient, challenge)
	tok2, err := svc.Token(context.Background(), TokenInput{
		GrantType:    "authorization_code",
		Code:         authz2.Code,
		RedirectURI:  testRedirectURI,
		ClientID:     testPublicClient,
		CodeVerifier: verifier,
	})
	if err != nil {
		t.Fatalf("exchange2: %v", err)
	}
	_, err = svc.Token(context.Background(), TokenInput{
		GrantType:    "refresh_token",
		RefreshToken: tok2.RefreshToken,
		ClientID:     testConfClient,
		ClientSecret: testConfSecret,
	})
	assertOAuthErr(t, err, "invalid_grant")
}

func TestAuthRefresh_RejectsOAuthRefreshToken(t *testing.T) {
	svc, auth := newTestOAuthService(t)
	user := seedUser(t, auth)
	verifier, challenge := pkcePair(t)

	authz := authorizeOK(t, svc, user.Id, testPublicClient, challenge)
	tok, err := svc.Token(context.Background(), TokenInput{
		GrantType:    "authorization_code",
		Code:         authz.Code,
		RedirectURI:  testRedirectURI,
		ClientID:     testPublicClient,
		CodeVerifier: verifier,
	})
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}

	_, err = auth.Refresh(context.Background(), RefreshInput{RefreshToken: tok.RefreshToken})
	if err == nil {
		t.Fatal("expected first-party refresh to reject oauth refresh token")
	}
}

func assertOAuthErr(t *testing.T, err error, wantCode string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %s, got nil", wantCode)
	}
	var oe *oauth.Error
	if !errors.As(err, &oe) {
		t.Fatalf("expected *oauth.Error, got %T: %v", err, err)
	}
	if oe.Code != wantCode {
		t.Fatalf("error code = %q, want %q (%v)", oe.Code, wantCode, err)
	}
}
