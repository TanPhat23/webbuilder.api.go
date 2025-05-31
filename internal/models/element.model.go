package models

type Element struct {
    Type           string                 `json:"type" db:"type"`
    ID             string                 `json:"id" db:"id"`
    Content        string                 `json:"content" db:"content"`
    IsSelected     bool                   `json:"isSelected" db:"is_selected"`
    Name           *string                `json:"name,omitempty" db:"name"`
    Styles         map[string]interface{} `json:"styles,omitempty" db:"styles"`
    TailwindStyles *string                `json:"tailwindStyles,omitempty" db:"tailwind_styles"`
    X              float64                `json:"x" db:"x"`
    Y              float64                `json:"y" db:"y"`
    Src            *string                `json:"src,omitempty" db:"src"`
    Href           *string                `json:"href,omitempty" db:"href"`
    ParentID       *string                `json:"parentId,omitempty" db:"parent_id"`
    ProjectID      *string                `json:"projectId,omitempty" db:"project_id"`
}

// FrameElement extends Element
type FrameElement struct {
    Element
    Elements []Element `json:"elements" db:"-"`
}

// CarouselElement extends Element
type CarouselElement struct {
    Element
    CarouselSettings map[string]interface{} `json:"carouselSettings" db:"carousel_settings"`
    Elements         []Element              `json:"elements" db:"-"`
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
    Elements []Element `json:"elements" db:"-"`
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

// ChartData for chart elements
type ChartData struct {
    Labels   []string       `json:"labels"`
    Datasets []ChartDataset `json:"datasets"`
}

// ChartElement extends Element
type ChartElement struct {
    Element
    ChartType    string                 `json:"chartType" db:"chart_type"`
    ChartData    ChartData              `json:"chartData" db:"chart_data"`
    ChartOptions map[string]interface{} `json:"chartOptions,omitempty" db:"chart_options"`
}

// TableSettings for data table elements
type TableSettings struct {
    Sortable    *bool `json:"sortable,omitempty"`
    Searchable  *bool `json:"searchable,omitempty"`
    Pagination  *bool `json:"pagination,omitempty"`
    RowsPerPage *int  `json:"rowsPerPage,omitempty"`
    Striped     *bool `json:"striped,omitempty"`
    Bordered    *bool `json:"bordered,omitempty"`
    HoverEffect *bool `json:"hoverEffect,omitempty"`
}

// DataTableElement extends Element
type DataTableElement struct {
    Element
    Headers       []string               `json:"headers" db:"headers"`
    Rows          [][]interface{}        `json:"rows" db:"rows"`
    TableSettings *TableSettings         `json:"tableSettings,omitempty" db:"table_settings"`
}

// FormElement extends Element
type FormElement struct {
    Element
    Elements     []Element              `json:"elements" db:"-"`
    FormSettings map[string]interface{} `json:"formSettings,omitempty" db:"form_settings"`
}

// EditorElement is a union type for all element types
type EditorElement interface {
    GetID() string
    GetType() string
}
	