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

// ─── mock ElementRepository (reused from snapshot tests, local alias) ──────────
// We need a concrete type that satisfies ElementRepositoryInterface for the
// EventWorkflowHandler constructor. We reuse the same stub pattern.

type mockEWElementRepo struct{}

func (m *mockEWElementRepo) GetElements(ctx context.Context, projectID string, pageID ...string) ([]models.EditorElement, error) {
	return nil, nil
}
func (m *mockEWElementRepo) ReplaceElements(ctx context.Context, projectID string, elements []models.EditorElement) error {
	return nil
}
func (m *mockEWElementRepo) GetElementByID(ctx context.Context, elementID string) (*models.Element, error) {
	return nil, nil
}
func (m *mockEWElementRepo) GetElementsByPageID(ctx context.Context, pageID string) ([]models.Element, error) {
	return nil, nil
}
func (m *mockEWElementRepo) GetElementsByPageIds(ctx context.Context, pageIDs []string) ([]models.EditorElement, error) {
	return nil, nil
}
func (m *mockEWElementRepo) GetChildElements(ctx context.Context, parentID string) ([]models.Element, error) {
	return nil, nil
}
func (m *mockEWElementRepo) GetRootElements(ctx context.Context, projectID string) ([]models.Element, error) {
	return nil, nil
}
func (m *mockEWElementRepo) CreateElement(ctx context.Context, element *models.Element) error {
	return nil
}
func (m *mockEWElementRepo) UpdateElement(ctx context.Context, element *models.Element) error {
	return nil
}
func (m *mockEWElementRepo) UpdateEventWorkflows(ctx context.Context, elementID string, workflows []byte) error {
	return nil
}
func (m *mockEWElementRepo) DeleteElementByID(ctx context.Context, elementID string) error {
	return nil
}
func (m *mockEWElementRepo) DeleteElementsByPageID(ctx context.Context, pageID string) error {
	return nil
}
func (m *mockEWElementRepo) DeleteElementsByProjectID(ctx context.Context, projectID string) error {
	return nil
}
func (m *mockEWElementRepo) CountElementsByProjectID(ctx context.Context, projectID string) (int64, error) {
	return 0, nil
}
func (m *mockEWElementRepo) GetElementWithRelations(ctx context.Context, elementID string) (*models.Element, error) {
	return nil, nil
}
func (m *mockEWElementRepo) GetElementsByIDs(ctx context.Context, elementIDs []string) ([]models.Element, error) {
	return nil, nil
}

// ─── test app factory ─────────────────────────────────────────────────────────

// newEventWorkflowTestApp builds a minimal Fiber app wired to EventWorkflowHandler.
// It injects "test-user-id" into locals so auth middleware is bypassed.
func newEventWorkflowTestApp(
	workflowRepo repositories.EventWorkflowRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
	eewRepo repositories.ElementEventWorkflowRepositoryInterface,
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

	// Inject a fake userId so every handler that calls ValidateUserID / MustUserAndParams succeeds.
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", "test-user-id")
		return c.Next()
	})

	h := handlers.NewEventWorkflowHandler(workflowRepo, projectRepo, &mockEWElementRepo{}, eewRepo)

	app.Post("/workflows", h.CreateEventWorkflow)
	app.Get("/workflows/:id", h.GetEventWorkflowByID)
	app.Get("/projects/:projectid/workflows", h.GetEventWorkflowsByProject)
	app.Put("/workflows/:id", h.UpdateEventWorkflow)
	app.Patch("/workflows/:id/enabled", h.UpdateEventWorkflowEnabled)
	app.Delete("/workflows/:id", h.DeleteEventWorkflow)
	app.Get("/workflows/:id/elements", h.GetEventWorkflowElements)

	return app
}

// newEventWorkflowTestAppNoAuth builds the same app WITHOUT injecting userId.
func newEventWorkflowTestAppNoAuth(
	workflowRepo repositories.EventWorkflowRepositoryInterface,
	projectRepo repositories.ProjectRepositoryInterface,
	eewRepo repositories.ElementEventWorkflowRepositoryInterface,
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

	h := handlers.NewEventWorkflowHandler(workflowRepo, projectRepo, &mockEWElementRepo{}, eewRepo)

	app.Post("/workflows", h.CreateEventWorkflow)
	app.Get("/workflows/:id", h.GetEventWorkflowByID)
	app.Get("/projects/:projectid/workflows", h.GetEventWorkflowsByProject)
	app.Put("/workflows/:id", h.UpdateEventWorkflow)
	app.Patch("/workflows/:id/enabled", h.UpdateEventWorkflowEnabled)
	app.Delete("/workflows/:id", h.DeleteEventWorkflow)
	app.Get("/workflows/:id/elements", h.GetEventWorkflowElements)

	return app
}

