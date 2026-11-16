package handlers

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"my-go-app/internal/agentproto"
	domainEntity "my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/utils/logger"
)

type fakeTab struct {
	mu   sync.Mutex
	sent []domainEntity.Message
}

func (f *fakeTab) Send(m domainEntity.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, m)
	return nil
}
func (f *fakeTab) count() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.sent) }

type fakeLocator struct{ byUser map[string]agentTarget } // key: project/user

func (f fakeLocator) MostRecentClient(project, user string) agentTarget {
	if t, ok := f.byUser[project+"/"+user]; ok {
		return t
	}
	return nil
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func request(bus agentproto.Bus, id string) {
	b, _ := json.Marshal(agentproto.Request{RequestID: id, UserID: "u1", ProjectID: "p1", PageID: "pg1", Tool: "get_code", Args: json.RawMessage(`{}`)})
	_ = bus.Publish(context.Background(), agentproto.RequestChannelPrefix+"u1", b)
}

func startBridge(t *testing.T, loc agentLocator, bus agentproto.Bus) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	b := newAgentBridge(loc, bus, newMemClaimer(), logger.Default())
	ready := make(chan struct{})
	go func() { b.Run(ctx); close(ready) }()
	time.Sleep(20 * time.Millisecond) // Run subscribes first thing
}

type memClaimer struct {
	mu   sync.Mutex
	seen map[string]bool
}

func newMemClaimer() *memClaimer { return &memClaimer{seen: map[string]bool{}} }
func (m *memClaimer) Claim(_ context.Context, id string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.seen[id] {
		return false, nil
	}
	m.seen[id] = true
	return true, nil
}

// Review Focus 1: two API instances both hear the request; only one delivers.
func TestTwoInstancesDeliverARequestToExactlyOneTab(t *testing.T) {
	bus := &agentproto.MemBus{}
	claims := newMemClaimer() // shared, like Redis
	tabA, tabB := &fakeTab{}, &fakeTab{}
	for _, tab := range []*fakeTab{tabA, tabB} {
		b := newAgentBridge(fakeLocator{byUser: map[string]agentTarget{"p1/u1": tab}}, bus, claims, logger.Default())
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		go b.Run(ctx)
	}
	time.Sleep(30 * time.Millisecond)
	request(bus, "r4")
	waitFor(t, "a delivery", func() bool { return tabA.count()+tabB.count() >= 1 })
	time.Sleep(100 * time.Millisecond)
	if got := tabA.count() + tabB.count(); got != 1 {
		t.Fatalf("the request reached %d tabs, want exactly 1", got)
	}
}

func TestBridgeDeliversToTheTabAcksAndReturnsItsResponse(t *testing.T) {
	bus := &agentproto.MemBus{}
	tab := &fakeTab{}
	acks, stopAcks, _ := bus.PSubscribe(context.Background(), agentproto.AckChannelPrefix+"r1")
	defer stopAcks()
	resps, stopResps, _ := bus.PSubscribe(context.Background(), agentproto.ResponseChannelPrefix+"r1")
	defer stopResps()
	loc := fakeLocator{byUser: map[string]agentTarget{"p1/u1": tab}}
	b := newAgentBridge(loc, bus, newMemClaimer(), logger.Default())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)
	time.Sleep(20 * time.Millisecond)

	request(bus, "r1")
	waitFor(t, "ws delivery", func() bool { return tab.count() == 1 })
	if m := tab.sent[0]; m.Type != domainEntity.TypeAgentRequest || m.RequestID != "r1" || m.PageID != "pg1" {
		t.Fatalf("delivered %+v", m)
	}
	select {
	case <-acks:
	case <-time.After(time.Second):
		t.Fatal("no ack")
	}

	reply := domainEntity.Message{Type: domainEntity.TypeAgentResponse, RequestID: "r1", Payload: json.RawMessage(`{"ok":true,"result":{"code":"x"}}`)}
	if err := b.handleResponse(context.Background(), tab, reply); err != nil {
		t.Fatal(err)
	}
	select {
	case m := <-resps:
		var r agentproto.Response
		_ = json.Unmarshal(m.Payload, &r)
		if !r.OK || r.RequestID != "r1" {
			t.Fatalf("%+v", r)
		}
	case <-time.After(time.Second):
		t.Fatal("no response published")
	}
}

func TestBridgeStaysSilentWhenTheUserHasNoTab(t *testing.T) {
	bus := &agentproto.MemBus{}
	acks, stop, _ := bus.PSubscribe(context.Background(), agentproto.AckChannelPrefix+"r2")
	defer stop()
	startBridge(t, fakeLocator{}, bus)
	request(bus, "r2")
	select {
	case <-acks:
		t.Fatal("must not ack when there is nobody to deliver to")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestResponseFromANonAddressedSessionIsRejected(t *testing.T) {
	bus := &agentproto.MemBus{}
	addressed, other := &fakeTab{}, &fakeTab{}
	resps, stop, _ := bus.PSubscribe(context.Background(), agentproto.ResponseChannelPrefix+"r3")
	defer stop()
	b := newAgentBridge(fakeLocator{byUser: map[string]agentTarget{"p1/u1": addressed}}, bus, newMemClaimer(), logger.Default())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)
	time.Sleep(20 * time.Millisecond)
	request(bus, "r3")
	waitFor(t, "delivery", func() bool { return addressed.count() == 1 })

	forged := domainEntity.Message{Type: domainEntity.TypeAgentResponse, RequestID: "r3", Payload: json.RawMessage(`{"ok":true}`)}
	if err := b.handleResponse(context.Background(), other, forged); err == nil {
		t.Fatal("a response from a session that was not sent the request must be rejected")
	}
	select {
	case <-resps:
		t.Fatal("nothing may be published for a rejected response")
	case <-time.After(50 * time.Millisecond):
	}
	// A request id is single use: the right session's first answer works, a replay does not.
	if err := b.handleResponse(context.Background(), addressed, forged); err != nil {
		t.Fatal(err)
	}
	if err := b.handleResponse(context.Background(), addressed, forged); err == nil {
		t.Fatal("a replayed response must be rejected")
	}
}
