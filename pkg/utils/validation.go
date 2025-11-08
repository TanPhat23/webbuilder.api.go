package utils

import (
	"github.com/gofiber/fiber/v2"
)

// ValidateRequiredParam checks if a required parameter is present
func ValidateRequiredParam(c *fiber.Ctx, paramName string) (string, error) {
	value := c.Params(paramName)
	if value == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, paramName+" is required")
	}
	return value, nil
}

// ValidateUserID extracts and validates user ID from context locals
func ValidateUserID(c *fiber.Ctx) (string, error) {
	userID, ok := c.Locals("userId").(string)
	if !ok || userID == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Unauthorized: You must be logged in")
	}
	return userID, nil
}

// ValidateJSONBody parses JSON body into the provided struct
func ValidateJSONBody(c *fiber.Ctx, v any) error {
	if err := c.BodyParser(v); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON body: "+err.Error())
	}
	return nil
}

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

func ValidateCollaboratorID(collaboratorID string) error {
	if collaboratorID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Collaborator ID is required")
	}
	return nil
}