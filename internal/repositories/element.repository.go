package repositories

import (
	"database/sql"
	"my-go-app/internal/models"
)

type ElementRepository struct {
	*sql.DB
}

func (r *ElementRepository) GetElements(projectID string) ([]*models.Element, error) {
	query := `
		SELECT id, type, content, is_selected, name, styles, tailwind_styles, 
		       x, y, src, href, parent_id, project_id 
		FROM elements 
		WHERE project_id = $1
	`

	rows, err := r.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var elements []*models.Element

	for rows.Next() {
		element := &models.Element{}
		err := rows.Scan(
			&element.ID,
			&element.Type,
			&element.Content,
			&element.IsSelected,
			&element.Name,
			&element.Styles,
			&element.TailwindStyles,
			&element.X,
			&element.Y,
			&element.Src,
			&element.Href,
			&element.ParentID,
			&element.ProjectID,
		)
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return elements, nil
}
