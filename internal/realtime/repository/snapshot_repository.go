package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
)

type snapshotRepository struct {
	db *gorm.DB
}

// NewSnapshotRepository creates a new instance of SnapshotRepository
func NewSnapshotRepository(db *gorm.DB) domainRepo.SnapshotRepository {
	return &snapshotRepository{db: db}
}

// CreateSnapshot persists a new snapshot to the database, verifying page ownership inside the transaction.
func (r *snapshotRepository) CreateSnapshot(ctx context.Context, projectID, pageID, name string, snapshotType entity.SnapshotType, elements []entity.Element) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := verifyPageOwnership(ctx, tx, projectID, pageID); err != nil {
			return err
		}
		snapshot := entity.Snapshot{
			ID:        uuid.New().String(),
			Name:      name,
			Type:      snapshotType,
			PageID:    pageID,
			ProjectID: projectID,
			Timestamp: time.Now().UnixMilli(),
			Elements:  elements,
			CreatedAt: time.Now(),
		}
		if err := tx.Create(&snapshot).Error; err != nil {
			return fmt.Errorf("failed to save snapshot: %w", err)
		}
		return nil
	})
}

// GetSnapshot retrieves the latest snapshot for a given project and page, verifying page ownership.
func (r *snapshotRepository) GetSnapshot(ctx context.Context, projectID, pageID string, snapshotType entity.SnapshotType) (*entity.Snapshot, error) {
	var snapshot entity.Snapshot
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := verifyPageOwnership(ctx, tx, projectID, pageID); err != nil {
			return err
		}
		return tx.
			Table(entity.Snapshot{}.TableName()).
			Where(`"ProjectId" = ? AND "PageId" = ? AND "Type" = ?`, projectID, pageID, snapshotType).
			Order(`"Timestamp" DESC`).
			First(&snapshot).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}
	return &snapshot, nil
}

// DeleteSnapshot removes snapshots for a given page and project, verifying page ownership inside the transaction.
func (r *snapshotRepository) DeleteSnapshot(ctx context.Context, pageID, projectID string, snapshotType entity.SnapshotType) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := verifyPageOwnership(ctx, tx, projectID, pageID); err != nil {
			return err
		}
		if err := tx.
			Table(entity.Snapshot{}.TableName()).
			Where(`"ProjectId" = ? AND "PageId" = ? AND "Type" = ?`, projectID, pageID, snapshotType).
			Delete(&entity.Snapshot{}).Error; err != nil {
			return fmt.Errorf("failed to delete snapshot: %w", err)
		}
		return nil
	})
}
