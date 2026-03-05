package testutil

import (
	"context"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MockSnapshotRepository struct {
	SaveSnapshotFn            func(ctx context.Context, projectID string, snapshot *models.Snapshot) error
	GetSnapshotsByProjectIDFn func(ctx context.Context, projectID string) ([]models.Snapshot, error)
	GetSnapshotByIDFn         func(ctx context.Context, snapshotID string) (*models.Snapshot, error)
	DeleteSnapshotFn          func(ctx context.Context, snapshotID string) error
}

func (m *MockSnapshotRepository) SaveSnapshot(ctx context.Context, projectID string, snapshot *models.Snapshot) error {
	if m.SaveSnapshotFn != nil {
		return m.SaveSnapshotFn(ctx, projectID, snapshot)
	}
	return nil
}

func (m *MockSnapshotRepository) GetSnapshotsByProjectID(ctx context.Context, projectID string) ([]models.Snapshot, error) {
	if m.GetSnapshotsByProjectIDFn != nil {
		return m.GetSnapshotsByProjectIDFn(ctx, projectID)
	}
	return []models.Snapshot{}, nil
}

func (m *MockSnapshotRepository) GetSnapshotByID(ctx context.Context, snapshotID string) (*models.Snapshot, error) {
	if m.GetSnapshotByIDFn != nil {
		return m.GetSnapshotByIDFn(ctx, snapshotID)
	}
	return nil, repositories.ErrSnapshotNotFound
}

func (m *MockSnapshotRepository) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	if m.DeleteSnapshotFn != nil {
		return m.DeleteSnapshotFn(ctx, snapshotID)
	}
	return nil
}