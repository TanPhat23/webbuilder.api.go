package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"my-go-app/internal/agentproto"
	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
	domainUsecase "my-go-app/internal/realtime/domain/usecase"
	"my-go-app/internal/realtime/utils/logger"
)

type agentHandlersRedisFailBus struct{ subErr error }

func (agentHandlersRedisFailBus) Publish(context.Context, string, []byte) error { return nil }
func (b agentHandlersRedisFailBus) PSubscribe(context.Context, string) (<-chan agentproto.BusMessage, func(), error) {
	return nil, nil, b.subErr
}

type agentHandlersRedisClaimer struct {
	won bool
	err error
}

func (c agentHandlersRedisClaimer) Claim(context.Context, string) (bool, error) { return c.won, c.err }

type agentHandlersRedisTab struct{ err error }

func (t agentHandlersRedisTab) Send(domainEntity.Message) error { return t.err }

func agentHandlersRedisRequestBytes(id string) []byte {
	b, _ := json.Marshal(agentproto.Request{RequestID: id, UserID: "u1", ProjectID: "p1", PageID: "pg1", Tool: "get_code", Args: json.RawMessage(`{}`)})
	return b
}

func TestRedisClaimer_OnlyFirstClaimWinsAndErrorsSurface(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()
	c := redisClaimer{rdb}
	won, err := c.Claim(context.Background(), "req-1")
	if err != nil || !won {
		t.Fatalf("first claim: %v %v", won, err)
	}
	if won, err = c.Claim(context.Background(), "req-1"); err != nil || won {
		t.Fatalf("second claim must lose: %v %v", won, err)
	}
	if won, _ = c.Claim(context.Background(), "req-2"); !won {
		t.Fatal("a different request must be claimable")
	}
	if ttl := mr.TTL("agent:claim:req-1"); ttl <= 0 || ttl > 30*time.Second {
		t.Fatalf("claim ttl = %v", ttl)
	}
	mr.Close()
	if _, err = c.Claim(context.Background(), "req-3"); err == nil {
		t.Fatal("claim against a dead Redis must fail")
	}
}

func TestHubLocator_ReturnsNilInterfaceWhenNoTabAndTheClientWhenPresent(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	loc := hubLocator{env.h.hub}
	if got := loc.MostRecentClient(handlerTestProject, "u1"); got != nil {
		t.Fatalf("no tab: got non-nil %v", got)
	}
	c := env.dial("u1")
	c.join(handlerTestPage)
	got := loc.MostRecentClient(handlerTestProject, "u1")
	if got == nil {
		t.Fatal("expected the joined tab")
	}
	if loc.MostRecentClient(handlerTestProject, "someone-else") != nil {
		t.Fatal("other user must not match")
	}
	if loc.MostRecentClient("other-project", "u1") != nil {
		t.Fatal("other project must not match")
	}
}

func TestHubLocator_SkipsWriterSockets(t *testing.T) {
	env := handlerTestNew(t, nil, nil)
	w := env.dialQuery("u=u1&conn=writer")
	w.join(handlerTestPage)
	if env.h.hub.MostRecentClient(handlerTestProject, "u1") != nil {
		t.Fatal("a writer socket must never be chosen as the editor tab")
	}
}

func TestAgentBridge_RunStopsWhenSubscribeFails(t *testing.T) {
	b := newAgentBridge(fakeLocator{}, agentHandlersRedisFailBus{subErr: errors.New("no bus")}, newMemClaimer(), logger.New(nil, "", 0))
	done := make(chan struct{})
	go func() { b.Run(context.Background()); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run must return when it cannot subscribe")
	}
}

func TestAgentBridge_RunIgnoresBadPayloadsAndStopsWhenBusCloses(t *testing.T) {
	tab := &fakeTab{}
	bus := &agentproto.MemBus{}
	b := newAgentBridge(fakeLocator{byUser: map[string]agentTarget{"p1/u1": tab}}, bus, newMemClaimer(), handlerTestLogger())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { b.Run(ctx); close(done) }()
	time.Sleep(30 * time.Millisecond)

	for _, bad := range [][]byte{
		[]byte("not json"),
		[]byte(`{"userId":"u1","projectId":"p1"}`),   // no requestId
		[]byte(`{"requestId":"x","projectId":"p1"}`), // no userId
	} {
		_ = bus.Publish(context.Background(), agentproto.RequestChannelPrefix+"u1", bad)
	}
	_ = bus.Publish(context.Background(), agentproto.RequestChannelPrefix+"u1", agentHandlersRedisRequestBytes("good"))
	waitFor(t, "good request delivered", func() bool { return tab.count() == 1 })
	if tab.count() != 1 {
		t.Fatalf("bad payloads must not be delivered, tab got %d", tab.count())
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run must stop on ctx cancel")
	}
}

func TestAgentBridge_DeliverSkipsWhenClaimLostOrFails(t *testing.T) {
	for name, claim := range map[string]agentHandlersRedisClaimer{
		"lost":  {won: false},
		"error": {err: errors.New("redis down")},
	} {
		bus := &agentproto.MemBus{}
		acks, stop, _ := bus.PSubscribe(context.Background(), agentproto.AckChannelPrefix+"r")
		tab := &fakeTab{}
		b := newAgentBridge(fakeLocator{byUser: map[string]agentTarget{"p1/u1": tab}}, bus, claim, handlerTestLogger())
		b.deliver(context.Background(), agentproto.Request{RequestID: "r", UserID: "u1", ProjectID: "p1"})
		if tab.count() != 0 || len(b.issued) != 0 {
			t.Fatalf("%s: delivered despite no claim", name)
		}
		select {
		case <-acks:
			t.Fatalf("%s: acked without claim", name)
		case <-time.After(30 * time.Millisecond):
		}
		stop()
	}
}

