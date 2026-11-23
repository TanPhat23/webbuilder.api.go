package handlers

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"my-go-app/internal/models"
	"my-go-app/internal/services"
)

// invitationHandlerStub fails the method named in errs and records each call.
type invitationHandlerStub struct {
	services.InvitationServiceInterface
	errs  map[string]error
	calls []string
}

func (s *invitationHandlerStub) rec(name string, args ...any) error {
	s.calls = append(s.calls, name+fmt.Sprint(args...))
	return s.errs[name]
}

func (s *invitationHandlerStub) CreateInvitation(_ context.Context, projectID, email string, role models.CollaboratorRole, invitedBy string) (*models.Invitation, error) {
	if err := s.rec("Create", "[", projectID, "|", email, "|", role, "|", invitedBy, "]"); err != nil {
		return nil, err
	}
	return &models.Invitation{Id: "inv-1", ProjectId: projectID, Email: email, Role: role, Status: models.InvitationStatusPending}, nil
}
func (s *invitationHandlerStub) AcceptInvitation(_ context.Context, token, userID string) error {
	return s.rec("Accept", "[", token, "|", userID, "]")
}
func (s *invitationHandlerStub) GetInvitationsByProject(_ context.Context, projectID, userID string) ([]models.Invitation, error) {
	if err := s.rec("List", "[", projectID, "]"); err != nil {
		return nil, err
	}
	return []models.Invitation{{Id: "inv-1", ProjectId: projectID, Email: "a@b.co"}}, nil
}
func (s *invitationHandlerStub) GetPendingInvitationsByProject(_ context.Context, projectID, userID string) ([]models.Invitation, error) {
	if err := s.rec("Pending", "[", projectID, "]"); err != nil {
		return nil, err
	}
	return []models.Invitation{{Id: "inv-2", ProjectId: projectID, Status: models.InvitationStatusPending}}, nil
}
func (s *invitationHandlerStub) GetInvitationByID(_ context.Context, id string) (*models.Invitation, error) {
	if err := s.rec("Get", "[", id, "]"); err != nil {
		return nil, err
	}
	return &models.Invitation{Id: id, ProjectId: "proj-of-" + id}, nil
}
func (s *invitationHandlerStub) CheckProjectOwnership(_ context.Context, projectID, userID string) error {
	return s.rec("Owner", "[", projectID, "|", userID, "]")
}
func (s *invitationHandlerStub) CancelInvitation(_ context.Context, id string) error {
	return s.rec("Cancel", "[", id, "]")
}
func (s *invitationHandlerStub) UpdateInvitationStatus(_ context.Context, id string, status models.InvitationStatus) error {
	return s.rec("Status", "[", id, "|", status, "]")
}
func (s *invitationHandlerStub) DeleteInvitation(_ context.Context, id string) error {
	return s.rec("Delete", "[", id, "]")
}

func invitationHandlerApp(user string, svc *invitationHandlerStub) *fiber.App {
	h := NewInvitationHandler(svc)
	return newKitApp(user, func(app *fiber.App) {
		app.Post("/invitations", h.CreateInvitation)
		app.Post("/invitations/accept", h.AcceptInvitation)
		app.Get("/projects/:projectid/invitations", h.GetInvitationsByProject)
		app.Get("/projects/:projectid/invitations/pending", h.GetPendingInvitationsByProject)
		app.Patch("/invitations/:invitationid", h.UpdateInvitationStatus)
		app.Post("/invitations/:invitationid/cancel", h.CancelInvitation)
		app.Delete("/invitations/:invitationid", h.DeleteInvitation)
		app.Post("/noid/cancel", h.CancelInvitation)
		app.Patch("/noid", h.UpdateInvitationStatus)
		app.Delete("/noid", h.DeleteInvitation)
		app.Get("/noproject", h.GetInvitationsByProject)
	})
}

func TestCreateInvitation(t *testing.T) {
	svc := &invitationHandlerStub{}
	r := kitSend(t, invitationHandlerApp("u1", svc), "POST", "/invitations", `{"project_id":"p1","email":"a@b.co","role":"viewer"}`)
	kitExpect(t, r, 201, "inv-1", "viewer", "pending")
	if want := []string{"Create[p1|a@b.co|viewer|u1]"}; !reflect.DeepEqual(svc.calls, want) {
		t.Fatalf("calls = %v, want %v", svc.calls, want)
	}
}

