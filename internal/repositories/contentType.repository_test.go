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

// contentTypeNew returns the concrete repository: ExistsByName, ExistsByID and
// GetContentTypeByName are not part of ContentTypeRepositoryInterface.
func contentTypeNew(db *gorm.DB) *ContentTypeRepository { return &ContentTypeRepository{db: db} }

var errContentTypeBoom = errors.New("content type boom")

const contentTypeSelect = `SELECT * FROM "public"."ContentType"`

var contentTypeCols = []string{"Id", "Name", "Description"}

func contentTypeCount(f *fakeDB, n int64) *fakeRule {
	return f.on("count(*)", []string{"count"}, []driver.Value{n})
}

func TestContentTypeValidationErrors(t *testing.T) {
	ctx := context.Background()
	cases := map[string]func(*ContentTypeRepository) error{
		"id is required": func(r *ContentTypeRepository) error { _, err := r.GetContentTypeByID(ctx, ""); return err },
		"name is required": func(r *ContentTypeRepository) error {
			_, err := r.GetContentTypeByName(ctx, "")
			return err
		},
		"content type cannot be nil": func(r *ContentTypeRepository) error { _, err := r.CreateContentType(ctx, nil); return err },
		"content type name is required": func(r *ContentTypeRepository) error {
			_, err := r.CreateContentType(ctx, &models.ContentType{})
			return err
		},
		"id is required (update)": func(r *ContentTypeRepository) error {
			_, err := r.UpdateContentType(ctx, "", map[string]any{"a": 1})
			return err
		},
		"no updates provided": func(r *ContentTypeRepository) error {
			_, err := r.UpdateContentType(ctx, "ct1", map[string]any{})
			return err
		},
		"id is required (delete)": func(r *ContentTypeRepository) error { return r.DeleteContentType(ctx, "") },
		"name is required (exists)": func(r *ContentTypeRepository) error {
			_, err := r.ExistsByName(ctx, "")
			return err
		},
		"id is required (exists)": func(r *ContentTypeRepository) error { _, err := r.ExistsByID(ctx, ""); return err },
	}
	for want, call := range cases {
		db, f := newFakeDB(t)
		err := call(contentTypeNew(db))
		if err == nil || !strings.HasPrefix(want, err.Error()) {
			t.Errorf("want error %q, got %v", want, err)
		}
		if f.queryCount() != 0 {
			t.Errorf("%s: validation must not touch the database", want)
		}
	}
}

