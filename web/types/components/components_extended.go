// Package components - Extended component definitions for the remaining top 20 priority components
package components

import (
	"fmt"
)

// 9. Select Control Schema - Dropdown selections
type SelectControlSchema struct {
	BaseComponentProps
	Type             string         `json:"type"` // "select"
	Name             string         `json:"name"`
	Label            string         `json:"label,omitempty"`
	Value            any            `json:"value,omitempty"`
	Placeholder      string         `json:"placeholder,omitempty"`
	Size             string         `json:"size,omitempty"` // "xs", "sm", "md", "lg"
	Required         bool           `json:"required,omitempty"`
	ReadOnly         bool           `json:"readOnly,omitempty"`
	Multiple         bool           `json:"multiple,omitempty"`
	Searchable       bool           `json:"searchable,omitempty"`
	Clearable        bool           `json:"clearable,omitempty"`
	Options          []SelectOption `json:"options,omitempty"`
	Source           string         `json:"source,omitempty"`
	API              *APIConfig     `json:"api,omitempty"`
	AutoComplete     *APIConfig     `json:"autoComplete,omitempty"`
	CheckAll         bool           `json:"checkAll,omitempty"`
	CheckAllLabel    string         `json:"checkAllLabel,omitempty"`
	DefaultCheckAll  bool           `json:"defaultCheckAll,omitempty"`
	ExtractValue     bool           `json:"extractValue,omitempty"`
	JoinValues       bool           `json:"joinValues,omitempty"`
	Delimiter        string         `json:"delimiter,omitempty"`
	LabelField       string         `json:"labelField,omitempty"`
	ValueField       string         `json:"valueField,omitempty"`
	IconField        string         `json:"iconField,omitempty"`
	DeferField       string         `json:"deferField,omitempty"`
	MenuTpl          string         `json:"menuTpl,omitempty"`
	CreateBtnLabel   string         `json:"createBtnLabel,omitempty"`
	AddControls      []any          `json:"addControls,omitempty"`
	AddAPI           *APIConfig     `json:"addApi,omitempty"`
	EditControls     []any          `json:"editControls,omitempty"`
	EditAPI          *APIConfig     `json:"editApi,omitempty"`
	RemovableOn      string         `json:"removableOn,omitempty"`
	EditableOn       string         `json:"editableOn,omitempty"`
	OptionClassName  string         `json:"optionClassName,omitempty"`
	PopOverContainer string         `json:"popOverContainer,omitempty"`
	Overlay          *PopOverConfig `json:"overlay,omitempty"`
}

// SelectOption represents an option in a select control
type SelectOption struct {
	Label       string         `json:"label"`
	Value       any            `json:"value"`
	Icon        string         `json:"icon,omitempty"`
	Image       string         `json:"image,omitempty"`
	Disabled    bool           `json:"disabled,omitempty"`
	Children    []SelectOption `json:"children,omitempty"`
	Defer       bool           `json:"defer,omitempty"`
	Hidden      bool           `json:"hidden,omitempty"`
	Description string         `json:"description,omitempty"`
}

// 10. Date Control Schema - Date inputs for transactions
type DateControlSchema struct {
	BaseComponentProps
	Type             string            `json:"type"` // "input-date"
	Name             string            `json:"name"`
	Label            string            `json:"label,omitempty"`
	Value            string            `json:"value,omitempty"`
	Placeholder      string            `json:"placeholder,omitempty"`
	Size             string            `json:"size,omitempty"` // "xs", "sm", "md", "lg"
	Required         bool              `json:"required,omitempty"`
	ReadOnly         bool              `json:"readOnly,omitempty"`
	Format           string            `json:"format,omitempty"`
	InputFormat      string            `json:"inputFormat,omitempty"`
	DisplayFormat    string            `json:"displayFormat,omitempty"`
	ValueFormat      string            `json:"valueFormat,omitempty"`
	CloseOnSelect    bool              `json:"closeOnSelect,omitempty"`
	TimeConstraints  any               `json:"timeConstraints,omitempty"`
	MinDate          string            `json:"minDate,omitempty"`
	MaxDate          string            `json:"maxDate,omitempty"`
	DisabledDate     string            `json:"disabledDate,omitempty"`
	Clearable        bool              `json:"clearable,omitempty"`
	EmbedMode        bool              `json:"embedMode,omitempty"`
	Transform        map[string]string `json:"transform,omitempty"`
	BorderMode       string            `json:"borderMode,omitempty"`
	Shortcuts        []DateShortcut    `json:"shortcuts,omitempty"`
	UTC              bool              `json:"utc,omitempty"`
	PopOverContainer string            `json:"popOverContainer,omitempty"`
}

