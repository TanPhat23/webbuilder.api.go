package services

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"strconv"
	"strings"
)

type CustomElementService struct {
	customElementRepo repositories.CustomElementRepositoryInterface
}

var customElementAllowedCols = map[string]string{
	"name": "Name", "typeId": "TypeId", "description": "Description", "category": "Category",
	"icon": "Icon", "thumbnail": "Thumbnail", "tags": "Tags", "isPublic": "IsPublic",
	"version": "Version", "structure": "Structure", "defaultProps": "DefaultProps",
}

func NewCustomElementService(customElementRepo repositories.CustomElementRepositoryInterface) *CustomElementService {
	return &CustomElementService{
		customElementRepo: customElementRepo,
	}
}

func (s *CustomElementService) GetCustomElements(ctx context.Context, userID string, isPublic *bool) ([]models.CustomElement, error) {
	if userID == "" {
		return nil, errors.New("userId is required")
	}

	return s.customElementRepo.GetCustomElements(ctx, userID, isPublic)
}

func (s *CustomElementService) GetCustomElementByID(ctx context.Context, id string, userID string) (*models.CustomElement, error) {
	if id == "" {
		return nil, errors.New("custom element id is required")
	}
	if userID == "" {
		return nil, errors.New("userId is required")
	}

	return s.customElementRepo.GetCustomElementByID(ctx, id, userID)
}

func (s *CustomElementService) CreateCustomElement(ctx context.Context, element *models.CustomElement) (*models.CustomElement, error) {
	if element == nil {
		return nil, errors.New("custom element cannot be nil")
	}
	if strings.TrimSpace(element.Name) == "" {
		return nil, errors.New("name is required")
	}
	if element.UserId == "" {
		return nil, errors.New("userId is required")
	}
	if len(element.Structure) == 0 {
		return nil, errors.New("structure is required")
	}
	if len(element.DefaultProps) == 0 {
		element.DefaultProps = []byte("{}")
	}
	if strings.TrimSpace(element.Version) == "" {
		element.Version = "1.0.0"
	}

	return s.customElementRepo.CreateCustomElement(ctx, element)
}

func (s *CustomElementService) UpdateCustomElement(ctx context.Context, id string, userID string, updates map[string]any) (*models.CustomElement, error) {
	if id == "" {
		return nil, errors.New("custom element id is required")
	}
	if userID == "" {
		return nil, errors.New("userId is required")
	}
	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	current, err := s.GetCustomElementByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, errors.New("custom element does not exist")
	}

	if name, ok := updates["name"]; ok {
		if value, ok := name.(string); ok && strings.TrimSpace(value) == "" {
			return nil, errors.New("name cannot be empty")
		}
	}
	if structure, ok := updates["structure"]; ok {
		if value, ok := structure.([]byte); ok && len(value) == 0 {
			return nil, errors.New("structure cannot be empty")
		}
	}
	if version, ok := updates["version"]; ok {
		if value, ok := version.(string); ok && strings.TrimSpace(value) == "" {
			return nil, errors.New("version cannot be empty")
		}
	}
	columns, err := utils.BuildColumnUpdates(updates, customElementAllowedCols)
	if err != nil {
		return nil, err
	}
	return s.customElementRepo.UpdateCustomElement(ctx, id, userID, columns)
}

func (s *CustomElementService) DeleteCustomElement(ctx context.Context, id string, userID string) error {
	if id == "" {
		return errors.New("custom element id is required")
	}
	if userID == "" {
		return errors.New("userId is required")
	}

	current, err := s.GetCustomElementByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if current == nil {
		return errors.New("custom element does not exist")
	}

	return s.customElementRepo.DeleteCustomElement(ctx, id, userID)
}

func (s *CustomElementService) GetPublicCustomElements(ctx context.Context, category *string, limit int, offset int) ([]models.CustomElement, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be greater than zero")
	}
	if offset < 0 {
		return nil, errors.New("offset cannot be negative")
	}
	if category != nil && strings.TrimSpace(*category) == "" {
		return nil, errors.New("category cannot be empty")
	}

	return s.customElementRepo.GetPublicCustomElements(ctx, category, limit, offset)
}

func (s *CustomElementService) DuplicateCustomElement(ctx context.Context, id string, userID string, newName string) (*models.CustomElement, error) {
	if id == "" {
		return nil, errors.New("custom element id is required")
	}
	if userID == "" {
		return nil, errors.New("userId is required")
	}
	if strings.TrimSpace(newName) == "" {
		return nil, errors.New("newName is required")
	}

	current, err := s.GetCustomElementByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, errors.New("custom element does not exist")
	}

	duplicate, err := s.customElementRepo.DuplicateCustomElement(ctx, id, userID, newName)
	if err != nil {
		return nil, err
	}

	return duplicate, nil
}

func (s *CustomElementService) ValidateCustomElementVersion(version string) error {
	if strings.TrimSpace(version) == "" {
		return errors.New("version is required")
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid version format: %s", version)
	}

	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return fmt.Errorf("invalid version format: %s", version)
		}
	}

	return nil
}
