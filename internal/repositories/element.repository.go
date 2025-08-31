package repositories

import (
	"encoding/json"
	"my-go-app/internal/models"
	"my-go-app/pkg/utils"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ElementRepository struct {
	DB *gorm.DB
}

func (r *ElementRepository) GetElements(projectID string) ([]models.EditorElement, error) {
	type elementWithSettings struct {
		models.Element
		Settings json.RawMessage `gorm:"column:Settings"`
	}

	var rows []elementWithSettings
	err := r.DB.Table(TableElement.String()+" as e").
		Select(`e.*, s."Settings"`).
		Joins(`LEFT JOIN LATERAL (
			SELECT "Settings" FROM `+TableSetting.String()+` WHERE "ElementId" = e."Id" LIMIT 1
		) s ON true`).
		Where(`e."ProjectId" = ?`, projectID).
		Order(`e."Order"`).
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}


	if len(rows) == 0 {
		return []models.EditorElement{}, nil
	}

	elements := make([]models.EditorElement, len(rows))
	for i, row := range rows {
		el := row.Element
		s := row.Settings
		el.Settings = &s
		elements[i] = &el
	}
	return utils.BuildElementTree(elements), nil
}

func (r *ElementRepository) CreateElement(elements []models.EditorElement, projectID string) error {
	if len(elements) == 0 {
		return nil
	}

	tx := r.DB.Begin()
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
		}
	}()

	flatElements, flatSettings, err := r.flattenElementsForInsert(tx, elements, projectID)
	if err != nil {
		tx.Rollback()
		return err
	}

	const batchSize = 500

	if len(flatElements) > 0 {
		if err := tx.Table(TableElement.String()).CreateInBatches(flatElements, batchSize).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if len(flatSettings) > 0 {
		if err := tx.Table(TableSetting.String()).CreateInBatches(flatSettings, batchSize).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
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

		if base.Settings != nil {
			setting := models.Setting{
				Id:          uuid.NewString(),
				Name:        "default",
				SettingType: "element",
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

	// Query for highest orders
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

	// Ensure all parentKeys have an entry
	for key := range parentKeys {
		if _, exists := orderCounters[key]; !exists {
			orderCounters[key] = 0
		}
	}

	// Assign orders
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
