package configs

import (
	"errors"
	"my-go-app/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

func FiberConfig() fiber.Config {
	return fiber.Config{
		Prefork:       false,
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Fiber",
		AppName:       "Webbuilder v1.0.1",
		ErrorHandler:  jsonErrorHandler,
	}
}

// jsonErrorHandler is the single error-handling entry point for the entire API.
// It guarantees that every error — whether a built-in *fiber.Error, our own
// *utils.ValidationError, or any unexpected error — is returned as JSON so
// clients always receive a consistent error shape.
func jsonErrorHandler(c *fiber.Ctx, err error) error {
	// 1. Structured validation errors (422 Unprocessable Entity)
	var valErr *utils.ValidationError
	if errors.As(err, &valErr) {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"error":  "Validation failed",
			"fields": valErr.Fields,
		})
	}

	// 2. Explicit fiber errors (e.g. fiber.NewError(404, "not found"))
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(fiber.Map{
			"error":   fiberErr.Message,
			"message": fiberErr.Message,
		})
	}

	// 3. Fallback — unexpected / unhandled errors
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "Internal server error",
	})
}