func TestAgentBridge_FailedSendForgetsTheRequestAndDoesNotAck(t *testing.T) {
	bus := &agentproto.MemBus{}
	acks, stop, _ := bus.PSubscribe(context.Background(), agentproto.AckChannelPrefix+"r")
	defer stop()
	tab := agentHandlersRedisTab{err: errors.New("socket closed")}
	b := newAgentBridge(fakeLocator{byUser: map[string]agentTarget{"p1/u1": tab}}, bus, newMemClaimer(), handlerTestLogger())
	b.deliver(context.Background(), agentproto.Request{RequestID: "r", UserID: "u1", ProjectID: "p1"})
	if len(b.issued) != 0 {
		t.Fatal("a request that could not be sent must not stay addressable")
	}
	select {
	case <-acks:
		t.Fatal("must not ack a failed delivery")
	case <-time.After(30 * time.Millisecond):
	}
	err := b.handleResponse(context.Background(), tab, domainEntity.Message{RequestID: "r", Payload: json.RawMessage(`{"ok":true}`)})
	if !domainErrors.IsCode(err, domainErrors.CodeValidation) {
		t.Fatalf("response for an unsent request: %v", err)
	}
}

func TestAgentBridge_UnreadableResponseIsReportedAsInternalErrorToTheRelay(t *testing.T) {
	bus := &agentproto.MemBus{}
	resps, stop, _ := bus.PSubscribe(context.Background(), agentproto.ResponseChannelPrefix+"r")
	defer stop()
	tab := &fakeTab{}
	b := newAgentBridge(fakeLocator{byUser: map[string]agentTarget{"p1/u1": tab}}, bus, newMemClaimer(), handlerTestLogger())
	b.deliver(context.Background(), agentproto.Request{RequestID: "r", UserID: "u1", ProjectID: "p1"})
	if err := b.handleResponse(context.Background(), tab, domainEntity.Message{RequestID: "r", Payload: json.RawMessage(`"just a string"`)}); err != nil {
		t.Fatal(err)
	}
	select {
	case m := <-resps:
		var r agentproto.Response
		_ = json.Unmarshal(m.Payload, &r)
		if r.OK || r.RequestID != "r" || r.Error == nil || r.Error.Code != agentproto.CodeInternal {
			t.Fatalf("%+v", r)
		}
	case <-time.After(time.Second):
		t.Fatal("no response published")
	}
}

// Full path over real sockets and a (fake) Redis: relay request -> editor tab -> relay response.
func TestAgentBridge_EndToEndThroughHandlerWithRedis(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	els := &handlerFakeElements{}
	h := NewHandler(&domainUsecase.UseCase{Element: els}, rdb, handlerTestLogger())
	t.Cleanup(h.Shutdown)
	if h.agents == nil {
		t.Fatal("agent bridge must exist when Redis is configured")
	}
	env := handlerTestServe(t, h, els)
	tab := env.dial("u1")
	tab.join(handlerTestPage)

	bus := agentproto.NewRedisBus(rdb)
	acks, stopAck, err := bus.PSubscribe(context.Background(), agentproto.AckChannelPrefix+"e2e")
	if err != nil {
		t.Fatal(err)
	}
	defer stopAck()
	resps, stopResp, err := bus.PSubscribe(context.Background(), agentproto.ResponseChannelPrefix+"e2e")
	if err != nil {
		t.Fatal(err)
	}
	defer stopResp()

	req := agentproto.Request{RequestID: "e2e", UserID: "u1", ProjectID: handlerTestProject, PageID: handlerTestPage, Tool: "get_code", Args: json.RawMessage(`{"a":1}`)}
	body, _ := json.Marshal(req)
	// Run subscribes asynchronously; publish until the tab hears it.
	deadline := time.Now().Add(3 * time.Second)
	var delivered handlerMsg
	got := make(chan handlerMsg, 1)
	go func() { got <- tab.next("agent:request") }()
	for published := false; !published || time.Now().Before(deadline); {
		_ = bus.Publish(context.Background(), agentproto.RequestChannelPrefix+"u1", body)
		published = true
		select {
		case delivered = <-got:
			goto delivered
		case <-time.After(100 * time.Millisecond):
		}
		if time.Now().After(deadline) {
			t.Fatal("tab never received the agent request")
		}
	}
delivered:
	if delivered.RequestID != "e2e" || delivered.PageID != handlerTestPage {
		t.Fatalf("%+v", delivered)
	}
	var payload struct {
		Tool   string          `json:"tool"`
		Args   json.RawMessage `json:"args"`
		PageID string          `json:"pageId"`
	}
	delivered.decode(t, &payload)
	if payload.Tool != "get_code" || string(payload.Args) != `{"a":1}` || payload.PageID != handlerTestPage {
		t.Fatalf("payload %+v", payload)
	}
	select {
	case <-acks:
	case <-time.After(2 * time.Second):
		t.Fatal("no ack")
	}

	// A different user's tab cannot answer it.
	intruder := env.dial("u2")
	intruder.send("agent:response", map[string]any{"ok": true}, "", "e2e")
	intruder.expectError("VALIDATION_ERROR", "unknown agent request")

	tab.send("agent:response", map[string]any{"ok": true, "result": map[string]any{"code": "x"}}, "", "e2e")
	select {
	case m := <-resps:
		var r agentproto.Response
		_ = json.Unmarshal(m.Payload, &r)
		if !r.OK || r.RequestID != "e2e" {
			t.Fatalf("%+v", r)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no response published")
	}
}
