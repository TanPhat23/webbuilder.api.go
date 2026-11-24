package models

import (
	"regexp"
	"strings"
	"time"
)

var nonSlugCharsRe = regexp.MustCompile(`[^a-z0-9]+`)

type ContentType struct {
	Fields  []ContentField `gorm:"foreignKey:ContentTypeId" json:"fields,omitempty"`
	Items   []ContentItem  `gorm:"foreignKey:ContentTypeId" json:"items,omitempty"`
	Project *Project       `gorm:"foreignKey:ProjectId;references:ID" json:"-"`

	Id               string  `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ProjectId        *string `gorm:"column:ProjectId;type:varchar(255);index" json:"projectId,omitempty"`
	Name             string  `gorm:"column:Name;type:varchar(255);not null" json:"name"`
	Slug             string  `gorm:"column:Slug;type:varchar(255);not null" json:"slug"`
	Description      *string `gorm:"column:Description;type:text" json:"description,omitempty"`
	PublicRead       bool    `gorm:"column:PublicRead;not null;default:true" json:"publicRead"`
	PublicCreate     bool    `gorm:"column:PublicCreate;not null;default:false" json:"publicCreate"`
	DefaultPublished bool    `gorm:"column:DefaultPublished;not null;default:false" json:"defaultPublished"`
	RateLimitPerMin  int     `gorm:"column:RateLimitPerMin;not null;default:30" json:"rateLimitPerMin"`

	CreatedAt time.Time `gorm:"column:CreatedAt" json:"-"`
	UpdatedAt time.Time `gorm:"column:UpdatedAt" json:"-"`
}

func (ContentType) TableName() string {
	return `public."ContentType"`
}

func SlugifyCollectionName(name, fallbackID string) string {
	slug := nonSlugCharsRe.ReplaceAllString(strings.ToLower(strings.TrimSpace(name)), "-")
	slug = strings.Trim(slug, "-")
	if slug != "" {
		return slug
	}
	cleanID := strings.ReplaceAll(strings.TrimSpace(fallbackID), "-", "")
	if len(cleanID) > 8 {
		cleanID = cleanID[len(cleanID)-8:]
	}
	if cleanID == "" {
		return "collection"
	}
	return "collection-" + cleanID
}
