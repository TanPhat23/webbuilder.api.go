package agentproto

import (
	"context"
	"testing"
	"time"
)

func recv(t *testing.T, ch <-chan BusMessage) BusMessage {
	t.Helper()
	select {
	case m := <-ch:
		return m
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for a message")
		return BusMessage{}
	}
}

func TestMemBusExactAndPrefix(t *testing.T) {
	var b MemBus
	ctx := context.Background()
	exact, stopExact, _ := b.PSubscribe(ctx, "agent:ack:r1")
	defer stopExact()
	prefix, stopPrefix, _ := b.PSubscribe(ctx, "agent:req:*")
	defer stopPrefix()

	_ = b.Publish(ctx, "agent:req:user-1", []byte("a"))
	_ = b.Publish(ctx, "agent:ack:r1", []byte("b"))
	_ = b.Publish(ctx, "agent:ack:r2", []byte("not for us"))

	if m := recv(t, prefix); string(m.Payload) != "a" || m.Channel != "agent:req:user-1" {
		t.Fatalf("prefix got %+v", m)
	}
	if m := recv(t, exact); string(m.Payload) != "b" {
		t.Fatalf("exact got %+v", m)
	}
	select {
	case m := <-exact:
		t.Fatalf("exact subscription received another channel: %+v", m)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestMemBusStopEndsDelivery(t *testing.T) {
	var b MemBus
	ch, stop, _ := b.PSubscribe(context.Background(), "x:*")
	stop()
	stop() // idempotent
	_ = b.Publish(context.Background(), "x:1", []byte("late"))
	if _, ok := <-ch; ok {
		t.Fatal("expected the channel to be closed after stop")
	}
}
