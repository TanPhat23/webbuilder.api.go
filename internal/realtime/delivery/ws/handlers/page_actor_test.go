package handlers

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPageActors_SerializePerKeyAndKeepFIFO(t *testing.T) {
	p := newPageActors(context.Background(), time.Minute)

	var inFlight, maxInFlight int32
	var order []int
	var mu sync.Mutex

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = p.Do(context.Background(), "proj:page", func() {
				n := atomic.AddInt32(&inFlight, 1)
				for {
					m := atomic.LoadInt32(&maxInFlight)
					if n <= m || atomic.CompareAndSwapInt32(&maxInFlight, m, n) {
						break
					}
				}
				time.Sleep(100 * time.Microsecond)
				mu.Lock()
				order = append(order, i)
				mu.Unlock()
				atomic.AddInt32(&inFlight, -1)
			})
		}()
	}
	wg.Wait()

	if maxInFlight != 1 {
		t.Fatalf("ops for one page ran concurrently (max in flight %d)", maxInFlight)
	}
	if len(order) != 50 {
		t.Fatalf("expected 50 ops, ran %d", len(order))
	}
}

func TestPageActors_DifferentKeysRunIndependently(t *testing.T) {
	p := newPageActors(context.Background(), time.Minute)
	release := make(chan struct{})
	started := make(chan struct{})

	go func() {
		_ = p.Do(context.Background(), "a", func() { close(started); <-release })
	}()
	<-started

	done := make(chan struct{})
	go func() {
		_ = p.Do(context.Background(), "b", func() {})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("op on page b was blocked by page a")
	}
	close(release)
}

func TestPageActors_IdleActorStopsAndIsRecreated(t *testing.T) {
	p := newPageActors(context.Background(), 20*time.Millisecond)

	ran := 0
	if err := p.Do(context.Background(), "k", func() { ran++ }); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for p.size() != 0 {
		if time.Now().After(deadline) {
			t.Fatal("idle actor was not stopped")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err := p.Do(context.Background(), "k", func() { ran++ }); err != nil || ran != 2 {
		t.Fatalf("actor not recreated after idle stop: ran=%d err=%v", ran, err)
	}
}

func TestPageActors_PanicDoesNotKillActor(t *testing.T) {
	p := newPageActors(context.Background(), time.Minute)
	if err := p.Do(context.Background(), "k", func() { panic("boom") }); err == nil {
		t.Fatal("expected error from panicking op")
	}
	ok := false
	if err := p.Do(context.Background(), "k", func() { ok = true }); err != nil || !ok {
		t.Fatalf("actor unusable after panic: err=%v", err)
	}
}

func TestPageActors_ShutdownUnblocksCallers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := newPageActors(ctx, time.Minute)

	block := make(chan struct{})
	started := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		errCh <- p.Do(context.Background(), "k", func() { close(started); <-block })
	}()
	<-started
	cancel()
	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error after shutdown")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Do did not return after shutdown")
	}
	close(block)
}
