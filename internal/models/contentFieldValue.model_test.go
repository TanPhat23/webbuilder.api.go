package models

import (
	"encoding/json"
	"testing"
)

func TestContentFieldValueTableName(t *testing.T) {
	if got := (ContentFieldValue{}).TableName(); got != `public."ContentFieldValue"` {
		t.Fatalf("table name = %s", got)
	}
}

func TestContentFieldValueJSONRoundTrip(t *testing.T) {
	v := "hello"
	b, err := json.Marshal(ContentFieldValue{Id: "v", ContentItemId: "i", FieldId: "f", Value: &v})
	if err != nil {
		t.Fatal(err)
	}
	var back ContentFieldValue
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.Value == nil || *back.Value != "hello" || back.ContentItemId != "i" || back.FieldId != "f" {
		t.Fatalf("round trip lost data: %s", b)
	}
	b, _ = json.Marshal(ContentFieldValue{Id: "v"})
	var got map[string]any
	_ = json.Unmarshal(b, &got)
	if _, ok := got["value"]; ok {
		t.Fatalf("nil value must be omitted: %s", b)
	}
}
