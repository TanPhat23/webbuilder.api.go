package authz

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/auth"
)

// SubjectAccess holds a user's resolved role and effective permissions on a project.
type SubjectAccess struct {
	UserID         string       `json:"userId"`
	ProjectID      string       `json:"projectId"`
	Role           Role         `json:"role"`
	IsOwner        bool         `json:"isOwner"`
	IsCollaborator bool         `json:"isCollaborator"`
	Permissions    []Permission `json:"permissions"`
}

// Has reports whether this resolved subject grants perm.
func (s *SubjectAccess) Has(perm Permission) bool {
	if s == nil {
		return false
	}
	return HasPermission(s.Role, perm)
}

// AccessLookup resolves a user's membership on a project from persistence.
// Implementations must return ErrNotFound when the project does not exist (or is
// deleted) or when the user is neither the owner nor a collaborator.
type AccessLookup interface {
	LookupProjectAccess(ctx context.Context, projectID, userID string) (*SubjectAccess, error)
}

// Authorizer evaluates project-scoped permissions with request-level memoization.
type Authorizer struct {
	lookup AccessLookup
}

// NewAuthorizer constructs an Authorizer backed by lookup.
func NewAuthorizer(lookup AccessLookup) *Authorizer {
	return &Authorizer{lookup: lookup}
}

type accessCacheKey struct{}

type accessCache struct {
	entries map[string]*SubjectAccess
}

// WithRequestCache attaches a per-request memoization cache to ctx so multiple
// authorization checks for the same (projectID, userID) issue at most one DB read.
func WithRequestCache(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, ok := ctx.Value(accessCacheKey{}).(*accessCache); ok {
		return ctx
	}
	return context.WithValue(ctx, accessCacheKey{}, &accessCache{
		entries: make(map[string]*SubjectAccess),
	})
}

// WithSubjectAccess stores a resolved SubjectAccess in ctx's request cache
// (creating a cache if needed) so downstream services reuse it with zero queries.
func WithSubjectAccess(ctx context.Context, access *SubjectAccess) context.Context {
	ctx = WithRequestCache(ctx)
	if access == nil {
		return ctx
	}
	if cache, ok := ctx.Value(accessCacheKey{}).(*accessCache); ok {
		cache.entries[cacheKey(access.ProjectID, access.UserID)] = access
	}
	return ctx
}

// SubjectFromFiber returns the SubjectAccess cached on c by RequireProjectPermission, if any.
func SubjectFromFiber(c fiber.Ctx) (*SubjectAccess, bool) {
	if c == nil {
		return nil, false
	}
	access, ok := c.Locals("authz.access").(*SubjectAccess)
	return access, ok && access != nil
}

func cacheKey(projectID, userID string) string {
	return projectID + "\x00" + userID
}

// Resolve returns the user's SubjectAccess on projectID, memoizing the result if
// ctx carries a request cache.
func (a *Authorizer) Resolve(ctx context.Context, projectID, userID string) (*SubjectAccess, error) {
	if projectID == "" || userID == "" {
		return nil, ErrNotFound
	}
	if a == nil || a.lookup == nil {
		return nil, errors.New("authorizer is not configured")
	}

	key := cacheKey(projectID, userID)
	if cache, ok := ctx.Value(accessCacheKey{}).(*accessCache); ok {
		if hit, exists := cache.entries[key]; exists {
			return hit, nil
		}
	}

	access, err := a.lookup.LookupProjectAccess(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if access == nil {
		return nil, ErrNotFound
	}
	if len(access.Permissions) == 0 {
		access.Permissions = PermissionsForRole(access.Role)
	}
	if cache, ok := ctx.Value(accessCacheKey{}).(*accessCache); ok {
		cache.entries[key] = access
	}
	return access, nil
}

// Authorize checks that userID holds perm on projectID. Non-members receive
// ErrForbidden so mutation callers cannot probe vs. access-denied ambiguously,
// while missing/deleted projects still surface ErrNotFound.
func (a *Authorizer) Authorize(ctx context.Context, projectID, userID string, perm Permission) (*SubjectAccess, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}
	if projectID == "" {
		return nil, fmt.Errorf("%w: projectId is required", ErrNotFound)
	}
	access, err := a.Resolve(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}
	if !access.Has(perm) {
		return nil, fmt.Errorf("%w: role %q lacks %q", ErrForbidden, access.Role, perm)
	}
	return access, nil
}

// AuthorizeAny checks that userID holds at least one of perms on projectID.
func (a *Authorizer) AuthorizeAny(ctx context.Context, projectID, userID string, perms ...Permission) (*SubjectAccess, error) {
	if userID == "" {
		return nil, ErrUnauthorized
	}
	if projectID == "" {
		return nil, fmt.Errorf("%w: projectId is required", ErrNotFound)
	}
	access, err := a.Resolve(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}
	for _, perm := range perms {
		if access.Has(perm) {
			return access, nil
		}
	}
	return nil, fmt.Errorf("%w: role %q lacks required permission", ErrForbidden, access.Role)
}

// AuthorizeFiber resolves the authenticated user from c and checks perm on the
// project ID found in route param paramName (with optional request caching).
func (a *Authorizer) AuthorizeFiber(c fiber.Ctx, paramName string, perm Permission) (*SubjectAccess, error) {
	id, ok := auth.FromFiber(c)
	if !ok {
		return nil, ErrUnauthorized
	}
	projectID := c.Params(paramName)
	if projectID == "" {
		return nil, fiber.NewError(fiber.StatusBadRequest, paramName+" is required")
	}
	ctx := WithRequestCache(c.RequestCtx())
	c.SetContext(ctx)
	return a.Authorize(ctx, projectID, id.UserID, perm)
}

// RequireProjectPermission returns Fiber middleware that enforces perm on the
// project identified by route parameter paramName and caches SubjectAccess on
// both c.Locals and c.RequestCtx().
func (a *Authorizer) RequireProjectPermission(paramName string, perm Permission) fiber.Handler {
	return func(c fiber.Ctx) error {
		access, err := a.AuthorizeFiber(c, paramName, perm)
		if err != nil {
			status, ok := HTTPStatus(err)
			if ok {
				return c.Status(status).JSON(fiber.Map{"error": err.Error()})
			}
			return err
		}
		c.Locals("authz.access", access)
		c.SetContext(WithSubjectAccess(c.RequestCtx(), access))
		return c.Next()
	}
}
