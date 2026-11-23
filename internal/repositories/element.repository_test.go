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

func assertElementErrIs(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("want error %v, got %v", want, err)
	}
}

func assertElementNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

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

func TestElementGetElementsWholeProjectAndPageFilter(t *testing.T) {
	ctx := context.Background()
	rows := [][]driver.Value{
		{"root", "Frame", "pg1", nil, int64(1)},
		{"child", "Text", "pg1", "root", int64(2)},
	}
	db, f := newFakeDB(t)
	f.on(elementSelect, elementCols, rows...)
	got, err := elementNew(db).GetElements(ctx, "proj")
	assertElementNoErr(t, err)
	assertElementTreeRootWithChild(t, got, "root")
	if f.sawQuery(`"Element"."PageId" = $2`) || !f.sawQuery(`"Page"."ProjectId" =`, `"DeletedAt" IS NULL`) || !elementArg(f, "proj") {
		t.Errorf("query %q", f.queries[0])
	}

	dbPage, fPage := newFakeDB(t)
	fPage.on(elementSelect, elementCols, rows[0])
	_, err = elementNew(dbPage).GetElements(ctx, "proj", "pg1")
	assertElementNoErr(t, err)
	if !fPage.sawQuery(`"Element"."PageId" = $2`) || !elementArg(fPage, "pg1") {
		t.Errorf("page filter missing: %q", fPage.queries[0])
	}
}

func TestElementGetElementsEdgeCases(t *testing.T) {
	ctx := context.Background()
	dbEmptyPage, fEmptyPage := newFakeDB(t)
	_, err := elementNew(dbEmptyPage).GetElements(ctx, "proj", "")
	assertElementNoErr(t, err)
	if fEmptyPage.sawQuery(`"Element"."PageId" = $2`) {
		t.Error("empty page id must not filter")
	}

	dbNone, _ := newFakeDB(t)
	got, err := elementNew(dbNone).GetElements(ctx, "proj")
	assertElementNoErr(t, err)
	if got == nil || len(got) != 0 {
		t.Errorf("got %#v", got)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("SELECT", errElementBoom)
	_, err = elementNew(dbErr).GetElements(ctx, "proj")
	assertElementErrIs(t, err, errElementBoom)
}

func assertElementTreeRootWithChild(t *testing.T, got []models.EditorElement, rootID string) {
	t.Helper()
	if len(got) != 1 || got[0].GetElement().Id != rootID {
		t.Fatalf("want single root %q, got %+v", rootID, got)
	}
	if len(utils.GetChildrenFromEditorElement(got[0])) != 1 {
		t.Fatalf("root should carry one child")
	}
}

func TestElementGetByIDAndRelations(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(elementSelect, elementCols, []driver.Value{"e1", "Text", "pg1", nil, int64(1)})
	repo := elementNew(db)
	got, err := repo.GetElementByID(ctx, "e1")
	assertElementNoErr(t, err)
	if got.Id != "e1" || got.PageId == nil || *got.PageId != "pg1" {
		t.Errorf("GetElementByID: %+v", got)
	}
	rel, err := repo.GetElementWithRelations(ctx, "e1")
	assertElementNoErr(t, err)
	if rel.Id != "e1" {
		t.Errorf("GetElementWithRelations: %+v", rel)
	}
}

func TestElementGetByIDAndRelationsErrors(t *testing.T) {
	ctx := context.Background()
	empty, _ := newFakeDB(t)
	er := elementNew(empty)
	_, err := er.GetElementByID(ctx, "x")
	assertElementErrIs(t, err, ErrElementNotFound)
	_, err = er.GetElementWithRelations(ctx, "x")
	assertElementErrIs(t, err, ErrElementNotFound)

	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errElementBoom)
	br := elementNew(bad)
	_, err = br.GetElementByID(ctx, "x")
	assertElementErrIs(t, err, errElementBoom)
	if errors.Is(err, ErrElementNotFound) {
		t.Errorf("unexpected ErrElementNotFound: %v", err)
	}
	_, err = br.GetElementWithRelations(ctx, "x")
	assertElementErrIs(t, err, errElementBoom)
}

