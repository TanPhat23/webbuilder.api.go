package models

import "testing"

func TestSlugifyCollectionName(t *testing.T) {
	cases := []struct {
		name     string
		fallback string
		want     string
	}{
		{"Catering Inquiries", "id1234", "catering-inquiries"},
		{"  Menu & Specials! ", "id1234", "menu-specials"},
		{"!!!", "abc12345", "collection-abc12345"},
	}
	for _, tc := range cases {
		if got := SlugifyCollectionName(tc.name, tc.fallback); got != tc.want {
			t.Errorf("SlugifyCollectionName(%q, %q) = %q, want %q", tc.name, tc.fallback, got, tc.want)
		}
	}
}

func TestIsValidWebhookEvent(t *testing.T) {
	if !IsValidWebhookEvent(WebhookEventItemCreated) || !IsValidWebhookEvent(WebhookEventItemUpdated) || !IsValidWebhookEvent(WebhookEventItemDeleted) {
		t.Fatal("expected standard item events to be valid")
	}
	if IsValidWebhookEvent("item.exploded") {
		t.Fatal("expected unknown event to be invalid")
	}
}
