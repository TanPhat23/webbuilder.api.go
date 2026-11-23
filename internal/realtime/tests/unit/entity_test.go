package unit

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	wsentity "my-go-app/internal/realtime/delivery/ws/entity"
	domainEntity "my-go-app/internal/realtime/domain/entity"
)

// mockLogger implements logger.Logger for testing purposes.
type mockLogger struct{}

func (m *mockLogger) Info(args ...interface{})                  {}
func (m *mockLogger) Infof(format string, args ...interface{})  {}
func (m *mockLogger) Error(args ...interface{})                 {}
func (m *mockLogger) Errorf(format string, args ...interface{}) {}
func (m *mockLogger) Debug(args ...interface{})                 {}
func (m *mockLogger) Debugf(format string, args ...interface{}) {}
func (m *mockLogger) Fatal(args ...interface{})                 {}
func (m *mockLogger) Fatalf(format string, args ...interface{}) {}

func TestClient_RoomID(t *testing.T) {
	cases := []struct {
		name      string
		projectID string
		pageID    string
		want      string
	}{
		{"both IDs present", "proj-1", "page-1", "proj-1:page-1"},
		{"only project ID", "proj-1", "", "proj-1"},
		{"only page ID", "", "page-1", "page-1"},
		{"no IDs", "", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// websocket.Conn is nil here as we are only testing RoomID logic
			c := wsentity.NewClient(nil, tc.projectID, nil)
			c.SetPageID(tc.pageID)
			if got := c.RoomID(); got != tc.want {
				t.Errorf("RoomID() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMessageFactory_CreateWithPayload(t *testing.T) {
	projectID, pageID := "test-project", "test-page"
	f := wsentity.NewMessageFactory(projectID, pageID)
	payload := domainEntity.PayloadPresence{UserID: "user-123", UserName: "Tester"}
	msg, err := f.Create(domainEntity.TypePresence, payload)
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}
	if msg.Type != domainEntity.TypePresence || msg.ProjectID != projectID {
		t.Errorf("unexpected header: %+v", msg)
	}
	if msg.PageID != pageID {
		t.Errorf("expected pageID %s, got %s", pageID, msg.PageID)
	}
	var decoded domainEntity.PayloadPresence
	if err := json.Unmarshal(msg.Payload, &decoded); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if decoded.UserID != payload.UserID || decoded.UserName != payload.UserName {
		t.Errorf("payload data mismatch")
	}
}

func TestMessageFactory_CreateWithoutPayload(t *testing.T) {
	f := wsentity.NewMessageFactory("test-project", "test-page")
	msg, err := f.Create(domainEntity.TypeLeave, nil)
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}
	if msg.Payload != nil {
		t.Errorf("expected nil payload, got %v", msg.Payload)
	}
}

func TestMessageFactory_Helpers(t *testing.T) {
	f := wsentity.NewMessageFactory("p1", "pg1")
	id := "el-delete-id"
	msg, err := f.DeleteElement(id)
	if err != nil || msg.Type != domainEntity.TypeElementDelete {
		t.Fatalf("DeleteElement failed: %v %+v", err, msg)
	}
	var delPayload domainEntity.PayloadElementDelete
	if err := json.Unmarshal(msg.Payload, &delPayload); err != nil {
		t.Fatalf("unmarshal delete payload: %v", err)
	}
	if delPayload.ID != id {
		t.Errorf("expected id %s, got %s", id, delPayload.ID)
	}

	errMsg := f.Error("TEST_ERROR", "something went wrong")
	if errMsg.Type != domainEntity.TypeError {
		t.Errorf("wrong type: %s", errMsg.Type)
	}
	var errPayload domainEntity.PayloadError
	if err := json.Unmarshal(errMsg.Payload, &errPayload); err != nil {
		t.Fatalf("unmarshal error payload: %v", err)
	}
	if errPayload.Code != "TEST_ERROR" || errPayload.Message != "something went wrong" {
		t.Errorf("unexpected error payload: %+v", errPayload)
	}
}

func TestDecodePayload(t *testing.T) {
	data := []byte(`{"id": "element-1", "order": 5}`)
	msg := domainEntity.Message{
		Payload: json.RawMessage(data),
	}

	decoded, err := wsentity.DecodePayload[domainEntity.PayloadElementDelete](msg)
	if err != nil {
		t.Fatalf("DecodePayload failed: %v", err)
	}

	if decoded.ID != "element-1" {
		t.Errorf("expected ID element-1, got %s", decoded.ID)
	}
}

func assertPresenceEntry(t *testing.T, u domainEntity.UserPresence, id, name string, cx, cy float64, wantElem *string) {
	t.Helper()
	if u.UserID != id || u.UserName != name {
		t.Fatalf("unexpected user identity: %+v", u)
	}
	if u.Presence.CursorX != cx || u.Presence.CursorY != cy {
		t.Fatalf("unexpected coords: %+v", u.Presence)
	}
	if wantElem == nil {
		if u.Presence.ElementID != nil {
			t.Fatalf("expected nil elementId, got %+v", u.Presence)
		}
		return
	}
	if u.Presence.ElementID == nil || *u.Presence.ElementID != *wantElem {
		t.Fatalf("expected elementId %q, got %+v", *wantElem, u.Presence)
	}
}

func TestMessageFactory_SyncPage_SerializesUserPresence(t *testing.T) {
	f := wsentity.NewMessageFactory("p1", "pg1")
	elem := "elem-42"
	users := []domainEntity.UserPresence{
		{UserID: "user-1", UserName: "Tester", Presence: domainEntity.Presence{CursorX: 0.12, CursorY: 0.34}},
		{UserID: "user-2", UserName: "Bob", Presence: domainEntity.Presence{CursorX: 0.56, CursorY: 0.78, ElementID: &elem}},
	}

	msg, err := f.SyncPage(nil, nil, users)
	if err != nil {
		t.Fatalf("SyncPage failed: %v", err)
	}
	var payload domainEntity.PayloadSync
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if len(payload.Users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(payload.Users))
	}
	assertPresenceEntry(t, payload.Users[0], "user-1", "Tester", 0.12, 0.34, nil)
	assertPresenceEntry(t, payload.Users[1], "user-2", "Bob", 0.56, 0.78, &elem)
}

func TestHub_Basic(t *testing.T) {
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lg := &mockLogger{}

	hub := wsentity.NewHub(ctx, wg, 64, nil, lg)
	if hub == nil {
		t.Fatal("failed to create hub")
	}

	t.Run("get participants of non-existent room", func(t *testing.T) {
		users := hub.GetRoomParticipants("none")
		if len(users) != 0 {
			t.Errorf("expected 0 users, got %d", len(users))
		}
	})

	t.Run("register nil client", func(t *testing.T) {
		err := hub.Register(ctx, nil)
		if err != nil {
			t.Errorf("expected no error registering nil client, got %v", err)
		}
	})

	t.Run("unregister nil client", func(t *testing.T) {
		err := hub.Unregister(ctx, nil)
		if err != nil {
			t.Errorf("expected no error unregistering nil client, got %v", err)
		}
	})
}

func TestRoom_Initialization(t *testing.T) {
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	lg := &mockLogger{}
	roomID := "room-123"

	room := wsentity.NewRoom(roomID, ctx, wg, 10, lg)
	if room == nil {
		t.Fatal("failed to create room")
	}

	if room.PageID != roomID {
		t.Errorf("expected room ID %s, got %s", roomID, room.PageID)
	}

	t.Run("get participants from empty room", func(t *testing.T) {
		users := room.GetParticipants()
		if len(users) != 0 {
			t.Errorf("expected 0 participants, got %d", len(users))
		}
	})
}

func TestMessageFactory_SyncPage_SanitizesElements(t *testing.T) {
	f := wsentity.NewMessageFactory("p1", "pg1")

	elements := []domainEntity.Element{
		{
			ID:       "el-1",
			Type:     "text",
			Order:    0,
			IsLocked: false,
			PageID:   "pg1",
		},
		{
			ID:       "el-2",
			Type:     "container",
			Order:    1,
			IsLocked: false,
			PageID:   "pg1",
			Elements: []domainEntity.Element{
				{
					ID:       "el-2-1",
					Type:     "text",
					Order:    0,
					IsLocked: false,
					PageID:   "pg1",
				},
			},
		},
	}

	msg, err := f.SyncPage(elements, nil, nil)
	if err != nil {
		t.Fatalf("SyncPage failed: %v", err)
	}

	var payload domainEntity.PayloadSync
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		t.Fatalf("failed to unmarshal sync payload: %v", err)
	}

	if len(payload.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(payload.Elements))
	}

	// Top-level element should have content and styles set (non-nil)
	el := payload.Elements[0]
	if el.Content == nil {
		t.Errorf("expected element content to be set (non-nil), got nil")
	}
	if el.Styles == nil {
		t.Errorf("expected element styles to be set (non-nil), got nil")
	}

	// Nested child should also be sanitized
	if len(payload.Elements[1].Elements) != 1 {
		t.Fatalf("expected 1 child element, got %d", len(payload.Elements[1].Elements))
	}
	child := payload.Elements[1].Elements[0]
	if child.Content == nil {
		t.Errorf("expected child content to be set (non-nil), got nil")
	}
	if child.Styles == nil {
		t.Errorf("expected child styles to be set (non-nil), got nil")
	}
}
