package testutil

import (
	"context"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type MockProjectRepository struct {
	GetPublicProjectByIDFn    func(ctx context.Context, projectID string) (*models.Project, error)
	GetProjectByIDFn          func(ctx context.Context, projectID, userID string) (*models.Project, error)
	GetProjectWithAccessFn    func(ctx context.Context, projectID, userID string) (*models.Project, error)
	GetProjectsByUserIDFn     func(ctx context.Context, userID string) ([]models.Project, error)
	GetCollaboratorProjectsFn func(ctx context.Context, userID string) ([]models.Project, error)
	GetProjectPagesFn         func(ctx context.Context, projectID, userID string) ([]models.Page, error)
	CreateProjectFn           func(ctx context.Context, project *models.Project) error
	UpdateProjectFn           func(ctx context.Context, projectID, userID string, updates map[string]any) (*models.Project, error)
	DeleteProjectFn           func(ctx context.Context, projectID, userID string) error
	HardDeleteProjectFn       func(ctx context.Context, projectID, userID string) error
	RestoreProjectFn          func(ctx context.Context, projectID, userID string) error
	ExistsForUserFn           func(ctx context.Context, projectID, userID string) (bool, error)
	GetProjectWithLockFn      func(ctx context.Context, projectID, userID string) (*models.Project, error)
}

func (m *MockProjectRepository) GetPublicProjectByID(ctx context.Context, projectID string) (*models.Project, error) {
	if m.GetPublicProjectByIDFn != nil {
		return m.GetPublicProjectByIDFn(ctx, projectID)
	}
	return nil, repositories.ErrProjectNotFound
}

func (m *MockProjectRepository) GetProjectByID(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if m.GetProjectByIDFn != nil {
		return m.GetProjectByIDFn(ctx, projectID, userID)
	}
	return nil, repositories.ErrProjectNotFound
}

func (m *MockProjectRepository) GetProjectWithAccess(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if m.GetProjectWithAccessFn != nil {
		return m.GetProjectWithAccessFn(ctx, projectID, userID)
	}
	return nil, repositories.ErrProjectNotFound
}

func (m *MockProjectRepository) GetProjectsByUserID(ctx context.Context, userID string) ([]models.Project, error) {
	if m.GetProjectsByUserIDFn != nil {
		return m.GetProjectsByUserIDFn(ctx, userID)
	}
	return []models.Project{}, nil
}

func (m *MockProjectRepository) GetCollaboratorProjects(ctx context.Context, userID string) ([]models.Project, error) {
	if m.GetCollaboratorProjectsFn != nil {
		return m.GetCollaboratorProjectsFn(ctx, userID)
	}
	return []models.Project{}, nil
}

func (m *MockProjectRepository) GetProjectPages(ctx context.Context, projectID, userID string) ([]models.Page, error) {
	if m.GetProjectPagesFn != nil {
		return m.GetProjectPagesFn(ctx, projectID, userID)
	}
	return []models.Page{}, nil
}

func (m *MockProjectRepository) CreateProject(ctx context.Context, project *models.Project) error {
	if m.CreateProjectFn != nil {
		return m.CreateProjectFn(ctx, project)
	}
	return nil
}

func (m *MockProjectRepository) UpdateProject(ctx context.Context, projectID, userID string, updates map[string]any) (*models.Project, error) {
	if m.UpdateProjectFn != nil {
		return m.UpdateProjectFn(ctx, projectID, userID, updates)
	}
	return nil, repositories.ErrProjectNotFound
}

func (m *MockProjectRepository) DeleteProject(ctx context.Context, projectID, userID string) error {
	if m.DeleteProjectFn != nil {
		return m.DeleteProjectFn(ctx, projectID, userID)
	}
	return nil
}

func (m *MockProjectRepository) HardDeleteProject(ctx context.Context, projectID, userID string) error {
	if m.HardDeleteProjectFn != nil {
		return m.HardDeleteProjectFn(ctx, projectID, userID)
	}
	return nil
}

func (m *MockProjectRepository) RestoreProject(ctx context.Context, projectID, userID string) error {
	if m.RestoreProjectFn != nil {
		return m.RestoreProjectFn(ctx, projectID, userID)
	}
	return nil
}

func (m *MockProjectRepository) ExistsForUser(ctx context.Context, projectID, userID string) (bool, error) {
	if m.ExistsForUserFn != nil {
		return m.ExistsForUserFn(ctx, projectID, userID)
	}
	return false, nil
}

func (m *MockProjectRepository) GetProjectWithLock(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if m.GetProjectWithLockFn != nil {
		return m.GetProjectWithLockFn(ctx, projectID, userID)
	}
	return nil, repositories.ErrProjectNotFound
}