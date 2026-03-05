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

// newSnapshotTestApp builds a minimal Fiber app wired to SnapshotHandler.
// It injects "test-user-id" into locals so auth middleware is bypassed.
// elementRepo is accepted as an interface so callers can pass a nil-safe stub
// for tests that never exercise the element-sync path.
func newSnapshotTestApp(
	snapshotRepo repositories.SnapshotRepositoryInterface,
	elementRepo repositories.ElementRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
) *fiber.App {
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

	h := handlers.NewSnapshotHandler(snapshotRepo, elementRepo, projectRepo)

	app.Post("/projects/:projectid/snapshots", h.SaveSnapshot)
	app.Get("/projects/:projectid/snapshots", h.GetSnapshots)
	app.Get("/snapshots/:snapshotid", h.GetSnapshotByID)
	app.Delete("/projects/:projectid/snapshots/:snapshotid", h.DeleteSnapshot)

	return app
}

// newSnapshotTestAppNoAuth builds the same app WITHOUT injecting userId, used
// to exercise the 401 Unauthorized paths on endpoints that require auth.
func newSnapshotTestAppNoAuth(
	snapshotRepo repositories.SnapshotRepositoryInterface,
	elementRepo repositories.ElementRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
) *fiber.App {
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

	h := handlers.NewSnapshotHandler(snapshotRepo, elementRepo, projectRepo)

	app.Delete("/projects/:projectid/snapshots/:snapshotid", h.DeleteSnapshot)

	return app
}

// ─── request helpers ──────────────────────────────────────────────────────────

// snapshotStatusOf fires a request and returns only the HTTP status code.
func snapshotStatusOf(t *testing.T, app *fiber.App, method, path string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp.StatusCode
}

