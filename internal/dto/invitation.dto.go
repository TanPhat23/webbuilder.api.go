package dto

import "my-go-app/internal/models"

// UpdateInvitationStatusRequest contains the required fields to update an invitation's status.
type UpdateInvitationStatusRequest struct {
	Status models.InvitationStatus `json:"status" validate:"required,oneof=pending accepted expired cancelled"`
}