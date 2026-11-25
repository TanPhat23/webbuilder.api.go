package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"my-go-app/internal/models"
)

var errContentFieldValueBoom = errors.New("content field value boom")

const contentFieldValueSelect = `SELECT * FROM "public"."ContentFieldValue"`

var contentFieldValueCols = []string{"Id", "ContentItemId", "FieldId", "Value"}

func TestContentFieldValueValidationErrors(t *testing.T) {
	ctx := context.Background()
	cases := map[string]func(ContentFieldValueRepositoryInterface) error{
		"contentItemID is required": func(r ContentFieldValueRepositoryInterface) error {
			_, err := r.ListByContentItem(ctx, "")
			return err
		},
		"id is required": func(r ContentFieldValueRepositoryInterface) error {
			_, err := r.GetByID(ctx, "")
			return err
		},
		"content field value cannot be nil": func(r ContentFieldValueRepositoryInterface) error {
			_, err := r.Create(ctx, nil)
			return err
		},
		"content item ID is required": func(r ContentFieldValueRepositoryInterface) error {
			_, err := r.Create(ctx, &models.ContentFieldValue{FieldId: "f"})
			return err
		},
		"field ID is required": func(r ContentFieldValueRepositoryInterface) error {
			_, err := r.Create(ctx, &models.ContentFieldValue{ContentItemId: "i"})
			return err
		},
		"id is required (update)": func(r ContentFieldValueRepositoryInterface) error {
			_, err := r.Update(ctx, "", nil)
			return err
		},
		"id is required (delete)": func(r ContentFieldValueRepositoryInterface) error { return r.Delete(ctx, "") },
	}
	for want, call := range cases {
		db, f := newFakeDB(t)
		err := call(NewContentFieldValueRepository(db))
		if err == nil || !strings.HasPrefix(want, err.Error()) {
			t.Errorf("want error %q, got %v", want, err)
		}
		if f.queryCount() != 0 {
			t.Errorf("%s: validation must not touch the database", want)
		}
	}
}

func TestContentFieldValueGetByContentItem(t *testing.T) {
	v := "hello"
	db, f := newFakeDB(t)
	f.on(contentFieldValueSelect, contentFieldValueCols,
		[]driver.Value{"v1", "i1", "f1", v},
		[]driver.Value{"v2", "i1", "f2", nil})
	got, err := NewContentFieldValueRepository(db).ListByContentItem(context.Background(), "i1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Value == nil || *got[0].Value != v || got[1].Value != nil {
		t.Fatalf("unexpected values: %+v", got)
	}
	if !f.sawQuery(`"ContentItemId" =`) {
		t.Errorf("query: %q", f.lastQuery())
	}

	db2, f2 := newFakeDB(t)
	f2.onErr("SELECT", errContentFieldValueBoom)
	if _, err := NewContentFieldValueRepository(db2).ListByContentItem(context.Background(), "i1"); !errors.Is(err, errContentFieldValueBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestContentFieldValueGetByID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(contentFieldValueSelect, contentFieldValueCols, []driver.Value{"v1", "i1", "f1", "x"})
	got, err := NewContentFieldValueRepository(db).GetByID(ctx, "v1")
	if err != nil || got.Id != "v1" || got.FieldId != "f1" {
		t.Fatalf("got %+v, %v", got, err)
	}
	db2, _ := newFakeDB(t)
	if _, err := NewContentFieldValueRepository(db2).GetByID(ctx, "x"); !errors.Is(err, ErrContentFieldValueNotFound) {
		t.Errorf("err = %v", err)
	}
	db3, f3 := newFakeDB(t)
	f3.onErr("SELECT", errContentFieldValueBoom)
	if _, err := NewContentFieldValueRepository(db3).GetByID(ctx, "x"); !errors.Is(err, errContentFieldValueBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestContentFieldValueCreate(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("INSERT INTO", 1)
	val := "v"
	got, err := NewContentFieldValueRepository(db).Create(ctx, &models.ContentFieldValue{ContentItemId: "i1", FieldId: "f1", Value: &val})
	if err != nil || got.Id == "" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if !f.sawQuery("INSERT INTO", "ContentFieldValue") {
		t.Error("no insert issued")
	}
	got, err = NewContentFieldValueRepository(db).Create(ctx, &models.ContentFieldValue{Id: "keep", ContentItemId: "i1", FieldId: "f1"})
	if err != nil || got.Id != "keep" {
		t.Errorf("supplied id changed: %+v, %v", got, err)
	}

	db2, f2 := newFakeDB(t)
	f2.onErr("INSERT INTO", errContentFieldValueBoom)
	if _, err := NewContentFieldValueRepository(db2).Create(ctx, &models.ContentFieldValue{ContentItemId: "i1", FieldId: "f1"}); !errors.Is(err, errContentFieldValueBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestContentFieldValueUpdate(t *testing.T) {
	ctx := context.Background()
	newVal := "new"
	t.Run("updates and refetches", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("UPDATE", 1)
		f.on(contentFieldValueSelect, contentFieldValueCols, []driver.Value{"v1", "i1", "f1", newVal})
		got, err := NewContentFieldValueRepository(db).Update(ctx, "v1", &newVal)
		if err != nil || got.Value == nil || *got.Value != newVal {
			t.Fatalf("got %+v, %v", got, err)
		}
		if !f.sawQuery("UPDATE", "ContentFieldValue", `"Value"`) {
			t.Errorf("no update issued: %v", f.queries)
		}
	})
	t.Run("not found", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("UPDATE", 0)
		if _, err := NewContentFieldValueRepository(db).Update(ctx, "v1", &newVal); !errors.Is(err, ErrContentFieldValueNotFound) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("update fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("UPDATE", errContentFieldValueBoom)
		if _, err := NewContentFieldValueRepository(db).Update(ctx, "v1", &newVal); !errors.Is(err, errContentFieldValueBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("refetch fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("UPDATE", 1)
		f.onErr("SELECT", errContentFieldValueBoom)
		if _, err := NewContentFieldValueRepository(db).Update(ctx, "v1", &newVal); !errors.Is(err, errContentFieldValueBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestContentFieldValueDelete(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("DELETE FROM", 1)
	if err := NewContentFieldValueRepository(db).Delete(ctx, "v1"); err != nil {
		t.Fatal(err)
	}
	db2, _ := newFakeDB(t)
	if err := NewContentFieldValueRepository(db2).Delete(ctx, "v1"); !errors.Is(err, ErrContentFieldValueNotFound) {
		t.Errorf("err = %v", err)
	}
	db3, f3 := newFakeDB(t)
	f3.onErr("DELETE FROM", errContentFieldValueBoom)
	if err := NewContentFieldValueRepository(db3).Delete(ctx, "v1"); !errors.Is(err, errContentFieldValueBoom) {
		t.Errorf("err = %v", err)
	}
}
