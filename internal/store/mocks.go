package store

import "context"

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct {
}

func (m *MockUserStore) Create(ctx context.Context, user *User) error {
	return nil
}

func (m *MockUserStore) GetById(ctx context.Context, id int64) (*User, error) {
	return &User{}, nil
}
func (m *MockUserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	return &User{}, nil
}
func (m *MockUserStore) CreateAndInvite(ctx context.Context, user *User, token string) error {
	return nil
}
func (m *MockUserStore) Activate(ctx context.Context, token string) error {
	return nil
}
