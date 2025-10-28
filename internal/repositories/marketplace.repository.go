package repositories

import (
	"fmt"
	"my-go-app/internal/models"
	"time"

	"github.com/google/uuid"
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
	if err := query.Session(&gorm.Session{PrepareStmt: false}).Count(&total).Error; err != nil {
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

	// Execute query
	if err := query.Session(&gorm.Session{PrepareStmt: false}).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	// Manually load associations for all items
	if len(items) > 0 {
		itemIds := make([]string, len(items))
		for i, item := range items {
			itemIds[i] = item.Id
		}

		// Load all tags
		var itemTags []struct {
			ItemId string
			models.Tag
		}
		r.DB.Table(`"Tag"`).
			Select(`"MarketplaceItemTag"."ItemId", "Tag".*`).
			Joins(`INNER JOIN "MarketplaceItemTag" ON "MarketplaceItemTag"."TagId" = "Tag"."Id"`).
			Where(`"MarketplaceItemTag"."ItemId" IN ?`, itemIds).
			Scan(&itemTags)

		// Load all categories
		var itemCategories []struct {
			ItemId string
			models.Category
		}
		r.DB.Table(`"Category"`).
			Select(`"MarketplaceItemCategory"."ItemId", "Category".*`).
			Joins(`INNER JOIN "MarketplaceItemCategory" ON "MarketplaceItemCategory"."CategoryId" = "Category"."Id"`).
			Where(`"MarketplaceItemCategory"."ItemId" IN ?`, itemIds).
			Scan(&itemCategories)

		// Map tags and categories to items
		for i := range items {
			for _, it := range itemTags {
				if it.ItemId == items[i].Id {
					items[i].Tags = append(items[i].Tags, it.Tag)
				}
			}
			for _, ic := range itemCategories {
				if ic.ItemId == items[i].Id {
					items[i].Categories = append(items[i].Categories, ic.Category)
				}
			}
		}
	}

	return items, total, nil
}

func (r *MarketplaceRepository) GetMarketplaceItemByID(id string) (*models.MarketplaceItem, error) {
	var item models.MarketplaceItem
	err := r.DB.Session(&gorm.Session{PrepareStmt: false}).Where(&models.MarketplaceItem{Id: id, DeletedAt: nil}).
		First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// Manually load Tags
	var tags []models.Tag
	r.DB.Table(`"Tag"`).
		Joins(`INNER JOIN "MarketplaceItemTag" ON "MarketplaceItemTag"."TagId" = "Tag"."Id"`).
		Where(`"MarketplaceItemTag"."ItemId" = ?`, id).
		Find(&tags)
	item.Tags = tags

	// Manually load Categories
	var categories []models.Category
	r.DB.Table(`"Category"`).
		Joins(`INNER JOIN "MarketplaceItemCategory" ON "MarketplaceItemCategory"."CategoryId" = "Category"."Id"`).
		Where(`"MarketplaceItemCategory"."ItemId" = ?`, id).
		Find(&categories)
	item.Categories = categories

	return &item, nil
}

func (r *MarketplaceRepository) UpdateMarketplaceItem(id string, userId string, updates map[string]any) (*models.MarketplaceItem, error) {
	// Verify ownership
	var count int64
	if err := r.DB.Model(&models.MarketplaceItem{}).
		Where(&models.MarketplaceItem{Id: id, AuthorId: userId, DeletedAt: nil}).
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
			if err := tx.Where(&models.MarketplaceItemTag{ItemId: id}).Delete(&models.MarketplaceItemTag{}).Error; err != nil {
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
			if err := tx.Where(&models.MarketplaceItemCategory{ItemId: id}).Delete(&models.MarketplaceItemCategory{}).Error; err != nil {
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
				Where(&models.MarketplaceItem{Id: id}).
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

func (r *MarketplaceRepository) DeleteMarketplaceItem(id string, userId string) error {
	result := r.DB.Model(&models.MarketplaceItem{}).
		Where(&models.MarketplaceItem{Id: id, AuthorId: userId, DeletedAt: nil}).
		Update("\"DeletedAt\"", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *MarketplaceRepository) DownloadMarketplaceItem(itemId string, userId string) (*models.Project, error) {
	// Get the marketplace item with its ProjectId
	item, err := r.GetMarketplaceItemByID(itemId)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("marketplace item not found")
	}

	// Check if the item has a ProjectId
	if item.ProjectId == nil || *item.ProjectId == "" {
		return nil, fmt.Errorf("marketplace item does not have an associated project")
	}

	// Get the original project
	var originalProject models.Project
	if err := r.DB.Table(`"Project"`).Where(`"Id" = ?`, *item.ProjectId).First(&originalProject).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("original project not found")
		}
		return nil, err
	}

	// Clone the project
	now := time.Now()
	newProject := models.Project{
		ID:         	uuid.New().String(),
		Name:        item.Title + " (Copy)",
		Description: &item.Description,
		Styles:      originalProject.Styles,
		Header:      originalProject.Header,
		Published:   false,
		Subdomain:   nil,
		OwnerId:     userId,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Create the new project
	if err := r.DB.Table(`"Project"`).Create(&newProject).Error; err != nil {
		return nil, err
	}

	// Clone all pages from the original project
	var originalPages []models.Page
	if err := r.DB.Table(`"Page"`).Where(`"ProjectId" = ?`, *item.ProjectId).Find(&originalPages).Error; err != nil {
		return nil, err
	}

	pageIdMap := make(map[string]string)

	for _, originalPage := range originalPages {
		newPageId := uuid.New().String()
		pageIdMap[originalPage.Id] = newPageId

		newPage := models.Page{
			Id:        newPageId,
			Name:      originalPage.Name,
			Type:      originalPage.Type,
			Styles:    originalPage.Styles,
			ProjectId: newProject.ID,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := r.DB.Table(`"Page"`).Create(&newPage).Error; err != nil {
			return nil, err
		}
	}

	// Clone all elements from the original project
	var originalElements []models.Element
	if err := r.DB.Table(`"Element"`).Where(`"ProjectId" = ?`, *item.ProjectId).Order(`"Order" ASC`).Find(&originalElements).Error; err != nil {
		return nil, err
	}

	elementIdMap := make(map[string]string) // old element id -> new element id

	for _, originalElement := range originalElements {
		newElementId := uuid.New().String()
		elementIdMap[originalElement.Id] = newElementId

		// Update PageId if it exists
		var newPageId *string
		if originalElement.PageId != nil {
			if mappedPageId, ok := pageIdMap[*originalElement.PageId]; ok {
				newPageId = &mappedPageId
			}
		}

		// Update ParentId if it exists (will be updated in second pass)
		newElement := models.Element{
			Id:             newElementId,
			Name:           originalElement.Name,
			Type:           originalElement.Type,
			Content:        originalElement.Content,
			Href:           originalElement.Href,
			Src:            originalElement.Src,
			Styles:         originalElement.Styles,
			TailwindStyles: originalElement.TailwindStyles,
			Order:          originalElement.Order,
			ParentId:       originalElement.ParentId, // Will update this next
			PageId:         newPageId,
			ProjectId:      newProject.ID,
		}

		if err := r.DB.Table(`"Element"`).Create(&newElement).Error; err != nil {
			return nil, err
		}
	}

	// Update ParentId references in the new elements
	for oldParentId, newParentId := range elementIdMap {
		r.DB.Table(`"Element"`).
			Where(`"ProjectId" = ? AND "ParentId" = ?`, newProject.ID, oldParentId).
			Update("ParentId", newParentId)
	}

	// Increment download count
	if err := r.IncrementDownloads(itemId); err != nil {
		return nil, err
	}

	return &newProject, nil
}

func (r *MarketplaceRepository) IncrementDownloads(id string) error {
	return r.DB.Model(&models.MarketplaceItem{}).
		Where(&models.MarketplaceItem{Id: id}).
		Update("\"Downloads\"", gorm.Expr(`"Downloads" + 1`)).Error
}

func (r *MarketplaceRepository) IncrementLikes(id string) error {
	return r.DB.Model(&models.MarketplaceItem{}).
		Where(&models.MarketplaceItem{Id: id}).
		Update("\"Likes\"", gorm.Expr(`"Likes" + 1`)).Error
}

func (r *MarketplaceRepository) CreateCategory(category models.Category) (*models.Category, error) {
	if err := r.DB.Create(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *MarketplaceRepository) GetCategories() ([]models.Category, error) {
	var categories []models.Category
	if err := r.DB.Order(`"Name" ASC`).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *MarketplaceRepository) GetCategoryByID(id string) (*models.Category, error) {
	var category models.Category
	err := r.DB.Where(&models.Category{Id: id}).First(&category).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *MarketplaceRepository) GetCategoryByName(name string) (*models.Category, error) {
	var category models.Category
	err := r.DB.Where(&models.Category{Name: name}).First(&category).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *MarketplaceRepository) DeleteCategory(id string) error {
	result := r.DB.Where(&models.Category{Id: id}).Delete(&models.Category{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *MarketplaceRepository) CreateTag(tag models.Tag) (*models.Tag, error) {
	if err := r.DB.Create(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *MarketplaceRepository) GetTags() ([]models.Tag, error) {
	var tags []models.Tag
	if err := r.DB.Order(`"Name" ASC`).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *MarketplaceRepository) GetTagByID(id string) (*models.Tag, error) {
	var tag models.Tag
	err := r.DB.Where(&models.Tag{Id: id}).First(&tag).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

func (r *MarketplaceRepository) GetTagByName(name string) (*models.Tag, error) {
	var tag models.Tag
	err := r.DB.Where(&models.Tag{Name: name}).First(&tag).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

func (r *MarketplaceRepository) DeleteTag(id string) error {
	result := r.DB.Where(&models.Tag{Id: id}).Delete(&models.Tag{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
