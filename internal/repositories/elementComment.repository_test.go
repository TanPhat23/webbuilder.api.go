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

func TestElementCommentCreate(t *testing.T) {
	ctx := context.Background()
	t.Run("generates id, inserts and refetches", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onExec("INSERT INTO", 1)
		f.on(elementCommentSelect, elementCommentCols, []driver.Value{"stored", "hi", "u1", "e1", false})
		in := &models.ElementComment{Content: "hi", AuthorId: "u1", ElementId: "e1"}
		got, err := elementCommentNew(db).CreateElementComment(ctx, in)
		if err != nil || got.Id != "stored" || got.Content != "hi" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if in.Id == "" {
			t.Error("id should be generated")
		}
		if !f.sawQuery("INSERT INTO", "ElementComment") || !elementCommentArg(f, in.Id) {
			t.Errorf("queries: %v", f.queries)
		}
	})
	t.Run("keeps supplied id", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(elementCommentSelect, elementCommentCols, []driver.Value{"mine", "hi", "u1", "e1", false})
		in := &models.ElementComment{Id: "mine", Content: "hi", AuthorId: "u1", ElementId: "e1"}
		if _, err := elementCommentNew(db).CreateElementComment(ctx, in); err != nil || in.Id != "mine" {
			t.Fatalf("id=%q err=%v", in.Id, err)
		}
	})
	t.Run("insert fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("INSERT INTO", errElementCommentBoom)
		if _, err := elementCommentNew(db).CreateElementComment(ctx, &models.ElementComment{}); !errors.Is(err, errElementCommentBoom) {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("refetch finds nothing", func(t *testing.T) {
		db, _ := newFakeDB(t)
		if _, err := elementCommentNew(db).CreateElementComment(ctx, &models.ElementComment{}); err == nil || err.Error() != "element comment not found" {
			t.Errorf("err = %v", err)
		}
	})
}

func TestElementCommentGetByID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on(elementCommentSelect, elementCommentCols, []driver.Value{"c1", "hi", "u1", "e1", true})
	got, err := elementCommentNew(db).GetElementCommentByID(ctx, "c1")
	if err != nil || got.Id != "c1" || !got.Resolved {
		t.Fatalf("got %+v, %v", got, err)
	}
	if !f.sawQuery(`"DeletedAt" IS NULL`) {
		t.Errorf("soft-deleted comments must be excluded: %q", f.queries[0])
	}
	empty, _ := newFakeDB(t)
	if _, err := elementCommentNew(empty).GetElementCommentByID(ctx, "x"); err == nil || err.Error() != "element comment not found" {
		t.Errorf("missing: %v", err)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errElementCommentBoom)
	if _, err := elementCommentNew(bad).GetElementCommentByID(ctx, "x"); !errors.Is(err, errElementCommentBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestElementCommentGetElementComments(t *testing.T) {
	ctx := context.Background()
	resolved := true
	for _, tc := range []struct {
		name      string
		filter    *models.ElementCommentFilter
		wantSQL   []string
		denySQL   []string
		wantInts  []int
		wantAnyOf []any
	}{
		{"nil filter sorts newest first", nil, []string{`ORDER BY "CreatedAt" DESC`}, []string{"LIMIT", "OFFSET", `"AuthorId"`}, nil, nil},
		{"empty filter sorts newest first", &models.ElementCommentFilter{}, []string{`ORDER BY "CreatedAt" DESC`}, []string{"LIMIT", "OFFSET"}, nil, nil},
		{"all filters ascending", &models.ElementCommentFilter{AuthorId: "u9", Resolved: &resolved, SortBy: "UpdatedAt", Limit: 5, Offset: 10},
			[]string{`"AuthorId" =`, `"Resolved" =`, `ORDER BY "UpdatedAt" ASC`, "LIMIT", "OFFSET"}, nil, []int{5, 10}, []any{"u9", true}},
		{"descending", &models.ElementCommentFilter{SortBy: "Content", SortOrder: "DESC"}, []string{`ORDER BY "Content" DESC`}, nil, nil, nil},
		{"unknown order direction means ascending", &models.ElementCommentFilter{SortBy: "Content", SortOrder: "desc"}, []string{`ORDER BY "Content" ASC`}, nil, nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, f := newFakeDB(t)
			f.on(elementCommentSelect, elementCommentCols, []driver.Value{"c1", "hi", "u1", "e1", false})
			got, err := elementCommentNew(db).GetElementComments(ctx, "e1", tc.filter)
			if err != nil || len(got) != 1 || got[0].Id != "c1" {
				t.Fatalf("got %+v, %v", got, err)
			}
			q := f.queries[0]
			if !f.sawQuery(tc.wantSQL...) {
				t.Errorf("want %v in %q", tc.wantSQL, q)
			}
			for _, d := range tc.denySQL {
				if strings.Contains(q, d) {
					t.Errorf("unexpected %q in %q", d, q)
				}
			}
			for _, n := range tc.wantInts {
				if !elementCommentIntArg(f, n) {
					t.Errorf("%d not bound: %v", n, f.args[0])
				}
			}
			for _, a := range tc.wantAnyOf {
				if !elementCommentArg(f, a) {
					t.Errorf("%v not bound: %v", a, f.args[0])
				}
			}
		})
	}
	bad, fb := newFakeDB(t)
	fb.onErr("SELECT", errElementCommentBoom)
	if _, err := elementCommentNew(bad).GetElementComments(ctx, "e1", nil); !errors.Is(err, errElementCommentBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestElementCommentUpdate(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"c1", "edited", "u1", "e1", false}
	t.Run("ok", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(elementCommentSelect, elementCommentCols, row)
		f.onExec("UPDATE", 1)
		got, err := elementCommentNew(db).UpdateElementComment(ctx, "c1", map[string]any{"Content": "edited"})
		if err != nil || got.Content != "edited" {
			t.Fatalf("got %+v, %v", got, err)
		}
		if !f.sawQuery("UPDATE", "ElementComment", `"Content"`) || !elementCommentArg(f, "edited") {
			t.Errorf("queries: %v", f.queries)
		}
	})
	t.Run("not found", func(t *testing.T) {
		db, f := newFakeDB(t)
		if _, err := elementCommentNew(db).UpdateElementComment(ctx, "c1", map[string]any{"Content": "x"}); err == nil {
			t.Error("expected not found error")
		}
		if f.sawQuery("UPDATE") {
			t.Error("must not update a missing comment")
		}
	})
	t.Run("update fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(elementCommentSelect, elementCommentCols, row)
		f.onErr("UPDATE", errElementCommentBoom)
		if _, err := elementCommentNew(db).UpdateElementComment(ctx, "c1", map[string]any{"Content": "x"}); !errors.Is(err, errElementCommentBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestElementCommentDelete(t *testing.T) {
	ctx := context.Background()
	row := []driver.Value{"c1", "hi", "u1", "e1", false}
	t.Run("soft deletes", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(elementCommentSelect, elementCommentCols, row)
		f.onExec("UPDATE", 1)
		if err := elementCommentNew(db).DeleteElementComment(ctx, "c1"); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("UPDATE", "ElementComment", "CURRENT_TIMESTAMP", `"DeletedAt"`) || f.sawQuery("DELETE FROM") {
			t.Errorf("expected soft delete: %v", f.queries)
		}
	})
	t.Run("missing", func(t *testing.T) {
		db, f := newFakeDB(t)
		if err := elementCommentNew(db).DeleteElementComment(ctx, "c1"); err == nil || f.sawQuery("UPDATE") {
			t.Errorf("err = %v, queries %v", err, f.queries)
		}
	})
	t.Run("update fails", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(elementCommentSelect, elementCommentCols, row)
		f.onErr("UPDATE", errElementCommentBoom)
		if err := elementCommentNew(db).DeleteElementComment(ctx, "c1"); !errors.Is(err, errElementCommentBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestElementCommentByAuthor(t *testing.T) {
	ctx := context.Background()
	t.Run("with paging", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on(elementCommentSelect, elementCommentCols, []driver.Value{"c1", "hi", "u1", "e1", false})
		got, err := elementCommentNew(db).GetElementCommentsByAuthorID(ctx, "u1", 3, 6)
		if err != nil || len(got) != 1 {
			t.Fatalf("got %+v, %v", got, err)
		}
		if !f.sawQuery(`"AuthorId" =`, "LIMIT", "OFFSET") || !elementCommentIntArg(f, 3) || !elementCommentIntArg(f, 6) {
			t.Errorf("query %q args %v", f.queries[0], f.args[0])
		}
	})
	t.Run("without paging", func(t *testing.T) {
		db, f := newFakeDB(t)
		if _, err := elementCommentNew(db).GetElementCommentsByAuthorID(ctx, "u1", 0, 0); err != nil {
			t.Fatal(err)
		}
		if f.sawQuery("LIMIT") || f.sawQuery("OFFSET") {
			t.Errorf("unexpected paging: %q", f.queries[0])
		}
	})
	t.Run("error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", errElementCommentBoom)
		if _, err := elementCommentNew(db).GetElementCommentsByAuthorID(ctx, "u1", 0, 0); !errors.Is(err, errElementCommentBoom) {
			t.Errorf("err = %v", err)
		}
	})
}

func TestElementCommentCounts(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.on("count(*)", []string{"count"}, []driver.Value{int64(9)})
	repo := elementCommentNew(db)
	if n, err := repo.CountElementComments(ctx, "e1"); err != nil || n != 9 {
		t.Errorf("CountElementComments = %d, %v", n, err)
	}
	if n, err := repo.CountElementCommentsByProjectID(ctx, "p1"); err != nil || n != 9 {
		t.Errorf("CountElementCommentsByProjectID = %d, %v", n, err)
	}
	if !f.sawQuery("JOIN", `"Page"`, `"ProjectId" =`) {
		t.Errorf("project count must join pages: %v", f.queries)
	}
	bad, fb := newFakeDB(t)
	fb.onErr("count(*)", errElementCommentBoom)
	br := elementCommentNew(bad)
	if n, err := br.CountElementComments(ctx, "e1"); n != 0 || !errors.Is(err, errElementCommentBoom) {
		t.Errorf("count error = %d, %v", n, err)
	}
	if n, err := br.CountElementCommentsByProjectID(ctx, "p1"); n != 0 || !errors.Is(err, errElementCommentBoom) {
		t.Errorf("project count error = %d, %v", n, err)
	}
}

func TestElementCommentToggleResolved(t *testing.T) {
	ctx := context.Background()
	for _, resolved := range []bool{false, true} {
		db, f := newFakeDB(t)
		f.on(elementCommentSelect, elementCommentCols, []driver.Value{"c1", "hi", "u1", "e1", resolved})
		f.onExec("UPDATE", 1)
		if _, err := elementCommentNew(db).ToggleResolvedStatus(ctx, "c1"); err != nil {
			t.Fatal(err)
		}
		if !f.sawQuery("UPDATE", "ElementComment", `"Resolved"`) {
			t.Fatalf("no update: %v", f.queries)
		}
		var updateArgs []driver.NamedValue
		for i, q := range f.queries {
			if strings.HasPrefix(q, "UPDATE") {
				updateArgs = f.args[i]
			}
		}
		found := false
		for _, a := range updateArgs {
			if v, ok := a.Value.(bool); ok && v == !resolved {
				found = true
			}
		}
		if !found {
			t.Errorf("resolved=%v should be flipped; args %+v", resolved, updateArgs)
		}
	}
	missing, _ := newFakeDB(t)
	if _, err := elementCommentNew(missing).ToggleResolvedStatus(ctx, "c1"); err == nil {
		t.Error("expected error for missing comment")
	}
	bad, fb := newFakeDB(t)
	fb.on(elementCommentSelect, elementCommentCols, []driver.Value{"c1", "hi", "u1", "e1", false})
	fb.onErr("UPDATE", errElementCommentBoom)
	if _, err := elementCommentNew(bad).ToggleResolvedStatus(ctx, "c1"); !errors.Is(err, errElementCommentBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestElementCommentDeleteByElementID(t *testing.T) {
	ctx := context.Background()
	db, f := newFakeDB(t)
	f.onExec("UPDATE", 3)
	if err := elementCommentNew(db).DeleteElementCommentsByElementID(ctx, "e1"); err != nil {
		t.Fatal(err)
	}
	if !f.sawQuery("UPDATE", "ElementComment", "CURRENT_TIMESTAMP", `"ElementId" =`) {
		t.Errorf("query: %q", f.lastQuery())
	}
	bad, fb := newFakeDB(t)
	fb.onErr("UPDATE", errElementCommentBoom)
	if err := elementCommentNew(bad).DeleteElementCommentsByElementID(ctx, "e1"); !errors.Is(err, errElementCommentBoom) {
		t.Errorf("err = %v", err)
	}
}

func TestElementCommentByProjectID(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	t.Run("with paging", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.on("JOIN public.\"Element\"", append([]string{"CreatedAt"}, elementCommentCols...), []driver.Value{now, "c1", "hi", "u1", "e1", false})
		got, err := elementCommentNew(db).GetElementCommentsByProjectID(ctx, "p1", 4, 8)
		if err != nil || len(got) != 1 || got[0].Id != "c1" || !got[0].CreatedAt.Equal(now) {
			t.Fatalf("got %+v, %v", got, err)
		}
		if !f.sawQuery("LIMIT", "OFFSET", `"ProjectId" =`) || !elementCommentIntArg(f, 4) || !elementCommentIntArg(f, 8) {
			t.Errorf("query %q args %v", f.queries[0], f.args[0])
		}
	})
	t.Run("without paging", func(t *testing.T) {
		db, f := newFakeDB(t)
		if _, err := elementCommentNew(db).GetElementCommentsByProjectID(ctx, "p1", 0, 0); err != nil {
			t.Fatal(err)
		}
		if f.sawQuery("LIMIT") || f.sawQuery("OFFSET") {
			t.Errorf("unexpected paging: %q", f.queries[0])
		}
	})
	t.Run("error", func(t *testing.T) {
		db, f := newFakeDB(t)
		f.onErr("SELECT", errElementCommentBoom)
		if _, err := elementCommentNew(db).GetElementCommentsByProjectID(ctx, "p1", 0, 0); !errors.Is(err, errElementCommentBoom) {
			t.Errorf("err = %v", err)
		}
	})
}
