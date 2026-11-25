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
			r.OnCreate = func(_ context.Context, e *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
				return one, err
			}
		}, func(s *ElementEventWorkflowService) (any, error) { return s.Create(ctx, one) }, "Create"},
		{"GetByID", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnGetByID = func(_ context.Context, id string) (*models.ElementEventWorkflow, error) {
				if id != "x" {
					t.Errorf("id = %q", id)
				}
				return one, err
			}
		}, func(s *ElementEventWorkflowService) (any, error) { return s.GetByID(ctx, "x") }, "GetByID"},
		{"GetAll", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnListAll = func(context.Context) ([]models.ElementEventWorkflow, error) { return list, err }
		}, func(s *ElementEventWorkflowService) (any, error) { return s.ListAll(ctx) }, "ListAll"},
		{"ByElement", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnListByElement = func(_ context.Context, id string) ([]models.ElementEventWorkflow, error) {
				if id != "e" {
					t.Errorf("id = %q", id)
				}
				return list, err
			}
		}, func(s *ElementEventWorkflowService) (any, error) {
			return s.ListByElement(ctx, "e")
		}, "ListByElement"},
		{"ByWorkflow", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnListByWorkflow = func(context.Context, string) ([]models.ElementEventWorkflow, error) { return list, err }
		}, func(s *ElementEventWorkflowService) (any, error) {
			return s.ListByWorkflow(ctx, "w")
		}, "ListByWorkflow"},
		{"ByEventName", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnListByEvent = func(context.Context, string) ([]models.ElementEventWorkflow, error) { return list, err }
		}, func(s *ElementEventWorkflowService) (any, error) {
			return s.ListByEvent(ctx, "click")
		}, "ListByEvent"},
		{"ByFilters", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnListByFilters = func(_ context.Context, e, w, n string) ([]models.ElementEventWorkflow, error) {
				if e != "e" || w != "w" || n != "n" {
					t.Errorf("args = %q %q %q", e, w, n)
				}
				return list, err
			}
		}, func(s *ElementEventWorkflowService) (any, error) {
			return s.ListByFilters(ctx, "e", "w", "n")
		}, "ListByFilters"},
		{"Update", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnUpdate = func(_ context.Context, id string, e *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
				if id != "x" || e != one {
					t.Errorf("args = %q %v", id, e)
				}
				return one, err
			}
		}, func(s *ElementEventWorkflowService) (any, error) { return s.Update(ctx, "x", one) }, "Update"},
		{"Delete", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnDelete = func(context.Context, string) error { return err }
		}, func(s *ElementEventWorkflowService) (any, error) { return nil, s.Delete(ctx, "x") }, "Delete"},
		{"DeleteByElement", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnDeleteByElement = func(context.Context, string) error { return err }
		}, func(s *ElementEventWorkflowService) (any, error) {
			return nil, s.DeleteByElement(ctx, "e")
		}, "DeleteByElement"},
		{"DeleteByWorkflow", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnDeleteByWorkflow = func(context.Context, string) error { return err }
		}, func(s *ElementEventWorkflowService) (any, error) {
			return nil, s.DeleteByWorkflow(ctx, "w")
		}, "DeleteByWorkflow"},
		{"ByPage", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnListByPage = func(context.Context, string) ([]models.ElementEventWorkflow, error) { return list, err }
		}, func(s *ElementEventWorkflowService) (any, error) { return s.ListByPage(ctx, "p") }, "ListByPage"},
		{"CheckLinked", func(r *fkElementEventWorkflowRepo, err error) {
			r.OnIsLinked = func(context.Context, string, string, string) (bool, error) { return true, err }
		}, func(s *ElementEventWorkflowService) (any, error) {
			return s.IsLinked(ctx, "e", "w", "n")
		}, "IsLinked"},
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
	r.OnListAll = func(context.Context) ([]models.ElementEventWorkflow, error) { return want, nil }
	r.OnIsLinked = func(context.Context, string, string, string) (bool, error) { return true, nil }
	s := NewElementEventWorkflowService(r, nil, nil, nil)
	got, err := s.ListAll(context.Background())
	wantNoErr(t, err)
	if len(got) != 2 || got[1].Id != "b" {
		t.Fatalf("got %v", got)
	}
	linked, err := s.IsLinked(context.Background(), "e", "w", "n")
	wantNoErr(t, err)
	if !linked {
		t.Fatal("expected linked")
	}
}
