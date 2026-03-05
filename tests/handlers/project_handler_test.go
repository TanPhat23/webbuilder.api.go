package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"my-go-app/internal/handlers"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/tests/testutil"

	"github.com/gofiber/fiber/v2"
)

// ─── test app factory ─────────────────────────────────────────────────────────

// newProjectTestApp builds a minimal Fiber app wired to ProjectHandler.
// It injects "test-user-id" into locals so auth middleware is bypassed.
func newProjectTestApp(projectRepo repositories.ProjectRepositoryInterface) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Inject a fake userId so every handler that calls ValidateUserID succeeds.
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", "test-user-id")
		return c.Next()
	})

	h := handlers.NewProjectHandler(projectRepo)

	app.Get("/projects", h.GetProjectsByUser)
	app.Get("/projects/:projectid", h.GetProjectByID)
	app.Get("/projects/:projectid/public", h.GetPublicProjectByID)
	app.Get("/projects/:projectid/pages", h.GetProjectPages)
	app.Patch("/projects/:projectid", h.UpdateProject)
	app.Delete("/projects/:projectid", h.DeleteProject)

	return app
}

// newProjectTestAppNoAuth builds a Fiber app WITHOUT injecting userId, used to
// test the 401 Unauthorized path.
func newProjectTestAppNoAuth(projectRepo repositories.ProjectRepositoryInterface) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	h := handlers.NewProjectHandler(projectRepo)

	app.Get("/projects", h.GetProjectsByUser)
	app.Get("/projects/:projectid", h.GetProjectByID)
	app.Get("/projects/:projectid/pages", h.GetProjectPages)
	app.Patch("/projects/:projectid", h.UpdateProject)
	app.Delete("/projects/:projectid", h.DeleteProject)

	return app
}

// ─── shared request helpers ───────────────────────────────────────────────────

// projectStatusOf fires a GET and returns only the status code.
func projectStatusOf(t *testing.T, app *fiber.App, method, path string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp.StatusCode
}

// projectBodyOf fires a request and returns status + raw response bytes.
func projectBodyOf(t *testing.T, app *fiber.App, method, path string, body io.Reader, contentType string) (int, []byte) {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("io.ReadAll error: %v", err)
	}
	return resp.StatusCode, b
}

// ─── fixture helpers ──────────────────────────────────────────────────────────

