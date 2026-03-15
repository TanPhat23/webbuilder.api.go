package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"my-go-app/internal/models"

	"gorm.io/gorm"
)

var ErrImageNotFound = errors.New("image not found")

type ImageRepository struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) ImageRepositoryInterface {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) CreateImage(ctx context.Context, image models.Image) (*models.Image, error) {
	if err := r.db.WithContext(ctx).Create(&image).Error; err != nil {
		return nil, fmt.Errorf("failed to create image: %w", err)
	}
	return &image, nil
}

func (r *ImageRepository) GetImagesByUserID(ctx context.Context, userID string) ([]models.Image, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	var images []models.Image
	err := r.db.WithContext(ctx).
		Where(`"UserId" = ? AND "DeletedAt" IS NULL`, userID).
		Order(`"CreatedAt" DESC`).
		Find(&images).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get images by user ID: %w", err)
	}
	return images, nil
}

func (r *ImageRepository) GetImageByID(ctx context.Context, imageID, userID string) (*models.Image, error) {
	if imageID == "" {
		return nil, errors.New("imageID is required")
	}
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	var image models.Image
	err := r.db.WithContext(ctx).
		Where(`"ImageId" = ? AND "UserId" = ? AND "DeletedAt" IS NULL`, imageID, userID).
		First(&image).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrImageNotFound
		}
		return nil, fmt.Errorf("failed to get image by ID: %w", err)
	}
	return &image, nil
}

func (r *ImageRepository) DeleteImage(ctx context.Context, imageID, userID string) error {
	if imageID == "" {
		return errors.New("imageID is required")
	}
	if userID == "" {
		return errors.New("userID is required")
	}

	result := r.db.WithContext(ctx).
		Where(`"ImageId" = ? AND "UserId" = ?`, imageID, userID).
		Delete(&models.Image{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete image: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrImageNotFound
	}
	return nil
}

func (r *ImageRepository) SoftDeleteImage(ctx context.Context, imageID, userID string) error {
	if imageID == "" {
		return errors.New("imageID is required")
	}
	if userID == "" {
		return errors.New("userID is required")
	}

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.Image{}).
		Where(`"ImageId" = ? AND "UserId" = ? AND "DeletedAt" IS NULL`, imageID, userID).
		Update(`"DeletedAt"`, now)
	if result.Error != nil {
		return fmt.Errorf("failed to soft-delete image: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrImageNotFound
	}
	return nil
}

func (r *ImageRepository) GetAllImages(ctx context.Context, limit, offset int) ([]models.Image, error) {
	var images []models.Image
	query := r.db.WithContext(ctx).
		Where(`"DeletedAt" IS NULL`).
		Order(`"CreatedAt" DESC`)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&images).Error; err != nil {
		return nil, fmt.Errorf("failed to get all images: %w", err)
	}
	return images, nil
}