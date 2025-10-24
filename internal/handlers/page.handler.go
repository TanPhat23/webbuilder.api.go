package handlers

import (
	"my-go-app/internal/repositories"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
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
