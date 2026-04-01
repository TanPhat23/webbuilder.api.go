package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type ContentFieldService struct {
	contentFieldRepo repositories.ContentFieldRepositoryInterface
}

func NewContentFieldService(contentFieldRepo repositories.ContentFieldRepositoryInterface) *ContentFieldService {
	return &ContentFieldService{
		contentFieldRepo: contentFieldRepo,
	}
}

func (s *ContentFieldService) GetContentFieldsByContentType(ctx context.Context, contentTypeID string) ([]models.ContentField, error) {
	if strings.TrimSpace(contentTypeID) == "" {
		return nil, errors.New("contentTypeId is required")
	}

	return s.contentFieldRepo.GetContentFieldsByContentType(ctx, contentTypeID)
}

func (s *ContentFieldService) GetContentFieldByID(ctx context.Context, id string) (*models.ContentField, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("content field id is required")
	}

	field, err := s.contentFieldRepo.GetContentFieldByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if field == nil {
		return nil, errors.New("content field does not exist")
	}

	return field, nil
}

func (s *ContentFieldService) CreateContentField(ctx context.Context, field *models.ContentField) (*models.ContentField, error) {
	if field == nil {
		return nil, errors.New("content field cannot be nil")
	}
	if strings.TrimSpace(field.ContentTypeId) == "" {
		return nil, errors.New("contentTypeId is required")
	}
	if strings.TrimSpace(field.Name) == "" {
		return nil, errors.New("name is required")
	}
	if strings.TrimSpace(field.Type) == "" {
		return nil, errors.New("type is required")
	}

	created, err := s.contentFieldRepo.CreateContentField(ctx, field)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *ContentFieldService) UpdateContentField(ctx context.Context, id string, updates map[string]any) (*models.ContentField, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("content field id is required")
	}
	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	current, err := s.GetContentFieldByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, errors.New("content field does not exist")
	}

	if name, ok := updates["Name"]; ok {
		if value, ok := name.(string); ok && strings.TrimSpace(value) == "" {
			return nil, errors.New("name cannot be empty")
		}
	}
	if typ, ok := updates["Type"]; ok {
		if value, ok := typ.(string); ok && strings.TrimSpace(value) == "" {
			return nil, errors.New("type cannot be empty")
		}
	}
	if required, ok := updates["Required"]; ok {
		if _, ok := required.(bool); !ok {
			return nil, fmt.Errorf("required must be a boolean")
		}
	}

	updated, err := s.contentFieldRepo.UpdateContentField(ctx, id, updates)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *ContentFieldService) DeleteContentField(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("content field id is required")
	}

	current, err := s.GetContentFieldByID(ctx, id)
	if err != nil {
		return err
	}
	if current == nil {
		return errors.New("content field does not exist")
	}

	return s.contentFieldRepo.DeleteContentField(ctx, id)
}