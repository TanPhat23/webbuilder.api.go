package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"my-go-app/internal/models"
	"my-go-app/internal/services"
	test "my-go-app/test/mockrepo"
)

func TestCreateImage_Success(t *testing.T) {
	mock := test.NewMockImageRepo()
	mock.SetCreateImage(func(ctx context.Context, image models.Image) (*models.Image, error) {
		return &image, nil
	})

	service := services.NewImageService(mock, nil)

	image := models.Image{
		ImageLink: "https://example.com/image.jpg",
		UserId:    "user123",
	}

	result, err := service.CreateImage(context.Background(), image)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Errorf("expected image, got nil")
	}
	if result.UserId != "user123" {
		t.Errorf("expected userId user123, got %v", result.UserId)
	}
	if result.ImageId == "" {
		t.Errorf("expected imageId to be set, got empty")
	}
	if result.CreatedAt.IsZero() {
		t.Errorf("expected createdAt to be set, got zero time")
	}
}

func TestCreateImage_MissingUserID(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	image := models.Image{
		ImageLink: "https://example.com/image.jpg",
	}

	result, err := service.CreateImage(context.Background(), image)
	if err == nil {
		t.Errorf("expected error for missing userId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "userId is required" {
		t.Errorf("expected 'userId is required', got %v", err.Error())
	}
}

func TestCreateImage_MissingImageLink(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	image := models.Image{
		UserId: "user123",
	}

	result, err := service.CreateImage(context.Background(), image)
	if err == nil {
		t.Errorf("expected error for missing image link")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "image link is required" {
		t.Errorf("expected 'image link is required', got %v", err.Error())
	}
}

func TestCreateImage_WithCustomID(t *testing.T) {
	mock := test.NewMockImageRepo()
	customID := "custom-image-id"
	mock.SetCreateImage(func(ctx context.Context, image models.Image) (*models.Image, error) {
		return &image, nil
	})

	service := services.NewImageService(mock, nil)

	image := models.Image{
		ImageId:   customID,
		ImageLink: "https://example.com/image.jpg",
		UserId:    "user123",
	}

	result, err := service.CreateImage(context.Background(), image)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.ImageId != customID {
		t.Errorf("expected imageId %v, got %v", customID, result.ImageId)
	}
}

func TestCreateImage_SetsTimestamps(t *testing.T) {
	mock := test.NewMockImageRepo()
	mock.SetCreateImage(func(ctx context.Context, image models.Image) (*models.Image, error) {
		return &image, nil
	})

	service := services.NewImageService(mock, nil)

	beforeTime := time.Now()
	image := models.Image{
		ImageLink: "https://example.com/image.jpg",
		UserId:    "user123",
	}

	result, err := service.CreateImage(context.Background(), image)
	afterTime := time.Now()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.CreatedAt.Before(beforeTime) || result.CreatedAt.After(afterTime.Add(time.Second)) {
		t.Errorf("expected createdAt to be set near current time")
	}
	if result.UpdatedAt.Before(beforeTime) || result.UpdatedAt.After(afterTime.Add(time.Second)) {
		t.Errorf("expected updatedAt to be set near current time")
	}
}

func TestGetImagesByUserID_Success(t *testing.T) {
	images := []models.Image{
		{
			ImageId:   "img1",
			ImageLink: "https://example.com/img1.jpg",
			UserId:    "user123",
			CreatedAt: time.Now(),
		},
		{
			ImageId:   "img2",
			ImageLink: "https://example.com/img2.jpg",
			UserId:    "user123",
			CreatedAt: time.Now(),
		},
	}

	mock := test.NewMockImageRepo()
	mock.SetGetImagesByUserID(func(ctx context.Context, userID string) ([]models.Image, error) {
		return images, nil
	})

	service := services.NewImageService(mock, nil)

	result, err := service.GetImagesByUserID(context.Background(), "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 images, got %d", len(result))
	}
}

func TestGetImagesByUserID_EmptyUserID(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	result, err := service.GetImagesByUserID(context.Background(), "")
	if err == nil {
		t.Errorf("expected error for empty userId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "userId is required" {
		t.Errorf("expected 'userId is required', got %v", err.Error())
	}
}

func TestGetImagesByUserID_RepositoryError(t *testing.T) {
	mock := test.NewMockImageRepo()
	mock.SetGetImagesByUserID(func(ctx context.Context, userID string) ([]models.Image, error) {
		return nil, errors.New("database error")
	})

	service := services.NewImageService(mock, nil)

	result, err := service.GetImagesByUserID(context.Background(), "user123")
	if err == nil {
		t.Errorf("expected database error")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetImageByID_Success(t *testing.T) {
	image := &models.Image{
		ImageId:   "img123",
		ImageLink: "https://example.com/img.jpg",
		UserId:    "user123",
		CreatedAt: time.Now(),
	}

	mock := test.NewMockImageRepo()
	mock.SetGetImageByID(func(ctx context.Context, imageID string, userID string) (*models.Image, error) {
		return image, nil
	})

	service := services.NewImageService(mock, nil)

	result, err := service.GetImageByID(context.Background(), "img123", "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Errorf("expected image, got nil")
	}
	if result.ImageId != "img123" {
		t.Errorf("expected imageId img123, got %v", result.ImageId)
	}
}

func TestGetImageByID_EmptyImageID(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	result, err := service.GetImageByID(context.Background(), "", "user123")
	if err == nil {
		t.Errorf("expected error for empty imageId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "image id is required" {
		t.Errorf("expected 'image id is required', got %v", err.Error())
	}
}

func TestGetImageByID_EmptyUserID(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	result, err := service.GetImageByID(context.Background(), "img123", "")
	if err == nil {
		t.Errorf("expected error for empty userId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "userId is required" {
		t.Errorf("expected 'userId is required', got %v", err.Error())
	}
}

func TestGetImageByID_RepositoryError(t *testing.T) {
	mock := test.NewMockImageRepo()
	mock.SetGetImageByID(func(ctx context.Context, imageID string, userID string) (*models.Image, error) {
		return nil, errors.New("database error")
	})

	service := services.NewImageService(mock, nil)

	result, err := service.GetImageByID(context.Background(), "img123", "user123")
	if err == nil {
		t.Errorf("expected database error")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestDeleteImage_Success(t *testing.T) {
	mock := test.NewMockImageRepo()
	mock.SetDeleteImage(func(ctx context.Context, imageID string, userID string) error {
		return nil
	})

	service := services.NewImageService(mock, nil)

	err := service.DeleteImage(context.Background(), "img123", "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestDeleteImage_EmptyImageID(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	err := service.DeleteImage(context.Background(), "", "user123")
	if err == nil {
		t.Errorf("expected error for empty imageId")
	}
	if err.Error() != "image id is required" {
		t.Errorf("expected 'image id is required', got %v", err.Error())
	}
}

func TestDeleteImage_EmptyUserID(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	err := service.DeleteImage(context.Background(), "img123", "")
	if err == nil {
		t.Errorf("expected error for empty userId")
	}
	if err.Error() != "userId is required" {
		t.Errorf("expected 'userId is required', got %v", err.Error())
	}
}

func TestDeleteImage_RepositoryError(t *testing.T) {
	mock := test.NewMockImageRepo()
	mock.SetDeleteImage(func(ctx context.Context, imageID string, userID string) error {
		return errors.New("database error")
	})

	service := services.NewImageService(mock, nil)

	err := service.DeleteImage(context.Background(), "img123", "user123")
	if err == nil {
		t.Errorf("expected database error")
	}
}

func TestSoftDeleteImage_Success(t *testing.T) {
	mock := test.NewMockImageRepo()
	mock.SetSoftDeleteImage(func(ctx context.Context, imageID string, userID string) error {
		return nil
	})

	service := services.NewImageService(mock, nil)

	err := service.SoftDeleteImage(context.Background(), "img123", "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestSoftDeleteImage_EmptyImageID(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	err := service.SoftDeleteImage(context.Background(), "", "user123")
	if err == nil {
		t.Errorf("expected error for empty imageId")
	}
	if err.Error() != "image id is required" {
		t.Errorf("expected 'image id is required', got %v", err.Error())
	}
}

func TestSoftDeleteImage_EmptyUserID(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	err := service.SoftDeleteImage(context.Background(), "img123", "")
	if err == nil {
		t.Errorf("expected error for empty userId")
	}
	if err.Error() != "userId is required" {
		t.Errorf("expected 'userId is required', got %v", err.Error())
	}
}

func TestSoftDeleteImage_RepositoryError(t *testing.T) {
	mock := test.NewMockImageRepo()
	mock.SetSoftDeleteImage(func(ctx context.Context, imageID string, userID string) error {
		return errors.New("database error")
	})

	service := services.NewImageService(mock, nil)

	err := service.SoftDeleteImage(context.Background(), "img123", "user123")
	if err == nil {
		t.Errorf("expected database error")
	}
}

func TestGetAllImages_Success(t *testing.T) {
	images := []models.Image{
		{
			ImageId:   "img1",
			ImageLink: "https://example.com/img1.jpg",
			UserId:    "user123",
		},
		{
			ImageId:   "img2",
			ImageLink: "https://example.com/img2.jpg",
			UserId:    "user456",
		},
	}

	mock := test.NewMockImageRepo()
	mock.SetGetAllImages(func(ctx context.Context, limit int, offset int) ([]models.Image, error) {
		return images, nil
	})

	service := services.NewImageService(mock, nil)

	result, err := service.GetAllImages(context.Background(), 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 images, got %d", len(result))
	}
}

func TestGetAllImages_WithZeroLimit(t *testing.T) {
	mock := test.NewMockImageRepo()
	mock.SetGetAllImages(func(ctx context.Context, limit int, offset int) ([]models.Image, error) {
		return []models.Image{}, nil
	})

	service := services.NewImageService(mock, nil)

	result, err := service.GetAllImages(context.Background(), 0, 0)
	if err != nil {
		t.Errorf("expected no error with zero limit, got %v", err)
	}
	if result == nil {
		t.Errorf("expected empty result, got nil")
	}
}

func TestGetAllImages_NegativeLimit(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	result, err := service.GetAllImages(context.Background(), -1, 0)
	if err == nil {
		t.Errorf("expected error for negative limit")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "limit cannot be negative" {
		t.Errorf("expected 'limit cannot be negative', got %v", err.Error())
	}
}

func TestGetAllImages_NegativeOffset(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	result, err := service.GetAllImages(context.Background(), 10, -1)
	if err == nil {
		t.Errorf("expected error for negative offset")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "offset cannot be negative" {
		t.Errorf("expected 'offset cannot be negative', got %v", err.Error())
	}
}

func TestGetAllImages_RepositoryError(t *testing.T) {
	mock := test.NewMockImageRepo()
	mock.SetGetAllImages(func(ctx context.Context, limit int, offset int) ([]models.Image, error) {
		return nil, errors.New("database error")
	})

	service := services.NewImageService(mock, nil)

	result, err := service.GetAllImages(context.Background(), 10, 0)
	if err == nil {
		t.Errorf("expected database error")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetAllImages_WithPagination(t *testing.T) {
	mock := test.NewMockImageRepo()
	var capturedLimit, capturedOffset int
	mock.SetGetAllImages(func(ctx context.Context, limit int, offset int) ([]models.Image, error) {
		capturedLimit = limit
		capturedOffset = offset
		return []models.Image{}, nil
	})

	service := services.NewImageService(mock, nil)

	_, err := service.GetAllImages(context.Background(), 50, 100)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if capturedLimit != 50 {
		t.Errorf("expected limit 50, got %d", capturedLimit)
	}
	if capturedOffset != 100 {
		t.Errorf("expected offset 100, got %d", capturedOffset)
	}
}

func TestCreateUploadedImage_MissingUserID(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	mockFile := test.NewMockFile([]byte("test data"))

	result, err := service.CreateUploadedImage(context.Background(), "", "test.jpg", nil, mockFile)
	if err == nil {
		t.Errorf("expected error for missing userId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "userId is required" {
		t.Errorf("expected 'userId is required', got %v", err.Error())
	}
}

func TestCreateUploadedImage_NilFile(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	result, err := service.CreateUploadedImage(context.Background(), "user123", "test.jpg", nil, nil)
	if err == nil {
		t.Errorf("expected error for nil file")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "image file is required" {
		t.Errorf("expected 'image file is required', got %v", err.Error())
	}
}

func TestCreateBase64UploadedImage_MissingUserID(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	result, err := service.CreateBase64UploadedImage(context.Background(), "", "data:image/jpeg;base64,test", nil)
	if err == nil {
		t.Errorf("expected error for missing userId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "userId is required" {
		t.Errorf("expected 'userId is required', got %v", err.Error())
	}
}

func TestCreateBase64UploadedImage_MissingImageData(t *testing.T) {
	mock := test.NewMockImageRepo()
	service := services.NewImageService(mock, nil)

	result, err := service.CreateBase64UploadedImage(context.Background(), "user123", "", nil)
	if err == nil {
		t.Errorf("expected error for missing image data")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "image data is required" {
		t.Errorf("expected 'image data is required', got %v", err.Error())
	}
}

func TestCreateImage_RepositoryError(t *testing.T) {
	mock := test.NewMockImageRepo()
	mock.SetCreateImage(func(ctx context.Context, image models.Image) (*models.Image, error) {
		return nil, errors.New("database error")
	})

	service := services.NewImageService(mock, nil)

	image := models.Image{
		ImageLink: "https://example.com/image.jpg",
		UserId:    "user123",
	}

	result, err := service.CreateImage(context.Background(), image)
	if err == nil {
		t.Errorf("expected database error")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetImagesByUserID_EmptyResult(t *testing.T) {
	mock := test.NewMockImageRepo()
	mock.SetGetImagesByUserID(func(ctx context.Context, userID string) ([]models.Image, error) {
		return []models.Image{}, nil
	})

	service := services.NewImageService(mock, nil)

	result, err := service.GetImagesByUserID(context.Background(), "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 images, got %d", len(result))
	}
}

func TestGetAllImages_EmptyResult(t *testing.T) {
	mock := test.NewMockImageRepo()
	mock.SetGetAllImages(func(ctx context.Context, limit int, offset int) ([]models.Image, error) {
		return []models.Image{}, nil
	})

	service := services.NewImageService(mock, nil)

	result, err := service.GetAllImages(context.Background(), 10, 0)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 images, got %d", len(result))
	}
}