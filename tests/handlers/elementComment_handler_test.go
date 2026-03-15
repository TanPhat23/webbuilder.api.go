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
	"my-go-app/pkg/utils"
	"my-go-app/tests/testutil"

	"github.com/gofiber/fiber/v2"
)

// ─── test app factories ───────────────────────────────────────────────────────

func newElementCommentTestApp(repo *testutil.MockElementCommentRepository) *fiber.App {
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

	h := handlers.NewElementCommentHandler(repo)

	app.Post("/element-comments", h.CreateElementComment)
	app.Get("/element-comments/:id", h.GetElementCommentByID)
	app.Get("/elements/:elementId/comments", h.GetElementComments)
	app.Patch("/element-comments/:id", h.UpdateElementComment)
	app.Delete("/element-comments/:id", h.DeleteElementComment)
	app.Patch("/element-comments/:id/toggle-resolved", h.ToggleResolvedStatus)
	app.Get("/element-comments/author/:authorId", h.GetCommentsByAuthorID)
	app.Get("/projects/:projectId/comments", h.GetCommentsByProjectID)

	return app
}

func newElementCommentTestAppNoAuth(repo *testutil.MockElementCommentRepository) *fiber.App {
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

	h := handlers.NewElementCommentHandler(repo)

	app.Post("/element-comments", h.CreateElementComment)
	app.Patch("/element-comments/:id", h.UpdateElementComment)
	app.Delete("/element-comments/:id", h.DeleteElementComment)

	return app
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func ecStatusOf(t *testing.T, app *fiber.App, method, path string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp.StatusCode
}

func ecBodyOf(t *testing.T, app *fiber.App, method, path string, body io.Reader, contentType string) (int, []byte) {
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

func makeElementComment(id, elementID, authorID, content string) *models.ElementComment {
	return &models.ElementComment{
		Id:        id,
		Content:   content,
		AuthorId:  authorID,
		ElementId: elementID,
		Resolved:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ─── CreateElementComment ─────────────────────────────────────────────────────

func TestCreateElementComment_Returns201OnSuccess(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		CreateElementCommentFn: func(_ context.Context, comment *models.ElementComment) (*models.ElementComment, error) {
			return comment, nil
		},
	}
	app := newElementCommentTestApp(repo)

	payload := strings.NewReader(`{"content":"Nice element!","elementId":"el-1"}`)
	status, body := ecBodyOf(t, app, "POST", "/element-comments", payload, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["content"] != "Nice element!" {
		t.Errorf("content: got %v, want %q", result["content"], "Nice element!")
	}
}

func TestCreateElementComment_Returns422WhenContentMissing(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	app := newElementCommentTestApp(repo)

	payload := strings.NewReader(`{"elementId":"el-1"}`)
	status, _ := ecBodyOf(t, app, "POST", "/element-comments", payload, "application/json")
	if status != fiber.StatusUnprocessableEntity {
		t.Errorf("expected 422 for missing content, got %d", status)
	}
}

func TestCreateElementComment_Returns422WhenElementIDMissing(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	app := newElementCommentTestApp(repo)

	payload := strings.NewReader(`{"content":"Nice element!"}`)
	status, _ := ecBodyOf(t, app, "POST", "/element-comments", payload, "application/json")
	if status != fiber.StatusUnprocessableEntity {
		t.Errorf("expected 422 for missing elementId, got %d", status)
	}
}

func TestCreateElementComment_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	app := newElementCommentTestApp(repo)

	payload := strings.NewReader(`{invalid}`)
	status, _ := ecBodyOf(t, app, "POST", "/element-comments", payload, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", status)
	}
}

func TestCreateElementComment_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		CreateElementCommentFn: func(_ context.Context, _ *models.ElementComment) (*models.ElementComment, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newElementCommentTestApp(repo)

	payload := strings.NewReader(`{"content":"Nice!","elementId":"el-1"}`)
	status, _ := ecBodyOf(t, app, "POST", "/element-comments", payload, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestCreateElementComment_Returns401WhenNoUserID(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	app := newElementCommentTestAppNoAuth(repo)

	payload := strings.NewReader(`{"content":"Nice!","elementId":"el-1"}`)
	status, _ := ecBodyOf(t, app, "POST", "/element-comments", payload, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

func TestCreateElementComment_SetsAuthorIDFromLocals(t *testing.T) {
	var capturedAuthorID string
	repo := &testutil.MockElementCommentRepository{
		CreateElementCommentFn: func(_ context.Context, comment *models.ElementComment) (*models.ElementComment, error) {
			capturedAuthorID = comment.AuthorId
			return comment, nil
		},
	}
	app := newElementCommentTestApp(repo)

	payload := strings.NewReader(`{"content":"Nice!","elementId":"el-1"}`)
	status, _ := ecBodyOf(t, app, "POST", "/element-comments", payload, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedAuthorID != "test-user-id" {
		t.Errorf("authorId: got %q, want %q", capturedAuthorID, "test-user-id")
	}
}

func TestCreateElementComment_AssignsNewID(t *testing.T) {
	var capturedID string
	repo := &testutil.MockElementCommentRepository{
		CreateElementCommentFn: func(_ context.Context, comment *models.ElementComment) (*models.ElementComment, error) {
			capturedID = comment.Id
			return comment, nil
		},
	}
	app := newElementCommentTestApp(repo)

	payload := strings.NewReader(`{"content":"Nice!","elementId":"el-1"}`)
	status, _ := ecBodyOf(t, app, "POST", "/element-comments", payload, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedID == "" {
		t.Error("expected a non-empty ID to be assigned, got empty string")
	}
}

// ─── GetElementCommentByID ────────────────────────────────────────────────────

func TestGetElementCommentByID_ReturnsCommentWhenFound(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			if id == "ec-1" {
				return makeElementComment(id, "el-1", "test-user-id", "Hello"), nil
			}
			return nil, nil
		},
	}
	app := newElementCommentTestApp(repo)

	status, body := ecBodyOf(t, app, "GET", "/element-comments/ec-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["id"] != "ec-1" {
		t.Errorf("id: got %v, want %q", result["id"], "ec-1")
	}
}

func TestGetElementCommentByID_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, _ string) (*models.ElementComment, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newElementCommentTestApp(repo)

	if code := ecStatusOf(t, app, "GET", "/element-comments/ec-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetElementCommentByID_PassesIDToRepository(t *testing.T) {
	var capturedID string
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			capturedID = id
			return makeElementComment(id, "el-1", "test-user-id", "Hello"), nil
		},
	}
	app := newElementCommentTestApp(repo)

	status := ecStatusOf(t, app, "GET", "/element-comments/my-comment-id")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedID != "my-comment-id" {
		t.Errorf("id passed to repo: got %q, want %q", capturedID, "my-comment-id")
	}
}

// ─── GetElementComments ───────────────────────────────────────────────────────

func TestGetElementComments_ReturnsCommentsForElement(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsFn: func(_ context.Context, elementID string, _ *models.ElementCommentFilter) ([]models.ElementComment, error) {
			if elementID != "el-1" {
				return nil, errors.New("unexpected elementID: " + elementID)
			}
			return []models.ElementComment{
				*makeElementComment("ec-1", elementID, "test-user-id", "First"),
				*makeElementComment("ec-2", elementID, "test-user-id", "Second"),
			}, nil
		},
		CountElementCommentsFn: func(_ context.Context, _ string) (int64, error) {
			return 2, nil
		},
	}
	app := newElementCommentTestApp(repo)

	status, body := ecBodyOf(t, app, "GET", "/elements/el-1/comments", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	data, ok := result["data"].([]any)
	if !ok {
		t.Fatalf("expected data to be an array, got: %T", result["data"])
	}
	if len(data) != 2 {
		t.Errorf("expected 2 comments, got %d", len(data))
	}
	if result["total"] != float64(2) {
		t.Errorf("total: got %v, want 2", result["total"])
	}
}

func TestGetElementComments_ReturnsEmptySliceWhenNone(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsFn: func(_ context.Context, _ string, _ *models.ElementCommentFilter) ([]models.ElementComment, error) {
			return []models.ElementComment{}, nil
		},
		CountElementCommentsFn: func(_ context.Context, _ string) (int64, error) {
			return 0, nil
		},
	}
	app := newElementCommentTestApp(repo)

	status, body := ecBodyOf(t, app, "GET", "/elements/el-1/comments", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	data, ok := result["data"].([]any)
	if !ok {
		t.Fatalf("expected data to be an array")
	}
	if len(data) != 0 {
		t.Errorf("expected empty data array, got %d", len(data))
	}
}

func TestGetElementComments_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsFn: func(_ context.Context, _ string, _ *models.ElementCommentFilter) ([]models.ElementComment, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newElementCommentTestApp(repo)

	if code := ecStatusOf(t, app, "GET", "/elements/el-1/comments"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetElementComments_PassesElementIDToRepository(t *testing.T) {
	var capturedElementID string
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsFn: func(_ context.Context, elementID string, _ *models.ElementCommentFilter) ([]models.ElementComment, error) {
			capturedElementID = elementID
			return []models.ElementComment{}, nil
		},
	}
	app := newElementCommentTestApp(repo)

	status := ecStatusOf(t, app, "GET", "/elements/my-element/comments")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedElementID != "my-element" {
		t.Errorf("elementID passed to repo: got %q, want %q", capturedElementID, "my-element")
	}
}

// ─── UpdateElementComment ─────────────────────────────────────────────────────

func TestUpdateElementComment_Returns200OnSuccess(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			return makeElementComment(id, "el-1", "test-user-id", "Original"), nil
		},
		UpdateElementCommentFn: func(_ context.Context, id string, updates map[string]any) (*models.ElementComment, error) {
			content := updates["Content"].(string)
			return makeElementComment(id, "el-1", "test-user-id", content), nil
		},
	}
	app := newElementCommentTestApp(repo)

	payload := strings.NewReader(`{"content":"Updated content"}`)
	status, body := ecBodyOf(t, app, "PATCH", "/element-comments/ec-1", payload, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["content"] != "Updated content" {
		t.Errorf("content: got %v, want %q", result["content"], "Updated content")
	}
}

func TestUpdateElementComment_Returns400WhenBodyIsEmpty(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			return makeElementComment(id, "el-1", "test-user-id", "Original"), nil
		},
	}
	app := newElementCommentTestApp(repo)

	payload := strings.NewReader(`{}`)
	status, _ := ecBodyOf(t, app, "PATCH", "/element-comments/ec-1", payload, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for empty update body, got %d", status)
	}
}

func TestUpdateElementComment_Returns400WhenBodyIsInvalidJSON(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	app := newElementCommentTestApp(repo)

	payload := strings.NewReader(`not-json`)
	status, _ := ecBodyOf(t, app, "PATCH", "/element-comments/ec-1", payload, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", status)
	}
}

func TestUpdateElementComment_Returns403WhenNotAuthor(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			return makeElementComment(id, "el-1", "another-user-id", "Original"), nil
		},
	}
	app := newElementCommentTestApp(repo)

	payload := strings.NewReader(`{"content":"Hijacked!"}`)
	status, _ := ecBodyOf(t, app, "PATCH", "/element-comments/ec-1", payload, "application/json")
	if status != fiber.StatusForbidden {
		t.Errorf("expected 403 when non-author tries to update, got %d", status)
	}
}

func TestUpdateElementComment_Returns500OnGetError(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, _ string) (*models.ElementComment, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newElementCommentTestApp(repo)

	payload := strings.NewReader(`{"content":"Updated"}`)
	status, _ := ecBodyOf(t, app, "PATCH", "/element-comments/ec-1", payload, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestUpdateElementComment_Returns500OnUpdateError(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			return makeElementComment(id, "el-1", "test-user-id", "Original"), nil
		},
		UpdateElementCommentFn: func(_ context.Context, _ string, _ map[string]any) (*models.ElementComment, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newElementCommentTestApp(repo)

	payload := strings.NewReader(`{"content":"Updated"}`)
	status, _ := ecBodyOf(t, app, "PATCH", "/element-comments/ec-1", payload, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestUpdateElementComment_Returns401WhenNoUserID(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	app := newElementCommentTestAppNoAuth(repo)

	payload := strings.NewReader(`{"content":"Updated"}`)
	status, _ := ecBodyOf(t, app, "PATCH", "/element-comments/ec-1", payload, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

func TestUpdateElementComment_CanToggleResolvedField(t *testing.T) {
	var capturedUpdates map[string]any
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			return makeElementComment(id, "el-1", "test-user-id", "Original"), nil
		},
		UpdateElementCommentFn: func(_ context.Context, id string, updates map[string]any) (*models.ElementComment, error) {
			capturedUpdates = updates
			ce := makeElementComment(id, "el-1", "test-user-id", "Original")
			ce.Resolved = true
			return ce, nil
		},
	}
	app := newElementCommentTestApp(repo)

	payload := strings.NewReader(`{"resolved":true}`)
	status, _ := ecBodyOf(t, app, "PATCH", "/element-comments/ec-1", payload, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if resolved, ok := capturedUpdates["Resolved"].(bool); !ok || !resolved {
		t.Errorf("expected Resolved=true in updates, got %v", capturedUpdates)
	}
}

// ─── DeleteElementComment ─────────────────────────────────────────────────────

func TestDeleteElementComment_Returns204OnSuccess(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			return makeElementComment(id, "el-1", "test-user-id", "Hello"), nil
		},
		DeleteElementCommentFn: func(_ context.Context, _ string) error {
			return nil
		},
	}
	app := newElementCommentTestApp(repo)

	status, _ := ecBodyOf(t, app, "DELETE", "/element-comments/ec-1", nil, "")
	if status != fiber.StatusNoContent {
		t.Errorf("expected 204, got %d", status)
	}
}

func TestDeleteElementComment_Returns403WhenNotAuthor(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			return makeElementComment(id, "el-1", "another-user-id", "Hello"), nil
		},
	}
	app := newElementCommentTestApp(repo)

	if code := ecStatusOf(t, app, "DELETE", "/element-comments/ec-1"); code != fiber.StatusForbidden {
		t.Errorf("expected 403 when non-author tries to delete, got %d", code)
	}
}

func TestDeleteElementComment_Returns500OnGetError(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, _ string) (*models.ElementComment, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newElementCommentTestApp(repo)

	if code := ecStatusOf(t, app, "DELETE", "/element-comments/ec-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestDeleteElementComment_Returns500OnDeleteError(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			return makeElementComment(id, "el-1", "test-user-id", "Hello"), nil
		},
		DeleteElementCommentFn: func(_ context.Context, _ string) error {
			return errors.New("db write error")
		},
	}
	app := newElementCommentTestApp(repo)

	if code := ecStatusOf(t, app, "DELETE", "/element-comments/ec-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestDeleteElementComment_Returns401WhenNoUserID(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{}
	app := newElementCommentTestAppNoAuth(repo)

	if code := ecStatusOf(t, app, "DELETE", "/element-comments/ec-1"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

func TestDeleteElementComment_PassesCorrectIDToRepository(t *testing.T) {
	var capturedID string
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentByIDFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			return makeElementComment(id, "el-1", "test-user-id", "Hello"), nil
		},
		DeleteElementCommentFn: func(_ context.Context, id string) error {
			capturedID = id
			return nil
		},
	}
	app := newElementCommentTestApp(repo)

	if code := ecStatusOf(t, app, "DELETE", "/element-comments/target-comment"); code != fiber.StatusNoContent {
		t.Fatalf("expected 204, got %d", code)
	}
	if capturedID != "target-comment" {
		t.Errorf("id passed to repo: got %q, want %q", capturedID, "target-comment")
	}
}

// ─── ToggleResolvedStatus ─────────────────────────────────────────────────────

func TestToggleResolvedStatus_Returns200OnSuccess(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		ToggleResolvedStatusFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			ce := makeElementComment(id, "el-1", "test-user-id", "Hello")
			ce.Resolved = true
			return ce, nil
		},
	}
	app := newElementCommentTestApp(repo)

	status, body := ecBodyOf(t, app, "PATCH", "/element-comments/ec-1/toggle-resolved", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["resolved"] != true {
		t.Errorf("resolved: got %v, want true", result["resolved"])
	}
}

func TestToggleResolvedStatus_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		ToggleResolvedStatusFn: func(_ context.Context, _ string) (*models.ElementComment, error) {
			return nil, errors.New("db write error")
		},
	}
	app := newElementCommentTestApp(repo)

	if code := ecStatusOf(t, app, "PATCH", "/element-comments/ec-1/toggle-resolved"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestToggleResolvedStatus_PassesCorrectIDToRepository(t *testing.T) {
	var capturedID string
	repo := &testutil.MockElementCommentRepository{
		ToggleResolvedStatusFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			capturedID = id
			return makeElementComment(id, "el-1", "test-user-id", "Hello"), nil
		},
	}
	app := newElementCommentTestApp(repo)

	status := ecStatusOf(t, app, "PATCH", "/element-comments/toggle-target/toggle-resolved")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedID != "toggle-target" {
		t.Errorf("id passed to repo: got %q, want %q", capturedID, "toggle-target")
	}
}

