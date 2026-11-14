package unit

import (
	"context"
	"encoding/json"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"

	"my-go-app/internal/realtime/delivery/ws/handlers"
	"my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
	"my-go-app/internal/realtime/domain/logic"
	domainRepo "my-go-app/internal/realtime/domain/repository"
	domainUsecase "my-go-app/internal/realtime/domain/usecase"
	"my-go-app/internal/realtime/usecase"
)

// memLogic is an in-memory LogicUseCase that uses the real validator, so these
// tests exercise the websocket handler's behaviour, not a reimplementation of it.
type memLogic struct {
	mu      sync.Mutex
	read    map[string]bool
	write   map[string]bool
	project logic.Document
	pages   map[string]logic.Document
	// Revisions mirror the repository: +1 per committed update of a document.
	projectRev int64
	pageRevs   map[string]int64
}

func newMemLogic() *memLogic {
	return &memLogic{
		read:     map[string]bool{"editor": true, "editor2": true, "viewer": true},
		write:    map[string]bool{"editor": true, "editor2": true},
		project:  logic.Empty(),
		pages:    map[string]logic.Document{"page": logic.Empty(), "page2": logic.Empty()},
		pageRevs: map[string]int64{},
	}
}

func (m *memLogic) Access(_ context.Context, _, user string) (bool, bool, error) {
	return m.read[user], m.write[user], nil
}

func (m *memLogic) Update(_ context.Context, _, pageID, user string, patch map[string]any) (logic.Revisioned, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.write[user] {
		return logic.Revisioned{}, domainErrors.New(domainErrors.CodeAccessDenied, "no write access", domainErrors.ErrUnauthorized)
	}
	current := m.project
	if pageID != "" {
		var ok bool
		if current, ok = m.pages[pageID]; !ok {
			return logic.Revisioned{}, domainErrors.New(domainErrors.CodeNotFound, "page not found", domainErrors.ErrNotFound)
		}
	}
	updated, err := logic.Merge(current, patch)
	if err != nil {
		return logic.Revisioned{}, domainErrors.New(domainErrors.CodeValidation, err.Error(), domainErrors.ErrValidation)
	}
	project, pages := m.project, map[string]logic.Document{}
	for id, doc := range m.pages {
		pages[id] = doc
	}
	if pageID == "" {
		project = updated
	} else {
		pages[pageID] = updated
	}
	if err := logic.ValidateScopes(project, pages); err != nil {
		return logic.Revisioned{}, domainErrors.New(domainErrors.CodeValidation, err.Error(), domainErrors.ErrValidation)
	}
	m.project, m.pages = project, pages
	if pageID == "" {
		m.projectRev++
		return logic.Revisioned{Doc: updated, Revision: m.projectRev}, nil
	}
	m.pageRevs[pageID]++
	return logic.Revisioned{Doc: updated, Revision: m.pageRevs[pageID]}, nil
}

func (m *memLogic) Get(_ context.Context, _, pageID, user string) (logic.Revisioned, logic.Revisioned, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.read[user] {
		return logic.Revisioned{}, logic.Revisioned{}, domainErrors.New(domainErrors.CodeAccessDenied, "no access", domainErrors.ErrUnauthorized)
	}
	return logic.Revisioned{Doc: m.project, Revision: m.projectRev}, logic.Revisioned{Doc: m.pages[pageID], Revision: m.pageRevs[pageID]}, nil
}

func (m *memLogic) snapshot() (logic.Document, logic.Document) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.project, m.pages["page"]
}

