package dto

import "my-go-app/internal/models"

type CreateCollaboratorRequest struct {
	ProjectID string                  `json:"projectId" validate:"required,min=1"`
	UserID    string                  `json:"userId" validate:"required,min=1"`
	Role      models.CollaboratorRole `json:"role" validate:"required,oneof=owner editor viewer"`
}

type UpdateCollaboratorRoleRequest struct {
	Role models.CollaboratorRole `json:"role" validate:"required,oneof=owner editor viewer"`
}
