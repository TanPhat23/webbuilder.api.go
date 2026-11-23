package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"my-go-app/internal/models"
)

var errCustomElementTypeBoom = errors.New("custom element type boom")

const customElementTypeSelect = `SELECT * FROM "public"."CustomElementType"`

var customElementTypeCols = []string{"Id", "Name", "Category"}

func TestCustomElementTypeValidationErrors(t *testing.T) {
	ctx := context.Background()
	cases := map[string]func(CustomElementTypeRepositoryInterface) error{
		"id is required": func(r CustomElementTypeRepositoryInterface) error {
			_, err := r.GetCustomElementTypeByID(ctx, "")
			return err
		},
		"name is required": func(r CustomElementTypeRepositoryInterface) error {
			_, err := r.GetCustomElementTypeByName(ctx, "")
			return err
		},
		"customElementType is required": func(r CustomElementTypeRepositoryInterface) error {
			_, err := r.CreateCustomElementType(ctx, nil)
			return err
		},
		"name is required (create)": func(r CustomElementTypeRepositoryInterface) error {
			_, err := r.CreateCustomElementType(ctx, &models.CustomElementType{})
			return err
		},
		"id is required (update)": func(r CustomElementTypeRepositoryInterface) error {
			_, err := r.UpdateCustomElementType(ctx, "", map[string]any{"Name": "x"})
			return err
		},
		"id is required (delete)": func(r CustomElementTypeRepositoryInterface) error { return r.DeleteCustomElementType(ctx, "") },
	}
	for want, call := range cases {
		db, f := newFakeDB(t)
		err := call(NewCustomElementTypeRepository(db))
		if err == nil || !strings.HasPrefix(want, err.Error()) {
			t.Errorf("want error %q, got %v", want, err)
		}
		if f.queryCount() != 0 {
			t.Errorf("%s: validation must not touch the database", want)
		}
	}
}

