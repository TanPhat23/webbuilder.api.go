package repositories

import (
	"encoding/json"
	"my-go-app/internal/models"
	"my-go-app/pkg/utils"
	"strings"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *ElementRepository) ReplaceElements(projectID string, elements []models.EditorElement) error {
	mu := r.getProjectMutex(projectID)
	mu.Lock()
	defer mu.Unlock()

	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(TableElement.String()).
			Where(`"ProjectId" = ?`, projectID).
			Delete(&models.Element{}).Error; err != nil {
			return err
		}

		return r.createElementsAndSettings(tx, elements, projectID)
	})
}

func (r *ElementRepository) createElementsAndSettings(tx *gorm.DB, elements []models.EditorElement, projectID string) error {
	flatElements, flatSettings, err := r.flattenElementsForInsert(tx, elements, projectID)
	if err != nil {
		return err
	}

	const batchSize = 500

	// Create elements
	if len(flatElements) > 0 {
		if err := tx.Table(TableElement.String()).
			CreateInBatches(flatElements, batchSize).Error; err != nil {
			return err
		}
	}

	// Create settings
	if err := r.SettingRepository.CreateSettings(tx, flatSettings); err != nil {
		return err
	}

	return nil
}

type ElementRepository struct {
	DB               *gorm.DB
	SettingRepository SettingRepositoryInterface
	projectLocks     sync.Map
}

func (r *ElementRepository) getProjectMutex(projectID string) *sync.Mutex {
	mu, _ := r.projectLocks.LoadOrStore(projectID, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

func (r *ElementRepository) GetElements(projectID string) ([]models.EditorElement, error) {

	var elements []models.Element
	if err := r.DB.Table(TableElement.String()).
		Where(`"ProjectId" = ?`, projectID).
		Order(`"Order"`).
		Find(&elements).Error; err != nil {
		return nil, err
	}

	if len(elements) == 0 {
		return []models.EditorElement{}, nil
	}


	elementIDs := make([]string, len(elements))
	for i, elem := range elements {
		elementIDs[i] = elem.Id
	}


	settings, err := r.SettingRepository.GetSettingsByElementIDs(r.DB, elementIDs)
	if err != nil {
		return nil, err
	}


	settingsMap := make(map[string]json.RawMessage)
	for _, setting := range settings {
		settingsMap[setting.ElementId] = setting.Settings
	}


	editorElements := make([]models.EditorElement, len(elements))
	for i, elem := range elements {
		elemCopy := elem
		if settings, exists := settingsMap[elem.Id]; exists {
			elemCopy.Settings = &settings
		}
		editorElements[i] = &elemCopy
	}

	return utils.BuildElementTree(editorElements), nil
}

func (r *ElementRepository) flattenElementsForInsert(tx *gorm.DB, rootElements []models.EditorElement, projectID string) ([]models.Element, []models.Setting, error) {
	type queueItem struct {
		element  models.EditorElement
		parentID *string
	}

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

		base.ProjectId = projectID

		if base.ParentId == nil {
			hasNull = true
		}
		parentKeys[r.buildParentKey(projectID, base.ParentId)] = true

		flattened = append(flattened, *base)

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

		children := utils.GetChildrenFromEditorElement(item.element)
		for _, child := range children {
			childEditor, err := utils.ConvertToEditorElement(child)
			if err != nil {
				return nil, nil, err
			}
			parentID := base.Id
			parentKeys[r.buildParentKey(projectID, &parentID)] = true
			queue = append(queue, queueItem{
				element:  childEditor,
				parentID: &parentID,
			})
		}
	}


	var parentIDList []string
	for key := range parentKeys {
		if key != r.buildParentKey(projectID, nil) {
			parts := strings.Split(key, ":")
			if len(parts) == 2 {
				parentIDList = append(parentIDList, parts[1])
			}
		}
	}

	orderCounters := make(map[string]int)
	if len(parentIDList) > 0 || hasNull {
		var results []struct {
			ParentID *string `gorm:"column:ParentId"`
			MaxOrder int     `gorm:"column:max_order"`
		}
		q := tx.Table(TableElement.String()).
			Select(`"ParentId", COALESCE(MAX("Order"), 0) as max_order`).
			Where(`"ProjectId" = ?`, projectID)
		if len(parentIDList) > 0 {
			q = q.Where(`"ParentId" IN (?)`, parentIDList)
		}
		if hasNull {
			if len(parentIDList) > 0 {
				q = q.Or(`"ParentId" IS NULL`)
			} else {
				q = q.Where(`"ParentId" IS NULL`)
			}
		}
		err := q.Group(`"ParentId"`).Scan(&results).Error
		if err != nil {
			return nil, nil, err
		}
		for _, res := range results {
			key := r.buildParentKey(projectID, res.ParentID)
			orderCounters[key] = res.MaxOrder
		}
	}


	for key := range parentKeys {
		if _, exists := orderCounters[key]; !exists {
			orderCounters[key] = 0
		}
	}


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
