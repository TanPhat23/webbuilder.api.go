package models

import "time"

type Collaborator struct {
	User      User             `gorm:"foreignKey:UserId" json:"user,omitempty"`
	Project   Project          `gorm:"foreignKey:ProjectId" json:"project,omitempty"`
	Id        string           `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	UserId    string           `gorm:"column:UserId;type:varchar(255);not null;index" json:"userId"`
	ProjectId string           `gorm:"column:ProjectId;type:varchar(255);not null;index" json:"projectId"`
	Role      CollaboratorRole `gorm:"column:Role;type:varchar(50);not null;default:'editor'" json:"role"`
	BaseAuditFields
}

func (Collaborator) TableName() string {
	return `public."Collaborator"`
}

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
	Role CollaboratorRole `json:"role" validate:"required,oneof=owner editor viewer"`
}