// DateShortcut for date picker shortcuts
type DateShortcut struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// 11. Date Range Control Schema - Date range for reporting
type DateRangeControlSchema struct {
	BaseComponentProps
	Type             string            `json:"type"` // "input-date-range"
	Name             string            `json:"name"`
	Label            string            `json:"label,omitempty"`
	Value            string            `json:"value,omitempty"`
	Placeholder      string            `json:"placeholder,omitempty"`
	Size             string            `json:"size,omitempty"` // "xs", "sm", "md", "lg"
	Required         bool              `json:"required,omitempty"`
	ReadOnly         bool              `json:"readOnly,omitempty"`
	Format           string            `json:"format,omitempty"`
	InputFormat      string            `json:"inputFormat,omitempty"`
	DisplayFormat    string            `json:"displayFormat,omitempty"`
	ValueFormat      string            `json:"valueFormat,omitempty"`
	Delimiter        string            `json:"delimiter,omitempty"`
	MinDate          string            `json:"minDate,omitempty"`
	MaxDate          string            `json:"maxDate,omitempty"`
	MinDuration      int               `json:"minDuration,omitempty"`
	MaxDuration      int               `json:"maxDuration,omitempty"`
	Clearable        bool              `json:"clearable,omitempty"`
	EmbedMode        bool              `json:"embedMode,omitempty"`
	Transform        map[string]string `json:"transform,omitempty"`
	BorderMode       string            `json:"borderMode,omitempty"`
	Shortcuts        []DateShortcut    `json:"shortcuts,omitempty"`
	UTC              bool              `json:"utc,omitempty"`
	PopOverContainer string            `json:"popOverContainer,omitempty"`
}

// 12. Navigation Schema - Menu navigation
type NavSchema struct {
	BaseComponentProps
	Type             string         `json:"type"` // "nav"
	Source           string         `json:"source,omitempty"`
	Links            []NavItem      `json:"links,omitempty"`
	Stacked          bool           `json:"stacked,omitempty"`
	Mode             string         `json:"mode,omitempty"` // "inline", "float"
	Level            int            `json:"level,omitempty"`
	Accordion        bool           `json:"accordion,omitempty"`
	Draggable        bool           `json:"draggable,omitempty"`
	SaveOrderAPI     *APIConfig     `json:"saveOrderApi,omitempty"`
	ItemActions      []ActionSchema `json:"itemActions,omitempty"`
	DefaultOpenLevel int            `json:"defaultOpenLevel,omitempty"`
	ThemeColor       string         `json:"themeColor,omitempty"`
}

// NavItem represents a navigation item
type NavItem struct {
	Label         string     `json:"label"`
	To            string     `json:"to,omitempty"`
	Target        string     `json:"target,omitempty"`
	Icon          string     `json:"icon,omitempty"`
	Children      []NavItem  `json:"children,omitempty"`
	Unfolded      bool       `json:"unfolded,omitempty"`
	Active        bool       `json:"active,omitempty"`
	ActiveOn      string     `json:"activeOn,omitempty"`
	Defer         bool       `json:"defer,omitempty"`
	DeferAPI      *APIConfig `json:"deferApi,omitempty"`
	Disabled      bool       `json:"disabled,omitempty"`
	DisabledOn    string     `json:"disabledOn,omitempty"`
	Hidden        bool       `json:"hidden,omitempty"`
	HiddenOn      string     `json:"hiddenOn,omitempty"`
	ClassName     string     `json:"className,omitempty"`
	ItemClassName string     `json:"itemClassName,omitempty"`
}