func TestElementListQueries(t *testing.T) {
	ctx := context.Background()
	two := [][]driver.Value{
		{"e1", "Text", "pg1", nil, int64(1)},
		{"e2", "Text", "pg1", nil, int64(2)},
	}
	cases := []struct {
		call     func(*ElementRepository) ([]models.Element, error)
		fragment string
		arg      any
	}{
		{func(r *ElementRepository) ([]models.Element, error) { return r.GetElementsByPageID(ctx, "pg1") }, `"PageId" =`, "pg1"},
		{func(r *ElementRepository) ([]models.Element, error) { return r.GetChildElements(ctx, "parent") }, `"ParentId" =`, "parent"},
		{func(r *ElementRepository) ([]models.Element, error) { return r.GetRootElements(ctx, "proj") }, `"ParentId" IS NULL`, "proj"},
		{func(r *ElementRepository) ([]models.Element, error) { return r.GetElementsByIDs(ctx, []string{"e1", "e2"}) }, `"Id" IN`, "e2"},
	}
	for _, tc := range cases {
		db, f := newFakeDB(t)
		f.on(elementSelect, elementCols, two...)
		got, err := tc.call(elementNew(db))
		assertElementNoErr(t, err)
		if len(got) != 2 || got[0].Id != "e1" || got[1].Id != "e2" {
			t.Fatalf("got %+v", got)
		}
		if !f.sawQuery(tc.fragment) || !elementArg(f, tc.arg) {
			t.Errorf("query %q args %v", f.queries[0], f.args[0])
		}
		bad, fb := newFakeDB(t)
		fb.onErr("SELECT", errElementBoom)
		_, err = tc.call(elementNew(bad))
		assertElementErrIs(t, err, errElementBoom)
	}
}

func TestElementGetElementsByIDsEmpty(t *testing.T) {
	db, f := newFakeDB(t)
	got, err := elementNew(db).GetElementsByIDs(context.Background(), nil)
	assertElementNoErr(t, err)
	if got == nil || len(got) != 0 || f.queryCount() != 0 {
		t.Errorf("got %#v, queries %d", got, f.queryCount())
	}
}

func TestElementGetElementsByPageIds(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(elementSelect, elementCols,
		[]driver.Value{"root", "Frame", "pg1", nil, int64(1)},
		[]driver.Value{"child", "Text", "pg1", "root", int64(2)})
	got, err := elementNew(db).GetElementsByPageIds(ctx, "proj", []string{"pg1", "pg2"})
	assertElementNoErr(t, err)
	assertElementTreeRootWithChild(t, got, "root")
	if !f.sawQuery(`"ProjectId" =`, `"DeletedAt" IS NULL`) || !elementArg(f, "proj") || !elementArg(f, "pg2") {
		t.Errorf("query %q args %v", f.queries[0], f.args[0])
	}

	emptyDB, _ := newFakeDB(t)
	emptyGot, err := elementNew(emptyDB).GetElementsByPageIds(ctx, "proj", []string{"pg1"})
	assertElementNoErr(t, err)
	if emptyGot == nil || len(emptyGot) != 0 {
		t.Errorf("got %#v", emptyGot)
	}

	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errElementBoom)
	_, err = elementNew(bad).GetElementsByPageIds(ctx, "proj", []string{"pg1"})
	assertElementErrIs(t, err, errElementBoom)
}

func TestElementCreate(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("INSERT INTO", 1)
	e := &models.Element{Type: "Text"}
	assertElementNoErr(t, elementNew(db).CreateElement(ctx, e))
	if e.Id == "" || !f.sawQuery("INSERT INTO", "Element") || !elementArg(f, e.Id) {
		t.Errorf("id=%q queries=%v", e.Id, f.queries)
	}

	dbKeep, _ := newFakeDB(t)
	kept := &models.Element{Id: "mine", Type: "Text"}
	assertElementNoErr(t, elementNew(dbKeep).CreateElement(ctx, kept))
	if kept.Id != "mine" {
		t.Errorf("id=%q", kept.Id)
	}

	bad, fb := newFakeDB(t)
	fb.onErr("INSERT INTO", errElementBoom)
	assertElementErrIs(t, elementNew(bad).CreateElement(ctx, &models.Element{Type: "Text"}), errElementBoom)
}

