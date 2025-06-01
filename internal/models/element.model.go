package models

import "encoding/json"

type Element struct {
    Type           string                 `json:"type" db:"type"`
    ID             string                 `json:"id" db:"id"`
    Content        string                 `json:"content" db:"content"`
    IsSelected     bool                   `json:"isSelected" db:"is_selected"`
    Name           *string                `json:"name,omitempty" db:"name"`
    Styles         json.RawMessage			`json:"styles,omitempty" db:"styles"`
    TailwindStyles *string                `json:"tailwindStyles,omitempty" db:"tailwind_styles"`
    X              float64                `json:"x" db:"x"`
    Y              float64                `json:"y" db:"y"`
    Src            *string                `json:"src,omitempty" db:"src"`
    Href           *string                `json:"href,omitempty" db:"href"`
    ParentID       *string                `json:"parentId,omitempty" db:"parent_id"`
    ProjectID      *string                `json:"projectId,omitempty" db:"project_id"`
    Order          *int                   `json:"order,omitempty" db:"order"`
}

// FrameElement extends Element
type FrameElement struct {
    Element
    Elements []interface{} `json:"elements" db:"-"`
}

// CarouselElement extends Element
type CarouselElement struct {
    Element
    CarouselSettings map[string]interface{} `json:"carouselSettings" db:"carousel_settings"`
    Elements         []interface{}          `json:"elements" db:"-"`
}

// ButtonElement extends Element
type ButtonElement struct {
    Element
    ButtonType    string        `json:"buttonType" db:"button_type"`
    FrameElement  *FrameElement `json:"element,omitempty" db:"-"`
}

// InputElement extends Element
type InputElement struct {
    Element
    InputSettings map[string]interface{} `json:"inputSettings" db:"input_settings"`
}

// ListElement extends Element
type ListElement struct {
    Element
    Elements []interface{} `json:"elements" db:"-"`
}

// SelectElement extends Element
type SelectElement struct {
    Element
    Options       []map[string]interface{} `json:"options" db:"options"`
    SelectSettings map[string]interface{}   `json:"selectSettings,omitempty" db:"select_settings"`
}

// ChartDataset for chart elements
type ChartDataset struct {
    Label           string      `json:"label"`
    Data            []float64   `json:"data"`
    BackgroundColor interface{} `json:"backgroundColor,omitempty"`
    BorderColor     interface{} `json:"borderColor,omitempty"`
    BorderWidth     *int        `json:"borderWidth,omitempty"`
    Fill            *bool       `json:"fill,omitempty"`
}


// FormElement extends Element
type FormElement struct {
    Element
    Elements     []interface{}          `json:"elements" db:"-"`
    FormSettings map[string]interface{} `json:"formSettings,omitempty" db:"form_settings"`
}

type EditorElement interface {
    GetElement() Element
    GetType() string
}

// Implement the interface for each element type
func (e Element) GetElement() Element {
    return e
}

func (e Element) GetType() string {
    return e.Type
}

func (fe FormElement) GetType() string {
    return fe.Type
}

func (fe FrameElement) GetElement() Element {
    return fe.Element
}

func (fe FrameElement) GetType() string {
    return fe.Type
}

func (ce CarouselElement) GetElement() Element {
    return ce.Element
}

func (ce CarouselElement) GetType() string {
    return ce.Type
}

func (be ButtonElement) GetElement() Element {
    return be.Element
}

func (be ButtonElement) GetType() string {
    return be.Type
}

func (ie InputElement) GetElement() Element {
    return ie.Element
}

func (ie InputElement) GetType() string {
    return ie.Type
}

func (le ListElement) GetElement() Element {
    return le.Element
}

func (le ListElement) GetType() string {
    return le.Type
}

func (se SelectElement) GetElement() Element {
    return se.Element
}

func (se SelectElement) GetType() string {
    return se.Type
}

func (fe FormElement) GetElement() Element {
    return fe.Element
}
