package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
)

func TestCustomElementTypeService(t *testing.T) {
	ctx := context.Background()
	r := &fkCustomElementTypeRepo{}
	s := NewCustomElementTypeService(r)
	one := &models.CustomElementType{}

	r.OnList = func(context.Context) ([]models.CustomElementType, error) {
		return []models.CustomElementType{*one}, nil
	}
	l, err := s.List(ctx)
	wantNoErr(t, err)
	if len(l) != 1 {
		t.Fatal(l)
	}
	r.OnList = func(context.Context) ([]models.CustomElementType, error) { return nil, errBoom }
	if _, err := s.List(ctx); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnGetByID = func(_ context.Context, id string) (*models.CustomElementType, error) {
		if id != "i" {
			t.Fatal(id)
		}
		return one, nil
	}
	g, err := s.GetByID(ctx, "i")
	wantNoErr(t, err)
	if g != one {
		t.Fatal("wrong")
	}
	r.OnGetByID = func(context.Context, string) (*models.CustomElementType, error) { return nil, errBoom }
	if _, err := s.GetByID(ctx, "i"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnGetByName = func(_ context.Context, n string) (*models.CustomElementType, error) {
		if n != "nm" {
			t.Fatal(n)
		}
		return one, nil
	}
	g, err = s.GetByName(ctx, "nm")
	wantNoErr(t, err)
	if g != one {
		t.Fatal("wrong")
	}
	r.OnGetByName = func(context.Context, string) (*models.CustomElementType, error) { return nil, errBoom }
	if _, err := s.GetByName(ctx, "nm"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnCreate = func(context.Context, *models.CustomElementType) (*models.CustomElementType, error) { return one, nil }
	g, err = s.Create(ctx, one)
	wantNoErr(t, err)
	if g != one {
		t.Fatal("wrong")
	}
	r.OnCreate = func(context.Context, *models.CustomElementType) (*models.CustomElementType, error) {
		return nil, errBoom
	}
	if _, err := s.Create(ctx, one); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnUpdate = func(_ context.Context, id string, u map[string]any) (*models.CustomElementType, error) {
		if id != "i" || u["k"] != "v" {
			t.Fatal(id, u)
		}
		return one, nil
	}
	g, err = s.Update(ctx, "i", map[string]any{"k": "v"})
	wantNoErr(t, err)
	if g != one {
		t.Fatal("wrong")
	}
	r.OnUpdate = func(context.Context, string, map[string]any) (*models.CustomElementType, error) { return nil, errBoom }
	if _, err := s.Update(ctx, "i", nil); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnDelete = func(context.Context, string) error { return errBoom }
	if err := s.Delete(ctx, "i"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	r.OnDelete = func(context.Context, string) error { return nil }
	wantNoErr(t, s.Delete(ctx, "i"))
}
