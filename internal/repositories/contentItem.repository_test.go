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

func assertContentItemNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertContentItemErrIs(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("want %v, got %v", want, err)
	}
}

func assertContentItemValidationMap(t *testing.T, cases map[string]func(*ContentItemRepository) error) {
	t.Helper()
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

func TestContentItemCreateValidationErrors(t *testing.T) {
	ctx := context.Background()
	full := func() *models.ContentItem {
		return &models.ContentItem{Title: "t", Slug: "s", ContentTypeId: "ct"}
	}
	assertContentItemValidationMap(t, map[string]func(*ContentItemRepository) error{
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
	})
}

func TestContentItemLookupAndMutationValidationErrors(t *testing.T) {
	ctx := context.Background()
	assertContentItemValidationMap(t, map[string]func(*ContentItemRepository) error{
		"contentTypeID is required": func(r *ContentItemRepository) error { _, err := r.GetContentItemsByContentType(ctx, ""); return err },
		"id is required":            func(r *ContentItemRepository) error { _, err := r.GetContentItemByID(ctx, ""); return err },
		"id is required (update)": func(r *ContentItemRepository) error {
			_, err := r.UpdateContentItem(ctx, "", map[string]any{"a": 1}, nil)
			return err
		},
		"no updates provided":     func(r *ContentItemRepository) error { _, err := r.UpdateContentItem(ctx, "i1", nil, nil); return err },
		"id is required (delete)": func(r *ContentItemRepository) error { return r.DeleteContentItem(ctx, "") },
		"contentTypeID is required (public)": func(r *ContentItemRepository) error {
			_, err := r.GetPublicContentItems(ctx, "", 5, "", "")
			return err
		},
		"contentTypeID and slug are required": func(r *ContentItemRepository) error { _, err := r.GetContentItemBySlug(ctx, "ct", ""); return err },
		"contentTypeID and slug are required (type)": func(r *ContentItemRepository) error {
			_, err := r.GetContentItemBySlug(ctx, "", "s")
			return err
		},
	})
}

func TestContentItemGetByContentTypeAndID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(contentItemSelect, contentItemCols, []driver.Value{"i1", "ct1", "a", "A", true}, []driver.Value{"i2", "ct1", "b", "B", false})
	repo := contentItemNew(db)
	items, err := repo.GetContentItemsByContentType(ctx, "ct1")
	assertContentItemNoErr(t, err)
	if len(items) != 2 || items[0].Slug != "a" {
		t.Fatalf("items = %+v", items)
	}
	if !items[0].Published || items[1].Published {
		t.Fatalf("published mismatch: %+v", items)
	}
	if !f.sawQuery(`ORDER BY "CreatedAt" DESC`) {
		t.Errorf("query: %v", f.queries)
	}
	item, err := repo.GetContentItemByID(ctx, "i1")
	assertContentItemNoErr(t, err)
	if item.Id != "i1" {
		t.Fatalf("item = %+v", item)
	}
}

func TestContentItemGetErrors(t *testing.T) {
	ctx := context.Background()
	empty, emptyFake := newFakeDB(t)
	if emptyFake == nil {
		t.Fatal("missing fake db")
	}
	_, err := contentItemNew(empty).GetContentItemByID(ctx, "x")
	assertContentItemErrIs(t, err, ErrContentItemNotFound)

	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errContentItemBoom)
	br := contentItemNew(bad)
	_, err = br.GetContentItemsByContentType(ctx, "ct1")
	assertContentItemErrIs(t, err, errContentItemBoom)
	_, err = br.GetContentItemByID(ctx, "x")
	assertContentItemErrIs(t, err, errContentItemBoom)
	if errors.Is(err, ErrContentItemNotFound) {
		t.Errorf("unexpected ErrContentItemNotFound: %v", err)
	}
}

