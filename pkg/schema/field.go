package schema

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/niiniyare/erp/pkg/condition"
)

// Field represents a form input or UI component
type Field struct {
	// Identity
	Name  string    `json:"name" validate:"required,min=1,max=100" example:"email"`
	Type  FieldType `json:"type" validate:"required" example:"email"`
	Label string    `json:"label" validate:"required,min=1,max=200" example:"Email Address"`

	// Description and help
	Description string `json:"description,omitempty" validate:"max=500" example:"Enter your primary email"`
	Placeholder string `json:"placeholder,omitempty" validate:"max=200" example:"user@example.com"`
	Help        string `json:"help,omitempty" validate:"max=500"`    // Help text below field
	Tooltip     string `json:"tooltip,omitempty" validate:"max=200"` // Tooltip on hover
	Icon        string `json:"icon,omitempty" validate:"icon_name"`  // Icon to display
	Error       string `json:"error,omitempty" validate:"max=200"`   // Custom error message

	// State flags
	Required bool `json:"required,omitempty"` // Field must have value
	Disabled bool `json:"disabled,omitempty"` // Field cannot be edited
	Readonly bool `json:"readonly,omitempty"` // Value visible but not editable
	Hidden   bool `json:"hidden,omitempty"`   // Field not displayed

	// Values
	Value   any      `json:"value,omitempty"`                   // Current value
	Default any      `json:"default,omitempty"`                 // Default value (supports ${expressions})
	Options []Option `json:"options,omitempty" validate:"dive"` // For select/radio/checkbox

	// Behavior configuration
	Validation   *FieldValidation  `json:"validation,omitempty"`                             // Validation rules
	Transform    *Transform        `json:"transform,omitempty"`                              // Value transformation
	Mask         *Mask             `json:"mask,omitempty"`                                   // Input masking
	Layout       *FieldLayout      `json:"layout,omitempty"`                                 // Positioning in grid
	Style        *Style            `json:"style,omitempty"`                                  // Custom styling
	Config       map[string]any    `json:"config,omitempty"`                                 // Field-specific config
	Events       *FieldEvents      `json:"events,omitempty"`                                 // Event handlers
	Conditional  *Conditional      `json:"conditional,omitempty"`                            // Show/hide conditions
	DataSource   *DataSource       `json:"dataSource,omitempty"`                             // Dynamic options source
	Permissions  *FieldPermissions `json:"permissions,omitempty"`                            // Access control
	Dependencies []string          `json:"dependencies,omitempty" validate:"dive,fieldname"` // Depends on these fields

	// Internationalization
	I18n *FieldI18n `json:"i18n,omitempty"` // Translations

	// Framework integration
	HTMX   *FieldHTMX   `json:"htmx,omitempty"`   // HTMX attributes
	Alpine *FieldAlpine `json:"alpine,omitempty"` // Alpine.js bindings

	// Security
	Security *FieldSecurity `json:"security,omitempty"` // Security settings

	// Enrichment fields (for runtime permission/role checking)
	RequirePermission string   `json:"require_permission,omitempty"` // Permission required to view/edit
	RequireRoles      []string `json:"require_roles,omitempty"`      // Roles required to view/edit

	// Runtime state (populated by enricher, not in JSON)
	Runtime *FieldRuntime `json:"-"` // Runtime state set by enricher

	// Internal state (not in JSON)
	evaluator *condition.Evaluator `json:"-"` // Condition evaluator (injected)
}

// GetLocalizedLabel returns the field label in the specified locale
// Uses embedded translations based on field ID, with I18n as fallback
func (f *Field) GetLocalizedLabel(locale string) string {
	// Check if explicit I18n translations exist (backwards compatibility)
	if f.I18n != nil && f.I18n.Label != nil {
		if label, exists := f.I18n.Label[locale]; exists && label != "" {
			return label
		}
	}
	
	// Use embedded translations with field ID as key
	return T_Field(locale, f.Name)
}

// GetLocalizedPlaceholder returns the placeholder text in the specified locale
func (f *Field) GetLocalizedPlaceholder(locale string) string {
	// Check if explicit I18n translations exist
	if f.I18n != nil && f.I18n.Placeholder != nil {
		if placeholder, exists := f.I18n.Placeholder[locale]; exists && placeholder != "" {
			return placeholder
		}
	}
	
	// For placeholder, fallback to field-specific key or empty
	placeholderKey := f.Name + "Placeholder"
	placeholder := T_Field(locale, placeholderKey)
	
	// If no specific placeholder translation exists, return the original or empty
	if placeholder == humanizeFieldID(placeholderKey) {
		return f.Placeholder // Return original placeholder
	}
	
	return placeholder
}

// GetLocalizedHelp returns the help text in the specified locale
func (f *Field) GetLocalizedHelp(locale string) string {
	// Check if explicit I18n translations exist
	if f.I18n != nil && f.I18n.Help != nil {
		if help, exists := f.I18n.Help[locale]; exists && help != "" {
			return help
		}
	}
	
	// For help text, fallback to field-specific key or empty
	helpKey := f.Name + "Help"
	help := T_Field(locale, helpKey)
	
	// If no specific help translation exists, return the original or empty
	if help == humanizeFieldID(helpKey) {
		return f.Help // Return original help text
	}
	
	return help
}

// GetLocalizedDescription returns the description in the specified locale
func (f *Field) GetLocalizedDescription(locale string) string {
	// Check if explicit I18n translations exist
	if f.I18n != nil && f.I18n.Description != nil {
		if desc, exists := f.I18n.Description[locale]; exists && desc != "" {
			return desc
		}
	}
	
	// For description, fallback to field-specific key or empty
	descKey := f.Name + "Description"
	desc := T_Field(locale, descKey)
	
	// If no specific description translation exists, return the original or empty
	if desc == humanizeFieldID(descKey) {
		return f.Description // Return original description
	}
	
	return desc
}

// FieldType defines all supported input types
type FieldType string

const (
	// Basic text inputs
	FieldText     FieldType = "text"
	FieldEmail    FieldType = "email"
	FieldPassword FieldType = "password"
	FieldNumber   FieldType = "number"
	FieldHidden   FieldType = "hidden"
	FieldPhone    FieldType = "phone"
	FieldURL      FieldType = "url"

	// Date and time
	FieldDate      FieldType = "date"
	FieldTime      FieldType = "time"
	FieldDateTime  FieldType = "datetime"
	FieldDateRange FieldType = "daterange"
	FieldMonth     FieldType = "month"   // Month picker (YYYY-MM)
	FieldYear      FieldType = "year"    // Year picker
	FieldQuarter   FieldType = "quarter" // Quarter picker (Q1-Q4)

	// Text content
	FieldTextarea FieldType = "textarea"
	FieldRichText FieldType = "richtext" // WYSIWYG editor
	FieldCode     FieldType = "code"     // Code editor with syntax highlighting
	FieldJSON     FieldType = "json"     // JSON editor with validation

	// Selection
	FieldSelect      FieldType = "select"      // Dropdown
	FieldMultiSelect FieldType = "multiselect" // Multiple selection dropdown
	FieldRadio       FieldType = "radio"       // Radio buttons
	FieldCheckbox    FieldType = "checkbox"    // Single checkbox
	FieldCheckboxes  FieldType = "checkboxes"  // Multiple checkboxes
	FieldTreeSelect  FieldType = "treeselect"  // Hierarchical select
	FieldCascader    FieldType = "cascader"    // Cascading dropdown
	FieldTransfer    FieldType = "transfer"    // Transfer list (left/right)

	// Interactive controls
	FieldSwitch FieldType = "switch" // Toggle switch
	FieldSlider FieldType = "slider" // Range slider
	FieldRating FieldType = "rating" // Star rating
	FieldColor  FieldType = "color"  // Color picker

	// File uploads
	FieldFile      FieldType = "file"      // Generic file upload
	FieldImage     FieldType = "image"     // Image upload with preview
	FieldVideo     FieldType = "video"     // Video upload
	FieldAudio     FieldType = "audio"     // Audio upload
	FieldSignature FieldType = "signature" // Signature pad

	// Specialized
	FieldCurrency     FieldType = "currency"     // Money input with formatting
	FieldTags         FieldType = "tags"         // Tag input
	FieldLocation     FieldType = "location"     // Address/location picker
	FieldRelation     FieldType = "relation"     // Foreign key relationship
	FieldAutoComplete FieldType = "autocomplete" // Autocomplete search
	FieldIconPicker   FieldType = "icon-picker"  // Icon selector
	FieldFormula      FieldType = "formula"      // Computed/calculated field

	// Display only
	FieldDisplay FieldType = "display" // Read-only display
	FieldDivider FieldType = "divider" // Visual separator
	FieldHTML    FieldType = "html"    // Raw HTML content
	FieldStatic  FieldType = "static"  // Static display text

	// Layout containers
	FieldGroup    FieldType = "group"    // Field grouping
	FieldFieldset FieldType = "fieldset" // HTML fieldset
	FieldTabs     FieldType = "tabs"     // Tabbed interface
	FieldPanel    FieldType = "panel"    // Panel container
	FieldCollapse FieldType = "collapse" // Collapsible section

	// Collections
	FieldRepeatable    FieldType = "repeatable"     // Repeatable field groups
	FieldTableRepeater FieldType = "table_repeater" // Table-style repeatable fields
)

