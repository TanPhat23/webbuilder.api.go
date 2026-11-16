package models

import "time"

// AgentToken is a personal access token for the MCP service. Only the hash of
// the secret is stored; the secret is shown once when the token is created.
// The MCP service reads this table directly to verify tokens.
type AgentToken struct {
	Id         string     `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	UserId     string     `gorm:"column:UserId;type:varchar(255);not null;index" json:"userId"`
	Name       string     `gorm:"column:Name;type:varchar(100);not null" json:"name"`
	Prefix     string     `gorm:"column:Prefix;type:varchar(16);not null" json:"prefix"`
	TokenHash  string     `gorm:"column:TokenHash;type:char(64);not null;uniqueIndex" json:"-"`
	CreatedAt  time.Time  `gorm:"column:CreatedAt" json:"createdAt"`
	LastUsedAt *time.Time `gorm:"column:LastUsedAt" json:"lastUsedAt,omitempty"`
	RevokedAt  *time.Time `gorm:"column:RevokedAt" json:"revokedAt,omitempty"`
}

func (AgentToken) TableName() string {
	return `public."AgentToken"`
}
