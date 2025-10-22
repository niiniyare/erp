package filterpanel

import (
	"github.com/niiniyare/erp/web/components/atoms"
	"github.com/niiniyare/erp/web/components/molecules"
)

// FilterType defines the type of filter input
type FilterType string

const (
	FilterTypeText      FilterType = "text"      // Text input
	FilterTypeSelect    FilterType = "select"    // Dropdown select
	FilterTypeDate      FilterType = "date"      // Date picker
	FilterTypeDateRange FilterType = "daterange" // Date range picker
	FilterTypeNumber    FilterType = "number"    // Number input
	FilterTypeCheckbox  FilterType = "checkbox"  // Checkbox group
	FilterTypeRadio     FilterType = "radio"     // Radio button group
	FilterTypeSearch    FilterType = "search"    // Search input with suggestions
)

// FilterOperator defines comparison operators for filters
type FilterOperator string

const (
	OperatorEquals       FilterOperator = "eq"        // Equals
	OperatorNotEquals    FilterOperator = "ne"        // Not equals
	OperatorContains     FilterOperator = "contains"  // Contains text
	OperatorStartsWith   FilterOperator = "starts"    // Starts with
	OperatorEndsWith     FilterOperator = "ends"      // Ends with
	OperatorGreaterThan  FilterOperator = "gt"        // Greater than
	OperatorLessThan     FilterOperator = "lt"        // Less than
	OperatorGreaterEqual FilterOperator = "gte"       // Greater than or equal
	OperatorLessEqual    FilterOperator = "lte"       // Less than or equal
	OperatorBetween      FilterOperator = "between"   // Between (range)
	OperatorIn           FilterOperator = "in"        // In list
	OperatorNotIn        FilterOperator = "notin"     // Not in list
	OperatorIsNull       FilterOperator = "isnull"    // Is null/empty
	OperatorIsNotNull    FilterOperator = "isnotnull" // Is not null/empty
)

// FilterOption represents an option for select/checkbox/radio filters
type FilterOption struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	Selected bool   `json:"selected"`
	Disabled bool   `json:"disabled"`
	Group    string `json:"group,omitempty"` // Option group
	Count    int    `json:"count,omitempty"` // Result count for this option
}

// FilterField defines a single filter field
type FilterField struct {
	// Basic properties
	Name        string     `json:"name"`  // Field name for form submission
	Label       string     `json:"label"` // Display label
	Type        FilterType `json:"type"`  // Input type
	Placeholder string     `json:"placeholder,omitempty"`
	Required    bool       `json:"required"`

	// Value and state
	Value        any            `json:"value,omitempty"`        // Current value
	DefaultValue any            `json:"defaultValue,omitempty"` // Default value
	Operator     FilterOperator `json:"operator,omitempty"`     // Comparison operator

	// Options for select/radio/checkbox types
	Options  []FilterOption `json:"options,omitempty"`
	Multiple bool           `json:"multiple"` // Allow multiple selections

	// Validation
	Min     any    `json:"min,omitempty"`     // Min value for numbers/dates
	Max     any    `json:"max,omitempty"`     // Max value for numbers/dates
	Pattern string `json:"pattern,omitempty"` // Regex pattern

	// UI behavior
	Collapsible bool   `json:"collapsible"`     // Can be collapsed
	Collapsed   bool   `json:"collapsed"`       // Initially collapsed
	Width       string `json:"width,omitempty"` // CSS width class

	// Help and description
	Help        string `json:"help,omitempty"`        // Help text
	Description string `json:"description,omitempty"` // Field description

	// HTMX integration
	HxGet     string `json:"hxGet,omitempty"`     // Load options dynamically
	HxPost    string `json:"hxPost,omitempty"`    // Submit filter
	HxTarget  string `json:"hxTarget,omitempty"`  // Target for results
	HxTrigger string `json:"hxTrigger,omitempty"` // Trigger events
	HxSwap    string `json:"hxSwap,omitempty"`    // Swap strategy

	// Dependencies
	DependsOn []string `json:"dependsOn,omitempty"` // Fields this depends on
	ShowWhen  string   `json:"showWhen,omitempty"`  // Condition to show field

	// Styling
	ID    string `json:"id,omitempty"`
	Class string `json:"class,omitempty"`
}