// FieldValidation defines validation rules for a field
type FieldValidation struct {
	// String validation
	MinLength *int   `json:"minLength,omitempty"` // Minimum string length
	MaxLength *int   `json:"maxLength,omitempty"` // Maximum string length
	Pattern   string `json:"pattern,omitempty"`   // Regex pattern
	Format    string `json:"format,omitempty"`    // Format validator (email, url, uuid, date, datetime, time)

	// Number validation
	Min          *float64 `json:"min,omitempty"`          // Minimum value
	Max          *float64 `json:"max,omitempty"`          // Maximum value
	Step         *float64 `json:"step,omitempty"`         // Value increment
	Integer      bool     `json:"integer,omitempty"`      // Must be integer
	Positive     bool     `json:"positive,omitempty"`     // Must be positive
	Negative     bool     `json:"negative,omitempty"`     // Must be negative
	MultipleOf   *float64 `json:"multipleOf,omitempty"`   // Must be multiple of
	ExclusiveMin bool     `json:"exclusiveMin,omitempty"` // Min is exclusive
	ExclusiveMax bool     `json:"exclusiveMax,omitempty"` // Max is exclusive

	// Array validation
	MinItems    *int `json:"minItems,omitempty"`    // Min array length
	MaxItems    *int `json:"maxItems,omitempty"`    // Max array length
	UniqueItems bool `json:"uniqueItems,omitempty"` // Items must be unique

	// File validation
	File  *FileValidation  `json:"file,omitempty"`  // File upload validation
	Image *ImageValidation `json:"image,omitempty"` // Image upload validation

	// Custom validation (uses condition engine formula syntax)
	Custom   string    `json:"custom,omitempty"` // Custom validation formula
	Messages *Messages `json:"messages"`         // Custom error messages
}

// FileValidation for file upload fields
type FileValidation struct {
	MaxSize  int64    `json:"maxSize,omitempty"`  // Max file size in bytes
	MinSize  int64    `json:"minSize,omitempty"`  // Min file size in bytes
	Accept   []string `json:"accept,omitempty"`   // Allowed MIME types
	MaxFiles int      `json:"maxFiles,omitempty"` // Max number of files (for multiple)
}

// ImageValidation extends file validation with image-specific rules
type ImageValidation struct {
	FileValidation
	MinWidth  int `json:"minWidth,omitempty"`  // Minimum image width
	MaxWidth  int `json:"maxWidth,omitempty"`  // Maximum image width
	MinHeight int `json:"minHeight,omitempty"` // Minimum image height
	MaxHeight int `json:"maxHeight,omitempty"` // Maximum image height
}

// Messages holds custom validation error messages
type Messages struct {
	Required  string `json:"required,omitempty"`
	MinLength string `json:"minLength,omitempty"`
	MaxLength string `json:"maxLength,omitempty"`
	Pattern   string `json:"pattern,omitempty"`
	Min       string `json:"min,omitempty"`
	Max       string `json:"max,omitempty"`
	Custom    string `json:"custom,omitempty"`
}

// Transform defines value transformation rules
type Transform struct {
	Type   string         `json:"type" validate:"oneof=uppercase lowercase trim capitalize slugify"` // Transform type
	Params map[string]any `json:"params,omitempty"`                                                  // Transform parameters
}

// Mask defines input masking
type Mask struct {
	Pattern     string `json:"pattern" example:"(999) 999-9999"`  // Mask pattern
	Placeholder string `json:"placeholder,omitempty" example:"_"` // Placeholder character
	ShowMask    bool   `json:"showMask,omitempty"`                // Show mask when empty
	Guide       bool   `json:"guide,omitempty"`                   // Show guide while typing
}

// FieldLayout controls field positioning in grid
type FieldLayout struct {
	Row     int    `json:"row,omitempty"`     // Grid row
	Column  int    `json:"column,omitempty"`  // Grid column
	ColSpan int    `json:"colSpan,omitempty"` // Columns to span
	RowSpan int    `json:"rowSpan,omitempty"` // Rows to span
	Order   int    `json:"order,omitempty"`   // Display order
	Width   string `json:"width,omitempty"`   // Custom width
	Offset  int    `json:"offset,omitempty"`  // Column offset
	Class   string `json:"class,omitempty"`   // CSS classes
}

// Style defines custom styling
type Style struct {
	Classes        string            `json:"classes,omitempty"`        // CSS classes
	Styles         map[string]string `json:"styles,omitempty"`         // Inline styles
	LabelClass     string            `json:"labelClass,omitempty"`     // Label CSS classes
	InputClass     string            `json:"inputClass,omitempty"`     // Input CSS classes
	ErrorClass     string            `json:"errorClass,omitempty"`     // Error CSS classes
	ContainerClass string            `json:"containerClass,omitempty"` // Container CSS
}

// FieldEvents defines field-level event handlers
type FieldEvents struct {
	// Form lifecycle
	OnMount   string `json:"onMount,omitempty" validate:"js_function"`   // Form mounted
	OnUnmount string `json:"onUnmount,omitempty" validate:"js_function"` // Form unmounted

	// Submission lifecycle
	OnSubmit        string `json:"onSubmit,omitempty" validate:"js_function"`        // Submit triggered
	BeforeSubmit    string `json:"beforeSubmit,omitempty" validate:"js_function"`    // Before submit
	AfterSubmit     string `json:"afterSubmit,omitempty" validate:"js_function"`     // After submit
	OnSubmitSuccess string `json:"onSubmitSuccess,omitempty" validate:"js_function"` // Submit success
	OnSubmitError   string `json:"onSubmitError,omitempty" validate:"js_function"`   // Submit error

	// Form state
	OnReset    string `json:"onReset,omitempty" validate:"js_function"`    // Form reset
	OnValidate string `json:"onValidate,omitempty" validate:"js_function"` // Validation
	OnChange   string `json:"onChange,omitempty" validate:"js_function"`   // Any field changed
	OnDirty    string `json:"onDirty,omitempty" validate:"js_function"`    // Form becomes dirty
	OnPristine string `json:"onPristine,omitempty" validate:"js_function"` // Form becomes pristine

	// Data events
	OnLoad  string `json:"onLoad,omitempty" validate:"js_function"`  // Data loaded
	OnError string `json:"onError,omitempty" validate:"js_function"` // Error occurred

	// Field events
	OnFieldChange   string `json:"onFieldChange,omitempty" validate:"js_function"`   // Individual field changed
	OnFieldFocus    string `json:"onFieldFocus,omitempty" validate:"js_function"`    // Field focused
	OnFieldBlur     string `json:"onFieldBlur,omitempty" validate:"js_function"`     // Field blurred
	OnFieldValidate string `json:"onFieldValidate,omitempty" validate:"js_function"` // Field validated
}

// Conditional defines when field is shown/required using condition engine
// Supports both simple format (for ease of use) and full condition package format (for advanced use)
type Conditional struct {
	// Simple format (recommended for most use cases)
	Show     *ConditionGroup `json:"show,omitempty"`     // Show field when true
	Hide     *ConditionGroup `json:"hide,omitempty"`     // Hide field when true
	Required *ConditionGroup `json:"required,omitempty"` // Required when true
	Disabled *ConditionGroup `json:"disabled,omitempty"` // Disabled when true

	// Advanced format (use condition package directly for complex scenarios)
	ShowAdvanced     *condition.ConditionGroup `json:"showAdvanced,omitempty"`     // Advanced show conditions
	HideAdvanced     *condition.ConditionGroup `json:"hideAdvanced,omitempty"`     // Advanced hide conditions
	RequiredAdvanced *condition.ConditionGroup `json:"requiredAdvanced,omitempty"` // Advanced required conditions
	DisabledAdvanced *condition.ConditionGroup `json:"disabledAdvanced,omitempty"` // Advanced disabled conditions
}

