// Package atoms provides atomic UI components with comprehensive type safety,
// validation, and styling capabilities following Go best practices.
//
// Design Principles:
// - Single source of truth for all component types
// - Type safety through const enums
// - Composition over inheritance
// - Idiomatic Go patterns (builder, options)
// - Thread-safe operations
// - Zero external dependencies for core types
package atoms

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
)

// Version of the atoms package
const Version = "2.0.0"

// ============================================================================
// SIZE ENUMERATION
// ============================================================================

// Size defines component sizing options used across all visual components.
// Use these constants instead of string literals for type safety.
type Size string

const (
	SizeXS Size = "xs" // Extra small - 12px base
	SizeSM Size = "sm" // Small - 14px base
	SizeMD Size = "md" // Medium - 16px base (default)
	SizeLG Size = "lg" // Large - 18px base
	SizeXL Size = "xl" // Extra large - 20px base
)

// InputSize defines input-specific sizing options.
type InputSize = Size

const (
	InputSizeXS = SizeXS
	InputSizeSM = SizeSM
	InputSizeMD = SizeMD
	InputSizeLG = SizeLG
	InputSizeXL = SizeXL
)

// IconSize defines icon-specific sizing options.
type IconSize = Size

const (
	IconSizeXS   = SizeXS
	IconSizeSM   = SizeSM
	IconSizeMD   = SizeMD
	IconSizeLG   = SizeLG
	IconSizeXL   = SizeXL
	IconSize2XL  Size = "2xl" // Extra large for special cases
)

// SpinnerSize defines spinner-specific sizing options.
type SpinnerSize = Size

const (
	SpinnerSizeXS = SizeXS
	SpinnerSizeSM = SizeSM
	SpinnerSizeMD = SizeMD
	SpinnerSizeLG = SizeLG
	SpinnerSizeXL = SizeXL
)

// Button-specific size constants for backward compatibility
const (
	ButtonSizeXS = SizeXS
	ButtonSizeSM = SizeSM
	ButtonSizeMD = SizeMD
	ButtonSizeLG = SizeLG
	ButtonSizeXL = SizeXL
)

// Additional size types for specific components
type ButtonSize = Size
type TextareaSize = Size
type SelectSize = Size
type CheckboxSize = Size
type RadioSize = Size

// Component variant aliases
type ButtonVariant = Variant

// IconButton is an alias for ButtonProps with icon-focused configuration
type IconButton = ButtonProps

// String returns the string representation of Size.
func (s Size) String() string {
	return string(s)
}

// IsValid checks if the size value is valid.
func (s Size) IsValid() bool {
	switch s {
	case SizeXS, SizeSM, SizeMD, SizeLG, SizeXL:
		return true
	}
	return false
}

// AllSizes returns all valid size options.
func AllSizes() []Size {
	return []Size{SizeXS, SizeSM, SizeMD, SizeLG, SizeXL}
}

// ============================================================================
// VALIDATION STATE ENUMERATION
// ============================================================================

// ValidationState defines the validation status of a form field.
type ValidationState string

const (
	StateDefault ValidationState = "default" // Neutral state, no validation
	StateSuccess ValidationState = "success" // Valid input, positive feedback
	StateError   ValidationState = "error"   // Invalid input, error feedback
	StateWarning ValidationState = "warning" // Valid but potentially problematic
	StateInfo    ValidationState = "info"    // Informational feedback
)

// String returns the string representation of ValidationState.
func (v ValidationState) String() string {
	return string(v)
}

// IsValid checks if the validation state is valid.
func (v ValidationState) IsValid() bool {
	switch v {
	case StateDefault, StateSuccess, StateError, StateWarning, StateInfo:
		return true
	}
	return false
}

// IsError returns true if the state indicates an error.
func (v ValidationState) IsError() bool {
	return v == StateError
}

// IsSuccess returns true if the state indicates success.
func (v ValidationState) IsSuccess() bool {
	return v == StateSuccess
}

// AllValidationStates returns all valid validation states.
func AllValidationStates() []ValidationState {
	return []ValidationState{StateDefault, StateSuccess, StateError, StateWarning, StateInfo}
}

// ============================================================================
// LABEL POSITION ENUMERATION
// ============================================================================

