package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/authz"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type elementCommentEnv struct {
	svc *ElementCommentService
	cr  *fkElementCommentRepo
	er  *fkElementRepo
	pr  *fkProjectRepo
}

func newElementCommentEnv() *elementCommentEnv {
	e := &elementCommentEnv{cr: &fkElementCommentRepo{}, er: &fkElementRepo{}, pr: &fkProjectRepo{}}
	e.svc = NewElementCommentService(e.cr, e.er, e.pr)
	return e
}

func (e *elementCommentEnv) elementExists() {
	e.er.OnGetByID = func(_ context.Context, id string) (*models.Element, error) {
		return &models.Element{Id: id, Page: &models.Page{ProjectId: "p"}}, nil
	}
	e.pr.OnGetAccess = func(_ context.Context, pid, uid string) (*models.ProjectAccess, error) {
		return &models.ProjectAccess{ProjectId: pid, UserId: uid, Role: models.RoleEditor}, nil
	}
}
func (e *elementCommentEnv) elementMissing() {
	e.er.OnGetByID = func(context.Context, string) (*models.Element, error) { return nil, nil }
}
func (e *elementCommentEnv) elementErr() {
	e.er.OnGetByID = func(context.Context, string) (*models.Element, error) { return nil, errBoom }
}
func (e *elementCommentEnv) commentOwnedBy(author string) {
	e.cr.OnGetByID = func(_ context.Context, id string) (*models.ElementComment, error) {
		return &models.ElementComment{Id: id, ElementId: "e", AuthorId: author}, nil
	}
}
func (e *elementCommentEnv) projectVisible() {
	e.pr.OnGetWithAccess = func(context.Context, string, string) (*models.Project, error) { return &models.Project{ID: "p"}, nil }
}
func (e *elementCommentEnv) projectHidden() {
	e.pr.OnGetWithAccess = func(context.Context, string, string) (*models.Project, error) {
		return nil, repositories.ErrProjectNotFound
	}
	e.pr.OnGetPublicByID = func(context.Context, string) (*models.Project, error) { return nil, nil }
}

func isBoom(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want wrapping errBoom", err)
	}
}

func TestElementCommentCreate(t *testing.T) {
	ctx := context.Background()
	valid := func() *models.ElementComment {
		return &models.ElementComment{ElementId: "e", AuthorId: "a", Content: "hi"}
	}

	e := newElementCommentEnv()
	_, err := e.svc.Create(ctx, nil)
	wantErr(t, err, "cannot be nil")
	_, err = e.svc.Create(ctx, &models.ElementComment{AuthorId: "a", Content: "c"})
	wantErr(t, err, "elementId is required")
	_, err = e.svc.Create(ctx, &models.ElementComment{ElementId: "e", Content: "c"})
	wantErr(t, err, "authorId is required")
	_, err = e.svc.Create(ctx, &models.ElementComment{ElementId: "e", AuthorId: "a"})
	wantErr(t, err, "content cannot be empty")
	wantCalls(t, e.er.Calls)

	e.elementErr()
	_, err = e.svc.Create(ctx, valid())
	wantErr(t, err, "element not found")
	isBoom(t, err)

	e.elementMissing()
	_, err = e.svc.Create(ctx, valid())
	wantErr(t, err, "element does not exist")
	wantCalls(t, e.cr.Calls)

	e.elementExists()
	e.cr.OnCreate = func(_ context.Context, c *models.ElementComment) (*models.ElementComment, error) {
		c.Id = "new"
		return c, nil
	}
	got, err := e.svc.Create(ctx, valid())
	wantNoErr(t, err)
	if got.Id != "new" {
		t.Fatalf("got %+v", got)
	}
	e.cr.OnCreate = func(context.Context, *models.ElementComment) (*models.ElementComment, error) { return nil, errBoom }
	_, err = e.svc.Create(ctx, valid())
	isBoom(t, err)
}

