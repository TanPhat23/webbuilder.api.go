package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/services"
	test "my-go-app/test/mockrepo"
)

func TestGetPublicProjectByID_Success(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	project := &models.Project{ID: "proj123", Name: "Test Project", Published: true}
	projectRepo.SetGetPublicProjectByID(func(ctx context.Context, id string) (*models.Project, error) {
		return project, nil
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetPublicProjectByID(context.Background(), "proj123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil || result.ID != "proj123" {
		t.Errorf("expected project with ID proj123, got %v", result)
	}
}

func TestGetPublicProjectByID_EmptyID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetPublicProjectByID(context.Background(), "")
	if err == nil {
		t.Errorf("expected error for empty projectId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "projectId is required" {
		t.Errorf("expected 'projectId is required', got %v", err.Error())
	}
}

func TestGetPublicProjectByID_RepositoryError(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetPublicProjectByID(func(ctx context.Context, id string) (*models.Project, error) {
		return nil, errors.New("database error")
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetPublicProjectByID(context.Background(), "proj123")
	if err == nil {
		t.Errorf("expected database error")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetProjectByID_Success(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	project := &models.Project{ID: "proj123", Name: "Test Project", OwnerId: "user123"}
	projectRepo.SetGetProjectByID(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return project, nil
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectByID(context.Background(), "proj123", "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil || result.ID != "proj123" {
		t.Errorf("expected project with ID proj123, got %v", result)
	}
}

func TestGetProjectByID_EmptyProjectID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectByID(context.Background(), "", "user123")
	if err == nil {
		t.Errorf("expected error for empty projectId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "projectId is required" {
		t.Errorf("expected 'projectId is required', got %v", err.Error())
	}
}

func TestGetProjectByID_EmptyUserID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectByID(context.Background(), "proj123", "")
	if err == nil {
		t.Errorf("expected error for empty userId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "userId is required" {
		t.Errorf("expected 'userId is required', got %v", err.Error())
	}
}

func TestGetProjectByID_NotFound(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetProjectByID(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return nil, nil
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectByID(context.Background(), "proj123", "user123")
	if err == nil {
		t.Errorf("expected error for project not found")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "project does not exist" {
		t.Errorf("expected 'project does not exist', got %v", err.Error())
	}
}

func TestGetProjectByID_RepositoryError(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetProjectByID(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return nil, errors.New("database error")
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectByID(context.Background(), "proj123", "user123")
	if err == nil {
		t.Errorf("expected database error")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetProjectWithAccess_Success(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	project := &models.Project{ID: "proj123", Name: "Test Project", OwnerId: "user123"}
	projectRepo.SetGetProjectWithAccess(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return project, nil
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectWithAccess(context.Background(), "proj123", "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil || result.ID != "proj123" {
		t.Errorf("expected project, got %v", result)
	}
}

func TestGetProjectWithAccess_RepositoryError(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetProjectWithAccess(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return nil, errors.New("database error")
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectWithAccess(context.Background(), "proj123", "user123")
	if err == nil {
		t.Errorf("expected database error")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetProjectWithAccess_EmptyProjectID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectWithAccess(context.Background(), "", "user123")
	if err == nil {
		t.Errorf("expected error for empty projectId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetProjectWithAccess_EmptyUserID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectWithAccess(context.Background(), "proj123", "")
	if err == nil {
		t.Errorf("expected error for empty userId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetProjectWithAccess_NotFound(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetProjectWithAccess(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return nil, nil
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectWithAccess(context.Background(), "proj123", "user123")
	if err == nil {
		t.Errorf("expected error for project not found")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetProjectsByUserID_Success(t *testing.T) {
	projects := []models.Project{
		{ID: "proj1", Name: "Project 1", OwnerId: "user123"},
		{ID: "proj2", Name: "Project 2", OwnerId: "user123"},
	}
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetProjectsByUserID(func(ctx context.Context, userID string) ([]models.Project, error) {
		return projects, nil
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectsByUserID(context.Background(), "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 projects, got %d", len(result))
	}
}

func TestGetProjectsByUserID_EmptyUserID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectsByUserID(context.Background(), "")
	if err == nil {
		t.Errorf("expected error for empty userId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetProjectsByUserID_EmptyResult(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetProjectsByUserID(func(ctx context.Context, userID string) ([]models.Project, error) {
		return []models.Project{}, nil
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectsByUserID(context.Background(), "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 projects, got %d", len(result))
	}
}

func TestGetCollaboratorProjects_Success(t *testing.T) {
	projects := []models.Project{
		{ID: "proj1", Name: "Collab Project 1"},
	}
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetCollaboratorProjects(func(ctx context.Context, userID string) ([]models.Project, error) {
		return projects, nil
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetCollaboratorProjects(context.Background(), "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 project, got %d", len(result))
	}
}

func TestGetCollaboratorProjects_EmptyUserID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetCollaboratorProjects(context.Background(), "")
	if err == nil {
		t.Errorf("expected error for empty userId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetProjectPages_Success(t *testing.T) {
	pages := []models.Page{
		{Id: "page1", Name: "Home", ProjectId: "proj123"},
	}
	project := &models.Project{ID: "proj123", Name: "Test Project", Published: true}

	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetPublicProjectByID(func(ctx context.Context, id string) (*models.Project, error) {
		return project, nil
	})
	projectRepo.SetGetProjectPages(func(ctx context.Context, projID, userID string) ([]models.Page, error) {
		return pages, nil
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectPages(context.Background(), "proj123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 page, got %d", len(result))
	}
}

func TestGetProjectPages_RepositoryErrorOnVerify(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetPublicProjectByID(func(ctx context.Context, id string) (*models.Project, error) {
		return nil, errors.New("database error")
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectPages(context.Background(), "proj123")
	if err == nil {
		t.Errorf("expected error while verifying project")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetProjectPages_RepositoryErrorOnPagesFetch(t *testing.T) {
	project := &models.Project{ID: "proj123", Name: "Test Project", Published: true}
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetPublicProjectByID(func(ctx context.Context, id string) (*models.Project, error) {
		return project, nil
	})
	projectRepo.SetGetProjectPages(func(ctx context.Context, projID, userID string) ([]models.Page, error) {
		return nil, errors.New("database error")
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectPages(context.Background(), "proj123")
	if err == nil {
		t.Errorf("expected error while fetching pages")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetProjectPages_EmptyProjectID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectPages(context.Background(), "")
	if err == nil {
		t.Errorf("expected error for empty projectId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetProjectPages_ProjectNotFound(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetPublicProjectByID(func(ctx context.Context, id string) (*models.Project, error) {
		return nil, nil
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectPages(context.Background(), "proj123")
	if err == nil {
		t.Errorf("expected error for project not found")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestCreateProject_Success(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetCreateProject(func(ctx context.Context, proj *models.Project) error {
		return nil
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	project := &models.Project{
		Name:    "New Project",
		OwnerId: "user123",
	}

	result, err := service.CreateProject(context.Background(), project)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Errorf("expected project, got nil")
	}
}

func TestCreateProject_NilProject(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.CreateProject(context.Background(), nil)
	if err == nil {
		t.Errorf("expected error for nil project")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "project cannot be nil" {
		t.Errorf("expected 'project cannot be nil', got %v", err.Error())
	}
}

func TestCreateProject_MissingName(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	project := &models.Project{
		OwnerId: "user123",
	}

	result, err := service.CreateProject(context.Background(), project)
	if err == nil {
		t.Errorf("expected error for missing name")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "project name is required" {
		t.Errorf("expected 'project name is required', got %v", err.Error())
	}
}

func TestCreateProject_MissingOwnerId(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	project := &models.Project{
		Name: "New Project",
	}

	result, err := service.CreateProject(context.Background(), project)
	if err == nil {
		t.Errorf("expected error for missing ownerId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "ownerId is required" {
		t.Errorf("expected 'ownerId is required', got %v", err.Error())
	}
}

func TestCreateProject_RepositoryError(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetCreateProject(func(ctx context.Context, proj *models.Project) error {
		return errors.New("database error")
	})
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	project := &models.Project{
		Name:    "New Project",
		OwnerId: "user123",
	}

	result, err := service.CreateProject(context.Background(), project)
	if err == nil {
		t.Errorf("expected database error")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestUpdateProject_Success(t *testing.T) {
	existingDescription := "Original Description"
	existingSubdomain := "original"
	existingStyles := json.RawMessage("{original}")
	existingHeader := json.RawMessage("{\"logo\":\"old\"}")
	existingProject := &models.Project{
		ID:          "proj123",
		Name:        "Original Project",
		Description: &existingDescription,
		Subdomain:   &existingSubdomain,
		Styles:      &existingStyles,
		Header:      &existingHeader,
		OwnerId:     "user123",
	}

	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetProjectWithAccess(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return existingProject, nil
	})
	projectRepo.SetUpdateProject(func(ctx context.Context, projID, userID string, updates map[string]any) (*models.Project, error) {
		return &models.Project{
			ID:          projID,
			Name:        updates["Name"].(string),
			Description: updates["Description"].(*string),
			Subdomain:   updates["Subdomain"].(*string),
			Styles:      updates["Styles"].(*json.RawMessage),
			Header:      updates["Header"].(*json.RawMessage),
			OwnerId:     userID,
		}, nil
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	updatedProject := &models.Project{
		Name: "Updated Project",
	}

	result, err := service.UpdateProject(context.Background(), "proj123", "user123", updatedProject)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Errorf("expected project, got nil")
	}
	if result.Name != "Updated Project" {
		t.Errorf("expected updated name, got %v", result.Name)
	}
}

func TestUpdateProject_RepositoryError(t *testing.T) {
	existingProject := &models.Project{
		ID:      "proj123",
		Name:    "Original Project",
		OwnerId: "user123",
	}

	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetProjectWithAccess(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return existingProject, nil
	})
	projectRepo.SetUpdateProject(func(ctx context.Context, projID, userID string, updates map[string]any) (*models.Project, error) {
		return nil, errors.New("database error")
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.UpdateProject(context.Background(), "proj123", "user123", &models.Project{Name: "Updated"})
	if err == nil {
		t.Errorf("expected database error")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestUpdateProject_UsesExistingFieldsWhenInputEmptyOrNil(t *testing.T) {
	existingDescription := "Original Description"
	existingSubdomain := "original"
	existingStyles := json.RawMessage("{original}")
	existingHeader := json.RawMessage("{\"logo\":\"old\"}")
	existingProject := &models.Project{
		ID:          "proj123",
		Name:        "Original Project",
		Description: &existingDescription,
		Subdomain:   &existingSubdomain,
		Styles:      &existingStyles,
		Header:      &existingHeader,
		OwnerId:     "user123",
	}

	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetProjectWithAccess(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return existingProject, nil
	})
	projectRepo.SetUpdateProject(func(ctx context.Context, projID, userID string, updates map[string]any) (*models.Project, error) {
		if updates["Name"].(string) != existingProject.Name {
			t.Errorf("expected name fallback, got %v", updates["Name"])
		}
		if updates["Description"].(*string) == nil || *updates["Description"].(*string) != *existingProject.Description {
			t.Errorf("expected description fallback, got %v", updates["Description"])
		}
		if updates["Subdomain"].(*string) == nil || *updates["Subdomain"].(*string) != *existingProject.Subdomain {
			t.Errorf("expected subdomain fallback, got %v", updates["Subdomain"])
		}
		if updates["Styles"].(*json.RawMessage) == nil || string(*updates["Styles"].(*json.RawMessage)) != string(*existingProject.Styles) {
			t.Errorf("expected styles fallback, got %v", updates["Styles"])
		}
		if updates["Header"].(*json.RawMessage) == nil || string(*updates["Header"].(*json.RawMessage)) != string(*existingProject.Header) {
			t.Errorf("expected header fallback, got %v", updates["Header"])
		}
		return existingProject, nil
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	_, err := service.UpdateProject(context.Background(), "proj123", "user123", &models.Project{})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestUpdateProject_VerificationError(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetProjectWithAccess(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return nil, errors.New("database error")
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.UpdateProject(context.Background(), "proj123", "user123", &models.Project{Name: "Updated"})
	if err == nil {
		t.Errorf("expected database error")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestUpdateProject_EmptyProjectID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.UpdateProject(context.Background(), "", "user123", &models.Project{Name: "Updated"})
	if err == nil {
		t.Errorf("expected error for empty projectId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestUpdateProject_EmptyUserID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.UpdateProject(context.Background(), "proj123", "", &models.Project{Name: "Updated"})
	if err == nil {
		t.Errorf("expected error for empty userId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestUpdateProject_NilProject(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.UpdateProject(context.Background(), "proj123", "user123", nil)
	if err == nil {
		t.Errorf("expected error for nil project")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestDeleteProject_Success(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	project := &models.Project{ID: "proj123", Name: "Test", OwnerId: "user123"}
	projectRepo.SetGetProjectWithAccess(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return project, nil
	})
	projectRepo.SetDeleteProject(func(ctx context.Context, projID, userID string) error {
		return nil
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	err := service.DeleteProject(context.Background(), "proj123", "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestDeleteProject_RepositoryError(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	project := &models.Project{ID: "proj123", Name: "Test", OwnerId: "user123"}
	projectRepo.SetGetProjectWithAccess(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return project, nil
	})
	projectRepo.SetDeleteProject(func(ctx context.Context, projID, userID string) error {
		return errors.New("database error")
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	err := service.DeleteProject(context.Background(), "proj123", "user123")
	if err == nil {
		t.Errorf("expected database error")
	}
}

func TestDeleteProject_EmptyProjectID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	err := service.DeleteProject(context.Background(), "", "user123")
	if err == nil {
		t.Errorf("expected error for empty projectId")
	}
}

func TestDeleteProject_EmptyUserID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	err := service.DeleteProject(context.Background(), "proj123", "")
	if err == nil {
		t.Errorf("expected error for empty userId")
	}
}

func TestHardDeleteProject_Success(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	project := &models.Project{ID: "proj123", Name: "Test", OwnerId: "user123"}
	projectRepo.SetGetProjectWithAccess(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return project, nil
	})
	projectRepo.SetHardDeleteProject(func(ctx context.Context, projID, userID string) error {
		return nil
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	err := service.HardDeleteProject(context.Background(), "proj123", "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestHardDeleteProject_RepositoryError(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	project := &models.Project{ID: "proj123", Name: "Test", OwnerId: "user123"}
	projectRepo.SetGetProjectWithAccess(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return project, nil
	})
	projectRepo.SetHardDeleteProject(func(ctx context.Context, projID, userID string) error {
		return errors.New("database error")
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	err := service.HardDeleteProject(context.Background(), "proj123", "user123")
	if err == nil {
		t.Errorf("expected database error")
	}
}

func TestRestoreProject_Success(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	project := &models.Project{ID: "proj123", Name: "Test", OwnerId: "user123"}
	projectRepo.SetGetProjectWithAccess(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return project, nil
	})
	projectRepo.SetRestoreProject(func(ctx context.Context, projID, userID string) error {
		return nil
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	err := service.RestoreProject(context.Background(), "proj123", "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestRestoreProject_RepositoryError(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	project := &models.Project{ID: "proj123", Name: "Test", OwnerId: "user123"}
	projectRepo.SetGetProjectWithAccess(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return project, nil
	})
	projectRepo.SetRestoreProject(func(ctx context.Context, projID, userID string) error {
		return errors.New("database error")
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	err := service.RestoreProject(context.Background(), "proj123", "user123")
	if err == nil {
		t.Errorf("expected database error")
	}
}

func TestExistsForUser_True(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetExistsForUser(func(ctx context.Context, projID, userID string) (bool, error) {
		return true, nil
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	exists, err := service.ExistsForUser(context.Background(), "proj123", "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !exists {
		t.Errorf("expected project to exist")
	}
}

func TestExistsForUser_RepositoryError(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetExistsForUser(func(ctx context.Context, projID, userID string) (bool, error) {
		return false, errors.New("database error")
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	exists, err := service.ExistsForUser(context.Background(), "proj123", "user123")
	if err == nil {
		t.Errorf("expected database error")
	}
	if exists {
		t.Errorf("expected false for existence check")
	}
}

func TestExistsForUser_False(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetExistsForUser(func(ctx context.Context, projID, userID string) (bool, error) {
		return false, nil
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	exists, err := service.ExistsForUser(context.Background(), "proj123", "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if exists {
		t.Errorf("expected project to not exist")
	}
}

func TestExistsForUser_EmptyProjectID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	exists, err := service.ExistsForUser(context.Background(), "", "user123")
	if err == nil {
		t.Errorf("expected error for empty projectId")
	}
	if exists {
		t.Errorf("expected false for existence check")
	}
}

func TestExistsForUser_EmptyUserID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	exists, err := service.ExistsForUser(context.Background(), "proj123", "")
	if err == nil {
		t.Errorf("expected error for empty userId")
	}
	if exists {
		t.Errorf("expected false for existence check")
	}
}

func TestGetProjectWithLock_Success(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	project := &models.Project{ID: "proj123", Name: "Test", OwnerId: "user123"}
	projectRepo.SetGetProjectWithLock(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return project, nil
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectWithLock(context.Background(), "proj123", "user123")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Errorf("expected project, got nil")
	}
}

func TestGetProjectWithLock_RepositoryError(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	projectRepo.SetGetProjectWithLock(func(ctx context.Context, projID, userID string) (*models.Project, error) {
		return nil, errors.New("database error")
	})

	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectWithLock(context.Background(), "proj123", "user123")
	if err == nil {
		t.Errorf("expected database error")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetProjectWithLock_EmptyProjectID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectWithLock(context.Background(), "", "user123")
	if err == nil {
		t.Errorf("expected error for empty projectId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestGetProjectWithLock_EmptyUserID(t *testing.T) {
	projectRepo := test.NewMockProjectRepo()
	collaboratorRepo := test.NewMockCollaboratorRepo()
	userRepo := test.NewMockUserRepo()

	service := services.NewProjectService(projectRepo, collaboratorRepo, userRepo)

	result, err := service.GetProjectWithLock(context.Background(), "proj123", "")
	if err == nil {
		t.Errorf("expected error for empty userId")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}