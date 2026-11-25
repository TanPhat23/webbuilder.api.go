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

func contentTypeNew(db *gorm.DB) *ContentTypeRepository { return &ContentTypeRepository{db: db} }

var errContentTypeBoom = errors.New("content type boom")

const contentTypeSelect = `SELECT * FROM "public"."ContentType"`

var contentTypeCols = []string{"Id", "Name", "Description"}

func contentTypeCount(f *fakeDB, n int64) *fakeRule {
	return f.on("count(*)", []string{"count"}, []driver.Value{n})
}

func assertContentTypeNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertContentTypeErrIs(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("want %v, got %v", want, err)
	}
}

func TestContentTypeValidationErrors(t *testing.T) {
	ctx := context.Background()
	cases := map[string]func(*ContentTypeRepository) error{
		"id is required": func(r *ContentTypeRepository) error { _, err := r.GetByID(ctx, ""); return err },
		"name is required": func(r *ContentTypeRepository) error {
			_, err := r.GetByName(ctx, "")
			return err
		},
		"content type cannot be nil": func(r *ContentTypeRepository) error { _, err := r.Create(ctx, nil); return err },
		"content type name is required": func(r *ContentTypeRepository) error {
			_, err := r.Create(ctx, &models.ContentType{})
			return err
		},
		"id is required (update)": func(r *ContentTypeRepository) error {
			_, err := r.Update(ctx, "", map[string]any{"a": 1})
			return err
		},
		"no updates provided": func(r *ContentTypeRepository) error {
			_, err := r.Update(ctx, "ct1", map[string]any{})
			return err
		},
		"id is required (delete)": func(r *ContentTypeRepository) error { return r.Delete(ctx, "") },
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
	got, err := contentTypeNew(db).List(ctx)
	assertContentTypeNoErr(t, err)
	if len(got) != 2 || got[0].Name != "Blog" || got[1].Description == nil || *got[1].Description != "desc" {
		t.Fatalf("got %+v", got)
	}
	if !f.sawQuery(`ORDER BY "CreatedAt" DESC`) {
		t.Errorf("query: %q", f.lastQuery())
	}
	db2, f2 := newFakeDB(t)
	f2.onErr("SELECT", errContentTypeBoom)
	_, err = contentTypeNew(db2).List(ctx)
	assertContentTypeErrIs(t, err, errContentTypeBoom)
}

func TestContentTypeGetByIDAndName(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(contentTypeSelect, contentTypeCols, []driver.Value{"ct1", "Blog", nil})
	repo := contentTypeNew(db)
	byID, err := repo.GetByID(ctx, "ct1")
	assertContentTypeNoErr(t, err)
	if byID.Id != "ct1" {
		t.Fatalf("by id: %+v", byID)
	}
	byName, err := repo.GetByName(ctx, "Blog")
	assertContentTypeNoErr(t, err)
	if byName.Name != "Blog" {
		t.Fatalf("by name: %+v", byName)
	}
}

func TestContentTypeGetByIDAndNameErrors(t *testing.T) {
	ctx := context.Background()
	empty, _ := newFakeDB(t)
	emptyRepo := contentTypeNew(empty)
	_, err := emptyRepo.GetByID(ctx, "x")
	assertContentTypeErrIs(t, err, ErrContentTypeNotFound)
	_, err = emptyRepo.GetByName(ctx, "x")
	assertContentTypeErrIs(t, err, ErrContentTypeNotFound)

	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errContentTypeBoom)
	badRepo := contentTypeNew(bad)
	_, err = badRepo.GetByID(ctx, "x")
	assertContentTypeErrIs(t, err, errContentTypeBoom)
	_, err = badRepo.GetByName(ctx, "x")
	assertContentTypeErrIs(t, err, errContentTypeBoom)
}