// ConditionGroup wraps simple condition logic (backward compatible)
// For more complex conditions, use ShowAdvanced/HideAdvanced with condition.ConditionGroup
type ConditionGroup struct {
	Logic      string      `json:"logic" validate:"oneof=AND OR"` // AND or OR
	Conditions []Condition `json:"conditions" validate:"dive"`    // List of conditions
}

// Condition represents a single condition check (backward compatible)
type Condition struct {
	Field    string `json:"field" validate:"required"`    // Field to check
	Operator string `json:"operator" validate:"required"` // Comparison operator (equal, not_equal, greater, less, etc.)
	Value    any    `json:"value"`                        // Value to compare against
}

// ToConditionGroup converts simple ConditionGroup to condition package format
func (cg *ConditionGroup) ToConditionGroup() *condition.ConditionGroup {
	if cg == nil {
		return nil
	}

	var conjunction condition.Conjunction
	if strings.ToUpper(cg.Logic) == "OR" {
		conjunction = condition.ConjunctionOr
	} else {
		conjunction = condition.ConjunctionAnd
	}

	builder := condition.NewBuilder(conjunction)

	for i, cond := range cg.Conditions {
		// Map operator string to OperatorType
		op := mapOperatorType(cond.Operator)

		// Add rule with unique ID
		ruleID := fmt.Sprintf("rule-%d", i+1)
		builder.AddRuleWithID(ruleID, cond.Field, op, cond.Value)
	}

	return builder.Build()
}

// mapOperatorType maps string operators to condition.OperatorType
func mapOperatorType(op string) condition.OperatorType {
	switch strings.ToLower(op) {
	case "equal", "equals", "==", "eq":
		return condition.OpEqual
	case "not_equal", "not_equals", "!=", "neq":
		return condition.OpNotEqual
	case "greater", "greater_than", ">", "gt":
		return condition.OpGreater
	case "greater_or_equal", "greater_than_or_equal", ">=", "gte":
		return condition.OpGreaterOrEqual
	case "less", "less_than", "<", "lt":
		return condition.OpLess
	case "less_or_equal", "less_than_or_equal", "<=", "lte":
		return condition.OpLessOrEqual
	case "between":
		return condition.OpBetween
	case "not_between":
		return condition.OpNotBetween
	case "contains":
		return condition.OpContains
	case "not_contains":
		return condition.OpNotContains
	case "starts_with":
		return condition.OpStartsWith
	case "ends_with":
		return condition.OpEndsWith
	case "in", "select_any_in":
		return condition.OpIn
	case "not_in", "select_not_any_in":
		return condition.OpNotIn
	case "is_empty":
		return condition.OpIsEmpty
	case "is_not_empty":
		return condition.OpIsNotEmpty
	case "match_regexp", "regex":
		return condition.OpMatchRegexp
	default:
		return condition.OpEqual // Default to equal
	}
}

// DataSource defines where to fetch dynamic options
type DataSource struct {
	Type      string            `json:"type" validate:"oneof=api static computed"` // Source type
	URL       string            `json:"url,omitempty" validate:"url"`              // API endpoint
	Method    string            `json:"method,omitempty" validate:"oneof=GET POST"`
	Headers   map[string]string `json:"headers,omitempty"`   // Request headers
	Params    map[string]string `json:"params,omitempty"`    // Query parameters
	CacheTTL  int               `json:"cacheTTL,omitempty"`  // Cache duration in seconds
	Static    []Option          `json:"static,omitempty"`    // Static options
	Computed  string            `json:"computed,omitempty"`  // JS function for computed options
	Transform string            `json:"transform,omitempty"` // Transform response data
}

// FieldPermissions controls field access
type FieldPermissions struct {
	View     []string `json:"view,omitempty"`     // Roles that can view
	Edit     []string `json:"edit,omitempty"`     // Roles that can edit
	Required []string `json:"required,omitempty"` // Permissions needed
}

// FieldRuntime holds runtime state set by the enricher
type FieldRuntime struct {
	Visible  bool   `json:"visible"`  // Whether field should be visible to user
	Editable bool   `json:"editable"` // Whether field can be edited by user
	Reason   string `json:"reason"`   // Reason for visibility/editability state
}

// FieldI18n holds field translations for multiple locales
type FieldI18n struct {
	Label       map[string]string `json:"label,omitempty"`       // Localized labels
	Placeholder map[string]string `json:"placeholder,omitempty"` // Localized placeholders
	Help        map[string]string `json:"help,omitempty"`        // Localized help text
	Tooltip     map[string]string `json:"tooltip,omitempty"`     // Localized tooltips
	Error       map[string]string `json:"error,omitempty"`       // Localized error messages
	// settings
	Enabled          bool              `json:"enabled"` // Enable i18n
	DefaultLocale    string            `json:"defaultLocale" validate:"locale" example:"en-US"`
	SupportedLocales []string          `json:"supportedLocales" validate:"dive,locale"`
	Translations     map[string]string `json:"translations,omitempty"` // Translation keys
	DateFormat       string            `json:"dateFormat,omitempty" example:"MM/DD/YYYY"`
	TimeFormat       string            `json:"timeFormat,omitempty" example:"HH:mm:ss"`
	NumberFormat     string            `json:"numberFormat,omitempty" example:"1,000.00"`
	CurrencyFormat   string            `json:"currencyFormat,omitempty" example:"$1,000.00"`
	Currency         string            `json:"currency,omitempty" validate:"iso4217" example:"USD"`
	Direction        string            `json:"direction,omitempty" validate:"oneof=ltr rtl" example:"ltr"`
	FallbackLocale   string            `json:"fallbackLocale,omitempty" validate:"locale"`
	LoadPath         string            `json:"loadPath,omitempty"` // Path to translation files
}

// FieldHTMX defines HTMX behavior for this field
type FieldHTMX struct {
	Enabled     bool   `json:"enabled"`                                     // Enable Alpine.js
	XData       string `json:"xData,omitempty" validate:"js_object"`        // Component data
	XInit       string `json:"xInit,omitempty" validate:"js_function"`      // Initialization
	XShow       string `json:"xShow,omitempty" validate:"js_expression"`    // Show/hide
	XIf         string `json:"xIf,omitempty" validate:"js_expression"`      // Conditional
	XModel      string `json:"xModel,omitempty" validate:"js_variable"`     // Two-way binding
	XBind       string `json:"xBind,omitempty" validate:"js_object"`        // Attribute bindings
	XOn         string `json:"xOn,omitempty" validate:"js_object"`          // Event handlers
	XText       string `json:"xText,omitempty" validate:"js_expression"`    // Text content
	XHTML       string `json:"xHtml,omitempty" validate:"js_expression"`    // HTML content
	XRef        string `json:"xRef,omitempty" validate:"js_variable"`       // Reference
	XCloak      bool   `json:"xCloak,omitempty"`                            // Cloak until ready
	XTransition string `json:"xTransition,omitempty"`                       // Transition effect
	XTeleport   string `json:"xTeleport,omitempty" validate:"css_selector"` // Teleport target
	XFor        string `json:"xFor,omitempty" validate:"js_expression"`     // Loop directive
	XEffect     string `json:"xEffect,omitempty" validate:"js_function"`    // Side effects
}

// FieldAlpine defines Alpine.js bindings
type FieldAlpine struct {
	XModel string `json:"xModel,omitempty" validate:"js_variable"`  // Two-way binding
	XBind  string `json:"xBind,omitempty" validate:"js_object"`     // Attribute bindings
	XOn    string `json:"xOn,omitempty" validate:"js_object"`       // Event handlers
	XShow  string `json:"xShow,omitempty" validate:"js_expression"` // Show/hide
	XIf    string `json:"xIf,omitempty" validate:"js_expression"`   // Conditional render
}

// FieldSecurity defines security settings
type FieldSecurity struct {
	Sanitize     bool     `json:"sanitize,omitempty"`     // XSS protection
	AllowedTags  []string `json:"allowedTags,omitempty"`  // For richtext (e.g., ["p", "strong", "em"])
	MaxFileSize  int64    `json:"maxFileSize,omitempty"`  // File security limit
	AllowedMimes []string `json:"allowedMimes,omitempty"` // Allowed MIME types for files
}

