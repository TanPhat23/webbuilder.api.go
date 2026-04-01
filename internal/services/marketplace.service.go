package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MarketplaceService struct {
	marketplaceRepo repositories.MarketplaceRepositoryInterface
}

func NewMarketplaceService(marketplaceRepo repositories.MarketplaceRepositoryInterface) *MarketplaceService {
	return &MarketplaceService{
		marketplaceRepo: marketplaceRepo,
	}
}

func (s *MarketplaceService) CreateMarketplaceItem(ctx context.Context, item models.MarketplaceItem, tagIds []string, categoryIds []string) (*models.MarketplaceItem, error) {
	if strings.TrimSpace(item.Title) == "" {
		return nil, errors.New("title is required")
	}
	if strings.TrimSpace(item.AuthorId) == "" {
		return nil, errors.New("authorId is required")
	}
	if item.ProjectId == nil || strings.TrimSpace(*item.ProjectId) == "" {
		return nil, errors.New("projectId is required")
	}

	if len(tagIds) > 0 {
		for _, tagID := range tagIds {
			if strings.TrimSpace(tagID) == "" {
				return nil, errors.New("tagId cannot be empty")
			}
			tag, err := s.GetTagByID(ctx, tagID)
			if err != nil {
				return nil, fmt.Errorf("failed to validate tag: %w", err)
			}
			if tag == nil {
				return nil, fmt.Errorf("tag with id %s does not exist", tagID)
			}
		}
	}

	if len(categoryIds) > 0 {
		for _, categoryID := range categoryIds {
			if strings.TrimSpace(categoryID) == "" {
				return nil, errors.New("categoryId cannot be empty")
			}
			category, err := s.GetCategoryByID(ctx, categoryID)
			if err != nil {
				return nil, fmt.Errorf("failed to validate category: %w", err)
			}
			if category == nil {
				return nil, fmt.Errorf("category with id %s does not exist", categoryID)
			}
		}
	}

	if strings.TrimSpace(item.TemplateType) == "" {
		item.TemplateType = "block"
	}

	if item.PageCount == nil {
		pageCount := 0
		item.PageCount = &pageCount
	}

	item.Featured = false
	item.Verified = false

	return s.marketplaceRepo.CreateMarketplaceItem(item, tagIds, categoryIds)
}

func (s *MarketplaceService) GetMarketplaceItems(ctx context.Context, filter repositories.MarketplaceFilter) ([]models.MarketplaceItem, int64, error) {
	if filter.Limit < 0 {
		return nil, 0, errors.New("limit cannot be negative")
	}
	if filter.Offset < 0 {
		return nil, 0, errors.New("offset cannot be negative")
	}
	if filter.SortBy == "" {
		filter.SortBy = "createdAt"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}

	return s.marketplaceRepo.GetMarketplaceItems(filter)
}

func (s *MarketplaceService) GetMarketplaceItemByID(ctx context.Context, id string) (*models.MarketplaceItem, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("marketplace item id is required")
	}

	item, err := s.marketplaceRepo.GetMarketplaceItemByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("marketplace item does not exist")
	}

	return item, nil
}

func (s *MarketplaceService) UpdateMarketplaceItem(ctx context.Context, id string, item *models.MarketplaceItem, userId string) (*models.MarketplaceItem, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("marketplace item id is required")
	}
	if strings.TrimSpace(userId) == "" {
		return nil, errors.New("userId is required")
	}
	if item == nil {
		return nil, errors.New("marketplace item cannot be nil")
	}

	existing, err := s.GetMarketplaceItemByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("marketplace item does not exist")
	}
	if existing.AuthorId != "" && existing.AuthorId != userId {
		return nil, errors.New("unauthorized: user is not the author")
	}

	updates := map[string]any{
		"Title":        item.Title,
		"Description":  item.Description,
		"Preview":      item.Preview,
		"TemplateType": item.TemplateType,
		"Featured":     item.Featured,
		"PageCount":    item.PageCount,
		"ProjectId":    item.ProjectId,
	}

	if title, ok := updates["Title"].(string); ok && strings.TrimSpace(title) == "" {
		return nil, errors.New("title cannot be empty")
	}
	if templateType, ok := updates["TemplateType"].(string); ok && strings.TrimSpace(templateType) == "" {
		return nil, errors.New("templateType cannot be empty")
	}

	return s.marketplaceRepo.UpdateMarketplaceItem(id, userId, updates)
}

