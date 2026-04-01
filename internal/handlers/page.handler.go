package handlers

import (
	"my-go-app/internal/dto"
	"my-go-app/internal/models"
	"my-go-app/internal/services"
	"my-go-app/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v2"
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

func (h *PageHandler) DeletePage(c *fiber.Ctx) error {
	userID, ids, err := utils.MustUserAndParams(c, "projectid", "pageid")
	if err != nil {
		return err
	}
	projectID := ids[0]
	pageID := ids[1]

	if err := h.pageService.DeletePageByProjectID(c.Context(), pageID, projectID, userID); err != nil {
		return utils.HandleRepoError(c, err, "Page not found or not owned by user", "Failed to delete page")
	}

	return utils.SendNoContent(c)
}

func (h *PageHandler) GetPagesByProjectID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	pages, err := h.pageService.GetPagesByProjectID(c.Context(), projectID)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to retrieve pages")
	}

	return utils.SendJSON(c, fiber.StatusOK, pages)
}

func (h *PageHandler) GetPageByID(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "projectid", "pageid")
	if err != nil {
		return err
	}
	pageID := ids[1]

	page, err := h.pageService.GetPageByID(c.Context(), pageID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Page not found", "Failed to retrieve page")
	}

	return utils.SendJSON(c, fiber.StatusOK, page)
}

func (h *PageHandler) CreatePage(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "projectid")
	if err != nil {
		return err
	}
	projectID := ids[0]

	var req dto.CreatePageRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	now := time.Now()
	page := &models.Page{
		Id:        uuid.New().String(),
		Name:      req.Name,
		Type:      req.Type,
		Styles:    req.Styles,
		ProjectId: projectID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdPage, err := h.pageService.CreatePage(c.Context(), page)
	if err != nil {
		return utils.HandleRepoError(c, err, "", "Failed to create page")
	}

	return utils.SendJSON(c, fiber.StatusCreated, createdPage)
}

func (h *PageHandler) UpdatePage(c *fiber.Ctx) error {
	ids, err := utils.MustParams(c, "projectid", "pageid")
	if err != nil {
		return err
	}
	pageID := ids[1]

	if _, err := h.pageService.GetPageByID(c.Context(), pageID); err != nil {
		return utils.HandleRepoError(c, err, "Page not found", "Failed to verify page")
	}

	var req dto.UpdatePageRequest
	if err := utils.ValidateAndParseBody(c, &req); err != nil {
		return err
	}

	rawBody := map[string]any{}
	if req.Name != nil   { rawBody["name"] = *req.Name }
	if req.Type != nil   { rawBody["type"] = *req.Type }
	if req.Styles != nil { rawBody["styles"] = req.Styles }

	updates, err := utils.BuildColumnUpdates(rawBody, pageAllowedCols)
	if err != nil {
		return err
	}
	if err := utils.RequireUpdates(updates); err != nil {
		return err
	}

	if err := h.pageService.UpdatePageFields(c.Context(), pageID, updates); err != nil {
		return utils.HandleRepoError(c, err, "Page not found", "Failed to update page")
	}

	updated, err := h.pageService.GetPageByID(c.Context(), pageID)
	if err != nil {
		return utils.HandleRepoError(c, err, "Page not found", "Failed to fetch updated page")
	}

	return utils.SendJSON(c, fiber.StatusOK, updated)
}