// Option represents a selectable option for select/radio/checkbox fields
type Option struct {
	Value       string   `json:"value" validate:"required"`                // Option value
	Label       string   `json:"label" validate:"required"`                // Display label
	Description string   `json:"description,omitempty" validate:"max=200"` // Extra description
	Icon        string   `json:"icon,omitempty" validate:"icon_name"`      // Icon to show
	Color       string   `json:"color,omitempty" validate:"css_color"`     // Color indicator
	Group       string   `json:"group,omitempty"`                          // Option group
	Disabled    bool     `json:"disabled,omitempty"`                       // Cannot be selected
	Selected    bool     `json:"selected,omitempty"`                       // Pre-selected
	Children    []Option `json:"children,omitempty" validate:"dive"`       // Nested options (tree)
	Meta        any      `json:"meta,omitempty"`                           // Custom metadata
}

// Typed Field Configurations

// TextFieldConfig for text input fields
type TextFieldConfig struct {
	MaxLength    int    `json:"maxLength,omitempty"`
	Autocomplete string `json:"autocomplete,omitempty"`
	Prefix       string `json:"prefix,omitempty"`
	Suffix       string `json:"suffix,omitempty"`
}

// SelectFieldConfig for selection fields
type SelectFieldConfig struct {
	Searchable    bool   `json:"searchable,omitempty"`
	Clearable     bool   `json:"clearable,omitempty"`
	Multiple      bool   `json:"multiple,omitempty"`
	Placeholder   string `json:"placeholder,omitempty"`
	MaxItems      *int   `json:"maxItems,omitempty"`
	OptionsSource string `json:"optionsSource,omitempty"` // API endpoint for dynamic options
	Createable    bool   `json:"createable,omitempty"`    // Allow creating new options
}

// FileFieldConfig for file upload fields
type FileFieldConfig struct {
	MaxSize     int64    `json:"maxSize,omitempty"`
	Accept      []string `json:"accept,omitempty"`
	Multiple    bool     `json:"multiple,omitempty"`
	UploadURL   string   `json:"uploadUrl,omitempty"`
	DownloadURL string   `json:"downloadUrl,omitempty"`
	AutoUpload  bool     `json:"autoUpload,omitempty"`
	Drag        bool     `json:"drag,omitempty"` // Enable drag-and-drop
}

// ImageFieldConfig for image upload fields
type ImageFieldConfig struct {
	FileFieldConfig
	Crop        bool    `json:"crop,omitempty"`
	AspectRatio float64 `json:"aspectRatio,omitempty"` // e.g., 1.0 for square, 16/9 for landscape
	Resize      *struct {
		MaxWidth  int     `json:"maxWidth,omitempty"`
		MaxHeight int     `json:"maxHeight,omitempty"`
		Quality   float64 `json:"quality,omitempty"` // 0.0 to 1.0
	} `json:"resize,omitempty"`
	Preview bool `json:"preview,omitempty"` // Show image preview
}

// RelationConfig for foreign key fields
type RelationConfig struct {
	TargetSchema string         `json:"targetSchema"`        // Related table/schema
	DisplayField string         `json:"displayField"`        // Field to display in dropdown
	SearchFields []string       `json:"searchFields"`        // Fields to search
	ValueField   string         `json:"valueField"`          // Field to use as value (typically "id")
	Filters      map[string]any `json:"filters,omitempty"`   // Additional filters for related records
	SortBy       string         `json:"sortBy,omitempty"`    // Sort field
	SortOrder    string         `json:"sortOrder,omitempty"` // "asc" or "desc"
	CreateNew    bool           `json:"createNew,omitempty"` // Allow creating new records inline
	CreateURL    string         `json:"createUrl,omitempty"` // URL for creating new records
	Template     string         `json:"template,omitempty"`  // Display template e.g., "${code} - ${name}"
}

// RepeatableConfig for repeatable field groups
type RepeatableConfig struct {
	MinItems         int     `json:"minItems,omitempty"`
	MaxItems         int     `json:"maxItems,omitempty"`
	AddButtonText    string  `json:"addButtonText,omitempty"`
	RemoveButtonText string  `json:"removeButtonText,omitempty"`
	Sortable         bool    `json:"sortable,omitempty"`
	Collapsible      bool    `json:"collapsible,omitempty"`
	ItemTitle        string  `json:"itemTitle,omitempty"` // Template for item title, e.g., "Item ${index + 1}"
	Fields           []Field `json:"fields"`
}

// FormulaConfig for computed fields
type FormulaConfig struct {
	Expression   string   `json:"expression" validate:"required,max=10000"`              // Formula expression using condition engine syntax
	Format       string   `json:"format,omitempty"`                                      // Output format: currency, number, percent, date, datetime
	Precision    int      `json:"precision,omitempty" validate:"min=0,max=10"`           // Decimal places
	Recalculate  string   `json:"recalculate" validate:"oneof=onChange onSubmit onLoad"` // When to recalculate
	Dependencies []string `json:"dependencies,omitempty"`                                // Fields this formula depends on
}

// TextareaConfig for textarea fields
type TextareaConfig struct {
	Rows      int    `json:"rows,omitempty"`
	MaxLength int    `json:"maxLength,omitempty"`
	ShowCount bool   `json:"showCount,omitempty"` // Display character counter
	AutoSize  bool   `json:"autoSize,omitempty"`  // Auto-grow height
	Resize    string `json:"resize,omitempty"`    // "none", "vertical", "horizontal", "both"
}

// RichTextConfig for WYSIWYG editor
type RichTextConfig struct {
	Toolbar      []string `json:"toolbar,omitempty"` // Available formatting tools
	Height       int      `json:"height,omitempty"`
	MaxLength    int      `json:"maxLength,omitempty"`
	AllowImages  bool     `json:"allowImages,omitempty"`
	AllowTables  bool     `json:"allowTables,omitempty"`
	Sanitize     bool     `json:"sanitize,omitempty"`     // XSS protection
	OutputFormat string   `json:"outputFormat,omitempty"` // "html", "markdown", "json"
}

// CodeConfig for code editor
type CodeConfig struct {
	Language    string `json:"language,omitempty"` // "sql", "javascript", "python", "json", etc.
	Theme       string `json:"theme,omitempty"`    // Editor theme
	Height      int    `json:"height,omitempty"`
	LineNumbers bool   `json:"lineNumbers,omitempty"`
	Minimap     bool   `json:"minimap,omitempty"`
	WordWrap    bool   `json:"wordWrap,omitempty"`
	ReadOnly    bool   `json:"readOnly,omitempty"`
}

// CurrencyConfig for currency fields
type CurrencyConfig struct {
	Currency       string `json:"currency,omitempty"`       // ISO currency code (USD, EUR, KES, etc.)
	Locale         string `json:"locale,omitempty"`         // Locale for formatting (en-US, en-GB, etc.)
	Decimals       int    `json:"decimals,omitempty"`       // Decimal places (typically 2)
	ShowSymbol     bool   `json:"showSymbol,omitempty"`     // Display currency symbol
	SymbolPosition string `json:"symbolPosition,omitempty"` // "before" or "after"
}

// AutocompleteConfig for autocomplete fields
type AutocompleteConfig struct {
	Source       string `json:"source"`                 // API endpoint for search
	MinChars     int    `json:"minChars,omitempty"`     // Minimum characters before search
	Debounce     int    `json:"debounce,omitempty"`     // Delay in milliseconds
	MaxResults   int    `json:"maxResults,omitempty"`   // Maximum suggestions to show
	DisplayField string `json:"displayField,omitempty"` // Field to display
	ValueField   string `json:"valueField,omitempty"`   // Field to use as value
	Template     string `json:"template,omitempty"`     // Display template with placeholders
}

// RatingConfig for rating fields
type RatingConfig struct {
	Max          int      `json:"max,omitempty"`          // Maximum rating value (typically 5)
	AllowHalf    bool     `json:"allowHalf,omitempty"`    // Allow half-star ratings
	AllowClear   bool     `json:"allowClear,omitempty"`   // Allow clearing rating
	Character    string   `json:"character,omitempty"`    // Rating character/icon (default "★")
	ShowTooltips bool     `json:"showTooltips,omitempty"` // Show tooltips for each rating
	Tooltips     []string `json:"tooltips,omitempty"`     // Tooltip text for each rating level
}

// SliderConfig for slider fields
type SliderConfig struct {
	Min       float64           `json:"min,omitempty"`
	Max       float64           `json:"max,omitempty"`
	Step      float64           `json:"step,omitempty"`
	Marks     map[string]string `json:"marks,omitempty"`     // Labels at specific values
	ShowValue bool              `json:"showValue,omitempty"` // Display current value
	Range     bool              `json:"range,omitempty"`     // Enable range selection (two handles)
}