func TestContentTypeCreate(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	contentTypeCount(f, 0)
	f.onExec("INSERT INTO", 1)
	before := time.Now().Add(-time.Second)
	got, err := contentTypeNew(db).Create(ctx, &models.ContentType{Name: "Blog"})
	assertContentTypeNoErr(t, err)
	if got.Id == "" || got.CreatedAt.Before(before) || !got.CreatedAt.Equal(got.UpdatedAt) || !f.sawQuery("INSERT INTO", "ContentType") {
		t.Errorf("id/timestamps/insert invalid: %+v %v", got, f.queries)
	}

	dbKeep, fKeep := newFakeDB(t)
	contentTypeCount(fKeep, 0)
	keepGot, err := contentTypeNew(dbKeep).Create(ctx, &models.ContentType{Id: "mine", Name: "Blog"})
	assertContentTypeNoErr(t, err)
	if keepGot.Id != "mine" {
		t.Fatalf("got %+v", keepGot)
	}
}

func TestContentTypeCreateErrors(t *testing.T) {
	ctx := context.Background()
	dbDup, fDup := newFakeDB(t)
	contentTypeCount(fDup, 1)
	_, err := contentTypeNew(dbDup).Create(ctx, &models.ContentType{Name: "Blog"})
	assertContentTypeErrIs(t, err, ErrContentTypeDuplicate)
	if fDup.sawQuery("INSERT INTO") {
		t.Error("duplicate inserted")
	}

	dbChk, fChk := newFakeDB(t)
	fChk.onErr("count(*)", errContentTypeBoom)
	_, err = contentTypeNew(dbChk).Create(ctx, &models.ContentType{Name: "Blog"})
	assertContentTypeErrIs(t, err, errContentTypeBoom)

	dbIns, fIns := newFakeDB(t)
	contentTypeCount(fIns, 0)
	fIns.onErr("INSERT INTO", errContentTypeBoom)
	_, err = contentTypeNew(dbIns).Create(ctx, &models.ContentType{Name: "Blog"})
	assertContentTypeErrIs(t, err, errContentTypeBoom)
}

func TestContentTypeUpdateSuccess(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"ct1", "Renamed", nil}
	db, f := newFakeDB(t)
	contentTypeCount(f, 1).once()
	contentTypeCount(f, 0)
	f.onExec("UPDATE", 1)
	f.on(contentTypeSelect, contentTypeCols, row)
	updates := map[string]any{"name": "Renamed"}
	got, err := contentTypeNew(db).Update(ctx, "ct1", updates)
	assertContentTypeNoErr(t, err)
	if got.Name != "Renamed" || !f.sawQuery("UPDATE", "ContentType") {
		t.Fatalf("got=%+v queries=%v", got, f.queries)
	}
	if _, ok := updates["UpdatedAt"].(time.Time); !ok {
		t.Error("UpdatedAt must be added")
	}
}

func TestContentTypeUpdateNoNameAndErrors(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"ct1", "Renamed", nil}
	db, f := newFakeDB(t)
	contentTypeCount(f, 1)
	f.onExec("UPDATE", 1)
	f.on(contentTypeSelect, contentTypeCols, row)
	_, err := contentTypeNew(db).Update(ctx, "ct1", map[string]any{"Description": "d"})
	assertContentTypeNoErr(t, err)
	assertSingleCountQuery(t, f.queries)

	for _, tc := range []struct {
		setup   func(*fakeDB)
		updates map[string]any
		want    error
	}{
		{func(f *fakeDB) { contentTypeCount(f, 0) }, map[string]any{"a": 1}, ErrContentTypeNotFound},
		{func(f *fakeDB) { f.onErr("count(*)", errContentTypeBoom) }, map[string]any{"a": 1}, errContentTypeBoom},
		{func(f *fakeDB) { contentTypeCount(f, 1) }, map[string]any{"name": "Taken"}, ErrContentTypeDuplicate},
		{func(f *fakeDB) { contentTypeCount(f, 1).once(); f.onErr("count(*)", errContentTypeBoom) }, map[string]any{"name": "x"}, errContentTypeBoom},
		{func(f *fakeDB) { contentTypeCount(f, 1); f.onErr("UPDATE", errContentTypeBoom) }, map[string]any{"Description": "d"}, errContentTypeBoom},
		{func(f *fakeDB) {
			contentTypeCount(f, 1)
			f.onExec("UPDATE", 1)
			f.onErr(contentTypeSelect, errContentTypeBoom)
		}, map[string]any{"Description": "d"}, errContentTypeBoom},
	} {
		d, fd := newFakeDB(t)
		tc.setup(fd)
		_, err := contentTypeNew(d).Update(ctx, "ct1", tc.updates)
		assertContentTypeErrIs(t, err, tc.want)
	}
}

