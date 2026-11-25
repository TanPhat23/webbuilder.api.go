package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"strings"
	"time"

	"github.com/lucsky/cuid"
)

const presignedImageURLTTL = 5 * time.Minute

// ImageService manages uploaded image metadata and backing object storage.
type ImageService struct {
	imageRepo     repositories.ImageRepositoryInterface
	storage       ObjectStorage
	publicBaseURL string
}

// NewImageService constructs an ImageService using PublicAPIURL() as the public base URL.
func NewImageService(imageRepo repositories.ImageRepositoryInterface, storage ObjectStorage) *ImageService {
	return NewImageServiceWithBaseURL(imageRepo, storage, PublicAPIURL())
}

// NewImageServiceWithBaseURL constructs an ImageService with an explicit public base URL.
func NewImageServiceWithBaseURL(imageRepo repositories.ImageRepositoryInterface, storage ObjectStorage, publicBaseURL string) *ImageService {
	return &ImageService{
		imageRepo:     imageRepo,
		storage:       storage,
		publicBaseURL: strings.TrimRight(strings.TrimSpace(publicBaseURL), "/"),
	}
}

// ImageObjectKey returns the deterministic storage object key for a user image.
func ImageObjectKey(userID, imageID string) string {
	return "images/" + userID + "/" + imageID
}

func (s *ImageService) publicImageURL(imageID string) string {
	return s.publicBaseURL + "/api/v1/public/images/" + imageID
}

func (s *ImageService) Create(ctx context.Context, image models.Image) (*models.Image, error) {
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

	return s.imageRepo.Create(ctx, image)
}

func (s *ImageService) persistUploadedObject(
	ctx context.Context,
	userID string,
	imageName *string,
	r io.Reader,
	size int64,
	contentType string,
) (*models.ImageUploadResponse, error) {
	imageID := cuid.New()
	key := ImageObjectKey(userID, imageID)

	if err := s.storage.PutObject(ctx, key, r, size, contentType); err != nil {
		return nil, fmt.Errorf("failed to upload image to object storage: %w", err)
	}

	now := time.Now()
	image := models.Image{
		ImageId:   imageID,
		ImageLink: s.publicImageURL(imageID),
		ImageName: imageName,
		UserId:    userID,
		AuditFields: models.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	created, err := s.Create(ctx, image)
	if err != nil {
		_ = s.storage.DeleteObject(ctx, key)
		return nil, err
	}

	return &models.ImageUploadResponse{
		ImageId:   created.ImageId,
		ImageLink: created.ImageLink,
		ImageName: created.ImageName,
		CreatedAt: created.CreatedAt,
	}, nil
}

func seekerSize(r io.Seeker) (int64, error) {
	size, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}
	return size, nil
}

func (s *ImageService) Upload(ctx context.Context, userID string, imageName *string, file multipart.File) (*models.ImageUploadResponse, error) {
	if userID == "" {
		return nil, errors.New("userId is required")
	}
	if file == nil {
		return nil, errors.New("image file is required")
	}
	if s.storage == nil {
		return nil, errors.New("image upload service is not configured")
	}

	contentType, err := DetectImageContentType(file)
	if err != nil {
		return nil, err
	}
	size, err := seekerSize(file)
	if err != nil {
		return nil, fmt.Errorf("failed to determine image size: %w", err)
	}

	return s.persistUploadedObject(ctx, userID, imageName, file, size, contentType)
}

func decodeBase64Image(raw string) ([]byte, string, error) {
	payload := strings.TrimSpace(raw)
	if strings.HasPrefix(strings.ToLower(payload), "data:") {
		comma := strings.IndexByte(payload, ',')
		if comma < 0 || !strings.HasSuffix(strings.ToLower(payload[:comma]), ";base64") {
			return nil, "", errors.New("invalid base64 image data URI")
		}
		payload = payload[comma+1:]
	}

	if base64.StdEncoding.DecodedLen(len(payload)) > int(maxImageUploadBytes)+3 {
		return nil, "", errors.New("image exceeds maximum allowed size of 10MB")
	}

	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, "", fmt.Errorf("invalid base64 image encoding: %w", err)
	}
	if int64(len(decoded)) > maxImageUploadBytes {
		return nil, "", errors.New("image exceeds maximum allowed size of 10MB")
	}

	contentType, err := DetectImageContentType(bytes.NewReader(decoded))
	if err != nil {
		return nil, "", err
	}
	return decoded, contentType, nil
}

func (s *ImageService) UploadBase64(ctx context.Context, userID string, imageData string, imageName *string) (*models.ImageUploadResponse, error) {
	if userID == "" {
		return nil, errors.New("userId is required")
	}
	if imageData == "" {
		return nil, errors.New("image data is required")
	}
	if s.storage == nil {
		return nil, errors.New("image upload service is not configured")
	}

	decoded, contentType, err := decodeBase64Image(imageData)
	if err != nil {
		return nil, err
	}

	return s.persistUploadedObject(ctx, userID, imageName, bytes.NewReader(decoded), int64(len(decoded)), contentType)
}

func (s *ImageService) ListByUser(ctx context.Context, userID string) ([]models.Image, error) {
	if userID == "" {
		return nil, errors.New("userId is required")
	}
	return s.imageRepo.ListByUser(ctx, userID)
}

func (s *ImageService) GetByID(ctx context.Context, imageID, userID string) (*models.Image, error) {
	if imageID == "" {
		return nil, errors.New("image id is required")
	}
	if userID == "" {
		return nil, errors.New("userId is required")
	}
	return s.imageRepo.GetByID(ctx, imageID, userID)
}

// PublicURL looks up an active image record and mints a short-lived presigned GET URL.
func (s *ImageService) PublicURL(ctx context.Context, imageID string) (string, error) {
	if imageID == "" {
		return "", errors.New("image id is required")
	}
	if s.storage == nil {
		return "", errors.New("image storage service is not configured")
	}
	img, err := s.imageRepo.GetPublicByID(ctx, imageID)
	if err != nil {
		return "", err
	}
	return s.storage.PresignedGetURL(ctx, ImageObjectKey(img.UserId, img.ImageId), presignedImageURLTTL)
}

func (s *ImageService) Delete(ctx context.Context, imageID, userID string) error {
	if imageID == "" {
		return errors.New("image id is required")
	}
	if userID == "" {
		return errors.New("userId is required")
	}
	if err := s.imageRepo.Delete(ctx, imageID, userID); err != nil {
		return err
	}
	if s.storage != nil {
		_ = s.storage.DeleteObject(ctx, ImageObjectKey(userID, imageID))
	}
	return nil
}

func (s *ImageService) SoftDelete(ctx context.Context, imageID, userID string) error {
	if imageID == "" {
		return errors.New("image id is required")
	}
	if userID == "" {
		return errors.New("userId is required")
	}
	return s.imageRepo.SoftDelete(ctx, imageID, userID)
}

func (s *ImageService) ListAll(ctx context.Context, limit, offset int) ([]models.Image, error) {
	if limit < 0 {
		return nil, errors.New("limit cannot be negative")
	}
	if offset < 0 {
		return nil, errors.New("offset cannot be negative")
	}
	return s.imageRepo.ListAll(ctx, limit, offset)
}