func TestToggleResolvedStatus_CanUnresolve(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		ToggleResolvedStatusFn: func(_ context.Context, id string) (*models.ElementComment, error) {
			ce := makeElementComment(id, "el-1", "test-user-id", "Hello")
			ce.Resolved = false
			return ce, nil
		},
	}
	app := newElementCommentTestApp(repo)

	status, body := ecBodyOf(t, app, "PATCH", "/element-comments/ec-1/toggle-resolved", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["resolved"] != false {
		t.Errorf("resolved: got %v, want false", result["resolved"])
	}
}

// ─── GetCommentsByAuthorID ────────────────────────────────────────────────────

func TestGetCommentsByAuthorID_ReturnsCommentsForAuthor(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsByAuthorIDFn: func(_ context.Context, authorID string, _, _ int) ([]models.ElementComment, error) {
			if authorID != "author-1" {
				return nil, errors.New("unexpected authorID: " + authorID)
			}
			return []models.ElementComment{
				*makeElementComment("ec-1", "el-1", authorID, "First"),
				*makeElementComment("ec-2", "el-2", authorID, "Second"),
			}, nil
		},
	}
	app := newElementCommentTestApp(repo)

	status, body := ecBodyOf(t, app, "GET", "/element-comments/author/author-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 comments, got %d", len(result))
	}
}

