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

func newCustomElementTestApp(repo *testutil.MockCustomElementRepository) *fiber.App {
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

	h := handlers.NewCustomElementHandler(repo)

	app.Get("/custom-elements", h.GetCustomElements)
	app.Get("/custom-elements/public", h.GetPublicCustomElements)
	app.Get("/custom-elements/:id", h.GetCustomElementByID)
	app.Post("/custom-elements", h.CreateCustomElement)
	app.Patch("/custom-elements/:id", h.UpdateCustomElement)
	app.Delete("/custom-elements/:id", h.DeleteCustomElement)
	app.Post("/custom-elements/:id/duplicate", h.DuplicateCustomElement)

	return app
}

func newCustomElementTestAppNoAuth(repo *testutil.MockCustomElementRepository) *fiber.App {
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

	h := handlers.NewCustomElementHandler(repo)

	app.Get("/custom-elements", h.GetCustomElements)
	app.Get("/custom-elements/:id", h.GetCustomElementByID)
	app.Post("/custom-elements", h.CreateCustomElement)
	app.Patch("/custom-elements/:id", h.UpdateCustomElement)
	app.Delete("/custom-elements/:id", h.DeleteCustomElement)
	app.Post("/custom-elements/:id/duplicate", h.DuplicateCustomElement)

	return app
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func ceStatusOf(t *testing.T, app *fiber.App, method, path string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp.StatusCode
}

func ceBodyOf(t *testing.T, app *fiber.App, method, path string, body io.Reader, contentType string) (int, []byte) {
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

func makeCustomElement(id, name, userID string) *models.CustomElement {
	structure := []byte(`{"type":"div"}`)
	return &models.CustomElement{
		Id:        id,
		Name:      name,
		Structure: structure,
		UserId:    userID,
		IsPublic:  false,
		Version:   "1.0.0",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ─── GetCustomElements ────────────────────────────────────────────────────────

func TestGetCustomElements_ReturnsEmptySliceWhenNone(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		GetCustomElementsFn: func(_ context.Context, _ string, _ *bool) ([]models.CustomElement, error) {
			return []models.CustomElement{}, nil
		},
	}
	app := newCustomElementTestApp(repo)

	status, body := ceBodyOf(t, app, "GET", "/custom-elements", nil, "")
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

func TestGetCustomElements_ReturnsElementsForUser(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		GetCustomElementsFn: func(_ context.Context, userID string, _ *bool) ([]models.CustomElement, error) {
			if userID != "test-user-id" {
				return nil, errors.New("unexpected userID: " + userID)
			}
			return []models.CustomElement{
				*makeCustomElement("ce-1", "MyButton", userID),
				*makeCustomElement("ce-2", "MyCard", userID),
			}, nil
		},
	}
	app := newCustomElementTestApp(repo)

	status, body := ceBodyOf(t, app, "GET", "/custom-elements", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 elements, got %d", len(result))
	}
}

func TestGetCustomElements_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		GetCustomElementsFn: func(_ context.Context, _ string, _ *bool) ([]models.CustomElement, error) {
			return nil, errors.New("db connection lost")
		},
	}
	app := newCustomElementTestApp(repo)

	if code := ceStatusOf(t, app, "GET", "/custom-elements"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetCustomElements_Returns401WhenNoUserID(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	app := newCustomElementTestAppNoAuth(repo)

	if code := ceStatusOf(t, app, "GET", "/custom-elements"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

func TestGetCustomElements_PassesUserIDToRepository(t *testing.T) {
	var capturedUserID string
	repo := &testutil.MockCustomElementRepository{
		GetCustomElementsFn: func(_ context.Context, userID string, _ *bool) ([]models.CustomElement, error) {
			capturedUserID = userID
			return []models.CustomElement{}, nil
		},
	}
	app := newCustomElementTestApp(repo)

	status := ceStatusOf(t, app, "GET", "/custom-elements")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedUserID != "test-user-id" {
		t.Errorf("userID passed to repo: got %q, want %q", capturedUserID, "test-user-id")
	}
}

func TestGetCustomElements_PassesIsPublicFilterWhenProvided(t *testing.T) {
	var capturedIsPublic *bool
	repo := &testutil.MockCustomElementRepository{
		GetCustomElementsFn: func(_ context.Context, _ string, isPublic *bool) ([]models.CustomElement, error) {
			capturedIsPublic = isPublic
			return []models.CustomElement{}, nil
		},
	}
	app := newCustomElementTestApp(repo)

	status := ceStatusOf(t, app, "GET", "/custom-elements?isPublic=true")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedIsPublic == nil {
		t.Fatal("expected isPublic to be non-nil")
	}
	if !*capturedIsPublic {
		t.Errorf("expected isPublic=true, got false")
	}
}

// ─── GetPublicCustomElements ──────────────────────────────────────────────────

func TestGetPublicCustomElements_ReturnsPublicElements(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		GetPublicCustomElementsFn: func(_ context.Context, _ *string, limit, offset int) ([]models.CustomElement, error) {
			el := makeCustomElement("ce-1", "PublicBtn", "owner-1")
			el.IsPublic = true
			return []models.CustomElement{*el}, nil
		},
	}
	app := newCustomElementTestApp(repo)

	status, body := ceBodyOf(t, app, "GET", "/custom-elements/public", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 element, got %d", len(result))
	}
}

func TestGetPublicCustomElements_ReturnsEmptySliceWhenNone(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		GetPublicCustomElementsFn: func(_ context.Context, _ *string, _, _ int) ([]models.CustomElement, error) {
			return []models.CustomElement{}, nil
		},
	}
	app := newCustomElementTestApp(repo)

	status, body := ceBodyOf(t, app, "GET", "/custom-elements/public", nil, "")
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

func TestGetPublicCustomElements_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		GetPublicCustomElementsFn: func(_ context.Context, _ *string, _, _ int) ([]models.CustomElement, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newCustomElementTestApp(repo)

	if code := ceStatusOf(t, app, "GET", "/custom-elements/public"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetPublicCustomElements_PassesCategoryFilterToRepository(t *testing.T) {
	var capturedCategory *string
	repo := &testutil.MockCustomElementRepository{
		GetPublicCustomElementsFn: func(_ context.Context, category *string, _, _ int) ([]models.CustomElement, error) {
			capturedCategory = category
			return []models.CustomElement{}, nil
		},
	}
	app := newCustomElementTestApp(repo)

	status := ceStatusOf(t, app, "GET", "/custom-elements/public?category=ui")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedCategory == nil || *capturedCategory != "ui" {
		t.Errorf("category passed to repo: got %v, want %q", capturedCategory, "ui")
	}
}

// ─── GetCustomElementByID ─────────────────────────────────────────────────────

func TestGetCustomElementByID_ReturnsElementWhenFound(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		GetCustomElementByIDFn: func(_ context.Context, id, userID string) (*models.CustomElement, error) {
			if id == "ce-1" && userID == "test-user-id" {
				return makeCustomElement(id, "MyButton", userID), nil
			}
			return nil, repositories.ErrCustomElementNotFound
		},
	}
	app := newCustomElementTestApp(repo)

	status, body := ceBodyOf(t, app, "GET", "/custom-elements/ce-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["id"] != "ce-1" {
		t.Errorf("id: got %v, want %q", result["id"], "ce-1")
	}
}

func TestGetCustomElementByID_Returns404WhenNotFound(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		GetCustomElementByIDFn: func(_ context.Context, _, _ string) (*models.CustomElement, error) {
			return nil, repositories.ErrCustomElementNotFound
		},
	}
	app := newCustomElementTestApp(repo)

	if code := ceStatusOf(t, app, "GET", "/custom-elements/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetCustomElementByID_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		GetCustomElementByIDFn: func(_ context.Context, _, _ string) (*models.CustomElement, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newCustomElementTestApp(repo)

	if code := ceStatusOf(t, app, "GET", "/custom-elements/ce-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetCustomElementByID_Returns401WhenNoUserID(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	app := newCustomElementTestAppNoAuth(repo)

	if code := ceStatusOf(t, app, "GET", "/custom-elements/ce-1"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── CreateCustomElement ──────────────────────────────────────────────────────

func TestCreateCustomElement_Returns201OnSuccess(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		CreateCustomElementFn: func(_ context.Context, ce *models.CustomElement) (*models.CustomElement, error) {
			return ce, nil
		},
	}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"name":"MyButton","structure":{"type":"div"}}`)
	status, body := ceBodyOf(t, app, "POST", "/custom-elements", payload, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["name"] != "MyButton" {
		t.Errorf("name: got %v, want %q", result["name"], "MyButton")
	}
}

func TestCreateCustomElement_Returns422WhenNameMissing(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"structure":{"type":"div"}}`)
	status, _ := ceBodyOf(t, app, "POST", "/custom-elements", payload, "application/json")
	if status != fiber.StatusUnprocessableEntity {
		t.Errorf("expected 422 for missing name, got %d", status)
	}
}

func TestCreateCustomElement_Returns422WhenStructureMissing(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"name":"MyButton"}`)
	status, _ := ceBodyOf(t, app, "POST", "/custom-elements", payload, "application/json")
	if status != fiber.StatusUnprocessableEntity {
		t.Errorf("expected 422 for missing structure, got %d", status)
	}
}

func TestCreateCustomElement_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{invalid}`)
	status, _ := ceBodyOf(t, app, "POST", "/custom-elements", payload, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", status)
	}
}

func TestCreateCustomElement_Returns409WhenNameAlreadyExists(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		CreateCustomElementFn: func(_ context.Context, _ *models.CustomElement) (*models.CustomElement, error) {
			return nil, repositories.ErrCustomElementAlreadyExists
		},
	}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"name":"Duplicate","structure":{"type":"div"}}`)
	status, _ := ceBodyOf(t, app, "POST", "/custom-elements", payload, "application/json")
	if status != fiber.StatusConflict {
		t.Errorf("expected 409 for duplicate name, got %d", status)
	}
}

func TestCreateCustomElement_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		CreateCustomElementFn: func(_ context.Context, _ *models.CustomElement) (*models.CustomElement, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"name":"MyButton","structure":{"type":"div"}}`)
	status, _ := ceBodyOf(t, app, "POST", "/custom-elements", payload, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestCreateCustomElement_Returns401WhenNoUserID(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	app := newCustomElementTestAppNoAuth(repo)

	payload := strings.NewReader(`{"name":"MyButton","structure":{"type":"div"}}`)
	status, _ := ceBodyOf(t, app, "POST", "/custom-elements", payload, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

func TestCreateCustomElement_DefaultsVersionTo1_0_0WhenNotProvided(t *testing.T) {
	var capturedVersion string
	repo := &testutil.MockCustomElementRepository{
		CreateCustomElementFn: func(_ context.Context, ce *models.CustomElement) (*models.CustomElement, error) {
			capturedVersion = ce.Version
			return ce, nil
		},
	}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"name":"MyButton","structure":{"type":"div"}}`)
	status, _ := ceBodyOf(t, app, "POST", "/custom-elements", payload, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedVersion != "1.0.0" {
		t.Errorf("version: got %q, want %q", capturedVersion, "1.0.0")
	}
}

func TestCreateCustomElement_SetsAuthorIDFromLocals(t *testing.T) {
	var capturedUserID string
	repo := &testutil.MockCustomElementRepository{
		CreateCustomElementFn: func(_ context.Context, ce *models.CustomElement) (*models.CustomElement, error) {
			capturedUserID = ce.UserId
			return ce, nil
		},
	}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"name":"MyButton","structure":{"type":"div"}}`)
	status, _ := ceBodyOf(t, app, "POST", "/custom-elements", payload, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedUserID != "test-user-id" {
		t.Errorf("userId: got %q, want %q", capturedUserID, "test-user-id")
	}
}

// ─── UpdateCustomElement ──────────────────────────────────────────────────────

func TestUpdateCustomElement_Returns200OnSuccess(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		UpdateCustomElementFn: func(_ context.Context, id, userID string, updates map[string]any) (*models.CustomElement, error) {
			ce := makeCustomElement(id, updates["name"].(string), userID)
			return ce, nil
		},
	}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"name":"UpdatedButton"}`)
	status, body := ceBodyOf(t, app, "PATCH", "/custom-elements/ce-1", payload, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["name"] != "UpdatedButton" {
		t.Errorf("name: got %v, want %q", result["name"], "UpdatedButton")
	}
}

func TestUpdateCustomElement_Returns400WhenBodyIsEmpty(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{}`)
	status, _ := ceBodyOf(t, app, "PATCH", "/custom-elements/ce-1", payload, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for empty update body, got %d", status)
	}
}

