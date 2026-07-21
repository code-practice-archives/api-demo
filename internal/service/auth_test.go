package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/code-practice-archives/api-demo/internal/pkg/errcode"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/pkg/logger"
	"github.com/code-practice-archives/api-demo/internal/pkg/loginjail"
	"github.com/code-practice-archives/api-demo/internal/repository"
)

func newTestAuthService(t *testing.T) *AuthService {
	t.Helper()
	return newTestAuthServiceWithJail(t, loginjail.NewMemory(5, 15*time.Minute))
}

func newTestAuthServiceWithJail(t *testing.T, jail loginjail.Jail) *AuthService {
	t.Helper()

	repos := &repository.Repositories{
		User:         repository.NewMockUserStore(),
		RefreshToken: repository.NewMockRefreshTokenStore(),
	}
	jwtMgr := jwtx.NewManager("test-secret", time.Hour, 7*24*time.Hour)
	return NewAuthService(repos, jwtMgr, jail, logger.Nop())
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	svc := newTestAuthService(t)

	ctx := context.Background()
	reg, err := svc.Register(ctx, RegisterInput{Username: "alice", Password: "secret123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if reg.Token == "" || reg.RefreshToken == "" {
		t.Fatal("expected non-empty access and refresh tokens")
	}
	if reg.ExpiresIn != int64(time.Hour/time.Second) {
		t.Fatalf("expires_in = %d, want %d", reg.ExpiresIn, int64(time.Hour/time.Second))
	}
	if reg.User == nil || reg.User.Username != "alice" || reg.User.Id == 0 {
		t.Fatalf("unexpected user: %+v", reg.User)
	}
	raw, err := json.Marshal(reg.User)
	if err != nil {
		t.Fatalf("marshal user: %v", err)
	}
	if strings.Contains(string(raw), "password") || strings.Contains(string(raw), "$2") {
		t.Fatalf("password must not appear in json: %s", raw)
	}

	login, err := svc.Login(ctx, LoginInput{Username: "alice", Password: "secret123"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if login.Token == "" || login.RefreshToken == "" {
		t.Fatal("expected non-empty login access and refresh tokens")
	}
}

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, svc *AuthService)
		in      RegisterInput
		wantErr error
	}{
		{
			name: "success",
			in:   RegisterInput{Username: "bob", Password: "secret123"},
		},
		{
			name:    "username too short",
			in:      RegisterInput{Username: "ab", Password: "secret123"},
			wantErr: errcode.ErrInvalidArgument,
		},
		{
			name:    "password too short",
			in:      RegisterInput{Username: "bob", Password: "123"},
			wantErr: errcode.ErrInvalidArgument,
		},
		{
			name: "username taken",
			setup: func(t *testing.T, svc *AuthService) {
				t.Helper()
				if _, err := svc.Register(context.Background(), RegisterInput{Username: "bob", Password: "secret123"}); err != nil {
					t.Fatalf("seed user: %v", err)
				}
			},
			in:      RegisterInput{Username: "bob", Password: "secret123"},
			wantErr: errcode.ErrUsernameTaken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestAuthService(t)
			if tt.setup != nil {
				tt.setup(t, svc)
			}

			got, err := svc.Register(context.Background(), tt.in)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil || got.User == nil || got.Token == "" || got.RefreshToken == "" {
				t.Fatalf("unexpected result: %+v", got)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, svc *AuthService)
		in      LoginInput
		wantErr error
	}{
		{
			name: "success",
			setup: func(t *testing.T, svc *AuthService) {
				t.Helper()
				if _, err := svc.Register(context.Background(), RegisterInput{Username: "carol", Password: "secret123"}); err != nil {
					t.Fatalf("seed user: %v", err)
				}
			},
			in: LoginInput{Username: "carol", Password: "secret123"},
		},
		{
			name: "wrong password",
			setup: func(t *testing.T, svc *AuthService) {
				t.Helper()
				if _, err := svc.Register(context.Background(), RegisterInput{Username: "carol", Password: "secret123"}); err != nil {
					t.Fatalf("seed user: %v", err)
				}
			},
			in:      LoginInput{Username: "carol", Password: "wrongpass"},
			wantErr: errcode.ErrInvalidCredentials,
		},
		{
			name:    "user not found",
			in:      LoginInput{Username: "nobody", Password: "secret123"},
			wantErr: errcode.ErrInvalidCredentials,
		},
		{
			name:    "empty username",
			in:      LoginInput{Username: "", Password: "secret123"},
			wantErr: errcode.ErrInvalidArgument,
		},
		{
			name:    "empty password",
			in:      LoginInput{Username: "carol", Password: ""},
			wantErr: errcode.ErrInvalidArgument,
		},
		{
			// 短用户名/密码不是登录参数错误，走凭证校验失败
			name:    "short credentials treated as invalid credentials",
			in:      LoginInput{Username: "ab", Password: "123"},
			wantErr: errcode.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestAuthService(t)
			if tt.setup != nil {
				tt.setup(t, svc)
			}

			got, err := svc.Login(context.Background(), tt.in)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil || got.Token == "" || got.RefreshToken == "" {
				t.Fatalf("unexpected result: %+v", got)
			}
		})
	}
}

