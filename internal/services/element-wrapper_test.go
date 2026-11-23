package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
)

func TestElementWrapperService(t *testing.T) {
	ctx := context.Background()
	r := &fkElementRepo{}
	s := NewElementWrapperService(r)

	r.OnGetElements = func(_ context.Context, pid string, pages ...string) ([]models.EditorElement, error) {
		if pid != "p" || len(pages) != 2 || pages[0] != "a" || pages[1] != "b" {
			t.Fatal(pid, pages)
		}
		return []models.EditorElement{nil}, nil
	}
	e, err := s.GetElements(ctx, "p", "a", "b")
	wantNoErr(t, err)
	if len(e) != 1 {
		t.Fatal(e)
	}
	r.OnGetElements = func(context.Context, string, ...string) ([]models.EditorElement, error) { return nil, errBoom }
	if _, err := s.GetElements(ctx, "p"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnGetElementsByPageID = func(_ context.Context, id string) ([]models.Element, error) {
		if id != "pg" {
			t.Fatal(id)
		}
		return []models.Element{{}, {}}, nil
	}
	m, err := s.GetElementsByPageID(ctx, "pg")
	wantNoErr(t, err)
	if len(m) != 2 {
		t.Fatal(m)
	}
	r.OnGetElementsByPageID = func(context.Context, string) ([]models.Element, error) { return nil, errBoom }
	if _, err := s.GetElementsByPageID(ctx, "pg"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnGetElementsByPageIds = func(_ context.Context, pid string, ids []string) ([]models.EditorElement, error) {
		if pid != "p" || len(ids) != 1 || ids[0] != "x" {
			t.Fatal(pid, ids)
		}
		return []models.EditorElement{nil}, nil
	}
	e, err = s.GetElementsByPageIds(ctx, "p", []string{"x"})
	wantNoErr(t, err)
	if len(e) != 1 {
		t.Fatal(e)
	}
	r.OnGetElementsByPageIds = func(context.Context, string, []string) ([]models.EditorElement, error) { return nil, errBoom }
	if _, err := s.GetElementsByPageIds(ctx, "p", nil); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnGetElementsByIDs = func(_ context.Context, ids []string) ([]models.Element, error) {
		if len(ids) != 1 || ids[0] != "e1" {
			t.Fatal(ids)
		}
		return []models.Element{{}}, nil
	}
	m, err = s.GetElementsByIDs(ctx, []string{"e1"})
	wantNoErr(t, err)
	if len(m) != 1 {
		t.Fatal(m)
	}
	r.OnGetElementsByIDs = func(context.Context, []string) ([]models.Element, error) { return nil, errBoom }
	if _, err := s.GetElementsByIDs(ctx, nil); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
}
