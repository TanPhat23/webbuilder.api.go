package middleware

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func ValidateMiddleware[T any](validate *validator.Validate) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body T
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body: "+err.Error())
		}

		if err := validate.Struct(body); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+validationErrors.Error())
			}
			return fiber.NewError(fiber.StatusBadRequest, "Validation error: "+err.Error())
		}
		return c.Next()
	}
}
