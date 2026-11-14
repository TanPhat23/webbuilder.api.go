package unit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"

	wsentity "my-go-app/internal/realtime/delivery/ws/entity"
	"my-go-app/internal/realtime/delivery/ws/handlers"
	"my-go-app/internal/realtime/domain/entity"
	domainRepo "my-go-app/internal/realtime/domain/repository"
	domainUsecase "my-go-app/internal/realtime/domain/usecase"
	"my-go-app/internal/realtime/usecase"
)

// Stubs for the use cases join/sync touch besides the element use case.
type stubElementUC struct{ domainUsecase.ElementUseCase }

func (stubElementUC) GetElementsByProjectID(_ context.Context, _ string) ([]entity.Element, error) {
	return nil, nil
}

type stubWorkflowUC struct{ domainUsecase.WorkflowUseCase }

func (stubWorkflowUC) GetWorkflowsByProject(_ context.Context, _ string) ([]entity.EventWorkflow, error) {
	return nil, nil
}

type stubElementWorkflowsUC struct {
	domainUsecase.ElementWorkflowsUseCase
}

func (stubElementWorkflowsUC) GetWorkflowsByPage(_ context.Context, _, _ string) ([]entity.ElementEventWorkflow, error) {
	return nil, nil
}
func (stubElementWorkflowsUC) GetWorkflowsByProject(_ context.Context, _ string) ([]entity.ElementEventWorkflow, error) {
	return nil, nil
}

type wsTestClient struct {
	t    *testing.T
	conn *websocket.Conn
}

func (c *wsTestClient) send(typ string, payload any, clientOpID string) {
	c.t.Helper()
	msg := map[string]any{"type": typ, "pageId": testPage, "timestamp": time.Now().UnixMilli(), "payload": payload}
	if clientOpID != "" {
		msg["clientOpId"] = clientOpID
	}
	if err := c.conn.WriteJSON(msg); err != nil {
		c.t.Fatalf("write: %v", err)
	}
}

type wsIn struct {
	Type       string          `json:"type"`
	Seq        int64           `json:"seq"`
	ClientOpID string          `json:"clientOpId"`
	RequestID  string          `json:"requestId"`
	Payload    json.RawMessage `json:"payload"`
}

func (c *wsTestClient) sendRequest(typ string, payload any, requestID string) {
	c.t.Helper()
	msg := map[string]any{"type": typ, "pageId": testPage, "timestamp": time.Now().UnixMilli(), "payload": payload, "requestId": requestID}
	if err := c.conn.WriteJSON(msg); err != nil {
		c.t.Fatalf("write: %v", err)
	}
}

// next reads the next message of a type, skipping others (e.g. sync:project, presence).
func (c *wsTestClient) next(typ string, wait time.Duration) (wsIn, bool) {
	c.t.Helper()
	deadline := time.Now().Add(wait)
	for {
		_ = c.conn.SetReadDeadline(deadline)
		var m wsIn
		if err := c.conn.ReadJSON(&m); err != nil {
			return wsIn{}, false
		}
		if m.Type == typ {
			return m, true
		}
	}
}

func startWSServer(t *testing.T) (url string, docs *memDocRepo) {
	t.Helper()
	return startWSServerAs(t, func(n int) string { return fmt.Sprintf("user-%d", n) })
}

