package entity

import (
	"time"
)

// EventWorkflow represents a defined set of actions that can be triggered by events.
// This matches the Prisma EventWorkflow model.
type EventWorkflow struct {
	ID          string         `json:"id" gorm:"primaryKey;column:Id"`
	ProjectID   string         `json:"projectId" gorm:"column:ProjectId;index;uniqueIndex:ux_project_workflow_name"`
	Name        string         `json:"name" gorm:"column:Name;type:varchar(255);uniqueIndex:ux_project_workflow_name"`
	Description *string        `json:"description,omitempty" gorm:"column:Description;type:text"`
	Handlers    []any          `json:"handlers" gorm:"column:Handlers;serializer:json;type:jsonb;default:'[]'"`
	Enabled     bool           `json:"enabled" gorm:"column:Enabled;default:true;index"`
	CreatedAt   time.Time      `json:"createdAt" gorm:"column:CreatedAt"`
	UpdatedAt   time.Time      `json:"updatedAt" gorm:"column:UpdatedAt"`
	CanvasData  map[string]any `json:"canvasData,omitempty" gorm:"column:CanvasData;serializer:json;type:jsonb"`

	ElementEventWorkflows []ElementEventWorkflow `json:"elementEventWorkflows,omitempty" gorm:"foreignKey:WorkflowID"`
}

// TableName sets the database table name for EventWorkflow.
func (EventWorkflow) TableName() string {
	return "EventWorkflow"
}

// ElementEventWorkflow represents the junction between an Element and an EventWorkflow.
// It defines which workflow is triggered by a specific event (e.g., "onClick") on an element.
type ElementEventWorkflow struct {
	ID         string    `json:"id" gorm:"primaryKey;column:Id"`
	ElementID  string    `json:"elementId" gorm:"column:ElementId;index;uniqueIndex:ux_element_workflow_event"`
	WorkflowID string    `json:"workflowId" gorm:"column:WorkflowId;index;uniqueIndex:ux_element_workflow_event"`
	EventName  string    `json:"eventName" gorm:"column:EventName;type:varchar(100);index;uniqueIndex:ux_element_workflow_event"`
	CreatedAt  time.Time `json:"createdAt" gorm:"column:CreatedAt"`

	Element  *Element       `json:"element,omitempty" gorm:"foreignKey:ElementID;references:ID"`
	Workflow *EventWorkflow `json:"workflow,omitempty" gorm:"foreignKey:WorkflowID;references:ID"`
}

// TableName sets the database table name for ElementEventWorkflow.
func (ElementEventWorkflow) TableName() string {
	return "ElementEventWorkflow"
}
