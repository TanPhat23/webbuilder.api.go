package dto

type CreateElementEventWorkflowRequest struct {
	ElementID  string `json:"elementId"  validate:"required,min=1,max=255"`
	WorkflowID string `json:"workflowId" validate:"required,min=1,max=255"`
	EventName  string `json:"eventName"  validate:"required,min=1,max=255"`
}

type UpdateElementEventWorkflowRequest struct {
	EventName string `json:"eventName" validate:"required,min=1,max=255"`
}
