package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"my-go-app/internal/models"
)

var errContentFieldBoom = errors.New("content field boom")

const contentFieldSelect = `SELECT * FROM "public"."ContentField"`

func contentFieldCount(f *fakeDB, n int64) *fakeRule {
	return f.on("count(*)", []string{"count"}, []driver.Value{n})
}

var contentFieldCols = []string{"Id", "ContentTypeId", "Name", "Type", "Required"}

func TestContentFieldValidationErrors(t *testing.T) {
	ctx := context.Background()
	cases := map[string]func(ContentFieldRepositoryInterface) error{
		"contentTypeID is required": func(r ContentFieldRepositoryInterface) error {
			_, err := r.GetContentFieldsByContentType(ctx, "")
			return err
		},
		"id is required": func(r ContentFieldRepositoryInterface) error { _, err := r.GetContentFieldByID(ctx, ""); return err },
		"content field cannot be nil": func(r ContentFieldRepositoryInterface) error {
			_, err := r.CreateContentField(ctx, nil)
			return err
		},
		"id is required (update)": func(r ContentFieldRepositoryInterface) error {
			_, err := r.UpdateContentField(ctx, "", map[string]any{"Name": "x"})
			return err
		},
		"no updates provided": func(r ContentFieldRepositoryInterface) error {
			_, err := r.UpdateContentField(ctx, "f1", nil)
			return err
		},
		"id is required (delete)": func(r ContentFieldRepositoryInterface) error { return r.DeleteContentField(ctx, "") },
	}
	for want, call := range cases {
		db, f := newFakeDB(t)
		err := call(NewContentFieldRepository(db))
		if err == nil || !strings.HasPrefix(want, err.Error()) {
			t.Errorf("want error %q, got %v", want, err)
		}
		if f.queryCount() != 0 {
			t.Errorf("%s: validation must not touch the database", want)
		}
	}
}

func TestContentFieldGetByContentType(t *testing.T) {
	db, f := newFakeDB(t)
	f.on(contentFieldSelect, contentFieldCols,
		[]driver.Value{"f1", "ct1", "alpha", "text", true},
		[]driver.Value{"f2", "ct1", "beta", "number", false})
	got, err := NewContentFieldRepository(db).GetContentFieldsByContentType(context.Background(), "ct1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "alpha" || !got[0].Required || got[1].Type != "number" {
		t.Fatalf("unexpected fields: %+v", got)
	}
	if !f.sawQuery(`"ContentTypeId" =`, `ORDER BY "Name" ASC`) {
		t.Errorf("query: %q", f.lastQuery())
	}

	db2, f2 := newFakeDB(t)
	f2.onErr("SELECT", errContentFieldBoom)
	if _, err := NewContentFieldRepository(db2).GetContentFieldsByContentType(context.Background(), "ct1"); !errors.Is(err, errContentFieldBoom) {
		t.Errorf("error not wrapped: %v", err)
	}
}

func TestContentFieldGetByID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(contentFieldSelect, contentFieldCols, []driver.Value{"f1", "ct1", "alpha", "text", false})
	got, err := NewContentFieldRepository(db).GetContentFieldByID(ctx, "f1")
	if err != nil || got.Id != "f1" || got.ContentTypeId != "ct1" {
		t.Fatalf("got %+v, %v", got, err)
	}

	db2, _ := newFakeDB(t)
	if _, err := NewContentFieldRepository(db2).GetContentFieldByID(ctx, "nope"); !errors.Is(err, ErrContentFieldNotFound) {
		t.Errorf("want ErrContentFieldNotFound, got %v", err)
	}

	db3, f3 := newFakeDB(t)
	f3.onErr("SELECT", errContentFieldBoom)
	if _, err := NewContentFieldRepository(db3).GetContentFieldByID(ctx, "f1"); !errors.Is(err, errContentFieldBoom) || errors.Is(err, ErrContentFieldNotFound) {
		t.Errorf("want wrapped boom, got %v", err)
	}
}

