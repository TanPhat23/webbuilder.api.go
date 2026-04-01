package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"time"

	"github.com/lucsky/cuid"
)

type ImageService struct {
	imageRepo     repositories.ImageRepositoryInterface
	cloudinarySvc *CloudinaryService
}

func NewImageService(imageRepo repositories.ImageRepositoryInterface, cloudinarySvc *CloudinaryService) *ImageService {
	return &ImageService{
		imageRepo:     imageRepo,
		cloudinarySvc: cloudinarySvc,
	}
}

func (s *ImageService) CreateImage(ctx context.Context, image models.Image) (*models.Image, error) {
	if image.ImageId == "" {
		image.ImageId = cuid.New()
	}
	if image.CreatedAt.IsZero() {
		image.CreatedAt = time.Now()
	}
	if image.UpdatedAt.IsZero() {
		image.UpdatedAt = image.CreatedAt
	}
	if image.UserId == "" {
		return nil, errors.New("userId is required")
	}
	if image.ImageLink == "" {
		return nil, errors.New("image link is required")
	}

	return s.imageRepo.CreateImage(ctx, image)
}

func (s *ImageService) CreateUploadedImage(ctx context.Context, userID string, fileName string, imageName *string, file multipart.File) (*models.ImageUploadResponse, error) {
	if userID == "" {
		return nil, errors.New("userId is required")
	}
	if file == nil {
		return nil, errors.New("image file is required")
	}

	uploadResult, err := s.cloudinarySvc.UploadImage(ctx, file, fileName, "webbuilder/"+userID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image to Cloudinary: %w", err)
	}

	now := time.Now()
	image := models.Image{
		ImageId:   cuid.New(),
		ImageLink: uploadResult.SecureURL,
		ImageName: imageName,
		UserId:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	created, err := s.CreateImage(ctx, image)
	if err != nil {
		_ = s.cloudinarySvc.DeleteImage(ctx, uploadResult.PublicID)
		return nil, err
	}

	return &models.ImageUploadResponse{
		ImageId:   created.ImageId,
		ImageLink: created.ImageLink,
		ImageName: created.ImageName,
		CreatedAt: created.CreatedAt,
	}, nil
}

func (s *ImageService) CreateBase64UploadedImage(ctx context.Context, userID string, imageData string, imageName *string) (*models.ImageUploadResponse, error) {
	if userID == "" {
		return nil, errors.New("userId is required")
	}
	if imageData == "" {
		return nil, errors.New("image data is required")
	}

	uploadResult, err := s.cloudinarySvc.UploadBase64Image(ctx, imageData, "webbuilder/"+userID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image to Cloudinary: %w", err)
	}

	now := time.Now()
	image := models.Image{
		ImageId:   cuid.New(),
		ImageLink: uploadResult.SecureURL,
		ImageName: imageName,
		UserId:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	created, err := s.CreateImage(ctx, image)
	if err != nil {
		_ = s.cloudinarySvc.DeleteImage(ctx, uploadResult.PublicID)
		return nil, err
	}

	return &models.ImageUploadResponse{
		ImageId:   created.ImageId,
		ImageLink: created.ImageLink,
		ImageName: created.ImageName,
		CreatedAt: created.CreatedAt,
	}, nil
}

func (s *ImageService) GetImagesByUserID(ctx context.Context, userID string) ([]models.Image, error) {
	if userID == "" {
		return nil, errors.New("userId is required")
	}
	return s.imageRepo.GetImagesByUserID(ctx, userID)
}

func (s *ImageService) GetImageByID(ctx context.Context, imageID, userID string) (*models.Image, error) {
	if imageID == "" {
		return nil, errors.New("image id is required")
	}
	if userID == "" {
		return nil, errors.New("userId is required")
	}
	return s.imageRepo.GetImageByID(ctx, imageID, userID)
}

func (s *ImageService) DeleteImage(ctx context.Context, imageID, userID string) error {
	if imageID == "" {
		return errors.New("image id is required")
	}
	if userID == "" {
		return errors.New("userId is required")
	}
	return s.imageRepo.DeleteImage(ctx, imageID, userID)
}

func (s *ImageService) SoftDeleteImage(ctx context.Context, imageID, userID string) error {
	if imageID == "" {
		return errors.New("image id is required")
	}
	if userID == "" {
		return errors.New("userId is required")
	}
	return s.imageRepo.SoftDeleteImage(ctx, imageID, userID)
}

func (s *ImageService) GetAllImages(ctx context.Context, limit, offset int) ([]models.Image, error) {
	if limit < 0 {
		return nil, errors.New("limit cannot be negative")
	}
	if offset < 0 {
		return nil, errors.New("offset cannot be negative")
	}
	return s.imageRepo.GetAllImages(ctx, limit, offset)
}