package repositories

import (
	"context"
	"errors"
	"fmt"
	"my-go-app/internal/models"
	"time"

	"gorm.io/gorm"
)

var (
	ErrSnapshotNotFound = errors.New("snapshot not found")
	ErrInvalidSnapshotType = errors.New("invalid snapshot type")
)

const (
	SnapshotTypeWorking = "working"
	SnapshotTypeSaved = "saved"
	SnapshotTypePublished = "published"
)

	type SnapshotRepository struct {
		db *gorm.DB
	}

	func NewSnapshotRepository(db *gorm.DB) SnapshotRepositoryInterface {
		return &SnapshotRepository{db: db}
	}

	func (r *SnapshotRepository) SaveSnapshot(ctx context.Context, projectID string, snapshot *models.Snapshot) error {
	if snapshot == nil {
		return errors.New("snapshot cannot be nil")
	}

	if projectID == "" {
		return errors.New("projectID is required")
	}

	if snapshot.Type == "" {
		return errors.New("snapshot type is required")
	}

	// Validate snapshot type
	if !isValidSnapshotType(snapshot.Type) {
		return ErrInvalidSnapshotType
	}

	snapshot.ProjectId = projectID

	// For working snapshots, update if exists, otherwise create
	if snapshot.Type == SnapshotTypeWorking {
		return r.upsertWorkingSnapshot(ctx, snapshot)
	}

	// For other snapshot types, always create new
	return r.createSnapshot(ctx, snapshot)
}

	func (r *SnapshotRepository) upsertWorkingSnapshot(ctx context.Context, snapshot *models.Snapshot) error {
	var existing models.Snapshot

	err := r.db.WithContext(ctx).
		Model(&models.Snapshot{}).
		Where("\"ProjectId\" = ? AND \"Type\" = ?", snapshot.ProjectId, SnapshotTypeWorking).
		Order("\"CreatedAt\" DESC").
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.createSnapshot(ctx, snapshot)
		}
		return fmt.Errorf("failed to check existing working snapshot: %w", err)
	}

	snapshot.Id = existing.Id
	snapshot.CreatedAt = existing.CreatedAt

	err = r.db.WithContext(ctx).
		Model(&models.Snapshot{}).
		Where("\"Id\" = ?", existing.Id).
		Updates(snapshot).Error

	if err != nil {
		return fmt.Errorf("failed to update working snapshot: %w", err)
	}

	return nil
}

	func (r *SnapshotRepository) createSnapshot(ctx context.Context, snapshot *models.Snapshot) error {
	now := time.Now()
	if snapshot.CreatedAt.IsZero() {
		snapshot.CreatedAt = now
	}

	err := r.db.WithContext(ctx).
		Model(&models.Snapshot{}).
		Create(snapshot).Error

	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	return nil
}

	func (r *SnapshotRepository) GetSnapshotsByProjectID(ctx context.Context, projectID string) ([]models.Snapshot, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}

	var snapshots []models.Snapshot

	err := r.db.WithContext(ctx).
		Model(&models.Snapshot{}).
		Where("\"ProjectId\" = ?", projectID).
		Order("\"CreatedAt\" DESC").
		Find(&snapshots).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots by project ID: %w", err)
	}

	return snapshots, nil
}

	func (r *SnapshotRepository) GetSnapshotsByProjectIDAndType(ctx context.Context, projectID, snapshotType string) ([]models.Snapshot, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}

	if snapshotType == "" {
		return nil, errors.New("snapshot type is required")
	}

	if !isValidSnapshotType(snapshotType) {
		return nil, ErrInvalidSnapshotType
	}

	var snapshots []models.Snapshot

	err := r.db.WithContext(ctx).
		Model(&models.Snapshot{}).
		Where("\"ProjectId\" = ? AND \"Type\" = ?", projectID, snapshotType).
		Order("\"CreatedAt\" DESC").
		Find(&snapshots).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots by project ID and type: %w", err)
	}

	return snapshots, nil
}

	func (r *SnapshotRepository) GetSnapshotByID(ctx context.Context, snapshotID string) (*models.Snapshot, error) {
	if snapshotID == "" {
		return nil, errors.New("snapshotID is required")
	}

	var snapshot models.Snapshot

	err := r.db.WithContext(ctx).
		Model(&models.Snapshot{}).
		Where("\"Id\" = ?", snapshotID).
		First(&snapshot).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSnapshotNotFound
		}
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	return &snapshot, nil
}

	func (r *SnapshotRepository) GetLatestWorkingSnapshot(ctx context.Context, projectID string) (*models.Snapshot, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}

	var snapshot models.Snapshot

	err := r.db.WithContext(ctx).
		Model(&models.Snapshot{}).
		Where("\"ProjectId\" = ? AND \"Type\" = ?", projectID, SnapshotTypeWorking).
		Order("\"CreatedAt\" DESC").
		First(&snapshot).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSnapshotNotFound
		}
		return nil, fmt.Errorf("failed to get latest working snapshot: %w", err)
	}

	return &snapshot, nil
}

	func (r *SnapshotRepository) GetLatestPublishedSnapshot(ctx context.Context, projectID string) (*models.Snapshot, error) {
	if projectID == "" {
		return nil, errors.New("projectID is required")
	}

	var snapshot models.Snapshot

	err := r.db.WithContext(ctx).
		Model(&models.Snapshot{}).
		Where("\"ProjectId\" = ? AND \"Type\" = ?", projectID, SnapshotTypePublished).
		Order("\"CreatedAt\" DESC").
		First(&snapshot).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSnapshotNotFound
		}
		return nil, fmt.Errorf("failed to get latest published snapshot: %w", err)
	}

	return &snapshot, nil
}

	func (r *SnapshotRepository) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	if snapshotID == "" {
		return errors.New("snapshotID is required")
	}

	result := r.db.WithContext(ctx).
		Where("\"Id\" = ?", snapshotID).
		Delete(&models.Snapshot{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete snapshot: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrSnapshotNotFound
	}

	return nil
}

	func (r *SnapshotRepository) DeleteSnapshotsByProjectID(ctx context.Context, projectID string) error {
	if projectID == "" {
		return errors.New("projectID is required")
	}

	result := r.db.WithContext(ctx).
		Where("\"ProjectId\" = ?", projectID).
		Delete(&models.Snapshot{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete snapshots by project ID: %w", result.Error)
	}

	return nil
}

	func (r *SnapshotRepository) DeleteOldSnapshots(ctx context.Context, projectID string, olderThan time.Duration) error {
	if projectID == "" {
		return errors.New("projectID is required")
	}

	cutoffTime := time.Now().Add(-olderThan)

	result := r.db.WithContext(ctx).
		Where("\"ProjectId\" = ? AND \"Type\" != ? AND \"CreatedAt\" < ?", projectID, SnapshotTypeWorking, cutoffTime).
		Delete(&models.Snapshot{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete old snapshots: %w", result.Error)
	}

	return nil
}

	func (r *SnapshotRepository) CountSnapshotsByProjectID(ctx context.Context, projectID string) (int64, error) {
	if projectID == "" {
		return 0, errors.New("projectID is required")
	}

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Snapshot{}).
		Where("\"ProjectId\" = ?", projectID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count snapshots: %w", err)
	}

	return count, nil
}

	func (r *SnapshotRepository) CountSnapshotsByType(ctx context.Context, projectID, snapshotType string) (int64, error) {
	if projectID == "" {
		return 0, errors.New("projectID is required")
	}

	if snapshotType == "" {
		return 0, errors.New("snapshot type is required")
	}

	if !isValidSnapshotType(snapshotType) {
		return 0, ErrInvalidSnapshotType
	}

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Snapshot{}).
		Where("\"ProjectId\" = ? AND \"Type\" = ?", projectID, snapshotType).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count snapshots by type: %w", err)
	}

	return count, nil
}

	func (r *SnapshotRepository) ExistsByID(ctx context.Context, snapshotID string) (bool, error) {
	if snapshotID == "" {
		return false, errors.New("snapshotID is required")
	}

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Snapshot{}).
		Where("\"Id\" = ?", snapshotID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check snapshot existence: %w", err)
	}

	return count > 0, nil
}

	func isValidSnapshotType(snapshotType string) bool {
	switch snapshotType {
	case SnapshotTypeWorking, SnapshotTypeSaved, SnapshotTypePublished:
		return true
	default:
		return false
	}
}