func (s *MarketplaceService) DeleteMarketplaceItem(ctx context.Context, id string, userId string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("marketplace item id is required")
	}
	if strings.TrimSpace(userId) == "" {
		return errors.New("userId is required")
	}

	existing, err := s.GetMarketplaceItemByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("marketplace item does not exist")
	}

	return s.marketplaceRepo.DeleteMarketplaceItem(id, userId)
}

func (s *MarketplaceService) DownloadMarketplaceItem(ctx context.Context, itemID, userID string) error {
	if strings.TrimSpace(itemID) == "" {
		return errors.New("marketplace item id is required")
	}
	if strings.TrimSpace(userID) == "" {
		return errors.New("userId is required")
	}

	_, err := s.marketplaceRepo.DownloadMarketplaceItem(itemID, userID)
	return err
}

func (s *MarketplaceService) IncrementDownloads(ctx context.Context, itemID string) error {
	if strings.TrimSpace(itemID) == "" {
		return errors.New("marketplace item id is required")
	}

	return s.marketplaceRepo.IncrementDownloads(itemID)
}

func (s *MarketplaceService) IncrementLikes(ctx context.Context, itemID string) error {
	if strings.TrimSpace(itemID) == "" {
		return errors.New("marketplace item id is required")
	}

	return s.marketplaceRepo.IncrementLikes(itemID)
}

func (s *MarketplaceService) CreateCategory(ctx context.Context, category *models.Category) (*models.Category, error) {
	if category == nil {
		return nil, errors.New("category cannot be nil")
	}
	if strings.TrimSpace(category.Name) == "" {
		return nil, errors.New("name is required")
	}

	return s.marketplaceRepo.CreateCategory(*category)
}

func (s *MarketplaceService) GetCategories(ctx context.Context) ([]models.Category, error) {
	return s.marketplaceRepo.GetCategories()
}

func (s *MarketplaceService) GetCategoryByID(ctx context.Context, id string) (*models.Category, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("category id is required")
	}

	category, err := s.marketplaceRepo.GetCategoryByID(id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, errors.New("category does not exist")
	}

	return category, nil
}

func (s *MarketplaceService) GetCategoryByName(ctx context.Context, name string) (*models.Category, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("category name is required")
	}

	return s.marketplaceRepo.GetCategoryByName(name)
}

func (s *MarketplaceService) DeleteCategory(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("category id is required")
	}

	_, err := s.GetCategoryByID(ctx, id)
	if err != nil {
		return err
	}

	return s.marketplaceRepo.DeleteCategory(id)
}

func (s *MarketplaceService) CreateTag(ctx context.Context, tag *models.Tag) (*models.Tag, error) {
	if tag == nil {
		return nil, errors.New("tag cannot be nil")
	}
	if strings.TrimSpace(tag.Name) == "" {
		return nil, errors.New("name is required")
	}

	return s.marketplaceRepo.CreateTag(*tag)
}

func (s *MarketplaceService) GetTags(ctx context.Context) ([]models.Tag, error) {
	return s.marketplaceRepo.GetTags()
}

func (s *MarketplaceService) GetTagByID(ctx context.Context, id string) (*models.Tag, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("tag id is required")
	}

	tag, err := s.marketplaceRepo.GetTagByID(id)
	if err != nil {
		return nil, err
	}
	if tag == nil {
		return nil, errors.New("tag does not exist")
	}

	return tag, nil
}

func (s *MarketplaceService) GetTagByName(ctx context.Context, name string) (*models.Tag, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("tag name is required")
	}

	return s.marketplaceRepo.GetTagByName(name)
}

func (s *MarketplaceService) DeleteTag(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("tag id is required")
	}

	_, err := s.GetTagByID(ctx, id)
	if err != nil {
		return err
	}

	return s.marketplaceRepo.DeleteTag(id)
}