func TestGetCommentsByAuthorID_ReturnsEmptySliceWhenNone(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsByAuthorIDFn: func(_ context.Context, _ string, _, _ int) ([]models.ElementComment, error) {
			return []models.ElementComment{}, nil
		},
	}
	app := newElementCommentTestApp(repo)

	status, body := ecBodyOf(t, app, "GET", "/element-comments/author/author-1", nil, "")
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

func TestGetCommentsByAuthorID_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsByAuthorIDFn: func(_ context.Context, _ string, _, _ int) ([]models.ElementComment, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newElementCommentTestApp(repo)

	if code := ecStatusOf(t, app, "GET", "/element-comments/author/author-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetCommentsByAuthorID_PassesAuthorIDToRepository(t *testing.T) {
	var capturedAuthorID string
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsByAuthorIDFn: func(_ context.Context, authorID string, _, _ int) ([]models.ElementComment, error) {
			capturedAuthorID = authorID
			return []models.ElementComment{}, nil
		},
	}
	app := newElementCommentTestApp(repo)

	status := ecStatusOf(t, app, "GET", "/element-comments/author/my-author-id")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedAuthorID != "my-author-id" {
		t.Errorf("authorID passed to repo: got %q, want %q", capturedAuthorID, "my-author-id")
	}
}

