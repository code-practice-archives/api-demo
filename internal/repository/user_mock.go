package repository

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/code-practice-archives/api-demo/internal/model"
	"gorm.io/gorm"
)

// MockUserStore 内存版 UserStore，供单测替换真实数据库。
type MockUserStore struct {
	mu     sync.Mutex
	seq    atomic.Int64
	byID   map[int64]*model.User
	byName map[string]int64
}

func NewMockUserStore() *MockUserStore {
	return &MockUserStore{
		byID:   make(map[int64]*model.User),
		byName: make(map[string]int64),
	}
}

func (s *MockUserStore) Create(_ context.Context, user *model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byName[user.Username]; ok {
		return gorm.ErrDuplicatedKey
	}

	id := s.seq.Add(1)
	cp := *user
	cp.Id = id
	s.byID[id] = &cp
	s.byName[user.Username] = id
	user.Id = id
	return nil
}

func (s *MockUserStore) FindByUsername(_ context.Context, username string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := s.byName[username]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *s.byID[id]
	return &cp, nil
}

func (s *MockUserStore) FindByID(_ context.Context, id int64) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.byID[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *u
	return &cp, nil
}

func (s *MockUserStore) ExistsByUsername(_ context.Context, username string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.byName[username]
	return ok, nil
}