func TestContentItemCreateSuccess(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	contentItemCount(f, 0)
	f.onExec("INSERT INTO", 1)
	f.on(contentItemSelect, contentItemCols, []driver.Value{"stored", "ct1", "s", "T", false})
	val := "v"
	values := []models.ContentFieldValue{{FieldId: "f1", Value: &val}, {Id: "given", FieldId: "f2"}}
	ci := &models.ContentItem{Title: "T", Slug: "s", ContentTypeId: "ct1"}
	before := time.Now().Add(-time.Second)
	got, err := contentItemNew(db).CreateContentItem(ctx, ci, values)
	assertContentItemNoErr(t, err)
	assertCreatedContentItemFields(t, f, got, ci, values, before)
}

func assertCreatedContentItemFields(t *testing.T, f *fakeDB, got, ci *models.ContentItem, values []models.ContentFieldValue, before time.Time) {
	t.Helper()
	if got.Id != "stored" || ci.Id == "" {
		t.Fatalf("got=%+v ci=%+v", got, ci)
	}
	if ci.CreatedAt.Before(before) || !ci.CreatedAt.Equal(ci.UpdatedAt) {
		t.Fatalf("timestamps mismatch: ci=%+v", ci)
	}
	for i, v := range values {
		if v.ContentItemId != ci.Id || v.Id == "" {
			t.Errorf("value %d not linked: %+v", i, v)
		}
	}
	if values[1].Id != "given" || !f.sawQuery(contentItemValueInsert) {
		t.Errorf("values mismatch: %+v queries=%v", values, f.queries)
	}
	if f.begins != 1 || f.commits != 1 || f.aborts != 0 {
		t.Errorf("tx mismatch: begins=%d commits=%d aborts=%d", f.begins, f.commits, f.aborts)
	}
}

func TestContentItemCreateErrors(t *testing.T) {
	ctx := context.Background()
	item := &models.ContentItem{Title: "T", Slug: "s", ContentTypeId: "ct1"}

	dbDup, fDup := newFakeDB(t)
	contentItemCount(fDup, 1)
	_, err := contentItemNew(dbDup).CreateContentItem(ctx, item, nil)
	assertContentItemErrIs(t, err, ErrContentItemDuplicate)
	if fDup.begins != 0 {
		t.Error("no transaction expected")
	}

	dbChk, fChk := newFakeDB(t)
	fChk.onErr("count(*)", errContentItemBoom)
	_, err = contentItemNew(dbChk).CreateContentItem(ctx, item, nil)
	assertContentItemErrIs(t, err, errContentItemBoom)

	dbIns, fIns := newFakeDB(t)
	contentItemCount(fIns, 0)
	fIns.onErr(`INSERT INTO "public"."ContentItem"`, errContentItemBoom)
	_, err = contentItemNew(dbIns).CreateContentItem(ctx, item, nil)
	assertContentItemErrIs(t, err, errContentItemBoom)
	if fIns.commits != 0 || fIns.aborts != 1 {
		t.Errorf("commits=%d aborts=%d", fIns.commits, fIns.aborts)
	}

	dbVal, fVal := newFakeDB(t)
	contentItemCount(fVal, 0)
	fVal.onErr(contentItemValueInsert, errContentItemBoom)
	_, err = contentItemNew(dbVal).CreateContentItem(ctx, item, []models.ContentFieldValue{{FieldId: "f"}})
	assertContentItemErrIs(t, err, errContentItemBoom)
	if fVal.commits != 0 || fVal.aborts != 1 {
		t.Errorf("commits=%d aborts=%d", fVal.commits, fVal.aborts)
	}
}

