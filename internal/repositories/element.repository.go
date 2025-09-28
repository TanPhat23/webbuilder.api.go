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

func (r *ElementRepository) UpdateElement(element models.EditorElement) error {
	if element == nil {
		return nil
	}

	base := element.GetElement()
	if base == nil {
		return gorm.ErrRecordNotFound
	}

	var settings *string
	if base.Settings != nil {
		settingsStr := string(*base.Settings)
		settings = &settingsStr
	}

	return r.DB.Transaction(func(tx *gorm.DB) error {
		updateData := map[string]any{
			"Type": base.Type,
		}

		if base.Content != nil {
			updateData["Content"] = *base.Content
		}

		if base.Name != nil {
			updateData["Name"] = *base.Name
		}

		if len(base.Styles) > 0 {
			updateData["Styles"] = base.Styles
		}

		if base.TailwindStyles != nil {
			updateData["TailwindStyles"] = *base.TailwindStyles
		}

		if base.Src != nil {
			updateData["Src"] = *base.Src
		}

		if base.Href != nil {
			updateData["Href"] = *base.Href
		}

		if err := tx.Table(TableElement.String()).
			Where(`"Id" = ?`, base.Id).
			Updates(updateData).Error; err != nil {
			return err
		}


		if err := r.updateElementSettings(tx, base.Id, settings, base.Type); err != nil {
			return err
		}

		return nil
	})
}



func (r *ElementRepository) updateElementSettings(tx *gorm.DB, elementID string, settings *string, settingType string) error {
	if settings == nil {
		return nil
	}

	if err := r.SettingRepository.DeleteSetting(tx, elementID); err != nil {
		return err
	}

	if *settings != "" {
		setting := models.Setting{
			Id:          uuid.NewString(),
			Name:        "default",
			SettingType:  settingType,
			Settings:    json.RawMessage(*settings),
			ElementId:   elementID,
		}
		return r.SettingRepository.CreateSetting(tx, setting)
	}

	return nil
}

func (r *ElementRepository) DeleteElement(elementID string) error {
	if elementID == "" {
		return gorm.ErrInvalidData
	}

	return r.DB.Transaction(func(tx *gorm.DB) error {
		elementIDs, err := r.getElementIDsForDeletion(tx, elementID)
		if err != nil {
			return err
		}

		if len(elementIDs) == 0 {
			return gorm.ErrRecordNotFound
		}

		return r.deleteElementsAndSettings(tx, elementIDs)
	})
}

func (r *ElementRepository) getElementIDsForDeletion(tx *gorm.DB, elementID string) ([]string, error) {
	var elementIDs []string
	err := tx.Table(TableElement.String()).
		Select(`"Id"`).
		Where(`"Id" = ? OR "ParentId" = ?`, elementID, elementID).
		Pluck(`"Id"`, &elementIDs).Error
	return elementIDs, err
}

func (r *ElementRepository) deleteElementsAndSettings(tx *gorm.DB, elementIDs []string) error {

	if err := r.SettingRepository.DeleteSettings(tx, elementIDs); err != nil {
		return err
	}


	return tx.Table(TableElement.String()).
		Where(`"Id" IN (?)`, elementIDs).
		Delete(&models.Element{}).Error
}

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

func (r *ElementRepository) CreateElement(elements []models.EditorElement, projectID string) error {
	if len(elements) == 0 {
		return nil
	}

	return r.DB.Transaction(func(tx *gorm.DB) error {
		return r.createElementsAndSettings(tx, elements, projectID)
	})
}

func (r *ElementRepository) InsertElementAfter(projectID string, previousElementID string, element models.EditorElement) error {
	if element == nil {
		return nil
	}

	return r.DB.Transaction(func(tx *gorm.DB) error {

		var previousElement models.Element
		if err := tx.Table(TableElement.String()).
			Where(`"Id" = ? AND "ProjectId" = ?`, previousElementID, projectID).
			First(&previousElement).Error; err != nil {
			return err
		}


		elements := []models.EditorElement{element}
		flatElements, flatSettings, err := r.flattenElementsForInsert(tx, elements, projectID)
		if err != nil {
			return err
		}

		if len(flatElements) == 0 {
			return nil
		}


		if err := r.insertElementsWithOrderUpdate(tx, flatElements, flatSettings, previousElement); err != nil {
			return err
		}

		return nil
	})
}

func (r *ElementRepository) insertElementsWithOrderUpdate(tx *gorm.DB, flatElements []models.Element, flatSettings []models.Setting, previousElement models.Element) error {
	parentID := previousElement.ParentId


	newOrder, err := r.getAndUpdateSiblingsOrder(tx, previousElement, parentID)
	if err != nil {
		return err
	}

	// Set order for new elements
	for i := range flatElements {
		elem := &flatElements[i]
		elem.ParentId = parentID
		elem.Order = newOrder
		newOrder++
	}

	// Insert elements
	if len(flatElements) > 0 {
		if err := tx.Table(TableElement.String()).
			CreateInBatches(flatElements, 500).Error; err != nil {
			return err
		}
	}

	// Insert settings
	if err := r.SettingRepository.CreateSettings(tx, flatSettings); err != nil {
		return err
	}

	return nil
}

func (r *ElementRepository) getAndUpdateSiblingsOrder(tx *gorm.DB, previousElement models.Element, parentID *string) (int, error) {
	var siblings []models.Element
	if err := tx.Table(TableElement.String()).
		Where(`"ProjectId" = ? AND "ParentId" IS NOT DISTINCT FROM ? AND "Order" > ?`,
			previousElement.ProjectId, parentID, previousElement.Order).
		Order(`"Order"`).
		Find(&siblings).Error; err != nil {
		return 0, err
	}

	newOrder := previousElement.Order + 2
	for i := range siblings {
		sibling := &siblings[i]
		if err := tx.Table(TableElement.String()).
			Where(`"Id" = ?`, sibling.Id).
			Update(`"Order"`, newOrder).Error; err != nil {
			return 0, err
		}
		newOrder++
	}

	return previousElement.Order + 1, nil
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

func (r *ElementRepository) SwapElements(projectID string, elementID1 string, elementID2 string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var elem1, elem2 models.Element
		if err := tx.Table(TableElement.String()).
			Where(`"Id" = ? AND "ProjectId" = ?`, elementID1, projectID).
			First(&elem1).Error; err != nil {
			return err
		}
		if err := tx.Table(TableElement.String()).
			Where(`"Id" = ? AND "ProjectId" = ?`, elementID2, projectID).
			First(&elem2).Error; err != nil {
			return err
		}


		if (elem1.ParentId == nil && elem2.ParentId != nil) || (elem1.ParentId != nil && elem2.ParentId == nil) || (elem1.ParentId != nil && elem2.ParentId != nil && *elem1.ParentId != *elem2.ParentId) {
			return gorm.ErrInvalidData
		}

		temp := elem1.Order
		elem1.Order = elem2.Order
		elem2.Order = temp

		if err := tx.Table(TableElement.String()).
			Where(`"Id" = ?`, elem1.Id).
			Update("Order", elem1.Order).Error; err != nil {
			return err
		}
		if err := tx.Table(TableElement.String()).
			Where(`"Id" = ?`, elem2.Id).
			Update("Order", elem2.Order).Error; err != nil {
			return err
		}

		return nil
	})
}
