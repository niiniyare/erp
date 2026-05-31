package schema

import (
	"encoding/json"
	"fmt"
)

// ============================================================================
// ROOT SCHEMA
// ============================================================================

// SchemaObject is a generic container that can hold any Amis component
type SchemaObject struct {
	Type string          `json:"type"`
	Body json.RawMessage `json:"body,omitempty"`
}

// ============================================================================
// PAGE & CONTAINER COMPONENTS (Extended)
// ============================================================================

// Page root page component
type Page struct {
	Type        string         `json:"type"`
	Title       *string        `json:"title,omitempty"`
	Description *string        `json:"description,omitempty"`
	Body        any            `json:"body,omitempty"`
	Toolbar     []any          `json:"toolbar,omitempty"`
	Aside       any            `json:"aside,omitempty"`
	ClassName   *string        `json:"className,omitempty"`
	Style       map[string]any `json:"style,omitempty"`
}

// Container flexible container
type Container struct {
	Type      string         `json:"type"`
	Body      any            `json:"body,omitempty"`
	ClassName *string        `json:"className,omitempty"`
	Style     map[string]any `json:"style,omitempty"`
}

// Grid layout component
type Grid struct {
	Type      string       `json:"type"`
	Columns   []GridColumn `json:"columns,omitempty"`
	Gap       *string      `json:"gap,omitempty"`
	ClassName *string      `json:"className,omitempty"`
}

// GridColumn grid column
type GridColumn struct {
	Body      any            `json:"body,omitempty"`
	Width     *string        `json:"width,omitempty"`
	ClassName *string        `json:"className,omitempty"`
	Style     map[string]any `json:"style,omitempty"`
}

// HBox horizontal box layout
type HBox struct {
	Type      string       `json:"type"`
	Columns   []HBoxColumn `json:"columns,omitempty"`
	Gap       *string      `json:"gap,omitempty"`
	ClassName *string      `json:"className,omitempty"`
}

// HBoxColumn horizontal box column
type HBoxColumn struct {
	Body      any     `json:"body,omitempty"`
	Width     *string `json:"width,omitempty"`
	ClassName *string `json:"className,omitempty"`
}

// VBox vertical box layout
type VBox struct {
	Type      string  `json:"type"`
	Rows      []any   `json:"rows,omitempty"`
	Gap       *string `json:"gap,omitempty"`
	ClassName *string `json:"className,omitempty"`
}

// Flex flexible layout
type Flex struct {
	Type           string  `json:"type"`
	Items          []any   `json:"items,omitempty"`
	Direction      *string `json:"direction,omitempty"`
	JustifyContent *string `json:"justifyContent,omitempty"`
	AlignItems     *string `json:"alignItems,omitempty"`
	ClassName      *string `json:"className,omitempty"`
}

// Panel collapsible panel
type Panel struct {
	Type      string  `json:"type"`
	Title     string  `json:"title"`
	Body      any     `json:"body,omitempty"`
	Expanded  *bool   `json:"expanded,omitempty"`
	ClassName *string `json:"className,omitempty"`
	Icon      *string `json:"icon,omitempty"`
}

// Collapse collapse group
type Collapse struct {
	Type      string         `json:"type"`
	Panels    []CollapseItem `json:"items,omitempty"`
	ClassName *string        `json:"className,omitempty"`
	Accordion *bool          `json:"accordion,omitempty"`
}

// CollapseItem collapse panel item
type CollapseItem struct {
	Title string  `json:"title"`
	Body  any     `json:"body,omitempty"`
	Icon  *string `json:"icon,omitempty"`
}

// Card card component
type Card struct {
	Type     string   `json:"type"`
	Title    *string  `json:"title,omitempty"`
	SubTitle *string  `json:"subTitle,omitempty"`
	Desc     *string  `json:"desc,omitempty"`
	Body     any      `json:"body,omitempty"`
	Avatar   *string  `json:"avatar,omitempty"`
	Image    *string  `json:"image,omitempty"`
	Link     *string  `json:"link,omitempty"`
	Actions  []Action `json:"actions,omitempty"`
}

// Cards card list component
type Cards struct {
	Type        string       `json:"type"`
	CardOptions *CardOptions `json:"cardOptions,omitempty"`
	Data        []any        `json:"data,omitempty"`
	Source      *API         `json:"source,omitempty"`
	ClassName   *string      `json:"className,omitempty"`
}

// CardOptions card display options
type CardOptions struct {
	Title    *string `json:"title,omitempty"`
	SubTitle *string `json:"subTitle,omitempty"`
	Desc     *string `json:"desc,omitempty"`
	Avatar   *string `json:"avatar,omitempty"`
	Image    *string `json:"image,omitempty"`
	Link     *string `json:"link,omitempty"`
}

// Carousel carousel component
type Carousel struct {
	Type      string  `json:"type"`
	Items     []any   `json:"items,omitempty"`
	Auto      *bool   `json:"auto,omitempty"`
	Duration  *int    `json:"duration,omitempty"`
	ClassName *string `json:"className,omitempty"`
}

// ============================================================================
// TABS & NAVIGATION COMPONENTS
// ============================================================================

// Tabs tabbed panel component
type Tabs struct {
	Type      string      `json:"type"`
	Tabs      []TabConfig `json:"tabs,omitempty"`
	Active    *string     `json:"active,omitempty"`
	ClassName *string     `json:"className,omitempty"`
}

