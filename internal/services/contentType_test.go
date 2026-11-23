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

	r.OnGetContentTypes = func(context.Context) ([]models.ContentType, error) { return []models.ContentType{*one}, nil }
	l, err := s.GetContentTypes(ctx)
	wantNoErr(t, err)
	if len(l) != 1 || l[0].Id != "1" {
		t.Fatal(l)
	}
	r.OnGetContentTypes = func(context.Context) ([]models.ContentType, error) { return nil, errBoom }
	if _, err := s.GetContentTypes(ctx); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnGetContentTypeByID = func(_ context.Context, id string) (*models.ContentType, error) {
		if id != "1" {
			t.Fatal(id)
		}
		return one, nil
	}
	g, err := s.GetContentTypeByID(ctx, "1")
	wantNoErr(t, err)
	if g != one {
		t.Fatal("wrong")
	}
	r.OnGetContentTypeByID = func(context.Context, string) (*models.ContentType, error) { return nil, errBoom }
	if _, err := s.GetContentTypeByID(ctx, "1"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnCreateContentType = func(context.Context, *models.ContentType) (*models.ContentType, error) { return one, nil }
	c, err := s.CreateContentType(ctx, one)
	wantNoErr(t, err)
	if c != one {
		t.Fatal("wrong")
	}
	r.OnCreateContentType = func(context.Context, *models.ContentType) (*models.ContentType, error) { return nil, errBoom }
	if _, err := s.CreateContentType(ctx, one); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnDeleteContentType = func(_ context.Context, id string) error { return errBoom }
	if err := s.DeleteContentType(ctx, "1"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	r.OnDeleteContentType = func(context.Context, string) error { return nil }
	wantNoErr(t, s.DeleteContentType(ctx, "1"))
}

func TestContentTypeUpdateBuildsNameDescriptionMap(t *testing.T) {
	ctx := context.Background()
	r := &fkContentTypeRepo{}
	s := NewContentTypeService(r)
	var got map[string]any
	r.OnUpdateContentType = func(_ context.Context, id string, u map[string]any) (*models.ContentType, error) {
		if id != "7" {
			t.Fatal(id)
		}
		got = u
		return &models.ContentType{Id: id}, nil
	}
	desc := strp("d")
	res, err := s.UpdateContentType(ctx, "7", &models.ContentType{Id: "ignored", Name: "n", Description: desc})
	wantNoErr(t, err)
	if res.Id != "7" || len(got) != 2 || got["name"] != "n" || got["description"] != desc {
		t.Fatalf("updates %v", got)
	}
	r.OnUpdateContentType = func(context.Context, string, map[string]any) (*models.ContentType, error) { return nil, errBoom }
	if _, err := s.UpdateContentType(ctx, "7", &models.ContentType{}); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
}
