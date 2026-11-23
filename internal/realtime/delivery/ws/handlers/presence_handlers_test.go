package handlers

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	wsentity "my-go-app/internal/realtime/delivery/ws/entity"
	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
)

func TestJoin_RequiresPageID(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.sendPage("join", "", map[string]any{}, "", "")
	c.expectError("VALIDATION_ERROR", "pageId is required to join")
	if len(env.els.synced) != 0 {
		t.Fatal("rejected join must not sync")
	}
}

func TestJoin_SendsPageSyncThenProjectSyncEchoingRequestID(t *testing.T) {
	env := handlerTestNew(t, newLogicHandlersFake(), nil)
	env.els.seq = 12
	env.els.pageElems = []domainEntity.Element{{ID: handlerTestElemA, Type: "text", PageID: handlerTestPage}}
	env.els.projElems = []domainEntity.Element{
		{ID: handlerTestElemA, Type: "text", PageID: handlerTestPage},
		{ID: handlerTestElemB, Type: "div", PageID: handlerTestPage2},
	}
	c := env.dial("u1")
	msgs := c.join(handlerTestPage)
	if len(msgs) != 3 || msgs[0].Type != "sync:page" || msgs[1].Type != "logic:sync" || msgs[2].Type != "sync:project" {
		t.Fatalf("join replies = %+v", msgs)
	}
	for _, m := range msgs {
		if m.RequestID != "join-"+handlerTestPage {
			t.Fatalf("%s lost the requestId: %q", m.Type, m.RequestID)
		}
	}
	var page domainEntity.PayloadSync
	msgs[0].decode(t, &page)
	if page.Seq != 12 || len(page.Elements) != 1 || page.Elements[0].ID != handlerTestElemA {
		t.Fatalf("page sync %+v", page)
	}
	if len(page.Users) != 1 || page.Users[0].UserID != "u1" || page.Users[0].UserName != "u1@example.com" {
		t.Fatalf("page sync must list the joiner as a participant: %+v", page.Users)
	}
	var proj domainEntity.PayloadSyncProject
	msgs[2].decode(t, &proj)
	if len(proj.Elements) != 2 {
		t.Fatalf("project sync %+v", proj)
	}
	if got := env.els.synced[0]; got.project != handlerTestProject || got.page != handlerTestPage {
		t.Fatalf("SyncPageWithSeq called with %+v", got)
	}
}

func TestJoin_PageSyncErrorIsReported(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	env.els.syncErr = errors.New("cache miss storm")
	c := env.dial("u1")
	c.sendPage("join", handlerTestPage, map[string]any{}, "j", "")
	m, _ := c.expectError("PROCESS_ERROR", "cache miss storm")
	if m.ClientOpID != "j" {
		t.Fatalf("clientOpId = %q", m.ClientOpID)
	}
}

func TestJoin_ProjectSyncFailureIsReported(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	env.els.projErr = errors.New("project read failed")
	c := env.dial("u1")
	c.sendPage("join", handlerTestPage, map[string]any{}, "j", "")
	c.expectError("PROCESS_ERROR", "project read failed")
}

func TestJoin_SecondUserSeesBothParticipants(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	a, b := env.dial("alice"), env.dial("bob")
	a.join(handlerTestPage)
	msgs := b.join(handlerTestPage)
	var page domainEntity.PayloadSync
	msgs[0].decode(t, &page)
	ids := map[string]bool{}
	for _, u := range page.Users {
		ids[u.UserID] = true
	}
	if len(page.Users) != 2 || !ids["alice"] || !ids["bob"] {
		t.Fatalf("users = %+v", page.Users)
	}
}

// presenceHandlersWaitClosed reads until the server closes the socket.
func presenceHandlersWaitClosed(t *testing.T, c *handlerConn) {
	t.Helper()
	_ = c.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				t.Fatal("connection was not closed")
			}
			return
		}
	}
}

func presenceHandlersParticipants(t *testing.T, c *handlerConn) []domainEntity.UserPresence {
	t.Helper()
	c.send("sync:page", map[string]any{}, "", "s")
	var page domainEntity.PayloadSync
	c.next("sync:page").decode(t, &page)
	return page.Users
}

func TestJoin_SwitchingPagesKeepsConnectionOpen(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	a, b := env.dial("alice"), env.dial("bob")
	a.join(handlerTestPage)
	b.join(handlerTestPage)

	a.sendPage("join", handlerTestPage2, map[string]any{}, "", "")
	a.next("sync:page")

	if users := presenceHandlersParticipants(t, b); len(users) != 1 || users[0].UserID != "bob" {
		t.Fatalf("old room participants = %+v", users)
	}
	b.send("element:delete", map[string]any{"id": handlerTestElemA}, "d", "")
	if m := b.next("element:delete"); m.ClientOpID != "d" {
		t.Fatalf("remaining client must keep working: %+v", m)
	}
}