// LabelPosition defines where the label appears relative to the input.
type LabelPosition string

const (
	LabelTop    LabelPosition = "top"    // Above input (default for text inputs)
	LabelBottom LabelPosition = "bottom" // Below input (rare)
	LabelLeft   LabelPosition = "left"   // Left of input (for inline forms)
	LabelRight  LabelPosition = "right"  // Right of input (default for checkboxes)
	LabelNone   LabelPosition = "none"   // No visible label (aria-label only)
)

// String returns the string representation of LabelPosition.
func (l LabelPosition) String() string {
	return string(l)
}

// ============================================================================
// INPUT TYPE ENUMERATION
// ============================================================================

// InputType defines HTML input types for form elements.
type InputType string

const (
	InputTypeText     InputType = "text"
	InputTypeEmail    InputType = "email"
	InputTypePassword InputType = "password"
	InputTypeNumber   InputType = "number"
	InputTypeTel      InputType = "tel"
	InputTypeURL      InputType = "url"
	InputTypeSearch   InputType = "search"
	InputTypeDate     InputType = "date"
	InputTypeTime     InputType = "time"
	InputTypeDatetime InputType = "datetime-local"
	InputTypeMonth    InputType = "month"
	InputTypeWeek     InputType = "week"
	InputTypeColor    InputType = "color"
	InputTypeFile     InputType = "file"
	InputTypeHidden   InputType = "hidden"
	InputTypeRange    InputType = "range"
	InputTypeCheckbox InputType = "checkbox"
	InputTypeRadio    InputType = "radio"
)

// Input type aliases for backward compatibility
const (
	InputText     = InputTypeText
	InputPassword = InputTypePassword
	InputEmail    = InputTypeEmail
	InputNumber   = InputTypeNumber
	InputTel      = InputTypeTel
	InputURL      = InputTypeURL
)

// String returns the string representation of InputType.
func (i InputType) String() string {
	return string(i)
}

// GetAriaRole returns the appropriate ARIA role for the input type.
func (i InputType) GetAriaRole() string {
	roles := map[InputType]string{
		InputTypeSearch: "searchbox",
	}
	if role, ok := roles[i]; ok {
		return role
	}
	return "" // Use native role
}

// ============================================================================
// VARIANT ENUMERATION
// ============================================================================

// Variant defines visual style variants for components.
type Variant string

const (
	VariantDefault    Variant = "default"    // Standard styling
	VariantOutlined   Variant = "outlined"   // Outlined/bordered
	VariantFilled     Variant = "filled"     // Filled background
	VariantGhost      Variant = "ghost"      // Minimal styling
	VariantUnderlined Variant = "underlined" // Only bottom border
	VariantSolid      Variant = "solid"      // Solid background
	VariantSecondary  Variant = "secondary"  // Secondary styling
	VariantPrimary    Variant = "primary"    // Primary styling
	VariantLight      Variant = "light"      // Light styling
	VariantWarning    Variant = "warning"    // Warning styling
	VariantDark       Variant = "dark"       // Dark styling
	VariantGradient   Variant = "gradient"   // Gradient styling
)

// Button-specific variants for backward compatibility
const (
	ButtonPrimary   = VariantPrimary
	ButtonSecondary = VariantSecondary
	ButtonDanger    = ColorDanger
	ButtonWarning   = VariantWarning
	ButtonLight     = VariantLight
	ButtonGhost     = VariantGhost
)

// String returns the string representation of Variant.
func (v Variant) String() string {
	return string(v)
}

// ============================================================================
// COLOR SCHEME ENUMERATION
// ============================================================================

// ColorScheme defines color themes for components.
type ColorScheme string

const (
	ColorDefault     ColorScheme = "default"     // Blue/primary colors
	ColorPrimary     ColorScheme = "primary"     // Primary brand color
	ColorSecondary   ColorScheme = "secondary"   // Secondary brand color
	ColorSuccess     ColorScheme = "success"     // Green - success states
	ColorDanger      ColorScheme = "danger"      // Red - error/danger states
	ColorWarning     ColorScheme = "warning"     // Yellow/orange - warning states
	ColorInfo        ColorScheme = "info"        // Blue - informational states
	ColorGray        ColorScheme = "gray"        // Neutral gray
	ColorNeutral     ColorScheme = "neutral"     // Alias for gray
	ColorLight       ColorScheme = "light"       // Light colors
	ColorDark        ColorScheme = "dark"        // Dark colors
	ColorPurple      ColorScheme = "purple"      // Purple colors
	ColorAlternative ColorScheme = "alternative" // Alternative color scheme
)