// TabConfig tab configuration
type TabConfig struct {
	Title    string  `json:"title"`
	Key      *string `json:"key,omitempty"`
	Body     any     `json:"body,omitempty"`
	Icon     *string `json:"icon,omitempty"`
	Closable *bool   `json:"closable,omitempty"`
}

// Steps step progress component
type Steps struct {
	Type      string       `json:"type"`
	Steps     []StepConfig `json:"steps,omitempty"`
	Active    *int         `json:"active,omitempty"`
	Direction *string      `json:"direction,omitempty"`
	ClassName *string      `json:"className,omitempty"`
}

// StepConfig step configuration
type StepConfig struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Status      *string `json:"status,omitempty"`
	Body        any     `json:"body,omitempty"`
}

// Nav navigation component
type Nav struct {
	Type      string    `json:"type"`
	Items     []NavItem `json:"items,omitempty"`
	Mode      *string   `json:"mode,omitempty"`
	ClassName *string   `json:"className,omitempty"`
}

// NavItem navigation item
type NavItem struct {
	Label    string    `json:"label"`
	Icon     *string   `json:"icon,omitempty"`
	Link     *string   `json:"link,omitempty"`
	Badge    *string   `json:"badge,omitempty"`
	Children []NavItem `json:"children,omitempty"`
}

// AnchorNav anchor navigation component
type AnchorNav struct {
	Type      string       `json:"type"`
	Links     []AnchorLink `json:"links,omitempty"`
	Direction *string      `json:"direction,omitempty"`
	ClassName *string      `json:"className,omitempty"`
}

// AnchorLink anchor link
type AnchorLink struct {
	Label  string `json:"label"`
	Target string `json:"target"`
	Body   any    `json:"body,omitempty"`
}

// ============================================================================
// FORM COMPONENTS (Extended)
// ============================================================================

// Form data input component
type Form struct {
	Type             string        `json:"type"`
	Title            *string       `json:"title,omitempty"`
	Body             any           `json:"body,omitempty"`
	SubmitText       *string       `json:"submitText,omitempty"`
	ResetText        *string       `json:"resetText,omitempty"`
	Layout           *string       `json:"layout,omitempty"`
	Mode             *string       `json:"mode,omitempty"`
	LabelWidth       *string       `json:"labelWidth,omitempty"`
	LabelAlign       *string       `json:"labelAlign,omitempty"`
	ClassName        *string       `json:"className,omitempty"`
	API              *API          `json:"api,omitempty"`
	InitAPI          *API          `json:"initApi,omitempty"`
	Redirect         *string       `json:"redirect,omitempty"`
	ResetAfterSubmit *bool         `json:"resetAfterSubmit,omitempty"`
	Messages         *FormMessages `json:"messages,omitempty"`
}

// FormControl base form field - compatible with FormControlSchema.json
type FormControl struct {
	// Identity & Display
	SchemaID       *string `json:"$id,omitempty"`
	Type           string  `json:"type"`
	Name           string  `json:"name"`
	Label          any     `json:"label,omitempty"` // string or false
	LabelAlign     *string `json:"labelAlign,omitempty"`
	LabelWidth     any     `json:"labelWidth,omitempty"` // number or string
	LabelClassName *string `json:"labelClassName,omitempty"`
	Placeholder    *string `json:"placeholder,omitempty"`
	ClassName      *string `json:"className,omitempty"`
	InputClassName *string `json:"inputClassName,omitempty"`

	// Values & Defaults
	Value        any `json:"value,omitempty"`
	DefaultValue any `json:"defaultValue,omitempty"`

	// States
	Required   *bool   `json:"required,omitempty"`
	Disabled   *bool   `json:"disabled,omitempty"`
	DisabledOn *string `json:"disabledOn,omitempty"` // expression
	Visible    *bool   `json:"visible,omitempty"`
	VisibleOn  *string `json:"visibleOn,omitempty"` // expression
	Hidden     *bool   `json:"hidden,omitempty"`
	HiddenOn   *string `json:"hiddenOn,omitempty"` // expression
	ReadOnly   *bool   `json:"readOnly,omitempty"`
	ReadOnlyOn *string `json:"readOnlyOn,omitempty"`

	// Static Display
	Static               *bool   `json:"static,omitempty"`
	StaticOn             *string `json:"staticOn,omitempty"`
	StaticPlaceholder    *string `json:"staticPlaceholder,omitempty"`
	StaticClassName      *string `json:"staticClassName,omitempty"`
	StaticLabelClassName *string `json:"staticLabelClassName,omitempty"`
	StaticInputClassName *string `json:"staticInputClassName,omitempty"`
	StaticSchema         any     `json:"staticSchema,omitempty"`

	// Description & Help
	Description          *string `json:"description,omitempty"`
	Desc                 *string `json:"desc,omitempty"`
	DescriptionClassName *string `json:"descriptionClassName,omitempty"`
	Remark               any     `json:"remark,omitempty"` // SchemaRemark
	LabelRemark          any     `json:"labelRemark,omitempty"`
	Hint                 *string `json:"hint,omitempty"`

	// Validation
	ValidationErrors map[string]string `json:"validationErrors,omitempty"`
	Validations      any               `json:"validations,omitempty"` // string or object
	ValidateOnChange *bool             `json:"validateOnChange,omitempty"`
	ValidateAPI      any               `json:"validateApi,omitempty"` // API config

	// Layout & Display
	Size       *string `json:"size,omitempty"`       // xs, sm, md, lg, full
	Mode       *string `json:"mode,omitempty"`       // normal, inline, horizontal
	Horizontal any     `json:"horizontal,omitempty"` // FormHorizontal
	Inline     *bool   `json:"inline,omitempty"`

	// Behavior
	SubmitOnChange     *bool `json:"submitOnChange,omitempty"`
	ClearValueOnHidden *bool `json:"clearValueOnHidden,omitempty"`

	// AutoFill
	AutoFill     any `json:"autoFill,omitempty"`     // map or AutoFillConfig
	InitAutoFill any `json:"initAutoFill,omitempty"` // bool or "fillIfNotSet"

	// Events & Actions
	OnEvent map[string]EventConfig `json:"onEvent,omitempty"`

	// Editor & Meta
	EditorSetting *EditorConfig  `json:"editorSetting,omitempty"`
	TestIDBuilder any            `json:"testIdBuilder,omitempty"`
	UseMobileUI   *bool          `json:"useMobileUI,omitempty"`
	Style         map[string]any `json:"style,omitempty"`

	// Extra Fields
	ExtraName *string `json:"extraName,omitempty"`
	Row       *int    `json:"row,omitempty"`

	// Body (for composite controls)
	Body any `json:"body,omitempty"` // SchemaCollection
}

