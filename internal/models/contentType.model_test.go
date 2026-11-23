package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestContentTypeTableName(t *testing.T) {
	if got := (ContentType{}).TableName(); got != `public."ContentType"` {
		t.Fatalf("table name = %s", got)
	}
}

func TestContentTypeJSONHidesTimestamps(t *testing.T) {
	desc := "d"
	b, err := json.Marshal(ContentType{Id: "c", Name: "Post", Description: &desc, CreatedAt: time.Now(), UpdatedAt: time.Now(), Fields: []ContentField{{Id: "f"}}})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got["id"] != "c" || got["name"] != "Post" || got["description"] != "d" {
		t.Fatalf("unexpected json: %s", b)
	}
	for _, k := range []string{"CreatedAt", "UpdatedAt", "createdAt", "updatedAt", "items"} {
		if _, ok := got[k]; ok {
			t.Fatalf("%s must not be serialised: %s", k, b)
		}
	}
	if f, ok := got["fields"].([]any); !ok || len(f) != 1 {
		t.Fatalf("fields missing: %s", b)
	}
}
