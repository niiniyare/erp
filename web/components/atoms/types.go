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
	ColorDefault   ColorScheme = "default"   // Blue/primary colors
	ColorPrimary   ColorScheme = "primary"   // Primary brand color
	ColorSecondary ColorScheme = "secondary" // Secondary brand color
	ColorSuccess   ColorScheme = "success"   // Green - success states
	ColorDanger    ColorScheme = "danger"    // Red - error/danger states
	ColorWarning   ColorScheme = "warning"   // Yellow/orange - warning states
	ColorInfo      ColorScheme = "info"      // Blue - informational states
	ColorGray      ColorScheme = "gray"      // Neutral gray
	ColorNeutral   ColorScheme = "neutral"   // Alias for gray
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
// COMPONENT TYPE REGISTRY
// ============================================================================

// ComponentType represents the type of atomic component.
type ComponentType string

const (
	ComponentButton   ComponentType = "button"
	ComponentInput    ComponentType = "input"
	ComponentCheckbox ComponentType = "checkbox"
	ComponentRadio    ComponentType = "radio"
	ComponentSelect   ComponentType = "select"
	ComponentTextarea ComponentType = "textarea"
	ComponentIcon     ComponentType = "icon"
	ComponentSpinner  ComponentType = "spinner"
	ComponentToggle   ComponentType = "toggle"
	ComponentLabel    ComponentType = "label"
	ComponentBadge    ComponentType = "badge"
	ComponentAvatar   ComponentType = "avatar"
)

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
func ToJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// FromJSON converts JSON bytes to a value.
func FromJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ToJSONString converts any value to a JSON string.
func ToJSONString(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
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