func TestContentFieldCreate(t *testing.T) {
	ctx := context.Background()
	t.Run("generates id and inserts", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentFieldCount(f, 0)
		f.onExec("INSERT INTO", 1)
		cf := &models.ContentField{ContentTypeId: "ct1", Name: "alpha", Type: "text"}
		got, err := NewContentFieldRepository(db).CreateContentField(ctx, cf)
		if err != nil || got.Id == "" || got.Name != "alpha" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if !f.sawQuery("INSERT INTO", "ContentField") {
			t.Error("no insert issued")
		}
	})
	t.Run("keeps supplied id", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentFieldCount(f, 0)
		got, err := NewContentFieldRepository(db).CreateContentField(ctx, &models.ContentField{Id: "mine", ContentTypeId: "ct1", Name: "a"})
		if err != nil || got.Id != "mine" {
			t.Fatalf("got %+v, %v", got, err)
		}
	})
	t.Run("duplicate name", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentFieldCount(f, 1)
		_, err := NewContentFieldRepository(db).CreateContentField(ctx, &models.ContentField{ContentTypeId: "ct1", Name: "alpha"})
		if !errors.Is(err, ErrContentFieldDuplicate) {
			t.Errorf("err = %v", err)
		}
		if f.sawQuery("INSERT INTO") {
			t.Error("duplicate must not be inserted")
		}
	})
	t.Run("existence check fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("count(*)", errContentFieldBoom)
		_, err := NewContentFieldRepository(db).CreateContentField(ctx, &models.ContentField{ContentTypeId: "ct1", Name: "alpha"})
		if !errors.Is(err, errContentFieldBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("insert fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		contentFieldCount(f, 0)
		f.onErr("INSERT INTO", errContentFieldBoom)
		_, err := NewContentFieldRepository(db).CreateContentField(ctx, &models.ContentField{ContentTypeId: "ct1", Name: "alpha"})
		if !errors.Is(err, errContentFieldBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestContentFieldUpdate(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"f1", "ct1", "renamed", "text", false}
	t.Run("updates and returns fresh row", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentFieldSelect, contentFieldCols, row)
		contentFieldCount(f, 0)
		f.onExec("UPDATE", 1)
		got, err := NewContentFieldRepository(db).UpdateContentField(ctx, "f1", map[string]any{"name": "renamed"})
		if err != nil || got.Name != "renamed" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if !f.sawQuery("count(*)", `"Id" !=`) || !f.sawQuery("UPDATE", "ContentField") {
			t.Errorf("expected uniqueness check and update, queries: %v", f.queries)
		}
	})
	t.Run("lower-case name key conflicting", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentFieldSelect, contentFieldCols, row)
		contentFieldCount(f, 1)
		_, err := NewContentFieldRepository(db).UpdateContentField(ctx, "f1", map[string]any{"name": "taken"})
		if !errors.Is(err, ErrContentFieldDuplicate) {
			t.Errorf("err = %v", err)
		}
		if f.sawQuery("UPDATE") {
			t.Error("conflicting rename must not be written")
		}
	})
	t.Run("uniqueness check failure", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentFieldSelect, contentFieldCols, row)
		f.onErr("count(*)", errContentFieldBoom)
		_, err := NewContentFieldRepository(db).UpdateContentField(ctx, "f1", map[string]any{"name": "x"})
		if !errors.Is(err, errContentFieldBoom) {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("capitalized Name rejects duplicate", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentFieldSelect, contentFieldCols, row)
		contentFieldCount(f, 1)
		_, err := NewContentFieldRepository(db).UpdateContentField(ctx, "f1", map[string]any{"Name": "taken"})
		if !errors.Is(err, ErrContentFieldDuplicate) {
			t.Fatalf("err = %v", err)
		}
		if f.sawQuery("UPDATE") {
			t.Fatal("duplicate update was written")
		}
	})
	t.Run("not found", func(t *testing.T) {
		db, _ := newFakeDB(t)
		_, err := NewContentFieldRepository(db).UpdateContentField(ctx, "f1", map[string]any{"Type": "x"})
		if !errors.Is(err, ErrContentFieldNotFound) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("update fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentFieldSelect, contentFieldCols, row)
		f.onErr("UPDATE", errContentFieldBoom)
		_, err := NewContentFieldRepository(db).UpdateContentField(ctx, "f1", map[string]any{"Type": "x"})
		if !errors.Is(err, errContentFieldBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("refetch fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(contentFieldSelect, contentFieldCols, row).once()
		f.onExec("UPDATE", 1)
		f.onErr(contentFieldSelect, errContentFieldBoom)
		_, err := NewContentFieldRepository(db).UpdateContentField(ctx, "f1", map[string]any{"Type": "x"})
		if !errors.Is(err, errContentFieldBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestContentFieldDelete(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("DELETE FROM", 1)
	if err := NewContentFieldRepository(db).DeleteContentField(ctx, "f1"); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery("DELETE FROM", "ContentField", `"Id" =`) {
		t.Errorf("query: %q", f.lastQuery())
	}

	db2, _ := newFakeDB(t)
	if err := NewContentFieldRepository(db2).DeleteContentField(ctx, "f1"); !errors.Is(err, ErrContentFieldNotFound) {
		t.Errorf("no rows deleted: err = %v", err)
	}

	db3, f3 := newFakeDB(t)
	f3.onErr("DELETE FROM", errContentFieldBoom)
	if err := NewContentFieldRepository(db3).DeleteContentField(ctx, "f1"); !errors.Is(err, errContentFieldBoom) {
		t.Errorf("err = %v", err)
	}
}
