package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/redis/go-redis/v9"

	"my-go-app/internal/agentproto"
	wsentity "my-go-app/internal/realtime/delivery/ws/entity"
	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
	domainUsecase "my-go-app/internal/realtime/domain/usecase"
	"my-go-app/internal/realtime/utils/logger"
	"my-go-app/internal/realtime/utils/validation"
)

const (
	maxConnsPerUserProject = 8
	maxWSMessageBytes      = 512 << 10 // 512 KB
	wsPongWait             = 60 * time.Second
	wsPingPeriod           = 25 * time.Second
	wsRateBurst            = 100.0
	wsRateRefillPerSec     = 50.0
)

type MessageHandler func(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error

// connMapKey avoids separator collisions in activeConns map keys.
type connMapKey struct {
	projectID string
	userID    string
}

type Handler struct {
	uc       *domainUsecase.UseCase
	hub      *wsentity.Hub
	agents   *agentBridge
	handlers map[domainEntity.MessageType]MessageHandler
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	sendBuf  int
	logger   logger.Logger
	pages    *pageActors

	// access, when set, decides who may connect and who may only read. Without
	// it every authenticated user may do everything (tests, dev mode).
	access ProjectAccess

	// pageChecker, when set, verifies page ownership within a project on join.
	pageChecker PageChecker

	allowedOrigins    []string
	originsConfigured bool

	// activeConns tracks all active clients (including Writer sockets) per user+project
	// so RBAC revocation and connection limits apply uniformly.
	activeConnsMu sync.Mutex
	activeConns   map[connMapKey]map[*wsentity.Client]struct{}
}

// ProjectAccess reports a user's permissions on a project.
type ProjectAccess interface {
	Access(ctx context.Context, projectID, userID string) (read bool, write bool, err error)
}

// PageChecker verifies whether a page belongs to a project.
type PageChecker interface {
	PageBelongsToProject(ctx context.Context, projectID, pageID string) (bool, error)
}

// RequireAccess makes the handler refuse users with no access to a project and
// reject every mutating message from read-only collaborators.
func (h *Handler) RequireAccess(a ProjectAccess) { h.access = a }

// RequirePageOwnership enables page-to-project ownership checks before joining rooms.
func (h *Handler) RequirePageOwnership(p PageChecker) { h.pageChecker = p }

// SetAllowedOrigins configures the allowed Origin headers for WebSocket upgrades.
func (h *Handler) SetAllowedOrigins(origins []string) {
	h.originsConfigured = true
	if len(origins) == 0 {
		h.allowedOrigins = nil
		return
	}
	h.allowedOrigins = append([]string(nil), origins...)
}

// mutating lists the message types that change project data.
var mutating = map[domainEntity.MessageType]bool{
	domainEntity.TypeElementCreate:  true,
	domainEntity.TypeElementUpdate:  true,
	domainEntity.TypeElementDelete:  true,
	domainEntity.TypeElementMove:    true,
	domainEntity.TypeElementContent: true,
	domainEntity.TypeCommentCreate:  true,
	domainEntity.TypeCommentUpdate:  true,
	domainEntity.TypeCommentDelete:  true,
	domainEntity.TypeLogicUpdate:    true,
	domainEntity.TypeLogicMove:      true,
	domainEntity.TypeSnapshot:       true,
}

// connKey returns a unique key for tracking a user's active connection to a project.
func connKey(projectID, userID string) string {
	return projectID + "::" + userID
}

func NewHandler(uc *domainUsecase.UseCase, rdb *redis.Client, lg logger.Logger) *Handler {
	ctx, cancel := context.WithCancel(context.Background())
	h := &Handler{
		uc:          uc,
		ctx:         ctx,
		cancel:      cancel,
		sendBuf:     64,
		logger:      lg,
		activeConns: make(map[connMapKey]map[*wsentity.Client]struct{}),
	}
	h.pages = newPageActors(ctx, pageActorIdle)
	h.hub = wsentity.NewHub(h.ctx, &h.wg, h.sendBuf, rdb, lg)
	if rdb != nil {
		h.agents = newAgentBridge(hubLocator{h.hub}, agentproto.NewRedisBus(rdb), redisClaimer{rdb}, lg)
		h.wg.Add(2)
		go func() { defer h.wg.Done(); h.agents.Run(h.ctx) }()
		go func() { defer h.wg.Done(); h.listenRBACSync(rdb) }()
	}
	h.registerHandlers()
	return h
}

// registerHandlers maps message types to their respective handling functions (Command Pattern).
func (h *Handler) registerHandlers() {
	h.handlers = map[domainEntity.MessageType]MessageHandler{
		// Presence & Sync
		domainEntity.TypeJoin:        h.handleJoin,
		domainEntity.TypeLeave:       h.handleLeave,
		domainEntity.TypePresence:    h.handlePresence,
		domainEntity.TypeSync:        h.handleSync,
		domainEntity.TypeSyncProject: h.handleSyncProject,

		// Element Operations
		domainEntity.TypeElementCreate: h.handleElementCreate,
		domainEntity.TypeElementUpdate: h.handleElementUpdate,
		domainEntity.TypeElementDelete: h.handleElementDelete,
		domainEntity.TypeElementMove:   h.handleElementMove,

		// Logic (variables and flows)
		domainEntity.TypeLogicUpdate: h.handleLogicUpdate,
		domainEntity.TypeLogicMove:   h.handleLogicMove,
		domainEntity.TypeLogicSync:   h.handleLogicSync,
	}
	if h.agents != nil {
		h.handlers[domainEntity.TypeAgentResponse] = func(ctx context.Context, c *wsentity.Client, msg domainEntity.Message) error {
			return h.agents.handleResponse(ctx, c, msg)
		}
	}
}

// HandleWebSocket is the main entry point for upgraded WebSocket connections.
func (h *Handler) HandleWebSocket(c *websocket.Conn) {
	projectID := c.Params("project")
	user, _ := c.Locals("user").(*domainEntity.User)
	if user == nil || user.ID == "" {
		_ = c.WriteJSON(wsentity.NewMessageFactory(projectID, "").Error(domainErrors.CodeUnauthorized, "unauthorized"))
		_ = c.Close()
		return
	}

	h.logger.Infof("New WebSocket connection: projectID=%s, userID=%s", projectID, user.ID)
	client := wsentity.NewClient(c, projectID, user)
	// A helper socket (?conn=writer) coexists with the editor's own connection.
	client.Writer = c.Query("conn") == "writer"

	if ticketProj, _ := c.Locals("ticketProjectId").(string); ticketProj != "" && ticketProj != projectID {
		_ = client.WriteJSON(client.Factory().Error(domainErrors.CodeAccessDenied, "ticket is not valid for this project"))
		client.Release()
		return
	}

	if !h.authorizeWSClient(client, projectID, user.ID) {
		client.Release()
		return
	}

	key := connMapKey{projectID: projectID, userID: user.ID}
	if !h.registerActiveConn(key, client) {
		_ = client.WriteJSON(client.Factory().Error(domainErrors.CodeAccessDenied, "too many concurrent connections"))
		client.Release()
		return
	}

	defer h.cleanupWSClient(key, client)

	ctx := context.Background()
	h.readWSLoop(ctx, c, client, user.ID)
}

func (h *Handler) authorizeWSClient(client *wsentity.Client, projectID, userID string) bool {
	if projectID == "" {
		_ = client.WriteJSON(client.Factory().Error(domainErrors.CodeValidation, "project ID is required"))
		return false
	}
	if !validation.IsValidSafeID(projectID) {
		_ = client.WriteJSON(client.Factory().Error(domainErrors.CodeValidation, "invalid project ID"))
		return false
	}
	if h.access == nil {
		return true
	}
	read, write, err := h.access.Access(context.Background(), projectID, userID)
	if err != nil {
		h.logger.Errorf("access check failed for user=%s project=%s: %v", userID, projectID, err)
		_ = client.WriteJSON(client.Factory().Error(domainErrors.CodeInternal, "could not check access"))
		return false
	}
	if !read {
		// Same reply whether the project is missing or private.
		_ = client.WriteJSON(client.Factory().Error(domainErrors.CodeAccessDenied, "you do not have access to this project"))
		return false
	}
	client.SetReadOnly(!write)
	return true
}

// RBACSyncChannel is the Redis Pub/Sub channel used to notify realtime nodes
// whenever a user's project role or membership changes via REST.
const RBACSyncChannel = "rbac:access:changed"

type rbacSyncEvent struct {
	ProjectID string `json:"projectId"`
	UserID    string `json:"userId"`
}

func (h *Handler) listenRBACSync(rdb *redis.Client) {
	pubsub := rdb.Subscribe(h.ctx, RBACSyncChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-h.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			var ev rbacSyncEvent
			if err := json.Unmarshal([]byte(msg.Payload), &ev); err != nil {
				continue
			}
			h.SyncUserAccess(ev.ProjectID, ev.UserID)
		}
	}
}

