package database

import (
	"my-go-app/internal/repositories"

	"gorm.io/gorm"
)

func NewRepositories(db *gorm.DB) *repositories.RepositoriesInterface {
	settingRepo := &repositories.SettingRepository{DB: db}
	return &repositories.RepositoriesInterface{
		ElementRepository:           &repositories.ElementRepository{DB: db, SettingRepository: settingRepo},
		ProjectRepository:           &repositories.ProjectRepository{DB: db},
		SnapshotRepository:          &repositories.SnapshotRepository{DB: db},
		SettingRepository:           settingRepo,
		PageRepository:              &repositories.PageRepository{DB: db},
		ContentTypeRepository:       repositories.NewContentTypeRepository(db),
		ContentFieldRepository:      repositories.NewContentFieldRepository(db),
		ContentItemRepository:       repositories.NewContentItemRepository(db),
		ContentFieldValueRepository: repositories.NewContentFieldValueRepository(db),
		ImageRepository:             repositories.NewImageRepository(db),
		MarketplaceRepository:       repositories.NewMarketplaceRepository(db),
	}
}
