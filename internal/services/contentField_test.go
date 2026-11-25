package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
)

func TestContentFieldReads(t *testing.T) {
	ctx := context.Background()
	r := &fkContentFieldRepo{}
	s := NewContentFieldService(r)

	_, err := s.ListByContentType(ctx, "  ")
	wantErr(t, err, "contentTypeId is required")
	r.OnListByContentType = func(_ context.Context, id string) ([]models.ContentField, error) {
		if id != "t" {
			t.Fatal(id)
		}
		return []models.ContentField{{Id: "1"}}, nil
	}
	l, err := s.ListByContentType(ctx, "t")
	wantNoErr(t, err)
	if len(l) != 1 {
		t.Fatal(l)
	}

	_, err = s.GetByID(ctx, " ")
	wantErr(t, err, "content field id is required")
	r.OnGetByID = func(context.Context, string) (*models.ContentField, error) { return nil, errBoom }
	if _, err := s.GetByID(ctx, "1"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	r.OnGetByID = func(context.Context, string) (*models.ContentField, error) { return nil, nil }
	_, err = s.GetByID(ctx, "1")
	wantErr(t, err, "content field does not exist")
	r.OnGetByID = func(context.Context, string) (*models.ContentField, error) { return &models.ContentField{Id: "1"}, nil }
	f, err := s.GetByID(ctx, "1")
	wantNoErr(t, err)
	if f.Id != "1" {
		t.Fatal(f)
	}
}

func TestContentFieldCreate(t *testing.T) {
	ctx := context.Background()
	r := &fkContentFieldRepo{}
	s := NewContentFieldService(r)
	_, err := s.Create(ctx, nil)
	wantErr(t, err, "cannot be nil")
	_, err = s.Create(ctx, &models.ContentField{ContentTypeId: " ", Name: "n", Type: "t"})
	wantErr(t, err, "contentTypeId is required")
	_, err = s.Create(ctx, &models.ContentField{ContentTypeId: "c", Name: " ", Type: "t"})
	wantErr(t, err, "name is required")
	_, err = s.Create(ctx, &models.ContentField{ContentTypeId: "c", Name: "n", Type: " "})
	wantErr(t, err, "type is required")
	wantCalls(t, r.Calls)

	in := &models.ContentField{ContentTypeId: "c", Name: "n", Type: "t"}
	r.OnCreate = func(_ context.Context, f *models.ContentField) (*models.ContentField, error) { return f, nil }
	out, err := s.Create(ctx, in)
	wantNoErr(t, err)
	if out != in {
		t.Fatal("wrong")
	}
	r.OnCreate = func(context.Context, *models.ContentField) (*models.ContentField, error) { return nil, errBoom }
	if _, err := s.Create(ctx, in); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
}

func TestContentFieldUpdate(t *testing.T) {
	ctx := context.Background()
	r := &fkContentFieldRepo{}
	s := NewContentFieldService(r)
	ok := map[string]any{"Name": "x"}
	_, err := s.Update(ctx, " ", ok)
	wantErr(t, err, "content field id is required")
	_, err = s.Update(ctx, "1", nil)
	wantErr(t, err, "no updates provided")

	r.OnGetByID = func(context.Context, string) (*models.ContentField, error) { return nil, errBoom }
	if _, err := s.Update(ctx, "1", ok); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	r.OnGetByID = func(context.Context, string) (*models.ContentField, error) { return nil, nil }
	_, err = s.Update(ctx, "1", ok)
	wantErr(t, err, "does not exist")

	r.OnGetByID = func(context.Context, string) (*models.ContentField, error) { return &models.ContentField{}, nil }
	_, err = s.Update(ctx, "1", map[string]any{"Name": " "})
	wantErr(t, err, "name cannot be empty")
	_, err = s.Update(ctx, "1", map[string]any{"Type": ""})
	wantErr(t, err, "type cannot be empty")
	_, err = s.Update(ctx, "1", map[string]any{"Required": "yes"})
	wantErr(t, err, "required must be a boolean")
	wantCalls(t, r.Calls[:0])

	var got map[string]any
	r.OnUpdate = func(_ context.Context, id string, u map[string]any) (*models.ContentField, error) {
		got = u
		return &models.ContentField{Id: id}, nil
	}
	out, err := s.Update(ctx, "1", map[string]any{"Name": "n", "Type": "t", "Required": true})
	wantNoErr(t, err)
	if out.Id != "1" || got["Required"] != true {
		t.Fatal(out, got)
	}
	r.OnUpdate = func(context.Context, string, map[string]any) (*models.ContentField, error) { return nil, errBoom }
	if _, err := s.Update(ctx, "1", ok); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
}

func TestContentFieldDelete(t *testing.T) {
	ctx := context.Background()
	r := &fkContentFieldRepo{}
	s := NewContentFieldService(r)
	wantErr(t, s.Delete(ctx, ""), "content field id is required")
	r.OnGetByID = func(context.Context, string) (*models.ContentField, error) { return nil, errBoom }
	if err := s.Delete(ctx, "1"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	r.OnGetByID = func(context.Context, string) (*models.ContentField, error) { return nil, nil }
	wantErr(t, s.Delete(ctx, "1"), "does not exist")
	r.OnGetByID = func(context.Context, string) (*models.ContentField, error) { return &models.ContentField{}, nil }
	r.OnDelete = func(context.Context, string) error { return errBoom }
	if err := s.Delete(ctx, "1"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	r.OnDelete = func(context.Context, string) error { return nil }
	wantNoErr(t, s.Delete(ctx, "1"))
}
