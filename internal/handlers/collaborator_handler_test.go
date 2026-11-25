package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/services"
)

type collaboratorSvcStub struct {
	services.CollaboratorServiceInterface
	err       error
	created   *models.Collaborator
	listProj  string
	listUser  string
	roleID    string
	role      models.CollaboratorRole
	deletedID string
}

func (s *collaboratorSvcStub) ListByProject(_ context.Context, projectID, userID string) ([]models.Collaborator, error) {
	s.listProj, s.listUser = strings.Clone(projectID), strings.Clone(userID)
	if s.err != nil {
		return nil, s.err
	}
	return []models.Collaborator{{Id: "c1", ProjectId: projectID, UserId: "u2", Role: models.CollaboratorRole("editor")}}, nil
}
func (s *collaboratorSvcStub) GetByID(_ context.Context, id, _ string) (*models.Collaborator, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &models.Collaborator{Id: id, ProjectId: "p1"}, nil
}
func (s *collaboratorSvcStub) UpdateRole(_ context.Context, id, _ string, role models.CollaboratorRole) error {
	s.roleID, s.role = strings.Clone(id), models.CollaboratorRole(strings.Clone(string(role)))
	return s.err
}
func (s *collaboratorSvcStub) Delete(_ context.Context, id, _ string) error {
	s.deletedID = strings.Clone(id)
	return s.err
}
func (s *collaboratorSvcStub) LeaveProject(_ context.Context, _, _ string) error {
	return s.err
}
func (s *collaboratorSvcStub) Create(_ context.Context, _ string, c *models.Collaborator) (*models.Collaborator, error) {
	s.created = c
	if s.err != nil {
		return nil, s.err
	}
	c.Id = "c-new"
	return c, nil
}

func collaboratorKit(userID string, svc *collaboratorSvcStub) *fiber.App {
	h := NewCollaboratorHandler(svc)
	return newKitApp(userID, func(app *fiber.App) {
		app.Get("/projects/:projectid/collaborators", h.ListByProject)
		app.Get("/collaborators/:collaboratorid", h.GetByID)
		app.Patch("/collaborators/:collaboratorid", h.UpdateRole)
		app.Delete("/collaborators/:collaboratorid", h.Delete)
		app.Post("/collaborators", h.Create)
		// Routes without the parameter exercise the "missing param" branches.
		app.Get("/noparam/projects", h.ListByProject)
		app.Get("/noparam", h.GetByID)
		app.Patch("/noparam", h.UpdateRole)
		app.Delete("/noparam", h.Delete)
	})
}

func TestCollaboratorHandlerStatuses(t *testing.T) {
	cases := []struct {
		name         string
		user         string
		method, path string
		body         string
		err          error
		want         int
		contains     string
	}{
		{"list ok", "u1", "GET", "/projects/p1/collaborators", "", nil, 200, `"collaborators":[{`},
		{"list unauthenticated", "", "GET", "/projects/p1/collaborators", "", nil, 401, "Unauthorized"},
		{"list missing param", "u1", "GET", "/noparam/projects", "", nil, 400, "projectid is required"},
		{"list not found", "u1", "GET", "/projects/p1/collaborators", "", gorm.ErrRecordNotFound, 404, "Project not found"},
		{"list internal", "u1", "GET", "/projects/p1/collaborators", "", errInternal, 500, "Failed to retrieve collaborators"},
		{"get ok", "u1", "GET", "/collaborators/c9", "", nil, 200, `"id":"c9"`},
		{"get missing param", "u1", "GET", "/noparam", "", nil, 400, "collaboratorid is required"},
		{"get not found", "u1", "GET", "/collaborators/c9", "", gorm.ErrRecordNotFound, 404, "Collaborator not found"},
		{"get internal", "u1", "GET", "/collaborators/c9", "", errInternal, 500, "Failed to retrieve collaborator"},
		{"role ok", "u1", "PATCH", "/collaborators/c9", `{"role":"viewer"}`, nil, 200, "Role updated successfully"},
		{"role missing param", "u1", "PATCH", "/noparam", `{"role":"viewer"}`, nil, 400, "collaboratorid is required"},
		{"role bad json", "u1", "PATCH", "/collaborators/c9", `{`, nil, 0, ""},
		{"role invalid value", "u1", "PATCH", "/collaborators/c9", `{"role":"admin"}`, nil, 0, ""},
		{"role missing value", "u1", "PATCH", "/collaborators/c9", `{}`, nil, 0, ""},
		{"role internal", "u1", "PATCH", "/collaborators/c9", `{"role":"viewer"}`, errInternal, 500, "Failed to update collaborator role"},
		{"delete ok", "u1", "DELETE", "/collaborators/c9", "", nil, 204, ""},
		{"delete missing param", "u1", "DELETE", "/noparam", "", nil, 400, "collaboratorid is required"},
		{"delete internal", "u1", "DELETE", "/collaborators/c9", "", errInternal, 500, "Failed to delete collaborator"},
		{"create ok", "u1", "POST", "/collaborators", `{"projectId":"p1","userId":"u2","role":"editor"}`, nil, 201, "c-new"},
		{"create bad json", "u1", "POST", "/collaborators", `{`, nil, 0, ""},
		{"create invalid role", "u1", "POST", "/collaborators", `{"projectId":"p1","userId":"u2","role":"boss"}`, nil, 0, ""},
		{"create missing user", "u1", "POST", "/collaborators", `{"projectId":"p1","role":"editor"}`, nil, 0, ""},
		{"create internal", "u1", "POST", "/collaborators", `{"projectId":"p1","userId":"u2","role":"editor"}`, errInternal, 500, "Failed to create collaborator"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := kitSend(t, collaboratorKit(tc.user, &collaboratorSvcStub{err: tc.err}), tc.method, tc.path, tc.body)
			if tc.contains == "" {
				kitExpect(t, r, tc.want)
			} else {
				kitExpect(t, r, tc.want, tc.contains)
			}
		})
	}
}

func TestCollaboratorHandlerForwardsArguments(t *testing.T) {
	svc := &collaboratorSvcStub{}
	app := collaboratorKit("u1", svc)

	kitExpect(t, kitSend(t, app, "GET", "/projects/p7/collaborators", ""), 200)
	if svc.listProj != "p7" || svc.listUser != "u1" {
		t.Fatalf("list called with project=%q user=%q", svc.listProj, svc.listUser)
	}

	kitExpect(t, kitSend(t, app, "PATCH", "/collaborators/c5", `{"role":"viewer"}`), 200)
	if svc.roleID != "c5" || svc.role != models.CollaboratorRole("viewer") {
		t.Fatalf("role update id=%q role=%q", svc.roleID, svc.role)
	}

	r := kitSend(t, app, "POST", "/collaborators", `{"projectId":"p3","userId":"u4","role":"editor"}`)
	kitExpect(t, r, 201, `"id":"c-new"`)
	if svc.created == nil || svc.created.ProjectId != "p3" || svc.created.UserId != "u4" || svc.created.Role != models.CollaboratorRole("editor") {
		t.Fatalf("created %+v", svc.created)
	}

	kitExpect(t, kitSend(t, app, "POST", "/collaborators", `{"projectId":"p3","userId":"u4","role":"owner"}`), 0)
}
