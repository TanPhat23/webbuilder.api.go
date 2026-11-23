package handlers

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	domainErrors "my-go-app/internal/realtime/domain/errors"
	"my-go-app/internal/realtime/domain/logic"
	domainUsecase "my-go-app/internal/realtime/domain/usecase"

	domainEntity "my-go-app/internal/realtime/domain/entity"
)

type logicHandlersUpdateCall struct {
	project, page, user string
	patch               map[string]any
	nullPaths           [][][]string
}

type logicHandlersMoveCall struct{ project, page, user, kind, name, to string }

type logicHandlersFake struct {
	domainUsecase.LogicUseCase
	mu      sync.Mutex
	err     error
	getErr  error
	project logic.Revisioned
	page    logic.Revisioned
	updates []logicHandlersUpdateCall
	moves   []logicHandlersMoveCall
	gets    []logicHandlersMoveCall
}

func newLogicHandlersFake() *logicHandlersFake {
	return &logicHandlersFake{
		project: logic.Revisioned{Doc: logic.Document{"scope": "project"}, Revision: 4},
		page:    logic.Revisioned{Doc: logic.Document{"scope": "page"}, Revision: 7},
	}
}

func (f *logicHandlersFake) Update(_ context.Context, project, page, user string, patch map[string]any, nullPaths ...[][]string) (logic.Revisioned, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.updates = append(f.updates, logicHandlersUpdateCall{project, page, user, patch, nullPaths})
	if f.err != nil {
		return logic.Revisioned{}, f.err
	}
	if page == "" {
		return f.project, nil
	}
	return f.page, nil
}

func (f *logicHandlersFake) Move(_ context.Context, project, page, user, kind, name, to string) (logic.Revisioned, logic.Revisioned, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.moves = append(f.moves, logicHandlersMoveCall{project, page, user, kind, name, to})
	if f.err != nil {
		return logic.Revisioned{}, logic.Revisioned{}, f.err
	}
	return f.project, f.page, nil
}

func (f *logicHandlersFake) Get(_ context.Context, project, page, user string) (logic.Revisioned, logic.Revisioned, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.gets = append(f.gets, logicHandlersMoveCall{project: project, page: page, user: user})
	if f.getErr != nil {
		return logic.Revisioned{}, logic.Revisioned{}, f.getErr
	}
	return f.project, f.page, nil
}

func TestLogic_Key(t *testing.T) {
	if got := logicActorKey("p"); got != "p:logic" {
		t.Fatalf("got %q", got)
	}
}

func TestLogic_UnavailableWhenNoLogicUseCase(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.join(handlerTestPage)
	for _, typ := range []string{"logic:update", "logic:move"} {
		c.send(typ, map[string]any{}, "op", "")
		c.expectError("INTERNAL_ERROR", "logic editing is not available")
	}
	// logic:sync silently does nothing (join relies on it).
	c.send("logic:sync", nil, "", "")
	if d := c.drain(); len(handlerOfType(d, "logic:sync")) != 0 || len(handlerOfType(d, "error")) != 0 {
		t.Fatalf("unexpected %+v", d)
	}
}

func TestLogicSync_RepliesWithBothDocumentsAndRevisions(t *testing.T) {
	lg := newLogicHandlersFake()
	env := handlerTestNew(t, lg, nil)
	c := env.dial("u1")
	c.join(handlerTestPage)
	c.send("logic:sync", nil, "", "sync-1")
	m := c.next("logic:sync")
	var p domainEntity.PayloadLogicSync
	m.decode(t, &p)
	if m.RequestID != "sync-1" || p.PageID != handlerTestPage || p.ProjectRevision != 4 || p.PageRevision != 7 ||
		p.Project["scope"] != "project" || p.Page["scope"] != "page" {
		t.Fatalf("%+v %+v", m, p)
	}
	last := lg.gets[len(lg.gets)-1]
	if last.project != handlerTestProject || last.page != handlerTestPage || last.user != "u1" {
		t.Fatalf("Get called with %+v", last)
	}
}

