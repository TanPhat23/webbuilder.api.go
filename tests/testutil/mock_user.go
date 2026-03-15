package testutil

import (
	"context"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MockUserRepository struct {
	GetUserByIDFn       func(ctx context.Context, userID string) (*models.User, error)
	GetUserByEmailFn    func(ctx context.Context, email string) (*models.User, error)
	GetUserByUsernameFn func(ctx context.Context, username string) (*models.User, error)
	SearchUsersFn       func(ctx context.Context, query string) ([]models.User, error)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	if m.GetUserByIDFn != nil {
		return m.GetUserByIDFn(ctx, userID)
	}
	return nil, repositories.ErrUserNotFound
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.GetUserByEmailFn != nil {
		return m.GetUserByEmailFn(ctx, email)
	}
	return nil, repositories.ErrUserNotFound
}

func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	if m.GetUserByUsernameFn != nil {
		return m.GetUserByUsernameFn(ctx, username)
	}
	return nil, repositories.ErrUserNotFound
}

func (m *MockUserRepository) SearchUsers(ctx context.Context, query string) ([]models.User, error) {
	if m.SearchUsersFn != nil {
		return m.SearchUsersFn(ctx, query)
	}
	return []models.User{}, nil
}