// snapshotBodyOf fires a request with an optional body and returns status + raw bytes.
func snapshotBodyOf(t *testing.T, app *fiber.App, method, path string, body io.Reader, contentType string) (int, []byte) {
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

// ─── mock ElementRepository stub ─────────────────────────────────────────────

// mockElementRepo is a minimal stub that satisfies ElementRepositoryInterface
// for tests that go through SaveSnapshot (which calls ReplaceElements).
// Only ReplaceElements needs a real implementation; every other method panics
// so a missing stub is caught immediately.
type mockElementRepo struct {
	ReplaceElementsFn func(ctx context.Context, projectID string, elements []models.EditorElement) error
}

func (m *mockElementRepo) ReplaceElements(ctx context.Context, projectID string, elements []models.EditorElement) error {
	if m.ReplaceElementsFn != nil {
		return m.ReplaceElementsFn(ctx, projectID, elements)
	}
	return nil
}

// The remaining methods are stubs that return safe zero values so that the
// interface is satisfied without needing to import the full implementation.
func (m *mockElementRepo) GetElements(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error) {
	return nil, nil
}
func (m *mockElementRepo) GetElementByID(ctx context.Context, elementID string) (*models.Element, error) {
	return nil, nil
}
func (m *mockElementRepo) GetElementsByPageID(ctx context.Context, pageID string) ([]models.Element, error) {
	return nil, nil
}
func (m *mockElementRepo) GetElementsByPageIds(ctx context.Context, pageIDs []string) ([]models.EditorElement, error) {
	return nil, nil
}
func (m *mockElementRepo) GetChildElements(ctx context.Context, parentID string) ([]models.Element, error) {
	return nil, nil
}
func (m *mockElementRepo) GetRootElements(ctx context.Context, pageID string) ([]models.Element, error) {
	return nil, nil
}
func (m *mockElementRepo) CreateElement(ctx context.Context, element *models.Element) error {
	return nil
}
func (m *mockElementRepo) UpdateElement(ctx context.Context, element *models.Element) error {
	return nil
}
func (m *mockElementRepo) UpdateEventWorkflows(ctx context.Context, elementID string, workflows []byte) error {
	return nil
}
func (m *mockElementRepo) DeleteElementByID(ctx context.Context, elementID string) error {
	return nil
}
func (m *mockElementRepo) DeleteElementsByPageID(ctx context.Context, pageID string) error {
	return nil
}
func (m *mockElementRepo) DeleteElementsByProjectID(ctx context.Context, projectID string) error {
	return nil
}
func (m *mockElementRepo) CountElementsByProjectID(ctx context.Context, projectID string) (int64, error) {
	return 0, nil
}
func (m *mockElementRepo) GetElementWithRelations(ctx context.Context, elementID string) (*models.Element, error) {
	return nil, nil
}
func (m *mockElementRepo) GetElementsByIDs(ctx context.Context, elementIDs []string) ([]models.Element, error) {
	return nil, nil
}

// ─── fixture helpers ──────────────────────────────────────────────────────────

func makeSnapshot(id, projectID, name, snapshotType string) *models.Snapshot {
	return &models.Snapshot{
		Id:        id,
		ProjectId: projectID,
		Name:      name,
		Type:      snapshotType,
		Elements:  []byte(`[]`),
		Timestamp: time.Now().UnixMilli(),
	}
}

// ─── GetSnapshots ─────────────────────────────────────────────────────────────

func TestGetSnapshots_ReturnsEmptySliceWhenNone(t *testing.T) {
	// When no snapshots exist the handler must respond 200 with an empty JSON array.
	snapshotRepo := &testutil.MockSnapshotRepository{
		GetSnapshotsByProjectIDFn: func(_ context.Context, _ string) ([]models.Snapshot, error) {
			return []models.Snapshot{}, nil
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	status, body := snapshotBodyOf(t, app, "GET", "/projects/proj-1/snapshots", nil, "")
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

func TestGetSnapshots_ReturnsSnapshotsForProject(t *testing.T) {
	// The repository must be called with the route param and the resulting slice
	// forwarded verbatim to the caller.
	snapshotRepo := &testutil.MockSnapshotRepository{
		GetSnapshotsByProjectIDFn: func(_ context.Context, projectID string) ([]models.Snapshot, error) {
			if projectID != "proj-1" {
				return nil, errors.New("unexpected projectID: " + projectID)
			}
			return []models.Snapshot{
				*makeSnapshot("snap-1", projectID, "v1", "manual"),
				*makeSnapshot("snap-2", projectID, "v2", "working"),
			}, nil
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	status, body := snapshotBodyOf(t, app, "GET", "/projects/proj-1/snapshots", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(result))
	}
}

func TestGetSnapshots_Returns500OnRepositoryError(t *testing.T) {
	// A repository failure must surface as 500 Internal Server Error.
	snapshotRepo := &testutil.MockSnapshotRepository{
		GetSnapshotsByProjectIDFn: func(_ context.Context, _ string) ([]models.Snapshot, error) {
			return nil, errors.New("db connection lost")
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	if code := snapshotStatusOf(t, app, "GET", "/projects/proj-1/snapshots"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── GetSnapshotByID ──────────────────────────────────────────────────────────

func TestGetSnapshotByID_ReturnsSnapshotWhenFound(t *testing.T) {
	// Happy path: repository finds the snapshot; handler serialises it with 200.
	snapshotRepo := &testutil.MockSnapshotRepository{
		GetSnapshotByIDFn: func(_ context.Context, snapshotID string) (*models.Snapshot, error) {
			if snapshotID == "snap-1" {
				return makeSnapshot("snap-1", "proj-1", "v1", "manual"), nil
			}
			return nil, repositories.ErrSnapshotNotFound
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	status, body := snapshotBodyOf(t, app, "GET", "/snapshots/snap-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["id"] != "snap-1" {
		t.Errorf("id: got %v, want %q", result["id"], "snap-1")
	}
}

func TestGetSnapshotByID_Returns404WhenNotFound(t *testing.T) {
	// The ErrSnapshotNotFound sentinel must translate to a 404 response.
	snapshotRepo := &testutil.MockSnapshotRepository{
		GetSnapshotByIDFn: func(_ context.Context, _ string) (*models.Snapshot, error) {
			return nil, repositories.ErrSnapshotNotFound
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	if code := snapshotStatusOf(t, app, "GET", "/snapshots/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetSnapshotByID_Returns500OnRepositoryError(t *testing.T) {
	// An unexpected repository error must yield 500.
	snapshotRepo := &testutil.MockSnapshotRepository{
		GetSnapshotByIDFn: func(_ context.Context, _ string) (*models.Snapshot, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	if code := snapshotStatusOf(t, app, "GET", "/snapshots/snap-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── SaveSnapshot ─────────────────────────────────────────────────────────────

func TestSaveSnapshot_Returns201OnSuccess(t *testing.T) {
	// A valid body must trigger SaveSnapshot + ReplaceElements and return 201.
	snapshotRepo := &testutil.MockSnapshotRepository{
		SaveSnapshotFn: func(_ context.Context, _ string, _ *models.Snapshot) error {
			return nil
		},
	}
	elemRepo := &mockElementRepo{
		ReplaceElementsFn: func(_ context.Context, _ string, _ []models.EditorElement) error {
			return nil
		},
	}
	app := newSnapshotTestApp(snapshotRepo, elemRepo, &testutil.MockProjectRepository{})

	body := strings.NewReader(`{"name":"v1","elements":[]}`)
	status, respBody := snapshotBodyOf(t, app, "POST", "/projects/proj-1/snapshots", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d – body: %s", status, respBody)
	}
}

func TestSaveSnapshot_OmittedElementsDefaultsToEmptyAndSucceeds(t *testing.T) {
	// SaveSnapshot uses ValidateJSONBody (parse-only, no struct validation), so
	// omitting "elements" results in a nil slice that is treated as an empty
	// array — the handler still succeeds and returns 201.
	snapshotRepo := &testutil.MockSnapshotRepository{
		SaveSnapshotFn: func(_ context.Context, _ string, _ *models.Snapshot) error {
			return nil
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	body := strings.NewReader(`{"name":"v1"}`)
	status, _ := snapshotBodyOf(t, app, "POST", "/projects/proj-1/snapshots", body, "application/json")
	if status != fiber.StatusCreated {
		t.Errorf("expected 201 when elements is omitted (nil treated as empty slice), got %d", status)
	}
}

func TestSaveSnapshot_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	// Malformed JSON must cause BodyParser to fail with 400.
	app := newSnapshotTestApp(
		&testutil.MockSnapshotRepository{},
		&mockElementRepo{},
		&testutil.MockProjectRepository{},
	)

	body := strings.NewReader(`not-json`)
	status, _ := snapshotBodyOf(t, app, "POST", "/projects/proj-1/snapshots", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestSaveSnapshot_Returns500WhenSaveSnapshotFails(t *testing.T) {
	// If SaveSnapshot returns an error the handler must yield 500.
	snapshotRepo := &testutil.MockSnapshotRepository{
		SaveSnapshotFn: func(_ context.Context, _ string, _ *models.Snapshot) error {
			return errors.New("db write error")
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	body := strings.NewReader(`{"name":"v1","elements":[]}`)
	status, _ := snapshotBodyOf(t, app, "POST", "/projects/proj-1/snapshots", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestSaveSnapshot_Returns500WhenReplaceElementsFails(t *testing.T) {
	// If ReplaceElements (the element-sync step) fails, the handler must still
	// propagate the error as 500.
	snapshotRepo := &testutil.MockSnapshotRepository{
		SaveSnapshotFn: func(_ context.Context, _ string, _ *models.Snapshot) error {
			return nil
		},
	}
	elemRepo := &mockElementRepo{
		ReplaceElementsFn: func(_ context.Context, _ string, _ []models.EditorElement) error {
			return errors.New("element sync failed")
		},
	}
	app := newSnapshotTestApp(snapshotRepo, elemRepo, &testutil.MockProjectRepository{})

	body := strings.NewReader(`{"name":"v1","elements":[]}`)
	status, _ := snapshotBodyOf(t, app, "POST", "/projects/proj-1/snapshots", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestSaveSnapshot_UsesProvidedIDWhenPresent(t *testing.T) {
	// If the request body already includes an "id", the snapshot must be saved
	// with that exact ID (not a newly generated one).
	var capturedID string
	snapshotRepo := &testutil.MockSnapshotRepository{
		SaveSnapshotFn: func(_ context.Context, _ string, snapshot *models.Snapshot) error {
			capturedID = snapshot.Id
			return nil
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	body := strings.NewReader(`{"id":"my-fixed-id","name":"v1","elements":[]}`)
	status, _ := snapshotBodyOf(t, app, "POST", "/projects/proj-1/snapshots", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedID != "my-fixed-id" {
		t.Errorf("snapshot ID: got %q, want %q", capturedID, "my-fixed-id")
	}
}

func TestSaveSnapshot_GeneratesIDWhenNotProvided(t *testing.T) {
	// If no "id" is supplied in the body the handler must auto-generate one so
	// the repository never receives an empty ID.
	var capturedID string
	snapshotRepo := &testutil.MockSnapshotRepository{
		SaveSnapshotFn: func(_ context.Context, _ string, snapshot *models.Snapshot) error {
			capturedID = snapshot.Id
			return nil
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	body := strings.NewReader(`{"name":"v1","elements":[]}`)
	status, _ := snapshotBodyOf(t, app, "POST", "/projects/proj-1/snapshots", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedID == "" {
		t.Error("expected a non-empty auto-generated snapshot ID")
	}
}

func TestSaveSnapshot_DefaultsTypeToWorking(t *testing.T) {
	// When no "type" is provided the snapshot type must default to "working".
	var capturedType string
	snapshotRepo := &testutil.MockSnapshotRepository{
		SaveSnapshotFn: func(_ context.Context, _ string, snapshot *models.Snapshot) error {
			capturedType = snapshot.Type
			return nil
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	body := strings.NewReader(`{"name":"v1","elements":[]}`)
	status, _ := snapshotBodyOf(t, app, "POST", "/projects/proj-1/snapshots", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedType != "working" {
		t.Errorf("snapshot type: got %q, want %q", capturedType, "working")
	}
}

func TestSaveSnapshot_UsesProvidedType(t *testing.T) {
	// When "type" is explicitly set in the body it must be forwarded unchanged.
	var capturedType string
	snapshotRepo := &testutil.MockSnapshotRepository{
		SaveSnapshotFn: func(_ context.Context, _ string, snapshot *models.Snapshot) error {
			capturedType = snapshot.Type
			return nil
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	body := strings.NewReader(`{"name":"release-1","type":"manual","elements":[]}`)
	status, _ := snapshotBodyOf(t, app, "POST", "/projects/proj-1/snapshots", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedType != "manual" {
		t.Errorf("snapshot type: got %q, want %q", capturedType, "manual")
	}
}

// ─── DeleteSnapshot ───────────────────────────────────────────────────────────

func TestDeleteSnapshot_Returns200OnSuccess(t *testing.T) {
	// A non-working snapshot that belongs to the requested project and whose
	// owner matches the caller must be deleted successfully.
	snapshotRepo := &testutil.MockSnapshotRepository{
		GetSnapshotByIDFn: func(_ context.Context, snapshotID string) (*models.Snapshot, error) {
			return makeSnapshot(snapshotID, "proj-1", "v1", "manual"), nil
		},
		DeleteSnapshotFn: func(_ context.Context, _ string) error {
			return nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, projectID, _ string) (*models.Project, error) {
			return makeProject(projectID, "test-user-id", "My Project"), nil
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, projectRepo)

	status, body := snapshotBodyOf(t, app, "DELETE", "/projects/proj-1/snapshots/snap-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}
}

func TestDeleteSnapshot_Returns404WhenSnapshotNotFound(t *testing.T) {
	// If the initial GetSnapshotByID lookup fails with a not-found sentinel the
	// handler must return 404.
	snapshotRepo := &testutil.MockSnapshotRepository{
		GetSnapshotByIDFn: func(_ context.Context, _ string) (*models.Snapshot, error) {
			return nil, repositories.ErrSnapshotNotFound
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	if code := snapshotStatusOf(t, app, "DELETE", "/projects/proj-1/snapshots/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeleteSnapshot_Returns400WhenSnapshotBelongsToDifferentProject(t *testing.T) {
	// A snapshot whose ProjectId does not match the route's projectid must be
	// rejected with 400 (mismatched ownership guard in the handler).
	snapshotRepo := &testutil.MockSnapshotRepository{
		GetSnapshotByIDFn: func(_ context.Context, snapshotID string) (*models.Snapshot, error) {
			// Snapshot belongs to "proj-other", not "proj-1".
			return makeSnapshot(snapshotID, "proj-other", "v1", "manual"), nil
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	if code := snapshotStatusOf(t, app, "DELETE", "/projects/proj-1/snapshots/snap-1"); code != fiber.StatusBadRequest {
		t.Errorf("expected 400 for mismatched project, got %d", code)
	}
}

func TestDeleteSnapshot_Returns403WhenUserDoesNotOwnProject(t *testing.T) {
	// After the project/snapshot ownership check the handler verifies the caller
	// owns the project via GetProjectByID. A not-found / forbidden error from
	// that call must yield 403.
	snapshotRepo := &testutil.MockSnapshotRepository{
		GetSnapshotByIDFn: func(_ context.Context, snapshotID string) (*models.Snapshot, error) {
			return makeSnapshot(snapshotID, "proj-1", "v1", "manual"), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, _, _ string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, projectRepo)

	if code := snapshotStatusOf(t, app, "DELETE", "/projects/proj-1/snapshots/snap-1"); code != fiber.StatusForbidden {
		t.Errorf("expected 403 when user does not own project, got %d", code)
	}
}

func TestDeleteSnapshot_Returns400WhenTryingToDeleteWorkingSnapshot(t *testing.T) {
	// "working" snapshots are protected; attempting to delete one must yield 400.
	snapshotRepo := &testutil.MockSnapshotRepository{
		GetSnapshotByIDFn: func(_ context.Context, snapshotID string) (*models.Snapshot, error) {
			return makeSnapshot(snapshotID, "proj-1", "autosave", "working"), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, projectRepo)

	if code := snapshotStatusOf(t, app, "DELETE", "/projects/proj-1/snapshots/snap-1"); code != fiber.StatusBadRequest {
		t.Errorf("expected 400 when deleting working snapshot, got %d", code)
	}
}

func TestDeleteSnapshot_Returns500WhenDeleteFails(t *testing.T) {
	// If DeleteSnapshot returns an unexpected error the handler must yield 500.
	snapshotRepo := &testutil.MockSnapshotRepository{
		GetSnapshotByIDFn: func(_ context.Context, snapshotID string) (*models.Snapshot, error) {
			return makeSnapshot(snapshotID, "proj-1", "v1", "manual"), nil
		},
		DeleteSnapshotFn: func(_ context.Context, _ string) error {
			return errors.New("db write error")
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, projectRepo)

	if code := snapshotStatusOf(t, app, "DELETE", "/projects/proj-1/snapshots/snap-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestDeleteSnapshot_Returns401WhenNoUserID(t *testing.T) {
	// MustUserAndParams validates userId first; missing → 401.
	app := newSnapshotTestAppNoAuth(
		&testutil.MockSnapshotRepository{},
		&mockElementRepo{},
		&testutil.MockProjectRepository{},
	)

	if code := snapshotStatusOf(t, app, "DELETE", "/projects/proj-1/snapshots/snap-1"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

func TestDeleteSnapshot_Returns500WhenGetSnapshotFails(t *testing.T) {
	// A non-sentinel error during the initial fetch must yield 500.
	snapshotRepo := &testutil.MockSnapshotRepository{
		GetSnapshotByIDFn: func(_ context.Context, _ string) (*models.Snapshot, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newSnapshotTestApp(snapshotRepo, &mockElementRepo{}, &testutil.MockProjectRepository{})

	if code := snapshotStatusOf(t, app, "DELETE", "/projects/proj-1/snapshots/snap-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}