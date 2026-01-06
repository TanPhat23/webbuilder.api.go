package models

import "time"

type ElementEventWorkflow struct {
	Id         string    `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ElementId  string    `gorm:"column:ElementId;type:varchar(255);not null;index" json:"elementId"`
	WorkflowId string    `gorm:"column:WorkflowId;type:varchar(255);not null;index" json:"workflowId"`
	EventName  string    `gorm:"column:EventName;type:varchar(255);not null;index" json:"eventName"`
	CreatedAt  time.Time `gorm:"column:CreatedAt;type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`

	// Relations
	Element  *Element       `gorm:"foreignKey:ElementId;references:Id;constraint:OnDelete:Cascade" json:"-"`
	Workflow *EventWorkflow `gorm:"foreignKey:WorkflowId;references:Id;constraint:OnDelete:Cascade" json:"-"`
}

func (ElementEventWorkflow) TableName() string {
	return `public."ElementEventWorkflow"`
}
