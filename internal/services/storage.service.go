package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	maxImageUploadBytes = int64(10 * 1024 * 1024)
	defaultS3Region     = "auto"
)

// ObjectStorage defines object-storage operations required by image workflows.
type ObjectStorage interface {
	PutObject(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	DeleteObject(ctx context.Context, key string) error
	PresignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}

// S3StorageConfig configures an S3-compatible / MinIO object storage client.
type S3StorageConfig struct {
	Endpoint     string
	Bucket       string
	AccessKeyID  string
	SecretKey    string
	Region       string
	UseSSL       bool
	BucketLookup minio.BucketLookupType
}

// S3StorageService implements ObjectStorage using github.com/minio/minio-go/v7.
type S3StorageService struct {
	client *minio.Client
	bucket string
}

// PublicAPIURL returns the externally reachable base URL of this API service.
func PublicAPIURL() string {
	if v := strings.TrimSpace(os.Getenv("API_PUBLIC_URL")); v != "" {
		return strings.TrimRight(v, "/")
	}
	if d := strings.TrimSpace(os.Getenv("RAILWAY_PUBLIC_DOMAIN")); d != "" {
		return "https://" + strings.TrimRight(d, "/")
	}
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	return "http://localhost:" + port
}

func firstNonEmptyEnv(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

// ParseS3Endpoint parses an endpoint string (with or without http/https scheme)
// and determines host:port and whether TLS should be used.
func ParseS3Endpoint(raw string, sslOverride string) (string, bool, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false, fmt.Errorf("storage endpoint is required")
	}

	secure := true
	host := trimmed
	if strings.Contains(trimmed, "://") {
		u, err := url.Parse(trimmed)
		if err != nil {
			return "", false, fmt.Errorf("invalid storage endpoint %q: %w", raw, err)
		}
		if u.Host == "" || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
			return "", false, fmt.Errorf("storage endpoint %q must be host[:port] without path or query", raw)
		}
		switch strings.ToLower(u.Scheme) {
		case "https":
			secure = true
		case "http":
			secure = false
		default:
			return "", false, fmt.Errorf("unsupported storage endpoint scheme %q", u.Scheme)
		}
		host = u.Host
	}

	switch strings.ToLower(strings.TrimSpace(sslOverride)) {
	case "true", "1":
		secure = true
	case "false", "0":
		secure = false
	}

	return host, secure, nil
}

func parseBucketLookup(urlStyle string, usePathStyle string, secure bool) minio.BucketLookupType {
	switch strings.ToLower(strings.TrimSpace(urlStyle)) {
	case "path":
		return minio.BucketLookupPath
	case "virtual-host", "dns":
		return minio.BucketLookupDNS
	}
	switch strings.ToLower(strings.TrimSpace(usePathStyle)) {
	case "true", "1":
		return minio.BucketLookupPath
	case "false", "0":
		return minio.BucketLookupDNS
	}
	if !secure {
		return minio.BucketLookupPath
	}
	return minio.BucketLookupAuto
}

