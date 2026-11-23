package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"

	wsentity "my-go-app/internal/realtime/delivery/ws/entity"
	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
	domainUsecase "my-go-app/internal/realtime/domain/usecase"
	"my-go-app/internal/realtime/utils/logger"
)

const (
	handlerTestProject = "proj-1"
	handlerTestPage    = "page-1"
	handlerTestPage2   = "page-2"
	handlerTestElemA   = "11111111-1111-1111-1111-111111111111"
	handlerTestElemB   = "22222222-2222-2222-2222-222222222222"
)

func handlerTestLogger() logger.Logger { return logger.New(io.Discard, "", 0) }

// handlerFakeElements is an in-memory ElementUseCase that records every call.
type handlerFakeElements struct {
	domainUsecase.ElementUseCase
	mu        sync.Mutex
	seq       int64
	opErr     error
	syncErr   error
	projErr   error
	pageElems []domainEntity.Element
	projElems []domainEntity.Element
	moveRes   *domainEntity.Element

	created []handlerCreateCall
	patched []handlerPatchCall
	removed []handlerRefCall
	moved   []handlerMoveCall
	synced  []handlerRefCall
}

type handlerCreateCall struct {
	project string
	el      domainEntity.Element
	pos     *int
}
type handlerPatchCall struct {
	project, page, id string
	props, styles     map[string]any
	settings          map[string]any
}
type handlerRefCall struct{ project, page, id string }
type handlerMoveCall struct {
	project, page, id, parent string
	index                     int
}

func (f *handlerFakeElements) next() (int64, error) {
	if f.opErr != nil {
		return 0, f.opErr
	}
	f.seq++
	return f.seq, nil
}

func (f *handlerFakeElements) CreateElementAt(_ context.Context, project string, el *domainEntity.Element, pos *int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.created = append(f.created, handlerCreateCall{project, *el, pos})
	return f.next()
}

func (f *handlerFakeElements) PatchElement(_ context.Context, project, page, id string, props, styles, settings map[string]any) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.patched = append(f.patched, handlerPatchCall{project, page, id, props, styles, settings})
	return f.next()
}

func (f *handlerFakeElements) RemoveElement(_ context.Context, project, page, id string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.removed = append(f.removed, handlerRefCall{project, page, id})
	return f.next()
}

func (f *handlerFakeElements) MoveElement(_ context.Context, project, page, id, parent string, index int) (*domainEntity.Element, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.moved = append(f.moved, handlerMoveCall{project, page, id, parent, index})
	seq, err := f.next()
	if err != nil {
		return nil, 0, err
	}
	if f.moveRes != nil {
		return f.moveRes, seq, nil
	}
	return &domainEntity.Element{ID: id, Order: index}, seq, nil
}

func (f *handlerFakeElements) SyncPageWithSeq(_ context.Context, project, page string) ([]domainEntity.Element, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.synced = append(f.synced, handlerRefCall{project, page, ""})
	if f.syncErr != nil {
		return nil, 0, f.syncErr
	}
	return f.pageElems, f.seq, nil
}

func (f *handlerFakeElements) GetElementsByProjectID(_ context.Context, _ string) ([]domainEntity.Element, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.projErr != nil {
		return nil, f.projErr
	}
	return f.projElems, nil
}

// handlerFakeAccess implements ProjectAccess.
type handlerFakeAccess struct {
	read, write bool
	err         error
}

func (a handlerFakeAccess) Access(context.Context, string, string) (bool, bool, error) {
	return a.read, a.write, a.err
}

type handlerTestEnv struct {
	t    *testing.T
	h    *Handler
	app  *fiber.App
	base string
	els  *handlerFakeElements
}

// handlerTestNew builds a Handler over fakes (no Redis) and serves it on loopback.
// Connections authenticate as the user named by the ?u= query parameter (default u1).
func handlerTestNew(t *testing.T, lg domainUsecase.LogicUseCase, access ProjectAccess) *handlerTestEnv {
	t.Helper()
	els := &handlerFakeElements{}
	uc := &domainUsecase.UseCase{Element: els, Logic: lg}
	h := NewHandler(uc, nil, handlerTestLogger())
	if access != nil {
		h.RequireAccess(access)
	}
	t.Cleanup(h.Shutdown)
	return handlerTestServe(t, h, els)
}

func handlerTestServe(t *testing.T, h *Handler, els *handlerFakeElements) *handlerTestEnv {
	t.Helper()
	app := fiber.New()
	app.Use("/ws", func(c fiber.Ctx) error {
		u := c.Query("u")
		if u == "" {
			u = "u1"
		}
		c.Locals("user", &domainEntity.User{ID: u, Email: u + "@example.com"})
		return c.Next()
	})
	h.SetupRoutes(app)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("cannot listen on loopback: %v", err)
	}
	go func() { _ = app.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true}) }()
	t.Cleanup(func() { _ = app.Shutdown() })
	return &handlerTestEnv{t: t, h: h, app: app, base: "ws://" + ln.Addr().String() + "/ws/", els: els}
}