// String returns the string representation of ColorScheme.
func (c ColorScheme) String() string {
	return string(c)
}

// ============================================================================
// ICON POSITION ENUMERATION
// ============================================================================

// IconPosition defines where an icon appears relative to content.
type IconPosition string

const (
	IconLeft   IconPosition = "left"
	IconRight  IconPosition = "right"
	IconTop    IconPosition = "top"
	IconBottom IconPosition = "bottom"
)

// String returns the string representation of IconPosition.
func (i IconPosition) String() string {
	return string(i)
}

// ============================================================================
// BREAKPOINT ENUMERATION
// ============================================================================

// Breakpoint represents responsive design breakpoints.
type Breakpoint string

const (
	BreakpointSM  Breakpoint = "sm"  // 640px
	BreakpointMD  Breakpoint = "md"  // 768px
	BreakpointLG  Breakpoint = "lg"  // 1024px
	BreakpointXL  Breakpoint = "xl"  // 1280px
	BreakpointXXL Breakpoint = "2xl" // 1536px
)

// String returns the string representation of Breakpoint.
func (b Breakpoint) String() string {
	return string(b)
}

// ============================================================================
// ID GENERATOR (Thread-Safe)
// ============================================================================

var idCounter int64

// GenerateID generates a unique ID with the given prefix in a thread-safe manner.
// Uses atomic operations to ensure uniqueness in concurrent environments.
func GenerateID(prefix string) string {
	id := atomic.AddInt64(&idCounter, 1)
	return fmt.Sprintf("%s-%d", prefix, id)
}

// EnsureID returns the provided ID or generates one if empty.
func EnsureID(id, prefix string) string {
	if id != "" {
		return id
	}
	return GenerateID(prefix)
}

// ResetIDCounter resets the ID counter (useful for testing).
// WARNING: Not thread-safe, should only be used in tests.
func ResetIDCounter() {
	atomic.StoreInt64(&idCounter, 0)
}

// ============================================================================
// COMPONENT TYPE REGISTRY & SCHEMA ALIGNMENT
// ============================================================================

// ComponentType represents the type of atomic component.
// Aligned with schema definitions in docs/ui/schema/definitions/components/atoms/
type ComponentType string

const (
	ComponentButton      ComponentType = "button"
	ComponentInput       ComponentType = "input"
	ComponentCheckbox    ComponentType = "checkbox"
	ComponentRadio       ComponentType = "radio"
	ComponentSelect      ComponentType = "select"
	ComponentTextarea    ComponentType = "textarea"
	ComponentIcon        ComponentType = "icon"
	ComponentSpinner     ComponentType = "spinner"
	ComponentToggle      ComponentType = "toggle"
	ComponentLabel       ComponentType = "label"
	ComponentBadge       ComponentType = "badge"
	ComponentAvatar      ComponentType = "avatar"
	ComponentAction      ComponentType = "action"
	ComponentDivider     ComponentType = "divider"
	ComponentHidden      ComponentType = "hidden"
	ComponentImage       ComponentType = "image"
	ComponentLink        ComponentType = "link"
	ComponentProgress    ComponentType = "progress"
	ComponentStatic      ComponentType = "static"
	ComponentStatus      ComponentType = "status"
	ComponentTag         ComponentType = "tag"
	ComponentUUID        ComponentType = "uuid"
	ComponentColorInput  ComponentType = "color-input"
)

// ============================================================================
// SCHEMA EXPRESSION SUPPORT
// ============================================================================

// SchemaExpression represents conditional expressions used in schemas
// Supports both static boolean values and dynamic expressions
type SchemaExpression struct {
	Static    *bool   `json:"static,omitempty"`    // Direct boolean value
	Expression string `json:"expression,omitempty"` // Dynamic expression string
}

// NewStaticExpression creates a static boolean expression
func NewStaticExpression(value bool) *SchemaExpression {
	return &SchemaExpression{Static: &value}
}

