package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestContentItemTableName(t *testing.T) {
	if got := (ContentItem{}).TableName(); got != `public."ContentItem"` {
		t.Fatalf("table name = %s", got)
	}
}

func TestContentItemJSON(t *testing.T) {
	val := "x"
	item := ContentItem{
		Id: "i", ContentTypeId: "ct", Slug: "a-b", Title: "T", Published: true,
		FieldValues: []ContentFieldValue{{Id: "v", Value: &val}},
		AuditFields: AuditFields{CreatedAt: time.Date(2026, 5, 6, 7, 8, 9, 0, time.UTC)},
	}
	b, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got["slug"] != "a-b" || got["title"] != "T" || got["published"] != true || got["createdAt"] != "2026-05-06T07:08:09Z" {
		t.Fatalf("unexpected json: %s", b)
	}
	if fv, ok := got["fieldValues"].([]any); !ok || len(fv) != 1 {
		t.Fatalf("fieldValues missing: %s", b)
	}
	if _, ok := got["deletedAt"]; ok {
		t.Fatalf("nil deletedAt must be omitted: %s", b)
	}
	if _, ok := got["contentType"]; ok {
		t.Fatalf("nil contentType must be omitted: %s", b)
	}
}
