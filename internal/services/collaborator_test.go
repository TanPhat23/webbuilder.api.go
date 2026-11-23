package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

func collabSvc() (*CollaboratorService, *fkCollaboratorRepo, *fkProjectRepo) {
	c, p := &fkCollaboratorRepo{}, &fkProjectRepo{}
	return NewCollaboratorService(c, p), c, p
}

func TestCollaboratorCreate(t *testing.T) {
	ctx := context.Background()
	valid := func() *models.Collaborator { return &models.Collaborator{ProjectId: "p", UserId: "u"} }
	for _, tc := range []struct {
		name string
		in   *models.Collaborator
		msg  string
	}{
		{"nil", nil, "cannot be nil"},
		{"no project", &models.Collaborator{UserId: "u"}, "projectId is required"},
		{"no user", &models.Collaborator{ProjectId: "p"}, "userId is required"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, c, p := collabSvc()
			_, err := s.CreateCollaborator(ctx, tc.in)
			wantErr(t, err, tc.msg)
			wantCalls(t, c.Calls)
			wantCalls(t, p.Calls)
		})
	}
	t.Run("project lookup error wrapped", func(t *testing.T) {
		s, _, p := collabSvc()
		p.OnGetPublicProjectByID = func(context.Context, string) (*models.Project, error) { return nil, errBoom }
		_, err := s.CreateCollaborator(ctx, valid())
		wantErr(t, err, "failed to verify project")
		if !errors.Is(err, errBoom) {
			t.Fatal("not wrapped")
		}
	})
	t.Run("project missing", func(t *testing.T) {
		s, c, p := collabSvc()
		p.OnGetPublicProjectByID = func(context.Context, string) (*models.Project, error) { return nil, nil }
		_, err := s.CreateCollaborator(ctx, valid())
		wantErr(t, err, "project does not exist")
		wantCalls(t, c.Calls)
	})
	t.Run("success and repo error", func(t *testing.T) {
		s, c, p := collabSvc()
		p.OnGetPublicProjectByID = func(_ context.Context, id string) (*models.Project, error) {
			if id != "p" {
				t.Fatalf("id %s", id)
			}
			return &models.Project{}, nil
		}
		want := &models.Collaborator{Id: "c"}
		c.OnCreateCollaborator = func(context.Context, *models.Collaborator) (*models.Collaborator, error) { return want, nil }
		got, err := s.CreateCollaborator(ctx, valid())
		wantNoErr(t, err)
		if got != want {
			t.Fatal("wrong result")
		}
		c.OnCreateCollaborator = func(context.Context, *models.Collaborator) (*models.Collaborator, error) { return nil, errBoom }
		if _, err := s.CreateCollaborator(ctx, valid()); !errors.Is(err, errBoom) {
			t.Fatalf("err %v", err)
		}
	})
}

func TestCollaboratorGetByProject(t *testing.T) {
	ctx := context.Background()
	s, c, p := collabSvc()
	_, err := s.GetCollaboratorsByProject(ctx, "", "u")
	wantErr(t, err, "projectId is required")

	p.OnGetProjectWithAccess = func(context.Context, string, string) (*models.Project, error) { return nil, errBoom }
	_, err = s.GetCollaboratorsByProject(ctx, "p", "u")
	wantErr(t, err, "failed to verify project")
	if !errors.Is(err, errBoom) {
		t.Fatal("not wrapped")
	}

	p.OnGetProjectWithAccess = func(context.Context, string, string) (*models.Project, error) { return &models.Project{}, nil }
	list := []models.Collaborator{{Id: "1"}}
	c.OnGetCollaboratorsByProject = func(_ context.Context, id string) ([]models.Collaborator, error) {
		if id != "p" {
			t.Fatalf("id %s", id)
		}
		return list, nil
	}
	got, err := s.GetCollaboratorsByProject(ctx, "p", "u")
	wantNoErr(t, err)
	if len(got) != 1 || got[0].Id != "1" {
		t.Fatalf("got %v", got)
	}
}

func TestCollaboratorGetByID(t *testing.T) {
	ctx := context.Background()
	s, c, _ := collabSvc()
	_, err := s.GetCollaboratorByID(ctx, "")
	wantErr(t, err, "collaborator id is required")
	c.OnGetCollaboratorByID = func(context.Context, string) (*models.Collaborator, error) { return &models.Collaborator{Id: "x"}, nil }
	got, err := s.GetCollaboratorByID(ctx, "x")
	wantNoErr(t, err)
	if got.Id != "x" {
		t.Fatal("wrong")
	}
}

