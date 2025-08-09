package models

type Element struct {
    Type           string                 `json:"type" db:"type"`
    ID             string                 `json:"id" db:"id"`
    Content        string                 `json:"content" db:"content"`
    IsSelected     bool                   `json:"isSelected" db:"is_selected"`
    Name           *string                `json:"name,omitempty" db:"name"`
    Styles         map[string]any			`json:"styles,omitempty" db:"styles"`
    TailwindStyles *string                `json:"tailwindStyles,omitempty" db:"tailwind_styles"`
    X              float64                `json:"x" db:"x"`
    Y              float64                `json:"y" db:"y"`
    Src            *string                `json:"src,omitempty" db:"src"`
    Href           *string                `json:"href,omitempty" db:"href"`
    ParentID       *string                `json:"parentId,omitempty" db:"parent_id"`
    ProjectID      *string                `json:"projectId,omitempty" db:"project_id"`
    Order          *int                   `json:"order,omitempty" db:"order"`
}

type SectionElement struct {
    Element
    Elements []any `json:"elements" db:"-"`
}

// FrameElement extends Element
type FrameElement struct {
    Element
    Elements []any `json:"elements" db:"-"`
}

// CarouselElement extends Element
type CarouselElement struct {
    Element
    Settings map[string]any `json:"settings" db:"carousel_settings"`
    Elements []any          `json:"elements" db:"-"`
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
    Settings map[string]any `json:"settings" db:"carousel_settings"`
}

// ListElement extends Element
type ListElement struct {
    Element
    Elements []any `json:"elements" db:"-"`
}

// SelectElement extends Element
type SelectElement struct {
    Element
    Options  []map[string]any `json:"options" db:"options"`
    Settings map[string]any   `json:"settings,omitempty" db:"carousel_settings"`
}

// ChartDataset for chart elements
type ChartDataset struct {
    Label           string      `json:"label"`
    Data            []float64   `json:"data"`
    BackgroundColor any `json:"backgroundColor,omitempty"`
    BorderColor     any `json:"borderColor,omitempty"`
    BorderWidth     *int        `json:"borderWidth,omitempty"`
    Fill            *bool       `json:"fill,omitempty"`
}


// FormElement extends Element
type FormElement struct {
    Element
    Elements []any           `json:"elements" db:"-"`
    Settings map[string]any  `json:"settings,omitempty" db:"carousel_settings"`
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
