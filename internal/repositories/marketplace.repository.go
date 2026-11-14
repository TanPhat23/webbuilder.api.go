package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"time"

	"gorm.io/gorm"
	"my-go-app/pkg/ids"
)

type MarketplaceRepository struct {
	db *gorm.DB
}

func NewMarketplaceRepository(db *gorm.DB) MarketplaceRepositoryInterface {
	return &MarketplaceRepository{db: db}
}

func (r *MarketplaceRepository) CreateMarketplaceItem(item models.MarketplaceItem, tagIds []string, categoryIds []string) (*models.MarketplaceItem, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Create the item
		if err := gorm.G[models.MarketplaceItem](tx).Create(tx.Statement.Context, &item); err != nil {
			return err
		}

		// Associate tags
		if len(tagIds) > 0 {
			for _, tagId := range tagIds {
				itemTag := models.MarketplaceItemTag{
					ItemId: item.Id,
					TagId:  tagId,
				}
				if err := gorm.G[models.MarketplaceItemTag](tx).Create(tx.Statement.Context, &itemTag); err != nil {
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
				if err := gorm.G[models.MarketplaceItemCategory](tx).Create(tx.Statement.Context, &itemCategory); err != nil {
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

	query := r.db.Model(&models.MarketplaceItem{}).
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
		r.db.Table(`"Tag"`).
			Select(`"MarketplaceItemTag"."ItemId", "Tag".*`).
			Joins(`INNER JOIN "MarketplaceItemTag" ON "MarketplaceItemTag"."TagId" = "Tag"."Id"`).
			Where(`"MarketplaceItemTag"."ItemId" IN ?`, itemIds).
			Scan(&itemTags)

		// Load all categories
		var itemCategories []struct {
			ItemId string
			models.Category
		}
		r.db.Table(`"Category"`).
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
	item, err := gorm.G[models.MarketplaceItem](r.db.Session(&gorm.Session{PrepareStmt: false})).Where(&models.MarketplaceItem{Id: id, DeletedAt: nil}).
		First(r.db.Statement.Context)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// Manually load Tags
	var tags []models.Tag
	r.db.Table(`"Tag"`).
		Joins(`INNER JOIN "MarketplaceItemTag" ON "MarketplaceItemTag"."TagId" = "Tag"."Id"`).
		Where(`"MarketplaceItemTag"."ItemId" = ?`, id).
		Find(&tags)
	item.Tags = tags

	// Manually load Categories
	var categories []models.Category
	r.db.Table(`"Category"`).
		Joins(`INNER JOIN "MarketplaceItemCategory" ON "MarketplaceItemCategory"."CategoryId" = "Category"."Id"`).
		Where(`"MarketplaceItemCategory"."ItemId" = ?`, id).
		Find(&categories)
	item.Categories = categories

	return &item, nil
}

func (r *MarketplaceRepository) UpdateMarketplaceItem(id string, userID string, updates map[string]any) (*models.MarketplaceItem, error) {
	// Verify ownership
	count, err := gorm.G[models.MarketplaceItem](r.db).
		Where(&models.MarketplaceItem{Id: id, AuthorId: userID, DeletedAt: nil}).
		Count(r.db.Statement.Context, "*")
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		// Handle tag updates
		if tagIds, ok := updates["TagIds"].([]string); ok {
			if _, err := gorm.G[models.MarketplaceItemTag](tx).Where(&models.MarketplaceItemTag{ItemId: id}).Delete(tx.Statement.Context); err != nil {
				return err
			}
			for _, tagId := range tagIds {
				itemTag := models.MarketplaceItemTag{ItemId: id, TagId: tagId}
				if err := gorm.G[models.MarketplaceItemTag](tx).Create(tx.Statement.Context, &itemTag); err != nil {
					return err
				}
			}
			delete(updates, "TagIds")
		}

		// Handle category updates
		if categoryIds, ok := updates["CategoryIds"].([]string); ok {
			if _, err := gorm.G[models.MarketplaceItemCategory](tx).Where(&models.MarketplaceItemCategory{ItemId: id}).Delete(tx.Statement.Context); err != nil {
				return err
			}
			for _, categoryId := range categoryIds {
				itemCategory := models.MarketplaceItemCategory{ItemId: id, CategoryId: categoryId}
				if err := gorm.G[models.MarketplaceItemCategory](tx).Create(tx.Statement.Context, &itemCategory); err != nil {
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
	rows, err := gorm.G[models.MarketplaceItem](r.db).
		Where(&models.MarketplaceItem{Id: id, AuthorId: userId, DeletedAt: nil}).
		Update(r.db.Statement.Context, "\"DeletedAt\"", time.Now())
	if err != nil {
		return err
	}
	if rows == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *MarketplaceRepository) DownloadMarketplaceItem(itemID string, userID string) (*models.Project, error) {
	// Get the marketplace item with its ProjectId
	item, err := r.GetMarketplaceItemByID(itemID)
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
	originalProject, err := gorm.G[models.Project](r.db).Table(`"Project"`).Where(`"Id" = ?`, *item.ProjectId).First(r.db.Statement.Context)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("original project not found")
		}
		return nil, err
	}

	// Prepare every ID before the first write so entropy errors cannot leave a partial clone.
	originalPages, err := gorm.G[models.Page](r.db).Table(`"Page"`).Where(`"ProjectId" = ?`, *item.ProjectId).Find(r.db.Statement.Context)
	if err != nil {
		return nil, err
	}
	var originalElements []models.Element
	if err := r.db.Table(`"Element"`).Joins("Page").Where(`"Page"."ProjectId" = ?`, *item.ProjectId).Order(`"Element"."Order" ASC`).Find(&originalElements).Error; err != nil {
		return nil, err
	}
	newProjectID, pageIdMap, elementIdMap, err := prepareCloneIDs(originalPages, originalElements, ids.NewV7)
	if err != nil {
		return nil, err
	}

	// Clone the project
	now := time.Now()
	newProject := models.Project{
		ID:          newProjectID,
		Name:        item.Title + " (Copy)",
		Description: &item.Description,
		Styles:      originalProject.Styles,
		Header:      originalProject.Header,
		Published:   false,
		Subdomain:   nil,
		OwnerId:     userID,
		AuditFields: &models.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// Create the new project
	if err := gorm.G[models.Project](r.db).Table(`"Project"`).Create(r.db.Statement.Context, &newProject); err != nil {
		return nil, err
	}

	// Clone all pages from the original project
	for _, originalPage := range originalPages {
		newPageId := pageIdMap[originalPage.Id]

		newPage := models.Page{
			Id:        newPageId,
			Name:      originalPage.Name,
			Type:      originalPage.Type,
			Styles:    originalPage.Styles,
			ProjectId: newProject.ID,
			AuditFields: models.AuditFields{
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		if err := gorm.G[models.Page](r.db).Table(`"Page"`).Create(r.db.Statement.Context, &newPage); err != nil {
			return nil, err
		}
	}

	// Clone all elements from the original project
	for _, originalElement := range originalElements {
		newElementId := elementIdMap[originalElement.Id]

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
			OrderKey:       originalElement.OrderKey,
			ParentId:       originalElement.ParentId, // Will update this next
			PageId:         newPageId,
		}

		if err := gorm.G[models.Element](r.db).Table(`"Element"`).Create(r.db.Statement.Context, &newElement); err != nil {
			return nil, err
		}
	}

	// Update ParentId references in the new elements
	for oldParentId, newParentId := range elementIdMap {
		r.db.Table(`"Element"`).
			Joins("Page").
			Where(`"Page"."ProjectId" = ? AND "Element"."ParentId" = ?`, newProject.ID, oldParentId).
			Update("ParentId", newParentId)
	}

	// Increment download count
	if err := r.IncrementDownloads(itemID); err != nil {
		return nil, err
	}

	return &newProject, nil
}

// prepareCloneIDs allocates every ID before a marketplace clone writes any rows.
func prepareCloneIDs(pages []models.Page, elements []models.Element, generate func() (string, error)) (string, map[string]string, map[string]string, error) {
	projectID, err := generate()
	if err != nil {
		return "", nil, nil, err
	}
	pageIDs := make(map[string]string, len(pages))
	for _, page := range pages {
		id, err := generate()
		if err != nil {
			return "", nil, nil, err
		}
		pageIDs[page.Id] = id
	}
	elementIDs := make(map[string]string, len(elements))
	for _, element := range elements {
		id, err := generate()
		if err != nil {
			return "", nil, nil, err
		}
		elementIDs[element.Id] = id
	}
	return projectID, pageIDs, elementIDs, nil
}

func (r *MarketplaceRepository) IncrementDownloads(id string) error {
	_, err := gorm.G[models.MarketplaceItem](r.db).
		Where(&models.MarketplaceItem{Id: id}).
		Update(r.db.Statement.Context, "\"Downloads\"", gorm.Expr(`"Downloads" + 1`))
	return err
}

func (r *MarketplaceRepository) IncrementLikes(id string) error {
	_, err := gorm.G[models.MarketplaceItem](r.db).
		Where(&models.MarketplaceItem{Id: id}).
		Update(r.db.Statement.Context, "\"Likes\"", gorm.Expr(`"Likes" + 1`))
	return err
}

func (r *MarketplaceRepository) CreateCategory(category models.Category) (*models.Category, error) {
	if err := gorm.G[models.Category](r.db).Create(r.db.Statement.Context, &category); err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *MarketplaceRepository) GetCategories() ([]models.Category, error) {
	categories, err := gorm.G[models.Category](r.db).Order(`"Name" ASC`).Find(r.db.Statement.Context)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *MarketplaceRepository) GetCategoryByID(id string) (*models.Category, error) {
	category, err := gorm.G[models.Category](r.db).Where(&models.Category{Id: id}).First(r.db.Statement.Context)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *MarketplaceRepository) GetCategoryByName(name string) (*models.Category, error) {
	category, err := gorm.G[models.Category](r.db).Where(&models.Category{Name: name}).First(r.db.Statement.Context)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *MarketplaceRepository) DeleteCategory(id string) error {
	rows, err := gorm.G[models.Category](r.db).Where(&models.Category{Id: id}).Delete(r.db.Statement.Context)
	if err != nil {
		return err
	}
	if rows == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *MarketplaceRepository) CreateTag(tag models.Tag) (*models.Tag, error) {
	if err := gorm.G[models.Tag](r.db).Create(r.db.Statement.Context, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *MarketplaceRepository) GetTags() ([]models.Tag, error) {
	tags, err := gorm.G[models.Tag](r.db).Order(`"Name" ASC`).Find(r.db.Statement.Context)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *MarketplaceRepository) GetTagByID(id string) (*models.Tag, error) {
	tag, err := gorm.G[models.Tag](r.db).Where(&models.Tag{Id: id}).First(r.db.Statement.Context)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

func (r *MarketplaceRepository) GetTagByName(name string) (*models.Tag, error) {
	tag, err := gorm.G[models.Tag](r.db).Where(&models.Tag{Name: name}).First(r.db.Statement.Context)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

func (r *MarketplaceRepository) DeleteTag(id string) error {
	rows, err := gorm.G[models.Tag](r.db).Where(&models.Tag{Id: id}).Delete(r.db.Statement.Context)
	if err != nil {
		return err
	}
	if rows == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *MarketplaceRepository) GetItemsByAuthor(ctx context.Context, authorID string) ([]models.MarketplaceItem, error) {
	if authorID == "" {
		return nil, errors.New("authorID is required")
	}
	items, err := gorm.G[models.MarketplaceItem](r.db).
		Preload("Categories", nil).
		Where(`"AuthorId" = ? AND "DeletedAt" IS NULL`, authorID).
		Order(`"CreatedAt" DESC`).
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get author marketplace items: %w", err)
	}
	return items, nil
}
