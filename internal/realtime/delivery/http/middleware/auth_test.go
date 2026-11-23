package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/realtime/config"
)

func TestRealtimeAuthMiddlewareDevModeAndErrors(t *testing.T) {
	t.Setenv("DEV_USER_ID", "dev-42")
	devApp := fiber.New()
	devApp.Use(AuthMiddleware(&config.Config{DevMode: true}))
	devApp.Get("/ws", func(c fiber.Ctx) error {
		return c.SendString(c.Locals("userId").(string))
	})
	res, err := devApp.Test(httptest.NewRequest("GET", "/ws", nil))
	if err != nil || res.StatusCode != 200 {
		t.Fatalf("dev mode status=%d err=%v", res.StatusCode, err)
	}

	unconfigured := fiber.New()
	unconfigured.Use(AuthMiddleware(&config.Config{DevMode: false}))
	unconfigured.Get("/ws", func(c fiber.Ctx) error { return c.SendStatus(200) })

	noTokRes, err := unconfigured.Test(httptest.NewRequest("GET", "/ws", nil))
	if err != nil || noTokRes.StatusCode != 401 {
		t.Fatalf("missing token status=%d err=%v", noTokRes.StatusCode, err)
	}

	reqWithHeader := httptest.NewRequest("GET", "/ws", nil)
	reqWithHeader.Header.Set("Authorization", "Bearer abc")
	misconfRes, err := unconfigured.Test(reqWithHeader)
	if err != nil || misconfRes.StatusCode != 500 {
		t.Fatalf("missing clerk secret status=%d err=%v", misconfRes.StatusCode, err)
	}

	configured := fiber.New()
	configured.Use(AuthMiddleware(&config.Config{ClerkSecretKey: "sk_test_dummy"}))
	configured.Get("/ws", func(c fiber.Ctx) error { return c.SendStatus(200) })

	reqCookie := httptest.NewRequest("GET", "/ws", nil)
	reqCookie.AddCookie(&http.Cookie{Name: "__session", Value: "bad-jwt"})
	cookieRes, err := configured.Test(reqCookie)
	if err != nil || cookieRes.StatusCode != 401 {
		t.Fatalf("bad cookie token status=%d err=%v", cookieRes.StatusCode, err)
	}

	queryRes, err := configured.Test(httptest.NewRequest("GET", "/ws?token=bad-jwt", nil))
	if err != nil || queryRes.StatusCode != 401 {
		t.Fatalf("bad query token status=%d err=%v", queryRes.StatusCode, err)
	}
}