// 13. Alert Schema - Status notifications
type AlertSchema struct {
	BaseComponentProps
	Type            string `json:"type"`            // "alert"
	Level           string `json:"level,omitempty"` // "info", "success", "warning", "danger"
	Title           string `json:"title,omitempty"`
	Body            string `json:"body,omitempty"`
	Closable        bool   `json:"closable,omitempty"`
	CloseText       string `json:"closeText,omitempty"`
	ShowCloseButton bool   `json:"showCloseButton,omitempty"`
	Icon            string `json:"icon,omitempty"`
	ShowIcon        bool   `json:"showIcon,omitempty"`
	IconClassName   string `json:"iconClassName,omitempty"`
}

// 14. Pagination Schema - Data pagination
type PaginationSchema struct {
	BaseComponentProps
	Type             string   `json:"type"`           // "pagination"
	Mode             string   `json:"mode,omitempty"` // "normal", "simple"
	Layout           []string `json:"layout,omitempty"`
	MaxButtons       int      `json:"maxButtons,omitempty"`
	Total            int      `json:"total,omitempty"`
	PerPage          int      `json:"perPage,omitempty"`
	PerPageAvailable []int    `json:"perPageAvailable,omitempty"`
	ShowPerPage      bool     `json:"showPerPage,omitempty"`
	PerPageText      string   `json:"perPageText,omitempty"`
	ShowPageInput    bool     `json:"showPageInput,omitempty"`
	ShowQuickJumper  bool     `json:"showQuickJumper,omitempty"`
	Size             string   `json:"size,omitempty"`  // "sm", "md", "lg"
	Align            string   `json:"align,omitempty"` // "left", "center", "right"
	OnPageChange     string   `json:"onPageChange,omitempty"`
}

// 15. Search Box Schema - Search functionality
type SearchBoxSchema struct {
	BaseComponentProps
	Type          string     `json:"type"` // "search-box"
	Name          string     `json:"name"`
	Label         string     `json:"label,omitempty"`
	Placeholder   string     `json:"placeholder,omitempty"`
	Mini          bool       `json:"mini,omitempty"`
	SearchAPI     *APIConfig `json:"searchApi,omitempty"`
	Clearable     bool       `json:"clearable,omitempty"`
	Enhanced      bool       `json:"enhanced,omitempty"`
	Loading       bool       `json:"loading,omitempty"`
	LoadingConfig any        `json:"loadingConfig,omitempty"`
	OnSearch      string     `json:"onSearch,omitempty"`
	OnClear       string     `json:"onClear,omitempty"`
	OnCancel      string     `json:"onCancel,omitempty"`
}

// 16. Tabs Schema - Tabbed interfaces
type TabsSchema struct {
	BaseComponentProps
	Type             string    `json:"type"`               // "tabs"
	Mode             string    `json:"mode,omitempty"`     // "line", "card", "radio", "vertical", "chrome", "simple"
	TabsMode         string    `json:"tabsMode,omitempty"` // "line", "card", "radio"
	Tabs             []TabItem `json:"tabs,omitempty"`
	Source           string    `json:"source,omitempty"`
	TabClassName     string    `json:"tabClassName,omitempty"`
	ContentClassName string    `json:"contentClassName,omitempty"`
	LinksClassName   string    `json:"linksClassName,omitempty"`
	Mountable        bool      `json:"mountable,omitempty"`
	Unmountable      bool      `json:"unmountable,omitempty"`
	AddBtnText       string    `json:"addBtnText,omitempty"`
	Addable          bool      `json:"addable,omitempty"`
	Closable         bool      `json:"closable,omitempty"`
	Draggable        bool      `json:"draggable,omitempty"`
	ShowTip          bool      `json:"showTip,omitempty"`
	ShowTipClassName string    `json:"showTipClassName,omitempty"`
	EdgeMode         bool      `json:"edgeMode,omitempty"`
}

// TabItem represents a tab item
type TabItem struct {
	Title         string `json:"title"`
	Tab           string `json:"tab,omitempty"`
	Hash          string `json:"hash,omitempty"`
	Icon          string `json:"icon,omitempty"`
	IconPosition  string `json:"iconPosition,omitempty"`
	Body          []any  `json:"body,omitempty"`
	Disabled      bool   `json:"disabled,omitempty"`
	DisabledOn    string `json:"disabledOn,omitempty"`
	Hidden        bool   `json:"hidden,omitempty"`
	HiddenOn      string `json:"hiddenOn,omitempty"`
	ClassName     string `json:"className,omitempty"`
	UnmountOnExit bool   `json:"unmountOnExit,omitempty"`
	Reload        bool   `json:"reload,omitempty"`
	Closable      bool   `json:"closable,omitempty"`
	Tip           string `json:"tip,omitempty"`
}

