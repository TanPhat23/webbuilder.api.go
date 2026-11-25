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
	s := NewElementQueryService(r)

	r.OnList = func(_ context.Context, pid string, pages ...string) ([]models.EditorElement, error) {
		if pid != "p" || len(pages) != 2 || pages[0] != "a" || pages[1] != "b" {
			t.Fatal(pid, pages)
		}
		return []models.EditorElement{nil}, nil
	}
	e, err := s.List(ctx, "p", "a", "b")
	wantNoErr(t, err)
	if len(e) != 1 {
		t.Fatal(e)
	}
	r.OnList = func(context.Context, string, ...string) ([]models.EditorElement, error) { return nil, errBoom }
	if _, err := s.List(ctx, "p"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnListByPage = func(_ context.Context, id string) ([]models.Element, error) {
		if id != "pg" {
			t.Fatal(id)
		}
		return []models.Element{{}, {}}, nil
	}
	m, err := s.ListByPage(ctx, "pg")
	wantNoErr(t, err)
	if len(m) != 2 {
		t.Fatal(m)
	}
	r.OnListByPage = func(context.Context, string) ([]models.Element, error) { return nil, errBoom }
	if _, err := s.ListByPage(ctx, "pg"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnListByPageIDs = func(_ context.Context, pid string, ids []string) ([]models.EditorElement, error) {
		if pid != "p" || len(ids) != 1 || ids[0] != "x" {
			t.Fatal(pid, ids)
		}
		return []models.EditorElement{nil}, nil
	}
	e, err = s.ListByPageIDs(ctx, "p", []string{"x"})
	wantNoErr(t, err)
	if len(e) != 1 {
		t.Fatal(e)
	}
	r.OnListByPageIDs = func(context.Context, string, []string) ([]models.EditorElement, error) { return nil, errBoom }
	if _, err := s.ListByPageIDs(ctx, "p", nil); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}

	r.OnListByIDs = func(_ context.Context, ids []string) ([]models.Element, error) {
		if len(ids) != 1 || ids[0] != "e1" {
			t.Fatal(ids)
		}
		return []models.Element{{}}, nil
	}
	m, err = s.ListByIDs(ctx, []string{"e1"})
	wantNoErr(t, err)
	if len(m) != 1 {
		t.Fatal(m)
	}
	r.OnListByIDs = func(context.Context, []string) ([]models.Element, error) { return nil, errBoom }
	if _, err := s.ListByIDs(ctx, nil); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
}
