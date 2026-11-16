package agentproto

import (
	"context"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
)

type BusMessage struct {
	Channel string
	Payload []byte
}

// Bus is the slice of Redis pub/sub the agent relay needs.
type Bus interface {
	Publish(ctx context.Context, channel string, payload []byte) error
	// PSubscribe returns once the subscription is active, so a publish that
	// follows is never missed. A trailing * in pattern matches any suffix.
	PSubscribe(ctx context.Context, pattern string) (<-chan BusMessage, func(), error)
}

type RedisBus struct{ rdb *redis.Client }

func NewRedisBus(rdb *redis.Client) Bus { return &RedisBus{rdb: rdb} }

func (b *RedisBus) Publish(ctx context.Context, channel string, payload []byte) error {
	return b.rdb.Publish(ctx, channel, payload).Err()
}

func (b *RedisBus) PSubscribe(ctx context.Context, pattern string) (<-chan BusMessage, func(), error) {
	ps := b.rdb.PSubscribe(ctx, pattern)
	if _, err := ps.Receive(ctx); err != nil { // waits for the subscribe confirmation
		_ = ps.Close()
		return nil, nil, err
	}
	out := make(chan BusMessage, 64)
	done := make(chan struct{})
	go func() {
		defer close(out)
		in := ps.Channel()
		for {
			select {
			case m, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- BusMessage{Channel: m.Channel, Payload: []byte(m.Payload)}:
				default: // slow consumer: drop rather than block Redis
				}
			case <-done:
				return
			}
		}
	}()
	var once sync.Once
	return out, func() { once.Do(func() { close(done); _ = ps.Close() }) }, nil
}

// MemBus is an in-process Bus for tests. The zero value is ready to use.
type MemBus struct {
	mu   sync.Mutex
	subs map[*memSub]struct{}
}

type memSub struct {
	pattern string
	ch      chan BusMessage
	closed  bool
}

func matches(pattern, channel string) bool {
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(channel, strings.TrimSuffix(pattern, "*"))
	}
	return pattern == channel
}

func (b *MemBus) Publish(_ context.Context, channel string, payload []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for s := range b.subs {
		if s.closed || !matches(s.pattern, channel) {
			continue
		}
		select {
		case s.ch <- BusMessage{Channel: channel, Payload: payload}:
		default:
		}
	}
	return nil
}

func (b *MemBus) PSubscribe(_ context.Context, pattern string) (<-chan BusMessage, func(), error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.subs == nil {
		b.subs = map[*memSub]struct{}{}
	}
	s := &memSub{pattern: pattern, ch: make(chan BusMessage, 64)}
	b.subs[s] = struct{}{}
	return s.ch, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if !s.closed {
			s.closed = true
			delete(b.subs, s)
			close(s.ch)
		}
	}, nil
}