func TestCreateInvitationDefaultsRoleToEditorWithoutValidator(t *testing.T) {
	// The production validator requires a role, so the default only applies where
	// the app has no struct validator installed.
	svc := &invitationHandlerStub{}
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error { c.Locals("userId", "u1"); return c.Next() })
	app.Post("/invitations", NewInvitationHandler(svc).CreateInvitation)
	kitExpect(t, kitSend(t, app, "POST", "/invitations", `{"project_id":"p1","email":"a@b.co"}`), 201, "editor")
	if want := []string{"Create[p1|a@b.co|editor|u1]"}; !reflect.DeepEqual(svc.calls, want) {
		t.Fatalf("calls = %v, want %v", svc.calls, want)
	}
}

func TestCreateInvitationRejections(t *testing.T) {
	cases := []struct {
		name, user, body string
		err              error
		want             int
		expect           string
	}{
		{"anonymous", "", `{"project_id":"p1","email":"a@b.co","role":"viewer"}`, nil, 401, ""},
		{"bad json", "u1", `{`, nil, 0, ""},
		{"bad email", "u1", `{"project_id":"p1","email":"nope","role":"viewer"}`, nil, 0, ""},
		{"missing project", "u1", `{"email":"a@b.co","role":"viewer"}`, nil, 0, ""},
		{"missing role", "u1", `{"project_id":"p1","email":"a@b.co"}`, nil, 0, ""},
		{"unknown role", "u1", `{"project_id":"p1","email":"a@b.co","role":"boss"}`, nil, 0, ""},
		{"service refuses", "u1", `{"project_id":"p1","email":"a@b.co","role":"viewer"}`, fmt.Errorf("already invited"), 400, "already invited"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &invitationHandlerStub{errs: map[string]error{"Create": tc.err}}
			r := kitSend(t, invitationHandlerApp(tc.user, svc), "POST", "/invitations", tc.body)
			if tc.expect == "" {
				kitExpect(t, r, tc.want)
			} else {
				kitExpect(t, r, tc.want, "Failed to create invitation", tc.expect)
			}
			if tc.err == nil && len(svc.calls) != 0 {
				t.Fatalf("service called: %v", svc.calls)
			}
		})
	}
}

func TestListInvitations(t *testing.T) {
	for _, tc := range []struct{ name, path, call, internal string }{
		{"all", "/projects/p1/invitations", "List[p1]", "Failed to retrieve invitations"},
		{"pending", "/projects/p1/invitations/pending", "Pending[p1]", "Failed to retrieve pending invitations"},
	} {
		method := map[string]string{"all": "List", "pending": "Pending"}[tc.name]
		t.Run(tc.name+" ok", func(t *testing.T) {
			svc := &invitationHandlerStub{}
			kitExpect(t, kitSend(t, invitationHandlerApp("u1", svc), "GET", tc.path, ""), 200, "p1")
			if want := []string{"Owner[p1|u1]", tc.call}; !reflect.DeepEqual(svc.calls, want) {
				t.Fatalf("calls = %v, want %v", svc.calls, want)
			}
		})
		t.Run(tc.name+" anonymous", func(t *testing.T) {
			kitExpect(t, kitSend(t, invitationHandlerApp("", &invitationHandlerStub{}), "GET", tc.path, ""), 401)
		})
		t.Run(tc.name+" forbidden", func(t *testing.T) {
			svc := &invitationHandlerStub{errs: map[string]error{"Owner": fmt.Errorf("not the owner")}}
			kitExpect(t, kitSend(t, invitationHandlerApp("u1", svc), "GET", tc.path, ""), 403, "Access denied", "not the owner")
			if len(svc.calls) != 1 {
				t.Fatalf("listing ran after the ownership check failed: %v", svc.calls)
			}
		})
		t.Run(tc.name+" failure", func(t *testing.T) {
			svc := &invitationHandlerStub{errs: map[string]error{method: errInternal}}
			kitExpect(t, kitSend(t, invitationHandlerApp("u1", svc), "GET", tc.path, ""), 500, tc.internal)
		})
	}
	kitExpect(t, kitSend(t, invitationHandlerApp("u1", &invitationHandlerStub{}), "GET", "/noproject", ""), 400, "projectid is required")
}

func TestAcceptInvitation(t *testing.T) {
	svc := &invitationHandlerStub{}
	kitExpect(t, kitSend(t, invitationHandlerApp("u1", svc), "POST", "/invitations/accept", `{"token":"tok"}`), 200, "Invitation accepted successfully")
	if want := []string{"Accept[tok|u1]"}; !reflect.DeepEqual(svc.calls, want) {
		t.Fatalf("calls = %v", svc.calls)
	}
	kitExpect(t, kitSend(t, invitationHandlerApp("", &invitationHandlerStub{}), "POST", "/invitations/accept", `{"token":"tok"}`), 401)
	kitExpect(t, kitSend(t, invitationHandlerApp("u1", &invitationHandlerStub{}), "POST", "/invitations/accept", `{`), 0)
	kitExpect(t, kitSend(t, invitationHandlerApp("u1", &invitationHandlerStub{}), "POST", "/invitations/accept", `{}`), 0)
	bad := &invitationHandlerStub{errs: map[string]error{"Accept": fmt.Errorf("invitation expired")}}
	kitExpect(t, kitSend(t, invitationHandlerApp("u1", bad), "POST", "/invitations/accept", `{"token":"tok"}`), 400, "Failed to accept invitation", "invitation expired")
}