// EventConfig event configuration with actions
type EventConfig struct {
	Weight   *int            `json:"weight,omitempty"`
	Actions  []ActionConfig  `json:"actions"`
	Debounce *DebounceConfig `json:"debounce,omitempty"`
	Track    *TrackConfig    `json:"track,omitempty"`
}

// DebounceConfig debounce configuration
type DebounceConfig struct {
	Wait     *int  `json:"wait,omitempty"`
	Leading  *bool `json:"leading,omitempty"`
	Trailing *bool `json:"trailing,omitempty"`
	MaxWait  *int  `json:"maxWait,omitempty"`
}

// TrackConfig track configuration
type TrackConfig struct {
	Off      *bool   `json:"off,omitempty"`
	Exp      *string `json:"exp,omitempty"`
	MaskKeys *string `json:"maskKeys,omitempty"`
}

// ActionConfig event action configuration
type ActionConfig struct {
	ActionType string  `json:"actionType"`
	Dialog     any     `json:"dialog,omitempty"`
	Drawer     any     `json:"drawer,omitempty"`
	Toast      any     `json:"toast,omitempty"`
	API        any     `json:"api,omitempty"`
	Redirect   *string `json:"redirect,omitempty"`
	Reload     *string `json:"reload,omitempty"`
	// Add more as needed
}

// EditorConfig editor configuration
type EditorConfig struct {
	Behavior    *string `json:"behavior,omitempty"`
	DisplayName *string `json:"displayName,omitempty"`
	Mock        any     `json:"mock,omitempty"`
}

// AutoFillConfig auto-fill configuration
type AutoFillConfig struct {
	ShowSuggestion   *bool          `json:"showSuggestion,omitempty"`
	DefaultSelection any            `json:"defaultSelection,omitempty"`
	API              any            `json:"api,omitempty"`
	Silent           *bool          `json:"silent,omitempty"`
	FillMapping      map[string]any `json:"fillMappinng,omitempty"`
	Trigger          *string        `json:"trigger,omitempty"` // change, focus, blur
	Mode             *string        `json:"mode,omitempty"`    // popOver, dialog, drawer
	Position         *string        `json:"position,omitempty"`
	Size             *string        `json:"size,omitempty"`
	Columns          []any          `json:"columns,omitempty"`
	Filter           any            `json:"filter,omitempty"`
}

// FieldValidationsExtended comprehensive validation rules
type FieldValidations struct {
	// Type validations
	IsAlpha        *bool `json:"isAlpha,omitempty"`
	IsAlphanumeric *bool `json:"isAlphanumeric,omitempty"`
	IsEmail        *bool `json:"isEmail,omitempty"`
	IsFloat        *bool `json:"isFloat,omitempty"`
	IsInt          *bool `json:"isInt,omitempty"`
	IsJSON         *bool `json:"isJson,omitempty"`
	IsNumeric      *bool `json:"isNumeric,omitempty"`
	IsRequired     *bool `json:"isRequired,omitempty"`
	IsURL          *bool `json:"isUrl,omitempty"`

	// Length/Range validations
	IsLength  any `json:"isLength,omitempty"`  // number
	MinLength any `json:"minLength,omitempty"` // number
	MaxLength any `json:"maxLength,omitempty"` // number
	Minimum   any `json:"minimum,omitempty"`   // number
	Maximum   any `json:"maximum,omitempty"`   // number

	// Pattern validations
	MatchRegexp  *string `json:"matchRegexp,omitempty"`
	MatchRegexp1 *string `json:"matchRegexp1,omitempty"`
	MatchRegexp2 *string `json:"matchRegexp2,omitempty"`
	MatchRegexp3 *string `json:"matchRegexp3,omitempty"`
	MatchRegexp4 *string `json:"matchRegexp4,omitempty"`
	MatchRegexp5 *string `json:"matchRegexp5,omitempty"`

	// DateTime validations
	IsDateTimeSame         any `json:"isDateTimeSame,omitempty"`
	IsDateTimeBefore       any `json:"isDateTimeBefore,omitempty"`
	IsDateTimeAfter        any `json:"isDateTimeAfter,omitempty"`
	IsDateTimeSameOrBefore any `json:"isDateTimeSameOrBefore,omitempty"`
	IsDateTimeSameOrAfter  any `json:"isDateTimeSameOrAfter,omitempty"`
	IsDateTimeBetween      any `json:"isDateTimeBetween,omitempty"`

	// Time validations
	IsTimeSame         any `json:"isTimeSame,omitempty"`
	IsTimeBefore       any `json:"isTimeBefore,omitempty"`
	IsTimeAfter        any `json:"isTimeAfter,omitempty"`
	IsTimeSameOrBefore any `json:"isTimeSameOrBefore,omitempty"`
	IsTimeSameOrAfter  any `json:"isTimeSameOrAfter,omitempty"`
	IsTimeBetween      any `json:"isTimeBetween,omitempty"`
}

