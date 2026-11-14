package entity

import "time"

type SnapshotType string

type snapshotTypes struct {
	Working SnapshotType
	Version SnapshotType
}

var SnapshotTypes = snapshotTypes{
	Working: "working",
	Version: "version",
}

type Snapshot struct {
	ID        string       `json:"id" gorm:"primaryKey;column:Id"`
	ProjectID string       `json:"projectId" gorm:"column:ProjectId;index:idx_project_timestamp,priority:1;index:idx_project_type,priority:1"`
	PageID    string       `json:"pageId" gorm:"column:PageId;index:idx_page_timestamp,priority:1;index:idx_page_type,priority:1"`
	Elements  []Element    `json:"elements" gorm:"serializer:json;type:jsonb;column:Elements"`
	Timestamp int64        `json:"timestamp" gorm:"column:Timestamp;index:idx_project_timestamp,priority:2"`
	CreatedAt time.Time    `json:"createdAt" gorm:"column:CreatedAt;index;default:now()"`
	Name      string       `json:"name" gorm:"type:varchar(255);column:Name"`
	Type      SnapshotType `json:"type" gorm:"type:varchar(50);column:Type;index:idx_project_type,priority:2"`

	Project *Project `json:"project,omitempty" gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE"`
}

func (Snapshot) TableName() string {
	return "Snapshot"
}
