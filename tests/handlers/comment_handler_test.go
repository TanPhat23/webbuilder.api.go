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

// ─── mock MarketplaceRepository (minimal — only methods called by CommentHandler) ──

// mockMarketplaceRepo is a minimal test double for MarketplaceRepositoryInterface.
// CommentHandler only calls GetMarketplaceItemByID; every other method is a
// no-op stub so the interface is satisfied without importing the full impl.
type mockMarketplaceRepo struct {
	GetMarketplaceItemByIDFn func(id string) (*models.MarketplaceItem, error)
	GetMarketplaceItemsFn    func(filter repositories.MarketplaceFilter) ([]models.MarketplaceItem, int64, error)
	CreateMarketplaceItemFn  func(item models.MarketplaceItem, tagIds []string, categoryIds []string) (*models.MarketplaceItem, error)
	UpdateMarketplaceItemFn  func(id string, userId string, updates map[string]any) (*models.MarketplaceItem, error)
	DeleteMarketplaceItemFn  func(id string, userId string) error
	DownloadMarketplaceItemFn func(itemId string, userId string) (*models.Project, error)
	IncrementDownloadsFn     func(id string) error
	IncrementLikesFn         func(id string) error
	CreateCategoryFn         func(category models.Category) (*models.Category, error)
	GetCategoriesFn          func() ([]models.Category, error)
	GetCategoryByIDFn        func(id string) (*models.Category, error)
	GetCategoryByNameFn      func(name string) (*models.Category, error)
	DeleteCategoryFn         func(id string) error
	CreateTagFn              func(tag models.Tag) (*models.Tag, error)
	GetTagsFn                func() ([]models.Tag, error)
	GetTagByIDFn             func(id string) (*models.Tag, error)
	GetTagByNameFn           func(name string) (*models.Tag, error)
	DeleteTagFn              func(id string) error
}

var _ repositories.MarketplaceRepositoryInterface = (*mockMarketplaceRepo)(nil)

func (m *mockMarketplaceRepo) GetMarketplaceItemByID(id string) (*models.MarketplaceItem, error) {
	if m.GetMarketplaceItemByIDFn != nil {
		return m.GetMarketplaceItemByIDFn(id)
	}
	return nil, nil
}
func (m *mockMarketplaceRepo) GetMarketplaceItems(filter repositories.MarketplaceFilter) ([]models.MarketplaceItem, int64, error) {
	if m.GetMarketplaceItemsFn != nil {
		return m.GetMarketplaceItemsFn(filter)
	}
	return nil, 0, nil
}
func (m *mockMarketplaceRepo) CreateMarketplaceItem(item models.MarketplaceItem, tagIds []string, categoryIds []string) (*models.MarketplaceItem, error) {
	if m.CreateMarketplaceItemFn != nil {
		return m.CreateMarketplaceItemFn(item, tagIds, categoryIds)
	}
	return &item, nil
}
func (m *mockMarketplaceRepo) UpdateMarketplaceItem(id string, userId string, updates map[string]any) (*models.MarketplaceItem, error) {
	if m.UpdateMarketplaceItemFn != nil {
		return m.UpdateMarketplaceItemFn(id, userId, updates)
	}
	return nil, nil
}
func (m *mockMarketplaceRepo) DeleteMarketplaceItem(id string, userId string) error {
	if m.DeleteMarketplaceItemFn != nil {
		return m.DeleteMarketplaceItemFn(id, userId)
	}
	return nil
}
func (m *mockMarketplaceRepo) DownloadMarketplaceItem(itemId string, userId string) (*models.Project, error) {
	if m.DownloadMarketplaceItemFn != nil {
		return m.DownloadMarketplaceItemFn(itemId, userId)
	}
	return nil, nil
}
func (m *mockMarketplaceRepo) IncrementDownloads(id string) error {
	if m.IncrementDownloadsFn != nil {
		return m.IncrementDownloadsFn(id)
	}
	return nil
}
func (m *mockMarketplaceRepo) IncrementLikes(id string) error {
	if m.IncrementLikesFn != nil {
		return m.IncrementLikesFn(id)
	}
	return nil
}
func (m *mockMarketplaceRepo) CreateCategory(category models.Category) (*models.Category, error) {
	if m.CreateCategoryFn != nil {
		return m.CreateCategoryFn(category)
	}
	return &category, nil
}
func (m *mockMarketplaceRepo) GetCategories() ([]models.Category, error) {
	if m.GetCategoriesFn != nil {
		return m.GetCategoriesFn()
	}
	return nil, nil
}
func (m *mockMarketplaceRepo) GetCategoryByID(id string) (*models.Category, error) {
	if m.GetCategoryByIDFn != nil {
		return m.GetCategoryByIDFn(id)
	}
	return nil, nil
}
func (m *mockMarketplaceRepo) GetCategoryByName(name string) (*models.Category, error) {
	if m.GetCategoryByNameFn != nil {
		return m.GetCategoryByNameFn(name)
	}
	return nil, nil
}
func (m *mockMarketplaceRepo) DeleteCategory(id string) error {
	if m.DeleteCategoryFn != nil {
		return m.DeleteCategoryFn(id)
	}
	return nil
}
func (m *mockMarketplaceRepo) CreateTag(tag models.Tag) (*models.Tag, error) {
	if m.CreateTagFn != nil {
		return m.CreateTagFn(tag)
	}
	return &tag, nil
}
func (m *mockMarketplaceRepo) GetTags() ([]models.Tag, error) {
	if m.GetTagsFn != nil {
		return m.GetTagsFn()
	}
	return nil, nil
}
func (m *mockMarketplaceRepo) GetTagByID(id string) (*models.Tag, error) {
	if m.GetTagByIDFn != nil {
		return m.GetTagByIDFn(id)
	}
	return nil, nil
}
func (m *mockMarketplaceRepo) GetTagByName(name string) (*models.Tag, error) {
	if m.GetTagByNameFn != nil {
		return m.GetTagByNameFn(name)
	}
	return nil, nil
}
func (m *mockMarketplaceRepo) DeleteTag(id string) error {
	if m.DeleteTagFn != nil {
		return m.DeleteTagFn(id)
	}
	return nil
}

