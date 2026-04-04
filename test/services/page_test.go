package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/services"
	mockrepo "my-go-app/test/mockrepo"
)

func TestGetPagesByProjectID(t *testing.T) {
	t.Run("valid project ID returns pages", func(t *testing.T) {
		pageRepo := mockrepo.NewMockPageRepo().SetGetPagesByProjectID(func(ctx context.Context, projectID string) ([]models.Page, error) {
			if projectID == "valid-project-id" {
				return []models.Page{
					{Id: "page1", ProjectId: "valid-project-id"},
					{Id: "page2", ProjectId: "valid-project-id"},
				}, nil
			}
			return nil, nil
		})
		projectRepo := mockrepo.NewMockProjectRepo().SetGetPublicProjectByID(func(ctx context.Context, id string) (*models.Project, error) {
			if id == "valid-project-id" {
				return &models.Project{ID: "valid-project-id"}, nil
			}
			return nil, nil
		})

		service := services.NewPageService(pageRepo, projectRepo)

		pages, err := service.GetPagesByProjectID(context.Background(), "valid-project-id")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(pages) != 2 {
			t.Fatalf("expected 2 pages, got %d", len(pages))
		}
		if pages[0].Id != "page1" || pages[1].Id != "page2" {
			t.Fatalf("unexpected pages returned: %+v", pages)
		}
	})

	t.Run("missing project ID returns error", func(t *testing.T) {
		service := services.NewPageService(mockrepo.NewMockPageRepo(), mockrepo.NewMockProjectRepo())

		pages, err := service.GetPagesByProjectID(context.Background(), "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "projectId is required" {
			t.Fatalf("expected projectId error, got %v", err)
		}
		if pages != nil {
			t.Fatalf("expected nil pages, got %+v", pages)
		}
	})

	t.Run("project lookup error is returned", func(t *testing.T) {
		pageRepo := mockrepo.NewMockPageRepo()
		projectRepo := mockrepo.NewMockProjectRepo().SetGetPublicProjectByID(func(ctx context.Context, id string) (*models.Project, error) {
			return nil, errors.New("project lookup failed")
		})

		service := services.NewPageService(pageRepo, projectRepo)

		pages, err := service.GetPagesByProjectID(context.Background(), "valid-project-id")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "project lookup failed" {
			t.Fatalf("expected propagated error, got %v", err)
		}
		if pages != nil {
			t.Fatalf("expected nil pages, got %+v", pages)
		}
	})
}

func TestGetPageByID(t *testing.T) {
	pageRepo := mockrepo.NewMockPageRepo().SetGetPageByID(func(ctx context.Context, pageID, userID string) (*models.Page, error) {
		if pageID == "page1" {
			return &models.Page{Id: "page1", ProjectId: "valid-project-id"}, nil
		}
		return nil, nil
	})
	projectRepo := mockrepo.NewMockProjectRepo().SetGetPublicProjectByID(func(ctx context.Context, id string) (*models.Project, error) {
		if id == "valid-project-id" {
			return &models.Project{ID: "valid-project-id"}, nil
		}
		return nil, nil
	})

	service := services.NewPageService(pageRepo, projectRepo)

	t.Run("valid project ID returns pages", func(t *testing.T) {
		page, err := service.GetPageByID(context.Background(), "page1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if page == nil || page.Id != "page1" {
			t.Fatalf("expected page with ID 'page1', got %v", page)
		}
	})
	t.Run("missing page ID returns error", func(t *testing.T) {
		page, err := service.GetPageByID(context.Background(), "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if page != nil {
			t.Fatalf("expected nil page, got %v", page)
		}
	})
}
func TestCreatePage(t *testing.T) {
	pageRepo := mockrepo.NewMockPageRepo().SetCreatePage(func(ctx context.Context, page *models.Page) error {
		if page.Id == "new-page" && page.ProjectId == "valid-project-id" {
			return nil
		}
		return errors.New("failed to create page")
	})
	projectRepo := mockrepo.NewMockProjectRepo().SetGetPublicProjectByID(func(ctx context.Context, id string) (*models.Project, error) {
		if id == "valid-project-id" {
			return &models.Project{ID: "valid-project-id"}, nil
		}
		return nil, nil
	})

	service := services.NewPageService(pageRepo, projectRepo)

	t.Run("valid page is created successfully", func(t *testing.T) {
		page := &models.Page{Id: "new-page", ProjectId: "valid-project-id", Name: "New Page", Type: "html"}
		createdPage, err := service.CreatePage(context.Background(), page)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if createdPage == nil || createdPage.Id != "new-page" {
			t.Fatalf("expected created page with ID 'new-page', got %v", createdPage)
		}
	})

	t.Run("missing project ID returns error", func(t *testing.T) {
		page := &models.Page{Id: "new-page", Name: "New Page", Type: "html"}
		createdPage, err := service.CreatePage(context.Background(), page)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "projectId is required" {
			t.Fatalf("expected projectId error, got %v", err)
		}
		if createdPage != nil {
			t.Fatalf("expected nil page, got %v", createdPage)
		}
	})
	t.Run("missing page name returns error", func(t *testing.T) {
		page := &models.Page{Id: "new-page", ProjectId: "valid-project-id", Type: "html"}
		createdPage, err := service.CreatePage(context.Background(), page)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "page name is required" {
			t.Fatalf("expected page name error, got %v", err)
		}
		if createdPage != nil {
			t.Fatalf("expected nil page, got %v", createdPage)
		}
	})
	
	t.Run("missing page type returns error", func(t *testing.T) {
		page := &models.Page{Id: "new-page", ProjectId: "valid-project-id", Name: "New Page"}
		createdPage, err := service.CreatePage(context.Background(), page)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "page type is required" {
			t.Fatalf("expected page type error, got %v", err)
		}
		if createdPage != nil {
			t.Fatalf("expected nil page, got %v", createdPage)
		}
	})

	t.Run("project lookup error is returned", func(t *testing.T) {
		projectRepo.SetGetPublicProjectByID(func(ctx context.Context, id string) (*models.Project, error) {
			return nil, errors.New("project lookup failed")
		})
		page := &models.Page{Id: "new-page", ProjectId: "valid-project-id", Name: "New Page", Type: "html"}
		createdPage, err := service.CreatePage(context.Background(), page)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "project lookup failed" {
			t.Fatalf("expected propagated error, got %v", err)
		}
		if createdPage != nil {
			t.Fatalf("expected nil page, got %v", createdPage)
		}
	})
}

