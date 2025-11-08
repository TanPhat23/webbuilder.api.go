package models

import (
	"time"
)

type Collaborator struct {
	Id        string           `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	UserId    string           `gorm:"column:UserId;type:varchar(255);not null;index" json:"userId"`
	ProjectId string           `gorm:"column:ProjectId;type:varchar(255);not null;index" json:"projectId"`
	Role      CollaboratorRole `gorm:"column:Role;type:varchar(50);not null;default:'editor'" json:"role"`
	CreatedAt time.Time        `gorm:"column:CreatedAt;default:CURRENT_TIMESTAMP" json:"createdAt,omitempty"`
	UpdatedAt time.Time        `gorm:"column:UpdatedAt;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updatedAt,omitempty"`
	User      User             `gorm:"foreignKey:UserId" json:"user,omitempty"`
	Project   Project          `gorm:"foreignKey:ProjectId" json:"project,omitempty"`
}

func (Collaborator) TableName() string {
	return `public."Collaborator"`
}

// Request/Response DTOs
type CollaboratorResponse struct {
	Id        string           `json:"id"`
	UserId    string           `json:"userId"`
	ProjectId string           `json:"projectId"`
	Role      CollaboratorRole `json:"role"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
	User      *User            `json:"user,omitempty"`
}

type UpdateCollaboratorRoleRequest struct {
	Role CollaboratorRole `json:"role" validate:"required"`
}