// FilterGroup defines a group of related filter fields
type FilterGroup struct {
	Title       string        `json:"title"`
	Fields      []FilterField `json:"fields"`
	Collapsible bool          `json:"collapsible"`
	Collapsed   bool          `json:"collapsed"`
	Icon        string        `json:"icon,omitempty"`
	Description string        `json:"description,omitempty"`

	// Layout
	Columns int    `json:"columns,omitempty"` // Number of columns for fields
	Spacing string `json:"spacing,omitempty"` // Custom spacing classes

	// Visibility
	ShowWhen string `json:"showWhen,omitempty"` // Condition to show group

	ID    string `json:"id,omitempty"`
	Class string `json:"class,omitempty"`
}

// FilterPreset defines a saved filter configuration
type FilterPreset struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Public      bool           `json:"public"`  // Shared with other users
	Default     bool           `json:"default"` // Default preset
	Values      map[string]any `json:"values"`  // Field values
	CreatedBy   string         `json:"createdBy,omitempty"`
	CreatedAt   string         `json:"createdAt,omitempty"`
}

// FilterAction defines action buttons for the filter panel
type FilterAction struct {
	Text    string        `json:"text"`
	Icon    string        `json:"icon,omitempty"`
	Variant atoms.Variant `json:"variant"`
	Size    atoms.Size    `json:"size"`
	Type    string        `json:"type,omitempty"` // submit, button, reset
	OnClick string        `json:"onclick,omitempty"`

	// HTMX attributes
	HxPost   string `json:"hxPost,omitempty"`
	HxGet    string `json:"hxGet,omitempty"`
	HxTarget string `json:"hxTarget,omitempty"`
	HxSwap   string `json:"hxSwap,omitempty"`

	// Behavior
	Position  string `json:"position,omitempty"` // left, right
	AutoFocus bool   `json:"autoFocus"`
	Disabled  bool   `json:"disabled"`

	ID    string `json:"id,omitempty"`
	Class string `json:"class,omitempty"`
}

// FilterPanelProps defines properties for the FilterPanel organism
type FilterPanelProps struct {
	// Core configuration
	ID    string `json:"id,omitempty"`
	Title string `json:"title"`

	// Filter structure
	Groups []FilterGroup `json:"groups"`           // Filter groups
	Fields []FilterField `json:"fields,omitempty"` // Standalone fields (ungrouped)

	// Form configuration
	FormAction string `json:"formAction,omitempty"` // Form submission URL
	FormMethod string `json:"formMethod,omitempty"` // GET, POST

	// HTMX integration
	HxPost    string `json:"hxPost,omitempty"`    // Submit filters via HTMX
	HxGet     string `json:"hxGet,omitempty"`     // Load filter options
	HxTarget  string `json:"hxTarget,omitempty"`  // Target for results
	HxSwap    string `json:"hxSwap,omitempty"`    // Swap strategy
	HxTrigger string `json:"hxTrigger,omitempty"` // Trigger events

	// Actions
	Actions   []FilterAction `json:"actions,omitempty"` // Custom action buttons
	ShowApply bool           `json:"showApply"`         // Show Apply button
	ShowReset bool           `json:"showReset"`         // Show Reset button
	ShowSave  bool           `json:"showSave"`          // Show Save Preset button

	// Presets
	Presets     []FilterPreset `json:"presets,omitempty"` // Available presets
	ShowPresets bool           `json:"showPresets"`       // Show preset dropdown

	// Layout
	Collapsible bool   `json:"collapsible"`        // Panel can be collapsed
	Collapsed   bool   `json:"collapsed"`          // Initially collapsed
	Sticky      bool   `json:"sticky"`             // Sticky positioning
	Position    string `json:"position,omitempty"` // top, left, right
	Width       string `json:"width,omitempty"`    // Panel width

	// Behavior
	AutoApply bool `json:"autoApply"` // Apply filters on change
	SaveState bool `json:"saveState"` // Save state to localStorage
	ClearAll  bool `json:"clearAll"`  // Show clear all button

	// Count and results
	ShowCount bool   `json:"showCount"`           // Show result count
	CountText string `json:"countText,omitempty"` // Custom count text

	// Search integration
	SearchProps *molecules.SearchProps `json:"searchProps,omitempty"` // Quick search

	// Styling
	Variant    string `json:"variant,omitempty"` // default, minimal, card
	Class      string `json:"class,omitempty"`
	DataTestID string `json:"dataTestId,omitempty"`

	// Accessibility
	AriaLabel string `json:"ariaLabel,omitempty"`
}

// QuickFilterProps defines a simplified filter interface
type QuickFilterProps struct {
	Fields     []FilterField `json:"fields"`
	HxTarget   string        `json:"hxTarget"`
	HxPost     string        `json:"hxPost,omitempty"`
	Horizontal bool          `json:"horizontal"` // Horizontal layout
	Compact    bool          `json:"compact"`    // Compact spacing
}