// InputText text input field
type InputText struct {
	FormControl
	MinLength   *int    `json:"minLength,omitempty"`
	MaxLength   *int    `json:"maxLength,omitempty"`
	Pattern     *string `json:"pattern,omitempty"`
	Clearable   *bool   `json:"clearable,omitempty"`
	ShowCounter *bool   `json:"showCounter,omitempty"`
}

// InputNumber numeric input field
type InputNumber struct {
	FormControl
	Minimum   *float64 `json:"minimum,omitempty"`
	Maximum   *float64 `json:"maximum,omitempty"`
	Step      *float64 `json:"step,omitempty"`
	Precision *int     `json:"precision,omitempty"`
}

// InputCity city input field
type InputCity struct {
	FormControl
	Source *API `json:"source,omitempty"`
}

// InputColor color input field
type InputColor struct {
	FormControl
	ShowAlpha *bool   `json:"showAlpha,omitempty"`
	Format    *string `json:"format,omitempty"`
}

// Select dropdown selection field
type Select struct {
	FormControl
	Options     []Option `json:"options,omitempty"`
	Multiple    *bool    `json:"multiple,omitempty"`
	Searchable  *bool    `json:"searchable,omitempty"`
	Creatable   *bool    `json:"creatable,omitempty"`
	Source      *API     `json:"source,omitempty"`
	MaxTagCount *int     `json:"maxTagCount,omitempty"`
}

// NestedSelect nested select field
type NestedSelect struct {
	FormControl
	Options []NestedOption `json:"options,omitempty"`
	Source  *API           `json:"source,omitempty"`
}

// NestedOption nested select option
type NestedOption struct {
	Label    string         `json:"label"`
	Value    any            `json:"value"`
	Children []NestedOption `json:"children,omitempty"`
}

// ChainedSelect chained select field
type ChainedSelect struct {
	FormControl
	Options []any `json:"options,omitempty"`
	Source  *API  `json:"source,omitempty"`
}

// Checkbox checkbox field
type Checkbox struct {
	FormControl
	Option *Option `json:"option,omitempty"`
}

// Checkboxes multiple checkbox field
type Checkboxes struct {
	FormControl
	Options []Option `json:"options,omitempty"`
}

// Radio radio button field
type Radio struct {
	FormControl
	Options   []Option `json:"options,omitempty"`
	Direction *string  `json:"direction,omitempty"`
}

// Radios radio group field
type Radios struct {
	FormControl
	Options   []Option `json:"options,omitempty"`
	Direction *string  `json:"direction,omitempty"`
}

// Textarea multiline text field
type Textarea struct {
	FormControl
	MinLength   *int  `json:"minLength,omitempty"`
	MaxLength   *int  `json:"maxLength,omitempty"`
	Rows        *int  `json:"rows,omitempty"`
	ShowCounter *bool `json:"showCounter,omitempty"`
}

// DatePicker date selection field
type DatePicker struct {
	FormControl
	Format    *string `json:"format,omitempty"`
	MinDate   *string `json:"minDate,omitempty"`
	MaxDate   *string `json:"maxDate,omitempty"`
	StartDate *string `json:"startDate,omitempty"`
}

// DateRange date range selection field
type DateRange struct {
	FormControl
	Format  *string `json:"format,omitempty"`
	MinDate *string `json:"minDate,omitempty"`
	MaxDate *string `json:"maxDate,omitempty"`
}

// TimePicker time selection field
type TimePicker struct {
	FormControl
	Format  *string `json:"format,omitempty"`
	MinTime *string `json:"minTime,omitempty"`
	MaxTime *string `json:"maxTime,omitempty"`
}

// MonthPicker month selection field
type MonthPicker struct {
	FormControl
	Format *string `json:"format,omitempty"`
}

// YearPicker year selection field
type YearPicker struct {
	FormControl
	Format *string `json:"format,omitempty"`
}

// QuarterPicker quarter selection field
type QuarterPicker struct {
	FormControl
	Format *string `json:"format,omitempty"`
}

// File file upload field
type File struct {
	FormControl
	Accept      *string `json:"accept,omitempty"`
	Multiple    *bool   `json:"multiple,omitempty"`
	MaxSize     *int64  `json:"maxSize,omitempty"`
	MaxCount    *int    `json:"maxCount,omitempty"`
	UploadAPI   *API    `json:"uploadApi,omitempty"`
	DownloadAPI *API    `json:"downloadApi,omitempty"`
}

