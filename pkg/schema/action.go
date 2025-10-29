package schema

import "context"

// Action represents a button or actionable element in the form
type Action struct {
	ID   string     `json:"id" validate:"required" example:"submit-btn"`
	Type ActionType `json:"type" validate:"required" example:"submit"`
	Text string     `json:"text" validate:"required,min=1,max=100" example:"Create User"`

	// Appearance
	Variant  string `json:"variant,omitempty" validate:"oneof=primary secondary outline ghost destructive" example:"primary"`
	Size     string `json:"size,omitempty" validate:"oneof=sm md lg xl" example:"md"`
	Icon     string `json:"icon,omitempty" validate:"icon_name" example:"user-plus"`
	Position string `json:"position,omitempty" validate:"oneof=left right center" example:"left"` // Icon position

	// State
	Loading  bool `json:"loading,omitempty"`  // Show loading spinner
	Disabled bool `json:"disabled,omitempty"` // Cannot be clicked
	Hidden   bool `json:"hidden,omitempty"`   // Not displayed

	// Behavior
	Config      *ActionConfig      `json:"config,omitempty"`      // Action configuration
	Confirm     *Confirm           `json:"confirm,omitempty"`     // Confirmation dialog
	Conditional *Conditional       `json:"conditional,omitempty"` // Show/hide conditions
	Permissions *ActionPermissions `json:"permissions,omitempty"` // Access control

	// Framework integration
	HTMX   *ActionHTMX   `json:"htmx,omitempty"`   // HTMX configuration
	Alpine *ActionAlpine `json:"alpine,omitempty"` // Alpine.js configuration
}

// ActionType defines the type of action
type ActionType string

const (
	ActionSubmit ActionType = "submit" // Submit form
	ActionReset  ActionType = "reset"  // Reset form to defaults
	ActionButton ActionType = "button" // Generic button (no default behavior)
	ActionLink   ActionType = "link"   // Navigate to URL
	ActionCustom ActionType = "custom" // Custom handler
)

// ActionConfig holds action-specific configuration
type ActionConfig struct {
	URL             string            `json:"url,omitempty" validate:"url"` // For link actions
	Target          string            `json:"target,omitempty"`             // Link target (_blank, _self)
	Handler         string            `json:"handler,omitempty"`            // Custom JS handler function
	Params          map[string]string `json:"params,omitempty"`             // Additional parameters
	Debounce        int               `json:"debounce,omitempty"`           // Debounce delay in ms
	Throttle        int               `json:"throttle,omitempty"`           // Throttle delay in ms
	PreventDefault  bool              `json:"preventDefault,omitempty"`     // Prevent default action
	StopPropagation bool              `json:"stopPropagation,omitempty"`    // Stop event propagation
}

// Confirm defines a confirmation dialog before action executes
type Confirm struct {
	Enabled bool   `json:"enabled"`                                                // Show confirmation
	Title   string `json:"title,omitempty" example:"Confirm"`                      // Dialog title
	Message string `json:"message" example:"Are you sure?"`                        // Confirmation message
	Confirm string `json:"confirm,omitempty" example:"Yes"`                        // Confirm button text
	Cancel  string `json:"cancel,omitempty" example:"No"`                          // Cancel button text
	Variant string `json:"variant,omitempty" validate:"oneof=info warning danger"` // Dialog type
}

// ActionPermissions controls who can see/use the action
type ActionPermissions struct {
	View     []string `json:"view,omitempty"`     // Roles that can view action
	Execute  []string `json:"execute,omitempty"`  // Roles that can execute action
	Required []string `json:"required,omitempty"` // Required permissions
}

