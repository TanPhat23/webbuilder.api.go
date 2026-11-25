package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/auth"
)

// IsAdminUser reports whether userID is listed in the comma-separated ADMIN_USER_IDS env var.
func IsAdminUser(userID string) bool {
	if userID == "" {
		return false
	}
	raw := os.Getenv("ADMIN_USER_IDS")
	for entry := range strings.SplitSeq(raw, ",") {
		if strings.TrimSpace(entry) == userID {
			return true
		}
	}
	return false
}

// RequireAdmin allows only authenticated users listed in ADMIN_USER_IDS.
func RequireAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		id, ok := auth.FromFiber(c)
		if !ok || id.UserID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}
		if !IsAdminUser(id.UserID) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: admin access required",
			})
		}
		return c.Next()
	}
}
