package entity

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	domainEntity "my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/utils/logger"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const redisChannelPrefix = "ws:room:"

// broadcastEnqueueTimeout bounds how long localBroadcast waits for a full room queue to drain.
const broadcastEnqueueTimeout = 2 * time.Second

// serverID uniquely identifies this application instance to avoid message loops in Pub/Sub
var serverID = uuid.NewString()

// registerRequest represents a request to add a client to a room.
type registerRequest struct {
	client *Client
	done   chan struct{}
	// rejected is set (before done closes) when the room was shutting down and did not
	// take the client; the caller then joins a fresh room.
	rejected *bool
}

// unregisterRequest represents a request to remove a client from a room.
type unregisterRequest struct {
	detach bool
	client *Client
	done   chan struct{}
}

// redisEnvelope wraps the domain message with metadata for distributed broadcasting
type redisEnvelope struct {
	SourceServerID string               `json:"sourceServerId"`
	Message        domainEntity.Message `json:"message"`
}

// Hub manages multiple rooms and synchronizes state across multiple nodes via Redis.
type Hub struct {
	// rooms maps room ID to *Room. Every broadcast, presence update and lookup reads
	// it from its own goroutine while rooms open and close rarely, which is what
	// sync.Map is for: reads take no lock, whereas an RWMutex reader count is one
	// cache line every core writes (about 30x slower at 16 cores, see lookup_bench_test.go).
	rooms       sync.Map
	redisClient *redis.Client
	parentCtx   context.Context
	wg          *sync.WaitGroup
	sendBuf     int
	logger      logger.Logger
}

// NewHub creates a new Hub. If rdb is provided, it enables horizontal scalability using Redis Pub/Sub.
func NewHub(parentCtx context.Context, wg *sync.WaitGroup, sendBuf int, rdb *redis.Client, lg logger.Logger) *Hub {
	h := &Hub{
		redisClient: rdb,
		parentCtx:   parentCtx,
		wg:          wg,
		sendBuf:     sendBuf,
		logger:      lg,
	}

	if h.redisClient != nil {
		go h.listenRedis()
	}

	return h
}

// room returns the open room with that ID, or nil.
func (h *Hub) room(roomID string) *Room {
	if v, ok := h.rooms.Load(roomID); ok {
		return v.(*Room)
	}
	return nil
}

// getOrCreateRoom returns an existing room or initializes a new one.
func (h *Hub) getOrCreateRoom(roomID string) *Room {
	for {
		if r := h.room(roomID); r != nil {
			if r.ctx.Err() == nil {
				return r
			}
			// The room is shutting down (it emptied, or the hub is stopping) and has not
			// left the table yet: replace it rather than hand out a room that never answers.
			h.rooms.CompareAndDelete(roomID, r)
			if h.parentCtx.Err() != nil {
				return r
			}
			continue
		}

		r := NewRoom(roomID, h.parentCtx, h.wg, h.sendBuf, h.logger)
		r.onEmpty = func(rid string) {
			// Only remove this room: a newer room may already own the ID.
			h.rooms.CompareAndDelete(rid, r)
		}
		if existing, loaded := h.rooms.LoadOrStore(roomID, r); loaded {
			r.cancel() // lost the race; this room never ran
			if er := existing.(*Room); er.ctx.Err() == nil || h.parentCtx.Err() != nil {
				return er
			}
			continue
		}
		go r.run()
		return r
	}
}

// Register adds a client to a room and waits for confirmation. If the room closes
// while the client is joining (its last client left), the client joins a new room.
func (h *Hub) Register(ctx context.Context, client *Client) error {
	if client == nil || client.RoomID() == "" {
		return nil
	}
	roomID := client.RoomID()
	for {
		joined, err := h.tryRegisterInRoom(ctx, h.getOrCreateRoom(roomID), client)
		if err != nil || joined {
			return err
		}
	}
}

