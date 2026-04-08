package models

import "encoding/json"

// Type classification maps for optimized type checking
var (
	containerTypes = map[string]bool{
		"Section":          true,
		"Frame":            true,
		"Carousel":         true,
		"List":             true,
		"Select":           true,
		"Form":             true,
		"DataLoader":       true,
		"CMSContentList":   true,
		"CMSContentItem":   true,
		"CMSContentGrid":   true,
	}

	cmsTypes = map[string]bool{
		"CMSContentList": true,
		"CMSContentItem": true,
		"CMSContentGrid": true,
	}

	mediaTypes = map[string]bool{
		"Image":  true,
		"Video":  true,
		"Iframe": true,
	}

	contentTypes = map[string]bool{
		"Button": true,
		"Text":   true,
		"Input":  true,
		"Select": true,
	}

	hrefTypes = map[string]bool{
		"Button": true,
		"Text":   true,
	}
)

type Element struct {
	Elements       []Element              `gorm:"foreignKey:ParentId;references:Id;constraint:OnDelete:Cascade" json:"-"`
	EventWorkflows []ElementEventWorkflow `gorm:"foreignKey:ElementId;references:Id;constraint:OnDelete:Cascade" json:"-"`
	Comments       []ElementComment       `gorm:"foreignKey:ElementId;references:Id" json:"-"`
	Page           *Page                  `gorm:"foreignKey:PageId;references:Id;constraint:OnDelete:Cascade" json:"-"`
	Parent         *Element               `gorm:"foreignKey:ParentId;references:Id;constraint:OnDelete:Cascade" json:"-"`
	Id             string                 `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Type           string                 `gorm:"column:Type;type:varchar(32);not null" json:"type"`
	PageId         *string                `gorm:"column:PageId;type:varchar(255);index" json:"pageId,omitempty"`
	ParentId       *string                `gorm:"column:ParentId;type:varchar(255);index" json:"parentId,omitempty"`
	Name           *string                `gorm:"column:Name;type:varchar(255)" json:"name,omitempty"`
	Content        *string                `gorm:"column:Content;type:text" json:"content,omitempty"`
	Href           *string                `gorm:"column:Href;type:varchar(255)" json:"href,omitempty"`
	Src            *string                `gorm:"column:Src;type:varchar(255)" json:"src,omitempty"`
	TailwindStyles *string                `gorm:"column:TailwindStyles;type:varchar(255)" json:"tailwindStyles,omitempty"`
	Styles         json.RawMessage        `gorm:"column:Styles;type:jsonb" json:"styles,omitempty"`
	Settings       *json.RawMessage       `gorm:"column:Settings;type:jsonb" json:"settings,omitempty"`
	Order          int                    `gorm:"column:Order;default:0" json:"order"`
}

type EditorElement interface {
	GetElement() *Element
	GetType() string
}

func (e *Element) GetElement() *Element {
	return e
}

func (e *Element) GetType() string {
	return e.Type
}

func (e *Element) IsContainer() bool {
	return containerTypes[e.Type]
}

func (e *Element) HasElements() bool {
	return e.IsContainer()
}

func (e *Element) IsButton() bool {
	return e.Type == "Button"
}

func (e *Element) IsInput() bool {
	return e.Type == "Input"
}

func (e *Element) IsText() bool {
	return e.Type == "Text"
}

func (e *Element) IsSection() bool {
	return e.Type == "Section"
}

func (e *Element) IsFrame() bool {
	return e.Type == "Frame"
}

func (e *Element) IsCarousel() bool {
	return e.Type == "Carousel"
}

func (e *Element) IsList() bool {
	return e.Type == "List"
}

func (e *Element) IsSelect() bool {
	return e.Type == "Select"
}

func (e *Element) IsForm() bool {
	return e.Type == "Form"
}

func (e *Element) IsDataLoader() bool {
	return e.Type == "DataLoader"
}

func (e *Element) IsCMSContent() bool {
	return cmsTypes[e.Type]
}

func (e *Element) IsCMSContentList() bool {
	return e.Type == "CMSContentList"
}

func (e *Element) IsCMSContentItem() bool {
	return e.Type == "CMSContentItem"
}

func (e *Element) IsCMSContentGrid() bool {
	return e.Type == "CMSContentGrid"
}

func (e *Element) IsCustomElement() bool {
	return e.Type == "CustomElement"
}

func (e *Element) CanHaveHref() bool {
	return hrefTypes[e.Type]
}

func (e *Element) CanHaveSrc() bool {
	return mediaTypes[e.Type]
}

func (e *Element) CanHaveContent() bool {
	return contentTypes[e.Type]
}

func (Element) TableName() string {
	return `public."Element"`
}