// ─── request helpers ──────────────────────────────────────────────────────────

func ewStatusOf(t *testing.T, app *fiber.App, method, path string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp.StatusCode
}

func ewBodyOf(t *testing.T, app *fiber.App, method, path string, body io.Reader, contentType string) (int, []byte) {
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

func makeWorkflow(id, projectID, name string, enabled bool) *models.EventWorkflow {
	now := time.Now()
	return &models.EventWorkflow{
		Id:        id,
		ProjectId: projectID,
		Name:      name,
		Enabled:   enabled,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ─── CreateEventWorkflow ──────────────────────────────────────────────────────

func TestCreateEventWorkflow_Returns201OnSuccess(t *testing.T) {
	// Happy path: caller has project access, name is unique, workflow is created.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		CheckIfWorkflowNameExistsFn: func(_ context.Context, _, _, _ string) (bool, error) {
			return false, nil // name is available
		},
		CreateEventWorkflowFn: func(_ context.Context, wf *models.EventWorkflow) (*models.EventWorkflow, error) {
			wf.Id = "wf-1"
			return wf, nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"projectId":"proj-1","name":"My Workflow"}`)
	status, respBody := ewBodyOf(t, app, "POST", "/workflows", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d – body: %s", status, respBody)
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["name"] != "My Workflow" {
		t.Errorf("name: got %v, want %q", result["name"], "My Workflow")
	}
}

func TestCreateEventWorkflow_Returns403WhenProjectAccessDenied(t *testing.T) {
	// GetProjectWithAccess fails → the handler must return 403.
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, _, _ string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newEventWorkflowTestApp(&testutil.MockEventWorkflowRepository{}, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"projectId":"proj-1","name":"My Workflow"}`)
	status, _ := ewBodyOf(t, app, "POST", "/workflows", body, "application/json")
	if status != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", status)
	}
}

func TestCreateEventWorkflow_Returns400WhenNameAlreadyExists(t *testing.T) {
	// If a workflow with the same name already exists in the project → 400.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		CheckIfWorkflowNameExistsFn: func(_ context.Context, _, _, _ string) (bool, error) {
			return true, nil // name is taken
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"projectId":"proj-1","name":"Duplicate"}`)
	status, _ := ewBodyOf(t, app, "POST", "/workflows", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for duplicate name, got %d", status)
	}
}

func TestCreateEventWorkflow_Returns400WhenProjectIDMissing(t *testing.T) {
	// "projectId" is required; omitting it must fail validation.
	app := newEventWorkflowTestApp(&testutil.MockEventWorkflowRepository{}, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"name":"My Workflow"}`)
	status, _ := ewBodyOf(t, app, "POST", "/workflows", body, "application/json")
	if status == fiber.StatusCreated {
		t.Errorf("expected 4xx for missing projectId, got 201")
	}
}

