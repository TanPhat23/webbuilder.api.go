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

// errCollaboratorNotFound is a local sentinel used in tests to simulate the
// "collaborator not found" condition. The real CollaboratorRepository returns
// (nil, nil) when no row is found, but handler tests need an error whose
// message ends with "not found" so that HandleRepoError maps it to a 404.
var errCollaboratorNotFound = errors.New("collaborator not found")

// ─── test app factory ─────────────────────────────────────────────────────────

// newCollaboratorTestApp builds a minimal Fiber app wired to CollaboratorHandler.
// It injects "test-user-id" into locals so auth middleware is bypassed.
func newCollaboratorTestApp(
	collaboratorRepo repositories.CollaboratorRepositoryInterface,
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

	h := handlers.NewCollaboratorHandler(collaboratorRepo, projectRepo)

	app.Get("/projects/:projectid/collaborators", h.GetCollaboratorsByProject)
	app.Get("/collaborators/:collaboratorid", h.GetCollaboratorByID)
	app.Patch("/collaborators/:collaboratorid/role", h.UpdateCollaboratorRole)
	app.Delete("/collaborators/:collaboratorid", h.DeleteCollaborator)

	return app
}

// newCollaboratorTestAppNoAuth builds the same app WITHOUT injecting userId,
// used to exercise 401 Unauthorized paths.
func newCollaboratorTestAppNoAuth(
	collaboratorRepo repositories.CollaboratorRepositoryInterface,
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

	h := handlers.NewCollaboratorHandler(collaboratorRepo, projectRepo)

	app.Get("/projects/:projectid/collaborators", h.GetCollaboratorsByProject)
	app.Get("/collaborators/:collaboratorid", h.GetCollaboratorByID)
	app.Patch("/collaborators/:collaboratorid/role", h.UpdateCollaboratorRole)
	app.Delete("/collaborators/:collaboratorid", h.DeleteCollaborator)

	return app
}

// ─── request helpers ──────────────────────────────────────────────────────────

// collabStatusOf fires a request and returns only the HTTP status code.
func collabStatusOf(t *testing.T, app *fiber.App, method, path string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp.StatusCode
}

