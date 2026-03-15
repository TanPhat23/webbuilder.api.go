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
	"my-go-app/pkg/utils"
	"my-go-app/tests/testutil"

	"github.com/gofiber/fiber/v2"
)

// ─── test app factories ───────────────────────────────────────────────────────

func newCustomElementTypeTestApp(repo *testutil.MockCustomElementTypeRepository) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			var ve *utils.ValidationError
			if errors.As(err, &ve) {
				return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error(), "fields": ve.Fields})
			}
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", "test-user-id")
		return c.Next()
	})

	h := handlers.NewCustomElementTypeHandler(repo)

	app.Get("/custom-element-types", h.GetCustomElementTypes)
	app.Get("/custom-element-types/:id", h.GetCustomElementTypeByID)
	app.Post("/custom-element-types", h.CreateCustomElementType)
	app.Patch("/custom-element-types/:id", h.UpdateCustomElementType)
	app.Delete("/custom-element-types/:id", h.DeleteCustomElementType)

	return app
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func cetStatusOf(t *testing.T, app *fiber.App, method, path string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp.StatusCode
}

func cetBodyOf(t *testing.T, app *fiber.App, method, path string, body io.Reader, contentType string) (int, []byte) {
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

func makeCustomElementType(id, name string) *models.CustomElementType {
	desc := "A test element type"
	cat := "layout"
	return &models.CustomElementType{
		Id:          id,
		Name:        name,
		Description: &desc,
		Category:    &cat,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ─── GetCustomElementTypes ────────────────────────────────────────────────────

func TestGetCustomElementTypes_ReturnsEmptySliceWhenNone(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		GetCustomElementTypesFn: func(_ context.Context) ([]models.CustomElementType, error) {
			return []models.CustomElementType{}, nil
		},
	}
	app := newCustomElementTypeTestApp(repo)

	status, body := cetBodyOf(t, app, "GET", "/custom-element-types", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(result))
	}
}

func TestGetCustomElementTypes_ReturnsAllTypes(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		GetCustomElementTypesFn: func(_ context.Context) ([]models.CustomElementType, error) {
			return []models.CustomElementType{
				*makeCustomElementType("cet-1", "Button"),
				*makeCustomElementType("cet-2", "Card"),
			}, nil
		},
	}
	app := newCustomElementTypeTestApp(repo)

	status, body := cetBodyOf(t, app, "GET", "/custom-element-types", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 types, got %d", len(result))
	}
}

func TestGetCustomElementTypes_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		GetCustomElementTypesFn: func(_ context.Context) ([]models.CustomElementType, error) {
			return nil, errors.New("db connection lost")
		},
	}
	app := newCustomElementTypeTestApp(repo)

	if code := cetStatusOf(t, app, "GET", "/custom-element-types"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── GetCustomElementTypeByID ─────────────────────────────────────────────────

func TestGetCustomElementTypeByID_ReturnsTypeWhenFound(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		GetCustomElementTypeByIDFn: func(_ context.Context, id string) (*models.CustomElementType, error) {
			if id == "cet-1" {
				return makeCustomElementType(id, "Button"), nil
			}
			return nil, repositories.ErrCustomElementTypeNotFound
		},
	}
	app := newCustomElementTypeTestApp(repo)

	status, body := cetBodyOf(t, app, "GET", "/custom-element-types/cet-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["id"] != "cet-1" {
		t.Errorf("id: got %v, want %q", result["id"], "cet-1")
	}
	if result["name"] != "Button" {
		t.Errorf("name: got %v, want %q", result["name"], "Button")
	}
}

