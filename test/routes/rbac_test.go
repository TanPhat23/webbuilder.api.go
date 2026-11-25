package routes_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"my-go-app/internal/auth"
	"my-go-app/internal/authz"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/routes"
	"my-go-app/internal/services"
	"my-go-app/pkg/configs"
)

type rbacProjectRepo struct {
	repositories.ProjectRepositoryInterface
	roles map[string]models.CollaboratorRole
}

func (r *rbacProjectRepo) GetAccess(_ context.Context, projectID, userID string) (*models.ProjectAccess, error) {
	if projectID != "p1" {
		return nil, repositories.ErrProjectNotFound
	}
	role, ok := r.roles[userID]
	if !ok {
		return nil, repositories.ErrProjectNotFound
	}
	return &models.ProjectAccess{
		ProjectId:      projectID,
		UserId:         userID,
		Role:           role,
		IsOwner:        role == models.RoleOwner,
		IsCollaborator: role != models.RoleOwner,
		Permissions:    authz.PermissionsForRole(authz.Role(role)),
	}, nil
}

func (r *rbacProjectRepo) UpdateFields(_ context.Context, projectID, _ string, _ map[string]any) (*models.Project, error) {
	return &models.Project{ID: projectID, Name: "Updated"}, nil
}

func (r *rbacProjectRepo) GetWithAccess(_ context.Context, projectID, userID string) (*models.Project, error) {
	if projectID == "p1" {
		if _, ok := r.roles[userID]; ok {
			return &models.Project{ID: projectID, OwnerId: "owner-1", Name: "Site"}, nil
		}
	}
	return nil, repositories.ErrProjectNotFound
}

func (r *rbacProjectRepo) ListPages(_ context.Context, projectID, _ string) ([]models.Page, error) {
	return []models.Page{{Id: "pg1", ProjectId: projectID, Name: "Home"}}, nil
}

type rbacPageRepo struct {
	repositories.PageRepositoryInterface
}

func (r *rbacPageRepo) Create(context.Context, *models.Page) error { return nil }
func (r *rbacPageRepo) ListByProject(context.Context, string) ([]models.Page, error) {
	return []models.Page{{Id: "pg1", ProjectId: "p1", Name: "Home"}}, nil
}

type rbacCollabRepo struct {
	repositories.CollaboratorRepositoryInterface
}

func (r *rbacCollabRepo) ListByProject(context.Context, string) ([]models.Collaborator, error) {
	return []models.Collaborator{{Id: "c1", ProjectId: "p1", UserId: "editor-1", Role: models.RoleEditor}}, nil
}

type rbacContentTypeRepo struct {
	repositories.ContentTypeRepositoryInterface
}

func (r *rbacContentTypeRepo) Create(_ context.Context, ct *models.ContentType) (*models.ContentType, error) {
	ct.Id = "ct1"
	return ct, nil
}
func (r *rbacContentTypeRepo) ListByProject(context.Context, string) ([]models.ContentType, error) {
	return []models.ContentType{{Id: "ct1", Name: "Posts"}}, nil
}

type rbacSnapshotRepo struct {
	repositories.SnapshotRepositoryInterface
}

func (r *rbacSnapshotRepo) ListByProject(context.Context, string) ([]models.Snapshot, error) {
	return []models.Snapshot{{Id: "s1", ProjectId: "p1", Name: "v1"}}, nil
}
func (r *rbacSnapshotRepo) GetByID(_ context.Context, id string) (*models.Snapshot, error) {
	return &models.Snapshot{Id: id, ProjectId: "p1", Name: "v1"}, nil
}
func (r *rbacSnapshotRepo) Delete(context.Context, string) error { return nil }

func buildRBACTestApp() *fiber.App {
	roles := map[string]models.CollaboratorRole{
		"owner-1":  models.RoleOwner,
		"editor-1": models.RoleEditor,
		"viewer-1": models.RoleViewer,
	}

	repos := &repositories.Repositories{
		ProjectRepository:      &rbacProjectRepo{roles: roles},
		PageRepository:         &rbacPageRepo{},
		CollaboratorRepository: &rbacCollabRepo{},
		ContentTypeRepository:  &rbacContentTypeRepo{},
		SnapshotRepository:     &rbacSnapshotRepo{},
	}

	app := fiber.New(configs.Fiber())
	publicGroup := app.Group("/api/v1")
	protectedGroup := app.Group("/api/v1", func(c fiber.Ctx) error {
		uid := c.Get("X-Test-User")
		if uid != "" {
			identity := auth.Identity{UserID: uid}
			auth.SetFiber(c, identity)
			c.SetContext(authz.WithRequestCache(auth.WithIdentity(c.RequestCtx(), identity)))
		}
		return c.Next()
	})

	routes.RegisterProjects(publicGroup, protectedGroup, repos)
	routes.RegisterPages(publicGroup, protectedGroup, repos, nil)
	routes.RegisterCollaborators(protectedGroup, repos)
	routes.RegisterContent(publicGroup, protectedGroup, repos)
	routes.RegisterSnapshots(protectedGroup, repos)
	routes.RegisterInvitations(protectedGroup, &services.InvitationService{})

	return app
}

