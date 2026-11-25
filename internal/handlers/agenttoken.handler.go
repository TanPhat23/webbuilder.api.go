package handlers

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/agenttokens"
	"my-go-app/internal/auth"
	"my-go-app/internal/models"
	"my-go-app/pkg/utils"
)

type AgentTokenManager interface {
	Create(ctx context.Context, userID, name string) (string, models.AgentToken, error)
	List(ctx context.Context, userID string) ([]models.AgentToken, error)
	Revoke(ctx context.Context, userID, id string) error
}

type AgentTokenHandler struct {
	tokens AgentTokenManager
	// mcpBaseURL is the public address of the MCP service; the editor prints
	// mcpBaseURL + "/mcp" for the user to paste into their CLI.
	mcpBaseURL string
}

func NewAgentTokenHandler(tokens AgentTokenManager, mcpBaseURL string) *AgentTokenHandler {
	return &AgentTokenHandler{tokens: tokens, mcpBaseURL: strings.TrimRight(mcpBaseURL, "/")}
}

// PublicMCPURL is read at startup from PUBLIC_MCP_URL (for example
// https://mcp.example.com). The default suits local development.
func PublicMCPURL() string {
	if v := os.Getenv("PUBLIC_MCP_URL"); v != "" {
		return v
	}
	return "http://localhost:8090"
}

func requireInteractiveUser(c fiber.Ctx) (string, bool, error) {
	userID, err := utils.RequireUserID(c)
	if err != nil {
		return "", false, err
	}
	if id, ok := auth.FromFiber(c); ok && id.AuthMethod == "agent_token" {
		return "", false, c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "agent tokens cannot manage agent tokens; interactive session required",
		})
	}
	return userID, true, nil
}

func (h *AgentTokenHandler) Create(c fiber.Ctx) error {
	userID, ok, err := requireInteractiveUser(c)
	if !ok {
		return err
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	secret, tok, err := h.tokens.Create(c.Context(), userID, body.Name)
	switch {
	case errors.Is(err, agenttokens.ErrBadName):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, agenttokens.ErrTooManyTokens):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case err != nil:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create token"})
	}
	// The secret appears in this response only.
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": tok.Id, "name": tok.Name, "prefix": tok.Prefix, "createdAt": tok.CreatedAt,
		"token": secret, "mcpUrl": h.mcpBaseURL + "/mcp",
	})
}

func (h *AgentTokenHandler) List(c fiber.Ctx) error {
	userID, ok, err := requireInteractiveUser(c)
	if !ok {
		return err
	}
	tokens, err := h.tokens.List(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not list tokens"})
	}
	return c.JSON(tokens)
}

func (h *AgentTokenHandler) Revoke(c fiber.Ctx) error {
	userID, ok, err := requireInteractiveUser(c)
	if !ok {
		return err
	}
	err = h.tokens.Revoke(c.Context(), userID, c.Params("id"))
	if errors.Is(err, agenttokens.ErrTokenNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "token not found"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not revoke token"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