// DateConfig for date fields
type DateConfig struct {
	Format         string `json:"format,omitempty"`         // Internal date format (YYYY-MM-DD)
	DisplayFormat  string `json:"displayFormat,omitempty"`  // User-facing format (MM/DD/YYYY)
	MinDate        string `json:"minDate,omitempty"`        // Minimum selectable date
	MaxDate        string `json:"maxDate,omitempty"`        // Maximum selectable date
	FirstDayOfWeek int    `json:"firstDayOfWeek,omitempty"` // 0=Sunday, 1=Monday
	Shortcuts      bool   `json:"shortcuts,omitempty"`      // Show "Today", "Yesterday" buttons
}

// TimeConfig for time fields
type TimeConfig struct {
	Format     string `json:"format,omitempty"`     // Time format (HH:mm or HH:mm:ss)
	Use24Hour  bool   `json:"use24Hour,omitempty"`  // 24-hour vs 12-hour format
	MinuteStep int    `json:"minuteStep,omitempty"` // Minute increment (e.g., 15, 30)
}

// DateTimeConfig for datetime fields
type DateTimeConfig struct {
	Format        string `json:"format,omitempty"`        // Internal format
	DisplayFormat string `json:"displayFormat,omitempty"` // User-facing format
	Timezone      string `json:"timezone,omitempty"`      // Timezone (UTC, America/New_York, etc.)
	UTC           bool   `json:"utc,omitempty"`           // Store in UTC
}

// DateRangeConfig for daterange fields
type DateRangeConfig struct {
	StartField     string `json:"startField,omitempty"`     // Field name for start date
	EndField       string `json:"endField,omitempty"`       // Field name for end date
	MinDays        int    `json:"minDays,omitempty"`        // Minimum range in days
	MaxDays        int    `json:"maxDays,omitempty"`        // Maximum range in days
	SingleCalendar bool   `json:"singleCalendar,omitempty"` // Show both dates in one calendar
}

// TagsConfig for tags fields
type TagsConfig struct {
	Delimiter         string   `json:"delimiter,omitempty"`         // Tag separator (typically ",")
	MaxTags           int      `json:"maxTags,omitempty"`           // Maximum number of tags
	AllowCustom       bool     `json:"allowCustom,omitempty"`       // Allow creating new tags
	Suggestions       []string `json:"suggestions,omitempty"`       // Pre-defined tag suggestions
	SuggestionsSource string   `json:"suggestionsSource,omitempty"` // API for tag suggestions
}

// ColorConfig for color picker
type ColorConfig struct {
	Format  string   `json:"format,omitempty"`  // "hex", "rgb", "rgba", "hsl"
	Alpha   bool     `json:"alpha,omitempty"`   // Support transparency
	Presets []string `json:"presets,omitempty"` // Quick-select colors
}

// GroupConfig for field groups
type GroupConfig struct {
	Collapsible bool `json:"collapsible,omitempty"` // Can collapse/expand
	Collapsed   bool `json:"collapsed,omitempty"`   // Initial collapsed state
	Border      bool `json:"border,omitempty"`      // Show border around group
}

// SetEvaluator injects the condition evaluator for conditional logic
func (f *Field) SetEvaluator(evaluator *condition.Evaluator) {
	f.evaluator = evaluator
}

// IsVisible checks if field should be displayed given current form data
// Supports both simple and advanced condition formats
func (f *Field) IsVisible(ctx context.Context, data map[string]any) (bool, error) {
	if f.Hidden {
		return false, nil
	}

	if f.Conditional == nil {
		return true, nil
	}

	// Check Hide conditions first (takes precedence)
	// Try advanced format first, then simple format
	hideCondition := f.Conditional.HideAdvanced
	if hideCondition == nil && f.Conditional.Hide != nil {
		// Convert simple format to advanced format
		hideCondition = f.Conditional.Hide.ToConditionGroup()
	}

	if hideCondition != nil {
		if f.evaluator == nil {
			// For visibility checks, return false (visible) if evaluator is not set
			// This allows graceful degradation when conditions are defined but evaluator is missing
			return false, nil
		}
		evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
		hide, err := f.evaluator.Evaluate(ctx, hideCondition, evalCtx)
		if err != nil {
			return false, fmt.Errorf("error evaluating hide condition for field %s: %w", f.Name, err)
		}
		if hide {
			return false, nil
		}
	}

	// Check Show conditions
	// Try advanced format first, then simple format
	showCondition := f.Conditional.ShowAdvanced
	if showCondition == nil && f.Conditional.Show != nil {
		// Convert simple format to advanced format
		showCondition = f.Conditional.Show.ToConditionGroup()
	}

	if showCondition != nil {
		if f.evaluator == nil {
			// For visibility checks, return false (visible) if evaluator is not set
			// This allows graceful degradation when conditions are defined but evaluator is missing
			return false, nil
		}
		evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
		show, err := f.evaluator.Evaluate(ctx, showCondition, evalCtx)
		if err != nil {
			return false, fmt.Errorf("error evaluating show condition for field %s: %w", f.Name, err)
		}
		return show, nil
	}

	return true, nil
}

// IsRequired checks if field is required given current form data
// Supports both simple and advanced condition formats
func (f *Field) IsRequired(ctx context.Context, data map[string]any) (bool, error) {
	if f.Required {
		return true, nil
	}

	if f.Conditional == nil {
		return false, nil
	}

	// Try advanced format first, then simple format
	requiredCondition := f.Conditional.RequiredAdvanced
	if requiredCondition == nil && f.Conditional.Required != nil {
		// Convert simple format to advanced format
		requiredCondition = f.Conditional.Required.ToConditionGroup()
	}

	if requiredCondition != nil {
		if f.evaluator == nil {
			// For required checks, return false (not required) if evaluator is not set
			// This allows graceful degradation when conditions are defined but evaluator is missing
			return false, nil
		}
		evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
		required, err := f.evaluator.Evaluate(ctx, requiredCondition, evalCtx)
		if err != nil {
			return false, fmt.Errorf("error evaluating required condition for field %s: %w", f.Name, err)
		}
		return required, nil
	}

	return false, nil
}

// IsDisabled checks if field is disabled given current form data
// Supports both simple and advanced condition formats
func (f *Field) IsDisabled(ctx context.Context, data map[string]any) (bool, error) {
	if f.Disabled {
		return true, nil
	}

	if f.Conditional == nil {
		return false, nil
	}

	// Try advanced format first, then simple format
	disabledCondition := f.Conditional.DisabledAdvanced
	if disabledCondition == nil && f.Conditional.Disabled != nil {
		// Convert simple format to advanced format
		disabledCondition = f.Conditional.Disabled.ToConditionGroup()
	}

	if disabledCondition != nil {
		if f.evaluator == nil {
			// For disabled checks, return false (not disabled) if evaluator is not set
			// This allows graceful degradation when conditions are defined but evaluator is missing
			return false, nil
		}
		evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
		disabled, err := f.evaluator.Evaluate(ctx, disabledCondition, evalCtx)
		if err != nil {
			return false, fmt.Errorf("error evaluating disabled condition for field %s: %w", f.Name, err)
		}
		return disabled, nil
	}

	return false, nil
}

// Validate checks if field configuration is valid
func (f *Field) Validate(ctx context.Context) error {
	if f.Name == "" {
		return ErrInvalidFieldName
	}
	if f.Type == "" {
		return ErrInvalidFieldType
	}

	// Validate validation rules
	if f.Validation != nil {
		if f.Validation.Min != nil && f.Validation.Max != nil {
			if *f.Validation.Min > *f.Validation.Max {
				return NewValidationError(
					"invalid_validation_range",
					"min cannot be greater than max",
				).WithField(f.Name)
			}
		}
		if f.Validation.MinLength != nil && f.Validation.MaxLength != nil {
			if *f.Validation.MinLength > *f.Validation.MaxLength {
				return NewValidationError(
					"invalid_length_range",
					"minLength cannot be greater than maxLength",
				).WithField(f.Name)
			}
		}
		if f.Validation.MinItems != nil && f.Validation.MaxItems != nil {
			if *f.Validation.MinItems > *f.Validation.MaxItems {
				return NewValidationError(
					"invalid_items_range",
					"minItems cannot be greater than maxItems",
				).WithField(f.Name)
			}
		}

		// Validate pattern if present
		if f.Validation.Pattern != "" {
			if _, err := regexp.Compile(f.Validation.Pattern); err != nil {
				return NewValidationError(
					"invalid_pattern",
					fmt.Sprintf("invalid regex pattern: %v", err),
				).WithField(f.Name)
			}
		}
	}

	// Validate options for select-type fields
	if f.requiresOptions() && len(f.Options) == 0 && f.DataSource == nil {
		return NewValidationError(
			"missing_options",
			"field type requires options or dataSource",
		).WithField(f.Name)
	}

	// Validate formula fields
	if f.Type == FieldFormula {
		var formulaCfg FormulaConfig
		if err := mapstructure.Decode(f.Config, &formulaCfg); err != nil {
			return NewValidationError(
				"invalid_formula_config",
				fmt.Sprintf("invalid formula configuration: %v", err),
			).WithField(f.Name)
		}
		if formulaCfg.Expression == "" {
			return NewValidationError(
				"missing_formula_expression",
				"formula field requires expression",
			).WithField(f.Name)
		}
	}

	return nil
}

