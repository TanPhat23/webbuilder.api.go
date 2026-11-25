package handlers

import (
	"errors"
	"reflect"
	"testing"

	domainEntity "my-go-app/internal/realtime/domain/entity"
)

func elementHandlersStr(s string) *string { return &s }

// elementHandlersPair connects two clients of different users to the same page.
func elementHandlersPair(t *testing.T) (*handlerTestEnv, *handlerConn, *handlerConn) {
	t.Helper()
	env := handlerTestNew(t, nil, nil)
	a, b := env.dial("alice"), env.dial("bob")
	a.join(handlerTestPage)
	b.join(handlerTestPage)
	return env, a, b
}

func assertElementCreateBroadcasts(t *testing.T, clients map[string]*handlerConn) {
	t.Helper()
	for name, c := range clients {
		m := c.next("element:create")
		if m.Seq != 1 || m.ClientOpID != "op-1" || m.RequestID != "req-1" {
			t.Fatalf("%s: %+v", name, m)
		}
		var p domainEntity.PayloadElementCreate
		m.decode(t, &p)
		if p.Element.ID != handlerTestElemA || p.Element.Type != "text" {
			t.Fatalf("%s payload: %+v", name, p.Element)
		}
		if p.Element.Name == nil || p.Element.Styles == nil {
			t.Fatalf("%s: broadcast element must be sanitized (name/styles present)", name)
		}
	}
}

func TestElementCreate_BroadcastsToEveryoneWithSeqAndIDs(t *testing.T) {
	env, a, b := elementHandlersPair(t)
	a.send("element:create", map[string]any{
		"element":  map[string]any{"type": "text", "pageId": "ignored-page", "id": handlerTestElemA},
		"position": 3,
	}, "op-1", "req-1")

	assertElementCreateBroadcasts(t, map[string]*handlerConn{"sender": a, "other": b})
	if len(env.els.created) != 1 {
		t.Fatalf("created %d", len(env.els.created))
	}
	call := env.els.created[0]
	if call.project != handlerTestProject || call.el.PageID != handlerTestPage {
		t.Fatalf("call project=%q page=%q", call.project, call.el.PageID)
	}
	if call.pos == nil || *call.pos != 3 {
		t.Fatalf("position = %v", call.pos)
	}
}

func TestElementCreate_RootParentIsAcceptedAndPassedThrough(t *testing.T) {
	env, a, peer := elementHandlersPair(t)
	if peer == nil {
		t.Fatal("missing peer")
	}
	a.send("element:create", map[string]any{"element": map[string]any{"type": "div", "parentId": "root"}}, "", "")
	a.next("element:create")
	if p := env.els.created[0].el.ParentID; p == nil || *p != "root" {
		t.Fatalf("parentId = %v, want root preserved", p)
	}
}

func TestElementCreate_ValidationErrors(t *testing.T) {
	env, a, b := elementHandlersPair(t)
	cases := []struct {
		name    string
		payload any
	}{
		{"malformed payload", "not an object"},
		{"missing type", map[string]any{"element": map[string]any{"pageId": handlerTestPage}}},
		{"bad parent uuid", map[string]any{"element": map[string]any{"type": "text", "parentId": "nope"}}},
		{"negative order", map[string]any{"element": map[string]any{"type": "text", "order": -1}}},
	}
	for _, tc := range cases {
		a.send("element:create", tc.payload, "op", "")
		a.expectError("VALIDATION_ERROR", "")
		_ = tc.name
	}
	if len(env.els.created) != 0 {
		t.Fatalf("invalid creates reached the use case: %d", len(env.els.created))
	}
	if d := b.drain(); len(d) != 0 {
		t.Fatalf("invalid creates must not broadcast: %+v", d)
	}
}

func TestElementCreate_RequiresPageWhenNotJoined(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.sendPage("element:create", "", map[string]any{"element": map[string]any{"type": "text"}}, "", "")
	c.expectError("PROCESS_ERROR", "not joined")
	if len(env.els.created) != 0 {
		t.Fatal("use case must not be called")
	}
}

func TestElementCreate_UseCaseErrorGoesToSenderOnlyAndConsumesNoSeq(t *testing.T) {
	env, a, b := elementHandlersPair(t)
	env.els.opErr = errors.New("db write failed")
	a.send("element:create", map[string]any{"element": map[string]any{"type": "text"}}, "op-bad", "")
	m, p := a.expectError("PROCESS_ERROR", "internal error")
	if m.ClientOpID != "op-bad" || p.ClientOpID != "op-bad" {
		t.Fatalf("clientOpId not echoed: %+v %+v", m, p)
	}
	env.els.mu.Lock()
	env.els.opErr = nil
	env.els.mu.Unlock()
	a.send("element:create", map[string]any{"element": map[string]any{"type": "text"}}, "op-ok", "")
	// b's first element:create must be the successful one: nothing was broadcast for the failure.
	if m := b.next("element:create"); m.ClientOpID != "op-ok" || m.Seq != 1 {
		t.Fatalf("other client saw %+v", m)
	}
}