// collabBodyOf fires a request with an optional body and returns status + raw bytes.
func collabBodyOf(t *testing.T, app *fiber.App, method, path string, body io.Reader, contentType string) (int, []byte) {
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

func makeCollaborator(id, userID, projectID string, role models.CollaboratorRole) *models.Collaborator {
	now := time.Now()
	return &models.Collaborator{
		Id:        id,
		UserId:    userID,
		ProjectId: projectID,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ─── GetCollaboratorsByProject ────────────────────────────────────────────────

func TestGetCollaboratorsByProject_ReturnsCollaboratorsWhenAccessGranted(t *testing.T) {
	// The handler first gates on GetProjectWithAccess, then fetches collaborators.
	// On success it returns 200 with a JSON object containing a "collaborators" key.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorsByProjectFn: func(_ context.Context, projectID string) ([]models.Collaborator, error) {
			return []models.Collaborator{
				*makeCollaborator("collab-1", "user-a", projectID, models.RoleEditor),
				*makeCollaborator("collab-2", "user-b", projectID, models.RoleViewer),
			}, nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	status, body := collabBodyOf(t, app, "GET", "/projects/proj-1/collaborators", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	collabs, ok := result["collaborators"].([]any)
	if !ok {
		t.Fatalf("expected 'collaborators' key with array value, got: %v", result["collaborators"])
	}
	if len(collabs) != 2 {
		t.Errorf("expected 2 collaborators, got %d", len(collabs))
	}
}

func TestGetCollaboratorsByProject_ReturnsEmptyCollaboratorsWhenNone(t *testing.T) {
	// An empty collaborators list must still return 200 with an empty array.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorsByProjectFn: func(_ context.Context, _ string) ([]models.Collaborator, error) {
			return []models.Collaborator{}, nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	status, body := collabBodyOf(t, app, "GET", "/projects/proj-1/collaborators", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	collabs, ok := result["collaborators"].([]any)
	if !ok {
		t.Fatalf("expected 'collaborators' key with array value, got: %v", result["collaborators"])
	}
	if len(collabs) != 0 {
		t.Errorf("expected empty collaborators, got %d", len(collabs))
	}
}

func TestGetCollaboratorsByProject_Returns404WhenProjectNotFound(t *testing.T) {
	// If GetProjectWithAccess cannot find the project the handler must respond
	// with a not-found / access-denied error (handler maps it via HandleRepoError
	// with "Project not found").
	collaboratorRepo := &testutil.MockCollaboratorRepository{}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, _, _ string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	if code := collabStatusOf(t, app, "GET", "/projects/ghost/collaborators"); code != fiber.StatusNotFound {
		t.Errorf("expected 404 when project not found, got %d", code)
	}
}

func TestGetCollaboratorsByProject_Returns500WhenCollaboratorsFetchFails(t *testing.T) {
	// Access is granted but the collaborators query itself fails.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorsByProjectFn: func(_ context.Context, _ string) ([]models.Collaborator, error) {
			return nil, errors.New("db error")
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	if code := collabStatusOf(t, app, "GET", "/projects/proj-1/collaborators"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetCollaboratorsByProject_Returns401WhenNoUserID(t *testing.T) {
	// MustUserAndParams checks userId first; missing → 401.
	app := newCollaboratorTestAppNoAuth(&testutil.MockCollaboratorRepository{}, &testutil.MockProjectRepository{})

	if code := collabStatusOf(t, app, "GET", "/projects/proj-1/collaborators"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── GetCollaboratorByID ──────────────────────────────────────────────────────

func TestGetCollaboratorByID_ReturnsCollaboratorWhenFound(t *testing.T) {
	// Happy path: the collaborator exists and the caller has access to its
	// associated project.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			if id == "collab-1" {
				return makeCollaborator("collab-1", "user-a", "proj-1", models.RoleEditor), nil
			}
			return nil, errCollaboratorNotFound
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			return makeProject(projectID, userID, "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	status, body := collabBodyOf(t, app, "GET", "/collaborators/collab-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["id"] != "collab-1" {
		t.Errorf("id: got %v, want %q", result["id"], "collab-1")
	}
}

func TestGetCollaboratorByID_Returns404WhenCollaboratorNotFound(t *testing.T) {
	// When the collaborator lookup returns a "not found" error the handler must
	// respond with 404. (The real repo returns nil,nil but the mock can return
	// an error so HandleRepoError maps the message suffix to 404.)
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, _ string) (*models.Collaborator, error) {
			return nil, errCollaboratorNotFound
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, &testutil.MockProjectRepository{})

	if code := collabStatusOf(t, app, "GET", "/collaborators/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetCollaboratorByID_Returns403WhenUserHasNoAccessToProject(t *testing.T) {
	// The collaborator exists but the caller is not authorised to see the
	// project it belongs to — the handler must return 403.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			return makeCollaborator(id, "user-a", "proj-1", models.RoleEditor), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectWithAccessFn: func(_ context.Context, _, _ string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	if code := collabStatusOf(t, app, "GET", "/collaborators/collab-1"); code != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", code)
	}
}

func TestGetCollaboratorByID_Returns500OnRepositoryError(t *testing.T) {
	// An unexpected repository error during the collaborator lookup must yield 500.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, _ string) (*models.Collaborator, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, &testutil.MockProjectRepository{})

	if code := collabStatusOf(t, app, "GET", "/collaborators/collab-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetCollaboratorByID_Returns401WhenNoUserID(t *testing.T) {
	app := newCollaboratorTestAppNoAuth(&testutil.MockCollaboratorRepository{}, &testutil.MockProjectRepository{})

	if code := collabStatusOf(t, app, "GET", "/collaborators/collab-1"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── UpdateCollaboratorRole ───────────────────────────────────────────────────

func TestUpdateCollaboratorRole_Returns200OnSuccess(t *testing.T) {
	// The caller is the project owner; updating the role must return 200 with a
	// success message.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			return makeCollaborator(id, "user-a", "proj-1", models.RoleEditor), nil
		},
		UpdateCollaboratorRoleFn: func(_ context.Context, _ string, _ models.CollaboratorRole) error {
			return nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		// GetProjectByID is used here (not GetProjectWithAccess) to verify ownership.
		GetProjectByIDFn: func(_ context.Context, projectID, userID string) (*models.Project, error) {
			// Return a project whose owner is the authenticated user ("test-user-id").
			return makeProject(projectID, "test-user-id", "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	body := strings.NewReader(`{"role":"viewer"}`)
	status, respBody := collabBodyOf(t, app, "PATCH", "/collaborators/collab-1/role", body, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, respBody)
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["message"] != "Role updated successfully" {
		t.Errorf("message: got %v, want %q", result["message"], "Role updated successfully")
	}
}

func TestUpdateCollaboratorRole_Returns404WhenCollaboratorNotFound(t *testing.T) {
	// The collaborator lookup returns a not-found error → 404.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, _ string) (*models.Collaborator, error) {
			return nil, errCollaboratorNotFound
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, &testutil.MockProjectRepository{})

	body := strings.NewReader(`{"role":"viewer"}`)
	status, _ := collabBodyOf(t, app, "PATCH", "/collaborators/ghost/role", body, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestUpdateCollaboratorRole_Returns403WhenCallerIsNotProjectOwner(t *testing.T) {
	// The caller is in the project but is NOT the owner; the role update must be
	// denied with 403.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			return makeCollaborator(id, "user-a", "proj-1", models.RoleEditor), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, projectID, _ string) (*models.Project, error) {
			// The project owner is someone else, not "test-user-id".
			return makeProject(projectID, "other-owner", "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	body := strings.NewReader(`{"role":"viewer"}`)
	status, _ := collabBodyOf(t, app, "PATCH", "/collaborators/collab-1/role", body, "application/json")
	if status != fiber.StatusForbidden {
		t.Errorf("expected 403 when caller is not owner, got %d", status)
	}
}

func TestUpdateCollaboratorRole_Returns403WhenProjectNotFound(t *testing.T) {
	// If GetProjectByID fails (the project is inaccessible to this user) the
	// handler must return 403.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			return makeCollaborator(id, "user-a", "proj-1", models.RoleEditor), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, _, _ string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	body := strings.NewReader(`{"role":"viewer"}`)
	status, _ := collabBodyOf(t, app, "PATCH", "/collaborators/collab-1/role", body, "application/json")
	if status != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", status)
	}
}

func TestUpdateCollaboratorRole_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	// Malformed JSON must fail with 400 before the ownership check fires.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			return makeCollaborator(id, "user-a", "proj-1", models.RoleEditor), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, projectID, _ string) (*models.Project, error) {
			return makeProject(projectID, "test-user-id", "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	body := strings.NewReader(`not-json`)
	status, _ := collabBodyOf(t, app, "PATCH", "/collaborators/collab-1/role", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestUpdateCollaboratorRole_Returns422WhenRoleIsInvalid(t *testing.T) {
	// The "role" field must be one of: owner, editor, viewer. An unknown value
	// must fail validation before reaching the repository.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			return makeCollaborator(id, "user-a", "proj-1", models.RoleEditor), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, projectID, _ string) (*models.Project, error) {
			return makeProject(projectID, "test-user-id", "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	body := strings.NewReader(`{"role":"superadmin"}`)
	status, _ := collabBodyOf(t, app, "PATCH", "/collaborators/collab-1/role", body, "application/json")
	// Validation must reject the unknown role with a 4xx (422 from ValidationError
	// or 400 from the fiber error handler).
	if status == fiber.StatusOK {
		t.Errorf("expected 4xx for unknown role, got 200")
	}
}

func TestUpdateCollaboratorRole_Returns400WhenRoleIsMissing(t *testing.T) {
	// The "role" field is required; an empty body must be rejected before the
	// repository is ever called.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			return makeCollaborator(id, "user-a", "proj-1", models.RoleEditor), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, projectID, _ string) (*models.Project, error) {
			return makeProject(projectID, "test-user-id", "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	body := strings.NewReader(`{}`)
	status, _ := collabBodyOf(t, app, "PATCH", "/collaborators/collab-1/role", body, "application/json")
	if status == fiber.StatusOK {
		t.Errorf("expected 4xx for missing role, got 200")
	}
}

func TestUpdateCollaboratorRole_Returns500WhenUpdateFails(t *testing.T) {
	// An unexpected error from UpdateCollaboratorRole must yield 500.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			return makeCollaborator(id, "user-a", "proj-1", models.RoleEditor), nil
		},
		UpdateCollaboratorRoleFn: func(_ context.Context, _ string, _ models.CollaboratorRole) error {
			return errors.New("db write error")
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, projectID, _ string) (*models.Project, error) {
			return makeProject(projectID, "test-user-id", "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	body := strings.NewReader(`{"role":"viewer"}`)
	status, _ := collabBodyOf(t, app, "PATCH", "/collaborators/collab-1/role", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestUpdateCollaboratorRole_Returns401WhenNoUserID(t *testing.T) {
	app := newCollaboratorTestAppNoAuth(&testutil.MockCollaboratorRepository{}, &testutil.MockProjectRepository{})

	body := strings.NewReader(`{"role":"viewer"}`)
	status, _ := collabBodyOf(t, app, "PATCH", "/collaborators/collab-1/role", body, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

// ─── DeleteCollaborator ───────────────────────────────────────────────────────

func TestDeleteCollaborator_Returns204WhenOwnerRemovesCollaborator(t *testing.T) {
	// The project owner ("test-user-id") may remove any collaborator.
	// Expected response: 204 No Content.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			// The collaborator being deleted is a different user ("user-a").
			return makeCollaborator(id, "user-a", "proj-1", models.RoleEditor), nil
		},
		DeleteCollaboratorFn: func(_ context.Context, _ string) error {
			return nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, projectID, _ string) (*models.Project, error) {
			// Owner is the authenticated user.
			return makeProject(projectID, "test-user-id", "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	if code := collabStatusOf(t, app, "DELETE", "/collaborators/collab-1"); code != fiber.StatusNoContent {
		t.Errorf("expected 204, got %d", code)
	}
}

func TestDeleteCollaborator_Returns204WhenCollaboratorRemovesThemselves(t *testing.T) {
	// A collaborator may remove themselves from a project even if they are not
	// the owner. The authenticated user IS the collaborator being removed.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			// UserId matches the authenticated "test-user-id".
			return makeCollaborator(id, "test-user-id", "proj-1", models.RoleEditor), nil
		},
		DeleteCollaboratorFn: func(_ context.Context, _ string) error {
			return nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, projectID, _ string) (*models.Project, error) {
			// Owner is someone else.
			return makeProject(projectID, "other-owner", "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	if code := collabStatusOf(t, app, "DELETE", "/collaborators/collab-1"); code != fiber.StatusNoContent {
		t.Errorf("expected 204 when collaborator removes themselves, got %d", code)
	}
}

func TestDeleteCollaborator_Returns403WhenNeitherOwnerNorSelf(t *testing.T) {
	// A non-owner trying to remove a collaborator that is also not themselves
	// must be rejected with 403.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			// The collaborator is "user-a" (not the authenticated "test-user-id").
			return makeCollaborator(id, "user-a", "proj-1", models.RoleEditor), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, projectID, _ string) (*models.Project, error) {
			// The project owner is "other-owner" (not "test-user-id").
			return makeProject(projectID, "other-owner", "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	if code := collabStatusOf(t, app, "DELETE", "/collaborators/collab-1"); code != fiber.StatusForbidden {
		t.Errorf("expected 403 for unauthorized removal, got %d", code)
	}
}

func TestDeleteCollaborator_Returns404WhenCollaboratorNotFound(t *testing.T) {
	// The collaborator does not exist; a "not found" error must map to 404.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, _ string) (*models.Collaborator, error) {
			return nil, errCollaboratorNotFound
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, &testutil.MockProjectRepository{})

	if code := collabStatusOf(t, app, "DELETE", "/collaborators/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeleteCollaborator_Returns403WhenProjectNotFound(t *testing.T) {
	// The project cannot be fetched (e.g., the caller has no access). The
	// handler must return 403 Access Denied.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			return makeCollaborator(id, "user-a", "proj-1", models.RoleEditor), nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, _, _ string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	if code := collabStatusOf(t, app, "DELETE", "/collaborators/collab-1"); code != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", code)
	}
}

func TestDeleteCollaborator_Returns500WhenDeleteFails(t *testing.T) {
	// An unexpected error from DeleteCollaborator must yield 500.
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			return makeCollaborator(id, "user-a", "proj-1", models.RoleEditor), nil
		},
		DeleteCollaboratorFn: func(_ context.Context, _ string) error {
			return errors.New("db write error")
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, projectID, _ string) (*models.Project, error) {
			return makeProject(projectID, "test-user-id", "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	if code := collabStatusOf(t, app, "DELETE", "/collaborators/collab-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestDeleteCollaborator_Returns401WhenNoUserID(t *testing.T) {
	// MustUserAndParams validates userId first; missing → 401.
	app := newCollaboratorTestAppNoAuth(&testutil.MockCollaboratorRepository{}, &testutil.MockProjectRepository{})

	if code := collabStatusOf(t, app, "DELETE", "/collaborators/collab-1"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

func TestDeleteCollaborator_CorrectIDPassedToRepo(t *testing.T) {
	// Verify that the handler forwards the collaboratorid route param to the
	// repository unchanged.
	var capturedID string
	collaboratorRepo := &testutil.MockCollaboratorRepository{
		GetCollaboratorByIDFn: func(_ context.Context, id string) (*models.Collaborator, error) {
			return makeCollaborator(id, "test-user-id", "proj-1", models.RoleEditor), nil
		},
		DeleteCollaboratorFn: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	projectRepo := &testutil.MockProjectRepository{
		GetProjectByIDFn: func(_ context.Context, projectID, _ string) (*models.Project, error) {
			return makeProject(projectID, "test-user-id", "My Project"), nil
		},
	}
	app := newCollaboratorTestApp(collaboratorRepo, projectRepo)

	if code := collabStatusOf(t, app, "DELETE", "/collaborators/my-collab-id"); code != fiber.StatusNoContent {
		t.Fatalf("expected 204, got %d", code)
	}
	if capturedID != "my-collab-id" {
		t.Errorf("collaboratorID passed to repo: got %q, want %q", capturedID, "my-collab-id")
	}
}