// Array array field (repeating group)
type Array struct {
	FormControl
	Controls  []FormControl `json:"controls,omitempty"`
	AddButton *string       `json:"addButton,omitempty"`
	MinLength *int          `json:"minLength,omitempty"`
	MaxLength *int          `json:"maxLength,omitempty"`
}

// Switch toggle field
type Switch struct {
	FormControl
	Option *Option `json:"option,omitempty"`
}

// Rating rating field
type Rating struct {
	FormControl
	Count     *int  `json:"count,omitempty"`
	AllowHalf *bool `json:"allowHalf,omitempty"`
	ReadOnly  *bool `json:"readOnly,omitempty"`
}

// Combo combo field (multiple related inputs)
type Combo struct {
	FormControl
	Controls []FormControl `json:"controls,omitempty"`
}

// FormMessages form-level messages
type FormMessages struct {
	ValidateFailed *string `json:"validateFailed,omitempty"`
	SubmitSuccess  *string `json:"submitSuccess,omitempty"`
	SubmitFailed   *string `json:"submitFailed,omitempty"`
}

// ============================================================================
// TABLE COMPONENTS (Extended)
// ============================================================================

// Table data table component
type Table struct {
	Type            string        `json:"type"`
	Title           *string       `json:"title,omitempty"`
	Columns         []TableColumn `json:"columns,omitempty"`
	Data            []any         `json:"data,omitempty"`
	Source          *API          `json:"source,omitempty"`
	Selectable      *bool         `json:"selectable,omitempty"`
	Sortable        *bool         `json:"sortable,omitempty"`
	Searchable      *bool         `json:"searchable,omitempty"`
	PageSize        *int          `json:"pageSize,omitempty"`
	ClassName       *string       `json:"className,omitempty"`
	HeaderClassName *string       `json:"headerClassName,omitempty"`
	FooterClassName *string       `json:"footerClassName,omitempty"`
	RowClassName    *string       `json:"rowClassName,omitempty"`
}

// TableColumn table column definition
type TableColumn struct {
	Name       string  `json:"name"`
	Label      string  `json:"label"`
	Type       *string `json:"type,omitempty"`
	Sortable   *bool   `json:"sortable,omitempty"`
	Searchable *bool   `json:"searchable,omitempty"`
	Width      *string `json:"width,omitempty"`
	Align      *string `json:"align,omitempty"`
	Fixed      *string `json:"fixed,omitempty"`
	Remark     *string `json:"remark,omitempty"`
	Template   *string `json:"template,omitempty"`
}

// CRUD CRUD component
type CRUD struct {
	Type         string        `json:"type"`
	Title        *string       `json:"title,omitempty"`
	Mode         *string       `json:"mode,omitempty"`
	Columns      []TableColumn `json:"columns,omitempty"`
	ListAPI      *API          `json:"api,omitempty"`
	CreateAPI    *API          `json:"createApi,omitempty"`
	UpdateAPI    *API          `json:"updateApi,omitempty"`
	DeleteAPI    *API          `json:"deleteApi,omitempty"`
	CreateButton *bool         `json:"createButton,omitempty"`
	UpdateButton *bool         `json:"updateButton,omitempty"`
	DeleteButton *bool         `json:"deleteButton,omitempty"`
	ClassName    *string       `json:"className,omitempty"`
}

// List list component
type List struct {
	Type       string      `json:"type"`
	Title      *string     `json:"title,omitempty"`
	ListFields []ListField `json:"listFields,omitempty"`
	Data       []any       `json:"data,omitempty"`
	Source     *API        `json:"source,omitempty"`
	ClassName  *string     `json:"className,omitempty"`
}

// ListField list item field
type ListField struct {
	Name  string  `json:"name"`
	Label string  `json:"label"`
	Type  *string `json:"type,omitempty"`
}

// ============================================================================
// ACTION & BUTTON COMPONENTS (Extended)
// ============================================================================

// Action button action component
type Action struct {
	Type       string         `json:"type"`
	Label      string         `json:"label"`
	Level      *string        `json:"level,omitempty"`
	Icon       *string        `json:"icon,omitempty"`
	Disabled   *bool          `json:"disabled,omitempty"`
	OnClick    *string        `json:"onClick,omitempty"`
	ActionType *string        `json:"actionType,omitempty"`
	Dialog     any            `json:"dialog,omitempty"`
	Drawer     any            `json:"drawer,omitempty"`
	Redirect   *string        `json:"redirect,omitempty"`
	Blank      *bool          `json:"blank,omitempty"`
	API        *API           `json:"api,omitempty"`
	ClassName  *string        `json:"className,omitempty"`
	Style      map[string]any `json:"style,omitempty"`
}

// Button button component
type Button struct {
	Type       string  `json:"type"`
	Label      string  `json:"label"`
	Level      *string `json:"level,omitempty"`
	Icon       *string `json:"icon,omitempty"`
	Disabled   *bool   `json:"disabled,omitempty"`
	OnClick    *string `json:"onClick,omitempty"`
	ActionType *string `json:"actionType,omitempty"`
	ClassName  *string `json:"className,omitempty"`
}

// ButtonGroup button group component
type ButtonGroup struct {
	Type      string   `json:"type"`
	Buttons   []Button `json:"buttons,omitempty"`
	ClassName *string  `json:"className,omitempty"`
}

// ButtonToolbar button toolbar component
type ButtonToolbar struct {
	Type      string  `json:"type"`
	Buttons   []any   `json:"buttons,omitempty"`
	ClassName *string `json:"className,omitempty"`
}

