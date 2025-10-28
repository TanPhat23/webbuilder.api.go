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

type Invitation struct {
	Id         string           `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Email      string           `gorm:"column:Email;type:varchar(255);not null" json:"email"`
	ProjectId  string           `gorm:"column:ProjectId;type:varchar(255);not null;index" json:"projectId"`
	Role       CollaboratorRole `gorm:"column:Role;type:varchar(50);not null;default:'editor'" json:"role"`
	Token      string           `gorm:"column:Token;type:varchar(255);not null;uniqueIndex" json:"token"`
	ExpiresAt  time.Time        `gorm:"column:ExpiresAt;not null" json:"expiresAt"`
	CreatedAt  time.Time        `gorm:"column:CreatedAt;default:CURRENT_TIMESTAMP" json:"createdAt,omitempty"`
	AcceptedAt *time.Time       `gorm:"column:AcceptedAt" json:"acceptedAt,omitempty"`
	Project    Project          `gorm:"foreignKey:ProjectId" json:"project,omitempty"`
}

func (Invitation) TableName() string {
	return `public."Invitation"`
}

// Request/Response DTOs
type CreateInvitationRequest struct {
	ProjectID string              `json:"projectId" validate:"required"`
	Email     string              `json:"email" validate:"required,email"`
	Role      CollaboratorRole    `json:"role"`
}

type InvitationResponse struct {
	Id         string              `json:"id"`
	Email      string              `json:"email"`
	ProjectId  string              `json:"projectId"`
	Role       CollaboratorRole    `json:"role"`
	Token      string              `json:"token"`
	ExpiresAt  time.Time           `json:"expiresAt"`
	CreatedAt  time.Time           `json:"createdAt"`
	AcceptedAt *time.Time          `json:"acceptedAt,omitempty"`
}

type AcceptInvitationRequest struct {
	Token string `json:"token" validate:"required"`
}