// 17. File Control Schema - File uploads (moved to separate file: file_control.go)

// 18. Checkbox Control Schema - Boolean selections
type CheckboxControlSchema struct {
	BaseComponentProps
	Type           string       `json:"type"` // "checkbox"
	Name           string       `json:"name"`
	Label          string       `json:"label,omitempty"`
	Option         string       `json:"option,omitempty"`
	Text           string       `json:"text,omitempty"`
	TrueValue      any          `json:"trueValue,omitempty"`
	FalseValue     any          `json:"falseValue,omitempty"`
	Checked        bool         `json:"checked,omitempty"`
	Partial        bool         `json:"partial,omitempty"`
	CheckedOn      string       `json:"checkedOn,omitempty"`
	Size           string       `json:"size,omitempty"` // "sm", "md", "lg"
	Badge          *BadgeConfig `json:"badge,omitempty"`
	InputClassName string       `json:"inputClassName,omitempty"`
}

// 19. Panel Schema - Content organization
type PanelSchema struct {
	BaseComponentProps
	Type              string          `json:"type"` // "panel"
	Title             string          `json:"title,omitempty"`
	Header            any             `json:"header,omitempty"`
	Body              []any           `json:"body,omitempty"`
	Footer            any             `json:"footer,omitempty"`
	Actions           []ActionSchema  `json:"actions,omitempty"`
	HeaderClassName   string          `json:"headerClassName,omitempty"`
	FooterClassName   string          `json:"footerClassName,omitempty"`
	ActionsClassName  string          `json:"actionsClassName,omitempty"`
	Collapsable       bool            `json:"collapsable,omitempty"`
	Collapsed         bool            `json:"collapsed,omitempty"`
	CollapsedOn       string          `json:"collapsedOn,omitempty"`
	SubFormMode       string          `json:"subFormMode,omitempty"`
	SubFormHorizontal *FormHorizontal `json:"subFormHorizontal,omitempty"`
}

// 20. Status Schema - Record status indicators
type StatusSchema struct {
	BaseComponentProps
	Type        string                   `json:"type"` // "status"
	Map         map[string]StatusMapItem `json:"map,omitempty"`
	LabelMap    map[string]string        `json:"labelMap,omitempty"`
	Source      string                   `json:"source,omitempty"`
	Placeholder string                   `json:"placeholder,omitempty"`
}

// StatusMapItem represents a status mapping
type StatusMapItem struct {
	Label     string `json:"label"`
	Color     string `json:"color,omitempty"`
	Icon      string `json:"icon,omitempty"`
	ClassName string `json:"className,omitempty"`
}

// Extended factory methods for new components
func (f *DefaultComponentFactory) CreateSelect(config SelectControlSchema) (*SelectControlSchema, error) {
	if config.Type == "" {
		config.Type = "select"
	}
	if config.Size == "" {
		config.Size = "md"
	}
	if config.LabelField == "" {
		config.LabelField = "label"
	}
	if config.ValueField == "" {
		config.ValueField = "value"
	}
	if config.Delimiter == "" && config.Multiple {
		config.Delimiter = ","
	}

	return &config, nil
}

func (f *DefaultComponentFactory) CreateDateControl(config DateControlSchema) (*DateControlSchema, error) {
	if config.Type == "" {
		config.Type = "input-date"
	}
	if config.Size == "" {
		config.Size = "md"
	}
	if config.Format == "" {
		config.Format = "DD-MM-YYYY"
	}
	if config.Placeholder == "" {
		config.Placeholder = "Select date"
	}

	return &config, nil
}

func (f *DefaultComponentFactory) CreateDateRange(config DateRangeControlSchema) (*DateRangeControlSchema, error) {
	if config.Type == "" {
		config.Type = "input-date-range"
	}
	if config.Size == "" {
		config.Size = "md"
	}
	if config.Format == "" {
		config.Format = "DD-MM-YYYY"
	}
	if config.Delimiter == "" {
		config.Delimiter = ","
	}
	if config.Placeholder == "" {
		config.Placeholder = "Select date range"
	}

	return &config, nil
}

