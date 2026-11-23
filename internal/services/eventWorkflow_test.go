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
	e.proj.OnGetPublicProjectByID = func(context.Context, string) (*models.Project, error) {
		if projErr != nil {
			return nil, projErr
		}
		if projectExists {
			return &models.Project{}, nil
		}
		return nil, nil
	}
	e.proj.OnGetProjectWithAccess = func(context.Context, string, string) (*models.Project, error) {
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
		_, err := e.svc.CreateEventWorkflow(ctx, nil, "u")
		wantErr(t, err, "workflow cannot be nil")
		_, err = e.svc.CreateEventWorkflow(ctx, &models.EventWorkflow{Name: "n"}, "u")
		wantErr(t, err, "projectId is required")
		_, err = e.svc.CreateEventWorkflow(ctx, &models.EventWorkflow{ProjectId: "p"}, "u")
		wantErr(t, err, "workflow name is required")
		wantCalls(t, e.repo.Calls)
	})
	t.Run("project not visible wraps error", func(t *testing.T) {
		e := newEventWorkflowEnv(false, nil)
		_, err := e.svc.CreateEventWorkflow(ctx, &models.EventWorkflow{ProjectId: "p", Name: "n"}, "u")
		wantErr(t, err, "failed to verify project")
		if !errors.Is(err, repositories.ErrProjectNotFound) {
			t.Fatalf("err = %v", err)
		}
		wantCalls(t, e.repo.Calls)
	})
	t.Run("defaults applied", func(t *testing.T) {
		e := newEventWorkflowEnv(true, nil)
		var got *models.EventWorkflow
		e.repo.OnCreateEventWorkflow = func(_ context.Context, w *models.EventWorkflow) (*models.EventWorkflow, error) {
			got = w
			return w, nil
		}
		w := &models.EventWorkflow{ProjectId: "p", Name: "n"}
		_, err := e.svc.CreateEventWorkflow(ctx, w, "u")
		wantNoErr(t, err)
		if string(got.CanvasData) != "{}" || string(got.Handlers) != "[]" {
			t.Fatalf("canvas=%s handlers=%s", got.CanvasData, got.Handlers)
		}
	})
	t.Run("existing data preserved and repo error", func(t *testing.T) {
		e := newEventWorkflowEnv(true, nil)
		e.repo.OnCreateEventWorkflow = func(_ context.Context, w *models.EventWorkflow) (*models.EventWorkflow, error) {
			if string(w.CanvasData) != "{\"a\":1}" || string(w.Handlers) != "[1]" {
				t.Errorf("data overwritten: %s %s", w.CanvasData, w.Handlers)
			}
			return nil, errBoom
		}
		w := &models.EventWorkflow{ProjectId: "p", Name: "n", CanvasData: []byte("{\"a\":1}"), Handlers: []byte("[1]")}
		_, err := e.svc.CreateEventWorkflow(ctx, w, "u")
		if !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestEventWorkflowService_GetByID(t *testing.T) {
	ctx := context.Background()
	e := newEventWorkflowEnv(true, nil)
	_, err := e.svc.GetEventWorkflowByID(ctx, "")
	wantErr(t, err, "workflow id is required")

	e.repo.OnGetEventWorkflowByID = func(_ context.Context, id string) (*models.EventWorkflow, error) {
		return &models.EventWorkflow{Id: id}, nil
	}
	w, err := e.svc.GetEventWorkflowByID(ctx, "w1")
	wantNoErr(t, err)
	if w.Id != "w1" {
		t.Fatalf("w = %+v", w)
	}
}

// Methods that verify the project through GetPublicProjectByID and then
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
			r.OnGetEventWorkflowsByProjectID = func(context.Context, string) ([]models.EventWorkflow, error) { return list, nil }
		}, func(s *EventWorkflowService, p string) (int, error) {
			return listLen(s.GetEventWorkflowsByProjectID(ctx, p))
		},
			"GetEventWorkflowsByProjectID", "projectId is required"},
		{"WithElements", func(r *fkEventWorkflowRepo) {
			r.OnGetEventWorkflowsByProjectIDWithElements = func(context.Context, string) ([]models.EventWorkflow, error) { return list, nil }
		}, func(s *EventWorkflowService, p string) (int, error) {
			return listLen(s.GetEventWorkflowsByProjectIDWithElements(ctx, p))
		}, "GetEventWorkflowsByProjectIDWithElements", "projectId is required"},
		{"Enabled", func(r *fkEventWorkflowRepo) {
			r.OnGetEnabledEventWorkflowsByProjectID = func(context.Context, string) ([]models.EventWorkflow, error) { return list, nil }
		}, func(s *EventWorkflowService, p string) (int, error) {
			return listLen(s.GetEnabledEventWorkflowsByProjectID(ctx, p))
		}, "GetEnabledEventWorkflowsByProjectID", "projectId is required"},
		{"ByName", func(r *fkEventWorkflowRepo) {
			r.OnGetEventWorkflowsByName = func(_ context.Context, p, n string) ([]models.EventWorkflow, error) {
				if n != "nm" {
					return nil, errors.New("bad name")
				}
				return list, nil
			}
		}, func(s *EventWorkflowService, p string) (int, error) {
			return listLen(s.GetEventWorkflowsByName(ctx, p, "nm"))
		},
			"GetEventWorkflowsByName", "projectId is required"},
		{"DeleteByProject", func(r *fkEventWorkflowRepo) {
			r.OnDeleteEventWorkflowsByProjectID = func(context.Context, string) error { return nil }
		}, func(s *EventWorkflowService, p string) (int, error) {
			return 0, s.DeleteEventWorkflowsByProjectID(ctx, p)
		},
			"DeleteEventWorkflowsByProjectID", "projectId is required"},
		{"Count", func(r *fkEventWorkflowRepo) {
			r.OnCountEventWorkflowsByProjectID = func(context.Context, string) (int64, error) { return 1, nil }
		}, func(s *EventWorkflowService, p string) (int, error) {
			n, err := s.CountEventWorkflowsByProjectID(ctx, p)
			return int(n), err
		}, "CountEventWorkflowsByProjectID", "projectId is required"},
		{"NameExists", func(r *fkEventWorkflowRepo) {
			r.OnCheckIfWorkflowNameExists = func(_ context.Context, p, n, ex string) (bool, error) { return ex == "ex", nil }
		}, func(s *EventWorkflowService, p string) (int, error) {
			ok, err := s.CheckIfWorkflowNameExists(ctx, p, "nm", "ex")
			if ok {
				return 1, err
			}
			return 0, err
		}, "CheckIfWorkflowNameExists", "projectId is required"},
	}

	for _, o := range ops {
		t.Run(o.name, func(t *testing.T) {
			t.Run("missing project id", func(t *testing.T) {
				e := newEventWorkflowEnv(true, nil)
				_, err := o.call(e.svc, "")
				wantErr(t, err, o.noArg)
				wantCalls(t, e.proj.Calls)
			})
			t.Run("project lookup error wrapped", func(t *testing.T) {
				e := newEventWorkflowEnv(true, errBoom)
				_, err := o.call(e.svc, "p")
				wantErr(t, err, "failed to verify project")
				if !errors.Is(err, errBoom) {
					t.Fatalf("err = %v", err)
				}
				wantCalls(t, e.repo.Calls)
			})
			t.Run("project missing", func(t *testing.T) {
				e := newEventWorkflowEnv(false, nil)
				_, err := o.call(e.svc, "p")
				wantErr(t, err, "project does not exist")
				wantCalls(t, e.repo.Calls)
			})
			t.Run("delegates", func(t *testing.T) {
				e := newEventWorkflowEnv(true, nil)
				o.hook(e.repo)
				n, err := o.call(e.svc, "p")
				wantNoErr(t, err)
				if o.name != "DeleteByProject" && n != 1 {
					t.Fatalf("n = %d", n)
				}
				wantCalls(t, e.repo.Calls, o.method)
			})
		})
	}

	t.Run("ByName and NameExists require name", func(t *testing.T) {
		e := newEventWorkflowEnv(true, nil)
		_, err := e.svc.GetEventWorkflowsByName(ctx, "p", "")
		wantErr(t, err, "workflow name is required")
		_, err = e.svc.CheckIfWorkflowNameExists(ctx, "p", "", "")
		wantErr(t, err, "workflow name is required")
	})
	t.Run("repo errors propagate", func(t *testing.T) {
		e := newEventWorkflowEnv(true, nil)
		e.repo.OnGetEventWorkflowsByProjectID = func(context.Context, string) ([]models.EventWorkflow, error) { return nil, errBoom }
		e.repo.OnGetEventWorkflowsByProjectIDWithElements = func(context.Context, string) ([]models.EventWorkflow, error) { return nil, errBoom }
		e.repo.OnGetEnabledEventWorkflowsByProjectID = func(context.Context, string) ([]models.EventWorkflow, error) { return nil, errBoom }
		e.repo.OnGetEventWorkflowsByName = func(context.Context, string, string) ([]models.EventWorkflow, error) { return nil, errBoom }
		e.repo.OnDeleteEventWorkflowsByProjectID = func(context.Context, string) error { return errBoom }
		e.repo.OnCountEventWorkflowsByProjectID = func(context.Context, string) (int64, error) { return 0, errBoom }
		e.repo.OnCheckIfWorkflowNameExists = func(context.Context, string, string, string) (bool, error) { return false, errBoom }
		errs := []error{}
		_, err := e.svc.GetEventWorkflowsByProjectID(ctx, "p")
		errs = append(errs, err)
		_, err = e.svc.GetEventWorkflowsByProjectIDWithElements(ctx, "p")
		errs = append(errs, err)
		_, err = e.svc.GetEnabledEventWorkflowsByProjectID(ctx, "p")
		errs = append(errs, err)
		_, err = e.svc.GetEventWorkflowsByName(ctx, "p", "n")
		errs = append(errs, err)
		errs = append(errs, e.svc.DeleteEventWorkflowsByProjectID(ctx, "p"))
		_, err = e.svc.CountEventWorkflowsByProjectID(ctx, "p")
		errs = append(errs, err)
		_, err = e.svc.CheckIfWorkflowNameExists(ctx, "p", "n", "")
		errs = append(errs, err)
		for i, err := range errs {
			if !errors.Is(err, errBoom) {
				t.Errorf("op %d: err = %v", i, err)
			}
		}
	})
}