func TestCreateEventWorkflow_Returns400WhenNameMissing(t *testing.T) {
	// "name" is required; omitting it must fail validation.
	app := newEventWorkflowTestApp(&testutil.MockEventWorkflowRepository{}, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"projectId":"proj-1"}`)
	status, _ := ewBodyOf(t, app, "POST", "/workflows", body, "application/json")
	if status == fiber.StatusCreated {
		t.Errorf("expected 4xx for missing name, got 201")
	}
}

func TestCreateEventWorkflow_Returns500WhenCreateFails(t *testing.T) {
	// An unexpected repository error during creation must yield 500.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		CheckIfWorkflowNameExistsFn: func(_ context.Context, _, _, _ string) (bool, error) {
			return false, nil
		},
		CreateEventWorkflowFn: func(_ context.Context, _ *models.EventWorkflow) (*models.EventWorkflow, error) {
			return nil, errors.New("db write error")
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"projectId":"proj-1","name":"My Workflow"}`)
	status, _ := ewBodyOf(t, app, "POST", "/workflows", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestCreateEventWorkflow_Returns401WhenNoUserID(t *testing.T) {
	app := newEventWorkflowTestAppNoAuth(&testutil.MockEventWorkflowRepository{}, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"projectId":"proj-1","name":"My Workflow"}`)
	status, _ := ewBodyOf(t, app, "POST", "/workflows", body, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

func TestCreateEventWorkflow_DefaultsEnabledToTrue(t *testing.T) {
	// When "enabled" is not provided it must default to true.
	var capturedEnabled bool
	workflowRepo := &testutil.MockEventWorkflowRepository{
		CheckIfWorkflowNameExistsFn: func(_ context.Context, _, _, _ string) (bool, error) {
			return false, nil
		},
		CreateEventWorkflowFn: func(_ context.Context, wf *models.EventWorkflow) (*models.EventWorkflow, error) {
			capturedEnabled = wf.Enabled
			return wf, nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"projectId":"proj-1","name":"My Workflow"}`)
	status, _ := ewBodyOf(t, app, "POST", "/workflows", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if !capturedEnabled {
		t.Error("expected enabled to default to true when not provided")
	}
}

// ─── GetEventWorkflowByID ─────────────────────────────────────────────────────

func TestGetEventWorkflowByID_ReturnsWorkflowWhenFound(t *testing.T) {
	// Happy path: workflow exists and the caller has access to its project.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, id string) (*models.EventWorkflow, error) {
			return makeWorkflow(id, "proj-1", "My Workflow", true), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	status, body := ewBodyOf(t, app, "GET", "/workflows/wf-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["name"] != "My Workflow" {
		t.Errorf("name: got %v, want %q", result["name"], "My Workflow")
	}
}

func TestGetEventWorkflowByID_Returns404WhenWorkflowIsNil(t *testing.T) {
	// When the repository returns (nil, nil) the handler must respond with 404.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, _ string) (*models.EventWorkflow, error) {
			return nil, nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "GET", "/workflows/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetEventWorkflowByID_Returns500OnRepositoryError(t *testing.T) {
	// A non-nil repository error must yield 500.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, _ string) (*models.EventWorkflow, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "GET", "/workflows/wf-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetEventWorkflowByID_Returns403WhenProjectAccessDenied(t *testing.T) {
	// The workflow exists but the caller has no access to its project → 403.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, id string) (*models.EventWorkflow, error) {
			return makeWorkflow(id, "proj-1", "My Workflow", true), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, _, _ string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "GET", "/workflows/wf-1"); code != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", code)
	}
}

