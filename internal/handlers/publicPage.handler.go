package handlers

import (
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v3"
)

// GetPublicByID serves a page of a published project; anything else
// (a draft project, or a page of another project) is not found.
func (h *PageHandler) GetPublicByID(c fiber.Ctx) error {
	ids, err := utils.RequireParams(c, "projectid", "pageid")
	if err != nil {
		return err
	}
	page, err := h.pageService.GetPublicByID(c.Context(), ids[0], ids[1])
	if err != nil {
		return utils.RespondError(c, err, "Page not found", "Failed to retrieve page")
	}
	return utils.SendJSON(c, fiber.StatusOK, page)
}