func assertElementUpdateEcho(t *testing.T, clients []*handlerConn) {
	t.Helper()
	for _, c := range clients {
		m := c.next("element:update")
		if m.Seq != 1 || m.ClientOpID != "op-u" || m.RequestID != "req-u" {
			t.Fatalf("%+v", m)
		}
		var p domainEntity.PayloadElementUpdate
		m.decode(t, &p)
		if p.ID != handlerTestElemA || p.Styles["color"] != "red" {
			t.Fatalf("payload %+v", p)
		}
	}
}

func TestElementUpdate_MergesTopLevelFieldsAndProperties(t *testing.T) {
	env, a, b := elementHandlersPair(t)
	a.send("element:update", map[string]any{
		"id":             handlerTestElemA,
		"name":           "N",
		"content":        "C",
		"src":            "S",
		"href":           "H",
		"tailwindStyles": "T",
		"properties":     map[string]any{"content": "explicit", "isLocked": true},
		"styles":         map[string]any{"color": "red"},
		"settings":       map[string]any{"k": nil},
	}, "op-u", "req-u")

	assertElementUpdateEcho(t, []*handlerConn{a, b})
	got := env.els.patched[0]
	wantProps := map[string]any{"name": "N", "content": "explicit", "src": "S", "href": "H", "tailwindStyles": "T", "isLocked": true}
	if got.project != handlerTestProject || got.page != handlerTestPage || got.id != handlerTestElemA {
		t.Fatalf("call %+v", got)
	}
	if !reflect.DeepEqual(got.props, wantProps) || got.styles["color"] != "red" {
		t.Fatalf("props=%v styles=%v", got.props, got.styles)
	}
	if v, ok := got.settings["k"]; !ok || v != nil {
		t.Fatalf("settings null must be forwarded, got %v", got.settings)
	}
}

func TestElementUpdate_Errors(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.send("element:update", map[string]any{"id": handlerTestElemA}, "", "")
	c.expectError("PROCESS_ERROR", "not joined")

	c.join(handlerTestPage)
	c.send("element:update", "garbage", "", "")
	c.expectError("VALIDATION_ERROR", "")
	c.send("element:update", map[string]any{"id": "not-a-uuid"}, "", "")
	c.expectError("VALIDATION_ERROR", "invalid element id")
	c.send("element:update", map[string]any{}, "", "")
	c.expectError("VALIDATION_ERROR", "invalid element id")

	env.els.opErr = errors.New("conflict")
	c.send("element:update", map[string]any{"id": handlerTestElemA}, "op", "")
	c.expectError("PROCESS_ERROR", "internal error")
}

