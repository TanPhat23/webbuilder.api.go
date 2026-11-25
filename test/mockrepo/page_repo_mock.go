package test

import (
	"context"
	"my-go-app/internal/models"
)

// MockPageRepo implements PageRepositoryInterface for testing.
type MockPageRepo struct {
	*GenericMock
}

// NewMockPageRepo creates a new MockPageRepo instance.
func NewMockPageRepo() *MockPageRepo {
	return &MockPageRepo{GenericMock: &GenericMock{funcs: make(map[string]any)}}
}

func (m *MockPageRepo) SetListByProject(fn func(context.Context, string) ([]models.Page, error)) *MockPageRepo {
	m.Set("ListByProject", fn)
	return m
}

func (m *MockPageRepo) SetGetByID(fn func(context.Context, string, string) (*models.Page, error)) *MockPageRepo {
	m.Set("GetByID", fn)
	return m
}

func (m *MockPageRepo) SetCreate(fn func(context.Context, *models.Page) error) *MockPageRepo {
	m.Set("Create", fn)
	return m
}

func (m *MockPageRepo) SetUpdate(fn func(context.Context, *models.Page) error) *MockPageRepo {
	m.Set("Update", fn)
	return m
}

func (m *MockPageRepo) SetUpdateFields(fn func(context.Context, string, map[string]any) error) *MockPageRepo {
	m.Set("UpdateFields", fn)
	return m
}

func (m *MockPageRepo) SetDelete(fn func(context.Context, string) error) *MockPageRepo {
	m.Set("Delete", fn)
	return m
}

func (m *MockPageRepo) SetDeleteForUser(fn func(context.Context, string, string, string) error) *MockPageRepo {
	m.Set("DeleteForUser", fn)
	return m
}

func (m *MockPageRepo) ListByProject(ctx context.Context, projectID string) ([]models.Page, error) {
	if fn := m.Get("ListByProject"); fn != nil {
		return fn.(func(context.Context, string) ([]models.Page, error))(ctx, projectID)
	}
	return nil, nil
}

func (m *MockPageRepo) GetByID(ctx context.Context, pageID, projectID string) (*models.Page, error) {
	if fn := m.Get("GetByID"); fn != nil {
		return fn.(func(context.Context, string, string) (*models.Page, error))(ctx, pageID, projectID)
	}
	return nil, nil
}

func (m *MockPageRepo) Create(ctx context.Context, page *models.Page) error {
	if fn := m.Get("Create"); fn != nil {
		return fn.(func(context.Context, *models.Page) error)(ctx, page)
	}
	return nil
}

func (m *MockPageRepo) Update(ctx context.Context, page *models.Page) error {
	if fn := m.Get("Update"); fn != nil {
		return fn.(func(context.Context, *models.Page) error)(ctx, page)
	}
	return nil
}

func (m *MockPageRepo) UpdateFields(ctx context.Context, pageID string, updates map[string]any) error {
	if fn := m.Get("UpdateFields"); fn != nil {
		return fn.(func(context.Context, string, map[string]any) error)(ctx, pageID, updates)
	}
	return nil
}

func (m *MockPageRepo) Delete(ctx context.Context, pageID string) error {
	if fn := m.Get("Delete"); fn != nil {
		return fn.(func(context.Context, string) error)(ctx, pageID)
	}
	return nil
}

func (m *MockPageRepo) DeleteForUser(ctx context.Context, pageID, projectID, userID string) error {
	if fn := m.Get("DeleteForUser"); fn != nil {
		return fn.(func(context.Context, string, string, string) error)(ctx, pageID, projectID, userID)
	}
	return nil
}
