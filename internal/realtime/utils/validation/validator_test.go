package validation

import (
	"testing"

	domainEntity "my-go-app/internal/realtime/domain/entity"
)

func sPtr(s string) *string { return &s }

func TestValidateRealtimePayloads(t *testing.T) {
	const validUUID = "11111111-1111-4111-8111-111111111111"

	if ValidatePresence(nil) == nil || ValidatePresence(&domainEntity.PayloadPresence{}) == nil {
		t.Fatal("expected presence errors")
	}
	if err := ValidatePresence(&domainEntity.PayloadPresence{UserID: "u1"}); err != nil {
		t.Fatalf("unexpected presence error: %v", err)
	}

	if ValidateElementUpdate(nil) == nil || ValidateElementUpdate(&domainEntity.PayloadElementUpdate{ID: "bad"}) == nil {
		t.Fatal("expected update errors")
	}
	if err := ValidateElementUpdate(&domainEntity.PayloadElementUpdate{ID: validUUID}); err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	if ValidateElementMove(nil) == nil ||
		ValidateElementMove(&domainEntity.PayloadElementMove{ID: "bad"}) == nil ||
		ValidateElementMove(&domainEntity.PayloadElementMove{ID: validUUID, Order: -1}) == nil ||
		ValidateElementMove(&domainEntity.PayloadElementMove{ID: validUUID, Order: 0, ParentID: sPtr("not-uuid")}) == nil {
		t.Fatal("expected move errors")
	}
	if err := ValidateElementMove(&domainEntity.PayloadElementMove{ID: validUUID, Order: 1, ParentID: sPtr(validUUID)}); err != nil {
		t.Fatalf("unexpected move error: %v", err)
	}

	if ValidateElementDelete(nil) == nil || ValidateElementDelete(&domainEntity.PayloadElementDelete{ID: "bad"}) == nil {
		t.Fatal("expected delete errors")
	}
	if err := ValidateElementDelete(&domainEntity.PayloadElementDelete{ID: validUUID}); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	if ValidateElementContent(nil) == nil || ValidateElementContent(&domainEntity.PayloadElementContent{ID: "bad"}) == nil {
		t.Fatal("expected content errors")
	}
	if err := ValidateElementContent(&domainEntity.PayloadElementContent{ID: validUUID}); err != nil {
		t.Fatalf("unexpected content error: %v", err)
	}

	if ValidateComment(nil) == nil ||
		ValidateComment(&domainEntity.PayloadComment{ElementID: "bad"}) == nil ||
		ValidateComment(&domainEntity.PayloadComment{ElementID: validUUID}) == nil {
		t.Fatal("expected comment errors")
	}
	if err := ValidateComment(&domainEntity.PayloadComment{
		ElementID: validUUID,
		Comment:   domainEntity.ElementComment{Content: "Nice layout"},
	}); err != nil {
		t.Fatalf("unexpected comment error: %v", err)
	}
}

func TestValidateElementForAddAndUpdate(t *testing.T) {
	const validUUID = "11111111-1111-4111-8111-111111111111"

	if ValidateElementCreate(nil) == nil || ValidateElementForAdd(nil) == nil || ValidateElementForUpdate(nil) == nil {
		t.Fatal("expected nil element errors")
	}

	el := &domainEntity.Element{Type: "Text", PageID: "page-1", ParentID: sPtr("root")}
	if err := ValidateElementCreate(&domainEntity.PayloadElementCreate{Element: *el}); err != nil {
		t.Fatalf("unexpected root parent add error: %v", err)
	}

	if err := ValidateElementForAdd(el); err != nil {
		t.Fatalf("unexpected add error: %v", err)
	}
	if el.ParentID == nil || *el.ParentID != "root" {
		t.Fatalf("ParentID should be restored after validation, got %v", el.ParentID)
	}

	badEl := &domainEntity.Element{Type: "Text", PageID: "page-1", ID: "bad-id", ParentID: sPtr("")}
	if ValidateElementForAdd(badEl) == nil {
		t.Fatal("expected invalid id error")
	}
	if badEl.ParentID == nil || *badEl.ParentID != "" {
		t.Fatalf("ParentID should be restored on validation failure, got %v", badEl.ParentID)
	}

	if ValidateElementForUpdate(&domainEntity.Element{ID: "bad", Type: "Text"}) == nil {
		t.Fatal("expected invalid update id")
	}
	if err := ValidateElementForUpdate(&domainEntity.Element{ID: validUUID, Type: "Text"}); err != nil {
		t.Fatalf("unexpected valid update error: %v", err)
	}
}
