package unit

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/realtime/config"
	"my-go-app/internal/realtime/delivery/http/middleware"
	"my-go-app/internal/realtime/domain/entity"
)

// A client must never choose its own identity: no ?devUser= or X-Dev-User,
// in dev mode or not.
func TestAuthMiddlewareIgnoresClientChosenIdentity(t *testing.T) {
	t.Setenv("DEV_USER_ID", "configured-dev-user")
	whoami := func(cfg *config.Config) *fiber.App {
		app := fiber.New()
		app.Use(middleware.AuthMiddleware(cfg))
		app.Get("/", func(c fiber.Ctx) error {
			user, _ := c.Locals("user").(*entity.User)
			if user == nil {
				return c.SendString("")
			}
			return c.SendString(user.ID)
		})
		return app
	}
	dev := whoami(&config.Config{DevMode: true})
	for _, r := range []struct{ path, header string }{{path: "/?devUser=owner"}, {path: "/", header: "owner"}} {
		httpReq := httptest.NewRequest("GET", r.path, nil)
		if r.header != "" {
			httpReq.Header.Set("X-Dev-User", r.header)
		}
		res, err := dev.Test(httpReq)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(res.Body)
		if string(body) != "configured-dev-user" {
			t.Fatalf("%s: identity %q came from the client", r.path, body)
		}
	}
	prod := whoami(&config.Config{})
	res, err := prod.Test(httptest.NewRequest("GET", "/?devUser=owner", nil))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("without a token outside dev mode: status %d", res.StatusCode)
	}
}
