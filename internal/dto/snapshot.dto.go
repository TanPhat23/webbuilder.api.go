package dto

type SaveSnapshotRequest struct {
	Elements  []any  `json:"elements"  validate:"required,min=1"`
	Name      string `json:"name"      validate:"required,min=1,max=255"`
	Type      string `json:"type"      validate:"omitempty,min=1,max=100"`
	Id        string `json:"id"        validate:"required,min=1,max=255"`
	Timestamp int64  `json:"timestamp"`
}
