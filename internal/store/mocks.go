package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/stretchr/testify/mock"
)

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	args := m.Called(user.ID)
	return args.Error(1)
}

func (m *MockUserStore) GetById(ctx context.Context, id int64) (*User, error) {
	return &User{}, nil
}
func (m *MockUserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	return &User{}, nil
}
func (m *MockUserStore) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return nil
}
func (m *MockUserStore) Activate(ctx context.Context, token string) error {
	return nil
}
