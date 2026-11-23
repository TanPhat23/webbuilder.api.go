package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"my-go-app/internal/models"
)

var errContentItemBoom = errors.New("content item boom")

const (
	contentItemSelect      = `SELECT * FROM "public"."ContentItem"`
	contentItemValueInsert = `INSERT INTO "public"."ContentFieldValue"`
	contentItemValueDelete = `DELETE FROM "public"."ContentFieldValue"`
)

var contentItemCols = []string{"Id", "ContentTypeId", "Slug", "Title", "Published"}

func contentItemNew(db *gorm.DB) *ContentItemRepository { return &ContentItemRepository{db: db} }

func contentItemCount(f *fakeDB, n int64) *fakeRule {
	return f.on("count(*)", []string{"count"}, []driver.Value{n})
}

func contentItemHasArg(f *fakeDB, want any) bool {
	for _, q := range f.args {
		for _, a := range q {
			if a.Value == want {
				return true
			}
		}
	}
	return false
}

func TestContentItemValidationErrors(t *testing.T) {
	ctx := context.Background()
	full := func() *models.ContentItem {
		return &models.ContentItem{Title: "t", Slug: "s", ContentTypeId: "ct"}
	}
	cases := map[string]func(*ContentItemRepository) error{
		"contentTypeID is required": func(r *ContentItemRepository) error { _, err := r.GetContentItemsByContentType(ctx, ""); return err },
		"id is required":            func(r *ContentItemRepository) error { _, err := r.GetContentItemByID(ctx, ""); return err },
		"content item cannot be nil": func(r *ContentItemRepository) error {
			_, err := r.CreateContentItem(ctx, nil, nil)
			return err
		},
		"content item title is required": func(r *ContentItemRepository) error {
			ci := full()
			ci.Title = ""
			_, err := r.CreateContentItem(ctx, ci, nil)
			return err
		},
		"content item slug is required": func(r *ContentItemRepository) error {
			ci := full()
			ci.Slug = ""
			_, err := r.CreateContentItem(ctx, ci, nil)
			return err
		},
		"content type ID is required": func(r *ContentItemRepository) error {
			ci := full()
			ci.ContentTypeId = ""
			_, err := r.CreateContentItem(ctx, ci, nil)
			return err
		},
		"id is required (update)": func(r *ContentItemRepository) error {
			_, err := r.UpdateContentItem(ctx, "", map[string]any{"a": 1}, nil)
			return err
		},
		"no updates provided": func(r *ContentItemRepository) error {
			_, err := r.UpdateContentItem(ctx, "i1", nil, nil)
			return err
		},
		"id is required (delete)": func(r *ContentItemRepository) error { return r.DeleteContentItem(ctx, "") },
		"contentTypeID is required (public)": func(r *ContentItemRepository) error {
			_, err := r.GetPublicContentItems(ctx, "", 5, "", "")
			return err
		},
		"contentTypeID and slug are required": func(r *ContentItemRepository) error {
			_, err := r.GetContentItemBySlug(ctx, "ct", "")
			return err
		},
		"contentTypeID and slug are required (type)": func(r *ContentItemRepository) error {
			_, err := r.GetContentItemBySlug(ctx, "", "s")
			return err
		},
	}
	for want, call := range cases {
		db, f := newFakeDB(t)
		err := call(contentItemNew(db))
		if err == nil || !strings.HasPrefix(want, err.Error()) {
			t.Errorf("want error %q, got %v", want, err)
		}
		if f.queryCount() != 0 {
			t.Errorf("%s: validation must not touch the database", want)
		}
	}
}