func TestUpdateCustomElement_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`not-json`)
	status, _ := ceBodyOf(t, app, "PATCH", "/custom-elements/ce-1", payload, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", status)
	}
}

func TestUpdateCustomElement_Returns404WhenNotFound(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		UpdateCustomElementFn: func(_ context.Context, _, _ string, _ map[string]any) (*models.CustomElement, error) {
			return nil, repositories.ErrCustomElementNotFound
		},
	}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"name":"NewName"}`)
	status, _ := ceBodyOf(t, app, "PATCH", "/custom-elements/ghost", payload, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestUpdateCustomElement_Returns403WhenUnauthorized(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		UpdateCustomElementFn: func(_ context.Context, _, _ string, _ map[string]any) (*models.CustomElement, error) {
			return nil, repositories.ErrCustomElementUnauthorized
		},
	}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"name":"NewName"}`)
	status, _ := ceBodyOf(t, app, "PATCH", "/custom-elements/ce-1", payload, "application/json")
	if status != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", status)
	}
}

func TestUpdateCustomElement_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		UpdateCustomElementFn: func(_ context.Context, _, _ string, _ map[string]any) (*models.CustomElement, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"name":"NewName"}`)
	status, _ := ceBodyOf(t, app, "PATCH", "/custom-elements/ce-1", payload, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestUpdateCustomElement_Returns401WhenNoUserID(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	app := newCustomElementTestAppNoAuth(repo)

	payload := strings.NewReader(`{"name":"NewName"}`)
	status, _ := ceBodyOf(t, app, "PATCH", "/custom-elements/ce-1", payload, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

func TestUpdateCustomElement_OnlyAllowsKnownColumns(t *testing.T) {
	var capturedUpdates map[string]any
	repo := &testutil.MockCustomElementRepository{
		UpdateCustomElementFn: func(_ context.Context, id, userID string, updates map[string]any) (*models.CustomElement, error) {
			capturedUpdates = updates
			return makeCustomElement(id, "MyButton", userID), nil
		},
	}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"name":"MyButton","unknownField":"evil","id":"hacked","userId":"attacker"}`)
	status, _ := ceBodyOf(t, app, "PATCH", "/custom-elements/ce-1", payload, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if _, ok := capturedUpdates["unknownField"]; ok {
		t.Error("unknown field should not be passed to the repository")
	}
	if _, ok := capturedUpdates["id"]; ok {
		t.Error("immutable field 'id' should not be passed to the repository")
	}
	if _, ok := capturedUpdates["userId"]; ok {
		t.Error("immutable field 'userId' should not be passed to the repository")
	}
}

