package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"my-go-app/pkg/utils"
	"sync"

	"gorm.io/gorm"
	"my-go-app/pkg/ids"
	"roci.dev/fracdex"
)

var (
	ErrElementNotFound = errors.New("element not found")
)

type ElementRepository struct {
	db           *gorm.DB
	projectLocks sync.Map
}

func NewElementRepository(db *gorm.DB) ElementRepositoryInterface {
	return &ElementRepository{
		db: db,
	}
}

func (r *ElementRepository) getProjectMutex(projectID string) *sync.Mutex {
	mu, _ := r.projectLocks.LoadOrStore(projectID, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// GetElements retrieves all elements for a project with tree structure
// Optional pageID parameter filters elements by specific page
func (r *ElementRepository) GetElements(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}

	var elements []models.Element

	query := r.db.WithContext(ctx).
		Joins("Page").
		Where(`"Page"."ProjectId" = ? AND "Page"."DeletedAt" IS NULL`, projectID)

	// Add pageID filter if provided
	if len(pageID) > 0 && pageID[0] != "" {
		query = query.Where("\"Element\".\"PageId\" = ?", pageID[0])
	}

	err := query.Order("\"Element\".\"Order\"").
		Find(&elements).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get elements: %w", err)
	}

	if len(elements) == 0 {
		return []models.EditorElement{}, nil
	}

	editorElements := make([]models.EditorElement, len(elements))
	for i, elem := range elements {
		elemCopy := elem
		editorElements[i] = &elemCopy
	}

	return utils.BuildElementTree(editorElements), nil
}

// GetElementByID retrieves a single element by ID with all relations
func (r *ElementRepository) GetElementByID(ctx context.Context, elementID string) (*models.Element, error) {
	if elementID == "" {
		return nil, errors.New("elementID is required")
	}

	element, err := gorm.G[models.Element](r.db).
		Preload("Page", nil).
		Where("\"Id\" = ?", elementID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrElementNotFound
		}
		return nil, fmt.Errorf("failed to get element by ID: %w", err)
	}

	return &element, nil
}

// GetElementWithRelations retrieves element with all related data loaded
func (r *ElementRepository) GetElementWithRelations(ctx context.Context, elementID string) (*models.Element, error) {
	if elementID == "" {
		return nil, errors.New("elementID is required")
	}

	element, err := gorm.G[models.Element](r.db).
		Preload("Page", nil).
		Preload("Parent", nil).
		Where("\"Id\" = ?", elementID).
		First(ctx)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrElementNotFound
		}
		return nil, fmt.Errorf("failed to get element with relations: %w", err)
	}

	return &element, nil
}

// GetElementsByPageID retrieves all elements for a specific page
func (r *ElementRepository) GetElementsByPageID(ctx context.Context, pageID string) ([]models.Element, error) {
	if pageID == "" {
		return nil, errors.New("pageID is required")
	}

	elements, err := gorm.G[models.Element](r.db).
		Where("\"PageId\" = ?", pageID).
		Order("\"Order\"").
		Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get elements by page ID: %w", err)
	}

	return elements, nil
}

// GetElementsByPageIds retrieves the elements of several pages as a tree. Only pages
// that belong to projectID are read: page IDs from other projects match nothing.
func (r *ElementRepository) GetElementsByPageIds(ctx context.Context, projectID string, pageIDs []string) ([]models.EditorElement, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}
	if len(pageIDs) == 0 {
		return nil, errors.New("pageIDs is required")
	}

	elements, err := gorm.G[models.Element](r.db).
		Where(`"PageId" IN (SELECT "Id" FROM "Page" WHERE "ProjectId" = ? AND "Id" IN ? AND "DeletedAt" IS NULL)`, projectID, pageIDs).
		Order("\"Order\"").
		Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get elements by page IDs: %w", err)
	}

	if len(elements) == 0 {
		return []models.EditorElement{}, nil
	}

	editorElements := make([]models.EditorElement, len(elements))
	for i, elem := range elements {
		elemCopy := elem
		editorElements[i] = &elemCopy
	}

	return utils.BuildElementTree(editorElements), nil
}

// GetChildElements retrieves child elements of a parent element
func (r *ElementRepository) GetChildElements(ctx context.Context, parentID string) ([]models.Element, error) {
	if parentID == "" {
		return nil, errors.New("parentID is required")
	}

	elements, err := gorm.G[models.Element](r.db).
		Where("\"ParentId\" = ?", parentID).
		Order("\"Order\"").
		Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get child elements: %w", err)
	}

	return elements, nil
}

