package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/code-practice-archives/api-demo/internal/pkg/database"
	"github.com/code-practice-archives/api-demo/internal/pkg/jwtx"
	"github.com/code-practice-archives/api-demo/internal/repository"
)

func newTestAuthService(t *testing.T) *AuthService {
	t.Helper()

	db, err := database.OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	users := repository.NewUserRepository(db)
	jwtMgr := jwtx.NewManager("test-secret", time.Hour)
	return NewAuthService(users, jwtMgr)
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
			wantErr: ErrInvalidInput,
		},
		{
			name:    "password too short",
			in:      RegisterInput{Username: "bob", Password: "123"},
			wantErr: ErrInvalidInput,
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
			wantErr: ErrUsernameTaken,
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
			wantErr: ErrInvalidCredentials,
		},
		{
			name:    "user not found",
			in:      LoginInput{Username: "nobody", Password: "secret123"},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:    "invalid input",
			in:      LoginInput{Username: "ab", Password: "123"},
			wantErr: ErrInvalidInput,
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
