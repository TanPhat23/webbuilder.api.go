package testutil

import (
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MockMarketplaceRepository struct {
	CreateMarketplaceItemFn   func(item models.MarketplaceItem, tagIds []string, categoryIds []string) (*models.MarketplaceItem, error)
	GetMarketplaceItemsFn     func(filter repositories.MarketplaceFilter) ([]models.MarketplaceItem, int64, error)
	GetMarketplaceItemByIDFn  func(id string) (*models.MarketplaceItem, error)
	UpdateMarketplaceItemFn   func(id string, userId string, updates map[string]any) (*models.MarketplaceItem, error)
	DeleteMarketplaceItemFn   func(id string, userId string) error
	DownloadMarketplaceItemFn func(itemId string, userId string) (*models.Project, error)
	IncrementDownloadsFn      func(id string) error
	IncrementLikesFn          func(id string) error
	CreateCategoryFn          func(category models.Category) (*models.Category, error)
	GetCategoriesFn           func() ([]models.Category, error)
	GetCategoryByIDFn         func(id string) (*models.Category, error)
	GetCategoryByNameFn       func(name string) (*models.Category, error)
	DeleteCategoryFn          func(id string) error
	CreateTagFn               func(tag models.Tag) (*models.Tag, error)
	GetTagsFn                 func() ([]models.Tag, error)
	GetTagByIDFn              func(id string) (*models.Tag, error)
	GetTagByNameFn            func(name string) (*models.Tag, error)
	DeleteTagFn               func(id string) error
}

func (m *MockMarketplaceRepository) CreateMarketplaceItem(item models.MarketplaceItem, tagIds []string, categoryIds []string) (*models.MarketplaceItem, error) {
	if m.CreateMarketplaceItemFn != nil {
		return m.CreateMarketplaceItemFn(item, tagIds, categoryIds)
	}
	return &item, nil
}

func (m *MockMarketplaceRepository) GetMarketplaceItems(filter repositories.MarketplaceFilter) ([]models.MarketplaceItem, int64, error) {
	if m.GetMarketplaceItemsFn != nil {
		return m.GetMarketplaceItemsFn(filter)
	}
	return []models.MarketplaceItem{}, 0, nil
}

func (m *MockMarketplaceRepository) GetMarketplaceItemByID(id string) (*models.MarketplaceItem, error) {
	if m.GetMarketplaceItemByIDFn != nil {
		return m.GetMarketplaceItemByIDFn(id)
	}
	return nil, nil
}

func (m *MockMarketplaceRepository) UpdateMarketplaceItem(id string, userId string, updates map[string]any) (*models.MarketplaceItem, error) {
	if m.UpdateMarketplaceItemFn != nil {
		return m.UpdateMarketplaceItemFn(id, userId, updates)
	}
	return nil, nil
}

func (m *MockMarketplaceRepository) DeleteMarketplaceItem(id string, userId string) error {
	if m.DeleteMarketplaceItemFn != nil {
		return m.DeleteMarketplaceItemFn(id, userId)
	}
	return nil
}

func (m *MockMarketplaceRepository) DownloadMarketplaceItem(itemId string, userId string) (*models.Project, error) {
	if m.DownloadMarketplaceItemFn != nil {
		return m.DownloadMarketplaceItemFn(itemId, userId)
	}
	return nil, nil
}

func (m *MockMarketplaceRepository) IncrementDownloads(id string) error {
	if m.IncrementDownloadsFn != nil {
		return m.IncrementDownloadsFn(id)
	}
	return nil
}

func (m *MockMarketplaceRepository) IncrementLikes(id string) error {
	if m.IncrementLikesFn != nil {
		return m.IncrementLikesFn(id)
	}
	return nil
}

func (m *MockMarketplaceRepository) CreateCategory(category models.Category) (*models.Category, error) {
	if m.CreateCategoryFn != nil {
		return m.CreateCategoryFn(category)
	}
	return &category, nil
}

func (m *MockMarketplaceRepository) GetCategories() ([]models.Category, error) {
	if m.GetCategoriesFn != nil {
		return m.GetCategoriesFn()
	}
	return []models.Category{}, nil
}

func (m *MockMarketplaceRepository) GetCategoryByID(id string) (*models.Category, error) {
	if m.GetCategoryByIDFn != nil {
		return m.GetCategoryByIDFn(id)
	}
	return nil, nil
}

func (m *MockMarketplaceRepository) GetCategoryByName(name string) (*models.Category, error) {
	if m.GetCategoryByNameFn != nil {
		return m.GetCategoryByNameFn(name)
	}
	return nil, nil
}

func (m *MockMarketplaceRepository) DeleteCategory(id string) error {
	if m.DeleteCategoryFn != nil {
		return m.DeleteCategoryFn(id)
	}
	return nil
}

func (m *MockMarketplaceRepository) CreateTag(tag models.Tag) (*models.Tag, error) {
	if m.CreateTagFn != nil {
		return m.CreateTagFn(tag)
	}
	return &tag, nil
}

func (m *MockMarketplaceRepository) GetTags() ([]models.Tag, error) {
	if m.GetTagsFn != nil {
		return m.GetTagsFn()
	}
	return []models.Tag{}, nil
}

func (m *MockMarketplaceRepository) GetTagByID(id string) (*models.Tag, error) {
	if m.GetTagByIDFn != nil {
		return m.GetTagByIDFn(id)
	}
	return nil, nil
}

func (m *MockMarketplaceRepository) GetTagByName(name string) (*models.Tag, error) {
	if m.GetTagByNameFn != nil {
		return m.GetTagByNameFn(name)
	}
	return nil, nil
}

func (m *MockMarketplaceRepository) DeleteTag(id string) error {
	if m.DeleteTagFn != nil {
		return m.DeleteTagFn(id)
	}
	return nil
}