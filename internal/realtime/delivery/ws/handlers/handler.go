package handlers

import (
	"context"
	"errors"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/redis/go-redis/v9"

	wsentity "my-go-app/internal/realtime/delivery/ws/entity"
	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
	domainUsecase "my-go-app/internal/realtime/domain/usecase"
	"my-go-app/internal/realtime/utils/logger"
)

type MessageHandler func(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error

type Handler struct {
	uc       *domainUsecase.UseCase
	hub      *wsentity.Hub
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

	// activeConns tracks the currently active client per user+project so that
	// duplicate connections (e.g. rapid reconnects) can be cleaned up.
	activeConnsMu sync.Mutex
	activeConns   map[string]*wsentity.Client
}

// ProjectAccess reports a user's permissions on a project.
type ProjectAccess interface {
	Access(ctx context.Context, projectID, userID string) (read bool, write bool, err error)
}

// RequireAccess makes the handler refuse users with no access to a project and
// reject every mutating message from read-only collaborators.
func (h *Handler) RequireAccess(a ProjectAccess) { h.access = a }

// mutating lists the message types that change project data.
var mutating = map[domainEntity.MessageType]bool{
	domainEntity.TypeElementCreate:         true,
	domainEntity.TypeElementUpdate:         true,
	domainEntity.TypeElementDelete:         true,
	domainEntity.TypeElementMove:           true,
	domainEntity.TypeElementContent:        true,
	domainEntity.TypeCommentCreate:         true,
	domainEntity.TypeCommentUpdate:         true,
	domainEntity.TypeCommentDelete:         true,
	domainEntity.TypeWorkflowCreate:        true,
	domainEntity.TypeWorkflowUpdate:        true,
	domainEntity.TypeWorkflowDelete:        true,
	domainEntity.TypeElementWorkflowCreate: true,
	domainEntity.TypeElementWorkflowUpdate: true,
	domainEntity.TypeElementWorkflowDelete: true,
	domainEntity.TypeLogicUpdate:           true,
	domainEntity.TypeSnapshot:              true,
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
		activeConns: make(map[string]*wsentity.Client),
	}
	h.pages = newPageActors(ctx, pageActorIdle)
	h.hub = wsentity.NewHub(h.ctx, &h.wg, h.sendBuf, rdb, lg)
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
		domainEntity.TypeLogicSync:   h.handleLogicSync,

		// Workflow Operations
		domainEntity.TypeWorkflowCreate:        h.handleWorkflowCreate,
		domainEntity.TypeWorkflowUpdate:        h.handleWorkflowUpdate,
		domainEntity.TypeWorkflowDelete:        h.handleWorkflowDelete,
		domainEntity.TypeElementWorkflowCreate: h.handleElementWorkflowCreate,
		domainEntity.TypeElementWorkflowUpdate: h.handleElementWorkflowUpdate,
		domainEntity.TypeElementWorkflowDelete: h.handleElementWorkflowDelete,
	}
}

// HandleWebSocket is the main entry point for upgraded WebSocket connections.
func (h *Handler) HandleWebSocket(c *websocket.Conn) {
	projectID := c.Params("project")
	user, _ := c.Locals("user").(*domainEntity.User)
	h.logger.Infof("New WebSocket connection: projectID=%s, userID=%s", projectID, user.ID)
	client := wsentity.NewClient(c, projectID, user)

	if projectID != "" && h.access != nil {
		read, write, err := h.access.Access(context.Background(), projectID, user.ID)
		if err != nil {
			h.logger.Errorf("access check failed for user=%s project=%s: %v", user.ID, projectID, err)
			_ = client.WriteJSON(client.Factory().Error(domainErrors.CodeInternal, "could not check access"))
			client.Release()
			return
		}
		if !read {
			// Same reply whether the project is missing or private.
			_ = client.WriteJSON(client.Factory().Error(domainErrors.CodeAccessDenied, "you do not have access to this project"))
			client.Release()
			return
		}
		client.ReadOnly = !write
	}

	if projectID == "" {
		errMsg := client.Factory().Error(domainErrors.CodeValidation, "project ID is required")
		_ = client.WriteJSON(errMsg)
		client.Release()
		return
	}

	ck := connKey(projectID, user.ID)
	h.activeConnsMu.Lock()
	if old, exists := h.activeConns[ck]; exists {
		h.logger.Infof("Closing stale connection for user=%s project=%s", user.ID, projectID)
		_ = h.hub.Unregister(h.ctx, old)
		old.Close()
	}
	h.activeConns[ck] = client
	h.activeConnsMu.Unlock()

	ctx := context.Background()

	defer func() {
		_ = h.hub.Unregister(h.ctx, client)

		h.activeConnsMu.Lock()
		if h.activeConns[ck] == client {
			delete(h.activeConns, ck)
		}
		h.activeConnsMu.Unlock()

		// Must be last: fences off the writer goroutine before the library recycles the conn.
		client.Release()
	}()

	for {
		var msg domainEntity.Message
		if err := c.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Errorf("websocket error: %v", err)
			}
			break
		}
		h.logger.Infof("Received message: type=%s from userID=%s", msg.Type, user.ID)
		h.logger.Debugf("Message details: %+v", msg)

		if err := h.processMessage(ctx, client, msg); err != nil {
			h.logger.Errorf("Error processing message %s: %v", msg.Type, err)

			// Extract structured error code if available
			errCode := domainErrors.CodeProcessError
			// Clients get the structured message only: err.Error() would add the
			// code again and the wrapped cause, which can be a database error.
			errText := err.Error()
			var appErr *domainErrors.AppError
			if errors.As(err, &appErr) {
				errCode = appErr.Code
				errText = appErr.Message
			}

			errMsg := client.Factory().ErrorForOp(errCode, errText, msg.ClientOpID)
			errMsg.RequestID = msg.RequestID
			if sendErr := client.Send(errMsg); sendErr != nil {
				h.logger.Errorf("failed to send error to client: %v", sendErr)
			}
		}
	}
}

// processMessage resolves the handler for a message and executes it.
func (h *Handler) processMessage(ctx context.Context, client *wsentity.Client, msg domainEntity.Message) error {
	// Ensure consistency in the incoming message metadata
	msg.ProjectID = client.ProjectID
	if msg.PageID == "" {
		msg.PageID = client.PageID
	}

	if client.ReadOnly && mutating[msg.Type] {
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
