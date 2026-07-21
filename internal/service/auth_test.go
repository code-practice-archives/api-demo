package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/code-practice-archives/api-demo/internal/pkg/database"
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

	db, err := database.OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	repos := repository.New(db)
	jwtMgr := jwtx.NewManager("test-secret", time.Hour)
	return NewAuthService(repos, jwtMgr, jail, logger.Nop())
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	svc := newTestAuthService(t)

	ctx := context.Background()
	reg, err := svc.Register(ctx, RegisterInput{Username: "alice", Password: "secret123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if reg.Token == "" {
		t.Fatal("expected non-empty token")
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
	if login.Token == "" {
		t.Fatal("expected non-empty login token")
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
			if got == nil || got.User == nil || got.Token == "" {
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
			name:    "invalid input",
			in:      LoginInput{Username: "ab", Password: "123"},
			wantErr: errcode.ErrInvalidArgument,
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
			if got == nil || got.Token == "" {
				t.Fatalf("unexpected result: %+v", got)
			}
		})
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