func TestLogicSync_Errors(t *testing.T) {
	lg := newLogicHandlersFake()
	env := handlerTestNew(t, lg, nil)
	c := env.dial("u1")
	c.send("logic:sync", nil, "", "")
	c.expectError("PROCESS_ERROR", "not joined")

	c.join(handlerTestPage)
	lg.getErr = domainErrors.New(domainErrors.CodeAccessDenied, "no logic for you", nil)
	c.send("logic:sync", nil, "", "")
	c.expectError("ACCESS_DENIED", "no logic for you")
}

func TestJoin_SucceedsEvenWhenInitialLogicSyncFails(t *testing.T) {
	lg := newLogicHandlersFake()
	lg.getErr = errors.New("logic db down")
	env := handlerTestNew(t, lg, nil)
	c := env.dial("u1")
	msgs := c.join(handlerTestPage)
	if len(handlerOfType(msgs, "sync:page")) != 1 || len(handlerOfType(msgs, "sync:project")) != 1 {
		t.Fatalf("join must still complete: %+v", msgs)
	}
	if len(handlerOfType(msgs, "logic:sync")) != 0 || len(handlerOfType(msgs, "error")) != 0 {
		t.Fatalf("failed logic sync must be swallowed: %+v", msgs)
	}
}

func TestLogicUpdate_ProjectScopeBroadcastsToEveryPageOfTheProject(t *testing.T) {
	lg := newLogicHandlersFake()
	env := handlerTestNew(t, lg, nil)
	a, b := env.dial("alice"), env.dial("bob")
	a.join(handlerTestPage)
	b.join(handlerTestPage2)

	a.send("logic:update", map[string]any{
		"scope": "project", "patch": map[string]any{"variables": map[string]any{"x": 1}},
		"nullPaths": [][]string{{"variables", "y"}},
	}, "op-l", "req-l")
	for name, c := range map[string]*handlerConn{"sender": a, "other page": b} {
		m := c.next("logic:update")
		var p domainEntity.PayloadLogicChanged
		m.decode(t, &p)
		if m.ClientOpID != "op-l" || m.RequestID != "req-l" || p.Scope != "project" || p.PageID != "" ||
			p.Revision != 4 || p.Logic["scope"] != "project" {
			t.Fatalf("%s: %+v %+v", name, m, p)
		}
	}
	got := lg.updates[0]
	if got.project != handlerTestProject || got.page != "" || got.user != "alice" {
		t.Fatalf("call %+v", got)
	}
	if !reflect.DeepEqual(got.nullPaths, [][][]string{{{"variables", "y"}}}) {
		t.Fatalf("nullPaths = %v", got.nullPaths)
	}
	if vars, _ := got.patch["variables"].(map[string]any); vars["x"] != float64(1) {
		t.Fatalf("patch = %v", got.patch)
	}
}

func TestLogicUpdate_PageScopeStaysOnThePage(t *testing.T) {
	lg := newLogicHandlersFake()
	env := handlerTestNew(t, lg, nil)
	a, same, other := env.dial("alice"), env.dial("bob"), env.dial("carol")
	a.join(handlerTestPage)
	same.join(handlerTestPage)
	other.join(handlerTestPage2)

	a.send("logic:update", map[string]any{"scope": "page", "patch": map[string]any{"a": 1}}, "op", "")
	for _, c := range []*handlerConn{a, same} {
		m := c.next("logic:update")
		var p domainEntity.PayloadLogicChanged
		m.decode(t, &p)
		if p.Scope != "page" || p.PageID != handlerTestPage || p.Revision != 7 || m.PageID != handlerTestPage {
			t.Fatalf("%+v %+v", m, p)
		}
	}
	if d := other.drain(); len(handlerOfType(d, "logic:update")) != 0 {
		t.Fatalf("page-scoped update leaked to another page: %+v", d)
	}
	if got := lg.updates[0]; got.page != handlerTestPage {
		t.Fatalf("call %+v", got)
	}
}

