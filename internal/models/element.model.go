package models

import "encoding/json"

type Element struct {
	Id             string          `gorm:"primaryKey;column:Id;type:varchar(255)" json:"id"`
	Type           string          `gorm:"column:Type;type:varchar(32);not null" json:"type"`
	Content        *string         `gorm:"column:Content;type:text" json:"content,omitempty"`
	Name           *string         `gorm:"column:Name;type:varchar(255)" json:"name,omitempty"`
	Styles         json.RawMessage `gorm:"column:Styles;type:jsonb" json:"styles,omitempty"`
	TailwindStyles *string         `gorm:"column:TailwindStyles;type:varchar(255)" json:"tailwindStyles,omitempty"`
	Src            *string         `gorm:"column:Src;type:varchar(255)" json:"src,omitempty"`
	Href           *string         `gorm:"column:Href;type:varchar(255)" json:"href,omitempty"`
	ParentId       *string         `gorm:"column:ParentId;type:varchar(255);index" json:"parentId,omitempty"`
	PageId         *string         `gorm:"column:PageId;type:varchar(255);index" json:"pageId,omitempty"`
	Order          int             `gorm:"column:Order;default:0" json:"order"`

	// Temporary field for backward compatibility (not persisted)
	Settings *json.RawMessage `gorm:"-" json:"settings,omitempty"`

	// Relations
	Page               *Page                    `gorm:"foreignKey:PageId;references:Id;constraint:OnDelete:Cascade" json:"-"`
	Parent             *Element                 `gorm:"foreignKey:ParentId;references:Id;constraint:OnDelete:Cascade" json:"-"`
	Elements           []Element                `gorm:"foreignKey:ParentId;references:Id;constraint:OnDelete:Cascade" json:"-"`
	EventWorkflows     []ElementEventWorkflow   `gorm:"foreignKey:ElementId;references:Id;constraint:OnDelete:Cascade" json:"-"`
	SettingRecord      *Setting                 `gorm:"foreignKey:ElementId;references:Id;constraint:OnDelete:Cascade" json:"-"`
	Comments           []ElementComment         `gorm:"foreignKey:ElementId;references:Id" json:"-"`
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
	switch e.Type {
	case "Section", "Frame", "Carousel", "List", "Select", "Form", "DataLoader",
		"CMSContentList", "CMSContentItem", "CMSContentGrid":
		return true
	default:
		return false
	}
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
	switch e.Type {
	case "CMSContentList", "CMSContentItem", "CMSContentGrid":
		return true
	default:
		return false
	}
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
	return e.Type == "Button" || e.Type == "Text"
}

func (e *Element) CanHaveSrc() bool {
	switch e.Type {
	case "Image", "Video", "Iframe":
		return true
	default:
		return false
	}
}

func (e *Element) CanHaveContent() bool {
	switch e.Type {
	case "Button", "Text", "Input", "Select":
		return true
	default:
		return false
	}
}

func (Element) TableName() string {
	return `public."Element"`
}
