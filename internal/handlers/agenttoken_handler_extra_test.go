package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/auth"
	"my-go-app/internal/models"
)

type agenttokenListFailStub struct {
	tokenManagerStub
	listErr, revokeErr error
}

func (s *agenttokenListFailStub) List(context.Context, string) ([]models.AgentToken, error) {
	return nil, s.listErr
}
func (s *agenttokenListFailStub) Revoke(context.Context, string, string) error { return s.revokeErr }

func agenttokenKit(userID string, stub AgentTokenManager) *fiber.App {
	h := NewAgentTokenHandler(stub, "https://mcp.example.test")
	return newKitApp(userID, func(app *fiber.App) {
		app.Post("/agent-tokens", h.Create)
		app.Get("/agent-tokens", h.List)
		app.Delete("/agent-tokens/:id", h.Revoke)
	})
}

func TestAgentTokenPublicMCPURL(t *testing.T) {
	t.Setenv("PUBLIC_MCP_URL", "")
	if got := PublicMCPURL(); got != "http://localhost:8090" {
		t.Fatalf("default = %q", got)
	}
	t.Setenv("PUBLIC_MCP_URL", "https://mcp.example.com")
	if got := PublicMCPURL(); got != "https://mcp.example.com" {
		t.Fatalf("configured = %q", got)
	}
}

func TestAgentTokenHandlerRejectsUnauthenticated(t *testing.T) {
	app := agenttokenKit("", &tokenManagerStub{})
	for _, c := range [][2]string{{"POST", "/agent-tokens"}, {"GET", "/agent-tokens"}, {"DELETE", "/agent-tokens/t1"}} {
		body := ""
		if c[0] == "POST" {
			body = `{"name":"x"}`
		}
		kitExpect(t, kitSend(t, app, c[0], c[1], body), 401, "Unauthorized")
	}
}

func TestAgentTokenCreateRejectsMalformedBody(t *testing.T) {
	stub := &tokenManagerStub{}
	kitExpect(t, kitSend(t, agenttokenKit("u1", stub), "POST", "/agent-tokens", `{`), 400, "invalid body")
	if len(stub.created) != 0 {
		t.Fatalf("service must not be called, created=%v", stub.created)
	}
}

func TestAgentTokenInternalErrorsAreGeneric(t *testing.T) {
	create := kitSend(t, agenttokenKit("u1", &tokenManagerStub{err: errInternal}), "POST", "/agent-tokens", `{"name":"x"}`)
	kitExpect(t, create, 500, "could not create token")

	fail := &agenttokenListFailStub{listErr: errInternal, revokeErr: errors.New("db down 10.0.0.5")}
	kitExpect(t, kitSend(t, agenttokenKit("u1", fail), "GET", "/agent-tokens", ""), 500, "could not list tokens")
	kitExpect(t, kitSend(t, agenttokenKit("u1", fail), "DELETE", "/agent-tokens/t1", ""), 500, "could not revoke token")
}

func TestAgentTokenHandlerRejectsAgentTokenSession(t *testing.T) {
	stub := &tokenManagerStub{}
	h := NewAgentTokenHandler(stub, "https://mcp.example.test")
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		auth.SetFiber(c, auth.Identity{UserID: "u1", AuthMethod: "agent_token"})
		return c.Next()
	})
	app.Post("/agent-tokens", h.Create)
	app.Get("/agent-tokens", h.List)
	app.Delete("/agent-tokens/:id", h.Revoke)

	for _, tc := range [][2]string{{"POST", "/agent-tokens"}, {"GET", "/agent-tokens"}, {"DELETE", "/agent-tokens/t1"}} {
		body := ""
		if tc[0] == "POST" {
			body = `{"name":"nested"}`
		}
		kitExpect(t, kitSend(t, app, tc[0], tc[1], body), 403, "interactive session required")
	}
	if len(stub.created) != 0 || stub.revoked != "" {
		t.Fatalf("service must not be invoked with agent_token auth: created=%v revoked=%q", stub.created, stub.revoked)
	}
}
