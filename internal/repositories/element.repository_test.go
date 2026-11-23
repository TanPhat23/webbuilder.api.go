package repositories

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/pkg/utils"
)

var errElementBoom = errors.New("element boom")

const elementSelect = `FROM "public"."Element"`

var elementCols = []string{"Id", "Type", "PageId", "ParentId", "Order"}

func elementNew(db *gorm.DB) *ElementRepository { return &ElementRepository{db: db} }

func elementArg(f *fakeDB, want any) bool {
	for _, q := range f.args {
		for _, a := range q {
			if a.Value == want {
				return true
			}
		}
	}
	return false
}

func elementStrPtr(s string) *string { return &s }

func TestElementValidationErrors(t *testing.T) {
	ctx := context.Background()
	cases := map[string]func(*ElementRepository) error{
		"projectID is required": func(r *ElementRepository) error { _, err := r.GetElements(ctx, ""); return err },
		"elementID is required": func(r *ElementRepository) error { _, err := r.GetElementByID(ctx, ""); return err },
		"elementID is required (relations)": func(r *ElementRepository) error {
			_, err := r.GetElementWithRelations(ctx, "")
			return err
		},
		"pageID is required": func(r *ElementRepository) error { _, err := r.GetElementsByPageID(ctx, ""); return err },
		"projectID is required (page ids)": func(r *ElementRepository) error {
			_, err := r.GetElementsByPageIds(ctx, "", []string{"p"})
			return err
		},
		"pageIDs is required":  func(r *ElementRepository) error { _, err := r.GetElementsByPageIds(ctx, "p", nil); return err },
		"parentID is required": func(r *ElementRepository) error { _, err := r.GetChildElements(ctx, ""); return err },
		"projectID is required (roots)": func(r *ElementRepository) error {
			_, err := r.GetRootElements(ctx, "")
			return err
		},
		"element cannot be nil":          func(r *ElementRepository) error { return r.CreateElement(ctx, nil) },
		"type is required":               func(r *ElementRepository) error { return r.CreateElement(ctx, &models.Element{}) },
		"element cannot be nil (update)": func(r *ElementRepository) error { return r.UpdateElement(ctx, nil) },
		"element ID is required":         func(r *ElementRepository) error { return r.UpdateElement(ctx, &models.Element{}) },
		"elementID is required (workflows)": func(r *ElementRepository) error {
			return r.UpdateEventWorkflows(ctx, "", []byte("[]"))
		},
		"elementID is required (delete)":  func(r *ElementRepository) error { return r.DeleteElementByID(ctx, "") },
		"pageID is required (delete)":     func(r *ElementRepository) error { return r.DeleteElementsByPageID(ctx, "") },
		"projectID is required (delete)":  func(r *ElementRepository) error { return r.DeleteElementsByProjectID(ctx, "") },
		"projectID is required (count)":   func(r *ElementRepository) error { _, err := r.CountElementsByProjectID(ctx, ""); return err },
		"projectID is required (replace)": func(r *ElementRepository) error { return r.ReplaceElements(ctx, "", nil) },
	}
	for want, call := range cases {
		db, f := newFakeDB(t)
		err := call(elementNew(db))
		if err == nil || !strings.HasPrefix(want, err.Error()) {
			t.Errorf("want error %q, got %v", want, err)
		}
		if f.queryCount() != 0 || f.begins != 0 {
			t.Errorf("%s: validation must not touch the database", want)
		}
	}
}

func TestElementProjectMutexIsPerProject(t *testing.T) {
	r := elementNew(nil)
	if r.getProjectMutex("a") != r.getProjectMutex("a") {
		t.Error("same project must share a mutex")
	}
	if r.getProjectMutex("a") == r.getProjectMutex("b") {
		t.Error("different projects must not share a mutex")
	}
}

