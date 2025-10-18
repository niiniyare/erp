package atoms

import (
	"fmt"
	"strings"
)

// ============================================================================
// BASE PROPERTIES
// ============================================================================

// BaseProps contains fundamental HTML attributes shared by all components.
// Every interactive component should embed this struct for consistency.
type BaseProps struct {
	ID         string `json:"id,omitempty"`         // HTML id attribute
	Name       string `json:"name,omitempty"`       // Form field name
	Class      string `json:"class,omitempty"`      // Custom CSS classes
	Style      string `json:"style,omitempty"`      // Inline styles (use sparingly)
	DataTestID string `json:"dataTestId,omitempty"` // For testing frameworks
	TabIndex   int    `json:"tabIndex,omitempty"`   // Custom tab order (0=default, -1=not focusable)
}

// HasID checks if the component has an ID set.
func (b BaseProps) HasID() bool {
	return strings.TrimSpace(b.ID) != ""
}

// GetID returns the ID or generates one if not set.
func (b BaseProps) GetID(prefix string) string {
	return EnsureID(b.ID, prefix)
}

// GetClasses returns combined class string with optional additional classes.
func (b BaseProps) GetClasses(additionalClasses ...string) string {
	classes := make([]string, 0, len(additionalClasses)+1)
	if b.Class != "" {
		classes = append(classes, b.Class)
	}
	classes = append(classes, additionalClasses...)
	return JoinClasses(classes...)
}

// ============================================================================
// ACCESSIBILITY PROPERTIES
// ============================================================================

// AccessibilityProps contains ARIA and accessibility-related attributes.
// These improve screen reader support and overall accessibility compliance.
type AccessibilityProps struct {
	AriaLabel       string `json:"ariaLabel,omitempty"`       // Accessible name
	AriaDescribedBy string `json:"ariaDescribedBy,omitempty"` // ID of describing element
	AriaLabelledBy  string `json:"ariaLabelledBy,omitempty"`  // ID of labeling element
	AriaRequired    bool   `json:"ariaRequired,omitempty"`    // Marks field as required
	AriaInvalid     bool   `json:"ariaInvalid,omitempty"`     // Marks field as invalid
	AriaDisabled    bool   `json:"ariaDisabled,omitempty"`    // Marks field as disabled
	AriaPlaceholder string `json:"ariaPlaceholder,omitempty"` // Accessible placeholder
	AriaControls    string `json:"ariaControls,omitempty"`    // ID of controlled element
	AriaExpanded    *bool  `json:"ariaExpanded,omitempty"`    // Expandable state (pointer for 3-state)
	AriaHasPopup    string `json:"ariaHasPopup,omitempty"`    // Popup type (menu, dialog, etc.)
	AriaLive        string `json:"ariaLive,omitempty"`        // Live region (polite, assertive)
	AriaHidden      bool   `json:"ariaHidden,omitempty"`      // Hidden from screen readers
	Role            string `json:"role,omitempty"`            // ARIA role override
}

// GetAriaAttributes returns a map of aria attributes for rendering.
// Returns only non-empty values to keep HTML clean.
func (a AccessibilityProps) GetAriaAttributes() map[string]string {
	attrs := make(map[string]string)

	if a.AriaLabel != "" {
		attrs["aria-label"] = a.AriaLabel
	}
	if a.AriaDescribedBy != "" {
		attrs["aria-describedby"] = a.AriaDescribedBy
	}
	if a.AriaLabelledBy != "" {
		attrs["aria-labelledby"] = a.AriaLabelledBy
	}
	if a.AriaRequired {
		attrs["aria-required"] = "true"
	}
	if a.AriaInvalid {
		attrs["aria-invalid"] = "true"
	}
	if a.AriaDisabled {
		attrs["aria-disabled"] = "true"
	}
	if a.AriaPlaceholder != "" {
		attrs["aria-placeholder"] = a.AriaPlaceholder
	}
	if a.AriaControls != "" {
		attrs["aria-controls"] = a.AriaControls
	}
	if a.AriaExpanded != nil {
		attrs["aria-expanded"] = fmt.Sprintf("%t", *a.AriaExpanded)
	}
	if a.AriaHasPopup != "" {
		attrs["aria-haspopup"] = a.AriaHasPopup
	}
	if a.AriaLive != "" {
		attrs["aria-live"] = a.AriaLive
	}
	if a.AriaHidden {
		attrs["aria-hidden"] = "true"
	}
	if a.Role != "" {
		attrs["role"] = a.Role
	}

	return attrs
}

