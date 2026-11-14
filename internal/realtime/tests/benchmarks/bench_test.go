package benchmarks

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	domainEntity "my-go-app/internal/realtime/domain/entity"
)

// MockConn simulates a websocket connection. It counts writes and optionally
// sleeps to simulate network/write latency.
type MockConn struct {
	delay  time.Duration
	writes uint64
}

func (m *MockConn) WriteJSON(v interface{}) error {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	atomic.AddUint64(&m.writes, 1)
	return nil
}

// mutexClient simulates the "direct function call protected by a mutex" approach.
type mutexClient struct {
	conn *MockConn
	mu   sync.Mutex
}

// chanClient simulates the "channel + single writer goroutine" approach.
type chanClient struct {
	conn *MockConn
	send chan domainEntity.Message
}

// helper to create many channel clients and start their writers
func startChanClients(n, buf int, delay time.Duration) ([]*chanClient, *sync.WaitGroup) {
	clients := make([]*chanClient, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		mc := &MockConn{delay: delay}
		ch := make(chan domainEntity.Message, buf)
		clients[i] = &chanClient{
			conn: mc,
			send: ch,
		}
		wg.Add(1)
		go func(c *chanClient) {
			defer wg.Done()
			for msg := range c.send {
				_ = c.conn.WriteJSON(msg)
			}
		}(clients[i])
	}
	return clients, &wg
}

// ---- Single-client benchmarks ----

// BenchmarkMutexWrite_SingleClient measures concurrent writers doing direct
// WriteJSON protected by a mutex on a single connection.
func BenchmarkMutexWrite_SingleClient(b *testing.B) {
	b.ReportAllocs()
	mock := &MockConn{delay: 0}
	client := &mutexClient{conn: mock}
	payload := domainEntity.Message{Type: domainEntity.TypeElementUpdate, Payload: []byte(`{"x":1}`)}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.mu.Lock()
			_ = client.conn.WriteJSON(payload)
			client.mu.Unlock()
		}
	})
}

// BenchmarkChannelWriter_SingleClient measures many producers sending to a
// buffered channel consumed by a single writer goroutine for one connection.
func BenchmarkChannelWriter_SingleClient(b *testing.B) {
	b.ReportAllocs()
	mock := &MockConn{delay: 0}
	send := make(chan domainEntity.Message, 64)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range send {
			_ = mock.WriteJSON(msg)
		}
	}()

	payload := domainEntity.Message{Type: domainEntity.TypeElementUpdate, Payload: []byte(`{"x":1}`)}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			send <- payload
		}
	})
	b.StopTimer()

	close(send)
	wg.Wait()
}

// Variants with artificial write delay (to simulate network)
func BenchmarkMutexWrite_SingleClient_Delay(b *testing.B) {
	b.ReportAllocs()
	mock := &MockConn{delay: 50 * time.Microsecond}
	client := &mutexClient{conn: mock}
	payload := domainEntity.Message{Type: domainEntity.TypeElementUpdate, Payload: []byte(`{"x":1}`)}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client.mu.Lock()
			_ = client.conn.WriteJSON(payload)
			client.mu.Unlock()
		}
	})
}

func BenchmarkChannelWriter_SingleClient_Delay(b *testing.B) {
	b.ReportAllocs()
	mock := &MockConn{delay: 50 * time.Microsecond}
	send := make(chan domainEntity.Message, 8) // small buffer to create backpressure
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range send {
			_ = mock.WriteJSON(msg)
		}
	}()

	payload := domainEntity.Message{Type: domainEntity.TypeElementUpdate, Payload: []byte(`{"x":1}`)}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			send <- payload
		}
	})
	b.StopTimer()

	close(send)
	wg.Wait()
}

// ---- Many-client benchmarks ----

func BenchmarkMutexWrite_ManyClients(b *testing.B) {
	const numClients = 100
	b.ReportAllocs()

	clients := make([]*mutexClient, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = &mutexClient{conn: &MockConn{delay: 0}}
	}

	var counter uint64
	payload := domainEntity.Message{Type: domainEntity.TypeElementUpdate, Payload: []byte(`{"x":1}`)}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx := int(atomic.AddUint64(&counter, 1) % numClients)
			c := clients[idx]
			c.mu.Lock()
			_ = c.conn.WriteJSON(payload)
			c.mu.Unlock()
		}
	})
}

func BenchmarkChannelWriter_ManyClients(b *testing.B) {
	const numClients = 100
	const bufSize = 32
	b.ReportAllocs()

	clients, wg := startChanClients(numClients, bufSize, 0)

	var counter uint64
	payload := domainEntity.Message{Type: domainEntity.TypeElementUpdate, Payload: []byte(`{"x":1}`)}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx := int(atomic.AddUint64(&counter, 1) % numClients)
			clients[idx].send <- payload
		}
	})
	b.StopTimer()

	// cleanup: close channels and wait for writers
	for _, c := range clients {
		close(c.send)
	}
	wg.Wait()
}

// Many clients with write delay: simulates heavier I/O so contention/backpressure appears.
func BenchmarkMutexWrite_ManyClients_Delay(b *testing.B) {
	const numClients = 100
	b.ReportAllocs()

	clients := make([]*mutexClient, numClients)
	for i := 0; i < numClients; i++ {
		clients[i] = &mutexClient{conn: &MockConn{delay: 50 * time.Microsecond}}
	}

	var counter uint64
	payload := domainEntity.Message{Type: domainEntity.TypeElementUpdate, Payload: []byte(`{"x":1}`)}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx := int(atomic.AddUint64(&counter, 1) % numClients)
			c := clients[idx]
			c.mu.Lock()
			_ = c.conn.WriteJSON(payload)
			c.mu.Unlock()
		}
	})
}

func BenchmarkChannelWriter_ManyClients_Delay(b *testing.B) {
	const numClients = 100
	const bufSize = 8
	b.ReportAllocs()

	clients, wg := startChanClients(numClients, bufSize, 50*time.Microsecond)

	var counter uint64
	payload := domainEntity.Message{Type: domainEntity.TypeElementUpdate, Payload: []byte(`{"x":1}`)}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			idx := int(atomic.AddUint64(&counter, 1) % numClients)
			clients[idx].send <- payload
		}
	})
	b.StopTimer()

	for _, c := range clients {
		close(c.send)
	}
	wg.Wait()
}