// NewDynamicExpression creates a dynamic expression
func NewDynamicExpression(expr string) *SchemaExpression {
	return &SchemaExpression{Expression: expr}
}

// IsTrue evaluates the expression to determine if it's true
func (se *SchemaExpression) IsTrue() bool {
	if se == nil {
		return false
	}
	if se.Static != nil {
		return *se.Static
	}
	// For dynamic expressions, we'd need context evaluation
	// For now, return false as safe default
	return false
}

// ============================================================================
// EDITOR CONFIGURATION
// ============================================================================

// EditorSetting contains metadata for visual editor integration
// Maps to schema editorSetting property
type EditorSetting struct {
	Behavior    string      `json:"behavior,omitempty"`    // create, update, remove
	DisplayName string      `json:"displayName,omitempty"` // Business-friendly name
	Mock        any `json:"mock,omitempty"`        // Editor mock data
}

// ============================================================================
// VALIDATION RULES EXTENSION
// ============================================================================

// ValidationRules provides comprehensive validation configuration
// Aligned with schema validation properties
type ValidationRules struct {
	// Basic validation
	Required  bool   `json:"isRequired,omitempty"`
	Email     bool   `json:"isEmail,omitempty"`
	URL       bool   `json:"isUrl,omitempty"`
	Numeric   bool   `json:"isNumeric,omitempty"`
	Integer   bool   `json:"isInt,omitempty"`
	Float     bool   `json:"isFloat,omitempty"`
	Alpha     bool   `json:"isAlpha,omitempty"`
	Alphanumeric bool `json:"isAlphanumeric,omitempty"`
	JSON      bool   `json:"isJson,omitempty"`
	
	// Length validation
	Length    *int `json:"isLength,omitempty"`
	MinLength *int `json:"minLength,omitempty"`
	MaxLength *int `json:"maxLength,omitempty"`
	Minimum   *float64 `json:"minimum,omitempty"`
	Maximum   *float64 `json:"maximum,omitempty"`
	
	// Pattern matching
	Regex1 string `json:"matchRegexp,omitempty"`
	Regex2 string `json:"matchRegexp2,omitempty"`
	Regex3 string `json:"matchRegexp3,omitempty"`
	Regex4 string `json:"matchRegexp4,omitempty"`
	Regex5 string `json:"matchRegexp5,omitempty"`
	
	// Date/time validation
	DateTimeSame          []string `json:"isDateTimeSame,omitempty"`
	DateTimeBefore        []string `json:"isDateTimeBefore,omitempty"`
	DateTimeAfter         []string `json:"isDateTimeAfter,omitempty"`
	DateTimeSameOrBefore  []string `json:"isDateTimeSameOrBefore,omitempty"`
	DateTimeSameOrAfter   []string `json:"isDateTimeSameOrAfter,omitempty"`
	DateTimeBetween       []string `json:"isDateTimeBetween,omitempty"`
	
	// Time validation
	TimeSame          []string `json:"isTimeSame,omitempty"`
	TimeBefore        []string `json:"isTimeBefore,omitempty"`
	TimeAfter         []string `json:"isTimeAfter,omitempty"`
	TimeSameOrBefore  []string `json:"isTimeSameOrBefore,omitempty"`
	TimeSameOrAfter   []string `json:"isTimeSameOrAfter,omitempty"`
	TimeBetween       []string `json:"isTimeBetween,omitempty"`
}

// ValidationErrors provides custom error messages for validation failures
type ValidationErrors struct {
	Required       string `json:"isRequired,omitempty"`
	Email          string `json:"isEmail,omitempty"`
	URL            string `json:"isUrl,omitempty"`
	Numeric        string `json:"isNumeric,omitempty"`
	Integer        string `json:"isInt,omitempty"`
	Float          string `json:"isFloat,omitempty"`
	Alpha          string `json:"isAlpha,omitempty"`
	Alphanumeric   string `json:"isAlphanumeric,omitempty"`
	JSON           string `json:"isJson,omitempty"`
	Length         string `json:"isLength,omitempty"`
	MinLength      string `json:"minLength,omitempty"`
	MaxLength      string `json:"maxLength,omitempty"`
	Minimum        string `json:"minimum,omitempty"`
	Maximum        string `json:"maximum,omitempty"`
	Regex1         string `json:"matchRegexp,omitempty"`
	Regex2         string `json:"matchRegexp2,omitempty"`
	Regex3         string `json:"matchRegexp3,omitempty"`
	Regex4         string `json:"matchRegexp4,omitempty"`
	Regex5         string `json:"matchRegexp5,omitempty"`
	DateTimeSame   string `json:"isDateTimeSame,omitempty"`
	DateTimeBefore string `json:"isDateTimeBefore,omitempty"`
	DateTimeAfter  string `json:"isDateTimeAfter,omitempty"`
	TimeSame       string `json:"isTimeSame,omitempty"`
	TimeBefore     string `json:"isTimeBefore,omitempty"`
	TimeAfter      string `json:"isTimeAfter,omitempty"`
}

