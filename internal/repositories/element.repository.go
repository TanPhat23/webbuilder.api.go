package repositories

import (
	"encoding/json"
	"my-go-app/internal/models"
	"my-go-app/pkg/utils"

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

	err := r.DB.Table((models.Element{}).TableName() + " as e").
		Joins(`LEFT JOIN public."Setting" as s ON e."Id" = s."ElementId"`).
		Where(`e."ProjectId" = ?`, projectID).
		Order(`e."Order"`).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	var elements []models.EditorElement

	for _, row := range rows {
		element := &row.Element
		elements = append(elements, element)
	}
	if len(elements) == 0 {
		return elements, nil
	}
	return utils.BuildElementTree(elements), nil
}
