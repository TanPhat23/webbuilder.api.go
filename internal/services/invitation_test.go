package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"my-go-app/internal/authz"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type invitationFakeMailer struct {
	err   error
	calls []string
}

func (m *invitationFakeMailer) SendInvitationEmail(to, projectName, link string) error {
	m.calls = append(m.calls, to+"|"+projectName+"|"+link)
	return m.err
}

type invitationEnv struct {
	inv  *fkInvitationRepo
	proj *fkProjectRepo
	user *fkUserRepo
	mail *invitationFakeMailer
	svc  *InvitationService
}

func newInvitationEnv() *invitationEnv {
	e := &invitationEnv{inv: &fkInvitationRepo{}, proj: &fkProjectRepo{}, user: &fkUserRepo{}, mail: &invitationFakeMailer{}}
	e.svc = NewInvitationService(e.inv, &fkCollaboratorRepo{}, e.proj, e.user, e.mail, "http://app")
	return e
}

func (e *invitationEnv) project(p *models.Project, err error) {
	e.proj.OnGetByID = func(context.Context, string, string) (*models.Project, error) { return p, err }
}

func (e *invitationEnv) inviter(u *models.User, err error) {
	e.user.OnGetByID = func(context.Context, string) (*models.User, error) { return u, err }
}

func TestInvitationService_CreateValidationAndLookups(t *testing.T) {
	ctx := context.Background()
	proj := &models.Project{Name: "Site"}
	user := &models.User{Email: "me@x.com"}

	e := newInvitationEnv()
	_, err := e.svc.Create(ctx, "", "a@x.com", "", "u")
	wantErr(t, err, "projectId is required")
	_, err = e.svc.Create(ctx, "p", "", "", "u")
	wantErr(t, err, "email is required")
	_, err = e.svc.Create(ctx, "p", "a@x.com", "", "")
	wantErr(t, err, "invitedBy is required")
	wantCalls(t, e.proj.Calls)

	e.project(nil, errBoom)
	if _, err = e.svc.Create(ctx, "p", "a@x.com", "", "u"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
	e.project(nil, nil)
	_, err = e.svc.Create(ctx, "p", "a@x.com", "", "u")
	wantErr(t, err, "project does not exist")

	e.project(proj, nil)
	e.inviter(nil, errBoom)
	_, err = e.svc.Create(ctx, "p", "a@x.com", "", "u")
	wantErr(t, err, "failed to get inviter details")
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}

	e.inviter(nil, nil)
	_, err = e.svc.Create(ctx, "p", "a@x.com", "", "u")
	wantErr(t, err, "inviter does not exist")

	e.inviter(user, nil)
	_, err = e.svc.Create(ctx, "p", "me@x.com", "", "u")
	wantErr(t, err, "cannot invite yourself")
	wantCalls(t, e.inv.Calls)
}

