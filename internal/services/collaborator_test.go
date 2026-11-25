package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/authz"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

func collabSvc() (*CollaboratorService, *fkCollaboratorRepo, *fkProjectRepo) {
	c, p := &fkCollaboratorRepo{}, &fkProjectRepo{}
	return NewCollaboratorService(c, p), c, p
}

func ownerProjectAccess(p *fkProjectRepo, projectID, ownerID string) {
	p.OnGetAccess = func(_ context.Context, pid, uid string) (*models.ProjectAccess, error) {
		if pid == projectID && uid == ownerID {
			return &models.ProjectAccess{ProjectId: pid, UserId: uid, Role: models.RoleOwner, IsOwner: true}, nil
		}
		return nil, repositories.ErrProjectNotFound
	}
}

func TestCollaboratorCreate(t *testing.T) {
	ctx := context.Background()
	valid := func() *models.Collaborator {
		return &models.Collaborator{ProjectId: "p", UserId: "u", Role: models.RoleEditor}
	}
	for _, tc := range []struct {
		name   string
		caller string
		in     *models.Collaborator
		msg    string
	}{
		{"unauthenticated", "", valid(), "unauthorized"},
		{"nil", "owner", nil, "cannot be nil"},
		{"no project", "owner", &models.Collaborator{UserId: "u"}, "projectId is required"},
		{"no user", "owner", &models.Collaborator{ProjectId: "p"}, "userId is required"},
		{"owner role forbidden", "owner", &models.Collaborator{ProjectId: "p", UserId: "u", Role: models.RoleOwner}, "cannot be assigned"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, c, p := collabSvc()
			_, err := s.Create(ctx, tc.caller, tc.in)
			wantErr(t, err, tc.msg)
			wantCalls(t, c.Calls)
			wantCalls(t, p.Calls)
		})
	}
	t.Run("non owner forbidden", func(t *testing.T) {
		s, c, p := collabSvc()
		p.OnGetAccess = func(context.Context, string, string) (*models.ProjectAccess, error) {
			return &models.ProjectAccess{ProjectId: "p", UserId: "editor", Role: models.RoleEditor}, nil
		}
		_, err := s.Create(ctx, "editor", valid())
		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("err = %v, want ErrForbidden", err)
		}
		wantCalls(t, c.Calls)
	})
	t.Run("success and repo error", func(t *testing.T) {
		s, c, p := collabSvc()
		ownerProjectAccess(p, "p", "owner")
		c.OnGetByProjectAndUser = func(context.Context, string, string) (*models.Collaborator, error) {
			return nil, nil
		}
		want := &models.Collaborator{Id: "c"}
		c.OnCreate = func(context.Context, *models.Collaborator) (*models.Collaborator, error) { return want, nil }
		got, err := s.Create(ctx, "owner", valid())
		wantNoErr(t, err)
		if got != want {
			t.Fatal("wrong result")
		}
		c.OnCreate = func(context.Context, *models.Collaborator) (*models.Collaborator, error) { return nil, errBoom }
		if _, err := s.Create(ctx, "owner", valid()); !errors.Is(err, errBoom) {
			t.Fatalf("err %v", err)
		}
	})
}

func TestCollaboratorGetByProject(t *testing.T) {
	ctx := context.Background()
	s, c, p := collabSvc()
	_, err := s.ListByProject(ctx, "", "u")
	wantErr(t, err, "projectId is required")

	p.OnGetAccess = func(context.Context, string, string) (*models.ProjectAccess, error) { return nil, errBoom }
	_, err = s.ListByProject(ctx, "p", "u")
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want errBoom", err)
	}

	p.OnGetAccess = func(context.Context, string, string) (*models.ProjectAccess, error) {
		return &models.ProjectAccess{ProjectId: "p", UserId: "u", Role: models.RoleViewer}, nil
	}
	list := []models.Collaborator{{Id: "1"}}
	c.OnListByProject = func(_ context.Context, id string) ([]models.Collaborator, error) {
		if id != "p" {
			t.Fatalf("id %s", id)
		}
		return list, nil
	}
	got, err := s.ListByProject(ctx, "p", "u")
	wantNoErr(t, err)
	if len(got) != 1 || got[0].Id != "1" {
		t.Fatalf("got %v", got)
	}
}

func TestCollaboratorGetByID(t *testing.T) {
	ctx := context.Background()
	s, c, p := collabSvc()
	_, err := s.GetByID(ctx, "", "u")
	wantErr(t, err, "collaborator id is required")
	ownerProjectAccess(p, "p", "u")
	c.OnGetByID = func(context.Context, string) (*models.Collaborator, error) {
		return &models.Collaborator{Id: "x", ProjectId: "p"}, nil
	}
	got, err := s.GetByID(ctx, "x", "u")
	wantNoErr(t, err)
	if got.Id != "x" {
		t.Fatal("wrong")
	}
}

