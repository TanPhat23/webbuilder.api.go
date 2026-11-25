package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

func TestContentItemReads(t *testing.T) {
	ctx := context.Background()
	r := &fkContentItemRepo{}
	s := NewContentItemService(r)

	_, err := s.GetContentItemsByContentType(ctx, "")
	wantErr(t, err, "contentTypeId is required")
	r.OnGetContentItemsByContentType = func(context.Context, string) ([]models.ContentItem, error) {
		return []models.ContentItem{{Id: "1"}}, nil
	}
	l, err := s.GetContentItemsByContentType(ctx, "t")
	wantNoErr(t, err)
	if len(l) != 1 {
		t.Fatal(l)
	}

	_, err = s.GetContentItemByID(ctx, "")
	wantErr(t, err, "content item id is required")
	r.OnGetContentItemByID = func(context.Context, string) (*models.ContentItem, error) { return nil, errBoom }
	if _, err := s.GetContentItemByID(ctx, "1"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	r.OnGetContentItemByID = func(context.Context, string) (*models.ContentItem, error) { return &models.ContentItem{Id: "1"}, nil }
	it, err := s.GetContentItemByID(ctx, "1")
	wantNoErr(t, err)
	if it.Id != "1" {
		t.Fatal(it)
	}

	_, err = s.GetContentItemBySlug(ctx, "", "s")
	wantErr(t, err, "contentTypeId is required")
	_, err = s.GetContentItemBySlug(ctx, "t", "")
	wantErr(t, err, "slug is required")
	r.OnGetContentItemBySlug = func(_ context.Context, ct, sl string) (*models.ContentItem, error) {
		if ct != "t" || sl != "s" {
			t.Fatal(ct, sl)
		}
		return &models.ContentItem{Id: "2"}, nil
	}
	it, err = s.GetContentItemBySlug(ctx, "t", "s")
	wantNoErr(t, err)
	if it.Id != "2" {
		t.Fatal(it)
	}
}

func TestContentItemGetPublic(t *testing.T) {
	ctx := context.Background()
	r := &fkContentItemRepo{}
	s := NewContentItemService(r)
	for _, tc := range []struct {
		name, ct string
		limit    int
		by, ord  string
		msg      string
	}{
		{"no type", "", 1, "a", "asc", "contentTypeId is required"},
		{"zero limit", "t", 0, "a", "asc", "limit must be greater than zero"},
		{"negative limit", "t", -1, "a", "asc", "limit must be greater than zero"},
		{"no sortBy", "t", 1, "", "asc", "sortBy is required"},
		{"no sortOrder", "t", 1, "a", "", "sortOrder is required"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.GetPublicContentItems(ctx, tc.ct, tc.limit, tc.by, tc.ord)
			wantErr(t, err, tc.msg)
		})
	}
	r.OnGetPublicContentItems = func(_ context.Context, ct string, lim int, by, ord string) ([]models.ContentItem, error) {
		if ct != "t" || lim != 5 || by != "title" || ord != "desc" {
			t.Fatal(ct, lim, by, ord)
		}
		return []models.ContentItem{{}}, nil
	}
	l, err := s.GetPublicContentItems(ctx, "t", 5, "title", "desc")
	wantNoErr(t, err)
	if len(l) != 1 {
		t.Fatal(l)
	}
}

func TestContentItemCreate(t *testing.T) {
	ctx := context.Background()
	r := &fkContentItemRepo{}
	s := NewContentItemService(r)
	_, err := s.CreateContentItem(ctx, nil, nil)
	wantErr(t, err, "cannot be nil")
	_, err = s.CreateContentItem(ctx, &models.ContentItem{Slug: "s", Title: "t"}, nil)
	wantErr(t, err, "contentTypeId is required")
	_, err = s.CreateContentItem(ctx, &models.ContentItem{ContentTypeId: "c", Title: "t"}, nil)
	wantErr(t, err, "slug is required")
	_, err = s.CreateContentItem(ctx, &models.ContentItem{ContentTypeId: "c", Slug: "s"}, nil)
	wantErr(t, err, "title is required")
	wantCalls(t, r.Calls)

	in := &models.ContentItem{ContentTypeId: "c", Slug: "s", Title: "t", FieldValues: []models.ContentFieldValue{{}}}
	var gotFV []models.ContentFieldValue
	var sawFV int
	r.OnCreateContentItem = func(_ context.Context, ci *models.ContentItem, fv []models.ContentFieldValue) (*models.ContentItem, error) {
		gotFV = fv
		sawFV = len(ci.FieldValues)
		return ci, nil
	}
	out, err := s.CreateContentItem(ctx, in, nil)
	wantNoErr(t, err)
	if out != in || gotFV == nil || len(gotFV) != 0 || sawFV != 0 {
		t.Fatalf("fv=%v saw=%d", gotFV, sawFV)
	}

	r.OnCreateContentItem = func(context.Context, *models.ContentItem, []models.ContentFieldValue) (*models.ContentItem, error) {
		return nil, errBoom
	}
	if _, err := s.CreateContentItem(ctx, in, nil); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
}

