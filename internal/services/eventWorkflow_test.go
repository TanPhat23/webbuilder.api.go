package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
)

type eventWorkflowEnv struct {
	repo *fkEventWorkflowRepo
	proj *fkProjectRepo
	svc  *EventWorkflowService
}

func newEventWorkflowEnv(projectExists bool, projErr error) *eventWorkflowEnv {
	e := &eventWorkflowEnv{repo: &fkEventWorkflowRepo{}, proj: &fkProjectRepo{}}
	e.proj.OnGetPublicByID = func(context.Context, string) (*models.Project, error) {
		if projErr != nil {
			return nil, projErr
		}
		if projectExists {
			return &models.Project{}, nil
		}
		return nil, nil
	}
	e.proj.OnGetWithAccess = func(context.Context, string, string) (*models.Project, error) {
		if projErr != nil {
			return nil, projErr
		}
		if projectExists {
			return &models.Project{}, nil
		}
		return nil, repositories.ErrProjectNotFound
	}
	e.svc = NewEventWorkflowService(e.repo, e.proj)
	return e
}

func TestEventWorkflowService_CreateEventWorkflow(t *testing.T) {
	ctx := context.Background()

	t.Run("validation", func(t *testing.T) {
		e := newEventWorkflowEnv(true, nil)
		_, err := e.svc.Create(ctx, nil, "u")
		wantErr(t, err, "workflow cannot be nil")
		_, err = e.svc.Create(ctx, &models.EventWorkflow{Name: "n"}, "u")
		wantErr(t, err, "projectId is required")
		_, err = e.svc.Create(ctx, &models.EventWorkflow{ProjectId: "p"}, "u")
		wantErr(t, err, "workflow name is required")
		wantCalls(t, e.repo.Calls)
	})
	t.Run("project not visible wraps error", func(t *testing.T) {
		e := newEventWorkflowEnv(false, nil)
		_, err := e.svc.Create(ctx, &models.EventWorkflow{ProjectId: "p", Name: "n"}, "u")
		wantErr(t, err, "failed to verify project")
		if !errors.Is(err, repositories.ErrProjectNotFound) {
			t.Fatalf("err = %v", err)
		}
		wantCalls(t, e.repo.Calls)
	})
	t.Run("defaults applied", func(t *testing.T) {
		e := newEventWorkflowEnv(true, nil)
		var got *models.EventWorkflow
		e.repo.OnCreate = func(_ context.Context, w *models.EventWorkflow) (*models.EventWorkflow, error) {
			got = w
			return w, nil
		}
		w := &models.EventWorkflow{ProjectId: "p", Name: "n"}
		_, err := e.svc.Create(ctx, w, "u")
		wantNoErr(t, err)
		if string(got.CanvasData) != "{}" || string(got.Handlers) != "[]" {
			t.Fatalf("canvas=%s handlers=%s", got.CanvasData, got.Handlers)
		}
	})
	t.Run("existing data preserved and repo error", func(t *testing.T) {
		e := newEventWorkflowEnv(true, nil)
		e.repo.OnCreate = func(_ context.Context, w *models.EventWorkflow) (*models.EventWorkflow, error) {
			if string(w.CanvasData) != "{\"a\":1}" || string(w.Handlers) != "[1]" {
				t.Errorf("data overwritten: %s %s", w.CanvasData, w.Handlers)
			}
			return nil, errBoom
		}
		w := &models.EventWorkflow{ProjectId: "p", Name: "n", CanvasData: []byte("{\"a\":1}"), Handlers: []byte("[1]")}
		_, err := e.svc.Create(ctx, w, "u")
		if !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestEventWorkflowService_GetByID(t *testing.T) {
	ctx := context.Background()
	e := newEventWorkflowEnv(true, nil)
	_, err := e.svc.GetByID(ctx, "")
	wantErr(t, err, "workflow id is required")

	e.repo.OnGetByID = func(_ context.Context, id string) (*models.EventWorkflow, error) {
		return &models.EventWorkflow{Id: id}, nil
	}
	w, err := e.svc.GetByID(ctx, "w1")
	wantNoErr(t, err)
	if w.Id != "w1" {
		t.Fatalf("w = %+v", w)
	}
}

// Methods that verify the project through GetPublicByID and then
// delegate to a single repository call.
func TestEventWorkflowService_PublicProjectGuarded(t *testing.T) {
	ctx := context.Background()
	list := []models.EventWorkflow{{Id: "a"}}

	type op struct {
		name   string
		hook   func(r *fkEventWorkflowRepo)
		call   func(s *EventWorkflowService, pid string) (int, error)
		method string
		noArg  string
	}
	listLen := func(l []models.EventWorkflow, err error) (int, error) { return len(l), err }
	ops := []op{
		{"ByProjectID", func(r *fkEventWorkflowRepo) {
			r.OnListByProject = func(context.Context, string) ([]models.EventWorkflow, error) { return list, nil }
		}, func(s *EventWorkflowService, p string) (int, error) {
			return listLen(s.ListByProject(ctx, p))
		},
			"ListByProject", "projectId is required"},
		{"WithElements", func(r *fkEventWorkflowRepo) {
			r.OnListByProjectWithElements = func(context.Context, string) ([]models.EventWorkflow, error) { return list, nil }
		}, func(s *EventWorkflowService, p string) (int, error) {
			return listLen(s.ListByProjectWithElements(ctx, p))
		}, "ListByProjectWithElements", "projectId is required"},
		{"Enabled", func(r *fkEventWorkflowRepo) {
			r.OnListEnabledByProject = func(context.Context, string) ([]models.EventWorkflow, error) { return list, nil }
		}, func(s *EventWorkflowService, p string) (int, error) {
			return listLen(s.ListEnabledByProject(ctx, p))
		}, "ListEnabledByProject", "projectId is required"},
		{"ByName", func(r *fkEventWorkflowRepo) {
			r.OnListByName = func(_ context.Context, p, n string) ([]models.EventWorkflow, error) {
				if n != "nm" {
					return nil, errors.New("bad name")
				}
				return list, nil
			}
		}, func(s *EventWorkflowService, p string) (int, error) {
			return listLen(s.ListByName(ctx, p, "nm"))
		},
			"ListByName", "projectId is required"},
		{"DeleteByProject", func(r *fkEventWorkflowRepo) {
			r.OnDeleteByProject = func(context.Context, string) error { return nil }
		}, func(s *EventWorkflowService, p string) (int, error) {
			return 0, s.DeleteByProject(ctx, p)
		},
			"DeleteByProject", "projectId is required"},
		{"Count", func(r *fkEventWorkflowRepo) {
			r.OnCountByProject = func(context.Context, string) (int64, error) { return 1, nil }
		}, func(s *EventWorkflowService, p string) (int, error) {
			n, err := s.CountByProject(ctx, p)
			return int(n), err
		}, "CountByProject", "projectId is required"},
		{"NameExists", func(r *fkEventWorkflowRepo) {
			r.OnExistsByName = func(_ context.Context, p, n, ex string) (bool, error) { return ex == "ex", nil }
		}, func(s *EventWorkflowService, p string) (int, error) {
			ok, err := s.ExistsByName(ctx, p, "nm", "ex")
			if ok {
				return 1, err
			}
			return 0, err
		}, "ExistsByName", "projectId is required"},
	}

	for _, o := range ops {
		assertPublicProjectGuardedOp(t, ctx, o.name, o.hook, o.call, o.method, o.noArg)
	}

	e := newEventWorkflowEnv(true, nil)
	_, err := e.svc.ListByName(ctx, "p", "")
	wantErr(t, err, "workflow name is required")
	_, err = e.svc.ExistsByName(ctx, "p", "", "")
	wantErr(t, err, "workflow name is required")
	assertPublicProjectGuardedErrors(t, ctx)
}

func assertPublicProjectGuardedOp(
	t *testing.T,
	ctx context.Context,
	name string,
	hook func(*fkEventWorkflowRepo),
	call func(*EventWorkflowService, string) (int, error),
	method, noArg string,
) {
	t.Helper()
	e1 := newEventWorkflowEnv(true, nil)
	_, err := call(e1.svc, "")
	wantErr(t, err, noArg)
	wantCalls(t, e1.proj.Calls)

	e2 := newEventWorkflowEnv(true, errBoom)
	_, err = call(e2.svc, "p")
	wantErr(t, err, "failed to verify project")
	if !errors.Is(err, errBoom) {
		t.Fatalf("%s: err = %v", name, err)
	}
	wantCalls(t, e2.repo.Calls)

	e3 := newEventWorkflowEnv(false, nil)
	_, err = call(e3.svc, "p")
	wantErr(t, err, "project does not exist")
	wantCalls(t, e3.repo.Calls)

	e4 := newEventWorkflowEnv(true, nil)
	hook(e4.repo)
	n, err := call(e4.svc, "p")
	wantNoErr(t, err)
	if name != "DeleteByProject" && n != 1 {
		t.Fatalf("%s: n = %d", name, n)
	}
	wantCalls(t, e4.repo.Calls, method)
}

func assertPublicProjectGuardedErrors(t *testing.T, ctx context.Context) {
	t.Helper()
	e := newEventWorkflowEnv(true, nil)
	e.repo.OnListByProject = func(context.Context, string) ([]models.EventWorkflow, error) { return nil, errBoom }
	e.repo.OnListByProjectWithElements = func(context.Context, string) ([]models.EventWorkflow, error) { return nil, errBoom }
	e.repo.OnListEnabledByProject = func(context.Context, string) ([]models.EventWorkflow, error) { return nil, errBoom }
	e.repo.OnListByName = func(context.Context, string, string) ([]models.EventWorkflow, error) { return nil, errBoom }
	e.repo.OnDeleteByProject = func(context.Context, string) error { return errBoom }
	e.repo.OnCountByProject = func(context.Context, string) (int64, error) { return 0, errBoom }
	e.repo.OnExistsByName = func(context.Context, string, string, string) (bool, error) { return false, errBoom }

	errs := []error{}
	_, err := e.svc.ListByProject(ctx, "p")
	errs = append(errs, err)
	_, err = e.svc.ListByProjectWithElements(ctx, "p")
	errs = append(errs, err)
	_, err = e.svc.ListEnabledByProject(ctx, "p")
	errs = append(errs, err)
	_, err = e.svc.ListByName(ctx, "p", "n")
	errs = append(errs, err)
	errs = append(errs, e.svc.DeleteByProject(ctx, "p"))
	_, err = e.svc.CountByProject(ctx, "p")
	errs = append(errs, err)
	_, err = e.svc.ExistsByName(ctx, "p", "n", "")
	errs = append(errs, err)
	for i, err := range errs {
		if !errors.Is(err, errBoom) {
			t.Errorf("op %d: err = %v", i, err)
		}
	}
}

func TestEventWorkflowService_UpdateValidationAndLookups(t *testing.T) {
	ctx := context.Background()
	existing := &models.EventWorkflow{Id: "w1", ProjectId: "p1", Name: "old"}
	e := newEventWorkflowEnv(true, nil)
	_, err := e.svc.Update(ctx, "", &models.EventWorkflow{Name: "n"})
	wantErr(t, err, "workflow id is required")
	_, err = e.svc.Update(ctx, "w1", nil)
	wantErr(t, err, "workflow cannot be nil")
	_, err = e.svc.Update(ctx, "w1", &models.EventWorkflow{})
	wantErr(t, err, "workflow name is required")
	wantCalls(t, e.repo.Calls)

	eErr := newEventWorkflowEnv(true, nil)
	eErr.repo.OnGetByID = func(context.Context, string) (*models.EventWorkflow, error) { return nil, errBoom }
	if _, err := eErr.svc.Update(ctx, "w1", &models.EventWorkflow{Name: "n"}); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}

	eMiss := newEventWorkflowEnv(true, nil)
	eMiss.repo.OnGetByID = func(context.Context, string) (*models.EventWorkflow, error) { return nil, nil }
	_, err = eMiss.svc.Update(ctx, "w1", &models.EventWorkflow{Name: "n"})
	wantErr(t, err, "workflow does not exist")

	eProj := newEventWorkflowEnv(true, nil)
	eProj.repo.OnGetByID = func(context.Context, string) (*models.EventWorkflow, error) { return existing, nil }
	_, err = eProj.svc.Update(ctx, "w1", &models.EventWorkflow{Name: "n", ProjectId: "other"})
	wantErr(t, err, "projectId cannot be changed")
	wantCalls(t, eProj.repo.Calls, "GetByID")
}

func TestEventWorkflowService_UpdateFieldInheritanceAndOverrides(t *testing.T) {
	ctx := context.Background()
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	desc := "desc"
	existing := &models.EventWorkflow{
		Id: "w1", ProjectId: "p1", Name: "old", Description: &desc,
		CanvasData: []byte("{\"c\":1}"), Handlers: []byte("[2]"),
	}
	existing.CreatedAt = created

	e1 := newEventWorkflowEnv(true, nil)
	e1.repo.OnGetByID = func(context.Context, string) (*models.EventWorkflow, error) { return existing, nil }
	var got *models.EventWorkflow
	e1.repo.OnUpdate = func(_ context.Context, _ string, w *models.EventWorkflow) (*models.EventWorkflow, error) {
		got = w
		return w, nil
	}
	_, err := e1.svc.Update(ctx, "w1", &models.EventWorkflow{Name: "new"})
	wantNoErr(t, err)
	assertInheritedWorkflowFields(t, got, &desc, created)

	e2 := newEventWorkflowEnv(true, nil)
	e2.repo.OnGetByID = func(context.Context, string) (*models.EventWorkflow, error) { return existing, nil }
	e2.repo.OnUpdate = func(_ context.Context, _ string, w *models.EventWorkflow) (*models.EventWorkflow, error) {
		if *w.Description != "mine" || string(w.CanvasData) != "{}" || string(w.Handlers) != "[]" || !w.CreatedAt.Equal(created.Add(time.Hour)) {
			t.Errorf("overwritten: %+v", w)
		}
		return nil, errBoom
	}
	mine := "mine"
	in := &models.EventWorkflow{Name: "n", ProjectId: "p1", Description: &mine, CanvasData: []byte("{}"), Handlers: []byte("[]")}
	in.CreatedAt = created.Add(time.Hour)
	if _, err := e2.svc.Update(ctx, "w1", in); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
}

func assertInheritedWorkflowFields(t *testing.T, got *models.EventWorkflow, desc *string, created time.Time) {
	t.Helper()
	if got.Description != desc || got.ProjectId != "p1" || got.Name != "new" {
		t.Fatalf("scalar inheritance mismatch: %+v", got)
	}
	if string(got.CanvasData) != "{\"c\":1}" || string(got.Handlers) != "[2]" || !got.CreatedAt.Equal(created) {
		t.Fatalf("payload/time inheritance mismatch: %+v", got)
	}
}

func TestEventWorkflowService_EnabledAndDelete(t *testing.T) {
	ctx := context.Background()
	type op struct {
		name   string
		call   func(s *EventWorkflowService, id string) error
		hook   func(r *fkEventWorkflowRepo, err error)
		method string
	}
	ops := []op{
		{"enabled", func(s *EventWorkflowService, id string) error { return s.SetEnabled(ctx, id, true) },
			func(r *fkEventWorkflowRepo, err error) {
				r.OnSetEnabled = func(_ context.Context, id string, en bool) error {
					if !en {
						t.Error("enabled flag lost")
					}
					return err
				}
			}, "SetEnabled"},
		{"delete", func(s *EventWorkflowService, id string) error { return s.Delete(ctx, id) },
			func(r *fkEventWorkflowRepo, err error) {
				r.OnDelete = func(context.Context, string) error { return err }
			}, "Delete"},
	}
	for _, o := range ops {
		t.Run(o.name, func(t *testing.T) {
			e := newEventWorkflowEnv(true, nil)
			wantErr(t, o.call(e.svc, ""), "workflow id is required")

			e.repo.OnGetByID = func(context.Context, string) (*models.EventWorkflow, error) { return nil, errBoom }
			if err := o.call(e.svc, "w"); !errors.Is(err, errBoom) {
				t.Fatalf("err = %v", err)
			}
			e.repo.OnGetByID = func(context.Context, string) (*models.EventWorkflow, error) { return nil, nil }
			wantErr(t, o.call(e.svc, "w"), "workflow does not exist")

			e.repo.OnGetByID = func(context.Context, string) (*models.EventWorkflow, error) {
				return &models.EventWorkflow{}, nil
			}
			e.repo.Calls = nil
			o.hook(e.repo, nil)
			wantNoErr(t, o.call(e.svc, "w"))
			wantCalls(t, e.repo.Calls, "GetByID", o.method)
			o.hook(e.repo, errBoom)
			if err := o.call(e.svc, "w"); !errors.Is(err, errBoom) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestEventWorkflowService_GetWithFilters(t *testing.T) {
	ctx := context.Background()
	e := newEventWorkflowEnv(true, nil)
	_, err := e.svc.ListWithFilters(ctx, "", "u", nil, "")
	wantErr(t, err, "projectId is required")

	hidden := newEventWorkflowEnv(false, nil)
	_, err = hidden.svc.ListWithFilters(ctx, "p", "u", nil, "")
	wantErr(t, err, "failed to verify project")

	en := true
	e.repo.OnListWithFilters = func(_ context.Context, p string, enabled *bool, q string) ([]models.EventWorkflow, error) {
		if p != "p" || enabled == nil || !*enabled || q != "srch" {
			t.Errorf("args %q %v %q", p, enabled, q)
		}
		return []models.EventWorkflow{{Id: "a"}}, nil
	}
	res, err := e.svc.ListWithFilters(ctx, "p", "u", &en, "srch")
	wantNoErr(t, err)
	if len(res) != 1 {
		t.Fatalf("res = %+v", res)
	}
	e.repo.OnListWithFilters = func(context.Context, string, *bool, string) ([]models.EventWorkflow, error) {
		return nil, errBoom
	}
	if _, err := e.svc.ListWithFilters(ctx, "p", "u", nil, ""); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
}
