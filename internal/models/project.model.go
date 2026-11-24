package models

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

var (
	validSubdomainRe = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{1,61}[a-z0-9])$`)
	reservedSubdomains = map[string]bool{
		"www": true, "api": true, "app": true, "admin": true, "ws": true,
		"mcp": true, "preview": true, "editor": true, "dashboard": true,
		"sign-in": true, "sign-up": true, "help": true, "helppage": true,
		"static": true, "assets": true, "cdn": true, "mail": true, "webhook": true,
	}
)

type Project struct {
	Pages         []Page         `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"pages,omitempty"`
	Snapshots     []Snapshot     `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"snapshots,omitempty"`
	Collaborators []Collaborator `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"collaborators,omitempty"`
	Invitations   []Invitation   `gorm:"foreignKey:ProjectId;references:Id;constraint:OnDelete:Cascade" json:"invitations,omitempty"`

	ID      string `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Name    string `gorm:"column:Name;type:varchar(255);not null" json:"name"`
	OwnerId string `gorm:"column:OwnerId;type:varchar(255);not null" json:"ownerId"`

	Owner           User             `gorm:"foreignKey:OwnerId" json:"owner,omitempty"`
	Logic           json.RawMessage  `gorm:"column:Logic;type:jsonb" json:"logic,omitempty"`
	Styles          *json.RawMessage `gorm:"column:Styles;type:jsonb" json:"styles,omitempty"`
	Header          *json.RawMessage `gorm:"column:Header;type:jsonb" json:"header,omitempty"`
	Description     *string          `gorm:"column:Description;type:text" json:"description,omitempty"`
	Subdomain       *string          `gorm:"column:Subdomain;type:varchar(255)" json:"subdomain,omitempty"`

	Published bool `gorm:"column:Published;not null;default:false" json:"published"`

	*AuditFields

	// Owned by realtime: incremented on every committed logic update.
	LogicRevision int64 `gorm:"column:LogicRevision;not null;default:0;<-:false" json:"-"`
}

func (Project) TableName() string {
	return `public."Project"`
}

// NormalizeSubdomain validates and normalizes a subdomain label. Empty input
// returns ("", false, nil) so callers can clear an existing subdomain.
func NormalizeSubdomain(raw string) (string, bool, error) {
	cleaned := strings.ToLower(strings.TrimSpace(raw))
	if cleaned == "" {
		return "", false, nil
	}
	if !validSubdomainRe.MatchString(cleaned) || strings.Contains(cleaned, "--") {
		return "", false, errors.New("subdomain must be 3-63 characters using lowercase letters, numbers, and single hyphens")
	}
	if reservedSubdomains[cleaned] {
		return "", false, errors.New("this subdomain is reserved")
	}
	return cleaned, true, nil
}
