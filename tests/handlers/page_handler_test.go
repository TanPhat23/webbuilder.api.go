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

// newPageTestApp builds a minimal Fiber app wired to PageHandler.
// It injects "test-user-id" into locals so auth middleware is bypassed.
func newPageTestApp(pageRepo repositories.PageRepositoryInterface) *fiber.App {
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

	// Inject a fake userId so every handler that calls MustUserAndParams succeeds.
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", "test-user-id")
		return c.Next()
	})

	h := handlers.NewPageHandler(pageRepo)

	app.Get("/projects/:projectid/pages", h.GetPagesByProjectID)
	app.Get("/projects/:projectid/pages/:pageid", h.GetPageByID)
	app.Post("/projects/:projectid/pages", h.CreatePage)
	app.Patch("/projects/:projectid/pages/:pageid", h.UpdatePage)
	app.Delete("/projects/:projectid/pages/:pageid", h.DeletePage)

	return app
}

// newPageTestAppNoAuth builds the same app WITHOUT injecting userId, used to
// exercise the 401 Unauthorized paths.
func newPageTestAppNoAuth(pageRepo repositories.PageRepositoryInterface) *fiber.App {
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

	h := handlers.NewPageHandler(pageRepo)

	app.Get("/projects/:projectid/pages", h.GetPagesByProjectID)
	app.Get("/projects/:projectid/pages/:pageid", h.GetPageByID)
	app.Post("/projects/:projectid/pages", h.CreatePage)
	app.Patch("/projects/:projectid/pages/:pageid", h.UpdatePage)
	app.Delete("/projects/:projectid/pages/:pageid", h.DeletePage)

	return app
}

// ─── request helpers ──────────────────────────────────────────────────────────

// pageStatusOf fires a request and returns only the HTTP status code.
func pageStatusOf(t *testing.T, app *fiber.App, method, path string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp.StatusCode
}

