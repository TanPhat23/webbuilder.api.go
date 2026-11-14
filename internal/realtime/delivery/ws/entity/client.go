package entity

import (
	"errors"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"

	domainEntity "my-go-app/internal/realtime/domain/entity"
)

// Client represents a single websocket client connected to a page/room.
// It owns a buffered send channel which is used by a single writer goroutine.
type Client struct {
	ProjectID string
	Conn      *websocket.Conn
	PageID    string
	User      *domainEntity.User
	// ReadOnly clients may read and see presence but cannot change anything.
	ReadOnly bool
	// sendMu guards send/sendClosed so a send can never race with the channel being closed.
	sendMu     sync.RWMutex
	send       chan domainEntity.Message
	sendClosed bool

	// connMu serializes every write to and close of Conn: the websocket library allows only
	// one concurrent writer. Once released is set, Conn belongs to the library again (it is
	// pooled and reused for other connections) and must never be touched.
	connMu   sync.Mutex
	released bool

	// Presence state: stores the latest cursor/selection information for the client.
	// Protected by PresenceMu when accessed concurrently.
	PresenceMu sync.Mutex
	CursorX    float64
	CursorY    float64
	ElementID  *string
}

// NewClient constructs a new client for the given connection, project, and user.
func NewClient(conn *websocket.Conn, projectID string, user *domainEntity.User) *Client {
	return &Client{
		Conn:      conn,
		ProjectID: projectID,
		User:      user,
	}
}

// openSend (re)creates the client's send channel and returns it.
func (c *Client) openSend(buf int) chan domainEntity.Message {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	c.send = make(chan domainEntity.Message, buf)
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

// TrySend attempts to enqueue a message onto the client's send channel.
// Returns false if the channel is not initialized, already closed, or if the client
// is slow (i.e., the channel buffer is full).
func (c *Client) TrySend(msg domainEntity.Message) bool {
	c.sendMu.RLock()
	defer c.sendMu.RUnlock()
	if c.send == nil || c.sendClosed {
		return false
	}
	select {
	case c.send <- msg:
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
	if c.released || c.Conn == nil {
		return ErrConnReleased
	}
	if d, ok := any(c.Conn).(interface{ SetWriteDeadline(time.Time) error }); ok {
		_ = d.SetWriteDeadline(time.Now().Add(writeWait))
	}
	return c.Conn.WriteJSON(msg)
}

// CloseConn closes the connection unless it was already released. Safe for concurrent use.
func (c *Client) CloseConn() {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	if !c.released && c.Conn != nil {
		_ = c.Conn.Close()
	}
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
