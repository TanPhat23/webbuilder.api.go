package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
)

func TestContentTypeServicePassThrough(t *testing.T) {
	ctx := context.Background()
	r := &fkContentTypeRepo{}
	s := NewContentTypeService(r)
	one := &models.ContentType{Id: "1"}

	r.OnList = func(context.Context) ([]models.ContentType, error) { return []models.ContentType{*one}, nil }
	l, err := s.List(ctx)
	wantNoErr(t, err)
	if len(l) != 1 || l[0].Id != "1" {
		t.Fatal(l)
	}
	r.OnList = func(context.Context) ([]models.ContentType, error) { return nil, errBoom }
	if _, err := s.List(ctx); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnGetByID = func(_ context.Context, id string) (*models.ContentType, error) {
		if id != "1" {
			t.Fatal(id)
		}
		return one, nil
	}
	g, err := s.GetByID(ctx, "1")
	wantNoErr(t, err)
	if g != one {
		t.Fatal("wrong")
	}
	r.OnGetByID = func(context.Context, string) (*models.ContentType, error) { return nil, errBoom }
	if _, err := s.GetByID(ctx, "1"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnCreate = func(context.Context, *models.ContentType) (*models.ContentType, error) { return one, nil }
	c, err := s.Create(ctx, one)
	wantNoErr(t, err)
	if c != one {
		t.Fatal("wrong")
	}
	r.OnCreate = func(context.Context, *models.ContentType) (*models.ContentType, error) { return nil, errBoom }
	if _, err := s.Create(ctx, one); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnDelete = func(_ context.Context, id string) error { return errBoom }
	if err := s.Delete(ctx, "1"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	r.OnDelete = func(context.Context, string) error { return nil }
	wantNoErr(t, s.Delete(ctx, "1"))
}

func TestContentTypeUpdateBuildsNameDescriptionMap(t *testing.T) {
	ctx := context.Background()
	r := &fkContentTypeRepo{}
	s := NewContentTypeService(r)
	var got map[string]any
	r.OnUpdate = func(_ context.Context, id string, u map[string]any) (*models.ContentType, error) {
		if id != "7" {
			t.Fatal(id)
		}
		got = u
		return &models.ContentType{Id: id}, nil
	}
	desc := strp("d")
	res, err := s.Update(ctx, "7", &models.ContentType{Id: "ignored", Name: "n", Description: desc})
	wantNoErr(t, err)
	if res.Id != "7" || len(got) != 2 || got["name"] != "n" || got["description"] != desc {
		t.Fatalf("updates %v", got)
	}
	r.OnUpdate = func(context.Context, string, map[string]any) (*models.ContentType, error) { return nil, errBoom }
	if _, err := s.Update(ctx, "7", &models.ContentType{}); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
}
