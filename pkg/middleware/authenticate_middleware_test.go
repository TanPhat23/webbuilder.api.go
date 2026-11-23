package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func newAuthTestApp() *fiber.App {
	app := fiber.New()
	app.Use(AuthenticateMiddleware)
	app.Get("/api/v1/pages/public/:id", func(c fiber.Ctx) error { return c.SendString("ok") })
	app.Post("/api/v1/pages/public/:id", func(c fiber.Ctx) error { return c.SendString("ok") })
	app.Get("/api/v1/projects/public", func(c fiber.Ctx) error { return c.SendString("ok") })
	app.Get("/api/v1/elements/public/:id", func(c fiber.Ctx) error { return c.SendString("ok") })
	app.Get("/api/v1/public/content", func(c fiber.Ctx) error { return c.SendString("ok") })
	app.Get("/api/v1/private", func(c fiber.Ctx) error { return c.SendString("ok") })
	return app
}

func TestAuthenticateMiddlewarePublicBypassesAndRejections(t *testing.T) {
	app := newAuthTestApp()

	for _, tc := range []struct {
		method string
		path   string
		auth   string
		want   int
	}{
		{"GET", "/api/v1/pages/public/p1", "", 200},
		{"HEAD", "/api/v1/pages/public/p1", "", 200},
		{"POST", "/api/v1/pages/public/p1", "", 401},
		{"GET", "/api/v1/projects/public", "", 200},
		{"GET", "/api/v1/elements/public/p1", "", 200},
		{"GET", "/api/v1/public/content", "", 200},
		{"GET", "/api/v1/private", "", 401},
		{"GET", "/api/v1/private", "Bearer ", 401},
		{"GET", "/api/v1/private", "Bearer not-a-valid-jwt", 401},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		if tc.auth != "" {
			req.Header.Set("Authorization", tc.auth)
		}
		res, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s %s err=%v", tc.method, tc.path, err)
		}
		if res.StatusCode != tc.want {
			t.Errorf("%s %s status=%d want=%d", tc.method, tc.path, res.StatusCode, tc.want)
		}
	}
}