// SyncUserAccess refreshes the active WebSocket permissions for (projectID, userID),
// disconnecting all tracked sockets (primary and writer) immediately if access was revoked
// or toggling ReadOnly when downgraded/upgraded.
func (h *Handler) SyncUserAccess(projectID, userID string) {
	if h.access == nil || projectID == "" || userID == "" {
		return
	}
	key := connMapKey{projectID: projectID, userID: userID}
	h.activeConnsMu.Lock()
	set := h.activeConns[key]
	clients := make([]*wsentity.Client, 0, len(set))
	for c := range set {
		clients = append(clients, c)
	}
	h.activeConnsMu.Unlock()
	if len(clients) == 0 {
		return
	}

	read, write, err := h.access.Access(h.ctx, projectID, userID)
	if err != nil {
		return
	}
	for _, client := range clients {
		if !read {
			_ = client.WriteJSON(client.Factory().Error(domainErrors.CodeAccessDenied, "your access to this project was revoked"))
			client.CloseAsReplaced()
			_ = h.hub.Unregister(h.ctx, client)
			client.CloseSend()
			continue
		}
		client.SetReadOnly(!write)
	}
}

func (h *Handler) registerActiveConn(key connMapKey, client *wsentity.Client) bool {
	var toReplace []*wsentity.Client

	h.activeConnsMu.Lock()
	set := h.activeConns[key]
	if set == nil {
		set = make(map[*wsentity.Client]struct{})
		h.activeConns[key] = set
	}
	if !client.Writer {
		for existing := range set {
			if !existing.Writer {
				toReplace = append(toReplace, existing)
				delete(set, existing)
			}
		}
	}
	if len(set) >= maxConnsPerUserProject {
		if len(set) == 0 {
			delete(h.activeConns, key)
		}
		h.activeConnsMu.Unlock()
		return false
	}
	set[client] = struct{}{}
	h.activeConnsMu.Unlock()

	for _, old := range toReplace {
		h.logger.Infof("Closing stale connection for user=%s project=%s", client.User.ID, client.ProjectID)
		// Send the close frame before Unregister closes old's send channel: its writer
		// would otherwise close the socket first.
		old.CloseAsReplaced()
		_ = h.hub.Unregister(h.ctx, old)
		old.CloseSend()
	}
	return true
}

