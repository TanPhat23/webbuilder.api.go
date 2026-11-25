package handlers

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/services"
)

type elementCommentSvcStub struct {
	services.ElementCommentServiceInterface
	err      error
	countErr error

	created    *models.ElementComment
	filter     models.ElementCommentFilter
	filterElem string
	updID      string
	updUser    string
	updates    map[string]any
	delID      string
	delUser    string
	toggledID  string
	authorID   string
	limit      int
	offset     int
	projectID  string
	projUser   string
}

func (s *elementCommentSvcStub) Create(_ context.Context, c *models.ElementComment) (*models.ElementComment, error) {
	s.created = c
	return c, s.err
}
func (s *elementCommentSvcStub) GetByID(_ context.Context, id string) (*models.ElementComment, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &models.ElementComment{Id: id, Content: "hi"}, nil
}
func (s *elementCommentSvcStub) List(_ context.Context, elementID string, f models.ElementCommentFilter) ([]models.ElementComment, error) {
	s.filterElem, s.filter = strings.Clone(elementID), f
	if s.err != nil {
		return nil, s.err
	}
	return []models.ElementComment{{Id: "c1", ElementId: elementID}}, nil
}
func (s *elementCommentSvcStub) Count(context.Context, string) (int64, error) {
	if s.countErr != nil {
		return 0, s.countErr
	}
	return 42, nil
}
func (s *elementCommentSvcStub) Update(_ context.Context, id, userID string, u map[string]any) (*models.ElementComment, error) {
	s.updID, s.updUser, s.updates = strings.Clone(id), strings.Clone(userID), u
	if s.err != nil {
		return nil, s.err
	}
	return &models.ElementComment{Id: id, Content: "edited"}, nil
}
func (s *elementCommentSvcStub) Delete(_ context.Context, id, userID string) error {
	s.delID, s.delUser = strings.Clone(id), strings.Clone(userID)
	return s.err
}
func (s *elementCommentSvcStub) GetByIDForUser(_ context.Context, id, _ string) (*models.ElementComment, error) {
	return s.GetByID(context.Background(), id)
}
func (s *elementCommentSvcStub) ListForUser(_ context.Context, elementID, _ string, f models.ElementCommentFilter) ([]models.ElementComment, error) {
	return s.List(context.Background(), elementID, f)
}
func (s *elementCommentSvcStub) ToggleResolved(_ context.Context, id string) error {
	s.toggledID = strings.Clone(id)
	return s.err
}
func (s *elementCommentSvcStub) ToggleResolvedForUser(_ context.Context, id, _ string) error {
	return s.ToggleResolved(context.Background(), id)
}
func (s *elementCommentSvcStub) ListByAuthor(_ context.Context, a string, limit, offset int) ([]models.ElementComment, error) {
	s.authorID, s.limit, s.offset = strings.Clone(a), limit, offset
	if s.err != nil {
		return nil, s.err
	}
	return []models.ElementComment{{Id: "c2", AuthorId: a}}, nil
}
func (s *elementCommentSvcStub) ListByProject(_ context.Context, p, u string, limit, offset int) ([]models.ElementComment, error) {
	s.projectID, s.projUser, s.limit, s.offset = strings.Clone(p), strings.Clone(u), limit, offset
	if s.err != nil {
		return nil, s.err
	}
	return []models.ElementComment{{Id: "c3"}}, nil
}
func (s *elementCommentSvcStub) CountByProject(context.Context, string, string) (int64, error) {
	if s.countErr != nil {
		return 0, s.countErr
	}
	return 7, nil
}

func elementCommentKit(userID string, svc *elementCommentSvcStub) *fiber.App {
	h := NewElementCommentHandler(svc)
	return newKitApp(userID, func(app *fiber.App) {
		app.Post("/element-comments", h.Create)
		app.Get("/element-comments/author/:authorId", h.ListByAuthor)
		app.Get("/element-comments/:id", h.GetByID)
		app.Patch("/element-comments/:id/toggle-resolved", h.ToggleResolved)
		app.Patch("/element-comments/:id", h.Update)
		app.Delete("/element-comments/:id", h.Delete)
		app.Get("/elements/:elementId/comments", h.List)
		app.Get("/projects/:projectId/comments", h.ListByProject)
		app.Get("/noparam/get", h.GetByID)
		app.Get("/noparam/list", h.List)
		app.Patch("/noparam/update", h.Update)
		app.Delete("/noparam/delete", h.Delete)
		app.Patch("/noparam/toggle", h.ToggleResolved)
		app.Get("/noparam/author", h.ListByAuthor)
		app.Get("/noparam/project", h.ListByProject)
	})
}

