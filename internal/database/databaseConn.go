package database

import "my-go-app/internal/repositories"

type Repositories struct {
	*repositories.ElementRepository
	*repositories.ProjectRepository
}

// GetRepositories returns repositories using the shared DB connection
func GetRepositories() *Repositories {
	db := GetDB()

	return &Repositories{
		ElementRepository: &repositories.ElementRepository{DB: db},
		ProjectRepository: &repositories.ProjectRepository{DB: db},
	}
}