func TestContentItemUpdateSuccess(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"i1", "ct1", "slug", "Title", false}
	db, f := newFakeDB(t)
	f.on(contentItemSelect, contentItemCols, row)
	contentItemCount(f, 0)
	f.onExec("DELETE FROM", 2)
	f.onExec("INSERT INTO", 1)
	f.onExec("UPDATE", 1)
	values := []models.ContentFieldValue{{FieldId: "f1"}}
	updates := map[string]any{"slug": "new", "Title": "New"}
	got, err := contentItemNew(db).UpdateContentItem(ctx, "i1", updates, values)
	assertContentItemNoErr(t, err)
	if got.Id != "i1" || values[0].ContentItemId != "i1" {
		t.Fatalf("got=%+v val=%+v", got, values[0])
	}
	if values[0].Id == "" {
		t.Fatalf("expected generated value id: %+v", values[0])
	}
	if _, ok := updates["UpdatedAt"].(time.Time); !ok {
		t.Error("UpdatedAt must be added")
	}
	if !f.sawQuery(contentItemValueDelete) || !f.sawQuery(contentItemValueInsert) {
		t.Errorf("unexpected value queries: %v", f.queries)
	}
	if !f.sawQuery("UPDATE", "ContentItem") || f.commits != 1 {
		t.Errorf("unexpected update/commit: %v commits=%d", f.queries, f.commits)
	}
}

func TestContentItemUpdatePartialModes(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"i1", "ct1", "slug", "Title", false}
	dbVals, fVals := newFakeDB(t)
	fVals.on(contentItemSelect, contentItemCols, row)
	fVals.onExec("DELETE FROM", 0)
	fVals.onExec("INSERT INTO", 1)
	_, err := contentItemNew(dbVals).UpdateContentItem(ctx, "i1", nil, []models.ContentFieldValue{{Id: "v", FieldId: "f1"}})
	assertContentItemNoErr(t, err)
	if fVals.sawQuery("UPDATE") {
		t.Error("no column update expected")
	}

	dbCols, fCols := newFakeDB(t)
	fCols.on(contentItemSelect, contentItemCols, row)
	fCols.onExec("UPDATE", 1)
	_, err = contentItemNew(dbCols).UpdateContentItem(ctx, "i1", map[string]any{"Title": "x"}, nil)
	assertContentItemNoErr(t, err)
	if fCols.sawQuery(contentItemValueDelete) {
		t.Error("field values must stay untouched")
	}
}

func TestContentItemUpdateSlugConflicts(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"i1", "ct1", "slug", "Title", false}
	for _, key := range []string{"Slug", "slug"} {
		db, f := newFakeDB(t)
		f.on(contentItemSelect, contentItemCols, row)
		contentItemCount(f, 1)
		_, err := contentItemNew(db).UpdateContentItem(ctx, "i1", map[string]any{key: "taken"}, nil)
		assertContentItemErrIs(t, err, ErrContentItemDuplicate)
		if f.sawQuery("UPDATE") || f.commits != 0 {
			t.Fatal("duplicate slug must roll back without writing")
		}
	}
}

func TestContentItemUpdateErrors(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"i1", "ct1", "slug", "Title", false}
	empty, emptyFake := newFakeDB(t)
	if emptyFake == nil {
		t.Fatal("missing fake db")
	}
	_, err := contentItemNew(empty).UpdateContentItem(ctx, "i1", map[string]any{"Title": "x"}, nil)
	assertContentItemErrIs(t, err, ErrContentItemNotFound)

	for _, tc := range []struct {
		setup func(*fakeDB)
		call  func(*ContentItemRepository) error
		abort int
	}{
		{func(f *fakeDB) { f.onErr("count(*)", errContentItemBoom) }, func(r *ContentItemRepository) error { _, err := r.UpdateContentItem(ctx, "i1", map[string]any{"slug": "x"}, nil); return err }, 1},
		{func(f *fakeDB) { f.onErr("UPDATE", errContentItemBoom) }, func(r *ContentItemRepository) error { _, err := r.UpdateContentItem(ctx, "i1", map[string]any{"Title": "x"}, nil); return err }, 1},
		{func(f *fakeDB) { f.onErr("DELETE FROM", errContentItemBoom) }, func(r *ContentItemRepository) error { _, err := r.UpdateContentItem(ctx, "i1", nil, []models.ContentFieldValue{{FieldId: "f"}}); return err }, 1},
		{func(f *fakeDB) { f.onErr("INSERT INTO", errContentItemBoom) }, func(r *ContentItemRepository) error { _, err := r.UpdateContentItem(ctx, "i1", nil, []models.ContentFieldValue{{FieldId: "f"}}); return err }, 1},
	} {
		db, f := newFakeDB(t)
		f.on(contentItemSelect, contentItemCols, row)
		tc.setup(f)
		assertContentItemErrIs(t, tc.call(contentItemNew(db)), errContentItemBoom)
		if f.aborts != tc.abort {
			t.Errorf("aborts=%d want %d", f.aborts, tc.abort)
		}
	}
}