// startWSServerAs authenticates the n-th connection (from 1) as userID(n).
func startWSServerAs(t *testing.T, userID func(n int) string) (url string, docs *memDocRepo) {
	t.Helper()
	docs = newMemDocRepo()
	el := entity.Element{ID: testElemID, Type: "text", PageID: testPage, Styles: map[string]any{"color": "black"}}
	repo := &domainRepo.Repository{DocumentRepo: docs, ElementRepo: &memElementRepo{elements: []entity.Element{el}}}
	docUC := usecase.NewDocumentUseCase(repo)
	uc := &domainUsecase.UseCase{
		Element:          stubElementUC{usecase.NewElementUseCase(repo, docUC, nil)},
		Document:         docUC,
		Workflow:         stubWorkflowUC{},
		ElementWorkflows: stubElementWorkflowsUC{},
	}

	h := handlers.NewHandler(uc, nil, &mockLogger{})
	t.Cleanup(h.Shutdown)

	app := fiber.New()
	var n int
	var mu sync.Mutex
	app.Use("/ws", func(c fiber.Ctx) error {
		mu.Lock()
		n++
		id := userID(n)
		mu.Unlock()
		c.Locals("user", &entity.User{ID: id, Email: id + "@example.com"})
		return c.Next()
	})
	h.SetupRoutes(app)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("cannot listen on loopback: %v", err)
	}
	go func() { _ = app.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true}) }()
	t.Cleanup(func() { _ = app.Shutdown() })
	return "ws://" + ln.Addr().String() + "/ws/" + testProject, docs
}

func dialAndJoin(t *testing.T, url string) (*wsTestClient, wsIn) {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	c := &wsTestClient{t: t, conn: conn}
	c.send("join", map[string]any{"pageId": testPage}, "")
	sync, ok := c.next("sync:page", 3*time.Second)
	if !ok {
		t.Fatalf("no sync:page after join")
	}
	return c, sync
}

func TestWS_SenderReceivesOwnEchoWithSeqAndClientOpID(t *testing.T) {
	url, _ := startWSServer(t)
	a, syncA := dialAndJoin(t, url)
	b, _ := dialAndJoin(t, url)

	var syncPayload struct {
		Seq *int64 `json:"seq"`
	}
	if err := json.Unmarshal(syncA.Payload, &syncPayload); err != nil || syncPayload.Seq == nil || *syncPayload.Seq != 0 {
		t.Fatalf("sync:page must carry seq=0 for a fresh page, payload=%s", syncA.Payload)
	}

	// 1st op: success -> both sender and other client receive seq 1 with the sender's clientOpId.
	a.send("element:update", map[string]any{"id": testElemID, "styles": map[string]any{"color": "red"}}, "op-1")
	for name, c := range map[string]*wsTestClient{"sender": a, "other": b} {
		m, ok := c.next("element:update", 3*time.Second)
		if !ok {
			t.Fatalf("%s did not receive element:update", name)
		}
		if m.Seq != 1 || m.ClientOpID != "op-1" {
			t.Fatalf("%s got seq=%d clientOpId=%q, want 1/op-1", name, m.Seq, m.ClientOpID)
		}
	}

	// failed op: sender gets an error with its clientOpId, nobody gets a broadcast, seq not consumed.
	a.send("element:update", map[string]any{"id": "22222222-2222-2222-2222-222222222222", "styles": map[string]any{"x": 1}}, "op-bad")
	em, ok := a.next("error", 3*time.Second)
	if !ok || em.ClientOpID != "op-bad" {
		t.Fatalf("expected error echoing op-bad, got ok=%v msg=%+v", ok, em)
	}
	var ep struct {
		Code       string `json:"code"`
		ClientOpID string `json:"clientOpId"`
	}
	_ = json.Unmarshal(em.Payload, &ep)
	if ep.Code != "NOT_FOUND" || ep.ClientOpID != "op-bad" {
		t.Fatalf("unexpected error payload: %s", em.Payload)
	}
	// (A broadcast of the failed op would reach B first and break the seq checks below;
	// read timeouts are fatal for these connections, so we assert by ordering instead.)

	// next success gets seq 2 (no gap), delivered to the sender as its echo.
	b.send("element:update", map[string]any{"id": testElemID, "styles": map[string]any{"color": "blue"}}, "op-b")
	m, ok := a.next("element:update", 3*time.Second)
	if !ok || m.Seq != 2 || m.ClientOpID != "op-b" {
		t.Fatalf("expected seq=2 op-b at client A, got ok=%v %+v", ok, m)
	}
	if mb, ok := b.next("element:update", 3*time.Second); !ok || mb.Seq != 2 || mb.ClientOpID != "op-b" {
		t.Fatalf("sender B missing own echo: ok=%v %+v", ok, mb)
	}

	// a late joiner's sync:page starts from the current seq.
	_, syncC := dialAndJoin(t, url)
	_ = json.Unmarshal(syncC.Payload, &syncPayload)
	if syncPayload.Seq == nil || *syncPayload.Seq != 2 {
		t.Fatalf("late sync:page seq should be 2, payload=%s", syncC.Payload)
	}
}

