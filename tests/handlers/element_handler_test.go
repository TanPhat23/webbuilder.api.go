package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"my-go-app/internal/handlers"
	"my-go-app/internal/models"
	"my-go-app/tests/testutil"

	"github.com/gofiber/fiber/v2"
)

// ─── test app factory ─────────────────────────────────────────────────────────

func newElementTestApp(elementRepo *testutil.MockElementRepository) *fiber.App {
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

	h := handlers.NewElementHandler(elementRepo)

	app.Get("/projects/:projectid/elements", h.GetElements)
	app.Get("/elements", h.GetElementsByPageIds)

	return app
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func elementStatusOf(t *testing.T, app *fiber.App, method, path string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp.StatusCode
}

func elementBodyOf(t *testing.T, app *fiber.App, method, path string) (int, []byte) {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
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

func makeEditorElement(id, pageID string) models.Element {
	return models.Element{
		Id:     id,
		Type:   "Text",
		PageId: &pageID,
	}
}

// ─── GetElements ──────────────────────────────────────────────────────────────

func TestGetElements_ReturnsEmptySliceWhenNone(t *testing.T) {
	repo := &testutil.MockElementRepository{
		GetElementsFn: func(_ context.Context, projectID string, _ ...string) ([]models.EditorElement, error) {
			return []models.EditorElement{}, nil
		},
	}
	app := newElementTestApp(repo)

	status, body := elementBodyOf(t, app, "GET", "/projects/proj-1/elements")
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

func TestGetElements_ReturnsElementsForProject(t *testing.T) {
	pageID := "page-1"
	el1 := makeEditorElement("el-1", pageID)
	el2 := makeEditorElement("el-2", pageID)

	repo := &testutil.MockElementRepository{
		GetElementsFn: func(_ context.Context, projectID string, _ ...string) ([]models.EditorElement, error) {
			if projectID != "proj-1" {
				return nil, errors.New("unexpected projectID: " + projectID)
			}
			return []models.EditorElement{&el1, &el2}, nil
		},
	}
	app := newElementTestApp(repo)

	status, body := elementBodyOf(t, app, "GET", "/projects/proj-1/elements")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 elements, got %d", len(result))
	}
}

func TestGetElements_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockElementRepository{
		GetElementsFn: func(_ context.Context, _ string, _ ...string) ([]models.EditorElement, error) {
			return nil, errors.New("db connection lost")
		},
	}
	app := newElementTestApp(repo)

	if code := elementStatusOf(t, app, "GET", "/projects/proj-1/elements"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetElements_PassesProjectIDToRepository(t *testing.T) {
	var capturedProjectID string
	repo := &testutil.MockElementRepository{
		GetElementsFn: func(_ context.Context, projectID string, _ ...string) ([]models.EditorElement, error) {
			capturedProjectID = projectID
			return []models.EditorElement{}, nil
		},
	}
	app := newElementTestApp(repo)

	status := elementStatusOf(t, app, "GET", "/projects/my-project/elements")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedProjectID != "my-project" {
		t.Errorf("projectID passed to repo: got %q, want %q", capturedProjectID, "my-project")
	}
}

// ─── GetElementsByPageIds ─────────────────────────────────────────────────────

func TestGetElementsByPageIds_ReturnsElementsForPages(t *testing.T) {
	pageID := "page-1"
	el := makeEditorElement("el-1", pageID)

	repo := &testutil.MockElementRepository{
		GetElementsByPageIdsFn: func(_ context.Context, pageIDs []string) ([]models.EditorElement, error) {
			if len(pageIDs) != 1 || pageIDs[0] != "page-1" {
				return nil, errors.New("unexpected pageIDs")
			}
			return []models.EditorElement{&el}, nil
		},
	}
	app := newElementTestApp(repo)

	status, body := elementBodyOf(t, app, "GET", "/elements?pageIds=page-1")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 element, got %d", len(result))
	}
}

func TestGetElementsByPageIds_ReturnsElementsForMultiplePages(t *testing.T) {
	pageID1 := "page-1"
	pageID2 := "page-2"
	el1 := makeEditorElement("el-1", pageID1)
	el2 := makeEditorElement("el-2", pageID2)

	repo := &testutil.MockElementRepository{
		GetElementsByPageIdsFn: func(_ context.Context, pageIDs []string) ([]models.EditorElement, error) {
			if len(pageIDs) != 2 {
				return nil, errors.New("unexpected pageIDs count")
			}
			return []models.EditorElement{&el1, &el2}, nil
		},
	}
	app := newElementTestApp(repo)

	status, body := elementBodyOf(t, app, "GET", "/elements?pageIds=page-1,page-2")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 elements, got %d", len(result))
	}
}

func TestGetElementsByPageIds_Returns400WhenQueryParamMissing(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	app := newElementTestApp(repo)

	if code := elementStatusOf(t, app, "GET", "/elements"); code != fiber.StatusBadRequest {
		t.Errorf("expected 400 for missing pageIds, got %d", code)
	}
}

func TestGetElementsByPageIds_Returns400WhenQueryParamIsBlank(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	app := newElementTestApp(repo)

	if code := elementStatusOf(t, app, "GET", "/elements?pageIds="); code != fiber.StatusBadRequest {
		t.Errorf("expected 400 for blank pageIds, got %d", code)
	}
}

func TestGetElementsByPageIds_Returns400WhenAllIDsAreWhitespace(t *testing.T) {
	repo := &testutil.MockElementRepository{}
	app := newElementTestApp(repo)

	if code := elementStatusOf(t, app, "GET", "/elements?pageIds=+,+"); code != fiber.StatusBadRequest {
		t.Errorf("expected 400 when all IDs are whitespace, got %d", code)
	}
}

func TestGetElementsByPageIds_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockElementRepository{
		GetElementsByPageIdsFn: func(_ context.Context, _ []string) ([]models.EditorElement, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newElementTestApp(repo)

	if code := elementStatusOf(t, app, "GET", "/elements?pageIds=page-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetElementsByPageIds_PassesPageIDsToRepository(t *testing.T) {
	var capturedPageIDs []string
	repo := &testutil.MockElementRepository{
		GetElementsByPageIdsFn: func(_ context.Context, pageIDs []string) ([]models.EditorElement, error) {
			capturedPageIDs = pageIDs
			return []models.EditorElement{}, nil
		},
	}
	app := newElementTestApp(repo)

	status := elementStatusOf(t, app, "GET", "/elements?pageIds=page-a,page-b")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if len(capturedPageIDs) != 2 {
		t.Fatalf("expected 2 page IDs passed to repo, got %d", len(capturedPageIDs))
	}
	if capturedPageIDs[0] != "page-a" || capturedPageIDs[1] != "page-b" {
		t.Errorf("pageIDs passed to repo: got %v, want [page-a page-b]", capturedPageIDs)
	}
}

func TestGetElementsByPageIds_StripsWhitespaceAroundIDs(t *testing.T) {
	var capturedPageIDs []string
	repo := &testutil.MockElementRepository{
		GetElementsByPageIdsFn: func(_ context.Context, pageIDs []string) ([]models.EditorElement, error) {
			capturedPageIDs = pageIDs
			return []models.EditorElement{}, nil
		},
	}
	app := newElementTestApp(repo)

	status := elementStatusOf(t, app, "GET", "/elements?pageIds=+page-1+,+page-2+")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	for _, id := range capturedPageIDs {
		if id != "page-1" && id != "page-2" {
			t.Errorf("expected trimmed ID but got %q", id)
		}
	}
}

func TestGetElementsByPageIds_ReturnsEmptySliceWhenNone(t *testing.T) {
	repo := &testutil.MockElementRepository{
		GetElementsByPageIdsFn: func(_ context.Context, _ []string) ([]models.EditorElement, error) {
			return []models.EditorElement{}, nil
		},
	}
	app := newElementTestApp(repo)

	status, body := elementBodyOf(t, app, "GET", "/elements?pageIds=page-1")
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
