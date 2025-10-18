// Package components - ComboControlSchema for combination input fields
// Based on JSON schema: ComboControlSchema.json
package components

import (
	"encoding/json"
	"fmt"
)

// ComboControlSize represents the size options for combo controls  
type ComboControlSize string

const (
	ComboControlSizeXS   ComboControlSize = "xs"   // Extra small size
	ComboControlSizeSM   ComboControlSize = "sm"   // Small size
	ComboControlSizeMD   ComboControlSize = "md"   // Medium size (default)
	ComboControlSizeLG   ComboControlSize = "lg"   // Large size
	ComboControlSizeFull ComboControlSize = "full" // Full width size
)

// ComboSubFormMode represents the layout mode for sub-forms in combo controls
type ComboSubFormMode string

const (
	ComboSubFormModeNormal     ComboSubFormMode = "normal"     // Default normal layout
	ComboSubFormModeHorizontal ComboSubFormMode = "horizontal" // Horizontal layout with labels on left
	ComboSubFormModeInline     ComboSubFormMode = "inline"     // Inline layout, compact form
)

// ComboTabsStyle represents the display style for tabs mode
type ComboTabsStyle string

const (
	ComboTabsStyleDefault ComboTabsStyle = ""       // Default style
	ComboTabsStyleLine    ComboTabsStyle = "line"   // Line-style tabs
	ComboTabsStyleCard    ComboTabsStyle = "card"   // Card-style tabs
	ComboTabsStyleRadio   ComboTabsStyle = "radio"  // Radio button style tabs
)

// ComboMode represents the display mode for the combo control
type ComboMode string

const (
	ComboModeNormal     ComboMode = "normal"     // Normal display mode
	ComboModeInline     ComboMode = "inline"     // Inline display mode
	ComboModeHorizontal ComboMode = "horizontal" // Horizontal layout mode
)

// ComboControlMessages contains validation messages for combo controls
type ComboControlMessages struct {
	// General validation error message
	ValidateFailed string `json:"validateFailed,omitempty"`
	// Error message when minimum length validation fails
	MinLengthValidateFailed string `json:"minLengthValidateFailed,omitempty"`
	// Error message when maximum length validation fails
	MaxLengthValidateFailed string `json:"maxLengthValidateFailed,omitempty"`
}

// ComboCondition represents condition-based combo configuration
// This allows different form schemas to be rendered based on specific conditions
type ComboCondition struct {
	// JavaScript expression to test for this condition
	Test string `json:"test,omitempty"`
	// Label displayed for this condition type
	Label string `json:"label,omitempty"`
	// Form controls to render when this condition is met
	Items []any `json:"items,omitempty"`
	// Default values for this condition
	Scaffold any `json:"scaffold,omitempty"`
	// Type label for this condition (used in type switching)
	TypeLabel string `json:"typeLabel,omitempty"`
}

// ComboSubControl represents a control within the combo form
type ComboSubControl struct {
	// Control type (input-text, select, etc.)
	Type string `json:"type,omitempty"`
	// Field name
	Name string `json:"name,omitempty"`
	// Display label
	Label string `json:"label,omitempty"`
	// Control-specific configuration
	Config map[string]any `json:"config,omitempty"`
}

