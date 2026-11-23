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
		"userID is required": func(r *CustomElementRepository) error { _, err := r.GetCustomElements(ctx, "", nil); return err },
		"id is required":     func(r *CustomElementRepository) error { _, err := r.GetCustomElementByID(ctx, "", "u"); return err },
		"userID is required (get)": func(r *CustomElementRepository) error {
			_, err := r.GetCustomElementByID(ctx, "e", "")
			return err
		},
		"customElement is required": func(r *CustomElementRepository) error { _, err := r.CreateCustomElement(ctx, nil); return err },
		"userId is required": func(r *CustomElementRepository) error {
			_, err := r.CreateCustomElement(ctx, &models.CustomElement{Name: "n"})
			return err
		},
		"name is required": func(r *CustomElementRepository) error {
			_, err := r.CreateCustomElement(ctx, &models.CustomElement{UserId: "u"})
			return err
		},
		"id is required (update)": func(r *CustomElementRepository) error {
			_, err := r.UpdateCustomElement(ctx, "", "u", nil)
			return err
		},
		"userID is required (update)": func(r *CustomElementRepository) error {
			_, err := r.UpdateCustomElement(ctx, "e", "", nil)
			return err
		},
		"id is required (delete)":     func(r *CustomElementRepository) error { return r.DeleteCustomElement(ctx, "", "u") },
		"userID is required (delete)": func(r *CustomElementRepository) error { return r.DeleteCustomElement(ctx, "e", "") },
		"id is required (duplicate)": func(r *CustomElementRepository) error {
			_, err := r.DuplicateCustomElement(ctx, "", "u", "n")
			return err
		},
		"userID is required (duplicate)": func(r *CustomElementRepository) error {
			_, err := r.DuplicateCustomElement(ctx, "e", "", "n")
			return err
		},
		"newName is required": func(r *CustomElementRepository) error {
			_, err := r.DuplicateCustomElement(ctx, "e", "u", "")
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
	t.Run("all of the user's elements", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(customElementSelect, customElementCols, []driver.Value{"e1", "u1", "Card", nil, true, "1.0.0", nil, nil})
		got, err := customElementNew(db).GetCustomElements(ctx, "u1", nil)
		if err != nil || len(got) != 1 || got[0].Name != "Card" || !got[0].IsPublic {
			t.Fatalf("got %+v, %v", got, err)
		}
		if f.sawQuery(`"IsPublic" =`) || !f.sawQuery(`"UserId" =`, `ORDER BY "CreatedAt" DESC`) {
			t.Errorf("query: %q", f.queries[0])
		}
	})
	t.Run("public filter", func(t *testing.T) {
		db, f := newFakeDB(t)
		pub := false
		if _, err := customElementNew(db).GetCustomElements(ctx, "u1", &pub); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery(`"IsPublic" =`) || !customElementHasArg(f, false) {
			t.Errorf("filter missing: %q %v", f.queries[0], f.args[0])
		}
	})
	t.Run("error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", errCustomElementBoom)
		if _, err := customElementNew(db).GetCustomElements(ctx, "u1", nil); !errors.Is(err, errCustomElementBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestCustomElementGetByID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(customElementSelect, customElementCols, []driver.Value{"e1", "other", "Card", nil, true, "1.0.0", nil, nil})
	got, err := customElementNew(db).GetCustomElementByID(ctx, "e1", "u1")
	if err != nil || got.Id != "e1" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if !f.sawQuery(`"UserId" = $2 OR "IsPublic" = $3`) {
		t.Errorf("owner-or-public filter missing: %q", f.queries[0])
	}
	empty, _ := newFakeDB(t)
	if _, err := customElementNew(empty).GetCustomElementByID(ctx, "e1", "u1"); !errors.Is(err, ErrCustomElementNotFound) {
		t.Errorf("missing: %v", err)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errCustomElementBoom)
	if _, err := customElementNew(bad).GetCustomElementByID(ctx, "e1", "u1"); !errors.Is(err, errCustomElementBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestCustomElementCreate(t *testing.T) {
	ctx := context.Background()
	in := func() *models.CustomElement {
		return &models.CustomElement{Id: "e1", UserId: "u1", Name: "Card", Structure: []byte("{}")}
	}
	db, f := newFakeDB(t)
	f.onExec("INSERT INTO", 1)
	got, err := customElementNew(db).CreateCustomElement(ctx, in())
	if err != nil || got.Id != "e1" || !f.sawQuery("INSERT INTO", "CustomElement") {
		t.Fatalf("got %+v, %v", got, err)
	}

	dup, fd := newFakeDB(t)
	fd.onErr("INSERT INTO", customElementNameConflict())
	if _, err := customElementNew(dup).CreateCustomElement(ctx, in()); !errors.Is(err, ErrCustomElementAlreadyExists) {
		t.Errorf("conflict: %v", err)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("INSERT INTO", errCustomElementBoom)
	if _, err := customElementNew(bad).CreateCustomElement(ctx, in()); !errors.Is(err, errCustomElementBoom) || errors.Is(err, ErrCustomElementAlreadyExists) {
		t.Errorf("other error: %v", err)
	}
}

func TestCustomElementUpdate(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"e1", "u1", "Renamed", nil, false, "1.0.0", nil, nil}
	t.Run("ok", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(customElementSelect, customElementCols, row)
		f.onExec("UPDATE", 1)
		got, err := customElementNew(db).UpdateCustomElement(ctx, "e1", "u1", map[string]any{"Name": "Renamed"})
		if err != nil || got.Name != "Renamed" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if !f.sawQuery("UPDATE", "CustomElement", `"UserId" =`) {
			t.Errorf("update must be scoped to the owner: %v", f.queries)
		}
	})
	t.Run("not the owner", func(t *testing.T) {
		db, _ := newFakeDB(t)
		if _, err := customElementNew(db).UpdateCustomElement(ctx, "e1", "u1", map[string]any{"Name": "x"}); !errors.Is(err, ErrCustomElementUnauthorized) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("lookup fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", errCustomElementBoom)
		if _, err := customElementNew(db).UpdateCustomElement(ctx, "e1", "u1", map[string]any{"Name": "x"}); !errors.Is(err, errCustomElementBoom) || errors.Is(err, ErrCustomElementUnauthorized) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("name conflict", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(customElementSelect, customElementCols, row)
		f.onErr("UPDATE", customElementNameConflict())
		if _, err := customElementNew(db).UpdateCustomElement(ctx, "e1", "u1", map[string]any{"Name": "x"}); !errors.Is(err, ErrCustomElementAlreadyExists) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("update fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(customElementSelect, customElementCols, row)
		f.onErr("UPDATE", errCustomElementBoom)
		if _, err := customElementNew(db).UpdateCustomElement(ctx, "e1", "u1", map[string]any{"Name": "x"}); !errors.Is(err, errCustomElementBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("no rows affected", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(customElementSelect, customElementCols, row)
		f.onExec("UPDATE", 0)
		if _, err := customElementNew(db).UpdateCustomElement(ctx, "e1", "u1", map[string]any{"Name": "x"}); !errors.Is(err, ErrCustomElementUnauthorized) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("refetch fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(customElementSelect, customElementCols, row).once()
		f.onExec("UPDATE", 1)
		f.onErr(customElementSelect, errCustomElementBoom)
		if _, err := customElementNew(db).UpdateCustomElement(ctx, "e1", "u1", map[string]any{"Name": "x"}); !errors.Is(err, errCustomElementBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestCustomElementDelete(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("DELETE FROM", 1)
	if err := customElementNew(db).DeleteCustomElement(ctx, "e1", "u1"); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery("DELETE FROM", "CustomElement", `"UserId" =`) {
		t.Errorf("delete must be scoped to the owner: %q", f.lastQuery())
	}
	db2, _ := newFakeDB(t)
	if err := customElementNew(db2).DeleteCustomElement(ctx, "e1", "u1"); !errors.Is(err, ErrCustomElementUnauthorized) {
		t.Errorf("err = %v", err)
	}
	db3, f3 := newFakeDB(t)
	f3.onErr("DELETE FROM", errCustomElementBoom)
	if err := customElementNew(db3).DeleteCustomElement(ctx, "e1", "u1"); !errors.Is(err, errCustomElementBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestCustomElementGetPublic(t *testing.T) {
	ctx := context.Background()
	cat := "layout"
	empty := ""
	for _, tc := range []struct {
		name       string
		category   *string
		limit      int
		wantLimit  int
		wantFilter bool
	}{
		{"category and explicit limit", &cat, 7, 7, true},
		{"empty category ignored, default limit", &empty, 0, DefaultPageSize, false},
		{"nil category, limit capped", nil, MaxPageSize * 10, MaxPageSize, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			f.on(customElementSelect, customElementCols, []driver.Value{"e1", "u1", "Card", "layout", true, "1.0.0", nil, nil})
			got, err := customElementNew(db).GetPublicCustomElements(ctx, tc.category, tc.limit, 40)
			if err != nil || len(got) != 1 {
				t.Fatalf("got %+v, %v", got, err)
			}
			if f.sawQuery(`"Category" =`) != tc.wantFilter {
				t.Errorf("category filter = %v: %q", !tc.wantFilter, f.queries[0])
			}
			if !customElementHasArg(f, int64(tc.wantLimit)) && !customElementHasArg(f, tc.wantLimit) {
				t.Errorf("limit %d not bound: %v", tc.wantLimit, f.args[0])
			}
			if !customElementHasArg(f, int64(40)) && !customElementHasArg(f, 40) {
				t.Errorf("offset not bound: %v", f.args[0])
			}
		})
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errCustomElementBoom)
	if _, err := customElementNew(bad).GetPublicCustomElements(ctx, nil, 1, 0); !errors.Is(err, errCustomElementBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestCustomElementDuplicate(t *testing.T) {
	ctx := context.Background()
	t.Run("copies the original as a private element of the caller", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(customElementSelect, customElementCols, []driver.Value{"e1", "owner", "Card", "layout", true, "2.1.0", "desc", "a,b"})
		f.onExec("INSERT INTO", 1)
		got, err := customElementNew(db).DuplicateCustomElement(ctx, "e1", "u1", "Card copy")
		if err != nil {
			t.Fatal(err)
		}
		if got.Name != "Card copy" || got.UserId != "u1" || got.IsPublic || got.Version != "2.1.0" ||
			got.Category == nil || *got.Category != "layout" || got.Description == nil || *got.Description != "desc" || got.Tags == nil || *got.Tags != "a,b" {
			t.Errorf("unexpected copy: %+v", got)
		}
		if !f.sawQuery("INSERT INTO", "CustomElement") {
			t.Error("no insert")
		}
	})
	t.Run("original not visible", func(t *testing.T) {
		db, _ := newFakeDB(t)
		if _, err := customElementNew(db).DuplicateCustomElement(ctx, "e1", "u1", "x"); !errors.Is(err, ErrCustomElementNotFound) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("lookup fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", errCustomElementBoom)
		if _, err := customElementNew(db).DuplicateCustomElement(ctx, "e1", "u1", "x"); !errors.Is(err, errCustomElementBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("insert fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(customElementSelect, customElementCols, []driver.Value{"e1", "owner", "Card", nil, true, "1.0.0", nil, nil})
		f.onErr("INSERT INTO", errCustomElementBoom)
		if _, err := customElementNew(db).DuplicateCustomElement(ctx, "e1", "u1", "x"); !errors.Is(err, errCustomElementBoom) {
			t.Errorf("err = %v", err)
		}
	})
}