func TestGetCustomElementTypeByID_Returns404WhenNotFound(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		GetCustomElementTypeByIDFn: func(_ context.Context, _ string) (*models.CustomElementType, error) {
			return nil, repositories.ErrCustomElementTypeNotFound
		},
	}
	app := newCustomElementTypeTestApp(repo)

	if code := cetStatusOf(t, app, "GET", "/custom-element-types/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetCustomElementTypeByID_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		GetCustomElementTypeByIDFn: func(_ context.Context, _ string) (*models.CustomElementType, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newCustomElementTypeTestApp(repo)

	if code := cetStatusOf(t, app, "GET", "/custom-element-types/cet-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetCustomElementTypeByID_PassesIDToRepository(t *testing.T) {
	var capturedID string
	repo := &testutil.MockCustomElementTypeRepository{
		GetCustomElementTypeByIDFn: func(_ context.Context, id string) (*models.CustomElementType, error) {
			capturedID = id
			return makeCustomElementType(id, "Button"), nil
		},
	}
	app := newCustomElementTypeTestApp(repo)

	status := cetStatusOf(t, app, "GET", "/custom-element-types/my-type-id")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedID != "my-type-id" {
		t.Errorf("id passed to repo: got %q, want %q", capturedID, "my-type-id")
	}
}

// ─── CreateCustomElementType ──────────────────────────────────────────────────

func TestCreateCustomElementType_Returns201OnSuccess(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		CreateCustomElementTypeFn: func(_ context.Context, cet *models.CustomElementType) (*models.CustomElementType, error) {
			return cet, nil
		},
	}
	app := newCustomElementTypeTestApp(repo)

	payload := strings.NewReader(`{"name":"MyType","description":"desc","category":"ui"}`)
	status, body := cetBodyOf(t, app, "POST", "/custom-element-types", payload, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["name"] != "MyType" {
		t.Errorf("name: got %v, want %q", result["name"], "MyType")
	}
}

func TestCreateCustomElementType_Returns422WhenNameMissing(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{}
	app := newCustomElementTypeTestApp(repo)

	payload := strings.NewReader(`{"description":"desc"}`)
	status, _ := cetBodyOf(t, app, "POST", "/custom-element-types", payload, "application/json")
	if status != fiber.StatusUnprocessableEntity {
		t.Errorf("expected 422 for missing name, got %d", status)
	}
}

func TestCreateCustomElementType_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{}
	app := newCustomElementTypeTestApp(repo)

	payload := strings.NewReader(`{invalid}`)
	status, _ := cetBodyOf(t, app, "POST", "/custom-element-types", payload, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", status)
	}
}

func TestCreateCustomElementType_Returns409WhenNameAlreadyExists(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		CreateCustomElementTypeFn: func(_ context.Context, _ *models.CustomElementType) (*models.CustomElementType, error) {
			return nil, repositories.ErrCustomElementTypeAlreadyExists
		},
	}
	app := newCustomElementTypeTestApp(repo)

	payload := strings.NewReader(`{"name":"Duplicate"}`)
	status, _ := cetBodyOf(t, app, "POST", "/custom-element-types", payload, "application/json")
	if status != fiber.StatusConflict {
		t.Errorf("expected 409 for duplicate name, got %d", status)
	}
}

