package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/agenttokens"
	"my-go-app/internal/models"
)

type tokenManagerStub struct {
	created []string
	revoked string
	err     error
}

func (s *tokenManagerStub) Create(_ context.Context, userID, name string) (string, models.AgentToken, error) {
	if s.err != nil {
		return "", models.AgentToken{}, s.err
	}
	s.created = append(s.created, userID+":"+name)
	return "wbk_secretvalue", models.AgentToken{Id: "t1", UserId: userID, Name: name, Prefix: "wbk_secr"}, nil
}
func (s *tokenManagerStub) List(context.Context, string) ([]models.AgentToken, error) {
	return []models.AgentToken{{Id: "t1", Name: "cli", Prefix: "wbk_secr", TokenHash: "must-never-be-serialised"}}, nil
}
func (s *tokenManagerStub) Revoke(_ context.Context, _, id string) error {
	if id == "missing" {
		return agenttokens.ErrTokenNotFound
	}
	s.revoked = id
	return nil
}

func tokenApp(stub *tokenManagerStub) *fiber.App {
	app := fiber.New()
	h := NewAgentTokenHandler(stub, "https://mcp.example.test/")
	app.Use(func(c fiber.Ctx) error { c.Locals("userId", "user-1"); return c.Next() })
	app.Post("/agent-tokens", h.Create)
	app.Get("/agent-tokens", h.List)
	app.Delete("/agent-tokens/:id", h.Revoke)
	return app
}

func TestCreateReturnsTheSecretAndMcpUrlOnce(t *testing.T) {
	stub := &tokenManagerStub{}
	req := httptest.NewRequest("POST", "/agent-tokens", strings.NewReader(`{"name":"laptop"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := tokenApp(stub).Test(req)
	if err != nil || resp.StatusCode != 201 {
		t.Fatalf("status=%v err=%v", resp, err)
	}
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body["token"] != "wbk_secretvalue" || body["mcpUrl"] != "https://mcp.example.test/mcp" || body["id"] != "t1" {
		t.Fatalf("body=%v", body)
	}
	if len(stub.created) != 1 || stub.created[0] != "user-1:laptop" {
		t.Fatalf("created=%v", stub.created)
	}
}

func TestListNeverReturnsSecretMaterial(t *testing.T) {
	resp, _ := tokenApp(&tokenManagerStub{}).Test(httptest.NewRequest("GET", "/agent-tokens", nil))
	raw, _ := io.ReadAll(resp.Body)
	for _, leak := range []string{"must-never-be-serialised", `"token"`, "tokenHash", "TokenHash"} {
		if strings.Contains(string(raw), leak) {
			t.Fatalf("list leaked %q: %s", leak, string(raw))
		}
	}
}

func TestCreateMapsServiceErrors(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{{agenttokens.ErrBadName, 400}, {agenttokens.ErrTooManyTokens, 409}}
	for _, c := range cases {
		req := httptest.NewRequest("POST", "/agent-tokens", strings.NewReader(`{"name":"x"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := tokenApp(&tokenManagerStub{err: c.err}).Test(req)
		if resp.StatusCode != c.want {
			t.Fatalf("%v: status %d want %d", c.err, resp.StatusCode, c.want)
		}
	}
}

func TestRevokeStatuses(t *testing.T) {
	stub := &tokenManagerStub{}
	app := tokenApp(stub)
	if resp, _ := app.Test(httptest.NewRequest("DELETE", "/agent-tokens/t1", nil)); resp.StatusCode != 204 || stub.revoked != "t1" {
		t.Fatalf("status=%d revoked=%q", resp.StatusCode, stub.revoked)
	}
	if resp, _ := app.Test(httptest.NewRequest("DELETE", "/agent-tokens/missing", nil)); resp.StatusCode != 404 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}
