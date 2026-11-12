package models

import (
	"time"
)

type CollaboratorRole string

const (
	RoleOwner  CollaboratorRole = "owner"
	RoleEditor CollaboratorRole = "editor"
	RoleViewer CollaboratorRole = "viewer"
)

type InvitationStatus string

const (
	InvitationStatusPending   InvitationStatus = "pending"
	InvitationStatusAccepted  InvitationStatus = "accepted"
	InvitationStatusExpired   InvitationStatus = "expired"
	InvitationStatusCancelled InvitationStatus = "cancelled"
)

type Invitation struct {
	Project    Project          `gorm:"foreignKey:ProjectId" json:"project,omitempty"`
	ExpiresAt  time.Time        `gorm:"column:ExpiresAt;not null" json:"expires_at"`
	CreatedAt  time.Time        `gorm:"column:CreatedAt;default:CURRENT_TIMESTAMP" json:"created_at,omitempty"`
	Id         string           `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ProjectId  string           `gorm:"column:ProjectId;type:varchar(255);not null;index" json:"project_id"`
	Email      string           `gorm:"column:Email;type:varchar(255);not null" json:"email"`
	Token      string           `gorm:"column:Token;type:varchar(255);not null;uniqueIndex" json:"token"`
	Role       CollaboratorRole `gorm:"column:Role;type:varchar(50);not null;default:'editor'" json:"role"`
	Status     InvitationStatus `gorm:"column:Status;type:varchar(50);not null;default:'pending'" json:"status"`
	AcceptedAt *time.Time       `gorm:"column:AcceptedAt" json:"accepted_at,omitempty"`
}

func (Invitation) TableName() string {
	return `public."Invitation"`
}

// Request DTOs - Ordered by size for optimal alignment (16-byte strings → enums)
type CreateInvitationRequest struct {
	ProjectID string           `json:"project_id" validate:"required,min=1"`
	Email     string           `json:"email"      validate:"required,email,max=255"`
	Role      CollaboratorRole `json:"role"       validate:"required,oneof=owner editor viewer"`
}

type InvitationResponse struct {
	Id         string           `json:"id"`
	Email      string           `json:"email"`
	ProjectId  string           `json:"project_id"`
	Role       CollaboratorRole `json:"role"`
	Token      string           `json:"token"`
	Status     InvitationStatus `json:"status"`
	ExpiresAt  time.Time        `json:"expires_at"`
	CreatedAt  time.Time        `json:"created_at"`
	AcceptedAt *time.Time       `json:"accepted_at,omitempty"`
}

type AcceptInvitationRequest struct {
	Token string `json:"token" validate:"required,min=1,max=255"`
}

type UpdateInvitationStatusRequest struct {
	Status InvitationStatus `json:"status" validate:"required,oneof=pending accepted expired cancelled"`
}