// ─── mock CommentRepository (local, covers all methods) ───────────────────────

type mockCommentRepo struct {
	CreateCommentFn           func(ctx context.Context, comment models.Comment) (*models.Comment, error)
	GetCommentByIDFn          func(ctx context.Context, id string) (*models.Comment, error)
	GetCommentsFn             func(ctx context.Context, filter models.CommentFilter) ([]models.Comment, int64, error)
	UpdateCommentFn           func(ctx context.Context, id string, userID string, updates map[string]any) (*models.Comment, error)
	DeleteCommentFn           func(ctx context.Context, id string, userID string) error
	CreateReactionFn          func(ctx context.Context, reaction models.CommentReaction) (*models.CommentReaction, error)
	DeleteReactionFn          func(ctx context.Context, commentID string, userID string, reactionType string) error
	GetReactionsByCommentIDFn func(ctx context.Context, commentID string) ([]models.CommentReaction, error)
	GetReactionSummaryFn      func(ctx context.Context, commentID string) ([]models.ReactionSummary, error)
	GetCommentCountByItemIDFn func(ctx context.Context, itemID string) (int64, error)
	ModerateCommentFn         func(ctx context.Context, id string, status string) error
}

var _ repositories.CommentRepositoryInterface = (*mockCommentRepo)(nil)

func (m *mockCommentRepo) CreateComment(ctx context.Context, comment models.Comment) (*models.Comment, error) {
	if m.CreateCommentFn != nil {
		return m.CreateCommentFn(ctx, comment)
	}
	return &comment, nil
}
func (m *mockCommentRepo) GetCommentByID(ctx context.Context, id string) (*models.Comment, error) {
	if m.GetCommentByIDFn != nil {
		return m.GetCommentByIDFn(ctx, id)
	}
	return nil, repositories.ErrCommentNotFound
}
func (m *mockCommentRepo) GetComments(ctx context.Context, filter models.CommentFilter) ([]models.Comment, int64, error) {
	if m.GetCommentsFn != nil {
		return m.GetCommentsFn(ctx, filter)
	}
	return []models.Comment{}, 0, nil
}
func (m *mockCommentRepo) UpdateComment(ctx context.Context, id string, userID string, updates map[string]any) (*models.Comment, error) {
	if m.UpdateCommentFn != nil {
		return m.UpdateCommentFn(ctx, id, userID, updates)
	}
	return nil, repositories.ErrCommentNotFound
}
func (m *mockCommentRepo) DeleteComment(ctx context.Context, id string, userID string) error {
	if m.DeleteCommentFn != nil {
		return m.DeleteCommentFn(ctx, id, userID)
	}
	return nil
}
func (m *mockCommentRepo) CreateReaction(ctx context.Context, reaction models.CommentReaction) (*models.CommentReaction, error) {
	if m.CreateReactionFn != nil {
		return m.CreateReactionFn(ctx, reaction)
	}
	return &reaction, nil
}
func (m *mockCommentRepo) DeleteReaction(ctx context.Context, commentID string, userID string, reactionType string) error {
	if m.DeleteReactionFn != nil {
		return m.DeleteReactionFn(ctx, commentID, userID, reactionType)
	}
	return nil
}
func (m *mockCommentRepo) GetReactionsByCommentID(ctx context.Context, commentID string) ([]models.CommentReaction, error) {
	if m.GetReactionsByCommentIDFn != nil {
		return m.GetReactionsByCommentIDFn(ctx, commentID)
	}
	return []models.CommentReaction{}, nil
}
func (m *mockCommentRepo) GetReactionSummary(ctx context.Context, commentID string) ([]models.ReactionSummary, error) {
	if m.GetReactionSummaryFn != nil {
		return m.GetReactionSummaryFn(ctx, commentID)
	}
	return []models.ReactionSummary{}, nil
}
func (m *mockCommentRepo) GetCommentCountByItemID(ctx context.Context, itemID string) (int64, error) {
	if m.GetCommentCountByItemIDFn != nil {
		return m.GetCommentCountByItemIDFn(ctx, itemID)
	}
	return 0, nil
}
func (m *mockCommentRepo) ModerateComment(ctx context.Context, id string, status string) error {
	if m.ModerateCommentFn != nil {
		return m.ModerateCommentFn(ctx, id, status)
	}
	return nil
}

