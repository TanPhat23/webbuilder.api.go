package services

import (
	"context"
	"errors"

	"my-go-app/internal/repositories"
)

// ensureProjectVisible checks that projectID can be read by userID: the user
// owns or collaborates on it (published or not), or it is published. It returns
// repositories.ErrProjectNotFound otherwise, so unpublished projects are not
// revealed to strangers. An empty userID only passes for published projects.
func ensureProjectVisible(ctx context.Context, projects repositories.ProjectRepositoryInterface, projectID, userID string) error {
	if userID != "" {
		project, err := projects.GetProjectWithAccess(ctx, projectID, userID)
		if err == nil && project != nil {
			return nil
		}
		if err != nil && !errors.Is(err, repositories.ErrProjectNotFound) && !errors.Is(err, repositories.ErrProjectUnauthorized) {
			return err
		}
	}
	project, err := projects.GetPublicProjectByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return repositories.ErrProjectNotFound
	}
	return nil
}
