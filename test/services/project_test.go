package services

import (
	"context"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	mockrepo "my-go-app/test/mockrepo"
	"testing"
)

func TestGetProjectByID(t *testing.T) {
	projectRepo := mockrepo.NewMockProjectRepo().SetGetProjectByID(func(ctx context.Context, id, userID string) (*models.Project, error) {
		if id == "valid-id" && userID == "user-id" {
			return &models.Project{ID: "valid-id"}, nil
		}
		return nil, nil
	})
	service := services.NewProjectService(projectRepo, nil, nil)

	t.Run("valid ID", func(t *testing.T) {
		project, err := service.GetProjectByID(context.Background(), "valid-id", "user-id")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if project == nil || project.ID != "valid-id" {
			t.Fatalf("expected project with ID 'valid-id', got %v", project)
		}
	})
}