type handlerMsg struct {
	Type       string          `json:"type"`
	PageID     string          `json:"pageId"`
	UserID     string          `json:"userId"`
	Seq        int64           `json:"seq"`
	ClientOpID string          `json:"clientOpId"`
	RequestID  string          `json:"requestId"`
	Payload    json.RawMessage `json:"payload"`
}

func (m handlerMsg) decode(t *testing.T, v any) {
	t.Helper()
	if err := json.Unmarshal(m.Payload, v); err != nil {
		t.Fatalf("decode payload %s: %v", m.Payload, err)
	}
}

type handlerConn struct {
	t    *testing.T
	conn *websocket.Conn
	page string
}

func (e *handlerTestEnv) dialURL(project, query string) (*websocket.Conn, *http.Response, error) {
	u := e.base + project
	if query != "" {
		u += "?" + query
	}
	return websocket.DefaultDialer.Dial(u, nil)
}

func (e *handlerTestEnv) dial(user string) *handlerConn { return e.dialQuery("u=" + user) }

func (e *handlerTestEnv) dialQuery(query string) *handlerConn {
	e.t.Helper()
	conn, _, err := e.dialURL(handlerTestProject, query)
	if err != nil {
		e.t.Fatalf("dial: %v", err)
	}
	e.t.Cleanup(func() { _ = conn.Close() })
	return &handlerConn{t: e.t, conn: conn}
}

func (c *handlerConn) send(typ string, payload any, opID, reqID string) {
	c.t.Helper()
	c.sendPage(typ, c.page, payload, opID, reqID)
}

func (c *handlerConn) sendPage(typ, page string, payload any, opID, reqID string) {
	c.t.Helper()
	m := map[string]any{"type": typ, "pageId": page, "payload": payload}
	if opID != "" {
		m["clientOpId"] = opID
	}
	if reqID != "" {
		m["requestId"] = reqID
	}
	if err := c.conn.WriteJSON(m); err != nil {
		c.t.Fatalf("write: %v", err)
	}
}

// read returns the next message. A read timeout poisons the connection, so a
// miss is fatal; use drain to assert that something was NOT delivered.
func (c *handlerConn) read() handlerMsg {
	c.t.Helper()
	_ = c.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var m handlerMsg
	if err := c.conn.ReadJSON(&m); err != nil {
		c.t.Fatalf("read: %v", err)
	}
	return m
}

// next returns the next message of the given type, skipping others.
func (c *handlerConn) next(typ string) handlerMsg {
	c.t.Helper()
	for {
		if m := c.read(); m.Type == typ {
			return m
		}
	}
}

// drain returns every message delivered before a sync:project round trip completes.
func (c *handlerConn) drain() []handlerMsg {
	c.t.Helper()
	time.Sleep(60 * time.Millisecond)
	c.send("sync:project", struct{}{}, "", "barrier")
	var out []handlerMsg
	for {
		m := c.read()
		if m.Type == "sync:project" && m.RequestID == "barrier" {
			return out
		}
		out = append(out, m)
	}
}

func handlerOfType(msgs []handlerMsg, typ string) []handlerMsg {
	var out []handlerMsg
	for _, m := range msgs {
		if m.Type == typ {
			out = append(out, m)
		}
	}
	return out
}

// join joins a page and returns everything the server sent in reply.
func (c *handlerConn) join(page string) []handlerMsg {
	c.t.Helper()
	c.page = page
	c.sendPage("join", page, map[string]any{}, "", "join-"+page)
	var out []handlerMsg
	for {
		m := c.read()
		out = append(out, m)
		if m.Type == "sync:project" && m.RequestID == "join-"+page {
			return out
		}
	}
}

type handlerErrPayload struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	ClientOpID string `json:"clientOpId"`
}

func (c *handlerConn) expectError(code, msgContains string) (handlerMsg, handlerErrPayload) {
	c.t.Helper()
	m := c.next("error")
	var p handlerErrPayload
	m.decode(c.t, &p)
	if p.Code != code || !strings.Contains(p.Message, msgContains) {
		c.t.Fatalf("error = %+v, want code %s containing %q", p, code, msgContains)
	}
	return m, p
}

// ---- handler.go ----

func TestHandler_UnsupportedMessageTypeIsRejectedWithValidationError(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.join(handlerTestPage)
	c.send("no:such:type", map[string]any{}, "op-x", "req-x")
	m, _ := c.expectError("VALIDATION_ERROR", "unsupported message type")
	if m.ClientOpID != "op-x" || m.RequestID != "req-x" {
		t.Fatalf("error must echo ids, got op=%q req=%q", m.ClientOpID, m.RequestID)
	}
}

