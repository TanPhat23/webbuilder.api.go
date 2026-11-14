package entity

import (
	"context"
	"sync"

	domainEntity "my-go-app/internal/realtime/domain/entity"
	"my-go-app/internal/realtime/utils/logger"
)

// roomBroadcastBuffer is the capacity of a room's inbound broadcast queue.
const roomBroadcastBuffer = 1024

// roomBroadcast encapsulates a message targeted at a room and an optional
// sender which should not receive the message.
type roomBroadcast struct {
	message domainEntity.Message
	sender  *Client
}

// Room manages clients for a single page/project.
type Room struct {
	PageID     string
	clients    map[*Client]bool
	register   chan registerRequest
	unregister chan unregisterRequest
	broadcast  chan roomBroadcast
	ctx        context.Context
	cancel     context.CancelFunc
	wg         *sync.WaitGroup
	sendBuf    int
	logger     logger.Logger
	onEmpty    func(string) // optional callback called when room becomes empty
	mu         sync.Mutex
}

// NewRoom creates a room bound to the parentCtx. The room will stop when the
// parentCtx is canceled or when it empties itself (no clients remain).
func NewRoom(roomID string, parentCtx context.Context, wg *sync.WaitGroup, sendBuf int, lg logger.Logger) *Room {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Room{
		PageID:     roomID,
		clients:    make(map[*Client]bool),
		register:   make(chan registerRequest),
		unregister: make(chan unregisterRequest),
		broadcast:  make(chan roomBroadcast, roomBroadcastBuffer),
		ctx:        ctx,
		cancel:     cancel,
		wg:         wg,
		sendBuf:    sendBuf,
		logger:     lg,
	}
}

// run starts the room event loop. It manages client registration, unregistration
// and message broadcasting.
func (r *Room) run() {
	// Ensure cleanup of clients' send channels when the room stops
	defer func() {
		r.mu.Lock()
		for client := range r.clients {
			delete(r.clients, client)
			client.CloseSend()
		}
		r.mu.Unlock()
	}()

	for {
		select {
		case <-r.ctx.Done():
			return

		case req := <-r.register:
			client := req.client
			r.mu.Lock()
			if _, ok := r.clients[client]; !ok {
				send := client.openSend(r.sendBuf)
				r.clients[client] = true
				r.wg.Add(1)
				go r.clientWriter(client, send)
			}
			r.mu.Unlock()

			if req.done != nil {
				// close done asynchronously so the room loop isn't blocked by the
				// waiter's behavior
				go func(ch chan struct{}) { close(ch) }(req.done)
			}

		case req := <-r.unregister:
			client := req.client
			r.mu.Lock()
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				client.CloseSend()
			}
			empty := len(r.clients) == 0
			r.mu.Unlock()

			if req.done != nil {
				go func(ch chan struct{}) { close(ch) }(req.done)
			}

			if empty {
				// stop the room and tell the hub it is empty so it can be removed
				r.cancel()
				if r.onEmpty != nil {
					r.onEmpty(r.PageID)
				}
			}

		case msg := <-r.broadcast:
			r.mu.Lock()
			for client := range r.clients {
				if msg.sender != nil && client == msg.sender {
					continue
				}
				if !client.TrySend(msg.message) {
					// Slow client: never drop a single message silently (clients rely on
					// gap-free seq numbers). Close its send channel, which makes the writer
					// close the connection, so the client reconnects and resyncs.
					if r.logger != nil {
						r.logger.Errorf("room %s: client send buffer full (user=%s), closing connection to force resync", r.PageID, clientUserID(client))
					}
					client.CloseSend()
					delete(r.clients, client)
				}
			}
			r.mu.Unlock()
		}
	}
}

// clientWriter is responsible for writing messages from a client's send
// channel to the WebSocket connection. It uses the room's context to stop.
func (r *Room) clientWriter(c *Client, send <-chan domainEntity.Message) {
	defer r.wg.Done()

	for {
		select {
		case <-r.ctx.Done():
			c.CloseConn()
			return

		case msg, ok := <-send:
			if !ok {
				c.CloseConn()
				return
			}

			if err := c.WriteJSON(msg); err != nil {
				if r.logger != nil {
					r.logger.Errorf("websocket write error (page=%s): %v", c.PageID, err)
				}
				c.CloseConn()
				return
			}
		}
	}
}

// closeAllClients closes every client's send channel (and therefore their connections)
// and removes them from the room. Used when the room can no longer guarantee delivery.
func (r *Room) closeAllClients() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for client := range r.clients {
		client.CloseSend()
		delete(r.clients, client)
	}
}

func clientUserID(c *Client) string {
	if c != nil && c.User != nil {
		return c.User.ID
	}
	return ""
}

func (r *Room) GetParticipants() []domainEntity.UserPresence {
	r.mu.Lock()
	defer r.mu.Unlock()

	users := make([]domainEntity.UserPresence, 0, len(r.clients))
	for client := range r.clients {
		if client.User != nil {
			name := ""
			if client.User.FirstName != nil {
				name = *client.User.FirstName
			}
			if client.User.LastName != nil {
				if name != "" {
					name += " "
				}
				name += *client.User.LastName
			}
			if name == "" {
				name = client.User.Email
			}

			pres := client.GetPresence()
			users = append(users, domainEntity.UserPresence{
				UserID:   client.User.ID,
				UserName: name,
				Presence: pres,
			})
		}
	}
	return users
}
