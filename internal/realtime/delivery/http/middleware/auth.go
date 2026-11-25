package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"

	clerk "github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"my-go-app/internal/auth"
	"my-go-app/internal/realtime/config"
	"my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/utils/logger"
)

const wsTicketRedisPrefix = "ws:ticket:"

type wsTicketRecord struct {
	UserID    string `json:"userId"`
	ProjectID string `json:"projectId"`
}

type clerkUserPayload struct {
	EmailAddresses []struct {
		EmailAddress string `json:"emailAddress"`
	} `json:"emailAddresses"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	ImageURL  string `json:"imageUrl"`
}

// AuthMiddleware configures /ws authentication without a Redis ticket store (DEV_MODE only).
// Use AuthMiddlewareWithRedis in production to authenticate via single-use ?ticket= credentials.
func AuthMiddleware(cfg *config.Config) fiber.Handler {
	return AuthMiddlewareWithRedis(cfg, nil)
}

// AuthMiddlewareWithRedis authenticates /ws connections exclusively via:
//   - DEV_MODE=true (only when ENV is not production/prod/staging and Clerk key is not sk_live_)
//   - ?ticket=<opaque> single-use ticket consumed atomically from Redis
func AuthMiddlewareWithRedis(cfg *config.Config, rdb *redis.Client) fiber.Handler {
	userClient := newClerkUserClient(cfg)

	return func(c fiber.Ctx) error {
		if cfg != nil && cfg.DevMode && !isProdEnv(cfg) {
			return authenticateDevUser(c)
		}

		if ticket := strings.TrimSpace(c.Query("ticket")); ticket != "" {
			return authenticateWSTicket(c, rdb, userClient, ticket)
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: Missing or invalid WebSocket ticket",
		})
	}
}

func isProdEnv(cfg *config.Config) bool {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	if env == "production" || env == "prod" || env == "staging" {
		return true
	}
	return cfg != nil && strings.HasPrefix(cfg.ClerkSecretKey, "sk_live_")
}

func authenticateWSTicket(c fiber.Ctx, rdb *redis.Client, userClient *user.Client, rawTicket string) error {
	if rdb == nil {
		logger.Default().Errorf("AuthMiddleware: Redis is not configured for WebSocket ticket verification")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server misconfigured: missing ticket store",
		})
	}

	sum := sha256.Sum256([]byte(rawTicket))
	key := wsTicketRedisPrefix + hex.EncodeToString(sum[:])

	ctx := c.Context()
	raw, err := rdb.GetDel(ctx, key).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Default().Errorf("AuthMiddleware: failed to consume ws ticket: %v", err)
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: Invalid or expired ticket",
		})
	}

	var payload wsTicketRecord
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || payload.UserID == "" || payload.ProjectID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: Invalid ticket payload",
		})
	}

	mappedUser := &entity.User{ID: payload.UserID}
	if userClient != nil {
		mappedUser = fetchMappedUser(ctx, userClient, payload.UserID)
	}

	auth.SetFiber(c, auth.Identity{
		UserID:     mappedUser.ID,
		Email:      mappedUser.Email,
		AuthMethod: "ws_ticket",
	})
	c.Locals("userId", mappedUser.ID)
	c.Locals("user", mappedUser)
	c.Locals("ticketProjectId", payload.ProjectID)
	return c.Next()
}

func newClerkUserClient(cfg *config.Config) *user.Client {
	if cfg == nil || cfg.ClerkSecretKey == "" {
		return nil
	}
	clerkCfg := &clerk.ClientConfig{}
	clerkCfg.Key = clerk.String(cfg.ClerkSecretKey)
	return user.NewClient(clerkCfg)
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
	return c.Next()
}

func fetchMappedUser(ctx context.Context, userClient *user.Client, subject string) *entity.User {
	mapped := &entity.User{ID: subject}
	usr, err := userClient.Get(ctx, subject)
	if err != nil {
		logger.Default().Debugf("AuthMiddleware: failed to fetch user from Clerk: %v", err)
		return mapped
	}
	payload, ok := decodeClerkUserPayload(usr)
	if !ok {
		return mapped
	}
	applyClerkUserFields(mapped, payload)
	return mapped
}

func decodeClerkUserPayload(usr any) (clerkUserPayload, bool) {
	// Marshal through JSON so mapping does not depend on exact Clerk SDK field pointer shapes.
	raw, err := json.Marshal(usr)
	if err != nil {
		return clerkUserPayload{}, false
	}
	var payload clerkUserPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return clerkUserPayload{}, false
	}
	return payload, true
}

func applyClerkUserFields(mapped *entity.User, payload clerkUserPayload) {
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
}
