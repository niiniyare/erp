package schema

import (
	"context"
	"fmt"
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
	Default any      `json:"default,omitempty"`                 // Default value
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

	// Framework integration
	HTMX   *FieldHTMX   `json:"htmx,omitempty"`   // HTMX attributes
	Alpine *FieldAlpine `json:"alpine,omitempty"` // Alpine.js bindings

	// Internal state (not in JSON)
	compiledCondition any `json:"-"` // Pre-compiled condition for performance
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

	// Text content
	FieldTextarea FieldType = "textarea"
	FieldRichText FieldType = "richtext" // WYSIWYG editor
	FieldCode     FieldType = "code"     // Code editor with syntax highlighting
	FieldJSON     FieldType = "json"     // JSON editor with validation

	// Selection
	FieldSelect      FieldType = "select"      // Dropdown
	FieldMultiSelect FieldType = "multiselect" // Multiple selection dropdown
	FieldRadio       FieldType = "radio"       // Radio buttons
	FieldCheckbox    FieldType = "checkbox"    // Checkboxes
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
	FieldSignature FieldType = "signature" // Signature pad

	// Specialized
	FieldCurrency     FieldType = "currency"     // Money input with formatting
	FieldTags         FieldType = "tags"         // Tag input
	FieldLocation     FieldType = "location"     // Address/location picker
	FieldRelation     FieldType = "relation"     // Foreign key relationship
	FieldAutoComplete FieldType = "autocomplete" // Autocomplete search

	// Display only
	FieldDisplay FieldType = "display" // Read-only display
	FieldDivider FieldType = "divider" // Visual separator
	FieldHTML    FieldType = "html"    // Raw HTML content
	
	// Collections
	FieldRepeatable   FieldType = "repeatable"   // Repeatable field groups
	FieldTableRepeater FieldType = "table_repeater" // Table-style repeatable fields
)

// FieldValidation defines validation rules for a field
type FieldValidation struct {
	// String validation
	MinLength *int   `json:"minLength,omitempty"` // Minimum string length
	MaxLength *int   `json:"maxLength,omitempty"` // Maximum string length
	Pattern   string `json:"pattern,omitempty"`   // Regex pattern
	Format    string `json:"format,omitempty"`    // Format validator (email, url, uuid, etc)

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

	// Custom validation
	Custom   string   `json:"custom,omitempty"`   // Custom validation function
	Messages Messages `json:"messages,omitempty"` // Custom error messages
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
	OnChange  string `json:"onChange,omitempty" validate:"js_function"`  // Value changed
	OnBlur    string `json:"onBlur,omitempty" validate:"js_function"`    // Lost focus
	OnFocus   string `json:"onFocus,omitempty" validate:"js_function"`   // Gained focus
	OnInput   string `json:"onInput,omitempty" validate:"js_function"`   // Input event
	OnKeyDown string `json:"onKeyDown,omitempty" validate:"js_function"` // Key pressed
	OnKeyUp   string `json:"onKeyUp,omitempty" validate:"js_function"`   // Key released
	OnMount   string `json:"onMount,omitempty" validate:"js_function"`   // Field mounted
	OnUnmount string `json:"onUnmount,omitempty" validate:"js_function"` // Field unmounted
}

// Conditional defines when field is shown/required
type Conditional struct {
	Show     *ConditionGroup `json:"show,omitempty"`     // Show field when true
	Hide     *ConditionGroup `json:"hide,omitempty"`     // Hide field when true
	Required *ConditionGroup `json:"required,omitempty"` // Required when true
	Disabled *ConditionGroup `json:"disabled,omitempty"` // Disabled when true
}

// ConditionGroup wraps condition package's group
type ConditionGroup struct {
	Logic      string      `json:"logic" validate:"oneof=AND OR"` // AND or OR
	Conditions []Condition `json:"conditions" validate:"dive"`    // List of conditions
}