// ─── test app factory ─────────────────────────────────────────────────────────

// newCommentTestApp builds a minimal Fiber app wired to CommentHandler.
// It injects "test-user-id" into locals so auth middleware is bypassed.
func newCommentTestApp(
	commentRepo repositories.CommentRepositoryInterface,
	marketplaceRepo repositories.MarketplaceRepositoryInterface,
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

	// Inject fake userId so every handler that calls ValidateUserID succeeds.
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userId", "test-user-id")
		return c.Next()
	})

	h := handlers.NewCommentHandler(commentRepo, marketplaceRepo)

	app.Post("/comments", h.CreateComment)
	app.Get("/comments", h.GetComments)
	app.Get("/comments/:commentid", h.GetCommentByID)
	app.Get("/items/:itemid/comments", h.GetCommentsByItemID)
	app.Patch("/comments/:commentid", h.UpdateComment)
	app.Delete("/comments/:commentid", h.DeleteComment)
	app.Post("/comments/:commentid/reactions", h.CreateReaction)
	app.Delete("/comments/:commentid/reactions", h.DeleteReaction)
	app.Get("/comments/:commentid/reactions", h.GetReactionsByCommentID)
	app.Get("/comments/:commentid/reactions/summary", h.GetReactionSummary)
	app.Get("/items/:itemid/comments/count", h.GetCommentCount)
	app.Patch("/comments/:commentid/moderate", h.ModerateComment)

	return app
}

// newCommentTestAppNoAuth builds the same app WITHOUT injecting userId.
func newCommentTestAppNoAuth(
	commentRepo repositories.CommentRepositoryInterface,
	marketplaceRepo repositories.MarketplaceRepositoryInterface,
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

	h := handlers.NewCommentHandler(commentRepo, marketplaceRepo)

	app.Post("/comments", h.CreateComment)
	app.Patch("/comments/:commentid", h.UpdateComment)
	app.Delete("/comments/:commentid", h.DeleteComment)
	app.Post("/comments/:commentid/reactions", h.CreateReaction)
	app.Delete("/comments/:commentid/reactions", h.DeleteReaction)
	app.Patch("/comments/:commentid/moderate", h.ModerateComment)

	return app
}

// ─── request helpers ──────────────────────────────────────────────────────────

func commentStatusOf(t *testing.T, app *fiber.App, method, path string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	return resp.StatusCode
}