func TestElementGetElementsBuildsTree(t *testing.T) {
	ctx := context.Background()
	rows := [][]driver.Value{
		{"root", "Frame", "pg1", nil, int64(1)},
		{"child", "Text", "pg1", "root", int64(2)},
	}
	t.Run("whole project", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(elementSelect, elementCols, rows...)
		got, err := elementNew(db).GetElements(ctx, "proj")
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0].GetElement().Id != "root" {
			t.Fatalf("want a single root, got %+v", got)
		}
		kids := utils.GetChildrenFromEditorElement(got[0])
		if len(kids) != 1 {
			t.Fatalf("root should carry its child, got %v", kids)
		}
		if f.sawQuery(`"Element"."PageId" = $2`) || !f.sawQuery(`"Page"."ProjectId" =`, `"DeletedAt" IS NULL`) || !elementArg(f, "proj") {
			t.Errorf("query %q", f.queries[0])
		}
	})
	t.Run("page filter", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(elementSelect, elementCols, rows[0])
		if _, err := elementNew(db).GetElements(ctx, "proj", "pg1"); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery(`"Element"."PageId" = $2`) || !elementArg(f, "pg1") {
			t.Errorf("page filter missing: %q", f.queries[0])
		}
	})
	t.Run("empty page filter ignored", func(t *testing.T) {
		db, f := newFakeDB(t)
		if _, err := elementNew(db).GetElements(ctx, "proj", ""); err != nil {
			t.Fatal(err)
		}
		if f.sawQuery(`"Element"."PageId" = $2`) {
			t.Error("empty page id must not filter")
		}
	})
	t.Run("no elements gives empty non-nil slice", func(t *testing.T) {
		db, _ := newFakeDB(t)
		got, err := elementNew(db).GetElements(ctx, "proj")
		if err != nil || got == nil || len(got) != 0 {
			t.Errorf("got %#v, %v", got, err)
		}
	})
	t.Run("error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", errElementBoom)
		if _, err := elementNew(db).GetElements(ctx, "proj"); !errors.Is(err, errElementBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestElementGetByIDAndRelations(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(elementSelect, elementCols, []driver.Value{"e1", "Text", "pg1", nil, int64(1)})
	repo := elementNew(db)
	if got, err := repo.GetElementByID(ctx, "e1"); err != nil || got.Id != "e1" || got.PageId == nil || *got.PageId != "pg1" {
		t.Errorf("GetElementByID: %+v, %v", got, err)
	}
	if got, err := repo.GetElementWithRelations(ctx, "e1"); err != nil || got.Id != "e1" {
		t.Errorf("GetElementWithRelations: %+v, %v", got, err)
	}
	empty, _ := newFakeDB(t)
	er := elementNew(empty)
	if _, err := er.GetElementByID(ctx, "x"); !errors.Is(err, ErrElementNotFound) {
		t.Errorf("missing: %v", err)
	}
	if _, err := er.GetElementWithRelations(ctx, "x"); !errors.Is(err, ErrElementNotFound) {
		t.Errorf("missing relations: %v", err)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errElementBoom)
	br := elementNew(bad)
	if _, err := br.GetElementByID(ctx, "x"); !errors.Is(err, errElementBoom) || errors.Is(err, ErrElementNotFound) {
		t.Errorf("error: %v", err)
	}
	if _, err := br.GetElementWithRelations(ctx, "x"); !errors.Is(err, errElementBoom) || errors.Is(err, ErrElementNotFound) {
		t.Errorf("relations error: %v", err)
	}
}

func TestElementListQueries(t *testing.T) {
	ctx := context.Background()
	two := [][]driver.Value{
		{"e1", "Text", "pg1", nil, int64(1)},
		{"e2", "Text", "pg1", nil, int64(2)},
	}
	cases := []struct {
		name     string
		call     func(*ElementRepository) ([]models.Element, error)
		fragment string
		arg      any
	}{
		{"by page", func(r *ElementRepository) ([]models.Element, error) { return r.GetElementsByPageID(ctx, "pg1") }, `"PageId" =`, "pg1"},
		{"children", func(r *ElementRepository) ([]models.Element, error) { return r.GetChildElements(ctx, "parent") }, `"ParentId" =`, "parent"},
		{"roots", func(r *ElementRepository) ([]models.Element, error) { return r.GetRootElements(ctx, "proj") }, `"ParentId" IS NULL`, "proj"},
		{"by ids", func(r *ElementRepository) ([]models.Element, error) {
			return r.GetElementsByIDs(ctx, []string{"e1", "e2"})
		}, `"Id" IN`, "e2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			f.on(elementSelect, elementCols, two...)
			got, err := tc.call(elementNew(db))
			if err != nil || len(got) != 2 || got[0].Id != "e1" || got[1].Id != "e2" {
				t.Fatalf("got %+v, %v", got, err)
			}
			if !f.sawQuery(tc.fragment) || !elementArg(f, tc.arg) {
				t.Errorf("query %q args %v", f.queries[0], f.args[0])
			}
			bad, fb := newFakeDB(t)
			fb.onErr("SELECT", errElementBoom)
			if got, err := tc.call(elementNew(bad)); got != nil || !errors.Is(err, errElementBoom) {
				t.Errorf("error case: %+v, %v", got, err)
			}
		})
	}
	t.Run("by ids with none requested", func(t *testing.T) {
		db, f := newFakeDB(t)
		got, err := elementNew(db).GetElementsByIDs(ctx, nil)
		if err != nil || got == nil || len(got) != 0 || f.queryCount() != 0 {
			t.Errorf("got %#v, %v, queries %d", got, err, f.queryCount())
		}
	})
}

