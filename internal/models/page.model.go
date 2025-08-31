package models

import (
	"encoding/json"
	"time"
)

type Page struct {
	Id        string          `gorm:"primaryKey;column:Id;type:varchar(255)" json:"Id"`
	Name      string          `gorm:"column:Name;type:varchar(255);not null" json:"Name"`
	Type      string          `gorm:"column:Type;type:varchar(255);not null" json:"Type"`
	Styles    json.RawMessage `gorm:"column:Styles;type:jsonb" json:"Styles"`
	ProjectId string          `gorm:"column:ProjectId;type:varchar(255);not null;index" json:"ProjectId"`
	CreatedAt time.Time       `gorm:"column:CreatedAt;precision:6" json:"CreatedAt"`
	UpdatedAt time.Time       `gorm:"column:UpdatedAt;precision:6" json:"UpdatedAt"`
	DeletedAt *time.Time      `gorm:"column:DeletedAt;precision:6" json:"DeletedAt,omitempty"`
}

// Table name is managed by repositories
