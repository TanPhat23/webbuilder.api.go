package testutil

import (
	"context"

	"my-go-app/internal/models"
)

type MockElementCommentRepository struct {
	CreateElementCommentFn             func(ctx context.Context, comment *models.ElementComment) (*models.ElementComment, error)
	GetElementCommentByIDFn            func(ctx context.Context, id string) (*models.ElementComment, error)
	GetElementCommentsFn               func(ctx context.Context, elementID string, filter *models.ElementCommentFilter) ([]models.ElementComment, error)
	UpdateElementCommentFn             func(ctx context.Context, id string, updates map[string]any) (*models.ElementComment, error)
	DeleteElementCommentFn             func(ctx context.Context, id string) error
	GetElementCommentsByAuthorIDFn     func(ctx context.Context, authorID string, limit int, offset int) ([]models.ElementComment, error)
	CountElementCommentsFn             func(ctx context.Context, elementID string) (int64, error)
	ToggleResolvedStatusFn             func(ctx context.Context, id string) (*models.ElementComment, error)
	DeleteElementCommentsByElementIDFn func(ctx context.Context, elementID string) error
	GetElementCommentsByProjectIDFn    func(ctx context.Context, projectID string, limit int, offset int) ([]models.ElementComment, error)
	CountElementCommentsByProjectIDFn  func(ctx context.Context, projectID string) (int64, error)
}

func (m *MockElementCommentRepository) CreateElementComment(ctx context.Context, comment *models.ElementComment) (*models.ElementComment, error) {
	if m.CreateElementCommentFn != nil {
		return m.CreateElementCommentFn(ctx, comment)
	}
	return comment, nil
}

func (m *MockElementCommentRepository) GetElementCommentByID(ctx context.Context, id string) (*models.ElementComment, error) {
	if m.GetElementCommentByIDFn != nil {
		return m.GetElementCommentByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockElementCommentRepository) GetElementComments(ctx context.Context, elementID string, filter *models.ElementCommentFilter) ([]models.ElementComment, error) {
	if m.GetElementCommentsFn != nil {
		return m.GetElementCommentsFn(ctx, elementID, filter)
	}
	return []models.ElementComment{}, nil
}

func (m *MockElementCommentRepository) UpdateElementComment(ctx context.Context, id string, updates map[string]any) (*models.ElementComment, error) {
	if m.UpdateElementCommentFn != nil {
		return m.UpdateElementCommentFn(ctx, id, updates)
	}
	return nil, nil
}

func (m *MockElementCommentRepository) DeleteElementComment(ctx context.Context, id string) error {
	if m.DeleteElementCommentFn != nil {
		return m.DeleteElementCommentFn(ctx, id)
	}
	return nil
}

func (m *MockElementCommentRepository) GetElementCommentsByAuthorID(ctx context.Context, authorID string, limit int, offset int) ([]models.ElementComment, error) {
	if m.GetElementCommentsByAuthorIDFn != nil {
		return m.GetElementCommentsByAuthorIDFn(ctx, authorID, limit, offset)
	}
	return []models.ElementComment{}, nil
}

func (m *MockElementCommentRepository) CountElementComments(ctx context.Context, elementID string) (int64, error) {
	if m.CountElementCommentsFn != nil {
		return m.CountElementCommentsFn(ctx, elementID)
	}
	return 0, nil
}

func (m *MockElementCommentRepository) ToggleResolvedStatus(ctx context.Context, id string) (*models.ElementComment, error) {
	if m.ToggleResolvedStatusFn != nil {
		return m.ToggleResolvedStatusFn(ctx, id)
	}
	return nil, nil
}

func (m *MockElementCommentRepository) DeleteElementCommentsByElementID(ctx context.Context, elementID string) error {
	if m.DeleteElementCommentsByElementIDFn != nil {
		return m.DeleteElementCommentsByElementIDFn(ctx, elementID)
	}
	return nil
}

func (m *MockElementCommentRepository) GetElementCommentsByProjectID(ctx context.Context, projectID string, limit int, offset int) ([]models.ElementComment, error) {
	if m.GetElementCommentsByProjectIDFn != nil {
		return m.GetElementCommentsByProjectIDFn(ctx, projectID, limit, offset)
	}
	return []models.ElementComment{}, nil
}

func (m *MockElementCommentRepository) CountElementCommentsByProjectID(ctx context.Context, projectID string) (int64, error) {
	if m.CountElementCommentsByProjectIDFn != nil {
		return m.CountElementCommentsByProjectIDFn(ctx, projectID)
	}
	return 0, nil
}