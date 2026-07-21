package repository

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/code-practice-archives/api-demo/internal/model"
	"gorm.io/gorm"
)

// MockRefreshTokenStore 内存版 RefreshTokenStore，供单测替换真实数据库。
type MockRefreshTokenStore struct {
	mu     sync.Mutex
	seq    atomic.Int64
	byID   map[int64]*model.RefreshToken
	byHash map[string]int64
}

func NewMockRefreshTokenStore() *MockRefreshTokenStore {
	return &MockRefreshTokenStore{
		byID:   make(map[int64]*model.RefreshToken),
		byHash: make(map[string]int64),
	}
}

func (s *MockRefreshTokenStore) Create(_ context.Context, token *model.RefreshToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byHash[token.TokenHash]; ok {
		return gorm.ErrDuplicatedKey
	}

	id := s.seq.Add(1)
	cp := *token
	cp.Id = id
	s.byID[id] = &cp
	s.byHash[token.TokenHash] = id
	token.Id = id
	return nil
}

func (s *MockRefreshTokenStore) FindByTokenHash(_ context.Context, hash string) (*model.RefreshToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := s.byHash[hash]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *s.byID[id]
	return &cp, nil
}

func (s *MockRefreshTokenStore) Revoke(_ context.Context, id int64, revokedAt int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tok, ok := s.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if tok.RevokedAt != 0 {
		return nil
	}
	tok.RevokedAt = revokedAt
	tok.UpdatedAt = revokedAt
	return nil
}