func (h *Hub) tryRegisterInRoom(ctx context.Context, r *Room, client *Client) (bool, error) {
	done := make(chan struct{})
	var rejected bool
	select {
	case r.register <- registerRequest{client: client, done: done, rejected: &rejected}:
	case <-r.ctx.Done():
		// The room stopped before taking the request.
		if h.parentCtx.Err() != nil {
			return false, h.parentCtx.Err()
		}
		return false, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}

	select {
	case <-done:
		return !rejected, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

// Unregister removes a client from a room.
func (h *Hub) Unregister(ctx context.Context, client *Client) error {
	return h.unregister(ctx, client, false)
}

// Detach removes a client from its room while keeping its connection open.
func (h *Hub) Detach(ctx context.Context, client *Client) error {
	return h.unregister(ctx, client, true)
}

func (h *Hub) unregister(ctx context.Context, client *Client, detach bool) error {
	if client == nil {
		return nil
	}
	roomID := client.RoomID()
	if roomID == "" {
		return nil
	}

	r := h.room(roomID)
	if r == nil {
		return nil
	}

	done := make(chan struct{})
	select {
	case r.unregister <- unregisterRequest{client: client, done: done, detach: detach}:
		select {
		case <-done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	case <-r.ctx.Done():
		// The room already stopped and closed all its clients, this one included.
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Broadcast sends a message to all clients in a room across the entire cluster.
func (h *Hub) Broadcast(roomID string, msg domainEntity.Message, sender *Client) {
	// 1. Always broadcast to local clients for lowest latency
	h.localBroadcast(roomID, msg, sender)

	// 2. Publish to Redis for other server instances
	if h.redisClient != nil {
		envelope := redisEnvelope{
			SourceServerID: serverID,
			Message:        msg,
		}

		payload, err := json.Marshal(envelope)
		if err != nil {
			h.logger.Errorf("failed to marshal redis envelope: %v", err)
			return
		}

		channel := redisChannelPrefix + roomID
		if err := h.redisClient.Publish(h.parentCtx, channel, payload).Err(); err != nil {
			h.logger.Errorf("failed to publish to redis: %v", err)
		}
	}
}

// BroadcastProject sends a message to all rooms within a project (project-level and per-page rooms).
// It collects all rooms whose ID equals 'projectID' or starts with 'projectID:' and broadcasts to each.
func (h *Hub) BroadcastProject(projectID string, msg domainEntity.Message, sender *Client) {
	var roomIDs []string
	prefix := projectID + ":"
	h.rooms.Range(func(key, _ any) bool {
		if rid := key.(string); rid == projectID || strings.HasPrefix(rid, prefix) {
			roomIDs = append(roomIDs, rid)
		}
		return true
	})

	for _, rid := range roomIDs {
		h.Broadcast(rid, msg, sender)
	}
}

// localBroadcast sends a message to clients connected to this specific server instance.
func (h *Hub) localBroadcast(roomID string, msg domainEntity.Message, sender *Client) {
	r := h.room(roomID)
	if r == nil {
		return
	}

	rb := roomBroadcast{message: msg, sender: sender}
	select {
	case r.broadcast <- rb:
		return
	default:
	}

	// The room queue is full. Wait briefly for the room loop to drain it; if it still cannot
	// accept the message, do not drop it invisibly: log and disconnect the room's clients so
	// they reconnect and resync from the authoritative snapshot.
	timer := time.NewTimer(broadcastEnqueueTimeout)
	defer timer.Stop()
	select {
	case r.broadcast <- rb:
	case <-r.ctx.Done():
	case <-timer.C:
		if h.logger != nil {
			h.logger.Errorf("room %s broadcast queue full for %s, closing room clients to force resync (type=%s seq=%d)", roomID, broadcastEnqueueTimeout, msg.Type, msg.Seq)
		}
		r.closeAllClients()
	}
}

// listenRedis subscribes to Redis and handles messages from other cluster nodes.
func (h *Hub) listenRedis() {
	pubsub := h.redisClient.PSubscribe(h.parentCtx, redisChannelPrefix+"*")
	defer pubsub.Close()

	h.logger.Infof("Hub started Redis subscription (ServerID: %s)", serverID)

	ch := pubsub.Channel()
	for {
		select {
		case <-h.parentCtx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			var envelope redisEnvelope
			if err := json.Unmarshal([]byte(msg.Payload), &envelope); err != nil {
				h.logger.Errorf("failed to unmarshal redis message: %v", err)
				continue
			}

			// Ignore messages published by this instance (already handled by localBroadcast)
			if envelope.SourceServerID == serverID {
				continue
			}

			roomID := msg.Channel[len(redisChannelPrefix):]

			// Broadcast to local rooms. Since the original sender is on a different server,
			// we pass nil for sender to ensure all local clients receive it.
			h.localBroadcast(roomID, envelope.Message, nil)
		}
	}
}

func (h *Hub) GetRoomParticipants(roomID string) []domainEntity.UserPresence {
	if r := h.room(roomID); r != nil {
		return r.GetParticipants()
	}
	return []domainEntity.UserPresence{}
}

// MostRecentClient returns the user's most recently active connection on this instance.
func (h *Hub) MostRecentClient(projectID, userID string) *Client {
	var rooms []*Room
	prefix := projectID + ":"
	h.rooms.Range(func(key, value any) bool {
		if id := key.(string); id == projectID || strings.HasPrefix(id, prefix) {
			rooms = append(rooms, value.(*Room))
		}
		return true
	})

	var best *Client
	for _, r := range rooms {
		for _, c := range r.ClientsOf(userID) {
			if c.Writer {
				continue
			}
			if best == nil || c.LastActiveUnixNano() > best.LastActiveUnixNano() {
				best = c
			}
		}
	}
	return best
}
