package dto

import "my-go-app/internal/models"

type CreateCollaboratorRequest struct {
	ProjectID string                  `json:"projectId" validate:"required"`
	UserID    string                  `json:"userId" validate:"required"`
	Role      models.CollaboratorRole `json:"role" validate:"required,oneof=owner editor viewer"`
}

type UpdateCollaboratorRoleRequest struct {
	Role models.CollaboratorRole `json:"role" validate:"required,oneof=owner editor viewer"`
}