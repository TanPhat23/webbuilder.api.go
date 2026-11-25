package services

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"my-go-app/internal/models"
)

func TestPublicCollection_OffsetCap(t *testing.T) {
	projID := "p1"
	ctRepo := &fkContentTypeRepo{
		OnGetContentTypeByProjectAndSlugOrID: func(context.Context, string, string) (*models.ContentType, error) {
			return &models.ContentType{Id: "ct1", ProjectId: &projID, PublicRead: true}, nil
		},
	}
	var gotLimit int
	itemRepo := &fkContentItemRepo{
		OnGetPublicContentItems: func(_ context.Context, _ string, limit int, _, _ string) ([]models.ContentItem, error) {
			gotLimit = limit
			return []models.ContentItem{}, nil
		},
	}
	svc := NewPublicCollectionService(ctRepo, itemRepo)
	page, err := svc.ListPublicItems(context.Background(), projID, "posts", 20, 9999999, "CreatedAt", "desc")
	wantNoErr(t, err)
	if page.Offset != maxPublicCollectionOffset {
		t.Fatalf("offset = %d, want %d", page.Offset, maxPublicCollectionOffset)
	}
	if gotLimit != 20+maxPublicCollectionOffset {
		t.Fatalf("limit passed to repo = %d, want %d", gotLimit, 20+maxPublicCollectionOffset)
	}
}

func TestPublicCollection_SlidingWindowLimiterEvictsAndCaps(t *testing.T) {
	limiter := newSlidingWindowLimiter()
	t0 := time.Unix(1000, 0)

	// Populate 50,000 expired entries
	for i := 0; i < maxRateLimitBuckets; i++ {
		if !limiter.Allow(fmt.Sprintf("expired-%d", i), 10, time.Minute, t0) {
			t.Fatalf("expected allow at %d", i)
		}
	}
	if len(limiter.buckets) != maxRateLimitBuckets {
		t.Fatalf("buckets = %d, want %d", len(limiter.buckets), maxRateLimitBuckets)
	}

	// Next insert after window expiry should evict expired buckets rather than growing unbounded
	t1 := t0.Add(2 * time.Minute)
	if !limiter.Allow("fresh-1", 10, time.Minute, t1) {
		t.Fatal("expected allow for fresh-1")
	}
	if len(limiter.buckets) != 1 {
		t.Fatalf("expected expired buckets evicted, got %d buckets", len(limiter.buckets))
	}

	// Fill with active entries up to cap and verify it never exceeds maxRateLimitBuckets
	for i := 0; i < maxRateLimitBuckets+50; i++ {
		limiter.Allow(fmt.Sprintf("active-%d", i), 10, time.Minute, t1)
	}
	if len(limiter.buckets) > maxRateLimitBuckets {
		t.Fatalf("buckets exceeded cap: %d > %d", len(limiter.buckets), maxRateLimitBuckets)
	}
}

func TestPublicCollection_FieldMaxLength(t *testing.T) {
	fields := []models.ContentField{
		{Id: "f1", Name: "summary", Type: "text", Required: true},
		{Id: "f2", Name: "body", Type: "richtext", Required: false},
	}

	// 16 KB + 1 byte rejected for regular text field
	_, errs := ValidateSubmissionFields(fields, map[string]any{
		"summary": strings.Repeat("a", maxPublicFieldLen+1),
	})
	if errs["summary"] == "" {
		t.Fatal("expected length validation error for text field > 16KB")
	}

	// 100 KB allowed for richtext, 100 KB + 1 byte rejected
	_, errsOK := ValidateSubmissionFields(fields, map[string]any{
		"summary": "hello",
		"body":    strings.Repeat("b", maxPublicRichTextFieldLen),
	})
	if len(errsOK) != 0 {
		t.Fatalf("expected 100KB richtext allowed, got %v", errsOK)
	}

	_, errsTooBig := ValidateSubmissionFields(fields, map[string]any{
		"summary": "hello",
		"body":    strings.Repeat("b", maxPublicRichTextFieldLen+1),
	})
	if errsTooBig["body"] == "" {
		t.Fatal("expected length validation error for richtext field > 100KB")
	}
}
