package middleware

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/auth"
	"my-go-app/internal/models"
)

type fakeTokenLookup struct {
	tokens  map[string]*models.AgentToken
	touched []string
}

func (f *fakeTokenLookup) FindActiveByHash(_ context.Context, hash string) (*models.AgentToken, error) {
	return f.tokens[hash], nil
}

func (f *fakeTokenLookup) Touch(_ context.Context, id string, _ time.Time) error {
	f.touched = append(f.touched, id)
	return nil
}

func TestAuthenticateMiddleware_AgentTokenValidAndRevoked(t *testing.T) {
	secret := "wbk_test_secret_token_1234567890"
	lookup := &fakeTokenLookup{
		tokens: map[string]*models.AgentToken{
			hashAgentSecret(secret): {Id: "tok_1", UserId: "user_42"},
		},
	}

	app := fiber.New()
	app.Use(NewAuthenticateMiddleware(lookup))
	app.Get("/api/v1/private", func(c fiber.Ctx) error {
		id, ok := auth.FromFiber(c)
		if !ok {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.JSON(fiber.Map{"userId": id.UserID, "method": id.AuthMethod})
	})

	req := httptest.NewRequest(fiber.MethodGet, "/api/v1/private", nil)
	req.Header.Set("Authorization", "Bearer "+secret)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for valid agent token, got %d", resp.StatusCode)
	}
	if len(lookup.touched) != 1 || lookup.touched[0] != "tok_1" {
		t.Fatalf("expected Touch(tok_1), got %v", lookup.touched)
	}

	// Revoke token -> immediate 401
	delete(lookup.tokens, hashAgentSecret(secret))
	req2 := httptest.NewRequest(fiber.MethodGet, "/api/v1/private", nil)
	req2.Header.Set("Authorization", "Bearer "+secret)
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp2.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for revoked agent token, got %d", resp2.StatusCode)
	}
}

func TestAuthenticateMiddleware_TestBypassAndProdGuard(t *testing.T) {
	t.Setenv("TEST_AUTH_BYPASS", "true")
	t.Setenv("ENV", "production")
	if err := ValidateAuthEnv(); err == nil {
		t.Fatal("expected ValidateAuthEnv to reject TEST_AUTH_BYPASS in production")
	}

	t.Setenv("ENV", "test")
	if err := ValidateAuthEnv(); err != nil {
		t.Fatalf("expected ValidateAuthEnv to allow TEST_AUTH_BYPASS in test, got %v", err)
	}

	app := fiber.New()
	app.Use(NewAuthenticateMiddleware(nil))
	app.Get("/api/v1/private", func(c fiber.Ctx) error {
		id, ok := auth.FromFiber(c)
		if !ok {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.JSON(fiber.Map{"userId": id.UserID, "method": id.AuthMethod})
	})

	req := httptest.NewRequest(fiber.MethodGet, "/api/v1/private", nil)
	req.Header.Set("X-Test-User-Id", "test_user_1")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 with X-Test-User-Id in test mode, got %d", resp.StatusCode)
	}
}
