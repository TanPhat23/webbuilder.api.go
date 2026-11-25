package utils

import (
	"encoding/json"
	"reflect"
	"testing"

	"my-go-app/internal/models"
	"my-go-app/proto"
)

func TestConvertElementToProtoNil(t *testing.T) {
	if ConvertElementToProto(nil) != nil {
		t.Error("nil element must map to nil proto")
	}
	if ConvertProtoToElement(nil) != nil {
		t.Error("nil proto must map to nil element")
	}
}

func fullElement() *models.Element {
	settings := json.RawMessage(`{"autoplay":true}`)
	return &models.Element{
		Id: "root", Type: "Section", Order: 3,
		Content: strp("c"), Name: strp("n"), TailwindStyles: strp("p-4"),
		Src: strp("s"), Href: strp("h"), ParentId: strp("par"), PageId: strp("pg"),
		Styles:   json.RawMessage(`{"default":{"color":"red"}}`),
		Settings: &settings,
		EventWorkflows: []models.ElementEventWorkflow{
			{Id: "l1", EventName: "onClick", ElementId: "root", WorkflowId: "wf1"},
		},
		Elements: []models.Element{
			{Id: "kid", Type: "Text", Order: 1, Elements: []models.Element{{Id: "grandkid", Type: "Text"}}},
		},
	}
}

func TestConvertElementToProtoMapsEveryField(t *testing.T) {
	p := ConvertElementToProto(fullElement())

	if p.Id != "root" || p.Type != "Section" || p.Order != 3 {
		t.Errorf("scalars: %+v", p)
	}
	for name, pair := range map[string][2]*string{
		"content": {p.Content, strp("c")}, "name": {p.Name, strp("n")}, "tailwind": {p.TailwindStyles, strp("p-4")},
		"src": {p.Src, strp("s")}, "href": {p.Href, strp("h")}, "parent": {p.ParentId, strp("par")}, "page": {p.PageId, strp("pg")},
	} {
		if pair[0] == nil || *pair[0] != *pair[1] {
			t.Errorf("%s not mapped: %v", name, pair[0])
		}
	}
	if p.Styles != `{"default":{"color":"red"}}` {
		t.Errorf("styles = %q", p.Styles)
	}
	if p.Settings == nil || *p.Settings != `{"autoplay":true}` {
		t.Errorf("settings = %v", p.Settings)
	}
	if p.EventWorkflows == nil {
		t.Fatal("event workflows not serialised")
	}
	var wf []map[string]any
	if err := json.Unmarshal([]byte(*p.EventWorkflows), &wf); err != nil || len(wf) != 1 || wf[0]["id"] != "l1" || wf[0]["eventName"] != "onClick" {
		t.Errorf("event workflows JSON = %s (%v)", *p.EventWorkflows, err)
	}
	if len(p.Elements) != 1 || p.Elements[0].Id != "kid" || len(p.Elements[0].Elements) != 1 || p.Elements[0].Elements[0].Id != "grandkid" {
		t.Errorf("children not converted recursively: %+v", p.Elements)
	}
}

func TestConvertElementToProtoLeavesOptionalFieldsUnset(t *testing.T) {
	p := ConvertElementToProto(&models.Element{Id: "x", Type: "Text"})
	if p.Content != nil || p.Name != nil || p.TailwindStyles != nil || p.Src != nil || p.Href != nil ||
		p.ParentId != nil || p.PageId != nil || p.Settings != nil || p.EventWorkflows != nil || p.Elements != nil {
		t.Errorf("optional fields must stay nil: %+v", p)
	}
	if p.Styles != "" {
		t.Errorf("styles = %q", p.Styles)
	}
}