// HasAriaLabel checks if any form of ARIA label is set.
func (a AccessibilityProps) HasAriaLabel() bool {
	return a.AriaLabel != "" || a.AriaLabelledBy != ""
}

// ============================================================================
// VALIDATION PROPERTIES
// ============================================================================

// ValidationProps handles validation state and feedback messages.
// Supports multiple message types for different validation states.
type ValidationProps struct {
	State            ValidationState `json:"state,omitempty"`            // Current validation state
	HelpText         string          `json:"helpText,omitempty"`         // General guidance text
	ErrorText        string          `json:"errorText,omitempty"`        // Error message
	SuccessText      string          `json:"successText,omitempty"`      // Success message
	WarningText      string          `json:"warningText,omitempty"`      // Warning message
	InfoText         string          `json:"infoText,omitempty"`         // Info message
	ShowValidation   bool            `json:"showValidation,omitempty"`   // Whether to show validation UI
	ValidateOnBlur   bool            `json:"validateOnBlur,omitempty"`   // Trigger validation on blur
	ValidateOnChange bool            `json:"validateOnChange,omitempty"` // Trigger validation on change
}

// GetFeedbackMessage returns the appropriate message based on current state.
// Priority: Error > Success > Warning > Info > HelpText
func (v ValidationProps) GetFeedbackMessage() (string, ValidationState) {
	switch v.State {
	case StateError:
		if v.ErrorText != "" {
			return v.ErrorText, StateError
		}
	case StateSuccess:
		if v.SuccessText != "" {
			return v.SuccessText, StateSuccess
		}
	case StateWarning:
		if v.WarningText != "" {
			return v.WarningText, StateWarning
		}
	case StateInfo:
		if v.InfoText != "" {
			return v.InfoText, StateInfo
		}
	}

	// Fallback to help text with default state
	if v.HelpText != "" {
		return v.HelpText, StateDefault
	}

	return "", StateDefault
}

// HasFeedback checks if there's any feedback message to display.
func (v ValidationProps) HasFeedback() bool {
	msg, _ := v.GetFeedbackMessage()
	return msg != ""
}

// IsValid checks if the current state indicates a valid input.
func (v ValidationProps) IsValid() bool {
	return v.State == StateSuccess || v.State == StateDefault
}

// IsInvalid checks if the current state indicates an invalid input.
func (v ValidationProps) IsInvalid() bool {
	return v.State == StateError
}

// ============================================================================
// INTERACTION PROPERTIES
// ============================================================================

// InteractionProps handles user interaction states.
// Controls how users can interact with the component.
type InteractionProps struct {
	Disabled  bool `json:"disabled,omitempty"`  // Component is disabled
	Required  bool `json:"required,omitempty"`  // Field is required
	ReadOnly  bool `json:"readonly,omitempty"`  // Field is read-only
	AutoFocus bool `json:"autofocus,omitempty"` // Auto-focus on page load
}

// IsInteractive checks if the component accepts user input.
func (i InteractionProps) IsInteractive() bool {
	return !i.Disabled && !i.ReadOnly
}

// ShouldShowRequired checks if required indicator should be shown.
func (i InteractionProps) ShouldShowRequired() bool {
	return i.Required && !i.Disabled
}

// GetInteractionAttrs returns HTML attributes for interaction state.
func (i InteractionProps) GetInteractionAttrs() map[string]string {
	attrs := make(map[string]string)

	if i.Disabled {
		attrs["disabled"] = "disabled"
	}
	if i.Required {
		attrs["required"] = "required"
	}
	if i.ReadOnly {
		attrs["readonly"] = "readonly"
	}
	if i.AutoFocus {
		attrs["autofocus"] = "autofocus"
	}

	return attrs
}

