package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"my-go-app/internal/realtime/config"
	"my-go-app/internal/realtime/domain/entity"
)

func TestRealtimeAuthMiddlewareDevModeAndErrors(t *testing.T) {
	t.Setenv("ENV", "")
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

	// DevMode must not bypass auth in production.
	t.Setenv("ENV", "production")
	prodDevApp := fiber.New()
	prodDevApp.Use(AuthMiddleware(&config.Config{DevMode: true}))
	prodDevApp.Get("/ws", func(c fiber.Ctx) error { return c.SendStatus(200) })
	prodDevRes, err := prodDevApp.Test(httptest.NewRequest("GET", "/ws", nil))
	if err != nil || prodDevRes.StatusCode != 401 {
		t.Fatalf("dev mode in production must fail auth, got status=%d err=%v", prodDevRes.StatusCode, err)
	}
	t.Setenv("ENV", "")

	unconfigured := fiber.New()
	unconfigured.Use(AuthMiddleware(&config.Config{DevMode: false}))
	unconfigured.Get("/ws", func(c fiber.Ctx) error { return c.SendStatus(200) })

	noTokRes, err := unconfigured.Test(httptest.NewRequest("GET", "/ws", nil))
	if err != nil || noTokRes.StatusCode != 401 {
		t.Fatalf("missing token status=%d err=%v", noTokRes.StatusCode, err)
	}

	reqWithHeader := httptest.NewRequest("GET", "/ws", nil)
	reqWithHeader.Header.Set("Authorization", "Bearer abc")
	bearerRes, err := unconfigured.Test(reqWithHeader)
	if err != nil || bearerRes.StatusCode != 401 {
		t.Fatalf("bearer header must be rejected on /ws, status=%d err=%v", bearerRes.StatusCode, err)
	}

	misconfRes, err := unconfigured.Test(httptest.NewRequest("GET", "/ws?ticket=abc", nil))
	if err != nil || misconfRes.StatusCode != 500 {
		t.Fatalf("missing ticket store status=%d err=%v", misconfRes.StatusCode, err)
	}

	configured := fiber.New()
	configured.Use(AuthMiddleware(&config.Config{ClerkSecretKey: "sk_test_dummy"}))
	configured.Get("/ws", func(c fiber.Ctx) error { return c.SendStatus(200) })

	// Cookies and ?token= query parameters must be rejected in non-dev mode.
	reqCookie := httptest.NewRequest("GET", "/ws", nil)
	reqCookie.AddCookie(&http.Cookie{Name: "__session", Value: "bad-jwt"})
	cookieRes, err := configured.Test(reqCookie)
	if err != nil || cookieRes.StatusCode != 401 {
		t.Fatalf("cookie token status=%d err=%v", cookieRes.StatusCode, err)
	}

	queryRes, err := configured.Test(httptest.NewRequest("GET", "/ws?token=bad-jwt", nil))
	if err != nil || queryRes.StatusCode != 401 {
		t.Fatalf("query token status=%d err=%v", queryRes.StatusCode, err)
	}
}

func TestRealtimeAuthMiddlewareSingleUseTicketAndReplayRejection(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	rawTicket := "opaque-single-use-ticket-value"
	sum := sha256.Sum256([]byte(rawTicket))
	key := wsTicketRedisPrefix + hex.EncodeToString(sum[:])
	payload, _ := json.Marshal(wsTicketRecord{UserID: "user-99", ProjectID: "proj-42"})
	if err := rdb.Set(context.Background(), key, payload, 30*time.Second).Err(); err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Use(AuthMiddlewareWithRedis(&config.Config{DevMode: false}, rdb))
	app.Get("/ws/:project", func(c fiber.Ctx) error {
		u, _ := c.Locals("user").(*entity.User)
		tp, _ := c.Locals("ticketProjectId").(string)
		return c.SendString(u.ID + ":" + tp)
	})

	// First use succeeds and atomically consumes the ticket.
	firstRes, err := app.Test(httptest.NewRequest("GET", "/ws/proj-42?ticket="+rawTicket, nil))
	if err != nil || firstRes.StatusCode != 200 {
		t.Fatalf("first ticket status=%d err=%v", firstRes.StatusCode, err)
	}
	body, _ := io.ReadAll(firstRes.Body)
	if string(body) != "user-99:proj-42" {
		t.Fatalf("body = %q, want user-99:proj-42", body)
	}

	// Replay with the same ticket must be rejected.
	replayRes, err := app.Test(httptest.NewRequest("GET", "/ws/proj-42?ticket="+rawTicket, nil))
	if err != nil || replayRes.StatusCode != 401 {
		t.Fatalf("replay ticket status=%d err=%v", replayRes.StatusCode, err)
	}

	// Unknown ticket must be rejected.
	unknownRes, err := app.Test(httptest.NewRequest("GET", "/ws/proj-42?ticket=unknown", nil))
	if err != nil || unknownRes.StatusCode != 401 {
		t.Fatalf("unknown ticket status=%d err=%v", unknownRes.StatusCode, err)
	}
}
