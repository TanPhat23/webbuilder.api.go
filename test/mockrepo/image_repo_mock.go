package test

import (
	"context"
	"my-go-app/internal/models"
)

type MockImageRepo struct {
	*GenericMock
}

func NewMockImageRepo() *MockImageRepo {
	return &MockImageRepo{GenericMock: &GenericMock{funcs: make(map[string]any)}}
}

func (m *MockImageRepo) SetCreate(fn func(context.Context, models.Image) (*models.Image, error)) *MockImageRepo {
	m.Set("Create", fn)
	return m
}

func (m *MockImageRepo) SetListByUser(fn func(context.Context, string) ([]models.Image, error)) *MockImageRepo {
	m.Set("ListByUser", fn)
	return m
}

func (m *MockImageRepo) SetGetByID(fn func(context.Context, string, string) (*models.Image, error)) *MockImageRepo {
	m.Set("GetByID", fn)
	return m
}

func (m *MockImageRepo) SetGetPublicByID(fn func(context.Context, string) (*models.Image, error)) *MockImageRepo {
	m.Set("GetPublicByID", fn)
	return m
}

func (m *MockImageRepo) SetDelete(fn func(context.Context, string, string) error) *MockImageRepo {
	m.Set("Delete", fn)
	return m
}

func (m *MockImageRepo) SetSoftDelete(fn func(context.Context, string, string) error) *MockImageRepo {
	m.Set("SoftDelete", fn)
	return m
}

func (m *MockImageRepo) SetListAll(fn func(context.Context, int, int) ([]models.Image, error)) *MockImageRepo {
	m.Set("ListAll", fn)
	return m
}

func (m *MockImageRepo) Create(ctx context.Context, image models.Image) (*models.Image, error) {
	if fn := m.Get("Create"); fn != nil {
		return fn.(func(context.Context, models.Image) (*models.Image, error))(ctx, image)
	}
	return &image, nil
}

func (m *MockImageRepo) ListByUser(ctx context.Context, userID string) ([]models.Image, error) {
	if fn := m.Get("ListByUser"); fn != nil {
		return fn.(func(context.Context, string) ([]models.Image, error))(ctx, userID)
	}
	return nil, nil
}

func (m *MockImageRepo) GetByID(ctx context.Context, imageID string, userID string) (*models.Image, error) {
	if fn := m.Get("GetByID"); fn != nil {
		return fn.(func(context.Context, string, string) (*models.Image, error))(ctx, imageID, userID)
	}
	return nil, nil
}

func (m *MockImageRepo) GetPublicByID(ctx context.Context, imageID string) (*models.Image, error) {
	if fn := m.Get("GetPublicByID"); fn != nil {
		return fn.(func(context.Context, string) (*models.Image, error))(ctx, imageID)
	}
	return nil, nil
}

func (m *MockImageRepo) Delete(ctx context.Context, imageID string, userID string) error {
	if fn := m.Get("Delete"); fn != nil {
		return fn.(func(context.Context, string, string) error)(ctx, imageID, userID)
	}
	return nil
}

func (m *MockImageRepo) SoftDelete(ctx context.Context, imageID string, userID string) error {
	if fn := m.Get("SoftDelete"); fn != nil {
		return fn.(func(context.Context, string, string) error)(ctx, imageID, userID)
	}
	return nil
}

func (m *MockImageRepo) ListAll(ctx context.Context, limit int, offset int) ([]models.Image, error) {
	if fn := m.Get("ListAll"); fn != nil {
		return fn.(func(context.Context, int, int) ([]models.Image, error))(ctx, limit, offset)
	}
	return nil, nil
}