func TestContentItemGetByContentTypeAndID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(contentItemSelect, contentItemCols, []driver.Value{"i1", "ct1", "a", "A", true}, []driver.Value{"i2", "ct1", "b", "B", false})
	repo := contentItemNew(db)
	items, err := repo.GetContentItemsByContentType(ctx, "ct1")
	if err != nil || len(items) != 2 || items[0].Slug != "a" || !items[0].Published || items[1].Published {
		t.Fatalf("items = %+v, %v", items, err)
	}
	if !f.sawQuery(`ORDER BY "CreatedAt" DESC`) {
		t.Errorf("query: %v", f.queries)
	}
	item, err := repo.GetContentItemByID(ctx, "i1")
	if err != nil || item.Id != "i1" {
		t.Fatalf("item = %+v, %v", item, err)
	}

	empty, _ := newFakeDB(t)
	if _, err := contentItemNew(empty).GetContentItemByID(ctx, "x"); !errors.Is(err, ErrContentItemNotFound) {
		t.Errorf("missing: %v", err)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errContentItemBoom)
	br := contentItemNew(bad)
	if _, err := br.GetContentItemsByContentType(ctx, "ct1"); !errors.Is(err, errContentItemBoom) {
		t.Errorf("list error: %v", err)
	}
	if _, err := br.GetContentItemByID(ctx, "x"); !errors.Is(err, errContentItemBoom) || errors.Is(err, ErrContentItemNotFound) {
		t.Errorf("get error: %v", err)
	}
}

func TestContentItemCreate(t *testing.T) {
	ctx := context.Background()
	t.Run("creates item and field values in one transaction", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentItemCount(f, 0)
		f.onExec("INSERT INTO", 1)
		f.on(contentItemSelect, contentItemCols, []driver.Value{"stored", "ct1", "s", "T", false})
		val := "v"
		values := []models.ContentFieldValue{{FieldId: "f1", Value: &val}, {Id: "given", FieldId: "f2"}}
		ci := &models.ContentItem{Title: "T", Slug: "s", ContentTypeId: "ct1"}
		before := time.Now().Add(-time.Second)
		got, err := contentItemNew(db).CreateContentItem(ctx, ci, values)
		if err != nil {
			t.Fatal(err)
		}
		if got.Id != "stored" {
			t.Errorf("result should be the refetched row, got %+v", got)
		}
		if ci.Id == "" || ci.CreatedAt.Before(before) || !ci.CreatedAt.Equal(ci.UpdatedAt) {
			t.Errorf("id/timestamps not set: %+v", ci)
		}
		for i, v := range values {
			if v.ContentItemId != ci.Id || v.Id == "" {
				t.Errorf("value %d not linked: %+v", i, v)
			}
		}
		if values[1].Id != "given" {
			t.Error("supplied value id changed")
		}
		if f.begins != 1 || f.commits != 1 || f.aborts != 0 {
			t.Errorf("tx begins=%d commits=%d aborts=%d", f.begins, f.commits, f.aborts)
		}
		if !f.sawQuery(contentItemValueInsert) {
			t.Error("field values not inserted")
		}
	})
	t.Run("duplicate slug", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentItemCount(f, 1)
		_, err := contentItemNew(db).CreateContentItem(ctx, &models.ContentItem{Title: "T", Slug: "s", ContentTypeId: "ct1"}, nil)
		if !errors.Is(err, ErrContentItemDuplicate) {
			t.Errorf("err = %v", err)
		}
		if f.begins != 0 {
			t.Error("no transaction expected")
		}
	})
	t.Run("existence check fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("count(*)", errContentItemBoom)
		_, err := contentItemNew(db).CreateContentItem(ctx, &models.ContentItem{Title: "T", Slug: "s", ContentTypeId: "ct1"}, nil)
		if !errors.Is(err, errContentItemBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("item insert fails and rolls back", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentItemCount(f, 0)
		f.onErr(`INSERT INTO "public"."ContentItem"`, errContentItemBoom)
		_, err := contentItemNew(db).CreateContentItem(ctx, &models.ContentItem{Title: "T", Slug: "s", ContentTypeId: "ct1"}, nil)
		if !errors.Is(err, errContentItemBoom) || f.commits != 0 || f.aborts != 1 {
			t.Errorf("err=%v commits=%d aborts=%d", err, f.commits, f.aborts)
		}
	})
	t.Run("field value insert fails and rolls back", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentItemCount(f, 0)
		f.onErr(contentItemValueInsert, errContentItemBoom)
		_, err := contentItemNew(db).CreateContentItem(ctx, &models.ContentItem{Title: "T", Slug: "s", ContentTypeId: "ct1"}, []models.ContentFieldValue{{FieldId: "f"}})
		if !errors.Is(err, errContentItemBoom) || f.commits != 0 || f.aborts != 1 {
			t.Errorf("err=%v commits=%d aborts=%d", err, f.commits, f.aborts)
		}
	})
}

