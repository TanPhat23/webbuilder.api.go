package models

import "time"

type AuditFields struct {
	CreatedAt time.Time  `gorm:"column:CreatedAt" json:"createdAt,omitempty"`
	UpdatedAt time.Time  `gorm:"column:UpdatedAt" json:"updatedAt,omitempty"`
	DeletedAt *time.Time `gorm:"column:DeletedAt" json:"deletedAt,omitempty"`
}

type BaseAuditFields struct {
	CreatedAt time.Time `gorm:"column:CreatedAt" json:"createdAt,omitempty"`
	UpdatedAt time.Time `gorm:"column:UpdatedAt" json:"updatedAt,omitempty"`
}