func TestCustomElementTypeList(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(customElementTypeSelect, customElementTypeCols, []driver.Value{"t1", "Card", "layout"}, []driver.Value{"t2", "Hero", nil})
	got, err := NewCustomElementTypeRepository(db).GetCustomElementTypes(ctx)
	if err != nil || len(got) != 2 || got[0].Name != "Card" || got[0].Category == nil || *got[0].Category != "layout" || got[1].Category != nil {
		t.Fatalf("got %+v, %v", got, err)
	}
	if !f.sawQuery(`ORDER BY "Name" ASC`) {
		t.Errorf("query: %q", f.lastQuery())
	}
	db2, f2 := newFakeDB(t)
	f2.onErr("SELECT", errCustomElementTypeBoom)
	if _, err := NewCustomElementTypeRepository(db2).GetCustomElementTypes(ctx); !errors.Is(err, errCustomElementTypeBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestCustomElementTypeGetByIDAndName(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(customElementTypeSelect, customElementTypeCols, []driver.Value{"t1", "Card", nil})
	repo := NewCustomElementTypeRepository(db)
	if got, err := repo.GetCustomElementTypeByID(ctx, "t1"); err != nil || got.Id != "t1" {
		t.Errorf("by id: %+v, %v", got, err)
	}
	if got, err := repo.GetCustomElementTypeByName(ctx, "Card"); err != nil || got.Name != "Card" {
		t.Errorf("by name: %+v, %v", got, err)
	}
	empty, _ := newFakeDB(t)
	er := NewCustomElementTypeRepository(empty)
	if _, err := er.GetCustomElementTypeByID(ctx, "x"); !errors.Is(err, ErrCustomElementTypeNotFound) {
		t.Errorf("by id missing: %v", err)
	}
	if _, err := er.GetCustomElementTypeByName(ctx, "x"); !errors.Is(err, ErrCustomElementTypeNotFound) {
		t.Errorf("by name missing: %v", err)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errCustomElementTypeBoom)
	br := NewCustomElementTypeRepository(bad)
	if _, err := br.GetCustomElementTypeByID(ctx, "x"); !errors.Is(err, errCustomElementTypeBoom) {
		t.Errorf("by id error: %v", err)
	}
	if _, err := br.GetCustomElementTypeByName(ctx, "x"); !errors.Is(err, errCustomElementTypeBoom) {
		t.Errorf("by name error: %v", err)
	}
}

func TestCustomElementTypeCreate(t *testing.T) {
	ctx := context.Background()
	count := func(f *fakeDB, n int64) { f.on("count(*)", []string{"count"}, []driver.Value{n}) }
	t.Run("ok", func(t *testing.T) {
		db, f := newFakeDB(t)
		count(f, 0)
		f.onExec("INSERT INTO", 1)
		in := &models.CustomElementType{Id: "t1", Name: "Card"}
		got, err := NewCustomElementTypeRepository(db).CreateCustomElementType(ctx, in)
		if err != nil || got != in || !f.sawQuery("INSERT INTO", "CustomElementType") {
			t.Fatalf("got %+v, %v, queries %v", got, err, f.queries)
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		db, f := newFakeDB(t)
		count(f, 1)
		if _, err := NewCustomElementTypeRepository(db).CreateCustomElementType(ctx, &models.CustomElementType{Name: "Card"}); !errors.Is(err, ErrCustomElementTypeAlreadyExists) {
			t.Errorf("err = %v", err)
		}
		if f.sawQuery("INSERT INTO") {
			t.Error("duplicate inserted")
		}
	})
	t.Run("count fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("count(*)", errCustomElementTypeBoom)
		if _, err := NewCustomElementTypeRepository(db).CreateCustomElementType(ctx, &models.CustomElementType{Name: "Card"}); !errors.Is(err, errCustomElementTypeBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("insert fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		count(f, 0)
		f.onErr("INSERT INTO", errCustomElementTypeBoom)
		if _, err := NewCustomElementTypeRepository(db).CreateCustomElementType(ctx, &models.CustomElementType{Name: "Card"}); !errors.Is(err, errCustomElementTypeBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestCustomElementTypeUpdate(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"t1", "Renamed", nil}
	t.Run("ok", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(customElementTypeSelect, customElementTypeCols, row)
		f.onExec("UPDATE", 1)
		got, err := NewCustomElementTypeRepository(db).UpdateCustomElementType(ctx, "t1", map[string]any{"Name": "Renamed"})
		if err != nil || got.Name != "Renamed" || !f.sawQuery("UPDATE", "CustomElementType") {
			t.Fatalf("got %+v, %v, queries %v", got, err, f.queries)
		}
	})
	t.Run("not found", func(t *testing.T) {
		db, _ := newFakeDB(t)
		if _, err := NewCustomElementTypeRepository(db).UpdateCustomElementType(ctx, "t1", map[string]any{"Name": "x"}); !errors.Is(err, ErrCustomElementTypeNotFound) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("lookup fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", errCustomElementTypeBoom)
		if _, err := NewCustomElementTypeRepository(db).UpdateCustomElementType(ctx, "t1", map[string]any{"Name": "x"}); !errors.Is(err, errCustomElementTypeBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("update fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(customElementTypeSelect, customElementTypeCols, row)
		f.onErr("UPDATE", errCustomElementTypeBoom)
		if _, err := NewCustomElementTypeRepository(db).UpdateCustomElementType(ctx, "t1", map[string]any{"Name": "x"}); !errors.Is(err, errCustomElementTypeBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("refetch fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(customElementTypeSelect, customElementTypeCols, row).once()
		f.onExec("UPDATE", 1)
		f.onErr(customElementTypeSelect, errCustomElementTypeBoom)
		if _, err := NewCustomElementTypeRepository(db).UpdateCustomElementType(ctx, "t1", map[string]any{"Name": "x"}); !errors.Is(err, errCustomElementTypeBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestCustomElementTypeDelete(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("DELETE FROM", 1)
	if err := NewCustomElementTypeRepository(db).DeleteCustomElementType(ctx, "t1"); err != nil {
		t.Fatal(err)
	}
	db2, _ := newFakeDB(t)
	if err := NewCustomElementTypeRepository(db2).DeleteCustomElementType(ctx, "t1"); !errors.Is(err, ErrCustomElementTypeNotFound) {
		t.Errorf("err = %v", err)
	}
	db3, f3 := newFakeDB(t)
	f3.onErr("DELETE FROM", errCustomElementTypeBoom)
	if err := NewCustomElementTypeRepository(db3).DeleteCustomElementType(ctx, "t1"); !errors.Is(err, errCustomElementTypeBoom) {
		t.Errorf("err = %v", err)
	}
}
