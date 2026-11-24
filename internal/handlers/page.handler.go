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
	"name":   "Name",
	"type":   "Type",
	"styles": "Styles",
}

type PageHandler struct {
	pageService services.PageServiceInterface
}

func NewPageHandler(pageService services.PageServiceInterface) *PageHandler {
	return &PageHandler{
		pageService: pageService,
	}
}

func (h *PageHandler) DeletePage(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid", "pageid")
	if err != nil {
		return err
	}
	projectID := ids[0]
	pageID := ids[1]

	if err := h.pageService.DeletePageByProjectID(c.RequestCtx(), pageID, projectID, userID); err != nil {
		return utils.HandleRepoError(c, err, "Page not found or not owned by user", "Failed to delete page")
	}

	return utils.SendNoContent(c)
}

func (h *PageHandler) GetPagesByProjectID(c fiber.Ctx) error {
	userID, _ := c.Locals("userId").(string)
	return h.getPagesByProjectID(c, userID)
}

func (h *PageHandler) GetPublicPagesByProjectID(c fiber.Ctx) error {
	return h.getPagesByProjectID(c, "")
}

func (h *PageHandler) getPagesByProjectID(c fiber.Ctx, userID string) error {
	ids, err := utils.MustParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	pages, err := h.pageService.GetPagesByProjectID(c.RequestCtx(), projectID, userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Project not found", "Failed to retrieve pages")
	}

	return utils.SendJSON(c, fiber.StatusOK, pages)
}

func (h *PageHandler) GetPageByID(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid", "pageid")
	if err != nil {
		return err
	}
	projectID, pageID := ids[0], ids[1]

	page, err := h.pageService.GetPageByIDForUser(c.RequestCtx(), projectID, pageID, userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Page not found", "Failed to retrieve page")
	}

	return utils.SendJSON(c, fiber.StatusOK, page)
}

func (h *PageHandler) CreatePage(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid")
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
	page := &models.Page{
		Id:        pageID,
		Name:      req.Name,
		Type:      req.Type,
		Styles:    req.Styles,
		ProjectId: projectID,
		AuditFields: models.AuditFields{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	createdPage, err := h.pageService.CreatePageAsUser(c.RequestCtx(), page, userID)
	if errors.Is(err, repositories.ErrProjectUnauthorized) || errors.Is(err, authz.ErrForbidden) {
		return utils.SendError(c, fiber.StatusForbidden, "You cannot add pages to this project", err, userID)
	}
	if errors.Is(err, repositories.ErrPageNameTaken) {
		return utils.SendError(c, fiber.StatusConflict, "A page with this name already exists in the project", err, userID)
	}
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create page")
	}

	return utils.SendJSON(c, fiber.StatusCreated, createdPage)
}

func (h *PageHandler) UpdatePage(c fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid", "pageid")
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
	utils.SetIfNotNil(rawBody, "type", req.Type)
	if req.Styles != nil {
		rawBody["styles"] = req.Styles
	}

	updates, err := utils.BuildColumnUpdates(rawBody, pageAllowedCols)
	if err != nil {
		return err
	}
	if err := utils.RequireUpdates(updates); err != nil {
		return err
	}

	if err := h.pageService.UpdatePageFieldsForUser(c.RequestCtx(), projectID, pageID, userID, updates); err != nil {
		return utils.HandleRepoError(c, err, "Page not found", "Failed to update page")
	}

	updated, err := h.pageService.GetPageByIDForUser(c.RequestCtx(), projectID, pageID, userID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Page not found", "Failed to fetch updated page")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}

// fiber:context-methods migrated