func TestCreateCustomElementType_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		CreateCustomElementTypeFn: func(_ context.Context, _ *models.CustomElementType) (*models.CustomElementType, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newCustomElementTypeTestApp(repo)

	payload := strings.NewReader(`{"name":"MyType"}`)
	status, _ := cetBodyOf(t, app, "POST", "/custom-element-types", payload, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestCreateCustomElementType_AssignsNewID(t *testing.T) {
	var capturedID string
	repo := &testutil.MockCustomElementTypeRepository{
		CreateCustomElementTypeFn: func(_ context.Context, cet *models.CustomElementType) (*models.CustomElementType, error) {
			capturedID = cet.Id
			return cet, nil
		},
	}
	app := newCustomElementTypeTestApp(repo)

	payload := strings.NewReader(`{"name":"MyType"}`)
	status, _ := cetBodyOf(t, app, "POST", "/custom-element-types", payload, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedID == "" {
		t.Error("expected a non-empty ID to be assigned, got empty string")
	}
}

// ─── UpdateCustomElementType ──────────────────────────────────────────────────

func TestUpdateCustomElementType_Returns200OnSuccess(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		UpdateCustomElementTypeFn: func(_ context.Context, id string, updates map[string]any) (*models.CustomElementType, error) {
			cet := makeCustomElementType(id, updates["name"].(string))
			return cet, nil
		},
	}
	app := newCustomElementTypeTestApp(repo)

	payload := strings.NewReader(`{"name":"UpdatedName"}`)
	status, body := cetBodyOf(t, app, "PATCH", "/custom-element-types/cet-1", payload, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["name"] != "UpdatedName" {
		t.Errorf("name: got %v, want %q", result["name"], "UpdatedName")
	}
}

func TestUpdateCustomElementType_Returns400WhenBodyIsEmpty(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{}
	app := newCustomElementTypeTestApp(repo)

	payload := strings.NewReader(`{}`)
	status, _ := cetBodyOf(t, app, "PATCH", "/custom-element-types/cet-1", payload, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for empty update body, got %d", status)
	}
}

func TestUpdateCustomElementType_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{}
	app := newCustomElementTypeTestApp(repo)

	payload := strings.NewReader(`not-json`)
	status, _ := cetBodyOf(t, app, "PATCH", "/custom-element-types/cet-1", payload, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", status)
	}
}

func TestUpdateCustomElementType_Returns404WhenNotFound(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		UpdateCustomElementTypeFn: func(_ context.Context, _ string, _ map[string]any) (*models.CustomElementType, error) {
			return nil, repositories.ErrCustomElementTypeNotFound
		},
	}
	app := newCustomElementTypeTestApp(repo)

	payload := strings.NewReader(`{"name":"NewName"}`)
	status, _ := cetBodyOf(t, app, "PATCH", "/custom-element-types/ghost", payload, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestUpdateCustomElementType_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		UpdateCustomElementTypeFn: func(_ context.Context, _ string, _ map[string]any) (*models.CustomElementType, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newCustomElementTypeTestApp(repo)

	payload := strings.NewReader(`{"name":"NewName"}`)
	status, _ := cetBodyOf(t, app, "PATCH", "/custom-element-types/cet-1", payload, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestUpdateCustomElementType_OnlyAllowsKnownColumns(t *testing.T) {
	var capturedUpdates map[string]any
	repo := &testutil.MockCustomElementTypeRepository{
		UpdateCustomElementTypeFn: func(_ context.Context, _ string, updates map[string]any) (*models.CustomElementType, error) {
			capturedUpdates = updates
			return makeCustomElementType("cet-1", "Button"), nil
		},
	}
	app := newCustomElementTypeTestApp(repo)

	payload := strings.NewReader(`{"name":"Button","unknownField":"evil","id":"hacked"}`)
	status, _ := cetBodyOf(t, app, "PATCH", "/custom-element-types/cet-1", payload, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if _, ok := capturedUpdates["unknownField"]; ok {
		t.Error("unknown field should not be passed to the repository")
	}
	if _, ok := capturedUpdates["id"]; ok {
		t.Error("immutable field 'id' should not be passed to the repository")
	}
}

func TestUpdateCustomElementType_PassesCorrectIDToRepository(t *testing.T) {
	var capturedID string
	repo := &testutil.MockCustomElementTypeRepository{
		UpdateCustomElementTypeFn: func(_ context.Context, id string, _ map[string]any) (*models.CustomElementType, error) {
			capturedID = id
			return makeCustomElementType(id, "Updated"), nil
		},
	}
	app := newCustomElementTypeTestApp(repo)

	payload := strings.NewReader(`{"name":"Updated"}`)
	status := cetStatusOf(t, app, "PATCH", "/custom-element-types/my-type-id")
	_ = status

	payload = strings.NewReader(`{"name":"Updated"}`)
	cetBodyOf(t, app, "PATCH", "/custom-element-types/my-type-id", payload, "application/json")
	if capturedID != "my-type-id" {
		t.Errorf("id passed to repo: got %q, want %q", capturedID, "my-type-id")
	}
}

// ─── DeleteCustomElementType ──────────────────────────────────────────────────

func TestDeleteCustomElementType_Returns200OnSuccess(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		DeleteCustomElementTypeFn: func(_ context.Context, _ string) error {
			return nil
		},
	}
	app := newCustomElementTypeTestApp(repo)

	status, _ := cetBodyOf(t, app, "DELETE", "/custom-element-types/cet-1", nil, "")
	if status != fiber.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestDeleteCustomElementType_Returns404WhenNotFound(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		DeleteCustomElementTypeFn: func(_ context.Context, _ string) error {
			return repositories.ErrCustomElementTypeNotFound
		},
	}
	app := newCustomElementTypeTestApp(repo)

	if code := cetStatusOf(t, app, "DELETE", "/custom-element-types/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeleteCustomElementType_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockCustomElementTypeRepository{
		DeleteCustomElementTypeFn: func(_ context.Context, _ string) error {
			return errors.New("db error")
		},
	}
	app := newCustomElementTypeTestApp(repo)

	if code := cetStatusOf(t, app, "DELETE", "/custom-element-types/cet-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestDeleteCustomElementType_PassesCorrectIDToRepository(t *testing.T) {
	var capturedID string
	repo := &testutil.MockCustomElementTypeRepository{
		DeleteCustomElementTypeFn: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	app := newCustomElementTypeTestApp(repo)

	if code := cetStatusOf(t, app, "DELETE", "/custom-element-types/target-id"); code != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if capturedID != "target-id" {
		t.Errorf("id passed to repo: got %q, want %q", capturedID, "target-id")
	}
}