// ComboControlSchema represents a combination input field control
// This component allows for complex form structures with multiple sub-controls,
// conditional rendering, and dynamic adding/removing of form groups
// Based on AMis Combo control: https://aisuda.bce.baidu.com/amis/zh-CN/components/form/combo
type ComboControlSchema struct {
	BaseComponentProps

	// Component type identifier (required)
	Type string `json:"type"` // Must be "combo"

	// Field name for form submission, supports multi-level paths (e.g., "a.b.c") (required)
	Name string `json:"name"`

	// Display label for the combo control
	Label string `json:"label,omitempty"`

	// Default value for the combo control
	Value any `json:"value,omitempty"`

	// Placeholder text shown in empty state
	Placeholder string `json:"placeholder,omitempty"`

	// Size of the combo control
	Size ComboControlSize `json:"size,omitempty"`

	// Whether the field is required
	Required bool `json:"required,omitempty"`

	// Whether the control is read-only
	ReadOnly bool `json:"readOnly,omitempty"`

	// Whether multiple combo groups can be created
	Multiple bool `json:"multiple,omitempty"`

	// Form Structure Configuration
	// Array of form controls within each combo group (required for basic combo)
	Items []ComboSubControl `json:"items,omitempty"`

	// Conditional rendering configuration
	Conditions []ComboCondition `json:"conditions,omitempty"`

	// Whether users can switch between different condition types
	TypeSwitchable bool `json:"typeSwitchable,omitempty"`

	// Default values when adding new combo groups
	Scaffold any `json:"scaffold,omitempty"`

	// Data Processing Configuration
	// Whether to flatten results (remove field name wrapper)
	Flat bool `json:"flat,omitempty"`
	// Whether to join multiple values with delimiter
	JoinValues bool `json:"joinValues,omitempty"`
	// Whether to extract value from response
	ExtractValue bool `json:"extractValue,omitempty"`
	// Delimiter for joining values when multiple combo groups
	Delimiter string `json:"delimiter,omitempty"`

	// UI Configuration
	// CSS class name for the internal form container
	FormClassName string `json:"formClassName,omitempty"`
	// CSS class name for the add button
	AddButtonClassName string `json:"addButtonClassName,omitempty"`
	// Text displayed on the add button
	AddButtonText string `json:"addButtonText,omitempty"`
	// Whether the control should display without a border
	NoBorder bool `json:"noBorder,omitempty"`

	// Combo Management Configuration
	// Whether users can add new combo groups
	Addable bool `json:"addable,omitempty"`
	// Whether new combo groups are added at the top instead of bottom
	AddAtTop bool `json:"addattop,omitempty"`
	// Whether users can remove combo groups
	Removable bool `json:"removable,omitempty"`
	// Confirmation message shown before deleting a combo group
	DeleteConfirmText string `json:"deleteConfirmText,omitempty"`

	// Drag and Drop Configuration
	// Whether combo groups can be reordered by dragging
	Draggable bool `json:"draggable,omitempty"`
	// Tooltip text shown for draggable combo groups
	DraggableTip string `json:"draggableTip,omitempty"`

	// Layout Configuration
	// Sub-form layout mode
	SubFormMode ComboSubFormMode `json:"subFormMode,omitempty"`
	// Horizontal layout configuration for sub-forms
	SubFormHorizontal *FormHorizontal `json:"subFormHorizontal,omitempty"`

	// Tabs Display Mode (for multiple combo groups)
	// Whether to display combo groups as tabs
	TabsMode bool `json:"tabsMode,omitempty"`
	// Visual style of tabs when in tabs mode
	TabsStyle ComboTabsStyle `json:"tabsStyle,omitempty"`
	// Template for generating tab titles
	TabsLabelTpl any `json:"tabsLabelTpl,omitempty"` // SchemaTpl type

	// Validation Configuration
	// Maximum number of combo groups allowed
	MaxLength any `json:"maxLength,omitempty"` // number or SchemaTokenizeableString
	// Minimum number of combo groups required
	MinLength any `json:"minLength,omitempty"` // number or SchemaTokenizeableString
	// Custom validation messages
	Messages *ComboControlMessages `json:"messages,omitempty"`

	// Advanced Features
	// Whether sub-forms can access parent data context
	CanAccessSuperData bool `json:"canAccessSuperData,omitempty"`
	// Allow empty values when validator is configured
	Nullable bool `json:"nullable,omitempty"`

	// Performance Configuration
	// Enable lazy loading for large datasets
	LazyLoad bool `json:"lazyLoad,omitempty"`
	// Enable strict mode for better data consistency (disabled by default for performance)
	StrictMode bool `json:"strictMode,omitempty"`
	// Fields to synchronize when strictMode is false and combo is deeply nested
	SyncFields []string `json:"syncFields,omitempty"`

	// API Configuration
	// API configuration for delete operations
	DeleteAPI *APIConfig `json:"deleteApi,omitempty"`

	// Form Control Properties
	// Display mode for the current form item
	Mode ComboMode `json:"mode,omitempty"`
	// Label alignment
	LabelAlign string `json:"labelAlign,omitempty"`
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
	// Expression for conditional read-only state
	ReadOnlyOn string `json:"readOnlyOn,omitempty"`
	// Whether to trigger validation on every change
	ValidateOnChange bool `json:"validateOnChange,omitempty"`
	// Description content supporting HTML fragments
	Description string `json:"description,omitempty"`
	// CSS class name for description content
	DescriptionClassName string `json:"descriptionClassName,omitempty"`
	// Horizontal layout configuration
	Horizontal *FormHorizontal `json:"horizontal,omitempty"`
	// Whether form control is in inline mode
	Inline bool `json:"inline,omitempty"`
	// CSS class name for input element
	InputClassName string `json:"inputClassName,omitempty"`

	// Validation Configuration
	// Custom error messages for different validation rules
	ValidationErrors map[string]string `json:"validationErrors,omitempty"`
	// Validation rules (string or validation object)
	Validations any `json:"validations,omitempty"`
	// Whether to clear field value when form item is hidden
	ClearValueOnHidden bool `json:"clearValueOnHidden,omitempty"`
	// Remote validation API configuration
	ValidateAPI any `json:"validateApi,omitempty"`
	// Autofill configuration (simple map or advanced config)
	AutoFill any `json:"autoFill,omitempty"` // map[string]string or AutoFillConfig
	// Initial autofill behavior
	InitAutoFill any `json:"initAutoFill,omitempty"` // bool or "fillIfNotSet"
	// Number of rows for display
	Row int `json:"row,omitempty"`

	// Static Display Properties
	Static bool `json:"static,omitempty"`
	StaticOn string `json:"staticOn,omitempty"`
	StaticPlaceholder string `json:"staticPlaceholder,omitempty"`
	StaticClassName string `json:"staticClassName,omitempty"`
	StaticLabelClassName string `json:"staticLabelClassName,omitempty"`
	StaticInputClassName string `json:"staticInputClassName,omitempty"`
	StaticSchema any `json:"staticSchema,omitempty"`

	// Design and Testing Properties
	TestIdBuilder any `json:"testIdBuilder,omitempty"`
	UpdatePristineAfterStoreDataReInit bool `json:"updatePristineAfterStoreDataReInit,omitempty"`
	Remark any `json:"remark,omitempty"`
	LabelRemark any `json:"labelRemark,omitempty"`
	EditorSetting *EditorSetting `json:"editorSetting,omitempty"`
}

