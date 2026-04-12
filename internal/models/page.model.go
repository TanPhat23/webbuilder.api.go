package models

import (
	"encoding/json"
)

type Page struct {
	Elements  []Element       `gorm:"foreignKey:PageId;references:Id;constraint:OnDelete:Cascade" json:"Elements,omitempty"`
	Id        string          `gorm:"primaryKey;column:Id;type:varchar(255)" json:"Id"`
	ProjectId string          `gorm:"column:ProjectId;type:varchar(255);not null;index" json:"ProjectId"`
	Name      string          `gorm:"column:Name;type:varchar(255);not null" json:"Name"`
	Type      string          `gorm:"column:Type;type:varchar(255);not null" json:"Type"`
	Styles    json.RawMessage `gorm:"column:Styles;type:jsonb" json:"Styles,omitempty"`
	AuditFields
}

func (Page) TableName() string {
	return `public."Page"`
}