// GetRootElements retrieves root elements (without parent) for a project
func (r *ElementRepository) GetRootElements(ctx context.Context, projectID string) ([]models.Element, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}

	var elements []models.Element

	err := r.db.WithContext(ctx).
		Joins("Page").
		Where("\"Page\".\"ProjectId\" = ? AND \"Element\".\"ParentId\" IS NULL", projectID).
		Order("\"Element\".\"Order\"").
		Find(&elements).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get root elements: %w", err)
	}

	return elements, nil
}

// GetElementsByIDs retrieves multiple elements by their IDs
func (r *ElementRepository) GetElementsByIDs(ctx context.Context, elementIDs []string) ([]models.Element, error) {
	if len(elementIDs) == 0 {
		return []models.Element{}, nil
	}

	elements, err := gorm.G[models.Element](r.db).
		Where("\"Id\" IN ?", elementIDs).
		Find(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get elements by IDs: %w", err)
	}

	return elements, nil
}

// CreateElement creates a single element
func (r *ElementRepository) CreateElement(ctx context.Context, element *models.Element) error {
	if element == nil {
		return errors.New("element cannot be nil")
	}

	if element.Id == "" {
		id, err := ids.NewV7()
		if err != nil {
			return err
		}
		element.Id = id
	}

	if element.Type == "" {
		return errors.New("type is required")
	}

	err := gorm.G[models.Element](r.db).Create(ctx, element)

	if err != nil {
		return fmt.Errorf("failed to create element: %w", err)
	}

	return nil
}

// UpdateElement updates an element
func (r *ElementRepository) UpdateElement(ctx context.Context, element *models.Element) error {
	if element == nil {
		return errors.New("element cannot be nil")
	}

	if element.Id == "" {
		return errors.New("element ID is required")
	}

	result := r.db.WithContext(ctx).
		Model(&models.Element{}).
		Where("\"Id\" = ?", element.Id).
		Updates(element)

	if result.Error != nil {
		return fmt.Errorf("failed to update element: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrElementNotFound
	}

	return nil
}

// UpdateEventWorkflows updates the event workflows for an element
func (r *ElementRepository) UpdateEventWorkflows(ctx context.Context, elementID string, workflows []byte) error {
	if elementID == "" {
		return errors.New("elementID is required")
	}

	rows, err := gorm.G[models.Element](r.db).
		Where("\"Id\" = ?", elementID).
		Update(ctx, "\"EventWorkflows\"", workflows)

	if err != nil {
		return fmt.Errorf("failed to update event workflows: %w", err)
	}

	if rows == 0 {
		return ErrElementNotFound
	}

	return nil
}

// DeleteElementByID deletes a single element by ID
func (r *ElementRepository) DeleteElementByID(ctx context.Context, elementID string) error {
	if elementID == "" {
		return errors.New("elementID is required")
	}

	rowsAffected, err := gorm.G[models.Element](r.db).
		Where("\"Id\" = ?", elementID).
		Delete(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete element: %w", err)
	}

	if rowsAffected == 0 {
		return ErrElementNotFound
	}

	return nil
}

// DeleteElementsByPageID deletes all elements in a page
func (r *ElementRepository) DeleteElementsByPageID(ctx context.Context, pageID string) error {
	if pageID == "" {
		return errors.New("pageID is required")
	}

	_, err := gorm.G[models.Element](r.db).
		Where("\"PageId\" = ?", pageID).
		Delete(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete elements by page ID: %w", err)
	}

	return nil
}

func (r *ElementRepository) DeleteElementsByProjectID(ctx context.Context, projectID string) error {
	if projectID == "" {
		return errors.New("projectID is required")
	}

	result := r.db.WithContext(ctx).
		Joins("Page").
		Where("\"Page\".\"ProjectId\" = ?", projectID).
		Delete(&models.Element{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete elements: %w", result.Error)
	}

	return nil
}

func (r *ElementRepository) CountElementsByProjectID(ctx context.Context, projectID string) (int64, error) {
	if projectID == "" {
		return 0, errors.New("projectID is required")
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Element{}).
		Joins("Page").
		Where("\"Page\".\"ProjectId\" = ?", projectID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count elements: %w", err)
	}

	return count, nil
}

// ReplaceElements replaces all elements for a project
func (r *ElementRepository) ReplaceElements(ctx context.Context, projectID string, elements []models.EditorElement) error {
	if projectID == "" {
		return errors.New("projectID is required")
	}

	// Use project-level mutex to prevent concurrent modifications
	mu := r.getProjectMutex(projectID)
	mu.Lock()
	defer mu.Unlock()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing elements
		_, err := gorm.G[models.Element](tx).Where("\"PageId\" IN (SELECT \"Id\" FROM \"Page\" WHERE \"ProjectId\" = ?)", projectID).Delete(ctx)

		if err != nil {
			return fmt.Errorf("failed to delete existing elements: %w", err)
		}

		return r.createElementsAndSettings(ctx, tx, elements, projectID)
	})
}

