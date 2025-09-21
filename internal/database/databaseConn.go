package database

import (
	"my-go-app/internal/repositories"

	"gorm.io/gorm"
)

func NewRepositories(db *gorm.DB) *repositories.RepositoriesInterface {
	return &repositories.RepositoriesInterface{
		ElementRepository: &repositories.ElementRepository{DB: db},
		ProjectRepository: &repositories.ProjectRepository{DB: db},
	}
}
