package dto

// SaveSnapshotRequest contains the required fields to save a new snapshot.
type SaveSnapshotRequest struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type,omitempty"`
	Elements  []any  `json:"elements"             validate:"required"`
	Timestamp int64  `json:"timestamp,omitempty"`
}