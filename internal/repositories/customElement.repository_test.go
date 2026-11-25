package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"my-go-app/internal/models"
)

var errCustomElementBoom = errors.New("custom element boom")

const customElementSelect = `SELECT * FROM "public"."CustomElement"`

var customElementCols = []string{"Id", "UserId", "Name", "Category", "IsPublic", "Version", "Description", "Tags"}

func customElementNew(db *gorm.DB) *CustomElementRepository { return &CustomElementRepository{db: db} }

func customElementHasArg(f *fakeDB, want any) bool {
	for _, q := range f.args {
		for _, a := range q {
			if a.Value == want {
				return true
			}
		}
	}
	return false
}

func customElementNameConflict() error {
	return &pgconn.PgError{Code: "23505", ConstraintName: "CustomElement_UserId_Name_key"}
}

func assertCustomElementNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertCustomElementErrIs(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("want %v, got %v", want, err)
	}
}

func TestCustomElementNameConflictDetection(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{"matching constraint", customElementNameConflict(), true},
		{"wrapped", fmt.Errorf("wrap: %w", customElementNameConflict()), true},
		{"other constraint", &pgconn.PgError{Code: "23505", ConstraintName: "other"}, false},
		{"other code", &pgconn.PgError{Code: "23503", ConstraintName: "CustomElement_UserId_Name_key"}, false},
		{"not a pg error", errCustomElementBoom, false},
		{"nil", nil, false},
	} {
		if got := isCustomElementNameConflict(tc.err); got != tc.want {
			t.Errorf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}
}

func TestCustomElementValidationErrors(t *testing.T) {
	ctx := context.Background()
	cases := map[string]func(*CustomElementRepository) error{
		"userID is required": func(r *CustomElementRepository) error { _, err := r.List(ctx, "", nil); return err },
		"id is required":     func(r *CustomElementRepository) error { _, err := r.GetByID(ctx, "", "u"); return err },
		"userID is required (get)": func(r *CustomElementRepository) error {
			_, err := r.GetByID(ctx, "e", "")
			return err
		},
		"customElement is required": func(r *CustomElementRepository) error { _, err := r.Create(ctx, nil); return err },
		"userId is required": func(r *CustomElementRepository) error {
			_, err := r.Create(ctx, &models.CustomElement{Name: "n"})
			return err
		},
		"name is required": func(r *CustomElementRepository) error {
			_, err := r.Create(ctx, &models.CustomElement{UserId: "u"})
			return err
		},
		"id is required (update)": func(r *CustomElementRepository) error {
			_, err := r.Update(ctx, "", "u", nil)
			return err
		},
		"userID is required (update)": func(r *CustomElementRepository) error {
			_, err := r.Update(ctx, "e", "", nil)
			return err
		},
		"id is required (delete)":     func(r *CustomElementRepository) error { return r.Delete(ctx, "", "u") },
		"userID is required (delete)": func(r *CustomElementRepository) error { return r.Delete(ctx, "e", "") },
		"id is required (duplicate)": func(r *CustomElementRepository) error {
			_, err := r.Duplicate(ctx, "", "u", "n")
			return err
		},
		"userID is required (duplicate)": func(r *CustomElementRepository) error {
			_, err := r.Duplicate(ctx, "e", "", "n")
			return err
		},
		"newName is required": func(r *CustomElementRepository) error {
			_, err := r.Duplicate(ctx, "e", "u", "")
			return err
		},
	}
	for want, call := range cases {
		db, f := newFakeDB(t)
		err := call(customElementNew(db))
		if err == nil || !strings.HasPrefix(want, err.Error()) {
			t.Errorf("want error %q, got %v", want, err)
		}
		if f.queryCount() != 0 {
			t.Errorf("%s: validation must not touch the database", want)
		}
	}
}

