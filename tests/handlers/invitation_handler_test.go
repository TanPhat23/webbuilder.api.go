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
	"my-go-app/internal/services"

	"github.com/gofiber/fiber/v2"
)

// ─── mock InvitationService ───────────────────────────────────────────────────

// mockInvitationService is a test double for services.InvitationServiceInterface.
// Each method delegates to an optional function hook; when the hook is nil a
// safe zero-value is returned so tests only need to set the hooks they care about.
type mockInvitationService struct {
	CreateInvitationFn            func(ctx context.Context, projectID, email string, role models.CollaboratorRole, invitedBy string) (*models.Invitation, error)
	AcceptInvitationFn            func(ctx context.Context, token, userID string) error
	GetInvitationsByProjectFn     func(ctx context.Context, projectID string) ([]models.Invitation, error)
	GetInvitationByIDFn           func(ctx context.Context, id string) (*models.Invitation, error)
	DeleteInvitationFn            func(ctx context.Context, id string) error
	CheckProjectOwnershipFn       func(ctx context.Context, projectID, userID string) error
	CancelInvitationFn            func(ctx context.Context, id string) error
	UpdateInvitationStatusFn      func(ctx context.Context, id string, status models.InvitationStatus) error
	GetPendingInvitationsByProjectFn func(ctx context.Context, projectID string) ([]models.Invitation, error)
}

// Compile-time check that mockInvitationService satisfies the interface.
var _ services.InvitationServiceInterface = (*mockInvitationService)(nil)

func (m *mockInvitationService) CreateInvitation(ctx context.Context, projectID, email string, role models.CollaboratorRole, invitedBy string) (*models.Invitation, error) {
	if m.CreateInvitationFn != nil {
		return m.CreateInvitationFn(ctx, projectID, email, role, invitedBy)
	}
	return nil, errors.New("CreateInvitation not implemented")
}

func (m *mockInvitationService) AcceptInvitation(ctx context.Context, token, userID string) error {
	if m.AcceptInvitationFn != nil {
		return m.AcceptInvitationFn(ctx, token, userID)
	}
	return nil
}

func (m *mockInvitationService) GetInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	if m.GetInvitationsByProjectFn != nil {
		return m.GetInvitationsByProjectFn(ctx, projectID)
	}
	return []models.Invitation{}, nil
}

func (m *mockInvitationService) GetInvitationByID(ctx context.Context, id string) (*models.Invitation, error) {
	if m.GetInvitationByIDFn != nil {
		return m.GetInvitationByIDFn(ctx, id)
	}
	return nil, repositories.ErrInvitationNotFound
}

func (m *mockInvitationService) DeleteInvitation(ctx context.Context, id string) error {
	if m.DeleteInvitationFn != nil {
		return m.DeleteInvitationFn(ctx, id)
	}
	return nil
}

func (m *mockInvitationService) CheckProjectOwnership(ctx context.Context, projectID, userID string) error {
	if m.CheckProjectOwnershipFn != nil {
		return m.CheckProjectOwnershipFn(ctx, projectID, userID)
	}
	return nil
}

func (m *mockInvitationService) CancelInvitation(ctx context.Context, id string) error {
	if m.CancelInvitationFn != nil {
		return m.CancelInvitationFn(ctx, id)
	}
	return nil
}

func (m *mockInvitationService) UpdateInvitationStatus(ctx context.Context, id string, status models.InvitationStatus) error {
	if m.UpdateInvitationStatusFn != nil {
		return m.UpdateInvitationStatusFn(ctx, id, status)
	}
	return nil
}

func (m *mockInvitationService) GetPendingInvitationsByProject(ctx context.Context, projectID string) ([]models.Invitation, error) {
	if m.GetPendingInvitationsByProjectFn != nil {
		return m.GetPendingInvitationsByProjectFn(ctx, projectID)
	}
	return []models.Invitation{}, nil
}

// ─── test app factory ─────────────────────────────────────────────────────────