func (h *Handler) cleanupWSClient(key connMapKey, client *wsentity.Client) {
	_ = h.hub.Unregister(h.ctx, client)
	client.CloseSend()

	h.activeConnsMu.Lock()
	if set := h.activeConns[key]; set != nil {
		delete(set, client)
		if len(set) == 0 {
			delete(h.activeConns, key)
		}
	}
	h.activeConnsMu.Unlock()

	// Must be last: fences off the writer goroutine before the library recycles the conn.
	client.Release()
}

type wsRateLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
}

func newWSRateLimiter() *wsRateLimiter {
	return &wsRateLimiter{
		tokens:     wsRateBurst,
		maxTokens:  wsRateBurst,
		refillRate: wsRateRefillPerSec,
		lastRefill: time.Now(),
	}
}

func (rl *wsRateLimiter) allow() bool {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	if elapsed > 0 {
		rl.tokens += elapsed * rl.refillRate
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}
	if rl.tokens < 1.0 {
		return false
	}
	rl.tokens -= 1.0
	return true
}

func (h *Handler) readWSLoop(ctx context.Context, c *websocket.Conn, client *wsentity.Client, userID string) {
	c.SetReadLimit(maxWSMessageBytes)
	_ = c.SetReadDeadline(time.Now().Add(wsPongWait))
	c.SetPongHandler(func(string) error {
		return c.SetReadDeadline(time.Now().Add(wsPongWait))
	})

	stopPing := make(chan struct{})
	defer close(stopPing)
	go func() {
		ticker := time.NewTicker(wsPingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-stopPing:
				return
			case <-h.ctx.Done():
				return
			case <-ticker.C:
				if err := client.WritePing(); err != nil {
					return
				}
			}
		}
	}()

	limiter := newWSRateLimiter()
	for {
		var msg domainEntity.Message
		if err := c.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Errorf("websocket error: %v", err)
			}
			return
		}
		_ = c.SetReadDeadline(time.Now().Add(wsPongWait))
		h.logger.Infof("Received message: type=%s from userID=%s", msg.Type, userID)
		h.logger.Debugf("Message details: %+v", msg)

		if !limiter.allow() {
			errMsg := client.Factory().ErrorForOp(domainErrors.CodeValidation, "rate limit exceeded", msg.ClientOpID)
			errMsg.RequestID = msg.RequestID
			_ = client.Send(errMsg)
			continue
		}

		if err := h.processMessage(ctx, client, msg); err != nil {
			h.sendProcessError(client, msg, err)
		}
	}
}

