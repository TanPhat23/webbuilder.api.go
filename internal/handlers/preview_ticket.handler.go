package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
)

const (
	previewTicketPrefix     = "preview:ticket:"
	previewSessionPrefix    = "preview:session:"
	previewTicketTTL        = 60 * time.Second
	previewSessionTTL       = 15 * time.Minute
	previewTokenBytes       = 32
	previewSessionLocalsKey = "previewSession"
	previewMaxCMSLimit      = 100
)

type previewSessionPayload struct {
	UserID    string `json:"userId"`
	ProjectID string `json:"projectId"`
}

type previewExchangeRequest struct {
	Ticket string `json:"ticket"`
}

// PreviewTicketHandler mints short-lived single-use preview tickets, exchanges
// them for project-scoped read-only preview sessions, and serves read-only draft
// preview reads without exposing Clerk credentials to the site runtime origin.
type PreviewTicketHandler struct {
	redis           *redis.Client
	internalSecret  string
	projectService  services.ProjectServiceInterface
	elementService  services.ElementServiceInterface
	contentTypeRepo repositories.ContentTypeRepositoryInterface
	contentItemRepo repositories.ContentItemRepositoryInterface
}

// NewPreviewTicketHandler creates a PreviewTicketHandler backed by Redis and
// read-only project/element/CMS services.
func NewPreviewTicketHandler(
	redisClient *redis.Client,
	projectService services.ProjectServiceInterface,
	elementService services.ElementServiceInterface,
	contentTypeRepo repositories.ContentTypeRepositoryInterface,
	contentItemRepo repositories.ContentItemRepositoryInterface,
) *PreviewTicketHandler {
	return &PreviewTicketHandler{
		redis:           redisClient,
		internalSecret:  strings.TrimSpace(os.Getenv("PREVIEW_INTERNAL_SECRET")),
		projectService:  projectService,
		elementService:  elementService,
		contentTypeRepo: contentTypeRepo,
		contentItemRepo: contentItemRepo,
	}
}

// WithInternalSecret overrides the shared internal exchange secret (primarily for tests).
func (h *PreviewTicketHandler) WithInternalSecret(secret string) *PreviewTicketHandler {
	h.internalSecret = strings.TrimSpace(secret)
	return h
}

func generateRandomToken() (string, string, error) {
	var raw [previewTokenBytes]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw[:])
	sum := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(sum[:]), nil
}

func hashTokenHex(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Mint issues a single-use 60-second preview ticket bound to the authenticated
// user and requested project, storing only its SHA-256 digest in Redis.
func (h *PreviewTicketHandler) Mint(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]
	if !wsTicketSafeIDRe.MatchString(projectID) {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid project ID", nil, userID)
	}
	if h.redis == nil {
		return utils.SendError(c, fiber.StatusServiceUnavailable, "Preview ticket service unavailable", nil, userID)
	}

	ticket, digestHex, err := generateRandomToken()
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to generate preview ticket", err, userID)
	}

	payload, err := json.Marshal(previewSessionPayload{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to encode preview ticket", err, userID)
	}

	key := previewTicketPrefix + digestHex
	if err := h.redis.Set(c.RequestCtx(), key, payload, previewTicketTTL).Err(); err != nil {
		return utils.SendError(c, fiber.StatusServiceUnavailable, "Preview ticket service unavailable", err, userID)
	}

	c.Set("Cache-Control", "private, no-store")
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"ticket":    ticket,
		"expiresIn": int(previewTicketTTL.Seconds()),
	})
}

// Exchange atomically consumes a single-use preview ticket (GETDEL) and mints
// a 15-minute read-only preview session token for the Site Runtime server.
func (h *PreviewTicketHandler) Exchange(c fiber.Ctx) error {
	c.Set("Cache-Control", "private, no-store")

	if h.internalSecret == "" {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Preview exchange is not configured")
	}
	provided := strings.TrimSpace(c.Get("X-Preview-Internal-Secret"))
	if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(h.internalSecret)) != 1 {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid internal preview secret")
	}
	if h.redis == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Preview ticket service unavailable")
	}

	var req previewExchangeRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid preview exchange request")
	}
	ticket := strings.TrimSpace(req.Ticket)
	if ticket == "" || len(ticket) > 256 {
		return fiber.NewError(fiber.StatusBadRequest, "ticket is required")
	}

	ticketKey := previewTicketPrefix + hashTokenHex(ticket)
	stored, err := h.redis.GetDel(c.RequestCtx(), ticketKey).Result()
	if errors.Is(err, redis.Nil) {
		return fiber.NewError(fiber.StatusUnauthorized, "Preview ticket is invalid or expired")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Preview ticket service unavailable")
	}

	var payload previewSessionPayload
	if err := json.Unmarshal([]byte(stored), &payload); err != nil || payload.UserID == "" || payload.ProjectID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Preview ticket payload is invalid")
	}

	sessionToken, sessionDigestHex, err := generateRandomToken()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate preview session")
	}

	sessionKey := previewSessionPrefix + sessionDigestHex
	if err := h.redis.Set(c.RequestCtx(), sessionKey, stored, previewSessionTTL).Err(); err != nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Preview session service unavailable")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token":     sessionToken,
		"projectId": payload.ProjectID,
		"expiresIn": int(previewSessionTTL.Seconds()),
	})
}