func TestElementCommentHandlerStatuses(t *testing.T) {
	nf := gorm.ErrRecordNotFound
	cases := []struct {
		name         string
		user         string
		method, path string
		body         string
		err          error
		want         int
		contains     string
	}{
		{"create ok", "u1", "POST", "/element-comments", `{"content":"hello","elementId":"e1"}`, nil, 201, "hello"},
		{"create unauthenticated", "", "POST", "/element-comments", `{"content":"hello","elementId":"e1"}`, nil, 401, "Unauthorized"},
		{"create bad json", "u1", "POST", "/element-comments", `{`, nil, 0, ""},
		{"create empty content", "u1", "POST", "/element-comments", `{"content":"","elementId":"e1"}`, nil, 0, ""},
		{"create missing element", "u1", "POST", "/element-comments", `{"content":"x"}`, nil, 0, ""},
		{"create internal", "u1", "POST", "/element-comments", `{"content":"x","elementId":"e1"}`, errInternal, 500, "Failed to create comment"},

		{"get ok", "u1", "GET", "/element-comments/c1", "", nil, 200, `"id":"c1"`},
		{"get unauthenticated", "", "GET", "/element-comments/c1", "", nil, 401, "Unauthorized"},
		{"get missing param", "u1", "GET", "/noparam/get", "", nil, 400, "id is required"},
		{"get not found", "u1", "GET", "/element-comments/c1", "", nf, 404, "Comment not found"},
		{"get internal", "u1", "GET", "/element-comments/c1", "", errInternal, 500, "Failed to retrieve comment"},

		{"list ok", "u1", "GET", "/elements/e1/comments", "", nil, 200, `"total":42`},
		{"list unauthenticated", "", "GET", "/elements/e1/comments", "", nil, 401, "Unauthorized"},
		{"list missing param", "u1", "GET", "/noparam/list", "", nil, 400, "elementId is required"},
		{"list internal", "u1", "GET", "/elements/e1/comments", "", errInternal, 500, "Failed to retrieve comments"},

		{"update ok", "u1", "PATCH", "/element-comments/c1", `{"content":"edited"}`, nil, 200, "edited"},
		{"update unauthenticated", "", "PATCH", "/element-comments/c1", `{"content":"edited"}`, nil, 401, "Unauthorized"},
		{"update missing param", "u1", "PATCH", "/noparam/update", `{"content":"x"}`, nil, 400, "id is required"},
		{"update bad json", "u1", "PATCH", "/element-comments/c1", `[`, nil, 0, ""},
		{"update empty content", "u1", "PATCH", "/element-comments/c1", `{"content":""}`, nil, 0, ""},
		{"update nothing", "u1", "PATCH", "/element-comments/c1", `{}`, nil, 400, "No valid fields to update"},
		{"update not found", "u1", "PATCH", "/element-comments/c1", `{"resolved":true}`, nf, 404, "Comment not found"},
		{"update internal", "u1", "PATCH", "/element-comments/c1", `{"resolved":true}`, errInternal, 500, "Failed to update comment"},

		{"delete ok", "u1", "DELETE", "/element-comments/c1", "", nil, 204, ""},
		{"delete unauthenticated", "", "DELETE", "/element-comments/c1", "", nil, 401, "Unauthorized"},
		{"delete missing param", "u1", "DELETE", "/noparam/delete", "", nil, 400, "id is required"},
		{"delete internal", "u1", "DELETE", "/element-comments/c1", "", errInternal, 500, "Failed to delete comment"},

		{"toggle ok", "u1", "PATCH", "/element-comments/c1/toggle-resolved", "", nil, 200, "Resolved status toggled successfully"},
		{"toggle unauthenticated", "", "PATCH", "/element-comments/c1/toggle-resolved", "", nil, 401, "Unauthorized"},
		{"toggle missing param", "u1", "PATCH", "/noparam/toggle", "", nil, 400, "id is required"},
		{"toggle not found", "u1", "PATCH", "/element-comments/c1/toggle-resolved", "", nf, 404, "Comment not found"},
		{"toggle internal", "u1", "PATCH", "/element-comments/c1/toggle-resolved", "", errInternal, 500, "Failed to toggle resolved status"},

		{"author ok", "u1", "GET", "/element-comments/author/u1", "", nil, 200, `"authorId":"u1"`},
		{"author unauthenticated", "", "GET", "/element-comments/author/u1", "", nil, 401, "Unauthorized"},
		{"author missing param", "u1", "GET", "/noparam/author", "", nil, 400, "authorId is required"},
		{"author internal", "u1", "GET", "/element-comments/author/u1", "", errInternal, 500, "Failed to retrieve comments"},

		{"project ok", "u1", "GET", "/projects/p1/comments", "", nil, 200, `"total":7`},
		{"project unauthenticated", "", "GET", "/projects/p1/comments", "", nil, 401, "Unauthorized"},
		{"project missing param", "u1", "GET", "/noparam/project", "", nil, 400, "projectId is required"},
		{"project not found", "u1", "GET", "/projects/p1/comments", "", nf, 404, "Project not found"},
		{"project internal", "u1", "GET", "/projects/p1/comments", "", errInternal, 500, "Failed to retrieve comments"},
	}
	for _, tc := range cases {
		r := kitSend(t, elementCommentKit(tc.user, &elementCommentSvcStub{err: tc.err}), tc.method, tc.path, tc.body)
		assertCommentKitResult(t, r, tc.want, tc.contains)
	}
}

