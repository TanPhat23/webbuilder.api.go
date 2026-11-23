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

var errElementCommentBoom = errors.New("element comment boom")

const elementCommentSelect = `SELECT * FROM "public"."ElementComment"`

var elementCommentCols = []string{"Id", "Content", "AuthorId", "ElementId", "Resolved"}

func elementCommentNew(db *gorm.DB) *ElementCommentRepository {
	return &ElementCommentRepository{db: db}
}

func elementCommentArg(f *fakeDB, want any) bool {
	for _, q := range f.args {
		for _, a := range q {
			if a.Value == want {
				return true
			}
		}
	}
	return false
}

func elementCommentIntArg(f *fakeDB, want int) bool {
	return elementCommentArg(f, int64(want)) || elementCommentArg(f, want)
}

func assertCommentNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertCommentErrIs(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("want %v, got %v", want, err)
	}
}

func TestElementCommentValidationErrors(t *testing.T) {
	ctx := context.Background()
	cases := map[string]func(*ElementCommentRepository) error{
		"comment cannot be nil": func(r *ElementCommentRepository) error { _, err := r.CreateElementComment(ctx, nil); return err },
		"comment id is required": func(r *ElementCommentRepository) error {
			_, err := r.GetElementCommentByID(ctx, "")
			return err
		},
		"elementId is required": func(r *ElementCommentRepository) error {
			_, err := r.GetElementComments(ctx, "", nil)
			return err
		},
		"comment id is required (update)": func(r *ElementCommentRepository) error {
			_, err := r.UpdateElementComment(ctx, "", map[string]any{"Content": "x"})
			return err
		},
		"no updates provided": func(r *ElementCommentRepository) error {
			_, err := r.UpdateElementComment(ctx, "c1", nil)
			return err
		},
		"comment id is required (delete)": func(r *ElementCommentRepository) error { return r.DeleteElementComment(ctx, "") },
		"authorId is required": func(r *ElementCommentRepository) error {
			_, err := r.GetElementCommentsByAuthorID(ctx, "", 1, 0)
			return err
		},
		"elementId is required (count)": func(r *ElementCommentRepository) error {
			_, err := r.CountElementComments(ctx, "")
			return err
		},
		"comment id is required (toggle)": func(r *ElementCommentRepository) error {
			_, err := r.ToggleResolvedStatus(ctx, "")
			return err
		},
		"elementId is required (delete by element)": func(r *ElementCommentRepository) error {
			return r.DeleteElementCommentsByElementID(ctx, "")
		},
		"projectId is required": func(r *ElementCommentRepository) error {
			_, err := r.GetElementCommentsByProjectID(ctx, "", 1, 0)
			return err
		},
		"projectId is required (count)": func(r *ElementCommentRepository) error {
			_, err := r.CountElementCommentsByProjectID(ctx, "")
			return err
		},
	}
	for want, call := range cases {
		db, f := newFakeDB(t)
		err := call(elementCommentNew(db))
		if err == nil || !strings.HasPrefix(want, err.Error()) {
			t.Errorf("want error %q, got %v", want, err)
		}
		if f.queryCount() != 0 {
			t.Errorf("%s: validation must not touch the database", want)
		}
	}
}

func TestElementCommentCreateSuccess(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("INSERT INTO", 1)
	f.on(elementCommentSelect, elementCommentCols, []driver.Value{"stored", "hi", "u1", "e1", false})
	in := &models.ElementComment{Content: "hi", AuthorId: "u1", ElementId: "e1"}
	got, err := elementCommentNew(db).CreateElementComment(ctx, in)
	assertCommentNoErr(t, err)
	if got.Id != "stored" || got.Content != "hi" || in.Id == "" {
		t.Fatalf("got=%+v in=%+v", got, in)
	}
	if !f.sawQuery("INSERT INTO", "ElementComment") || !elementCommentArg(f, in.Id) {
		t.Errorf("queries: %v", f.queries)
	}
}

