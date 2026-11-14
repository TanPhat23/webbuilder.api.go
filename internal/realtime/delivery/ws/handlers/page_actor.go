package handlers

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	pageActorQueue = 256
	pageActorIdle  = 30 * time.Second
)

// pageActors serializes work per page on this server instance. Each page key (project+page)
// gets a lazily created actor: one goroutine draining an op queue in FIFO order. The actor
// stops itself after being idle (which is also what happens once a room empties).
//
// Running "mutate document + broadcast" inside one actor op guarantees that, on this instance,
// ops of a page are applied, sequenced and enqueued for broadcast strictly one at a time and in
// seq order. Cross-instance safety is provided by the optimistic transaction in the repository.
type pageActors struct {
	ctx  context.Context
	idle time.Duration

	mu     sync.Mutex
	actors map[string]*pageActor
}

type pageActor struct {
	ops  chan *pageOp
	refs int // ops submitted but not yet finished; guarded by pageActors.mu
}

type pageOp struct {
	fn   func()
	done chan struct{}
	err  error
}

func newPageActors(ctx context.Context, idle time.Duration) *pageActors {
	if idle <= 0 {
		idle = pageActorIdle
	}
	return &pageActors{ctx: ctx, idle: idle, actors: make(map[string]*pageActor)}
}

// Do runs fn on the actor for key and waits for it to finish. Calls for the same key never
// run concurrently and run in submission order. It returns an error if the actors were shut
// down, ctx ended before fn was queued, or fn panicked.
func (p *pageActors) Do(ctx context.Context, key string, fn func()) error {
	p.mu.Lock()
	a := p.actors[key]
	if a == nil {
		a = &pageActor{ops: make(chan *pageOp, pageActorQueue)}
		p.actors[key] = a
		go p.run(key, a)
	}
	a.refs++
	p.mu.Unlock()

	op := &pageOp{fn: fn, done: make(chan struct{})}
	select {
	case a.ops <- op:
	case <-ctx.Done():
		p.release(a)
		return ctx.Err()
	case <-p.ctx.Done():
		p.release(a)
		return p.ctx.Err()
	}

	select {
	case <-op.done:
		return op.err
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
}

func (p *pageActors) release(a *pageActor) {
	p.mu.Lock()
	a.refs--
	p.mu.Unlock()
}

func (p *pageActors) run(key string, a *pageActor) {
	timer := time.NewTimer(p.idle)
	defer timer.Stop()

	for {
		select {
		case op := <-a.ops:
			p.exec(op)
			p.release(a)
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(p.idle)

		case <-timer.C:
			p.mu.Lock()
			if a.refs == 0 {
				delete(p.actors, key)
				p.mu.Unlock()
				return
			}
			p.mu.Unlock()
			timer.Reset(p.idle)

		case <-p.ctx.Done():
			p.mu.Lock()
			if p.actors[key] == a {
				delete(p.actors, key)
			}
			p.mu.Unlock()
			return
		}
	}
}

func (p *pageActors) exec(op *pageOp) {
	defer func() {
		if r := recover(); r != nil {
			op.err = fmt.Errorf("page op panicked: %v", r)
		}
		close(op.done)
	}()
	op.fn()
}

// size reports the number of live actors (for tests).
func (p *pageActors) size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.actors)
}
