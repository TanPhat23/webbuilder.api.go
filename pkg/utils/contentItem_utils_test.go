package utils

import (
	"testing"
	"time"

	"my-go-app/internal/models"
)

func strp(s string) *string { return &s }

func TestFlattenContentItem(t *testing.T) {
	created := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	updated := created.Add(time.Hour)
	ct := &models.ContentType{Id: "ct1", Name: "Post"}
	item := &models.ContentItem{
		Id: "i1", ContentTypeId: "ct1", Slug: "hello", Title: "Hello", Published: true, ContentType: ct,
		AuditFields: models.AuditFields{CreatedAt: created, UpdatedAt: updated},
		FieldValues: []models.ContentFieldValue{
			{Field: &models.ContentField{Name: "body"}, Value: strp("text")},
			{Field: &models.ContentField{Name: "subtitle"}, Value: nil},
			{Field: &models.ContentField{Name: ""}, Value: strp("dropped: unnamed field")},
		},
	}

	got := FlattenContentItem(item)

	base := map[string]any{
		"id": "i1", "title": "Hello", "slug": "hello", "published": true,
		"createdAt": created, "updatedAt": updated, "contentTypeId": "ct1",
	}
	for k, want := range base {
		if got[k] != want {
			t.Errorf("%s = %v, want %v", k, got[k], want)
		}
	}
	if got["contentType"] != ct {
		t.Errorf("contentType not carried through: %v", got["contentType"])
	}
	if v, ok := got["body"].(*string); !ok || *v != "text" {
		t.Errorf("field value not flattened by field name: %#v", got["body"])
	}
	if v, ok := got["subtitle"]; !ok || v.(*string) != nil {
		t.Errorf("a field with a nil value must still appear as a key: %#v", got["subtitle"])
	}
	if len(got) != len(base)+1+2 { // + contentType + body + subtitle
		t.Errorf("unexpected keys in %v", got)
	}
}

func TestFlattenContentItemWithoutFieldValues(t *testing.T) {
	got := FlattenContentItem(&models.ContentItem{Id: "i", Slug: "s"})
	if got["id"] != "i" || got["slug"] != "s" || got["published"] != false {
		t.Errorf("base keys wrong: %v", got)
	}
	if ct, ok := got["contentType"]; !ok || ct.(*models.ContentType) != nil {
		t.Errorf("contentType key should exist with a nil value: %#v", ct)
	}
}

func TestFlattenContentItemsKeepsEachItemDistinct(t *testing.T) {
	items := []models.ContentItem{
		{Id: "a", Title: "A", FieldValues: []models.ContentFieldValue{{Field: &models.ContentField{Name: "x"}, Value: strp("1")}}},
		{Id: "b", Title: "B"},
		{Id: "c", Title: "C"},
	}
	got := FlattenContentItems(items)
	if len(got) != 3 {
		t.Fatalf("len = %d", len(got))
	}
	for i, id := range []string{"a", "b", "c"} {
		if got[i]["id"] != id {
			t.Errorf("item %d id = %v, want %s (loop variable aliasing?)", i, got[i]["id"], id)
		}
	}
	if _, leaked := got[1]["x"]; leaked {
		t.Error("field from item a leaked into item b")
	}
	if out := FlattenContentItems(nil); out == nil || len(out) != 0 {
		t.Errorf("nil input should give an empty non-nil slice, got %#v", out)
	}
}
