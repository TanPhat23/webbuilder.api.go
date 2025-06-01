package repositories

import (
	"database/sql"
	"encoding/json"
	"my-go-app/internal/models"
	"my-go-app/pkg/utils"
)

type ElementRepository struct {
	*sql.DB
}

func (r *ElementRepository) GetElements(projectID string) ([]models.EditorElement, error) {
	const query = `
    SELECT
        e."Id", e."Type", e."Content", e."IsSelected",
        e."Styles", e."X", e."Y", e."Src", e."Href", e."Order",
        e."ParentId", e."ProjectId", e."Name", e."TailwindStyles",
        s."Settings"
    FROM public."Elements" e
    LEFT JOIN public."Settings" s ON e."Id" = s."ElementId"
    WHERE e."ProjectId" = $1
    ORDER BY e."Order"
    `
	rows, err := r.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var elements []models.EditorElement


	for rows.Next() {
		element := &models.Element{}
		var settings *string
		var stylesJSON string

		err := rows.Scan(
			&element.ID,
			&element.Type,
			&element.Content,
			&element.IsSelected,
			&stylesJSON,
			&element.X,
			&element.Y,
			&element.Src,
			&element.Href,
			&element.Order,
			&element.ParentID,
			&element.ProjectID,
			&element.Name,
			&element.TailwindStyles,
			&settings,
		)
		if err != nil {
			return nil, err
		}
		var settingsMap map[string]interface{}
		if settings != nil {
			err = json.Unmarshal([]byte(*settings), &settingsMap)
			if err != nil {
				return nil, err
			}
			elementWithSettings := utils.ApplyElementSetting(element , settingsMap )
			elements = append(elements, elementWithSettings)
		}else{
			elements = append(elements, element)
		}
		
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return utils.BuildElementTree(elements), nil
}
