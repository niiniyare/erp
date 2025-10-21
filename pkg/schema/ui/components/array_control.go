// Package components - ArrayControlSchema for complex array/repeater fields
// Based on JSON schema: ArrayControlSchema.json
package components

import (
	"encoding/json"
	"fmt"
)

// ArrayControlSubFormMode represents the layout mode for sub-forms in array controls
type ArrayControlSubFormMode string

const (
	ArraySubFormModeNormal     ArrayControlSubFormMode = "normal"     // Default normal layout
	ArraySubFormModeHorizontal ArrayControlSubFormMode = "horizontal" // Horizontal layout with labels on left
	ArraySubFormModeInline     ArrayControlSubFormMode = "inline"     // Inline layout, compact form
)

// ArrayControlTabsStyle represents the display style for tabs mode
type ArrayControlTabsStyle string

const (
	ArrayTabsStyleDefault ArrayControlTabsStyle = ""      // Default style
	ArrayTabsStyleLine    ArrayControlTabsStyle = "line"  // Line-style tabs
	ArrayTabsStyleCard    ArrayControlTabsStyle = "card"  // Card-style tabs
	ArrayTabsStyleRadio   ArrayControlTabsStyle = "radio" // Radio button style tabs
)

// ArrayControlSize represents the form item size options
type ArrayControlSize string

const (
	ArraySizeXS   ArrayControlSize = "xs"   // Extra small size
	ArraySizeSM   ArrayControlSize = "sm"   // Small size
	ArraySizeMD   ArrayControlSize = "md"   // Medium size (default)
	ArraySizeLG   ArrayControlSize = "lg"   // Large size
	ArraySizeFull ArrayControlSize = "full" // Full width size
)

// ArrayControlMode represents the display mode for the form control
type ArrayControlMode string

const (
	ArrayModeNormal     ArrayControlMode = "normal"     // Normal display mode
	ArrayModeInline     ArrayControlMode = "inline"     // Inline display mode
	ArrayModeHorizontal ArrayControlMode = "horizontal" // Horizontal layout mode
)

// ArrayControlMessages contains validation messages for array controls
type ArrayControlMessages struct {
	// General validation error message
	ValidateFailed string `json:"validateFailed,omitempty"`
	// Error message when minimum length validation fails
	MinLengthValidateFailed string `json:"minLengthValidateFailed,omitempty"`
	// Error message when maximum length validation fails
	MaxLengthValidateFailed string `json:"maxLengthValidateFailed,omitempty"`
}

// AutoFillTrigger represents when autofill should be triggered
type AutoFillTrigger string

const (
	AutoFillTriggerChange AutoFillTrigger = "change" // Trigger on value change
	AutoFillTriggerFocus  AutoFillTrigger = "focus"  // Trigger on focus
	AutoFillTriggerBlur   AutoFillTrigger = "blur"   // Trigger on blur
)

// AutoFillMode represents the popup mode for reference entry
type AutoFillMode string

const (
	AutoFillModePopOver AutoFillMode = "popOver" // Popup overlay mode
	AutoFillModeDialog  AutoFillMode = "dialog"  // Dialog window mode
	AutoFillModeDrawer  AutoFillMode = "drawer"  // Side drawer mode
)

// AutoFillConfig represents advanced autofill configuration for reference entry
type AutoFillConfig struct {
	// Whether it is in reference entry mode, showing candidate values for selection
	ShowSuggestion bool `json:"showSuggestion,omitempty"`
	// Default value selected when in reference entry mode
	DefaultSelection any `json:"defaultSelection,omitempty"`
	// API configuration for autofill data source
	API *APIConfig `json:"api,omitempty"`
	// Whether to display data format error prompts (default: true)
	Silent bool `json:"silent,omitempty"`
	// Data mapping configuration when filling values
	FillMapping map[string]any `json:"fillMapping,omitempty"`
	// Trigger condition for autofill (default: change)
	Trigger AutoFillTrigger `json:"trigger,omitempty"`
	// Popup window mode when in reference entry
	Mode AutoFillMode `json:"mode,omitempty"`
	// Popup position when mode is drawer
	Position string `json:"position,omitempty"`
	// Size of popup container when in reference entry
	Size string `json:"size,omitempty"`
	// Displayed columns for reference entry
	Columns []any `json:"columns,omitempty"`
	// Filter conditions for reference entry
	Filter any `json:"filter,omitempty"`
}

