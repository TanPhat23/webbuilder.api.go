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

// Auth configures /ws authentication without a Redis ticket store (DEV_MODE only).
// Use AuthWithRedis in production to authenticate via single-use ?ticket= credentials.
func Auth(cfg *config.Config) fiber.Handler {
	return AuthWithRedis(cfg, nil)
}

// AuthWithRedis authenticates /ws connections exclusively via:
//   - DEV_MODE=true (only when ENV is not production/prod/staging and Clerk key is not sk_live_)
//   - ?ticket=<opaque> single-use ticket consumed atomically from Redis
func AuthWithRedis(cfg *config.Config, rdb *redis.Client) fiber.Handler {
	userClient := newClerkUserClient(cfg)

	return func(c fiber.Ctx) error {
		if cfg != nil && cfg.DevMode && !isProdEnv(cfg) {
			return authenticateDevUser(c)
		}

		if ticket := strings.TrimSpace(c.Query("ticket")); ticket != "" {
			return authenticateWSTicket(c, rdb, userClient, ticket)
		}

		return writeAuthError(c, fiber.StatusUnauthorized, "Unauthorized: Missing or invalid WebSocket ticket")
	}
}

func writeAuthError(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}

func attachAuthenticatedUser(c fiber.Ctx, u *entity.User) error {
	c.Locals("userId", u.ID)
	c.Locals("user", u)
	return c.Next()
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
		logger.Default().Errorf("Auth: Redis is not configured for WebSocket ticket verification")
		return writeAuthError(c, fiber.StatusInternalServerError, "Server misconfigured: missing ticket store")
	}

	sum := sha256.Sum256([]byte(rawTicket))
	key := wsTicketRedisPrefix + hex.EncodeToString(sum[:])

	ctx := c.Context()
	raw, err := rdb.GetDel(ctx, key).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Default().Errorf("Auth: failed to consume ws ticket: %v", err)
		}
		return writeAuthError(c, fiber.StatusUnauthorized, "Unauthorized: Invalid or expired ticket")
	}

	var payload wsTicketRecord
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || payload.UserID == "" || payload.ProjectID == "" {
		return writeAuthError(c, fiber.StatusUnauthorized, "Unauthorized: Invalid ticket payload")
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
	c.Locals("ticketProjectId", payload.ProjectID)
	return attachAuthenticatedUser(c, mappedUser)
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
	logger.Default().Debugf("Auth: DevMode active, using UserID: %s", devUserId)

	devUser := &entity.User{
		ID:    devUserId,
		Email: "dev@example.com",
	}
	return attachAuthenticatedUser(c, devUser)
}

func fetchMappedUser(ctx context.Context, userClient *user.Client, subject string) *entity.User {
	mapped := &entity.User{ID: subject}
	usr, err := userClient.Get(ctx, subject)
	if err != nil {
		logger.Default().Debugf("Auth: failed to fetch user from Clerk: %v", err)
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