func TestCustomElementGetCustomElements(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(customElementSelect, customElementCols, []driver.Value{"e1", "u1", "Card", nil, true, "1.0.0", nil, nil})
	got, err := customElementNew(db).List(ctx, "u1", nil)
	assertCustomElementNoErr(t, err)
	if len(got) != 1 || got[0].Name != "Card" || !got[0].IsPublic {
		t.Fatalf("got %+v", got)
	}
	if f.sawQuery(`"IsPublic" =`) || !f.sawQuery(`"UserId" =`, `ORDER BY "CreatedAt" DESC`) {
		t.Errorf("query: %q", f.queries[0])
	}

	dbPub, fPub := newFakeDB(t)
	pub := false
	_, err = customElementNew(dbPub).List(ctx, "u1", &pub)
	assertCustomElementNoErr(t, err)
	if !fPub.sawQuery(`"IsPublic" =`) || !customElementHasArg(fPub, false) {
		t.Errorf("filter missing: %q %v", fPub.queries[0], fPub.args[0])
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("SELECT", errCustomElementBoom)
	_, err = customElementNew(dbErr).List(ctx, "u1", nil)
	assertCustomElementErrIs(t, err, errCustomElementBoom)
}

func TestCustomElementGetByID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(customElementSelect, customElementCols, []driver.Value{"e1", "other", "Card", nil, true, "1.0.0", nil, nil})
	got, err := customElementNew(db).GetByID(ctx, "e1", "u1")
	assertCustomElementNoErr(t, err)
	if got.Id != "e1" || !f.sawQuery(`"UserId" = $2 OR "IsPublic" = $3`) {
		t.Fatalf("got=%+v queries=%v", got, f.queries)
	}

	empty, _ := newFakeDB(t)
	_, err = customElementNew(empty).GetByID(ctx, "e1", "u1")
	assertCustomElementErrIs(t, err, ErrCustomElementNotFound)

	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errCustomElementBoom)
	_, err = customElementNew(bad).GetByID(ctx, "e1", "u1")
	assertCustomElementErrIs(t, err, errCustomElementBoom)
}

func TestCustomElementCreate(t *testing.T) {
	ctx := context.Background()
	in := func() *models.CustomElement {
		return &models.CustomElement{Id: "e1", UserId: "u1", Name: "Card", Structure: []byte("{}")}
	}
	db, f := newFakeDB(t)
	f.onExec("INSERT INTO", 1)
	got, err := customElementNew(db).Create(ctx, in())
	assertCustomElementNoErr(t, err)
	if got.Id != "e1" || !f.sawQuery("INSERT INTO", "CustomElement") {
		t.Fatalf("got %+v", got)
	}

	dup, fd := newFakeDB(t)
	fd.onErr("INSERT INTO", customElementNameConflict())
	_, err = customElementNew(dup).Create(ctx, in())
	assertCustomElementErrIs(t, err, ErrCustomElementAlreadyExists)

	bad, fb := newFakeDB(t)
	fb.onErr("INSERT INTO", errCustomElementBoom)
	_, err = customElementNew(bad).Create(ctx, in())
	assertCustomElementErrIs(t, err, errCustomElementBoom)
	if errors.Is(err, ErrCustomElementAlreadyExists) {
		t.Errorf("unexpected conflict: %v", err)
	}
}

func TestCustomElementUpdateSuccess(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"e1", "u1", "Renamed", nil, false, "1.0.0", nil, nil}
	db, f := newFakeDB(t)
	f.on(customElementSelect, customElementCols, row)
	f.onExec("UPDATE", 1)
	got, err := customElementNew(db).Update(ctx, "e1", "u1", map[string]any{"Name": "Renamed"})
	assertCustomElementNoErr(t, err)
	if got.Name != "Renamed" || !f.sawQuery("UPDATE", "CustomElement", `"UserId" =`) {
		t.Fatalf("got=%+v queries=%v", got, f.queries)
	}
}

func TestCustomElementUpdateErrors(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"e1", "u1", "Renamed", nil, false, "1.0.0", nil, nil}
	for _, tc := range []struct {
		setup func(*fakeDB)
		want  error
	}{
		{func(*fakeDB) {}, ErrCustomElementUnauthorized},
		{func(f *fakeDB) { f.onErr("SELECT", errCustomElementBoom) }, errCustomElementBoom},
		{func(f *fakeDB) {
			f.on(customElementSelect, customElementCols, row)
			f.onErr("UPDATE", customElementNameConflict())
		}, ErrCustomElementAlreadyExists},
		{func(f *fakeDB) {
			f.on(customElementSelect, customElementCols, row)
			f.onErr("UPDATE", errCustomElementBoom)
		}, errCustomElementBoom},
		{func(f *fakeDB) { f.on(customElementSelect, customElementCols, row); f.onExec("UPDATE", 0) }, ErrCustomElementUnauthorized},
		{func(f *fakeDB) {
			f.on(customElementSelect, customElementCols, row).once()
			f.onExec("UPDATE", 1)
			f.onErr(customElementSelect, errCustomElementBoom)
		}, errCustomElementBoom},
	} {
		db, f := newFakeDB(t)
		tc.setup(f)
		_, err := customElementNew(db).Update(ctx, "e1", "u1", map[string]any{"Name": "x"})
		assertCustomElementErrIs(t, err, tc.want)
	}
}

