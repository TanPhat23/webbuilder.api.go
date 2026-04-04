package test

import (
	"context"
	"my-go-app/internal/models"
)

type MockImageRepo struct {
	*GenericMock
}

func NewMockImageRepo() *MockImageRepo {
	return &MockImageRepo{GenericMock: &GenericMock{funcs: make(map[string]any)}}
}

func (m *MockImageRepo) SetCreateImage(fn func(context.Context, models.Image) (*models.Image, error)) *MockImageRepo {
	m.Set("CreateImage", fn)
	return m
}

func (m *MockImageRepo) SetGetImagesByUserID(fn func(context.Context, string) ([]models.Image, error)) *MockImageRepo {
	m.Set("GetImagesByUserID", fn)
	return m
}

func (m *MockImageRepo) SetGetImageByID(fn func(context.Context, string, string) (*models.Image, error)) *MockImageRepo {
	m.Set("GetImageByID", fn)
	return m
}

func (m *MockImageRepo) SetDeleteImage(fn func(context.Context, string, string) error) *MockImageRepo {
	m.Set("DeleteImage", fn)
	return m
}

func (m *MockImageRepo) SetSoftDeleteImage(fn func(context.Context, string, string) error) *MockImageRepo {
	m.Set("SoftDeleteImage", fn)
	return m
}

func (m *MockImageRepo) SetGetAllImages(fn func(context.Context, int, int) ([]models.Image, error)) *MockImageRepo {
	m.Set("GetAllImages", fn)
	return m
}

func (m *MockImageRepo) CreateImage(ctx context.Context, image models.Image) (*models.Image, error) {
	if fn := m.Get("CreateImage"); fn != nil {
		return fn.(func(context.Context, models.Image) (*models.Image, error))(ctx, image)
	}
	return &image, nil
}

func (m *MockImageRepo) GetImagesByUserID(ctx context.Context, userID string) ([]models.Image, error) {
	if fn := m.Get("GetImagesByUserID"); fn != nil {
		return fn.(func(context.Context, string) ([]models.Image, error))(ctx, userID)
	}
	return nil, nil
}

func (m *MockImageRepo) GetImageByID(ctx context.Context, imageID string, userID string) (*models.Image, error) {
	if fn := m.Get("GetImageByID"); fn != nil {
		return fn.(func(context.Context, string, string) (*models.Image, error))(ctx, imageID, userID)
	}
	return nil, nil
}

func (m *MockImageRepo) DeleteImage(ctx context.Context, imageID string, userID string) error {
	if fn := m.Get("DeleteImage"); fn != nil {
		return fn.(func(context.Context, string, string) error)(ctx, imageID, userID)
	}
	return nil
}

func (m *MockImageRepo) SoftDeleteImage(ctx context.Context, imageID string, userID string) error {
	if fn := m.Get("SoftDeleteImage"); fn != nil {
		return fn.(func(context.Context, string, string) error)(ctx, imageID, userID)
	}
	return nil
}

func (m *MockImageRepo) GetAllImages(ctx context.Context, limit int, offset int) ([]models.Image, error) {
	if fn := m.Get("GetAllImages"); fn != nil {
		return fn.(func(context.Context, int, int) ([]models.Image, error))(ctx, limit, offset)
	}
	return nil, nil
}