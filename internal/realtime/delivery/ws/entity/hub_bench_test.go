package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	contribws "github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"

	domainEntity "my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/utils/logger"
)

func quietLogger() logger.Logger { return logger.New(io.Discard, "", 0) }

// fanoutRig is a Hub with n real websocket clients in one room. Each client
// counts the frames it receives, so a benchmark can wait for full delivery.
type fanoutRig struct {
	hub      *Hub
	roomID   string
	received atomic.Int64
	clients  int
	stop     func()
}

func newFanoutRig(tb testing.TB, clients int) *fanoutRig { return newFanoutRigWith(tb, clients, nil) }

// newFanoutRigWith is newFanoutRig that also hands every received frame to onFrame
// (called from the reader goroutine of client idx).
func newFanoutRigWith(tb testing.TB, clients int, onFrame func(idx int, frame []byte)) *fanoutRig {
	tb.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	hub := NewHub(ctx, &wg, 512, nil, quietLogger())
	rig := &fanoutRig{hub: hub, roomID: "proj:page", clients: clients}

	app := fiber.New()
	app.Get("/ws", contribws.New(func(conn *contribws.Conn) {
		c := NewClient(conn, "proj", &domainEntity.User{ID: fmt.Sprintf("u-%p", conn), Email: "u@example.com"})
		c.SetPageID("page")
		if err := hub.Register(context.Background(), c); err != nil {
			return
		}
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
		_ = hub.Unregister(context.Background(), c)
		c.Release()
	}))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		tb.Skipf("cannot listen on loopback: %v", err)
	}
	go func() { _ = app.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true}) }()

	var conns []*websocket.Conn
	var readers sync.WaitGroup
	for i := 0; i < clients; i++ {
		idx := i
		conn, _, err := websocket.DefaultDialer.Dial("ws://"+ln.Addr().String()+"/ws", nil)
		if err != nil {
			tb.Fatalf("dial: %v", err)
		}
		conns = append(conns, conn)
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				_, frame, err := conn.ReadMessage()
				if err != nil {
					return
				}
				if onFrame != nil {
					onFrame(idx, frame)
				}
				rig.received.Add(1)
			}
		}()
	}
	deadline := time.Now().Add(5 * time.Second)
	for len(hub.RoomParticipants(rig.roomID)) < clients {
		if time.Now().After(deadline) {
			tb.Fatalf("only %d of %d clients registered", len(hub.RoomParticipants(rig.roomID)), clients)
		}
		time.Sleep(5 * time.Millisecond)
	}

	rig.stop = func() {
		for _, c := range conns {
			_ = c.Close()
		}
		readers.Wait()
		cancel()
		_ = app.Shutdown()
	}
	tb.Cleanup(rig.stop)
	return rig
}

// waitFor blocks until the rig has received want frames.
func (r *fanoutRig) waitFor(tb testing.TB, want int64) {
	tb.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for r.received.Load() < want {
		if time.Now().After(deadline) {
			tb.Fatalf("delivery stalled: %d of %d frames", r.received.Load(), want)
		}
		time.Sleep(200 * time.Microsecond)
	}
}

func benchPayload(size int) json.RawMessage {
	body := `{"id":"11111111-1111-1111-1111-111111111111","styles":{"color":"red","padding":"` + strings.Repeat("x", size) + `"}}`
	return json.RawMessage(body)
}

// BenchmarkHubFanout broadcasts to N real websocket clients and waits until
// every client has the frame. ns/op is the time to deliver one message to all.
func BenchmarkHubFanout(b *testing.B) {
	for _, clients := range []int{1, 10, 50, 200} {
		b.Run(fmt.Sprintf("clients=%d", clients), func(b *testing.B) {
			rig := newFanoutRig(b, clients)
			msg := domainEntity.Message{Type: domainEntity.TypeElementUpdate, ProjectID: "proj", PageID: "page", Payload: benchPayload(200)}
			const batch = 64 // stay inside each client's send buffer
			b.ReportAllocs()
			b.ResetTimer()
			sent := int64(0)
			for i := 0; i < b.N; i++ {
				msg.Seq = int64(i + 1)
				rig.hub.Broadcast(rig.roomID, msg, nil)
				sent++
				if sent%batch == 0 {
					rig.waitFor(b, sent*int64(clients))
				}
			}
			rig.waitFor(b, sent*int64(clients))
			b.StopTimer()
			b.ReportMetric(float64(sent)*float64(clients)/b.Elapsed().Seconds(), "frames/s")
		})
	}
}

// BenchmarkHubRoomLookup is the read path every broadcast, presence update and
// participant list starts with, hit from many goroutines at once.
func BenchmarkHubRoomLookup(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	hub := NewHub(ctx, &wg, 8, nil, quietLogger())
	const rooms = 500
	for i := 0; i < rooms; i++ {
		hub.getOrCreateRoom(fmt.Sprintf("proj-%d:page", i))
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = hub.RoomParticipants(fmt.Sprintf("proj-%d:page", i%rooms))
			i++
		}
	})
}
