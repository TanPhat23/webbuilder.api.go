package database

import (
	"my-go-app/internal/repositories"

	"gorm.io/gorm"
)

func NewRepositories(db *gorm.DB) *repositories.RepositoriesInterface {
	return repositories.NewRepositories(db)
}
