package cache

import (
	"context"

	"github.com/rpstvs/social/internal/store"
	"github.com/stretchr/testify/mock"
)

func NewMockCache() Storage {
	return Storage{
		Users: &MockCacheStorage{},
	}
}

type MockCacheStorage struct {
	mock.Mock
}

func (m *MockCacheStorage) Get(ctx context.Context, id int64) (*store.User, error) {
	args := m.Called(id)
	return &store.User{}, args.Error(1)
}
func (m *MockCacheStorage) Set(context.Context, *store.User) error {
	return nil
}