// requiresOptions checks if field type requires options
func (f *Field) requiresOptions() bool {
	switch f.Type {
	case FieldSelect, FieldMultiSelect, FieldRadio, FieldCheckboxes, FieldTreeSelect, FieldCascader, FieldTransfer:
		return true
	}
	return false
}

// validateValueExcludingCustom validates a value against field rules, excluding custom condition engine validation
func (f *Field) validateValueExcludingCustom(ctx context.Context, value any) error {
	// Check for context cancellation before starting validation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue with validation
	}

	// Required check
	if f.Required && isEmpty(value) {
		var requiredMsg string
		if f.Validation != nil && f.Validation.Messages != nil {
			requiredMsg = f.Validation.Messages.Required
		}
		return f.validationError(requiredMsg, "%s is required", f.Label)
	}

	// Skip validation if no value and not required
	if isEmpty(value) {
		return nil
	}

	// Type compatibility check
	if !f.isValueTypeCompatible(value) {
		return f.validationError("", "invalid type for field %s", f.Label)
	}

	if f.Validation == nil {
		return nil
	}

	// Type-specific validation using switch for better performance and clarity
	switch val := value.(type) {
	case string:
		if err := f.validateString(val); err != nil {
			return fmt.Errorf("string validation failed for field %s: %w", f.Label, err)
		}
	case []any:
		if err := f.validateArray(val); err != nil {
			return fmt.Errorf("array validation failed for field %s: %w", f.Label, err)
		}
	default:
		// Handle numbers and other types
		if err := f.validateNumber(val); err != nil {
			return fmt.Errorf("number validation failed for field %s: %w", f.Label, err)
		}
	}

	// Skip custom validation - will be handled by registry
	return nil
}

// isValueTypeCompatible checks if the value type is compatible with field expectations
func (f *Field) isValueTypeCompatible(value any) bool {
	// Implement type compatibility logic based on field configuration
	// This could check against f.Type or other field metadata
	switch value.(type) {
	case string:
		return f.isStringTypeCompatible()
	case []any:
		return f.isArrayTypeCompatible()
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return f.isNumberTypeCompatible()
	default:
		return f.isOtherTypeCompatible()
	}
}

// Helper methods for type compatibility (to be implemented based on your field configuration)
func (f *Field) isStringTypeCompatible() bool {
	// Check if field expects string type
	return true // default implementation
}

func (f *Field) isArrayTypeCompatible() bool {
	// Check if field expects array type
	return true // default implementation
}

func (f *Field) isNumberTypeCompatible() bool {
	// Check if field expects number type
	return true // default implementation
}

func (f *Field) isOtherTypeCompatible() bool {
	// Check if field expects other specific types
	return true // default implementation
}

// ValidateValue validates a value against field rules
func (f *Field) ValidateValue(ctx context.Context, value any) error {
	// Required check
	if f.Required && isEmpty(value) {
		var requiredMsg string
		if f.Validation != nil {
			requiredMsg = f.Validation.Messages.Required
		}
		return f.validationError(requiredMsg, "%s is required", f.Label)
	}

	// Skip validation if no value and not required
	if isEmpty(value) {
		return nil
	}

	if f.Validation == nil {
		return nil
	}

	// String validation
	if strVal, ok := value.(string); ok {
		if err := f.validateString(strVal); err != nil {
			return err
		}
	}

	// Number validation
	if err := f.validateNumber(value); err != nil {
		return err
	}

	// Array validation
	if arr, ok := value.([]any); ok {
		if err := f.validateArray(arr); err != nil {
			return err
		}
	}

	// Custom validation using condition engine
	if f.Validation.Custom != "" {
		if err := f.validateCustom(ctx, value); err != nil {
			return err
		}
	}

	return nil
}

// validateString performs string-specific validation
func (f *Field) validateString(value string) error {
	v := f.Validation

	// Length validation
	if v.MinLength != nil && len(value) < *v.MinLength {
		msg := ""
		if v.Messages != nil {
			msg = v.Messages.MinLength
		}
		return f.validationError(msg,
			"%s must be at least %d characters", f.Label, *v.MinLength)
	}
	if v.MaxLength != nil && len(value) > *v.MaxLength {
			msg := ""
		if v.Messages != nil {
			msg = v.Messages.MaxLength
		}
		return f.validationError(msg,
			"%s must be at most %d characters", f.Label, *v.MaxLength)
	}

	// Pattern validation
	if v.Pattern != "" {
		matched, err := regexp.MatchString(v.Pattern, value)
		if err != nil {
			return NewValidationError("invalid_pattern",
				"invalid validation pattern").WithField(f.Name).WithDetail("pattern", v.Pattern).WithDetail("error", err.Error())
		}
		if !matched {
			msg := ""
			if v.Messages != nil {
				msg = v.Messages.Pattern
			}
			return f.validationError(msg,
				"%s format is invalid", f.Label)
		}
	}

	// Format validation
	if v.Format != "" {
		if err := f.validateFormat(value); err != nil {
			return err
		}
	}

	return nil
}

// validateFormat validates specific string formats
func (f *Field) validateFormat(value string) error {
	switch f.Validation.Format {
	case "email":
		if !isValidEmail(value) {
			return NewValidationError("invalid_email",
				fmt.Sprintf("%s must be a valid email address", f.Label)).WithField(f.Name).WithDetail("value", value)
		}
	case "url":
		parsedURL, err := url.Parse(value)
		if err != nil || (!strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://")) {
			return NewValidationError("invalid_url",
				fmt.Sprintf("%s must be a valid URL", f.Label)).WithField(f.Name).WithDetail("value", value)
		}
		if parsedURL.Scheme == "" || parsedURL.Host == "" {
			return NewValidationError("invalid_url",
				fmt.Sprintf("%s must be a valid URL", f.Label)).WithField(f.Name).WithDetail("value", value)
		}
	case "uuid":
		if _, err := uuid.Parse(value); err != nil {
			return NewValidationError("invalid_uuid",
				fmt.Sprintf("%s must be a valid UUID", f.Label)).WithField(f.Name).WithDetail("value", value).WithDetail("error", err.Error())
		}
	case "date":
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return NewValidationError("invalid_date",
				fmt.Sprintf("%s must be a valid date (YYYY-MM-DD)", f.Label)).WithField(f.Name).WithDetail("value", value).WithDetail("error", err.Error())
		}
	case "datetime":
		if _, err := time.Parse(time.RFC3339, value); err != nil {
			return NewValidationError("invalid_datetime",
				fmt.Sprintf("%s must be a valid datetime (RFC3339)", f.Label)).WithField(f.Name).WithDetail("value", value).WithDetail("error", err.Error())
		}
	case "time":
		if _, err := time.Parse("15:04:05", value); err != nil {
			return NewValidationError("invalid_time",
				fmt.Sprintf("%s must be a valid time (HH:MM:SS)", f.Label)).WithField(f.Name).WithDetail("value", value).WithDetail("error", err.Error())
		}
	}
	return nil
}