func TestInvitationService_CreateDuplicateAndRepoErrors(t *testing.T) {
	ctx := context.Background()
	proj := &models.Project{Name: "Site"}
	user := &models.User{Email: "me@x.com"}

	e := newInvitationEnv()
	e.project(proj, nil)
	e.inviter(user, nil)
	e.inv.OnListByProject = func(context.Context, string) ([]models.Invitation, error) { return nil, errBoom }
	if _, err := e.svc.Create(ctx, "p", "a@x.com", "", "u"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}

	now := time.Now()
	e.inv.OnListByProject = func(context.Context, string) ([]models.Invitation, error) {
		return []models.Invitation{{Email: "a@x.com", AcceptedAt: &now}, {Email: "a@x.com"}}, nil
	}
	_, err := e.svc.Create(ctx, "p", "a@x.com", "", "u")
	wantErr(t, err, "invitation already sent")

	e.inv.OnListByProject = func(context.Context, string) ([]models.Invitation, error) { return nil, nil }
	e.inv.OnCreate = func(context.Context, *models.Invitation) (*models.Invitation, error) { return nil, errBoom }
	if _, err = e.svc.Create(ctx, "p", "a@x.com", "", "u"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
	if len(e.mail.calls) != 0 {
		t.Fatalf("mail sent despite failure: %v", e.mail.calls)
	}
}

func TestInvitationService_CreateSuccessAndMailerTolerance(t *testing.T) {
	ctx := context.Background()
	proj := &models.Project{Name: "Site"}
	user := &models.User{Email: "me@x.com"}

	e := newInvitationEnv()
	e.project(proj, nil)
	e.inviter(user, nil)
	e.inv.OnListByProject = func(context.Context, string) ([]models.Invitation, error) { return nil, nil }
	var saved *models.Invitation
	e.inv.OnCreate = func(_ context.Context, i *models.Invitation) (*models.Invitation, error) {
		saved = i
		return i, nil
	}
	before := time.Now()
	got, err := e.svc.Create(ctx, "p", "a@x.com", "", "u")
	wantNoErr(t, err)
	if got != saved || saved.Role != models.RoleEditor || saved.ProjectId != "p" {
		t.Fatalf("saved = %+v", saved)
	}
	if saved.Email != "a@x.com" || saved.Id == "" || saved.Token == "" {
		t.Fatalf("saved identity = %+v", saved)
	}
	if len(saved.Token) != 43 { // 32 bytes base64.RawURLEncoding
		t.Fatalf("expected 43-char crypto/rand token, got len=%d (%q)", len(saved.Token), saved.Token)
	}
	if d := saved.ExpiresAt.Sub(before); d < 7*24*time.Hour-time.Minute || d > 7*24*time.Hour+time.Minute {
		t.Fatalf("expiry = %v", d)
	}
	want := "a@x.com|Site|http://app/accept-invitation?token=" + saved.Token
	if len(e.mail.calls) != 1 || e.mail.calls[0] != want {
		t.Fatalf("mail = %v, want %q", e.mail.calls, want)
	}

	e.mail.err = errBoom
	got, err = e.svc.Create(ctx, "p", "a@x.com", models.RoleViewer, "u")
	wantNoErr(t, err)
	if got.Role != models.RoleViewer {
		t.Fatalf("role = %v", got.Role)
	}
	_, err = NewInvitationService(e.inv, nil, e.proj, e.user, nil, "").Create(ctx, "p", "a@x.com", "", "u")
	wantNoErr(t, err)
}

func TestInvitationService_CheckProjectOwnership(t *testing.T) {
	ctx := context.Background()
	e := newInvitationEnv()
	wantErr(t, e.svc.RequireProjectOwnership(ctx, "", "u"), "projectId is required")
	wantErr(t, e.svc.RequireProjectOwnership(ctx, "p", ""), "userId is required")

	e.project(nil, errBoom)
	if err := e.svc.RequireProjectOwnership(ctx, "p", "u"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
	e.project(nil, nil)
	wantErr(t, e.svc.RequireProjectOwnership(ctx, "p", "u"), "project does not exist")
	e.project(&models.Project{OwnerId: "other"}, nil)
	wantErr(t, e.svc.RequireProjectOwnership(ctx, "p", "u"), "not the owner")
	e.project(&models.Project{OwnerId: "u"}, nil)
	wantNoErr(t, e.svc.RequireProjectOwnership(ctx, "p", "u"))
}

func TestInvitationService_GetAndAccept(t *testing.T) {
	ctx := context.Background()
	e := newInvitationEnv()
	_, err := e.svc.GetByID(ctx, "")
	wantErr(t, err, "invitation id is required")
	e.inv.OnGetByID = func(_ context.Context, id string) (*models.Invitation, error) {
		return &models.Invitation{Id: id}, nil
	}
	i, err := e.svc.GetByID(ctx, "i1")
	wantNoErr(t, err)
	if i.Id != "i1" {
		t.Fatalf("i = %+v", i)
	}

	wantErr(t, e.svc.Accept(ctx, "", "u"), "token is required")
	wantErr(t, e.svc.Accept(ctx, "t", ""), "userId is required")

	// Missing invitation
	e.inv.OnGetByToken = func(context.Context, string) (*models.Invitation, error) { return nil, nil }
	wantErr(t, e.svc.Accept(ctx, "t", "u"), "invitation does not exist")

	// Reject non-assignable role in invitation
	e.inv.OnGetByToken = func(context.Context, string) (*models.Invitation, error) {
		return &models.Invitation{ProjectId: "p", Email: "u@x.com", Role: models.RoleOwner}, nil
	}
	if err := e.svc.Accept(ctx, "t", "u"); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("err = %v, want authz.ErrForbidden for owner role", err)
	}

	// Reject email mismatch
	e.inv.OnGetByToken = func(context.Context, string) (*models.Invitation, error) {
		return &models.Invitation{ProjectId: "p", Email: "invited@x.com", Role: models.RoleEditor}, nil
	}
	e.inviter(&models.User{Id: "u", Email: "other@x.com"}, nil)
	if err := e.svc.Accept(ctx, "t", "u"); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("err = %v, want authz.ErrForbidden for email mismatch", err)
	}

	// Reject Create with owner role
	if _, err := e.svc.Create(ctx, "p", "a@x.com", models.RoleOwner, "owner"); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("err = %v, want authz.ErrForbidden when inviting owner", err)
	}

	// Matching case-insensitive email succeeds
	e.inviter(&models.User{Id: "u", Email: "INVITED@x.com"}, nil)
	e.proj.OnGetAccess = func(context.Context, string, string) (*models.ProjectAccess, error) {
		return nil, repositories.ErrProjectNotFound
	}
	e.inv.OnAccept = func(_ context.Context, tok, uid string) error {
		if tok != "t" || uid != "u" {
			t.Errorf("args %q %q", tok, uid)
		}
		return nil
	}
	wantNoErr(t, e.svc.Accept(ctx, "t", "u"))
	e.inv.OnAccept = func(context.Context, string, string) error { return errBoom }
	if err := e.svc.Accept(ctx, "t", "u"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
}