// RequirePreviewToken validates the X-Preview-Token header against Redis and
// verifies it is scoped to the :projectid route parameter.
func (h *PreviewTicketHandler) RequirePreviewToken(c fiber.Ctx) error {
	c.Set("Cache-Control", "private, no-store")

	projectID := strings.TrimSpace(c.Params("projectid"))
	if !wsTicketSafeIDRe.MatchString(projectID) {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid project ID")
	}
	if h.redis == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Preview session service unavailable")
	}

	rawToken := strings.TrimSpace(c.Get("X-Preview-Token"))
	if rawToken == "" || len(rawToken) > 256 {
		return fiber.NewError(fiber.StatusUnauthorized, "Preview token is required")
	}

	sessionKey := previewSessionPrefix + hashTokenHex(rawToken)
	stored, err := h.redis.Get(c.RequestCtx(), sessionKey).Result()
	if errors.Is(err, redis.Nil) {
		return fiber.NewError(fiber.StatusUnauthorized, "Preview session is invalid or expired")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Preview session service unavailable")
	}

	var payload previewSessionPayload
	if err := json.Unmarshal([]byte(stored), &payload); err != nil || payload.UserID == "" || payload.ProjectID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Preview session payload is invalid")
	}
	if payload.ProjectID != projectID {
		return fiber.NewError(fiber.StatusForbidden, "Preview token is not valid for this project")
	}

	c.Locals(previewSessionLocalsKey, payload)
	return c.Next()
}

func previewSessionFromCtx(c fiber.Ctx) (previewSessionPayload, error) {
	session, ok := c.Locals(previewSessionLocalsKey).(previewSessionPayload)
	if !ok || session.UserID == "" || session.ProjectID == "" {
		return previewSessionPayload{}, fiber.NewError(fiber.StatusUnauthorized, "Preview session is required")
	}
	return session, nil
}

// GetProject returns read-only draft project metadata for an active preview session.
func (h *PreviewTicketHandler) GetProject(c fiber.Ctx) error {
	session, err := previewSessionFromCtx(c)
	if err != nil {
		return err
	}
	project, err := h.projectService.GetWithAccess(c.RequestCtx(), session.ProjectID, session.UserID)
	if err != nil {
		return respondProjectError(c, err, "Failed to retrieve project")
	}
	return utils.SendJSON(c, fiber.StatusOK, project)
}

// ListPages returns all project pages (including logic and component pages) for preview.
func (h *PreviewTicketHandler) ListPages(c fiber.Ctx) error {
	session, err := previewSessionFromCtx(c)
	if err != nil {
		return err
	}
	pages, err := h.projectService.ListPages(c.RequestCtx(), session.ProjectID, session.UserID)
	if err != nil {
		return respondProjectError(c, err, "Failed to retrieve project pages")
	}
	return utils.SendJSON(c, fiber.StatusOK, pages)
}

// ListElements returns draft project elements after verifying live project access.
func (h *PreviewTicketHandler) ListElements(c fiber.Ctx) error {
	session, err := previewSessionFromCtx(c)
	if err != nil {
		return err
	}
	if _, err := h.projectService.GetWithAccess(c.RequestCtx(), session.ProjectID, session.UserID); err != nil {
		return respondProjectError(c, err, "Failed to verify project access")
	}
	elements, err := h.elementService.List(c.RequestCtx(), session.ProjectID)
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to retrieve elements")
	}
	return utils.SendJSON(c, fiber.StatusOK, elements)
}

