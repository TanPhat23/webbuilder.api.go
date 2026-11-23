package services

import (
	"context"
	"errors"
	"testing"

	"my-go-app/internal/models"
)

func TestElementEventWorkflowServiceDelegates(t *testing.T) {
	ctx := context.Background()
	list := []models.ElementEventWorkflow{{Id: "w1"}}
	one := &models.ElementEventWorkflow{Id: "w1"}

	type tc struct {
		name  string
		setup func(r *fkElementEventWorkflowRepo, err error)
		call  func(s *ElementEventWorkflowService) (any, error)
		want  string
	}
	cases := []tc{
		{"Create", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnCreateElementEventWorkflow = func(_ context.Context, e *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
				return one, err
			}
		}, func(s *ElementEventWorkflowService) (any, error) { return s.CreateElementEventWorkflow(ctx, one) }, "CreateElementEventWorkflow"},
		{"GetByID", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnGetElementEventWorkflowByID = func(_ context.Context, id string) (*models.ElementEventWorkflow, error) {
				if id != "x" {
					t.Errorf("id = %q", id)
				}
				return one, err
			}
		}, func(s *ElementEventWorkflowService) (any, error) { return s.GetElementEventWorkflowByID(ctx, "x") }, "GetElementEventWorkflowByID"},
		{"GetAll", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnGetAllElementEventWorkflows = func(context.Context) ([]models.ElementEventWorkflow, error) { return list, err }
		}, func(s *ElementEventWorkflowService) (any, error) { return s.GetAllElementEventWorkflows(ctx) }, "GetAllElementEventWorkflows"},
		{"ByElement", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnGetElementEventWorkflowsByElementID = func(_ context.Context, id string) ([]models.ElementEventWorkflow, error) {
				if id != "e" {
					t.Errorf("id = %q", id)
				}
				return list, err
			}
		}, func(s *ElementEventWorkflowService) (any, error) {
			return s.GetElementEventWorkflowsByElementID(ctx, "e")
		}, "GetElementEventWorkflowsByElementID"},
		{"ByWorkflow", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnGetElementEventWorkflowsByWorkflowID = func(context.Context, string) ([]models.ElementEventWorkflow, error) { return list, err }
		}, func(s *ElementEventWorkflowService) (any, error) {
			return s.GetElementEventWorkflowsByWorkflowID(ctx, "w")
		}, "GetElementEventWorkflowsByWorkflowID"},
		{"ByEventName", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnGetElementEventWorkflowsByEventName = func(context.Context, string) ([]models.ElementEventWorkflow, error) { return list, err }
		}, func(s *ElementEventWorkflowService) (any, error) {
			return s.GetElementEventWorkflowsByEventName(ctx, "click")
		}, "GetElementEventWorkflowsByEventName"},
		{"ByFilters", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnGetElementEventWorkflowsByFilters = func(_ context.Context, e, w, n string) ([]models.ElementEventWorkflow, error) {
				if e != "e" || w != "w" || n != "n" {
					t.Errorf("args = %q %q %q", e, w, n)
				}
				return list, err
			}
		}, func(s *ElementEventWorkflowService) (any, error) {
			return s.GetElementEventWorkflowsByFilters(ctx, "e", "w", "n")
		}, "GetElementEventWorkflowsByFilters"},
		{"Update", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnUpdateElementEventWorkflow = func(_ context.Context, id string, e *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
				if id != "x" || e != one {
					t.Errorf("args = %q %v", id, e)
				}
				return one, err
			}
		}, func(s *ElementEventWorkflowService) (any, error) { return s.UpdateElementEventWorkflow(ctx, "x", one) }, "UpdateElementEventWorkflow"},
		{"Delete", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnDeleteElementEventWorkflow = func(context.Context, string) error { return err }
		}, func(s *ElementEventWorkflowService) (any, error) { return nil, s.DeleteElementEventWorkflow(ctx, "x") }, "DeleteElementEventWorkflow"},
		{"DeleteByElement", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnDeleteElementEventWorkflowsByElementID = func(context.Context, string) error { return err }
		}, func(s *ElementEventWorkflowService) (any, error) {
			return nil, s.DeleteElementEventWorkflowsByElementID(ctx, "e")
		}, "DeleteElementEventWorkflowsByElementID"},
		{"DeleteByWorkflow", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnDeleteElementEventWorkflowsByWorkflowID = func(context.Context, string) error { return err }
		}, func(s *ElementEventWorkflowService) (any, error) {
			return nil, s.DeleteElementEventWorkflowsByWorkflowID(ctx, "w")
		}, "DeleteElementEventWorkflowsByWorkflowID"},
		{"ByPage", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnGetElementEventWorkflowsByPageID = func(context.Context, string) ([]models.ElementEventWorkflow, error) { return list, err }
		}, func(s *ElementEventWorkflowService) (any, error) { return s.GetElementEventWorkflowsByPageID(ctx, "p") }, "GetElementEventWorkflowsByPageID"},
		{"CheckLinked", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnCheckIfWorkflowLinkedToElement = func(context.Context, string, string, string) (bool, error) { return true, err }
		}, func(s *ElementEventWorkflowService) (any, error) {
			return s.CheckIfWorkflowLinkedToElement(ctx, "e", "w", "n")
		}, "CheckIfWorkflowLinkedToElement"},
	}
	for _, c := range cases {
		t.Run(c.name+"/ok", func(t *testing.T) {
			r := &fkElementEventWorkflowRepo{}
			c.setup(r, nil)
			s := NewElementEventWorkflowService(r, nil, nil, nil)
			_, err := c.call(s)
			wantNoErr(t, err)
			wantCalls(t, r.Calls, c.want)
		})
		t.Run(c.name+"/error", func(t *testing.T) {
			r := &fkElementEventWorkflowRepo{}
			c.setup(r, errBoom)
			s := NewElementEventWorkflowService(r, nil, nil, nil)
			_, err := c.call(s)
			if !errors.Is(err, errBoom) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestElementEventWorkflowServiceReturnsRepoValues(t *testing.T) {
	r := &fkElementEventWorkflowRepo{}
	want := []models.ElementEventWorkflow{{Id: "a"}, {Id: "b"}}
	r.OnGetAllElementEventWorkflows = func(context.Context) ([]models.ElementEventWorkflow, error) { return want, nil }
	r.OnCheckIfWorkflowLinkedToElement = func(context.Context, string, string, string) (bool, error) { return true, nil }
	s := NewElementEventWorkflowService(r, nil, nil, nil)
	got, err := s.GetAllElementEventWorkflows(context.Background())
	wantNoErr(t, err)
	if len(got) != 2 || got[1].Id != "b" {
		t.Fatalf("got %v", got)
	}
	linked, err := s.CheckIfWorkflowLinkedToElement(context.Background(), "e", "w", "n")
	wantNoErr(t, err)
	if !linked {
		t.Fatal("expected linked")
	}
}
