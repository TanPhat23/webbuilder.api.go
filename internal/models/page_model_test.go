package models

import "testing"

func TestNormalizeExplicitSlugForType(t *testing.T) {
	validCases := []struct {
		raw      string
		pageType string
		want     string
	}{
		{"/", PageTypeStatic, ""},
		{"/About-Us/", PageTypeStatic, "about-us"},
		{"blog/2026/launch", PageTypeStatic, "blog/2026/launch"},
		{"blog/:slug", PageTypeDynamic, "blog/:slug"},
		{"/Shop/:Category/:Slug/", PageTypeDynamic, "shop/:category/:slug"},
		{"docs/*slug", PageTypeDynamic, "docs/*slug"},
		{"docs/*", PageTypeDynamic, "docs/*"},
		{"*", PageTypeDynamic, "*"},
		{"/*path/", PageTypeDynamic, "*path"},
	}

	for _, tc := range validCases {
		got, err := NormalizeExplicitSlugForType(tc.raw, tc.pageType)
		if err != nil {
			t.Fatalf("NormalizeExplicitSlugForType(%q, %q) unexpected error: %v", tc.raw, tc.pageType, err)
		}
		if got != tc.want {
			t.Fatalf("NormalizeExplicitSlugForType(%q, %q) = %q, want %q", tc.raw, tc.pageType, got, tc.want)
		}
	}

	invalidCases := []struct {
		raw      string
		pageType string
	}{
		{"blog/:slug", PageTypeStatic},
		{"*", PageTypeStatic},
		{"docs/*path/edit", PageTypeDynamic},
		{"blog/:slug/:slug", PageTypeDynamic},
		{"blog/:slug/*slug", PageTypeDynamic},
		{"blog/:123", PageTypeDynamic},
		{"blog/**", PageTypeDynamic},
		{"blog/bad space", PageTypeDynamic},
	}

	for _, tc := range invalidCases {
		if got, err := NormalizeExplicitSlugForType(tc.raw, tc.pageType); err == nil {
			t.Fatalf("NormalizeExplicitSlugForType(%q, %q) = %q, want error", tc.raw, tc.pageType, got)
		}
	}
}

func TestCanonicalRoutePattern(t *testing.T) {
	if got := CanonicalRoutePattern("blog/:slug"); got != "blog/:" {
		t.Fatalf("CanonicalRoutePattern(blog/:slug) = %q", got)
	}
	if got := CanonicalRoutePattern("blog/:id"); got != "blog/:" {
		t.Fatalf("CanonicalRoutePattern(blog/:id) = %q", got)
	}
	if got := CanonicalRoutePattern("docs/*path"); got != "docs/*" {
		t.Fatalf("CanonicalRoutePattern(docs/*path) = %q", got)
	}
	if got := CanonicalRoutePattern("*"); got != "*" {
		t.Fatalf("CanonicalRoutePattern(*) = %q", got)
	}
}