// EditorSetting contains design-time configuration for the page editor
type EditorSetting struct {
	// Component behavior and usage (e.g., create, update, remove)
	Behavior string `json:"behavior,omitempty"`
	// Component display name for easy identification in editor
	DisplayName string `json:"displayName,omitempty"`
	// Mock data for editor preview purposes
	Mock any `json:"mock,omitempty"`
}

// ArrayControlSchema represents a complex array/repeater field control
// This component allows users to add, remove, and manage multiple instances of a form structure
// Based on the AMis InputArray component: https://aisuda.bce.baidu.com/amis/zh-CN/components/form/array
type ArrayControlSchema struct {
	BaseComponentProps

	// Component type identifier (required)
	Type string `json:"type"` // Must be "input-array"

	// Field name for form submission, supports multi-level paths (e.g., "a.b.c") (required)
	Name string `json:"name"`

	// Member renderer configuration - defines the structure of each array item (required)
	Items any `json:"items,omitempty"` // SchemaCollection type

	// Display label for the control, can be string or false to hide
	Label any `json:"label,omitempty"` // string or false

	// Default value when adding new array members
	Scaffold any `json:"scaffold,omitempty"`

	// UI Configuration
	// Whether the control should display without a border
	NoBorder bool `json:"noBorder,omitempty"`
	// CSS class name for the internal form item container
	FormClassName string `json:"formClassName,omitempty"`
	// CSS class name for the add button
	AddButtonClassName string `json:"addButtonClassName,omitempty"`
	// Text displayed on the add button
	AddButtonText string `json:"addButtonText,omitempty"`
	// Placeholder text shown when there are no array members
	Placeholder string `json:"placeholder,omitempty"`

	// Array Management Configuration
	// Whether users can add new items to the array
	Addable bool `json:"addable,omitempty"`
	// Whether new items are added at the top instead of bottom
	AddAtTop bool `json:"addattop,omitempty"`
	// Whether users can remove items from the array
	Removable bool `json:"removable,omitempty"`
	// Confirmation message shown before deleting an item
	DeleteConfirmText string `json:"deleteConfirmText,omitempty"`
	// API configuration for delete operations
	DeleteAPI *APIConfig `json:"deleteApi,omitempty"`

	// Drag and Drop Configuration
	// Whether array items can be reordered by dragging
	Draggable bool `json:"draggable,omitempty"`
	// Tooltip text shown for draggable items
	DraggableTip string `json:"draggableTip,omitempty"`

	// Data Format Configuration
	// Whether to flatten results (remove field name wrapper)
	// Only valid when controls length is 1 and multiple is true
	Flat bool `json:"flat,omitempty"`
	// Delimiter used when flattening is enabled and joinValues is true
	Delimiter string `json:"delimiter,omitempty"`
	// Whether to send flattened data as delimited string vs array
	JoinValues bool `json:"joinValues,omitempty"`

	// Validation Configuration
	// Maximum number of array items allowed
	MaxLength any `json:"maxLength,omitempty"` // number or SchemaTokenizeableString
	// Minimum number of array items required
	MinLength any `json:"minLength,omitempty"` // number or SchemaTokenizeableString
	// Whether multiple selection is possible
	Multiple bool `json:"multiple,omitempty"`
	// Custom validation messages
	Messages *ArrayControlMessages `json:"messages,omitempty"`

	// Layout Configuration
	// Whether to use multi-line display (default: single line)
	MultiLine bool `json:"multiLine,omitempty"`
	// Sub-form layout mode
	SubFormMode ArrayControlSubFormMode `json:"subFormMode,omitempty"`
	// Horizontal layout configuration for sub-forms
	SubFormHorizontal *FormHorizontal `json:"subFormHorizontal,omitempty"`

	// Advanced Features
	// Whether sub-forms can access parent data context
	CanAccessSuperData bool `json:"canAccessSuperData,omitempty"`
	// Whether conditions can be switched (used with conditions property)
	TypeSwitchable bool `json:"typeSwitchable,omitempty"`

	// Tabs Display Mode
	// Whether to display array items as tabs
	TabsMode bool `json:"tabsMode,omitempty"`
	// Visual style of tabs when in tabs mode
	TabsStyle ArrayControlTabsStyle `json:"tabsStyle,omitempty"`
	// Template for generating tab titles
	TabsLabelTpl any `json:"tabsLabelTpl,omitempty"` // SchemaTpl type

	// Performance Configuration
	// Enable lazy loading for large datasets to prevent performance issues
	LazyLoad bool `json:"lazyLoad,omitempty"`
	// Enable strict mode for better data consistency (disabled by default for performance)
	StrictMode bool `json:"strictMode,omitempty"`
	// Fields to synchronize when strictMode is false and combo is deeply nested
	SyncFields []string `json:"syncFields,omitempty"`
	// Allow empty values when validator is configured and in single mode
	Nullable bool `json:"nullable,omitempty"`

	// Form Control Properties
	// Size of the form item
	Size ArrayControlSize `json:"size,omitempty"`
	// Label alignment
	LabelAlign string `json:"labelAlign,omitempty"` // LabelAlign type
	// Custom label width (default unit: px)
	LabelWidth any `json:"labelWidth,omitempty"` // number or string
	// CSS class name for the label
	LabelClassName string `json:"labelClassName,omitempty"`
	// Additional field name for range components
	ExtraName string `json:"extraName,omitempty"`
	// Input prompt shown when focused
	Hint string `json:"hint,omitempty"`
	// Whether to submit form when modifications are completed
	SubmitOnChange bool `json:"submitOnChange,omitempty"`
	// Whether the control is read-only
	ReadOnly bool `json:"readOnly,omitempty"`
	// Expression for conditional read-only state
	ReadOnlyOn string `json:"readOnlyOn,omitempty"`
	// Whether to trigger validation on every change
	ValidateOnChange bool `json:"validateOnChange,omitempty"`
	// Description content supporting HTML fragments
	Description string `json:"description,omitempty"`
	// Short description (alias for description)
	Desc string `json:"desc,omitempty"`
	// CSS class name for description content
	DescriptionClassName string `json:"descriptionClassName,omitempty"`
	// Display mode for the current form item
	Mode ArrayControlMode `json:"mode,omitempty"`
	// Horizontal layout configuration
	Horizontal *FormHorizontal `json:"horizontal,omitempty"`
	// Whether form control is in inline mode
	Inline bool `json:"inline,omitempty"`
	// CSS class name for input element
	InputClassName string `json:"inputClassName,omitempty"`
	// Whether the field is required
	Required bool `json:"required,omitempty"`

	// Validation Configuration
	// Custom error messages for different validation rules
	ValidationErrors map[string]string `json:"validationErrors,omitempty"`
	// Validation rules (string or validation object)
	Validations any `json:"validations,omitempty"`
	// Default static value (variables not supported, use name attribute for data binding)
	Value any `json:"value,omitempty"`
	// Whether to clear field value when form item is hidden
	ClearValueOnHidden bool `json:"clearValueOnHidden,omitempty"`
	// Remote validation API configuration
	ValidateAPI any `json:"validateApi,omitempty"` // string or BaseApiObject
	// Autofill configuration (simple map or advanced config)
	AutoFill any `json:"autoFill,omitempty"` // map[string]string or AutoFillConfig
	// Initial autofill behavior
	InitAutoFill any `json:"initAutoFill,omitempty"` // bool or "fillIfNotSet"
	// Number of rows for display
	Row int `json:"row,omitempty"`

	// Static Display Properties
	// Whether to display in static mode
	Static bool `json:"static,omitempty"`
	// Expression for conditional static display
	StaticOn string `json:"staticOn,omitempty"`
	// Placeholder for empty values in static display
	StaticPlaceholder string `json:"staticPlaceholder,omitempty"`
	// CSS class name for static form item
	StaticClassName string `json:"staticClassName,omitempty"`
	// CSS class name for static label
	StaticLabelClassName string `json:"staticLabelClassName,omitempty"`
	// CSS class name for static input value
	StaticInputClassName string `json:"staticInputClassName,omitempty"`
	// Schema for static display mode
	StaticSchema any `json:"staticSchema,omitempty"`

	// Design and Testing Properties
	// Test ID builder configuration for automated testing
	TestIdBuilder any `json:"testIdBuilder,omitempty"` // TestIdBuilder type
	// Whether to update pristine state after store data reinitialization
	UpdatePristineAfterStoreDataReInit bool `json:"updatePristineAfterStoreDataReInit,omitempty"`
	// Remark tooltip configuration
	Remark any `json:"remark,omitempty"` // SchemaRemark type
	// Label remark tooltip configuration
	LabelRemark any `json:"labelRemark,omitempty"` // SchemaRemark type
	// Design-time editor configuration (ignored at runtime)
	EditorSetting *EditorSetting `json:"editorSetting,omitempty"`
}