func TestContentItemUpdate(t *testing.T) {
	ctx := context.Background()
	r := &fkContentItemRepo{}
	s := NewContentItemService(r)
	fv := []models.ContentFieldValue{{}}
	_, err := s.UpdateContentItem(ctx, "", map[string]any{"a": 1}, nil)
	wantErr(t, err, "content item id is required")
	_, err = s.UpdateContentItem(ctx, "1", nil, nil)
	wantErr(t, err, "no updates provided")
	_, err = s.UpdateContentItem(ctx, "1", map[string]any{"Slug": ""}, nil)
	wantErr(t, err, "slug cannot be empty")
	_, err = s.UpdateContentItem(ctx, "1", map[string]any{"Title": ""}, nil)
	wantErr(t, err, "title cannot be empty")
	wantCalls(t, r.Calls)

	var gotU map[string]any
	var gotFV []models.ContentFieldValue
	r.OnUpdateContentItem = func(_ context.Context, id string, u map[string]any, f []models.ContentFieldValue) (*models.ContentItem, error) {
		gotU, gotFV = u, f
		return &models.ContentItem{Id: id}, nil
	}
	out, err := s.UpdateContentItem(ctx, "1", map[string]any{"Title": "x"}, nil)
	wantNoErr(t, err)
	if out.Id != "1" || gotU["Title"] != "x" || gotFV == nil || len(gotFV) != 0 {
		t.Fatal(gotU, gotFV)
	}
	_, err = s.UpdateContentItem(ctx, "1", nil, fv)
	wantNoErr(t, err)
	if gotU == nil || len(gotFV) != 1 {
		t.Fatal(gotU, gotFV)
	}
	r.OnUpdateContentItem = func(context.Context, string, map[string]any, []models.ContentFieldValue) (*models.ContentItem, error) {
		return nil, errBoom
	}
	if _, err := s.UpdateContentItem(ctx, "1", map[string]any{"a": 1}, nil); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
}

func TestContentItemDeleteAndSort(t *testing.T) {
	ctx := context.Background()
	r := &fkContentItemRepo{}
	s := NewContentItemService(r)
	wantErr(t, s.DeleteContentItem(ctx, ""), "content item id is required")
	r.OnDeleteContentItem = func(context.Context, string) error { return errBoom }
	if err := s.DeleteContentItem(ctx, "1"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	r.OnDeleteContentItem = func(context.Context, string) error { return nil }
	wantNoErr(t, s.DeleteContentItem(ctx, "1"))

	_, _, err := s.ValidatePublicSort("", "asc")
	wantErr(t, err, "sortBy is required")
	_, _, err = s.ValidatePublicSort("a", "")
	wantErr(t, err, "sortOrder is required")
	_, _, err = s.ValidatePublicSort("a", "up")
	wantErr(t, err, "invalid sortOrder: up")
	for _, o := range []string{"asc", "desc"} {
		by, ord, err := s.ValidatePublicSort("title", o)
		wantNoErr(t, err)
		if by != "title" || ord != o {
			t.Fatal(by, ord)
		}
	}
}

func TestContentItemGetBySlugPublicIsolation(t *testing.T) {
	ctx := context.Background()
	projID := "p1"

	t.Run("rejects non-public content type", func(t *testing.T) {
		ciRepo := &fkContentItemRepo{}
		ctRepo := &fkContentTypeRepo{
			OnGetContentTypeByID: func(context.Context, string) (*models.ContentType, error) {
				return &models.ContentType{Id: "ct1", ProjectId: &projID, PublicRead: false}, nil
			},
		}
		projRepo := &fkProjectRepo{}
		s := NewContentItemService(ciRepo).WithAuth(ctRepo, projRepo)
		_, err := s.GetContentItemBySlug(ctx, "ct1", "hello")
		if !errors.Is(err, repositories.ErrContentItemNotFound) {
			t.Fatalf("expected ErrContentItemNotFound, got %v", err)
		}
	})

	t.Run("rejects unpublished or deleted project", func(t *testing.T) {
		ciRepo := &fkContentItemRepo{}
		ctRepo := &fkContentTypeRepo{
			OnGetContentTypeByID: func(context.Context, string) (*models.ContentType, error) {
				return &models.ContentType{Id: "ct1", ProjectId: &projID, PublicRead: true}, nil
			},
		}
		projRepo := &fkProjectRepo{
			OnGetPublicProjectByID: func(context.Context, string) (*models.Project, error) {
				return nil, repositories.ErrProjectNotFound
			},
		}
		s := NewContentItemService(ciRepo).WithAuth(ctRepo, projRepo)
		_, err := s.GetContentItemBySlug(ctx, "ct1", "hello")
		if !errors.Is(err, repositories.ErrContentItemNotFound) {
			t.Fatalf("expected ErrContentItemNotFound, got %v", err)
		}
	})

	t.Run("allows public content type on published project", func(t *testing.T) {
		ciRepo := &fkContentItemRepo{
			OnGetContentItemBySlug: func(_ context.Context, ct, sl string) (*models.ContentItem, error) {
				return &models.ContentItem{Id: "i1", ContentTypeId: ct, Slug: sl, Published: true}, nil
			},
		}
		ctRepo := &fkContentTypeRepo{
			OnGetContentTypeByID: func(context.Context, string) (*models.ContentType, error) {
				return &models.ContentType{Id: "ct1", ProjectId: &projID, PublicRead: true}, nil
			},
		}
		projRepo := &fkProjectRepo{
			OnGetPublicProjectByID: func(context.Context, string) (*models.Project, error) {
				return &models.Project{ID: projID, Published: true}, nil
			},
		}
		s := NewContentItemService(ciRepo).WithAuth(ctRepo, projRepo)
		got, err := s.GetContentItemBySlug(ctx, "ct1", "hello")
		wantNoErr(t, err)
		if got.Id != "i1" {
			t.Fatalf("got=%+v", got)
		}
	})
}