func TestEventWorkflowService_Update(t *testing.T) {
	ctx := context.Background()
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	desc := "desc"
	existing := &models.EventWorkflow{
		Id: "w1", ProjectId: "p1", Name: "old", Description: &desc,
		CanvasData: []byte("{\"c\":1}"), Handlers: []byte("[2]"),
	}
	existing.CreatedAt = created

	t.Run("validation", func(t *testing.T) {
		e := newEventWorkflowEnv(true, nil)
		_, err := e.svc.UpdateEventWorkflow(ctx, "", &models.EventWorkflow{Name: "n"})
		wantErr(t, err, "workflow id is required")
		_, err = e.svc.UpdateEventWorkflow(ctx, "w1", nil)
		wantErr(t, err, "workflow cannot be nil")
		_, err = e.svc.UpdateEventWorkflow(ctx, "w1", &models.EventWorkflow{})
		wantErr(t, err, "workflow name is required")
		wantCalls(t, e.repo.Calls)
	})
	t.Run("lookup error", func(t *testing.T) {
		e := newEventWorkflowEnv(true, nil)
		e.repo.OnGetEventWorkflowByID = func(context.Context, string) (*models.EventWorkflow, error) { return nil, errBoom }
		_, err := e.svc.UpdateEventWorkflow(ctx, "w1", &models.EventWorkflow{Name: "n"})
		if !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("not found", func(t *testing.T) {
		e := newEventWorkflowEnv(true, nil)
		e.repo.OnGetEventWorkflowByID = func(context.Context, string) (*models.EventWorkflow, error) { return nil, nil }
		_, err := e.svc.UpdateEventWorkflow(ctx, "w1", &models.EventWorkflow{Name: "n"})
		wantErr(t, err, "workflow does not exist")
	})
	t.Run("project change rejected", func(t *testing.T) {
		e := newEventWorkflowEnv(true, nil)
		e.repo.OnGetEventWorkflowByID = func(context.Context, string) (*models.EventWorkflow, error) { return existing, nil }
		_, err := e.svc.UpdateEventWorkflow(ctx, "w1", &models.EventWorkflow{Name: "n", ProjectId: "other"})
		wantErr(t, err, "projectId cannot be changed")
		wantCalls(t, e.repo.Calls, "GetEventWorkflowByID")
	})
	t.Run("missing fields inherited", func(t *testing.T) {
		e := newEventWorkflowEnv(true, nil)
		e.repo.OnGetEventWorkflowByID = func(context.Context, string) (*models.EventWorkflow, error) { return existing, nil }
		var got *models.EventWorkflow
		e.repo.OnUpdateEventWorkflow = func(_ context.Context, id string, w *models.EventWorkflow) (*models.EventWorkflow, error) {
			got = w
			return w, nil
		}
		_, err := e.svc.UpdateEventWorkflow(ctx, "w1", &models.EventWorkflow{Name: "new"})
		wantNoErr(t, err)
		if got.Description != &desc || string(got.CanvasData) != "{\"c\":1}" || string(got.Handlers) != "[2]" ||
			got.ProjectId != "p1" || !got.CreatedAt.Equal(created) || got.Name != "new" {
			t.Fatalf("got = %+v", got)
		}
	})
	t.Run("provided fields kept and repo error", func(t *testing.T) {
		e := newEventWorkflowEnv(true, nil)
		e.repo.OnGetEventWorkflowByID = func(context.Context, string) (*models.EventWorkflow, error) { return existing, nil }
		e.repo.OnUpdateEventWorkflow = func(_ context.Context, id string, w *models.EventWorkflow) (*models.EventWorkflow, error) {
			if *w.Description != "mine" || string(w.CanvasData) != "{}" || string(w.Handlers) != "[]" || !w.CreatedAt.Equal(created.Add(time.Hour)) {
				t.Errorf("overwritten: %+v", w)
			}
			return nil, errBoom
		}
		mine := "mine"
		in := &models.EventWorkflow{Name: "n", ProjectId: "p1", Description: &mine, CanvasData: []byte("{}"), Handlers: []byte("[]")}
		in.CreatedAt = created.Add(time.Hour)
		_, err := e.svc.UpdateEventWorkflow(ctx, "w1", in)
		if !errors.Is(err, errBoom) {
			t.Fatalf("err = %v", err)
		}
	})
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
		{"enabled", func(s *EventWorkflowService, id string) error { return s.UpdateEventWorkflowEnabled(ctx, id, true) },
			func(r *fkEventWorkflowRepo, err error) {
				r.OnUpdateEventWorkflowEnabled = func(_ context.Context, id string, en bool) error {
					if !en {
						t.Error("enabled flag lost")
					}
					return err
				}
			}, "UpdateEventWorkflowEnabled"},
		{"delete", func(s *EventWorkflowService, id string) error { return s.DeleteEventWorkflow(ctx, id) },
			func(r *fkEventWorkflowRepo, err error) {
				r.OnDeleteEventWorkflow = func(context.Context, string) error { return err }
			}, "DeleteEventWorkflow"},
	}
	for _, o := range ops {
		t.Run(o.name, func(t *testing.T) {
			e := newEventWorkflowEnv(true, nil)
			wantErr(t, o.call(e.svc, ""), "workflow id is required")

			e.repo.OnGetEventWorkflowByID = func(context.Context, string) (*models.EventWorkflow, error) { return nil, errBoom }
			if err := o.call(e.svc, "w"); !errors.Is(err, errBoom) {
				t.Fatalf("err = %v", err)
			}
			e.repo.OnGetEventWorkflowByID = func(context.Context, string) (*models.EventWorkflow, error) { return nil, nil }
			wantErr(t, o.call(e.svc, "w"), "workflow does not exist")

			e.repo.OnGetEventWorkflowByID = func(context.Context, string) (*models.EventWorkflow, error) {
				return &models.EventWorkflow{}, nil
			}
			e.repo.Calls = nil
			o.hook(e.repo, nil)
			wantNoErr(t, o.call(e.svc, "w"))
			wantCalls(t, e.repo.Calls, "GetEventWorkflowByID", o.method)
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
	_, err := e.svc.GetEventWorkflowsWithFilters(ctx, "", "u", nil, "")
	wantErr(t, err, "projectId is required")

	hidden := newEventWorkflowEnv(false, nil)
	_, err = hidden.svc.GetEventWorkflowsWithFilters(ctx, "p", "u", nil, "")
	wantErr(t, err, "failed to verify project")

	en := true
	e.repo.OnGetEventWorkflowsWithFilters = func(_ context.Context, p string, enabled *bool, q string) ([]models.EventWorkflow, error) {
		if p != "p" || enabled == nil || !*enabled || q != "srch" {
			t.Errorf("args %q %v %q", p, enabled, q)
		}
		return []models.EventWorkflow{{Id: "a"}}, nil
	}
	res, err := e.svc.GetEventWorkflowsWithFilters(ctx, "p", "u", &en, "srch")
	wantNoErr(t, err)
	if len(res) != 1 {
		t.Fatalf("res = %+v", res)
	}
	e.repo.OnGetEventWorkflowsWithFilters = func(context.Context, string, *bool, string) ([]models.EventWorkflow, error) {
		return nil, errBoom
	}
	if _, err := e.svc.GetEventWorkflowsWithFilters(ctx, "p", "u", nil, ""); !errors.Is(err, errBoom) {
		t.Fatalf("err = %v", err)
	}
}