func TestCollaboratorUpdateRole(t *testing.T) {
	ctx := context.Background()
	role := models.CollaboratorRole("viewer")
	get := func(c *fkCollaboratorRepo, col *models.Collaborator, err error) {
		c.OnGetByID = func(context.Context, string) (*models.Collaborator, error) { return col, err }
	}
	s, c, p := collabSvc()
	wantErr(t, s.UpdateRole(ctx, "", "owner", role), "collaborator id is required")
	if err := s.UpdateRole(ctx, "1", "owner", models.RoleOwner); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("err = %v, want forbidden for owner role", err)
	}

	get(c, nil, errBoom)
	err := s.UpdateRole(ctx, "1", "owner", role)
	wantErr(t, err, "failed to retrieve collaborator")
	if !errors.Is(err, errBoom) {
		t.Fatal("not wrapped")
	}
	get(c, nil, nil)
	wantErr(t, s.UpdateRole(ctx, "1", "owner", role), "collaborator does not exist")
	get(c, &models.Collaborator{UserId: "u"}, nil)
	wantErr(t, s.UpdateRole(ctx, "1", "owner", role), "collaborator projectId is required")

	// Reject self role change
	get(c, &models.Collaborator{ProjectId: "p", UserId: "owner"}, nil)
	if err := s.UpdateRole(ctx, "1", "owner", role); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("err = %v, want forbidden for self role change", err)
	}

	get(c, &models.Collaborator{ProjectId: "p", UserId: "u"}, nil)
	p.OnGetAccess = func(context.Context, string, string) (*models.ProjectAccess, error) { return nil, errBoom }
	err = s.UpdateRole(ctx, "1", "owner", role)
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want errBoom", err)
	}

	ownerProjectAccess(p, "p", "owner")
	var gotRole models.CollaboratorRole
	c.OnUpdateRole = func(_ context.Context, id string, r models.CollaboratorRole) error { gotRole = r; return errBoom }
	if err := s.UpdateRole(ctx, "1", "owner", role); !errors.Is(err, errBoom) {
		t.Fatalf("err %v", err)
	}
	c.OnUpdateRole = func(_ context.Context, id string, r models.CollaboratorRole) error { gotRole = r; return nil }
	wantNoErr(t, s.UpdateRole(ctx, "1", "owner", role))
	if gotRole != role {
		t.Fatal("role not forwarded")
	}
}

func TestCollaboratorDeleteAndLeave(t *testing.T) {
	ctx := context.Background()
	s, c, p := collabSvc()
	wantErr(t, s.Delete(ctx, "", "owner"), "collaborator id is required")
	c.OnGetByID = func(context.Context, string) (*models.Collaborator, error) { return nil, errBoom }
	err := s.Delete(ctx, "1", "owner")
	wantErr(t, err, "failed to retrieve collaborator")
	if !errors.Is(err, errBoom) {
		t.Fatal("not wrapped")
	}
	c.OnGetByID = func(context.Context, string) (*models.Collaborator, error) { return nil, nil }
	wantErr(t, s.Delete(ctx, "1", "owner"), "collaborator does not exist")

	c.OnGetByID = func(context.Context, string) (*models.Collaborator, error) {
		return &models.Collaborator{ProjectId: "p", UserId: "target"}, nil
	}
	ownerProjectAccess(p, "p", "owner")
	// Stranger cannot delete another collaborator
	if err := s.Delete(ctx, "1", "stranger"); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("err = %v, want forbidden", err)
	}
	// Self-removal succeeds even without owner role
	c.OnDelete = func(context.Context, string) error { return errBoom }
	if err := s.Delete(ctx, "1", "target"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	c.OnDelete = func(context.Context, string) error { return nil }
	wantNoErr(t, s.Delete(ctx, "1", "owner"))

	// LeaveProject rejects owner, succeeds for collaborator
	if err := s.LeaveProject(ctx, "p", "owner"); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("owner leave err = %v, want forbidden", err)
	}
	p.OnGetAccess = func(context.Context, string, string) (*models.ProjectAccess, error) {
		return &models.ProjectAccess{ProjectId: "p", UserId: "target", Role: models.RoleEditor}, nil
	}
	c.OnDeleteByProjectAndUser = func(context.Context, string, string) (bool, error) { return true, nil }
	wantNoErr(t, s.LeaveProject(ctx, "p", "target"))
}

func TestCollaboratorIsCollaborator(t *testing.T) {
	ctx := context.Background()
	s, c, _ := collabSvc()
	_, err := s.IsCollaborator(ctx, "", "p")
	wantErr(t, err, "userId is required")
	_, err = s.IsCollaborator(ctx, "u", "")
	wantErr(t, err, "projectId is required")
	c.OnIsCollaborator = func(_ context.Context, pid, uid string) (bool, error) {
		if pid != "p" || uid != "u" {
			t.Fatalf("arg order %s %s", pid, uid)
		}
		return true, nil
	}
	ok, err := s.IsCollaborator(ctx, "u", "p")
	wantNoErr(t, err)
	if !ok {
		t.Fatal("want true")
	}
}

func TestCollaboratorCheckOwnership(t *testing.T) {
	ctx := context.Background()
	s, _, p := collabSvc()
	wantErr(t, s.RequireOwnership(ctx, "", "u"), "projectId is required")
	wantErr(t, s.RequireOwnership(ctx, "p", ""), "userId is required")
	p.OnGetAccess = func(context.Context, string, string) (*models.ProjectAccess, error) { return nil, errBoom }
	if err := s.RequireOwnership(ctx, "p", "u"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	p.OnGetAccess = func(context.Context, string, string) (*models.ProjectAccess, error) {
		return nil, repositories.ErrProjectNotFound
	}
	if err := s.RequireOwnership(ctx, "p", "u"); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("err = %v, want forbidden", err)
	}
	p.OnGetAccess = func(context.Context, string, string) (*models.ProjectAccess, error) {
		return &models.ProjectAccess{ProjectId: "p", UserId: "u", Role: models.RoleEditor}, nil
	}
	if err := s.RequireOwnership(ctx, "p", "u"); !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("err = %v, want forbidden for editor", err)
	}
	ownerProjectAccess(p, "p", "u")
	wantNoErr(t, s.RequireOwnership(ctx, "p", "u"))
}
