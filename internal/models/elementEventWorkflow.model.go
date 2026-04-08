package models

import "time"

type ElementEventWorkflow struct {
	Element  *Element       `gorm:"foreignKey:ElementId;references:Id;constraint:OnDelete:Cascade" json:"-"`
	Workflow *EventWorkflow `gorm:"foreignKey:WorkflowId;references:Id;constraint:OnDelete:Cascade" json:"-"`

	CreatedAt time.Time `gorm:"column:CreatedAt;type:timestamp;default:CURRENT_TIMESTAMP" json:"createdAt"`

	Id         string `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	EventName  string `gorm:"column:EventName;type:varchar(255);not null;index" json:"eventName"`
	ElementId  string `gorm:"column:ElementId;type:varchar(255);not null;index" json:"elementId"`
	WorkflowId string `gorm:"column:WorkflowId;type:varchar(255);not null;index" json:"workflowId"`
}

func (ElementEventWorkflow) TableName() string {
	return `public."ElementEventWorkflow"`
}