// ============================================================================
// ALPINE.JS EVENT HANDLERS
// ============================================================================

// AlpineEventHandlers contains Alpine.js event handling directives.
// These will be rendered as x-on:* attributes.
type AlpineEventHandlers struct {
	OnChange     string `json:"onChange,omitempty"`     // x-on:change
	OnInput      string `json:"onInput,omitempty"`      // x-on:input
	OnFocus      string `json:"onFocus,omitempty"`      // x-on:focus
	OnBlur       string `json:"onBlur,omitempty"`       // x-on:blur
	OnClick      string `json:"onClick,omitempty"`      // x-on:click
	OnKeyDown    string `json:"onKeyDown,omitempty"`    // x-on:keydown
	OnKeyUp      string `json:"onKeyUp,omitempty"`      // x-on:keyup
	OnMouseEnter string `json:"onMouseEnter,omitempty"` // x-on:mouseenter
	OnMouseLeave string `json:"onMouseLeave,omitempty"` // x-on:mouseleave
	OnSubmit     string `json:"onSubmit,omitempty"`     // x-on:submit
}

// GetEventAttributes returns a map of Alpine.js event attributes.
// Automatically prefixes with x-on: for Alpine.js.
func (a AlpineEventHandlers) GetEventAttributes() map[string]string {
	attrs := make(map[string]string)

	if a.OnChange != "" {
		attrs["x-on:change"] = a.OnChange
	}
	if a.OnInput != "" {
		attrs["x-on:input"] = a.OnInput
	}
	if a.OnFocus != "" {
		attrs["x-on:focus"] = a.OnFocus
	}
	if a.OnBlur != "" {
		attrs["x-on:blur"] = a.OnBlur
	}
	if a.OnClick != "" {
		attrs["x-on:click"] = a.OnClick
	}
	if a.OnKeyDown != "" {
		attrs["x-on:keydown"] = a.OnKeyDown
	}
	if a.OnKeyUp != "" {
		attrs["x-on:keyup"] = a.OnKeyUp
	}
	if a.OnMouseEnter != "" {
		attrs["x-on:mouseenter"] = a.OnMouseEnter
	}
	if a.OnMouseLeave != "" {
		attrs["x-on:mouseleave"] = a.OnMouseLeave
	}
	if a.OnSubmit != "" {
		attrs["x-on:submit"] = a.OnSubmit
	}

	return attrs
}

// HasEventHandlers checks if any event handlers are defined.
func (a AlpineEventHandlers) HasEventHandlers() bool {
	return a.OnChange != "" || a.OnInput != "" || a.OnFocus != "" ||
		a.OnBlur != "" || a.OnClick != "" || a.OnKeyDown != "" ||
		a.OnKeyUp != "" || a.OnMouseEnter != "" || a.OnMouseLeave != "" ||
		a.OnSubmit != ""
}

// ============================================================================
// LABEL PROPERTIES
// ============================================================================

// LabelProps contains label-specific properties.
// Used by components that display labels.
type LabelProps struct {
	Label         string        `json:"label,omitempty"`         // Label text
	LabelPosition LabelPosition `json:"labelPosition,omitempty"` // Label position
	LabelClass    string        `json:"labelClass,omitempty"`    // Custom label classes
	HideLabel     bool          `json:"hideLabel,omitempty"`     // Visually hide label (keep for a11y)
}

// HasLabel checks if a label should be rendered.
func (l LabelProps) HasLabel() bool {
	return strings.TrimSpace(l.Label) != ""
}

// ShouldRenderLabel checks if label should be visible.
func (l LabelProps) ShouldRenderLabel() bool {
	return l.HasLabel() && !l.HideLabel && l.LabelPosition != LabelNone
}

