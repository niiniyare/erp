// Package components implements Go types for the top 20 priority ERP component schemas
// This provides type-safe component definitions for building ERP interfaces with Templ integration
package components

import (
	"encoding/json"
	"fmt"
)

// BaseComponentProps contains common properties shared by all components
type BaseComponentProps struct {
	ID          string                 `json:"id,omitempty"`
	ClassName   string                 `json:"className,omitempty"`
	Style       map[string]any         `json:"style,omitempty"`
	Disabled    bool                   `json:"disabled,omitempty"`
	DisabledOn  string                 `json:"disabledOn,omitempty"`
	Hidden      bool                   `json:"hidden,omitempty"`
	HiddenOn    string                 `json:"hiddenOn,omitempty"`
	Visible     bool                   `json:"visible,omitempty"`
	VisibleOn   string                 `json:"visibleOn,omitempty"`
	TestID      string                 `json:"testid,omitempty"`
	OnEvent     map[string]EventConfig `json:"onEvent,omitempty"`
	UseMobileUI bool                   `json:"useMobileUI,omitempty"`
}

// EventConfig represents event handling configuration
type EventConfig struct {
	Weight   int             `json:"weight,omitempty"`
	Actions  []ActionSchema  `json:"actions"`
	Debounce *DebounceConfig `json:"debounce,omitempty"`
	Track    *TrackConfig    `json:"track,omitempty"`
}

// DebounceConfig for event debouncing
type DebounceConfig struct {
	Wait     int  `json:"wait"`
	MaxWait  int  `json:"maxWait,omitempty"`
	Leading  bool `json:"leading,omitempty"`
	Trailing bool `json:"trailing,omitempty"`
}

// TrackConfig for event tracking
type TrackConfig struct {
	ID   string         `json:"id"`
	Name string         `json:"name"`
	Data map[string]any `json:"data,omitempty"`
}

