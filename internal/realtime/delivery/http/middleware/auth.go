package middleware

import (
	"context"
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

	"github.com/gofiber/fiber/v3"
)

type clerkUserPayload struct {
	EmailAddresses []struct {
		EmailAddress string `json:"emailAddress"`
	} `json:"emailAddresses"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	ImageURL  string `json:"imageUrl"`
}

func AuthMiddleware(cfg *config.Config) fiber.Handler {
	jwksClient, userClient := newClerkClients(cfg)

	return func(c fiber.Ctx) error {
		if cfg != nil && cfg.DevMode {
			return authenticateDevUser(c)
		}

		token := extractSessionToken(c)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Missing or invalid token",
			})
		}

		if jwksClient == nil || userClient == nil {
			logger.Default().Errorf("AuthMiddleware: Clerk is not configured (missing CLERK_SECRET_KEY)")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Server misconfigured: missing clerk secret key",
			})
		}

		ctx := c.Context()
		subject, err := verifyClerkToken(ctx, jwksClient, token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid session token",
			})
		}

		mappedUser := fetchMappedUser(ctx, userClient, subject)
		c.Locals("userId", mappedUser.ID)
		c.Locals("user", mappedUser)
		c.Locals("token", token)
		return c.Next()
	}
}

func newClerkClients(cfg *config.Config) (*jwks.Client, *user.Client) {
	if cfg == nil || cfg.ClerkSecretKey == "" {
		return nil, nil
	}
	clerkCfg := &clerk.ClientConfig{}
	clerkCfg.Key = clerk.String(cfg.ClerkSecretKey)
	return jwks.NewClient(clerkCfg), user.NewClient(clerkCfg)
}

func authenticateDevUser(c fiber.Ctx) error {
	// One fixed identity (DEV_USER_ID) from the server's environment, never from the client.
	devUserId := os.Getenv("DEV_USER_ID")
	if devUserId == "" {
		devUserId = "dev-user-id"
	}
	logger.Default().Debugf("AuthMiddleware: DevMode active, using UserID: %s", devUserId)

	devUser := &entity.User{
		ID:    devUserId,
		Email: "dev@example.com",
	}
	c.Locals("userId", devUser.ID)
	c.Locals("user", devUser)
	c.Locals("token", devUserId)
	return c.Next()
}

func extractSessionToken(c fiber.Ctx) string {
	if authHeader := c.Get("Authorization"); authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return parts[1]
		}
	}
	if cookie := c.Cookies("__session"); cookie != "" {
		return cookie
	}
	return c.Query("token")
}

func verifyClerkToken(ctx context.Context, jwksClient *jwks.Client, token string) (string, error) {
	unsafeClaims, err := jwt.Decode(ctx, &jwt.DecodeParams{Token: token})
	if err != nil {
		logger.Default().Debugf("AuthMiddleware: jwt.Decode failed: %v", err)
		return "", err
	}
	jwk, err := jwt.GetJSONWebKey(ctx, &jwt.GetJSONWebKeyParams{
		KeyID:      unsafeClaims.KeyID,
		JWKSClient: jwksClient,
	})
	if err != nil {
		logger.Default().Debugf("AuthMiddleware: GetJSONWebKey failed: %v", err)
		return "", err
	}
	claims, err := jwt.Verify(ctx, &jwt.VerifyParams{Token: token, JWK: jwk})
	if err != nil {
		logger.Default().Debugf("AuthMiddleware: jwt.Verify failed: %v", err)
		return "", err
	}
	return claims.Subject, nil
}

func fetchMappedUser(ctx context.Context, userClient *user.Client, subject string) *entity.User {
	mapped := &entity.User{ID: subject}
	usr, err := userClient.Get(ctx, subject)
	if err != nil {
		logger.Default().Debugf("AuthMiddleware: failed to fetch user from Clerk: %v", err)
		return mapped
	}
	// Marshal through JSON so mapping does not depend on exact Clerk SDK field pointer shapes.
	raw, err := json.Marshal(usr)
	if err != nil {
		return mapped
	}
	var payload clerkUserPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return mapped
	}
	if len(payload.EmailAddresses) > 0 && payload.EmailAddresses[0].EmailAddress != "" {
		mapped.Email = payload.EmailAddresses[0].EmailAddress
	}
	if payload.FirstName != "" {
		mapped.FirstName = &payload.FirstName
	}
	if payload.LastName != "" {
		mapped.LastName = &payload.LastName
	}
	if payload.ImageURL != "" {
		mapped.ImageUrl = &payload.ImageURL
	}
	return mapped
}
