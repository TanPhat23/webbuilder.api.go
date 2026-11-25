package validation

import (
	"testing"

	domainEntity "my-go-app/internal/realtime/domain/entity"
)

func sPtr(s string) *string { return &s }

func TestValidateRealtimePayloads(t *testing.T) {
	const validUUID = "11111111-1111-4111-8111-111111111111"

	if Presence(nil) == nil || Presence(&domainEntity.PayloadPresence{}) == nil {
		t.Fatal("expected presence errors")
	}
	if err := Presence(&domainEntity.PayloadPresence{UserID: "u1"}); err != nil {
		t.Fatalf("unexpected presence error: %v", err)
	}

	if ElementUpdate(nil) == nil || ElementUpdate(&domainEntity.PayloadElementUpdate{ID: "bad"}) == nil {
		t.Fatal("expected update errors")
	}
	if err := ElementUpdate(&domainEntity.PayloadElementUpdate{ID: validUUID}); err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	if ElementMove(nil) == nil ||
		ElementMove(&domainEntity.PayloadElementMove{ID: "bad"}) == nil ||
		ElementMove(&domainEntity.PayloadElementMove{ID: validUUID, Order: -1}) == nil ||
		ElementMove(&domainEntity.PayloadElementMove{ID: validUUID, Order: 0, ParentID: sPtr("not-uuid")}) == nil {
		t.Fatal("expected move errors")
	}
	if err := ElementMove(&domainEntity.PayloadElementMove{ID: validUUID, Order: 1, ParentID: sPtr(validUUID)}); err != nil {
		t.Fatalf("unexpected move error: %v", err)
	}

	if ElementDelete(nil) == nil || ElementDelete(&domainEntity.PayloadElementDelete{ID: "bad"}) == nil {
		t.Fatal("expected delete errors")
	}
	if err := ElementDelete(&domainEntity.PayloadElementDelete{ID: validUUID}); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	if ElementContent(nil) == nil || ElementContent(&domainEntity.PayloadElementContent{ID: "bad"}) == nil {
		t.Fatal("expected content errors")
	}
	if err := ElementContent(&domainEntity.PayloadElementContent{ID: validUUID}); err != nil {
		t.Fatalf("unexpected content error: %v", err)
	}

	if Comment(nil) == nil ||
		Comment(&domainEntity.PayloadComment{ElementID: "bad"}) == nil ||
		Comment(&domainEntity.PayloadComment{ElementID: validUUID}) == nil {
		t.Fatal("expected comment errors")
	}
	if err := Comment(&domainEntity.PayloadComment{
		ElementID: validUUID,
		Comment:   domainEntity.ElementComment{Content: "Nice layout"},
	}); err != nil {
		t.Fatalf("unexpected comment error: %v", err)
	}
}

func TestValidateElementForAddAndUpdate(t *testing.T) {
	const validUUID = "11111111-1111-4111-8111-111111111111"

	if ElementCreate(nil) == nil || ElementForCreate(nil) == nil || ElementForUpdate(nil) == nil {
		t.Fatal("expected nil element errors")
	}

	el := &domainEntity.Element{Type: "Text", PageID: "page-1", ParentID: sPtr("root")}
	if err := ElementCreate(&domainEntity.PayloadElementCreate{Element: *el}); err != nil {
		t.Fatalf("unexpected root parent add error: %v", err)
	}

	if err := ElementForCreate(el); err != nil {
		t.Fatalf("unexpected add error: %v", err)
	}
	if el.ParentID == nil || *el.ParentID != "root" {
		t.Fatalf("ParentID should be restored after validation, got %v", el.ParentID)
	}

	badEl := &domainEntity.Element{Type: "Text", PageID: "page-1", ID: "bad-id", ParentID: sPtr("")}
	if ElementForCreate(badEl) == nil {
		t.Fatal("expected invalid id error")
	}
	if badEl.ParentID == nil || *badEl.ParentID != "" {
		t.Fatalf("ParentID should be restored on validation failure, got %v", badEl.ParentID)
	}

	if ElementForUpdate(&domainEntity.Element{ID: "bad", Type: "Text"}) == nil {
		t.Fatal("expected invalid update id")
	}
	if err := ElementForUpdate(&domainEntity.Element{ID: validUUID, Type: "Text"}); err != nil {
		t.Fatalf("unexpected valid update error: %v", err)
	}
}

func TestValidateSafeIDAndBounds(t *testing.T) {
	for _, okID := range []string{"proj1", "p1", "11111111-1111-1111-1111-111111111111", "a_b-c"} {
		if !IsSafeID(okID) {
			t.Fatalf("expected %q to be valid", okID)
		}
	}
	for _, badID := range []string{"", "a:b", "a/b", "a b", "../x"} {
		if IsSafeID(badID) {
			t.Fatalf("expected %q to be invalid", badID)
		}
	}

	// Element depth > MaxElementTreeDepth must fail.
	deep := domainEntity.Element{Type: "Text", PageID: "page-1"}
	for i := 1; i < MaxElementTreeDepth+1; i++ {
		deep = domainEntity.Element{Type: "Frame", PageID: "page-1", Elements: []domainEntity.Element{deep}}
	}
	if err := ElementForCreate(&deep); err == nil {
		t.Fatal("expected depth limit error")
	}

	// Node count > MaxElementTreeNodes must fail.
	children := make([]domainEntity.Element, MaxElementTreeNodes)
	for i := range children {
		children[i] = domainEntity.Element{Type: "Text", PageID: "page-1"}
	}
	wide := domainEntity.Element{Type: "Frame", PageID: "page-1", Elements: children}
	if err := ElementForCreate(&wide); err == nil {
		t.Fatal("expected node count limit error")
	}
}
