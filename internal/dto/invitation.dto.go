package dto

import "my-go-app/internal/models"

type CreateInvitationRequest struct {
	ProjectID string                  `json:"projectId" validate:"required,min=1"`
	Email     string                  `json:"email" validate:"required,email,max=255"`
	Role      models.CollaboratorRole `json:"role" validate:"required,oneof=owner editor viewer"`
}

type AcceptInvitationRequest struct {
	Token string `json:"token" validate:"required,min=1,max=255"`
}

type UpdateInvitationStatusRequest struct {
	Status models.InvitationStatus `json:"status" validate:"required,oneof=pending accepted expired cancelled"`
}
