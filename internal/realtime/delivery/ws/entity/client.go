package entity

import (
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/contrib/v3/websocket"

	domainEntity "my-go-app/internal/realtime/domain/entity"
)

// Client represents a single websocket client connected to a page/room.
// It owns a buffered send channel which is used by a single writer goroutine.
type Client struct {
	lastActive atomic.Int64
	ProjectID  string
	Conn       *websocket.Conn
	PageID     string
	User       *domainEntity.User
	// ReadOnly clients may read and see presence but cannot change anything.
	ReadOnly bool
	// Writer marks a short-lived helper socket (e.g. writing a component master
	// while the editor stays open). It never replaces the editor's connection and
	// is never chosen as the editor tab for agent calls.
	Writer bool
	// sendMu guards send/sendClosed so a send can never race with the channel being closed.
	sendMu sync.RWMutex
	// send holds frames that are already JSON encoded, so a broadcast to a whole
	// room is marshalled once instead of once per client.
	send       chan []byte
	sendClosed bool

	// connMu serializes every write to and close of Conn: the websocket library allows only
	// one concurrent writer. Once released is set, Conn belongs to the library again (it is
	// pooled and reused for other connections) and must never be touched.
	connMu   sync.Mutex
	released bool
	// closing is set once a close frame was sent: CloseConn then leaves the connection
	// open so the peer can read that frame, and closeGrace closes it if the peer never answers.
	closing bool

	// Presence state: stores the latest cursor/selection information for the client.
	// Protected by PresenceMu when accessed concurrently.
	PresenceMu sync.Mutex
	CursorX    float64
	CursorY    float64
	ElementID  *string
}

// NewClient constructs a new client for the given connection, project, and user.
func NewClient(conn *websocket.Conn, projectID string, user *domainEntity.User) *Client {
	c := &Client{
		Conn:      conn,
		ProjectID: projectID,
		User:      user,
	}
	c.Touch()
	return c
}

// Touch marks the client as active now. Any message from the tab counts.
func (c *Client) Touch() { c.lastActive.Store(time.Now().UnixNano()) }

func (c *Client) LastActiveUnixNano() int64 { return c.lastActive.Load() }

// openSend (re)creates the client's send channel and returns it.
func (c *Client) openSend(buf int) chan []byte {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	c.send = make(chan []byte, buf)
	c.sendClosed = false
	return c.send
}

// CloseSend closes the client's send channel exactly once.
func (c *Client) CloseSend() {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if c.send != nil && !c.sendClosed {
		c.sendClosed = true
		close(c.send)
	}
}

// EncodeFrame marshals a message into the frame written to a websocket client.
func EncodeFrame(msg domainEntity.Message) ([]byte, error) { return json.Marshal(msg) }

// TrySend encodes msg and attempts to enqueue it. See TrySendFrame.
func (c *Client) TrySend(msg domainEntity.Message) bool {
	frame, err := EncodeFrame(msg)
	if err != nil {
		return false
	}
	return c.TrySendFrame(frame)
}

// TrySendFrame attempts to enqueue an encoded frame onto the client's send channel.
// Returns false if the channel is not initialized, already closed, or if the client
// is slow (i.e., the channel buffer is full). The frame must not be modified afterwards:
// a broadcast shares one frame between all its recipients.
func (c *Client) TrySendFrame(frame []byte) bool {
	c.sendMu.RLock()
	defer c.sendMu.RUnlock()
	if c.send == nil || c.sendClosed {
		return false
	}
	select {
	case c.send <- frame:
		return true
	default:
		return false
	}
}

// Send attempts to enqueue the message to the client's send channel, falling
// back to a direct WriteJSON if the client is slow.
func (c *Client) Send(msg domainEntity.Message) error {
	if c.TrySend(msg) {
		return nil
	}
	return c.WriteJSON(msg)
}

// ErrConnReleased is returned by writes after the connection handler has returned.
var ErrConnReleased = errors.New("websocket connection already released")

const writeWait = 10 * time.Second

