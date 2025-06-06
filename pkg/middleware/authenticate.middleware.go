package middleware

import (
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/gofiber/fiber/v2"
)

func AuthenticateMiddleware(c *fiber.Ctx) error {
	// Example authentication logic
	sessionToken := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")

	if sessionToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	claim, err := jwt.Verify(c.Context(), &jwt.VerifyParams{
		Token: sessionToken,
	})
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: " + err.Error(),
		})
	}

	usr, err := user.Get(c.Context(), claim.Subject)
	if err != nil || usr == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: " + err.Error(),
		})
	}

	c.Locals("user", usr)
	c.Locals("userId", claim.Subject)

	// Continue to next handler
	return c.Next()
}
