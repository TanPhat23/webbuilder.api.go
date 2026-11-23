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
	if len(got) != 2 || got[0].Name != "alpha" {
		t.Fatalf("unexpected fields: %+v", got)
	}
	if !got[0].Required || got[1].Type != "number" {
		t.Fatalf("unexpected field attrs: %+v", got)
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
	if err != nil || got.Id != "f1" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got.ContentTypeId != "ct1" {
		t.Fatalf("contentTypeId=%q", got.ContentTypeId)
	}

	db2, f2 := newFakeDB(t)
	if f2 == nil {
		t.Fatal("missing fake db")
	}
	if _, err := NewContentFieldRepository(db2).GetContentFieldByID(ctx, "nope"); !errors.Is(err, ErrContentFieldNotFound) {
		t.Errorf("want ErrContentFieldNotFound, got %v", err)
	}

	db3, f3 := newFakeDB(t)
	f3.onErr("SELECT", errContentFieldBoom)
	if _, err := NewContentFieldRepository(db3).GetContentFieldByID(ctx, "f1"); !errors.Is(err, errContentFieldBoom) || errors.Is(err, ErrContentFieldNotFound) {
		t.Errorf("want wrapped boom, got %v", err)
	}
}

func TestContentFieldCreateSuccess(t *testing.T) {
	ctx := context.Background()
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

	dbKeep, fKeep := newFakeDB(t)
	contentFieldCount(fKeep, 0)
	kept, err := NewContentFieldRepository(dbKeep).CreateContentField(ctx, &models.ContentField{Id: "mine", ContentTypeId: "ct1", Name: "a"})
	if err != nil || kept.Id != "mine" {
		t.Fatalf("got %+v, %v", kept, err)
	}
}

func TestContentFieldCreateErrors(t *testing.T) {
	ctx := context.Background()
	dbDup, fDup := newFakeDB(t)
	contentFieldCount(fDup, 1)
	_, err := NewContentFieldRepository(dbDup).CreateContentField(ctx, &models.ContentField{ContentTypeId: "ct1", Name: "alpha"})
	if !errors.Is(err, ErrContentFieldDuplicate) {
		t.Errorf("err = %v", err)
	}
	if fDup.sawQuery("INSERT INTO") {
		t.Error("duplicate must not be inserted")
	}

	dbChk, fChk := newFakeDB(t)
	fChk.onErr("count(*)", errContentFieldBoom)
	if _, err := NewContentFieldRepository(dbChk).CreateContentField(ctx, &models.ContentField{ContentTypeId: "ct1", Name: "alpha"}); !errors.Is(err, errContentFieldBoom) {
		t.Errorf("err = %v", err)
	}

	dbIns, fIns := newFakeDB(t)
	contentFieldCount(fIns, 0)
	fIns.onErr("INSERT INTO", errContentFieldBoom)
	if _, err := NewContentFieldRepository(dbIns).CreateContentField(ctx, &models.ContentField{ContentTypeId: "ct1", Name: "alpha"}); !errors.Is(err, errContentFieldBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestContentFieldUpdateSuccessAndConflicts(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"f1", "ct1", "renamed", "text", false}
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

	for _, key := range []string{"name", "Name"} {
		dbDup, fDup := newFakeDB(t)
		fDup.on(contentFieldSelect, contentFieldCols, row)
		contentFieldCount(fDup, 1)
		if _, err := NewContentFieldRepository(dbDup).UpdateContentField(ctx, "f1", map[string]any{key: "taken"}); !errors.Is(err, ErrContentFieldDuplicate) {
			t.Fatalf("%s: err = %v", key, err)
		}
		if fDup.sawQuery("UPDATE") {
			t.Fatalf("%s: conflicting rename must not be written", key)
		}
	}
}

func TestContentFieldUpdateErrors(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"f1", "ct1", "renamed", "text", false}

	dbChk, fChk := newFakeDB(t)
	fChk.on(contentFieldSelect, contentFieldCols, row)
	fChk.onErr("count(*)", errContentFieldBoom)
	if _, err := NewContentFieldRepository(dbChk).UpdateContentField(ctx, "f1", map[string]any{"name": "x"}); !errors.Is(err, errContentFieldBoom) {
		t.Errorf("err = %v", err)
	}

	dbMiss, missFake := newFakeDB(t)
	if missFake == nil {
		t.Fatal("missing fake db")
	}
	if _, err := NewContentFieldRepository(dbMiss).UpdateContentField(ctx, "f1", map[string]any{"Type": "x"}); !errors.Is(err, ErrContentFieldNotFound) {
		t.Errorf("err = %v", err)
	}

	dbUpd, fUpd := newFakeDB(t)
	fUpd.on(contentFieldSelect, contentFieldCols, row)
	fUpd.onErr("UPDATE", errContentFieldBoom)
	if _, err := NewContentFieldRepository(dbUpd).UpdateContentField(ctx, "f1", map[string]any{"Type": "x"}); !errors.Is(err, errContentFieldBoom) {
		t.Errorf("err = %v", err)
	}

	dbRef, fRef := newFakeDB(t)
	fRef.on(contentFieldSelect, contentFieldCols, row).once()
	fRef.onExec("UPDATE", 1)
	fRef.onErr(contentFieldSelect, errContentFieldBoom)
	if _, err := NewContentFieldRepository(dbRef).UpdateContentField(ctx, "f1", map[string]any{"Type": "x"}); !errors.Is(err, errContentFieldBoom) {
		t.Errorf("err = %v", err)
	}
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

	db2, f2 := newFakeDB(t)
	if f2 == nil {
		t.Fatal("missing fake db")
	}
	if err := NewContentFieldRepository(db2).DeleteContentField(ctx, "f1"); !errors.Is(err, ErrContentFieldNotFound) {
		t.Errorf("no rows deleted: err = %v", err)
	}

	db3, f3 := newFakeDB(t)
	f3.onErr("DELETE FROM", errContentFieldBoom)
	if err := NewContentFieldRepository(db3).DeleteContentField(ctx, "f1"); !errors.Is(err, errContentFieldBoom) {
		t.Errorf("err = %v", err)
	}
}
