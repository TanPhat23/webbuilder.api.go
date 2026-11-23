package models

import (
	"encoding/json"
	"testing"
)

func TestElementTableName(t *testing.T) {
	if got := (Element{}).TableName(); got != `public."Element"` {
		t.Fatalf("table name = %s", got)
	}
}

func TestElementGetters(t *testing.T) {
	e := &Element{Id: "e", Type: "Button"}
	if e.GetElement() != e {
		t.Fatal("GetElement must return the receiver")
	}
	if e.GetType() != "Button" {
		t.Fatalf("GetType = %q", e.GetType())
	}
	var _ EditorElement = e
}

func TestElementTypePredicates(t *testing.T) {
	cms := []string{"CMSContentList", "CMSContentItem", "CMSContentGrid"}
	containers := append([]string{"Section", "Frame", "Carousel", "List", "Select", "Form", "DataLoader"}, cms...)
	preds := map[string]struct {
		fn   func(*Element) bool
		want []string
	}{
		"IsContainer":      {(*Element).IsContainer, containers},
		"HasElements":      {(*Element).HasElements, containers},
		"IsButton":         {(*Element).IsButton, []string{"Button"}},
		"IsInput":          {(*Element).IsInput, []string{"Input"}},
		"IsText":           {(*Element).IsText, []string{"Text"}},
		"IsSection":        {(*Element).IsSection, []string{"Section"}},
		"IsFrame":          {(*Element).IsFrame, []string{"Frame"}},
		"IsCarousel":       {(*Element).IsCarousel, []string{"Carousel"}},
		"IsList":           {(*Element).IsList, []string{"List"}},
		"IsSelect":         {(*Element).IsSelect, []string{"Select"}},
		"IsForm":           {(*Element).IsForm, []string{"Form"}},
		"IsDataLoader":     {(*Element).IsDataLoader, []string{"DataLoader"}},
		"IsCMSContent":     {(*Element).IsCMSContent, cms},
		"IsCMSContentList": {(*Element).IsCMSContentList, []string{"CMSContentList"}},
		"IsCMSContentItem": {(*Element).IsCMSContentItem, []string{"CMSContentItem"}},
		"IsCMSContentGrid": {(*Element).IsCMSContentGrid, []string{"CMSContentGrid"}},
		"IsCustomElement":  {(*Element).IsCustomElement, []string{"CustomElement"}},
		"CanHaveHref":      {(*Element).CanHaveHref, []string{"Button", "Text"}},
		"CanHaveSrc":       {(*Element).CanHaveSrc, []string{"Image", "Video", "Iframe"}},
		"CanHaveContent":   {(*Element).CanHaveContent, []string{"Button", "Text", "Input", "Select"}},
	}
	types := []string{"Section", "Frame", "Carousel", "List", "Select", "Form", "DataLoader",
		"CMSContentList", "CMSContentItem", "CMSContentGrid", "Button", "Text", "Input",
		"Image", "Video", "Iframe", "CustomElement", "Unknown", "", "button"}
	for name, p := range preds {
		want := map[string]bool{}
		for _, w := range p.want {
			want[w] = true
		}
		for _, typ := range types {
			if got := p.fn(&Element{Type: typ}); got != want[typ] {
				t.Errorf("%s(%q) = %v, want %v", name, typ, got, want[typ])
			}
		}
	}
}

func TestElementJSON(t *testing.T) {
	name := "n"
	settings := json.RawMessage(`{"a":1}`)
	e := Element{
		Id: "e", Type: "Text", Name: &name, Styles: json.RawMessage(`{"color":"red"}`), Settings: &settings,
		Order: 2, Elements: []Element{{Id: "child"}}, Parent: &Element{Id: "parent"},
	}
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if string(got["id"]) != `"e"` || string(got["styles"]) != `{"color":"red"}` || string(got["settings"]) != `{"a":1}` || string(got["order"]) != "2" || string(got["name"]) != `"n"` {
		t.Fatalf("unexpected json: %s", b)
	}
	for _, k := range []string{"Elements", "elements", "parent", "Parent", "eventWorkflows", "comments", "page", "pageId", "orderKey"} {
		if _, ok := got[k]; ok {
			t.Errorf("%s must not be serialised: %s", k, b)
		}
	}
}