func TestContentItemUpdate(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"i1", "ct1", "slug", "Title", false}
	t.Run("replaces field values and updates columns", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentItemSelect, contentItemCols, row)
		contentItemCount(f, 0)
		f.onExec("DELETE FROM", 2)
		f.onExec("INSERT INTO", 1)
		f.onExec("UPDATE", 1)
		values := []models.ContentFieldValue{{FieldId: "f1"}}
		updates := map[string]any{"slug": "new", "Title": "New"}
		got, err := contentItemNew(db).UpdateContentItem(ctx, "i1", updates, values)
		if err != nil || got.Id != "i1" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if values[0].ContentItemId != "i1" || values[0].Id == "" {
			t.Errorf("value not linked: %+v", values[0])
		}
		if _, ok := updates["UpdatedAt"].(time.Time); !ok {
			t.Error("UpdatedAt must be added")
		}
		if !f.sawQuery(contentItemValueDelete) || !f.sawQuery(contentItemValueInsert) || !f.sawQuery("UPDATE", "ContentItem") {
			t.Errorf("expected delete+insert+update, got %v", f.queries)
		}
		if f.commits != 1 {
			t.Errorf("commits = %d", f.commits)
		}
	})
	t.Run("field values only", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentItemSelect, contentItemCols, row)
		f.onExec("DELETE FROM", 0)
		f.onExec("INSERT INTO", 1)
		if _, err := contentItemNew(db).UpdateContentItem(ctx, "i1", nil, []models.ContentFieldValue{{Id: "v", FieldId: "f1"}}); err != nil {
			t.Fatal(err)
		}
		if f.sawQuery("UPDATE") {
			t.Error("no column update expected")
		}
	})
	t.Run("columns only skips field values", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentItemSelect, contentItemCols, row)
		f.onExec("UPDATE", 1)
		if _, err := contentItemNew(db).UpdateContentItem(ctx, "i1", map[string]any{"Title": "x"}, nil); err != nil {
			t.Fatal(err)
		}
		if f.sawQuery(contentItemValueDelete) {
			t.Error("field values must stay untouched")
		}
	})
	t.Run("capitalized Slug rejects duplicate", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentItemSelect, contentItemCols, row)
		contentItemCount(f, 1)
		_, err := contentItemNew(db).UpdateContentItem(ctx, "i1", map[string]any{"Slug": "taken"}, nil)
		if !errors.Is(err, ErrContentItemDuplicate) {
			t.Fatalf("err = %v", err)
		}
		if f.sawQuery("UPDATE") {
			t.Fatal("duplicate update was written")
		}
	})
	t.Run("not found", func(t *testing.T) {
		db, _ := newFakeDB(t)
		if _, err := contentItemNew(db).UpdateContentItem(ctx, "i1", map[string]any{"Title": "x"}, nil); !errors.Is(err, ErrContentItemNotFound) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("slug taken", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentItemSelect, contentItemCols, row)
		contentItemCount(f, 1)
		_, err := contentItemNew(db).UpdateContentItem(ctx, "i1", map[string]any{"slug": "taken"}, nil)
		if !errors.Is(err, ErrContentItemDuplicate) {
			t.Errorf("err = %v", err)
		}
		if f.sawQuery("UPDATE") || f.commits != 0 {
			t.Error("conflicting slug must roll back without writing")
		}
	})
	t.Run("slug check fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentItemSelect, contentItemCols, row)
		f.onErr("count(*)", errContentItemBoom)
		if _, err := contentItemNew(db).UpdateContentItem(ctx, "i1", map[string]any{"slug": "x"}, nil); !errors.Is(err, errContentItemBoom) {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("update fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentItemSelect, contentItemCols, row)
		f.onErr("UPDATE", errContentItemBoom)
		if _, err := contentItemNew(db).UpdateContentItem(ctx, "i1", map[string]any{"Title": "x"}, nil); !errors.Is(err, errContentItemBoom) || f.aborts != 1 {
			t.Errorf("err=%v aborts=%d", err, f.aborts)
		}
	})
	t.Run("deleting old values fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentItemSelect, contentItemCols, row)
		f.onErr("DELETE FROM", errContentItemBoom)
		if _, err := contentItemNew(db).UpdateContentItem(ctx, "i1", nil, []models.ContentFieldValue{{FieldId: "f"}}); !errors.Is(err, errContentItemBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("inserting new values fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentItemSelect, contentItemCols, row)
		f.onErr("INSERT INTO", errContentItemBoom)
		if _, err := contentItemNew(db).UpdateContentItem(ctx, "i1", nil, []models.ContentFieldValue{{FieldId: "f"}}); !errors.Is(err, errContentItemBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestContentItemDelete(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("DELETE FROM", 1)
	if err := contentItemNew(db).DeleteContentItem(ctx, "i1"); err != nil {
		t.Fatal(err)
	}
	db2, _ := newFakeDB(t)
	if err := contentItemNew(db2).DeleteContentItem(ctx, "i1"); !errors.Is(err, ErrContentItemNotFound) {
		t.Errorf("err = %v", err)
	}
	db3, f3 := newFakeDB(t)
	f3.onErr("DELETE FROM", errContentItemBoom)
	if err := contentItemNew(db3).DeleteContentItem(ctx, "i1"); !errors.Is(err, errContentItemBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestContentItemGetPublicNormalisesSortAndLimit(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name            string
		limit           int
		sortBy, sortDir string
		wantOrder       string
		wantLimit       int
	}{
		{"valid values kept", 5, "Title", "asc", `ORDER BY "Title" asc`, 5},
		{"unknown sort field and direction fall back", 0, "Id; DROP TABLE x", "sideways", `ORDER BY "CreatedAt" desc`, DefaultPageSize},
		{"too large limit falls back", MaxPageSize + 1, "Slug", "desc", `ORDER BY "Slug" desc`, DefaultPageSize},
		{"negative limit falls back", -3, "UpdatedAt", "asc", `ORDER BY "UpdatedAt" asc`, DefaultPageSize},
		{"max limit allowed", MaxPageSize, "CreatedAt", "asc", `ORDER BY "CreatedAt" asc`, MaxPageSize},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			f.on(contentItemSelect, contentItemCols, []driver.Value{"i1", "ct1", "a", "A", true})
			got, err := contentItemNew(db).GetPublicContentItems(ctx, "ct1", tc.limit, tc.sortBy, tc.sortDir)
			if err != nil || len(got) != 1 || got[0].Id != "i1" {
				t.Fatalf("got %+v, %v", got, err)
			}
			if !f.sawQuery(tc.wantOrder, `"Published" =`) {
				t.Errorf("want %s in %q", tc.wantOrder, f.queries[0])
			}
			if !contentItemHasArg(f, int64(tc.wantLimit)) && !contentItemHasArg(f, tc.wantLimit) {
				t.Errorf("limit %d not bound: %v", tc.wantLimit, f.args[0])
			}
			if !contentItemHasArg(f, true) {
				t.Error("published=true must be bound")
			}
		})
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errContentItemBoom)
	if _, err := contentItemNew(bad).GetPublicContentItems(ctx, "ct1", 1, "", ""); !errors.Is(err, errContentItemBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestContentItemGetBySlug(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(contentItemSelect, contentItemCols, []driver.Value{"i1", "ct1", "hello", "Hello", true})
	got, err := contentItemNew(db).GetContentItemBySlug(ctx, "ct1", "hello")
	if err != nil || got.Slug != "hello" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if !contentItemHasArg(f, "hello") || !contentItemHasArg(f, "ct1") {
		t.Errorf("args: %v", f.args[0])
	}
	empty, _ := newFakeDB(t)
	if _, err := contentItemNew(empty).GetContentItemBySlug(ctx, "ct1", "x"); !errors.Is(err, ErrContentItemNotFound) {
		t.Errorf("missing: %v", err)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errContentItemBoom)
	if _, err := contentItemNew(bad).GetContentItemBySlug(ctx, "ct1", "x"); !errors.Is(err, errContentItemBoom) {
		t.Errorf("err = %v", err)
	}
}
