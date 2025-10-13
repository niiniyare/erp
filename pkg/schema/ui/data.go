package ui

import (
	"context"
	"encoding/json"
	"fmt"
)

// Table component configuration
type TableConfig struct {
	Columns      []TableColumn `json:"columns" validate:"required"`
	DataSource   string        `json:"data_source,omitempty"`
	Data         []any `json:"data,omitempty"`
	Pagination   *Pagination   `json:"pagination,omitempty"`
	Sorting      *Sorting      `json:"sorting,omitempty"`
	Filtering    *Filtering    `json:"filtering,omitempty"`
	Selection    *Selection    `json:"selection,omitempty"`
	Search       *Search       `json:"search,omitempty"`
	Actions      []Action      `json:"actions,omitempty"`
	RowActions   []Action      `json:"row_actions,omitempty"`
	Loading      bool          `json:"loading,omitempty"`
	Striped      bool          `json:"striped,omitempty"`
	Bordered     bool          `json:"bordered,omitempty"`
	Hoverable    bool          `json:"hoverable,omitempty"`
	Compact      bool          `json:"compact,omitempty"`
	Responsive   bool          `json:"responsive,omitempty"`
	EmptyMessage string        `json:"empty_message,omitempty"`
}

// TableColumn represents a table column definition
type TableColumn struct {
	Key        string      `json:"key" validate:"required"`
	Title      string      `json:"title" validate:"required"`
	DataType   DataType    `json:"data_type,omitempty"`
	Width      string      `json:"width,omitempty"`
	MinWidth   string      `json:"min_width,omitempty"`
	MaxWidth   string      `json:"max_width,omitempty"`
	Sortable   bool        `json:"sortable,omitempty"`
	Filterable bool        `json:"filterable,omitempty"`
	Searchable bool        `json:"searchable,omitempty"`
	Resizable  bool        `json:"resizable,omitempty"`
	Fixed      FixedColumn `json:"fixed,omitempty"`
	Align      Alignment   `json:"align,omitempty"`
	Format     string      `json:"format,omitempty"`
	Render     string      `json:"render,omitempty"`
	Hidden     bool        `json:"hidden,omitempty"`
	Ellipsis   bool        `json:"ellipsis,omitempty"`
}

// DataType is already defined in types.go - no need to duplicate

// FixedColumn represents column fixing options
type FixedColumn string

const (
	FixedLeft  FixedColumn = "left"
	FixedRight FixedColumn = "right"
)

// Alignment is already defined in types.go - no need to duplicate

// Pagination configuration
type Pagination struct {
	Page            int   `json:"page,omitempty"`
	PageSize        int   `json:"page_size,omitempty"`
	Total           int   `json:"total,omitempty"`
	ShowSizer       bool  `json:"show_sizer,omitempty"`
	ShowQuickJumper bool  `json:"show_quick_jumper,omitempty"`
	ShowTotal       bool  `json:"show_total,omitempty"`
	Simple          bool  `json:"simple,omitempty"`
	PageSizes       []int `json:"page_sizes,omitempty"`
}

// Sorting configuration
type Sorting struct {
	DefaultSort []SortOrder `json:"default_sort,omitempty"`
	Multiple    bool        `json:"multiple,omitempty"`
	Remote      bool        `json:"remote,omitempty"`
}

// SortOrder represents a sort configuration
type SortOrder struct {
	Column    string        `json:"column" validate:"required"`
	Direction SortDirection `json:"direction" validate:"required"`
}

// SortDirection represents sort directions
type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// Filtering configuration
type Filtering struct {
	Filters []Filter `json:"filters,omitempty"`
	Remote  bool     `json:"remote,omitempty"`
}

// Filter represents a column filter
type Filter struct {
	Column      string     `json:"column" validate:"required"`
	Type        FilterType `json:"type" validate:"required"`
	Options     []Option   `json:"options,omitempty"` // Option is defined in types.go
	Multiple    bool       `json:"multiple,omitempty"`
	SearchQuery string     `json:"search_query,omitempty"`
}

// FilterType represents different filter types
type FilterType string