// ============================================================================
// REMOTE VALIDATION & AUTOFILL
// ============================================================================

// APIObject represents API configuration for remote operations
type APIObject struct {
	URL     string                 `json:"url"`
	Method  string                 `json:"method,omitempty"`
	Headers map[string]string      `json:"headers,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
	Silent  bool                   `json:"silent,omitempty"`
}

// AutoFillConfig provides autofill/autocomplete functionality
type AutoFillConfig struct {
	API              *APIObject             `json:"api,omitempty"`
	ShowSuggestion   bool                   `json:"showSuggestion,omitempty"`
	DefaultSelection any            `json:"defaultSelection,omitempty"`
	FillMapping      map[string]string      `json:"fillMapping,omitempty"`
	Trigger          string                 `json:"trigger,omitempty"` // change, focus, blur
	Mode             string                 `json:"mode,omitempty"`    // popOver, dialog, drawer
	Position         string                 `json:"position,omitempty"`
	Size             string                 `json:"size,omitempty"`
	Columns          []any          `json:"columns,omitempty"`
	Filter           any            `json:"filter,omitempty"`
	Silent           bool                   `json:"silent,omitempty"`
}

// ============================================================================
// EVENT SYSTEM ENHANCEMENT
// ============================================================================

// EventAction represents an action in the event system
type EventAction struct {
	ActionType string                 `json:"actionType"`
	Args       map[string]any `json:"args,omitempty"`
}

// DebounceConfig controls event debouncing
type DebounceConfig struct {
	Wait    int  `json:"wait"`    // milliseconds
	Leading bool `json:"leading,omitempty"`
	Trailing bool `json:"trailing,omitempty"`
}

// TrackConfig controls event tracking
type TrackConfig struct {
	Enable bool                   `json:"enable"`
	Data   map[string]any `json:"data,omitempty"`
}

// EventListener represents a single event listener configuration
type EventListener struct {
	Weight   int             `json:"weight,omitempty"`
	Actions  []EventAction   `json:"actions"`
	Debounce *DebounceConfig `json:"debounce,omitempty"`
	Track    *TrackConfig    `json:"track,omitempty"`
}

// EventConfiguration maps event names to their listeners
type EventConfiguration map[string]EventListener

// ============================================================================
// DISPLAY MODE ENUMERATION
// ============================================================================

// DisplayMode defines how components render in different contexts
type DisplayMode string

const (
	DisplayModeNormal     DisplayMode = "normal"     // Standard interactive mode
	DisplayModeStatic     DisplayMode = "static"     // Static display only
	DisplayModeInline     DisplayMode = "inline"     // Inline layout
	DisplayModeHorizontal DisplayMode = "horizontal" // Horizontal form layout
)

// String returns the string representation of DisplayMode
func (d DisplayMode) String() string {
	return string(d)
}

// String returns the string representation of ComponentType.
func (c ComponentType) String() string {
	return string(c)
}

// ============================================================================
// DATA ATTRIBUTES
// ============================================================================

// DataAttributes manages data-* attributes for components.
// Provides a type-safe way to add custom data attributes.
type DataAttributes map[string]string

// NewDataAttributes creates a new DataAttributes map.
func NewDataAttributes() DataAttributes {
	return make(DataAttributes)
}

// Set adds or updates a data attribute.
func (da DataAttributes) Set(key, value string) DataAttributes {
	da[key] = value
	return da
}

// Get retrieves a data attribute value.
func (da DataAttributes) Get(key string) string {
	return da[key]
}

// Has checks if a data attribute exists.
func (da DataAttributes) Has(key string) bool {
	_, exists := da[key]
	return exists
}

// Delete removes a data attribute.
func (da DataAttributes) Delete(key string) {
	delete(da, key)
}

// ToAttributes converts to map with "data-" prefix for HTML rendering.
func (da DataAttributes) ToAttributes() map[string]string {
	attrs := make(map[string]string, len(da))
	for key, value := range da {
		attrs[fmt.Sprintf("data-%s", key)] = value
	}
	return attrs
}

// MarshalJSON implements json.Marshaler interface.
func (da DataAttributes) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string(da))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (da *DataAttributes) UnmarshalJSON(data []byte) error {
	m := make(map[string]string)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*da = m
	return nil
}

// ============================================================================
// RESPONSIVE VALUE
// ============================================================================

// ResponsiveValue represents a value that can change at different breakpoints.
// Useful for responsive design patterns like "w-full md:w-1/2 lg:w-1/3".
type ResponsiveValue struct {
	Base string                // Default value (mobile-first)
	At   map[Breakpoint]string // Values at specific breakpoints
}

// NewResponsiveValue creates a new responsive value with a base value.
func NewResponsiveValue(base string) *ResponsiveValue {
	return &ResponsiveValue{
		Base: base,
		At:   make(map[Breakpoint]string),
	}
}

// AtBreakpoint sets a value for a specific breakpoint.
func (rv *ResponsiveValue) AtBreakpoint(bp Breakpoint, value string) *ResponsiveValue {
	rv.At[bp] = value
	return rv
}

// ToClasses converts responsive value to Tailwind-style classes.
// Example: "w-full md:w-1/2 lg:w-1/3"
func (rv *ResponsiveValue) ToClasses() string {
	if rv.Base == "" && len(rv.At) == 0 {
		return ""
	}

	classes := []string{rv.Base}

	// Apply in ascending breakpoint order
	breakpointOrder := []Breakpoint{BreakpointSM, BreakpointMD, BreakpointLG, BreakpointXL, BreakpointXXL}
	for _, bp := range breakpointOrder {
		if value, ok := rv.At[bp]; ok && value != "" {
			classes = append(classes, fmt.Sprintf("%s:%s", bp, value))
		}
	}

	return joinStrings(classes, " ")
}

// ============================================================================
// RESPONSIVE SIZE
// ============================================================================

// ResponsiveSize allows different sizes at different breakpoints.
type ResponsiveSize struct {
	Base Size
	At   map[Breakpoint]Size
}

// NewResponsiveSize creates a responsive size configuration.
func NewResponsiveSize(base Size) *ResponsiveSize {
	return &ResponsiveSize{
		Base: base,
		At:   make(map[Breakpoint]Size),
	}
}

// AtBreakpoint sets size for specific breakpoint.
func (rs *ResponsiveSize) AtBreakpoint(bp Breakpoint, size Size) *ResponsiveSize {
	rs.At[bp] = size
	return rs
}

// GetCurrentSize returns the size for a given breakpoint (or base if not specified).
func (rs *ResponsiveSize) GetCurrentSize(bp Breakpoint) Size {
	if size, ok := rs.At[bp]; ok {
		return size
	}
	return rs.Base
}

// ============================================================================
// JSON HELPERS
// ============================================================================

// ToJSON converts any value to JSON bytes.
func ToJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

// FromJSON converts JSON bytes to a value.
func FromJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// ToJSONString converts any value to a JSON string.
func ToJSONString(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// ============================================================================
// BUSINESS DOMAIN TYPES
// ============================================================================

// Permission represents a system permission for user management
type Permission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Actions     []string `json:"actions,omitempty"` // create, read, update, delete
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// joinStrings joins non-empty strings with a separator.
func joinStrings(parts []string, sep string) string {
	var result []string
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	if len(result) == 0 {
		return ""
	}

	// Manual join to avoid importing strings package
	if len(result) == 1 {
		return result[0]
	}

	n := len(sep) * (len(result) - 1)
	for _, s := range result {
		n += len(s)
	}

	b := make([]byte, n)
	bp := copy(b, result[0])
	for _, s := range result[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], s)
	}
	return string(b)
}