func TestElementProtoRoundTrip(t *testing.T) {
	in := fullElement()
	out := ConvertProtoToElement(ConvertElementToProto(in))

	if out.Id != in.Id || out.Type != in.Type || out.Order != in.Order {
		t.Errorf("scalars: %+v", out)
	}
	if *out.Content != "c" || *out.Name != "n" || *out.TailwindStyles != "p-4" || *out.Src != "s" ||
		*out.Href != "h" || *out.ParentId != "par" || *out.PageId != "pg" {
		t.Errorf("optional strings: %+v", out)
	}
	if string(out.Styles) != string(in.Styles) || string(*out.Settings) != string(*in.Settings) {
		t.Errorf("json fields: %s / %s", out.Styles, *out.Settings)
	}
	if len(out.EventWorkflows) != 1 || out.EventWorkflows[0].Id != "l1" || out.EventWorkflows[0].EventName != "onClick" || out.EventWorkflows[0].WorkflowId != "wf1" {
		t.Errorf("workflows: %+v", out.EventWorkflows)
	}
	if len(out.Elements) != 1 || out.Elements[0].Id != "kid" || out.Elements[0].Elements[0].Id != "grandkid" {
		t.Errorf("children: %+v", out.Elements)
	}
}

func TestConvertProtoToElementToleratesBadWorkflowsAndNilChildren(t *testing.T) {
	bad := "not json"
	p := &proto.Element{
		Id: "x", Type: "Section", EventWorkflows: &bad,
		Elements: []*proto.Element{{Id: "a"}, nil, {Id: "c"}},
	}
	e := ConvertProtoToElement(p)
	if e.EventWorkflows != nil {
		t.Errorf("invalid workflow JSON must be dropped, got %+v", e.EventWorkflows)
	}
	if len(e.Elements) != 3 || e.Elements[0].Id != "a" || e.Elements[1].Id != "" || e.Elements[2].Id != "c" {
		t.Errorf("nil child must leave a zero slot so indexes stay aligned: %+v", e.Elements)
	}
	if e.Settings != nil {
		t.Errorf("unset settings must stay nil")
	}
}

func TestConvertElementsToProto(t *testing.T) {
	if got := ConvertElementsToProto(nil); got != nil {
		t.Errorf("nil slice: %v", got)
	}
	if got := ConvertElementsToProto([]models.Element{}); got != nil {
		t.Errorf("empty slice: %v", got)
	}
	got := ConvertElementsToProto([]models.Element{{Id: "a", Type: "Text"}, {Id: "b", Type: "Text"}})
	if len(got) != 2 || got[0].Id != "a" || got[1].Id != "b" {
		t.Errorf("each element must map to its own proto (loop aliasing): %+v", got)
	}
	if got[0] == got[1] {
		t.Error("distinct elements share one proto pointer")
	}
}

func TestConvertProtoElementsToModel(t *testing.T) {
	if got := ConvertProtoToElements(nil); got != nil {
		t.Errorf("nil: %v", got)
	}
	got := ConvertProtoToElements([]*proto.Element{{Id: "a"}, nil, {Id: "b"}})
	if len(got) != 2 || got[0].Id != "a" || got[1].Id != "b" {
		t.Errorf("nil entries must be skipped: %+v", got)
	}
}

func TestEventWorkflowStringConversion(t *testing.T) {
	if MarshalEventWorkflows(nil) != nil || MarshalEventWorkflows([]models.ElementEventWorkflow{}) != nil {
		t.Error("empty workflows must map to nil")
	}
	in := []models.ElementEventWorkflow{{Id: "1", EventName: "onClick", ElementId: "e", WorkflowId: "w"}}
	s := MarshalEventWorkflows(in)
	if s == nil {
		t.Fatal("nil result")
	}
	back := UnmarshalEventWorkflows(s)
	if len(back) != 1 || back[0].Id != "1" || back[0].EventName != "onClick" || back[0].ElementId != "e" || back[0].WorkflowId != "w" {
		t.Errorf("round trip: %+v", back)
	}

	empty := ""
	bad := "{oops"
	for name, in := range map[string]*string{"nil": nil, "empty": &empty, "invalid": &bad} {
		got := UnmarshalEventWorkflows(in)
		if got == nil || !reflect.DeepEqual(got, []models.ElementEventWorkflow{}) {
			t.Errorf("%s: want empty non-nil slice, got %#v", name, got)
		}
	}
}