// GetLabelClasses returns classes for the label element.
func (l LabelProps) GetLabelClasses(size Size, state ValidationState) string {
	classes := []string{
		"block", // Labels are typically block elements
		"font-medium",
	}

	// Add size-specific text size
	switch size {
	case SizeXS:
		classes = append(classes, "text-xs")
	case SizeSM:
		classes = append(classes, "text-sm")
	case SizeLG:
		classes = append(classes, "text-base")
	case SizeXL:
		classes = append(classes, "text-lg")
	default:
		classes = append(classes, "text-sm")
	}

	// Add state-specific colors
	classes = append(classes, GetLabelColorClasses(state)...)

	// Add custom classes
	if l.LabelClass != "" {
		classes = append(classes, l.LabelClass)
	}

	return JoinClasses(classes...)
}

// ============================================================================
// PLACEHOLDER PROPERTIES
// ============================================================================

// PlaceholderProps contains placeholder-related properties.
type PlaceholderProps struct {
	Placeholder      string `json:"placeholder,omitempty"`      // Placeholder text
	FloatingLabel    bool   `json:"floatingLabel,omitempty"`    // Material-style floating label
	PlaceholderClass string `json:"placeholderClass,omitempty"` // Custom placeholder styling
}

// HasPlaceholder checks if placeholder text is set.
func (p PlaceholderProps) HasPlaceholder() bool {
	return strings.TrimSpace(p.Placeholder) != ""
}

// ============================================================================
// ICON PROPERTIES
// ============================================================================

// IconProps contains icon configuration.
type IconProps struct {
	Name     string       `json:"name,omitempty"`     // Icon name (e.g., lucide icon name)
	Position IconPosition `json:"position,omitempty"` // Icon position
	Size     Size         `json:"size,omitempty"`     // Icon size
	Color    string       `json:"color,omitempty"`    // Icon color
	Class    string       `json:"class,omitempty"`    // Custom icon classes
}

// HasIcon checks if an icon is configured.
func (i IconProps) HasIcon() bool {
	return strings.TrimSpace(i.Name) != ""
}

// GetIconClasses returns complete icon class string.
func (i IconProps) GetIconClasses() string {
	classes := []string{}

	// Add size classes
	switch i.Size {
	case SizeXS:
		classes = append(classes, "w-3", "h-3")
	case SizeSM:
		classes = append(classes, "w-4", "h-4")
	case SizeLG:
		classes = append(classes, "w-6", "h-6")
	case SizeXL:
		classes = append(classes, "w-8", "h-8")
	default:
		classes = append(classes, "w-5", "h-5")
	}

	// Add color if specified
	if i.Color != "" {
		classes = append(classes, i.Color)
	}

	// Add position-based spacing
	switch i.Position {
	case IconLeft:
		classes = append(classes, "mr-2")
	case IconRight:
		classes = append(classes, "ml-2")
	case IconTop:
		classes = append(classes, "mb-2")
	case IconBottom:
		classes = append(classes, "mt-2")
	}

	// Add custom classes
	if i.Class != "" {
		classes = append(classes, i.Class)
	}

	return JoinClasses(classes...)
}

// ============================================================================
// COMPONENT-SPECIFIC PROPS
// ============================================================================

// ButtonProps defines properties for Button components.
type ButtonProps struct {
	BaseProps
	AccessibilityProps
	InteractionProps
	AlpineEventHandlers

	// Content
	Text string    `json:"text,omitempty"` // Button text
	Icon IconProps `json:"icon,omitempty"` // Icon configuration

	// Styling
	Variant     Variant     `json:"variant,omitempty"`     // Visual variant
	Size        Size        `json:"size,omitempty"`        // Button size
	ColorScheme ColorScheme `json:"colorScheme,omitempty"` // Color theme
	FullWidth   bool        `json:"fullWidth,omitempty"`   // Full width button

	// State
	Loading bool `json:"loading,omitempty"` // Loading state
	Active  bool `json:"active,omitempty"`  // Active state

	// Button-specific
	Type string `json:"type,omitempty"` // button, submit, reset
	Form string `json:"form,omitempty"` // Associated form ID
}