// APIConfig represents API configuration for data operations
type APIConfig struct {
	URL          string            `json:"url,omitempty"`
	Method       string            `json:"method,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	Data         map[string]any    `json:"data,omitempty"`
	DataType     string            `json:"dataType,omitempty"`
	ResponseType string            `json:"responseType,omitempty"`
	Cache        int               `json:"cache,omitempty"`
	QsOptions    map[string]any    `json:"qsOptions,omitempty"`
}

// FormHorizontal represents horizontal layout configuration
type FormHorizontal struct {
	Left       int          `json:"left"`
	Right      int          `json:"right"`
	Offset     int          `json:"offset,omitempty"`
	LeftFixed  *FixedConfig `json:"leftFixed,omitempty"`
	RightFixed *FixedConfig `json:"rightFixed,omitempty"`
}

// FixedConfig for fixed positioning
type FixedConfig struct {
	Value string `json:"value"`
	Unit  string `json:"unit,omitempty"`
}

// It been implemented validatio.go
// // ValidationRule represents form validation rules
//
//	type ValidationRule struct {
//		Rule    string   `json:"rule"`
//		Message string   `json:"message"`
//		Name    []string `json:"name,omitempty"`
//	}
//
// 1. CRUD Schema - Core data operations
type CRUDSchema struct {
	BaseComponentProps
	Type             string         `json:"type"`           // "crud"
	Mode             string         `json:"mode,omitempty"` // "table", "cards", "list"
	Title            string         `json:"title,omitempty"`
	API              *APIConfig     `json:"api,omitempty"`
	InitAPI          *APIConfig     `json:"initApi,omitempty"`
	SaveAPI          *APIConfig     `json:"saveApi,omitempty"`
	QuickSaveAPI     *APIConfig     `json:"quickSaveApi,omitempty"`
	QuickSaveItemAPI *APIConfig     `json:"quickSaveItemApi,omitempty"`
	BulkActions      []ActionSchema `json:"bulkActions,omitempty"`
	ItemActions      []ActionSchema `json:"itemActions,omitempty"`
	HeaderToolbar    []any          `json:"headerToolbar,omitempty"`
	FooterToolbar    []any          `json:"footerToolbar,omitempty"`
	Filter           *FormSchema    `json:"filter,omitempty"`
	Columns          []TableColumn  `json:"columns,omitempty"`
	PerPageAvailable []int          `json:"perPageAvailable,omitempty"`
	OrderBy          string         `json:"orderBy,omitempty"`
	OrderDir         string         `json:"orderDir,omitempty"` // "asc", "desc"
	DefaultParams    map[string]any `json:"defaultParams,omitempty"`
	PageField        string         `json:"pageField,omitempty"`
	PerPageField     string         `json:"perPageField,omitempty"`
	OrderField       string         `json:"orderField,omitempty"`
	Syncable         bool           `json:"syncable,omitempty"`
	Draggable        bool           `json:"draggable,omitempty"`
	LoadDataOnce     bool           `json:"loadDataOnce,omitempty"`
	Source           string         `json:"source,omitempty"`
}

// 2. Table Schema - Data tables
type TableSchema struct {
	BaseComponentProps
	Type               string         `json:"type"` // "table"
	Title              string         `json:"title,omitempty"`
	Source             string         `json:"source,omitempty"`
	Columns            []TableColumn  `json:"columns,omitempty"`
	Resizable          bool           `json:"resizable,omitempty"`
	Sortable           bool           `json:"sortable,omitempty"`
	Selectable         bool           `json:"selectable,omitempty"`
	Multiple           bool           `json:"multiple,omitempty"`
	FixedHeader        bool           `json:"fixedHeader,omitempty"`
	FixedColumns       bool           `json:"fixedColumns,omitempty"`
	ShowHeader         bool           `json:"showHeader,omitempty"`
	ShowFooter         bool           `json:"showFooter,omitempty"`
	AutoGenerateFilter bool           `json:"autoGenerateFilter,omitempty"`
	Sticky             bool           `json:"sticky,omitempty"`
	ItemActions        []ActionSchema `json:"itemActions,omitempty"`
	TableClassName     string         `json:"tableClassName,omitempty"`
	HeaderClassName    string         `json:"headerClassName,omitempty"`
	FooterClassName    string         `json:"footerClassName,omitempty"`
	ToolbarClassName   string         `json:"toolbarClassName,omitempty"`
	PrefixRow          []any          `json:"prefixRow,omitempty"`
	AffixRow           []any          `json:"affixRow,omitempty"`
	Placeholder        string         `json:"placeholder,omitempty"`
	ShowBadge          bool           `json:"showBadge,omitempty"`
	ExpandConfig       *ExpandConfig  `json:"expandConfig,omitempty"`
}

// TableColumn represents a table column definition
type TableColumn struct {
	Name           string           `json:"name,omitempty"`
	Label          string           `json:"label,omitempty"`
	Type           string           `json:"type,omitempty"`
	Tpl            string           `json:"tpl,omitempty"`
	Width          any              `json:"width,omitempty"` // number or string
	MinWidth       any              `json:"minWidth,omitempty"`
	MaxWidth       any              `json:"maxWidth,omitempty"`
	Fixed          string           `json:"fixed,omitempty"` // "left", "right"
	Sortable       bool             `json:"sortable,omitempty"`
	Searchable     bool             `json:"searchable,omitempty"`
	QuickEdit      *QuickEditConfig `json:"quickEdit,omitempty"`
	PopOver        *PopOverConfig   `json:"popOver,omitempty"`
	Copyable       *CopyableConfig  `json:"copyable,omitempty"`
	Align          string           `json:"align,omitempty"` // "left", "center", "right"
	ClassName      string           `json:"className,omitempty"`
	LabelClassName string           `json:"labelClassName,omitempty"`
	FilterMultiple bool             `json:"filterMultiple,omitempty"`
	Filters        []FilterOption   `json:"filters,omitempty"`
	Breakpoint     string           `json:"breakpoint,omitempty"`
	Remark         string           `json:"remark,omitempty"`
	Value          any              `json:"value,omitempty"`
}

// QuickEditConfig for inline editing
type QuickEditConfig struct {
	Type          string     `json:"type,omitempty"`
	Mode          string     `json:"mode,omitempty"`
	SaveAPI       *APIConfig `json:"saveApi,omitempty"`
	SaveImmediate bool       `json:"saveImmediate,omitempty"`
}

// PopOverConfig for popover display
type PopOverConfig struct {
	Mode     string `json:"mode,omitempty"` // "dialog", "drawer", "popOver"
	Title    string `json:"title,omitempty"`
	Body     any    `json:"body,omitempty"`
	Size     string `json:"size,omitempty"`
	Position string `json:"position,omitempty"`
}

// CopyableConfig for copyable content
type CopyableConfig struct {
	Content string `json:"content,omitempty"`
	Copy    string `json:"copy,omitempty"`
}

// FilterOption for column filtering
type FilterOption struct {
	Label string `json:"label"`
	Value any    `json:"value"`
}

// ExpandConfig for expandable rows
type ExpandConfig struct {
	Expand    any  `json:"expand,omitempty"`
	ExpandAll bool `json:"expandAll,omitempty"`
}

// 3. Form Schema - Data entry forms
type FormSchema struct {
	BaseComponentProps
	Type                        string           `json:"type"` // "form"
	Title                       string           `json:"title,omitempty"`
	Actions                     []ActionSchema   `json:"actions,omitempty"`
	Body                        []any            `json:"body,omitempty"`
	Data                        map[string]any   `json:"data,omitempty"`
	InitAPI                     *APIConfig       `json:"initApi,omitempty"`
	API                         *APIConfig       `json:"api,omitempty"`
	AsyncAPI                    *APIConfig       `json:"asyncApi,omitempty"`
	InitAsyncAPI                *APIConfig       `json:"initAsyncApi,omitempty"`
	CheckInterval               int              `json:"checkInterval,omitempty"`
	FinishedField               string           `json:"finishedField,omitempty"`
	InitFinishedField           string           `json:"initFinishedField,omitempty"`
	InitCheckInterval           int              `json:"initCheckInterval,omitempty"`
	InitFetch                   bool             `json:"initFetch,omitempty"`
	InitFetchOn                 string           `json:"initFetchOn,omitempty"`
	Interval                    int              `json:"interval,omitempty"`
	SilentPolling               bool             `json:"silentPolling,omitempty"`
	StopAutoRefreshWhen         string           `json:"stopAutoRefreshWhen,omitempty"`
	Mode                        string           `json:"mode,omitempty"` // "normal", "inline", "horizontal"
	Horizontal                  *FormHorizontal  `json:"horizontal,omitempty"`
	LabelAlign                  string           `json:"labelAlign,omitempty"` // "left", "center", "right"
	LabelWidth                  any              `json:"labelWidth,omitempty"` // number or string
	ColumnCount                 int              `json:"columnCount,omitempty"`
	AutoFocus                   bool             `json:"autoFocus,omitempty"`
	SubmitText                  string           `json:"submitText,omitempty"`
	SubmitOnChange              bool             `json:"submitOnChange,omitempty"`
	SubmitOnInit                bool             `json:"submitOnInit,omitempty"`
	ResetAfterSubmit            bool             `json:"resetAfterSubmit,omitempty"`
	ClearAfterSubmit            bool             `json:"clearAfterSubmit,omitempty"`
	Target                      string           `json:"target,omitempty"`
	Redirect                    string           `json:"redirect,omitempty"`
	Reload                      string           `json:"reload,omitempty"`
	Name                        string           `json:"name,omitempty"`
	PrimaryField                string           `json:"primaryField,omitempty"`
	Messages                    *FormMessages    `json:"messages,omitempty"`
	WrapWithPanel               bool             `json:"wrapWithPanel,omitempty"`
	PanelClassName              string           `json:"panelClassName,omitempty"`
	AffixFooter                 bool             `json:"affixFooter,omitempty"`
	Rules                       []ComponentValidationRule `json:"rules,omitempty"`
	PreventEnterSubmit          bool             `json:"preventEnterSubmit,omitempty"`
	PromptPageLeave             bool             `json:"promptPageLeave,omitempty"`
	PromptPageLeaveMessage      string           `json:"promptPageLeaveMessage,omitempty"`
	PersistData                 string           `json:"persistData,omitempty"`
	PersistDataKeys             []string         `json:"persistDataKeys,omitempty"`
	ClearPersistDataAfterSubmit bool             `json:"clearPersistDataAfterSubmit,omitempty"`
	Debug                       bool             `json:"debug,omitempty"`
	DebugConfig                 *DebugConfig     `json:"debugConfig,omitempty"`
}

// FormMessages for form message configuration
type FormMessages struct {
	ValidateFailed string `json:"validateFailed,omitempty"`
}

// DebugConfig for form debugging
type DebugConfig struct {
	LevelExpand       int    `json:"levelExpand,omitempty"`
	EnableClipboard   bool   `json:"enableClipboard,omitempty"`
	IconStyle         string `json:"iconStyle,omitempty"` // "square", "circle", "triangle"
	QuotesOnKeys      bool   `json:"quotesOnKeys,omitempty"`
	SortKeys          bool   `json:"sortKeys,omitempty"`
	EllipsisThreshold any    `json:"ellipsisThreshold,omitempty"` // number or false
}

// 4. Chart Schema - Analytics and KPI visualization
type ChartSchema struct {
	BaseComponentProps
	Type            string         `json:"type"` // "chart"
	API             *APIConfig     `json:"api,omitempty"`
	InitAPI         *APIConfig     `json:"initApi,omitempty"`
	Interval        int            `json:"interval,omitempty"`
	Config          map[string]any `json:"config,omitempty"`
	Style           map[string]any `json:"style,omitempty"`
	Width           any            `json:"width,omitempty"`  // number or string
	Height          any            `json:"height,omitempty"` // number or string
	Replacedata     bool           `json:"replacedata,omitempty"`
	TrackExpression string         `json:"trackExpression,omitempty"`
	DataFilter      string         `json:"dataFilter,omitempty"`
	Source          string         `json:"source,omitempty"`
	Name            string         `json:"name,omitempty"`
}

// 5. ActionSchema - User actions and buttons
type ActionSchema struct {
	BaseComponentProps
	Type             string         `json:"type,omitempty"`       // "action", "button", "submit", etc.
	ActionType       string         `json:"actionType,omitempty"` // "ajax", "link", "dialog", etc.
	Label            string         `json:"label,omitempty"`
	Icon             string         `json:"icon,omitempty"`
	RightIcon        string         `json:"rightIcon,omitempty"`
	Level            string         `json:"level,omitempty"` // "primary", "secondary", "success", etc.
	Size             string         `json:"size,omitempty"`  // "xs", "sm", "md", "lg"
	API              *APIConfig     `json:"api,omitempty"`
	Link             string         `json:"link,omitempty"`
	Dialog           *DialogSchema  `json:"dialog,omitempty"`
	Drawer           *DrawerSchema  `json:"drawer,omitempty"`
	Toast            *ToastConfig   `json:"toast,omitempty"`
	Required         []string       `json:"required,omitempty"`
	Args             map[string]any `json:"args,omitempty"`
	Script           string         `json:"script,omitempty"`
	Reload           string         `json:"reload,omitempty"`
	Target           string         `json:"target,omitempty"`
	Close            any            `json:"close,omitempty"` // bool or string
	Copy             string         `json:"copy,omitempty"`
	Payload          map[string]any `json:"payload,omitempty"`
	RequireSelected  bool           `json:"requireSelected,omitempty"`
	MergeData        bool           `json:"mergeData,omitempty"`
	ConfirmText      string         `json:"confirmText,omitempty"`
	DisabledTip      string         `json:"disabledTip,omitempty"`
	Block            bool           `json:"block,omitempty"`
	Loading          bool           `json:"loading,omitempty"`
	LoadingOn        string         `json:"loadingOn,omitempty"`
	Tooltip          string         `json:"tooltip,omitempty"`
	TooltipPlacement string         `json:"tooltipPlacement,omitempty"`
	Badge            *BadgeConfig   `json:"badge,omitempty"`
	HotKey           string         `json:"hotKey,omitempty"`
}

// ToastConfig for toast notifications
type ToastConfig struct {
	Items       []ToastItem `json:"items,omitempty"`
	Position    string      `json:"position,omitempty"`
	CloseButton bool        `json:"closeButton,omitempty"`
	ShowIcon    bool        `json:"showIcon,omitempty"`
	Timeout     int         `json:"timeout,omitempty"`
}

// ToastItem represents individual toast message
type ToastItem struct {
	Title   string `json:"title,omitempty"`
	Body    string `json:"body,omitempty"`
	Level   string `json:"level,omitempty"` // "info", "success", "error", "warning"
	Timeout int    `json:"timeout,omitempty"`
}

// BadgeConfig for button badges
type BadgeConfig struct {
	Text          string `json:"text,omitempty"`
	Level         string `json:"level,omitempty"`
	ClassName     string `json:"className,omitempty"`
	Offset        []int  `json:"offset,omitempty"`
	Position      string `json:"position,omitempty"`
	Overflowcount int    `json:"overflowcount,omitempty"`
	VisibleOn     string `json:"visibleOn,omitempty"`
}

// 6. Dialog Schema - Modal dialogs
type DialogSchema struct {
	BaseComponentProps
	Type            string         `json:"type"` // "dialog"
	Title           string         `json:"title,omitempty"`
	Body            []any          `json:"body,omitempty"`
	Size            string         `json:"size,omitempty"` // "xs", "sm", "md", "lg", "xl", "full"
	CloseOnEsc      bool           `json:"closeOnEsc,omitempty"`
	CloseOnOutside  bool           `json:"closeOnOutside,omitempty"`
	ShowCloseButton bool           `json:"showCloseButton,omitempty"`
	ShowErrorMsg    bool           `json:"showErrorMsg,omitempty"`
	ShowLoading     bool           `json:"showLoading,omitempty"`
	Draggable       bool           `json:"draggable,omitempty"`
	Position        string         `json:"position,omitempty"`
	Resizable       bool           `json:"resizable,omitempty"`
	Overlay         bool           `json:"overlay,omitempty"`
	Actions         []ActionSchema `json:"actions,omitempty"`
	Data            map[string]any `json:"data,omitempty"`
	DataMapping     map[string]any `json:"dataMapping,omitempty"`
	HeaderClassName string         `json:"headerClassName,omitempty"`
	BodyClassName   string         `json:"bodyClassName,omitempty"`
	FooterClassName string         `json:"footerClassName,omitempty"`
	Name            string         `json:"name,omitempty"`
	DialogType      string         `json:"dialogType,omitempty"`
	LazyRender      bool           `json:"lazyRender,omitempty"`
}

// 7. DrawerSchema - Side drawer/panel
type DrawerSchema struct {
	BaseComponentProps
	Type            string         `json:"type"` // "drawer"
	Title           string         `json:"title,omitempty"`
	Body            []any          `json:"body,omitempty"`
	Size            string         `json:"size,omitempty"`     // "xs", "sm", "md", "lg", "xl"
	Position        string         `json:"position,omitempty"` // "left", "right", "top", "bottom"
	Overlay         bool           `json:"overlay,omitempty"`
	CloseOnEsc      bool           `json:"closeOnEsc,omitempty"`
	CloseOnOutside  bool           `json:"closeOnOutside,omitempty"`
	ShowCloseButton bool           `json:"showCloseButton,omitempty"`
	Resizable       bool           `json:"resizable,omitempty"`
	Actions         []ActionSchema `json:"actions,omitempty"`
	Data            map[string]any `json:"data,omitempty"`
	HeaderClassName string         `json:"headerClassName,omitempty"`
	BodyClassName   string         `json:"bodyClassName,omitempty"`
	FooterClassName string         `json:"footerClassName,omitempty"`
	Width           any            `json:"width,omitempty"`  // number or string
	Height          any            `json:"height,omitempty"` // number or string
	Name            string         `json:"name,omitempty"`
}

// Additional essential components for the top 20...

// 8. Text Control Schema - Basic text input
type TextControlSchema struct {
	BaseComponentProps
	Type               string            `json:"type"` // "input-text"
	Name               string            `json:"name"`
	Label              string            `json:"label,omitempty"`
	Value              string            `json:"value,omitempty"`
	Placeholder        string            `json:"placeholder,omitempty"`
	Size               string            `json:"size,omitempty"` // "xs", "sm", "md", "lg"
	Required           bool              `json:"required,omitempty"`
	ReadOnly           bool              `json:"readOnly,omitempty"`
	Multiple           bool              `json:"multiple,omitempty"`
	Trim               bool              `json:"trim,omitempty"`
	ClearValueOnHidden bool              `json:"clearValueOnHidden,omitempty"`
	ResetValue         string            `json:"resetValue,omitempty"`
	AddOn              *AddOnConfig      `json:"addOn,omitempty"`
	AutoComplete       string            `json:"autoComplete,omitempty"`
	BorderMode         string            `json:"borderMode,omitempty"`
	MinLength          int               `json:"minLength,omitempty"`
	MaxLength          int               `json:"maxLength,omitempty"`
	Counter            *CounterConfig    `json:"counter,omitempty"`
	Transform          map[string]string `json:"transform,omitempty"`
	ShowCounter        bool              `json:"showCounter,omitempty"`
	Prefix             string            `json:"prefix,omitempty"`
	Suffix             string            `json:"suffix,omitempty"`
	ValidateApi        *APIConfig        `json:"validateApi,omitempty"`
	ValidateOnChange   bool              `json:"validateOnChange,omitempty"`
	SubmitOnChange     bool              `json:"submitOnChange,omitempty"`
}

// AddOnConfig for input add-ons
type AddOnConfig struct {
	Type     string `json:"type,omitempty"`
	Label    string `json:"label,omitempty"`
	Icon     string `json:"icon,omitempty"`
	Position string `json:"position,omitempty"` // "left", "right"
}

// CounterConfig for character counter
type CounterConfig struct {
	Enable    bool   `json:"enable,omitempty"`
	Max       int    `json:"max,omitempty"`
	ClassName string `json:"className,omitempty"`
}

// ComponentFactory is a factory interface
type ComponentFactory interface {
	CreateCRUD(config CRUDSchema) (*CRUDSchema, error)
	CreateTable(config TableSchema) (*TableSchema, error)
	CreateForm(config FormSchema) (*FormSchema, error)
	CreateChart(config ChartSchema) (*ChartSchema, error)
	CreateAction(config ActionSchema) (*ActionSchema, error)
	CreateDialog(config DialogSchema) (*DialogSchema, error)
	CreateTextInput(config TextControlSchema) (*TextControlSchema, error)
}

// DefaultComponentFactory provides default implementations
type DefaultComponentFactory struct{}

// NewComponentFactory creates a new component factory
func NewComponentFactory() ComponentFactory {
	return &DefaultComponentFactory{}
}

// CreateCRUD creates a CRUD component with defaults
func (f *DefaultComponentFactory) CreateCRUD(config CRUDSchema) (*CRUDSchema, error) {
	if config.Type == "" {
		config.Type = "crud"
	}
	if config.Mode == "" {
		config.Mode = "table"
	}
	if config.PerPageAvailable == nil {
		config.PerPageAvailable = []int{10, 20, 50, 100}
	}
	if config.OrderDir == "" {
		config.OrderDir = "asc"
	}
	if config.PageField == "" {
		config.PageField = "page"
	}
	if config.PerPageField == "" {
		config.PerPageField = "perPage"
	}

	return &config, nil
}

// CreateTable creates a table component with defaults
func (f *DefaultComponentFactory) CreateTable(config TableSchema) (*TableSchema, error) {
	if config.Type == "" {
		config.Type = "table"
	}
	if config.ShowHeader == false && config.ShowHeader != true {
		config.ShowHeader = true // Default to showing header
	}
	if config.Placeholder == "" {
		config.Placeholder = "No data available"
	}

	return &config, nil
}

// CreateForm creates a form component with defaults
func (f *DefaultComponentFactory) CreateForm(config FormSchema) (*FormSchema, error) {
	if config.Type == "" {
		config.Type = "form"
	}
	if config.Mode == "" {
		config.Mode = "normal"
	}
	if config.LabelAlign == "" {
		config.LabelAlign = "left"
	}
	if config.SubmitText == "" {
		config.SubmitText = "Submit"
	}
	if config.CheckInterval == 0 {
		config.CheckInterval = 3000 // 3 seconds default
	}

	return &config, nil
}

// CreateChart creates a chart component with defaults
func (f *DefaultComponentFactory) CreateChart(config ChartSchema) (*ChartSchema, error) {
	if config.Type == "" {
		config.Type = "chart"
	}
	if config.Width == nil {
		config.Width = "100%"
	}
	if config.Height == nil {
		config.Height = 400
	}

	return &config, nil
}

// CreateAction creates an action component with defaults
func (f *DefaultComponentFactory) CreateAction(config ActionSchema) (*ActionSchema, error) {
	if config.Type == "" {
		config.Type = "action"
	}
	if config.Level == "" {
		config.Level = "primary"
	}
	if config.Size == "" {
		config.Size = "md"
	}

	return &config, nil
}

// CreateDialog creates a dialog component with defaults
func (f *DefaultComponentFactory) CreateDialog(config DialogSchema) (*DialogSchema, error) {
	if config.Type == "" {
		config.Type = "dialog"
	}
	if config.Size == "" {
		config.Size = "md"
	}
	config.CloseOnEsc = true
	config.ShowCloseButton = true
	config.Overlay = true

	return &config, nil
}

// CreateTextInput creates a text input component with defaults
func (f *DefaultComponentFactory) CreateTextInput(config TextControlSchema) (*TextControlSchema, error) {
	if config.Type == "" {
		config.Type = "input-text"
	}
	if config.Size == "" {
		config.Size = "md"
	}
	config.Trim = true

	return &config, nil
}

// Validation methods

// ValidateComponent validates a component configuration
func ValidateComponent(component any) error {
	switch c := component.(type) {
	case *CRUDSchema:
		return validateCRUD(c)
	case *TableSchema:
		return validateTable(c)
	case *FormSchema:
		return validateForm(c)
	case *ChartSchema:
		return validateChart(c)
	case *ActionSchema:
		return validateAction(c)
	case *DialogSchema:
		return validateDialog(c)
	case *TextControlSchema:
		return validateTextControl(c)
	default:
		return fmt.Errorf("unsupported component type: %T", component)
	}
}

func validateCRUD(c *CRUDSchema) error {
	if c.Type != "crud" {
		return fmt.Errorf("invalid CRUD type: %s", c.Type)
	}
	if c.Mode != "" && c.Mode != "table" && c.Mode != "cards" && c.Mode != "list" {
		return fmt.Errorf("invalid CRUD mode: %s", c.Mode)
	}
	return nil
}

func validateTable(t *TableSchema) error {
	if t.Type != "table" {
		return fmt.Errorf("invalid table type: %s", t.Type)
	}
	return nil
}

func validateForm(f *FormSchema) error {
	if f.Type != "form" {
		return fmt.Errorf("invalid form type: %s", f.Type)
	}
	if f.Mode != "" && f.Mode != "normal" && f.Mode != "inline" && f.Mode != "horizontal" {
		return fmt.Errorf("invalid form mode: %s", f.Mode)
	}
	return nil
}

func validateChart(c *ChartSchema) error {
	if c.Type != "chart" {
		return fmt.Errorf("invalid chart type: %s", c.Type)
	}
	return nil
}

func validateAction(a *ActionSchema) error {
	if a.Type == "" {
		a.Type = "action"
	}
	return nil
}

func validateDialog(d *DialogSchema) error {
	if d.Type != "dialog" {
		return fmt.Errorf("invalid dialog type: %s", d.Type)
	}
	return nil
}

func validateTextControl(t *TextControlSchema) error {
	if t.Type != "input-text" {
		return fmt.Errorf("invalid text control type: %s", t.Type)
	}
	if t.Name == "" {
		return fmt.Errorf("text control name is required")
	}
	return nil
}

// Helper functions for JSON serialization

// ToJSON converts a component to JSON string
func ToJSON(component any) (string, error) {
	data, err := json.MarshalIndent(component, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal component to JSON: %w", err)
	}
	return string(data), nil
}

// FromJSON converts JSON string to a component
func FromJSON(jsonData string, component any) error {
	err := json.Unmarshal([]byte(jsonData), component)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON to component: %w", err)
	}
	return nil
}

// Component registry for managing component types
type ComponentRegistry struct {
	components map[string]any
	factory    ComponentFactory
}

// NewComponentRegistry creates a new component registry
func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		components: make(map[string]any),
		factory:    NewComponentFactory(),
	}
}

// Register registers a component with the registry
func (r *ComponentRegistry) Register(name string, component any) error {
	if err := ValidateComponent(component); err != nil {
		return fmt.Errorf("validation failed for component %s: %w", name, err)
	}
	r.components[name] = component
	return nil
}

// Get retrieves a component from the registry
func (r *ComponentRegistry) Get(name string) (any, bool) {
	component, exists := r.components[name]
	return component, exists
}

// List returns all registered component names
func (r *ComponentRegistry) List() []string {
	names := make([]string, 0, len(r.components))
	for name := range r.components {
		names = append(names, name)
	}
	return names
}

// Factory returns the component factory
func (r *ComponentRegistry) Factory() ComponentFactory {
	return r.factory
}
