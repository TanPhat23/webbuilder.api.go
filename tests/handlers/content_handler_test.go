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

	"github.com/gofiber/fiber/v2"
)

// ─── mock ContentTypeRepository ───────────────────────────────────────────────

type mockContentTypeRepo struct {
	GetContentTypesFn    func(ctx context.Context) ([]models.ContentType, error)
	GetContentTypeByIDFn func(ctx context.Context, id string) (*models.ContentType, error)
	CreateContentTypeFn  func(ctx context.Context, ct *models.ContentType) (*models.ContentType, error)
	UpdateContentTypeFn  func(ctx context.Context, id string, updates map[string]any) (*models.ContentType, error)
	DeleteContentTypeFn  func(ctx context.Context, id string) error
}

var _ repositories.ContentTypeRepositoryInterface = (*mockContentTypeRepo)(nil)

func (m *mockContentTypeRepo) GetContentTypes(ctx context.Context) ([]models.ContentType, error) {
	if m.GetContentTypesFn != nil {
		return m.GetContentTypesFn(ctx)
	}
	return []models.ContentType{}, nil
}
func (m *mockContentTypeRepo) GetContentTypeByID(ctx context.Context, id string) (*models.ContentType, error) {
	if m.GetContentTypeByIDFn != nil {
		return m.GetContentTypeByIDFn(ctx, id)
	}
	return nil, repositories.ErrContentTypeNotFound
}
func (m *mockContentTypeRepo) CreateContentType(ctx context.Context, ct *models.ContentType) (*models.ContentType, error) {
	if m.CreateContentTypeFn != nil {
		return m.CreateContentTypeFn(ctx, ct)
	}
	return ct, nil
}
func (m *mockContentTypeRepo) UpdateContentType(ctx context.Context, id string, updates map[string]any) (*models.ContentType, error) {
	if m.UpdateContentTypeFn != nil {
		return m.UpdateContentTypeFn(ctx, id, updates)
	}
	return nil, repositories.ErrContentTypeNotFound
}
func (m *mockContentTypeRepo) DeleteContentType(ctx context.Context, id string) error {
	if m.DeleteContentTypeFn != nil {
		return m.DeleteContentTypeFn(ctx, id)
	}
	return nil
}

// ─── mock ContentFieldRepository ─────────────────────────────────────────────

type mockContentFieldRepo struct {
	GetContentFieldsByContentTypeFn func(ctx context.Context, contentTypeID string) ([]models.ContentField, error)
	GetContentFieldByIDFn           func(ctx context.Context, id string) (*models.ContentField, error)
	CreateContentFieldFn            func(ctx context.Context, cf *models.ContentField) (*models.ContentField, error)
	UpdateContentFieldFn            func(ctx context.Context, id string, updates map[string]any) (*models.ContentField, error)
	DeleteContentFieldFn            func(ctx context.Context, id string) error
}

var _ repositories.ContentFieldRepositoryInterface = (*mockContentFieldRepo)(nil)

func (m *mockContentFieldRepo) GetContentFieldsByContentType(ctx context.Context, contentTypeID string) ([]models.ContentField, error) {
	if m.GetContentFieldsByContentTypeFn != nil {
		return m.GetContentFieldsByContentTypeFn(ctx, contentTypeID)
	}
	return []models.ContentField{}, nil
}
func (m *mockContentFieldRepo) GetContentFieldByID(ctx context.Context, id string) (*models.ContentField, error) {
	if m.GetContentFieldByIDFn != nil {
		return m.GetContentFieldByIDFn(ctx, id)
	}
	return nil, repositories.ErrContentFieldNotFound
}
func (m *mockContentFieldRepo) CreateContentField(ctx context.Context, cf *models.ContentField) (*models.ContentField, error) {
	if m.CreateContentFieldFn != nil {
		return m.CreateContentFieldFn(ctx, cf)
	}
	return cf, nil
}
func (m *mockContentFieldRepo) UpdateContentField(ctx context.Context, id string, updates map[string]any) (*models.ContentField, error) {
	if m.UpdateContentFieldFn != nil {
		return m.UpdateContentFieldFn(ctx, id, updates)
	}
	return nil, repositories.ErrContentFieldNotFound
}
func (m *mockContentFieldRepo) DeleteContentField(ctx context.Context, id string) error {
	if m.DeleteContentFieldFn != nil {
		return m.DeleteContentFieldFn(ctx, id)
	}
	return nil
}

// ─── mock ContentItemRepository ──────────────────────────────────────────────

