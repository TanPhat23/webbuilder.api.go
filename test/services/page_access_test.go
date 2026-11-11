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

func TestCreatePageAsUser(t *testing.T) {
	projectRepo := func(role models.CollaboratorRole, err error) *mockrepo.MockProjectRepo {
		return mockrepo.NewMockProjectRepo().
			SetGetPublicProjectByID(func(context.Context, string) (*models.Project, error) {
				return &models.Project{ID: "p1"}, nil
			}).
			SetGetProjectAccess(func(_ context.Context, projectID, userID string) (*models.ProjectAccess, error) {
				if err != nil {
					return nil, err
				}
				return &models.ProjectAccess{ProjectId: projectID, UserId: userID, Role: role}, nil
			})
	}
	newPage := func() *models.Page {
		return &models.Page{Id: "page-1", ProjectId: "p1", Name: "About", Type: "page"}
	}

	for _, tc := range []struct {
		name    string
		role    models.CollaboratorRole
		repoErr error
		wantErr error
	}{
		{"owner may create", models.RoleOwner, nil, nil},
		{"editor may create", models.RoleEditor, nil, nil},
		{"viewer is refused", models.RoleViewer, nil, repositories.ErrProjectUnauthorized},
		{"no access is refused", "", repositories.ErrProjectNotFound, repositories.ErrProjectUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			created := false
			pageRepo := mockrepo.NewMockPageRepo().SetCreatePage(func(context.Context, *models.Page) error {
				created = true
				return nil
			})
			svc := services.NewPageService(pageRepo, projectRepo(tc.role, tc.repoErr))

			_, err := svc.CreatePageAsUser(context.Background(), newPage(), "u1")
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if created != (tc.wantErr == nil) {
				t.Fatalf("page created = %v, want %v", created, tc.wantErr == nil)
			}
		})
	}
}

func TestCreateProjectWithPage(t *testing.T) {
	var gotProject *models.Project
	var gotPage *models.Page
	projectRepo := mockrepo.NewMockProjectRepo().
		SetCreateProjectWithPage(func(_ context.Context, p *models.Project, pg *models.Page) error {
			gotProject, gotPage = p, pg
			return nil
		})
	svc := services.NewProjectService(projectRepo, nil, nil)

	project, err := svc.CreateProjectWithPage(context.Background(), &models.Project{Name: "Site", OwnerId: "u1"})
	if err != nil {
		t.Fatal(err)
	}
	if project.ID == "" || gotProject != project {
		t.Fatalf("expected generated id on the stored project, got %+v", project)
	}
	if gotPage == nil || gotPage.Type != "page" || gotPage.Id == "" {
		t.Fatalf("expected a first page of type page, got %+v", gotPage)
	}

	if _, err := svc.CreateProjectWithPage(context.Background(), &models.Project{OwnerId: "u1"}); err == nil {
		t.Fatal("expected error for missing name")
	}
}
