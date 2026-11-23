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

	r.OnGetCustomElementTypes = func(context.Context) ([]models.CustomElementType, error) {
		return []models.CustomElementType{*one}, nil
	}
	l, err := s.GetCustomElementTypes(ctx)
	wantNoErr(t, err)
	if len(l) != 1 {
		t.Fatal(l)
	}
	r.OnGetCustomElementTypes = func(context.Context) ([]models.CustomElementType, error) { return nil, errBoom }
	if _, err := s.GetCustomElementTypes(ctx); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnGetCustomElementTypeByID = func(_ context.Context, id string) (*models.CustomElementType, error) {
		if id != "i" {
			t.Fatal(id)
		}
		return one, nil
	}
	g, err := s.GetCustomElementTypeByID(ctx, "i")
	wantNoErr(t, err)
	if g != one {
		t.Fatal("wrong")
	}
	r.OnGetCustomElementTypeByID = func(context.Context, string) (*models.CustomElementType, error) { return nil, errBoom }
	if _, err := s.GetCustomElementTypeByID(ctx, "i"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnGetCustomElementTypeByName = func(_ context.Context, n string) (*models.CustomElementType, error) {
		if n != "nm" {
			t.Fatal(n)
		}
		return one, nil
	}
	g, err = s.GetCustomElementTypeByName(ctx, "nm")
	wantNoErr(t, err)
	if g != one {
		t.Fatal("wrong")
	}
	r.OnGetCustomElementTypeByName = func(context.Context, string) (*models.CustomElementType, error) { return nil, errBoom }
	if _, err := s.GetCustomElementTypeByName(ctx, "nm"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnCreateCustomElementType = func(context.Context, *models.CustomElementType) (*models.CustomElementType, error) { return one, nil }
	g, err = s.CreateCustomElementType(ctx, one)
	wantNoErr(t, err)
	if g != one {
		t.Fatal("wrong")
	}
	r.OnCreateCustomElementType = func(context.Context, *models.CustomElementType) (*models.CustomElementType, error) {
		return nil, errBoom
	}
	if _, err := s.CreateCustomElementType(ctx, one); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnUpdateCustomElementType = func(_ context.Context, id string, u map[string]any) (*models.CustomElementType, error) {
		if id != "i" || u["k"] != "v" {
			t.Fatal(id, u)
		}
		return one, nil
	}
	g, err = s.UpdateCustomElementType(ctx, "i", map[string]any{"k": "v"})
	wantNoErr(t, err)
	if g != one {
		t.Fatal("wrong")
	}
	r.OnUpdateCustomElementType = func(context.Context, string, map[string]any) (*models.CustomElementType, error) { return nil, errBoom }
	if _, err := s.UpdateCustomElementType(ctx, "i", nil); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnDeleteCustomElementType = func(context.Context, string) error { return errBoom }
	if err := s.DeleteCustomElementType(ctx, "i"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	r.OnDeleteCustomElementType = func(context.Context, string) error { return nil }
	wantNoErr(t, s.DeleteCustomElementType(ctx, "i"))
}