func TestElementUpdate(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 1)
	err := elementNew(db).UpdateElement(ctx, &models.Element{Id: "e1", Type: "Text", Content: elementStrPtr("hello")})
	assertElementNoErr(t, err)
	if !f.sawQuery("UPDATE", "Element", `"Id" =`) || !elementArg(f, "hello") {
		t.Errorf("query %q", f.lastQuery())
	}

	miss, _ := newFakeDB(t)
	assertElementErrIs(t, elementNew(miss).UpdateElement(ctx, &models.Element{Id: "e1", Type: "Text"}), ErrElementNotFound)

	bad, fb := newFakeDB(t)
	fb.onErr("UPDATE", errElementBoom)
	assertElementErrIs(t, elementNew(bad).UpdateElement(ctx, &models.Element{Id: "e1", Type: "Text"}), errElementBoom)
}

func TestElementUpdateEventWorkflows(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 1)
	assertElementNoErr(t, elementNew(db).UpdateEventWorkflows(ctx, "e1", []byte(`[{"a":1}]`)))
	if !f.sawQuery("UPDATE", `"EventWorkflows"`) {
		t.Errorf("query %q", f.lastQuery())
	}

	miss, _ := newFakeDB(t)
	assertElementErrIs(t, elementNew(miss).UpdateEventWorkflows(ctx, "e1", nil), ErrElementNotFound)

	bad, fb := newFakeDB(t)
	fb.onErr("UPDATE", errElementBoom)
	assertElementErrIs(t, elementNew(bad).UpdateEventWorkflows(ctx, "e1", nil), errElementBoom)
}

func TestElementDeleteByID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("DELETE FROM", 1)
	assertElementNoErr(t, elementNew(db).DeleteElementByID(ctx, "e1"))
	if !f.sawQuery("DELETE FROM", `"Id" =`) {
		t.Errorf("query %q", f.lastQuery())
	}

	empty, _ := newFakeDB(t)
	assertElementErrIs(t, elementNew(empty).DeleteElementByID(ctx, "e1"), ErrElementNotFound)

	bad, fb := newFakeDB(t)
	fb.onErr("DELETE FROM", errElementBoom)
	assertElementErrIs(t, elementNew(bad).DeleteElementByID(ctx, "e1"), errElementBoom)
}

func TestElementDeletesByPageAndProject(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	assertElementNoErr(t, elementNew(db).DeleteElementsByPageID(ctx, "pg1"))
	if !f.sawQuery("DELETE FROM", `"PageId" =`) {
		t.Errorf("query %q", f.lastQuery())
	}
	badPage, fbPage := newFakeDB(t)
	fbPage.onErr("DELETE FROM", errElementBoom)
	assertElementErrIs(t, elementNew(badPage).DeleteElementsByPageID(ctx, "pg1"), errElementBoom)

	dbProj, fProj := newFakeDB(t)
	fProj.onExec("DELETE FROM", 5)
	assertElementNoErr(t, elementNew(dbProj).DeleteElementsByProjectID(ctx, "proj"))
	if !fProj.sawQuery("DELETE FROM", "Element") || !elementArg(fProj, "proj") {
		t.Errorf("query %q", fProj.lastQuery())
	}
	badProj, fbProj := newFakeDB(t)
	fbProj.onErr("DELETE", errElementBoom)
	assertElementErrIs(t, elementNew(badProj).DeleteElementsByProjectID(ctx, "proj"), errElementBoom)
}

