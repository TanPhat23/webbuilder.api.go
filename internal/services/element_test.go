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

func newSaveSnapshotReq() *proto.SaveSnapshotRequest {
	return &proto.SaveSnapshotRequest{
		Id: "s", Name: "n", Type: "auto", PageId: "pg", ProjectId: "pr", Timestamp: 9,
		Elements: `[{"id":"a","type":"Text"},{"id":"b","type":"Frame"}]`,
	}
}

func TestElementSaveSnapshotValidationAndRepoErrors(t *testing.T) {
	ctx := context.Background()

	for _, badJSON := range []string{"{", `[{"id":5}]`} {
		e := newElementEnv()
		r := newSaveSnapshotReq()
		r.Elements = badJSON
		if _, err := e.svc.SaveSnapshot(ctx, r); err == nil {
			t.Fatalf("expected error for elements=%q", badJSON)
		}
	}

	e := newElementEnv()
	r := newSaveSnapshotReq()
	r.PageId = ""
	_, err := e.svc.SaveSnapshot(ctx, r)
	wantErr(t, err, "pageId is required")
	r = newSaveSnapshotReq()
	r.ProjectId = ""
	_, err = e.svc.SaveSnapshot(ctx, r)
	wantErr(t, err, "projectId is required")
	wantCalls(t, e.sr.Calls)

	e.sr.OnSave = func(context.Context, string, *models.Snapshot) error { return errBoom }
	_, err = e.svc.SaveSnapshot(ctx, newSaveSnapshotReq())
	isBoom(t, err)
	wantCalls(t, e.er.Calls)

	e = newElementEnv()
	e.sr.OnSave = func(context.Context, string, *models.Snapshot) error { return nil }
	e.er.OnReplace = func(context.Context, string, []models.EditorElement) error { return errBoom }
	_, err = e.svc.SaveSnapshot(ctx, newSaveSnapshotReq())
	isBoom(t, err)
}

func TestElementSaveSnapshotSuccess(t *testing.T) {
	ctx := context.Background()
	e := newElementEnv()
	var snap *models.Snapshot
	var replaced []models.EditorElement
	var proj string
	e.sr.OnSave = func(_ context.Context, p string, s *models.Snapshot) error {
		proj, snap = p, s
		return nil
	}
	e.er.OnReplace = func(_ context.Context, p string, els []models.EditorElement) error {
		if p != "pr" {
			t.Errorf("project = %q", p)
		}
		replaced = els
		return nil
	}

	resp, err := e.svc.SaveSnapshot(ctx, newSaveSnapshotReq())
	wantNoErr(t, err)
	if resp.Message != "Snapshot saved successfully" {
		t.Fatalf("msg = %q", resp.Message)
	}
	if proj != "pr" || snap.Id != "s" || snap.Name != "n" {
		t.Fatalf("snapshot = %+v (proj=%q)", snap, proj)
	}
	if snap.Type != "auto" || snap.Timestamp != 9 || snap.ProjectId != "pr" {
		t.Fatalf("snapshot = %+v", snap)
	}
	if len(snap.Elements) == 0 {
		t.Fatal("expected snapshot elements")
	}
	if len(replaced) != 2 || replaced[0].GetElement().Id != "a" {
		t.Fatalf("replaced = %v", replaced)
	}
	if replaced[1].GetType() != "Frame" {
		t.Fatalf("replaced[1] = %v", replaced[1])
	}
}

func TestElementWorkflowLinksNilRepos(t *testing.T) {
	ctx := context.Background()
	s := NewElementService(nil, nil, nil, nil)

	ws, err := s.List(ctx, "e")
	wantNoErr(t, err)
	if ws == nil || len(ws) != 0 {
		t.Fatalf("ws = %#v", ws)
	}
	_, err = s.LinkElementToWorkflow(ctx, "e", "w", "n")
	wantErr(t, err, "not initialized")
	wantErr(t, s.UnlinkElementFromWorkflow(ctx, "e", "w", "n"), "not initialized")

	for _, fn := range []func() ([]models.EventWorkflow, error){
		func() ([]models.EventWorkflow, error) { return s.GetProjectWorkflows(ctx, "p") },
		func() ([]models.EventWorkflow, error) { return s.GetEnabledWorkflows(ctx, "p") },
	} {
		list, err := fn()
		wantNoErr(t, err)
		if list == nil || len(list) != 0 {
			t.Fatalf("list = %#v", list)
		}
	}

	bw, err := s.GetWorkflowsByEvent(ctx, "click")
	wantNoErr(t, err)
	if bw == nil || len(bw) != 0 {
		t.Fatalf("bw = %#v", bw)
	}
}