func TestElementGetElementsByPageIds(t *testing.T) {
	ctx := context.Background()
	t.Run("scopes pages to the project and builds a tree", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(elementSelect, elementCols,
			[]driver.Value{"root", "Frame", "pg1", nil, int64(1)},
			[]driver.Value{"child", "Text", "pg1", "root", int64(2)})
		got, err := elementNew(db).GetElementsByPageIds(ctx, "proj", []string{"pg1", "pg2"})
		if err != nil || len(got) != 1 || got[0].GetElement().Id != "root" || len(utils.GetChildrenFromEditorElement(got[0])) != 1 {
			t.Fatalf("got %+v, %v", got, err)
		}
		if !f.sawQuery(`"ProjectId" =`, `"DeletedAt" IS NULL`) || !elementArg(f, "proj") || !elementArg(f, "pg2") {
			t.Errorf("query %q args %v", f.queries[0], f.args[0])
		}
	})
	t.Run("empty result", func(t *testing.T) {
		db, _ := newFakeDB(t)
		got, err := elementNew(db).GetElementsByPageIds(ctx, "proj", []string{"pg1"})
		if err != nil || got == nil || len(got) != 0 {
			t.Errorf("got %#v, %v", got, err)
		}
	})
	t.Run("error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", errElementBoom)
		if _, err := elementNew(db).GetElementsByPageIds(ctx, "proj", []string{"pg1"}); !errors.Is(err, errElementBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestElementCreate(t *testing.T) {
	ctx := context.Background()
	t.Run("generates v7 id", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("INSERT INTO", 1)
		e := &models.Element{Type: "Text"}
		if err := elementNew(db).CreateElement(ctx, e); err != nil {
			t.Fatal(err)
		}
		if e.Id == "" || !f.sawQuery("INSERT INTO", "Element") || !elementArg(f, e.Id) {
			t.Errorf("id=%q queries=%v", e.Id, f.queries)
		}
	})
	t.Run("keeps supplied id", func(t *testing.T) {
		db, _ := newFakeDB(t)
		e := &models.Element{Id: "mine", Type: "Text"}
		if err := elementNew(db).CreateElement(ctx, e); err != nil || e.Id != "mine" {
			t.Errorf("id=%q err=%v", e.Id, err)
		}
	})
	t.Run("insert fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("INSERT INTO", errElementBoom)
		if err := elementNew(db).CreateElement(ctx, &models.Element{Type: "Text"}); !errors.Is(err, errElementBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestElementUpdateAndWorkflows(t *testing.T) {
	ctx := context.Background()
	t.Run("update ok", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("UPDATE", 1)
		if err := elementNew(db).UpdateElement(ctx, &models.Element{Id: "e1", Type: "Text", Content: elementStrPtr("hello")}); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("UPDATE", "Element", `"Id" =`) || !elementArg(f, "hello") {
			t.Errorf("query %q", f.lastQuery())
		}
	})
	t.Run("update not found", func(t *testing.T) {
		db, _ := newFakeDB(t)
		if err := elementNew(db).UpdateElement(ctx, &models.Element{Id: "e1", Type: "Text"}); !errors.Is(err, ErrElementNotFound) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("update fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("UPDATE", errElementBoom)
		if err := elementNew(db).UpdateElement(ctx, &models.Element{Id: "e1", Type: "Text"}); !errors.Is(err, errElementBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("workflows ok", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("UPDATE", 1)
		if err := elementNew(db).UpdateEventWorkflows(ctx, "e1", []byte(`[{"a":1}]`)); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("UPDATE", `"EventWorkflows"`) {
			t.Errorf("query %q", f.lastQuery())
		}
	})
	t.Run("workflows not found", func(t *testing.T) {
		db, _ := newFakeDB(t)
		if err := elementNew(db).UpdateEventWorkflows(ctx, "e1", nil); !errors.Is(err, ErrElementNotFound) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("workflows fail", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("UPDATE", errElementBoom)
		if err := elementNew(db).UpdateEventWorkflows(ctx, "e1", nil); !errors.Is(err, errElementBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestElementDeletes(t *testing.T) {
	ctx := context.Background()
	t.Run("by id", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("DELETE FROM", 1)
		if err := elementNew(db).DeleteElementByID(ctx, "e1"); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("DELETE FROM", `"Id" =`) {
			t.Errorf("query %q", f.lastQuery())
		}
		empty, _ := newFakeDB(t)
		if err := elementNew(empty).DeleteElementByID(ctx, "e1"); !errors.Is(err, ErrElementNotFound) {
			t.Errorf("missing: %v", err)
		}
		bad, fb := newFakeDB(t)
		fb.onErr("DELETE FROM", errElementBoom)
		if err := elementNew(bad).DeleteElementByID(ctx, "e1"); !errors.Is(err, errElementBoom) {
			t.Errorf("error: %v", err)
		}
	})
	t.Run("by page tolerates zero rows", func(t *testing.T) {
		db, f := newFakeDB(t)
		if err := elementNew(db).DeleteElementsByPageID(ctx, "pg1"); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("DELETE FROM", `"PageId" =`) {
			t.Errorf("query %q", f.lastQuery())
		}
		bad, fb := newFakeDB(t)
		fb.onErr("DELETE FROM", errElementBoom)
		if err := elementNew(bad).DeleteElementsByPageID(ctx, "pg1"); !errors.Is(err, errElementBoom) {
			t.Errorf("error: %v", err)
		}
	})
	t.Run("by project", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("DELETE FROM", 5)
		if err := elementNew(db).DeleteElementsByProjectID(ctx, "proj"); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("DELETE FROM", "Element") || !elementArg(f, "proj") {
			t.Errorf("query %q", f.lastQuery())
		}
		bad, fb := newFakeDB(t)
		fb.onErr("DELETE", errElementBoom)
		if err := elementNew(bad).DeleteElementsByProjectID(ctx, "proj"); !errors.Is(err, errElementBoom) {
			t.Errorf("error: %v", err)
		}
	})
}

func TestElementCountByProject(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("count(*)", []string{"count"}, []driver.Value{int64(12)})
	n, err := elementNew(db).CountElementsByProjectID(ctx, "proj")
	if err != nil || n != 12 {
		t.Errorf("count = %d, %v", n, err)
	}
	if !elementArg(f, "proj") {
		t.Errorf("project not bound: %v", f.args[0])
	}
	bad, fb := newFakeDB(t)
	fb.onErr("count(*)", errElementBoom)
	if n, err := elementNew(bad).CountElementsByProjectID(ctx, "proj"); n != 0 || !errors.Is(err, errElementBoom) {
		t.Errorf("error: %d, %v", n, err)
	}
}

func TestElementReplaceElements(t *testing.T) {
	ctx := context.Background()
	page := elementStrPtr("pg1")
	t.Run("deletes then inserts the flattened tree in one transaction", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("DELETE FROM", 2)
		f.onExec("INSERT INTO", 2)
		roots := []models.EditorElement{
			&models.Element{Id: "a", Type: "Text", PageId: page},
			&models.Element{Id: "b", Type: "Text", PageId: page},
		}
		if err := elementNew(db).ReplaceElements(ctx, "proj", roots); err != nil {
			t.Fatal(err)
		}
		if f.begins != 1 || f.commits != 1 || f.aborts != 0 {
			t.Errorf("tx begins=%d commits=%d aborts=%d", f.begins, f.commits, f.aborts)
		}
		del, ins := -1, -1
		for i, q := range f.queries {
			if strings.HasPrefix(q, "DELETE FROM") && del < 0 {
				del = i
			}
			if strings.HasPrefix(q, "INSERT INTO") && ins < 0 {
				ins = i
			}
		}
		if del < 0 || ins < 0 || del > ins {
			t.Fatalf("delete must precede insert: %v", f.queries)
		}
		if !elementArg(f, "proj") {
			t.Errorf("project id not bound on delete: %v", f.args[del])
		}
		inserted := map[any]bool{}
		for _, a := range f.args[ins] {
			inserted[a.Value] = true
		}
		if !inserted["a"] || !inserted["b"] {
			t.Errorf("both elements should be inserted: %v", f.args[ins])
		}
		if strings.Contains(f.queries[ins], "EventWorkflows") {
			t.Errorf("EventWorkflows must be omitted from the insert: %q", f.queries[ins])
		}
	})
	t.Run("empty replacement only deletes", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("DELETE FROM", 3)
		if err := elementNew(db).ReplaceElements(ctx, "proj", nil); err != nil {
			t.Fatal(err)
		}
		if f.sawQuery("INSERT INTO") || f.commits != 1 {
			t.Errorf("queries=%v commits=%d", f.queries, f.commits)
		}
	})
	t.Run("delete fails and rolls back", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("DELETE FROM", errElementBoom)
		err := elementNew(db).ReplaceElements(ctx, "proj", []models.EditorElement{&models.Element{Id: "a", Type: "Text"}})
		if !errors.Is(err, errElementBoom) || f.commits != 0 || f.aborts != 1 || f.sawQuery("INSERT INTO") {
			t.Errorf("err=%v commits=%d aborts=%d queries=%v", err, f.commits, f.aborts, f.queries)
		}
	})
	t.Run("insert fails and rolls back", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("INSERT INTO", errElementBoom)
		err := elementNew(db).ReplaceElements(ctx, "proj", []models.EditorElement{&models.Element{Id: "a", Type: "Text"}})
		if !errors.Is(err, errElementBoom) || f.commits != 0 || f.aborts != 1 {
			t.Errorf("err=%v commits=%d aborts=%d", err, f.commits, f.aborts)
		}
	})
	t.Run("unconvertible child aborts before inserting", func(t *testing.T) {
		db, f := newFakeDB(t)
		bad := &utils.ElementWithChildren{Element: &models.Element{Id: "root", Type: "Frame"}, Elements: []any{make(chan int)}}
		err := elementNew(db).ReplaceElements(ctx, "proj", []models.EditorElement{bad})
		if err == nil || !strings.Contains(err.Error(), "failed to convert child element") {
			t.Errorf("err = %v", err)
		}
		if f.sawQuery("INSERT INTO") || f.commits != 0 {
			t.Errorf("nothing may be inserted: %v", f.queries)
		}
	})
}

func TestElementFlattenSkipsNilAndNormalisesParents(t *testing.T) {
	var typedNil *models.Element
	empty := ""
	roots := []models.EditorElement{
		nil,
		typedNil,
		&models.Element{Id: "a", Type: "Text", ParentId: &empty},
	}
	got, err := elementNew(nil).flattenElementsForInsert(context.Background(), nil, roots, "proj")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Id != "a" || got[0].ParentId != nil {
		t.Fatalf("got %+v", got)
	}
	if got[0].Order != 1 || got[0].OrderKey == nil {
		t.Errorf("order not assigned: %+v", got[0])
	}
}

func TestElementFlattenReportsChildConversionError(t *testing.T) {
	bad := &utils.ElementWithChildren{Element: &models.Element{Id: "root", Type: "Frame"}, Elements: []any{make(chan int)}}
	_, err := elementNew(nil).flattenElementsForInsert(context.Background(), nil, []models.EditorElement{bad}, "proj")
	if err == nil || !strings.Contains(err.Error(), "failed to convert child element") {
		t.Errorf("err = %v", err)
	}
}