// Dialog modal dialog component
type Dialog struct {
	Type      string   `json:"type"`
	Title     string   `json:"title"`
	Body      any      `json:"body,omitempty"`
	Size      *string  `json:"size,omitempty"`
	Actions   []Action `json:"actions,omitempty"`
	ClassName *string  `json:"className,omitempty"`
}

// Drawer side drawer component
type Drawer struct {
	Type      string  `json:"type"`
	Title     string  `json:"title"`
	Body      any     `json:"body,omitempty"`
	Position  *string `json:"position,omitempty"`
	Size      *string `json:"size,omitempty"`
	ClassName *string `json:"className,omitempty"`
}

// ============================================================================
// DISPLAY COMPONENTS (Extended)
// ============================================================================

// Alert alert notification component
type Alert struct {
	Type      string  `json:"type"`
	Title     *string `json:"title,omitempty"`
	Body      string  `json:"body"`
	Level     *string `json:"level,omitempty"`
	Icon      *string `json:"icon,omitempty"`
	ClassName *string `json:"className,omitempty"`
	Closeable *bool   `json:"closeable,omitempty"`
}

// Progress progress bar component
type Progress struct {
	Type       string  `json:"type"`
	Percentage *int    `json:"percentage,omitempty"`
	Status     *string `json:"status,omitempty"`
	ClassName  *string `json:"className,omitempty"`
}

// Divider visual divider component
type Divider struct {
	Type      string  `json:"type"`
	Title     *string `json:"title,omitempty"`
	ClassName *string `json:"className,omitempty"`
	Direction *string `json:"direction,omitempty"`
}

// Image image component
type Image struct {
	Type      string         `json:"type"`
	Src       string         `json:"src"`
	Title     *string        `json:"title,omitempty"`
	Alt       *string        `json:"alt,omitempty"`
	Width     *string        `json:"width,omitempty"`
	Height    *string        `json:"height,omitempty"`
	ClassName *string        `json:"className,omitempty"`
	Style     map[string]any `json:"style,omitempty"`
}

// Images image gallery component
type Images struct {
	Type      string   `json:"type"`
	Images    []string `json:"images,omitempty"`
	ClassName *string  `json:"className,omitempty"`
}

// Video video component
type Video struct {
	Type      string  `json:"type"`
	Src       string  `json:"src"`
	Poster    *string `json:"poster,omitempty"`
	Controls  *bool   `json:"controls,omitempty"`
	Autoplay  *bool   `json:"autoplay,omitempty"`
	ClassName *string `json:"className,omitempty"`
}

// Audio audio component
type Audio struct {
	Type      string  `json:"type"`
	Src       string  `json:"src"`
	Controls  *bool   `json:"controls,omitempty"`
	Autoplay  *bool   `json:"autoplay,omitempty"`
	ClassName *string `json:"className,omitempty"`
}

// IFrame iframe component
type IFrame struct {
	Type      string  `json:"type"`
	Src       string  `json:"src"`
	Width     *string `json:"width,omitempty"`
	Height    *string `json:"height,omitempty"`
	ClassName *string `json:"className,omitempty"`
}

// QRCode QR code component
type QRCode struct {
	Type      string  `json:"type"`
	Value     string  `json:"value"`
	Size      *int    `json:"size,omitempty"`
	ClassName *string `json:"className,omitempty"`
}

// Icon icon component
type Icon struct {
	Type      string  `json:"type"`
	Icon      string  `json:"icon"`
	ClassName *string `json:"className,omitempty"`
}

// Badge badge component
type Badge struct {
	Type      string  `json:"type"`
	Badge     any     `json:"badge,omitempty"`
	Text      *string `json:"text,omitempty"`
	ClassName *string `json:"className,omitempty"`
}

// Tag tag component
type Tag struct {
	Type      string  `json:"type"`
	Label     string  `json:"label"`
	Color     *string `json:"color,omitempty"`
	Closeable *bool   `json:"closeable,omitempty"`
	ClassName *string `json:"className,omitempty"`
}

// Status status component
type Status struct {
	Type      string  `json:"type"`
	Value     any     `json:"value,omitempty"`
	Text      *string `json:"text,omitempty"`
	ClassName *string `json:"className,omitempty"`
}

// ============================================================================
// CHART & DATA VISUALIZATION COMPONENTS
// ============================================================================

// Chart chart component
type Chart struct {
	Type      string         `json:"type"`
	Config    map[string]any `json:"config,omitempty"`
	Width     *string        `json:"width,omitempty"`
	Height    *string        `json:"height,omitempty"`
	ClassName *string        `json:"className,omitempty"`
}

// Calendar calendar component
type Calendar struct {
	Type      string  `json:"type"`
	Events    []Event `json:"events,omitempty"`
	ClassName *string `json:"className,omitempty"`
}

// Event calendar event
type Event struct {
	Title     string `json:"title"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

// Timeline timeline component
type Timeline struct {
	Type      string         `json:"type"`
	Items     []TimelineItem `json:"items,omitempty"`
	Direction *string        `json:"direction,omitempty"`
	ClassName *string        `json:"className,omitempty"`
}

// TimelineItem timeline item
type TimelineItem struct {
	Title  string `json:"title"`
	Time   string `json:"time"`
	Detail any    `json:"detail,omitempty"`
}

// ============================================================================
// SHARED TYPES & HELPERS
// ============================================================================

// Option generic option for select components
type Option struct {
	Label string  `json:"label"`
	Value any     `json:"value"`
	Icon  *string `json:"icon,omitempty"`
	Desc  *string `json:"desc,omitempty"`
}

// API API configuration
type API struct {
	URL     string            `json:"url"`
	Method  *string           `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Data    map[string]any    `json:"data,omitempty"`
	Cache   *int              `json:"cache,omitempty"`
}