type mockContentItemRepo struct {
	GetContentItemsByContentTypeFn func(ctx context.Context, contentTypeID string) ([]models.ContentItem, error)
	GetContentItemByIDFn           func(ctx context.Context, id string) (*models.ContentItem, error)
	GetContentItemBySlugFn         func(ctx context.Context, contentTypeID, slug string) (*models.ContentItem, error)
	GetPublicContentItemsFn        func(ctx context.Context, contentTypeID string, limit int, sortBy, sortOrder string) ([]models.ContentItem, error)
	CreateContentItemFn            func(ctx context.Context, item *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	UpdateContentItemFn            func(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error)
	DeleteContentItemFn            func(ctx context.Context, id string) error
}

var _ repositories.ContentItemRepositoryInterface = (*mockContentItemRepo)(nil)

func (m *mockContentItemRepo) GetContentItemsByContentType(ctx context.Context, contentTypeID string) ([]models.ContentItem, error) {
	if m.GetContentItemsByContentTypeFn != nil {
		return m.GetContentItemsByContentTypeFn(ctx, contentTypeID)
	}
	return []models.ContentItem{}, nil
}
func (m *mockContentItemRepo) GetContentItemByID(ctx context.Context, id string) (*models.ContentItem, error) {
	if m.GetContentItemByIDFn != nil {
		return m.GetContentItemByIDFn(ctx, id)
	}
	return nil, repositories.ErrContentItemNotFound
}
func (m *mockContentItemRepo) GetContentItemBySlug(ctx context.Context, contentTypeID, slug string) (*models.ContentItem, error) {
	if m.GetContentItemBySlugFn != nil {
		return m.GetContentItemBySlugFn(ctx, contentTypeID, slug)
	}
	return nil, repositories.ErrContentItemNotFound
}
func (m *mockContentItemRepo) GetPublicContentItems(ctx context.Context, contentTypeID string, limit int, sortBy, sortOrder string) ([]models.ContentItem, error) {
	if m.GetPublicContentItemsFn != nil {
		return m.GetPublicContentItemsFn(ctx, contentTypeID, limit, sortBy, sortOrder)
	}
	return []models.ContentItem{}, nil
}
func (m *mockContentItemRepo) CreateContentItem(ctx context.Context, item *models.ContentItem, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	if m.CreateContentItemFn != nil {
		return m.CreateContentItemFn(ctx, item, fieldValues)
	}
	return item, nil
}
func (m *mockContentItemRepo) UpdateContentItem(ctx context.Context, id string, updates map[string]any, fieldValues []models.ContentFieldValue) (*models.ContentItem, error) {
	if m.UpdateContentItemFn != nil {
		return m.UpdateContentItemFn(ctx, id, updates, fieldValues)
	}
	return nil, repositories.ErrContentItemNotFound
}
func (m *mockContentItemRepo) DeleteContentItem(ctx context.Context, id string) error {
	if m.DeleteContentItemFn != nil {
		return m.DeleteContentItemFn(ctx, id)
	}
	return nil
}

// ─── test app factories ───────────────────────────────────────────────────────

func newContentTypeTestApp(repo repositories.ContentTypeRepositoryInterface) *fiber.App {
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
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", "test-user-id")
		return c.Next()
	})

	h := handlers.NewContentTypeHandler(repo)
	app.Get("/content-types", h.GetContentTypes)
	app.Get("/content-types/:id", h.GetContentTypeByID)
	app.Post("/content-types", h.CreateContentType)
	app.Patch("/content-types/:id", h.UpdateContentType)
	app.Delete("/content-types/:id", h.DeleteContentType)
	return app
}

func newContentFieldTestApp(repo repositories.ContentFieldRepositoryInterface) *fiber.App {
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
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", "test-user-id")
		return c.Next()
	})

	h := handlers.NewContentFieldHandler(repo)
	app.Get("/content-types/:contentTypeId/fields", h.GetContentFieldsByContentType)
	app.Get("/fields/:fieldId", h.GetContentFieldByID)
	app.Post("/content-types/:contentTypeId/fields", h.CreateContentField)
	app.Patch("/fields/:fieldId", h.UpdateContentField)
	app.Delete("/fields/:fieldId", h.DeleteContentField)
	return app
}

func newContentItemTestApp(repo repositories.ContentItemRepositoryInterface) *fiber.App {
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
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", "test-user-id")
		return c.Next()
	})

	h := handlers.NewContentItemHandler(repo)
	app.Get("/content-types/:contentTypeId/items", h.GetContentItemsByContentType)
	app.Get("/items/:itemId", h.GetContentItemByID)
	app.Post("/content-types/:contentTypeId/items", h.CreateContentItem)
	app.Patch("/items/:itemId", h.UpdateContentItem)
	app.Delete("/items/:itemId", h.DeleteContentItem)
	app.Get("/public/items", h.GetPublicContentItems)
	app.Get("/public/content-types/:contentTypeId/items/:slug", h.GetPublicContentItemBySlug)
	return app
}

// ─── request helpers ──────────────────────────────────────────────────────────

func contentStatusOf(t *testing.T, app *fiber.App, method, path string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp.StatusCode
}

