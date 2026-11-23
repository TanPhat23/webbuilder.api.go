package services

import (
	"context"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/proto"
)

type elementEnv struct {
	svc *ElementService
	sr  *fkSnapshotRepo
	er  *fkElementRepo
	wr  *fkEventWorkflowRepo
	eew *fkElementEventWorkflowRepo
}

func newElementEnv() *elementEnv {
	e := &elementEnv{sr: &fkSnapshotRepo{}, er: &fkElementRepo{}, wr: &fkEventWorkflowRepo{}, eew: &fkElementEventWorkflowRepo{}}
	e.svc = NewElementService(e.sr, e.er, e.wr, e.eew)
	return e
}

func TestElementSaveSnapshot(t *testing.T) {
	ctx := context.Background()
	req := func() *proto.SaveSnapshotRequest {
		return &proto.SaveSnapshotRequest{Id: "s", Name: "n", Type: "auto", PageId: "pg", ProjectId: "pr", Timestamp: 9,
			Elements: `[{"id":"a","type":"Text"},{"id":"b","type":"Frame"}]`}
	}

	t.Run("invalid json", func(t *testing.T) {
		e := newElementEnv()
		r := req()
		r.Elements = "{"
		_, err := e.svc.SaveSnapshot(ctx, r)
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("element not convertible", func(t *testing.T) {
		e := newElementEnv()
		r := req()
		r.Elements = `[{"id":5}]`
		_, err := e.svc.SaveSnapshot(ctx, r)
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("missing ids", func(t *testing.T) {
		e := newElementEnv()
		r := req()
		r.PageId = ""
		_, err := e.svc.SaveSnapshot(ctx, r)
		wantErr(t, err, "pageId is required")
		r = req()
		r.ProjectId = ""
		_, err = e.svc.SaveSnapshot(ctx, r)
		wantErr(t, err, "projectId is required")
		wantCalls(t, e.sr.Calls)
	})
	t.Run("snapshot save error stops before replace", func(t *testing.T) {
		e := newElementEnv()
		e.sr.OnSaveSnapshot = func(context.Context, string, *models.Snapshot) error { return errBoom }
		_, err := e.svc.SaveSnapshot(ctx, req())
		isBoom(t, err)
		wantCalls(t, e.er.Calls)
	})
	t.Run("replace error", func(t *testing.T) {
		e := newElementEnv()
		e.sr.OnSaveSnapshot = func(context.Context, string, *models.Snapshot) error { return nil }
		e.er.OnReplaceElements = func(context.Context, string, []models.EditorElement) error { return errBoom }
		_, err := e.svc.SaveSnapshot(ctx, req())
		isBoom(t, err)
	})
	t.Run("success", func(t *testing.T) {
		e := newElementEnv()
		var snap *models.Snapshot
		var replaced []models.EditorElement
		var proj string
		e.sr.OnSaveSnapshot = func(_ context.Context, p string, s *models.Snapshot) error {
			proj, snap = p, s
			return nil
		}
		e.er.OnReplaceElements = func(_ context.Context, p string, els []models.EditorElement) error {
			if p != "pr" {
				t.Errorf("project = %q", p)
			}
			replaced = els
			return nil
		}
		resp, err := e.svc.SaveSnapshot(ctx, req())
		wantNoErr(t, err)
		if resp.Message != "Snapshot saved successfully" {
			t.Fatalf("msg = %q", resp.Message)
		}
		if proj != "pr" || snap.Id != "s" || snap.Name != "n" || snap.Type != "auto" || snap.Timestamp != 9 || snap.ProjectId != "pr" || len(snap.Elements) == 0 {
			t.Fatalf("snapshot = %+v", snap)
		}
		if len(replaced) != 2 || replaced[0].GetElement().Id != "a" || replaced[1].GetType() != "Frame" {
			t.Fatalf("replaced = %v", replaced)
		}
	})
}

func TestElementWorkflowLinks(t *testing.T) {
	ctx := context.Background()

	t.Run("nil repos", func(t *testing.T) {
		s := NewElementService(nil, nil, nil, nil)
		ws, err := s.GetElementEventWorkflows(ctx, "e")
		wantNoErr(t, err)
		if ws == nil || len(ws) != 0 {
			t.Fatalf("ws = %#v", ws)
		}
		_, err = s.LinkElementToWorkflow(ctx, "e", "w", "n")
		wantErr(t, err, "not initialized")
		wantErr(t, s.UnlinkElementFromWorkflow(ctx, "e", "w", "n"), "not initialized")
		pw, err := s.GetProjectWorkflows(ctx, "p")
		wantNoErr(t, err)
		if pw == nil || len(pw) != 0 {
			t.Fatalf("pw = %#v", pw)
		}
		ew, err := s.GetEnabledWorkflows(ctx, "p")
		wantNoErr(t, err)
		if ew == nil || len(ew) != 0 {
			t.Fatalf("ew = %#v", ew)
		}
		bw, err := s.GetWorkflowsByEvent(ctx, "click")
		wantNoErr(t, err)
		if bw == nil || len(bw) != 0 {
			t.Fatalf("bw = %#v", bw)
		}
	})

	t.Run("GetElementEventWorkflows", func(t *testing.T) {
		e := newElementEnv()
		e.eew.OnGetElementEventWorkflowsByElementID = func(context.Context, string) ([]models.ElementEventWorkflow, error) {
			return []models.ElementEventWorkflow{{Id: "x"}}, nil
		}
		got, err := e.svc.GetElementEventWorkflows(ctx, "e")
		wantNoErr(t, err)
		if len(got) != 1 || got[0].Id != "x" {
			t.Fatalf("got %v", got)
		}
		e.eew.OnGetElementEventWorkflowsByElementID = func(context.Context, string) ([]models.ElementEventWorkflow, error) { return nil, errBoom }
		_, err = e.svc.GetElementEventWorkflows(ctx, "e")
		isBoom(t, err)
	})

	t.Run("Link", func(t *testing.T) {
		e := newElementEnv()
		var sent *models.ElementEventWorkflow
		e.eew.OnCreateElementEventWorkflow = func(_ context.Context, x *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
			sent = x
			return x, nil
		}
		got, err := e.svc.LinkElementToWorkflow(ctx, "e", "w", "click")
		wantNoErr(t, err)
		if got.ElementId != "e" || got.WorkflowId != "w" || got.EventName != "click" || sent != got {
			t.Fatalf("got %+v", got)
		}
		e.eew.OnCreateElementEventWorkflow = func(context.Context, *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
			return nil, errBoom
		}
		_, err = e.svc.LinkElementToWorkflow(ctx, "e", "w", "click")
		isBoom(t, err)
	})

	t.Run("Unlink", func(t *testing.T) {
		e := newElementEnv()
		e.eew.OnGetElementEventWorkflowsByFilters = func(_ context.Context, element, workflow, event string) ([]models.ElementEventWorkflow, error) {
			if element != "e" || workflow != "w" || event != "n" {
				t.Fatal("incorrect filters")
			}
			return []models.ElementEventWorkflow{{Id: "link"}}, nil
		}
		e.eew.OnDeleteElementEventWorkflow = func(_ context.Context, id string) error {
			if id != "link" {
				t.Fatal(id)
			}
			return nil
		}
		wantNoErr(t, e.svc.UnlinkElementFromWorkflow(ctx, "e", "w", "n"))
	})

	t.Run("project workflows", func(t *testing.T) {
		e := newElementEnv()
		e.wr.OnGetEventWorkflowsByProjectID = func(_ context.Context, p string) ([]models.EventWorkflow, error) {
			return []models.EventWorkflow{{}}, nil
		}
		e.wr.OnGetEnabledEventWorkflowsByProjectID = func(context.Context, string) ([]models.EventWorkflow, error) {
			return []models.EventWorkflow{{}, {}}, nil
		}
		a, err := e.svc.GetProjectWorkflows(ctx, "p")
		wantNoErr(t, err)
		b, err := e.svc.GetEnabledWorkflows(ctx, "p")
		wantNoErr(t, err)
		if len(a) != 1 || len(b) != 2 {
			t.Fatalf("%d %d", len(a), len(b))
		}
		e.wr.OnGetEventWorkflowsByProjectID = func(context.Context, string) ([]models.EventWorkflow, error) { return nil, errBoom }
		e.wr.OnGetEnabledEventWorkflowsByProjectID = func(context.Context, string) ([]models.EventWorkflow, error) { return nil, errBoom }
		_, err = e.svc.GetProjectWorkflows(ctx, "p")
		isBoom(t, err)
		_, err = e.svc.GetEnabledWorkflows(ctx, "p")
		isBoom(t, err)
	})

	t.Run("by event", func(t *testing.T) {
		e := newElementEnv()
		e.eew.OnGetElementEventWorkflowsByEventName = func(_ context.Context, n string) ([]models.ElementEventWorkflow, error) {
			return []models.ElementEventWorkflow{{EventName: n}}, nil
		}
		got, err := e.svc.GetWorkflowsByEvent(ctx, "click")
		wantNoErr(t, err)
		if len(got) != 1 || got[0].EventName != "click" {
			t.Fatalf("got %v", got)
		}
		e.eew.OnGetElementEventWorkflowsByEventName = func(context.Context, string) ([]models.ElementEventWorkflow, error) { return nil, errBoom }
		_, err = e.svc.GetWorkflowsByEvent(ctx, "click")
		isBoom(t, err)
	})
}

func TestElementGetPageElements(t *testing.T) {
	ctx := context.Background()
	e := newElementEnv()

	_, err := e.svc.GetPageElements(ctx, &proto.PageElementsRequest{ProjectId: "p"})
	wantErr(t, err, "pageId is required")
	_, err = e.svc.GetPageElements(ctx, &proto.PageElementsRequest{PageId: "pg"})
	wantErr(t, err, "projectId is required")

	e.er.OnGetElementsByPageID = func(context.Context, string) ([]models.Element, error) { return nil, errBoom }
	_, err = e.svc.GetPageElements(ctx, &proto.PageElementsRequest{PageId: "pg", ProjectId: "p"})
	isBoom(t, err)

	parent := "root"
	content := "hello"
	settings := rawp(`{"a":1}`)
	e.er.OnGetElementsByPageID = func(_ context.Context, id string) ([]models.Element, error) {
		if id != "pg" {
			t.Errorf("page = %q", id)
		}
		return []models.Element{
			{Id: "root", Type: "Frame", PageId: strp("pg")},
			{Id: "child", Type: "Text", ParentId: &parent, Content: &content, Settings: settings, Order: 2, Styles: []byte(`{"c":"red"}`)},
		}, nil
	}
	resp, err := e.svc.GetPageElements(ctx, &proto.PageElementsRequest{PageId: "pg", ProjectId: "p"})
	wantNoErr(t, err)
	if len(resp.Elements) != 1 || resp.Elements[0].Id != "root" {
		t.Fatalf("roots = %v", resp.Elements)
	}
	kids := resp.Elements[0].Elements
	if len(kids) != 1 || kids[0].Id != "child" || kids[0].Order != 2 || kids[0].Styles != `{"c":"red"}` {
		t.Fatalf("kids = %v", kids)
	}
	if kids[0].Settings == nil || *kids[0].Settings != `{"a":1}` {
		t.Fatalf("settings = %v", kids[0].Settings)
	}
}

func TestElementConvertEditorElementsToProtoSkipsNil(t *testing.T) {
	s := NewElementService(nil, nil, nil, nil)
	var nilElem *models.Element
	got := s.convertEditorElementsToProto([]models.EditorElement{nil, nilElem, &models.Element{Id: "k", Type: "Text"}})
	if len(got) != 1 || got[0].Id != "k" {
		t.Fatalf("got %v", got)
	}
}
