package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	mockrepo "my-go-app/test/mockrepo"
)

// An unpublished project: the public lookup fails, access lookup depends on user.
func unpublishedProjectRepo(accessibleTo string) *mockrepo.MockProjectRepo {
	return mockrepo.NewMockProjectRepo().
		SetGetPublicProjectByID(func(context.Context, string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		}).
		SetGetProjectWithAccess(func(_ context.Context, projectID, userID string) (*models.Project, error) {
			if userID == accessibleTo {
				return &models.Project{ID: projectID}, nil
			}
			return nil, repositories.ErrProjectUnauthorized
		})
}

func TestGetCollaboratorsByProject_UnpublishedProject(t *testing.T) {
	collabs := mockrepo.NewMockCollaboratorRepo().
		SetGetCollaboratorsByProject(func(context.Context, string) ([]models.Collaborator, error) {
			return []models.Collaborator{{UserId: "u1"}}, nil
		})
	svc := services.NewCollaboratorService(collabs, unpublishedProjectRepo("owner"))

	got, err := svc.GetCollaboratorsByProject(context.Background(), "p1", "owner")
	if err != nil || len(got) != 1 {
		t.Fatalf("owner of an unpublished project: got %v, %v", got, err)
	}

	_, err = svc.GetCollaboratorsByProject(context.Background(), "p1", "stranger")
	if !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Fatalf("stranger: err = %v, want ErrProjectNotFound", err)
	}
}

func TestGetCollaboratorsByProject_PublishedProjectStaysVisible(t *testing.T) {
	projects := mockrepo.NewMockProjectRepo().
		SetGetPublicProjectByID(func(_ context.Context, id string) (*models.Project, error) {
			return &models.Project{ID: id, Published: true}, nil
		}).
		SetGetProjectWithAccess(func(context.Context, string, string) (*models.Project, error) {
			return nil, repositories.ErrProjectUnauthorized
		})
	collabs := mockrepo.NewMockCollaboratorRepo().
		SetGetCollaboratorsByProject(func(context.Context, string) ([]models.Collaborator, error) {
			return nil, nil
		})
	svc := services.NewCollaboratorService(collabs, projects)

	if _, err := svc.GetCollaboratorsByProject(context.Background(), "p1", "stranger"); err != nil {
		t.Fatalf("published project should stay readable: %v", err)
	}
}

func TestCreatePageAsUser_UnpublishedProject(t *testing.T) {
	projects := unpublishedProjectRepo("owner").
		SetGetProjectAccess(func(_ context.Context, projectID, userID string) (*models.ProjectAccess, error) {
			return &models.ProjectAccess{ProjectId: projectID, UserId: userID, Role: models.RoleOwner}, nil
		})
	created := false
	pages := mockrepo.NewMockPageRepo().SetCreatePage(func(context.Context, *models.Page) error {
		created = true
		return nil
	})
	svc := services.NewPageService(pages, projects)

	page := &models.Page{Id: "pg1", ProjectId: "p1", Name: "Home", Type: "page"}
	if _, err := svc.CreatePageAsUser(context.Background(), page, "owner"); err != nil || !created {
		t.Fatalf("owner adding a page to an unpublished project: err=%v created=%v", err, created)
	}
}
