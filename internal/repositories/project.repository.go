package repositories

import (
	"database/sql"
	"encoding/json"
	"my-go-app/internal/models"
)

type ProjectRepository struct {
	*sql.DB
}

func (r *ProjectRepository) GetProjects() ([]models.Project, error) {
	const query = `
	SELECT
		p."Id", p."Name", p."Description", p."Styles", p."published", p."subdomain"
	FROM public."Projects" p
	ORDER BY p."Name"
	`
	rows, err := r.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project

	for rows.Next() {
		project := models.Project{}
		var projectStyles string
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&projectStyles,
			&project.Published,
			&project.Subdomain,
		)
		if err != nil {
			return nil, err
		}
		if projectStyles != "" {
			err = json.Unmarshal([]byte(projectStyles), &project.Styles)
			if err != nil {
				return nil, err // Handle JSON unmarshal error
			}
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func (r *ProjectRepository) GetProjectByID(projectID string, userId string) (*models.Project, error) {
	const query = `
	SELECT
		p."Id", p."Name", p."Description", p."Styles", p."published", p."subdomain"
	FROM public."Projects" p
	WHERE p."Id" = $1 AND p."OwnerId" = $2
	`
	row := r.QueryRow(query, projectID, userId)

	project := &models.Project{}
	var projectStyles string
	err := row.Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&projectStyles,
		&project.Published,
		&project.Subdomain,
	)
	if projectStyles != "" {
		err = json.Unmarshal([]byte(projectStyles), &project.Styles)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No project found
		}
		return nil, err // Other error
	}

	return project, nil
}
func (r *ProjectRepository) GetProjectsByUserID(userID string) ([]models.Project, error) {
	const query = `
	SELECT
		p."Id", p."Name", p."Description", p."Styles", p."published", p."subdomain"
	FROM public."Projects" p
	WHERE p."OwnerId" = $1
	ORDER BY p."Name"
	`
	rows, err := r.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project

	for rows.Next() {
		project := models.Project{}
		var projectStyles string
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&projectStyles,
			&project.Published,
			&project.Subdomain,
		)
		if err != nil {
			return nil, err
		}
		if projectStyles != "" {
			err = json.Unmarshal([]byte(projectStyles), &project.Styles)
			if err != nil {
				return nil, err // Handle JSON unmarshal error
			}
		}
		projects = append(projects, project)
	}

	return projects, nil
}
