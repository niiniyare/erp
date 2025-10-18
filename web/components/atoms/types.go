// Package atoms - Atomic UI components with schema integration
// This package contains all atom-level components, their types, utilities, and schema integration
package atoms

import (
	"encoding/json"
	"fmt"

	"github.com/niiniyare/erp/web/components/atoms/css"
)

// ============================================================================
// ATOM COMPONENT TYPES
// ============================================================================

// BaseAtomProps provides common properties for all atom components
type BaseAtomProps struct {
	// Common properties
	ID        string `json:"id,omitempty"`
	ClassName string `json:"className,omitempty"`
	TestID    string `json:"testId,omitempty"`

	// State properties
	Disabled bool `json:"disabled,omitempty"`
	Hidden   bool `json:"hidden,omitempty"`

	// Accessibility properties
	AriaLabel       string `json:"ariaLabel,omitempty"`
	AriaDescribedBy string `json:"ariaDescribedBy,omitempty"`
	Role            string `json:"role,omitempty"`
	TabIndex        int    `json:"tabIndex,omitempty"`
}

// ButtonProps defines properties for Button atom components
type ButtonProps struct {
	BaseAtomProps

	// Content properties
	Text  string `json:"text,omitempty"`
	Label string `json:"label,omitempty"` // Alternative to Text
	Icon  string `json:"icon,omitempty"`

	// Styling properties
	Variant string `json:"variant,omitempty"` // primary, secondary, success, danger, etc.
	Size    string `json:"size,omitempty"`    // xs, sm, md, lg, xl
	Block   bool   `json:"block,omitempty"`   // Full width button

	// State properties
	Loading bool `json:"loading,omitempty"`

	// Button-specific properties
	Type string `json:"type,omitempty"` // button, submit, reset
	Form string `json:"form,omitempty"` // Associated form ID
}

// ButtonSize defines the size variants for buttons
type ButtonSize string

const (
	ButtonSizeXS ButtonSize = "xs"
	ButtonSizeSM ButtonSize = "sm"
	ButtonSizeMD ButtonSize = "md"
	ButtonSizeLG ButtonSize = "lg"
	ButtonSizeXL ButtonSize = "xl"
)

// IconSize defines the size variants for icons
type IconSize string

const (
	IconSizeXS  IconSize = "xs"
	IconSizeSM  IconSize = "sm"
	IconSizeMD  IconSize = "md"
	IconSizeLG  IconSize = "lg"
	IconSizeXL  IconSize = "xl"
	IconSize2XL IconSize = "2xl"
)

// ============================================================================
// ATOM COMPONENT UTILITIES
// ============================================================================

// GetAtomCSS generates CSS for an atom component using the self-contained CSS factory
func GetAtomCSS(componentType string, variant string) (*css.Styles, error) {
	factory := css.NewFactory()

	switch componentType {
	case "button":
		return factory.ButtonStyles(variant), nil
	case "input":
		return factory.InputStyles(variant), nil
	case "checkbox":
		return factory.CheckboxStyles(variant), nil
	case "radio":
		return factory.RadioStyles(variant), nil
	case "select":
		return factory.SelectStyles(variant), nil
	case "textarea":
		return factory.TextareaStyles(variant), nil
	case "icon":
		return factory.IconStyles(variant), nil
	case "spinner":
		return factory.SpinnerStyles(variant), nil
	case "toggle":
		return factory.ToggleStyles(variant), nil
	default:
		return css.NewStyles(), fmt.Errorf("unsupported component type: %s", componentType)
	}
}

// CreateAtomFromJSON creates atom props from JSON configuration
func CreateAtomFromJSON[T any](jsonData []byte) (T, error) {
	var props T
	err := json.Unmarshal(jsonData, &props)
	return props, err
}

// FromJSON creates atom props from JSON configuration
func FromJSON[T any](jsonData []byte) (T, error) {
	var props T
	err := json.Unmarshal(jsonData, &props)
	return props, err
}

// ToJSON converts atom props to JSON
func ToJSON[T any](props T) ([]byte, error) {
	return json.Marshal(props)
}

