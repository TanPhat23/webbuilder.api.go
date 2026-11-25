package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"my-go-app/pkg/utils"
)

const (
	wsTicketPrefix = "ws:ticket:"
	wsTicketTTL    = 30 * time.Second
	wsTicketBytes  = 32
)

var wsTicketSafeIDRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

// WSTicketHandler mints short-lived, single-use WebSocket authentication tickets.
type WSTicketHandler struct {
	redis *redis.Client
}

// NewWSTicketHandler creates a WSTicketHandler backed by Redis.
func NewWSTicketHandler(redisClient *redis.Client) *WSTicketHandler {
	return &WSTicketHandler{redis: redisClient}
}

type wsTicketPayload struct {
	UserID    string `json:"userId"`
	ProjectID string `json:"projectId"`
}

// MintWSTicket issues a single-use 30-second WebSocket ticket bound to the authenticated
// user and requested project, storing only its SHA-256 digest in Redis.
func (h *WSTicketHandler) MintWSTicket(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]
	if !wsTicketSafeIDRe.MatchString(projectID) {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid project ID", nil, userID)
	}
	if h.redis == nil {
		return utils.SendError(c, fiber.StatusServiceUnavailable, "WebSocket ticket service unavailable", nil, userID)
	}

	var raw [wsTicketBytes]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to generate WebSocket ticket", err, userID)
	}
	ticket := base64.RawURLEncoding.EncodeToString(raw[:])
	sum := sha256.Sum256([]byte(ticket))
	key := wsTicketPrefix + hex.EncodeToString(sum[:])

	payload, err := json.Marshal(wsTicketPayload{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to encode WebSocket ticket", err, userID)
	}

	if err := h.redis.Set(c.RequestCtx(), key, payload, wsTicketTTL).Err(); err != nil {
		return utils.SendError(c, fiber.StatusServiceUnavailable, "WebSocket ticket service unavailable", err, userID)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"ticket":    ticket,
		"expiresIn": int(wsTicketTTL.Seconds()),
	})
}
