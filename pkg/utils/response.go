package utils

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error        string `json:"error"`
	ErrorMessage string `json:"message,omitempty"`
	UserID       string `json:"userId,omitempty"`
	RequestID    string `json:"requestId,omitempty"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// SendError sends a standardized error response
func SendError(c fiber.Ctx, status int, message string, err error, userID ...string) error {
	response := ErrorResponse{
		Error:        message,
		ErrorMessage: "",
	}
	if err != nil {
		response.ErrorMessage = err.Error()
	}
	if len(userID) > 0 {
		response.UserID = userID[0]
	}
	response.RequestID = RequestID(c)
	LogRequestError(c, status, message, err)
	return c.Status(status).JSON(response)
}

// LogRequestError records a failed request with slog. Server errors (5xx) are
// logged at error level and include the underlying error; client errors are
// logged at warn level so they stay visible without raising alarms.
func LogRequestError(c fiber.Ctx, status int, message string, err error) {
	attrs := []any{
		slog.Int("status", status),
		slog.String("method", c.Method()),
		slog.String("path", c.Path()),
		slog.String("message", message),
		slog.String("requestId", RequestID(c)),
	}
	if userID, ok := c.Locals("userId").(string); ok && userID != "" {
		attrs = append(attrs, slog.String("userId", userID))
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}
	if status >= fiber.StatusInternalServerError {
		slog.Error("request failed", attrs...)
		return
	}
	slog.Warn("request rejected", attrs...)
}

// SendSuccess sends a standardized success response
func SendSuccess(c fiber.Ctx, status int, message string, data ...any) error {
	response := SuccessResponse{
		Message: message,
	}
	if len(data) > 0 {
		response.Data = data[0]
	}
	return c.Status(status).JSON(response)
}

// SendJSON sends a JSON response with the given status
func SendJSON(c fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(data)
}

// SendNoContent sends a 204 No Content response
func SendNoContent(c fiber.Ctx) error {
	return c.Status(fiber.StatusNoContent).Send(nil)
}

// RequestID returns the id assigned by the requestid middleware, or "" when
// the middleware is not installed.
func RequestID(c fiber.Ctx) string {
	return requestid.FromContext(c)
}