func TestInvitationService_ListsByProject(t *testing.T) {
	ctx := context.Background()
	lists := []struct {
		name   string
		call   func(s *InvitationService, pid string) ([]models.Invitation, error)
		hook   func(r *fkInvitationRepo, err error)
		method string
	}{
		{"ListByProject", func(s *InvitationService, p string) ([]models.Invitation, error) {
			return s.ListByProject(ctx, p, "u")
		}, func(r *fkInvitationRepo, err error) {
			r.OnListByProject = func(context.Context, string) ([]models.Invitation, error) {
				return []models.Invitation{{Id: "a"}}, err
			}
		}, "ListByProject"},
		{"ListPendingByProject", func(s *InvitationService, p string) ([]models.Invitation, error) {
			return s.ListPendingByProject(ctx, p, "u")
		}, func(r *fkInvitationRepo, err error) {
			r.OnListPendingByProject = func(context.Context, string) ([]models.Invitation, error) {
				return []models.Invitation{{Id: "a"}}, err
			}
		}, "ListPendingByProject"},
	}
	for _, l := range lists {
		e := newInvitationEnv()
		_, err := l.call(e.svc, "")
		wantErr(t, err, "projectId is required")

		e.proj.OnGetByID = func(context.Context, string, string) (*models.Project, error) {
			return &models.Project{OwnerId: "u"}, nil
		}
		l.hook(e.inv, nil)
		_, err = l.call(e.svc, "p")
		wantNoErr(t, err)
		wantCalls(t, e.inv.Calls, l.method)
	}
}

func TestInvitationService_StatefulMutations(t *testing.T) {
	ctx := context.Background()
	ops := []struct {
		name   string
		call   func(s *InvitationService, id string) error
		hook   func(r *fkInvitationRepo, err error)
		method string
		noID   string
	}{
		{"Delete", func(s *InvitationService, id string) error { return s.Delete(ctx, id) },
			func(r *fkInvitationRepo, err error) {
				r.OnDelete = func(context.Context, string) error { return err }
			}, "Delete", "invitation id is required"},
		{"Cancel", func(s *InvitationService, id string) error { return s.Cancel(ctx, id) },
			func(r *fkInvitationRepo, err error) {
				r.OnCancel = func(context.Context, string) error { return err }
			}, "Cancel", "invitation id is required"},
		{"UpdateStatus", func(s *InvitationService, id string) error {
			return s.UpdateStatus(ctx, id, models.InvitationStatus("accepted"))
		}, func(r *fkInvitationRepo, err error) {
			r.OnUpdateStatus = func(_ context.Context, id string, st models.InvitationStatus) error {
				if st != "accepted" {
					t.Errorf("status = %q", st)
				}
				return err
			}
		}, "UpdateStatus", "invitation id is required"},
	}
	for _, o := range ops {
		t.Run(o.name, func(t *testing.T) {
			e := newInvitationEnv()
			wantErr(t, o.call(e.svc, ""), o.noID)

			e.inv.OnGetByID = func(context.Context, string) (*models.Invitation, error) { return nil, errBoom }
			if err := o.call(e.svc, "i"); !errors.Is(err, errBoom) {
				t.Fatalf("err = %v", err)
			}
			e.inv.OnGetByID = func(context.Context, string) (*models.Invitation, error) { return nil, nil }
			wantErr(t, o.call(e.svc, "i"), "invitation does not exist")

			e.inv.OnGetByID = func(context.Context, string) (*models.Invitation, error) {
				return &models.Invitation{}, nil
			}
			e.inv.Calls = nil
			o.hook(e.inv, nil)
			wantNoErr(t, o.call(e.svc, "i"))
			wantCalls(t, e.inv.Calls, "GetByID", o.method)
			o.hook(e.inv, errBoom)
			if err := o.call(e.svc, "i"); !errors.Is(err, errBoom) {
				t.Fatalf("err = %v", err)
			}
		})
	}

	t.Run("UpdateStatus requires status", func(t *testing.T) {
		e := newInvitationEnv()
		wantErr(t, e.svc.UpdateStatus(ctx, "i", ""), "status is required")
		wantCalls(t, e.inv.Calls)
	})
}

var _ = strings.Contains
