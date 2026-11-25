package handlers

import (
	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Use("/ws", func(c fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	wsCfg := websocket.Config{
		AllowEmptyOrigin: true,
	}
	if h.originsConfigured {
		if len(h.allowedOrigins) > 0 {
			wsCfg.Origins = h.allowedOrigins
		} else {
			// gofiber/contrib/v3/websocket treats an empty Origins slice as "*".
			// Use an unmatchable sentinel when origins were configured to empty (e.g. ENV=production without ALLOWED_ORIGINS).
			wsCfg.Origins = []string{"https://no-allowed-origins.invalid"}
		}
	}

	app.Get("/ws/:project", websocket.New(h.ServeWebSocket, wsCfg))
}
