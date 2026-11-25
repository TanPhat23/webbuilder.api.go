package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/services"
	mockrepo "my-go-app/test/mockrepo"
)

func assertPageServiceErr(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != want {
		t.Fatalf("expected error %q, got %v", want, err)
	}
}

func assertPageServiceNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func newTestPageService(pageRepo *mockrepo.MockPageRepo, projectRepo *mockrepo.MockProjectRepo) *services.PageService {
	return services.NewPageService(pageRepo, projectRepo)
}

func TestGetPagesByProjectIDValidAndMissing(t *testing.T) {
	pageRepo := mockrepo.NewMockPageRepo().SetListByProject(func(ctx context.Context, projectID string) ([]models.Page, error) {
		if projectID == "valid-project-id" {
			return []models.Page{{Id: "page1", ProjectId: "valid-project-id"}, {Id: "page2", ProjectId: "valid-project-id"}}, nil
		}
		return nil, nil
	})
	projectRepo := mockrepo.NewMockProjectRepo().SetGetPublicByID(func(ctx context.Context, id string) (*models.Project, error) {
		if id == "valid-project-id" {
			return &models.Project{ID: "valid-project-id"}, nil
		}
		return nil, nil
	})
	service := newTestPageService(pageRepo, projectRepo)

	pages, err := service.ListByProject(context.Background(), "valid-project-id", "")
	assertPageServiceNoErr(t, err)
	if len(pages) != 2 || pages[0].Id != "page1" || pages[1].Id != "page2" {
		t.Fatalf("unexpected pages returned: %+v", pages)
	}

	pages, err = service.ListByProject(context.Background(), "", "")
	assertPageServiceErr(t, err, "projectId is required")
	if pages != nil {
		t.Fatalf("expected nil pages, got %+v", pages)
	}
}

func TestGetPagesByProjectIDLookupError(t *testing.T) {
	projectRepo := mockrepo.NewMockProjectRepo().SetGetPublicByID(func(ctx context.Context, id string) (*models.Project, error) {
		return nil, errors.New("project lookup failed")
	})
	pages, err := newTestPageService(mockrepo.NewMockPageRepo(), projectRepo).ListByProject(context.Background(), "valid-project-id", "")
	assertPageServiceErr(t, err, "project lookup failed")
	if pages != nil {
		t.Fatalf("expected nil pages, got %+v", pages)
	}
}

func TestGetPageByID(t *testing.T) {
	pageRepo := mockrepo.NewMockPageRepo().SetGetByID(func(ctx context.Context, pageID, userID string) (*models.Page, error) {
		if pageID == "page1" {
			return &models.Page{Id: "page1", ProjectId: "valid-project-id"}, nil
		}
		return nil, nil
	})
	projectRepo := mockrepo.NewMockProjectRepo().SetGetPublicByID(func(ctx context.Context, id string) (*models.Project, error) {
		if id == "valid-project-id" {
			return &models.Project{ID: "valid-project-id"}, nil
		}
		return nil, nil
	})
	service := newTestPageService(pageRepo, projectRepo)

	page, err := service.GetByID(context.Background(), "page1")
	assertPageServiceNoErr(t, err)
	if page == nil || page.Id != "page1" {
		t.Fatalf("expected page1, got %v", page)
	}

	miss, err := service.GetByID(context.Background(), "")
	if err == nil || miss != nil {
		t.Fatalf("expected error and nil page, got %v, %v", err, miss)
	}
}

func newCreatePageTestService() (services.PageServiceInterface, *mockrepo.MockProjectRepo) {
	pageRepo := mockrepo.NewMockPageRepo().SetCreate(func(ctx context.Context, page *models.Page) error {
		if page.Id == "new-page" && page.ProjectId == "valid-project-id" {
			return nil
		}
		return errors.New("failed to create page")
	})
	projectRepo := mockrepo.NewMockProjectRepo().SetGetPublicByID(func(ctx context.Context, id string) (*models.Project, error) {
		if id == "valid-project-id" {
			return &models.Project{ID: "valid-project-id"}, nil
		}
		return nil, nil
	})
	return newTestPageService(pageRepo, projectRepo), projectRepo
}

func TestCreatePageSuccessAndValidation(t *testing.T) {
	service, _ := newCreatePageTestService()
	created, err := service.Create(context.Background(), &models.Page{Id: "new-page", ProjectId: "valid-project-id", Name: "New Page", Type: "html"})
	assertPageServiceNoErr(t, err)
	if created == nil || created.Id != "new-page" {
		t.Fatalf("expected new-page, got %v", created)
	}

	for _, tc := range []struct {
		page *models.Page
		want string
	}{
		{&models.Page{Id: "new-page", Name: "New Page", Type: "html"}, "projectId is required"},
		{&models.Page{Id: "new-page", ProjectId: "valid-project-id", Type: "html"}, "page name is required"},
		{&models.Page{Id: "new-page", ProjectId: "valid-project-id", Name: "New Page"}, "page type is required"},
	} {
		got, err := service.Create(context.Background(), tc.page)
		assertPageServiceErr(t, err, tc.want)
		if got != nil {
			t.Fatalf("expected nil page, got %v", got)
		}
	}
}

func TestCreatePageProjectLookupError(t *testing.T) {
	service, projectRepo := newCreatePageTestService()
	projectRepo.SetGetPublicByID(func(ctx context.Context, id string) (*models.Project, error) {
		return nil, errors.New("project lookup failed")
	})
	got, err := service.Create(context.Background(), &models.Page{Id: "new-page", ProjectId: "valid-project-id", Name: "New Page", Type: "html"})
	assertPageServiceErr(t, err, "project lookup failed")
	if got != nil {
		t.Fatalf("expected nil page, got %v", got)
	}
}

