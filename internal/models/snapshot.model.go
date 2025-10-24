package models

import (
	"encoding/json"
	"time"
)

type Snapshot struct {
	Id        string          `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ProjectId string          `gorm:"column:ProjectId;type:varchar(255);index" json:"projectId"`
	Name      string          `gorm:"column:Name;type:varchar(255)" json:"name"`
	Type      string          `gorm:"column:Type;type:varchar(50);default:'working'" json:"type"`
	Elements  json.RawMessage `gorm:"column:Elements;type:jsonb" json:"elements"`
	Timestamp int64           `gorm:"column:Timestamp;type:bigint" json:"timestamp"`
	CreatedAt time.Time       `gorm:"column:CreatedAt" json:"createdAt"`
}

func (Snapshot) TableName() string {
	return `public."Snapshot"`
}
