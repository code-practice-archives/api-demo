package repository

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/code-practice-archives/api-demo/internal/model"
	"gorm.io/gorm"
)

// MockOAuthClientStore 内存版 OAuthClientStore，供单测使用。
type MockOAuthClientStore struct {
	mu        sync.Mutex
	seq       atomic.Int64
	byClientID map[string]*model.OAuthClient
}

func NewMockOAuthClientStore() *MockOAuthClientStore {
	return &MockOAuthClientStore{
		byClientID: make(map[string]*model.OAuthClient),
	}
}

func (s *MockOAuthClientStore) Seed(client *model.OAuthClient) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := *client
	if cp.Id == 0 {
		cp.Id = s.seq.Add(1)
	}
	s.byClientID[cp.ClientID] = &cp
}

func (s *MockOAuthClientStore) FindByClientID(_ context.Context, clientID string) (*model.OAuthClient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.byClientID[clientID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *c
	return &cp, nil
}
