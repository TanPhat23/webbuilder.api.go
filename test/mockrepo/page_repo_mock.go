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

func (m *MockPageRepo) SetGetPagesByProjectID(fn func(context.Context, string) ([]models.Page, error)) *MockPageRepo {
	m.Set("GetPagesByProjectID", fn)
	return m
}

func (m *MockPageRepo) SetGetPageByID(fn func(context.Context, string, string) (*models.Page, error)) *MockPageRepo {
	m.Set("GetPageByID", fn)
	return m
}

func (m *MockPageRepo) SetCreatePage(fn func(context.Context, *models.Page) error) *MockPageRepo {
	m.Set("CreatePage", fn)
	return m
}

func (m *MockPageRepo) SetUpdatePage(fn func(context.Context, *models.Page) error) *MockPageRepo {
	m.Set("UpdatePage", fn)
	return m
}

func (m *MockPageRepo) SetUpdatePageFields(fn func(context.Context, string, map[string]any) error) *MockPageRepo {
	m.Set("UpdatePageFields", fn)
	return m
}

func (m *MockPageRepo) SetDeletePage(fn func(context.Context, string) error) *MockPageRepo {
	m.Set("DeletePage", fn)
	return m
}

func (m *MockPageRepo) SetDeletePageByProjectID(fn func(context.Context, string, string, string) error) *MockPageRepo {
	m.Set("DeletePageByProjectID", fn)
	return m
}

func (m *MockPageRepo) GetPagesByProjectID(ctx context.Context, projectID string) ([]models.Page, error) {
	if fn := m.Get("GetPagesByProjectID"); fn != nil {
		return fn.(func(context.Context, string) ([]models.Page, error))(ctx, projectID)
	}
	return nil, nil
}

func (m *MockPageRepo) GetPageByID(ctx context.Context, pageID, projectID string) (*models.Page, error) {
	if fn := m.Get("GetPageByID"); fn != nil {
		return fn.(func(context.Context, string, string) (*models.Page, error))(ctx, pageID, projectID)
	}
	return nil, nil
}

func (m *MockPageRepo) CreatePage(ctx context.Context, page *models.Page) error {
	if fn := m.Get("CreatePage"); fn != nil {
		return fn.(func(context.Context, *models.Page) error)(ctx, page)
	}
	return nil
}

func (m *MockPageRepo) UpdatePage(ctx context.Context, page *models.Page) error {
	if fn := m.Get("UpdatePage"); fn != nil {
		return fn.(func(context.Context, *models.Page) error)(ctx, page)
	}
	return nil
}

func (m *MockPageRepo) UpdatePageFields(ctx context.Context, pageID string, updates map[string]any) error {
	if fn := m.Get("UpdatePageFields"); fn != nil {
		return fn.(func(context.Context, string, map[string]any) error)(ctx, pageID, updates)
	}
	return nil
}

func (m *MockPageRepo) DeletePage(ctx context.Context, pageID string) error {
	if fn := m.Get("DeletePage"); fn != nil {
		return fn.(func(context.Context, string) error)(ctx, pageID)
	}
	return nil
}

func (m *MockPageRepo) DeletePageByProjectID(ctx context.Context, pageID, projectID, userID string) error {
	if fn := m.Get("DeletePageByProjectID"); fn != nil {
		return fn.(func(context.Context, string, string, string) error)(ctx, pageID, projectID, userID)
	}
	return nil
}