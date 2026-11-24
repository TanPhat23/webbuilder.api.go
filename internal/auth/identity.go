// Package auth defines the identity boundary between authentication providers
// (Clerk today, first-party auth later) and the rest of the application.
package auth

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

const (
	fiberIdentityKey = "auth.identity"
	fiberUserIDKey   = "userId"
)

type ctxKey struct{}

// Identity represents the authenticated principal for the current request.
type Identity struct {
	UserID     string
	Email      string
	Scopes     []string
	AuthMethod string
}

// SetFiber stores identity on the Fiber request locals and preserves the
// legacy "userId" local for handlers that read it directly.
func SetFiber(c fiber.Ctx, id Identity) {
	if c == nil {
		return
	}
	c.Locals(fiberIdentityKey, id)
	c.Locals(fiberUserIDKey, id.UserID)
}

// FromFiber extracts the authenticated Identity from Fiber locals. It falls
// back to the legacy "userId" local so unit tests that populate c.Locals("userId")
// continue to work seamlessly.
func FromFiber(c fiber.Ctx) (Identity, bool) {
	if c == nil {
		return Identity{}, false
	}
	if id, ok := c.Locals(fiberIdentityKey).(Identity); ok && id.UserID != "" {
		return id, true
	}
	if uid, ok := c.Locals(fiberUserIDKey).(string); ok && uid != "" {
		return Identity{UserID: uid}, true
	}
	return Identity{}, false
}

// WithIdentity attaches id to a standard context.Context.
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext retrieves the Identity attached to ctx, if any.
func FromContext(ctx context.Context) (Identity, bool) {
	if ctx == nil {
		return Identity{}, false
	}
	id, ok := ctx.Value(ctxKey{}).(Identity)
	return id, ok && id.UserID != ""
}
