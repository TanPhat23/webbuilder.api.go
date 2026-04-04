package test

import (
	"context"
	"mime/multipart"
	"my-go-app/internal/services"
)

type MockCloudinaryService struct {
	*GenericMock
}

func NewMockCloudinaryService() *MockCloudinaryService {
	return &MockCloudinaryService{GenericMock: &GenericMock{funcs: make(map[string]any)}}
}

func (m *MockCloudinaryService) SetUploadImage(fn func(context.Context, multipart.File, string, string) (*services.UploadResult, error)) *MockCloudinaryService {
	m.Set("UploadImage", fn)
	return m
}

func (m *MockCloudinaryService) SetDeleteImage(fn func(context.Context, string) error) *MockCloudinaryService {
	m.Set("DeleteImage", fn)
	return m
}

func (m *MockCloudinaryService) SetUploadBase64Image(fn func(context.Context, string, string) (*services.UploadResult, error)) *MockCloudinaryService {
	m.Set("UploadBase64Image", fn)
	return m
}

func (m *MockCloudinaryService) UploadImage(ctx context.Context, file multipart.File, filename string, folder string) (*services.UploadResult, error) {
	if fn := m.Get("UploadImage"); fn != nil {
		return fn.(func(context.Context, multipart.File, string, string) (*services.UploadResult, error))(ctx, file, filename, folder)
	}
	return &services.UploadResult{
		PublicID:  "test-public-id",
		SecureURL: "https://example.com/image.jpg",
		URL:       "http://example.com/image.jpg",
		Format:    "jpg",
		Width:     800,
		Height:    600,
	}, nil
}

func (m *MockCloudinaryService) DeleteImage(ctx context.Context, publicID string) error {
	if fn := m.Get("DeleteImage"); fn != nil {
		return fn.(func(context.Context, string) error)(ctx, publicID)
	}
	return nil
}

func (m *MockCloudinaryService) UploadBase64Image(ctx context.Context, base64Data string, folder string) (*services.UploadResult, error) {
	if fn := m.Get("UploadBase64Image"); fn != nil {
		return fn.(func(context.Context, string, string) (*services.UploadResult, error))(ctx, base64Data, folder)
	}
	return &services.UploadResult{
		PublicID:  "test-public-id",
		SecureURL: "https://example.com/image.jpg",
		URL:       "http://example.com/image.jpg",
		Format:    "jpg",
		Width:     800,
		Height:    600,
	}, nil
}