func TestWS_ConcurrentOpsAreTotallyOrdered(t *testing.T) {
	url, docs := startWSServer(t)
	const clients, perClient = 4, 5
	cs := make([]*wsTestClient, clients)
	for i := range cs {
		cs[i], _ = dialAndJoin(t, url)
	}

	var wg sync.WaitGroup
	for i, c := range cs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perClient; j++ {
				c.send("element:update", map[string]any{"id": testElemID, "styles": map[string]any{fmt.Sprintf("k%d_%d", i, j): j}}, fmt.Sprintf("c%d-%d", i, j))
			}
		}()
	}
	wg.Wait()

	total := clients * perClient
	for ci, c := range cs {
		for want := 1; want <= total; want++ {
			m, ok := c.next("element:update", 5*time.Second)
			if !ok {
				t.Fatalf("client %d: missing seq %d", ci, want)
			}
			if m.Seq != int64(want) {
				t.Fatalf("client %d: out of order, got seq %d want %d", ci, m.Seq, want)
			}
		}
	}

	el := findElementByID(docs.snapshot(testProject, testPage), testElemID)
	if el == nil || len(el.Styles) != total+1 {
		t.Fatalf("expected %d style keys, got %v", total+1, el)
	}
}

func TestWS_RequestIDIsEchoedOnReplies(t *testing.T) {
	url, _ := startWSServer(t)
	a, _ := dialAndJoin(t, url)

	a.sendRequest("element:update", map[string]any{"id": testElemID, "styles": map[string]any{"color": "red"}}, "req-ok")
	if m, ok := a.next("element:update", 3*time.Second); !ok || m.RequestID != "req-ok" {
		t.Fatalf("element echo must carry requestId, ok=%v %+v", ok, m)
	}

	a.sendRequest("element:update", map[string]any{"id": "22222222-2222-2222-2222-222222222222", "styles": map[string]any{"x": 1}}, "req-bad")
	if m, ok := a.next("error", 3*time.Second); !ok || m.RequestID != "req-bad" {
		t.Fatalf("error reply must carry requestId, ok=%v %+v", ok, m)
	}

	a.sendRequest("sync:page", map[string]any{}, "req-sync")
	if m, ok := a.next("sync:page", 3*time.Second); !ok || m.RequestID != "req-sync" {
		t.Fatalf("sync:page reply must carry requestId, ok=%v %+v", ok, m)
	}
}

// A user may hold one connection per project. A newer connection replaces the
// older one, which must be told so with CloseReplaced; otherwise the old tab
// reconnects and the two connections keep replacing each other.
func TestWS_ReplacedConnectionGetsReplacedCloseCode(t *testing.T) {
	url, _ := startWSServerAs(t, func(int) string { return "same-user" })
	old, _ := dialAndJoin(t, url)
	_, _ = dialAndJoin(t, url)

	_ = old.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		_, _, err := old.conn.ReadMessage()
		if err == nil {
			continue
		}
		var ce *websocket.CloseError
		if !errors.As(err, &ce) {
			t.Fatalf("old connection ended without a close frame: %v", err)
		}
		if ce.Code != wsentity.CloseReplaced {
			t.Fatalf("close code = %d, want %d", ce.Code, wsentity.CloseReplaced)
		}
		return
	}
}
