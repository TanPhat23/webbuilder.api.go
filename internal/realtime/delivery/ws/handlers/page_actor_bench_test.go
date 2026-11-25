package handlers

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkPageActorsDo runs trivial ops from many goroutines over many pages,
// which isolates the bookkeeping around each op (finding the page's actor and
// counting the op in and out) from the op itself.
func BenchmarkPageActorsDo(b *testing.B) {
	for _, pages := range []int{1, 64, 1024} {
		b.Run(fmt.Sprintf("pages=%d", pages), func(b *testing.B) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			p := newPageActors(ctx, time.Minute)
			keys := make([]string, pages)
			for i := range keys {
				keys[i] = fmt.Sprintf("proj:page-%d", i)
			}
			var seq atomic.Uint64
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				n := int(seq.Add(1)) * 7919
				for pb.Next() {
					n++
					if err := p.Run(ctx, keys[n%pages], func() {}); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}
