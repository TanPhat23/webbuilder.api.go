package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"my-go-app/internal/models"
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
	e.proj.OnGetProjectByID = func(context.Context, string, string) (*models.Project, error) { return p, err }
}

func (e *invitationEnv) inviter(u *models.User, err error) {
	e.user.OnGetUserByID = func(context.Context, string) (*models.User, error) { return u, err }
}

func TestInvitationService_CreateValidationAndLookups(t *testing.T) {
	ctx := context.Background()
	proj := &models.Project{Name: "Site"}
	user := &models.User{Email: "me@x.com"}

	e := newInvitationEnv()
	_, err := e.svc.CreateInvitation(ctx, "", "a@x.com", "", "u")
	wantErr(t, err, "projectId is required")
	_, err = e.svc.CreateInvitation(ctx, "p", "", "", "u")
	wantErr(t, err, "email is required")
	_, err = e.svc.CreateInvitation(ctx, "p", "a@x.com", "", "")
	wantErr(t, err, "invitedBy is required")
	wantCalls(t, e.proj.Calls)

	e.project(nil, errBoom)
	if _, err = e.svc.CreateInvitation(ctx, "p", "a@x.com", "", "u"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
	e.project(nil, nil)
	_, err = e.svc.CreateInvitation(ctx, "p", "a@x.com", "", "u")
	wantErr(t, err, "project does not exist")

	e.project(proj, nil)
	e.inviter(nil, errBoom)
	_, err = e.svc.CreateInvitation(ctx, "p", "a@x.com", "", "u")
	wantErr(t, err, "failed to get inviter details")
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}

	e.inviter(nil, nil)
	_, err = e.svc.CreateInvitation(ctx, "p", "a@x.com", "", "u")
	wantErr(t, err, "inviter does not exist")

	e.inviter(user, nil)
	_, err = e.svc.CreateInvitation(ctx, "p", "me@x.com", "", "u")
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
	e.inv.OnGetInvitationsByProject = func(context.Context, string) ([]models.Invitation, error) { return nil, errBoom }
	if _, err := e.svc.CreateInvitation(ctx, "p", "a@x.com", "", "u"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}

	now := time.Now()
	e.inv.OnGetInvitationsByProject = func(context.Context, string) ([]models.Invitation, error) {
		return []models.Invitation{{Email: "a@x.com", AcceptedAt: &now}, {Email: "a@x.com"}}, nil
	}
	_, err := e.svc.CreateInvitation(ctx, "p", "a@x.com", "", "u")
	wantErr(t, err, "invitation already sent")

	e.inv.OnGetInvitationsByProject = func(context.Context, string) ([]models.Invitation, error) { return nil, nil }
	e.inv.OnCreateInvitation = func(context.Context, *models.Invitation) (*models.Invitation, error) { return nil, errBoom }
	if _, err = e.svc.CreateInvitation(ctx, "p", "a@x.com", "", "u"); !errors.Is(err, errBoom) {
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
	e.inv.OnGetInvitationsByProject = func(context.Context, string) ([]models.Invitation, error) { return nil, nil }
	var saved *models.Invitation
	e.inv.OnCreateInvitation = func(_ context.Context, i *models.Invitation) (*models.Invitation, error) {
		saved = i
		return i, nil
	}
	before := time.Now()
	got, err := e.svc.CreateInvitation(ctx, "p", "a@x.com", "", "u")
	wantNoErr(t, err)
	if got != saved || saved.Role != models.RoleEditor || saved.ProjectId != "p" {
		t.Fatalf("saved = %+v", saved)
	}
	if saved.Email != "a@x.com" || saved.Id == "" || saved.Token == "" {
		t.Fatalf("saved identity = %+v", saved)
	}
	if d := saved.ExpiresAt.Sub(before); d < 7*24*time.Hour-time.Minute || d > 7*24*time.Hour+time.Minute {
		t.Fatalf("expiry = %v", d)
	}
	want := "a@x.com|Site|http://app/accept-invitation?token=" + saved.Token
	if len(e.mail.calls) != 1 || e.mail.calls[0] != want {
		t.Fatalf("mail = %v, want %q", e.mail.calls, want)
	}

	e.mail.err = errBoom
	got, err = e.svc.CreateInvitation(ctx, "p", "a@x.com", models.RoleViewer, "u")
	wantNoErr(t, err)
	if got.Role != models.RoleViewer {
		t.Fatalf("role = %v", got.Role)
	}
	_, err = NewInvitationService(e.inv, nil, e.proj, e.user, nil, "").CreateInvitation(ctx, "p", "a@x.com", "", "u")
	wantNoErr(t, err)
}

func TestInvitationService_CheckProjectOwnership(t *testing.T) {
	ctx := context.Background()
	e := newInvitationEnv()
	wantErr(t, e.svc.CheckProjectOwnership(ctx, "", "u"), "projectId is required")
	wantErr(t, e.svc.CheckProjectOwnership(ctx, "p", ""), "userId is required")

	e.project(nil, errBoom)
	if err := e.svc.CheckProjectOwnership(ctx, "p", "u"); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
	e.project(nil, nil)
	wantErr(t, e.svc.CheckProjectOwnership(ctx, "p", "u"), "project does not exist")
	e.project(&models.Project{OwnerId: "other"}, nil)
	wantErr(t, e.svc.CheckProjectOwnership(ctx, "p", "u"), "not the owner")
	e.project(&models.Project{OwnerId: "u"}, nil)
	wantNoErr(t, e.svc.CheckProjectOwnership(ctx, "p", "u"))
}

func TestInvitationService_GetAndAccept(t *testing.T) {
	ctx := context.Background()
	e := newInvitationEnv()
	_, err := e.svc.GetInvitationByID(ctx, "")
	wantErr(t, err, "invitation id is required")
	e.inv.OnGetInvitationByID = func(_ context.Context, id string) (*models.Invitation, error) {
		return &models.Invitation{Id: id}, nil
	}
	i, err := e.svc.GetInvitationByID(ctx, "i1")
	wantNoErr(t, err)
	if i.Id != "i1" {
		t.Fatalf("i = %+v", i)
	}

	wantErr(t, e.svc.AcceptInvitation(ctx, "", "u"), "token is required")
	wantErr(t, e.svc.AcceptInvitation(ctx, "t", ""), "userId is required")
	e.inv.OnAcceptInvitation = func(_ context.Context, tok, uid string) error {
		if tok != "t" || uid != "u" {
			t.Errorf("args %q %q", tok, uid)
		}
		return nil
	}
	wantNoErr(t, e.svc.AcceptInvitation(ctx, "t", "u"))
	e.inv.OnAcceptInvitation = func(context.Context, string, string) error { return errBoom }
	if err := e.svc.AcceptInvitation(ctx, "t", "u"); !errors.Is(err, errBoom) {
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
		{"GetInvitationsByProject", func(s *InvitationService, p string) ([]models.Invitation, error) {
			return s.GetInvitationsByProject(ctx, p, "u")
		}, func(r *fkInvitationRepo, err error) {
			r.OnGetInvitationsByProject = func(context.Context, string) ([]models.Invitation, error) {
				return []models.Invitation{{Id: "a"}}, err
			}
		}, "GetInvitationsByProject"},
		{"GetPendingInvitationsByProject", func(s *InvitationService, p string) ([]models.Invitation, error) {
			return s.GetPendingInvitationsByProject(ctx, p, "u")
		}, func(r *fkInvitationRepo, err error) {
			r.OnGetPendingInvitationsByProject = func(context.Context, string) ([]models.Invitation, error) {
				return []models.Invitation{{Id: "a"}}, err
			}
		}, "GetPendingInvitationsByProject"},
	}
	for _, l := range lists {
		e := newInvitationEnv()
		_, err := l.call(e.svc, "")
		wantErr(t, err, "projectId is required")

		e.proj.OnGetProjectByID = func(context.Context, string, string) (*models.Project, error) {
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
		{"Delete", func(s *InvitationService, id string) error { return s.DeleteInvitation(ctx, id) },
			func(r *fkInvitationRepo, err error) {
				r.OnDeleteInvitation = func(context.Context, string) error { return err }
			}, "DeleteInvitation", "invitation id is required"},
		{"Cancel", func(s *InvitationService, id string) error { return s.CancelInvitation(ctx, id) },
			func(r *fkInvitationRepo, err error) {
				r.OnCancelInvitation = func(context.Context, string) error { return err }
			}, "CancelInvitation", "invitation id is required"},
		{"UpdateStatus", func(s *InvitationService, id string) error {
			return s.UpdateInvitationStatus(ctx, id, models.InvitationStatus("accepted"))
		}, func(r *fkInvitationRepo, err error) {
			r.OnUpdateInvitationStatus = func(_ context.Context, id string, st models.InvitationStatus) error {
				if st != "accepted" {
					t.Errorf("status = %q", st)
				}
				return err
			}
		}, "UpdateInvitationStatus", "invitation id is required"},
	}
	for _, o := range ops {
		t.Run(o.name, func(t *testing.T) {
			e := newInvitationEnv()
			wantErr(t, o.call(e.svc, ""), o.noID)

			e.inv.OnGetInvitationByID = func(context.Context, string) (*models.Invitation, error) { return nil, errBoom }
			if err := o.call(e.svc, "i"); !errors.Is(err, errBoom) {
				t.Fatalf("err = %v", err)
			}
			e.inv.OnGetInvitationByID = func(context.Context, string) (*models.Invitation, error) { return nil, nil }
			wantErr(t, o.call(e.svc, "i"), "invitation does not exist")

			e.inv.OnGetInvitationByID = func(context.Context, string) (*models.Invitation, error) {
				return &models.Invitation{}, nil
			}
			e.inv.Calls = nil
			o.hook(e.inv, nil)
			wantNoErr(t, o.call(e.svc, "i"))
			wantCalls(t, e.inv.Calls, "GetInvitationByID", o.method)
			o.hook(e.inv, errBoom)
			if err := o.call(e.svc, "i"); !errors.Is(err, errBoom) {
				t.Fatalf("err = %v", err)
			}
		})
	}

	t.Run("UpdateStatus requires status", func(t *testing.T) {
		e := newInvitationEnv()
		wantErr(t, e.svc.UpdateInvitationStatus(ctx, "i", ""), "status is required")
		wantCalls(t, e.inv.Calls)
	})
}

var _ = strings.Contains
