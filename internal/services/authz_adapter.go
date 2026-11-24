package services

import (
	"context"
	"errors"

	"my-go-app/internal/authz"
	"my-go-app/internal/repositories"
)

type projectAccessLookup struct {
	projects repositories.ProjectRepositoryInterface
}

// NewProjectAuthorizer creates an authz.Authorizer backed by a ProjectRepositoryInterface.
func NewProjectAuthorizer(projects repositories.ProjectRepositoryInterface) *authz.Authorizer {
	return authz.NewAuthorizer(&projectAccessLookup{projects: projects})
}

func (l *projectAccessLookup) LookupProjectAccess(ctx context.Context, projectID, userID string) (*authz.SubjectAccess, error) {
	if l.projects == nil {
		return nil, errors.New("project repository is not configured")
	}
	access, err := l.projects.GetProjectAccess(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrProjectNotFound) {
			return nil, authz.ErrNotFound
		}
		if errors.Is(err, repositories.ErrProjectUnauthorized) {
			return nil, authz.ErrForbidden
		}
		return nil, err
	}
	if access == nil {
		return nil, authz.ErrNotFound
	}
	role := authz.Role(access.Role)
	perms := access.Permissions
	if len(perms) == 0 {
		perms = authz.PermissionsForRole(role)
	}
	return &authz.SubjectAccess{
		UserID:         userID,
		ProjectID:      projectID,
		Role:           role,
		IsOwner:        access.IsOwner,
		IsCollaborator: access.IsCollaborator,
		Permissions:    perms,
	}, nil
}

func authorizeProjectPerm(ctx context.Context, projects repositories.ProjectRepositoryInterface, projectID, userID string, perm authz.Permission) (*authz.SubjectAccess, error) {
	return NewProjectAuthorizer(projects).Authorize(ctx, projectID, userID, perm)
}
