package authz

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestRolePermissionMatrix(t *testing.T) {
	t.Run("owner has every defined permission", func(t *testing.T) {
		all := PermissionsForRole(RoleOwner)
		if len(all) == 0 {
			t.Fatal("owner permissions must not be empty")
		}
		for _, perm := range all {
			if !HasPermission(RoleOwner, perm) {
				t.Errorf("owner missing %q", perm)
			}
		}
	})

	t.Run("editor permissions match policy", func(t *testing.T) {
		allowed := []Permission{
			PermProjectView, PermProjectEdit,
			PermElementCreate, PermElementEdit, PermElementDelete, PermElementReorder, PermLogicEdit,
			PermPageCreate, PermPageEdit, PermPageDelete,
			PermCollaboratorView,
			PermCommentCreate, PermCommentEditOwn, PermCommentDeleteOwn, PermCommentResolve,
			PermSnapshotView, PermSnapshotCreate,
			PermExportCode,
			PermCMSView, PermCMSEdit,
		}
		for _, p := range allowed {
			if !HasPermission(RoleEditor, p) {
				t.Errorf("editor should have %q", p)
			}
		}

		denied := []Permission{
			PermProjectDelete, PermProjectPublish, PermProjectSettings, PermProjectTransfer,
			PermCollaboratorInvite, PermCollaboratorEdit, PermCollaboratorRemove,
			PermCommentEditAll, PermCommentDeleteAll,
			PermSnapshotRestore, PermSnapshotDelete,
			PermExportTemplate, PermCMSPublish,
		}
		for _, p := range denied {
			if HasPermission(RoleEditor, p) {
				t.Errorf("editor must not have %q", p)
			}
		}
	})

	t.Run("viewer permissions match policy", func(t *testing.T) {
		allowed := []Permission{
			PermProjectView, PermElementView, PermPageView, PermCollaboratorView,
			PermCommentView, PermCommentCreate, PermCommentEditOwn, PermCommentDeleteOwn,
			PermSnapshotView, PermCMSView,
		}
		for _, p := range allowed {
			if !HasPermission(RoleViewer, p) {
				t.Errorf("viewer should have %q", p)
			}
		}

		denied := []Permission{
			PermProjectEdit, PermElementEdit, PermLogicEdit,
			PermPageCreate, PermPageEdit, PermPageDelete,
			PermCommentResolve, PermCommentEditAll, PermCommentDeleteAll,
			PermSnapshotCreate, PermSnapshotDelete, PermCMSEdit, PermCMSPublish,
			PermCollaboratorInvite, PermCollaboratorEdit, PermCollaboratorRemove,
		}
		for _, p := range denied {
			if HasPermission(RoleViewer, p) {
				t.Errorf("viewer must not have %q", p)
			}
		}
	})

	t.Run("assignable collaborator roles exclude owner", func(t *testing.T) {
		if IsAssignableCollaboratorRole(RoleOwner) {
			t.Error("owner must not be assignable as a collaborator role")
		}
		if !IsAssignableCollaboratorRole(RoleEditor) || !IsAssignableCollaboratorRole(RoleViewer) {
			t.Error("editor and viewer must be assignable")
		}
		if IsAssignableCollaboratorRole(Role("admin")) {
			t.Error("unknown role must not be assignable")
		}
	})
}

type stubLookup struct {
	calls int
	roles map[string]Role
	err   error
}

func (s *stubLookup) LookupProjectAccess(_ context.Context, projectID, userID string) (*SubjectAccess, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	role, ok := s.roles[projectID+":"+userID]
	if !ok {
		return nil, ErrNotFound
	}
	return &SubjectAccess{
		UserID:         userID,
		ProjectID:      projectID,
		Role:           role,
		IsOwner:        role == RoleOwner,
		IsCollaborator: role != RoleOwner,
		Permissions:    PermissionsForRole(role),
	}, nil
}

func TestAuthorizerMemoizationAndEnforcement(t *testing.T) {
	lookup := &stubLookup{
		roles: map[string]Role{
			"p1:u-owner":  RoleOwner,
			"p1:u-editor": RoleEditor,
			"p1:u-viewer": RoleViewer,
		},
	}
	authz := NewAuthorizer(lookup)
	ctx := WithRequestCache(context.Background())

	if _, err := authz.Authorize(ctx, "p1", "u-editor", PermPageDelete); err != nil {
		t.Fatalf("editor PageDelete failed: %v", err)
	}
	if _, err := authz.Authorize(ctx, "p1", "u-editor", PermProjectEdit); err != nil {
		t.Fatalf("editor ProjectEdit failed: %v", err)
	}
	if lookup.calls != 1 {
		t.Fatalf("expected 1 lookup with request cache, got %d", lookup.calls)
	}

	if _, err := authz.Authorize(ctx, "p1", "u-editor", PermProjectPublish); !errors.Is(err, ErrForbidden) {
		t.Fatalf("editor ProjectPublish = %v, want ErrForbidden", err)
	}
	if _, err := authz.Authorize(ctx, "p1", "u-viewer", PermCommentCreate); err != nil {
		t.Fatalf("viewer CommentCreate = %v", err)
	}
	if _, err := authz.Authorize(ctx, "p1", "u-viewer", PermCommentResolve); !errors.Is(err, ErrForbidden) {
		t.Fatalf("viewer CommentResolve = %v, want ErrForbidden", err)
	}
	if _, err := authz.Authorize(ctx, "p1", "u-stranger", PermProjectView); !errors.Is(err, ErrForbidden) {
		t.Fatalf("stranger ProjectView = %v, want ErrForbidden", err)
	}
	if _, err := authz.Authorize(ctx, "p1", "", PermProjectView); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("empty user = %v, want ErrUnauthorized", err)
	}
	if status, ok := HTTPStatus(ErrForbidden); !ok || status != http.StatusForbidden {
		t.Fatalf("HTTPStatus(ErrForbidden) = (%d, %v)", status, ok)
	}
}
