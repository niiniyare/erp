package atoms

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
)

// ============================================================================
// BASE PROPERTIES - SCHEMA ALIGNED
// ============================================================================

// BaseProps contains fundamental HTML attributes shared by all components.
// Enhanced with schema-aligned properties for consistency.
type BaseProps struct {
	// Core HTML attributes
	ID         string `json:"id,omitempty"`         // HTML id attribute
	Name       string `json:"name,omitempty"`       // Form field name
	Class      string `json:"className,omitempty"`  // Custom CSS classes (aligned with schema)
	Style      interface{} `json:"style,omitempty"` // Inline styles (object or string)
	DataTestID string `json:"dataTestId,omitempty"` // For testing frameworks
	TabIndex   int    `json:"tabIndex,omitempty"`   // Custom tab order (0=default, -1=not focusable)
	
	// Schema-aligned properties
	SchemaID   string `json:"$$id,omitempty"`       // Schema designer unique ID
	Ref        string `json:"$ref,omitempty"`       // Schema reference for definitions
	UseMobileUI *bool `json:"useMobileUI,omitempty"` // Mobile UI override
	
	// Conditional display (disabled is handled by InteractionProps)
	DisabledOn *SchemaExpression  `json:"disabledOn,omitempty"` // Dynamic disabled condition
	Hidden     bool               `json:"hidden,omitempty"`     // Static hidden state
	HiddenOn   *SchemaExpression  `json:"hiddenOn,omitempty"`   // Dynamic hidden condition
	Visible    bool               `json:"visible,omitempty"`    // Static visible state
	VisibleOn  *SchemaExpression  `json:"visibleOn,omitempty"`  // Dynamic visible condition
	
	// Static display mode
	Static              bool               `json:"static,omitempty"`              // Enable static display mode
	StaticOn            *SchemaExpression  `json:"staticOn,omitempty"`            // Dynamic static condition
	StaticPlaceholder   string             `json:"staticPlaceholder,omitempty"`   // Static mode placeholder
	StaticClassName     string             `json:"staticClassName,omitempty"`     // Static mode CSS classes
	StaticLabelClassName string            `json:"staticLabelClassName,omitempty"` // Static label CSS classes
	StaticInputClassName string            `json:"staticInputClassName,omitempty"` // Static input CSS classes
	StaticSchema        interface{}        `json:"staticSchema,omitempty"`        // Static display schema
	
	// Event configuration
	OnEvent     EventConfiguration `json:"onEvent,omitempty"`     // Event listeners configuration
	
	// Editor metadata
	EditorSetting *EditorSetting `json:"editorSetting,omitempty"` // Visual editor configuration
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

// IsDisabledByExpression checks if the component should be disabled by dynamic condition.
func (b BaseProps) IsDisabledByExpression() bool {
	if b.DisabledOn != nil {
		return b.DisabledOn.IsTrue()
	}
	return false
}

// IsHidden checks if the component should be hidden (static or dynamic).
func (b BaseProps) IsHidden() bool {
	if b.Hidden {
		return true
	}
	if b.HiddenOn != nil {
		return b.HiddenOn.IsTrue()
	}
	return false
}

// IsVisible checks if the component should be visible (considering all visibility rules).
func (b BaseProps) IsVisible() bool {
	// If explicitly hidden, not visible
	if b.IsHidden() {
		return false
	}
	
	// If visible is explicitly set to false, not visible
	if !b.Visible && b.VisibleOn == nil {
		return true // Default to visible if not specified
	}
	
	// Check dynamic visibility
	if b.VisibleOn != nil {
		return b.VisibleOn.IsTrue()
	}
	
	return b.Visible
}

// IsStaticMode checks if the component should render in static mode.
func (b BaseProps) IsStaticMode() bool {
	if b.Static {
		return true
	}
	if b.StaticOn != nil {
		return b.StaticOn.IsTrue()
	}
	return false
}

// GetStyleAttribute returns the style attribute as a string for HTML rendering.
func (b BaseProps) GetStyleAttribute() string {
	if b.Style == nil {
		return ""
	}
	
	switch v := b.Style.(type) {
	case string:
		return v
	case map[string]interface{}:
		// Convert object-style styles to CSS string
		var styles []string
		for prop, value := range v {
			if valueStr, ok := value.(string); ok {
				styles = append(styles, fmt.Sprintf("%s: %s", prop, valueStr))
			}
		}
		return strings.Join(styles, "; ")
	default:
		return ""
	}
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
	AriaPressed     *bool  `json:"ariaPressed,omitempty"`     // Pressed state (pointer for 3-state)
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
// VALIDATION PROPERTIES - SCHEMA ALIGNED
// ============================================================================

// ValidationProps handles validation state and feedback messages.
// Enhanced with schema-aligned validation capabilities.
type ValidationProps struct {
	// Basic validation state
	State            ValidationState `json:"state,omitempty"`            // Current validation state
	HelpText         string          `json:"helpText,omitempty"`         // General guidance text
	ErrorText        string          `json:"errorText,omitempty"`        // Error message
	SuccessText      string          `json:"successText,omitempty"`      // Success message
	WarningText      string          `json:"warningText,omitempty"`      // Warning message
	InfoText         string          `json:"infoText,omitempty"`         // Info message
	ShowValidation   bool            `json:"showValidation,omitempty"`   // Whether to show validation UI
	ValidateOnBlur   bool            `json:"validateOnBlur,omitempty"`   // Trigger validation on blur
	ValidateOnChange bool            `json:"validateOnChange,omitempty"` // Trigger validation on change
	
	// Schema-aligned validation
	Rules            *ValidationRules  `json:"validations,omitempty"`      // Validation rules configuration
	ErrorMessages    *ValidationErrors `json:"validationErrors,omitempty"` // Custom error messages
	ValidateAPI      *APIObject        `json:"validateApi,omitempty"`      // Remote validation API
	AutoFill         *AutoFillConfig   `json:"autoFill,omitempty"`         // Autofill configuration
	InitAutoFill     interface{}       `json:"initAutoFill,omitempty"`     // Initial autofill behavior
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

// GetValidationMessage returns just the message string for validation display.
func (v ValidationProps) GetValidationMessage() string {
	msg, _ := v.GetFeedbackMessage()
	return msg
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

// EventProps contains event handling properties for components.
// Alias for AlpinEventHandlers for backward compatibility.
type EventProps = AlpinEventHandlers

// AlpinEventHandlers contains Alpine.js event handling directives.
// These will be rendered as x-on:* attributes.
type AlpinEventHandlers struct {
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
	OnLoad       string `json:"onLoad,omitempty"`       // x-on:load
}

// GetEventAttributes returns a map of Alpine.js event attributes.
// Automatically prefixes with x-on: for Alpine.js.
func (a AlpinEventHandlers) GetEventAttributes() map[string]string {
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
	if a.OnLoad != "" {
		attrs["x-on:load"] = a.OnLoad
	}

	return attrs
}

// HasEventHandlers checks if any event handlers are defined.
func (a AlpinEventHandlers) HasEventHandlers() bool {
	return a.OnChange != "" || a.OnInput != "" || a.OnFocus != "" ||
		a.OnBlur != "" || a.OnClick != "" || a.OnKeyDown != "" ||
		a.OnKeyUp != "" || a.OnMouseEnter != "" || a.OnMouseLeave != "" ||
		a.OnSubmit != "" || a.OnLoad != ""
}

// ============================================================================
// LABEL PROPERTIES - SCHEMA ALIGNED
// ============================================================================

// LabelProps contains label-specific properties.
// Enhanced with schema-aligned label configuration.
type LabelProps struct {
	// Basic label properties
	Label         interface{}   `json:"label,omitempty"`         // Label text (string or false to hide)
	LabelPosition LabelPosition `json:"labelPosition,omitempty"` // Label position
	LabelClass    string        `json:"labelClassName,omitempty"` // Custom label classes (aligned with schema)
	HideLabel     bool          `json:"hideLabel,omitempty"`     // Visually hide label (keep for a11y)
	
	// Schema-aligned label properties
	LabelAlign    string        `json:"labelAlign,omitempty"`    // Label alignment (left, right, center)
	LabelWidth    interface{}   `json:"labelWidth,omitempty"`    // Label width (number or string)
	LabelRemark   interface{}   `json:"labelRemark,omitempty"`   // Label remark/tooltip
}

// FormLayoutProps contains form layout configuration
// Aligned with schema form layout properties
type FormLayoutProps struct {
	Mode         DisplayMode     `json:"mode,omitempty"`         // Form layout mode
	Inline       bool            `json:"inline,omitempty"`       // Inline form control
	Horizontal   interface{}     `json:"horizontal,omitempty"`   // Horizontal layout configuration
	Size         Size            `json:"size,omitempty"`         // Form item size
	ExtraName    string          `json:"extraName,omitempty"`    // Additional field name for range components
	Row          int             `json:"row,omitempty"`          // Row position in grid layouts
}

// FieldProps contains comprehensive field configuration
// Combines validation, interaction, and display properties for form fields
type FieldProps struct {
	BaseProps
	AccessibilityProps
	ValidationProps
	InteractionProps
	LabelProps
	PlaceholderProps
	FormLayoutProps
	AlpinEventHandlers
	
	// Field-specific properties
	Description           string             `json:"description,omitempty"`           // Field description
	Desc                  string             `json:"desc,omitempty"`                  // Alias for description
	DescriptionClassName  string             `json:"descriptionClassName,omitempty"`  // Description CSS classes
	Hint                  string             `json:"hint,omitempty"`                  // Input hint (focus tooltip)
	Remark                interface{}        `json:"remark,omitempty"`                // Field remark/tooltip
	SubmitOnChange        bool               `json:"submitOnChange,omitempty"`        // Submit form on field change
	ClearValueOnHidden    bool               `json:"clearValueOnHidden,omitempty"`    // Clear value when hidden
	InputClassName        string             `json:"inputClassName,omitempty"`        // Input-specific CSS classes
	
	// Default value
	Value                 interface{}        `json:"value,omitempty"`                 // Default field value
}

// HasLabel checks if a label should be rendered.
func (l LabelProps) HasLabel() bool {
	if l.Label == nil {
		return false
	}
	
	switch v := l.Label.(type) {
	case string:
		return strings.TrimSpace(v) != ""
	case bool:
		return v // true means show, false means hide
	default:
		return false
	}
}

// GetLabelText returns the label text as a string.
func (l LabelProps) GetLabelText() string {
	if l.Label == nil {
		return ""
	}
	
	switch v := l.Label.(type) {
	case string:
		return v
	case bool:
		if v {
			return "" // Has label but no text specified
		}
		return ""
	default:
		return ""
	}
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
	BaseProps
	AccessibilityProps
	AlpinEventHandlers

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
	AlpinEventHandlers

	// Content
	Text         string       `json:"text,omitempty"`         // Button text
	Icon         IconProps    `json:"icon,omitempty"`         // Icon configuration
	IconName     string       `json:"iconName,omitempty"`     // Simple icon name (for backward compatibility)
	IconOnly     bool         `json:"iconOnly,omitempty"`     // Icon-only button
	IconPosition IconPosition `json:"iconPosition,omitempty"` // Icon position (left, right, top, bottom)

	// Styling
	Variant     Variant     `json:"variant,omitempty"`     // Visual variant
	Size        Size        `json:"size,omitempty"`        // Button size
	ColorScheme ColorScheme `json:"colorScheme,omitempty"` // Color theme
	Color       string      `json:"color,omitempty"`       // Color (simple string for backward compatibility)
	FullWidth   bool        `json:"fullWidth,omitempty"`   // Full width button
	Pill        bool        `json:"pill,omitempty"`        // Pill-shaped (rounded-full) button

	// State
	Loading bool `json:"loading,omitempty"` // Loading state
	Active  bool `json:"active,omitempty"`  // Active state

	// Button-specific
	Type string `json:"type,omitempty"` // button, submit, reset
	Form string `json:"form,omitempty"` // Associated form ID

	// HTMX attributes
	HxGet     string `json:"hxGet,omitempty"`     // HTMX GET request
	HxPost    string `json:"hxPost,omitempty"`    // HTMX POST request
	HxPut     string `json:"hxPut,omitempty"`     // HTMX PUT request
	HxPatch   string `json:"hxPatch,omitempty"`   // HTMX PATCH request
	HxDelete  string `json:"hxDelete,omitempty"`  // HTMX DELETE request
	HxTarget  string `json:"hxTarget,omitempty"`  // HTMX target selector
	HxSwap    string `json:"hxSwap,omitempty"`    // HTMX swap strategy
	HxTrigger string `json:"hxTrigger,omitempty"` // HTMX trigger event
	HxConfirm string `json:"hxConfirm,omitempty"` // HTMX confirmation message

	// Navigation
	Href   string `json:"href,omitempty"`   // Link URL
	Target string `json:"target,omitempty"` // Link target
	
	// Additional commonly used fields (for template compatibility)
	OnClick    string            `json:"onClick,omitempty"`    // Click handler (also available via embedded AlpinEventHandlers)
	AriaLabel  string            `json:"ariaLabel,omitempty"`  // Accessible label (also available via embedded AccessibilityProps)
	EventProps AlpinEventHandlers `json:"eventProps,omitempty"` // Event handlers (also available via embedded AlpinEventHandlers)
}

// InputProps defines properties for Input components.
type InputProps struct {
	BaseProps
	AccessibilityProps
	ValidationProps
	InteractionProps
	LabelProps
	PlaceholderProps
	AlpinEventHandlers

	// Input-specific
	Type         InputType `json:"type,omitempty"`         // Input type
	Value        string    `json:"value,omitempty"`        // Input value
	Variant      Variant   `json:"variant,omitempty"`      // Input variant
	Size         Size      `json:"size,omitempty"`         // Input size
	AutoComplete string    `json:"autoComplete,omitempty"` // Autocomplete attribute
	MaxLength    int       `json:"maxLength,omitempty"`    // Maximum length
	MinLength    int       `json:"minLength,omitempty"`    // Minimum length
	Pattern      string    `json:"pattern,omitempty"`      // Validation pattern
	Min          string    `json:"min,omitempty"`          // Min value (for number/date)
	Max          string    `json:"max,omitempty"`          // Max value (for number/date)
	Step         string    `json:"step,omitempty"`         // Step value (for number)

	// Icons
	LeftIcon  IconProps `json:"leftIcon,omitempty"`  // Icon on left
	RightIcon IconProps `json:"rightIcon,omitempty"` // Icon on right
	
	// Additional commonly used fields (for template compatibility)
	EventProps AlpinEventHandlers `json:"eventProps,omitempty"` // Event handlers (also available via embedded AlpinEventHandlers)
}

// CheckboxProps defines properties for Checkbox components.
type CheckboxProps struct {
	BaseProps
	AccessibilityProps
	ValidationProps
	InteractionProps
	LabelProps
	AlpinEventHandlers

	// Checkbox-specific
	Checked       bool   `json:"checked,omitempty"`       // Checked state
	Indeterminate bool   `json:"indeterminate,omitempty"` // Indeterminate state
	Value         string `json:"value,omitempty"`         // Checkbox value
	Size          Size   `json:"size,omitempty"`          // Checkbox size
	Rounded       bool   `json:"rounded,omitempty"`       // Rounded corners
}

// RadioProps defines properties for Radio button components.
type RadioProps struct {
	BaseProps
	AccessibilityProps
	ValidationProps
	InteractionProps
	LabelProps
	AlpinEventHandlers

	// Radio-specific
	Checked bool   `json:"checked,omitempty"` // Checked state
	Value   string `json:"value,omitempty"`   // Radio value
	Size    Size   `json:"size,omitempty"`    // Radio size
}

// RadioOption represents a single radio option for groups
type RadioOption struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	Disabled bool   `json:"disabled,omitempty"`
	Checked  bool   `json:"checked,omitempty"`
	HelpText string `json:"helpText,omitempty"`
}

// RadioGroupProps defines properties for Radio Group components.
type RadioGroupProps struct {
	BaseProps
	AccessibilityProps
	ValidationProps
	InteractionProps

	// Radio group content
	Label   string        `json:"label,omitempty"`
	Options []RadioOption `json:"options"` // Radio options in the group

	// Radio group attributes
	Name        string `json:"name,omitempty"`        // Group name
	Value       string `json:"value,omitempty"`       // Selected value
	Orientation string `json:"orientation,omitempty"` // horizontal, vertical
	Layout      string `json:"layout,omitempty"`      // Layout style (vertical, horizontal)
	Required    bool   `json:"required,omitempty"`    // Required field

	// Radio group styling
	Size          Size          `json:"size"`              // Radio size for all options
	LabelPosition LabelPosition `json:"labelPosition"`     // Label position for all options
	Spacing       string        `json:"spacing,omitempty"` // Spacing between options
}

// GetLabelText returns the label text for the radio group.
func (r RadioGroupProps) GetLabelText() string {
	return r.Label
}

// SelectProps defines properties for Select components.
type SelectProps struct {
	BaseProps
	AccessibilityProps
	ValidationProps
	InteractionProps
	LabelProps
	PlaceholderProps
	AlpinEventHandlers

	// Select-specific
	Options        []SelectOption `json:"options,omitempty"`        // Select options
	Value          string         `json:"value,omitempty"`          // Selected value
	Multiple       bool           `json:"multiple,omitempty"`       // Allow multiple selection
	Size           Size           `json:"size,omitempty"`           // Select styling size
	VisibleOptions int            `json:"visibleOptions,omitempty"` // Number of visible options (HTML size attr)
}

// SelectOption represents an option in a select dropdown.
type SelectOption struct {
	Value    string         `json:"value"`              // Option value
	Label    string         `json:"label"`              // Option label
	Disabled bool           `json:"disabled,omitempty"` // Option disabled state
	Selected bool           `json:"selected,omitempty"` // Option selected state
	Group    string         `json:"group,omitempty"`    // Option group
	Children []SelectOption `json:"children,omitempty"` // Nested options for optgroups
}

// TextareaProps defines properties for Textarea components.
type TextareaProps struct {
	BaseProps
	AccessibilityProps
	ValidationProps
	InteractionProps
	LabelProps
	PlaceholderProps
	AlpinEventHandlers

	// Textarea-specific
	Value         string `json:"value,omitempty"`         // Textarea value
	Rows          int    `json:"rows,omitempty"`          // Number of rows
	Cols          int    `json:"cols,omitempty"`          // Number of columns
	MinLength     int    `json:"minLength,omitempty"`     // Minimum length
	MaxLength     int    `json:"maxLength,omitempty"`     // Maximum length
	ComponentSize Size   `json:"componentSize,omitempty"` // Textarea size
	Resizable     bool   `json:"resizable,omitempty"`     // Allow resizing
	Wrap          string `json:"wrap,omitempty"`          // Text wrapping mode
}

// ToggleProps defines properties for Toggle/Switch components.
type ToggleProps struct {
	BaseProps
	AccessibilityProps
	InteractionProps
	LabelProps
	ValidationProps
	AlpinEventHandlers

	// Toggle-specific
	Checked     bool        `json:"checked,omitempty"`     // Checked state
	Size        Size        `json:"size,omitempty"`        // Toggle size
	ColorScheme ColorScheme `json:"colorScheme,omitempty"` // Color theme
}

// SpinnerProps defines properties for Spinner/Loader components.
type SpinnerProps struct {
	BaseProps
	AccessibilityProps
	AlpinEventHandlers

	// Spinner-specific
	Size    Size        `json:"size,omitempty"`    // Spinner size
	Color   ColorScheme `json:"color,omitempty"`   // Spinner color
	Variant Variant     `json:"variant,omitempty"` // Spinner variant (border, dots, pulse)
	Speed   string      `json:"speed,omitempty"`   // Animation speed (slow, normal, fast)
	Label   string      `json:"label,omitempty"`   // Screen reader label
}

// ActionProps defines properties for Action/Command components.
type ActionProps struct {
	BaseProps
	AccessibilityProps
	InteractionProps

	// Action content
	Label   string `json:"label,omitempty"`
	Command string `json:"command,omitempty"` // JavaScript command to execute
	URL     string `json:"url,omitempty"`     // For navigation actions

	// Action behavior
	ActionType     string `json:"actionType,omitempty"`     // click, submit, navigate, etc.
	ExecutionMode  string `json:"executionMode,omitempty"`  // immediate, deferred, confirm
	Element        string `json:"element,omitempty"`        // button, a, div, span
	Type           string `json:"type,omitempty"`           // button type attribute
	Target         string `json:"target,omitempty"`         // link target
	Delay          int    `json:"delay,omitempty"`          // Delay in ms for deferred execution
	ThrottleMs     int    `json:"throttleMs,omitempty"`     // Throttle execution
	DebounceMs     int    `json:"debounceMs,omitempty"`     // Debounce execution
	ConfirmMessage string `json:"confirmMessage,omitempty"` // Confirmation dialog message
	LoadingText    string `json:"loadingText,omitempty"`    // Text during loading state
	ApiEndpoint    string `json:"apiEndpoint,omitempty"`    // For API actions
	ApiMethod      string `json:"apiMethod,omitempty"`      // HTTP method for API actions

	// Action styling
	Variant Variant   `json:"variant,omitempty"` // Action variant
	Size    Size      `json:"size,omitempty"`    // Action size
	Icon    IconProps `json:"icon,omitempty"`    // Optional icon

	// Action state
	Loading  bool `json:"loading,omitempty"`  // Loading state
	Disabled bool `json:"disabled,omitempty"` // Disabled state

	// Children for complex actions
	Children []templ.Component `json:"-"` // Child components
}

// BadgeProps defines properties for Badge components.
type BadgeProps struct {
	BaseProps
	AccessibilityProps
	AlpinEventHandlers

	// Badge content
	Text  string `json:"text,omitempty"`
	Value string `json:"value,omitempty"` // Numeric value for count badges

	// Badge styling
	Variant       Variant     `json:"variant,omitempty"`       // Badge variant
	Size          Size        `json:"size,omitempty"`          // Badge size
	ComponentSize Size        `json:"componentSize,omitempty"` // Component size (alias for Size)
	Color         ColorScheme `json:"color,omitempty"`         // Badge color scheme
	Rounded       bool        `json:"rounded,omitempty"`       // Rounded corners

	// Badge behavior
	Dismissible bool   `json:"dismissible,omitempty"` // Can be dismissed
	MaxValue    int    `json:"maxValue,omitempty"`    // Max value for count display (shows 99+)
	Icon        string `json:"icon,omitempty"`        // Optional icon name
	Dot         bool   `json:"dot,omitempty"`         // Render as dot indicator
}

// ColorInputProps defines properties for Color Input components.
type ColorInputProps struct {
	BaseProps
	AccessibilityProps
	ValidationProps
	InteractionProps
	AlpinEventHandlers

	// Color input content
	Label       string `json:"label,omitempty"`
	Value       string `json:"value,omitempty"` // Current color value
	Placeholder string `json:"placeholder,omitempty"`
	HelpText    string `json:"helpText,omitempty"`

	// Color input behavior
	Format       string   `json:"format,omitempty"`       // hex, rgb, hsl, named
	ShowPalette  bool     `json:"showPalette,omitempty"`  // Show color palette
	ShowPreview  bool     `json:"showPreview,omitempty"`  // Show color preview
	AllowAlpha   bool     `json:"allowAlpha,omitempty"`   // Allow alpha channel
	CustomColors []string `json:"customColors,omitempty"` // Custom color palette

	// Color input styling
	Variant Variant `json:"variant,omitempty"` // Input variant
	Size    Size    `json:"size,omitempty"`    // Input size
}

// GetLabelText returns the label text for the color input.
func (c ColorInputProps) GetLabelText() string {
	return c.Label
}

// DividerProps defines properties for Divider components.
type DividerProps struct {
	BaseProps
	AccessibilityProps
	AlpinEventHandlers

	// Divider content
	Text string `json:"text,omitempty"` // Optional divider text

	// Divider styling
	Orientation   string `json:"orientation,omitempty"`   // horizontal, vertical
	Variant       string `json:"variant,omitempty"`       // solid, dashed, dotted
	Color         string `json:"color,omitempty"`         // Divider color
	Thickness     string `json:"thickness,omitempty"`     // Divider thickness
	ComponentSize Size   `json:"componentSize,omitempty"` // Component size

	// Divider spacing
	Margin  string `json:"margin,omitempty"`  // Margin around divider
	Padding string `json:"padding,omitempty"` // Padding around text
}

// HiddenProps defines properties for Hidden Input components.
type HiddenProps struct {
	BaseProps
	InteractionProps

	// Hidden input attributes
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
	Form  string `json:"form,omitempty"` // Associated form ID
}

// ImageProps defines properties for Image components.
type ImageProps struct {
	BaseProps
	AccessibilityProps
	AlpinEventHandlers

	// Image attributes
	Src    string `json:"src,omitempty"`
	Srcset string `json:"srcset,omitempty"` // Responsive image source set
	Alt    string `json:"alt,omitempty"`
	Title  string `json:"title,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`

	// Image behavior
	Loading         string `json:"loading,omitempty"`         // Loading attribute (lazy, eager)
	Decoding        string `json:"decoding,omitempty"`        // Decoding attribute (sync, async, auto)
	CrossOrigin     string `json:"crossOrigin,omitempty"`     // Cross-origin attribute (anonymous, use-credentials)
	Lazy            bool   `json:"lazy,omitempty"`            // Lazy loading
	Placeholder     string `json:"placeholder,omitempty"`     // Placeholder image URL
	ShowPlaceholder bool   `json:"showPlaceholder,omitempty"` // Show placeholder while loading
	FallbackSrc     string `json:"fallbackSrc,omitempty"`     // Fallback image URL
	Sizes           string `json:"sizes,omitempty"`           // Responsive image sizes attribute
	OnError         string `json:"onError,omitempty"`         // Error handler (e.js)

	// Image styling
	Variant    Variant `json:"variant,omitempty"`    // Image variant
	Size       Size    `json:"size,omitempty"`       // Image size
	Shape      string  `json:"shape,omitempty"`      // Image shape (circle, rounded, square)
	Rounded    bool    `json:"rounded,omitempty"`    // Rounded corners
	Shadow     bool    `json:"shadow,omitempty"`     // Drop shadow
	Border     bool    `json:"border,omitempty"`     // Border
	Responsive bool    `json:"responsive,omitempty"` // Responsive sizing
	ObjectFit  string  `json:"objectFit,omitempty"`  // CSS object-fit property

	// Image caption
	Caption     string `json:"caption,omitempty"`     // Image caption
	CaptionSize Size   `json:"captionSize,omitempty"` // Caption text size

	// Image wrapper
	WrapperClass string `json:"wrapperClass,omitempty"` // Custom wrapper classes
}

// LinkProps defines properties for Link components.
type LinkProps struct {
	BaseProps
	AccessibilityProps
	InteractionProps
	AlpinEventHandlers

	// Link attributes
	Href   string `json:"href,omitempty"`
	Target string `json:"target,omitempty"` // _blank, _self, etc.
	Rel    string `json:"rel,omitempty"`    // Link relationship

	// Link content
	Text         string `json:"text,omitempty"`
	Icon         string `json:"icon,omitempty"`         // Optional icon name
	IconPosition string `json:"iconPosition,omitempty"` // Icon position (left, right)

	// Link styling
	Variant       Variant `json:"variant,omitempty"`       // Link variant
	Size          Size    `json:"size,omitempty"`          // Link size
	ComponentSize Size    `json:"componentSize,omitempty"` // Component size
	Color         string  `json:"color,omitempty"`         // Link color
	Underline     bool    `json:"underline,omitempty"`     // Show underline
	External      bool    `json:"external,omitempty"`      // External link styling
	Download      string  `json:"download,omitempty"`      // Download attribute
	NoOpener      bool    `json:"noOpener,omitempty"`      // Add rel="noopener"
	NoReferrer    bool    `json:"noReferrer,omitempty"`    // Add rel="noreferrer"
}

// ProgressProps defines properties for Progress components.
type ProgressProps struct {
	BaseProps
	AccessibilityProps
	AlpinEventHandlers

	// Progress attributes
	Value float64 `json:"value,omitempty"` // Current progress value
	Max   float64 `json:"max,omitempty"`   // Maximum progress value
	Min   float64 `json:"min,omitempty"`   // Minimum progress value

	// Progress content
	Label       string `json:"label,omitempty"`       // Progress label
	Description string `json:"description,omitempty"` // Progress description

	// Progress styling
	Variant       Variant       `json:"variant,omitempty"`       // Progress variant
	Size          Size          `json:"size,omitempty"`          // Progress size
	Color         ColorScheme   `json:"color,omitempty"`         // Progress color
	LabelPosition LabelPosition `json:"labelPosition,omitempty"` // Label position
	ShowPercent   bool          `json:"showPercent,omitempty"`   // Show percentage text
	ShowValue     bool          `json:"showValue,omitempty"`     // Show current value
	Animated      bool          `json:"animated,omitempty"`      // Animated progress
	Striped       bool          `json:"striped,omitempty"`       // Striped pattern
}

// StaticProps defines properties for Static Text components.
type StaticProps struct {
	BaseProps
	AccessibilityProps
	InteractionProps

	// Static content
	Text string `json:"text,omitempty"` // Text content
	Html string `json:"html,omitempty"` // HTML content (use with caution)

	// Static styling
	TextStyle          string `json:"textStyle,omitempty"`          // Text style variant
	Size               Size   `json:"size,omitempty"`               // Text size
	Element            string `json:"element,omitempty"`            // HTML element (span, p, h1, etc.)
	Alignment          string `json:"alignment,omitempty"`          // Text alignment
	Truncate           bool   `json:"truncate,omitempty"`           // Truncate text
	LineClamp          int    `json:"lineClamp,omitempty"`          // Line clamp
	PreserveWhitespace bool   `json:"preserveWhitespace,omitempty"` // Preserve whitespace

	// Static behavior
	OnClick string `json:"onClick,omitempty"` // Click handler
	For     string `json:"for,omitempty"`     // For label elements
	Style   string `json:"style,omitempty"`   // Inline styles

	// Children for complex text
	Children []templ.Component `json:"-"` // Child components
}

// StatusProps defines properties for Status Indicator components.
type StatusProps struct {
	BaseProps
	AccessibilityProps
	InteractionProps
	AlpinEventHandlers

	// Status content
	Text        string `json:"text,omitempty"`
	Description string `json:"description,omitempty"`

	// Status value
	Status string          `json:"status,omitempty"` // success, error, warning, info, pending
	Type   ValidationState `json:"type,omitempty"`   // Status type using ValidationState

	// Status styling
	Variant Variant `json:"variant,omitempty"` // Status variant
	Size    Size    `json:"size,omitempty"`    // Status size
	Icon    string  `json:"icon,omitempty"`    // Custom icon name

	// Status behavior
	ShowIcon bool `json:"showIcon,omitempty"` // Show status icon
	Pulsing  bool `json:"pulsing,omitempty"`  // Pulsing animation
}

// TagProps defines properties for Tag components.
type TagProps struct {
	BaseProps
	AccessibilityProps
	InteractionProps
	AlpinEventHandlers

	// Tag content
	Text  string `json:"text,omitempty"`
	Value string `json:"value,omitempty"` // Tag value for forms

	// Tag styling
	Variant Variant     `json:"variant,omitempty"` // Tag variant
	Size    Size        `json:"size,omitempty"`    // Tag size
	Color   ColorScheme `json:"color,omitempty"`   // Tag color scheme

	// Tag behavior
	Removable  bool   `json:"removable,omitempty"`  // Can be removed
	Selectable bool   `json:"selectable,omitempty"` // Can be selected/clicked
	Selected   bool   `json:"selected,omitempty"`   // Selected state
	OnRemove   string `json:"onRemove,omitempty"`   // Remove handler
	Icon       string `json:"icon,omitempty"`       // Optional icon name
}

// UUIDProps defines properties for UUID components.
type UUIDProps struct {
	BaseProps
	AccessibilityProps
	ValidationProps
	InteractionProps
	AlpinEventHandlers

	// UUID content
	Label       string `json:"label,omitempty"`
	Value       string `json:"value,omitempty"` // UUID value
	Placeholder string `json:"placeholder,omitempty"`
	HelpText    string `json:"helpText,omitempty"`

	// UUID behavior
	Editable           bool   `json:"editable,omitempty"`           // Can be edited
	ShowCopyButton     bool   `json:"showCopyButton,omitempty"`     // Show copy button
	ShowGenerateButton bool   `json:"showGenerateButton,omitempty"` // Show generate button
	AutoGenerate       bool   `json:"autoGenerate,omitempty"`       // Auto-generate on create
	DisplayFormat      string `json:"displayFormat,omitempty"`      // full, short, minimal

	// UUID styling
	Variant Variant `json:"variant,omitempty"` // UUID variant
	Size    Size    `json:"size,omitempty"`    // UUID size
}

// GetLabelText returns the label text for the UUID component.
func (u UUIDProps) GetLabelText() string {
	return u.Label
}