func TestAuthService_Refresh(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, svc *AuthService) RefreshInput
		wantErr error
	}{
		{
			name: "success rotates refresh token",
			setup: func(t *testing.T, svc *AuthService) RefreshInput {
				t.Helper()
				reg, err := svc.Register(context.Background(), RegisterInput{Username: "refresh_ok", Password: "secret123"})
				if err != nil {
					t.Fatalf("seed: %v", err)
				}
				return RefreshInput{RefreshToken: reg.RefreshToken}
			},
		},
		{
			name: "reuse after rotate fails",
			setup: func(t *testing.T, svc *AuthService) RefreshInput {
				t.Helper()
				reg, err := svc.Register(context.Background(), RegisterInput{Username: "refresh_reuse", Password: "secret123"})
				if err != nil {
					t.Fatalf("seed: %v", err)
				}
				if _, err := svc.Refresh(context.Background(), RefreshInput{RefreshToken: reg.RefreshToken}); err != nil {
					t.Fatalf("first refresh: %v", err)
				}
				return RefreshInput{RefreshToken: reg.RefreshToken}
			},
			wantErr: errcode.ErrInvalidRefreshToken,
		},
		{
			name: "unknown token",
			setup: func(t *testing.T, svc *AuthService) RefreshInput {
				t.Helper()
				return RefreshInput{RefreshToken: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"}
			},
			wantErr: errcode.ErrInvalidRefreshToken,
		},
		{
			name: "empty token",
			setup: func(t *testing.T, svc *AuthService) RefreshInput {
				t.Helper()
				return RefreshInput{RefreshToken: ""}
			},
			wantErr: errcode.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestAuthService(t)
			in := tt.setup(t, svc)

			got, err := svc.Refresh(context.Background(), in)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Token == "" || got.RefreshToken == "" || got.RefreshToken == in.RefreshToken {
				t.Fatalf("expected new token pair, got %+v", got)
			}
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	svc := newTestAuthService(t)
	ctx := context.Background()

	reg, err := svc.Register(ctx, RegisterInput{Username: "logout_user", Password: "secret123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if err := svc.Logout(ctx, LogoutInput{RefreshToken: reg.RefreshToken}); err != nil {
		t.Fatalf("logout: %v", err)
	}

	_, err = svc.Refresh(ctx, RefreshInput{RefreshToken: reg.RefreshToken})
	if !errors.Is(err, errcode.ErrInvalidRefreshToken) {
		t.Fatalf("refresh after logout: %v, want invalid refresh token", err)
	}

	// 幂等：再次 logout 不报错
	if err := svc.Logout(ctx, LogoutInput{RefreshToken: reg.RefreshToken}); err != nil {
		t.Fatalf("logout again: %v", err)
	}
}

func TestAuthService_LoginJail(t *testing.T) {
	ctx := context.Background()
	jail := loginjail.NewMemory(3, 15*time.Minute)
	svc := newTestAuthServiceWithJail(t, jail)

	if _, err := svc.Register(ctx, RegisterInput{Username: "dave", Password: "secret123"}); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	for i := 0; i < 2; i++ {
		_, err := svc.Login(ctx, LoginInput{Username: "dave", Password: "wrongpass"})
		if !errors.Is(err, errcode.ErrInvalidCredentials) {
			t.Fatalf("fail %d: error = %v, want invalid credentials", i+1, err)
		}
	}

	_, err := svc.Login(ctx, LoginInput{Username: "dave", Password: "wrongpass"})
	if !errors.Is(err, errcode.ErrAccountLocked) {
		t.Fatalf("3rd fail: error = %v, want account locked", err)
	}

	// 锁定期间即使密码正确也拒绝
	_, err = svc.Login(ctx, LoginInput{Username: "dave", Password: "secret123"})
	if !errors.Is(err, errcode.ErrAccountLocked) {
		t.Fatalf("locked login: error = %v, want account locked", err)
	}
}

func TestAuthService_LoginJailResetOnSuccess(t *testing.T) {
	ctx := context.Background()
	jail := loginjail.NewMemory(3, 15*time.Minute)
	svc := newTestAuthServiceWithJail(t, jail)

	if _, err := svc.Register(ctx, RegisterInput{Username: "erin", Password: "secret123"}); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	_, err := svc.Login(ctx, LoginInput{Username: "erin", Password: "wrongpass"})
	if !errors.Is(err, errcode.ErrInvalidCredentials) {
		t.Fatalf("fail: %v", err)
	}

	if _, err := svc.Login(ctx, LoginInput{Username: "erin", Password: "secret123"}); err != nil {
		t.Fatalf("success login should reset: %v", err)
	}

	// 重置后可再失败 maxRetries-1 次而不锁定
	for i := 0; i < 2; i++ {
		_, err := svc.Login(ctx, LoginInput{Username: "erin", Password: "wrongpass"})
		if !errors.Is(err, errcode.ErrInvalidCredentials) {
			t.Fatalf("fail %d after reset: error = %v, want invalid credentials", i+1, err)
		}
	}
}

func TestAuthService_Me(t *testing.T) {
	svc := newTestAuthService(t)
	ctx := context.Background()

	reg, err := svc.Register(ctx, RegisterInput{Username: "frank", Password: "secret123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	user, err := svc.Me(ctx, reg.User.Id)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	if user.Username != "frank" {
		t.Fatalf("username = %q, want frank", user.Username)
	}

	_, err = svc.Me(ctx, 99999)
	if !errors.Is(err, errcode.ErrUnauthorized) {
		t.Fatalf("error = %v, want unauthorized", err)
	}
}