func TestCustomElementDelete(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("DELETE FROM", 1)
	assertCustomElementNoErr(t, customElementNew(db).Delete(ctx, "e1", "u1"))
	if !f.sawQuery("DELETE FROM", "CustomElement", `"UserId" =`) {
		t.Errorf("delete must be scoped to owner: %q", f.lastQuery())
	}

	db2, _ := newFakeDB(t)
	assertCustomElementErrIs(t, customElementNew(db2).Delete(ctx, "e1", "u1"), ErrCustomElementUnauthorized)

	db3, f3 := newFakeDB(t)
	f3.onErr("DELETE FROM", errCustomElementBoom)
	assertCustomElementErrIs(t, customElementNew(db3).Delete(ctx, "e1", "u1"), errCustomElementBoom)
}

func TestCustomElementGetPublic(t *testing.T) {
	ctx := context.Background()
	cat, empty := "layout", ""
	for _, tc := range []struct {
		category   *string
		limit      int
		wantLimit  int
		wantFilter bool
	}{
		{&cat, 7, 7, true},
		{&empty, 0, DefaultPageSize, false},
		{nil, MaxPageSize * 10, MaxPageSize, false},
	} {
		db, f := newFakeDB(t)
		f.on(customElementSelect, customElementCols, []driver.Value{"e1", "u1", "Card", "layout", true, "1.0.0", nil, nil})
		got, err := customElementNew(db).ListPublic(ctx, tc.category, tc.limit, 40)
		assertCustomElementNoErr(t, err)
		assertPublicCustomElementsPage(t, f, got, tc.wantFilter, tc.wantLimit)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errCustomElementBoom)
	_, err := customElementNew(bad).ListPublic(ctx, nil, 1, 0)
	assertCustomElementErrIs(t, err, errCustomElementBoom)
}

func assertPublicCustomElementsPage(t *testing.T, f *fakeDB, got []models.CustomElement, wantFilter bool, wantLimit int) {
	t.Helper()
	if len(got) != 1 || f.sawQuery(`"Category" =`) != wantFilter {
		t.Fatalf("got=%+v filter=%v", got, f.queries[0])
	}
	if !customElementHasArg(f, int64(wantLimit)) && !customElementHasArg(f, wantLimit) {
		t.Errorf("limit %d not bound: %v", wantLimit, f.args[0])
	}
	if !customElementHasArg(f, int64(40)) && !customElementHasArg(f, 40) {
		t.Errorf("offset not bound: %v", f.args[0])
	}
}

func TestCustomElementDuplicate(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(customElementSelect, customElementCols, []driver.Value{"e1", "owner", "Card", "layout", true, "2.1.0", "desc", "a,b"})
	f.onExec("INSERT INTO", 1)
	got, err := customElementNew(db).Duplicate(ctx, "e1", "u1", "Card copy")
	assertCustomElementNoErr(t, err)
	assertDuplicatedCustomElement(t, f, got)

	dbMiss, _ := newFakeDB(t)
	_, err = customElementNew(dbMiss).Duplicate(ctx, "e1", "u1", "x")
	assertCustomElementErrIs(t, err, ErrCustomElementNotFound)

	dbSel, fSel := newFakeDB(t)
	fSel.onErr("SELECT", errCustomElementBoom)
	_, err = customElementNew(dbSel).Duplicate(ctx, "e1", "u1", "x")
	assertCustomElementErrIs(t, err, errCustomElementBoom)

	dbIns, fIns := newFakeDB(t)
	fIns.on(customElementSelect, customElementCols, []driver.Value{"e1", "owner", "Card", nil, true, "1.0.0", nil, nil})
	fIns.onErr("INSERT INTO", errCustomElementBoom)
	_, err = customElementNew(dbIns).Duplicate(ctx, "e1", "u1", "x")
	assertCustomElementErrIs(t, err, errCustomElementBoom)
}

func assertDuplicatedCustomElement(t *testing.T, f *fakeDB, got *models.CustomElement) {
	t.Helper()
	if got.Name != "Card copy" || got.UserId != "u1" || got.IsPublic || got.Version != "2.1.0" {
		t.Errorf("unexpected copy: %+v", got)
	}
	if got.Category == nil || *got.Category != "layout" || got.Description == nil || *got.Description != "desc" {
		t.Errorf("metadata mismatch: %+v", got)
	}
	if got.Tags == nil || *got.Tags != "a,b" || !f.sawQuery("INSERT INTO", "CustomElement") {
		t.Errorf("tags/insert mismatch: %+v queries=%v", got, f.queries)
	}
}