func TestElementCommentCreatePreservesIDAndErrors(t *testing.T) {
	ctx := context.Background()
	dbKeep, fKeep := newFakeDB(t)
	fKeep.on(elementCommentSelect, elementCommentCols, []driver.Value{"mine", "hi", "u1", "e1", false})
	keepIn := &models.ElementComment{Id: "mine", Content: "hi", AuthorId: "u1", ElementId: "e1"}
	_, err := elementCommentNew(dbKeep).CreateElementComment(ctx, keepIn)
	assertCommentNoErr(t, err)
	if keepIn.Id != "mine" {
		t.Fatalf("id=%q", keepIn.Id)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.onErr("INSERT INTO", errElementCommentBoom)
	_, err = elementCommentNew(dbErr).CreateElementComment(ctx, &models.ElementComment{})
	assertCommentErrIs(t, err, errElementCommentBoom)

	dbMiss, fMiss := newFakeDB(t)
	if fMiss == nil {
		t.Fatal("missing fake db")
	}
	_, err = elementCommentNew(dbMiss).CreateElementComment(ctx, &models.ElementComment{})
	if err == nil || err.Error() != "element comment not found" {
		t.Errorf("err = %v", err)
	}
}

func TestElementCommentGetByID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(elementCommentSelect, elementCommentCols, []driver.Value{"c1", "hi", "u1", "e1", true})
	got, err := elementCommentNew(db).GetElementCommentByID(ctx, "c1")
	assertCommentNoErr(t, err)
	if got.Id != "c1" || !got.Resolved {
		t.Fatalf("got=%+v", got)
	}
	if !f.sawQuery(`"DeletedAt" IS NULL`) {
		t.Fatalf("queries=%v", f.queries)
	}

	empty, emptyFake := newFakeDB(t)
	if emptyFake == nil {
		t.Fatal("missing fake db")
	}
	if _, err := elementCommentNew(empty).GetElementCommentByID(ctx, "x"); err == nil || err.Error() != "element comment not found" {
		t.Errorf("missing: %v", err)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errElementCommentBoom)
	_, err = elementCommentNew(bad).GetElementCommentByID(ctx, "x")
	assertCommentErrIs(t, err, errElementCommentBoom)
}

func TestElementCommentGetElementComments(t *testing.T) {
	ctx := context.Background()
	resolved := true
	for _, tc := range []struct {
		filter    *models.ElementCommentFilter
		wantSQL   []string
		denySQL   []string
		wantInts  []int
		wantAnyOf []any
	}{
		{nil, []string{`ORDER BY "CreatedAt" DESC`}, []string{"LIMIT", "OFFSET", `"AuthorId"`}, nil, nil},
		{&models.ElementCommentFilter{}, []string{`ORDER BY "CreatedAt" DESC`}, []string{"LIMIT", "OFFSET"}, nil, nil},
		{&models.ElementCommentFilter{AuthorId: "u9", Resolved: &resolved, SortBy: "UpdatedAt", Limit: 5, Offset: 10},
			[]string{`"AuthorId" =`, `"Resolved" =`, `ORDER BY "UpdatedAt" ASC`, "LIMIT", "OFFSET"}, nil, []int{5, 10}, []any{"u9", true}},
		{&models.ElementCommentFilter{SortBy: "Content", SortOrder: "DESC"}, []string{`ORDER BY "Content" DESC`}, nil, nil, nil},
		{&models.ElementCommentFilter{SortBy: "Content", SortOrder: "desc"}, []string{`ORDER BY "Content" ASC`}, nil, nil, nil},
	} {
		db, f := newFakeDB(t)
		f.on(elementCommentSelect, elementCommentCols, []driver.Value{"c1", "hi", "u1", "e1", false})
		got, err := elementCommentNew(db).GetElementComments(ctx, "e1", tc.filter)
		assertCommentNoErr(t, err)
		assertCommentListFilterMatch(t, f, got, tc.wantSQL, tc.denySQL, tc.wantInts, tc.wantAnyOf)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errElementCommentBoom)
	_, err := elementCommentNew(bad).GetElementComments(ctx, "e1", nil)
	assertCommentErrIs(t, err, errElementCommentBoom)
}

func assertCommentListFilterMatch(t *testing.T, f *fakeDB, got []models.ElementComment, wantSQL, denySQL []string, wantInts []int, wantAny []any) {
	t.Helper()
	if len(got) != 1 || got[0].Id != "c1" {
		t.Fatalf("got %+v", got)
	}
	q := f.queries[0]
	if !f.sawQuery(wantSQL...) {
		t.Errorf("want %v in %q", wantSQL, q)
	}
	for _, d := range denySQL {
		if strings.Contains(q, d) {
			t.Errorf("unexpected %q in %q", d, q)
		}
	}
	assertCommentFilterArgs(t, f, wantInts, wantAny)
}

func assertCommentFilterArgs(t *testing.T, f *fakeDB, wantInts []int, wantAny []any) {
	t.Helper()
	for _, n := range wantInts {
		if !elementCommentIntArg(f, n) {
			t.Errorf("%d not bound: %v", n, f.args[0])
		}
	}
	for _, a := range wantAny {
		if !elementCommentArg(f, a) {
			t.Errorf("%v not bound: %v", a, f.args[0])
		}
	}
}

func TestElementCommentUpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"c1", "edited", "u1", "e1", false}
	db, f := newFakeDB(t)
	f.on(elementCommentSelect, elementCommentCols, row)
	f.onExec("UPDATE", 1)
	got, err := elementCommentNew(db).UpdateElementComment(ctx, "c1", map[string]any{"Content": "edited"})
	assertCommentNoErr(t, err)
	if got.Content != "edited" || !f.sawQuery("UPDATE", "ElementComment", `"Content"`) || !elementCommentArg(f, "edited") {
		t.Fatalf("got=%+v queries=%v", got, f.queries)
	}

	dbMiss, fMiss := newFakeDB(t)
	if _, err := elementCommentNew(dbMiss).UpdateElementComment(ctx, "c1", map[string]any{"Content": "x"}); err == nil || fMiss.sawQuery("UPDATE") {
		t.Errorf("err=%v queries=%v", err, fMiss.queries)
	}

	dbErr, fErr := newFakeDB(t)
	fErr.on(elementCommentSelect, elementCommentCols, row)
	fErr.onErr("UPDATE", errElementCommentBoom)
	_, err = elementCommentNew(dbErr).UpdateElementComment(ctx, "c1", map[string]any{"Content": "x"})
	assertCommentErrIs(t, err, errElementCommentBoom)
}

func TestElementCommentDelete(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"c1", "hi", "u1", "e1", false}
	db, f := newFakeDB(t)
	f.on(elementCommentSelect, elementCommentCols, row)
	f.onExec("UPDATE", 1)
	assertCommentNoErr(t, elementCommentNew(db).DeleteElementComment(ctx, "c1"))
	if !f.sawQuery("UPDATE", "ElementComment", "CURRENT_TIMESTAMP", `"DeletedAt"`) || f.sawQuery("DELETE FROM") {
		t.Errorf("expected soft delete: %v", f.queries)
	}

	dbMiss, fMiss := newFakeDB(t)
	if err := elementCommentNew(dbMiss).DeleteElementComment(ctx, "c1"); err == nil || fMiss.sawQuery("UPDATE") {
		t.Errorf("err=%v queries=%v", err, fMiss.queries)
	}
	dbErr, fErr := newFakeDB(t)
	fErr.on(elementCommentSelect, elementCommentCols, row)
	fErr.onErr("UPDATE", errElementCommentBoom)
	assertCommentErrIs(t, elementCommentNew(dbErr).DeleteElementComment(ctx, "c1"), errElementCommentBoom)
}

func TestElementCommentByAuthorAndCounts(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(elementCommentSelect, elementCommentCols, []driver.Value{"c1", "hi", "u1", "e1", false})
	got, err := elementCommentNew(db).GetElementCommentsByAuthorID(ctx, "u1", 3, 6)
	assertCommentNoErr(t, err)
	if len(got) != 1 || !f.sawQuery(`"AuthorId" =`, "LIMIT", "OFFSET") {
		t.Fatalf("got=%+v queries=%v", got, f.queries)
	}
	if !elementCommentIntArg(f, 3) || !elementCommentIntArg(f, 6) {
		t.Fatalf("args=%v", f.args)
	}

	dbNoPage, fNoPage := newFakeDB(t)
	_, err = elementCommentNew(dbNoPage).GetElementCommentsByAuthorID(ctx, "u1", 0, 0)
	assertCommentNoErr(t, err)
	if fNoPage.sawQuery("LIMIT") || fNoPage.sawQuery("OFFSET") {
		t.Errorf("unexpected paging: %q", fNoPage.queries[0])
	}

	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errElementCommentBoom)
	_, err = elementCommentNew(bad).GetElementCommentsByAuthorID(ctx, "u1", 0, 0)
	assertCommentErrIs(t, err, errElementCommentBoom)
}

func TestElementCommentCounts(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("count(*)", []string{"count"}, []driver.Value{int64(9)})
	repo := elementCommentNew(db)
	n1, err := repo.CountElementComments(ctx, "e1")
	assertCommentNoErr(t, err)
	n2, err := repo.CountElementCommentsByProjectID(ctx, "p1")
	assertCommentNoErr(t, err)
	if n1 != 9 || n2 != 9 || !f.sawQuery("JOIN", `"Page"`, `"ProjectId" =`) {
		t.Errorf("n1=%d n2=%d queries=%v", n1, n2, f.queries)
	}

	bad, fb := newFakeDB(t)
	fb.onErr("count(*)", errElementCommentBoom)
	br := elementCommentNew(bad)
	_, err = br.CountElementComments(ctx, "e1")
	assertCommentErrIs(t, err, errElementCommentBoom)
	_, err = br.CountElementCommentsByProjectID(ctx, "p1")
	assertCommentErrIs(t, err, errElementCommentBoom)
}

