package utils

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Validate is the shared validator instance used across all handlers.
// Initialised once at startup with the recommended v11-ready option.
var Validate = validator.New(validator.WithRequiredStructEnabled())

// ValidateRequiredParam checks if a required URL parameter is present.
func ValidateRequiredParam(c *fiber.Ctx, paramName string) (string, error) {
	value := c.Params(paramName)
	if value == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, paramName+" is required")
	}
	return value, nil
}

// ValidateUserID extracts and validates the authenticated user ID from context locals.
func ValidateUserID(c *fiber.Ctx) (string, error) {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: You must be logged in")
	}
	return userID, nil
}

// ValidateJSONBody parses the raw JSON body into the provided value.
// It does NOT run struct-level validation; use ValidateAndParseBody for that.
func ValidateJSONBody(c *fiber.Ctx, v any) error {
	if err := c.BodyParser(v); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON body: "+err.Error())
	}
	return nil
}

// ValidateStruct runs go-playground/validator against the given struct.
// On failure it returns a *ValidationError whose fields are picked up by the
// custom Fiber ErrorHandler and rendered as a structured 422 response.
func ValidateStruct(v any) error {
	if err := Validate.Struct(v); err != nil {
		valErrs, ok := err.(validator.ValidationErrors)
		if !ok {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return NewValidationError(valErrs)
	}
	return nil
}

// ValidateAndParseBody is the one-stop helper for request body handling.
// It parses the JSON body into v and then validates the resulting struct.
// Handlers should simply do:
//
//	if err := utils.ValidateAndParseBody(c, &req); err != nil {
//	    return err
//	}
func ValidateAndParseBody(c *fiber.Ctx, v any) error {
	if err := c.BodyParser(v); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON body: "+err.Error())
	}
	return ValidateStruct(v)
}

// ValidateCollaboratorRole checks that the supplied role string is a known value.
func ValidateCollaboratorRole(role string) error {
	validRoles := map[string]bool{
		"owner":  true,
		"editor": true,
		"viewer": true,
	}
	if !validRoles[role] {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid collaborator role: "+role)
	}
	return nil
}

// ValidateCollaboratorID ensures a collaborator ID is non-empty.
func ValidateCollaboratorID(collaboratorID string) error {
	if collaboratorID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Collaborator ID is required")
	}
	return nil
}