const (
	FilterText        FilterType = "text"
	FilterSelect      FilterType = "select"
	FilterDate        FilterType = "date"
	FilterDateRange   FilterType = "date-range"
	FilterNumber      FilterType = "number"
	FilterNumberRange FilterType = "number-range"
	FilterBoolean     FilterType = "boolean"
)

// Selection configuration
type Selection struct {
	Type             SelectionType `json:"type,omitempty"`
	SelectedRows     []string      `json:"selected_rows,omitempty"`
	OnChange         string        `json:"on_change,omitempty"`
	GetCheckboxProps string        `json:"get_checkbox_props,omitempty"`
}

// SelectionType represents selection types
type SelectionType string

const (
	SelectionCheckbox SelectionType = "checkbox"
	SelectionRadio    SelectionType = "radio"
)

// Search configuration
type Search struct {
	Placeholder string   `json:"placeholder,omitempty"`
	Columns     []string `json:"columns,omitempty"`
	Remote      bool     `json:"remote,omitempty"`
	Debounce    int      `json:"debounce,omitempty"`
}

// Action, ActionType, and ConfirmDialog are already defined in types.go - no need to duplicate

// List component configuration
type ListConfig struct {
	DataSource   string        `json:"data_source,omitempty"`
	Data         []any `json:"data,omitempty"`
	ItemRender   string        `json:"item_render,omitempty"`
	Pagination   *Pagination   `json:"pagination,omitempty"`
	Loading      bool          `json:"loading,omitempty"`
	Split        bool          `json:"split,omitempty"`
	Bordered     bool          `json:"bordered,omitempty"`
	Header       *Component    `json:"header,omitempty"`
	Footer       *Component    `json:"footer,omitempty"`
	EmptyMessage string        `json:"empty_message,omitempty"`
}

// Tree component configuration
type TreeConfig struct {
	DataSource      string     `json:"data_source,omitempty"`
	Data            []TreeNode `json:"data,omitempty"`
	Checkable       bool       `json:"checkable,omitempty"`
	Selectable      bool       `json:"selectable,omitempty"`
	Multiple        bool       `json:"multiple,omitempty"`
	Expandable      bool       `json:"expandable,omitempty"`
	DefaultExpanded []string   `json:"default_expanded,omitempty"`
	ShowLine        bool       `json:"show_line,omitempty"`
	ShowIcon        bool       `json:"show_icon,omitempty"`
	Draggable       bool       `json:"draggable,omitempty"`
	VirtualScroll   bool       `json:"virtual_scroll,omitempty"`
	Height          string     `json:"height,omitempty"`
}

// TreeNode represents a tree node
type TreeNode struct {
	Key        string     `json:"key" validate:"required"`
	Title      string     `json:"title" validate:"required"`
	Icon       string     `json:"icon,omitempty"`
	Disabled   bool       `json:"disabled,omitempty"`
	Selectable bool       `json:"selectable,omitempty"`
	Checkable  bool       `json:"checkable,omitempty"`
	Children   []TreeNode `json:"children,omitempty"`
	IsLeaf     bool       `json:"is_leaf,omitempty"`
}

// Chart component configuration
type ChartConfig struct {
	Type       ChartType   `json:"type" validate:"required"`
	DataSource string      `json:"data_source,omitempty"`
	Data       any `json:"data,omitempty"`
	XAxis      *Axis       `json:"x_axis,omitempty"`
	YAxis      *Axis       `json:"y_axis,omitempty"`
	Legend     *Legend     `json:"legend,omitempty"`
	Tooltip    *Tooltip    `json:"tooltip,omitempty"`
	Colors     []string    `json:"colors,omitempty"`
	Width      string      `json:"width,omitempty"`
	Height     string      `json:"height,omitempty"`
	Animation  bool        `json:"animation,omitempty"`
	Responsive bool        `json:"responsive,omitempty"`
}

// ChartType represents different chart types
type ChartType string

const (
	ChartLine      ChartType = "line"
	ChartBar       ChartType = "bar"
	ChartPie       ChartType = "pie"
	ChartDoughnut  ChartType = "doughnut"
	ChartArea      ChartType = "area"
	ChartScatter   ChartType = "scatter"
	ChartRadar     ChartType = "radar"
	ChartPolarArea ChartType = "polar-area"
	ChartBubble    ChartType = "bubble"
)

