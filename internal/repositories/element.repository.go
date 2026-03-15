package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"my-go-app/pkg/utils"
	"strings"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
		Where("\"Page\".\"ProjectId\" = ?", projectID)

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

	var element models.Element

	err := r.db.WithContext(ctx).
		Preload("Page").
		Where("\"Id\" = ?", elementID).
		First(&element).Error

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

	var element models.Element

	err := r.db.WithContext(ctx).
		Preload("Page").
		Preload("Parent").
		Where("\"Id\" = ?", elementID).
		First(&element).Error

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

	var elements []models.Element

	err := r.db.WithContext(ctx).
		Where("\"PageId\" = ?", pageID).
		Order("\"Order\"").
		Find(&elements).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get elements by page ID: %w", err)
	}

	return elements, nil
}

// GetElementsByPageIds retrieves all elements for multiple pages with tree structure
func (r *ElementRepository) GetElementsByPageIds(ctx context.Context, pageIDs []string) ([]models.EditorElement, error) {
	if len(pageIDs) == 0 {
		return nil, errors.New("pageIDs is required")
	}

	var elements []models.Element

	err := r.db.WithContext(ctx).
		Where("\"PageId\" IN ?", pageIDs).
		Order("\"Order\"").
		Find(&elements).Error

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

	var elements []models.Element

	err := r.db.WithContext(ctx).
		Where("\"ParentId\" = ?", parentID).
		Order("\"Order\"").
		Find(&elements).Error

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

	var elements []models.Element

	err := r.db.WithContext(ctx).
		Where("\"Id\" IN ?", elementIDs).
		Find(&elements).Error

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
		element.Id = uuid.NewString()
	}

	if element.Type == "" {
		return errors.New("type is required")
	}

	err := r.db.WithContext(ctx).
		Create(element).Error

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

	result := r.db.WithContext(ctx).
		Model(&models.Element{}).
		Where("\"Id\" = ?", elementID).
		Update("\"EventWorkflows\"", workflows)

	if result.Error != nil {
		return fmt.Errorf("failed to update event workflows: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrElementNotFound
	}

	return nil
}

// DeleteElementByID deletes a single element by ID
func (r *ElementRepository) DeleteElementByID(ctx context.Context, elementID string) error {
	if elementID == "" {
		return errors.New("elementID is required")
	}

	result := r.db.WithContext(ctx).
		Where("\"Id\" = ?", elementID).
		Delete(&models.Element{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete element: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrElementNotFound
	}

	return nil
}

// DeleteElementsByPageID deletes all elements in a page
func (r *ElementRepository) DeleteElementsByPageID(ctx context.Context, pageID string) error {
	if pageID == "" {
		return errors.New("pageID is required")
	}

	result := r.db.WithContext(ctx).
		Where("\"PageId\" = ?", pageID).
		Delete(&models.Element{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete elements by page ID: %w", result.Error)
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
		err := tx.Where("\"PageId\" IN (SELECT \"Id\" FROM \"Page\" WHERE \"ProjectId\" = ?)", projectID).Delete(&models.Element{}).Error

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
		err := tx.Model(&models.Element{}).
			Omit("EventWorkflows").
			CreateInBatches(flatElements, DefaultBatchSize).Error

		if err != nil {
			return fmt.Errorf("failed to create elements: %w", err)
		}
	}

	return nil
}

func (r *ElementRepository) flattenElementsForInsert(ctx context.Context, tx *gorm.DB, rootElements []models.EditorElement, projectID string) ([]models.Element, error) {
	type queueItem struct {
		element  models.EditorElement
		parentID *string
	}

	// Initialize queue with root elements
	queue := make([]queueItem, 0, len(rootElements))
	parentKeys := make(map[string]bool)
	hasNull := false

	for _, e := range rootElements {
		if e != nil {
			queue = append(queue, queueItem{element: e, parentID: nil})
		}
	}

	flattened := make([]models.Element, 0, 256)

	// Process queue
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
			base.Id = uuid.NewString()
		}

		if base.ParentId != nil && *base.ParentId == "" {
			base.ParentId = nil
		}

		if item.parentID != nil {
			base.ParentId = item.parentID
		}

		if base.ParentId == nil {
			hasNull = true
		}
		parentKeys[r.buildParentKey(projectID, base.ParentId)] = true

		flattened = append(flattened, *base)

		children := utils.GetChildrenFromEditorElement(item.element)
		for _, child := range children {
			childEditor, err := utils.ConvertToEditorElement(child)
			if err != nil {
				return nil, fmt.Errorf("failed to convert child element: %w", err)
			}
			parentID := base.Id
			parentKeys[r.buildParentKey(projectID, &parentID)] = true
			queue = append(queue, queueItem{
				element:  childEditor,
				parentID: &parentID,
			})
		}
	}

	// Build parent ID list for query
	var parentIDList []string
	for key := range parentKeys {
		if key != r.buildParentKey(projectID, nil) {
			parts := strings.Split(key, ":")
			if len(parts) == 2 {
				parentIDList = append(parentIDList, parts[1])
			}
		}
	}

	// Get existing order counters
	orderCounters := make(map[string]int)
	if len(parentIDList) > 0 || hasNull {
		var results []struct {
			ParentID *string `gorm:"column:ParentId"`
			MaxOrder int     `gorm:"column:max_order"`
		}

		q := tx.Model(&models.Element{}).
			Select("\"Element\".\"ParentId\", COALESCE(MAX(\"Element\".\"Order\"), 0) as max_order").
			Where("\"PageId\" IN (SELECT \"Id\" FROM \"Page\" WHERE \"ProjectId\" = ?)", projectID)

		if len(parentIDList) > 0 {
			q = q.Where("\"Element\".\"ParentId\" IN (?)", parentIDList)
		}
		if hasNull {
			if len(parentIDList) > 0 {
				q = q.Or("\"Element\".\"ParentId\" IS NULL")
			} else {
				q = q.Where("\"Element\".\"ParentId\" IS NULL")
			}
		}

		err := q.Group("\"Element\".\"ParentId\"").Scan(&results).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get max order: %w", err)
		}

		for _, res := range results {
			key := r.buildParentKey(projectID, res.ParentID)
			orderCounters[key] = res.MaxOrder
		}
	}

	// Initialize missing counters
	for key := range parentKeys {
		if _, exists := orderCounters[key]; !exists {
			orderCounters[key] = 0
		}
	}

	// Assign order values
	for i := range flattened {
		elem := &flattened[i]
		parentID := elem.ParentId
		key := r.buildParentKey(projectID, parentID)
		orderCounters[key]++
		elem.Order = orderCounters[key]
	}

	return flattened, nil
}

func (r *ElementRepository) buildParentKey(projectID string, parentID *string) string {
	if parentID != nil {
		return projectID + ":" + *parentID
	}
	return projectID + ":root"
}
