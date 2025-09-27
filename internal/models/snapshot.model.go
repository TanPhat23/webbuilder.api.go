package models

import (
	"encoding/json"
	"time"
)

type Snapshot struct {
	Id        string          `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ProjectId string          `gorm:"column:ProjectId;type:varchar(255);index" json:"projectId"`
	Elements  json.RawMessage `gorm:"column:Elements;type:jsonb" json:"elements"`
	Timestamp int64           `gorm:"column:Timestamp;type:bigint" json:"timestamp"`
	CreatedAt time.Time       `gorm:"column:CreatedAt" json:"createdAt"`
}

func (Snapshot) TableName() string {
	return `public."Snapshot"`
}
