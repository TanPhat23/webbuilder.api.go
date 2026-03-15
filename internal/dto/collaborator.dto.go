package dto

import "my-go-app/internal/models"

// UpdateCollaboratorRoleRequest contains the required fields to update a collaborator's role.
type UpdateCollaboratorRoleRequest struct {
	Role models.CollaboratorRole `json:"role" validate:"required,oneof=owner editor viewer"`
}