package unit

import (
	"testing"

	domainEntity "my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/utils/validation"
)

const validUUID = "3fa85f64-5717-4562-b3fc-2c963f66afa6"

func TestValidatePresence(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := domainEntity.PayloadPresence{
			UserID:   validUUID,
			UserName: "Test User",
			CursorX:  100,
			CursorY:  200,
		}
		if err := validation.ValidatePresence(&p); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("missing_userId", func(t *testing.T) {
		p := domainEntity.PayloadPresence{UserID: ""}
		if err := validation.ValidatePresence(&p); err == nil {
			t.Fatalf("expected error for missing userId, got nil")
		}
	})
}

func TestValidateElementCreate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := domainEntity.PayloadElementCreate{
			Element: domainEntity.Element{
				PageID: "page-123",
				Type:   "button",
			},
		}
		if err := validation.ValidateElementCreate(&p); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid_element", func(t *testing.T) {
		p := domainEntity.PayloadElementCreate{
			Element: domainEntity.Element{
				PageID: "", // Missing PageID
				Type:   "", // Missing Type
			},
		}
		if err := validation.ValidateElementCreate(&p); err == nil {
			t.Fatalf("expected error for invalid element, got nil")
		}
	})
}

func TestValidateElementUpdate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := domainEntity.PayloadElementUpdate{
			ID: validUUID,
			Styles: map[string]any{
				"color": "red",
			},
		}
		if err := validation.ValidateElementUpdate(&p); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid_id", func(t *testing.T) {
		p := domainEntity.PayloadElementUpdate{ID: "not-a-uuid"}
		if err := validation.ValidateElementUpdate(&p); err == nil {
			t.Fatalf("expected error for invalid id, got nil")
		}
	})
}

func TestValidateElementMove(t *testing.T) {
	t.Run("valid_root_move", func(t *testing.T) {
		p := domainEntity.PayloadElementMove{
			ID:       validUUID,
			ParentID: nil,
			Order:    5,
		}
		if err := validation.ValidateElementMove(&p); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("valid_reparent", func(t *testing.T) {
		parentID := validUUID
		p := domainEntity.PayloadElementMove{
			ID:       validUUID,
			ParentID: &parentID,
			Order:    0,
		}
		if err := validation.ValidateElementMove(&p); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("negative_order", func(t *testing.T) {
		p := domainEntity.PayloadElementMove{
			ID:    validUUID,
			Order: -1,
		}
		if err := validation.ValidateElementMove(&p); err == nil {
			t.Fatalf("expected error for negative order, got nil")
		}
	})
}

func TestValidateElementDelete(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := domainEntity.PayloadElementDelete{ID: validUUID}
		if err := validation.ValidateElementDelete(&p); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("missing_id", func(t *testing.T) {
		p := domainEntity.PayloadElementDelete{ID: ""}
		if err := validation.ValidateElementDelete(&p); err == nil {
			t.Fatalf("expected error for missing id, got nil")
		}
	})
}

func TestValidateElementContent(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := domainEntity.PayloadElementContent{
			ID:      validUUID,
			Content: "Updated text",
		}
		if err := validation.ValidateElementContent(&p); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid_id", func(t *testing.T) {
		p := domainEntity.PayloadElementContent{ID: "invalid"}
		if err := validation.ValidateElementContent(&p); err == nil {
			t.Fatalf("expected error for invalid id, got nil")
		}
	})
}

func TestValidateComment(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := domainEntity.PayloadComment{
			ElementID: validUUID,
			Comment: domainEntity.ElementComment{
				Content: "This looks great!",
			},
		}
		if err := validation.ValidateComment(&p); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("missing_content", func(t *testing.T) {
		p := domainEntity.PayloadComment{
			ElementID: validUUID,
			Comment:   domainEntity.ElementComment{Content: ""},
		}
		if err := validation.ValidateComment(&p); err == nil {
			t.Fatalf("expected error for missing content, got nil")
		}
	})
}

func TestValidateElementForAdd(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		e := domainEntity.Element{
			PageID: "doc-123",
			Type:   "container",
			Order:  0,
		}
		if err := validation.ValidateElementForAdd(&e); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("missing_page_id", func(t *testing.T) {
		e := domainEntity.Element{
			Type: "button",
		}
		if err := validation.ValidateElementForAdd(&e); err == nil {
			t.Fatalf("expected error for missing pageId, got nil")
		}
	})
}

func TestValidateElementForUpdate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		e := domainEntity.Element{
			ID:    validUUID,
			Type:  "text",
			Order: 10,
		}
		if err := validation.ValidateElementForUpdate(&e); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("missing_id", func(t *testing.T) {
		e := domainEntity.Element{
			ID:   "",
			Type: "text",
		}
		if err := validation.ValidateElementForUpdate(&e); err == nil {
			t.Fatalf("expected error for missing id, got nil")
		}
	})
}
