package repositories

import (
	"my-go-app/internal/models"
	"time"

	"gorm.io/gorm"
)

type MarketplaceRepository struct {
	DB *gorm.DB
}



func NewMarketplaceRepository(db *gorm.DB) *MarketplaceRepository {
	return &MarketplaceRepository{
		DB: db,
	}
}

// CreateMarketplaceItem creates a new marketplace item with tags and categories
func (r *MarketplaceRepository) CreateMarketplaceItem(item models.MarketplaceItem, tagIds []string, categoryIds []string) (*models.MarketplaceItem, error) {
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		// Create the item
		if err := tx.Create(&item).Error; err != nil {
			return err
		}

		// Associate tags
		if len(tagIds) > 0 {
			for _, tagId := range tagIds {
				itemTag := models.MarketplaceItemTag{
					ItemId: item.Id,
					TagId:  tagId,
				}
				if err := tx.Create(&itemTag).Error; err != nil {
					return err
				}
			}
		}

		// Associate categories
		if len(categoryIds) > 0 {
			for _, categoryId := range categoryIds {
				itemCategory := models.MarketplaceItemCategory{
					ItemId:     item.Id,
					CategoryId: categoryId,
				}
				if err := tx.Create(&itemCategory).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Fetch the complete item with associations
	return r.GetMarketplaceItemByID(item.Id)
}

// GetMarketplaceItems retrieves marketplace items with filtering and pagination
func (r *MarketplaceRepository) GetMarketplaceItems(filter MarketplaceFilter) ([]models.MarketplaceItem, int64, error) {
	var items []models.MarketplaceItem
	var total int64

	query := r.DB.Model(&models.MarketplaceItem{}).
		Where(`"DeletedAt" IS NULL`)

	// Apply filters
	if filter.TemplateType != "" {
		query = query.Where(`"TemplateType" = ?`, filter.TemplateType)
	}

	if filter.Featured != nil {
		query = query.Where(`"Featured" = ?`, *filter.Featured)
	}

	if filter.AuthorId != "" {
		query = query.Where(`"AuthorId" = ?`, filter.AuthorId)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where(`("Title" ILIKE ? OR "Description" ILIKE ?)`, searchPattern, searchPattern)
	}

	// Filter by category
	if filter.CategoryId != "" {
		query = query.Joins(`INNER JOIN "MarketplaceItemCategory" ON "MarketplaceItemCategory"."ItemId" = "MarketplaceItem"."Id"`).
			Where(`"MarketplaceItemCategory"."CategoryId" = ?`, filter.CategoryId)
	}

	// Filter by tag
	if filter.TagId != "" {
		query = query.Joins(`INNER JOIN "MarketplaceItemTag" ON "MarketplaceItemTag"."ItemId" = "MarketplaceItem"."Id"`).
			Where(`"MarketplaceItemTag"."TagId" = ?`, filter.TagId)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "CreatedAt"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "downloads":
			sortBy = "Downloads"
		case "likes":
			sortBy = "Likes"
		case "createdAt":
			sortBy = "CreatedAt"
		case "updatedAt":
			sortBy = "UpdatedAt"
		case "title":
			sortBy = "Title"
		}
	}

	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query = query.Order(`"` + sortBy + `" ` + sortOrder)

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Preload associations
	query = query.Preload("Tags").Preload("Categories")

	if err := query.Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetMarketplaceItemByID retrieves a single marketplace item by ID
func (r *MarketplaceRepository) GetMarketplaceItemByID(id string) (*models.MarketplaceItem, error) {
	var item models.MarketplaceItem
	err := r.DB.Where(`"Id" = ? AND "DeletedAt" IS NULL`, id).
		Preload("Tags").
		Preload("Categories").
		First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// UpdateMarketplaceItem updates a marketplace item
func (r *MarketplaceRepository) UpdateMarketplaceItem(id string, userId string, updates map[string]any) (*models.MarketplaceItem, error) {
	// Verify ownership
	var count int64
	if err := r.DB.Model(&models.MarketplaceItem{}).
		Where(`"Id" = ? AND "AuthorId" = ? AND "DeletedAt" IS NULL`, id, userId).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		// Handle tag updates
		if tagIds, ok := updates["TagIds"].([]string); ok {
			// Delete existing tags
			if err := tx.Where(`"ItemId" = ?`, id).Delete(&models.MarketplaceItemTag{}).Error; err != nil {
				return err
			}
			// Add new tags
			for _, tagId := range tagIds {
				itemTag := models.MarketplaceItemTag{
					ItemId: id,
					TagId:  tagId,
				}
				if err := tx.Create(&itemTag).Error; err != nil {
					return err
				}
			}
			delete(updates, "TagIds")
		}

		// Handle category updates
		if categoryIds, ok := updates["CategoryIds"].([]string); ok {
			// Delete existing categories
			if err := tx.Where(`"ItemId" = ?`, id).Delete(&models.MarketplaceItemCategory{}).Error; err != nil {
				return err
			}
			// Add new categories
			for _, categoryId := range categoryIds {
				itemCategory := models.MarketplaceItemCategory{
					ItemId:     id,
					CategoryId: categoryId,
				}
				if err := tx.Create(&itemCategory).Error; err != nil {
					return err
				}
			}
			delete(updates, "CategoryIds")
		}

		// Update the item fields
		if len(updates) > 0 {
			updates["UpdatedAt"] = time.Now()
			if err := tx.Model(&models.MarketplaceItem{}).
				Where(`"Id" = ?`, id).
				Updates(updates).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return r.GetMarketplaceItemByID(id)
}

// DeleteMarketplaceItem soft deletes a marketplace item
func (r *MarketplaceRepository) DeleteMarketplaceItem(id string, userId string) error {
	result := r.DB.Model(&models.MarketplaceItem{}).
		Where(`"Id" = ? AND "AuthorId" = ? AND "DeletedAt" IS NULL`, id, userId).
		Update("DeletedAt", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// IncrementDownloads increments the download count
func (r *MarketplaceRepository) IncrementDownloads(id string) error {
	return r.DB.Model(&models.MarketplaceItem{}).
		Where(`"Id" = ?`, id).
		Update("Downloads", gorm.Expr(`"Downloads" + 1`)).Error
}

// IncrementLikes increments the like count
func (r *MarketplaceRepository) IncrementLikes(id string) error {
	return r.DB.Model(&models.MarketplaceItem{}).
		Where(`"Id" = ?`, id).
		Update("Likes", gorm.Expr(`"Likes" + 1`)).Error
}

// CreateCategory creates a new category
func (r *MarketplaceRepository) CreateCategory(category models.Category) (*models.Category, error) {
	if err := r.DB.Create(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

// GetCategories retrieves all categories
func (r *MarketplaceRepository) GetCategories() ([]models.Category, error) {
	var categories []models.Category
	if err := r.DB.Order(`"Name" ASC`).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// GetCategoryByID retrieves a category by ID
func (r *MarketplaceRepository) GetCategoryByID(id string) (*models.Category, error) {
	var category models.Category
	err := r.DB.Where(`"Id" = ?`, id).First(&category).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

// GetCategoryByName retrieves a category by name
func (r *MarketplaceRepository) GetCategoryByName(name string) (*models.Category, error) {
	var category models.Category
	err := r.DB.Where(`"Name" = ?`, name).First(&category).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

// DeleteCategory deletes a category
func (r *MarketplaceRepository) DeleteCategory(id string) error {
	result := r.DB.Where(`"Id" = ?`, id).Delete(&models.Category{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CreateTag creates a new tag
func (r *MarketplaceRepository) CreateTag(tag models.Tag) (*models.Tag, error) {
	if err := r.DB.Create(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetTags retrieves all tags
func (r *MarketplaceRepository) GetTags() ([]models.Tag, error) {
	var tags []models.Tag
	if err := r.DB.Order(`"Name" ASC`).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

// GetTagByID retrieves a tag by ID
func (r *MarketplaceRepository) GetTagByID(id string) (*models.Tag, error) {
	var tag models.Tag
	err := r.DB.Where(`"Id" = ?`, id).First(&tag).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

// GetTagByName retrieves a tag by name
func (r *MarketplaceRepository) GetTagByName(name string) (*models.Tag, error) {
	var tag models.Tag
	err := r.DB.Where(`"Name" = ?`, name).First(&tag).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

// DeleteTag deletes a tag
func (r *MarketplaceRepository) DeleteTag(id string) error {
	result := r.DB.Where(`"Id" = ?`, id).Delete(&models.Tag{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