// invitationHandlerMutations covers cancel, status update and delete, which
// share the load invitation / check ownership / apply sequence.
var invitationHandlerMutations = []struct {
	name, method, path, body, op, okMsg, internal, noID string
	okStatus                                            int
}{
	{"cancel", "POST", "/invitations/inv-9/cancel", "", "Cancel[inv-9]", "Invitation cancelled successfully", "Failed to cancel invitation", "/noid/cancel", 200},
	{"update status", "PATCH", "/invitations/inv-9", `{"status":"cancelled"}`, "Status[inv-9|cancelled]", "Invitation status updated successfully", "Failed to update invitation status", "/noid", 200},
	{"delete", "DELETE", "/invitations/inv-9", "", "Delete[inv-9]", "", "Failed to delete invitation", "/noid", 204},
}

func TestInvitationMutations(t *testing.T) {
	for _, m := range invitationHandlerMutations {
		opName := map[string]string{"cancel": "Cancel", "update status": "Status", "delete": "Delete"}[m.name]
		t.Run(m.name+" ok", func(t *testing.T) {
			svc := &invitationHandlerStub{}
			r := kitSend(t, invitationHandlerApp("u1", svc), m.method, m.path, m.body)
			if m.okMsg != "" {
				kitExpect(t, r, m.okStatus, m.okMsg)
			} else {
				kitExpect(t, r, m.okStatus)
			}
			// Ownership is checked against the invitation's own project.
			if want := []string{"Get[inv-9]", "Owner[proj-of-inv-9|u1]", m.op}; !reflect.DeepEqual(svc.calls, want) {
				t.Fatalf("calls = %v, want %v", svc.calls, want)
			}
		})
		t.Run(m.name+" anonymous", func(t *testing.T) {
			svc := &invitationHandlerStub{}
			kitExpect(t, kitSend(t, invitationHandlerApp("", svc), m.method, m.path, m.body), 401)
			if len(svc.calls) != 0 {
				t.Fatalf("service called: %v", svc.calls)
			}
		})
		t.Run(m.name+" missing id", func(t *testing.T) {
			kitExpect(t, kitSend(t, invitationHandlerApp("u1", &invitationHandlerStub{}), m.method, m.noID, m.body), 400, "invitationid is required")
		})
		t.Run(m.name+" invitation not found", func(t *testing.T) {
			svc := &invitationHandlerStub{errs: map[string]error{"Get": gorm.ErrRecordNotFound}}
			kitExpect(t, kitSend(t, invitationHandlerApp("u1", svc), m.method, m.path, m.body), 404, "Invitation not found")
			if len(svc.calls) != 1 {
				t.Fatalf("calls = %v", svc.calls)
			}
		})
		t.Run(m.name+" lookup failure", func(t *testing.T) {
			svc := &invitationHandlerStub{errs: map[string]error{"Get": errInternal}}
			kitExpect(t, kitSend(t, invitationHandlerApp("u1", svc), m.method, m.path, m.body), 500, "Failed to retrieve invitation")
		})
		t.Run(m.name+" forbidden", func(t *testing.T) {
			svc := &invitationHandlerStub{errs: map[string]error{"Owner": fmt.Errorf("not the owner")}}
			kitExpect(t, kitSend(t, invitationHandlerApp("u1", svc), m.method, m.path, m.body), 403, "Access denied")
			if len(svc.calls) != 2 {
				t.Fatalf("mutation ran after the ownership check failed: %v", svc.calls)
			}
		})
		t.Run(m.name+" failure", func(t *testing.T) {
			svc := &invitationHandlerStub{errs: map[string]error{opName: errInternal}}
			kitExpect(t, kitSend(t, invitationHandlerApp("u1", svc), m.method, m.path, m.body), 500, m.internal)
		})
	}
}

func TestUpdateInvitationStatusValidation(t *testing.T) {
	for _, body := range []string{`{`, `{}`, `{"status":"archived"}`} {
		svc := &invitationHandlerStub{}
		kitExpect(t, kitSend(t, invitationHandlerApp("u1", svc), "PATCH", "/invitations/inv-9", body), 0)
		if len(svc.calls) != 0 {
			t.Fatalf("body %s reached the service: %v", body, svc.calls)
		}
	}
}