// Condition represents a single condition check
type Condition struct {
	Field    string `json:"field" validate:"required"`    // Field to check
	Operator string `json:"operator" validate:"required"` // Comparison operator
	Value    any    `json:"value"`                        // Value to compare against
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

// FieldHTMX defines HTMX behavior for this field
type FieldHTMX struct {
	Trigger   string            `json:"trigger,omitempty"`   // HTMX trigger event
	Post      string            `json:"post,omitempty"`      // POST endpoint
	Get       string            `json:"get,omitempty"`       // GET endpoint
	Target    string            `json:"target,omitempty"`    // Update target
	Swap      string            `json:"swap,omitempty"`      // Swap method
	Indicator string            `json:"indicator,omitempty"` // Loading indicator
	Headers   map[string]string `json:"headers,omitempty"`   // Extra headers
	Validate  bool              `json:"validate,omitempty"`  // Validate before request
}

// FieldAlpine defines Alpine.js bindings
type FieldAlpine struct {
	XModel string `json:"xModel,omitempty" validate:"js_variable"`  // Two-way binding
	XBind  string `json:"xBind,omitempty" validate:"js_object"`     // Attribute bindings
	XOn    string `json:"xOn,omitempty" validate:"js_object"`       // Event handlers
	XShow  string `json:"xShow,omitempty" validate:"js_expression"` // Show/hide
	XIf    string `json:"xIf,omitempty" validate:"js_expression"`   // Conditional render
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

// IsVisible checks if field should be displayed given current form data
func (f *Field) IsVisible(data map[string]any) bool {
	if f.Hidden {
		return false
	}
	if f.Conditional != nil {
		// TODO: Integrate with condition evaluator
		// For now, assume visible
		return true
	}
	return true
}

// IsRequired checks if field is required given current form data
func (f *Field) IsRequired(data map[string]any) bool {
	if f.Required {
		return true
	}
	if f.Conditional != nil && f.Conditional.Required != nil {
		// TODO: Integrate with condition evaluator
		return false
	}
	return false
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
	}

	// Validate options for select-type fields
	if f.requiresOptions() && len(f.Options) == 0 && f.DataSource == nil {
		return NewValidationError(
			"missing_options",
			"field type requires options or dataSource",
		).WithField(f.Name)
	}

	return nil
}

// requiresOptions checks if field type requires options
func (f *Field) requiresOptions() bool {
	switch f.Type {
	case FieldSelect, FieldMultiSelect, FieldRadio, FieldTreeSelect, FieldCascader, FieldTransfer:
		return true
	}
	return false
}

// ValidateValue validates a value against field rules
func (f *Field) ValidateValue(value any) error {
	// Required check
	if f.Required && (value == nil || value == "") {
		return fmt.Errorf("%s is required", f.Label)
	}

	// Skip validation if no value and not required
	if value == nil || value == "" {
		return nil
	}

	if f.Validation == nil {
		return nil
	}

	// String validation
	if strVal, ok := value.(string); ok {
		if f.Validation.MinLength != nil && len(strVal) < *f.Validation.MinLength {
			if f.Validation.Messages.MinLength != "" {
				return fmt.Errorf("%s", f.Validation.Messages.MinLength)
			}
			return fmt.Errorf("%s must be at least %d characters", f.Label, *f.Validation.MinLength)
		}
		if f.Validation.MaxLength != nil && len(strVal) > *f.Validation.MaxLength {
			if f.Validation.Messages.MaxLength != "" {
				return fmt.Errorf("%s", f.Validation.Messages.MaxLength)
			}
			return fmt.Errorf("%s must be at most %d characters", f.Label, *f.Validation.MaxLength)
		}
		// TODO: Pattern validation
	}

	// Number validation
	if numVal, ok := value.(float64); ok {
		if f.Validation.Min != nil {
			if f.Validation.ExclusiveMin && numVal <= *f.Validation.Min {
				return fmt.Errorf("%s must be greater than %v", f.Label, *f.Validation.Min)
			}
			if !f.Validation.ExclusiveMin && numVal < *f.Validation.Min {
				if f.Validation.Messages.Min != "" {
					return fmt.Errorf("%s", f.Validation.Messages.Min)
				}
				return fmt.Errorf("%s must be at least %v", f.Label, *f.Validation.Min)
			}
		}
		if f.Validation.Max != nil {
			if f.Validation.ExclusiveMax && numVal >= *f.Validation.Max {
				return fmt.Errorf("%s must be less than %v", f.Label, *f.Validation.Max)
			}
			if !f.Validation.ExclusiveMax && numVal > *f.Validation.Max {
				if f.Validation.Messages.Max != "" {
					return fmt.Errorf("%s", f.Validation.Messages.Max)
				}
				return fmt.Errorf("%s must be at most %v", f.Label, *f.Validation.Max)
			}
		}
	}

	return nil
}

// GetDefaultValue returns the field's default value
func (f *Field) GetDefaultValue() any {
	if f.Default != nil {
		return f.Default
	}
	// Type-specific defaults
	switch f.Type {
	case FieldCheckbox:
		return false
	case FieldMultiSelect, FieldTags:
		return []string{}
	case FieldNumber, FieldCurrency, FieldSlider, FieldRating:
		return 0
	default:
		return ""
	}
}
