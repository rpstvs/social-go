package cache

import (
	"context"

	"github.com/rpstvs/social/internal/store"
)

func NewMockCache() Storage {
	return Storage{
		Users: &MockCacheStorage{},
	}
}

type MockCacheStorage struct {
}

func (m *MockCacheStorage) Get(context.Context, int64) (*store.User, error) {
	return &store.User{}, nil
}
func (m *MockCacheStorage) Set(context.Context, *store.User) error {
	return nil
}
