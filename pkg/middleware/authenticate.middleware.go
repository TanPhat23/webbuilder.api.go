package middleware

import (
	"fmt"
	"log"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/gofiber/fiber/v2"
)

func AuthenticateMiddleware(c *fiber.Ctx) error {
	// Example authentication logic
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		log.Println("Missing Authorization header")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: Missing Authorization header",
		})
	}
	sessionToken := strings.TrimPrefix(authHeader, "Bearer ")

	if sessionToken == "" {
		log.Printf("Authorization header present but token missing or malformed: %s", authHeader)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: Token missing or malformed",
		})
	}

	claim, err := jwt.Verify(c.Context(), &jwt.VerifyParams{
		Token: sessionToken,
	})
	if err != nil {
		log.Printf("JWT verification failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: " + err.Error(),
		})
	}

	usr, err := user.Get(c.Context(), claim.Subject)
	if err != nil || usr == nil {
		log.Printf("User lookup failed for subject %s: %v", claim.Subject, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: " + err.Error(),
		})
	}

	c.Locals("user", usr)
	fmt.Println(claim.Subject)
	c.Locals("userId", claim.Subject)

	// Continue to next handler
	return c.Next()
}
