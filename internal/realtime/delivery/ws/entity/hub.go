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
}

// unregisterRequest represents a request to remove a client from a room.
type unregisterRequest struct {
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
	mu          sync.RWMutex
	rooms       map[string]*Room
	redisClient *redis.Client
	parentCtx   context.Context
	wg          *sync.WaitGroup
	sendBuf     int
	logger      logger.Logger
}

// NewHub creates a new Hub. If rdb is provided, it enables horizontal scalability using Redis Pub/Sub.
func NewHub(parentCtx context.Context, wg *sync.WaitGroup, sendBuf int, rdb *redis.Client, lg logger.Logger) *Hub {
	h := &Hub{
		rooms:       make(map[string]*Room),
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

// getOrCreateRoom returns an existing room or initializes a new one.
func (h *Hub) getOrCreateRoom(roomID string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	r := h.rooms[roomID]
	if r == nil {
		r = NewRoom(roomID, h.parentCtx, h.wg, h.sendBuf, h.logger)
		r.onEmpty = func(rid string) {
			h.mu.Lock()
			delete(h.rooms, rid)
			h.mu.Unlock()
		}
		h.rooms[roomID] = r
		go r.run()
	}
	return r
}

// Register adds a client to a room and waits for confirmation.
func (h *Hub) Register(ctx context.Context, client *Client) error {
	if client == nil {
		return nil
	}
	roomID := client.RoomID()
	if roomID == "" {
		return nil
	}

	r := h.getOrCreateRoom(roomID)
	done := make(chan struct{})

	select {
	case r.register <- registerRequest{client: client, done: done}:
		select {
		case <-done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Unregister removes a client from a room.
func (h *Hub) Unregister(ctx context.Context, client *Client) error {
	if client == nil {
		return nil
	}
	roomID := client.RoomID()
	if roomID == "" {
		return nil
	}

	h.mu.RLock()
	r := h.rooms[roomID]
	h.mu.RUnlock()

	if r == nil {
		return nil
	}

	done := make(chan struct{})
	select {
	case r.unregister <- unregisterRequest{client: client, done: done}:
		select {
		case <-done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
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
	h.mu.RLock()
	roomIDs := make([]string, 0, len(h.rooms))
	for rid := range h.rooms {
		if rid == projectID || strings.HasPrefix(rid, projectID+":") {
			roomIDs = append(roomIDs, rid)
		}
	}
	h.mu.RUnlock()

	for _, rid := range roomIDs {
		h.Broadcast(rid, msg, sender)
	}
}

// localBroadcast sends a message to clients connected to this specific server instance.
func (h *Hub) localBroadcast(roomID string, msg domainEntity.Message, sender *Client) {
	h.mu.RLock()
	r := h.rooms[roomID]
	h.mu.RUnlock()

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
	h.mu.RLock()
	r := h.rooms[roomID]
	h.mu.RUnlock()

	if r != nil {
		return r.GetParticipants()
	}
	return []domainEntity.UserPresence{}
}