func (f *DefaultComponentFactory) CreateNav(config NavSchema) (*NavSchema, error) {
	if config.Type == "" {
		config.Type = "nav"
	}
	if config.Mode == "" {
		config.Mode = "inline"
	}
	config.Stacked = true

	return &config, nil
}

func (f *DefaultComponentFactory) CreateAlert(config AlertSchema) (*AlertSchema, error) {
	if config.Type == "" {
		config.Type = "alert"
	}
	if config.Level == "" {
		config.Level = "info"
	}
	config.ShowIcon = true
	config.Closable = true

	return &config, nil
}

func (f *DefaultComponentFactory) CreatePagination(config PaginationSchema) (*PaginationSchema, error) {
	if config.Type == "" {
		config.Type = "pagination"
	}
	if config.Mode == "" {
		config.Mode = "normal"
	}
	if config.MaxButtons == 0 {
		config.MaxButtons = 7
	}
	if config.PerPage == 0 {
		config.PerPage = 10
	}
	if config.PerPageAvailable == nil {
		config.PerPageAvailable = []int{10, 20, 50, 100}
	}
	config.ShowPerPage = true

	return &config, nil
}

func (f *DefaultComponentFactory) CreateSearchBox(config SearchBoxSchema) (*SearchBoxSchema, error) {
	if config.Type == "" {
		config.Type = "search-box"
	}
	if config.Placeholder == "" {
		config.Placeholder = "Enter keywords..."
	}
	config.Clearable = true
	config.Enhanced = true

	return &config, nil
}

func (f *DefaultComponentFactory) CreateTabs(config TabsSchema) (*TabsSchema, error) {
	if config.Type == "" {
		config.Type = "tabs"
	}
	if config.Mode == "" {
		config.Mode = "line"
	}
	config.Mountable = true
	config.Unmountable = true

	return &config, nil
}

// CreateFileControl moved to file_control.go

func (f *DefaultComponentFactory) CreateCheckbox(config CheckboxControlSchema) (*CheckboxControlSchema, error) {
	if config.Type == "" {
		config.Type = "checkbox"
	}
	if config.Size == "" {
		config.Size = "md"
	}
	if config.TrueValue == nil {
		config.TrueValue = true
	}
	if config.FalseValue == nil {
		config.FalseValue = false
	}

	return &config, nil
}

func (f *DefaultComponentFactory) CreatePanel(config PanelSchema) (*PanelSchema, error) {
	if config.Type == "" {
		config.Type = "panel"
	}
	config.Collapsable = false
	config.Collapsed = false

	return &config, nil
}

func (f *DefaultComponentFactory) CreateStatus(config StatusSchema) (*StatusSchema, error) {
	if config.Type == "" {
		config.Type = "status"
	}
	if config.Placeholder == "" {
		config.Placeholder = "-"
	}

	return &config, nil
}

// Extended ComponentFactory interface for new components
type ExtendedComponentFactory interface {
	ComponentFactory
	CreateSelect(config SelectControlSchema) (*SelectControlSchema, error)
	CreateDateControl(config DateControlSchema) (*DateControlSchema, error)
	CreateDateRange(config DateRangeControlSchema) (*DateRangeControlSchema, error)
	CreateNav(config NavSchema) (*NavSchema, error)
	CreateAlert(config AlertSchema) (*AlertSchema, error)
	CreatePagination(config PaginationSchema) (*PaginationSchema, error)
	CreateSearchBox(config SearchBoxSchema) (*SearchBoxSchema, error)
	CreateTabs(config TabsSchema) (*TabsSchema, error)
	// CreateFileControl moved to file_control.go
	CreateCheckbox(config CheckboxControlSchema) (*CheckboxControlSchema, error)
	CreatePanel(config PanelSchema) (*PanelSchema, error)
	CreateStatus(config StatusSchema) (*StatusSchema, error)
}

// Ensure DefaultComponentFactory implements ExtendedComponentFactory
var _ ExtendedComponentFactory = (*DefaultComponentFactory)(nil)