func TestContentTypeGetContentTypes(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(contentTypeSelect, contentTypeCols, []driver.Value{"ct1", "Blog", nil}, []driver.Value{"ct2", "News", "desc"})
	got, err := contentTypeNew(db).GetContentTypes(ctx)
	if err != nil || len(got) != 2 || got[0].Name != "Blog" || got[1].Description == nil || *got[1].Description != "desc" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if !f.sawQuery(`ORDER BY "CreatedAt" DESC`) {
		t.Errorf("query: %q", f.lastQuery())
	}
	db2, f2 := newFakeDB(t)
	f2.onErr("SELECT", errContentTypeBoom)
	if _, err := contentTypeNew(db2).GetContentTypes(ctx); !errors.Is(err, errContentTypeBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestContentTypeGetByIDAndName(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(contentTypeSelect, contentTypeCols, []driver.Value{"ct1", "Blog", nil})
	repo := contentTypeNew(db)
	byID, err := repo.GetContentTypeByID(ctx, "ct1")
	if err != nil || byID.Id != "ct1" {
		t.Fatalf("by id: %+v, %v", byID, err)
	}
	byName, err := repo.GetContentTypeByName(ctx, "Blog")
	if err != nil || byName.Name != "Blog" {
		t.Fatalf("by name: %+v, %v", byName, err)
	}

	empty, _ := newFakeDB(t)
	emptyRepo := contentTypeNew(empty)
	if _, err := emptyRepo.GetContentTypeByID(ctx, "x"); !errors.Is(err, ErrContentTypeNotFound) {
		t.Errorf("by id missing: %v", err)
	}
	if _, err := emptyRepo.GetContentTypeByName(ctx, "x"); !errors.Is(err, ErrContentTypeNotFound) {
		t.Errorf("by name missing: %v", err)
	}

	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errContentTypeBoom)
	badRepo := contentTypeNew(bad)
	if _, err := badRepo.GetContentTypeByID(ctx, "x"); !errors.Is(err, errContentTypeBoom) || errors.Is(err, ErrContentTypeNotFound) {
		t.Errorf("by id error: %v", err)
	}
	if _, err := badRepo.GetContentTypeByName(ctx, "x"); !errors.Is(err, errContentTypeBoom) {
		t.Errorf("by name error: %v", err)
	}
}

func TestContentTypeCreate(t *testing.T) {
	ctx := context.Background()
	t.Run("creates with id and timestamps", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentTypeCount(f, 0)
		f.onExec("INSERT INTO", 1)
		before := time.Now().Add(-time.Second)
		got, err := contentTypeNew(db).CreateContentType(ctx, &models.ContentType{Name: "Blog"})
		if err != nil {
			t.Fatal(err)
		}
		if got.Id == "" || got.CreatedAt.Before(before) || !got.CreatedAt.Equal(got.UpdatedAt) {
			t.Errorf("id/timestamps not set: %+v", got)
		}
		if !f.sawQuery("INSERT INTO", "ContentType") {
			t.Error("no insert")
		}
	})
	t.Run("keeps supplied id", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentTypeCount(f, 0)
		got, err := contentTypeNew(db).CreateContentType(ctx, &models.ContentType{Id: "mine", Name: "Blog"})
		if err != nil || got.Id != "mine" {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentTypeCount(f, 1)
		if _, err := contentTypeNew(db).CreateContentType(ctx, &models.ContentType{Name: "Blog"}); !errors.Is(err, ErrContentTypeDuplicate) {
			t.Errorf("err = %v", err)
		}
		if f.sawQuery("INSERT INTO") {
			t.Error("duplicate inserted")
		}
	})
	t.Run("existence check fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("count(*)", errContentTypeBoom)
		if _, err := contentTypeNew(db).CreateContentType(ctx, &models.ContentType{Name: "Blog"}); !errors.Is(err, errContentTypeBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("insert fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentTypeCount(f, 0)
		f.onErr("INSERT INTO", errContentTypeBoom)
		if _, err := contentTypeNew(db).CreateContentType(ctx, &models.ContentType{Name: "Blog"}); !errors.Is(err, errContentTypeBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestContentTypeUpdate(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"ct1", "Renamed", nil}
	t.Run("renames, stamps UpdatedAt, returns fresh row", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentTypeCount(f, 1).once() // exists
		contentTypeCount(f, 0)        // no other type uses the name
		f.onExec("UPDATE", 1)
		f.on(contentTypeSelect, contentTypeCols, row)
		updates := map[string]any{"name": "Renamed"}
		got, err := contentTypeNew(db).UpdateContentType(ctx, "ct1", updates)
		if err != nil || got.Name != "Renamed" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if _, ok := updates["UpdatedAt"].(time.Time); !ok {
			t.Error("UpdatedAt must be added to the updates")
		}
		if !f.sawQuery("UPDATE", "ContentType") {
			t.Error("no update issued")
		}
	})
	t.Run("update without name skips uniqueness check", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentTypeCount(f, 1)
		f.onExec("UPDATE", 1)
		f.on(contentTypeSelect, contentTypeCols, row)
		if _, err := contentTypeNew(db).UpdateContentType(ctx, "ct1", map[string]any{"Description": "d"}); err != nil {
			t.Fatal(err)
		}
		n := 0
		for _, q := range f.queries {
			if strings.Contains(q, "count(*)") {
				n++
			}
		}
		if n != 1 {
			t.Errorf("want only the existence count, got %d count queries", n)
		}
	})
	t.Run("missing", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentTypeCount(f, 0)
		if _, err := contentTypeNew(db).UpdateContentType(ctx, "ct1", map[string]any{"a": 1}); !errors.Is(err, ErrContentTypeNotFound) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("existence check fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("count(*)", errContentTypeBoom)
		if _, err := contentTypeNew(db).UpdateContentType(ctx, "ct1", map[string]any{"a": 1}); !errors.Is(err, errContentTypeBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("name taken", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentTypeCount(f, 1)
		if _, err := contentTypeNew(db).UpdateContentType(ctx, "ct1", map[string]any{"name": "Taken"}); !errors.Is(err, ErrContentTypeDuplicate) {
			t.Errorf("err = %v", err)
		}
		if f.sawQuery("UPDATE") {
			t.Error("duplicate rename written")
		}
	})
	t.Run("name check fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentTypeCount(f, 1).once()
		f.onErr("count(*)", errContentTypeBoom)
		if _, err := contentTypeNew(db).UpdateContentType(ctx, "ct1", map[string]any{"name": "x"}); !errors.Is(err, errContentTypeBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("update fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentTypeCount(f, 1)
		f.onErr("UPDATE", errContentTypeBoom)
		if _, err := contentTypeNew(db).UpdateContentType(ctx, "ct1", map[string]any{"Description": "d"}); !errors.Is(err, errContentTypeBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("refetch fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentTypeCount(f, 1)
		f.onExec("UPDATE", 1)
		f.onErr(contentTypeSelect, errContentTypeBoom)
		if _, err := contentTypeNew(db).UpdateContentType(ctx, "ct1", map[string]any{"Description": "d"}); !errors.Is(err, errContentTypeBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestContentTypeDelete(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("DELETE FROM", 1)
	if err := contentTypeNew(db).DeleteContentType(ctx, "ct1"); err != nil {
		t.Fatal(err)
	}
	db2, _ := newFakeDB(t)
	if err := contentTypeNew(db2).DeleteContentType(ctx, "ct1"); !errors.Is(err, ErrContentTypeNotFound) {
		t.Errorf("err = %v", err)
	}
	db3, f3 := newFakeDB(t)
	f3.onErr("DELETE FROM", errContentTypeBoom)
	if err := contentTypeNew(db3).DeleteContentType(ctx, "ct1"); !errors.Is(err, errContentTypeBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestContentTypeExists(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		n    int64
		want bool
	}{{0, false}, {2, true}} {
		db, f := newFakeDB(t)
		contentTypeCount(f, tc.n)
		repo := contentTypeNew(db)
		if got, err := repo.ExistsByName(ctx, "Blog"); err != nil || got != tc.want {
			t.Errorf("ExistsByName(count=%d) = %v, %v", tc.n, got, err)
		}
		if got, err := repo.ExistsByID(ctx, "ct1"); err != nil || got != tc.want {
			t.Errorf("ExistsByID(count=%d) = %v, %v", tc.n, got, err)
		}
	}
	db, f := newFakeDB(t)
	f.onErr("count(*)", errContentTypeBoom)
	repo := contentTypeNew(db)
	if got, err := repo.ExistsByName(ctx, "Blog"); got || !errors.Is(err, errContentTypeBoom) {
		t.Errorf("ExistsByName error = %v, %v", got, err)
	}
	if got, err := repo.ExistsByID(ctx, "ct1"); got || !errors.Is(err, errContentTypeBoom) {
		t.Errorf("ExistsByID error = %v, %v", got, err)
	}
}

func TestContentTypeWithCounts(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("field_count", []string{"Id", "Name", "field_count"}, []driver.Value{"ct1", "Blog", int64(3)})
	f.on("item_count", []string{"Id", "Name", "item_count"}, []driver.Value{"ct1", "Blog", int64(7)})
	repo := contentTypeNew(db)

	fields, err := repo.GetContentTypesWithFieldCount(ctx)
	if err != nil || len(fields) != 1 || fields[0]["Name"] != "Blog" || fields[0]["field_count"] != int64(3) {
		t.Fatalf("field counts: %+v, %v", fields, err)
	}
	items, err := repo.GetContentTypesWithItemCount(ctx)
	if err != nil || len(items) != 1 || items[0]["item_count"] != int64(7) {
		t.Fatalf("item counts: %+v, %v", items, err)
	}
	if !f.sawQuery("LEFT JOIN", "ContentField", "GROUP BY") || !f.sawQuery("LEFT JOIN", "ContentItem", "GROUP BY") {
		t.Errorf("queries: %v", f.queries)
	}

	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errContentTypeBoom)
	badRepo := contentTypeNew(bad)
	if _, err := badRepo.GetContentTypesWithFieldCount(ctx); !errors.Is(err, errContentTypeBoom) {
		t.Errorf("field count error: %v", err)
	}
	if _, err := badRepo.GetContentTypesWithItemCount(ctx); !errors.Is(err, errContentTypeBoom) {
		t.Errorf("item count error: %v", err)
	}
}