// newInvitationTestApp builds a minimal Fiber app wired to InvitationHandler.
// It injects "test-user-id" into locals so auth middleware is bypassed.
func newInvitationTestApp(svc services.InvitationServiceInterface) *fiber.App {
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

	h := handlers.NewInvitationHandler(svc)

	app.Post("/invitations", h.CreateInvitation)
	app.Get("/projects/:projectid/invitations", h.GetInvitationsByProject)
	app.Get("/projects/:projectid/invitations/pending", h.GetPendingInvitationsByProject)
	app.Post("/invitations/accept", h.AcceptInvitation)
	app.Patch("/invitations/:invitationid/cancel", h.CancelInvitation)
	app.Patch("/invitations/:invitationid/status", h.UpdateInvitationStatus)
	app.Delete("/invitations/:invitationid", h.DeleteInvitation)

	return app
}

// newInvitationTestAppNoAuth builds the same app WITHOUT injecting userId, used
// to exercise the 401 Unauthorized paths.
func newInvitationTestAppNoAuth(svc services.InvitationServiceInterface) *fiber.App {
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

	h := handlers.NewInvitationHandler(svc)

	app.Post("/invitations", h.CreateInvitation)
	app.Get("/projects/:projectid/invitations", h.GetInvitationsByProject)
	app.Post("/invitations/accept", h.AcceptInvitation)
	app.Patch("/invitations/:invitationid/cancel", h.CancelInvitation)
	app.Patch("/invitations/:invitationid/status", h.UpdateInvitationStatus)
	app.Delete("/invitations/:invitationid", h.DeleteInvitation)

	return app
}

// ─── request helpers ──────────────────────────────────────────────────────────

// invitationStatusOf fires a request and returns only the HTTP status code.
func invitationStatusOf(t *testing.T, app *fiber.App, method, path string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp.StatusCode
}