func TestElementWorkflowLinksCRUD(t *testing.T) {
	ctx := context.Background()
	e := newElementEnv()
	e.eew.OnListByElement = func(context.Context, string) ([]models.ElementEventWorkflow, error) {
		return []models.ElementEventWorkflow{{Id: "x"}}, nil
	}
	got, err := e.svc.List(ctx, "e")
	wantNoErr(t, err)
	if len(got) != 1 || got[0].Id != "x" {
		t.Fatalf("got %v", got)
	}
	e.eew.OnListByElement = func(context.Context, string) ([]models.ElementEventWorkflow, error) { return nil, errBoom }
	_, err = e.svc.List(ctx, "e")
	isBoom(t, err)

	var sent *models.ElementEventWorkflow
	e.eew.OnCreate = func(_ context.Context, x *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
		sent = x
		return x, nil
	}
	link, err := e.svc.LinkElementToWorkflow(ctx, "e", "w", "click")
	wantNoErr(t, err)
	if link.ElementId != "e" || link.WorkflowId != "w" {
		t.Fatalf("got %+v", link)
	}
	if link.EventName != "click" || sent != link {
		t.Fatalf("got %+v", link)
	}
	e.eew.OnCreate = func(context.Context, *models.ElementEventWorkflow) (*models.ElementEventWorkflow, error) {
		return nil, errBoom
	}
	_, err = e.svc.LinkElementToWorkflow(ctx, "e", "w", "click")
	isBoom(t, err)
}

func TestElementWorkflowLinksUnlinkAndLookups(t *testing.T) {
	ctx := context.Background()
	e := newElementEnv()
	e.eew.OnListByFilters = func(_ context.Context, element, workflow, event string) ([]models.ElementEventWorkflow, error) {
		if element != "e" || workflow != "w" || event != "n" {
			t.Fatal("incorrect filters")
		}
		return []models.ElementEventWorkflow{{Id: "link"}}, nil
	}
	e.eew.OnDelete = func(_ context.Context, id string) error {
		if id != "link" {
			t.Fatal(id)
		}
		return nil
	}
	wantNoErr(t, e.svc.UnlinkElementFromWorkflow(ctx, "e", "w", "n"))

	e.wr.OnListByProject = func(_ context.Context, p string) ([]models.EventWorkflow, error) {
		return []models.EventWorkflow{{}}, nil
	}
	e.wr.OnListEnabledByProject = func(context.Context, string) ([]models.EventWorkflow, error) {
		return []models.EventWorkflow{{}, {}}, nil
	}
	a, err := e.svc.GetProjectWorkflows(ctx, "p")
	wantNoErr(t, err)
	b, err := e.svc.GetEnabledWorkflows(ctx, "p")
	wantNoErr(t, err)
	if len(a) != 1 || len(b) != 2 {
		t.Fatalf("%d %d", len(a), len(b))
	}
	e.wr.OnListByProject = func(context.Context, string) ([]models.EventWorkflow, error) { return nil, errBoom }
	e.wr.OnListEnabledByProject = func(context.Context, string) ([]models.EventWorkflow, error) { return nil, errBoom }
	_, err = e.svc.GetProjectWorkflows(ctx, "p")
	isBoom(t, err)
	_, err = e.svc.GetEnabledWorkflows(ctx, "p")
	isBoom(t, err)
}

func TestElementWorkflowLinksByEvent(t *testing.T) {
	ctx := context.Background()
	e := newElementEnv()
	e.eew.OnListByEvent = func(_ context.Context, n string) ([]models.ElementEventWorkflow, error) {
		return []models.ElementEventWorkflow{{EventName: n}}, nil
	}
	got, err := e.svc.GetWorkflowsByEvent(ctx, "click")
	wantNoErr(t, err)
	if len(got) != 1 || got[0].EventName != "click" {
		t.Fatalf("got %v", got)
	}
	e.eew.OnListByEvent = func(context.Context, string) ([]models.ElementEventWorkflow, error) { return nil, errBoom }
	_, err = e.svc.GetWorkflowsByEvent(ctx, "click")
	isBoom(t, err)
}

func TestElementGetPageElements(t *testing.T) {
	ctx := context.Background()
	e := newElementEnv()

	_, err := e.svc.GetPageElements(ctx, &proto.PageElementsRequest{ProjectId: "p"})
	wantErr(t, err, "pageId is required")
	_, err = e.svc.GetPageElements(ctx, &proto.PageElementsRequest{PageId: "pg"})
	wantErr(t, err, "projectId is required")

	e.er.OnListByPage = func(context.Context, string) ([]models.Element, error) { return nil, errBoom }
	_, err = e.svc.GetPageElements(ctx, &proto.PageElementsRequest{PageId: "pg", ProjectId: "p"})
	isBoom(t, err)

	parent, content, settings := "root", "hello", rawp(`{"a":1}`)
	e.er.OnListByPage = func(_ context.Context, id string) ([]models.Element, error) {
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
	assertChildProtoElement(t, resp.Elements[0].Elements)
}

func assertChildProtoElement(t *testing.T, kids []*proto.Element) {
	t.Helper()
	if len(kids) != 1 || kids[0].Id != "child" {
		t.Fatalf("kids = %v", kids)
	}
	if kids[0].Order != 2 || kids[0].Styles != `{"c":"red"}` {
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