func makeProject(id, ownerID, name string) *models.Project {
	now := time.Now()
	return &models.Project{
		ID:        id,
		Name:      name,
		OwnerId:   ownerID,
		Published: false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ─── GetProjectsByUser ────────────────────────────────────────────────────────

func TestGetProjectsByUser_ReturnsEmptySliceWhenNone(t *testing.T) {
	// When the repository returns an empty slice, the handler should respond
	// with 200 and an empty JSON array.
	repo := &testutil.MockProjectRepository{
		GetProjectsByUserIDFn: func(_ context.Context, _ string) ([]models.Project, error) {
			return []models.Project{}, nil
		},
	}
	app := newProjectTestApp(repo)

	status, body := projectBodyOf(t, app, "GET", "/projects", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d elements", len(result))
	}
}

func TestGetProjectsByUser_ReturnsProjectsForAuthenticatedUser(t *testing.T) {
	// The repository should be called with the local userId and the resulting
	// slice should be forwarded verbatim to the caller.
	repo := &testutil.MockProjectRepository{
		GetProjectsByUserIDFn: func(_ context.Context, userID string) ([]models.Project, error) {
			if userID != "test-user-id" {
				return nil, errors.New("unexpected userID: " + userID)
			}
			return []models.Project{
				*makeProject("proj-1", userID, "My First Project"),
				*makeProject("proj-2", userID, "My Second Project"),
			}, nil
		},
	}
	app := newProjectTestApp(repo)

	status, body := projectBodyOf(t, app, "GET", "/projects", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 projects, got %d", len(result))
	}
}

func TestGetProjectsByUser_Returns500OnRepositoryError(t *testing.T) {
	// A repository error should bubble up as a 500 Internal Server Error.
	repo := &testutil.MockProjectRepository{
		GetProjectsByUserIDFn: func(_ context.Context, _ string) ([]models.Project, error) {
			return nil, errors.New("db connection lost")
		},
	}
	app := newProjectTestApp(repo)

	if code := projectStatusOf(t, app, "GET", "/projects"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetProjectsByUser_Returns401WhenNoUserID(t *testing.T) {
	// Without a userId local the handler must return 401 Unauthorized.
	app := newProjectTestAppNoAuth(&testutil.MockProjectRepository{})

	if code := projectStatusOf(t, app, "GET", "/projects"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── GetProjectByID ───────────────────────────────────────────────────────────

func TestGetProjectByID_ReturnsProjectWhenFound(t *testing.T) {
	// Happy path: the repository finds the project and the handler serialises
	// it with 200.
	repo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			if projectID == "proj-1" && userID == "test-user-id" {
				return makeProject("proj-1", userID, "Test Project"), nil
			}
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newProjectTestApp(repo)

	status, body := projectBodyOf(t, app, "GET", "/projects/proj-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["id"] != "proj-1" {
		t.Errorf("id: got %v, want %q", result["id"], "proj-1")
	}
}

func TestGetProjectByID_Returns404WhenNotFound(t *testing.T) {
	// The sentinel ErrProjectNotFound must translate to a 404.
	repo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, _, _ string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newProjectTestApp(repo)

	if code := projectStatusOf(t, app, "GET", "/projects/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetProjectByID_Returns500OnRepositoryError(t *testing.T) {
	// Any non-sentinel error from the repository should yield 500.
	repo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, _, _ string) (*models.Project, error) {
			return nil, errors.New("unexpected db error")
		},
	}
	app := newProjectTestApp(repo)

	if code := projectStatusOf(t, app, "GET", "/projects/proj-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetProjectByID_Returns401WhenNoUserID(t *testing.T) {
	// MustUserAndParams checks userId first; missing → 401.
	app := newProjectTestAppNoAuth(&testutil.MockProjectRepository{})

	if code := projectStatusOf(t, app, "GET", "/projects/proj-1"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── GetPublicProjectByID ─────────────────────────────────────────────────────

func TestGetPublicProjectByID_ReturnsProjectWhenFound(t *testing.T) {
	// Public endpoint does NOT require auth; it uses GetPublicProjectByID.
	repo := &testutil.MockProjectRepository{
		GetPublicProjectByIDFn: func(_ context.Context, projectID string) (*models.Project, error) {
			if projectID == "pub-1" {
				p := makeProject("pub-1", "owner-id", "Public Project")
				p.Published = true
				return p, nil
			}
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newProjectTestApp(repo)

	status, body := projectBodyOf(t, app, "GET", "/projects/pub-1/public", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["id"] != "pub-1" {
		t.Errorf("id: got %v, want %q", result["id"], "pub-1")
	}
}

func TestGetPublicProjectByID_Returns404WhenNotFound(t *testing.T) {
	repo := &testutil.MockProjectRepository{
		GetPublicProjectByIDFn: func(_ context.Context, _ string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newProjectTestApp(repo)

	if code := projectStatusOf(t, app, "GET", "/projects/ghost/public"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetPublicProjectByID_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockProjectRepository{
		GetPublicProjectByIDFn: func(_ context.Context, _ string) (*models.Project, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newProjectTestApp(repo)

	if code := projectStatusOf(t, app, "GET", "/projects/proj-1/public"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── GetProjectPages ──────────────────────────────────────────────────────────

func TestGetProjectPages_ReturnsPagesWhenAccessGranted(t *testing.T) {
	// The handler first gates on GetProjectWithAccess then fetches pages.
	now := time.Now()
	repo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "Project"), nil
		},
		GetProjectPagesFn: func(_ context.Context, projectID, _ string) ([]models.Page, error) {
			return []models.Page{
				{Id: "page-1", Name: "Home", ProjectId: projectID, Type: "page", CreatedAt: now, UpdatedAt: now},
			}, nil
		},
	}
	app := newProjectTestApp(repo)

	status, body := projectBodyOf(t, app, "GET", "/projects/proj-1/pages", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var pages []map[string]any
	if err := json.Unmarshal(body, &pages); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(pages) != 1 {
		t.Errorf("expected 1 page, got %d", len(pages))
	}
}

func TestGetProjectPages_Returns403WhenAccessDenied(t *testing.T) {
	// If GetProjectWithAccess returns a not-found/access-denied error, the
	// handler should respond with 403 Forbidden (not 404, because access was
	// explicitly denied — see handler implementation).
	repo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, _, _ string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newProjectTestApp(repo)

	if code := projectStatusOf(t, app, "GET", "/projects/proj-x/pages"); code != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", code)
	}
}

func TestGetProjectPages_Returns500WhenPagesFetchFails(t *testing.T) {
	// Access is granted but the pages query itself blows up.
	repo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "Project"), nil
		},
		GetProjectPagesFn: func(_ context.Context, _, _ string) ([]models.Page, error) {
			return nil, errors.New("db error while fetching pages")
		},
	}
	app := newProjectTestApp(repo)

	if code := projectStatusOf(t, app, "GET", "/projects/proj-1/pages"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetProjectPages_Returns401WhenNoUserID(t *testing.T) {
	app := newProjectTestAppNoAuth(&testutil.MockProjectRepository{})

	if code := projectStatusOf(t, app, "GET", "/projects/proj-1/pages"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── UpdateProject ────────────────────────────────────────────────────────────

func TestUpdateProject_Returns200OnSuccess(t *testing.T) {
	// A valid PATCH body with at least one allowed field should update the
	// project and return the updated resource.
	repo := &testutil.MockProjectRepository{
		UpdateProjectFn: func(_ context.Context, projectID, userID string, updates map[string]any) (*models.Project, error) {
			p := makeProject(projectID, userID, "Updated Name")
			return p, nil
		},
	}
	app := newProjectTestApp(repo)

	body := strings.NewReader(`{"name":"Updated Name"}`)
	status, respBody := projectBodyOf(t, app, "PATCH", "/projects/proj-1", body, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, respBody)
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["name"] != "Updated Name" {
		t.Errorf("name: got %v, want %q", result["name"], "Updated Name")
	}
}

func TestUpdateProject_Returns400WhenBodyIsEmpty(t *testing.T) {
	// An empty JSON object has no recognised fields; BuildColumnUpdates will
	// produce an empty map and RequireUpdates will reject it with 400.
	repo := &testutil.MockProjectRepository{}
	app := newProjectTestApp(repo)

	body := strings.NewReader(`{}`)
	status, _ := projectBodyOf(t, app, "PATCH", "/projects/proj-1", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestUpdateProject_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	// Malformed JSON should cause BodyParser to fail → 400.
	repo := &testutil.MockProjectRepository{}
	app := newProjectTestApp(repo)

	body := strings.NewReader(`not-json`)
	status, _ := projectBodyOf(t, app, "PATCH", "/projects/proj-1", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON body, got %d", status)
	}
}

func TestUpdateProject_Returns404WhenProjectNotFound(t *testing.T) {
	// If the repository cannot find the project for this user, 404 is expected.
	repo := &testutil.MockProjectRepository{
		UpdateProjectFn: func(_ context.Context, _, _ string, _ map[string]any) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newProjectTestApp(repo)

	body := strings.NewReader(`{"name":"New Name"}`)
	status, _ := projectBodyOf(t, app, "PATCH", "/projects/ghost", body, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestUpdateProject_Returns500OnRepositoryError(t *testing.T) {
	// A non-sentinel error from UpdateProject must yield 500.
	repo := &testutil.MockProjectRepository{
		UpdateProjectFn: func(_ context.Context, _, _ string, _ map[string]any) (*models.Project, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newProjectTestApp(repo)

	body := strings.NewReader(`{"name":"New Name"}`)
	status, _ := projectBodyOf(t, app, "PATCH", "/projects/proj-1", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestUpdateProject_Returns401WhenNoUserID(t *testing.T) {
	app := newProjectTestAppNoAuth(&testutil.MockProjectRepository{})

	body := strings.NewReader(`{"name":"New Name"}`)
	status, _ := projectBodyOf(t, app, "PATCH", "/projects/proj-1", body, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

func TestUpdateProject_OnlyAllowsKnownColumns(t *testing.T) {
	// A body containing only an unknown field (e.g. "ownerId") must be treated
	// as a no-op and rejected with 400 — the allowlist strips it.
	var capturedUpdates map[string]any
	repo := &testutil.MockProjectRepository{
		UpdateProjectFn: func(_ context.Context, _, _ string, updates map[string]any) (*models.Project, error) {
			capturedUpdates = updates
			return makeProject("proj-1", "test-user-id", "Project"), nil
		},
	}
	app := newProjectTestApp(repo)

	body := strings.NewReader(`{"ownerId":"hacker"}`)
	status, _ := projectBodyOf(t, app, "PATCH", "/projects/proj-1", body, "application/json")

	// "ownerId" is not in projectAllowedCols so RequireUpdates should abort with 400.
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for unknown-only fields, got %d", status)
	}
	if _, ok := capturedUpdates["OwnerId"]; ok {
		t.Error("OwnerId must not appear in column updates (should be stripped by allowlist)")
	}
}

func TestUpdateProject_PublishedBooleanIsAccepted(t *testing.T) {
	// Booleans in the allowed-columns map must be forwarded correctly.
	repo := &testutil.MockProjectRepository{
		UpdateProjectFn: func(_ context.Context, projectID, userID string, updates map[string]any) (*models.Project, error) {
			p := makeProject(projectID, userID, "Project")
			p.Published = true
			return p, nil
		},
	}
	app := newProjectTestApp(repo)

	body := strings.NewReader(`{"published":true}`)
	status, respBody := projectBodyOf(t, app, "PATCH", "/projects/proj-1", body, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, respBody)
	}
}

// ─── DeleteProject ────────────────────────────────────────────────────────────

func TestDeleteProject_Returns204OnSuccess(t *testing.T) {
	// A successful soft-delete should return 204 No Content with no body.
	repo := &testutil.MockProjectRepository{
		DeleteProjectFn: func(_ context.Context, _, _ string) error {
			return nil
		},
	}
	app := newProjectTestApp(repo)

	if code := projectStatusOf(t, app, "DELETE", "/projects/proj-1"); code != fiber.StatusNoContent {
		t.Errorf("expected 204, got %d", code)
	}
}

func TestDeleteProject_Returns404WhenProjectNotFound(t *testing.T) {
	// The sentinel error must map to 404.
	repo := &testutil.MockProjectRepository{
		DeleteProjectFn: func(_ context.Context, _, _ string) error {
			return repositories.ErrProjectNotFound
		},
	}
	app := newProjectTestApp(repo)

	if code := projectStatusOf(t, app, "DELETE", "/projects/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeleteProject_Returns500OnRepositoryError(t *testing.T) {
	// Any unexpected error must produce 500.
	repo := &testutil.MockProjectRepository{
		DeleteProjectFn: func(_ context.Context, _, _ string) error {
			return errors.New("db error")
		},
	}
	app := newProjectTestApp(repo)

	if code := projectStatusOf(t, app, "DELETE", "/projects/proj-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestDeleteProject_Returns401WhenNoUserID(t *testing.T) {
	app := newProjectTestAppNoAuth(&testutil.MockProjectRepository{})

	if code := projectStatusOf(t, app, "DELETE", "/projects/proj-1"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

func TestDeleteProject_CorrectUserAndProjectIDPassedToRepo(t *testing.T) {
	// Verify that the handler forwards the route param and the local userId
	// to the repository unchanged.
	var (
		capturedProjectID string
		capturedUserID    string
	)
	repo := &testutil.MockProjectRepository{
		DeleteProjectFn: func(_ context.Context, projectID, userID string) error {
			capturedProjectID = projectID
			capturedUserID = userID
			return nil
		},
	}
	app := newProjectTestApp(repo)

	if code := projectStatusOf(t, app, "DELETE", "/projects/my-proj-id"); code != fiber.StatusNoContent {
		t.Fatalf("expected 204, got %d", code)
	}
	if capturedProjectID != "my-proj-id" {
		t.Errorf("projectID: got %q, want %q", capturedProjectID, "my-proj-id")
	}
	if capturedUserID != "test-user-id" {
		t.Errorf("userID: got %q, want %q", capturedUserID, "test-user-id")
	}
}