func TestUpdatePage(t *testing.T) {
	pageRepo := mockrepo.NewMockPageRepo().SetUpdate(func(ctx context.Context, page *models.Page) error {
		if page.Id == "page1" {
			return nil
		}
		return errors.New("failed to update page")
	})
	service := newTestPageService(pageRepo, mockrepo.NewMockProjectRepo())

	updated, err := service.Update(context.Background(), "page1", &models.Page{Id: "page1", Name: "Updated Page", Type: "html"})
	assertPageServiceNoErr(t, err)
	if updated == nil || updated.Id != "page1" {
		t.Fatalf("expected page1, got %v", updated)
	}

	for _, tc := range []struct {
		id   string
		page *models.Page
		want string
	}{
		{"", &models.Page{Name: "Updated Page", Type: "html"}, "pageId is required"},
		{"page1", nil, "page cannot be nil"},
		{"invalid", &models.Page{Id: "invalid", Name: "Invalid Page", Type: "html"}, "failed to update page"},
	} {
		got, err := service.Update(context.Background(), tc.id, tc.page)
		assertPageServiceErr(t, err, tc.want)
		if got != nil {
			t.Fatalf("expected nil page, got %v", got)
		}
	}
}

func TestUpdatePageFields(t *testing.T) {
	pageRepo := mockrepo.NewMockPageRepo().SetUpdateFields(func(ctx context.Context, pageID string, updates map[string]any) error {
		if pageID == "page1" && len(updates) > 0 {
			return nil
		}
		return errors.New("failed to update page fields")
	})
	service := newTestPageService(pageRepo, mockrepo.NewMockProjectRepo())

	assertPageServiceNoErr(t, service.UpdateFields(context.Background(), "page1", map[string]any{"Name": "Updated Name"}))
	assertPageServiceErr(t, service.UpdateFields(context.Background(), "", map[string]any{"Name": "Updated Name"}), "pageId is required")
	assertPageServiceErr(t, service.UpdateFields(context.Background(), "page1", map[string]any{}), "no updates provided")
	assertPageServiceErr(t, service.UpdateFields(context.Background(), "invalid", map[string]any{"Name": "Updated Name"}), "failed to update page fields")
}

func TestDeletePage(t *testing.T) {
	pageRepo := mockrepo.NewMockPageRepo().SetDelete(func(ctx context.Context, pageID string) error {
		if pageID == "page1" {
			return nil
		}
		return errors.New("failed to delete page")
	})
	service := newTestPageService(pageRepo, mockrepo.NewMockProjectRepo())

	assertPageServiceNoErr(t, service.Delete(context.Background(), "page1"))
	assertPageServiceErr(t, service.Delete(context.Background(), ""), "pageId is required")
	assertPageServiceErr(t, service.Delete(context.Background(), "invalid"), "failed to delete page")
}

func TestDeletePageByProjectID(t *testing.T) {
	pageRepo := mockrepo.NewMockPageRepo().SetDeleteForUser(func(ctx context.Context, pageID, projectID, userID string) error {
		if pageID == "page1" && projectID == "proj1" && userID == "user1" {
			return nil
		}
		return errors.New("failed to delete page")
	})
	projectRepo := mockrepo.NewMockProjectRepo().
		SetGetWithAccess(func(ctx context.Context, projectID, userID string) (*models.Project, error) {
			if projectID == "proj1" && userID == "user1" {
				return &models.Project{ID: "proj1"}, nil
			}
			return nil, errors.New("access denied")
		}).
		SetGetAccess(func(ctx context.Context, projectID, userID string) (*models.ProjectAccess, error) {
			if projectID == "proj1" && userID == "user1" {
				return &models.ProjectAccess{ProjectId: projectID, UserId: userID, Role: models.RoleEditor}, nil
			}
			return nil, errors.New("access denied")
		})
	service := newTestPageService(pageRepo, projectRepo)

	assertPageServiceNoErr(t, service.DeleteForUser(context.Background(), "page1", "proj1", "user1"))
	assertPageServiceErr(t, service.DeleteForUser(context.Background(), "", "proj1", "user1"), "pageId is required")
	assertPageServiceErr(t, service.DeleteForUser(context.Background(), "page1", "", "user1"), "projectId is required")
	assertPageServiceErr(t, service.DeleteForUser(context.Background(), "page1", "proj1", ""), "userId is required")
	assertPageServiceErr(t, service.DeleteForUser(context.Background(), "page1", "proj2", "user2"), "access denied")
}

func TestDeletePageByProjectIDWithoutVerification(t *testing.T) {
	pageRepo := mockrepo.NewMockPageRepo().SetDeleteForUser(func(ctx context.Context, pageID, projectID, userID string) error {
		if pageID == "page1" && projectID == "proj1" && userID == "" {
			return nil
		}
		return errors.New("failed to delete page")
	})
	service := newTestPageService(pageRepo, mockrepo.NewMockProjectRepo())

	assertPageServiceNoErr(t, service.DeleteInProject(context.Background(), "page1", "proj1"))
	assertPageServiceErr(t, service.DeleteInProject(context.Background(), "", "proj1"), "pageId is required")
	assertPageServiceErr(t, service.DeleteInProject(context.Background(), "page1", ""), "projectId is required")
	assertPageServiceErr(t, service.DeleteInProject(context.Background(), "invalid", "invalid"), "failed to delete page")
}
