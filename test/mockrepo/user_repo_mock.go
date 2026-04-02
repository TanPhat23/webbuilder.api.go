package test

import (
	"context"
	"my-go-app/internal/models"
)

// MockUserRepo implements UserRepositoryInterface for testing.
type MockUserRepo struct {
	*GenericMock
}

// NewMockUserRepo creates a new MockUserRepo instance.
func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{GenericMock: &GenericMock{funcs: make(map[string]any)}}
}

func (m *MockUserRepo) SetGetUserByID(fn func(context.Context, string) (*models.User, error)) *MockUserRepo {
	m.Set("GetUserByID", fn)
	return m
}

func (m *MockUserRepo) SetGetUserByEmail(fn func(context.Context, string) (*models.User, error)) *MockUserRepo {
	m.Set("GetUserByEmail", fn)
	return m
}

func (m *MockUserRepo) SetGetUserByUsername(fn func(context.Context, string) (*models.User, error)) *MockUserRepo {
	m.Set("GetUserByUsername", fn)
	return m
}

func (m *MockUserRepo) SetSearchUsers(fn func(context.Context, string) ([]models.User, error)) *MockUserRepo {
	m.Set("SearchUsers", fn)
	return m
}

func (m *MockUserRepo) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	if fn := m.Get("GetUserByID"); fn != nil {
		return fn.(func(context.Context, string) (*models.User, error))(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if fn := m.Get("GetUserByEmail"); fn != nil {
		return fn.(func(context.Context, string) (*models.User, error))(ctx, email)
	}
	return nil, nil
}

func (m *MockUserRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if fn := m.Get("GetUserByUsername"); fn != nil {
		return fn.(func(context.Context, string) (*models.User, error))(ctx, username)
	}
	return nil, nil
}

func (m *MockUserRepo) SearchUsers(ctx context.Context, query string) ([]models.User, error) {
	if fn := m.Get("SearchUsers"); fn != nil {
		return fn.(func(context.Context, string) ([]models.User, error))(ctx, query)
	}
	return nil, nil
}