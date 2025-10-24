package repositories

import (
	"context"
	"encoding/json"
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
	db                *gorm.DB
	settingRepository SettingRepositoryInterface
	projectLocks      sync.Map
}

func NewElementRepository(db *gorm.DB, settingRepo SettingRepositoryInterface) ElementRepositoryInterface {
	return &ElementRepository{
		db:                db,
		settingRepository: settingRepo,
	}
}

func (r *ElementRepository) getProjectMutex(projectID string) *sync.Mutex {
	mu, _ := r.projectLocks.LoadOrStore(projectID, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

func (r *ElementRepository) GetElements(ctx context.Context, projectID string) ([]models.EditorElement, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}

	var elements []models.Element

	err := r.db.WithContext(ctx).
		Model(&models.Element{}).
		Where("\"ProjectId\" = ?", projectID).
		Order("\"Order\"").
		Find(&elements).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get elements: %w", err)
	}

	if len(elements) == 0 {
		return []models.EditorElement{}, nil
	}

	// Extract element IDs
	elementIDs := make([]string, len(elements))
	for i, elem := range elements {
		elementIDs[i] = elem.Id
	}

	// Get settings for all elements
	settings, err := r.settingRepository.GetSettingsByElementIDs(ctx, r.db, elementIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	// Map settings to element IDs
	settingsMap := make(map[string]json.RawMessage)
	for _, setting := range settings {
		settingsMap[setting.ElementId] = setting.Settings
	}

	// Build editor elements with settings
	editorElements := make([]models.EditorElement, len(elements))
	for i, elem := range elements {
		elemCopy := elem
		if settingsData, exists := settingsMap[elem.Id]; exists {
			elemCopy.Settings = &settingsData
		}
		editorElements[i] = &elemCopy
	}

	// Build and return tree structure
	return utils.BuildElementTree(editorElements), nil
}

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
		err := tx.Model(&models.Element{}).
			Where("\"ProjectId\" = ?", projectID).
			Delete(&models.Element{}).Error

		if err != nil {
			return fmt.Errorf("failed to delete existing elements: %w", err)
		}

		// Create new elements and settings
		return r.createElementsAndSettings(ctx, tx, elements, projectID)
	})
}

func (r *ElementRepository) createElementsAndSettings(ctx context.Context, tx *gorm.DB, elements []models.EditorElement, projectID string) error {
	flatElements, flatSettings, err := r.flattenElementsForInsert(ctx, tx, elements, projectID)
	if err != nil {
		return err
	}

	// Create elements in batches
	if len(flatElements) > 0 {
		err := tx.Model(&models.Element{}).
			CreateInBatches(flatElements, DefaultBatchSize).Error

		if err != nil {
			return fmt.Errorf("failed to create elements: %w", err)
		}
	}

	// Create settings
	if err := r.settingRepository.CreateSettings(ctx, tx, flatSettings); err != nil {
		return fmt.Errorf("failed to create settings: %w", err)
	}

	return nil
}

func (r *ElementRepository) flattenElementsForInsert(ctx context.Context, tx *gorm.DB, rootElements []models.EditorElement, projectID string) ([]models.Element, []models.Setting, error) {
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
	settings := make([]models.Setting, 0, 128)

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

		// Generate ID if not provided
		if base.Id == "" {
			base.Id = uuid.NewString()
		}

		// Clean up empty parent ID
		if base.ParentId != nil && *base.ParentId == "" {
			base.ParentId = nil
		}

		// Set parent from queue item if provided
		if item.parentID != nil {
			base.ParentId = item.parentID
		}

		base.ProjectId = projectID

		// Track parent keys
		if base.ParentId == nil {
			hasNull = true
		}
		parentKeys[r.buildParentKey(projectID, base.ParentId)] = true

		flattened = append(flattened, *base)

		// Create setting if settings exist
		if base.Settings != nil && string(*base.Settings) != "{}" {
			setting := models.Setting{
				Id:          uuid.NewString(),
				Name:        "default",
				SettingType: base.GetType(),
				Settings:    *base.Settings,
				ElementId:   base.Id,
			}
			settings = append(settings, setting)
		}

		// Add children to queue
		children := utils.GetChildrenFromEditorElement(item.element)
		for _, child := range children {
			childEditor, err := utils.ConvertToEditorElement(child)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to convert child element: %w", err)
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
			Select("\"ParentId\", COALESCE(MAX(\"Order\"), 0) as max_order").
			Where("\"ProjectId\" = ?", projectID)

		if len(parentIDList) > 0 {
			q = q.Where("\"ParentId\" IN (?)", parentIDList)
		}
		if hasNull {
			if len(parentIDList) > 0 {
				q = q.Or("\"ParentId\" IS NULL")
			} else {
				q = q.Where("\"ParentId\" IS NULL")
			}
		}

		err := q.Group("\"ParentId\"").Scan(&results).Error
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get max order: %w", err)
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

	return flattened, settings, nil
}

func (r *ElementRepository) buildParentKey(projectID string, parentID *string) string {
	if parentID != nil {
		return projectID + ":" + *parentID
	}
	return projectID + ":root"
}

func (r *ElementRepository) DeleteElementsByProjectID(ctx context.Context, projectID string) error {
	if projectID == "" {
		return errors.New("projectID is required")
	}

	result := r.db.WithContext(ctx).
		Where("\"ProjectId\" = ?", projectID).
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
		Where("\"ProjectId\" = ?", projectID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count elements: %w", err)
	}

	return count, nil
}
