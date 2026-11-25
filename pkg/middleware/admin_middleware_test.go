package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/auth"
)

func TestRequireAdmin(t *testing.T) {
	t.Setenv("ADMIN_USER_IDS", "admin_1, admin_2")

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		if uid := c.Get("X-Test-User"); uid != "" {
			auth.SetFiber(c, auth.Identity{UserID: uid, AuthMethod: "clerk"})
		}
		return c.Next()
	})
	app.Post("/custom-element-types", RequireAdmin(), func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusCreated)
	})

	// Unauthenticated -> 401
	resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/custom-element-types", nil))
	if err != nil || resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("unauthenticated: status=%v err=%v", resp, err)
	}

	// Non-admin -> 403
	reqNonAdmin := httptest.NewRequest(fiber.MethodPost, "/custom-element-types", nil)
	reqNonAdmin.Header.Set("X-Test-User", "user_99")
	resp, err = app.Test(reqNonAdmin)
	if err != nil || resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("non-admin: status=%v err=%v", resp, err)
	}

	// Admin -> 201
	reqAdmin := httptest.NewRequest(fiber.MethodPost, "/custom-element-types", nil)
	reqAdmin.Header.Set("X-Test-User", "admin_2")
	resp, err = app.Test(reqAdmin)
	if err != nil || resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("admin: status=%v err=%v", resp, err)
	}
}