func startLogicServer(t *testing.T, mem domainUsecase.LogicUseCase) string {
	t.Helper()
	repo := &domainRepo.Repository{DocumentRepo: newMemDocRepo(), ElementRepo: &memElementRepo{elements: []entity.Element{
		{ID: testElemID, Type: "text", PageID: "page", Styles: map[string]any{"color": "black"}},
	}}}
	docUC := usecase.NewDocumentUseCase(repo)
	uc := &domainUsecase.UseCase{
		Element:          stubElementUC{usecase.NewElementUseCase(repo, docUC, nil)},
		Document:         docUC,
		Workflow:         stubWorkflowUC{},
		ElementWorkflows: stubElementWorkflowsUC{},
		Logic:            mem,
	}
	h := handlers.NewHandler(uc, nil, &mockLogger{})
	h.RequireAccess(mem)
	t.Cleanup(h.Shutdown)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use("/ws", func(c *fiber.Ctx) error {
		c.Locals("user", &entity.User{ID: c.Query("u"), Email: c.Query("u") + "@example.com"})
		return c.Next()
	})
	h.SetupRoutes(app)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("cannot listen on loopback: %v", err)
	}
	go func() { _ = app.Listener(ln) }()
	t.Cleanup(func() { _ = app.Shutdown() })
	return "ws://" + ln.Addr().String() + "/ws/" + testProject
}

type logicClient struct{ *wsTestClient }

func dialLogic(t *testing.T, url, user, page string) (*logicClient, bool) {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url+"?u="+user, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	c := &logicClient{&wsTestClient{t: t, conn: conn}}
	c.raw("join", page, nil, "")
	_, joined := c.next("sync:page", 3*time.Second)
	return c, joined
}

func (c *logicClient) raw(typ, page string, payload any, opID string) {
	c.t.Helper()
	msg := map[string]any{"type": typ, "pageId": page, "timestamp": time.Now().UnixMilli(), "payload": payload}
	if opID != "" {
		msg["clientOpId"] = opID
	}
	if err := c.conn.WriteJSON(msg); err != nil {
		c.t.Fatalf("write: %v", err)
	}
}

type logicChanged struct {
	Scope    string         `json:"scope"`
	PageID   string         `json:"pageId"`
	Logic    map[string]any `json:"logic"`
	Revision *int64         `json:"revision"`
}

type logicSync struct {
	Project         map[string]any `json:"project"`
	PageID          string         `json:"pageId"`
	ProjectRevision *int64         `json:"projectRevision"`
	PageRevision    *int64         `json:"pageRevision"`
}

// Every logic:update (ack and broadcast) and logic:sync carries the document's
// revision, which increases with each commit, so a client can drop an older
// sync that arrives after a newer broadcast.
func TestLogicWS_RevisionsIncreaseAndTravelWithEveryDocument(t *testing.T) {
	mem := newMemLogic()
	url := startLogicServer(t, mem)
	editor, _ := dialLogic(t, url, "editor", "page")
	joinSync, ok := editor.next("logic:sync", 3*time.Second)
	if !ok {
		t.Fatal("no logic:sync after join")
	}
	if got := decode[logicSync](t, joinSync); got.ProjectRevision == nil || *got.ProjectRevision != 0 || got.PageRevision == nil || *got.PageRevision != 0 {
		t.Fatalf("fresh documents must have revision 0: %s", joinSync.Payload)
	}
	var last int64
	for i, name := range []string{"a", "b", "c"} {
		editor.raw("logic:update", "page", map[string]any{"scope": "project", "patch": map[string]any{"variables": map[string]any{name: variable("number")}}}, name)
		msg, ok := editor.next("logic:update", 3*time.Second)
		if !ok || msg.ClientOpID != name {
			t.Fatalf("update %s not acknowledged: %+v", name, msg)
		}
		got := decode[logicChanged](t, msg)
		if got.Revision == nil || *got.Revision <= last || *got.Revision != int64(i+1) {
			t.Fatalf("revision did not increase: %s (previous %d)", msg.Payload, last)
		}
		last = *got.Revision
	}
	editor.raw("logic:update", "page", map[string]any{"scope": "page", "patch": map[string]any{"flows": map[string]any{"f": map[string]any{"do": []any{map[string]any{"show": ""}}}}}}, "page-draft")
	msg, ok := editor.next("logic:update", 3*time.Second)
	if !ok {
		t.Fatal("incomplete page flow was not accepted")
	}
	if got := decode[logicChanged](t, msg); got.Revision == nil || *got.Revision != 1 {
		t.Fatalf("page revision is counted separately: %s", msg.Payload)
	}
	editor.raw("logic:sync", "page", map[string]any{}, "")
	msg, ok = editor.next("logic:sync", 3*time.Second)
	if !ok {
		t.Fatal("no logic:sync reply")
	}
	if got := decode[logicSync](t, msg); got.ProjectRevision == nil || *got.ProjectRevision != 3 || got.PageRevision == nil || *got.PageRevision != 1 {
		t.Fatalf("sync must carry the latest revisions: %s", msg.Payload)
	}
}