func TestUpdateProperties(t *testing.T) {
	if got := extractElementProperties(domainEntity.PayloadElementUpdate{}); len(got) != 0 {
		t.Fatalf("empty payload gave %v", got)
	}
	empty := elementHandlersStr("")
	got := extractElementProperties(domainEntity.PayloadElementUpdate{
		Name:       empty,
		Content:    elementHandlersStr("c"),
		Properties: map[string]any{"name": "from-props", "extra": 1},
	})
	want := map[string]any{"name": "from-props", "content": "c", "extra": 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	// An explicitly empty string is a value, not "absent".
	got = extractElementProperties(domainEntity.PayloadElementUpdate{Href: empty})
	if v, ok := got["href"]; !ok || v != "" {
		t.Fatalf("empty href must be kept: %v", got)
	}
}

func TestElementDelete(t *testing.T) {
	env, a, b := elementHandlersPair(t)
	a.send("element:delete", map[string]any{"id": handlerTestElemA}, "op-d", "req-d")
	for _, c := range []*handlerConn{a, b} {
		m := c.next("element:delete")
		var p domainEntity.PayloadElementDelete
		m.decode(t, &p)
		if m.Seq != 1 || m.ClientOpID != "op-d" || m.RequestID != "req-d" || p.ID != handlerTestElemA {
			t.Fatalf("%+v %+v", m, p)
		}
	}
	if got := env.els.removed[0]; got != (handlerRefCall{handlerTestProject, handlerTestPage, handlerTestElemA}) {
		t.Fatalf("call %+v", got)
	}

	a.send("element:delete", "garbage", "", "")
	a.expectError("VALIDATION_ERROR", "")

	env.els.opErr = errors.New("nope")
	a.send("element:delete", map[string]any{"id": handlerTestElemB}, "op-e", "")
	a.expectError("PROCESS_ERROR", "internal error")
}

func TestElementDelete_RequiresJoin(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.send("element:delete", map[string]any{"id": handlerTestElemA}, "", "")
	c.expectError("PROCESS_ERROR", "not joined")
	if len(env.els.removed) != 0 {
		t.Fatal("use case must not be called")
	}
}

func assertElementMoveBroadcasts(t *testing.T, clients []*handlerConn, parent string) {
	t.Helper()
	for _, c := range clients {
		m := c.next("element:move")
		var p domainEntity.PayloadElementMove
		m.decode(t, &p)
		if m.Seq != 1 || m.ClientOpID != "op-m" || m.RequestID != "req-m" {
			t.Fatalf("%+v", m)
		}
		if p.Order != 9 || p.OrderKey != "a9" {
			t.Fatalf("broadcast order/key mismatch: %+v", p)
		}
		if p.ParentID == nil || *p.ParentID != parent {
			t.Fatalf("broadcast parent mismatch: %+v", p)
		}
	}
}

func TestElementMove_BroadcastsTheUseCasesResolvedPosition(t *testing.T) {
	env, a, b := elementHandlersPair(t)
	parent := handlerTestElemB
	env.els.moveRes = &domainEntity.Element{ID: handlerTestElemA, Order: 9, OrderKey: "a9", ParentID: &parent}
	a.send("element:move", map[string]any{"id": handlerTestElemA, "parentId": parent, "order": 2}, "op-m", "req-m")
	assertElementMoveBroadcasts(t, []*handlerConn{a, b}, parent)
	if got := env.els.moved[0]; got != (handlerMoveCall{handlerTestProject, handlerTestPage, handlerTestElemA, parent, 2}) {
		t.Fatalf("call %+v", got)
	}
}

func TestElementMove_NullParentMovesToRoot(t *testing.T) {
	env, a, peer := elementHandlersPair(t)
	if peer == nil {
		t.Fatal("missing peer")
	}
	a.send("element:move", map[string]any{"id": handlerTestElemA, "parentId": nil, "order": 0}, "", "")
	a.next("element:move")
	if got := env.els.moved[0]; got.parent != "" || got.index != 0 {
		t.Fatalf("call %+v", got)
	}
}

func TestElementMove_Errors(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.send("element:move", map[string]any{"id": handlerTestElemA}, "", "")
	c.expectError("PROCESS_ERROR", "not joined")

	c.join(handlerTestPage)
	c.send("element:move", "garbage", "", "")
	c.expectError("VALIDATION_ERROR", "")
	c.send("element:move", map[string]any{"id": "bad", "order": 0}, "", "")
	c.expectError("VALIDATION_ERROR", "invalid element id")
	c.send("element:move", map[string]any{"id": handlerTestElemA, "order": -2}, "", "")
	c.expectError("VALIDATION_ERROR", "non-negative")
	c.send("element:move", map[string]any{"id": handlerTestElemA, "order": 0, "parentId": "zzz"}, "", "")
	c.expectError("VALIDATION_ERROR", "invalid parentId")
	if len(env.els.moved) != 0 {
		t.Fatalf("invalid moves reached the use case: %d", len(env.els.moved))
	}

	env.els.opErr = errors.New("cycle")
	c.send("element:move", map[string]any{"id": handlerTestElemA, "order": 1}, "op", "")
	c.expectError("PROCESS_ERROR", "internal error")
}

func TestElementOps_SequenceNumbersAreContiguousAcrossOpTypes(t *testing.T) {
	_, a, b := elementHandlersPair(t)
	a.send("element:create", map[string]any{"element": map[string]any{"type": "text"}}, "1", "")
	a.send("element:update", map[string]any{"id": handlerTestElemA}, "2", "")
	a.send("element:move", map[string]any{"id": handlerTestElemA, "order": 0}, "3", "")
	a.send("element:delete", map[string]any{"id": handlerTestElemA}, "4", "")
	for want := int64(1); want <= 4; want++ {
		var m handlerMsg
		for {
			m = b.read()
			if m.Seq != 0 {
				break
			}
		}
		if m.Seq != want {
			t.Fatalf("seq = %d, want %d (%s)", m.Seq, want, m.Type)
		}
	}
}