// Factory function to create ArrayControlSchema with sensible defaults
func NewArrayControl(name string, items any) *ArrayControlSchema {
	return &ArrayControlSchema{
		Type:          "input-array",
		Name:          name,
		Items:         items,
		AddButtonText: "Add Item",
		Addable:       true,
		Removable:     true,
		SubFormMode:   ArraySubFormModeNormal,
		TabsStyle:     ArrayTabsStyleLine,
		Size:          ArraySizeMD,
		Mode:          ArrayModeNormal,
		MultiLine:     false,
		LazyLoad:      false,
		StrictMode:    false,
		Nullable:      false,
	}
}

// Validation function for ArrayControlSchema
func (a *ArrayControlSchema) Validate() error {
	if a.Type != "input-array" {
		return fmt.Errorf("invalid array control type: %s, must be 'input-array'", a.Type)
	}
	if a.Name == "" {
		return fmt.Errorf("array control name is required")
	}
	if a.Items == nil {
		return fmt.Errorf("array control items configuration is required")
	}

	// Validate enum values
	if a.SubFormMode != "" &&
		a.SubFormMode != ArraySubFormModeNormal &&
		a.SubFormMode != ArraySubFormModeHorizontal &&
		a.SubFormMode != ArraySubFormModeInline {
		return fmt.Errorf("invalid subFormMode: %s", a.SubFormMode)
	}

	if a.TabsStyle != "" &&
		a.TabsStyle != ArrayTabsStyleDefault &&
		a.TabsStyle != ArrayTabsStyleLine &&
		a.TabsStyle != ArrayTabsStyleCard &&
		a.TabsStyle != ArrayTabsStyleRadio {
		return fmt.Errorf("invalid tabsStyle: %s", a.TabsStyle)
	}

	if a.Size != "" &&
		a.Size != ArraySizeXS &&
		a.Size != ArraySizeSM &&
		a.Size != ArraySizeMD &&
		a.Size != ArraySizeLG &&
		a.Size != ArraySizeFull {
		return fmt.Errorf("invalid size: %s", a.Size)
	}

	if a.Mode != "" &&
		a.Mode != ArrayModeNormal &&
		a.Mode != ArrayModeInline &&
		a.Mode != ArrayModeHorizontal {
		return fmt.Errorf("invalid mode: %s", a.Mode)
	}

	return nil
}

// ToJSON converts ArrayControlSchema to JSON string
func (a *ArrayControlSchema) ToJSON() (string, error) {
	if err := a.Validate(); err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal ArrayControlSchema to JSON: %w", err)
	}
	return string(data), nil
}

// FromJSON creates ArrayControlSchema from JSON string
func ArrayControlFromJSON(jsonData string) (*ArrayControlSchema, error) {
	var config ArrayControlSchema
	err := json.Unmarshal([]byte(jsonData), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to ArrayControlSchema: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &config, nil
}