func decode[T any](t *testing.T, m wsIn) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(m.Payload, &v); err != nil {
		t.Fatalf("decode %s: %v", m.Payload, err)
	}
	return v
}

func variable(kind string) map[string]any {
	return map[string]any{"type": kind, "default": map[string]any{"string": "", "number": float64(0)}[kind]}
}

func TestLogicWS_JoinSendsCurrentLogic(t *testing.T) {
	mem := newMemLogic()
	seeded, err := logic.Merge(mem.project, map[string]any{"variables": map[string]any{"count": variable("number")}})
	if err != nil {
		t.Fatal(err)
	}
	mem.project = seeded
	url := startLogicServer(t, mem)
	c, joined := dialLogic(t, url, "viewer", "page")
	if !joined {
		t.Fatal("viewer could not join")
	}
	m, ok := c.next("logic:sync", 3*time.Second)
	if !ok {
		t.Fatal("no logic:sync after join")
	}
	got := decode[struct {
		Project map[string]any `json:"project"`
		PageID  string         `json:"pageId"`
	}](t, m)
	if got.PageID != "page" || got.Project["variables"].(map[string]any)["count"] == nil {
		t.Fatalf("unexpected logic:sync: %s", m.Payload)
	}
}

func TestLogicWS_ProjectUpdateReachesEveryPageAndPageUpdateOnlyItsPage(t *testing.T) {
	mem := newMemLogic()
	url := startLogicServer(t, mem)
	a, _ := dialLogic(t, url, "editor", "page")
	b, _ := dialLogic(t, url, "editor2", "page2")

	a.raw("logic:update", "page", map[string]any{"scope": "project", "patch": map[string]any{"variables": map[string]any{"count": variable("number")}}}, "op-1")
	for name, c := range map[string]*logicClient{"sender": a, "other page": b} {
		m, ok := c.next("logic:update", 3*time.Second)
		if !ok {
			t.Fatalf("%s did not receive the project update", name)
		}
		got := decode[logicChanged](t, m)
		if got.Scope != "project" || got.Logic["variables"].(map[string]any)["count"] == nil {
			t.Fatalf("%s got %s", name, m.Payload)
		}
		if name == "sender" && m.ClientOpID != "op-1" {
			t.Fatalf("sender ack lost clientOpId: %+v", m)
		}
	}

	a.raw("logic:update", "page", map[string]any{"scope": "page", "patch": map[string]any{"flows": map[string]any{"greet": map[string]any{"do": []any{}}}}}, "op-2")
	m, ok := a.next("logic:update", 3*time.Second)
	if !ok || decode[logicChanged](t, m).PageID != "page" {
		t.Fatalf("page editor did not get its page update: ok=%v %s", ok, m.Payload)
	}
	if other, leaked := b.next("logic:update", 400*time.Millisecond); leaked {
		t.Fatalf("a page-scoped update leaked to another page: %s", other.Payload)
	}
}