// NewS3StorageConfigFromEnv loads S3/MinIO configuration from environment variables.
func NewS3StorageConfigFromEnv() (S3StorageConfig, error) {
	rawEndpoint := firstNonEmptyEnv("S3_ENDPOINT", "RAILWAY_BUCKET_ENDPOINT", "MINIO_ENDPOINT")
	bucket := firstNonEmptyEnv("S3_BUCKET", "RAILWAY_BUCKET_NAME", "MINIO_BUCKET")
	accessKey := firstNonEmptyEnv("S3_ACCESS_KEY_ID", "RAILWAY_BUCKET_ACCESS_KEY_ID", "MINIO_ACCESS_KEY")
	secretKey := firstNonEmptyEnv("S3_SECRET_ACCESS_KEY", "RAILWAY_BUCKET_SECRET_ACCESS_KEY", "MINIO_SECRET_KEY")
	region := firstNonEmptyEnv("S3_REGION", "RAILWAY_BUCKET_REGION", "MINIO_REGION")
	if region == "" {
		region = defaultS3Region
	}

	if rawEndpoint == "" || bucket == "" || accessKey == "" || secretKey == "" {
		return S3StorageConfig{}, fmt.Errorf("object storage credentials not found in environment variables")
	}

	host, secure, err := ParseS3Endpoint(rawEndpoint, os.Getenv("S3_USE_SSL"))
	if err != nil {
		return S3StorageConfig{}, err
	}

	lookup := parseBucketLookup(
		firstNonEmptyEnv("S3_URL_STYLE", "RAILWAY_BUCKET_URL_STYLE"),
		os.Getenv("S3_USE_PATH_STYLE"),
		secure,
	)

	return S3StorageConfig{
		Endpoint:     host,
		Bucket:       bucket,
		AccessKeyID:  accessKey,
		SecretKey:    secretKey,
		Region:       region,
		UseSSL:       secure,
		BucketLookup: lookup,
	}, nil
}

// NewS3StorageService constructs an S3StorageService from environment variables.
func NewS3StorageService() (*S3StorageService, error) {
	cfg, err := NewS3StorageConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return NewS3StorageServiceWithConfig(cfg)
}

// NewS3StorageServiceWithConfig constructs an S3StorageService from explicit config.
func NewS3StorageServiceWithConfig(cfg S3StorageConfig) (*S3StorageService, error) {
	if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.AccessKeyID == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("incomplete object storage configuration")
	}
	region := cfg.Region
	if region == "" {
		region = defaultS3Region
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretKey, ""),
		Secure:       cfg.UseSSL,
		Region:       region,
		BucketLookup: cfg.BucketLookup,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize object storage client: %w", err)
	}
	return &S3StorageService{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// PutObject uploads object content under key with the verified MIME content type.
func (s *S3StorageService) PutObject(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("object storage client is not initialized")
	}
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return nil
}

// DeleteObject removes an object by key from storage.
func (s *S3StorageService) DeleteObject(ctx context.Context, key string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("object storage client is not initialized")
	}
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// PresignedGetURL returns a time-limited signed URL for reading key from storage.
func (s *S3StorageService) PresignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("object storage client is not initialized")
	}
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, ttl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to presign object url: %w", err)
	}
	return u.String(), nil
}

var allowedImageContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	"image/bmp":  true,
}

// ValidateImageFile checks if the file metadata describes a valid raster image.
func ValidateImageFile(header *multipart.FileHeader) error {
	if header.Size > maxImageUploadBytes {
		return fmt.Errorf("file size exceeds maximum allowed size of 10MB")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".bmp":  true,
	}

	if !validExtensions[ext] {
		return fmt.Errorf("invalid file type: %s. Allowed types: jpg, jpeg, png, gif, webp, bmp", ext)
	}

	return nil
}

// DetectImageContentType inspects magic bytes, rewinds r to offset 0, and
// returns the verified image MIME type or an error for unsupported/unsafe payloads.
func DetectImageContentType(r io.ReadSeeker) (string, error) {
	if r == nil {
		return "", fmt.Errorf("image file is required")
	}
	buf := make([]byte, 512)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read image content: %w", err)
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset image reader: %w", err)
	}
	contentType := http.DetectContentType(buf[:n])
	if !allowedImageContentTypes[contentType] {
		return "", fmt.Errorf("invalid image content type: %s", contentType)
	}
	return contentType, nil
}

// ValidateImageContent inspects magic bytes and rewinds file to offset 0.
func ValidateImageContent(file multipart.File) error {
	_, err := DetectImageContentType(file)
	return err
}

// ReadFileContent reads the content of a multipart file.
func ReadFileContent(file multipart.File) ([]byte, error) {
	defer file.Close()
	return io.ReadAll(file)
}