func TestLeave_RemovesClientFromRoomAndKeepsConnectionOpen(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	a, b := env.dial("alice"), env.dial("bob")
	a.join(handlerTestPage)
	b.join(handlerTestPage)

	a.send("leave", map[string]any{}, "", "")
	a.join(handlerTestPage2)

	if users := presenceHandlersParticipants(t, b); len(users) != 1 || users[0].UserID != "bob" {
		t.Fatalf("participants after leave = %+v", users)
	}
}

func TestLeave_ClearsPageSoLaterRequestsNeedAJoin(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	client := wsentity.NewClient(nil, handlerTestProject, &domainEntity.User{ID: "u1"})
	client.SetPageID(handlerTestPage)
	if err := env.h.handleLeave(context.Background(), client, domainEntity.Message{}); err != nil {
		t.Fatal(err)
	}
	if client.PageID != "" {
		t.Fatalf("PageID = %q after leave", client.PageID)
	}
	if err := env.h.handleSync(context.Background(), client, domainEntity.Message{}); !errors.Is(err, domainErrors.ErrNotJoined) {
		t.Fatalf("sync after leave: %v", err)
	}
}

func TestLeave_WithoutJoinIsANoop(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.send("leave", map[string]any{}, "", "")
	if d := c.drain(); len(d) != 0 {
		t.Fatalf("unexpected %+v", d)
	}
}

func TestPresence_GoesToOthersNotSenderAndIsRemembered(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	a, b := env.dial("alice"), env.dial("bob")
	a.join(handlerTestPage)
	b.join(handlerTestPage)

	a.send("presence", map[string]any{"userId": "alice", "userName": "Al", "cursorX": 10.5, "cursorY": 20, "elementId": handlerTestElemA}, "", "")
	m := b.next("presence")
	var p domainEntity.PayloadPresence
	m.decode(t, &p)
	if p.UserID != "alice" || p.CursorX != 10.5 || p.CursorY != 20 || p.ElementID != handlerTestElemA {
		t.Fatalf("presence = %+v", p)
	}
	if d := a.drain(); len(handlerOfType(d, "presence")) != 0 {
		t.Fatalf("sender must not get its own presence: %+v", d)
	}

	// A later sync reflects the stored cursor.
	b.send("sync:page", map[string]any{}, "", "s")
	var page domainEntity.PayloadSync
	b.next("sync:page").decode(t, &page)
	for _, u := range page.Users {
		if u.UserID == "alice" {
			if u.Presence.CursorX != 10.5 || u.Presence.CursorY != 20 {
				t.Fatalf("stored presence = %+v", u.Presence)
			}
			return
		}
	}
	t.Fatalf("alice missing from %+v", page.Users)
}

func TestPresence_MalformedPayloadIsAValidationError(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.join(handlerTestPage)
	c.send("presence", "garbage", "", "")
	c.expectError("VALIDATION_ERROR", "")
}

func TestSync_RepliesWithPageStateAndRequestID(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.join(handlerTestPage)
	env.els.mu.Lock()
	env.els.seq = 5
	env.els.pageElems = []domainEntity.Element{{ID: handlerTestElemB, Type: "div", PageID: handlerTestPage}}
	env.els.mu.Unlock()
	c.send("sync:page", map[string]any{}, "", "again")
	m := c.next("sync:page")
	var page domainEntity.PayloadSync
	m.decode(t, &page)
	if m.RequestID != "again" || page.Seq != 5 || len(page.Elements) != 1 || page.Elements[0].ID != handlerTestElemB {
		t.Fatalf("%+v %+v", m, page)
	}
}

func TestSync_ErrorsAndRequiresJoin(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.send("sync:page", map[string]any{}, "", "")
	c.expectError("PROCESS_ERROR", "not joined")
	c.join(handlerTestPage)
	env.els.mu.Lock()
	env.els.syncErr = errors.New("sync failed")
	env.els.mu.Unlock()
	c.send("sync:page", map[string]any{}, "", "")
	c.expectError("PROCESS_ERROR", "sync failed")
}

func TestSyncProject_ReturnsAllProjectElementsWithoutJoining(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	env.els.projElems = []domainEntity.Element{
		{ID: handlerTestElemA, Type: "text", PageID: handlerTestPage},
		{ID: handlerTestElemB, Type: "div", PageID: handlerTestPage2},
	}
	c := env.dial("u1")
	c.send("sync:project", map[string]any{}, "", "sp")
	m := c.next("sync:project")
	var p domainEntity.PayloadSyncProject
	m.decode(t, &p)
	if m.RequestID != "sp" || len(p.Elements) != 2 {
		t.Fatalf("%+v %+v", m, p)
	}
}

func TestSyncProject_ErrorIsPropagated(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	env.els.projErr = errors.New("scan failed")
	c := env.dial("u1")
	c.send("sync:project", map[string]any{}, "", "")
	c.expectError("PROCESS_ERROR", "scan failed")
}
