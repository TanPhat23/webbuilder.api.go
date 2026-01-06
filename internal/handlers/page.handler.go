package handlers

import (
	"encoding/json"
	"log"
	"my-go-app/internal/models"
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PageHandler struct {
	pageRepository repositories.PageRepositoryInterface
}

func NewPageHandler(pageRepo repositories.PageRepositoryInterface) *PageHandler {
	return &PageHandler{
		pageRepository: pageRepo,
	}
}

func (h *PageHandler) DeletePage(c *fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	pageID, err := utils.ValidateRequiredParam(c, "pageid")
	if err != nil {
		return err
	}

	userID, err := utils.ValidateUserID(c)
	if err != nil {
		return err
	}

	err = h.pageRepository.DeletePageByProjectID(c.Context(), pageID, projectID, userID)
	if err != nil {
		if err.Error() == "record not found" {
			return utils.SendError(c, fiber.StatusNotFound, "Page not found or not owned by user", nil)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to delete page", err)
	}

	return utils.SendNoContent(c)
}

// GetPagesByProjectID retrieves all pages for a project
func (h *PageHandler) GetPagesByProjectID(c *fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	pages, err := h.pageRepository.GetPagesByProjectID(c.Context(), projectID)
	if err != nil {
		log.Println("Error retrieving pages:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve pages", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, pages)
}

// GetPageByID retrieves a single page by ID
func (h *PageHandler) GetPageByID(c *fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	pageID, err := utils.ValidateRequiredParam(c, "pageid")
	if err != nil {
		return err
	}

	page, err := h.pageRepository.GetPageByID(c.Context(), pageID, projectID)
	if err != nil {
		if err == repositories.ErrPageNotFound {
			return utils.SendError(c, fiber.StatusNotFound, "Page not found", err)
		}
		log.Println("Error retrieving page:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to retrieve page", err)
	}

	return utils.SendJSON(c, fiber.StatusOK, page)
}

// CreatePage creates a new page
func (h *PageHandler) CreatePage(c *fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	var req struct {
		Name   string          `json:"name"`
		Type   string          `json:"type"`
		Styles json.RawMessage `json:"styles,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	if req.Name == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Page name is required", nil)
	}

	if req.Type == "" {
		return utils.SendError(c, fiber.StatusBadRequest, "Page type is required", nil)
	}

	page := &models.Page{
		Id:        uuid.New().String(),
		Name:      req.Name,
		Type:      req.Type,
		Styles:    req.Styles,
		ProjectId: projectID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.pageRepository.CreatePage(c.Context(), page); err != nil {
		log.Println("Error creating page:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to create page", err)
	}

	return utils.SendJSON(c, fiber.StatusCreated, page)
}

// UpdatePage updates a page
func (h *PageHandler) UpdatePage(c *fiber.Ctx) error {
	projectID, err := utils.ValidateRequiredParam(c, "projectid")
	if err != nil {
		return err
	}

	pageID, err := utils.ValidateRequiredParam(c, "pageid")
	if err != nil {
		return err
	}

	// First verify the page exists and belongs to the project
	existingPage, err := h.pageRepository.GetPageByID(c.Context(), pageID, projectID)
	if err != nil {
		if err == repositories.ErrPageNotFound {
			return utils.SendError(c, fiber.StatusNotFound, "Page not found", err)
		}
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to verify page", err)
	}

	var req map[string]any
	if err := c.BodyParser(&req); err != nil {
		return utils.SendError(c, fiber.StatusBadRequest, "Invalid request body", err)
	}

	// Build update fields
	updates := make(map[string]any)

	if name, ok := req["name"].(string); ok && name != "" {
		updates["Name"] = name
	}

	if pageType, ok := req["type"].(string); ok && pageType != "" {
		updates["Type"] = pageType
	}

	if styles, ok := req["styles"]; ok {
		updates["Styles"] = styles
	}

	if len(updates) == 0 {
		return utils.SendError(c, fiber.StatusBadRequest, "No valid fields to update", nil)
	}

	if err := h.pageRepository.UpdatePageFields(c.Context(), pageID, updates); err != nil {
		if err == repositories.ErrPageNotFound {
			return utils.SendError(c, fiber.StatusNotFound, "Page not found", err)
		}
		log.Println("Error updating page:", err)
		return utils.SendError(c, fiber.StatusInternalServerError, "Failed to update page", err)
	}

	// Get the updated page
	updatedPage, err := h.pageRepository.GetPageByID(c.Context(), pageID, projectID)
	if err != nil {
		// If we can't get the updated page, return the existing one with updates applied
		for key, value := range updates {
			switch key {
			case "Name":
				existingPage.Name = value.(string)
			case "Type":
				existingPage.Type = value.(string)
			case "Styles":
				if styles, ok := value.(json.RawMessage); ok {
					existingPage.Styles = styles
				}
			}
		}
		existingPage.UpdatedAt = time.Now()
		return utils.SendJSON(c, fiber.StatusOK, existingPage)
	}

	return utils.SendJSON(c, fiber.StatusOK, updatedPage)
}
