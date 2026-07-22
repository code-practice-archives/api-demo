package repository

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/code-practice-archives/api-demo/internal/model"
	"gorm.io/gorm"
)

// MockOAuthAuthorizationCodeStore 内存版 OAuthAuthorizationCodeStore，供单测使用。
type MockOAuthAuthorizationCodeStore struct {
	mu     sync.Mutex
	seq    atomic.Int64
	byID   map[int64]*model.OAuthAuthorizationCode
	byHash map[string]int64
}

func NewMockOAuthAuthorizationCodeStore() *MockOAuthAuthorizationCodeStore {
	return &MockOAuthAuthorizationCodeStore{
		byID:   make(map[int64]*model.OAuthAuthorizationCode),
		byHash: make(map[string]int64),
	}
}

func (s *MockOAuthAuthorizationCodeStore) Create(_ context.Context, code *model.OAuthAuthorizationCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byHash[code.CodeHash]; ok {
		return gorm.ErrDuplicatedKey
	}

	id := s.seq.Add(1)
	cp := *code
	cp.Id = id
	s.byID[id] = &cp
	s.byHash[code.CodeHash] = id
	code.Id = id
	return nil
}

func (s *MockOAuthAuthorizationCodeStore) FindByCodeHash(_ context.Context, hash string) (*model.OAuthAuthorizationCode, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := s.byHash[hash]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *s.byID[id]
	return &cp, nil
}

func (s *MockOAuthAuthorizationCodeStore) MarkUsed(_ context.Context, id int64, usedAt int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	code, ok := s.byID[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	if code.UsedAt != 0 {
		return nil
	}
	code.UsedAt = usedAt
	return nil
}