func assertSingleCountQuery(t *testing.T, queries []string) {
	t.Helper()
	n := 0
	for _, q := range queries {
		if strings.Contains(q, "count(*)") {
			n++
		}
	}
	if n != 1 {
		t.Errorf("want only the existence count, got %d", n)
	}
}

func TestContentTypeDelete(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("DELETE FROM", 1)
	assertContentTypeNoErr(t, contentTypeNew(db).Delete(ctx, "ct1"))

	db2, _ := newFakeDB(t)
	assertContentTypeErrIs(t, contentTypeNew(db2).Delete(ctx, "ct1"), ErrContentTypeNotFound)

	db3, f3 := newFakeDB(t)
	f3.onErr("DELETE FROM", errContentTypeBoom)
	assertContentTypeErrIs(t, contentTypeNew(db3).Delete(ctx, "ct1"), errContentTypeBoom)
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
		gotName, err := repo.ExistsByName(ctx, "Blog")
		assertContentTypeNoErr(t, err)
		gotID, err := repo.ExistsByID(ctx, "ct1")
		assertContentTypeNoErr(t, err)
		if gotName != tc.want || gotID != tc.want {
			t.Errorf("count=%d: name=%v id=%v", tc.n, gotName, gotID)
		}
	}
	db, f := newFakeDB(t)
	f.onErr("count(*)", errContentTypeBoom)
	repo := contentTypeNew(db)
	_, err := repo.ExistsByName(ctx, "Blog")
	assertContentTypeErrIs(t, err, errContentTypeBoom)
	_, err = repo.ExistsByID(ctx, "ct1")
	assertContentTypeErrIs(t, err, errContentTypeBoom)
}

func TestContentTypeWithCounts(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("field_count", []string{"Id", "Name", "field_count"}, []driver.Value{"ct1", "Blog", int64(3)})
	f.on("item_count", []string{"Id", "Name", "item_count"}, []driver.Value{"ct1", "Blog", int64(7)})
	repo := contentTypeNew(db)

	fields, err := repo.ListWithFieldCount(ctx)
	assertContentTypeNoErr(t, err)
	items, err := repo.ListWithItemCount(ctx)
	assertContentTypeNoErr(t, err)
	assertContentTypeCountsResult(t, f, fields, items)

	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errContentTypeBoom)
	badRepo := contentTypeNew(bad)
	_, err = badRepo.ListWithFieldCount(ctx)
	assertContentTypeErrIs(t, err, errContentTypeBoom)
	_, err = badRepo.ListWithItemCount(ctx)
	assertContentTypeErrIs(t, err, errContentTypeBoom)
}

func assertContentTypeCountsResult(t *testing.T, f *fakeDB, fields, items []map[string]any) {
	t.Helper()
	if len(fields) != 1 || fields[0]["Name"] != "Blog" || fields[0]["field_count"] != int64(3) {
		t.Fatalf("field counts: %+v", fields)
	}
	if len(items) != 1 || items[0]["item_count"] != int64(7) {
		t.Fatalf("item counts: %+v", items)
	}
	if !f.sawQuery("LEFT JOIN", "ContentField", "GROUP BY") || !f.sawQuery("LEFT JOIN", "ContentItem", "GROUP BY") {
		t.Errorf("queries: %v", f.queries)
	}
}

func TestContentTypeGetByProjectAndSlugOrIDRequiresPublishedProject(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(`JOIN public."Project"`, contentTypeCols, []driver.Value{"ct1", "Blog", nil})
	got, err := contentTypeNew(db).GetByProjectAndSlugOrID(ctx, "proj1", "blog")
	assertContentTypeNoErr(t, err)
	if got.Id != "ct1" {
		t.Fatalf("got=%+v", got)
	}
	if !f.sawQuery(`JOIN public."Project"`, `public."Project"."Published" = true`, `public."Project"."DeletedAt" IS NULL`) {
		t.Fatalf("expected published non-deleted project join, got queries=%v", f.queries)
	}
}
