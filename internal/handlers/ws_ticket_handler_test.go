package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"my-go-app/internal/auth"
)

func TestMintWSTicket_StoresSHA256AndReturns201(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		auth.SetFiber(c, auth.Identity{UserID: "user-123", AuthMethod: "clerk"})
		return c.Next()
	})
	h := NewWSTicketHandler(rdb)
	app.Post("/api/v1/projects/:projectid/ws-ticket", h.Mint)

	res, err := app.Test(httptest.NewRequest("POST", "/api/v1/projects/proj-1/ws-ticket", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d, want 201", res.StatusCode)
	}

	body, _ := io.ReadAll(res.Body)
	var out struct {
		Ticket    string `json:"ticket"`
		ExpiresIn int    `json:"expiresIn"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if out.Ticket == "" || out.ExpiresIn != 30 {
		t.Fatalf("unexpected response: %+v", out)
	}

	sum := sha256.Sum256([]byte(out.Ticket))
	key := wsTicketPrefix + hex.EncodeToString(sum[:])
	stored, err := rdb.Get(context.Background(), key).Result()
	if err != nil {
		t.Fatalf("expected ticket hash in redis: %v", err)
	}
	var payload wsTicketPayload
	if err := json.Unmarshal([]byte(stored), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.UserID != "user-123" || payload.ProjectID != "proj-1" {
		t.Fatalf("stored payload = %+v", payload)
	}
	ttl := mr.TTL(key)
	if ttl <= 0 || ttl > 30*time.Second {
		t.Fatalf("unexpected ttl: %v", ttl)
	}
}

func TestMintWSTicket_Returns503WithoutRedis(t *testing.T) {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		auth.SetFiber(c, auth.Identity{UserID: "user-123", AuthMethod: "clerk"})
		return c.Next()
	})
	h := NewWSTicketHandler(nil)
	app.Post("/api/v1/projects/:projectid/ws-ticket", h.Mint)

	res, err := app.Test(httptest.NewRequest("POST", "/api/v1/projects/proj-1/ws-ticket", nil))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", res.StatusCode)
	}
}

func TestMintWSTicket_RejectsInvalidProjectID(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		auth.SetFiber(c, auth.Identity{UserID: "user-123", AuthMethod: "clerk"})
		return c.Next()
	})
	h := NewWSTicketHandler(rdb)
	app.Post("/api/v1/projects/:projectid/ws-ticket", h.Mint)

	res, err := app.Test(httptest.NewRequest("POST", "/api/v1/projects/bad:id/ws-ticket", nil))
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.StatusCode)
	}
}
