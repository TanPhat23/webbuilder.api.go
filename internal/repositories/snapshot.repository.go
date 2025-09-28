package repositories

import (
	"my-go-app/internal/models"

	"gorm.io/gorm"
)

type SnapshotRepository struct {
	DB *gorm.DB
}


func (r *SnapshotRepository) SaveSnapshot(projectId string, snapshot models.Snapshot) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(`"ProjectId" = ?`, projectId).Delete(&models.Snapshot{}).Error; err != nil {
			return err
		}

		return tx.Create(&snapshot).Error
	})
}
