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

	_, err := s.GetContentFieldsByContentType(ctx, "  ")
	wantErr(t, err, "contentTypeId is required")
	r.OnGetContentFieldsByContentType = func(_ context.Context, id string) ([]models.ContentField, error) {
		if id != "t" {
			t.Fatal(id)
		}
		return []models.ContentField{{Id: "1"}}, nil
	}
	l, err := s.GetContentFieldsByContentType(ctx, "t")
	wantNoErr(t, err)
	if len(l) != 1 {
		t.Fatal(l)
	}

	_, err = s.GetContentFieldByID(ctx, " ")
	wantErr(t, err, "content field id is required")
	r.OnGetContentFieldByID = func(context.Context, string) (*models.ContentField, error) { return nil, errBoom }
	if _, err := s.GetContentFieldByID(ctx, "1"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	r.OnGetContentFieldByID = func(context.Context, string) (*models.ContentField, error) { return nil, nil }
	_, err = s.GetContentFieldByID(ctx, "1")
	wantErr(t, err, "content field does not exist")
	r.OnGetContentFieldByID = func(context.Context, string) (*models.ContentField, error) { return &models.ContentField{Id: "1"}, nil }
	f, err := s.GetContentFieldByID(ctx, "1")
	wantNoErr(t, err)
	if f.Id != "1" {
		t.Fatal(f)
	}
}

func TestContentFieldCreate(t *testing.T) {
	ctx := context.Background()
	r := &fkContentFieldRepo{}
	s := NewContentFieldService(r)
	_, err := s.CreateContentField(ctx, nil)
	wantErr(t, err, "cannot be nil")
	_, err = s.CreateContentField(ctx, &models.ContentField{ContentTypeId: " ", Name: "n", Type: "t"})
	wantErr(t, err, "contentTypeId is required")
	_, err = s.CreateContentField(ctx, &models.ContentField{ContentTypeId: "c", Name: " ", Type: "t"})
	wantErr(t, err, "name is required")
	_, err = s.CreateContentField(ctx, &models.ContentField{ContentTypeId: "c", Name: "n", Type: " "})
	wantErr(t, err, "type is required")
	wantCalls(t, r.Calls)

	in := &models.ContentField{ContentTypeId: "c", Name: "n", Type: "t"}
	r.OnCreateContentField = func(_ context.Context, f *models.ContentField) (*models.ContentField, error) { return f, nil }
	out, err := s.CreateContentField(ctx, in)
	wantNoErr(t, err)
	if out != in {
		t.Fatal("wrong")
	}
	r.OnCreateContentField = func(context.Context, *models.ContentField) (*models.ContentField, error) { return nil, errBoom }
	if _, err := s.CreateContentField(ctx, in); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
}

func TestContentFieldUpdate(t *testing.T) {
	ctx := context.Background()
	r := &fkContentFieldRepo{}
	s := NewContentFieldService(r)
	ok := map[string]any{"Name": "x"}
	_, err := s.UpdateContentField(ctx, " ", ok)
	wantErr(t, err, "content field id is required")
	_, err = s.UpdateContentField(ctx, "1", nil)
	wantErr(t, err, "no updates provided")

	r.OnGetContentFieldByID = func(context.Context, string) (*models.ContentField, error) { return nil, errBoom }
	if _, err := s.UpdateContentField(ctx, "1", ok); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	r.OnGetContentFieldByID = func(context.Context, string) (*models.ContentField, error) { return nil, nil }
	_, err = s.UpdateContentField(ctx, "1", ok)
	wantErr(t, err, "does not exist")

	r.OnGetContentFieldByID = func(context.Context, string) (*models.ContentField, error) { return &models.ContentField{}, nil }
	_, err = s.UpdateContentField(ctx, "1", map[string]any{"Name": " "})
	wantErr(t, err, "name cannot be empty")
	_, err = s.UpdateContentField(ctx, "1", map[string]any{"Type": ""})
	wantErr(t, err, "type cannot be empty")
	_, err = s.UpdateContentField(ctx, "1", map[string]any{"Required": "yes"})
	wantErr(t, err, "required must be a boolean")
	wantCalls(t, r.Calls[:0])

	var got map[string]any
	r.OnUpdateContentField = func(_ context.Context, id string, u map[string]any) (*models.ContentField, error) {
		got = u
		return &models.ContentField{Id: id}, nil
	}
	out, err := s.UpdateContentField(ctx, "1", map[string]any{"Name": "n", "Type": "t", "Required": true})
	wantNoErr(t, err)
	if out.Id != "1" || got["Required"] != true {
		t.Fatal(out, got)
	}
	r.OnUpdateContentField = func(context.Context, string, map[string]any) (*models.ContentField, error) { return nil, errBoom }
	if _, err := s.UpdateContentField(ctx, "1", ok); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
}

func TestContentFieldDelete(t *testing.T) {
	ctx := context.Background()
	r := &fkContentFieldRepo{}
	s := NewContentFieldService(r)
	wantErr(t, s.DeleteContentField(ctx, ""), "content field id is required")
	r.OnGetContentFieldByID = func(context.Context, string) (*models.ContentField, error) { return nil, errBoom }
	if err := s.DeleteContentField(ctx, "1"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	r.OnGetContentFieldByID = func(context.Context, string) (*models.ContentField, error) { return nil, nil }
	wantErr(t, s.DeleteContentField(ctx, "1"), "does not exist")
	r.OnGetContentFieldByID = func(context.Context, string) (*models.ContentField, error) { return &models.ContentField{}, nil }
	r.OnDeleteContentField = func(context.Context, string) error { return errBoom }
	if err := s.DeleteContentField(ctx, "1"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	r.OnDeleteContentField = func(context.Context, string) error { return nil }
	wantNoErr(t, s.DeleteContentField(ctx, "1"))
}