func assertCommentKitResult(t *testing.T, r kitReply, want int, contains string) {
	t.Helper()
	if contains == "" {
		kitExpect(t, r, want)
		return
	}
	kitExpect(t, r, want, contains)
}

func TestElementCommentCreateStampsAuthorFromSession(t *testing.T) {
	svc := &elementCommentSvcStub{}
	r := kitSend(t, elementCommentKit("u7", svc), "POST", "/element-comments", `{"content":"hello","elementId":"e1","authorId":"spoofed"}`)
	kitExpect(t, r, 201)
	if svc.created == nil || svc.created.AuthorId != "u7" || svc.created.Id == "" {
		t.Fatalf("created %+v", svc.created)
	}
	if svc.created.ElementId != "e1" || svc.created.Content != "hello" {
		t.Fatalf("created %+v", svc.created)
	}
	if got := r.object(t); got["authorId"] != "u7" {
		t.Fatalf("response %v", got)
	}
}

func TestElementCommentListDefaultsAndPagination(t *testing.T) {
	svc := &elementCommentSvcStub{}
	r := kitSend(t, elementCommentKit("u1", svc), "GET", "/elements/e1/comments", "")
	kitExpect(t, r, 200)
	want := models.ElementCommentFilter{Limit: 20, Offset: 0, SortBy: "CreatedAt", SortOrder: "DESC"}
	if svc.filterElem != "e1" || svc.filter != want {
		t.Fatalf("default filter = %+v (element %q)", svc.filter, svc.filterElem)
	}
	body := r.object(t)
	if body["limit"] != float64(20) || body["offset"] != float64(0) || body["total"] != float64(42) {
		t.Fatalf("body %v", body)
	}
}

func TestElementCommentListHonoursQueryFilters(t *testing.T) {
	svc := &elementCommentSvcStub{}
	kitExpect(t, kitSend(t, elementCommentKit("u1", svc), "GET",
		"/elements/e1/comments?limit=5&offset=10&authorId=a9&resolved=true&sortBy=Content&sortOrder=ASC", ""), 200)
	f := svc.filter
	if f.Limit != 5 || f.Offset != 10 || f.AuthorId != "a9" {
		t.Fatalf("filter paging/author %+v", f)
	}
	if f.Resolved == nil || !*f.Resolved || f.SortBy != "Content" || f.SortOrder != "ASC" {
		t.Fatalf("filter sort/resolved %+v", f)
	}
}

func TestElementCommentListIgnoresInvalidQueryValues(t *testing.T) {
	svc := &elementCommentSvcStub{}
	kitExpect(t, kitSend(t, elementCommentKit("u1", svc), "GET",
		"/elements/e1/comments?limit=0&offset=-3&resolved=maybe&sortOrder=sideways", ""), 200)
	f := svc.filter
	if f.Limit != 20 || f.Offset != 0 || f.Resolved != nil || f.SortOrder != "DESC" {
		t.Fatalf("invalid values must fall back to defaults, got %+v", f)
	}
	kitExpect(t, kitSend(t, elementCommentKit("u1", svc), "GET", "/elements/e1/comments?limit=abc&offset=xyz", ""), 200)
	if svc.filter.Limit != 20 || svc.filter.Offset != 0 {
		t.Fatalf("non-numeric values must fall back to defaults, got %+v", svc.filter)
	}
}