func (r *ElementRepository) createElementsAndSettings(ctx context.Context, tx *gorm.DB, elements []models.EditorElement, projectID string) error {
	flatElements, err := r.flattenElementsForInsert(ctx, tx, elements, projectID)
	if err != nil {
		return err
	}

	if len(flatElements) > 0 {
		err := gorm.G[models.Element](tx).Omit("EventWorkflows").CreateInBatches(ctx, &flatElements, DefaultBatchSize)

		if err != nil {
			return fmt.Errorf("failed to create elements: %w", err)
		}
	}

	return nil
}

func (r *ElementRepository) flattenElementsForInsert(_ context.Context, _ *gorm.DB, rootElements []models.EditorElement, _ string) ([]models.Element, error) {
	flattened, err := flattenEditorElementTree(rootElements)
	if err != nil {
		return nil, err
	}
	if err := assignSiblingOrderKeys(flattened); err != nil {
		return nil, err
	}
	return flattened, nil
}

func flattenEditorElementTree(rootElements []models.EditorElement) ([]models.Element, error) {
	type queueItem struct {
		element  models.EditorElement
		parentID *string
	}

	queue := make([]queueItem, 0, len(rootElements))
	for _, e := range rootElements {
		if e != nil {
			queue = append(queue, queueItem{element: e})
		}
	}

	flattened := make([]models.Element, 0, 256)
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		if item.element == nil {
			continue
		}

		base := item.element.GetElement()
		if base == nil {
			continue
		}
		if base.Id == "" {
			id, err := ids.NewV7()
			if err != nil {
				return nil, err
			}
			base.Id = id
		}
		if base.ParentId != nil && *base.ParentId == "" {
			base.ParentId = nil
		}
		if item.parentID != nil {
			base.ParentId = item.parentID
		}
		flattened = append(flattened, *base)

		for _, child := range utils.GetChildrenFromEditorElement(item.element) {
			childEditor, err := utils.ConvertToEditorElement(child)
			if err != nil {
				return nil, fmt.Errorf("failed to convert child element: %w", err)
			}
			parentID := base.Id
			queue = append(queue, queueItem{element: childEditor, parentID: &parentID})
		}
	}
	return flattened, nil
}

// assignSiblingOrderKeys ensures every sibling group carries monotonic valid fracdex OrderKeys.
func assignSiblingOrderKeys(flattened []models.Element) error {
	type siblingGroup struct{ pageID, parentID string }
	groups := make(map[siblingGroup][]int)
	for i := range flattened {
		elem := &flattened[i]
		var group siblingGroup
		if elem.PageId != nil {
			group.pageID = *elem.PageId
		}
		if elem.ParentId != nil {
			group.parentID = *elem.ParentId
		}
		groups[group] = append(groups[group], i)
	}

	for _, indices := range groups {
		if hasValidOrderKeys(flattened, indices) {
			continue
		}
		keys, err := fracdex.NKeysBetween("", "", uint(len(indices)))
		if err != nil {
			return fmt.Errorf("failed to assign element order keys: %w", err)
		}
		for position, index := range indices {
			flattened[index].OrderKey = &keys[position]
		}
	}
	return nil
}

func hasValidOrderKeys(flattened []models.Element, indices []int) bool {
	valid := true
	previous := ""
	for position, index := range indices {
		elem := &flattened[index]
		elem.Order = position + 1
		if elem.OrderKey == nil || *elem.OrderKey == "" || (previous != "" && *elem.OrderKey <= previous) {
			valid = false
			continue
		}
		if _, err := fracdex.KeyBetween(*elem.OrderKey, ""); err != nil {
			valid = false
		}
		previous = *elem.OrderKey
	}
	return valid
}
