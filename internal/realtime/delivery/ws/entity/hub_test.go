package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	domainEntity "my-go-app/internal/realtime/domain/entity"
)

func newTestHub(tb testing.TB) (*Hub, context.CancelFunc) {
	tb.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	hub := NewHub(ctx, &wg, 8, nil, quietLogger())
	tb.Cleanup(cancel)
	return hub, cancel
}

func roomCount(h *Hub) int {
	n := 0
	h.rooms.Range(func(_, _ any) bool { n++; return true })
	return n
}

func TestHubConcurrentGetOrCreateRoomYieldsOneRoom(t *testing.T) {
	hub, _ := newTestHub(t)
	const goroutines = 64
	got := make([]*Room, goroutines)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := range got {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			got[i] = hub.getOrCreateRoom("proj:page")
		}()
	}
	close(start)
	wg.Wait()

	for i, r := range got {
		if r != got[0] {
			t.Fatalf("goroutine %d got a different room", i)
		}
	}
	if n := roomCount(hub); n != 1 {
		t.Fatalf("hub holds %d rooms, want 1", n)
	}
	if got[0].ctx.Err() != nil {
		t.Fatal("the surviving room was cancelled by a losing creator")
	}
}

func TestHubReplacesAClosingRoom(t *testing.T) {
	hub, _ := newTestHub(t)
	old := hub.getOrCreateRoom("proj:page")
	old.cancel() // the room emptied but has not left the table yet

	fresh := hub.getOrCreateRoom("proj:page")
	if fresh == old {
		t.Fatal("handed out a room that is shutting down")
	}
	if fresh.ctx.Err() != nil {
		t.Fatal("replacement room is not running")
	}

	// The old room finishing its shutdown must not remove the room that replaced it.
	old.onEmpty("proj:page")
	if hub.room("proj:page") != fresh {
		t.Fatal("the old room's onEmpty removed its replacement")
	}
}

func TestHubBroadcastProjectFindsEveryRoomOfTheProject(t *testing.T) {
	hub, _ := newTestHub(t)
	for _, id := range []string{"p1", "p1:a", "p1:b", "p10:a", "p2:a"} {
		hub.getOrCreateRoom(id)
	}
	var seen []string
	hub.rooms.Range(func(k, _ any) bool { seen = append(seen, k.(string)); return true })
	if len(seen) != 5 {
		t.Fatalf("rooms = %v", seen)
	}
	// "p10:a" must not match project "p1".
	var matched int
	prefix := "p1:"
	for _, id := range seen {
		if id == "p1" || len(id) >= len(prefix) && id[:len(prefix)] == prefix {
			matched++
		}
	}
	if matched != 3 {
		t.Fatalf("project p1 matches %d rooms, want 3", matched)
	}
}

// Many goroutines opening rooms, registering, broadcasting and leaving at once.
// Run under -race: it fails on any unsynchronised access to the room table.
func TestHubConcurrentRoomChurn(t *testing.T) {
	hub, _ := newTestHub(t)
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 60; i++ {
				project := fmt.Sprintf("proj-%d", (g+i)%5)
				c := NewClient(nil, project, &domainEntity.User{ID: fmt.Sprintf("u-%d-%d", g, i)})
				c.SetPageID("page")
				ctx := context.Background()
				if err := hub.Register(ctx, c); err != nil {
					t.Errorf("register: %v", err)
					return
				}
				hub.Broadcast(c.RoomID(), domainEntity.Message{Type: domainEntity.TypePresence, ProjectID: project}, c)
				hub.BroadcastProject(project, domainEntity.Message{Type: domainEntity.TypePresence}, nil)
				_ = hub.GetRoomParticipants(c.RoomID())
				_ = hub.MostRecentClient(project, c.User.ID)
				if err := hub.Unregister(ctx, c); err != nil {
					t.Errorf("unregister: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

// Every client receives every broadcast exactly once, in order, as valid JSON, and the
// sender does not receive its own message.
func TestHubFanoutDeliversEachFrameOnceInOrder(t *testing.T) {
	const clients, messages = 12, 150
	var mu sync.Mutex
	seqs := make([][]int64, clients)
	rig := newFanoutRigWith(t, clients, func(idx int, frame []byte) {
		var m domainEntity.Message
		if err := json.Unmarshal(frame, &m); err != nil {
			t.Errorf("client %d got invalid JSON %q: %v", idx, frame, err)
			return
		}
		if m.Type != domainEntity.TypeElementUpdate || m.ProjectID != "proj" || m.PageID != "page" {
			t.Errorf("client %d got the wrong message: %s", idx, frame)
		}
		mu.Lock()
		seqs[idx] = append(seqs[idx], m.Seq)
		mu.Unlock()
	})

	msg := domainEntity.Message{Type: domainEntity.TypeElementUpdate, ProjectID: "proj", PageID: "page", Payload: benchPayload(64)}
	for i := 1; i <= messages; i++ {
		msg.Seq = int64(i)
		rig.hub.Broadcast(rig.roomID, msg, nil)
		if i%50 == 0 {
			rig.waitFor(t, int64(i*clients))
		}
	}
	rig.waitFor(t, int64(messages*clients))

	mu.Lock()
	defer mu.Unlock()
	for idx, got := range seqs {
		if len(got) != messages {
			t.Fatalf("client %d received %d frames, want %d", idx, len(got), messages)
		}
		for i, seq := range got {
			if seq != int64(i+1) {
				t.Fatalf("client %d frame %d has seq %d: out of order or duplicated", idx, i, seq)
			}
		}
	}
}

func TestHubFanoutSkipsTheSender(t *testing.T) {
	rig := newFanoutRig(t, 3)
	room := rig.hub.room(rig.roomID)
	var sender *Client
	room.mu.Lock()
	for c := range room.clients {
		sender = c
		break
	}
	room.mu.Unlock()

	rig.hub.Broadcast(rig.roomID, domainEntity.Message{Type: domainEntity.TypeElementUpdate, ProjectID: "proj", PageID: "page", Seq: 1}, sender)
	rig.waitFor(t, 2)
	// Give a wrongly delivered third frame time to show up.
	rig.waitFor(t, 2)
	if got := rig.received.Load(); got != 2 {
		t.Fatalf("%d clients received the frame, want 2 (everyone but the sender)", got)
	}
}