// ActionHTMX defines HTMX behavior for the action
type ActionHTMX struct {
	Method    string            `json:"method,omitempty" validate:"oneof=GET POST PUT PATCH DELETE"`
	URL       string            `json:"url,omitempty" validate:"url"`
	Target    string            `json:"target,omitempty" validate:"css_selector"` // Where to put response
	Swap      string            `json:"swap,omitempty" validate:"oneof=innerHTML outerHTML beforebegin afterbegin beforeend afterend delete none"`
	Trigger   string            `json:"trigger,omitempty"`   // HTMX trigger specification
	Indicator string            `json:"indicator,omitempty"` // Loading indicator selector
	Confirm   string            `json:"confirm,omitempty"`   // Confirmation prompt
	Headers   map[string]string `json:"headers,omitempty"`   // Request headers
	Vals      string            `json:"vals,omitempty"`      // Additional values to include
	Include   string            `json:"include,omitempty"`   // Elements to include
	PushURL   string            `json:"pushUrl,omitempty"`   // URL to push to history
	Select    string            `json:"select,omitempty"`    // CSS selector for response content
	Sync      string            `json:"sync,omitempty"`      // Sync specification
}

// ActionAlpine defines Alpine.js bindings for the action
type ActionAlpine struct {
	XOn   string `json:"xOn,omitempty" validate:"js_object"`       // Event handlers
	XBind string `json:"xBind,omitempty" validate:"js_object"`     // Attribute bindings
	XShow string `json:"xShow,omitempty" validate:"js_expression"` // Show/hide condition
	XIf   string `json:"xIf,omitempty" validate:"js_expression"`   // Conditional render
	XText string `json:"xText,omitempty" validate:"js_expression"` // Text content
}

// Validate checks if action configuration is valid
func (a *Action) Validate(ctx context.Context) error {
	if a.ID == "" {
		return NewValidationError("action_id", "action ID is required")
	}
	if a.Text == "" {
		return NewValidationError("action_text", "action text is required").WithField(a.ID)
	}
	if a.Type == "" {
		a.Type = ActionButton // Default to button
	}

	// Validate link actions have URL
	if a.Type == ActionLink {
		if a.Config == nil || a.Config.URL == "" {
			return NewValidationError(
				"missing_url",
				"link action requires URL in config",
			).WithField(a.ID)
		}
	}

	// Validate custom actions have handler
	if a.Type == ActionCustom {
		if a.Config == nil || a.Config.Handler == "" {
			return NewValidationError(
				"missing_handler",
				"custom action requires handler in config",
			).WithField(a.ID)
		}
	}

	return nil
}

// IsVisible checks if action should be displayed given current form data
func (a *Action) IsVisible(data map[string]any) bool {
	if a.Hidden {
		return false
	}
	if a.Conditional != nil {
		// TODO: Integrate with condition evaluator
		return true
	}
	return true
}

// IsEnabled checks if action can be executed
func (a *Action) IsEnabled() bool {
	return !a.Disabled && !a.Loading
}

// GetVariantClass returns Flowbite CSS classes for the variant
func (a *Action) GetVariantClass() string {
	switch a.Variant {
	case "primary":
		return "text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
	case "secondary":
		return "text-gray-900 bg-white border border-gray-300 hover:bg-gray-100 focus:ring-4 focus:ring-gray-200 dark:bg-gray-800 dark:text-white dark:border-gray-600 dark:hover:bg-gray-700 dark:focus:ring-gray-700"
	case "outline":
		return "text-blue-700 border border-blue-700 hover:bg-blue-700 hover:text-white focus:ring-4 focus:ring-blue-300 dark:border-blue-500 dark:text-blue-500 dark:hover:bg-blue-500 dark:hover:text-white dark:focus:ring-blue-800"
	case "ghost":
		return "text-gray-500 hover:text-gray-900 hover:bg-gray-100 focus:ring-4 focus:ring-gray-200 dark:text-gray-400 dark:hover:text-white dark:hover:bg-gray-800 dark:focus:ring-gray-700"
	case "destructive":
		return "text-white bg-red-700 hover:bg-red-800 focus:ring-4 focus:ring-red-300 dark:bg-red-600 dark:hover:bg-red-700 dark:focus:ring-red-900"
	default:
		return "text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300"
	}
}

// GetSizeClass returns Flowbite CSS classes for the size
func (a *Action) GetSizeClass() string {
	switch a.Size {
	case "sm":
		return "px-3 py-2 text-sm"
	case "md":
		return "px-5 py-2.5 text-sm"
	case "lg":
		return "px-5 py-3 text-base"
	case "xl":
		return "px-6 py-3.5 text-base"
	default:
		return "px-5 py-2.5 text-sm"
	}
}