func (h *Handler) sendProcessError(client *wsentity.Client, msg domainEntity.Message, err error) {
	h.logger.Errorf("Error processing message %s: %v", msg.Type, err)
	errCode := domainErrors.CodeProcessError
	// Clients never receive raw non-AppError strings (e.g. database/driver internals).
	errText := "internal error"
	var appErr *domainErrors.AppError
	switch {
	case errors.As(err, &appErr):
		errCode = appErr.Code
		errText = appErr.Message
	case errors.Is(err, domainErrors.ErrNotJoined):
		errText = domainErrors.ErrNotJoined.Error()
	}
	errMsg := client.Factory().ErrorForOp(errCode, errText, msg.ClientOpID)
	errMsg.RequestID = msg.RequestID
	if sendErr := client.Send(errMsg); sendErr != nil {
		h.logger.Errorf("failed to send error to client: %v", sendErr)
	}
}

// processMessage resolves the handler for a message and executes it.
func (h *Handler) processMessage(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	// Ensure consistency in the incoming message metadata
	msg.ProjectID = client.ProjectID
	if msg.PageID == "" {
		msg.PageID = client.GetPageID()
	}

	client.Touch()

	if client.IsReadOnly() && mutating[msg.Type] {
		return domainErrors.New(domainErrors.CodeAccessDenied, "you have read-only access to this project", nil)
	}

	handler, ok := h.handlers[msg.Type]
	if !ok {
		return domainErrors.New(domainErrors.CodeValidation, "unsupported message type", nil)
	}

	return handler(ctx, client, msg)
}

// Shutdown gracefully stops the handler and waits for client writer goroutines to exit.
func (h *Handler) Shutdown() {
	h.cancel()
	h.wg.Wait()
}
