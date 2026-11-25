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

func (m *MockUserRepo) SetGetByID(fn func(context.Context, string) (*models.User, error)) *MockUserRepo {
	m.Set("GetByID", fn)
	return m
}

func (m *MockUserRepo) SetGetByEmail(fn func(context.Context, string) (*models.User, error)) *MockUserRepo {
	m.Set("GetByEmail", fn)
	return m
}

func (m *MockUserRepo) SetGetByUsername(fn func(context.Context, string) (*models.User, error)) *MockUserRepo {
	m.Set("GetByUsername", fn)
	return m
}

func (m *MockUserRepo) SetSearch(fn func(context.Context, string) ([]models.User, error)) *MockUserRepo {
	m.Set("Search", fn)
	return m
}

func (m *MockUserRepo) GetByID(ctx context.Context, userID string) (*models.User, error) {
	if fn := m.Get("GetByID"); fn != nil {
		return fn.(func(context.Context, string) (*models.User, error))(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if fn := m.Get("GetByEmail"); fn != nil {
		return fn.(func(context.Context, string) (*models.User, error))(ctx, email)
	}
	return nil, nil
}

func (m *MockUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if fn := m.Get("GetByUsername"); fn != nil {
		return fn.(func(context.Context, string) (*models.User, error))(ctx, username)
	}
	return nil, nil
}

func (m *MockUserRepo) Search(ctx context.Context, query string) ([]models.User, error) {
	if fn := m.Get("Search"); fn != nil {
		return fn.(func(context.Context, string) ([]models.User, error))(ctx, query)
	}
	return nil, nil
}
