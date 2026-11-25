package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type projectCapture struct {
	repositories.ProjectRepositoryInterface
	project *models.Project
	page    *models.Page
}

func (r *projectCapture) CreateWithPage(_ context.Context, p *models.Project, page *models.Page) error {
	r.project = p
	r.page = page
	return nil
}

func TestCreateProjectWithPageIDs(t *testing.T) {
	for _, supplied := range []string{"", uuid.NewString(), uuid.Must(uuid.NewV7()).String()} {
		t.Run(supplied, func(t *testing.T) {
			repo := &projectCapture{}
			service := NewProjectService(repo, nil, nil)
			result, err := service.CreateWithPage(context.Background(), &models.Project{ID: supplied, Name: "test", OwnerId: "owner"})
			if err != nil {
				t.Fatal(err)
			}
			if supplied != "" && result.ID != supplied {
				t.Fatalf("supplied ID changed: %q", result.ID)
			}
			for _, id := range []string{repo.page.Id, result.ID} {
				parsed, err := uuid.Parse(id)
				if err != nil {
					t.Fatal(err)
				}
				if id != supplied && parsed.Version() != 7 {
					t.Fatalf("expected UUIDv7, got %s", id)
				}
			}
		})
	}
}
