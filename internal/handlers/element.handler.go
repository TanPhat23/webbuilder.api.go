package handlers

import (
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetElements(c *fiber.Ctx) error {

	return c.JSON(fiber.Map{
		"message": "This is a placeholder for the GetElements handler",
		"status":  "success",
		"data":    nil})
}
