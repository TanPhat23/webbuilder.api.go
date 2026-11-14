package repository

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	domainRepo "my-go-app/internal/realtime/domain/repository"
)

// NewRepository creates a new instance of domain Repository struct that contains all sub-repositories.
// It requires both a GORM DB connection for persistent storage and a Redis client for caching/real-time document state.
func NewRepository(db *gorm.DB, rdb *redis.Client) *domainRepo.Repository {
	return &domainRepo.Repository{
		LogicRepo:           NewLogicRepository(db),
		ElementRepo:         NewElementRepository(db),
		SnapshotRepo:        NewSnapshotRepository(db),
		DocumentRepo:        NewDocumentRepository(rdb),
		ProjectRepo:         NewProjectRepository(db),
		WorkflowRepo:        NewWorkflowRepository(db),
		ElementWorkflowRepo: NewElementWorkflowRepository(db),
	}
}