func TestHandler_AgentResponseIsUnsupportedWithoutRedis(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	if env.h.agents != nil {
		t.Fatal("agent bridge must not exist without Redis")
	}
	c := env.dial("u1")
	c.join(handlerTestPage)
	c.send("agent:response", map[string]any{}, "", "")
	c.expectError("VALIDATION_ERROR", "unsupported message type")
}

func TestHandler_ErrorReplyUsesStructuredMessageOnly(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	env.els.projErr = domainErrors.New(domainErrors.CodeNotFound, "project gone", errors.New("pq: relation secret_table"))
	c := env.dial("u1")
	c.send("sync:project", map[string]any{}, "", "r1")
	_, p := c.expectError("NOT_FOUND", "project gone")
	if strings.Contains(p.Message, "secret_table") || strings.Contains(p.Message, "NOT_FOUND") {
		t.Fatalf("client must not see the cause or the code in the message: %q", p.Message)
	}
}

func TestHandler_PlainErrorsAreReportedAsProcessError(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	env.els.projErr = errors.New("boom")
	c := env.dial("u1")
	c.send("sync:project", map[string]any{}, "", "")
	c.expectError("PROCESS_ERROR", "boom")
}

func TestHandler_ConnectionSurvivesAFailedMessage(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.send("bogus", nil, "", "")
	c.expectError("VALIDATION_ERROR", "unsupported")
	c.send("sync:project", map[string]any{}, "", "after")
	if m := c.next("sync:project"); m.RequestID != "after" {
		t.Fatalf("unexpected reply %+v", m)
	}
}

func TestHandler_MalformedFrameClosesTheConnection(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	if err := c.conn.WriteMessage(websocket.TextMessage, []byte("{not json")); err != nil {
		t.Fatal(err)
	}
	_ = c.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func TestHandler_AccessCheckFailureReportsInternalErrorAndCloses(t *testing.T) {
	env := handlerTestNew(t, nil, handlerFakeAccess{err: errors.New("db down")})
	c := env.dial("u1")
	_, p := c.expectError("INTERNAL_ERROR", "could not check access")
	if strings.Contains(p.Message, "db down") {
		t.Fatalf("access error detail leaked: %q", p.Message)
	}
}

func TestHandler_UserWithoutReadAccessIsDenied(t *testing.T) {
	env := handlerTestNew(t, nil, handlerFakeAccess{})
	c := env.dial("u1")
	c.expectError("ACCESS_DENIED", "do not have access")
	_ = c.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	if _, _, err := c.conn.ReadMessage(); err == nil {
		t.Fatal("connection must be closed after denial")
	}
}

func TestHandler_ReadOnlyCollaboratorCanReadButNotMutate(t *testing.T) {
	env := handlerTestNew(t, nil, handlerFakeAccess{read: true})
	c := env.dial("viewer")
	c.join(handlerTestPage)

	for _, typ := range []string{"element:create", "element:update", "element:delete", "element:move", "logic:update", "logic:move"} {
		c.send(typ, map[string]any{}, "op-"+typ, "")
		m, _ := c.expectError("ACCESS_DENIED", "read-only")
		if m.ClientOpID != "op-"+typ {
			t.Fatalf("%s: clientOpId not echoed: %q", typ, m.ClientOpID)
		}
	}
	env.els.mu.Lock()
	n := len(env.els.created) + len(env.els.patched) + len(env.els.removed) + len(env.els.moved)
	env.els.mu.Unlock()
	if n != 0 {
		t.Fatalf("read-only client reached the use case %d times", n)
	}

	// Reads and presence are still allowed.
	c.send("sync:page", map[string]any{}, "", "s1")
	if m := c.next("sync:page"); m.RequestID != "s1" {
		t.Fatalf("sync reply %+v", m)
	}
	c.send("presence", map[string]any{"userId": "viewer", "cursorX": 1, "cursorY": 2}, "", "")
	if d := c.drain(); len(handlerOfType(d, "error")) != 0 {
		t.Fatalf("presence must not be rejected for read-only users: %+v", d)
	}
}

func TestHandler_WritableCollaboratorPassesAccessCheck(t *testing.T) {
	env := handlerTestNew(t, nil, handlerFakeAccess{read: true, write: true})
	c := env.dial("editor")
	c.join(handlerTestPage)
	c.send("element:delete", map[string]any{"id": handlerTestElemA}, "d1", "")
	if m := c.next("element:delete"); m.Seq != 1 || m.ClientOpID != "d1" {
		t.Fatalf("delete ack %+v", m)
	}
}

func TestHandler_NewConnectionReplacesStaleOneOfSameUser(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	old := env.dial("u1")
	old.join(handlerTestPage)
	fresh := env.dial("u1")
	fresh.join(handlerTestPage)

	_ = old.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		if _, _, err := old.conn.ReadMessage(); err != nil {
			break
		}
	}
	// The replacement keeps working and is the only registered tab.
	fresh.send("element:delete", map[string]any{"id": handlerTestElemA}, "d1", "")
	fresh.next("element:delete")
	env.h.activeConnsMu.Lock()
	n := len(env.h.activeConns)
	env.h.activeConnsMu.Unlock()
	if n != 1 {
		t.Fatalf("activeConns = %d, want 1", n)
	}
}