func TestElementCommentGetters(t *testing.T) {
	ctx := context.Background()

	t.Run("ByID", func(t *testing.T) {
		e := newElementCommentEnv()
		_, err := e.svc.GetByID(ctx, "")
		wantErr(t, err, "comment id is required")
		e.commentOwnedBy("a")
		got, err := e.svc.GetByID(ctx, "c")
		wantNoErr(t, err)
		if got.Id != "c" {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("List", func(t *testing.T) {
		e := newElementCommentEnv()
		_, err := e.svc.List(ctx, "", models.ElementCommentFilter{})
		wantErr(t, err, "elementId is required")
		e.elementErr()
		_, err = e.svc.List(ctx, "e", models.ElementCommentFilter{})
		wantErr(t, err, "failed to verify element")
		isBoom(t, err)
		e.elementMissing()
		_, err = e.svc.List(ctx, "e", models.ElementCommentFilter{})
		wantErr(t, err, "element does not exist")
		e.elementExists()
		var gotEl string
		e.cr.OnList = func(_ context.Context, id string, f *models.ElementCommentFilter) ([]models.ElementComment, error) {
			gotEl = id
			if f == nil {
				t.Error("nil filter")
			}
			return []models.ElementComment{{Id: "c"}}, nil
		}
		got, err := e.svc.List(ctx, "e", models.ElementCommentFilter{})
		wantNoErr(t, err)
		if len(got) != 1 || gotEl != "e" {
			t.Fatalf("got %v %q", got, gotEl)
		}
	})

	t.Run("ByAuthor", func(t *testing.T) {
		e := newElementCommentEnv()
		_, err := e.svc.ListByAuthor(ctx, "", 1, 0)
		wantErr(t, err, "authorId is required")
		_, err = e.svc.ListByAuthor(ctx, "a", -1, 0)
		wantErr(t, err, "limit cannot be negative")
		_, err = e.svc.ListByAuthor(ctx, "a", 1, -1)
		wantErr(t, err, "offset cannot be negative")
		e.cr.OnListByAuthor = func(_ context.Context, a string, l, o int) ([]models.ElementComment, error) {
			if a != "a" || l != 2 || o != 3 {
				t.Errorf("args %q %d %d", a, l, o)
			}
			return nil, errBoom
		}
		_, err = e.svc.ListByAuthor(ctx, "a", 2, 3)
		isBoom(t, err)
	})

	t.Run("ByProject", func(t *testing.T) {
		e := newElementCommentEnv()
		_, err := e.svc.ListByProject(ctx, "", "u", 1, 0)
		wantErr(t, err, "projectId is required")
		_, err = e.svc.ListByProject(ctx, "p", "u", -1, 0)
		wantErr(t, err, "limit cannot be negative")
		_, err = e.svc.ListByProject(ctx, "p", "u", 1, -1)
		wantErr(t, err, "offset cannot be negative")
		e.projectHidden()
		_, err = e.svc.ListByProject(ctx, "p", "u", 1, 0)
		wantErr(t, err, "failed to verify project")
		if !errors.Is(err, repositories.ErrProjectNotFound) {
			t.Fatalf("err = %v", err)
		}
		wantCalls(t, e.cr.Calls)
		e.projectVisible()
		e.cr.OnListByProject = func(context.Context, string, int, int) ([]models.ElementComment, error) {
			return []models.ElementComment{{Id: "c"}}, nil
		}
		got, err := e.svc.ListByProject(ctx, "p", "u", 1, 0)
		wantNoErr(t, err)
		if len(got) != 1 {
			t.Fatalf("got %v", got)
		}
	})
}

func TestElementCommentUpdateDelete(t *testing.T) {
	ctx := context.Background()
	upd := map[string]any{"content": "x"}

	t.Run("Update", func(t *testing.T) {
		e := newElementCommentEnv()
		_, err := e.svc.Update(ctx, "", "u", upd)
		wantErr(t, err, "comment id is required")
		_, err = e.svc.Update(ctx, "c", "", upd)
		wantErr(t, err, "userID is required")
		_, err = e.svc.Update(ctx, "c", "u", nil)
		wantErr(t, err, "no updates provided")
		e.cr.OnGetByID = func(context.Context, string) (*models.ElementComment, error) { return nil, errBoom }
		_, err = e.svc.Update(ctx, "c", "u", upd)
		wantErr(t, err, "failed to retrieve comment")
		isBoom(t, err)
		e.cr.OnGetByID = func(context.Context, string) (*models.ElementComment, error) { return nil, nil }
		_, err = e.svc.Update(ctx, "c", "u", upd)
		wantErr(t, err, "comment does not exist")
		e.elementExists()
		e.commentOwnedBy("other")
		_, err = e.svc.Update(ctx, "c", "u", upd)
		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("err = %v, want authz.ErrForbidden", err)
		}
		for _, c := range e.cr.Calls {
			if c == "Update" {
				t.Fatal("update must not run for non-author")
			}
		}
		e.commentOwnedBy("u")
		e.cr.OnUpdate = func(_ context.Context, id string, u map[string]any) (*models.ElementComment, error) {
			return &models.ElementComment{Id: id, Content: u["content"].(string)}, nil
		}
		got, err := e.svc.Update(ctx, "c", "u", upd)
		wantNoErr(t, err)
		if got.Content != "x" {
			t.Fatalf("got %+v", got)
		}

		// Viewer updating own comment's resolved status must be rejected (requires PermCommentResolve)
		e.pr.OnGetAccess = func(_ context.Context, pid, uid string) (*models.ProjectAccess, error) {
			return &models.ProjectAccess{ProjectId: pid, UserId: uid, Role: models.RoleViewer}, nil
		}
		for _, resolvedKey := range []string{"resolved", "Resolved"} {
			_, err := e.svc.Update(ctx, "c", "u", map[string]any{resolvedKey: true})
			if !errors.Is(err, authz.ErrForbidden) {
				t.Fatalf("viewer updating %q: err = %v, want authz.ErrForbidden", resolvedKey, err)
			}
		}
		// Editor updating resolved succeeds
		e.pr.OnGetAccess = func(_ context.Context, pid, uid string) (*models.ProjectAccess, error) {
			return &models.ProjectAccess{ProjectId: pid, UserId: uid, Role: models.RoleEditor}, nil
		}
		e.cr.OnUpdate = func(_ context.Context, id string, u map[string]any) (*models.ElementComment, error) {
			return &models.ElementComment{Id: id, Resolved: true}, nil
		}
		resolvedComment, err := e.svc.Update(ctx, "c", "u", map[string]any{"resolved": true})
		wantNoErr(t, err)
		if !resolvedComment.Resolved {
			t.Fatal("expected resolved comment")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		e := newElementCommentEnv()
		wantErr(t, e.svc.Delete(ctx, "", "u"), "comment id is required")
		wantErr(t, e.svc.Delete(ctx, "c", ""), "userID is required")
		e.cr.OnGetByID = func(context.Context, string) (*models.ElementComment, error) { return nil, errBoom }
		err := e.svc.Delete(ctx, "c", "u")
		wantErr(t, err, "failed to retrieve comment")
		isBoom(t, err)
		e.cr.OnGetByID = func(context.Context, string) (*models.ElementComment, error) { return nil, nil }
		wantErr(t, e.svc.Delete(ctx, "c", "u"), "comment does not exist")
		e.elementExists()
		e.commentOwnedBy("other")
		if err := e.svc.Delete(ctx, "c", "u"); !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("err = %v, want authz.ErrForbidden", err)
		}
		e.commentOwnedBy("u")
		e.cr.OnDelete = func(context.Context, string) error { return errBoom }
		isBoom(t, e.svc.Delete(ctx, "c", "u"))
		e.cr.OnDelete = func(context.Context, string) error { return nil }
		wantNoErr(t, e.svc.Delete(ctx, "c", "u"))
	})
}

func TestElementCommentToggleCountDeleteByElement(t *testing.T) {
	ctx := context.Background()

	t.Run("Toggle", func(t *testing.T) {
		e := newElementCommentEnv()
		wantErr(t, e.svc.ToggleResolved(ctx, ""), "comment id is required")
		e.cr.OnGetByID = func(context.Context, string) (*models.ElementComment, error) { return nil, errBoom }
		err := e.svc.ToggleResolved(ctx, "c")
		wantErr(t, err, "failed to retrieve comment")
		isBoom(t, err)
		e.commentOwnedBy("a")
		e.cr.OnToggleResolved = func(context.Context, string) (*models.ElementComment, error) { return nil, errBoom }
		isBoom(t, e.svc.ToggleResolved(ctx, "c"))
		e.cr.OnToggleResolved = func(context.Context, string) (*models.ElementComment, error) { return &models.ElementComment{}, nil }
		wantNoErr(t, e.svc.ToggleResolved(ctx, "c"))
	})

	t.Run("Count", func(t *testing.T) {
		e := newElementCommentEnv()
		_, err := e.svc.Count(ctx, "")
		wantErr(t, err, "elementId is required")
		e.elementErr()
		_, err = e.svc.Count(ctx, "e")
		wantErr(t, err, "failed to verify element")
		isBoom(t, err)
		e.elementMissing()
		_, err = e.svc.Count(ctx, "e")
		wantErr(t, err, "element does not exist")
		e.elementExists()
		e.cr.OnCount = func(context.Context, string) (int64, error) { return 7, nil }
		n, err := e.svc.Count(ctx, "e")
		wantNoErr(t, err)
		if n != 7 {
			t.Fatalf("n = %d", n)
		}
	})

	t.Run("CountByProject", func(t *testing.T) {
		e := newElementCommentEnv()
		_, err := e.svc.CountByProject(ctx, "", "u")
		wantErr(t, err, "projectId is required")
		e.projectHidden()
		_, err = e.svc.CountByProject(ctx, "p", "u")
		wantErr(t, err, "failed to verify project")
		e.projectVisible()
		e.cr.OnCountByProject = func(context.Context, string) (int64, error) { return 4, nil }
		n, err := e.svc.CountByProject(ctx, "p", "u")
		wantNoErr(t, err)
		if n != 4 {
			t.Fatalf("n = %d", n)
		}
	})

	t.Run("DeleteByElement", func(t *testing.T) {
		e := newElementCommentEnv()
		wantErr(t, e.svc.DeleteByElement(ctx, ""), "elementId is required")
		e.elementErr()
		err := e.svc.DeleteByElement(ctx, "e")
		wantErr(t, err, "failed to verify element")
		isBoom(t, err)
		e.elementMissing()
		wantErr(t, e.svc.DeleteByElement(ctx, "e"), "element does not exist")
		e.elementExists()
		e.cr.OnDeleteByElement = func(context.Context, string) error { return nil }
		wantNoErr(t, e.svc.DeleteByElement(ctx, "e"))
		wantCalls(t, e.cr.Calls, "DeleteByElement")
	})
}

func TestElementCommentUnsyncedElement(t *testing.T) {
	ctx := context.Background()
	e := newElementCommentEnv()
	e.er.OnGetByID = func(context.Context, string) (*models.Element, error) {
		return nil, repositories.ErrElementNotFound
	}

	got, err := e.svc.ListForUser(ctx, "unsynced-el", "user-1", models.ElementCommentFilter{})
	wantNoErr(t, err)
	if got == nil || len(got) != 0 {
		t.Fatalf("expected non-nil empty comment slice for unsynced element, got %#v", got)
	}
	wantCalls(t, e.cr.Calls)

	_, err = e.svc.Create(ctx, &models.ElementComment{ElementId: "unsynced-el", AuthorId: "user-1", Content: "hi"})
	if !errors.Is(err, authz.ErrNotFound) {
		t.Fatalf("expected authz.ErrNotFound for Create on unsynced element, got %v", err)
	}
}
