package models

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

const PageTypeComponent = "cp"

var (
	nonSlugRunRe   = regexp.MustCompile(`[^a-z0-9]+`)
	validPageSegRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

type Page struct {
	Elements  []Element       `gorm:"foreignKey:PageId;references:Id;constraint:OnDelete:Cascade" json:"elements,omitempty"`
	Id        string          `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ProjectId string          `gorm:"column:ProjectId;type:varchar(255);not null;index" json:"projectId"`
	Name      string          `gorm:"column:Name;type:varchar(255);not null" json:"name"`
	Slug      string          `gorm:"column:Slug;type:varchar(255);not null;default:''" json:"slug"`
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

// SlugifyPageName returns a normalized URL slug candidate for a page name.
// Empty, "home", and "index" map to "" (the site root "/").
func SlugifyPageName(name string) string {
	trimmed := strings.ToLower(strings.TrimSpace(name))
	if trimmed == "" || trimmed == "home" || trimmed == "index" {
		return ""
	}
	slug := strings.Trim(nonSlugRunRe.ReplaceAllString(trimmed, "-"), "-")
	return slug
}

// NormalizeExplicitSlug validates and normalizes a user-supplied page URL slug.
// "/" or "" normalizes to "" (root). Segments must contain only lowercase
// alphanumeric characters and hyphens separated by single "/".
func NormalizeExplicitSlug(raw string) (string, error) {
	cleaned := strings.ToLower(strings.TrimSpace(raw))
	cleaned = strings.Trim(cleaned, "/")
	if cleaned == "" {
		return "", nil
	}
	if len(cleaned) > 200 {
		return "", errors.New("page slug must be 200 characters or fewer")
	}
	parts := strings.Split(cleaned, "/")
	for _, part := range parts {
		if !validPageSegRe.MatchString(part) {
			return "", errors.New("page slug segments must use lowercase letters, numbers, and hyphens")
		}
	}
	return cleaned, nil
}
