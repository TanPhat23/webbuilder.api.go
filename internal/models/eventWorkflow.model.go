package models

import (
	"encoding/json"
	"time"
)

type EventWorkflow struct {
	Id          string          `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ProjectId   string          `gorm:"column:ProjectId;type:varchar(255);not null;index" json:"projectId"`
	Name        string          `gorm:"column:Name;type:varchar(255);not null" json:"name"`
	Description *string         `gorm:"column:Description;type:text" json:"description,omitempty"`
	CanvasData  json.RawMessage `gorm:"column:CanvasData;type:jsonb" json:"canvasData,omitempty"`
	Handlers		json.RawMessage `gorm:"column:Handlers;type:jsonb" json:"handlers,omitempty"`
	Enabled     bool            `gorm:"column:Enabled;not null;default:true" json:"enabled"`
	CreatedAt   time.Time       `gorm:"column:CreatedAt;type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time       `gorm:"column:UpdatedAt;type:timestamp;default:CURRENT_TIMESTAMP" json:"updatedAt"`

	// Relations
	Project               *Project               `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"-"`
	ElementEventWorkflows []ElementEventWorkflow `gorm:"foreignKey:WorkflowId;references:Id" json:"elementEventWorkflows,omitempty"`
}

func (EventWorkflow) TableName() string {
	return `public."EventWorkflow"`
}