// Expression expression for conditionals
type Expression struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// NewPage creates a new page
func NewPage(title string) *Page {
	return &Page{
		Type:  "page",
		Title: &title,
	}
}

// NewForm creates a new form
func NewForm(title string) *Form {
	return &Form{
		Type:  "form",
		Title: &title,
	}
}

// NewInputText creates a new text input field
func NewInputText(name, label string) *InputText {
	return &InputText{
		FormControl: FormControl{
			Type:  "input-text",
			Name:  name,
			Label: &label,
		},
	}
}

// NewInputNumber creates a new numeric input field
func NewInputNumber(name, label string) *InputNumber {
	return &InputNumber{
		FormControl: FormControl{
			Type:  "input-number",
			Name:  name,
			Label: &label,
		},
	}
}

// NewSelect creates a new select field
func NewSelect(name, label string, options []Option) *Select {
	return &Select{
		FormControl: FormControl{
			Type:  "select",
			Name:  name,
			Label: &label,
		},
		Options: options,
	}
}

// NewCheckbox creates a new checkbox field
func NewCheckbox(name, label string) *Checkbox {
	return &Checkbox{
		FormControl: FormControl{
			Type:  "checkbox",
			Name:  name,
			Label: &label,
		},
	}
}

// NewDatePicker creates a new date picker field
func NewDatePicker(name, label string) *DatePicker {
	return &DatePicker{
		FormControl: FormControl{
			Type:  "date",
			Name:  name,
			Label: &label,
		},
	}
}

// NewTextarea creates a new textarea field
func NewTextarea(name, label string) *Textarea {
	return &Textarea{
		FormControl: FormControl{
			Type:  "textarea",
			Name:  name,
			Label: &label,
		},
	}
}

// NewTable creates a new table component
func NewTable(title string) *Table {
	return &Table{
		Type:  "table",
		Title: &title,
	}
}

// NewButton creates a new button
func NewButton(label string) *Button {
	return &Button{
		Type:  "button",
		Label: label,
	}
}

// NewAlert creates a new alert
func NewAlert(level, body string) *Alert {
	return &Alert{
		Type:  "alert",
		Level: &level,
		Body:  body,
	}
}

// NewDialog creates a new dialog
func NewDialog(title string) *Dialog {
	return &Dialog{
		Type:  "dialog",
		Title: title,
	}
}

// NewCard creates a new card
func NewCard() *Card {
	return &Card{
		Type: "card",
	}
}

// NewTabs creates a new tabs component
func NewTabs() *Tabs {
	return &Tabs{
		Type: "tabs",
	}
}

// NewSteps creates a new steps component
func NewSteps() *Steps {
	return &Steps{
		Type: "steps",
	}
}

// NewGrid creates a new grid layout
func NewGrid() *Grid {
	return &Grid{
		Type: "grid",
	}
}

// ============================================================================
// BUILDER PATTERNS
// ============================================================================

// FormBuilder for building forms fluently
type FormBuilder struct {
	form *Form
}

// NewFormBuilder creates a new form builder
func NewFormBuilder(title string) *FormBuilder {
	return &FormBuilder{
		form: NewForm(title),
	}
}

// AddControl adds a form control to the form
func (fb *FormBuilder) AddControl(ctrl FormControl) *FormBuilder {
	// Body is handled as any, need to append to slice
	// This is simplified - actual implementation would marshal/unmarshal JSON
	return fb
}

// WithLayout sets the form layout
func (fb *FormBuilder) WithLayout(layout string) *FormBuilder {
	fb.form.Layout = &layout
	return fb
}

// WithSubmitText sets the submit button text
func (fb *FormBuilder) WithSubmitText(text string) *FormBuilder {
	fb.form.SubmitText = &text
	return fb
}

// WithAPI sets the form submission API
func (fb *FormBuilder) WithAPI(url string) *FormBuilder {
	fb.form.API = &API{URL: url}
	return fb
}

// Build returns the constructed form
func (fb *FormBuilder) Build() *Form {
	return fb.form
}

// ============================================================================
// OPTION HELPERS
// ============================================================================

// NewOption creates a new option
func NewOption(label string, value any) Option {
	return Option{
		Label: label,
		Value: value,
	}
}

// NewOptions creates multiple options
func NewOptions(items ...map[string]any) []Option {
	opts := make([]Option, len(items))
	for i, item := range items {
		opts[i] = Option{
			Label: item["label"].(string),
			Value: item["value"],
		}
	}
	return opts
}

// ============================================================================
// API HELPERS
// ============================================================================

// NewAPI creates a new API configuration
func NewAPI(url string) *API {
	return &API{
		URL: url,
	}
}

// WithMethod sets the HTTP method
func (a *API) WithMethod(method string) *API {
	a.Method = &method
	return a
}

// WithHeaders sets HTTP headers
func (a *API) WithHeaders(headers map[string]string) *API {
	a.Headers = headers
	return a
}

// WithData sets request data
func (a *API) WithData(data map[string]any) *API {
	a.Data = data
	return a
}

