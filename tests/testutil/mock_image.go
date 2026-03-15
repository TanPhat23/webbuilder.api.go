package testutil

import (
	"context"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MockImageRepository struct {
	CreateImageFn       func(ctx context.Context, image models.Image) (*models.Image, error)
	GetImagesByUserIDFn func(ctx context.Context, userID string) ([]models.Image, error)
	GetImageByIDFn      func(ctx context.Context, imageID string, userID string) (*models.Image, error)
	DeleteImageFn       func(ctx context.Context, imageID string, userID string) error
	SoftDeleteImageFn   func(ctx context.Context, imageID string, userID string) error
	GetAllImagesFn      func(ctx context.Context, limit int, offset int) ([]models.Image, error)
}

func (m *MockImageRepository) CreateImage(ctx context.Context, image models.Image) (*models.Image, error) {
	if m.CreateImageFn != nil {
		return m.CreateImageFn(ctx, image)
	}
	return &image, nil
}

func (m *MockImageRepository) GetImagesByUserID(ctx context.Context, userID string) ([]models.Image, error) {
	if m.GetImagesByUserIDFn != nil {
		return m.GetImagesByUserIDFn(ctx, userID)
	}
	return []models.Image{}, nil
}

func (m *MockImageRepository) GetImageByID(ctx context.Context, imageID string, userID string) (*models.Image, error) {
	if m.GetImageByIDFn != nil {
		return m.GetImageByIDFn(ctx, imageID, userID)
	}
	return nil, repositories.ErrImageNotFound
}

func (m *MockImageRepository) DeleteImage(ctx context.Context, imageID string, userID string) error {
	if m.DeleteImageFn != nil {
		return m.DeleteImageFn(ctx, imageID, userID)
	}
	return nil
}

func (m *MockImageRepository) SoftDeleteImage(ctx context.Context, imageID string, userID string) error {
	if m.SoftDeleteImageFn != nil {
		return m.SoftDeleteImageFn(ctx, imageID, userID)
	}
	return nil
}

func (m *MockImageRepository) GetAllImages(ctx context.Context, limit int, offset int) ([]models.Image, error) {
	if m.GetAllImagesFn != nil {
		return m.GetAllImagesFn(ctx, limit, offset)
	}
	return []models.Image{}, nil
}