func commentBodyOf(t *testing.T, app *fiber.App, method, path string, body io.Reader, contentType string) (int, []byte) {
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

func makeComment(id, itemID, authorID, content string) *models.Comment {
	now := time.Now()
	return &models.Comment{
		Id:        id,
		Content:   content,
		AuthorId:  authorID,
		ItemId:    itemID,
		Status:    "published",
		Edited:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func makeMarketplaceItem(id, title string) *models.MarketplaceItem {
	now := time.Now()
	return &models.MarketplaceItem{
		Id:           id,
		Title:        title,
		Description:  "A test item",
		TemplateType: "block",
		AuthorId:     "owner-id",
		AuthorName:   "Owner",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// ─── CreateComment ────────────────────────────────────────────────────────────

func TestCreateComment_Returns201OnSuccess(t *testing.T) {
	// Happy path: valid body, item exists, no parent — the handler must return 201
	// with the created comment.
	commentRepo := &mockCommentRepo{
		CreateCommentFn: func(_ context.Context, c models.Comment) (*models.Comment, error) {
			return &c, nil
		},
	}
	marketplaceRepo := &mockMarketplaceRepo{
		GetMarketplaceItemByIDFn: func(_ string) (*models.MarketplaceItem, error) {
			return makeMarketplaceItem("item-1", "My Template"), nil
		},
	}
	app := newCommentTestApp(commentRepo, marketplaceRepo)

	body := strings.NewReader(`{"content":"Great template!","itemId":"item-1"}`)
	status, respBody := commentBodyOf(t, app, "POST", "/comments", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d – body: %s", status, respBody)
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["content"] != "Great template!" {
		t.Errorf("content: got %v, want %q", result["content"], "Great template!")
	}
}

func TestCreateComment_Returns400WhenContentMissing(t *testing.T) {
	// The "content" field is required; omitting it must fail validation.
	app := newCommentTestApp(&mockCommentRepo{}, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"itemId":"item-1"}`)
	status, _ := commentBodyOf(t, app, "POST", "/comments", body, "application/json")
	if status == fiber.StatusCreated || status == fiber.StatusOK {
		t.Errorf("expected 4xx for missing content, got %d", status)
	}
}

func TestCreateComment_Returns400WhenItemIDMissing(t *testing.T) {
	// The "itemId" field is required; omitting it must fail validation.
	app := newCommentTestApp(&mockCommentRepo{}, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"content":"Hello"}`)
	status, _ := commentBodyOf(t, app, "POST", "/comments", body, "application/json")
	if status == fiber.StatusCreated || status == fiber.StatusOK {
		t.Errorf("expected 4xx for missing itemId, got %d", status)
	}
}

func TestCreateComment_Returns404WhenItemNotFound(t *testing.T) {
	// If GetMarketplaceItemByID returns nil the handler treats it as not found → 404.
	marketplaceRepo := &mockMarketplaceRepo{
		GetMarketplaceItemByIDFn: func(_ string) (*models.MarketplaceItem, error) {
			return nil, nil // nil item → 404
		},
	}
	app := newCommentTestApp(&mockCommentRepo{}, marketplaceRepo)

	body := strings.NewReader(`{"content":"Hello","itemId":"ghost"}`)
	status, _ := commentBodyOf(t, app, "POST", "/comments", body, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404 when item not found, got %d", status)
	}
}

func TestCreateComment_Returns500WhenItemLookupFails(t *testing.T) {
	// An unexpected error from GetMarketplaceItemByID must yield 500.
	marketplaceRepo := &mockMarketplaceRepo{
		GetMarketplaceItemByIDFn: func(_ string) (*models.MarketplaceItem, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newCommentTestApp(&mockCommentRepo{}, marketplaceRepo)

	body := strings.NewReader(`{"content":"Hello","itemId":"item-1"}`)
	status, _ := commentBodyOf(t, app, "POST", "/comments", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestCreateComment_Returns404WhenParentCommentNotFound(t *testing.T) {
	// If a parentId is supplied but the parent comment doesn't exist → 404.
	commentRepo := &mockCommentRepo{
		GetCommentByIDFn: func(_ context.Context, _ string) (*models.Comment, error) {
			return nil, repositories.ErrCommentNotFound
		},
	}
	marketplaceRepo := &mockMarketplaceRepo{
		GetMarketplaceItemByIDFn: func(_ string) (*models.MarketplaceItem, error) {
			return makeMarketplaceItem("item-1", "My Template"), nil
		},
	}
	app := newCommentTestApp(commentRepo, marketplaceRepo)

	body := strings.NewReader(`{"content":"Reply","itemId":"item-1","parentId":"ghost-parent"}`)
	status, _ := commentBodyOf(t, app, "POST", "/comments", body, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404 when parent comment not found, got %d", status)
	}
}

func TestCreateComment_Returns400WhenParentBelongsToDifferentItem(t *testing.T) {
	// A parent comment that belongs to a different item must be rejected with 400.
	commentRepo := &mockCommentRepo{
		GetCommentByIDFn: func(_ context.Context, id string) (*models.Comment, error) {
			// The parent comment belongs to "other-item", not "item-1".
			return makeComment(id, "other-item", "author-x", "Original"), nil
		},
	}
	marketplaceRepo := &mockMarketplaceRepo{
		GetMarketplaceItemByIDFn: func(_ string) (*models.MarketplaceItem, error) {
			return makeMarketplaceItem("item-1", "My Template"), nil
		},
	}
	app := newCommentTestApp(commentRepo, marketplaceRepo)

	body := strings.NewReader(`{"content":"Reply","itemId":"item-1","parentId":"parent-1"}`)
	status, _ := commentBodyOf(t, app, "POST", "/comments", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 when parent belongs to different item, got %d", status)
	}
}

func TestCreateComment_Returns500WhenCreateFails(t *testing.T) {
	// An unexpected error from CreateComment must yield 500.
	commentRepo := &mockCommentRepo{
		CreateCommentFn: func(_ context.Context, _ models.Comment) (*models.Comment, error) {
			return nil, errors.New("db write error")
		},
	}
	marketplaceRepo := &mockMarketplaceRepo{
		GetMarketplaceItemByIDFn: func(_ string) (*models.MarketplaceItem, error) {
			return makeMarketplaceItem("item-1", "My Template"), nil
		},
	}
	app := newCommentTestApp(commentRepo, marketplaceRepo)

	body := strings.NewReader(`{"content":"Hello","itemId":"item-1"}`)
	status, _ := commentBodyOf(t, app, "POST", "/comments", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestCreateComment_Returns401WhenNoUserID(t *testing.T) {
	app := newCommentTestAppNoAuth(&mockCommentRepo{}, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"content":"Hello","itemId":"item-1"}`)
	status, _ := commentBodyOf(t, app, "POST", "/comments", body, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

func TestCreateComment_SetsAuthorIDFromLocals(t *testing.T) {
	// The handler must populate AuthorId from the locals userId, not from the body.
	var capturedAuthorID string
	commentRepo := &mockCommentRepo{
		CreateCommentFn: func(_ context.Context, c models.Comment) (*models.Comment, error) {
			capturedAuthorID = c.AuthorId
			return &c, nil
		},
	}
	marketplaceRepo := &mockMarketplaceRepo{
		GetMarketplaceItemByIDFn: func(_ string) (*models.MarketplaceItem, error) {
			return makeMarketplaceItem("item-1", "Template"), nil
		},
	}
	app := newCommentTestApp(commentRepo, marketplaceRepo)

	body := strings.NewReader(`{"content":"Hello","itemId":"item-1"}`)
	status, _ := commentBodyOf(t, app, "POST", "/comments", body, "application/json")
	if status != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", status)
	}
	if capturedAuthorID != "test-user-id" {
		t.Errorf("authorId: got %q, want %q", capturedAuthorID, "test-user-id")
	}
}

// ─── GetComments ──────────────────────────────────────────────────────────────

func TestGetComments_Returns200WithResults(t *testing.T) {
	// The handler reads filters from query params and returns a paginated envelope.
	commentRepo := &mockCommentRepo{
		GetCommentsFn: func(_ context.Context, f models.CommentFilter) ([]models.Comment, int64, error) {
			return []models.Comment{
				*makeComment("c-1", "item-1", "author-x", "Hello"),
			}, 1, nil
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	status, body := commentBodyOf(t, app, "GET", "/comments?itemId=item-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["total"] != float64(1) {
		t.Errorf("total: got %v, want 1", result["total"])
	}
	data, ok := result["data"].([]any)
	if !ok || len(data) != 1 {
		t.Errorf("expected 1 comment in data, got %v", result["data"])
	}
}

func TestGetComments_Returns200WithEmptySlice(t *testing.T) {
	// An empty result must return 200 with data:[] and total:0.
	commentRepo := &mockCommentRepo{
		GetCommentsFn: func(_ context.Context, _ models.CommentFilter) ([]models.Comment, int64, error) {
			return []models.Comment{}, 0, nil
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	status, body := commentBodyOf(t, app, "GET", "/comments", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["total"] != float64(0) {
		t.Errorf("total: got %v, want 0", result["total"])
	}
}

func TestGetComments_Returns500OnRepositoryError(t *testing.T) {
	// A repository failure must surface as 500.
	commentRepo := &mockCommentRepo{
		GetCommentsFn: func(_ context.Context, _ models.CommentFilter) ([]models.Comment, int64, error) {
			return nil, 0, errors.New("db error")
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "GET", "/comments"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── GetCommentByID ───────────────────────────────────────────────────────────

func TestGetCommentByID_ReturnsCommentWhenFound(t *testing.T) {
	// Happy path: repository finds the comment; handler returns 200.
	commentRepo := &mockCommentRepo{
		GetCommentByIDFn: func(_ context.Context, id string) (*models.Comment, error) {
			if id == "c-1" {
				return makeComment("c-1", "item-1", "author-x", "Hello"), nil
			}
			return nil, repositories.ErrCommentNotFound
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	status, body := commentBodyOf(t, app, "GET", "/comments/c-1", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["id"] != "c-1" {
		t.Errorf("id: got %v, want %q", result["id"], "c-1")
	}
}

func TestGetCommentByID_Returns404WhenNotFound(t *testing.T) {
	// ErrCommentNotFound must translate to a 404 response.
	commentRepo := &mockCommentRepo{
		GetCommentByIDFn: func(_ context.Context, _ string) (*models.Comment, error) {
			return nil, repositories.ErrCommentNotFound
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "GET", "/comments/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetCommentByID_Returns500OnRepositoryError(t *testing.T) {
	// A non-sentinel error must yield 500.
	commentRepo := &mockCommentRepo{
		GetCommentByIDFn: func(_ context.Context, _ string) (*models.Comment, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "GET", "/comments/c-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── GetCommentsByItemID ──────────────────────────────────────────────────────

func TestGetCommentsByItemID_ReturnsCommentsForItem(t *testing.T) {
	// All comments for the requested item must be returned in the paginated
	// envelope.
	commentRepo := &mockCommentRepo{
		GetCommentsFn: func(_ context.Context, f models.CommentFilter) ([]models.Comment, int64, error) {
			if f.ItemId != "item-1" {
				return nil, 0, errors.New("unexpected itemId: " + f.ItemId)
			}
			return []models.Comment{
				*makeComment("c-1", "item-1", "author-x", "Great"),
				*makeComment("c-2", "item-1", "author-y", "Nice"),
			}, 2, nil
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	status, body := commentBodyOf(t, app, "GET", "/items/item-1/comments", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["total"] != float64(2) {
		t.Errorf("total: got %v, want 2", result["total"])
	}
}

func TestGetCommentsByItemID_Returns500OnRepositoryError(t *testing.T) {
	// A repository failure must surface as 500.
	commentRepo := &mockCommentRepo{
		GetCommentsFn: func(_ context.Context, _ models.CommentFilter) ([]models.Comment, int64, error) {
			return nil, 0, errors.New("db error")
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "GET", "/items/item-1/comments"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── UpdateComment ────────────────────────────────────────────────────────────

func TestUpdateComment_Returns200OnSuccess(t *testing.T) {
	// A valid PATCH body must update the comment and return 200 with the updated
	// resource.
	newContent := "Updated content"
	commentRepo := &mockCommentRepo{
		UpdateCommentFn: func(_ context.Context, id, _ string, _ map[string]any) (*models.Comment, error) {
			c := makeComment(id, "item-1", "test-user-id", newContent)
			return c, nil
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"content":"Updated content"}`)
	status, respBody := commentBodyOf(t, app, "PATCH", "/comments/c-1", body, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, respBody)
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["content"] != newContent {
		t.Errorf("content: got %v, want %q", result["content"], newContent)
	}
}

func TestUpdateComment_Returns400WhenBodyIsEmpty(t *testing.T) {
	// An empty JSON object has no valid update fields → 400 from RequireUpdates.
	app := newCommentTestApp(&mockCommentRepo{}, &mockMarketplaceRepo{})

	body := strings.NewReader(`{}`)
	status, _ := commentBodyOf(t, app, "PATCH", "/comments/c-1", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for empty update body, got %d", status)
	}
}

func TestUpdateComment_Returns400WhenContentIsEmpty(t *testing.T) {
	// An explicitly empty string for content must be rejected with 400.
	app := newCommentTestApp(&mockCommentRepo{}, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"content":""}`)
	status, _ := commentBodyOf(t, app, "PATCH", "/comments/c-1", body, "application/json")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for empty content string, got %d", status)
	}
}

func TestUpdateComment_Returns404WhenCommentNotFound(t *testing.T) {
	// ErrCommentNotFound from the repo must translate to 404.
	commentRepo := &mockCommentRepo{
		UpdateCommentFn: func(_ context.Context, _, _ string, _ map[string]any) (*models.Comment, error) {
			return nil, repositories.ErrCommentNotFound
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"content":"Updated"}`)
	status, _ := commentBodyOf(t, app, "PATCH", "/comments/ghost", body, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestUpdateComment_Returns403WhenUnauthorized(t *testing.T) {
	// ErrCommentUnauthorized from the repo must translate to 403 Forbidden.
	commentRepo := &mockCommentRepo{
		UpdateCommentFn: func(_ context.Context, _, _ string, _ map[string]any) (*models.Comment, error) {
			return nil, repositories.ErrCommentUnauthorized
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"content":"Updated"}`)
	status, _ := commentBodyOf(t, app, "PATCH", "/comments/c-1", body, "application/json")
	if status != fiber.StatusForbidden {
		t.Errorf("expected 403 for unauthorized update, got %d", status)
	}
}

func TestUpdateComment_Returns401WhenNoUserID(t *testing.T) {
	app := newCommentTestAppNoAuth(&mockCommentRepo{}, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"content":"Updated"}`)
	status, _ := commentBodyOf(t, app, "PATCH", "/comments/c-1", body, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

// ─── DeleteComment ────────────────────────────────────────────────────────────

func TestDeleteComment_Returns204OnSuccess(t *testing.T) {
	// A successful delete must return 204 No Content.
	commentRepo := &mockCommentRepo{
		DeleteCommentFn: func(_ context.Context, _, _ string) error {
			return nil
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "DELETE", "/comments/c-1"); code != fiber.StatusNoContent {
		t.Errorf("expected 204, got %d", code)
	}
}

func TestDeleteComment_Returns403WhenUnauthorized(t *testing.T) {
	// ErrCommentUnauthorized from the repo must translate to 403 Forbidden.
	commentRepo := &mockCommentRepo{
		DeleteCommentFn: func(_ context.Context, _, _ string) error {
			return repositories.ErrCommentUnauthorized
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "DELETE", "/comments/c-1"); code != fiber.StatusForbidden {
		t.Errorf("expected 403, got %d", code)
	}
}

func TestDeleteComment_Returns404WhenNotFound(t *testing.T) {
	// ErrCommentNotFound from the repo must map to 404.
	commentRepo := &mockCommentRepo{
		DeleteCommentFn: func(_ context.Context, _, _ string) error {
			return repositories.ErrCommentNotFound
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "DELETE", "/comments/ghost"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeleteComment_Returns500OnRepositoryError(t *testing.T) {
	commentRepo := &mockCommentRepo{
		DeleteCommentFn: func(_ context.Context, _, _ string) error {
			return errors.New("db error")
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "DELETE", "/comments/c-1"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

func TestDeleteComment_Returns401WhenNoUserID(t *testing.T) {
	app := newCommentTestAppNoAuth(&mockCommentRepo{}, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "DELETE", "/comments/c-1"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── CreateReaction ───────────────────────────────────────────────────────────

func TestCreateReaction_Returns201OnSuccess(t *testing.T) {
	// A valid reaction type on an existing comment must return 201.
	commentRepo := &mockCommentRepo{
		GetCommentByIDFn: func(_ context.Context, id string) (*models.Comment, error) {
			return makeComment(id, "item-1", "author-x", "Hello"), nil
		},
		CreateReactionFn: func(_ context.Context, r models.CommentReaction) (*models.CommentReaction, error) {
			return &r, nil
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"type":"like"}`)
	status, _ := commentBodyOf(t, app, "POST", "/comments/c-1/reactions", body, "application/json")
	if status != fiber.StatusCreated {
		t.Errorf("expected 201, got %d", status)
	}
}

func TestCreateReaction_Returns404WhenCommentNotFound(t *testing.T) {
	// If the comment does not exist the handler must return 404 before creating
	// the reaction.
	commentRepo := &mockCommentRepo{
		GetCommentByIDFn: func(_ context.Context, _ string) (*models.Comment, error) {
			return nil, repositories.ErrCommentNotFound
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"type":"like"}`)
	status, _ := commentBodyOf(t, app, "POST", "/comments/ghost/reactions", body, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestCreateReaction_Returns400WhenTypeMissing(t *testing.T) {
	// The "type" field is required; omitting it must fail validation.
	commentRepo := &mockCommentRepo{
		GetCommentByIDFn: func(_ context.Context, id string) (*models.Comment, error) {
			return makeComment(id, "item-1", "author-x", "Hello"), nil
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	body := strings.NewReader(`{}`)
	status, _ := commentBodyOf(t, app, "POST", "/comments/c-1/reactions", body, "application/json")
	if status == fiber.StatusCreated {
		t.Errorf("expected 4xx for missing type, got 201")
	}
}

func TestCreateReaction_Returns401WhenNoUserID(t *testing.T) {
	app := newCommentTestAppNoAuth(&mockCommentRepo{}, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"type":"like"}`)
	status, _ := commentBodyOf(t, app, "POST", "/comments/c-1/reactions", body, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

// ─── DeleteReaction ───────────────────────────────────────────────────────────

func TestDeleteReaction_Returns204OnSuccess(t *testing.T) {
	// A successful delete must return 204.
	commentRepo := &mockCommentRepo{
		DeleteReactionFn: func(_ context.Context, _, _, _ string) error {
			return nil
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "DELETE", "/comments/c-1/reactions?type=like"); code != fiber.StatusNoContent {
		t.Errorf("expected 204, got %d", code)
	}
}

func TestDeleteReaction_Returns400WhenTypeMissing(t *testing.T) {
	// The "type" query parameter is required; omitting it must yield 400.
	app := newCommentTestApp(&mockCommentRepo{}, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "DELETE", "/comments/c-1/reactions"); code != fiber.StatusBadRequest {
		t.Errorf("expected 400 for missing type param, got %d", code)
	}
}

func TestDeleteReaction_Returns404WhenReactionNotFound(t *testing.T) {
	// ErrCommentNotFound when deleting a reaction maps to 404 "Reaction not found".
	commentRepo := &mockCommentRepo{
		DeleteReactionFn: func(_ context.Context, _, _, _ string) error {
			return repositories.ErrCommentNotFound
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "DELETE", "/comments/c-1/reactions?type=like"); code != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestDeleteReaction_Returns401WhenNoUserID(t *testing.T) {
	app := newCommentTestAppNoAuth(&mockCommentRepo{}, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "DELETE", "/comments/c-1/reactions?type=like"); code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", code)
	}
}

// ─── GetReactionsByCommentID ──────────────────────────────────────────────────

func TestGetReactionsByCommentID_ReturnsReactions(t *testing.T) {
	commentRepo := &mockCommentRepo{
		GetReactionsByCommentIDFn: func(_ context.Context, _ string) ([]models.CommentReaction, error) {
			return []models.CommentReaction{{Id: "r-1", CommentId: "c-1", UserId: "u-1", Type: "like"}}, nil
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	status, body := commentBodyOf(t, app, "GET", "/comments/c-1/reactions", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 reaction, got %d", len(result))
	}
}

func TestGetReactionsByCommentID_Returns500OnRepositoryError(t *testing.T) {
	commentRepo := &mockCommentRepo{
		GetReactionsByCommentIDFn: func(_ context.Context, _ string) ([]models.CommentReaction, error) {
			return nil, errors.New("db error")
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "GET", "/comments/c-1/reactions"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── GetReactionSummary ───────────────────────────────────────────────────────

func TestGetReactionSummary_ReturnsSummary(t *testing.T) {
	commentRepo := &mockCommentRepo{
		GetReactionSummaryFn: func(_ context.Context, _ string) ([]models.ReactionSummary, error) {
			return []models.ReactionSummary{{Type: "like", Count: 5}}, nil
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	status, body := commentBodyOf(t, app, "GET", "/comments/c-1/reactions/summary", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 summary entry, got %d", len(result))
	}
	if result[0]["count"] != float64(5) {
		t.Errorf("count: got %v, want 5", result[0]["count"])
	}
}

// ─── GetCommentCount ──────────────────────────────────────────────────────────

func TestGetCommentCount_ReturnsCount(t *testing.T) {
	// The handler returns an object with itemId and count fields.
	commentRepo := &mockCommentRepo{
		GetCommentCountByItemIDFn: func(_ context.Context, itemID string) (int64, error) {
			return 42, nil
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	status, body := commentBodyOf(t, app, "GET", "/items/item-1/comments/count", nil, "")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["count"] != float64(42) {
		t.Errorf("count: got %v, want 42", result["count"])
	}
	if result["itemId"] != "item-1" {
		t.Errorf("itemId: got %v, want %q", result["itemId"], "item-1")
	}
}

func TestGetCommentCount_Returns500OnRepositoryError(t *testing.T) {
	commentRepo := &mockCommentRepo{
		GetCommentCountByItemIDFn: func(_ context.Context, _ string) (int64, error) {
			return 0, errors.New("db error")
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	if code := commentStatusOf(t, app, "GET", "/items/item-1/comments/count"); code != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", code)
	}
}

// ─── ModerateComment ──────────────────────────────────────────────────────────

func TestModerateComment_Returns200OnSuccess(t *testing.T) {
	// A valid status value must update the comment and return 200 with a message.
	commentRepo := &mockCommentRepo{
		ModerateCommentFn: func(_ context.Context, _, _ string) error {
			return nil
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"status":"flagged"}`)
	status, respBody := commentBodyOf(t, app, "PATCH", "/comments/c-1/moderate", body, "application/json")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, respBody)
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, respBody)
	}
	if result["message"] != "Comment moderated successfully" {
		t.Errorf("message: got %v, want %q", result["message"], "Comment moderated successfully")
	}
	if result["status"] != "flagged" {
		t.Errorf("status: got %v, want %q", result["status"], "flagged")
	}
}

func TestModerateComment_Returns400WhenStatusMissing(t *testing.T) {
	// The "status" field is required; an empty body must fail validation.
	app := newCommentTestApp(&mockCommentRepo{}, &mockMarketplaceRepo{})

	body := strings.NewReader(`{}`)
	status, _ := commentBodyOf(t, app, "PATCH", "/comments/c-1/moderate", body, "application/json")
	if status == fiber.StatusOK {
		t.Errorf("expected 4xx for missing status, got 200")
	}
}

func TestModerateComment_Returns400WhenStatusIsInvalid(t *testing.T) {
	// Only published/pending/flagged/deleted are valid; an unknown value must
	// be rejected by the "oneof" validator tag.
	app := newCommentTestApp(&mockCommentRepo{}, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"status":"banned"}`)
	status, _ := commentBodyOf(t, app, "PATCH", "/comments/c-1/moderate", body, "application/json")
	if status == fiber.StatusOK {
		t.Errorf("expected 4xx for invalid status value, got 200")
	}
}

func TestModerateComment_Returns404WhenCommentNotFound(t *testing.T) {
	// ErrCommentNotFound from ModerateComment must map to 404.
	commentRepo := &mockCommentRepo{
		ModerateCommentFn: func(_ context.Context, _, _ string) error {
			return repositories.ErrCommentNotFound
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"status":"flagged"}`)
	status, _ := commentBodyOf(t, app, "PATCH", "/comments/ghost/moderate", body, "application/json")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestModerateComment_Returns500OnRepositoryError(t *testing.T) {
	commentRepo := &mockCommentRepo{
		ModerateCommentFn: func(_ context.Context, _, _ string) error {
			return errors.New("db write error")
		},
	}
	app := newCommentTestApp(commentRepo, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"status":"flagged"}`)
	status, _ := commentBodyOf(t, app, "PATCH", "/comments/c-1/moderate", body, "application/json")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestModerateComment_Returns401WhenNoUserID(t *testing.T) {
	app := newCommentTestAppNoAuth(&mockCommentRepo{}, &mockMarketplaceRepo{})

	body := strings.NewReader(`{"status":"flagged"}`)
	status, _ := commentBodyOf(t, app, "PATCH", "/comments/c-1/moderate", body, "application/json")
	if status != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}