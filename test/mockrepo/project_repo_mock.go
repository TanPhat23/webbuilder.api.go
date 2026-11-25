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

func (m *MockProjectRepo) SetGetPublicByID(fn func(context.Context, string) (*models.Project, error)) *MockProjectRepo {
	m.Set("GetPublicByID", fn)
	return m
}

func (m *MockProjectRepo) SetGetPublicBySubdomain(fn func(context.Context, string) (*models.Project, error)) *MockProjectRepo {
	m.Set("GetPublicBySubdomain", fn)
	return m
}

func (m *MockProjectRepo) SetGetByID(fn func(context.Context, string, string) (*models.Project, error)) *MockProjectRepo {
	m.Set("GetByID", fn)
	return m
}

func (m *MockProjectRepo) SetGetWithAccess(fn func(context.Context, string, string) (*models.Project, error)) *MockProjectRepo {
	m.Set("GetWithAccess", fn)
	return m
}

func (m *MockProjectRepo) SetListByUser(fn func(context.Context, string) ([]models.Project, error)) *MockProjectRepo {
	m.Set("ListByUser", fn)
	return m
}

func (m *MockProjectRepo) SetListByCollaborator(fn func(context.Context, string) ([]models.Project, error)) *MockProjectRepo {
	m.Set("ListByCollaborator", fn)
	return m
}

func (m *MockProjectRepo) SetListPages(fn func(context.Context, string, string) ([]models.Page, error)) *MockProjectRepo {
	m.Set("ListPages", fn)
	return m
}

func (m *MockProjectRepo) SetCreate(fn func(context.Context, *models.Project) error) *MockProjectRepo {
	m.Set("Create", fn)
	return m
}

func (m *MockProjectRepo) SetUpdate(fn func(context.Context, string, string, map[string]any) (*models.Project, error)) *MockProjectRepo {
	m.Set("Update", fn)
	return m
}

func (m *MockProjectRepo) SetDelete(fn func(context.Context, string, string) error) *MockProjectRepo {
	m.Set("Delete", fn)
	return m
}

func (m *MockProjectRepo) SetHardDelete(fn func(context.Context, string, string) error) *MockProjectRepo {
	m.Set("HardDelete", fn)
	return m
}

func (m *MockProjectRepo) SetRestore(fn func(context.Context, string, string) error) *MockProjectRepo {
	m.Set("Restore", fn)
	return m
}

func (m *MockProjectRepo) SetExistsForUser(fn func(context.Context, string, string) (bool, error)) *MockProjectRepo {
	m.Set("ExistsForUser", fn)
	return m
}

func (m *MockProjectRepo) SetGetWithLock(fn func(context.Context, string, string) (*models.Project, error)) *MockProjectRepo {
	m.Set("GetWithLock", fn)
	return m
}

func (m *MockProjectRepo) GetPublicByID(ctx context.Context, projectID string) (*models.Project, error) {
	if fn := m.Get("GetPublicByID"); fn != nil {
		return fn.(func(context.Context, string) (*models.Project, error))(ctx, projectID)
	}
	return nil, nil
}

func (m *MockProjectRepo) GetPublicBySubdomain(ctx context.Context, subdomain string) (*models.Project, error) {
	if fn := m.Get("GetPublicBySubdomain"); fn != nil {
		return fn.(func(context.Context, string) (*models.Project, error))(ctx, subdomain)
	}
	return nil, nil
}

func (m *MockProjectRepo) GetByID(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if fn := m.Get("GetByID"); fn != nil {
		return fn.(func(context.Context, string, string) (*models.Project, error))(ctx, projectID, userID)
	}
	return nil, nil
}

func (m *MockProjectRepo) GetWithAccess(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if fn := m.Get("GetWithAccess"); fn != nil {
		return fn.(func(context.Context, string, string) (*models.Project, error))(ctx, projectID, userID)
	}
	return nil, nil
}

func (m *MockProjectRepo) ListByUser(ctx context.Context, userID string) ([]models.Project, error) {
	if fn := m.Get("ListByUser"); fn != nil {
		return fn.(func(context.Context, string) ([]models.Project, error))(ctx, userID)
	}
	return nil, nil
}

