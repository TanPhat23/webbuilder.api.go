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
	el := entity.Element{ID: testElemID, Type: "text", PageID: testPage, Styles: map[string]any{"color": "black"}}
	return startWSServerWithElementsAs(t, &memElementRepo{elements: []entity.Element{el}}, userID)
}

func startWSServerWithElements(t *testing.T, elements domainRepo.ElementRepository) (url string, docs *memDocRepo) {
	t.Helper()
	return startWSServerWithElementsAs(t, elements, func(n int) string { return fmt.Sprintf("user-%d", n) })
}

func startWSServerWithElementsAs(t *testing.T, elements domainRepo.ElementRepository, userID func(n int) string) (url string, docs *memDocRepo) {
	t.Helper()
	docs = newMemDocRepo()
	repo := &domainRepo.Repository{DocumentRepo: docs, ElementRepo: elements}
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

type seqPayload struct {
	Seq *int64 `json:"seq"`
}

func assertWSSyncSeq(t *testing.T, payload json.RawMessage, want int64, label string) {
	t.Helper()
	var syncPayload seqPayload
	if err := json.Unmarshal(payload, &syncPayload); err != nil {
		t.Fatalf("%s unmarshal sync:page: %v", label, err)
	}
	if syncPayload.Seq == nil || *syncPayload.Seq != want {
		t.Fatalf("%s sync:page seq=%v want=%d, payload=%s", label, syncPayload.Seq, want, payload)
	}
}

func assertWSEchoBroadcast(t *testing.T, clients map[string]*wsTestClient, wantSeq int64, wantOpID string) {
	t.Helper()
	for name, c := range clients {
		m, ok := c.next("element:update", 3*time.Second)
		if !ok {
			t.Fatalf("%s did not receive element:update", name)
		}
		if m.Seq != wantSeq || m.ClientOpID != wantOpID {
			t.Fatalf("%s got seq=%d clientOpId=%q, want %d/%s", name, m.Seq, m.ClientOpID, wantSeq, wantOpID)
		}
	}
}

func assertWSEchoErrorReply(t *testing.T, a *wsTestClient, wantCode, wantOpID string) {
	t.Helper()
	em, ok := a.next("error", 3*time.Second)
	if !ok || em.ClientOpID != wantOpID {
		t.Fatalf("expected error echoing %s, got ok=%v msg=%+v", wantOpID, ok, em)
	}
	var ep struct {
		Code       string `json:"code"`
		ClientOpID string `json:"clientOpId"`
	}
	if err := json.Unmarshal(em.Payload, &ep); err != nil {
		t.Fatalf("unmarshal error payload: %v", err)
	}
	if ep.Code != wantCode || ep.ClientOpID != wantOpID {
		t.Fatalf("unexpected error payload: %s", em.Payload)
	}
}

func TestWS_SenderReceivesOwnEchoWithSeqAndClientOpID(t *testing.T) {
	url, docs := startWSServer(t)
	if docs == nil {
		t.Fatal("missing docs repo")
	}
	a, syncA := dialAndJoin(t, url)
	b, syncB := dialAndJoin(t, url)
	if syncB.Type != "sync:page" {
		t.Fatalf("unexpected syncB type: %s", syncB.Type)
	}
	assertWSSyncSeq(t, syncA.Payload, 0, "fresh page")

	// 1st op: success -> both sender and other client receive seq 1 with the sender's clientOpId.
	a.send("element:update", map[string]any{"id": testElemID, "styles": map[string]any{"color": "red"}}, "op-1")
	assertWSEchoBroadcast(t, map[string]*wsTestClient{"sender": a, "other": b}, 1, "op-1")

	// failed op: sender gets an error with its clientOpId, nobody gets a broadcast, seq not consumed.
	a.send("element:update", map[string]any{"id": "22222222-2222-2222-2222-222222222222", "styles": map[string]any{"x": 1}}, "op-bad")
	assertWSEchoErrorReply(t, a, "NOT_FOUND", "op-bad")

	// next success gets seq 2 (no gap), delivered to the sender as its echo.
	b.send("element:update", map[string]any{"id": testElemID, "styles": map[string]any{"color": "blue"}}, "op-b")
	assertWSEchoBroadcast(t, map[string]*wsTestClient{"client A": a, "sender B": b}, 2, "op-b")

	// a late joiner's sync:page starts from the current seq.
	c, syncC := dialAndJoin(t, url)
	if c == nil {
		t.Fatal("missing client c")
	}
	assertWSSyncSeq(t, syncC.Payload, 2, "late joiner")
}

func sendClientOpsBatch(cs []*wsTestClient, perClient int) {
	var wg sync.WaitGroup
	for i, c := range cs {
		wg.Add(1)
		go func(idx int, client *wsTestClient) {
			defer wg.Done()
			for j := 0; j < perClient; j++ {
				client.send("element:update", map[string]any{"id": testElemID, "styles": map[string]any{fmt.Sprintf("k%d_%d", idx, j): j}}, fmt.Sprintf("c%d-%d", idx, j))
			}
		}(i, c)
	}
	wg.Wait()
}

func assertOrderedSeqForClients(t *testing.T, cs []*wsTestClient, total int) {
	t.Helper()
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
}

func TestWS_ConcurrentOpsAreTotallyOrdered(t *testing.T) {
	url, docs := startWSServer(t)
	const clients, perClient = 4, 5
	cs := make([]*wsTestClient, clients)
	for i := range cs {
		var sync wsIn
		cs[i], sync = dialAndJoin(t, url)
		if sync.Type != "sync:page" {
			t.Fatalf("client %d unexpected sync: %+v", i, sync)
		}
	}

	sendClientOpsBatch(cs, perClient)
	total := clients * perClient
	assertOrderedSeqForClients(t, cs, total)

	el := findElementByID(docs.snapshot(testProject, testPage), testElemID)
	if el == nil || len(el.Styles) != total+1 {
		t.Fatalf("expected %d style keys, got %v", total+1, el)
	}
}

func TestWS_RequestIDIsEchoedOnReplies(t *testing.T) {
	url, docs := startWSServer(t)
	if docs == nil {
		t.Fatal("missing docs repo")
	}
	a, syncA := dialAndJoin(t, url)
	if syncA.Type != "sync:page" {
		t.Fatalf("unexpected syncA: %+v", syncA)
	}

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
	url, docs := startWSServerAs(t, func(int) string { return "same-user" })
	if docs == nil {
		t.Fatal("missing docs repo")
	}
	old, syncOld := dialAndJoin(t, url)
	newer, syncNew := dialAndJoin(t, url)
	if syncOld.Type != "sync:page" || syncNew.Type != "sync:page" || newer == nil {
		t.Fatalf("join failed: %+v / %+v", syncOld, syncNew)
	}

	if err := old.conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}
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