func TestCollaboratorUpdateRole(t *testing.T) {
	ctx := context.Background()
	role := models.CollaboratorRole("viewer")
	get := func(c *fkCollaboratorRepo, col *models.Collaborator, err error) {
		c.OnGetCollaboratorByID = func(context.Context, string) (*models.Collaborator, error) { return col, err }
	}
	s, c, p := collabSvc()
	wantErr(t, s.UpdateCollaboratorRole(ctx, "", role), "collaborator id is required")

	get(c, nil, errBoom)
	err := s.UpdateCollaboratorRole(ctx, "1", role)
	wantErr(t, err, "failed to retrieve collaborator")
	if !errors.Is(err, errBoom) {
		t.Fatal("not wrapped")
	}
	get(c, nil, nil)
	wantErr(t, s.UpdateCollaboratorRole(ctx, "1", role), "collaborator does not exist")
	get(c, &models.Collaborator{UserId: "u"}, nil)
	wantErr(t, s.UpdateCollaboratorRole(ctx, "1", role), "collaborator projectId is required")

	get(c, &models.Collaborator{ProjectId: "p", UserId: "u"}, nil)
	p.OnGetProjectByID = func(context.Context, string, string) (*models.Project, error) { return nil, errBoom }
	err = s.UpdateCollaboratorRole(ctx, "1", role)
	wantErr(t, err, "failed to verify project ownership")
	if !errors.Is(err, errBoom) {
		t.Fatal("not wrapped")
	}
	p.OnGetProjectByID = func(context.Context, string, string) (*models.Project, error) { return nil, nil }
	wantErr(t, s.UpdateCollaboratorRole(ctx, "1", role), "project does not exist")

	p.OnGetProjectByID = func(_ context.Context, pid, uid string) (*models.Project, error) {
		if pid != "p" || uid != "u" {
			t.Fatalf("args %s %s", pid, uid)
		}
		return &models.Project{}, nil
	}
	var gotRole models.CollaboratorRole
	c.OnUpdateCollaboratorRole = func(_ context.Context, id string, r models.CollaboratorRole) error { gotRole = r; return errBoom }
	if err := s.UpdateCollaboratorRole(ctx, "1", role); !errors.Is(err, errBoom) {
		t.Fatalf("err %v", err)
	}
	c.OnUpdateCollaboratorRole = func(_ context.Context, id string, r models.CollaboratorRole) error { gotRole = r; return nil }
	wantNoErr(t, s.UpdateCollaboratorRole(ctx, "1", role))
	if gotRole != role {
		t.Fatal("role not forwarded")
	}
}

func TestCollaboratorDelete(t *testing.T) {
	ctx := context.Background()
	s, c, _ := collabSvc()
	wantErr(t, s.DeleteCollaborator(ctx, ""), "collaborator id is required")
	c.OnGetCollaboratorByID = func(context.Context, string) (*models.Collaborator, error) { return nil, errBoom }
	err := s.DeleteCollaborator(ctx, "1")
	wantErr(t, err, "failed to retrieve collaborator")
	if !errors.Is(err, errBoom) {
		t.Fatal("not wrapped")
	}
	c.OnGetCollaboratorByID = func(context.Context, string) (*models.Collaborator, error) { return nil, nil }
	wantErr(t, s.DeleteCollaborator(ctx, "1"), "collaborator does not exist")
	c.OnGetCollaboratorByID = func(context.Context, string) (*models.Collaborator, error) { return &models.Collaborator{}, nil }
	c.OnDeleteCollaborator = func(context.Context, string) error { return errBoom }
	if err := s.DeleteCollaborator(ctx, "1"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	c.OnDeleteCollaborator = func(context.Context, string) error { return nil }
	wantNoErr(t, s.DeleteCollaborator(ctx, "1"))
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
	wantErr(t, s.CheckOwnership(ctx, "", "u"), "projectId is required")
	wantErr(t, s.CheckOwnership(ctx, "p", ""), "userId is required")
	p.OnGetProjectByID = func(context.Context, string, string) (*models.Project, error) { return nil, errBoom }
	if err := s.CheckOwnership(ctx, "p", "u"); !errors.Is(err, errBoom) {
		t.Fatal(err)
	}
	p.OnGetProjectByID = func(context.Context, string, string) (*models.Project, error) { return nil, nil }
	wantErr(t, s.CheckOwnership(ctx, "p", "u"), "project does not exist")
	p.OnGetProjectByID = func(context.Context, string, string) (*models.Project, error) {
		return &models.Project{OwnerId: "other"}, nil
	}
	wantErr(t, s.CheckOwnership(ctx, "p", "u"), "not the owner")
	p.OnGetProjectByID = func(context.Context, string, string) (*models.Project, error) {
		return &models.Project{OwnerId: "u"}, nil
	}
	wantNoErr(t, s.CheckOwnership(ctx, "p", "u"))
	_ = repositories.ErrProjectNotFound
}