func TestContentItemDelete(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("DELETE FROM", 1)
	assertContentItemNoErr(t, contentItemNew(db).DeleteContentItem(ctx, "i1"))

	db2, f2 := newFakeDB(t)
	if f2 == nil {
		t.Fatal("missing fake db")
	}
	assertContentItemErrIs(t, contentItemNew(db2).DeleteContentItem(ctx, "i1"), ErrContentItemNotFound)

	db3, f3 := newFakeDB(t)
	f3.onErr("DELETE FROM", errContentItemBoom)
	assertContentItemErrIs(t, contentItemNew(db3).DeleteContentItem(ctx, "i1"), errContentItemBoom)
}

func TestContentItemGetPublicNormalisesSortAndLimit(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		limit           int
		sortBy, sortDir string
		wantOrder       string
		wantLimit       int
	}{
		{5, "Title", "asc", `ORDER BY "Title" asc`, 5},
		{0, "Id; DROP TABLE x", "sideways", `ORDER BY "CreatedAt" desc`, DefaultPageSize},
		{MaxPageSize + 1, "Slug", "desc", `ORDER BY "Slug" desc`, DefaultPageSize},
		{-3, "UpdatedAt", "asc", `ORDER BY "UpdatedAt" asc`, DefaultPageSize},
		{MaxPageSize, "CreatedAt", "asc", `ORDER BY "CreatedAt" asc`, MaxPageSize},
	} {
		db, f := newFakeDB(t)
		f.on(contentItemSelect, contentItemCols, []driver.Value{"i1", "ct1", "a", "A", true})
		got, err := contentItemNew(db).GetPublicContentItems(ctx, "ct1", tc.limit, tc.sortBy, tc.sortDir)
		assertContentItemNoErr(t, err)
		if len(got) != 1 || got[0].Id != "i1" {
			t.Fatalf("got=%+v", got)
		}
		if !f.sawQuery(tc.wantOrder, `"Published" =`) {
			t.Fatalf("queries=%v", f.queries)
		}
		if !contentItemHasArg(f, int64(tc.wantLimit)) && !contentItemHasArg(f, tc.wantLimit) {
			t.Errorf("limit arg mismatch: %v", f.args[0])
		}
		if !contentItemHasArg(f, true) {
			t.Errorf("published arg mismatch: %v", f.args[0])
		}
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errContentItemBoom)
	_, err := contentItemNew(bad).GetPublicContentItems(ctx, "ct1", 1, "", "")
	assertContentItemErrIs(t, err, errContentItemBoom)
}

func TestContentItemGetBySlug(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(contentItemSelect, contentItemCols, []driver.Value{"i1", "ct1", "hello", "Hello", true})
	got, err := contentItemNew(db).GetContentItemBySlug(ctx, "ct1", "hello")
	assertContentItemNoErr(t, err)
	if got.Slug != "hello" || !contentItemHasArg(f, "hello") {
		t.Fatalf("got %+v args=%v", got, f.args[0])
	}
	if !contentItemHasArg(f, "ct1") {
		t.Fatalf("missing ct1 arg: %v", f.args[0])
	}
	empty, emptyFake := newFakeDB(t)
	if emptyFake == nil {
		t.Fatal("missing fake db")
	}
	_, err = contentItemNew(empty).GetContentItemBySlug(ctx, "ct1", "x")
	assertContentItemErrIs(t, err, ErrContentItemNotFound)

	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errContentItemBoom)
	_, err = contentItemNew(bad).GetContentItemBySlug(ctx, "ct1", "x")
	assertContentItemErrIs(t, err, errContentItemBoom)
}
