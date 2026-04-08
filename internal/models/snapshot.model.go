package models

import (
	"encoding/json"
)

type Snapshot struct {
	Elements  json.RawMessage `gorm:"column:Elements;type:jsonb" json:"elements"`
	Timestamp int64           `gorm:"column:Timestamp;type:bigint" json:"timestamp"`
	BaseAuditFields
	Id        string `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ProjectId string `gorm:"column:ProjectId;type:varchar(255);index" json:"projectId"`
	Name      string `gorm:"column:Name;type:varchar(255)" json:"name"`
	Type      string `gorm:"column:Type;type:varchar(50);default:'working'" json:"type"`
}

func (Snapshot) TableName() string {
	return `public."Snapshot"`
}