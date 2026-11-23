package models

import (
	"encoding/json"
	"testing"
)

func TestContentFieldTableName(t *testing.T) {
	if got := (ContentField{}).TableName(); got != `public."ContentField"` {
		t.Fatalf("table name = %s", got)
	}
}

func TestContentFieldJSON(t *testing.T) {
	b, err := json.Marshal(ContentField{Id: "f", ContentTypeId: "ct", Name: "title", Type: "text", Required: true})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got["id"] != "f" || got["contentTypeId"] != "ct" || got["name"] != "title" || got["type"] != "text" || got["required"] != true {
		t.Fatalf("unexpected json: %s", b)
	}
	if _, ok := got["values"]; ok {
		t.Fatalf("empty values must be omitted: %s", b)
	}
	if _, ok := got["contentType"]; ok {
		t.Fatalf("nil contentType must be omitted: %s", b)
	}
}
