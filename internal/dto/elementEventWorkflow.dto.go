package dto

// CreateElementEventWorkflowRequest contains the required fields to create
// a new element event workflow association.
type CreateElementEventWorkflowRequest struct {
	ElementID  string `json:"elementId"  validate:"required"`
	WorkflowID string `json:"workflowId" validate:"required"`
	EventName  string `json:"eventName"  validate:"required"`
}

// UpdateElementEventWorkflowRequest contains the patchable fields on an
// element event workflow association.
type UpdateElementEventWorkflowRequest struct {
	EventName string `json:"eventName" validate:"required"`
}