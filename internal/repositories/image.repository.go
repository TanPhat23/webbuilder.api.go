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

func (r *ImageRepository) CreateImage(image models.Image) (*models.Image, error) {
	if err := r.DB.Create(&image).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *ImageRepository) GetImagesByUserID(userID string) ([]models.Image, error) {
	var images []models.Image
	err := r.DB.Where(&models.Image{UserId: userID, DeletedAt: nil}).
		Order("\"CreatedAt\" DESC").
		Find(&images).Error
	if err != nil {
		return nil, err
	}
	return images, nil
}

func (r *ImageRepository) GetImageByID(imageID string, userID string) (*models.Image, error) {
	var image models.Image
	err := r.DB.Where(&models.Image{ImageId: imageID, UserId: userID, DeletedAt: nil}).
		First(&image).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &image, nil
}

func (r *ImageRepository) DeleteImage(imageID string, userID string) error {
	result := r.DB.Where(&models.Image{ImageId: imageID, UserId: userID}).
		Delete(&models.Image{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ImageRepository) SoftDeleteImage(imageID string, userID string) error {
	now := time.Now()
	result := r.DB.Model(&models.Image{}).
		Where(&models.Image{ImageId: imageID, UserId: userID, DeletedAt: nil}).
		Update("\"DeletedAt\"", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ImageRepository) GetAllImages(limit int, offset int) ([]models.Image, error) {
	var images []models.Image
	query := r.DB.Where(&models.Image{DeletedAt: nil}).
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
