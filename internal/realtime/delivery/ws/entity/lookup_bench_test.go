package entity

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

// The hub resolves a room on every broadcast, presence update and participant
// list, from one goroutine per connection, while rooms are created and removed
// rarely. These benchmarks compare the candidate structures on exactly that.

type mutexRooms struct {
	mu sync.RWMutex
	m  map[string]*Room
}

func (r *mutexRooms) get(id string) *Room {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.m[id]
}
func (r *mutexRooms) put(id string, room *Room) { r.mu.Lock(); r.m[id] = room; r.mu.Unlock() }
func (r *mutexRooms) del(id string)             { r.mu.Lock(); delete(r.m, id); r.mu.Unlock() }

type syncMapRooms struct{ m sync.Map }

func (r *syncMapRooms) get(id string) *Room {
	if v, ok := r.m.Load(id); ok {
		return v.(*Room)
	}
	return nil
}
func (r *syncMapRooms) put(id string, room *Room) { r.m.Store(id, room) }
func (r *syncMapRooms) del(id string)             { r.m.Delete(id) }

type roomIndex interface {
	get(string) *Room
	put(string, *Room)
	del(string)
}

func benchLookup(b *testing.B, idx roomIndex, writePerMille int) {
	const rooms = 500
	ids := make([]string, rooms)
	for i := range ids {
		ids[i] = fmt.Sprintf("proj-%d:page", i)
		idx.put(ids[i], &Room{PageID: ids[i]})
	}
	var seq atomic.Uint64
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		n := int(seq.Add(1)) * 7919
		for pb.Next() {
			n++
			id := ids[n%rooms]
			if writePerMille > 0 && n%1000 < writePerMille {
				// a room closing and reopening: the rare write
				idx.del(id)
				idx.put(id, &Room{PageID: id})
				continue
			}
			// A room being replaced by another goroutine reads as missing, as it
			// would for a caller racing a room closing in the hub.
			_ = idx.get(id)
		}
	})
}

func BenchmarkRoomLookup(b *testing.B) {
	for _, mix := range []struct {
		name string
		w    int
	}{{"readonly", 0}, {"0.1pct-writes", 1}, {"1pct-writes", 10}} {
		b.Run("mutex/"+mix.name, func(b *testing.B) { benchLookup(b, &mutexRooms{m: map[string]*Room{}}, mix.w) })
		b.Run("syncmap/"+mix.name, func(b *testing.B) { benchLookup(b, &syncMapRooms{}, mix.w) })
	}
}
