package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/authz"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	mockrepo "my-go-app/test/mockrepo"
)

// An unpublished project: the public lookup fails, access lookup depends on user.
func unpublishedProjectRepo(accessibleTo string) *mockrepo.MockProjectRepo {
	return mockrepo.NewMockProjectRepo().
		SetGetPublicByID(func(context.Context, string) (*models.Project, error) {
			return nil, repositories.ErrProjectNotFound
		}).
		SetGetWithAccess(func(_ context.Context, projectID, userID string) (*models.Project, error) {
			if userID == accessibleTo {
				return &models.Project{ID: projectID}, nil
			}
			return nil, repositories.ErrProjectUnauthorized
		}).
		SetGetAccess(func(_ context.Context, projectID, userID string) (*models.ProjectAccess, error) {
			if userID == accessibleTo {
				return &models.ProjectAccess{ProjectId: projectID, UserId: userID, Role: models.RoleOwner, IsOwner: true}, nil
			}
			return nil, repositories.ErrProjectNotFound
		})
}

func TestGetCollaboratorsByProject_UnpublishedProject(t *testing.T) {
	collabs := mockrepo.NewMockCollaboratorRepo().
		SetListByProject(func(context.Context, string) ([]models.Collaborator, error) {
			return []models.Collaborator{{UserId: "u1"}}, nil
		})
	svc := services.NewCollaboratorService(collabs, unpublishedProjectRepo("owner"))

	got, err := svc.ListByProject(context.Background(), "p1", "owner")
	if err != nil || len(got) != 1 {
		t.Fatalf("owner of an unpublished project: got %v, %v", got, err)
	}

	_, err = svc.ListByProject(context.Background(), "p1", "stranger")
	if !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("stranger: err = %v, want authz.ErrForbidden", err)
	}
}

func TestGetCollaboratorsByProject_PublishedProjectRejectsStrangers(t *testing.T) {
	projects := mockrepo.NewMockProjectRepo().
		SetGetPublicByID(func(_ context.Context, id string) (*models.Project, error) {
			return &models.Project{ID: id, Published: true}, nil
		}).
		SetGetAccess(func(context.Context, string, string) (*models.ProjectAccess, error) {
			return nil, repositories.ErrProjectNotFound
		})
	collabs := mockrepo.NewMockCollaboratorRepo().
		SetListByProject(func(context.Context, string) ([]models.Collaborator, error) {
			return nil, nil
		})
	svc := services.NewCollaboratorService(collabs, projects)

	if _, err := svc.ListByProject(context.Background(), "p1", "stranger"); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("non-member should not see collaborator list: %v", err)
	}
}

func TestCreatePageAsUser_UnpublishedProject(t *testing.T) {
	projects := unpublishedProjectRepo("owner").
		SetGetAccess(func(_ context.Context, projectID, userID string) (*models.ProjectAccess, error) {
			return &models.ProjectAccess{ProjectId: projectID, UserId: userID, Role: models.RoleOwner}, nil
		})
	created := false
	pages := mockrepo.NewMockPageRepo().SetCreate(func(context.Context, *models.Page) error {
		created = true
		return nil
	})
	svc := services.NewPageService(pages, projects)

	page := &models.Page{Id: "pg1", ProjectId: "p1", Name: "Home", Type: "page"}
	if _, err := svc.CreateForUser(context.Background(), page, "owner"); err != nil || !created {
		t.Fatalf("owner adding a page to an unpublished project: err=%v created=%v", err, created)
	}
}

func TestGetPagesByProjectID_Visibility(t *testing.T) {
	pages := mockrepo.NewMockPageRepo().
		SetListByProject(func(context.Context, string) ([]models.Page, error) {
			return []models.Page{{Id: "pg1"}}, nil
		})
	svc := services.NewPageService(pages, unpublishedProjectRepo("owner"))

	if got, err := svc.ListByProject(context.Background(), "p1", "owner"); err != nil || len(got) != 1 {
		t.Fatalf("owner: got %v, %v", got, err)
	}
	// "" is the public route: an unpublished project must stay hidden.
	if _, err := svc.ListByProject(context.Background(), "p1", ""); !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Fatalf("public route: err = %v, want ErrProjectNotFound", err)
	}
	if _, err := svc.ListByProject(context.Background(), "p1", "stranger"); !errors.Is(err, repositories.ErrProjectNotFound) {
		t.Fatalf("stranger: err = %v, want ErrProjectNotFound", err)
	}
}
