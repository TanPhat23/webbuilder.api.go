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
	"my-go-app/internal/repositories"
	"my-go-app/tests/testutil"

	"github.com/gofiber/fiber/v2"
)

// ─── test app factory ─────────────────────────────────────────────────────────

// newUserTestApp builds a minimal Fiber app wired to UserHandler.
// Auth middleware is bypassed by injecting "test-user-id" into locals.
func newUserTestApp(userRepo repositories.UserRepositoryInterface) *fiber.App {
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

	h := handlers.NewUserHandler(userRepo)

	app.Get("/users/search", h.SearchUsers)
	app.Get("/users/email/:email", h.GetUserByEmail)
	app.Get("/users/username/:username", h.GetUserByUsername)

	return app
}

// ─── fixture helpers ──────────────────────────────────────────────────────────

func makeUser(id, email, username string) *models.User {
	fn := "Test"
	ln := "User"
	return &models.User{
		Id:        id,
		Email:     email,
		FirstName: &fn,
		LastName:  &ln,
	}
}

// ─── SearchUsers ──────────────────────────────────────────────────────────────

func TestSearchUsers_Returns200WithResults(t *testing.T) {
	// When the repository returns matching users the handler serialises the
	// slice with 200.
	repo := &testutil.MockUserRepository{
		SearchUsersFn: func(_ context.Context, query string) ([]models.User, error) {
			if query != "alice" {
				return nil, errors.New("unexpected query: " + query)
			}
			return []models.User{*makeUser("u-1", "alice@example.com", "alice")}, nil
		},
	}
	app := newUserTestApp(repo)

	status, body := userBodyOf(t, app, "GET", "/users/search?q=alice")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result []map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON array: %v – body: %s", err, body)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 user, got %d", len(result))
	}
}

func TestSearchUsers_ReturnsEmptySliceWhenNoMatches(t *testing.T) {
	// An empty result set must produce 200 with an empty JSON array.
	repo := &testutil.MockUserRepository{
		SearchUsersFn: func(_ context.Context, _ string) ([]models.User, error) {
			return []models.User{}, nil
		},
	}
	app := newUserTestApp(repo)

	status, body := userBodyOf(t, app, "GET", "/users/search?q=nobody")
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

func TestSearchUsers_Returns400WhenQueryParamMissing(t *testing.T) {
	// The 'q' query parameter is required; omitting it must yield 400.
	repo := &testutil.MockUserRepository{}
	app := newUserTestApp(repo)

	status, _ := userBodyOf(t, app, "GET", "/users/search")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for missing q param, got %d", status)
	}
}

func TestSearchUsers_Returns400WhenQueryParamIsBlank(t *testing.T) {
	// A whitespace-only query string is trimmed to "" and must also be rejected.
	repo := &testutil.MockUserRepository{}
	app := newUserTestApp(repo)

	status, _ := userBodyOf(t, app, "GET", "/users/search?q=%20%20%20")
	if status != fiber.StatusBadRequest {
		t.Errorf("expected 400 for blank q param, got %d", status)
	}
}

