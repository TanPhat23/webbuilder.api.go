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
	FindActiveByHash(ctx context.Context, hash string) (*models.AgentToken, error)
	Touch(ctx context.Context, id string, at time.Time) error
}

// ValidateAuthEnv refuses to start when TEST_AUTH_BYPASS is enabled in production.
func ValidateAuthEnv() error {
	if strings.EqualFold(os.Getenv("TEST_AUTH_BYPASS"), "true") &&
		strings.EqualFold(os.Getenv("ENV"), "production") {
		return errors.New("TEST_AUTH_BYPASS cannot be enabled when ENV=production")
	}
	return nil
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

func AuthenticateMiddleware(c fiber.Ctx) error {
	return NewAuthenticateMiddleware(nil)(c)
}

func NewAuthenticateMiddleware(tokenLookup AgentTokenLookup) fiber.Handler {
	return func(c fiber.Ctx) error {
		// The public page routes permit only reads. The trailing slash keeps
		// the project ID slot on private page routes distinct from this prefix.
		publicPageRead := (c.Method() == fiber.MethodGet || c.Method() == fiber.MethodHead) &&
			strings.HasPrefix(c.Path(), "/api/v1/pages/public/")
		if publicPageRead || strings.HasPrefix(c.Path(), "/api/v1/projects/public") || strings.HasPrefix(c.Path(), "/api/v1/public") || strings.HasPrefix(c.Path(), "/api/v1/elements/public") {
			return c.Next()
		}

		// Guarded test bypass for isolated local test stack only.
		if strings.EqualFold(os.Getenv("TEST_AUTH_BYPASS"), "true") &&
			!strings.EqualFold(os.Getenv("ENV"), "production") {
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
			log.Printf("Authorization header present but token missing or malformed: %s", authHeader)
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
			rec, err := tokenLookup.FindActiveByHash(c.RequestCtx(), hashAgentSecret(sessionToken))
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
				"error": "Unauthorized: " + err.Error(),
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