// WriteJSON writes one message to the connection. Safe for concurrent use.
func (c *Client) WriteJSON(msg any) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.released || c.closing || c.Conn == nil {
		return ErrConnReleased
	}
	if d, ok := any(c.Conn).(interface{ SetWriteDeadline(time.Time) error }); ok {
		_ = d.SetWriteDeadline(time.Now().Add(writeWait))
	}
	return c.Conn.WriteJSON(msg)
}

// WriteFrame writes one already encoded JSON frame as a text message. Safe for concurrent use.
func (c *Client) WriteFrame(frame []byte) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if c.released || c.closing || c.Conn == nil {
		return ErrConnReleased
	}
	if d, ok := any(c.Conn).(interface{ SetWriteDeadline(time.Time) error }); ok {
		_ = d.SetWriteDeadline(time.Now().Add(writeWait))
	}
	return c.Conn.WriteMessage(websocket.TextMessage, frame)
}

// CloseConn closes the connection unless it was already released. After a close frame
// (see CloseAsReplaced) it leaves the connection to the close handshake. Safe for concurrent use.
func (c *Client) CloseConn() {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if !c.released && !c.closing && c.Conn != nil {
		_ = c.Conn.Close()
	}
}

// CloseReplaced is the close code sent to a connection that a newer connection
// from the same user and project replaced. Clients must not reconnect after it.
const CloseReplaced = 4409

// closeGrace is how long a peer has to answer a close frame before the connection is closed.
const closeGrace = time.Second

// CloseAsReplaced sends the peer a CloseReplaced close frame and stops sending to it.
// The connection ends when the peer answers the close frame (the read loop then returns
// and the handler releases it), or after closeGrace. Closing the socket right after the
// frame can reset the connection before the peer reads it. Safe for concurrent use.
func (c *Client) CloseAsReplaced() {
	c.connMu.Lock()
	if c.released || c.closing || c.Conn == nil {
		c.connMu.Unlock()
		c.CloseSend()
		return
	}
	c.closing = true
	msg := websocket.FormatCloseMessage(CloseReplaced, "replaced by a newer connection")
	if err := c.Conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(closeGrace)); err != nil {
		_ = c.Conn.Close()
	}
	c.connMu.Unlock()
	c.CloseSend()
	time.AfterFunc(closeGrace, func() {
		c.connMu.Lock()
		defer c.connMu.Unlock()
		if !c.released && c.Conn != nil {
			_ = c.Conn.Close()
		}
	})
}

// Release closes the connection and fences off all further use of Conn. The connection
// handler must call it before returning: after that the library recycles the Conn object.
// It waits for any in-flight write to finish.
func (c *Client) Release() {
	c.CloseSend()
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if !c.released && c.Conn != nil {
		_ = c.Conn.Close()
	}
	c.released = true
}

// SetPageID sets the client's current page ID.
func (c *Client) SetPageID(pageID string) {
	c.PageID = pageID
}

// RoomID returns a composite room identifier to ensure uniqueness across projects.
func (c *Client) RoomID() string {
	if c.ProjectID == "" {
		return c.PageID
	}
	if c.PageID == "" {
		return c.ProjectID
	}
	return c.ProjectID + ":" + c.PageID
}

// Close gracefully closes the client connection and channel.
func (c *Client) Close() {
	c.CloseSend()
	c.CloseConn()
}

func (c *Client) Factory() *MessageFactory {
	if c.User != nil && c.User.ID != "" {
		return NewMessageFactory(c.ProjectID, c.PageID, c.User.ID)
	}
	return NewMessageFactory(c.ProjectID, c.PageID)
}

func (c *Client) SetPresence(p domainEntity.PayloadPresence) {
	c.PresenceMu.Lock()
	defer c.PresenceMu.Unlock()
	c.CursorX = p.CursorX
	c.CursorY = p.CursorY
	if p.ElementID == "" {
		c.ElementID = nil
	} else {
		e := p.ElementID
		c.ElementID = &e
	}
}

func (c *Client) GetPresence() domainEntity.Presence {
	c.PresenceMu.Lock()
	defer c.PresenceMu.Unlock()
	return domainEntity.Presence{
		CursorX:   c.CursorX,
		CursorY:   c.CursorY,
		ElementID: c.ElementID,
	}
}
