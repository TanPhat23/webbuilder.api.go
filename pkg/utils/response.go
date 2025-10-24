package utils

import (
	"github.com/gofiber/fiber/v2"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error        string `json:"error"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	UserID       string `json:"userId,omitempty"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    any `json:"data,omitempty"`
}

// SendError sends a standardized error response
func SendError(c *fiber.Ctx, status int, message string, err error, userID ...string) error {
	response := ErrorResponse{
		Error:        message,
		ErrorMessage: err.Error(),
	}
	if len(userID) > 0 {
		response.UserID = userID[0]
	}
	return c.Status(status).JSON(response)
}

// SendSuccess sends a standardized success response
func SendSuccess(c *fiber.Ctx, status int, message string, data ...any) error {
	response := SuccessResponse{
		Message: message,
	}
	if len(data) > 0 {
		response.Data = data[0]
	}
	return c.Status(status).JSON(response)
}

// SendJSON sends a JSON response with the given status
func SendJSON(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(data)
}

// SendNoContent sends a 204 No Content response
func SendNoContent(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNoContent).Send(nil)
}