func TestRBACRouteMatrix(t *testing.T) {
	app := buildRBACTestApp()

	cases := []struct {
		name   string
		user   string
		method string
		path   string
		body   string
		want   int
	}{
		// Project update field-level split: editor can rename, cannot publish/change subdomain
		{"editor rename project", "editor-1", http.MethodPatch, "/api/v1/projects/p1", `{"name":"Renamed"}`, 200},
		{"editor publish project forbidden", "editor-1", http.MethodPatch, "/api/v1/projects/p1", `{"published":true}`, 403},
		{"editor subdomain forbidden", "editor-1", http.MethodPatch, "/api/v1/projects/p1", `{"subdomain":"abc"}`, 403},
		{"owner publish project allowed", "owner-1", http.MethodPatch, "/api/v1/projects/p1", `{"published":true}`, 200},
		{"viewer edit project forbidden", "viewer-1", http.MethodPatch, "/api/v1/projects/p1", `{"name":"Nope"}`, 403},
		{"stranger project access forbidden", "stranger-1", http.MethodPatch, "/api/v1/projects/p1", `{"name":"Nope"}`, 403},

		// Pages
		{"viewer list pages", "viewer-1", http.MethodGet, "/api/v1/pages/p1", "", 200},
		{"viewer create page forbidden", "viewer-1", http.MethodPost, "/api/v1/pages/p1", `{"name":"Blog","type":"page"}`, 403},
		{"editor create page allowed", "editor-1", http.MethodPost, "/api/v1/pages/p1", `{"name":"Blog","type":"page"}`, 201},

		// Snapshots
		{"viewer list snapshots", "viewer-1", http.MethodGet, "/api/v1/snapshots/p1", "", 200},
		{"viewer delete snapshot forbidden", "viewer-1", http.MethodDelete, "/api/v1/snapshots/p1/s1", "", 403},
		{"editor delete snapshot forbidden", "editor-1", http.MethodDelete, "/api/v1/snapshots/p1/s1", "", 403},
		{"owner delete snapshot allowed", "owner-1", http.MethodDelete, "/api/v1/snapshots/p1/s1", "", 200},

		// Collaborators
		{"viewer list collaborators allowed", "viewer-1", http.MethodGet, "/api/v1/collaborators/project/p1", "", 200},
		{"stranger list collaborators forbidden", "stranger-1", http.MethodGet, "/api/v1/collaborators/project/p1", "", 403},

		// Project-scoped CMS
		{"viewer list project content types", "viewer-1", http.MethodGet, "/api/v1/projects/p1/content-types", "", 200},
		{"viewer create content type forbidden", "viewer-1", http.MethodPost, "/api/v1/projects/p1/content-types", `{"name":"News"}`, 403},
		{"editor create content type allowed", "editor-1", http.MethodPost, "/api/v1/projects/p1/content-types", `{"name":"News"}`, 201},
		{"stranger list content types forbidden", "stranger-1", http.MethodGet, "/api/v1/projects/p1/content-types", "", 403},

		// Ownership transfer only owner
		{"editor transfer ownership forbidden", "editor-1", http.MethodPost, "/api/v1/projects/p1/transfer-ownership", `{"newOwnerId":"editor-1"}`, 403},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var reader io.Reader
			if tc.body != "" {
				reader = strings.NewReader(tc.body)
			}
			req := httptest.NewRequest(tc.method, tc.path, reader)
			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			if tc.user != "" {
				req.Header.Set("X-Test-User", tc.user)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			raw, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != tc.want {
				t.Fatalf("%s %s (user=%s): status = %d, want %d (body=%s)", tc.method, tc.path, tc.user, resp.StatusCode, tc.want, string(raw))
			}
		})
	}
}