func TestLogicWS_RejectedPatchesChangeNothingAndOnlyTheSenderHearsOfIt(t *testing.T) {
	mem := newMemLogic()
	url := startLogicServer(t, mem)
	a, _ := dialLogic(t, url, "editor", "page")
	b, _ := dialLogic(t, url, "editor2", "page")
	a.raw("logic:update", "page", map[string]any{"scope": "project", "patch": map[string]any{"variables": map[string]any{"count": variable("number")}}}, "ok")
	a.next("logic:update", 3*time.Second)
	b.next("logic:update", 3*time.Second)
	before, _ := mem.snapshot()

	cases := []struct {
		name    string
		payload map[string]any
		code    string
	}{
		{"name already used by the project", map[string]any{"scope": "page", "patch": map[string]any{"variables": map[string]any{"count": variable("string")}}}, "VALIDATION_ERROR"},
		{"default does not match the type", map[string]any{"scope": "project", "patch": map[string]any{"variables": map[string]any{"x": map[string]any{"type": "number", "default": "no"}}}}, "VALIDATION_ERROR"},
		{"unknown scope", map[string]any{"scope": "team", "patch": map[string]any{}}, "VALIDATION_ERROR"},
		{"missing patch", map[string]any{"scope": "project"}, "VALIDATION_ERROR"},
	}
	for _, tc := range cases {
		a.raw("logic:update", "page", tc.payload, "bad")
		em, ok := a.next("error", 3*time.Second)
		if !ok || em.ClientOpID != "bad" {
			t.Fatalf("%s: expected an error echoing the op, got ok=%v %+v", tc.name, ok, em)
		}
		reply := decode[struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}](t, em)
		if reply.Code != tc.code {
			t.Fatalf("%s: code %q, want %q", tc.name, reply.Code, tc.code)
		}
		// The message is the reason only: no repeated code prefix, no repeated cause.
		if strings.Contains(reply.Message, "[") || strings.Count(reply.Message, "validation failed") > 1 {
			t.Fatalf("%s: message is not clean: %q", tc.name, reply.Message)
		}
	}
	if other, leaked := b.next("logic:update", 400*time.Millisecond); leaked {
		t.Fatalf("a rejected patch was broadcast: %s", other.Payload)
	}
	after, _ := mem.snapshot()
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)
	if string(beforeJSON) != string(afterJSON) {
		t.Fatalf("rejected patches changed the stored logic: %s -> %s", beforeJSON, afterJSON)
	}
}

func TestLogicWS_ViewersReadButCannotChangeAnything(t *testing.T) {
	mem := newMemLogic()
	url := startLogicServer(t, mem)
	editor, _ := dialLogic(t, url, "editor", "page")
	viewer, joined := dialLogic(t, url, "viewer", "page")
	if !joined {
		t.Fatal("a viewer must be able to join")
	}

	for _, tc := range []struct {
		typ     string
		payload any
	}{
		{"logic:update", map[string]any{"scope": "project", "patch": map[string]any{"variables": map[string]any{"v": variable("string")}}}},
		{"element:update", map[string]any{"id": testElemID, "styles": map[string]any{"color": "red"}}},
	} {
		viewer.raw(tc.typ, "page", tc.payload, "v-op")
		em, ok := viewer.next("error", 3*time.Second)
		if !ok {
			t.Fatalf("%s from a viewer was not rejected", tc.typ)
		}
		if code := decode[struct {
			Code string `json:"code"`
		}](t, em).Code; code != "ACCESS_DENIED" {
			t.Fatalf("%s: code %q, want ACCESS_DENIED", tc.typ, code)
		}
	}
	if project, _ := mem.snapshot(); len(project["variables"].(map[string]any)) != 0 {
		t.Fatal("a viewer changed the project logic")
	}

	editor.raw("logic:update", "page", map[string]any{"scope": "project", "patch": map[string]any{"variables": map[string]any{"v": variable("string")}}}, "e-op")
	if _, ok := viewer.next("logic:update", 3*time.Second); !ok {
		t.Fatal("a viewer should still receive the editor's changes")
	}
}

func TestLogicWS_UsersWithoutAccessAreTurnedAway(t *testing.T) {
	url := startLogicServer(t, newMemLogic())
	conn, _, err := websocket.DefaultDialer.Dial(url+"?u=stranger", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var first wsIn
	if err := conn.ReadJSON(&first); err != nil || first.Type != "error" {
		t.Fatalf("expected an error first, got %v %+v", err, first)
	}
	if code := decode[struct {
		Code string `json:"code"`
	}](t, first).Code; code != "ACCESS_DENIED" {
		t.Fatalf("code %q, want ACCESS_DENIED", code)
	}
	_ = conn.WriteJSON(map[string]any{"type": "join", "pageId": "page", "timestamp": 1})
	var next wsIn
	if err := conn.ReadJSON(&next); err == nil && next.Type == "sync:page" {
		t.Fatal("a user without access received the page")
	}
}
