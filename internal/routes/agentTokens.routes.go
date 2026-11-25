package routes

import (
	"my-go-app/internal/agenttokens"
	"my-go-app/internal/handlers"
	"my-go-app/internal/repositories"

	"github.com/gofiber/fiber/v3"
)

func RegisterAgentTokens(group fiber.Router, repos *repositories.Repositories) {
	handler := handlers.NewAgentTokenHandler(agenttokens.NewService(repos.AgentTokenRepository), handlers.PublicMCPURL())

	group.Post("/agent-tokens", handler.Create)
	group.Get("/agent-tokens", handler.List)
	group.Delete("/agent-tokens/:id", handler.Revoke)
}