// ValidateAtomProps validates atom component properties
func ValidateAtomProps(props any) error {
	// Basic validation - can be extended with schema validation
	return nil
}

// ============================================================================
// COMPONENT FACTORIES AND HELPERS
// ============================================================================

// GetButtonCSS generates CSS string for button component
func GetButtonCSS(variant string) string {
	if variant == "" {
		variant = "primary"
	}
	factory := css.NewFactory()
	styles := factory.ButtonStyles(variant)
	return styles.ToCSS()
}

// GetButtonDisplayText returns the appropriate display text for button
func GetButtonDisplayText(props ButtonProps) string {
	if props.Text != "" {
		return props.Text
	}
	if props.Label != "" {
		return props.Label
	}
	return ""
}

// GetIconCSS generates CSS string for icon component
func GetIconCSS(size, color string) string {
	if size == "" {
		size = "md"
	}
	factory := css.NewFactory()
	styles := factory.IconStyles(size)
	
	// Apply color if specified
	if color != "" {
		styles.WithColor(color)
	}
	
	return styles.ToCSS()
}

// GetInputCSS generates CSS string for input component
func GetInputCSS(state string) string {
	if state == "" {
		state = "default"
	}
	factory := css.NewFactory()
	styles := factory.InputStyles(state)
	return styles.ToCSS()
}

// GetRadioCSS generates CSS string for radio component
func GetRadioCSS(state string) string {
	if state == "" {
		state = "default"
	}
	factory := css.NewFactory()
	styles := factory.RadioStyles(state)
	return styles.ToCSS()
}

// ButtonFromJSON creates a ButtonProps from JSON configuration
func ButtonFromJSON(jsonData []byte) (ButtonProps, error) {
	var props ButtonProps
	err := json.Unmarshal(jsonData, &props)
	return props, err
}

// CreateButtonProps creates ButtonProps from configuration
func CreateButtonProps(id, text, variant string) ButtonProps {
	return ButtonProps{
		BaseAtomProps: BaseAtomProps{
			ID: id,
		},
		Text:    text,
		Variant: variant,
	}
}

// CheckboxFromJSON creates a CheckboxProps from JSON configuration
func CheckboxFromJSON(jsonData []byte) (CheckboxProps, error) {
	var props CheckboxProps
	err := json.Unmarshal(jsonData, &props)
	return props, err
}

// InputFromJSON creates a InputProps from JSON configuration
func InputFromJSON(jsonData []byte) (InputProps, error) {
	var props InputProps
	err := json.Unmarshal(jsonData, &props)
	return props, err
}

// SelectFromJSON creates a SelectProps from JSON configuration
func SelectFromJSON(jsonData []byte) (SelectProps, error) {
	var props SelectProps
	err := json.Unmarshal(jsonData, &props)
	return props, err
}

// RadioFromJSON creates a RadioProps from JSON configuration
func RadioFromJSON(jsonData []byte) (RadioProps, error) {
	var props RadioProps
	err := json.Unmarshal(jsonData, &props)
	return props, err
}

// IconFromJSON creates a IconProps from JSON configuration
func IconFromJSON(jsonData []byte) (IconProps, error) {
	var props IconProps
	err := json.Unmarshal(jsonData, &props)
	return props, err
}

// SpinnerFromJSON creates a SpinnerProps from JSON configuration
func SpinnerFromJSON(jsonData []byte) (SpinnerProps, error) {
	var props SpinnerProps
	err := json.Unmarshal(jsonData, &props)
	return props, err
}

// TextareaFromJSON creates a TextareaProps from JSON configuration
func TextareaFromJSON(jsonData []byte) (TextareaProps, error) {
	var props TextareaProps
	err := json.Unmarshal(jsonData, &props)
	return props, err
}

// ToggleFromJSON creates a ToggleProps from JSON configuration
func ToggleFromJSON(jsonData []byte) (ToggleProps, error) {
	var props ToggleProps
	err := json.Unmarshal(jsonData, &props)
	return props, err
}

// helper function to safely extract string values from config
func getStringFromConfig(config map[string]any, key string, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}