// validateNumber performs number-specific validation
func (f *Field) validateNumber(value any) error {
	numVal, ok := toFloat64(value)
	if !ok {
		return nil // Not a number, skip
	}

	v := f.Validation

	if v.Min != nil {
		if v.ExclusiveMin && numVal <= *v.Min {
			return NewValidationError("number_too_small",
				fmt.Sprintf("%s must be greater than %v", f.Label, *v.Min)).WithField(f.Name).WithDetail("value", numVal).WithDetail("min", *v.Min).WithDetail("exclusive", true)
		}
		if !v.ExclusiveMin && numVal < *v.Min {
			msg := ""
			if v.Messages != nil {
				msg = v.Messages.Min
			}
			return f.validationError(msg,
				"%s must be at least %v", f.Label, *v.Min)
		}
	}

	if v.Max != nil {
		if v.ExclusiveMax && numVal >= *v.Max {
			return NewValidationError("number_too_large",
				fmt.Sprintf("%s must be less than %v", f.Label, *v.Max)).WithField(f.Name).WithDetail("value", numVal).WithDetail("max", *v.Max).WithDetail("exclusive", true)
		}
		if !v.ExclusiveMax && numVal > *v.Max {
			msg := ""
			if v.Messages != nil {
				msg = v.Messages.Max
			}
			return f.validationError(msg,
				"%s must be at most %v", f.Label, *v.Max)
		}
	}

	if v.Integer && numVal != math.Floor(numVal) {
		return NewValidationError("not_integer",
			fmt.Sprintf("%s must be an integer", f.Label)).WithField(f.Name).WithDetail("value", numVal)
	}

	if v.Positive && numVal <= 0 {
		return NewValidationError("not_positive",
			fmt.Sprintf("%s must be positive", f.Label)).WithField(f.Name).WithDetail("value", numVal)
	}

	if v.Negative && numVal >= 0 {
		return NewValidationError("not_negative",
			fmt.Sprintf("%s must be negative", f.Label)).WithField(f.Name).WithDetail("value", numVal)
	}

	if v.MultipleOf != nil && math.Mod(numVal, *v.MultipleOf) != 0 {
		return NewValidationError("not_multiple_of",
			fmt.Sprintf("%s must be a multiple of %v", f.Label, *v.MultipleOf)).WithField(f.Name).WithDetail("value", numVal).WithDetail("multipleOf", *v.MultipleOf)
	}

	if v.Step != nil && math.Mod(numVal, *v.Step) != 0 {
		return NewValidationError("invalid_step",
			fmt.Sprintf("%s must be in steps of %v", f.Label, *v.Step)).WithField(f.Name).WithDetail("value", numVal).WithDetail("step", *v.Step)
	}

	return nil
}

// validateArray performs array-specific validation
func (f *Field) validateArray(arr []any) error {
	v := f.Validation

	if v.MinItems != nil && len(arr) < *v.MinItems {
		return NewValidationError("array_too_short",
			fmt.Sprintf("%s must have at least %d items", f.Label, *v.MinItems)).WithField(f.Name).WithDetail("length", len(arr)).WithDetail("minItems", *v.MinItems)
	}

	if v.MaxItems != nil && len(arr) > *v.MaxItems {
		return NewValidationError("array_too_long",
			fmt.Sprintf("%s must have at most %d items", f.Label, *v.MaxItems)).WithField(f.Name).WithDetail("length", len(arr)).WithDetail("maxItems", *v.MaxItems)
	}

	if v.UniqueItems && !hasUniqueItems(arr) {
		return NewValidationError("array_not_unique",
			fmt.Sprintf("%s must have unique items", f.Label)).WithField(f.Name).WithDetail("length", len(arr))
	}

	return nil
}

// validateCustom performs custom validation using condition engine formulas
func (f *Field) validateCustom(ctx context.Context, value any) error {
	if f.evaluator == nil {
		return NewValidationError("condition_evaluator_missing",
			"condition evaluator not set for custom validation").WithField(f.Name)
	}

	// Build condition group with formula
	builder := condition.NewBuilder(condition.ConjunctionAnd)
	builder.AddFormula(f.Validation.Custom)
	group := builder.Build()

	// Create evaluation context with value
	data := map[string]any{
		"value": value,
		"field": f.Name,
	}

	evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
	valid, err := f.evaluator.Evaluate(ctx, group, evalCtx)
	if err != nil {
		return fmt.Errorf("custom validation error for %s: %w", f.Label, err)
	}

	if !valid {
		return f.validationError(f.Validation.Messages.Custom,
			"%s validation failed", f.Label)
	}

	return nil
}

// validationError returns custom error message if available, otherwise default
func (f *Field) validationError(customMsg, defaultFmt string, args ...any) SchemaError {
	var message string
	var code string

	if customMsg != "" {
		// For custom messages, return the message as-is for backward compatibility
		message = customMsg
		code = "custom_validation"
	} else {
		message = fmt.Sprintf(defaultFmt, args...)
		code = "field_validation"
	}

	// Create a custom error that returns just the message for custom validation
	if customMsg != "" {
		return &customValidationError{
			baseError: NewValidationError(code, message).WithField(f.Name),
			message:   customMsg,
		}
	}

	return NewValidationError(code, message).WithField(f.Name)
}

// customValidationError provides backward compatibility for custom messages
type customValidationError struct {
	baseError SchemaError
	message   string
}

func (e *customValidationError) Error() string {
	return e.message // Return just the custom message for backward compatibility
}

func (e *customValidationError) Code() string {
	return e.baseError.Code()
}

func (e *customValidationError) Type() ErrorType {
	return e.baseError.Type()
}

func (e *customValidationError) Field() string {
	return e.baseError.Field()
}

func (e *customValidationError) Details() map[string]any {
	return e.baseError.Details()
}

func (e *customValidationError) WithField(field string) SchemaError {
	return &customValidationError{
		baseError: e.baseError.WithField(field),
		message:   e.message,
	}
}

func (e *customValidationError) WithDetail(key string, value any) SchemaError {
	return &customValidationError{
		baseError: e.baseError.WithDetail(key, value),
		message:   e.message,
	}
}

// GetDefaultValue returns the field's default value, resolving dynamic expressions
func (f *Field) GetDefaultValue(ctx context.Context) (any, error) {
	if f.Default == nil {
		return f.getTypeDefaultValue(), nil
	}

	// Handle string defaults that may contain dynamic expressions
	if strDefault, ok := f.Default.(string); ok {
		// Check for dynamic values like ${now}, ${session.user_id}, etc.
		if strings.HasPrefix(strDefault, "${") && strings.HasSuffix(strDefault, "}") {
			return f.resolveDynamicDefault(ctx, strDefault)
		}
	}

	return f.Default, nil
}

// resolveDynamicDefault resolves dynamic default value expressions
func (f *Field) resolveDynamicDefault(ctx context.Context, expr string) (any, error) {
	// Remove ${ and }
	expr = strings.TrimPrefix(expr, "${")
	expr = strings.TrimSuffix(expr, "}")

	switch expr {
	case "now":
		return time.Now(), nil
	case "today":
		return time.Now().Truncate(24 * time.Hour).Format("2006-01-02"), nil
	case "uuid":
		return uuid.New().String(), nil
	case "session.user_id":
		if schemaCtx, ok := ctx.Value("schema_context").(*Context); ok {
			return schemaCtx.UserID, nil
		}
		return nil, fmt.Errorf("user_id not in context")
	case "session.tenant_id":
		if schemaCtx, ok := ctx.Value("schema_context").(*Context); ok {
			return schemaCtx.TenantID, nil
		}
		return nil, fmt.Errorf("tenant_id not in context")
	case "session.session_id":
		if schemaCtx, ok := ctx.Value("schema_context").(*Context); ok {
			return schemaCtx.SessionID, nil
		}
		return nil, fmt.Errorf("session_id not in context")
	default:
		// Try to evaluate as formula expression
		return f.evaluateDynamicFormula(ctx, expr)
	}
}

// evaluateDynamicFormula evaluates a formula expression for default value
func (f *Field) evaluateDynamicFormula(ctx context.Context, formula string) (any, error) {
	if f.evaluator == nil {
		return nil, NewInternalError("condition_evaluator_missing",
			"condition evaluator not set for formula evaluation").WithField(f.Name).WithDetail("formula", formula)
	}

	builder := condition.NewBuilder(condition.ConjunctionAnd)
	builder.AddFormula(formula)
	group := builder.Build()

	// Get context data
	data := make(map[string]any)
	if schemaCtx, ok := ctx.Value("schema_context").(*Context); ok {
		data["user_id"] = schemaCtx.UserID
		data["tenant_id"] = schemaCtx.TenantID
		data["now"] = time.Now()
	}

	evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
	result, err := f.evaluator.Evaluate(ctx, group, evalCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate dynamic default formula: %w", err)
	}

	return result, nil
}

// getTypeDefaultValue returns type-specific default values
func (f *Field) getTypeDefaultValue() any {
	switch f.Type {
	case FieldCheckbox, FieldSwitch:
		return false
	case FieldMultiSelect, FieldCheckboxes, FieldTags:
		return []string{}
	case FieldNumber, FieldCurrency, FieldSlider, FieldRating:
		return 0
	case FieldDate:
		return time.Now().Format("2006-01-02")
	case FieldDateTime:
		return time.Now().Format(time.RFC3339)
	case FieldTime:
		return time.Now().Format("15:04:05")
	default:
		return ""
	}
}