// ─── DeleteCustomElement ──────────────────────────────────────────────────────

func TestDeleteCustomElement_Returns200OnSuccess(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		DeleteCustomElementFn: func(_ context.Context, _, _ string) error {
			return nil
		},
	}
	app := newCustomElementTestApp(repo)

	status, _ := ceBodyOf(t, app, "DELETE", "/custom-elements/ce-1", nil, "")
	if status != fiber.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
}

func TestDeleteCustomElement_Returns403WhenUnauthorized(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		DeleteCustomElementFn: func(_ context.Context, _, _ string) error {
			return repositories.ErrCustomElementUnauthorized
		},
	}
	app := newCustomElementTestApp(repo)

	if code := ceStatusOf(t, app, "DELETE", "/custom-elements/ce-1"); code != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", code)
	}
}

func TestDeleteCustomElement_Returns404WhenNotFound(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		DeleteCustomElementFn: func(_ context.Context, _, _ string) error {
			return repositories.ErrCustomElementNotFound
		},
	}
	app := newCustomElementTestApp(repo)

	if code := ceStatusOf(t, app, "DELETE", "/custom-elements/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeleteCustomElement_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		DeleteCustomElementFn: func(_ context.Context, _, _ string) error {
			return errors.New("db error")
		},
	}
	app := newCustomElementTestApp(repo)

	if code := ceStatusOf(t, app, "DELETE", "/custom-elements/ce-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestDeleteCustomElement_Returns401WhenNoUserID(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	app := newCustomElementTestAppNoAuth(repo)

	if code := ceStatusOf(t, app, "DELETE", "/custom-elements/ce-1"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

func TestDeleteCustomElement_PassesCorrectIDAndUserIDToRepository(t *testing.T) {
	var capturedID, capturedUserID string
	repo := &testutil.MockCustomElementRepository{
		DeleteCustomElementFn: func(_ context.Context, id, userID string) error {
			capturedID = id
			capturedUserID = userID
			return nil
		},
	}
	app := newCustomElementTestApp(repo)

	if code := ceStatusOf(t, app, "DELETE", "/custom-elements/target-ce"); code != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if capturedID != "target-ce" {
		t.Errorf("id passed to repo: got %q, want %q", capturedID, "target-ce")
	}
	if capturedUserID != "test-user-id" {
		t.Errorf("userID passed to repo: got %q, want %q", capturedUserID, "test-user-id")
	}
}

// ─── DuplicateCustomElement ───────────────────────────────────────────────────

func TestDuplicateCustomElement_Returns201OnSuccess(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		DuplicateCustomElementFn: func(_ context.Context, id, userID, newName string) (*models.CustomElement, error) {
			ce := makeCustomElement(id+"-copy", newName, userID)
			return ce, nil
		},
	}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"newName":"MyButton Copy"}`)
	status, body := ceBodyOf(t, app, "POST", "/custom-elements/ce-1/duplicate", payload, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["name"] != "MyButton Copy" {
		t.Errorf("name: got %v, want %q", result["name"], "MyButton Copy")
	}
}

func TestDuplicateCustomElement_Returns422WhenNewNameMissing(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{}`)
	status, _ := ceBodyOf(t, app, "POST", "/custom-elements/ce-1/duplicate", payload, "application/json")
	if status != fiber.StatusUnprocessableEntity {
		t.Errorf("expected 422 for missing newName, got %d", status)
	}
}