func TestGetEventWorkflowByID_Returns401WhenNoUserID(t *testing.T) {
	app := newEventWorkflowTestAppNoAuth(&testutil.MockEventWorkflowRepository{}, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "GET", "/workflows/wf-1"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── GetEventWorkflowsByProject ───────────────────────────────────────────────

func TestGetEventWorkflowsByProject_ReturnsWorkflows(t *testing.T) {
	// The handler gates on project access, then fetches workflows with filters.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowsWithFiltersFn: func(_ context.Context, projectID string, _ *bool, _ string) ([]models.EventWorkflow, error) {
			return []models.EventWorkflow{
				*makeWorkflow("wf-1", projectID, "Alpha", true),
				*makeWorkflow("wf-2", projectID, "Beta", false),
			}, nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	status, body := ewBodyOf(t, app, "GET", "/projects/proj-1/workflows", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["count"] != float64(2) {
		t.Errorf("count: got %v, want 2", result["count"])
	}
	data, ok := result["data"].([]any)
	if !ok || len(data) != 2 {
		t.Errorf("expected 2 workflows in data, got %v", result["data"])
	}
}

func TestGetEventWorkflowsByProject_ReturnsEmptyWhenNone(t *testing.T) {
	// No workflows must return 200 with count:0 and data:[].
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowsWithFiltersFn: func(_ context.Context, _ string, _ *bool, _ string) ([]models.EventWorkflow, error) {
			return []models.EventWorkflow{}, nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	status, body := ewBodyOf(t, app, "GET", "/projects/proj-1/workflows", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if result["count"] != float64(0) {
		t.Errorf("count: got %v, want 0", result["count"])
	}
}

func TestGetEventWorkflowsByProject_Returns403WhenProjectAccessDenied(t *testing.T) {
	// GetProjectWithAccess fails → 403.
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, _, _ string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newEventWorkflowTestApp(&testutil.MockEventWorkflowRepository{}, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "GET", "/projects/proj-1/workflows"); code != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", code)
	}
}

func TestGetEventWorkflowsByProject_Returns500OnRepositoryError(t *testing.T) {
	// A repository failure during the workflows fetch must yield 500.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowsWithFiltersFn: func(_ context.Context, _ string, _ *bool, _ string) ([]models.EventWorkflow, error) {
			return nil, errors.New("db error")
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "GET", "/projects/proj-1/workflows"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetEventWorkflowsByProject_Returns401WhenNoUserID(t *testing.T) {
	app := newEventWorkflowTestAppNoAuth(&testutil.MockEventWorkflowRepository{}, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "GET", "/projects/proj-1/workflows"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── UpdateEventWorkflow ──────────────────────────────────────────────────────

func TestUpdateEventWorkflow_Returns200OnSuccess(t *testing.T) {
	// A valid body with a new name must update the workflow and return 200.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, id string) (*models.EventWorkflow, error) {
			return makeWorkflow(id, "proj-1", "Old Name", true), nil
		},
		CheckIfWorkflowNameExistsFn: func(_ context.Context, _, _, _ string) (bool, error) {
			return false, nil
		},
		UpdateEventWorkflowFn: func(_ context.Context, id string, wf *models.EventWorkflow) (*models.EventWorkflow, error) {
			return wf, nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"name":"New Name"}`)
	status, respBody := ewBodyOf(t, app, "PUT", "/workflows/wf-1", body, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, respBody)
	}
}

func TestUpdateEventWorkflow_Returns404WhenWorkflowIsNil(t *testing.T) {
	// When the repository returns (nil, nil) for the lookup → 404.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, _ string) (*models.EventWorkflow, error) {
			return nil, nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"name":"New Name"}`)
	status, _ := ewBodyOf(t, app, "PUT", "/workflows/ghost", body, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestUpdateEventWorkflow_Returns403WhenProjectAccessDenied(t *testing.T) {
	// Workflow found but the caller has no access to its project → 403.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, id string) (*models.EventWorkflow, error) {
			return makeWorkflow(id, "proj-1", "My Workflow", true), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, _, _ string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"name":"New Name"}`)
	status, _ := ewBodyOf(t, app, "PUT", "/workflows/wf-1", body, "application/json")
	if status != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", status)
	}
}

func TestUpdateEventWorkflow_Returns400WhenNameAlreadyExists(t *testing.T) {
	// If the new name conflicts with an existing workflow in the same project → 400.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, id string) (*models.EventWorkflow, error) {
			return makeWorkflow(id, "proj-1", "Old Name", true), nil
		},
		CheckIfWorkflowNameExistsFn: func(_ context.Context, _, _, _ string) (bool, error) {
			return true, nil // conflict
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"name":"Duplicate Name"}`)
	status, _ := ewBodyOf(t, app, "PUT", "/workflows/wf-1", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for duplicate name, got %d", status)
	}
}

func TestUpdateEventWorkflow_Returns401WhenNoUserID(t *testing.T) {
	app := newEventWorkflowTestAppNoAuth(&testutil.MockEventWorkflowRepository{}, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"name":"New Name"}`)
	status, _ := ewBodyOf(t, app, "PUT", "/workflows/wf-1", body, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

// ─── UpdateEventWorkflowEnabled ───────────────────────────────────────────────

func TestUpdateEventWorkflowEnabled_Returns200OnSuccess(t *testing.T) {
	// Toggling enabled must call UpdateEventWorkflowEnabled and return 200 with
	// the new enabled value.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, id string) (*models.EventWorkflow, error) {
			return makeWorkflow(id, "proj-1", "My Workflow", true), nil
		},
		UpdateEventWorkflowEnabledFn: func(_ context.Context, _ string, _ bool) error {
			return nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"enabled":false}`)
	status, respBody := ewBodyOf(t, app, "PATCH", "/workflows/wf-1/enabled", body, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, respBody)
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v", err)
	}
	if result["enabled"] != false {
		t.Errorf("enabled: got %v, want false", result["enabled"])
	}
}

func TestUpdateEventWorkflowEnabled_Returns404WhenWorkflowIsNil(t *testing.T) {
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, _ string) (*models.EventWorkflow, error) {
			return nil, nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"enabled":false}`)
	status, _ := ewBodyOf(t, app, "PATCH", "/workflows/ghost/enabled", body, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestUpdateEventWorkflowEnabled_Returns400WhenBodyMissing(t *testing.T) {
	// "enabled" is required; an empty body must fail validation.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, id string) (*models.EventWorkflow, error) {
			return makeWorkflow(id, "proj-1", "My Workflow", true), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{}`)
	status, _ := ewBodyOf(t, app, "PATCH", "/workflows/wf-1/enabled", body, "application/json")
	if status == fiber.StatusOK {
		t.Errorf("expected 4xx for missing enabled field, got 200")
	}
}

func TestUpdateEventWorkflowEnabled_Returns401WhenNoUserID(t *testing.T) {
	app := newEventWorkflowTestAppNoAuth(&testutil.MockEventWorkflowRepository{}, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	body := strings.NewReader(`{"enabled":false}`)
	status, _ := ewBodyOf(t, app, "PATCH", "/workflows/wf-1/enabled", body, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

// ─── DeleteEventWorkflow ──────────────────────────────────────────────────────

func TestDeleteEventWorkflow_Returns204OnSuccess(t *testing.T) {
	// A successful delete must return 204 No Content.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, id string) (*models.EventWorkflow, error) {
			return makeWorkflow(id, "proj-1", "My Workflow", true), nil
		},
		DeleteEventWorkflowFn: func(_ context.Context, _ string) error {
			return nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "DELETE", "/workflows/wf-1"); code != fiber.StatusNoContent {
		t.Errorf("expected 204, got %d", code)
	}
}

func TestDeleteEventWorkflow_Returns404WhenWorkflowIsNil(t *testing.T) {
	// (nil, nil) from the repository → 404.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, _ string) (*models.EventWorkflow, error) {
			return nil, nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "DELETE", "/workflows/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeleteEventWorkflow_Returns403WhenProjectAccessDenied(t *testing.T) {
	// Workflow found but caller has no project access → 403.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, id string) (*models.EventWorkflow, error) {
			return makeWorkflow(id, "proj-1", "My Workflow", true), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, _, _ string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "DELETE", "/workflows/wf-1"); code != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", code)
	}
}

func TestDeleteEventWorkflow_Returns500WhenDeleteFails(t *testing.T) {
	// An unexpected repository error during deletion must yield 500.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, id string) (*models.EventWorkflow, error) {
			return makeWorkflow(id, "proj-1", "My Workflow", true), nil
		},
		DeleteEventWorkflowFn: func(_ context.Context, _ string) error {
			return errors.New("db write error")
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "DELETE", "/workflows/wf-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestDeleteEventWorkflow_Returns401WhenNoUserID(t *testing.T) {
	app := newEventWorkflowTestAppNoAuth(&testutil.MockEventWorkflowRepository{}, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "DELETE", "/workflows/wf-1"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── GetEventWorkflowElements ─────────────────────────────────────────────────

func TestGetEventWorkflowElements_ReturnsElements(t *testing.T) {
	// The handler fetches the workflow, gates on project access, then returns
	// all element-workflow links for the workflow ID.
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, id string) (*models.EventWorkflow, error) {
			return makeWorkflow(id, "proj-1", "My Workflow", true), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	eewRepo := &testutil.MockElementEventWorkflowRepository{
		GetElementEventWorkflowsByWorkflowIDFn: func(_ context.Context, workflowID string) ([]models.ElementEventWorkflow, error) {
			return []models.ElementEventWorkflow{
				{Id: "eew-1", WorkflowId: workflowID, ElementId: "elem-1", EventName: "click"},
			}, nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, eewRepo)

	status, body := ewBodyOf(t, app, "GET", "/workflows/wf-1/elements", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v", err)
	}
	if result["count"] != float64(1) {
		t.Errorf("count: got %v, want 1", result["count"])
	}
}

func TestGetEventWorkflowElements_Returns404WhenWorkflowIsNil(t *testing.T) {
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, _ string) (*models.EventWorkflow, error) {
			return nil, nil
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "GET", "/workflows/ghost/elements"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetEventWorkflowElements_Returns403WhenProjectAccessDenied(t *testing.T) {
	workflowRepo := &testutil.MockEventWorkflowRepository{
		GetEventWorkflowByIDFn: func(_ context.Context, id string) (*models.EventWorkflow, error) {
			return makeWorkflow(id, "proj-1", "My Workflow", true), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, _, _ string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newEventWorkflowTestApp(workflowRepo, projectRepo, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "GET", "/workflows/wf-1/elements"); code != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", code)
	}
}

func TestGetEventWorkflowElements_Returns401WhenNoUserID(t *testing.T) {
	app := newEventWorkflowTestAppNoAuth(&testutil.MockEventWorkflowRepository{}, &testutil.MockProjectRepository{}, &testutil.MockElementEventWorkflowRepository{})

	if code := ewStatusOf(t, app, "GET", "/workflows/wf-1/elements"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}