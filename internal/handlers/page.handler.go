package handlers

import (
	"errors"
	"my-go-app/internal/authz"
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

var pageAllowedCols = map[string]string{
	"name":          "Name",
	"slug":          "Slug",
	"type":          "Type",
	"styles":        "Styles",
	"contentTypeId": "ContentTypeId",
	"requireAuth":   "RequireAuth",
}

type PageHandler struct {
	pageService services.PageServiceInterface
}

func NewPageHandler(pageService services.PageServiceInterface) *PageHandler {
	return &PageHandler{
		pageService: pageService,
	}
}

func (h *PageHandler) Delete(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid", "pageid")
	if err != nil {
		return err
	}
	projectID := ids[0]
	pageID := ids[1]

	if err := h.pageService.DeleteForUser(c.RequestCtx(), pageID, projectID, userID); err != nil {
		return utils.RespondError(c, err, "Page not found or not owned by user", "Failed to delete page")
	}

	return utils.SendNoContent(c)
}

func (h *PageHandler) ListByProject(c fiber.Ctx) error {
	userID, _ := c.Locals("userId").(string)
	return h.listByProject(c, userID)
}

func (h *PageHandler) ListPublic(c fiber.Ctx) error {
	return h.listByProject(c, "")
}

func (h *PageHandler) listByProject(c fiber.Ctx, userID string) error {
	ids, err := utils.RequireParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	pages, err := h.pageService.ListByProject(c.RequestCtx(), projectID, userID)
	if err != nil {
		return utils.RespondError(c, err, "Project not found", "Failed to retrieve pages")
	}

	return utils.SendJSON(c, fiber.StatusOK, pages)
}

func (h *PageHandler) GetByID(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid", "pageid")
	if err != nil {
		return err
	}
	projectID, pageID := ids[0], ids[1]

	page, err := h.pageService.GetByIDForUser(c.RequestCtx(), projectID, pageID, userID)
	if err != nil {
		return utils.RespondError(c, err, "Page not found", "Failed to retrieve page")
	}

	return utils.SendJSON(c, fiber.StatusOK, page)
}

func (h *PageHandler) Create(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	req := new(dto.CreatePageRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	now := time.Now()
	pageID := req.Id
	if pageID == "" {
		pageID = uuid.New().String()
	}
	pageSlug := ""
	if req.Type != models.PageTypeComponent {
		if req.Slug != nil {
			normalized, err := models.NormalizeExplicitSlugForType(*req.Slug, req.Type)
			if err != nil {
				return utils.SendError(c, fiber.StatusBadRequest, err.Error(), err, userID)
			}
			pageSlug = normalized
		} else {
			pageSlug = models.SlugifyPageName(req.Name)
		}
	}

	var contentTypeID *string
	if req.ContentTypeId != nil && *req.ContentTypeId != "" {
		v := *req.ContentTypeId
		contentTypeID = &v
	}
	requireAuth := req.RequireAuth != nil && *req.RequireAuth && req.Type != models.PageTypeComponent

	page := &models.Page{
		Id:            pageID,
		Name:          req.Name,
		Slug:          pageSlug,
		Type:          req.Type,
		ContentTypeId: contentTypeID,
		RequireAuth:   requireAuth,
		Styles:        req.Styles,
		ProjectId:     projectID,
		AuditFields: models.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	createdPage, err := h.pageService.CreateForUser(c.RequestCtx(), page, userID)
	if errors.Is(err, repositories.ErrProjectUnauthorized) || errors.Is(err, authz.ErrForbidden) {
		return utils.SendError(c, fiber.StatusForbidden, "You cannot add pages to this project", err, userID)
	}
	if errors.Is(err, repositories.ErrPageNameTaken) {
		return utils.SendError(c, fiber.StatusConflict, "A page with this name already exists in the project", err, userID)
	}
	if errors.Is(err, services.ErrInvalidPageSlug) {
		return utils.SendError(c, fiber.StatusBadRequest, err.Error(), err, userID)
	}
	if errors.Is(err, repositories.ErrPageSlugTaken) {
		return utils.SendError(c, fiber.StatusConflict, "A page with this URL path already exists in the project", err, userID)
	}
	if err != nil {
		return utils.RespondError(c, err, "", "Failed to create page")
	}

	return utils.SendJSON(c, fiber.StatusCreated, createdPage)
}

func (h *PageHandler) Update(c fiber.Ctx) error {
	userID, ids, err := utils.RequireUserAndParams(c, "projectid", "pageid")
	if err != nil {
		return err
	}
	projectID, pageID := ids[0], ids[1]

	req := new(dto.UpdatePageRequest)
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	rawBody := map[string]any{}
	utils.SetIfNotNil(rawBody, "name", req.Name)
	if req.Slug != nil {
		slugType := models.PageTypeDynamic
		if req.Type != nil {
			slugType = *req.Type
		}
		normalized, err := models.NormalizeExplicitSlugForType(*req.Slug, slugType)
		if err != nil {
			return utils.SendError(c, fiber.StatusBadRequest, err.Error(), err, userID)
		}
		rawBody["slug"] = normalized
	}
	utils.SetIfNotNil(rawBody, "type", req.Type)
	if req.ContentTypeId != nil {
		if *req.ContentTypeId == "" {
			rawBody["contentTypeId"] = nil
		} else {
			rawBody["contentTypeId"] = *req.ContentTypeId
		}
	}
	if req.Styles != nil {
		rawBody["styles"] = req.Styles
	}
	utils.SetIfNotNil(rawBody, "requireAuth", req.RequireAuth)

	updates, err := utils.BuildColumnUpdates(rawBody, pageAllowedCols)
	if err != nil {
		return err
	}
	if err := utils.RequireUpdates(updates); err != nil {
		return err
	}

	if err := h.pageService.UpdateFieldsForUser(c.RequestCtx(), projectID, pageID, userID, updates); err != nil {
		if errors.Is(err, services.ErrInvalidPageSlug) {
			return utils.SendError(c, fiber.StatusBadRequest, err.Error(), err, userID)
		}
		if errors.Is(err, repositories.ErrPageNameTaken) {
			return utils.SendError(c, fiber.StatusConflict, "A page with this name already exists in the project", err, userID)
		}
		if errors.Is(err, repositories.ErrPageSlugTaken) {
			return utils.SendError(c, fiber.StatusConflict, "A page with this URL path already exists in the project", err, userID)
		}
		return utils.RespondError(c, err, "Page not found", "Failed to update page")
	}

	updated, err := h.pageService.GetByIDForUser(c.RequestCtx(), projectID, pageID, userID)
	if err != nil {
		return utils.RespondError(c, err, "Page not found", "Failed to fetch updated page")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

// fiber:context-methods migrated