func TestDuplicateCustomElement_Returns404WhenSourceNotFound(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		DuplicateCustomElementFn: func(_ context.Context, _, _, _ string) (*models.CustomElement, error) {
			return nil, repositories.ErrCustomElementNotFound
		},
	}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"newName":"Copy"}`)
	status, _ := ceBodyOf(t, app, "POST", "/custom-elements/ghost/duplicate", payload, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestDuplicateCustomElement_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{
		DuplicateCustomElementFn: func(_ context.Context, _, _, _ string) (*models.CustomElement, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"newName":"Copy"}`)
	status, _ := ceBodyOf(t, app, "POST", "/custom-elements/ce-1/duplicate", payload, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestDuplicateCustomElement_Returns401WhenNoUserID(t *testing.T) {
	repo := &testutil.MockCustomElementRepository{}
	app := newCustomElementTestAppNoAuth(repo)

	payload := strings.NewReader(`{"newName":"Copy"}`)
	status, _ := ceBodyOf(t, app, "POST", "/custom-elements/ce-1/duplicate", payload, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

func TestDuplicateCustomElement_PassesCorrectArgsToRepository(t *testing.T) {
	var capturedID, capturedUserID, capturedNewName string
	repo := &testutil.MockCustomElementRepository{
		DuplicateCustomElementFn: func(_ context.Context, id, userID, newName string) (*models.CustomElement, error) {
			capturedID = id
			capturedUserID = userID
			capturedNewName = newName
			return makeCustomElement(id+"-copy", newName, userID), nil
		},
	}
	app := newCustomElementTestApp(repo)

	payload := strings.NewReader(`{"newName":"The Clone"}`)
	status, _ := ceBodyOf(t, app, "POST", "/custom-elements/source-id/duplicate", payload, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedID != "source-id" {
		t.Errorf("id passed to repo: got %q, want %q", capturedID, "source-id")
	}
	if capturedUserID != "test-user-id" {
		t.Errorf("userID passed to repo: got %q, want %q", capturedUserID, "test-user-id")
	}
	if capturedNewName != "The Clone" {
		t.Errorf("newName passed to repo: got %q, want %q", capturedNewName, "The Clone")
	}
}