func TestLogicUpdate_ValidationAndErrors(t *testing.T) {
	lg := newLogicHandlersFake()
	env := handlerTestNew(t, lg, nil)
	c := env.dial("u1")

	// Not joined: page scope needs a page.
	c.sendPage("logic:update", "", map[string]any{"scope": "page", "patch": map[string]any{}}, "", "")
	c.expectError("PROCESS_ERROR", "not joined")

	c.join(handlerTestPage)
	c.send("logic:update", "garbage", "", "")
	c.expectError("VALIDATION_ERROR", "")
	c.send("logic:update", map[string]any{"scope": "project"}, "", "")
	c.expectError("VALIDATION_ERROR", "patch is required")
	c.send("logic:update", map[string]any{"scope": "galaxy", "patch": map[string]any{}}, "", "")
	c.expectError("VALIDATION_ERROR", "scope must be")
	c.sendPage("logic:update", handlerTestPage2, map[string]any{"scope": "page", "patch": map[string]any{}}, "", "")
	c.expectError("VALIDATION_ERROR", "page does not match")
	if len(lg.updates) != 0 {
		t.Fatalf("rejected updates reached the use case: %d", len(lg.updates))
	}

	lg.err = domainErrors.New(domainErrors.CodeValidation, "name already used", nil)
	c.send("logic:update", map[string]any{"scope": "project", "patch": map[string]any{"v": 1}}, "op-r", "")
	m, _ := c.expectError("VALIDATION_ERROR", "name already used")
	if m.ClientOpID != "op-r" {
		t.Fatalf("clientOpId = %q", m.ClientOpID)
	}
}

func TestLogicUpdate_RejectedPatchIsNotBroadcast(t *testing.T) {
	lg := newLogicHandlersFake()
	lg.err = domainErrors.New(domainErrors.CodeAccessDenied, "no write access", nil)
	env := handlerTestNew(t, lg, nil)
	a, b := env.dial("alice"), env.dial("bob")
	a.join(handlerTestPage)
	b.join(handlerTestPage)
	a.send("logic:update", map[string]any{"scope": "project", "patch": map[string]any{"v": 1}}, "", "")
	a.expectError("ACCESS_DENIED", "no write access")
	if d := b.drain(); len(handlerOfType(d, "logic:update")) != 0 {
		t.Fatalf("rejected patch was broadcast: %+v", d)
	}
}

func TestLogicMove_BroadcastsBothDocumentsToTheWholeProject(t *testing.T) {
	lg := newLogicHandlersFake()
	env := handlerTestNew(t, lg, nil)
	a, b := env.dial("alice"), env.dial("bob")
	a.join(handlerTestPage)
	b.join(handlerTestPage2)

	a.send("logic:move", map[string]any{"kind": "variable", "name": "count", "to": "project"}, "op-mv", "req-mv")
	for name, c := range map[string]*handlerConn{"sender": a, "other page": b} {
		m := c.next("logic:move")
		var p domainEntity.PayloadLogicSync
		m.decode(t, &p)
		if m.ClientOpID != "op-mv" || m.RequestID != "req-mv" || p.PageID != handlerTestPage ||
			p.ProjectRevision != 4 || p.PageRevision != 7 || p.Project["scope"] != "project" || p.Page["scope"] != "page" {
			t.Fatalf("%s: %+v %+v", name, m, p)
		}
	}
	want := logicHandlersMoveCall{handlerTestProject, handlerTestPage, "alice", "variable", "count", "project"}
	if lg.moves[0] != want {
		t.Fatalf("call %+v, want %+v", lg.moves[0], want)
	}
}

func TestLogicMove_ValidationAndErrors(t *testing.T) {
	lg := newLogicHandlersFake()
	env := handlerTestNew(t, lg, nil)
	c := env.dial("u1")
	c.send("logic:move", map[string]any{}, "", "")
	c.expectError("PROCESS_ERROR", "not joined")

	c.join(handlerTestPage)
	c.sendPage("logic:move", handlerTestPage2, map[string]any{"kind": "variable"}, "", "")
	c.expectError("VALIDATION_ERROR", "page does not match")
	c.send("logic:move", "garbage", "", "")
	c.expectError("VALIDATION_ERROR", "")
	if len(lg.moves) != 0 {
		t.Fatalf("rejected moves reached the use case: %d", len(lg.moves))
	}

	lg.err = domainErrors.New(domainErrors.CodeNotFound, "no such variable", nil)
	c.send("logic:move", map[string]any{"kind": "variable", "name": "nope", "to": "page"}, "", "")
	c.expectError("NOT_FOUND", "no such variable")
}