func TestElementCommentListReportsZeroTotalWhenCountFails(t *testing.T) {
	svc := &elementCommentSvcStub{countErr: errInternal}
	r := kitSend(t, elementCommentKit("u1", svc), "GET", "/elements/e1/comments", "")
	kitExpect(t, r, 200)
	if r.object(t)["total"] != float64(0) {
		t.Fatalf("body %s", r.body)
	}
	svc = &elementCommentSvcStub{countErr: errors.New("count failed")}
	r = kitSend(t, elementCommentKit("u1", svc), "GET", "/projects/p1/comments", "")
	kitExpect(t, r, 200)
	if r.object(t)["total"] != float64(0) {
		t.Fatalf("body %s", r.body)
	}
}

func TestElementCommentUpdateForwardsOnlyProvidedFields(t *testing.T) {
	svc := &elementCommentSvcStub{}
	app := elementCommentKit("u3", svc)

	kitExpect(t, kitSend(t, app, "PATCH", "/element-comments/c8", `{"content":"new text"}`), 200)
	assertUpdatedCommentPayload(t, svc, "new text", true)

	kitExpect(t, kitSend(t, app, "PATCH", "/element-comments/c8", `{"resolved":false}`), 200)
	if len(svc.updates) != 1 || svc.updates["resolved"] != false {
		t.Fatalf("updates=%v", svc.updates)
	}

	kitExpect(t, kitSend(t, app, "PATCH", "/element-comments/c8", `{"content":"a","resolved":true}`), 200)
	if len(svc.updates) != 2 || svc.updates["content"] != "a" || svc.updates["resolved"] != true {
		t.Fatalf("updates=%v", svc.updates)
	}
}

func assertUpdatedCommentPayload(t *testing.T, svc *elementCommentSvcStub, wantContent string, single bool) {
	t.Helper()
	if svc.updID != "c8" || svc.updUser != "u3" {
		t.Fatalf("id=%q user=%q", svc.updID, svc.updUser)
	}
	if single && (len(svc.updates) != 1 || svc.updates["content"] != wantContent) {
		t.Fatalf("updates=%v", svc.updates)
	}
}

func TestElementCommentDeleteAndToggleForwardIDs(t *testing.T) {
	svc := &elementCommentSvcStub{}
	app := elementCommentKit("u3", svc)
	kitExpect(t, kitSend(t, app, "DELETE", "/element-comments/c4", ""), 204)
	if svc.delID != "c4" || svc.delUser != "u3" {
		t.Fatalf("delete id=%q user=%q", svc.delID, svc.delUser)
	}
	kitExpect(t, kitSend(t, app, "PATCH", "/element-comments/c4/toggle-resolved", ""), 200)
	if svc.toggledID != "c4" {
		t.Fatalf("toggled %q", svc.toggledID)
	}
}

func TestElementCommentAuthorPagination(t *testing.T) {
	svc := &elementCommentSvcStub{}
	app := elementCommentKit("u3", svc)
	kitExpect(t, kitSend(t, app, "GET", "/element-comments/author/u3?limit=3&offset=6", ""), 200)
	if svc.authorID != "u3" || svc.limit != 3 || svc.offset != 6 {
		t.Fatalf("author=%q limit=%d offset=%d", svc.authorID, svc.limit, svc.offset)
	}
	kitExpect(t, kitSend(t, app, "GET", "/element-comments/author/u3?limit=-1&offset=-1", ""), 200)
	if svc.limit != 20 || svc.offset != 0 {
		t.Fatalf("limit=%d offset=%d", svc.limit, svc.offset)
	}
}

func TestElementCommentProjectPagination(t *testing.T) {
	svc := &elementCommentSvcStub{}
	app := elementCommentKit("u3", svc)
	r := kitSend(t, app, "GET", "/projects/p5/comments?limit=4&offset=8", "")
	kitExpect(t, r, 200)
	if svc.projectID != "p5" || svc.projUser != "u3" || svc.limit != 4 || svc.offset != 8 {
		t.Fatalf("project=%q user=%q limit=%d offset=%d", svc.projectID, svc.projUser, svc.limit, svc.offset)
	}
	if b := r.object(t); b["limit"] != float64(4) || b["offset"] != float64(8) {
		t.Fatalf("body %v", b)
	}
	kitExpect(t, kitSend(t, app, "GET", "/projects/p5/comments?limit=x&offset=y", ""), 200)
	if svc.limit != 20 || svc.offset != 0 {
		t.Fatalf("limit=%d offset=%d", svc.limit, svc.offset)
	}
}
