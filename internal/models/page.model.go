package models

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

const (
	PageTypeStatic    = "sp"
	PageTypeDynamic   = "dp"
	PageTypeComponent = "cp"
)

var (
	nonSlugRunRe      = regexp.MustCompile(`[^a-z0-9]+`)
	validPageSegRe    = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	validParamSegRe   = regexp.MustCompile(`^:[a-z][a-z0-9_]*$`)
	validCatchAllRe   = regexp.MustCompile(`^\*(?:[a-z][a-z0-9_]*)?$`)
)

type Page struct {
	Elements      []Element       `gorm:"foreignKey:PageId;references:Id;constraint:OnDelete:Cascade" json:"elements,omitempty"`
	Id            string          `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	ProjectId     string          `gorm:"column:ProjectId;type:varchar(255);not null;index" json:"projectId"`
	Name          string          `gorm:"column:Name;type:varchar(255);not null" json:"name"`
	Slug          string          `gorm:"column:Slug;type:varchar(255);not null;default:''" json:"slug"`
	Type          string          `gorm:"column:Type;type:varchar(255);not null" json:"type"`
	ContentTypeId *string         `gorm:"column:ContentTypeId;type:text" json:"contentTypeId,omitempty"`
	RequireAuth   bool            `gorm:"column:RequireAuth;not null;default:false" json:"requireAuth"`
	Logic         json.RawMessage `gorm:"column:Logic;type:jsonb" json:"logic,omitempty"`
	Styles        json.RawMessage `gorm:"column:Styles;type:jsonb" json:"styles,omitempty"`
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

// NormalizeExplicitSlug validates and normalizes a static page URL slug.
func NormalizeExplicitSlug(raw string) (string, error) {
	return NormalizeExplicitSlugForType(raw, PageTypeStatic)
}

// NormalizeExplicitSlugForType validates and normalizes a user-supplied page
// URL slug. Dynamic pages ("dp") may additionally include `:param` segments
// and a single trailing catch-all (`*` or `*param`) segment.
func NormalizeExplicitSlugForType(raw string, pageType string) (string, error) {
	cleaned := strings.ToLower(strings.TrimSpace(raw))
	cleaned = strings.Trim(cleaned, "/")
	if cleaned == "" {
		return "", nil
	}
	if len(cleaned) > 200 {
		return "", errors.New("page slug must be 200 characters or fewer")
	}

	isDynamic := pageType == PageTypeDynamic
	parts := strings.Split(cleaned, "/")
	seenParams := make(map[string]struct{}, len(parts))

	for i, part := range parts {
		if validPageSegRe.MatchString(part) {
			continue
		}
		if validParamSegRe.MatchString(part) {
			if !isDynamic {
				return "", errors.New("dynamic :param segments are only allowed on dynamic pages")
			}
			paramName := part[1:]
			if _, dup := seenParams[paramName]; dup {
				return "", errors.New("duplicate route parameter :" + paramName)
			}
			seenParams[paramName] = struct{}{}
			continue
		}
		if validCatchAllRe.MatchString(part) {
			if !isDynamic {
				return "", errors.New("catch-all * segments are only allowed on dynamic pages")
			}
			if i != len(parts)-1 {
				return "", errors.New("catch-all * segment must be the last segment of the route")
			}
			if len(part) > 1 {
				paramName := part[1:]
				if _, dup := seenParams[paramName]; dup {
					return "", errors.New("duplicate route parameter *" + paramName)
				}
				seenParams[paramName] = struct{}{}
			}
			continue
		}
		if isDynamic {
			return "", errors.New("page slug segments must be lowercase letters/numbers/hyphens, :param, or a trailing *catchall")
		}
		return "", errors.New("page slug segments must use lowercase letters, numbers, and hyphens")
	}
	return cleaned, nil
}

// HasDynamicSegments reports whether a normalized slug contains :param or * segments.
func HasDynamicSegments(slug string) bool {
	if slug == "" {
		return false
	}
	for _, part := range strings.Split(slug, "/") {
		if strings.HasPrefix(part, ":") || strings.HasPrefix(part, "*") {
			return true
		}
	}
	return false
}

// CanonicalRoutePattern replaces parameter names so structurally identical
// dynamic patterns (e.g. "blog/:slug" and "blog/:id") normalize to the same key.
func CanonicalRoutePattern(slug string) string {
	if slug == "" {
		return ""
	}
	parts := strings.Split(slug, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = ":"
		} else if strings.HasPrefix(part, "*") {
			parts[i] = "*"
		}
	}
	return strings.Join(parts, "/")
}