func contentBodyOf(t *testing.T, app *fiber.App, method, path string, body io.Reader, contentType string) (int, []byte) {
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

func makeContentType(id, name string) *models.ContentType {
	return &models.ContentType{Id: id, Name: name}
}

func makeContentField(id, contentTypeID, name, fieldType string) *models.ContentField {
	return &models.ContentField{
		Id:            id,
		Name:          name,
		Type:          fieldType,
		Required:      false,
		ContentTypeId: contentTypeID,
	}
}

func makeContentItem(id, contentTypeID, title, slug string, published bool) *models.ContentItem {
	now := time.Now()
	return &models.ContentItem{
		Id:            id,
		Title:         title,
		Slug:          slug,
		Published:     published,
		ContentTypeId: contentTypeID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ContentTypeHandler
// ═══════════════════════════════════════════════════════════════════════════════

// ─── GetContentTypes ──────────────────────────────────────────────────────────

func TestGetContentTypes_ReturnsEmptySliceWhenNone(t *testing.T) {
	// An empty repository must produce 200 with an empty JSON array.
	repo := &mockContentTypeRepo{
		GetContentTypesFn: func(_ context.Context) ([]models.ContentType, error) {
			return []models.ContentType{}, nil
		},
	}
	app := newContentTypeTestApp(repo)

	status, body := contentBodyOf(t, app, "GET", "/content-types", nil, "")
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

func TestGetContentTypes_ReturnsAllContentTypes(t *testing.T) {
	// All content types must be returned as a JSON array.
	repo := &mockContentTypeRepo{
		GetContentTypesFn: func(_ context.Context) ([]models.ContentType, error) {
			return []models.ContentType{
				*makeContentType("ct-1", "Blog"),
				*makeContentType("ct-2", "Product"),
			}, nil
		},
	}
	app := newContentTypeTestApp(repo)

	status, body := contentBodyOf(t, app, "GET", "/content-types", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}
	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 content types, got %d", len(result))
	}
}

func TestGetContentTypes_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentTypeRepo{
		GetContentTypesFn: func(_ context.Context) ([]models.ContentType, error) {
			return nil, errors.New("db error")
		},
	}
	app := newContentTypeTestApp(repo)

	if code := contentStatusOf(t, app, "GET", "/content-types"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── GetContentTypeByID ───────────────────────────────────────────────────────

func TestGetContentTypeByID_ReturnsContentTypeWhenFound(t *testing.T) {
	repo := &mockContentTypeRepo{
		GetContentTypeByIDFn: func(_ context.Context, id string) (*models.ContentType, error) {
			if id == "ct-1" {
				return makeContentType("ct-1", "Blog"), nil
			}
			return nil, repositories.ErrContentTypeNotFound
		},
	}
	app := newContentTypeTestApp(repo)

	status, body := contentBodyOf(t, app, "GET", "/content-types/ct-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["id"] != "ct-1" {
		t.Errorf("id: got %v, want %q", result["id"], "ct-1")
	}
}

func TestGetContentTypeByID_Returns404WhenNotFound(t *testing.T) {
	// ErrContentTypeNotFound (ends with "not found") must map to 404.
	repo := &mockContentTypeRepo{
		GetContentTypeByIDFn: func(_ context.Context, _ string) (*models.ContentType, error) {
			return nil, repositories.ErrContentTypeNotFound
		},
	}
	app := newContentTypeTestApp(repo)

	if code := contentStatusOf(t, app, "GET", "/content-types/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetContentTypeByID_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentTypeRepo{
		GetContentTypeByIDFn: func(_ context.Context, _ string) (*models.ContentType, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newContentTypeTestApp(repo)

	if code := contentStatusOf(t, app, "GET", "/content-types/ct-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── CreateContentType ────────────────────────────────────────────────────────

func TestCreateContentType_Returns201OnSuccess(t *testing.T) {
	repo := &mockContentTypeRepo{
		CreateContentTypeFn: func(_ context.Context, ct *models.ContentType) (*models.ContentType, error) {
			ct.Id = "ct-new"
			return ct, nil
		},
	}
	app := newContentTypeTestApp(repo)

	body := strings.NewReader(`{"name":"Blog"}`)
	status, respBody := contentBodyOf(t, app, "POST", "/content-types", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d – body: %s", status, respBody)
	}
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["name"] != "Blog" {
		t.Errorf("name: got %v, want %q", result["name"], "Blog")
	}
}

func TestCreateContentType_Returns400WhenNameMissing(t *testing.T) {
	// "name" is required; omitting it must fail validation.
	app := newContentTypeTestApp(&mockContentTypeRepo{})

	body := strings.NewReader(`{}`)
	status, _ := contentBodyOf(t, app, "POST", "/content-types", body, "application/json")
	if status == fiber.StatusCreated || status == fiber.StatusOK {
		t.Errorf("expected 4xx for missing name, got %d", status)
	}
}

func TestCreateContentType_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	app := newContentTypeTestApp(&mockContentTypeRepo{})

	body := strings.NewReader(`not-json`)
	status, _ := contentBodyOf(t, app, "POST", "/content-types", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestCreateContentType_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentTypeRepo{
		CreateContentTypeFn: func(_ context.Context, _ *models.ContentType) (*models.ContentType, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newContentTypeTestApp(repo)

	body := strings.NewReader(`{"name":"Blog"}`)
	status, _ := contentBodyOf(t, app, "POST", "/content-types", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

// ─── UpdateContentType ────────────────────────────────────────────────────────

func TestUpdateContentType_Returns200OnSuccess(t *testing.T) {
	repo := &mockContentTypeRepo{
		UpdateContentTypeFn: func(_ context.Context, id string, updates map[string]any) (*models.ContentType, error) {
			ct := makeContentType(id, "Updated Blog")
			return ct, nil
		},
	}
	app := newContentTypeTestApp(repo)

	body := strings.NewReader(`{"name":"Updated Blog"}`)
	status, respBody := contentBodyOf(t, app, "PATCH", "/content-types/ct-1", body, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, respBody)
	}
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["name"] != "Updated Blog" {
		t.Errorf("name: got %v, want %q", result["name"], "Updated Blog")
	}
}

func TestUpdateContentType_Returns400WhenBodyIsEmpty(t *testing.T) {
	// An empty body means no valid fields to update; RequireUpdates rejects it.
	app := newContentTypeTestApp(&mockContentTypeRepo{})

	body := strings.NewReader(`{}`)
	status, _ := contentBodyOf(t, app, "PATCH", "/content-types/ct-1", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestUpdateContentType_Returns404WhenNotFound(t *testing.T) {
	repo := &mockContentTypeRepo{
		UpdateContentTypeFn: func(_ context.Context, _ string, _ map[string]any) (*models.ContentType, error) {
			return nil, repositories.ErrContentTypeNotFound
		},
	}
	app := newContentTypeTestApp(repo)

	body := strings.NewReader(`{"name":"New Name"}`)
	status, _ := contentBodyOf(t, app, "PATCH", "/content-types/ghost", body, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestUpdateContentType_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentTypeRepo{
		UpdateContentTypeFn: func(_ context.Context, _ string, _ map[string]any) (*models.ContentType, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newContentTypeTestApp(repo)

	body := strings.NewReader(`{"name":"New Name"}`)
	status, _ := contentBodyOf(t, app, "PATCH", "/content-types/ct-1", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestUpdateContentType_OnlyAllowsKnownColumns(t *testing.T) {
	// An unknown field like "ownerId" is stripped by the allowlist; with no
	// remaining fields RequireUpdates returns 400.
	var updateCalled bool
	repo := &mockContentTypeRepo{
		UpdateContentTypeFn: func(_ context.Context, _ string, _ map[string]any) (*models.ContentType, error) {
			updateCalled = true
			return makeContentType("ct-1", "Blog"), nil
		},
	}
	app := newContentTypeTestApp(repo)

	body := strings.NewReader(`{"unknownField":"value"}`)
	status, _ := contentBodyOf(t, app, "PATCH", "/content-types/ct-1", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for unknown-only fields, got %d", status)
	}
	if updateCalled {
		t.Error("UpdateContentType must not be called when all fields are stripped")
	}
}

// ─── DeleteContentType ────────────────────────────────────────────────────────

func TestDeleteContentType_Returns204OnSuccess(t *testing.T) {
	repo := &mockContentTypeRepo{
		DeleteContentTypeFn: func(_ context.Context, _ string) error { return nil },
	}
	app := newContentTypeTestApp(repo)

	if code := contentStatusOf(t, app, "DELETE", "/content-types/ct-1"); code != fiber.StatusNoContent {
		t.Errorf("expected 204, got %d", code)
	}
}

func TestDeleteContentType_Returns404WhenNotFound(t *testing.T) {
	repo := &mockContentTypeRepo{
		DeleteContentTypeFn: func(_ context.Context, _ string) error {
			return repositories.ErrContentTypeNotFound
		},
	}
	app := newContentTypeTestApp(repo)

	if code := contentStatusOf(t, app, "DELETE", "/content-types/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeleteContentType_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentTypeRepo{
		DeleteContentTypeFn: func(_ context.Context, _ string) error {
			return errors.New("db error")
		},
	}
	app := newContentTypeTestApp(repo)

	if code := contentStatusOf(t, app, "DELETE", "/content-types/ct-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ContentFieldHandler
// ═══════════════════════════════════════════════════════════════════════════════

// ─── GetContentFieldsByContentType ───────────────────────────────────────────

func TestGetContentFieldsByContentType_ReturnsEmptySlice(t *testing.T) {
	repo := &mockContentFieldRepo{
		GetContentFieldsByContentTypeFn: func(_ context.Context, _ string) ([]models.ContentField, error) {
			return []models.ContentField{}, nil
		},
	}
	app := newContentFieldTestApp(repo)

	status, body := contentBodyOf(t, app, "GET", "/content-types/ct-1/fields", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}
	var result []any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d", len(result))
	}
}

func TestGetContentFieldsByContentType_ReturnsFields(t *testing.T) {
	repo := &mockContentFieldRepo{
		GetContentFieldsByContentTypeFn: func(_ context.Context, contentTypeID string) ([]models.ContentField, error) {
			return []models.ContentField{
				*makeContentField("f-1", contentTypeID, "Title", "text"),
				*makeContentField("f-2", contentTypeID, "Body", "richtext"),
			}, nil
		},
	}
	app := newContentFieldTestApp(repo)

	status, body := contentBodyOf(t, app, "GET", "/content-types/ct-1/fields", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}
	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 fields, got %d", len(result))
	}
}

func TestGetContentFieldsByContentType_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentFieldRepo{
		GetContentFieldsByContentTypeFn: func(_ context.Context, _ string) ([]models.ContentField, error) {
			return nil, errors.New("db error")
		},
	}
	app := newContentFieldTestApp(repo)

	if code := contentStatusOf(t, app, "GET", "/content-types/ct-1/fields"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── GetContentFieldByID ──────────────────────────────────────────────────────

func TestGetContentFieldByID_ReturnsFieldWhenFound(t *testing.T) {
	repo := &mockContentFieldRepo{
		GetContentFieldByIDFn: func(_ context.Context, id string) (*models.ContentField, error) {
			if id == "f-1" {
				return makeContentField("f-1", "ct-1", "Title", "text"), nil
			}
			return nil, repositories.ErrContentFieldNotFound
		},
	}
	app := newContentFieldTestApp(repo)

	status, body := contentBodyOf(t, app, "GET", "/fields/f-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["id"] != "f-1" {
		t.Errorf("id: got %v, want %q", result["id"], "f-1")
	}
}

func TestGetContentFieldByID_Returns404WhenNotFound(t *testing.T) {
	repo := &mockContentFieldRepo{
		GetContentFieldByIDFn: func(_ context.Context, _ string) (*models.ContentField, error) {
			return nil, repositories.ErrContentFieldNotFound
		},
	}
	app := newContentFieldTestApp(repo)

	if code := contentStatusOf(t, app, "GET", "/fields/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetContentFieldByID_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentFieldRepo{
		GetContentFieldByIDFn: func(_ context.Context, _ string) (*models.ContentField, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newContentFieldTestApp(repo)

	if code := contentStatusOf(t, app, "GET", "/fields/f-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── CreateContentField ───────────────────────────────────────────────────────

func TestCreateContentField_Returns201OnSuccess(t *testing.T) {
	repo := &mockContentFieldRepo{
		CreateContentFieldFn: func(_ context.Context, cf *models.ContentField) (*models.ContentField, error) {
			cf.Id = "f-new"
			return cf, nil
		},
	}
	app := newContentFieldTestApp(repo)

	body := strings.NewReader(`{"name":"Title","type":"text"}`)
	status, respBody := contentBodyOf(t, app, "POST", "/content-types/ct-1/fields", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d – body: %s", status, respBody)
	}
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["name"] != "Title" {
		t.Errorf("name: got %v, want %q", result["name"], "Title")
	}
	if result["contentTypeId"] != "ct-1" {
		t.Errorf("contentTypeId: got %v, want %q", result["contentTypeId"], "ct-1")
	}
}

func TestCreateContentField_Returns400WhenNameMissing(t *testing.T) {
	// "name" is required; omitting it must fail validation.
	app := newContentFieldTestApp(&mockContentFieldRepo{})

	body := strings.NewReader(`{"type":"text"}`)
	status, _ := contentBodyOf(t, app, "POST", "/content-types/ct-1/fields", body, "application/json")
	if status == fiber.StatusCreated || status == fiber.StatusOK {
		t.Errorf("expected 4xx for missing name, got %d", status)
	}
}

func TestCreateContentField_Returns400WhenTypeMissing(t *testing.T) {
	// "type" is required; omitting it must fail validation.
	app := newContentFieldTestApp(&mockContentFieldRepo{})

	body := strings.NewReader(`{"name":"Title"}`)
	status, _ := contentBodyOf(t, app, "POST", "/content-types/ct-1/fields", body, "application/json")
	if status == fiber.StatusCreated || status == fiber.StatusOK {
		t.Errorf("expected 4xx for missing type, got %d", status)
	}
}

func TestCreateContentField_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentFieldRepo{
		CreateContentFieldFn: func(_ context.Context, _ *models.ContentField) (*models.ContentField, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newContentFieldTestApp(repo)

	body := strings.NewReader(`{"name":"Title","type":"text"}`)
	status, _ := contentBodyOf(t, app, "POST", "/content-types/ct-1/fields", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestCreateContentField_SetsContentTypeIDFromRouteParam(t *testing.T) {
	// The handler must populate ContentTypeId from the route param, not from the
	// request body.
	var capturedContentTypeID string
	repo := &mockContentFieldRepo{
		CreateContentFieldFn: func(_ context.Context, cf *models.ContentField) (*models.ContentField, error) {
			capturedContentTypeID = cf.ContentTypeId
			return cf, nil
		},
	}
	app := newContentFieldTestApp(repo)

	body := strings.NewReader(`{"name":"Title","type":"text"}`)
	status, _ := contentBodyOf(t, app, "POST", "/content-types/my-type-id/fields", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedContentTypeID != "my-type-id" {
		t.Errorf("contentTypeId: got %q, want %q", capturedContentTypeID, "my-type-id")
	}
}

// ─── UpdateContentField ───────────────────────────────────────────────────────

func TestUpdateContentField_Returns200OnSuccess(t *testing.T) {
	repo := &mockContentFieldRepo{
		UpdateContentFieldFn: func(_ context.Context, id string, _ map[string]any) (*models.ContentField, error) {
			return makeContentField(id, "ct-1", "Updated Title", "text"), nil
		},
	}
	app := newContentFieldTestApp(repo)

	body := strings.NewReader(`{"name":"Updated Title"}`)
	status, respBody := contentBodyOf(t, app, "PATCH", "/fields/f-1", body, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, respBody)
	}
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["name"] != "Updated Title" {
		t.Errorf("name: got %v, want %q", result["name"], "Updated Title")
	}
}

func TestUpdateContentField_Returns400WhenBodyIsEmpty(t *testing.T) {
	app := newContentFieldTestApp(&mockContentFieldRepo{})

	body := strings.NewReader(`{}`)
	status, _ := contentBodyOf(t, app, "PATCH", "/fields/f-1", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for empty update body, got %d", status)
	}
}

func TestUpdateContentField_Returns404WhenNotFound(t *testing.T) {
	repo := &mockContentFieldRepo{
		UpdateContentFieldFn: func(_ context.Context, _ string, _ map[string]any) (*models.ContentField, error) {
			return nil, repositories.ErrContentFieldNotFound
		},
	}
	app := newContentFieldTestApp(repo)

	body := strings.NewReader(`{"name":"New Name"}`)
	status, _ := contentBodyOf(t, app, "PATCH", "/fields/ghost", body, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestUpdateContentField_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentFieldRepo{
		UpdateContentFieldFn: func(_ context.Context, _ string, _ map[string]any) (*models.ContentField, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newContentFieldTestApp(repo)

	body := strings.NewReader(`{"name":"New Name"}`)
	status, _ := contentBodyOf(t, app, "PATCH", "/fields/f-1", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestUpdateContentField_OnlyAllowsKnownColumns(t *testing.T) {
	// Unknown fields are stripped; with nothing left RequireUpdates returns 400.
	var updateCalled bool
	repo := &mockContentFieldRepo{
		UpdateContentFieldFn: func(_ context.Context, _ string, _ map[string]any) (*models.ContentField, error) {
			updateCalled = true
			return makeContentField("f-1", "ct-1", "Title", "text"), nil
		},
	}
	app := newContentFieldTestApp(repo)

	body := strings.NewReader(`{"unknownField":"value"}`)
	status, _ := contentBodyOf(t, app, "PATCH", "/fields/f-1", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for unknown-only fields, got %d", status)
	}
	if updateCalled {
		t.Error("UpdateContentField must not be called when all fields are stripped")
	}
}

// ─── DeleteContentField ───────────────────────────────────────────────────────

func TestDeleteContentField_Returns204OnSuccess(t *testing.T) {
	repo := &mockContentFieldRepo{
		DeleteContentFieldFn: func(_ context.Context, _ string) error { return nil },
	}
	app := newContentFieldTestApp(repo)

	if code := contentStatusOf(t, app, "DELETE", "/fields/f-1"); code != fiber.StatusNoContent {
		t.Errorf("expected 204, got %d", code)
	}
}

func TestDeleteContentField_Returns404WhenNotFound(t *testing.T) {
	repo := &mockContentFieldRepo{
		DeleteContentFieldFn: func(_ context.Context, _ string) error {
			return repositories.ErrContentFieldNotFound
		},
	}
	app := newContentFieldTestApp(repo)

	if code := contentStatusOf(t, app, "DELETE", "/fields/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeleteContentField_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentFieldRepo{
		DeleteContentFieldFn: func(_ context.Context, _ string) error {
			return errors.New("db error")
		},
	}
	app := newContentFieldTestApp(repo)

	if code := contentStatusOf(t, app, "DELETE", "/fields/f-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ContentItemHandler
// ═══════════════════════════════════════════════════════════════════════════════

// ─── GetContentItemsByContentType ─────────────────────────────────────────────

func TestGetContentItemsByContentType_ReturnsEmptySlice(t *testing.T) {
	repo := &mockContentItemRepo{
		GetContentItemsByContentTypeFn: func(_ context.Context, _ string) ([]models.ContentItem, error) {
			return []models.ContentItem{}, nil
		},
	}
	app := newContentItemTestApp(repo)

	status, body := contentBodyOf(t, app, "GET", "/content-types/ct-1/items", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}
	var result []any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d", len(result))
	}
}

func TestGetContentItemsByContentType_ReturnsItems(t *testing.T) {
	repo := &mockContentItemRepo{
		GetContentItemsByContentTypeFn: func(_ context.Context, ctID string) ([]models.ContentItem, error) {
			return []models.ContentItem{
				*makeContentItem("i-1", ctID, "Hello World", "hello-world", true),
				*makeContentItem("i-2", ctID, "Post Two", "post-two", false),
			}, nil
		},
	}
	app := newContentItemTestApp(repo)

	status, body := contentBodyOf(t, app, "GET", "/content-types/ct-1/items", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}
	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
}

func TestGetContentItemsByContentType_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentItemRepo{
		GetContentItemsByContentTypeFn: func(_ context.Context, _ string) ([]models.ContentItem, error) {
			return nil, errors.New("db error")
		},
	}
	app := newContentItemTestApp(repo)

	if code := contentStatusOf(t, app, "GET", "/content-types/ct-1/items"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── GetContentItemByID ───────────────────────────────────────────────────────

func TestGetContentItemByID_ReturnsItemWhenFound(t *testing.T) {
	repo := &mockContentItemRepo{
		GetContentItemByIDFn: func(_ context.Context, id string) (*models.ContentItem, error) {
			if id == "i-1" {
				return makeContentItem("i-1", "ct-1", "Hello World", "hello-world", true), nil
			}
			return nil, repositories.ErrContentItemNotFound
		},
	}
	app := newContentItemTestApp(repo)

	status, body := contentBodyOf(t, app, "GET", "/items/i-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["id"] != "i-1" {
		t.Errorf("id: got %v, want %q", result["id"], "i-1")
	}
}

func TestGetContentItemByID_Returns404WhenNotFound(t *testing.T) {
	repo := &mockContentItemRepo{
		GetContentItemByIDFn: func(_ context.Context, _ string) (*models.ContentItem, error) {
			return nil, repositories.ErrContentItemNotFound
		},
	}
	app := newContentItemTestApp(repo)

	if code := contentStatusOf(t, app, "GET", "/items/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetContentItemByID_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentItemRepo{
		GetContentItemByIDFn: func(_ context.Context, _ string) (*models.ContentItem, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newContentItemTestApp(repo)

	if code := contentStatusOf(t, app, "GET", "/items/i-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── CreateContentItem ────────────────────────────────────────────────────────

func TestCreateContentItem_Returns201OnSuccess(t *testing.T) {
	// CreateContentItem uses ValidateJSONBody (parse-only), so all fields are
	// optional from a validation standpoint.
	repo := &mockContentItemRepo{
		CreateContentItemFn: func(_ context.Context, item *models.ContentItem, _ []models.ContentFieldValue) (*models.ContentItem, error) {
			item.Id = "i-new"
			return item, nil
		},
	}
	app := newContentItemTestApp(repo)

	body := strings.NewReader(`{"title":"New Post","slug":"new-post","published":false}`)
	status, respBody := contentBodyOf(t, app, "POST", "/content-types/ct-1/items", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d – body: %s", status, respBody)
	}
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["contentTypeId"] != "ct-1" {
		t.Errorf("contentTypeId: got %v, want %q", result["contentTypeId"], "ct-1")
	}
}

func TestCreateContentItem_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	app := newContentItemTestApp(&mockContentItemRepo{})

	body := strings.NewReader(`not-json`)
	status, _ := contentBodyOf(t, app, "POST", "/content-types/ct-1/items", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestCreateContentItem_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentItemRepo{
		CreateContentItemFn: func(_ context.Context, _ *models.ContentItem, _ []models.ContentFieldValue) (*models.ContentItem, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newContentItemTestApp(repo)

	body := strings.NewReader(`{"title":"New Post"}`)
	status, _ := contentBodyOf(t, app, "POST", "/content-types/ct-1/items", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestCreateContentItem_SetsContentTypeIDFromRouteParam(t *testing.T) {
	// ContentTypeId must be populated from the route, not the body.
	var capturedContentTypeID string
	repo := &mockContentItemRepo{
		CreateContentItemFn: func(_ context.Context, item *models.ContentItem, _ []models.ContentFieldValue) (*models.ContentItem, error) {
			capturedContentTypeID = item.ContentTypeId
			return item, nil
		},
	}
	app := newContentItemTestApp(repo)

	body := strings.NewReader(`{"title":"Post"}`)
	status, _ := contentBodyOf(t, app, "POST", "/content-types/my-ct-id/items", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedContentTypeID != "my-ct-id" {
		t.Errorf("contentTypeId: got %q, want %q", capturedContentTypeID, "my-ct-id")
	}
}

// ─── UpdateContentItem ────────────────────────────────────────────────────────

func TestUpdateContentItem_Returns200OnSuccess(t *testing.T) {
	repo := &mockContentItemRepo{
		UpdateContentItemFn: func(_ context.Context, id string, _ map[string]any, _ []models.ContentFieldValue) (*models.ContentItem, error) {
			return makeContentItem(id, "ct-1", "Updated Post", "updated-post", true), nil
		},
	}
	app := newContentItemTestApp(repo)

	body := strings.NewReader(`{"title":"Updated Post","published":true}`)
	status, respBody := contentBodyOf(t, app, "PATCH", "/items/i-1", body, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, respBody)
	}
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["title"] != "Updated Post" {
		t.Errorf("title: got %v, want %q", result["title"], "Updated Post")
	}
}

func TestUpdateContentItem_Returns404WhenNotFound(t *testing.T) {
	repo := &mockContentItemRepo{
		UpdateContentItemFn: func(_ context.Context, _ string, _ map[string]any, _ []models.ContentFieldValue) (*models.ContentItem, error) {
			return nil, repositories.ErrContentItemNotFound
		},
	}
	app := newContentItemTestApp(repo)

	body := strings.NewReader(`{"title":"Updated"}`)
	status, _ := contentBodyOf(t, app, "PATCH", "/items/ghost", body, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestUpdateContentItem_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentItemRepo{
		UpdateContentItemFn: func(_ context.Context, _ string, _ map[string]any, _ []models.ContentFieldValue) (*models.ContentItem, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newContentItemTestApp(repo)

	body := strings.NewReader(`{"title":"Updated"}`)
	status, _ := contentBodyOf(t, app, "PATCH", "/items/i-1", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

// ─── DeleteContentItem ────────────────────────────────────────────────────────

func TestDeleteContentItem_Returns204OnSuccess(t *testing.T) {
	repo := &mockContentItemRepo{
		DeleteContentItemFn: func(_ context.Context, _ string) error { return nil },
	}
	app := newContentItemTestApp(repo)

	if code := contentStatusOf(t, app, "DELETE", "/items/i-1"); code != fiber.StatusNoContent {
		t.Errorf("expected 204, got %d", code)
	}
}

func TestDeleteContentItem_Returns404WhenNotFound(t *testing.T) {
	repo := &mockContentItemRepo{
		DeleteContentItemFn: func(_ context.Context, _ string) error {
			return repositories.ErrContentItemNotFound
		},
	}
	app := newContentItemTestApp(repo)

	if code := contentStatusOf(t, app, "DELETE", "/items/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeleteContentItem_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentItemRepo{
		DeleteContentItemFn: func(_ context.Context, _ string) error {
			return errors.New("db error")
		},
	}
	app := newContentItemTestApp(repo)

	if code := contentStatusOf(t, app, "DELETE", "/items/i-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── GetPublicContentItems ────────────────────────────────────────────────────

func TestGetPublicContentItems_ReturnsItemsAsFlattened(t *testing.T) {
	// The handler flattens each item (merging field values into the top-level
	// map) before returning. With no field values the core fields should still
	// be present.
	repo := &mockContentItemRepo{
		GetPublicContentItemsFn: func(_ context.Context, contentTypeID string, limit int, _, _ string) ([]models.ContentItem, error) {
			return []models.ContentItem{
				*makeContentItem("i-1", contentTypeID, "Hello", "hello", true),
			}, nil
		},
	}
	app := newContentItemTestApp(repo)

	status, body := contentBodyOf(t, app, "GET", "/public/items?contentTypeId=ct-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}
	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 item, got %d", len(result))
	}
	if result[0]["title"] != "Hello" {
		t.Errorf("title: got %v, want %q", result[0]["title"], "Hello")
	}
}

func TestGetPublicContentItems_ReturnsEmptySlice(t *testing.T) {
	repo := &mockContentItemRepo{
		GetPublicContentItemsFn: func(_ context.Context, _ string, _ int, _, _ string) ([]models.ContentItem, error) {
			return []models.ContentItem{}, nil
		},
	}
	app := newContentItemTestApp(repo)

	status, body := contentBodyOf(t, app, "GET", "/public/items", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}
	var result []any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d", len(result))
	}
}

func TestGetPublicContentItems_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentItemRepo{
		GetPublicContentItemsFn: func(_ context.Context, _ string, _ int, _, _ string) ([]models.ContentItem, error) {
			return nil, errors.New("db error")
		},
	}
	app := newContentItemTestApp(repo)

	if code := contentStatusOf(t, app, "GET", "/public/items"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── GetPublicContentItemBySlug ───────────────────────────────────────────────

func TestGetPublicContentItemBySlug_ReturnsItemWhenPublished(t *testing.T) {
	// A published item matching slug + contentTypeId must be returned as a
	// flattened JSON object.
	repo := &mockContentItemRepo{
		GetContentItemBySlugFn: func(_ context.Context, contentTypeID, slug string) (*models.ContentItem, error) {
			if contentTypeID == "ct-1" && slug == "hello-world" {
				return makeContentItem("i-1", contentTypeID, "Hello World", slug, true), nil
			}
			return nil, repositories.ErrContentItemNotFound
		},
	}
	app := newContentItemTestApp(repo)

	status, body := contentBodyOf(t, app, "GET", "/public/content-types/ct-1/items/hello-world", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["slug"] != "hello-world" {
		t.Errorf("slug: got %v, want %q", result["slug"], "hello-world")
	}
}

func TestGetPublicContentItemBySlug_Returns404WhenItemIsUnpublished(t *testing.T) {
	// An item with published=false must be treated as not found even if it
	// exists in the database.
	repo := &mockContentItemRepo{
		GetContentItemBySlugFn: func(_ context.Context, _, _ string) (*models.ContentItem, error) {
			// published=false
			return makeContentItem("i-1", "ct-1", "Draft Post", "draft", false), nil
		},
	}
	app := newContentItemTestApp(repo)

	if code := contentStatusOf(t, app, "GET", "/public/content-types/ct-1/items/draft"); code != fiber.StatusNotFound {
		t.Errorf("expected 404 for unpublished item, got %d", code)
	}
}

func TestGetPublicContentItemBySlug_Returns404WhenNotFound(t *testing.T) {
	repo := &mockContentItemRepo{
		GetContentItemBySlugFn: func(_ context.Context, _, _ string) (*models.ContentItem, error) {
			return nil, repositories.ErrContentItemNotFound
		},
	}
	app := newContentItemTestApp(repo)

	if code := contentStatusOf(t, app, "GET", "/public/content-types/ct-1/items/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetPublicContentItemBySlug_Returns500OnRepositoryError(t *testing.T) {
	repo := &mockContentItemRepo{
		GetContentItemBySlugFn: func(_ context.Context, _, _ string) (*models.ContentItem, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newContentItemTestApp(repo)

	if code := contentStatusOf(t, app, "GET", "/public/content-types/ct-1/items/my-slug"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}