// invitationBodyOf fires a request with an optional body and returns status + raw bytes.
func invitationBodyOf(t *testing.T, app *fiber.App, method, path string, body io.Reader, contentType string) (int, []byte) {
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

func makeInvitation(id, projectID, email string, status models.InvitationStatus) *models.Invitation {
	return &models.Invitation{
		Id:        id,
		Email:     email,
		ProjectId: projectID,
		Role:      models.RoleEditor,
		Token:     "token-" + id,
		Status:    status,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}
}

// ─── CreateInvitation ─────────────────────────────────────────────────────────

func TestCreateInvitation_Returns201OnSuccess(t *testing.T) {
	// A valid body must cause the handler to call CreateInvitation on the service
	// and return 201 with the created invitation.
	svc := &mockInvitationService{
		CreateInvitationFn: func(_ context.Context, projectID, email string, role models.CollaboratorRole, invitedBy string) (*models.Invitation, error) {
			return makeInvitation("inv-1", projectID, email, models.InvitationStatusPending), nil
		},
	}
	app := newInvitationTestApp(svc)

	body := strings.NewReader(`{"projectId":"proj-1","email":"alice@example.com","role":"editor"}`)
	status, respBody := invitationBodyOf(t, app, "POST", "/invitations", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d – body: %s", status, respBody)
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["id"] != "inv-1" {
		t.Errorf("id: got %v, want %q", result["id"], "inv-1")
	}
}

func TestCreateInvitation_DefaultsRoleToEditorWhenNotProvided(t *testing.T) {
	// When "role" is omitted from the body the handler must default it to
	// models.RoleEditor before calling the service.
	var capturedRole models.CollaboratorRole
	svc := &mockInvitationService{
		CreateInvitationFn: func(_ context.Context, projectID, email string, role models.CollaboratorRole, _ string) (*models.Invitation, error) {
			capturedRole = role
			return makeInvitation("inv-1", projectID, email, models.InvitationStatusPending), nil
		},
	}
	app := newInvitationTestApp(svc)

	body := strings.NewReader(`{"projectId":"proj-1","email":"alice@example.com"}`)
	status, _ := invitationBodyOf(t, app, "POST", "/invitations", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedRole != models.RoleEditor {
		t.Errorf("default role: got %q, want %q", capturedRole, models.RoleEditor)
	}
}

func TestCreateInvitation_Returns400WhenProjectIDMissing(t *testing.T) {
	// The "projectId" field is required; missing it must yield a 4xx response.
	app := newInvitationTestApp(&mockInvitationService{})

	body := strings.NewReader(`{"email":"alice@example.com"}`)
	status, _ := invitationBodyOf(t, app, "POST", "/invitations", body, "application/json")
	if status == fiber.StatusCreated || status == fiber.StatusOK {
		t.Errorf("expected 4xx for missing projectId, got %d", status)
	}
}

func TestCreateInvitation_Returns400WhenEmailMissing(t *testing.T) {
	// The "email" field is required; missing it must yield a 4xx response.
	app := newInvitationTestApp(&mockInvitationService{})

	body := strings.NewReader(`{"projectId":"proj-1"}`)
	status, _ := invitationBodyOf(t, app, "POST", "/invitations", body, "application/json")
	if status == fiber.StatusCreated || status == fiber.StatusOK {
		t.Errorf("expected 4xx for missing email, got %d", status)
	}
}

func TestCreateInvitation_Returns400WhenEmailIsInvalid(t *testing.T) {
	// The "email" field must be a valid email address; an invalid value must
	// fail validation before the service is called.
	app := newInvitationTestApp(&mockInvitationService{})

	body := strings.NewReader(`{"projectId":"proj-1","email":"not-an-email"}`)
	status, _ := invitationBodyOf(t, app, "POST", "/invitations", body, "application/json")
	if status == fiber.StatusCreated || status == fiber.StatusOK {
		t.Errorf("expected 4xx for invalid email, got %d", status)
	}
}

func TestCreateInvitation_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	// Malformed JSON must fail with 400 before any validation runs.
	app := newInvitationTestApp(&mockInvitationService{})

	body := strings.NewReader(`not-json`)
	status, _ := invitationBodyOf(t, app, "POST", "/invitations", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestCreateInvitation_Returns400WhenServiceFails(t *testing.T) {
	// A service-level error (e.g. duplicate invitation) must be mapped to 400.
	svc := &mockInvitationService{
		CreateInvitationFn: func(_ context.Context, _, _ string, _ models.CollaboratorRole, _ string) (*models.Invitation, error) {
			return nil, errors.New("invitation already sent to this email")
		},
	}
	app := newInvitationTestApp(svc)

	body := strings.NewReader(`{"projectId":"proj-1","email":"alice@example.com","role":"editor"}`)
	status, _ := invitationBodyOf(t, app, "POST", "/invitations", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 when service fails, got %d", status)
	}
}

func TestCreateInvitation_Returns401WhenNoUserID(t *testing.T) {
	// Without a userId local the handler must return 401 Unauthorized.
	app := newInvitationTestAppNoAuth(&mockInvitationService{})

	body := strings.NewReader(`{"projectId":"proj-1","email":"alice@example.com"}`)
	status, _ := invitationBodyOf(t, app, "POST", "/invitations", body, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

func TestCreateInvitation_PassesCorrectUserIDToService(t *testing.T) {
	// The handler must forward the authenticated userId as the "invitedBy" arg.
	var capturedInvitedBy string
	svc := &mockInvitationService{
		CreateInvitationFn: func(_ context.Context, projectID, email string, role models.CollaboratorRole, invitedBy string) (*models.Invitation, error) {
			capturedInvitedBy = invitedBy
			return makeInvitation("inv-1", projectID, email, models.InvitationStatusPending), nil
		},
	}
	app := newInvitationTestApp(svc)

	body := strings.NewReader(`{"projectId":"proj-1","email":"alice@example.com","role":"editor"}`)
	status, _ := invitationBodyOf(t, app, "POST", "/invitations", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedInvitedBy != "test-user-id" {
		t.Errorf("invitedBy: got %q, want %q", capturedInvitedBy, "test-user-id")
	}
}

// ─── GetInvitationsByProject ──────────────────────────────────────────────────

func TestGetInvitationsByProject_ReturnsInvitationsWhenOwner(t *testing.T) {
	// The handler first checks project ownership then fetches invitations.
	// On success it returns 200 with the invitation slice.
	svc := &mockInvitationService{
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error {
			return nil // caller is the owner
		},
		GetInvitationsByProjectFn: func(_ context.Context, projectID string) ([]models.Invitation, error) {
			return []models.Invitation{
				*makeInvitation("inv-1", projectID, "alice@example.com", models.InvitationStatusPending),
				*makeInvitation("inv-2", projectID, "bob@example.com", models.InvitationStatusAccepted),
			}, nil
		},
	}
	app := newInvitationTestApp(svc)

	status, body := invitationBodyOf(t, app, "GET", "/projects/proj-1/invitations", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 invitations, got %d", len(result))
	}
}

func TestGetInvitationsByProject_ReturnsEmptySliceWhenNone(t *testing.T) {
	// An empty invitations list must still return 200 with an empty JSON array.
	svc := &mockInvitationService{
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error { return nil },
		GetInvitationsByProjectFn: func(_ context.Context, _ string) ([]models.Invitation, error) {
			return []models.Invitation{}, nil
		},
	}
	app := newInvitationTestApp(svc)

	status, body := invitationBodyOf(t, app, "GET", "/projects/proj-1/invitations", nil, "")
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

func TestGetInvitationsByProject_Returns403WhenNotOwner(t *testing.T) {
	// CheckProjectOwnership returns an error → the handler must respond 403.
	svc := &mockInvitationService{
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error {
			return errors.New("not the project owner")
		},
	}
	app := newInvitationTestApp(svc)

	if code := invitationStatusOf(t, app, "GET", "/projects/proj-1/invitations"); code != fiber.StatusForbidden {
		t.Errorf("expected 403 when not owner, got %d", code)
	}
}

func TestGetInvitationsByProject_Returns500WhenServiceFails(t *testing.T) {
	// An unexpected error from GetInvitationsByProject must yield 500.
	svc := &mockInvitationService{
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error { return nil },
		GetInvitationsByProjectFn: func(_ context.Context, _ string) ([]models.Invitation, error) {
			return nil, errors.New("db connection lost")
		},
	}
	app := newInvitationTestApp(svc)

	if code := invitationStatusOf(t, app, "GET", "/projects/proj-1/invitations"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetInvitationsByProject_Returns401WhenNoUserID(t *testing.T) {
	app := newInvitationTestAppNoAuth(&mockInvitationService{})

	if code := invitationStatusOf(t, app, "GET", "/projects/proj-1/invitations"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── GetPendingInvitationsByProject ──────────────────────────────────────────

func TestGetPendingInvitationsByProject_ReturnsPendingInvitations(t *testing.T) {
	// The handler should return only pending invitations for project owners.
	svc := &mockInvitationService{
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error { return nil },
		GetPendingInvitationsByProjectFn: func(_ context.Context, projectID string) ([]models.Invitation, error) {
			return []models.Invitation{
				*makeInvitation("inv-1", projectID, "alice@example.com", models.InvitationStatusPending),
			}, nil
		},
	}
	app := newInvitationTestApp(svc)

	status, body := invitationBodyOf(t, app, "GET", "/projects/proj-1/invitations/pending", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 pending invitation, got %d", len(result))
	}
}

func TestGetPendingInvitationsByProject_Returns403WhenNotOwner(t *testing.T) {
	// CheckProjectOwnership returns an error → 403.
	svc := &mockInvitationService{
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error {
			return errors.New("not the project owner")
		},
	}
	app := newInvitationTestApp(svc)

	if code := invitationStatusOf(t, app, "GET", "/projects/proj-1/invitations/pending"); code != fiber.StatusForbidden {
		t.Errorf("expected 403 when not owner, got %d", code)
	}
}

func TestGetPendingInvitationsByProject_Returns500WhenServiceFails(t *testing.T) {
	// An unexpected error from GetPendingInvitationsByProject must yield 500.
	svc := &mockInvitationService{
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error { return nil },
		GetPendingInvitationsByProjectFn: func(_ context.Context, _ string) ([]models.Invitation, error) {
			return nil, errors.New("db error")
		},
	}
	app := newInvitationTestApp(svc)

	if code := invitationStatusOf(t, app, "GET", "/projects/proj-1/invitations/pending"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── AcceptInvitation ─────────────────────────────────────────────────────────

func TestAcceptInvitation_Returns200OnSuccess(t *testing.T) {
	// A valid token in the body must call AcceptInvitation and return 200 with
	// a success message.
	svc := &mockInvitationService{
		AcceptInvitationFn: func(_ context.Context, token, userID string) error {
			return nil
		},
	}
	app := newInvitationTestApp(svc)

	body := strings.NewReader(`{"token":"valid-token"}`)
	status, respBody := invitationBodyOf(t, app, "POST", "/invitations/accept", body, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, respBody)
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["message"] != "Invitation accepted successfully" {
		t.Errorf("message: got %v, want %q", result["message"], "Invitation accepted successfully")
	}
}

func TestAcceptInvitation_Returns400WhenTokenMissing(t *testing.T) {
	// The "token" field is required; missing it must fail validation before the
	// service is called.
	app := newInvitationTestApp(&mockInvitationService{})

	body := strings.NewReader(`{}`)
	status, _ := invitationBodyOf(t, app, "POST", "/invitations/accept", body, "application/json")
	if status == fiber.StatusOK {
		t.Errorf("expected 4xx for missing token, got 200")
	}
}

func TestAcceptInvitation_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	// Malformed JSON must fail with 400.
	app := newInvitationTestApp(&mockInvitationService{})

	body := strings.NewReader(`not-json`)
	status, _ := invitationBodyOf(t, app, "POST", "/invitations/accept", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestAcceptInvitation_Returns400WhenServiceFails(t *testing.T) {
	// A service-level error (e.g. expired token) must be mapped to 400.
	svc := &mockInvitationService{
		AcceptInvitationFn: func(_ context.Context, _, _ string) error {
			return errors.New("invitation has expired")
		},
	}
	app := newInvitationTestApp(svc)

	body := strings.NewReader(`{"token":"expired-token"}`)
	status, _ := invitationBodyOf(t, app, "POST", "/invitations/accept", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 when token is expired, got %d", status)
	}
}

func TestAcceptInvitation_Returns401WhenNoUserID(t *testing.T) {
	// Without a userId local the handler must return 401 Unauthorized.
	app := newInvitationTestAppNoAuth(&mockInvitationService{})

	body := strings.NewReader(`{"token":"valid-token"}`)
	status, _ := invitationBodyOf(t, app, "POST", "/invitations/accept", body, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

func TestAcceptInvitation_PassesCorrectUserIDAndTokenToService(t *testing.T) {
	// The handler must forward both the token from the body and the local userId.
	var (
		capturedToken  string
		capturedUserID string
	)
	svc := &mockInvitationService{
		AcceptInvitationFn: func(_ context.Context, token, userID string) error {
			capturedToken = token
			capturedUserID = userID
			return nil
		},
	}
	app := newInvitationTestApp(svc)

	body := strings.NewReader(`{"token":"abc-token-123"}`)
	status, _ := invitationBodyOf(t, app, "POST", "/invitations/accept", body, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedToken != "abc-token-123" {
		t.Errorf("token: got %q, want %q", capturedToken, "abc-token-123")
	}
	if capturedUserID != "test-user-id" {
		t.Errorf("userID: got %q, want %q", capturedUserID, "test-user-id")
	}
}

// ─── CancelInvitation ─────────────────────────────────────────────────────────

func TestCancelInvitation_Returns200OnSuccess(t *testing.T) {
	// The caller owns the project; cancelling the invitation must return 200.
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, id string) (*models.Invitation, error) {
			return makeInvitation(id, "proj-1", "alice@example.com", models.InvitationStatusPending), nil
		},
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error { return nil },
		CancelInvitationFn:      func(_ context.Context, _ string) error { return nil },
	}
	app := newInvitationTestApp(svc)

	status, respBody := invitationBodyOf(t, app, "PATCH", "/invitations/inv-1/cancel", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, respBody)
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["message"] != "Invitation cancelled successfully" {
		t.Errorf("message: got %v, want %q", result["message"], "Invitation cancelled successfully")
	}
}

func TestCancelInvitation_Returns404WhenInvitationNotFound(t *testing.T) {
	// If GetInvitationByID returns not-found the handler must respond 404.
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, _ string) (*models.Invitation, error) {
			return nil, repositories.ErrInvitationNotFound
		},
	}
	app := newInvitationTestApp(svc)

	if code := invitationStatusOf(t, app, "PATCH", "/invitations/ghost/cancel"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestCancelInvitation_Returns403WhenNotOwner(t *testing.T) {
	// The invitation exists but the caller does not own the project → 403.
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, id string) (*models.Invitation, error) {
			return makeInvitation(id, "proj-1", "alice@example.com", models.InvitationStatusPending), nil
		},
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error {
			return errors.New("not the project owner")
		},
	}
	app := newInvitationTestApp(svc)

	if code := invitationStatusOf(t, app, "PATCH", "/invitations/inv-1/cancel"); code != fiber.StatusForbidden {
		t.Errorf("expected 403 when not owner, got %d", code)
	}
}

func TestCancelInvitation_Returns500WhenCancelFails(t *testing.T) {
	// An unexpected error from CancelInvitation must yield 500.
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, id string) (*models.Invitation, error) {
			return makeInvitation(id, "proj-1", "alice@example.com", models.InvitationStatusPending), nil
		},
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error { return nil },
		CancelInvitationFn: func(_ context.Context, _ string) error {
			return errors.New("db write error")
		},
	}
	app := newInvitationTestApp(svc)

	if code := invitationStatusOf(t, app, "PATCH", "/invitations/inv-1/cancel"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestCancelInvitation_Returns401WhenNoUserID(t *testing.T) {
	app := newInvitationTestAppNoAuth(&mockInvitationService{})

	if code := invitationStatusOf(t, app, "PATCH", "/invitations/inv-1/cancel"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── UpdateInvitationStatus ───────────────────────────────────────────────────

func TestUpdateInvitationStatus_Returns200OnSuccess(t *testing.T) {
	// A valid status value from an owner must update the invitation and return 200.
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, id string) (*models.Invitation, error) {
			return makeInvitation(id, "proj-1", "alice@example.com", models.InvitationStatusPending), nil
		},
		CheckProjectOwnershipFn:  func(_ context.Context, _, _ string) error { return nil },
		UpdateInvitationStatusFn: func(_ context.Context, _ string, _ models.InvitationStatus) error { return nil },
	}
	app := newInvitationTestApp(svc)

	body := strings.NewReader(`{"status":"cancelled"}`)
	status, respBody := invitationBodyOf(t, app, "PATCH", "/invitations/inv-1/status", body, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, respBody)
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["message"] != "Invitation status updated successfully" {
		t.Errorf("message: got %v, want %q", result["message"], "Invitation status updated successfully")
	}
}

func TestUpdateInvitationStatus_Returns400WhenStatusIsMissing(t *testing.T) {
	// The "status" field is required; an empty body must fail validation.
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, id string) (*models.Invitation, error) {
			return makeInvitation(id, "proj-1", "alice@example.com", models.InvitationStatusPending), nil
		},
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error { return nil },
	}
	app := newInvitationTestApp(svc)

	body := strings.NewReader(`{}`)
	status, _ := invitationBodyOf(t, app, "PATCH", "/invitations/inv-1/status", body, "application/json")
	if status == fiber.StatusOK {
		t.Errorf("expected 4xx for missing status, got 200")
	}
}

func TestUpdateInvitationStatus_Returns400WhenStatusValueIsInvalid(t *testing.T) {
	// Only the four defined InvitationStatus values are valid; an unknown value
	// must be rejected by the "oneof" validator tag.
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, id string) (*models.Invitation, error) {
			return makeInvitation(id, "proj-1", "alice@example.com", models.InvitationStatusPending), nil
		},
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error { return nil },
	}
	app := newInvitationTestApp(svc)

	body := strings.NewReader(`{"status":"unknown-status"}`)
	status, _ := invitationBodyOf(t, app, "PATCH", "/invitations/inv-1/status", body, "application/json")
	if status == fiber.StatusOK {
		t.Errorf("expected 4xx for invalid status value, got 200")
	}
}

func TestUpdateInvitationStatus_Returns404WhenInvitationNotFound(t *testing.T) {
	// If GetInvitationByID returns not-found the handler must respond 404.
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, _ string) (*models.Invitation, error) {
			return nil, repositories.ErrInvitationNotFound
		},
	}
	app := newInvitationTestApp(svc)

	body := strings.NewReader(`{"status":"cancelled"}`)
	if code, _ := invitationBodyOf(t, app, "PATCH", "/invitations/ghost/status", body, "application/json"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestUpdateInvitationStatus_Returns403WhenNotOwner(t *testing.T) {
	// If CheckProjectOwnership fails the handler must return 403.
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, id string) (*models.Invitation, error) {
			return makeInvitation(id, "proj-1", "alice@example.com", models.InvitationStatusPending), nil
		},
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error {
			return errors.New("not the project owner")
		},
	}
	app := newInvitationTestApp(svc)

	body := strings.NewReader(`{"status":"cancelled"}`)
	if code, _ := invitationBodyOf(t, app, "PATCH", "/invitations/inv-1/status", body, "application/json"); code != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", code)
	}
}

func TestUpdateInvitationStatus_Returns500WhenUpdateFails(t *testing.T) {
	// An unexpected error from UpdateInvitationStatus must yield 500.
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, id string) (*models.Invitation, error) {
			return makeInvitation(id, "proj-1", "alice@example.com", models.InvitationStatusPending), nil
		},
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error { return nil },
		UpdateInvitationStatusFn: func(_ context.Context, _ string, _ models.InvitationStatus) error {
			return errors.New("db write error")
		},
	}
	app := newInvitationTestApp(svc)

	body := strings.NewReader(`{"status":"cancelled"}`)
	if code, _ := invitationBodyOf(t, app, "PATCH", "/invitations/inv-1/status", body, "application/json"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestUpdateInvitationStatus_Returns401WhenNoUserID(t *testing.T) {
	app := newInvitationTestAppNoAuth(&mockInvitationService{})

	body := strings.NewReader(`{"status":"cancelled"}`)
	if code, _ := invitationBodyOf(t, app, "PATCH", "/invitations/inv-1/status", body, "application/json"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── DeleteInvitation ─────────────────────────────────────────────────────────

func TestDeleteInvitation_Returns204OnSuccess(t *testing.T) {
	// The caller owns the project; deleting the invitation must return 204.
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, id string) (*models.Invitation, error) {
			return makeInvitation(id, "proj-1", "alice@example.com", models.InvitationStatusPending), nil
		},
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error { return nil },
		DeleteInvitationFn:      func(_ context.Context, _ string) error { return nil },
	}
	app := newInvitationTestApp(svc)

	if code := invitationStatusOf(t, app, "DELETE", "/invitations/inv-1"); code != fiber.StatusNoContent {
		t.Errorf("expected 204, got %d", code)
	}
}

func TestDeleteInvitation_Returns404WhenInvitationNotFound(t *testing.T) {
	// If GetInvitationByID returns not-found the handler must respond 404.
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, _ string) (*models.Invitation, error) {
			return nil, repositories.ErrInvitationNotFound
		},
	}
	app := newInvitationTestApp(svc)

	if code := invitationStatusOf(t, app, "DELETE", "/invitations/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeleteInvitation_Returns403WhenNotOwner(t *testing.T) {
	// The invitation exists but the caller is not the project owner → 403.
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, id string) (*models.Invitation, error) {
			return makeInvitation(id, "proj-1", "alice@example.com", models.InvitationStatusPending), nil
		},
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error {
			return errors.New("not the project owner")
		},
	}
	app := newInvitationTestApp(svc)

	if code := invitationStatusOf(t, app, "DELETE", "/invitations/inv-1"); code != fiber.StatusForbidden {
		t.Errorf("expected 403 when not owner, got %d", code)
	}
}

func TestDeleteInvitation_Returns500WhenDeleteFails(t *testing.T) {
	// An unexpected error from DeleteInvitation must yield 500.
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, id string) (*models.Invitation, error) {
			return makeInvitation(id, "proj-1", "alice@example.com", models.InvitationStatusPending), nil
		},
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error { return nil },
		DeleteInvitationFn: func(_ context.Context, _ string) error {
			return errors.New("db write error")
		},
	}
	app := newInvitationTestApp(svc)

	if code := invitationStatusOf(t, app, "DELETE", "/invitations/inv-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestDeleteInvitation_Returns401WhenNoUserID(t *testing.T) {
	// MustUserAndParams checks userId first; missing → 401.
	app := newInvitationTestAppNoAuth(&mockInvitationService{})

	if code := invitationStatusOf(t, app, "DELETE", "/invitations/inv-1"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

func TestDeleteInvitation_CorrectIDPassedToService(t *testing.T) {
	// The handler must forward the invitationid route param to the service
	// unchanged.
	var capturedID string
	svc := &mockInvitationService{
		GetInvitationByIDFn: func(_ context.Context, id string) (*models.Invitation, error) {
			return makeInvitation(id, "proj-1", "alice@example.com", models.InvitationStatusPending), nil
		},
		CheckProjectOwnershipFn: func(_ context.Context, _, _ string) error { return nil },
		DeleteInvitationFn: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	app := newInvitationTestApp(svc)

	if code := invitationStatusOf(t, app, "DELETE", "/invitations/my-inv-id"); code != fiber.StatusNoContent {
		t.Fatalf("expected 204, got %d", code)
	}
	if capturedID != "my-inv-id" {
		t.Errorf("invitationID passed to service: got %q, want %q", capturedID, "my-inv-id")
	}
}