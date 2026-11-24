// Package authz implements Role-Based Access Control (RBAC) as the single
// source of truth for both REST handlers/services and the realtime WebSocket server.
package authz

import (
	"errors"
	"net/http"
)

// Role represents a user's effective role on a project.
type Role string

const (
	RoleOwner  Role = "owner"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

// Permission is a fine-grained capability on a project resource.
type Permission string

const (
	// Project permissions
	PermProjectView     Permission = "project:view"
	PermProjectEdit     Permission = "project:edit"
	PermProjectDelete   Permission = "project:delete"
	PermProjectPublish  Permission = "project:publish"
	PermProjectSettings Permission = "project:settings"
	PermProjectTransfer Permission = "project:transfer"

	// Element & realtime permissions
	PermElementView    Permission = "element:view"
	PermElementCreate  Permission = "element:create"
	PermElementEdit    Permission = "element:edit"
	PermElementDelete  Permission = "element:delete"
	PermElementReorder Permission = "element:reorder"
	PermLogicEdit      Permission = "logic:edit"

	// Page permissions
	PermPageView   Permission = "page:view"
	PermPageCreate Permission = "page:create"
	PermPageEdit   Permission = "page:edit"
	PermPageDelete Permission = "page:delete"

	// Collaborator & invitation permissions
	PermCollaboratorView   Permission = "collaborator:view"
	PermCollaboratorInvite Permission = "collaborator:invite"
	PermCollaboratorEdit   Permission = "collaborator:edit"
	PermCollaboratorRemove Permission = "collaborator:remove"

	// Comment permissions
	PermCommentView      Permission = "comment:view"
	PermCommentCreate    Permission = "comment:create"
	PermCommentEditOwn   Permission = "comment:edit:own"
	PermCommentEditAll   Permission = "comment:edit:all"
	PermCommentDeleteOwn Permission = "comment:delete:own"
	PermCommentDeleteAll Permission = "comment:delete:all"
	PermCommentResolve   Permission = "comment:resolve"

	// Snapshot permissions
	PermSnapshotView    Permission = "snapshot:view"
	PermSnapshotCreate  Permission = "snapshot:create"
	PermSnapshotRestore Permission = "snapshot:restore"
	PermSnapshotDelete  Permission = "snapshot:delete"

	// Export permissions
	PermExportCode     Permission = "export:code"
	PermExportTemplate Permission = "export:template"

	// CMS permissions
	PermCMSView    Permission = "cms:view"
	PermCMSEdit    Permission = "cms:edit"
	PermCMSPublish Permission = "cms:publish"
)

var (
	// ErrUnauthorized indicates missing or invalid authentication credentials (HTTP 401).
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden indicates the authenticated user lacks the required permission (HTTP 403).
	ErrForbidden = errors.New("forbidden")
	// ErrNotFound indicates the target project or resource does not exist or is hidden (HTTP 404).
	ErrNotFound = errors.New("not found")
)

var rolePermissions = map[Role][]Permission{
	RoleOwner: {
		PermProjectView,
		PermProjectEdit,
		PermProjectDelete,
		PermProjectPublish,
		PermProjectSettings,
		PermProjectTransfer,
		PermElementView,
		PermElementCreate,
		PermElementEdit,
		PermElementDelete,
		PermElementReorder,
		PermLogicEdit,
		PermPageView,
		PermPageCreate,
		PermPageEdit,
		PermPageDelete,
		PermCollaboratorView,
		PermCollaboratorInvite,
		PermCollaboratorEdit,
		PermCollaboratorRemove,
		PermCommentView,
		PermCommentCreate,
		PermCommentEditOwn,
		PermCommentEditAll,
		PermCommentDeleteOwn,
		PermCommentDeleteAll,
		PermCommentResolve,
		PermSnapshotView,
		PermSnapshotCreate,
		PermSnapshotRestore,
		PermSnapshotDelete,
		PermExportCode,
		PermExportTemplate,
		PermCMSView,
		PermCMSEdit,
		PermCMSPublish,
	},
	RoleEditor: {
		PermProjectView,
		PermProjectEdit,
		PermElementView,
		PermElementCreate,
		PermElementEdit,
		PermElementDelete,
		PermElementReorder,
		PermLogicEdit,
		PermPageView,
		PermPageCreate,
		PermPageEdit,
		PermPageDelete,
		PermCollaboratorView,
		PermCommentView,
		PermCommentCreate,
		PermCommentEditOwn,
		PermCommentDeleteOwn,
		PermCommentResolve,
		PermSnapshotView,
		PermSnapshotCreate,
		PermExportCode,
		PermCMSView,
		PermCMSEdit,
	},
	RoleViewer: {
		PermProjectView,
		PermElementView,
		PermPageView,
		PermCollaboratorView,
		PermCommentView,
		PermCommentCreate,
		PermCommentEditOwn,
		PermCommentDeleteOwn,
		PermSnapshotView,
		PermCMSView,
	},
}

var rolePermissionSets = buildPermissionSets(rolePermissions)

func buildPermissionSets(src map[Role][]Permission) map[Role]map[Permission]struct{} {
	out := make(map[Role]map[Permission]struct{}, len(src))
	for role, perms := range src {
		set := make(map[Permission]struct{}, len(perms))
		for _, p := range perms {
			set[p] = struct{}{}
		}
		out[role] = set
	}
	return out
}

// IsValidRole reports whether role is a recognized project role.
func IsValidRole(role Role) bool {
	_, ok := rolePermissions[role]
	return ok
}

// IsAssignableCollaboratorRole reports whether role may be stored on a Collaborator
// or Invitation row. Owner lives exclusively on Project.OwnerId.
func IsAssignableCollaboratorRole(role Role) bool {
	return role == RoleEditor || role == RoleViewer
}

// PermissionsForRole returns a fresh slice of permissions granted to role.
func PermissionsForRole(role Role) []Permission {
	perms := rolePermissions[role]
	if len(perms) == 0 {
		return []Permission{}
	}
	out := make([]Permission, len(perms))
	copy(out, perms)
	return out
}

// HasPermission checks whether role grants perm in O(1).
func HasPermission(role Role, perm Permission) bool {
	set, ok := rolePermissionSets[role]
	if !ok {
		return false
	}
	_, allowed := set[perm]
	return allowed
}

// HTTPStatus maps authz sentinel errors to standard HTTP status codes.
func HTTPStatus(err error) (int, bool) {
	switch {
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, true
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, true
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, true
	default:
		return 0, false
	}
}