// InputProps defines properties for Input components.
type InputProps struct {
	BaseProps
	AccessibilityProps
	ValidationProps
	InteractionProps
	LabelProps
	PlaceholderProps
	AlpineEventHandlers

	// Input-specific
	Type      InputType `json:"type,omitempty"`      // Input type
	Value     string    `json:"value,omitempty"`     // Input value
	Size      Size      `json:"size,omitempty"`      // Input size
	MaxLength int       `json:"maxLength,omitempty"` // Maximum length
	MinLength int       `json:"minLength,omitempty"` // Minimum length
	Pattern   string    `json:"pattern,omitempty"`   // Validation pattern
	Min       string    `json:"min,omitempty"`       // Min value (for number/date)
	Max       string    `json:"max,omitempty"`       // Max value (for number/date)
	Step      string    `json:"step,omitempty"`      // Step value (for number)

	// Icons
	LeftIcon  IconProps `json:"leftIcon,omitempty"`  // Icon on left
	RightIcon IconProps `json:"rightIcon,omitempty"` // Icon on right
}

// CheckboxProps defines properties for Checkbox components.
type CheckboxProps struct {
	BaseProps
	AccessibilityProps
	ValidationProps
	InteractionProps
	LabelProps
	AlpineEventHandlers

	// Checkbox-specific
	Checked       bool   `json:"checked,omitempty"`       // Checked state
	Indeterminate bool   `json:"indeterminate,omitempty"` // Indeterminate state
	Value         string `json:"value,omitempty"`         // Checkbox value
	Size          Size   `json:"size,omitempty"`          // Checkbox size
}

// RadioProps defines properties for Radio button components.
type RadioProps struct {
	BaseProps
	AccessibilityProps
	ValidationProps
	InteractionProps
	LabelProps
	AlpineEventHandlers

	// Radio-specific
	Checked bool   `json:"checked,omitempty"` // Checked state
	Value   string `json:"value,omitempty"`   // Radio value
	Size    Size   `json:"size,omitempty"`    // Radio size
}

// SelectProps defines properties for Select components.
type SelectProps struct {
	BaseProps
	AccessibilityProps
	ValidationProps
	InteractionProps
	LabelProps
	PlaceholderProps
	AlpineEventHandlers

	// Select-specific
	Options  []SelectOption `json:"options,omitempty"`  // Select options
	Value    string         `json:"value,omitempty"`    // Selected value
	Multiple bool           `json:"multiple,omitempty"` // Allow multiple selection
	Size     Size           `json:"size,omitempty"`     // Select size
}

// SelectOption represents an option in a select dropdown.
type SelectOption struct {
	Value    string `json:"value"`              // Option value
	Label    string `json:"label"`              // Option label
	Disabled bool   `json:"disabled,omitempty"` // Option disabled state
	Group    string `json:"group,omitempty"`    // Option group
}

// TextareaProps defines properties for Textarea components.
type TextareaProps struct {
	BaseProps
	AccessibilityProps
	ValidationProps
	InteractionProps
	LabelProps
	PlaceholderProps
	AlpineEventHandlers

	// Textarea-specific
	Value     string `json:"value,omitempty"`     // Textarea value
	Rows      int    `json:"rows,omitempty"`      // Number of rows
	Cols      int    `json:"cols,omitempty"`      // Number of columns
	MaxLength int    `json:"maxLength,omitempty"` // Maximum length
	Size      Size   `json:"size,omitempty"`      // Textarea size
	Resizable bool   `json:"resizable,omitempty"` // Allow resizing
}

// ToggleProps defines properties for Toggle/Switch components.
type ToggleProps struct {
	BaseProps
	AccessibilityProps
	InteractionProps
	LabelProps
	AlpineEventHandlers

	// Toggle-specific
	Checked     bool        `json:"checked,omitempty"`     // Checked state
	Size        Size        `json:"size,omitempty"`        // Toggle size
	ColorScheme ColorScheme `json:"colorScheme,omitempty"` // Color theme
}

// SpinnerProps defines properties for Spinner/Loader components.
type SpinnerProps struct {
	BaseProps
	AccessibilityProps

	// Spinner-specific
	Size    Size        `json:"size,omitempty"`    // Spinner size
	Color   ColorScheme `json:"color,omitempty"`   // Spinner color
	Variant Variant     `json:"variant,omitempty"` // Spinner variant (border, dots, pulse)
}