// ApplyTransform transforms the field value according to transform rules
func (f *Field) ApplyTransform(value any) (any, error) {
	if f.Transform == nil {
		return value, nil
	}

	strVal, ok := value.(string)
	if !ok {
		return value, nil // Only transform strings
	}

	switch f.Transform.Type {
	case "uppercase":
		return strings.ToUpper(strVal), nil
	case "lowercase":
		return strings.ToLower(strVal), nil
	case "trim":
		return strings.TrimSpace(strVal), nil
	case "capitalize":
		if len(strVal) == 0 {
			return strVal, nil
		}
		return strings.ToUpper(string(strVal[0])) + strings.ToLower(strVal[1:]), nil
	case "slugify":
		return f.slugify(strVal), nil
	default:
		return value, fmt.Errorf("unknown transform type: %s", f.Transform.Type)
	}
}

// slugify converts a string to a URL-friendly slug
func (f *Field) slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)
	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	// Remove non-alphanumeric characters except hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	s = reg.ReplaceAllString(s, "")
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	s = reg.ReplaceAllString(s, "-")
	// Trim hyphens from ends
	s = strings.Trim(s, "-")
	return s
}

// EvaluateFormula evaluates a formula field's expression
func (f *Field) EvaluateFormula(ctx context.Context, data map[string]any) (any, error) {
	if f.Type != FieldFormula {
		return nil, NewValidationError("invalid_field_type",
			"field is not a formula field").WithField(f.Name).WithDetail("actualType", string(f.Type)).WithDetail("expectedType", string(FieldFormula))
	}

	if f.evaluator == nil {
		return nil, NewInternalError("condition_evaluator_missing",
			"condition evaluator not set for formula field").WithField(f.Name)
	}

	var formulaCfg FormulaConfig
	if err := mapstructure.Decode(f.Config, &formulaCfg); err != nil {
		return nil, NewValidationError("invalid_formula_config",
			"invalid formula configuration").WithField(f.Name).WithDetail("config", f.Config).WithDetail("error", err.Error())
	}

	// Build condition group with formula
	builder := condition.NewBuilder(condition.ConjunctionAnd)
	builder.AddFormula(formulaCfg.Expression)
	group := builder.Build()

	// Evaluate using condition engine
	evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
	result, err := f.evaluator.Evaluate(ctx, group, evalCtx)
	if err != nil {
		return nil, NewValidationError("formula_evaluation_failed",
			"formula evaluation failed").WithField(f.Name).WithDetail("expression", formulaCfg.Expression).WithDetail("error", err.Error())
	}

	// Format result based on config
	return f.formatFormulaResult(result, formulaCfg)
}

// formatFormulaResult formats the formula result based on configuration
func (f *Field) formatFormulaResult(value any, cfg FormulaConfig) (any, error) {
	switch cfg.Format {
	case "currency":
		if num, ok := toFloat64(value); ok {
			return roundToPrecision(num, cfg.Precision), nil
		}
	case "number":
		if num, ok := toFloat64(value); ok {
			return roundToPrecision(num, cfg.Precision), nil
		}
	case "percent":
		if num, ok := toFloat64(value); ok {
			return roundToPrecision(num*100, cfg.Precision), nil
		}
	case "date":
		if t, ok := value.(time.Time); ok {
			return t.Format("2006-01-02"), nil
		}
	case "datetime":
		if t, ok := value.(time.Time); ok {
			return t.Format(time.RFC3339), nil
		}
	default:
		return value, nil
	}
	return value, nil
}

// Helper Methods

// IsSelectionType checks if field is a selection type
func (f *Field) IsSelectionType() bool {
	return f.requiresOptions()
}

// IsFileType checks if field handles files
func (f *Field) IsFileType() bool {
	return f.Type == FieldFile || f.Type == FieldImage ||
		f.Type == FieldSignature || f.Type == FieldVideo ||
		f.Type == FieldAudio
}

// IsNumericType checks if field is numeric
func (f *Field) IsNumericType() bool {
	return f.Type == FieldNumber || f.Type == FieldCurrency ||
		f.Type == FieldSlider || f.Type == FieldRating
}

// IsDateTimeType checks if field is date/time related
func (f *Field) IsDateTimeType() bool {
	return f.Type == FieldDate || f.Type == FieldTime ||
		f.Type == FieldDateTime || f.Type == FieldDateRange ||
		f.Type == FieldMonth || f.Type == FieldYear ||
		f.Type == FieldQuarter
}

// IsLayoutType checks if field is a layout container
func (f *Field) IsLayoutType() bool {
	return f.Type == FieldGroup || f.Type == FieldFieldset ||
		f.Type == FieldTabs || f.Type == FieldPanel ||
		f.Type == FieldCollapse
}

// GetOptionLabel returns label for option value
func (f *Field) GetOptionLabel(value string) string {
	return f.getOptionLabelRecursive(f.Options, value)
}

// getOptionLabelRecursive recursively searches for option label
func (f *Field) getOptionLabelRecursive(options []Option, value string) string {
	for _, opt := range options {
		if opt.Value == value {
			return opt.Label
		}
		// Check nested options
		if len(opt.Children) > 0 {
			if label := f.getOptionLabelRecursive(opt.Children, value); label != "" {
				return label
			}
		}
	}
	return value
}

// GetLabel returns localized label if available, otherwise default label
func (f *Field) GetLabel(locale string) string {
	if f.I18n != nil && f.I18n.Label != nil {
		if label, ok := f.I18n.Label[locale]; ok {
			return label
		}
	}
	return f.Label
}

// GetPlaceholder returns localized placeholder if available
func (f *Field) GetPlaceholder(locale string) string {
	if f.I18n != nil && f.I18n.Placeholder != nil {
		if placeholder, ok := f.I18n.Placeholder[locale]; ok {
			return placeholder
		}
	}
	return f.Placeholder
}

// GetHelp returns localized help text if available
func (f *Field) GetHelp(locale string) string {
	if f.I18n != nil && f.I18n.Help != nil {
		if help, ok := f.I18n.Help[locale]; ok {
			return help
		}
	}
	return f.Help
}

// GetTextConfig returns typed text field configuration
func (f *Field) GetTextConfig() (*TextFieldConfig, error) {
	var cfg TextFieldConfig
	if err := mapstructure.Decode(f.Config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid text config: %w", err)
	}
	return &cfg, nil
}

// GetSelectConfig returns typed select field configuration
func (f *Field) GetSelectConfig() (*SelectFieldConfig, error) {
	var cfg SelectFieldConfig
	if err := mapstructure.Decode(f.Config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid select config: %w", err)
	}
	return &cfg, nil
}

// GetFileConfig returns typed file field configuration
func (f *Field) GetFileConfig() (*FileFieldConfig, error) {
	var cfg FileFieldConfig
	if err := mapstructure.Decode(f.Config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid file config: %w", err)
	}
	return &cfg, nil
}

// GetRelationConfig returns typed relation field configuration
func (f *Field) GetRelationConfig() (*RelationConfig, error) {
	var cfg RelationConfig
	if err := mapstructure.Decode(f.Config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid relation config: %w", err)
	}
	return &cfg, nil
}

// GetRepeatableConfig returns typed repeatable field configuration
func (f *Field) GetRepeatableConfig() (*RepeatableConfig, error) {
	var cfg RepeatableConfig
	if err := mapstructure.Decode(f.Config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid repeatable config: %w", err)
	}
	return &cfg, nil
}

// GetFormulaConfig returns typed formula field configuration
func (f *Field) GetFormulaConfig() (*FormulaConfig, error) {
	var cfg FormulaConfig
	if err := mapstructure.Decode(f.Config, &cfg); err != nil {
		return nil, fmt.Errorf("invalid formula config: %w", err)
	}
	return &cfg, nil
}

// Utility Functions

// isEmpty checks if a value is empty/nil
func isEmpty(value any) bool {
	if value == nil {
		return true
	}
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		return false
	}
}

// hasUniqueItems checks if array items are unique
func hasUniqueItems(arr []any) bool {
	seen := make(map[string]bool)
	for _, item := range arr {
		key := fmt.Sprintf("%v", item)
		if seen[key] {
			return false
		}
		seen[key] = true
	}
	return true
}

// toFloat64 converts various numeric types to float64
func toFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case int16:
		return float64(v), true
	case int8:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint64:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint8:
		return float64(v), true
	default:
		return 0, false
	}
}

// roundToPrecision rounds a float to specified decimal places
func roundToPrecision(value float64, precision int) float64 {
	multiplier := math.Pow(10, float64(precision))
	return math.Round(value*multiplier) / multiplier
}

// isValidEmail validates email format using RFC 5322 basic validation
func isValidEmail(email string) bool {
	// Basic RFC 5322 validation
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