func TestElementCountByProject(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("count(*)", []string{"count"}, []driver.Value{int64(12)})
	n, err := elementNew(db).CountElementsByProjectID(ctx, "proj")
	assertElementNoErr(t, err)
	if n != 12 || !elementArg(f, "proj") {
		t.Errorf("count=%d args=%v", n, f.args)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("count(*)", errElementBoom)
	n, err = elementNew(bad).CountElementsByProjectID(ctx, "proj")
	if n != 0 {
		t.Errorf("want 0 count, got %d", n)
	}
	assertElementErrIs(t, err, errElementBoom)
}

func TestElementReplaceElementsTree(t *testing.T) {
	ctx := context.Background()
	page := elementStrPtr("pg1")
	db, f := newFakeDB(t)
	f.onExec("DELETE FROM", 2)
	f.onExec("INSERT INTO", 2)
	roots := []models.EditorElement{
		&models.Element{Id: "a", Type: "Text", PageId: page},
		&models.Element{Id: "b", Type: "Text", PageId: page},
	}
	assertElementNoErr(t, elementNew(db).ReplaceElements(ctx, "proj", roots))
	if f.begins != 1 || f.commits != 1 || f.aborts != 0 {
		t.Errorf("tx begins=%d commits=%d aborts=%d", f.begins, f.commits, f.aborts)
	}
	del, ins := findDeleteAndInsertIndices(f.queries)
	if del < 0 || ins < 0 || del > ins || !elementArg(f, "proj") {
		t.Fatalf("delete must precede insert: %v", f.queries)
	}
	assertInsertedElementArgs(t, f, ins)
}

func findDeleteAndInsertIndices(queries []string) (int, int) {
	del, ins := -1, -1
	for i, q := range queries {
		if strings.HasPrefix(q, "DELETE FROM") && del < 0 {
			del = i
		}
		if strings.HasPrefix(q, "INSERT INTO") && ins < 0 {
			ins = i
		}
	}
	return del, ins
}

func assertInsertedElementArgs(t *testing.T, f *fakeDB, ins int) {
	t.Helper()
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
}

func TestElementReplaceElementsEmptyAndRollbacks(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("DELETE FROM", 3)
	assertElementNoErr(t, elementNew(db).ReplaceElements(ctx, "proj", nil))
	if f.sawQuery("INSERT INTO") || f.commits != 1 {
		t.Errorf("queries=%v commits=%d", f.queries, f.commits)
	}

	dbDel, fDel := newFakeDB(t)
	fDel.onErr("DELETE FROM", errElementBoom)
	err := elementNew(dbDel).ReplaceElements(ctx, "proj", []models.EditorElement{&models.Element{Id: "a", Type: "Text"}})
	assertElementErrIs(t, err, errElementBoom)
	if fDel.commits != 0 || fDel.aborts != 1 || fDel.sawQuery("INSERT INTO") {
		t.Errorf("commits=%d aborts=%d queries=%v", fDel.commits, fDel.aborts, fDel.queries)
	}

	dbIns, fIns := newFakeDB(t)
	fIns.onErr("INSERT INTO", errElementBoom)
	err = elementNew(dbIns).ReplaceElements(ctx, "proj", []models.EditorElement{&models.Element{Id: "a", Type: "Text"}})
	assertElementErrIs(t, err, errElementBoom)
	if fIns.commits != 0 || fIns.aborts != 1 {
		t.Errorf("commits=%d aborts=%d", fIns.commits, fIns.aborts)
	}
}

func TestElementReplaceElementsUnconvertibleChild(t *testing.T) {
	db, f := newFakeDB(t)
	bad := &utils.ElementWithChildren{Element: &models.Element{Id: "root", Type: "Frame"}, Elements: []any{make(chan int)}}
	err := elementNew(db).ReplaceElements(context.Background(), "proj", []models.EditorElement{bad})
	if err == nil || !strings.Contains(err.Error(), "failed to convert child element") {
		t.Errorf("err = %v", err)
	}
	if f.sawQuery("INSERT INTO") || f.commits != 0 {
		t.Errorf("nothing may be inserted: %v", f.queries)
	}
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
	assertElementNoErr(t, err)
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
