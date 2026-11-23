package models

import (
	"encoding/json"
	"testing"
)

func TestCustomElementTableNames(t *testing.T) {
	if got := (CustomElement{}).TableName(); got != `public."CustomElement"` {
		t.Fatalf("table name = %s", got)
	}
	if got := (CustomElementType{}).TableName(); got != `public."CustomElementType"` {
		t.Fatalf("type table name = %s", got)
	}
}

func TestCustomElementJSONKeepsRawDocumentsVerbatim(t *testing.T) {
	icon := "star"
	ce := CustomElement{
		Id: "c", UserId: "u", Name: "Card", Version: "1.0.0", IsPublic: true, Icon: &icon,
		Structure:    json.RawMessage(`{"type":"Frame","children":[]}`),
		DefaultProps: json.RawMessage(`{"a":1}`),
	}
	b, err := json.Marshal(ce)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Structure    json.RawMessage `json:"structure"`
		DefaultProps json.RawMessage `json:"defaultProps"`
		IsPublic     bool            `json:"isPublic"`
		Icon         string          `json:"icon"`
		Description  *string         `json:"description"`
	}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if string(got.Structure) != `{"type":"Frame","children":[]}` || string(got.DefaultProps) != `{"a":1}` || !got.IsPublic || got.Icon != "star" || got.Description != nil {
		t.Fatalf("unexpected json: %s", b)
	}
}

func TestCustomElementJSONOmitsEmptyDefaultProps(t *testing.T) {
	b, err := json.Marshal(CustomElement{Id: "c", Structure: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	_ = json.Unmarshal(b, &got)
	if _, ok := got["defaultProps"]; ok {
		t.Fatalf("defaultProps must be omitted: %s", b)
	}
	if got["isPublic"] != false {
		t.Fatalf("isPublic must always be present: %s", b)
	}
}

func TestCustomElementTypeJSON(t *testing.T) {
	cat := "layout"
	b, err := json.Marshal(CustomElementType{Id: "t", Name: "Hero", Category: &cat})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	_ = json.Unmarshal(b, &got)
	if got["name"] != "Hero" || got["category"] != "layout" {
		t.Fatalf("unexpected json: %s", b)
	}
	if _, ok := got["icon"]; ok {
		t.Fatalf("nil icon must be omitted: %s", b)
	}
}
