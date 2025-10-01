package models

import (
	"encoding/json"
)

// Setting matches the Prisma schema for the public."Setting" table.
type Setting struct {
	Id          string         `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Name        string         `gorm:"column:Name;type:varchar(255)" json:"name"`
	SettingType string         `gorm:"column:SettingType;type:varchar(255)" json:"settingType"`
	Settings     json.RawMessage`gorm:"column:Settings;type:jsonb" json:"settings"`
	ElementId   string         `gorm:"column:ElementId;type:varchar(255);unique" json:"elementId"`
}