// Extended validation functions
func validateExtendedComponent(component any) error {
	switch c := component.(type) {
	case *SelectControlSchema:
		return validateSelect(c)
	case *DateControlSchema:
		return validateDateControl(c)
	case *DateRangeControlSchema:
		return validateDateRange(c)
	case *NavSchema:
		return validateNav(c)
	case *AlertSchema:
		return validateAlert(c)
	case *PaginationSchema:
		return validatePagination(c)
	case *SearchBoxSchema:
		return validateSearchBox(c)
	case *TabsSchema:
		return validateTabs(c)
	// FileControlSchema validation moved to file_control.go
	case *CheckboxControlSchema:
		return validateCheckbox(c)
	case *PanelSchema:
		return validatePanel(c)
	case *StatusSchema:
		return validateStatus(c)
	default:
		return ValidateComponent(component) // Fallback to base validation
	}
}

func validateSelect(s *SelectControlSchema) error {
	if s.Type != "select" {
		return fmt.Errorf("invalid select type: %s", s.Type)
	}
	if s.Name == "" {
		return fmt.Errorf("select control name is required")
	}
	return nil
}

func validateDateControl(d *DateControlSchema) error {
	if d.Type != "input-date" {
		return fmt.Errorf("invalid date control type: %s", d.Type)
	}
	if d.Name == "" {
		return fmt.Errorf("date control name is required")
	}
	return nil
}

func validateDateRange(d *DateRangeControlSchema) error {
	if d.Type != "input-date-range" {
		return fmt.Errorf("invalid date range control type: %s", d.Type)
	}
	if d.Name == "" {
		return fmt.Errorf("date range control name is required")
	}
	return nil
}

func validateNav(n *NavSchema) error {
	if n.Type != "nav" {
		return fmt.Errorf("invalid nav type: %s", n.Type)
	}
	return nil
}

func validateAlert(a *AlertSchema) error {
	if a.Type != "alert" {
		return fmt.Errorf("invalid alert type: %s", a.Type)
	}
	return nil
}

func validatePagination(p *PaginationSchema) error {
	if p.Type != "pagination" {
		return fmt.Errorf("invalid pagination type: %s", p.Type)
	}
	return nil
}

func validateSearchBox(s *SearchBoxSchema) error {
	if s.Type != "search-box" {
		return fmt.Errorf("invalid search box type: %s", s.Type)
	}
	if s.Name == "" {
		return fmt.Errorf("search box name is required")
	}
	return nil
}

func validateTabs(t *TabsSchema) error {
	if t.Type != "tabs" {
		return fmt.Errorf("invalid tabs type: %s", t.Type)
	}
	return nil
}

// validateFileControl moved to file_control.go

func validateCheckbox(c *CheckboxControlSchema) error {
	if c.Type != "checkbox" {
		return fmt.Errorf("invalid checkbox type: %s", c.Type)
	}
	if c.Name == "" {
		return fmt.Errorf("checkbox name is required")
	}
	return nil
}

func validatePanel(p *PanelSchema) error {
	if p.Type != "panel" {
		return fmt.Errorf("invalid panel type: %s", p.Type)
	}
	return nil
}

func validateStatus(s *StatusSchema) error {
	if s.Type != "status" {
		return fmt.Errorf("invalid status type: %s", s.Type)
	}
	return nil
}

// ExtendedComponentRegistry with support for all 20 priority components
type ExtendedComponentRegistry struct {
	*ComponentRegistry
	extendedFactory ExtendedComponentFactory
}

// NewExtendedComponentRegistry creates a registry with all 20 components
func NewExtendedComponentRegistry() *ExtendedComponentRegistry {
	return &ExtendedComponentRegistry{
		ComponentRegistry: NewComponentRegistry(),
		extendedFactory:   &DefaultComponentFactory{},
	}
}

// ExtendedFactory returns the extended component factory
func (r *ExtendedComponentRegistry) ExtendedFactory() ExtendedComponentFactory {
	return r.extendedFactory
}

// RegisterExtended registers a component with extended validation
func (r *ExtendedComponentRegistry) RegisterExtended(name string, component any) error {
	if err := validateExtendedComponent(component); err != nil {
		return fmt.Errorf("extended validation failed for component %s: %w", name, err)
	}
	r.components[name] = component
	return nil
}
