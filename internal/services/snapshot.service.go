package services

import (
	"context"
	"errors"
	"fmt"

	"my-go-app/internal/authz"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type SnapshotService struct {
	snapshotRepo repositories.SnapshotRepositoryInterface
	elementRepo  repositories.ElementRepositoryInterface
	projectRepo  repositories.ProjectRepositoryInterface
}

func NewSnapshotService(
	snapshotRepo repositories.SnapshotRepositoryInterface,
	elementRepo repositories.ElementRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
) *SnapshotService {
	return &SnapshotService{
		snapshotRepo: snapshotRepo,
		elementRepo:  elementRepo,
		projectRepo:  projectRepo,
	}
}

func (s *SnapshotService) SaveSnapshot(ctx context.Context, projectID string, snapshot *models.Snapshot) error {
	if err := s.ValidateSnapshot(ctx, snapshot); err != nil {
		return err
	}
	if projectID == "" {
		return errors.New("projectId is required")
	}

	return s.snapshotRepo.SaveSnapshot(ctx, projectID, snapshot)
}

// SaveSnapshotWithAccess authorizes snapshot:create (owner or editor) before
// saving the snapshot.
func (s *SnapshotService) SaveSnapshotWithAccess(ctx context.Context, projectID, userID string, snapshot *models.Snapshot) error {
	if projectID == "" {
		return errors.New("projectId is required")
	}
	if userID == "" {
		return authz.ErrUnauthorized
	}
	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermSnapshotCreate); err != nil {
		return err
	}
	return s.SaveSnapshot(ctx, projectID, snapshot)
}

func (s *SnapshotService) GetSnapshotsByProjectID(ctx context.Context, projectID, userID string) ([]models.Snapshot, error) {
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermSnapshotView); err != nil {
		return nil, err
	}

	return s.snapshotRepo.GetSnapshotsByProjectID(ctx, projectID)
}

func (s *SnapshotService) GetSnapshotByID(ctx context.Context, id string) (*models.Snapshot, error) {
	if id == "" {
		return nil, errors.New("snapshot id is required")
	}

	return s.snapshotRepo.GetSnapshotByID(ctx, id)
}

// GetSnapshotByIDWithAccess verifies snapshot:view permission on projectID and
// confirms the snapshot belongs to projectID.
func (s *SnapshotService) GetSnapshotByIDWithAccess(ctx context.Context, projectID, snapshotID, userID string) (*models.Snapshot, error) {
	if snapshotID == "" {
		return nil, errors.New("snapshot id is required")
	}
	if projectID == "" {
		return nil, errors.New("projectId is required")
	}
	if userID == "" {
		return nil, authz.ErrUnauthorized
	}
	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermSnapshotView); err != nil {
		return nil, err
	}

	snapshot, err := s.snapshotRepo.GetSnapshotByID(ctx, snapshotID)
	if err != nil {
		return nil, err
	}
	if snapshot == nil || snapshot.ProjectId != projectID {
		return nil, fmt.Errorf("%w: snapshot not found", authz.ErrNotFound)
	}
	return snapshot, nil
}

func (s *SnapshotService) DeleteSnapshot(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("snapshot id is required")
	}

	return s.snapshotRepo.DeleteSnapshot(ctx, id)
}

func (s *SnapshotService) DeleteSnapshotWithAccess(ctx context.Context, snapshotID, projectID, userID string) error {
	if snapshotID == "" {
		return errors.New("snapshot id is required")
	}
	if projectID == "" {
		return errors.New("projectId is required")
	}
	if userID == "" {
		return errors.New("userId is required")
	}

	snapshot, err := s.snapshotRepo.GetSnapshotByID(ctx, snapshotID)
	if err != nil {
		return err
	}
	if snapshot == nil {
		return errors.New("snapshot does not exist")
	}

	if snapshot.ProjectId != projectID {
		return errors.New("snapshot does not belong to the specified project")
	}

	if _, err := authorizeProjectPerm(ctx, s.projectRepo, projectID, userID, authz.PermSnapshotDelete); err != nil {
		return err
	}

	return s.snapshotRepo.DeleteSnapshot(ctx, snapshotID)
}

func (s *SnapshotService) ValidateSnapshot(ctx context.Context, snapshot *models.Snapshot) error {
	if snapshot == nil {
		return errors.New("snapshot cannot be nil")
	}

	if snapshot.ProjectId == "" {
		return errors.New("snapshot project ID is required")
	}

	if len(snapshot.Elements) == 0 {
		return errors.New("snapshot elements are required")
	}

	return nil
}