// ─── GetCommentsByProjectID ───────────────────────────────────────────────────

func TestGetCommentsByProjectID_ReturnsCommentsForProject(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsByProjectIDFn: func(_ context.Context, projectID string, _, _ int) ([]models.ElementComment, error) {
			if projectID != "proj-1" {
				return nil, errors.New("unexpected projectID: " + projectID)
			}
			return []models.ElementComment{
				*makeElementComment("ec-1", "el-1", "user-1", "First"),
				*makeElementComment("ec-2", "el-2", "user-2", "Second"),
			}, nil
		},
		CountElementCommentsByProjectIDFn: func(_ context.Context, _ string) (int64, error) {
			return 2, nil
		},
	}
	app := newElementCommentTestApp(repo)

	status, body := ecBodyOf(t, app, "GET", "/projects/proj-1/comments", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	data, ok := result["data"].([]any)
	if !ok {
		t.Fatalf("expected data to be an array, got: %T", result["data"])
	}
	if len(data) != 2 {
		t.Errorf("expected 2 comments, got %d", len(data))
	}
	if result["total"] != float64(2) {
		t.Errorf("total: got %v, want 2", result["total"])
	}
}

func TestGetCommentsByProjectID_ReturnsEmptySliceWhenNone(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsByProjectIDFn: func(_ context.Context, _ string, _, _ int) ([]models.ElementComment, error) {
			return []models.ElementComment{}, nil
		},
		CountElementCommentsByProjectIDFn: func(_ context.Context, _ string) (int64, error) {
			return 0, nil
		},
	}
	app := newElementCommentTestApp(repo)

	status, body := ecBodyOf(t, app, "GET", "/projects/proj-1/comments", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	data, ok := result["data"].([]any)
	if !ok {
		t.Fatalf("expected data to be an array")
	}
	if len(data) != 0 {
		t.Errorf("expected empty data, got %d", len(data))
	}
}