func (h *PreviewTicketHandler) verifyContentTypeInProject(c fiber.Ctx, session previewSessionPayload, contentTypeID string) error {
	if _, err := h.projectService.GetWithAccess(c.RequestCtx(), session.ProjectID, session.UserID); err != nil {
		return respondProjectError(c, err, "Failed to verify project access")
	}
	if h.contentTypeRepo == nil || h.contentItemRepo == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "CMS preview service unavailable")
	}
	ct, err := h.contentTypeRepo.GetByID(c.RequestCtx(), contentTypeID)
	if err != nil {
		return utils.RespondError(c, err, "Content type not found", "Failed to retrieve content type")
	}
	if ct == nil || ct.ProjectId == nil || *ct.ProjectId != session.ProjectID {
		return fiber.NewError(fiber.StatusNotFound, "Content type not found")
	}
	return nil
}

// ListCMSItems returns flattened CMS items (including drafts) belonging to the previewed project.
func (h *PreviewTicketHandler) ListCMSItems(c fiber.Ctx) error {
	session, err := previewSessionFromCtx(c)
	if err != nil {
		return err
	}
	contentTypeID, err := utils.RequireParam(c, "contentTypeId")
	if err != nil {
		return err
	}
	if err := h.verifyContentTypeInProject(c, session, contentTypeID); err != nil {
		return err
	}

	items, err := h.contentItemRepo.ListByContentType(c.RequestCtx(), contentTypeID)
	if err != nil {
		return utils.RespondError(c, err, "Content items not found", "Failed to retrieve content items")
	}

	sortBy := c.Query("sortBy", "createdAt")
	sortOrder := strings.ToLower(c.Query("sortOrder", "desc"))
	sortPreviewContentItems(items, sortBy, sortOrder)

	limit := utils.ParsePositiveIntQuery(c, "limit", 10)
	if limit > previewMaxCMSLimit {
		limit = previewMaxCMSLimit
	}
	if len(items) > limit {
		items = items[:limit]
	}

	return utils.SendJSON(c, fiber.StatusOK, flattenPreviewContentItems(items))
}

// GetCMSItemBySlug returns a flattened CMS item (including drafts) for a dynamic preview page.
func (h *PreviewTicketHandler) GetCMSItemBySlug(c fiber.Ctx) error {
	session, err := previewSessionFromCtx(c)
	if err != nil {
		return err
	}
	contentTypeID, err := utils.RequireParam(c, "contentTypeId")
	if err != nil {
		return err
	}
	rawSlug := c.Params("*")
	if decoded, err := url.PathUnescape(rawSlug); err == nil {
		rawSlug = decoded
	}
	slug := strings.Trim(strings.TrimSpace(rawSlug), "/")
	if slug == "" {
		return fiber.NewError(fiber.StatusBadRequest, "slug is required")
	}

	if err := h.verifyContentTypeInProject(c, session, contentTypeID); err != nil {
		return err
	}

	item, err := h.contentItemRepo.GetBySlug(c.RequestCtx(), contentTypeID, slug)
	if err != nil {
		return utils.RespondError(c, err, "Content item not found", "Failed to retrieve content item")
	}

	return utils.SendJSON(c, fiber.StatusOK, flattenPreviewContentItem(item))
}

func sortPreviewContentItems(items []models.ContentItem, sortBy, sortOrder string) {
	asc := sortOrder == "asc"
	sort.SliceStable(items, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "title":
			less = strings.ToLower(items[i].Title) < strings.ToLower(items[j].Title)
		case "updatedAt":
			less = items[i].UpdatedAt.Before(items[j].UpdatedAt)
		default:
			less = items[i].CreatedAt.Before(items[j].CreatedAt)
		}
		if asc {
			return less
		}
		return !less
	})
}

func flattenPreviewContentItem(item *models.ContentItem) map[string]any {
	flattened := map[string]any{
		"contentTypeId": item.ContentTypeId,
		"createdAt":     item.CreatedAt,
		"id":            item.Id,
		"published":     item.Published,
		"slug":          item.Slug,
		"title":         item.Title,
		"updatedAt":     item.UpdatedAt,
		"contentType":   item.ContentType,
	}
	for _, fv := range item.FieldValues {
		if fv.Field != nil && fv.Field.Name != "" {
			flattened[fv.Field.Name] = fv.Value
		}
	}
	return flattened
}

func flattenPreviewContentItems(items []models.ContentItem) []map[string]any {
	flattened := make([]map[string]any, 0, len(items))
	for i := range items {
		flattened = append(flattened, flattenPreviewContentItem(&items[i]))
	}
	return flattened
}
