package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/auth"
	"my-go-app/internal/authz"
	"my-go-app/internal/models"
)

const (
	agentTokenPrefix   = "wbk_"
	maxAgentTokenLen   = 128
	agentTokenTouchGap = time.Minute
)

// AgentTokenLookup resolves an active (non-revoked) agent token by its SHA-256 hex hash.
type AgentTokenLookup interface {
	GetActiveByHash(ctx context.Context, hash string) (*models.AgentToken, error)
	Touch(ctx context.Context, id string, at time.Time) error
}

// ValidateAuthEnv refuses to start when TEST_AUTH_BYPASS is enabled outside
// the explicit local/test allowlist or when a live Clerk key is configured.
func ValidateAuthEnv() error {
	if !strings.EqualFold(os.Getenv("TEST_AUTH_BYPASS"), "true") {
		return nil
	}
	if strings.HasPrefix(strings.TrimSpace(os.Getenv("CLERK_SECRET_KEY")), "sk_live_") {
		return errors.New("TEST_AUTH_BYPASS cannot be enabled with a live CLERK_SECRET_KEY")
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("ENV"))) {
	case "test", "local", "development":
		return nil
	default:
		return errors.New("TEST_AUTH_BYPASS is only allowed when ENV is test, local, or development")
	}
}

func isTestBypassAllowed() bool {
	return ValidateAuthEnv() == nil && strings.EqualFold(os.Getenv("TEST_AUTH_BYPASS"), "true")
}

func isPublicRoute(c fiber.Ctx) bool {
	method := c.Method()
	path := c.Path()
	if method == fiber.MethodGet || method == fiber.MethodHead {
		if strings.HasPrefix(path, "/api/v1/pages/public/") ||
			strings.HasPrefix(path, "/api/v1/projects/public/") ||
			strings.HasPrefix(path, "/api/v1/elements/public/") ||
			strings.HasPrefix(path, "/api/v1/public/") {
			return true
		}
	}
	if method == fiber.MethodPost && strings.HasPrefix(path, "/api/v1/public/projects/") {
		rest := strings.TrimPrefix(path, "/api/v1/public/projects/")
		parts := strings.Split(rest, "/")
		if len(parts) == 4 && parts[0] != "" && parts[1] == "collections" && parts[2] != "" && parts[3] == "items" {
			return true
		}
	}
	return false
}

func hashAgentSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func attachIdentity(c fiber.Ctx, id auth.Identity) {
	auth.SetFiber(c, id)
	ctx := auth.WithIdentity(c.RequestCtx(), id)
	ctx = authz.WithRequestCache(ctx)
	c.SetContext(ctx)
}

func Authenticate(c fiber.Ctx) error {
	return NewAuthenticate(nil)(c)
}

func NewAuthenticate(tokenLookup AgentTokenLookup) fiber.Handler {
	return func(c fiber.Ctx) error {
		if isPublicRoute(c) {
			return c.Next()
		}

		// Guarded test bypass for isolated local test stack only.
		if isTestBypassAllowed() {
			if testUserID := strings.TrimSpace(c.Get("X-Test-User-Id")); testUserID != "" {
				attachIdentity(c, auth.Identity{
					UserID:     testUserID,
					AuthMethod: "test_bypass",
				})
				return c.Next()
			}
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			log.Println("Missing Authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Missing Authorization header",
			})
		}
		sessionToken := strings.TrimPrefix(authHeader, "Bearer ")
		if sessionToken == "" || sessionToken == authHeader {
			log.Println("Authorization header present but token missing or malformed")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Token missing or malformed",
			})
		}

		// Personal access token authentication (wbk_...) for MCP & CLI delegation.
		if strings.HasPrefix(sessionToken, agentTokenPrefix) {
			if tokenLookup == nil || len(sessionToken) <= len(agentTokenPrefix) || len(sessionToken) > maxAgentTokenLen {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Unauthorized: Invalid agent token",
				})
			}
			rec, err := tokenLookup.GetActiveByHash(c.RequestCtx(), hashAgentSecret(sessionToken))
			if err != nil {
				log.Printf("Agent token lookup failed: %v", err)
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Unauthorized: Agent token verification failed",
				})
			}
			if rec == nil || rec.RevokedAt != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Unauthorized: Invalid or revoked agent token",
				})
			}
			now := time.Now()
			if rec.LastUsedAt == nil || now.Sub(*rec.LastUsedAt) >= agentTokenTouchGap {
				_ = tokenLookup.Touch(c.RequestCtx(), rec.Id, now)
			}
			attachIdentity(c, auth.Identity{
				UserID:     rec.UserId,
				AuthMethod: "agent_token",
			})
			return c.Next()
		}

		claim, err := jwt.Verify(c.RequestCtx(), &jwt.VerifyParams{
			Token: sessionToken,
		})
		if err != nil {
			log.Printf("JWT verification failed: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid token",
			})
		}

		attachIdentity(c, auth.Identity{
			UserID:     claim.Subject,
			AuthMethod: "clerk",
		})

		return c.Next()
	}
}

// fiber:context-methods migrated