func TestGetCommentsByProjectID_Returns500OnRepositoryError(t *testing.T) {
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsByProjectIDFn: func(_ context.Context, _ string, _, _ int) ([]models.ElementComment, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newElementCommentTestApp(repo)

	if code := ecStatusOf(t, app, "GET", "/projects/proj-1/comments"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestGetCommentsByProjectID_PassesProjectIDToRepository(t *testing.T) {
	var capturedProjectID string
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsByProjectIDFn: func(_ context.Context, projectID string, _, _ int) ([]models.ElementComment, error) {
			capturedProjectID = projectID
			return []models.ElementComment{}, nil
		},
	}
	app := newElementCommentTestApp(repo)

	status := ecStatusOf(t, app, "GET", "/projects/my-project/comments")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedProjectID != "my-project" {
		t.Errorf("projectID passed to repo: got %q, want %q", capturedProjectID, "my-project")
	}
}

func TestGetCommentsByProjectID_IncludesPaginationDefaults(t *testing.T) {
	var capturedLimit, capturedOffset int
	repo := &testutil.MockElementCommentRepository{
		GetElementCommentsByProjectIDFn: func(_ context.Context, _ string, limit, offset int) ([]models.ElementComment, error) {
			capturedLimit = limit
			capturedOffset = offset
			return []models.ElementComment{}, nil
		},
	}
	app := newElementCommentTestApp(repo)

	status := ecStatusOf(t, app, "GET", "/projects/proj-1/comments")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedLimit != 20 {
		t.Errorf("default limit: got %d, want 20", capturedLimit)
	}
	if capturedOffset != 0 {
		t.Errorf("default offset: got %d, want 0", capturedOffset)
	}
}