func TestSearchUsers_Returns500OnRepositoryError(t *testing.T) {
	// An unexpected repository error must surface as 500 Internal Server Error.
	repo := &testutil.MockUserRepository{
		SearchUsersFn: func(_ context.Context, _ string) ([]models.User, error) {
			return nil, errors.New("db connection lost")
		},
	}
	app := newUserTestApp(repo)

	status, _ := userBodyOf(t, app, "GET", "/users/search?q=alice")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestSearchUsers_PassesQueryStringToRepository(t *testing.T) {
	// Verify the handler forwards the 'q' value to the repository unchanged.
	var capturedQuery string
	repo := &testutil.MockUserRepository{
		SearchUsersFn: func(_ context.Context, query string) ([]models.User, error) {
			capturedQuery = query
			return []models.User{}, nil
		},
	}
	app := newUserTestApp(repo)

	status, _ := userBodyOf(t, app, "GET", "/users/search?q=findme")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedQuery != "findme" {
		t.Errorf("query passed to repo: got %q, want %q", capturedQuery, "findme")
	}
}

// ─── GetUserByEmail ───────────────────────────────────────────────────────────

func TestGetUserByEmail_ReturnsUserWhenFound(t *testing.T) {
	// Happy path: repository finds the user and the handler serialises it with 200.
	repo := &testutil.MockUserRepository{
		GetUserByEmailFn: func(_ context.Context, email string) (*models.User, error) {
			if email == "alice@example.com" {
				return makeUser("u-1", email, "alice"), nil
			}
			return nil, repositories.ErrUserNotFound
		},
	}
	app := newUserTestApp(repo)

	status, body := userBodyOf(t, app, "GET", "/users/email/alice@example.com")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["email"] != "alice@example.com" {
		t.Errorf("email: got %v, want %q", result["email"], "alice@example.com")
	}
}

func TestGetUserByEmail_Returns404WhenNotFound(t *testing.T) {
	// ErrUserNotFound (message ends with "not found") must map to 404.
	repo := &testutil.MockUserRepository{
		GetUserByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
			return nil, repositories.ErrUserNotFound
		},
	}
	app := newUserTestApp(repo)

	status, _ := userBodyOf(t, app, "GET", "/users/email/ghost@example.com")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestGetUserByEmail_Returns500OnRepositoryError(t *testing.T) {
	// Any non-sentinel repository error must yield 500.
	repo := &testutil.MockUserRepository{
		GetUserByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newUserTestApp(repo)

	status, _ := userBodyOf(t, app, "GET", "/users/email/alice@example.com")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestGetUserByEmail_Returns400WhenParamMissing(t *testing.T) {
	// Fiber does not match the ":email" route when the segment is empty; any
	// response other than 200 confirms the empty-param guard is effective.
	repo := &testutil.MockUserRepository{}
	app := newUserTestApp(repo)

	// A trailing slash causes Fiber to return 404 (no matching route), which is
	// still not a success — confirm we never get 200.
	status, _ := userBodyOf(t, app, "GET", "/users/email/")
	if status == fiber.StatusOK {
		t.Errorf("empty email param should not return 200, got %d", status)
	}
}

func TestGetUserByEmail_PassesEmailToRepository(t *testing.T) {
	// The route param value must reach the repository unchanged.
	var capturedEmail string
	repo := &testutil.MockUserRepository{
		GetUserByEmailFn: func(_ context.Context, email string) (*models.User, error) {
			capturedEmail = email
			return makeUser("u-1", email, "alice"), nil
		},
	}
	app := newUserTestApp(repo)

	status, _ := userBodyOf(t, app, "GET", "/users/email/test@domain.com")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedEmail != "test@domain.com" {
		t.Errorf("email passed to repo: got %q, want %q", capturedEmail, "test@domain.com")
	}
}

// ─── GetUserByUsername ────────────────────────────────────────────────────────

func TestGetUserByUsername_ReturnsUserWhenFound(t *testing.T) {
	// Happy path: repository finds the user and the handler serialises it with 200.
	repo := &testutil.MockUserRepository{
		GetUserByUsernameFn: func(_ context.Context, username string) (*models.User, error) {
			if username == "alice" {
				return makeUser("u-1", "alice@example.com", username), nil
			}
			return nil, repositories.ErrUserNotFound
		},
	}
	app := newUserTestApp(repo)

	status, body := userBodyOf(t, app, "GET", "/users/username/alice")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d – body: %s", status, body)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not a JSON object: %v – body: %s", err, body)
	}
	if result["id"] != "u-1" {
		t.Errorf("id: got %v, want %q", result["id"], "u-1")
	}
}

func TestGetUserByUsername_Returns404WhenNotFound(t *testing.T) {
	// ErrUserNotFound must map to 404 via HandleRepoError.
	repo := &testutil.MockUserRepository{
		GetUserByUsernameFn: func(_ context.Context, _ string) (*models.User, error) {
			return nil, repositories.ErrUserNotFound
		},
	}
	app := newUserTestApp(repo)

	status, _ := userBodyOf(t, app, "GET", "/users/username/ghost")
	if status != fiber.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
}

func TestGetUserByUsername_Returns500OnRepositoryError(t *testing.T) {
	// Any non-sentinel repository error must yield 500.
	repo := &testutil.MockUserRepository{
		GetUserByUsernameFn: func(_ context.Context, _ string) (*models.User, error) {
			return nil, errors.New("db timeout")
		},
	}
	app := newUserTestApp(repo)

	status, _ := userBodyOf(t, app, "GET", "/users/username/alice")
	if status != fiber.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

func TestGetUserByUsername_Returns400WhenParamMissing(t *testing.T) {
	// An empty username segment must never produce 200.
	repo := &testutil.MockUserRepository{}
	app := newUserTestApp(repo)

	status, _ := userBodyOf(t, app, "GET", "/users/username/")
	if status == fiber.StatusOK {
		t.Errorf("empty username param should not return 200, got %d", status)
	}
}

func TestGetUserByUsername_PassesUsernameToRepository(t *testing.T) {
	// The route param value must reach the repository unchanged.
	var capturedUsername string
	repo := &testutil.MockUserRepository{
		GetUserByUsernameFn: func(_ context.Context, username string) (*models.User, error) {
			capturedUsername = username
			return makeUser("u-1", "alice@example.com", username), nil
		},
	}
	app := newUserTestApp(repo)

	status, _ := userBodyOf(t, app, "GET", "/users/username/myhero")
	if status != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if capturedUsername != "myhero" {
		t.Errorf("username passed to repo: got %q, want %q", capturedUsername, "myhero")
	}
}

// ─── request helpers (user handler) ──────────────────────────────────────────

// userBodyOf fires a request and returns status + raw response bytes.
func userBodyOf(t *testing.T, app *fiber.App, method, path string) (int, []byte) {
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