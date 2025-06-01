package database

import "my-go-app/internal/repositories"


type Repositories struct {
	*repositories.ElementRepository
}

// GetRepositories returns repositories using the shared DB connection
func GetRepositories() *Repositories {
	db := GetDB()

	return &Repositories{
		ElementRepository:  &repositories.ElementRepository{DB: db},
	}
}
