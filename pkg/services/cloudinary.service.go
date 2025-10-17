package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	cld *cloudinary.Cloudinary
}

type UploadResult struct {
	PublicID  string
	SecureURL string
	URL       string
	Format    string
	Width     int
	Height    int
}

// NewCloudinaryService creates a new Cloudinary service instance
func NewCloudinaryService() (*CloudinaryService, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("cloudinary credentials not found in environment variables")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudinary: %w", err)
	}

	return &CloudinaryService{
		cld: cld,
	}, nil
}

// UploadImage uploads an image to Cloudinary
func (s *CloudinaryService) UploadImage(ctx context.Context, file multipart.File, filename string, folder string) (*UploadResult, error) {
	uploadParams := uploader.UploadParams{
		PublicID:     generatePublicID(filename),
		Folder:       folder,
		ResourceType: "image",
	}

	// Upload the file
	result, err := s.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image to cloudinary: %w", err)
	}

	return &UploadResult{
		PublicID:  result.PublicID,
		SecureURL: result.SecureURL,
		URL:       result.URL,
		Format:    result.Format,
		Width:     result.Width,
		Height:    result.Height,
	}, nil
}

// DeleteImage deletes an image from Cloudinary by public ID
func (s *CloudinaryService) DeleteImage(ctx context.Context, publicID string) error {
	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "image",
	})
	if err != nil {
		return fmt.Errorf("failed to delete image from cloudinary: %w", err)
	}
	return nil
}

// UploadBase64Image uploads a base64 encoded image to Cloudinary
func (s *CloudinaryService) UploadBase64Image(ctx context.Context, base64Data string, folder string) (*UploadResult, error) {
	uploadParams := uploader.UploadParams{
		Folder:       folder,
		ResourceType: "image",
	}

	result, err := s.cld.Upload.Upload(ctx, base64Data, uploadParams)
	if err != nil {
		return nil, fmt.Errorf("failed to upload base64 image to cloudinary: %w", err)
	}

	return &UploadResult{
		PublicID:  result.PublicID,
		SecureURL: result.SecureURL,
		URL:       result.URL,
		Format:    result.Format,
		Width:     result.Width,
		Height:    result.Height,
	}, nil
}

// ValidateImageFile checks if the file is a valid image
func ValidateImageFile(header *multipart.FileHeader) error {
	// Check file size (max 10MB)
	maxSize := int64(10 * 1024 * 1024)
	if header.Size > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed size of 10MB")
	}

	// Check file extension
	ext := filepath.Ext(header.Filename)
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".svg":  true,
		".bmp":  true,
	}

	if !validExtensions[ext] {
		return fmt.Errorf("invalid file type: %s. Allowed types: jpg, jpeg, png, gif, webp, svg, bmp", ext)
	}

	return nil
}

// ReadFileContent reads the content of a multipart file
func ReadFileContent(file multipart.File) ([]byte, error) {
	defer file.Close()
	return io.ReadAll(file)
}

// generatePublicID generates a unique public ID for the image
func generatePublicID(filename string) string {
	ext := filepath.Ext(filename)
	nameWithoutExt := filename[:len(filename)-len(ext)]
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d", nameWithoutExt, timestamp)
}