func TestHandler_WriterSocketCoexistsWithEditorConnection(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	editor := env.dial("u1")
	editor.join(handlerTestPage)
	writer := env.dialQuery("u=u1&conn=writer")
	writer.join(handlerTestPage)

	// The editor was not replaced: it still sees the writer's op.
	writer.send("element:delete", map[string]any{"id": handlerTestElemA}, "w1", "")
	if m := editor.next("element:delete"); m.ClientOpID != "w1" {
		t.Fatalf("editor missed writer op: %+v", m)
	}
	env.h.activeConnsMu.Lock()
	n := len(env.h.activeConns)
	env.h.activeConnsMu.Unlock()
	if n != 1 {
		t.Fatalf("writer socket must not be tracked, activeConns = %d", n)
	}
}

func TestHandler_ActiveConnIsForgottenOnDisconnect(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.join(handlerTestPage)
	_ = c.conn.Close()
	deadline := time.Now().Add(3 * time.Second)
	for {
		env.h.activeConnsMu.Lock()
		n := len(env.h.activeConns)
		env.h.activeConnsMu.Unlock()
		if n == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("activeConns not cleaned after disconnect")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestHandler_ConnKey(t *testing.T) {
	if got := connKey("p", "u"); got != "p::u" {
		t.Fatalf("connKey = %q", got)
	}
	if connKey("a", "b::c") == connKey("a::b", "c") {
		t.Log("note: connKey is ambiguous when ids contain the separator")
	}
}

func TestHandler_ShutdownStopsActorsAndRejectsNewOps(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.join(handlerTestPage)
	env.h.Shutdown()
	client := wsentity.NewClient(nil, handlerTestProject, &domainEntity.User{ID: "u1"})
	client.SetPageID(handlerTestPage)
	payload, _ := json.Marshal(map[string]any{"id": handlerTestElemA})
	err := env.h.handleElementDelete(context.Background(), client, domainEntity.Message{Payload: payload})
	if err == nil {
		t.Fatal("ops after shutdown must fail")
	}
}

func TestHandler_ApplyAndBroadcastReportsBuildFailureAsInternal(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	client := wsentity.NewClient(nil, handlerTestProject, &domainEntity.User{ID: "u1"})
	client.SetPageID(handlerTestPage)
	applied := false
	err := env.h.applyAndBroadcast(context.Background(), client, domainEntity.Message{Type: domainEntity.TypeElementDelete}, handlerTestPage,
		func(context.Context) (int64, error) { applied = true; return 5, nil },
		func(*wsentity.MessageFactory) (domainEntity.Message, error) {
			return domainEntity.Message{}, errors.New("cannot encode")
		})
	if !applied {
		t.Fatal("apply must have run")
	}
	if !domainErrors.IsCode(err, domainErrors.CodeInternal) {
		t.Fatalf("err = %v, want INTERNAL_ERROR", err)
	}
}

func TestHandler_ApplyAndBroadcastPropagatesApplyError(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	client := wsentity.NewClient(nil, handlerTestProject, &domainEntity.User{ID: "u1"})
	want := errors.New("apply failed")
	built := false
	err := env.h.applyAndBroadcast(context.Background(), client, domainEntity.Message{}, handlerTestPage,
		func(context.Context) (int64, error) { return 0, want },
		func(*wsentity.MessageFactory) (domainEntity.Message, error) {
			built = true
			return domainEntity.Message{}, nil
		})
	if !errors.Is(err, want) || built {
		t.Fatalf("err=%v built=%v", err, built)
	}
}

// ---- routes_handlers.go ----

func TestRoutes_NonUpgradeRequestToWSIsRefused(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	resp, err := env.app.Test(httptest.NewRequest(http.MethodGet, "/ws/"+handlerTestProject, nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusUpgradeRequired {
		t.Fatalf("status = %d, want 426", resp.StatusCode)
	}
}

func TestRoutes_UnknownPathIsNotFound(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	resp, err := env.app.Test(httptest.NewRequest(http.MethodGet, "/other", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestRoutes_WebSocketUpgradeIsServed(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	c := env.dial("u1")
	c.send("sync:project", map[string]any{}, "", "r")
	if m := c.next("sync:project"); m.RequestID != "r" {
		t.Fatalf("reply %+v", m)
	}
}
