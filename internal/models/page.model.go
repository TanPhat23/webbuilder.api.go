package models

import (
	"encoding/json"
)

type Page struct {
	Elements  []Element       `gorm:"foreignKey:PageId;references:Id;constraint:OnDelete:Cascade" json:"elements,omitempty"`
	Id        string          `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ProjectId string          `gorm:"column:ProjectId;type:varchar(255);not null;index" json:"projectId"`
	Name      string          `gorm:"column:Name;type:varchar(255);not null" json:"name"`
	Type      string          `gorm:"column:Type;type:varchar(255);not null" json:"type"`
	Logic     json.RawMessage `gorm:"column:Logic;type:jsonb" json:"logic,omitempty"`
	Styles    json.RawMessage `gorm:"column:Styles;type:jsonb" json:"styles,omitempty"`
	AuditFields

	// Owned by realtime: incremented on every committed logic update.
	LogicRevision int64 `gorm:"column:LogicRevision;not null;default:0;<-:false" json:"-"`
}

func (Page) TableName() string {
	return `public."Page"`
}
