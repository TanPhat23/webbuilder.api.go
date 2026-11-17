package services

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/handlers"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	mockrepo "my-go-app/test/mockrepo"
)

// The public page endpoint serves a page only when it belongs to the project in
// the URL and that project is published, like /projects/public and /pages/public.
func publicPageFixture() (*services.PageService, *[]string) {
	var pageQueries []string
	pageRepo := mockrepo.NewMockPageRepo().SetGetPageByID(func(_ context.Context, pageID, projectID string) (*models.Page, error) {
		pageQueries = append(pageQueries, pageID+"@"+projectID)
		pages := map[string]models.Page{
			"pub-page":   {Id: "pub-page", ProjectId: "published", Logic: []byte(`{"v":1}`)},
			"draft-page": {Id: "draft-page", ProjectId: "draft", Logic: []byte(`{"v":1}`)},
		}
		page, ok := pages[pageID]
		if !ok || (projectID != "" && page.ProjectId != projectID) {
			return nil, repositories.ErrPageNotFound
		}
		return &page, nil
	})
	projectRepo := mockrepo.NewMockProjectRepo().SetGetPublicProjectByID(func(_ context.Context, id string) (*models.Project, error) {
		if id == "published" {
			return &models.Project{ID: id}, nil
		}
		return nil, repositories.ErrProjectNotFound
	})
	return services.NewPageService(pageRepo, projectRepo), &pageQueries
}

func TestGetPublicPageByIDRequiresPublishedOwningProject(t *testing.T) {
	service, queries := publicPageFixture()
	ctx := context.Background()
	page, err := service.GetPublicPageByID(ctx, "published", "pub-page")
	if err != nil || page == nil || page.Id != "pub-page" {
		t.Fatalf("published page: %+v %v", page, err)
	}
	if _, err := service.GetPublicPageByID(ctx, "draft", "draft-page"); err == nil {
		t.Fatal("a page of an unpublished project was served")
	}
	if _, err := service.GetPublicPageByID(ctx, "published", "draft-page"); err == nil {
		t.Fatal("a page of another (unpublished) project was served through a published project id")
	}
	for _, q := range *queries {
		if q == "draft-page@" || q == "pub-page@" {
			t.Fatalf("page looked up without its project: %v", *queries)
		}
	}
	// A repository that finds no published project and no error still means not found.
	nilProject := services.NewPageService(mockrepo.NewMockPageRepo(), mockrepo.NewMockProjectRepo())
	if _, err := nilProject.GetPublicPageByID(ctx, "published", "pub-page"); err == nil {
		t.Fatal("missing project treated as published")
	}
}

func TestPublicPageRouteReturnsNotFoundForDrafts(t *testing.T) {
	service, _ := publicPageFixture()
	app := fiber.New()
	app.Get("/api/v1/pages/public/:projectid/:pageid", handlers.NewPageHandler(service).GetPublicPageByID)
	for path, want := range map[string]int{
		"/api/v1/pages/public/published/pub-page":   fiber.StatusOK,
		"/api/v1/pages/public/draft/draft-page":     fiber.StatusNotFound,
		"/api/v1/pages/public/published/draft-page": fiber.StatusNotFound,
	} {
		res, err := app.Test(httptest.NewRequest("GET", path, nil))
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != want {
			t.Errorf("%s: status %d, want %d", path, res.StatusCode, want)
		}
	}
}
