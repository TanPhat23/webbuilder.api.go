package test

import (
	"context"
	"my-go-app/internal/models"
)

// MockProjectRepo implements ProjectRepositoryInterface for testing.
type MockProjectRepo struct {
	*GenericMock
}

func NewMockProjectRepo() *MockProjectRepo {
	return &MockProjectRepo{GenericMock: &GenericMock{funcs: make(map[string]any)}}
}

func (m *MockProjectRepo) SetGetPublicProjectByID(fn func(context.Context, string) (*models.Project, error)) *MockProjectRepo {
	m.Set("GetPublicProjectByID", fn)
	return m
}

func (m *MockProjectRepo) SetGetProjectByID(fn func(context.Context, string, string) (*models.Project, error)) *MockProjectRepo {
	m.Set("GetProjectByID", fn)
	return m
}

func (m *MockProjectRepo) SetGetProjectWithAccess(fn func(context.Context, string, string) (*models.Project, error)) *MockProjectRepo {
	m.Set("GetProjectWithAccess", fn)
	return m
}

func (m *MockProjectRepo) SetGetProjectsByUserID(fn func(context.Context, string) ([]models.Project, error)) *MockProjectRepo {
	m.Set("GetProjectsByUserID", fn)
	return m
}

func (m *MockProjectRepo) SetGetCollaboratorProjects(fn func(context.Context, string) ([]models.Project, error)) *MockProjectRepo {
	m.Set("GetCollaboratorProjects", fn)
	return m
}

func (m *MockProjectRepo) SetGetProjectPages(fn func(context.Context, string, string) ([]models.Page, error)) *MockProjectRepo {
	m.Set("GetProjectPages", fn)
	return m
}

func (m *MockProjectRepo) SetCreateProject(fn func(context.Context, *models.Project) error) *MockProjectRepo {
	m.Set("CreateProject", fn)
	return m
}

func (m *MockProjectRepo) SetUpdateProject(fn func(context.Context, string, string, map[string]any) (*models.Project, error)) *MockProjectRepo {
	m.Set("UpdateProject", fn)
	return m
}

func (m *MockProjectRepo) SetDeleteProject(fn func(context.Context, string, string) error) *MockProjectRepo {
	m.Set("DeleteProject", fn)
	return m
}

func (m *MockProjectRepo) SetHardDeleteProject(fn func(context.Context, string, string) error) *MockProjectRepo {
	m.Set("HardDeleteProject", fn)
	return m
}

func (m *MockProjectRepo) SetRestoreProject(fn func(context.Context, string, string) error) *MockProjectRepo {
	m.Set("RestoreProject", fn)
	return m
}

func (m *MockProjectRepo) SetExistsForUser(fn func(context.Context, string, string) (bool, error)) *MockProjectRepo {
	m.Set("ExistsForUser", fn)
	return m
}

func (m *MockProjectRepo) SetGetProjectWithLock(fn func(context.Context, string, string) (*models.Project, error)) *MockProjectRepo {
	m.Set("GetProjectWithLock", fn)
	return m
}

func (m *MockProjectRepo) GetPublicProjectByID(ctx context.Context, projectID string) (*models.Project, error) {
	if fn := m.Get("GetPublicProjectByID"); fn != nil {
		return fn.(func(context.Context, string) (*models.Project, error))(ctx, projectID)
	}
	return nil, nil
}

func (m *MockProjectRepo) GetProjectByID(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if fn := m.Get("GetProjectByID"); fn != nil {
		return fn.(func(context.Context, string, string) (*models.Project, error))(ctx, projectID, userID)
	}
	return nil, nil
}

func (m *MockProjectRepo) GetProjectWithAccess(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if fn := m.Get("GetProjectWithAccess"); fn != nil {
		return fn.(func(context.Context, string, string) (*models.Project, error))(ctx, projectID, userID)
	}
	return nil, nil
}

func (m *MockProjectRepo) GetProjectsByUserID(ctx context.Context, userID string) ([]models.Project, error) {
	if fn := m.Get("GetProjectsByUserID"); fn != nil {
		return fn.(func(context.Context, string) ([]models.Project, error))(ctx, userID)
	}
	return nil, nil
}

func (m *MockProjectRepo) GetCollaboratorProjects(ctx context.Context, userID string) ([]models.Project, error) {
	if fn := m.Get("GetCollaboratorProjects"); fn != nil {
		return fn.(func(context.Context, string) ([]models.Project, error))(ctx, userID)
	}
	return nil, nil
}

func (m *MockProjectRepo) GetProjectPages(ctx context.Context, projectID, userID string) ([]models.Page, error) {
	if fn := m.Get("GetProjectPages"); fn != nil {
		return fn.(func(context.Context, string, string) ([]models.Page, error))(ctx, projectID, userID)
	}
	return nil, nil
}

func (m *MockProjectRepo) CreateProject(ctx context.Context, project *models.Project) error {
	if fn := m.Get("CreateProject"); fn != nil {
		return fn.(func(context.Context, *models.Project) error)(ctx, project)
	}
	return nil
}

func (m *MockProjectRepo) UpdateProject(ctx context.Context, projectID, userID string, updates map[string]any) (*models.Project, error) {
	if fn := m.Get("UpdateProject"); fn != nil {
		return fn.(func(context.Context, string, string, map[string]any) (*models.Project, error))(ctx, projectID, userID, updates)
	}
	return nil, nil
}

func (m *MockProjectRepo) DeleteProject(ctx context.Context, projectID, userID string) error {
	if fn := m.Get("DeleteProject"); fn != nil {
		return fn.(func(context.Context, string, string) error)(ctx, projectID, userID)
	}
	return nil
}

func (m *MockProjectRepo) HardDeleteProject(ctx context.Context, projectID, userID string) error {
	if fn := m.Get("HardDeleteProject"); fn != nil {
		return fn.(func(context.Context, string, string) error)(ctx, projectID, userID)
	}
	return nil
}

func (m *MockProjectRepo) RestoreProject(ctx context.Context, projectID, userID string) error {
	if fn := m.Get("RestoreProject"); fn != nil {
		return fn.(func(context.Context, string, string) error)(ctx, projectID, userID)
	}
	return nil
}

func (m *MockProjectRepo) ExistsForUser(ctx context.Context, projectID, userID string) (bool, error) {
	if fn := m.Get("ExistsForUser"); fn != nil {
		return fn.(func(context.Context, string, string) (bool, error))(ctx, projectID, userID)
	}
	return false, nil
}

func (m *MockProjectRepo) GetProjectWithLock(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if fn := m.Get("GetProjectWithLock"); fn != nil {
		return fn.(func(context.Context, string, string) (*models.Project, error))(ctx, projectID, userID)
	}
	return nil, nil
}