func TestElementCommentToggleResolved(t *testing.T) {
	ctx := context.Background()
	for _, resolved := range []bool{false, true} {
		db, f := newFakeDB(t)
		f.on(elementCommentSelect, elementCommentCols, []driver.Value{"c1", "hi", "u1", "e1", resolved})
		f.onExec("UPDATE", 1)
		_, err := elementCommentNew(db).ToggleResolvedStatus(ctx, "c1")
		assertCommentNoErr(t, err)
		assertResolvedFlipBound(t, f, !resolved)
	}
	missing, missingFake := newFakeDB(t)
	if missingFake == nil {
		t.Fatal("missing fake db")
	}
	if _, err := elementCommentNew(missing).ToggleResolvedStatus(ctx, "c1"); err == nil {
		t.Error("expected error for missing comment")
	}
	bad, fb := newFakeDB(t)
	fb.on(elementCommentSelect, elementCommentCols, []driver.Value{"c1", "hi", "u1", "e1", false})
	fb.onErr("UPDATE", errElementCommentBoom)
	_, err := elementCommentNew(bad).ToggleResolvedStatus(ctx, "c1")
	assertCommentErrIs(t, err, errElementCommentBoom)
}

func assertResolvedFlipBound(t *testing.T, f *fakeDB, want bool) {
	t.Helper()
	if !f.sawQuery("UPDATE", "ElementComment", `"Resolved"`) {
		t.Fatalf("no update: %v", f.queries)
	}
	for i, q := range f.queries {
		if strings.HasPrefix(q, "UPDATE") && hasBoolArg(f.args[i], want) {
			return
		}
	}
	t.Errorf("resolved=%v not bound in update", want)
}

func hasBoolArg(args []driver.NamedValue, want bool) bool {
	for _, a := range args {
		if v, ok := a.Value.(bool); ok && v == want {
			return true
		}
	}
	return false
}

func TestElementCommentDeleteByElementID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 3)
	assertCommentNoErr(t, elementCommentNew(db).DeleteElementCommentsByElementID(ctx, "e1"))
	if !f.sawQuery("UPDATE", "ElementComment", "CURRENT_TIMESTAMP", `"ElementId" =`) {
		t.Errorf("query: %q", f.lastQuery())
	}
	bad, fb := newFakeDB(t)
	fb.onErr("UPDATE", errElementCommentBoom)
	assertCommentErrIs(t, elementCommentNew(bad).DeleteElementCommentsByElementID(ctx, "e1"), errElementCommentBoom)
}

func TestElementCommentByProjectIDPaging(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	db, f := newFakeDB(t)
	f.on(`JOIN public."Element"`, append([]string{"CreatedAt"}, elementCommentCols...), []driver.Value{now, "c1", "hi", "u1", "e1", false})
	got, err := elementCommentNew(db).GetElementCommentsByProjectID(ctx, "p1", 4, 8)
	assertCommentNoErr(t, err)
	if len(got) != 1 || got[0].Id != "c1" {
		t.Fatalf("got %+v", got)
	}
	if !got[0].CreatedAt.Equal(now) {
		t.Fatalf("createdAt=%v want %v", got[0].CreatedAt, now)
	}
	if !f.sawQuery("LIMIT", "OFFSET", `"ProjectId" =`) || !elementCommentIntArg(f, 4) || !elementCommentIntArg(f, 8) {
		t.Errorf("query %q args %v", f.queries[0], f.args[0])
	}
}

func TestElementCommentByProjectIDUnboundedAndErr(t *testing.T) {
	ctx := context.Background()
	dbNoPage, fNoPage := newFakeDB(t)
	_, err := elementCommentNew(dbNoPage).GetElementCommentsByProjectID(ctx, "p1", 0, 0)
	assertCommentNoErr(t, err)
	if fNoPage.sawQuery("LIMIT") || fNoPage.sawQuery("OFFSET") {
		t.Errorf("unexpected paging: %q", fNoPage.queries[0])
	}

	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errElementCommentBoom)
	_, err = elementCommentNew(bad).GetElementCommentsByProjectID(ctx, "p1", 0, 0)
	assertCommentErrIs(t, err, errElementCommentBoom)
}