// Factory function to create a basic combo control
func NewComboControl(name string, items []ComboSubControl) *ComboControlSchema {
	return &ComboControlSchema{
		Type:           "combo",
		Name:           name,
		Items:          items,
		Size:           ComboControlSizeMD,
		SubFormMode:    ComboSubFormModeNormal,
		TabsStyle:      ComboTabsStyleLine,
		Mode:           ComboModeNormal,
		AddButtonText:  "Add",
		Addable:        true,
		Removable:      true,
		Flat:           true,
		Multiple:       false,
		LazyLoad:       false,
		StrictMode:     false,
		Nullable:       false,
	}
}

// Factory function to create a multiple combo control (allows multiple groups)
func NewMultipleComboControl(name string, items []ComboSubControl) *ComboControlSchema {
	combo := NewComboControl(name, items)
	combo.Multiple = true
	combo.TabsMode = false // Use standard layout for multiple groups
	return combo
}

// Factory function to create a conditional combo control
func NewConditionalComboControl(name string, conditions []ComboCondition) *ComboControlSchema {
	return &ComboControlSchema{
		Type:           "combo",
		Name:           name,
		Conditions:     conditions,
		TypeSwitchable: true,
		Size:           ComboControlSizeMD,
		SubFormMode:    ComboSubFormModeNormal,
		Mode:           ComboModeNormal,
		AddButtonText:  "Add",
		Addable:        true,
		Removable:      true,
		Flat:           true,
		Multiple:       false,
	}
}

