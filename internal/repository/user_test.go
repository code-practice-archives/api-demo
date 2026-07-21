package repository

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/code-practice-archives/api-demo/internal/model"
	"github.com/code-practice-archives/api-demo/internal/pkg/database"
	"gorm.io/gorm"
)

// 集成测试：需设置 TEST_MYSQL_DSN，例如
// TEST_MYSQL_DSN='root:root@tcp(127.0.0.1:3306)/api_demo_test?charset=utf8mb4&parseTime=True&loc=Local'
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN not set; skip repository integration tests")
	}

	db, err := database.Open(database.Config{DSN: dsn})
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}

	if err := db.Exec("DELETE FROM users").Error; err != nil {
		t.Fatalf("truncate users: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	return db
}

func newTestUserRepo(t *testing.T) *UserRepository {
	t.Helper()
	return NewUserRepository(openTestDB(t))
}

func TestUserRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *UserRepository)
		user    *model.User
		wantErr bool
	}{
		{
			name: "success",
			user: &model.User{Username: "alice", Password: "hash"},
		},
		{
			name: "duplicate username",
			setup: func(t *testing.T, repo *UserRepository) {
				t.Helper()
				if err := repo.Create(context.Background(), &model.User{
					Username: "alice",
					Password: "hash",
				}); err != nil {
					t.Fatalf("seed user: %v", err)
				}
			},
			user:    &model.User{Username: "alice", Password: "other"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestUserRepo(t)
			if tt.setup != nil {
				tt.setup(t, repo)
			}

			err := repo.Create(context.Background(), tt.user)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.user.Id == 0 {
				t.Fatal("expected non-zero id after create")
			}
		})
	}
}

func TestUserRepository_FindByUsername(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, repo *UserRepository)
		username string
		wantUser bool
		wantErr  error
		wantName string
	}{
		{
			name: "found",
			setup: func(t *testing.T, repo *UserRepository) {
				t.Helper()
				if err := repo.Create(context.Background(), &model.User{
					Username: "alice",
					Password: "hash",
				}); err != nil {
					t.Fatalf("seed user: %v", err)
				}
			},
			username: "alice",
			wantUser: true,
			wantName: "alice",
		},
		{
			name:     "not found",
			username: "nobody",
			wantErr:  gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestUserRepo(t)
			if tt.setup != nil {
				tt.setup(t, repo)
			}

			got, err := repo.FindByUsername(context.Background(), tt.username)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				if got != nil {
					t.Fatalf("expected nil user, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.wantUser || got == nil {
				t.Fatalf("unexpected user: %+v", got)
			}
			if got.Username != tt.wantName || got.Id == 0 || got.Password == "" {
				t.Fatalf("unexpected user fields: %+v", got)
			}
		})
	}
}

func TestUserRepository_ExistsByUsername(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, repo *UserRepository)
		username string
		want     bool
	}{
		{
			name: "exists",
			setup: func(t *testing.T, repo *UserRepository) {
				t.Helper()
				if err := repo.Create(context.Background(), &model.User{
					Username: "alice",
					Password: "hash",
				}); err != nil {
					t.Fatalf("seed user: %v", err)
				}
			},
			username: "alice",
			want:     true,
		},
		{
			name:     "not exists",
			username: "nobody",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestUserRepo(t)
			if tt.setup != nil {
				tt.setup(t, repo)
			}

			got, err := repo.ExistsByUsername(context.Background(), tt.username)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("exists = %v, want %v", got, tt.want)
			}
		})
	}
}