// Axis configuration for charts
type Axis struct {
	Label     string   `json:"label,omitempty"`
	Type      AxisType `json:"type,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	Step      *float64 `json:"step,omitempty"`
	Format    string   `json:"format,omitempty"`
	Show      bool     `json:"show,omitempty"`
	GridLines bool     `json:"grid_lines,omitempty"`
}

// AxisType represents axis types
type AxisType string

const (
	AxisLinear   AxisType = "linear"
	AxisCategory AxisType = "category"
	AxisTime     AxisType = "time"
	AxisLog      AxisType = "logarithmic"
)

// Legend configuration
type Legend struct {
	Show     bool      `json:"show,omitempty"`
	Position Position  `json:"position,omitempty"`
	Align    Alignment `json:"align,omitempty"`
}

// Tooltip configuration
type Tooltip struct {
	Show      bool   `json:"show,omitempty"`
	Format    string `json:"format,omitempty"`
	Trigger   string `json:"trigger,omitempty"`
	Animation bool   `json:"animation,omitempty"`
}

// Badge component configuration
type BadgeConfig struct {
	Count    int         `json:"count,omitempty"`
	Text     string      `json:"text,omitempty"`
	ShowZero bool        `json:"show_zero,omitempty"`
	Dot      bool        `json:"dot,omitempty"`
	Status   BadgeStatus `json:"status,omitempty"`
	Color    string      `json:"color,omitempty"`
	Offset   []int       `json:"offset,omitempty"`
}

// BadgeStatus represents badge status types
type BadgeStatus string

const (
	BadgeStatusSuccess    BadgeStatus = "success"
	BadgeStatusProcessing BadgeStatus = "processing"
	BadgeStatusDefault    BadgeStatus = "default"
	BadgeStatusError      BadgeStatus = "error"
	BadgeStatusWarning    BadgeStatus = "warning"
)

// Tag component configuration
type TagConfig struct {
	Text      string `json:"text,omitempty"`
	Color     string `json:"color,omitempty"`
	Icon      string `json:"icon,omitempty"`
	Closable  bool   `json:"closable,omitempty"`
	OnClose   string `json:"on_close,omitempty"`
	CheckTime bool   `json:"checkable,omitempty"`
	Checked   bool   `json:"checked,omitempty"`
}

// Data display component factories
type (
	TableFactory struct{}
	ListFactory  struct{}
	TreeFactory  struct{}
	ChartFactory struct{}
	BadgeFactory struct{}
	TagFactory   struct{}
)

// Table factory implementation
func (f *TableFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var tableConfig TableConfig
	if err := mapToStruct(config, &tableConfig); err != nil {
		return Component{}, fmt.Errorf("invalid table config: %w", err)
	}

	component := NewComponent(ComponentTable, generateID())
	return component.WithConfig(tableConfig).Build(), nil
}

func (f *TableFactory) Validate(ctx context.Context, component Component) error {
	var tableConfig TableConfig
	if err := json.Unmarshal(component.Config, &tableConfig); err != nil {
		return fmt.Errorf("invalid table config: %w", err)
	}

	if len(tableConfig.Columns) == 0 {
		return fmt.Errorf("table must have at least one column")
	}

	if tableConfig.DataSource == "" && len(tableConfig.Data) == 0 {
		return fmt.Errorf("table must have data source or static data")
	}

	return nil
}

func (f *TableFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentTable,
		Title:       "Data Table",
		Description: "A table component for displaying structured data",
		Properties: map[string]Property{
			"columns": {
				Type:        "array",
				Description: "Table column definitions",
			},
			"data_source": {
				Type:        "string",
				Description: "API endpoint for dynamic data loading",
			},
			"pagination": {
				Type:        "object",
				Description: "Pagination configuration",
			},
			"sorting": {
				Type:        "object",
				Description: "Column sorting configuration",
			},
		},
		Required: []string{"columns"},
		Examples: []map[string]any{
			{
				"columns": []TableColumn{
					{Key: "name", Title: "Name", DataType: DataTypeText, Sortable: true},
					{Key: "email", Title: "Email", DataType: DataTypeText, Filterable: true},
				},
				"pagination": map[string]any{
					"page_size":  10,
					"show_sizer": true,
				},
			},
		},
	}
}

// Chart factory implementation
func (f *ChartFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var chartConfig ChartConfig
	if err := mapToStruct(config, &chartConfig); err != nil {
		return Component{}, fmt.Errorf("invalid chart config: %w", err)
	}

	component := NewComponent(ComponentChart, generateID())
	return component.WithConfig(chartConfig).Build(), nil
}

func (f *ChartFactory) Validate(ctx context.Context, component Component) error {
	var chartConfig ChartConfig
	if err := json.Unmarshal(component.Config, &chartConfig); err != nil {
		return fmt.Errorf("invalid chart config: %w", err)
	}

	if chartConfig.Type == "" {
		return fmt.Errorf("chart type is required")
	}

	if chartConfig.DataSource == "" && chartConfig.Data == nil {
		return fmt.Errorf("chart must have data source or static data")
	}

	return nil
}

func (f *ChartFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentChart,
		Title:       "Chart",
		Description: "A chart component for data visualization",
		Properties: map[string]Property{
			"type": {
				Type:        "string",
				Description: "Type of chart to display",
				Enum:        []string{"line", "bar", "pie", "doughnut", "area", "scatter"},
			},
			"data_source": {
				Type:        "string",
				Description: "API endpoint for chart data",
			},
			"responsive": {
				Type:        "boolean",
				Description: "Whether chart should be responsive",
				Default:     true,
			},
		},
		Required: []string{"type"},
		Examples: []map[string]any{
			{
				"type":        "line",
				"data_source": "/api/chart-data",
				"responsive":  true,
			},
		},
	}
}

// List factory implementation
func (f *ListFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var listConfig ListConfig
	if err := mapToStruct(config, &listConfig); err != nil {
		return Component{}, fmt.Errorf("invalid list config: %w", err)
	}
	component := NewComponent(ComponentList, generateID())
	return component.WithConfig(listConfig).Build(), nil
}

func (f *ListFactory) Validate(ctx context.Context, component Component) error {
	return nil
}

func (f *ListFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentList,
		Title:       "List",
		Description: "A list component for displaying data",
	}
}

// Tree factory implementation
func (f *TreeFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var treeConfig TreeConfig
	if err := mapToStruct(config, &treeConfig); err != nil {
		return Component{}, fmt.Errorf("invalid tree config: %w", err)
	}
	component := NewComponent(ComponentTree, generateID())
	return component.WithConfig(treeConfig).Build(), nil
}

func (f *TreeFactory) Validate(ctx context.Context, component Component) error {
	return nil
}

func (f *TreeFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentTree,
		Title:       "Tree",
		Description: "A hierarchical tree component",
	}
}

// Badge factory implementation
func (f *BadgeFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var badgeConfig BadgeConfig
	if err := mapToStruct(config, &badgeConfig); err != nil {
		return Component{}, fmt.Errorf("invalid badge config: %w", err)
	}
	component := NewComponent(ComponentBadge, generateID())
	return component.WithConfig(badgeConfig).Build(), nil
}

func (f *BadgeFactory) Validate(ctx context.Context, component Component) error {
	return nil
}

func (f *BadgeFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentBadge,
		Title:       "Badge",
		Description: "A badge component for status or count display",
	}
}

// Tag factory implementation
func (f *TagFactory) Create(ctx context.Context, config map[string]any) (Component, error) {
	var tagConfig TagConfig
	if err := mapToStruct(config, &tagConfig); err != nil {
		return Component{}, fmt.Errorf("invalid tag config: %w", err)
	}
	component := NewComponent(ComponentTag, generateID())
	return component.WithConfig(tagConfig).Build(), nil
}

func (f *TagFactory) Validate(ctx context.Context, component Component) error {
	return nil
}

func (f *TagFactory) GetSchema() ComponentSchema {
	return ComponentSchema{
		Type:        ComponentTag,
		Title:       "Tag",
		Description: "A tag component for labeling and categorization",
	}
}

