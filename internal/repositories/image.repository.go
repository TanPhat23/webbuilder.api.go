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

func (r *ImageRepository) Create(ctx context.Context, image models.Image) (*models.Image, error) {
	if err := gorm.G[models.Image](r.db).Create(ctx, &image); err != nil {
		return nil, fmt.Errorf("failed to create image: %w", err)
	}
	return &image, nil
}

func (r *ImageRepository) ListByUser(ctx context.Context, userID string) ([]models.Image, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	images, err := gorm.G[models.Image](r.db).
		Where(`"UserId" = ? AND "DeletedAt" IS NULL`, userID).
		Order(`"CreatedAt" DESC`).
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get images by user ID: %w", err)
	}
	return images, nil
}

func (r *ImageRepository) GetByID(ctx context.Context, imageID, userID string) (*models.Image, error) {
	if imageID == "" {
		return nil, errors.New("imageID is required")
	}
	if userID == "" {
		return nil, errors.New("userID is required")
	}

	image, err := gorm.G[models.Image](r.db).
		Where(`"ImageId" = ? AND "UserId" = ? AND "DeletedAt" IS NULL`, imageID, userID).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrImageNotFound
		}
		return nil, fmt.Errorf("failed to get image by ID: %w", err)
	}
	return &image, nil
}

func (r *ImageRepository) GetPublicByID(ctx context.Context, imageID string) (*models.Image, error) {
	if imageID == "" {
		return nil, errors.New("imageID is required")
	}

	image, err := gorm.G[models.Image](r.db).
		Where(`"ImageId" = ? AND "DeletedAt" IS NULL`, imageID).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrImageNotFound
		}
		return nil, fmt.Errorf("failed to get public image by ID: %w", err)
	}
	return &image, nil
}

func (r *ImageRepository) Delete(ctx context.Context, imageID, userID string) error {
	if imageID == "" {
		return errors.New("imageID is required")
	}
	if userID == "" {
		return errors.New("userID is required")
	}

	rowsAffected, err := gorm.G[models.Image](r.db).
		Where(`"ImageId" = ? AND "UserId" = ?`, imageID, userID).
		Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	if rowsAffected == 0 {
		return ErrImageNotFound
	}
	return nil
}

func (r *ImageRepository) SoftDelete(ctx context.Context, imageID, userID string) error {
	if imageID == "" {
		return errors.New("imageID is required")
	}
	if userID == "" {
		return errors.New("userID is required")
	}

	now := time.Now()
	rowsAffected, err := gorm.G[models.Image](r.db).
		Where(`"ImageId" = ? AND "UserId" = ? AND "DeletedAt" IS NULL`, imageID, userID).
		Update(ctx, `"DeletedAt"`, now)
	if err != nil {
		return fmt.Errorf("failed to soft-delete image: %w", err)
	}
	if rowsAffected == 0 {
		return ErrImageNotFound
	}
	return nil
}

func (r *ImageRepository) ListAll(ctx context.Context, limit, offset int) ([]models.Image, error) {
	query := gorm.G[models.Image](r.db).
		Where(`"DeletedAt" IS NULL`).
		Order(`"CreatedAt" DESC`)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	images, err := query.Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all images: %w", err)
	}
	return images, nil
}