func TestUpdatePage(t *testing.T) {
	pageRepo := mockrepo.NewMockPageRepo().SetUpdatePage(func(ctx context.Context, page *models.Page) error {
		if page.Id == "page1" {
			return nil
		}
		return errors.New("failed to update page")
	})
	projectRepo := mockrepo.NewMockProjectRepo()

	service := services.NewPageService(pageRepo, projectRepo)

	t.Run("valid page is updated successfully", func(t *testing.T) {
		page := &models.Page{Id: "page1", Name: "Updated Page", Type: "html"}
		updatedPage, err := service.UpdatePage(context.Background(), "page1", page)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updatedPage == nil || updatedPage.Id != "page1" {
			t.Fatalf("expected updated page with ID 'page1', got %v", updatedPage)
		}
	})

	t.Run("missing page ID returns error", func(t *testing.T) {
		page := &models.Page{Name: "Updated Page", Type: "html"}
		updatedPage, err := service.UpdatePage(context.Background(), "", page)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "pageId is required" {
			t.Fatalf("expected pageId error, got %v", err)
		}
		if updatedPage != nil {
			t.Fatalf("expected nil page, got %v", updatedPage)
		}
	})

	t.Run("nil page returns error", func(t *testing.T) {
		updatedPage, err := service.UpdatePage(context.Background(), "page1", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "page cannot be nil" {
			t.Fatalf("expected page nil error, got %v", err)
		}
		if updatedPage != nil {
			t.Fatalf("expected nil page, got %v", updatedPage)
		}
	})

	t.Run("repository error is returned", func(t *testing.T) {
		page := &models.Page{Id: "invalid", Name: "Invalid Page", Type: "html"}
		updatedPage, err := service.UpdatePage(context.Background(), "invalid", page)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "failed to update page" {
			t.Fatalf("expected repo error, got %v", err)
		}
		if updatedPage != nil {
			t.Fatalf("expected nil page, got %v", updatedPage)
		}
	})
}

func TestUpdatePageFields(t *testing.T) {
	pageRepo := mockrepo.NewMockPageRepo().SetUpdatePageFields(func(ctx context.Context, pageID string, updates map[string]any) error {
		if pageID == "page1" && len(updates) > 0 {
			return nil
		}
		return errors.New("failed to update page fields")
	})
	projectRepo := mockrepo.NewMockProjectRepo()

	service := services.NewPageService(pageRepo, projectRepo)

	t.Run("valid page fields are updated successfully", func(t *testing.T) {
		updates := map[string]any{"Name": "Updated Name"}
		err := service.UpdatePageFields(context.Background(), "page1", updates)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("missing page ID returns error", func(t *testing.T) {
		updates := map[string]any{"Name": "Updated Name"}
		err := service.UpdatePageFields(context.Background(), "", updates)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "pageId is required" {
			t.Fatalf("expected pageId error, got %v", err)
		}
	})

	t.Run("no updates provided returns error", func(t *testing.T) {
		err := service.UpdatePageFields(context.Background(), "page1", map[string]any{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "no updates provided" {
			t.Fatalf("expected no updates error, got %v", err)
		}
	})

	t.Run("repository error is returned", func(t *testing.T) {
		updates := map[string]any{"Name": "Updated Name"}
		err := service.UpdatePageFields(context.Background(), "invalid", updates)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "failed to update page fields" {
			t.Fatalf("expected repo error, got %v", err)
		}
	})
}

func TestDeletePage(t *testing.T) {
	pageRepo := mockrepo.NewMockPageRepo().SetDeletePage(func(ctx context.Context, pageID string) error {
		if pageID == "page1" {
			return nil
		}
		return errors.New("failed to delete page")
	})
	projectRepo := mockrepo.NewMockProjectRepo()

	service := services.NewPageService(pageRepo, projectRepo)

	t.Run("valid page is deleted successfully", func(t *testing.T) {
		err := service.DeletePage(context.Background(), "page1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("missing page ID returns error", func(t *testing.T) {
		err := service.DeletePage(context.Background(), "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "pageId is required" {
			t.Fatalf("expected pageId error, got %v", err)
		}
	})

	t.Run("repository error is returned", func(t *testing.T) {
		err := service.DeletePage(context.Background(), "invalid")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "failed to delete page" {
			t.Fatalf("expected repo error, got %v", err)
		}
	})
}

func TestDeletePageByProjectID(t *testing.T) {
	pageRepo := mockrepo.NewMockPageRepo().SetDeletePageByProjectID(func(ctx context.Context, pageID, projectID, userID string) error {
		if pageID == "page1" && projectID == "proj1" && userID == "user1" {
			return nil
		}
		return errors.New("failed to delete page")
	})
	projectRepo := mockrepo.NewMockProjectRepo().SetGetProjectWithAccess(func(ctx context.Context, projectID, userID string) (*models.Project, error) {
		if projectID == "proj1" && userID == "user1" {
			return &models.Project{ID: "proj1"}, nil
		}
		return nil, errors.New("access denied")
	})

	service := services.NewPageService(pageRepo, projectRepo)

	t.Run("valid page is deleted by project ID", func(t *testing.T) {
		err := service.DeletePageByProjectID(context.Background(), "page1", "proj1", "user1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("missing page ID returns error", func(t *testing.T) {
		err := service.DeletePageByProjectID(context.Background(), "", "proj1", "user1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "pageId is required" {
			t.Fatalf("expected pageId error, got %v", err)
		}
	})

	t.Run("missing project ID returns error", func(t *testing.T) {
		err := service.DeletePageByProjectID(context.Background(), "page1", "", "user1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "projectId is required" {
			t.Fatalf("expected projectId error, got %v", err)
		}
	})

	t.Run("missing user ID returns error", func(t *testing.T) {
		err := service.DeletePageByProjectID(context.Background(), "page1", "proj1", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "userId is required" {
			t.Fatalf("expected userId error, got %v", err)
		}
	})

	t.Run("access denied error is returned", func(t *testing.T) {
		err := service.DeletePageByProjectID(context.Background(), "page1", "proj2", "user2")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "access denied" {
			t.Fatalf("expected access denied error, got %v", err)
		}
	})
}

func TestDeletePageByProjectIDWithoutVerification(t *testing.T) {
	pageRepo := mockrepo.NewMockPageRepo().SetDeletePageByProjectID(func(ctx context.Context, pageID, projectID, userID string) error {
		if pageID == "page1" && projectID == "proj1" && userID == "" {
			return nil
		}
		return errors.New("failed to delete page")
	})
	projectRepo := mockrepo.NewMockProjectRepo()

	service := services.NewPageService(pageRepo, projectRepo)

	t.Run("valid page is deleted without verification", func(t *testing.T) {
		err := service.DeletePageByProjectIDWithoutVerification(context.Background(), "page1", "proj1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("missing page ID returns error", func(t *testing.T) {
		err := service.DeletePageByProjectIDWithoutVerification(context.Background(), "", "proj1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "pageId is required" {
			t.Fatalf("expected pageId error, got %v", err)
		}
	})

	t.Run("missing project ID returns error", func(t *testing.T) {
		err := service.DeletePageByProjectIDWithoutVerification(context.Background(), "page1", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "projectId is required" {
			t.Fatalf("expected projectId error, got %v", err)
		}
	})

	t.Run("repository error is returned", func(t *testing.T) {
		err := service.DeletePageByProjectIDWithoutVerification(context.Background(), "invalid", "invalid")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "failed to delete page" {
			t.Fatalf("expected repo error, got %v", err)
		}
	})
}
