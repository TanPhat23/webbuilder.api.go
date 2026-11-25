package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/authz"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

func snapshotValid() *models.Snapshot {
	return &models.Snapshot{ProjectId: "p1", Elements: []byte("[1]")}
}

func TestSnapshotService_ValidateAndSave(t *testing.T) {
	ctx := context.Background()
	svc := func(sr *fkSnapshotRepo) *SnapshotService { return NewSnapshotService(sr, nil, nil) }

	bad := []struct {
		name string
		s    *models.Snapshot
		msg  string
	}{
		{"nil", nil, "snapshot cannot be nil"},
		{"no project", &models.Snapshot{Elements: []byte("[1]")}, "snapshot project ID is required"},
		{"no elements", &models.Snapshot{ProjectId: "p1"}, "snapshot elements are required"},
	}
	for _, b := range bad {
		t.Run("validate "+b.name, func(t *testing.T) {
			wantErr(t, svc(&fkSnapshotRepo{}).Validate(ctx, b.s), b.msg)
			sr := &fkSnapshotRepo{}
			wantErr(t, svc(sr).Save(ctx, "p1", b.s), b.msg)
			wantCalls(t, sr.Calls)
		})
	}

	t.Run("save requires projectId", func(t *testing.T) {
		sr := &fkSnapshotRepo{}
		wantErr(t, svc(sr).Save(ctx, "", snapshotValid()), "projectId is required")
		wantCalls(t, sr.Calls)
	})
	t.Run("save ok", func(t *testing.T) {
		sr := &fkSnapshotRepo{OnSave: func(_ context.Context, pid string, s *models.Snapshot) error {
			if pid != "p9" || s.ProjectId != "p1" {
				t.Errorf("pid=%q s=%+v", pid, s)
			}
			return nil
		}}
		wantNoErr(t, svc(sr).Save(ctx, "p9", snapshotValid()))
		wantCalls(t, sr.Calls, "Save")
	})
	t.Run("save repo error", func(t *testing.T) {
		sr := &fkSnapshotRepo{OnSave: func(context.Context, string, *models.Snapshot) error { return errBoom }}
		if err := svc(sr).Save(ctx, "p1", snapshotValid()); !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestSnapshotService_GetSnapshotsByProjectID(t *testing.T) {
	ctx := context.Background()

	t.Run("empty project", func(t *testing.T) {
		_, err := NewSnapshotService(&fkSnapshotRepo{}, nil, &fkProjectRepo{}).ListByProject(ctx, "", "u")
		wantErr(t, err, "projectId is required")
	})
	t.Run("not visible", func(t *testing.T) {
		pr := &fkProjectRepo{
			OnGetAccess: func(context.Context, string, string) (*models.ProjectAccess, error) {
				return nil, repositories.ErrProjectNotFound
			},
		}
		sr := &fkSnapshotRepo{}
		_, err := NewSnapshotService(sr, nil, pr).ListByProject(ctx, "p1", "u")
		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("err = %v, want authz.ErrForbidden", err)
		}
		wantCalls(t, sr.Calls)
	})
	t.Run("ok", func(t *testing.T) {
		pr := &fkProjectRepo{OnGetAccess: func(context.Context, string, string) (*models.ProjectAccess, error) {
			return &models.ProjectAccess{ProjectId: "p1", UserId: "u", Role: models.RoleViewer}, nil
		}}
		sr := &fkSnapshotRepo{OnListByProject: func(_ context.Context, pid string) ([]models.Snapshot, error) {
			if pid != "p1" {
				t.Errorf("pid = %q", pid)
			}
			return []models.Snapshot{{Id: "s1"}}, nil
		}}
		res, err := NewSnapshotService(sr, nil, pr).ListByProject(ctx, "p1", "u")
		wantNoErr(t, err)
		if len(res) != 1 || res[0].Id != "s1" {
			t.Fatalf("res = %+v", res)
		}
	})
}

func TestSnapshotService_GetAndDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("get empty id", func(t *testing.T) {
		_, err := NewSnapshotService(&fkSnapshotRepo{}, nil, nil).GetByID(ctx, "")
		wantErr(t, err, "snapshot id is required")
	})
	t.Run("get ok and error", func(t *testing.T) {
		sr := &fkSnapshotRepo{OnGetByID: func(_ context.Context, id string) (*models.Snapshot, error) {
			return &models.Snapshot{Id: id}, nil
		}}
		s, err := NewSnapshotService(sr, nil, nil).GetByID(ctx, "s1")
		wantNoErr(t, err)
		if s.Id != "s1" {
			t.Fatalf("s = %+v", s)
		}
		sr.OnGetByID = func(context.Context, string) (*models.Snapshot, error) { return nil, errBoom }
		if _, err := NewSnapshotService(sr, nil, nil).GetByID(ctx, "s1"); !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("delete", func(t *testing.T) {
		wantErr(t, NewSnapshotService(&fkSnapshotRepo{}, nil, nil).Delete(ctx, ""), "snapshot id is required")
		sr := &fkSnapshotRepo{OnDelete: func(_ context.Context, id string) error {
			if id != "s1" {
				t.Errorf("id = %q", id)
			}
			return nil
		}}
		wantNoErr(t, NewSnapshotService(sr, nil, nil).Delete(ctx, "s1"))
		sr.OnDelete = func(context.Context, string) error { return errBoom }
		if err := NewSnapshotService(sr, nil, nil).Delete(ctx, "s1"); !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestSnapshotService_DeleteSnapshotWithAccess(t *testing.T) {
	ctx := context.Background()
	access := func(role models.CollaboratorRole, err error) *fkProjectRepo {
		return &fkProjectRepo{OnGetAccess: func(context.Context, string, string) (*models.ProjectAccess, error) {
			if err != nil {
				return nil, err
			}
			return &models.ProjectAccess{ProjectId: "p", UserId: "u", Role: role, IsOwner: role == models.RoleOwner}, nil
		}}
	}
	found := func(pid string) *fkSnapshotRepo {
		return &fkSnapshotRepo{
			OnGetByID: func(context.Context, string) (*models.Snapshot, error) {
				return &models.Snapshot{ProjectId: pid}, nil
			},
			OnDelete: func(context.Context, string) error { return nil },
		}
	}

	argErrs := []struct{ name, sid, pid, uid, msg string }{
		{"snapshot id", "", "p", "u", "snapshot id is required"},
		{"project id", "s", "", "u", "projectId is required"},
		{"user id", "s", "p", "", "userId is required"},
	}
	for _, a := range argErrs {
		t.Run("missing "+a.name, func(t *testing.T) {
			sr := &fkSnapshotRepo{}
			wantErr(t, NewSnapshotService(sr, nil, &fkProjectRepo{}).DeleteForUser(ctx, a.sid, a.pid, a.uid), a.msg)
			wantCalls(t, sr.Calls)
		})
	}

	t.Run("lookup error", func(t *testing.T) {
		sr := &fkSnapshotRepo{OnGetByID: func(context.Context, string) (*models.Snapshot, error) { return nil, errBoom }}
		if err := NewSnapshotService(sr, nil, &fkProjectRepo{}).DeleteForUser(ctx, "s", "p", "u"); !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("snapshot missing", func(t *testing.T) {
		sr := &fkSnapshotRepo{OnGetByID: func(context.Context, string) (*models.Snapshot, error) { return nil, nil }}
		wantErr(t, NewSnapshotService(sr, nil, &fkProjectRepo{}).DeleteForUser(ctx, "s", "p", "u"), "snapshot does not exist")
	})
	t.Run("wrong project", func(t *testing.T) {
		sr := found("other")
		wantErr(t, NewSnapshotService(sr, nil, &fkProjectRepo{}).DeleteForUser(ctx, "s", "p", "u"), "does not belong")
		wantCalls(t, sr.Calls, "GetByID")
	})
	t.Run("access denied", func(t *testing.T) {
		sr := found("p")
		err := NewSnapshotService(sr, nil, access(models.RoleEditor, nil)).DeleteForUser(ctx, "s", "p", "u")
		if !errors.Is(err, authz.ErrForbidden) {
			t.Fatalf("err = %v, want authz.ErrForbidden", err)
		}
		wantCalls(t, sr.Calls, "GetByID")
	})
	t.Run("deleted", func(t *testing.T) {
		sr := found("p")
		wantNoErr(t, NewSnapshotService(sr, nil, access(models.RoleOwner, nil)).DeleteForUser(ctx, "s", "p", "u"))
		wantCalls(t, sr.Calls, "GetByID", "Delete")
	})
}
