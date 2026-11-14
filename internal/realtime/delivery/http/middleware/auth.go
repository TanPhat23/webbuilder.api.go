package middleware

import (
	"encoding/json"
	"os"
	"strings"

	clerk "github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"

	"my-go-app/internal/realtime/config"
	"my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/utils/logger"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(cfg *config.Config) fiber.Handler {
	// Setup Clerk clients once when middleware is constructed
	var jwksClient *jwks.Client
	var userClient *user.Client
	if cfg != nil && cfg.ClerkSecretKey != "" {
		clerkCfg := &clerk.ClientConfig{}
		clerkCfg.Key = clerk.String(cfg.ClerkSecretKey)
		jwksClient = jwks.NewClient(clerkCfg)
		userClient = user.NewClient(clerkCfg)
	}

	return func(c *fiber.Ctx) error {
		// 1. Handle DevMode from application configuration: one fixed identity
		// (DEV_USER_ID) from the server's environment, never from the client.
		if cfg != nil && cfg.DevMode {
			devUserId := os.Getenv("DEV_USER_ID")
			if devUserId == "" {
				devUserId = "dev-user-id"
			}

			logger.Default().Debugf("AuthMiddleware: DevMode active, using UserID: %s", devUserId)

			user := &entity.User{
				ID:    devUserId,
				Email: "dev@example.com",
			}
			c.Locals("userId", user.ID)
			c.Locals("user", user)
			c.Locals("token", devUserId)
			return c.Next()
		}

		// Helper to extract token from Authorization header, __session cookie, or query param
		getSessionToken := func(c *fiber.Ctx) string {
			authHeader := c.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					return parts[1]
				}
			}

			if cookie := c.Cookies("__session"); cookie != "" {
				return cookie
			}

			return c.Query("token")
		}

		token := getSessionToken(c)

		// 4. Validate token presence
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Missing or invalid token",
			})
		}

		// Ensure Clerk is configured
		if jwksClient == nil || userClient == nil {
			logger.Default().Errorf("AuthMiddleware: Clerk is not configured (missing CLERK_SECRET_KEY)")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Server misconfigured: missing clerk secret key",
			})
		}

		// Verify the Clerk session JWT
		ctx := c.UserContext()

		unsafeClaims, err := jwt.Decode(ctx, &jwt.DecodeParams{
			Token: token,
		})
		if err != nil {
			logger.Default().Debugf("AuthMiddleware: jwt.Decode failed: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid session token",
			})
		}

		jwk, err := jwt.GetJSONWebKey(ctx, &jwt.GetJSONWebKeyParams{
			KeyID:      unsafeClaims.KeyID,
			JWKSClient: jwksClient,
		})
		if err != nil {
			logger.Default().Debugf("AuthMiddleware: GetJSONWebKey failed: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid session token",
			})
		}

		claims, err := jwt.Verify(ctx, &jwt.VerifyParams{
			Token: token,
			JWK:   jwk,
		})
		if err != nil {
			logger.Default().Debugf("AuthMiddleware: jwt.Verify failed: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid session token",
			})
		}

		// Retrieve user from Clerk (populate minimal fields)
		var mappedUser = &entity.User{
			ID: claims.Subject,
		}

		usr, err := userClient.Get(ctx, claims.Subject)
		if err != nil {
			// If fetching user fails, continue with user ID only but log it
			logger.Default().Debugf("AuthMiddleware: failed to fetch user from Clerk: %v", err)
		} else {
			// Marshal to map to avoid depending on exact SDK types
			var raw map[string]any
			if b, err := json.Marshal(usr); err == nil {
				if err := json.Unmarshal(b, &raw); err == nil {
					if v, ok := raw["emailAddresses"]; ok {
						if arr, ok := v.([]interface{}); ok && len(arr) > 0 {
							if e0, ok := arr[0].(map[string]interface{}); ok {
								if ea, ok := e0["emailAddress"].(string); ok && ea != "" {
									mappedUser.Email = ea
								}
							}
						}
					}
					if v, ok := raw["firstName"].(string); ok && v != "" {
						mappedUser.FirstName = new(string)
						*mappedUser.FirstName = v
					}
					if v, ok := raw["lastName"].(string); ok && v != "" {
						mappedUser.LastName = new(string)
						*mappedUser.LastName = v
					}
					if v, ok := raw["imageUrl"].(string); ok && v != "" {
						mappedUser.ImageUrl = new(string)
						*mappedUser.ImageUrl = v
					}
				}
			}
		}

		c.Locals("userId", mappedUser.ID)
		c.Locals("user", mappedUser)
		c.Locals("token", token)

		return c.Next()
	}
}