// Factory function to create a tabbed combo control
func NewTabbedComboControl(name string, items []ComboSubControl) *ComboControlSchema {
	combo := NewComboControl(name, items)
	combo.Multiple = true
	combo.TabsMode = true
	combo.TabsStyle = ComboTabsStyleLine
	return combo
}

// Helper method to add a condition to the combo control
func (c *ComboControlSchema) AddCondition(condition ComboCondition) {
	c.Conditions = append(c.Conditions, condition)
}

// Helper method to add a sub-control to the combo
func (c *ComboControlSchema) AddSubControl(control ComboSubControl) {
	c.Items = append(c.Items, control)
}

// Helper method to create a simple condition
func NewComboCondition(test, label string, items []any) ComboCondition {
	return ComboCondition{
		Test:  test,
		Label: label,
		Items: items,
	}
}

// Helper method to create a simple sub-control
func NewComboSubControl(controlType, name, label string) ComboSubControl {
	return ComboSubControl{
		Type:  controlType,
		Name:  name,
		Label: label,
	}
}

// Validation function for ComboControlSchema
func (c *ComboControlSchema) Validate() error {
	if c.Type != "combo" {
		return fmt.Errorf("invalid combo control type: %s, must be 'combo'", c.Type)
	}
	if c.Name == "" {
		return fmt.Errorf("combo control name is required")
	}

	// Must have either Items or Conditions
	if len(c.Items) == 0 && len(c.Conditions) == 0 {
		return fmt.Errorf("combo control must have either items or conditions configured")
	}

	// Validate enum values
	if c.Size != "" &&
		c.Size != ComboControlSizeXS &&
		c.Size != ComboControlSizeSM &&
		c.Size != ComboControlSizeMD &&
		c.Size != ComboControlSizeLG &&
		c.Size != ComboControlSizeFull {
		return fmt.Errorf("invalid size: %s", c.Size)
	}

	if c.SubFormMode != "" &&
		c.SubFormMode != ComboSubFormModeNormal &&
		c.SubFormMode != ComboSubFormModeHorizontal &&
		c.SubFormMode != ComboSubFormModeInline {
		return fmt.Errorf("invalid subFormMode: %s", c.SubFormMode)
	}

	if c.TabsStyle != "" &&
		c.TabsStyle != ComboTabsStyleDefault &&
		c.TabsStyle != ComboTabsStyleLine &&
		c.TabsStyle != ComboTabsStyleCard &&
		c.TabsStyle != ComboTabsStyleRadio {
		return fmt.Errorf("invalid tabsStyle: %s", c.TabsStyle)
	}

	if c.Mode != "" &&
		c.Mode != ComboModeNormal &&
		c.Mode != ComboModeInline &&
		c.Mode != ComboModeHorizontal {
		return fmt.Errorf("invalid mode: %s", c.Mode)
	}

	// Validate conditions
	for i, condition := range c.Conditions {
		if condition.Test == "" {
			return fmt.Errorf("condition %d must have a test expression", i)
		}
		if len(condition.Items) == 0 {
			return fmt.Errorf("condition %d (%s) must have items", i, condition.Label)
		}
	}

	// Validate sub-controls
	for i, item := range c.Items {
		if item.Type == "" {
			return fmt.Errorf("sub-control %d must have a type", i)
		}
		if item.Name == "" {
			return fmt.Errorf("sub-control %d (%s) must have a name", i, item.Type)
		}
	}

	return nil
}

// ToJSON converts ComboControlSchema to JSON string
func (c *ComboControlSchema) ToJSON() (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal ComboControlSchema to JSON: %w", err)
	}
	return string(data), nil
}

// FromJSON creates ComboControlSchema from JSON string
func ComboControlFromJSON(jsonData string) (*ComboControlSchema, error) {
	var config ComboControlSchema
	err := json.Unmarshal([]byte(jsonData), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to ComboControlSchema: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &config, nil
}