// pageBodyOf fires a request with an optional body and returns status + raw bytes.
func pageBodyOf(t *testing.T, app *fiber.App, method, path string, body io.Reader, contentType string) (int, []byte) {
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

func makePage(id, projectID, name string) *models.Page {
	now := time.Now()
	return &models.Page{
		Id:        id,
		Name:      name,
		Type:      "page",
		ProjectId: projectID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ─── GetPagesByProjectID ──────────────────────────────────────────────────────

func TestGetPagesByProjectID_ReturnsEmptySliceWhenNone(t *testing.T) {
	// When no pages exist the handler must respond 200 with an empty JSON array.
	repo := &testutil.MockPageRepository{
		GetPagesByProjectIDFn: func(_ context.Context, _ string) ([]models.Page, error) {
			return []models.Page{}, nil
		},
	}
	app := newPageTestApp(repo)

	status, body := pageBodyOf(t, app, "GET", "/projects/proj-1/pages", nil, "")
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

func TestGetPagesByProjectID_ReturnsPagesForProject(t *testing.T) {
	// The repository is called with the route param and the slice is forwarded.
	repo := &testutil.MockPageRepository{
		GetPagesByProjectIDFn: func(_ context.Context, projectID string) ([]models.Page, error) {
			if projectID != "proj-1" {
				return nil, errors.New("unexpected projectID: " + projectID)
			}
			return []models.Page{
				*makePage("page-1", projectID, "Home"),
				*makePage("page-2", projectID, "About"),
			}, nil
		},
	}
	app := newPageTestApp(repo)

	status, body := pageBodyOf(t, app, "GET", "/projects/proj-1/pages", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 pages, got %d", len(result))
	}
}

func TestGetPagesByProjectID_Returns500OnRepositoryError(t *testing.T) {
	// A repository failure must surface as 500 Internal Server Error.
	repo := &testutil.MockPageRepository{
		GetPagesByProjectIDFn: func(_ context.Context, _ string) ([]models.Page, error) {
			return nil, errors.New("db connection lost")
		},
	}
	app := newPageTestApp(repo)

	if code := pageStatusOf(t, app, "GET", "/projects/proj-1/pages"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── GetPageByID ──────────────────────────────────────────────────────────────

func TestGetPageByID_ReturnsPageWhenFound(t *testing.T) {
	// Happy path: repository finds the page; handler serialises it with 200.
	repo := &testutil.MockPageRepository{
		GetPageByIDFn: func(_ context.Context, pageID, projectID string) (*models.Page, error) {
			if pageID == "page-1" && projectID == "proj-1" {
				return makePage("page-1", "proj-1", "Home"), nil
			}
			return nil, repositories.ErrPageNotFound
		},
	}
	app := newPageTestApp(repo)

	status, body := pageBodyOf(t, app, "GET", "/projects/proj-1/pages/page-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["Id"] != "page-1" {
		t.Errorf("Id: got %v, want %q", result["Id"], "page-1")
	}
}

func TestGetPageByID_Returns404WhenNotFound(t *testing.T) {
	// ErrPageNotFound sentinel must translate to a 404 response.
	repo := &testutil.MockPageRepository{
		GetPageByIDFn: func(_ context.Context, _, _ string) (*models.Page, error) {
			return nil, repositories.ErrPageNotFound
		},
	}
	app := newPageTestApp(repo)

	if code := pageStatusOf(t, app, "GET", "/projects/proj-1/pages/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetPageByID_Returns500OnRepositoryError(t *testing.T) {
	// An unexpected repository error must yield 500.
	repo := &testutil.MockPageRepository{
		GetPageByIDFn: func(_ context.Context, _, _ string) (*models.Page, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newPageTestApp(repo)

	if code := pageStatusOf(t, app, "GET", "/projects/proj-1/pages/page-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── CreatePage ───────────────────────────────────────────────────────────────

func TestCreatePage_Returns201OnSuccess(t *testing.T) {
	// A valid body must cause the handler to create the page and return 201.
	repo := &testutil.MockPageRepository{
		CreatePageFn: func(_ context.Context, _ *models.Page) error {
			return nil
		},
	}
	app := newPageTestApp(repo)

	body := strings.NewReader(`{"name":"Home","type":"page"}`)
	status, respBody := pageBodyOf(t, app, "POST", "/projects/proj-1/pages", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d – body: %s", status, respBody)
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["Name"] != "Home" {
		t.Errorf("Name: got %v, want %q", result["Name"], "Home")
	}
	if result["ProjectId"] != "proj-1" {
		t.Errorf("ProjectId: got %v, want %q", result["ProjectId"], "proj-1")
	}
}

func TestCreatePage_Returns400WhenNameMissing(t *testing.T) {
	// The "name" field is required; missing it must yield a 4xx response.
	repo := &testutil.MockPageRepository{}
	app := newPageTestApp(repo)

	body := strings.NewReader(`{"type":"page"}`)
	status, _ := pageBodyOf(t, app, "POST", "/projects/proj-1/pages", body, "application/json")
	if status == fiber.StatusCreated || status == fiber.StatusOK {
		t.Errorf("expected 4xx for missing name, got %d", status)
	}
}

func TestCreatePage_Returns400WhenTypeMissing(t *testing.T) {
	// The "type" field is required; missing it must yield a 4xx response.
	repo := &testutil.MockPageRepository{}
	app := newPageTestApp(repo)

	body := strings.NewReader(`{"name":"Home"}`)
	status, _ := pageBodyOf(t, app, "POST", "/projects/proj-1/pages", body, "application/json")
	if status == fiber.StatusCreated || status == fiber.StatusOK {
		t.Errorf("expected 4xx for missing type, got %d", status)
	}
}

func TestCreatePage_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	// Malformed JSON must cause BodyParser to fail with 400.
	repo := &testutil.MockPageRepository{}
	app := newPageTestApp(repo)

	body := strings.NewReader(`not-json`)
	status, _ := pageBodyOf(t, app, "POST", "/projects/proj-1/pages", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestCreatePage_Returns500OnRepositoryError(t *testing.T) {
	// If CreatePage returns an unexpected error the handler should yield 500.
	repo := &testutil.MockPageRepository{
		CreatePageFn: func(_ context.Context, _ *models.Page) error {
			return errors.New("db write error")
		},
	}
	app := newPageTestApp(repo)

	body := strings.NewReader(`{"name":"Home","type":"page"}`)
	status, _ := pageBodyOf(t, app, "POST", "/projects/proj-1/pages", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestCreatePage_AssignsNewUUIDToCreatedPage(t *testing.T) {
	// The handler must populate the Id field before passing to the repository,
	// so two consecutive creates must yield different IDs.
	var ids []string
	repo := &testutil.MockPageRepository{
		CreatePageFn: func(_ context.Context, page *models.Page) error {
			ids = append(ids, page.Id)
			return nil
		},
	}
	app := newPageTestApp(repo)

	for i := 0; i < 2; i++ {
		b := strings.NewReader(`{"name":"Page","type":"page"}`)
		status, _ := pageBodyOf(t, app, "POST", "/projects/proj-1/pages", b, "application/json")
		if status != fiber.StatusCreated {
			t.Fatalf("iteration %d: expected 201, got %d", i, status)
		}
	}

	if len(ids) != 2 {
		t.Fatalf("expected 2 captured IDs, got %d", len(ids))
	}
	if ids[0] == ids[1] {
		t.Errorf("expected unique IDs for each page, but both are %q", ids[0])
	}
}

// ─── UpdatePage ───────────────────────────────────────────────────────────────

func TestUpdatePage_Returns200OnSuccess(t *testing.T) {
	// A valid PATCH body must update the page fields and return the updated page.
	updatedName := "Updated Home"
	repo := &testutil.MockPageRepository{
		// First call: GetPageByID to verify the page exists.
		GetPageByIDFn: func(_ context.Context, pageID, projectID string) (*models.Page, error) {
			p := makePage(pageID, projectID, "Home")
			p.Name = updatedName
			return p, nil
		},
		// Second call: after UpdatePageFields, handler re-fetches via GetPageByID.
		UpdatePageFieldsFn: func(_ context.Context, _ string, _ map[string]any) error {
			return nil
		},
	}
	app := newPageTestApp(repo)

	body := strings.NewReader(`{"name":"Updated Home"}`)
	status, respBody := pageBodyOf(t, app, "PATCH", "/projects/proj-1/pages/page-1", body, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, respBody)
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["Name"] != updatedName {
		t.Errorf("Name: got %v, want %q", result["Name"], updatedName)
	}
}

func TestUpdatePage_Returns404WhenPageNotFoundOnPreCheck(t *testing.T) {
	// If the initial GetPageByID call returns a not-found error the handler must
	// respond with 404 before even attempting the update.
	repo := &testutil.MockPageRepository{
		GetPageByIDFn: func(_ context.Context, _, _ string) (*models.Page, error) {
			return nil, repositories.ErrPageNotFound
		},
	}
	app := newPageTestApp(repo)

	body := strings.NewReader(`{"name":"Updated Name"}`)
	status, _ := pageBodyOf(t, app, "PATCH", "/projects/proj-1/pages/ghost", body, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestUpdatePage_Returns400WhenBodyIsEmpty(t *testing.T) {
	// An empty JSON object has no recognised fields; RequireUpdates must reject
	// it with 400.
	repo := &testutil.MockPageRepository{
		GetPageByIDFn: func(_ context.Context, pageID, projectID string) (*models.Page, error) {
			return makePage(pageID, projectID, "Home"), nil
		},
	}
	app := newPageTestApp(repo)

	body := strings.NewReader(`{}`)
	status, _ := pageBodyOf(t, app, "PATCH", "/projects/proj-1/pages/page-1", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for empty update body, got %d", status)
	}
}

func TestUpdatePage_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	// Malformed JSON must fail with 400 before the pre-check even fires.
	// Note: the handler calls GetPageByID *first*, so wire it up to succeed.
	repo := &testutil.MockPageRepository{
		GetPageByIDFn: func(_ context.Context, pageID, projectID string) (*models.Page, error) {
			return makePage(pageID, projectID, "Home"), nil
		},
	}
	app := newPageTestApp(repo)

	body := strings.NewReader(`not-json`)
	status, _ := pageBodyOf(t, app, "PATCH", "/projects/proj-1/pages/page-1", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", status)
	}
}

func TestUpdatePage_Returns500WhenUpdateFieldsFails(t *testing.T) {
	// If UpdatePageFields itself errors the handler must yield 500.
	repo := &testutil.MockPageRepository{
		GetPageByIDFn: func(_ context.Context, pageID, projectID string) (*models.Page, error) {
			return makePage(pageID, projectID, "Home"), nil
		},
		UpdatePageFieldsFn: func(_ context.Context, _ string, _ map[string]any) error {
			return errors.New("db write error")
		},
	}
	app := newPageTestApp(repo)

	body := strings.NewReader(`{"name":"New Name"}`)
	status, _ := pageBodyOf(t, app, "PATCH", "/projects/proj-1/pages/page-1", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestUpdatePage_OnlyAllowsKnownColumns(t *testing.T) {
	// A body with only an unknown field (e.g. "projectId") must be treated as a
	// no-op and rejected with 400 by RequireUpdates — the allowlist strips it.
	var updateCalled bool
	repo := &testutil.MockPageRepository{
		GetPageByIDFn: func(_ context.Context, pageID, projectID string) (*models.Page, error) {
			return makePage(pageID, projectID, "Home"), nil
		},
		UpdatePageFieldsFn: func(_ context.Context, _ string, _ map[string]any) error {
			updateCalled = true
			return nil
		},
	}
	app := newPageTestApp(repo)

	body := strings.NewReader(`{"projectId":"hijack"}`)
	status, _ := pageBodyOf(t, app, "PATCH", "/projects/proj-1/pages/page-1", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for unknown-only fields, got %d", status)
	}
	if updateCalled {
		t.Error("UpdatePageFields must not be called when all fields are stripped by the allowlist")
	}
}

// ─── DeletePage ───────────────────────────────────────────────────────────────

func TestDeletePage_Returns204OnSuccess(t *testing.T) {
	// A successful delete must return 204 No Content with no body.
	repo := &testutil.MockPageRepository{
		DeletePageByProjectIDFn: func(_ context.Context, _, _, _ string) error {
			return nil
		},
	}
	app := newPageTestApp(repo)

	if code := pageStatusOf(t, app, "DELETE", "/projects/proj-1/pages/page-1"); code != fiber.StatusNoContent {
		t.Errorf("expected 204, got %d", code)
	}
}

func TestDeletePage_Returns404WhenPageNotFound(t *testing.T) {
	// The ErrPageNotFound sentinel must translate to 404.
	repo := &testutil.MockPageRepository{
		DeletePageByProjectIDFn: func(_ context.Context, _, _, _ string) error {
			return repositories.ErrPageNotFound
		},
	}
	app := newPageTestApp(repo)

	if code := pageStatusOf(t, app, "DELETE", "/projects/proj-1/pages/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeletePage_Returns500OnRepositoryError(t *testing.T) {
	// Any unexpected repository error must produce 500.
	repo := &testutil.MockPageRepository{
		DeletePageByProjectIDFn: func(_ context.Context, _, _, _ string) error {
			return errors.New("db error")
		},
	}
	app := newPageTestApp(repo)

	if code := pageStatusOf(t, app, "DELETE", "/projects/proj-1/pages/page-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestDeletePage_Returns401WhenNoUserID(t *testing.T) {
	// MustUserAndParams validates userId first; missing → 401.
	app := newPageTestAppNoAuth(&testutil.MockPageRepository{})

	if code := pageStatusOf(t, app, "DELETE", "/projects/proj-1/pages/page-1"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

func TestDeletePage_CorrectParamsPassedToRepo(t *testing.T) {
	// Verify that the handler forwards all three identifiers to the repository
	// in the correct order.
	var (
		capturedPageID    string
		capturedProjectID string
		capturedUserID    string
	)
	repo := &testutil.MockPageRepository{
		DeletePageByProjectIDFn: func(_ context.Context, pageID, projectID, userID string) error {
			capturedPageID = pageID
			capturedProjectID = projectID
			capturedUserID = userID
			return nil
		},
	}
	app := newPageTestApp(repo)

	if code := pageStatusOf(t, app, "DELETE", "/projects/my-project/pages/my-page"); code != fiber.StatusNoContent {
		t.Fatalf("expected 204, got %d", code)
	}
	if capturedPageID != "my-page" {
		t.Errorf("pageID: got %q, want %q", capturedPageID, "my-page")
	}
	if capturedProjectID != "my-project" {
		t.Errorf("projectID: got %q, want %q", capturedProjectID, "my-project")
	}
	if capturedUserID != "test-user-id" {
		t.Errorf("userID: got %q, want %q", capturedUserID, "test-user-id")
	}
}