func (m *MockProjectRepo) ListByCollaborator(ctx context.Context, userID string) ([]models.Project, error) {
	if fn := m.Get("ListByCollaborator"); fn != nil {
		return fn.(func(context.Context, string) ([]models.Project, error))(ctx, userID)
	}
	return nil, nil
}

func (m *MockProjectRepo) ListPages(ctx context.Context, projectID, userID string) ([]models.Page, error) {
	if fn := m.Get("ListPages"); fn != nil {
		return fn.(func(context.Context, string, string) ([]models.Page, error))(ctx, projectID, userID)
	}
	return nil, nil
}

func (m *MockProjectRepo) Create(ctx context.Context, project *models.Project) error {
	if fn := m.Get("Create"); fn != nil {
		return fn.(func(context.Context, *models.Project) error)(ctx, project)
	}
	return nil
}

func (m *MockProjectRepo) Update(ctx context.Context, projectID, userID string, updates map[string]any) (*models.Project, error) {
	if fn := m.Get("Update"); fn != nil {
		return fn.(func(context.Context, string, string, map[string]any) (*models.Project, error))(ctx, projectID, userID, updates)
	}
	return nil, nil
}

func (m *MockProjectRepo) Delete(ctx context.Context, projectID, userID string) error {
	if fn := m.Get("Delete"); fn != nil {
		return fn.(func(context.Context, string, string) error)(ctx, projectID, userID)
	}
	return nil
}

func (m *MockProjectRepo) HardDelete(ctx context.Context, projectID, userID string) error {
	if fn := m.Get("HardDelete"); fn != nil {
		return fn.(func(context.Context, string, string) error)(ctx, projectID, userID)
	}
	return nil
}

func (m *MockProjectRepo) Restore(ctx context.Context, projectID, userID string) error {
	if fn := m.Get("Restore"); fn != nil {
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

func (m *MockProjectRepo) GetWithLock(ctx context.Context, projectID, userID string) (*models.Project, error) {
	if fn := m.Get("GetWithLock"); fn != nil {
		return fn.(func(context.Context, string, string) (*models.Project, error))(ctx, projectID, userID)
	}
	return nil, nil
}
func (m *MockProjectRepo) CreateWithPage(ctx context.Context, project *models.Project, page *models.Page) error {
	if fn := m.Get("CreateWithPage"); fn != nil {
		return fn.(func(context.Context, *models.Project, *models.Page) error)(ctx, project, page)
	}
	return nil
}

func (m *MockProjectRepo) GetAccess(ctx context.Context, projectID, userID string) (*models.ProjectAccess, error) {
	if fn := m.Get("GetAccess"); fn != nil {
		return fn.(func(context.Context, string, string) (*models.ProjectAccess, error))(ctx, projectID, userID)
	}
	return nil, nil
}

func (m *MockProjectRepo) SetCreateWithPage(fn func(context.Context, *models.Project, *models.Page) error) *MockProjectRepo {
	m.Set("CreateWithPage", fn)
	return m
}

func (m *MockProjectRepo) SetGetAccess(fn func(context.Context, string, string) (*models.ProjectAccess, error)) *MockProjectRepo {
	m.Set("GetAccess", fn)
	return m
}

func (m *MockProjectRepo) UpdateFields(ctx context.Context, projectID, userID string, updates map[string]any) (*models.Project, error) {
	if fn := m.Get("UpdateFields"); fn != nil {
		return fn.(func(context.Context, string, string, map[string]any) (*models.Project, error))(ctx, projectID, userID, updates)
	}
	return m.Update(ctx, projectID, userID, updates)
}

func (m *MockProjectRepo) TransferOwnership(ctx context.Context, projectID, currentOwnerID, newOwnerID string) error {
	if fn := m.Get("TransferOwnership"); fn != nil {
		return fn.(func(context.Context, string, string, string) error)(ctx, projectID, currentOwnerID, newOwnerID)
	}
	return nil
}