// WithCache sets cache duration in seconds
func (a *API) WithCache(seconds int) *API {
	a.Cache = &seconds
	return a
}

// ============================================================================
// VALIDATION HELPERS
// ============================================================================

// NewValidations creates field validations
func NewValidations() *FieldValidations {
	return &FieldValidations{}
}

// SetRequired marks field as required
func (fv *FieldValidations) SetRequired(required bool) *FieldValidations {
	fv.IsRequired = &required
	return fv
}

// SetEmail validates email format
func (fv *FieldValidations) SetEmail(isEmail bool) *FieldValidations {
	fv.IsEmail = &isEmail
	return fv
}

// SetNumeric validates numeric format
func (fv *FieldValidations) SetNumeric(isNumeric bool) *FieldValidations {
	fv.IsNumeric = &isNumeric
	return fv
}

// SetMinLength sets minimum length
func (fv *FieldValidations) SetMinLength(min int) *FieldValidations {
	fv.MinLength = &min
	return fv
}

// SetMaxLength sets maximum length
func (fv *FieldValidations) SetMaxLength(max int) *FieldValidations {
	fv.MaxLength = &max
	return fv
}

// ============================================================================
// COMMON PATTERNS
// ============================================================================

// IsFieldRequired checks if field is required
func (fc *FormControl) IsFieldRequired() bool {
	return fc.Required != nil && *fc.Required
}

// IsFieldDisabled checks if field is disabled
func (fc *FormControl) IsFieldDisabled() bool {
	return fc.Disabled != nil && *fc.Disabled
}

// IsFieldHidden checks if field is hidden
func (fc *FormControl) IsFieldHidden() bool {
	return fc.Hidden != nil && *fc.Hidden
}

// SetFieldValue sets the field value
func (fc *FormControl) SetFieldValue(value any) {
	fc.Value = value
}

// SetFieldDefault sets the field default value
func (fc *FormControl) SetFieldDefault(value any) {
	fc.DefaultValue = value
}

// MarshalJSON custom marshaling for FormControl types
func (it *InputText) MarshalJSON() ([]byte, error) {
	type Alias InputText
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(it),
	})
}

// ============================================================================
// TABLE HELPERS
// ============================================================================

// AddColumn adds a column to the table
func (t *Table) AddColumn(col TableColumn) *Table {
	t.Columns = append(t.Columns, col)
	return t
}

// AddRow adds a data row to the table
func (t *Table) AddRow(row map[string]any) *Table {
	t.Data = append(t.Data, row)
	return t
}

// SetSource sets the data source API
func (t *Table) SetSource(api *API) *Table {
	t.Source = api
	return t
}

// ============================================================================
// ACTION HELPERS
// ============================================================================

// NewLinkAction creates a link action
func NewLinkAction(label, href string) *Action {
	return &Action{
		Type:       "button",
		Label:      label,
		ActionType: StrPtr("link"),
		Redirect:   &href,
	}
}

// NewDialogAction creates a dialog action
func NewDialogAction(label string, dialog *Dialog) *Action {
	return &Action{
		Type:       "button",
		Label:      label,
		ActionType: StrPtr("dialog"),
		Dialog:     dialog,
	}
}

// NewDrawerAction creates a drawer action
func NewDrawerAction(label string, drawer *Drawer) *Action {
	return &Action{
		Type:       "button",
		Label:      label,
		ActionType: StrPtr("drawer"),
		Drawer:     drawer,
	}
}

// NewAPIAction creates an API action
func NewAPIAction(label string, api *API) *Action {
	return &Action{
		Type:       "button",
		Label:      label,
		ActionType: StrPtr("ajax"),
		API:        api,
	}
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// StrPtr returns pointer to string
func StrPtr(s string) *string {
	return &s
}

// IntPtr returns pointer to int
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns pointer to bool
func BoolPtr(b bool) *bool {
	return &b
}

// Float64Ptr returns pointer to float64
func Float64Ptr(f float64) *float64 {
	return &f
}

// ToJSON converts component to JSON bytes
func ToJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

// ToJSONString converts component to JSON string
func ToJSONString(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON unmarshals JSON to SchemaObject
func FromJSON(data []byte) (*SchemaObject, error) {
	var obj SchemaObject
	err := json.Unmarshal(data, &obj)
	return &obj, err
}

// FromJSONString unmarshals JSON string to SchemaObject
func FromJSONString(jsonStr string) (*SchemaObject, error) {
	return FromJSON([]byte(jsonStr))
}

// ============================================================================
// TYPE CONVERTERS
// ============================================================================

// ToPage converts SchemaObject to Page
func (so *SchemaObject) ToPage() (*Page, error) {
	var page Page
	if err := json.Unmarshal(so.Body, &page); err != nil {
		return nil, fmt.Errorf("failed to convert to Page: %w", err)
	}
	return &page, nil
}

// ToForm converts SchemaObject to Form
func (so *SchemaObject) ToForm() (*Form, error) {
	var form Form
	if err := json.Unmarshal(so.Body, &form); err != nil {
		return nil, fmt.Errorf("failed to convert to Form: %w", err)
	}
	return &form, nil
}

// ToTable converts SchemaObject to Table
func (so *SchemaObject) ToTable() (*Table, error) {
	var table Table
	if err := json.Unmarshal(so.Body, &table); err != nil {
		return nil, fmt.Errorf("failed to convert to Table: %w", err)
	}
	return &table, nil
}
