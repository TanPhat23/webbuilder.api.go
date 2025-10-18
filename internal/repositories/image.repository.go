package repositories

import (
	"my-go-app/internal/models"
	"time"

	"gorm.io/gorm"
)

type ImageRepository struct {
	DB *gorm.DB
}

func NewImageRepository(db *gorm.DB) *ImageRepository {
	return &ImageRepository{
		DB: db,
	}
}

// CreateImage creates a new image record
func (r *ImageRepository) CreateImage(image models.Image) (*models.Image, error) {
	if err := r.DB.Create(&image).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

// GetImagesByUserID retrieves all images for a specific user (excluding soft-deleted)
func (r *ImageRepository) GetImagesByUserID(userID string) ([]models.Image, error) {
	var images []models.Image
	err := r.DB.Where("\"UserId\" = ? AND \"DeletedAt\" IS NULL", userID).
		Order("\"CreatedAt\" DESC").
		Find(&images).Error
	if err != nil {
		return nil, err
	}
	return images, nil
}

// GetImageByID retrieves a specific image by ID and user ID
func (r *ImageRepository) GetImageByID(imageID string, userID string) (*models.Image, error) {
	var image models.Image
	err := r.DB.Where("\"ImageId\" = ? AND \"UserId\" = ? AND \"DeletedAt\" IS NULL", imageID, userID).
		First(&image).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &image, nil
}

// DeleteImage permanently deletes an image record
func (r *ImageRepository) DeleteImage(imageID string, userID string) error {
	result := r.DB.Where("\"ImageId\" = ? AND \"UserId\" = ?", imageID, userID).
		Delete(&models.Image{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// SoftDeleteImage soft deletes an image record by setting DeletedAt timestamp
func (r *ImageRepository) SoftDeleteImage(imageID string, userID string) error {
	now := time.Now()
	result := r.DB.Model(&models.Image{}).
		Where("\"ImageId\" = ? AND \"UserId\" = ? AND \"DeletedAt\" IS NULL", imageID, userID).
		Update("DeletedAt", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetAllImages retrieves all images with pagination (excluding soft-deleted)
func (r *ImageRepository) GetAllImages(limit int, offset int) ([]models.Image, error) {
	var images []models.Image
	query := r.DB.Where("\"DeletedAt\" IS NULL").
		Order("\"CreatedAt\" DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&images).Error
	if err != nil {
		return nil, err
	}
	return images, nil
}
