package schema

import (
	"context"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/pkg/condition"
)

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
	Config      *ActionConfig             `json:"config,omitempty"`      // Action configuration
	Confirm     *Confirm                  `json:"confirm,omitempty"`     // Confirmation dialog
	Conditional *Conditional              `json:"conditional,omitempty"` // Show/hide conditions
	Condition   *condition.ConditionGroup `json:"condition,omitempty"`   // Advanced conditions
	Permissions *ActionPermissions        `json:"permissions,omitempty"` // Access control

	// Framework integration
	HTMX   *ActionHTMX   `json:"htmx,omitempty"`   // HTMX configuration
	Alpine *ActionAlpine `json:"alpine,omitempty"` // Alpine.js configuration

	// Theme
	Theme *ActionTheme `json:"theme,omitempty"` // Action-specific theming

	// Internal (not serialized)
	evaluator *condition.Evaluator `json:"-"` // Condition evaluator
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
	SuccessMessage  string            `json:"successMessage,omitempty"`     // Success message to display
	ErrorMessage    string            `json:"errorMessage,omitempty"`       // Error message to display
	RedirectURL     string            `json:"redirectUrl,omitempty"`        // Redirect after success
}

// Confirm defines a confirmation dialog before action executes
type Confirm struct {
	Enabled bool   `json:"enabled"`                                                // Show confirmation
	Title   string `json:"title,omitempty" example:"Confirm"`                      // Dialog title
	Message string `json:"message" example:"Are you sure?"`                        // Confirmation message
	Confirm string `json:"confirm,omitempty" example:"Yes"`                        // Confirm button text
	Cancel  string `json:"cancel,omitempty" example:"No"`                          // Cancel button text
	Variant string `json:"variant,omitempty" validate:"oneof=info warning danger"` // Dialog type
	Icon    string `json:"icon,omitempty"`                                         // Dialog icon
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

// ActionTheme defines theme overrides for this action
type ActionTheme struct {
	Colors       map[string]string `json:"colors,omitempty"`       // Color overrides
	Spacing      map[string]string `json:"spacing,omitempty"`      // Spacing overrides
	BorderRadius string            `json:"borderRadius,omitempty"` // Border radius override
	FontSize     string            `json:"fontSize,omitempty"`     // Font size override
	FontWeight   string            `json:"fontWeight,omitempty"`   // Font weight override
	CustomCSS    string            `json:"customCSS,omitempty"`    // Custom CSS
}

// SetEvaluator sets the condition evaluator for the action
func (a *Action) SetEvaluator(evaluator *condition.Evaluator) {
	a.evaluator = evaluator
}

// GetEvaluator returns the action's condition evaluator
func (a *Action) GetEvaluator() *condition.Evaluator {
	return a.evaluator
}

// Validate checks if action configuration is valid
func (a *Action) Validate(ctx context.Context) error {
	collector := NewErrorCollector()

	if a.ID == "" {
		collector.AddError(NewValidationError("action_id", "action ID is required"))
	}
	if a.Text == "" {
		collector.AddError(NewValidationError("action_text", "action text is required").WithField(a.ID))
	}
	if a.Type == "" {
		a.Type = ActionButton // Default to button
	}

	// Validate link actions have URL
	if a.Type == ActionLink {
		if a.Config == nil || a.Config.URL == "" {
			collector.AddError(NewValidationError(
				"missing_url",
				"link action requires URL in config",
			).WithField(a.ID))
		}
	}

	// Validate custom actions have handler
	if a.Type == ActionCustom {
		if a.Config == nil || a.Config.Handler == "" {
			collector.AddError(NewValidationError(
				"missing_handler",
				"custom action requires handler in config",
			).WithField(a.ID))
		}
	}

	// Validate confirmation
	if a.Confirm != nil && a.Confirm.Enabled {
		if a.Confirm.Message == "" {
			collector.AddError(NewValidationError(
				"missing_confirm_message",
				"confirmation requires message",
			).WithField(a.ID))
		}
	}

	if collector.HasErrors() {
		return collector.Errors()
	}

	return nil
}

// IsVisible checks if action should be displayed given current form data
func (a *Action) IsVisible(ctx context.Context, data map[string]any) (bool, error) {
	if a.Hidden {
		return false, nil
	}

	// Check advanced conditions with evaluator
	if a.Condition != nil && a.evaluator != nil {
		evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
		result, err := a.evaluator.Evaluate(ctx, a.Condition, evalCtx)
		if err != nil {
			return false, WrapError(err, "condition_evaluation_failed", "failed to evaluate action condition")
		}
		return result, nil
	}

	// Check legacy conditional
	if a.Conditional != nil {
		if a.evaluator != nil {
			// Convert legacy conditional to modern condition format
			convertedCondition := convertLegacyConditional(a.Conditional)
			if convertedCondition != nil {
				evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
				result, err := a.evaluator.Evaluate(ctx, convertedCondition, evalCtx)
				if err != nil {
					return false, WrapError(err, "legacy_condition_evaluation_failed", "failed to evaluate legacy action condition")
				}
				return result, nil
			}
		}
		// Fallback: always visible if no evaluator or conversion failed
		return true, nil
	}

	return true, nil
}

// IsEnabled checks if action can be executed
func (a *Action) IsEnabled() bool {
	return !a.Disabled && !a.Loading
}

// GetVariantClass returns semantic variant identifier for UI rendering
func (a *Action) GetVariantClass() string {
	switch a.Variant {
	case "primary", "secondary", "outline", "ghost", "destructive":
		return "action-" + a.Variant
	case "":
		return "action-primary" // Default fallback
	default:
		return "action-primary" // Unknown variants default to primary
	}
}

// GetSizeClass returns semantic size identifier for UI rendering
func (a *Action) GetSizeClass() string {
	switch a.Size {
	case "sm", "md", "lg", "xl":
		return "action-" + a.Size
	case "":
		return "action-md" // Default fallback
	default:
		return "action-md" // Unknown sizes default to medium
	}
}

// GetIconPosition returns the icon position with fallback
func (a *Action) GetIconPosition() string {
	if a.Position != "" {
		return a.Position
	}
	return "left" // Default to left
}

// HasIcon checks if action has an icon
func (a *Action) HasIcon() bool {
	return a.Icon != ""
}

// HasConfirmation checks if action requires confirmation
func (a *Action) HasConfirmation() bool {
	return a.Confirm != nil && a.Confirm.Enabled
}

// RequiresPermission checks if action requires specific permission
func (a *Action) RequiresPermission(permission string) bool {
	if a.Permissions == nil {
		return false
	}
	return slices.Contains(a.Permissions.Required, permission)
}

// CanView checks if user with given roles can view the action
func (a *Action) CanView(userRoles []string) bool {
	if a.Permissions == nil || len(a.Permissions.View) == 0 {
		return true // No restrictions
	}

	for _, role := range userRoles {
		if slices.Contains(a.Permissions.View, role) {
			return true
		}
	}
	return false
}

// CanExecute checks if user with given roles can execute the action
func (a *Action) CanExecute(userRoles []string) bool {
	if a.Permissions == nil || len(a.Permissions.Execute) == 0 {
		return true // No restrictions
	}

	for _, role := range userRoles {
		if slices.Contains(a.Permissions.Execute, role) {
			return true
		}
	}
	return false
}

// Clone creates a copy of the action with evaluator preservation
func (a *Action) Clone() *Action {
	clone := *a
	clone.evaluator = a.evaluator
	return &clone
}

// ApplyTheme applies theme settings to the action
func (a *Action) ApplyTheme(theme *Theme) {
	if theme == nil {
		return
	}

	// Apply theme overrides
	if a.Theme == nil {
		a.Theme = &ActionTheme{}
	}

	// NOTE:Theme will be applied by renderer
}

// GetConfig returns action config with fallback
func (a *Action) GetConfig() *ActionConfig {
	if a.Config == nil {
		a.Config = &ActionConfig{}
	}
	return a.Config
}

// SetLoading sets the loading state
func (a *Action) SetLoading(loading bool) {
	a.Loading = loading
}

// SetDisabled sets the disabled state
func (a *Action) SetDisabled(disabled bool) {
	a.Disabled = disabled
}

// SetHidden sets the hidden state
func (a *Action) SetHidden(hidden bool) {
	a.Hidden = hidden
}

// GetHTMXConfig returns HTMX configuration with fallback
func (a *Action) GetHTMXConfig() *ActionHTMX {
	if a.HTMX == nil {
		a.HTMX = &ActionHTMX{}
	}
	return a.HTMX
}

// GetAlpineConfig returns Alpine configuration with fallback
func (a *Action) GetAlpineConfig() *ActionAlpine {
	if a.Alpine == nil {
		a.Alpine = &ActionAlpine{}
	}
	return a.Alpine
}

// IsSubmitAction checks if this is a submit action
func (a *Action) IsSubmitAction() bool {
	return a.Type == ActionSubmit
}

// IsResetAction checks if this is a reset action
func (a *Action) IsResetAction() bool {
	return a.Type == ActionReset
}

// IsLinkAction checks if this is a link action
func (a *Action) IsLinkAction() bool {
	return a.Type == ActionLink
}

// IsCustomAction checks if this is a custom action
func (a *Action) IsCustomAction() bool {
	return a.Type == ActionCustom
}

// GetURL returns the action URL (for link/HTMX actions)
func (a *Action) GetURL() string {
	if a.Type == ActionLink && a.Config != nil {
		return a.Config.URL
	}
	if a.HTMX != nil {
		return a.HTMX.URL
	}
	return ""
}

// GetHTTPMethod returns the HTTP method for HTMX actions
func (a *Action) GetHTTPMethod() string {
	if a.HTMX != nil && a.HTMX.Method != "" {
		return a.HTMX.Method
	}
	if a.Type == ActionSubmit {
		return "POST"
	}
	return "GET"
}

// ShouldDebounce checks if action should be debounced
func (a *Action) ShouldDebounce() bool {
	return a.Config != nil && a.Config.Debounce > 0
}

// GetDebounceDelay returns debounce delay in milliseconds
func (a *Action) GetDebounceDelay() int {
	if a.Config != nil {
		return a.Config.Debounce
	}
	return 0
}

// ShouldThrottle checks if action should be throttled
func (a *Action) ShouldThrottle() bool {
	return a.Config != nil && a.Config.Throttle > 0
}

// GetThrottleDelay returns throttle delay in milliseconds
func (a *Action) GetThrottleDelay() int {
	if a.Config != nil {
		return a.Config.Throttle
	}
	return 0
}

// ActionBuilder provides a fluent interface for building actions
type ActionBuilder struct {
	action    *Action
	evaluator *condition.Evaluator
}

// NewAction starts building an action
func NewAction(id string, actionType ActionType, text string) *ActionBuilder {
	return &ActionBuilder{
		action: &Action{
			ID:   id,
			Type: actionType,
			Text: text,
		},
	}
}

// WithVariant sets the variant
func (ab *ActionBuilder) WithVariant(variant string) *ActionBuilder {
	ab.action.Variant = variant
	return ab
}

// WithSize sets the size
func (ab *ActionBuilder) WithSize(size string) *ActionBuilder {
	ab.action.Size = size
	return ab
}

// WithIcon sets the icon
func (ab *ActionBuilder) WithIcon(icon string, position string) *ActionBuilder {
	ab.action.Icon = icon
	ab.action.Position = position
	return ab
}

// WithConfig sets the configuration
func (ab *ActionBuilder) WithConfig(config *ActionConfig) *ActionBuilder {
	ab.action.Config = config
	return ab
}

// WithConfirmation adds confirmation dialog
func (ab *ActionBuilder) WithConfirmation(message string, title string) *ActionBuilder {
	ab.action.Confirm = &Confirm{
		Enabled: true,
		Title:   title,
		Message: message,
	}
	return ab
}

// WithPermissions sets permissions
func (ab *ActionBuilder) WithPermissions(permissions *ActionPermissions) *ActionBuilder {
	ab.action.Permissions = permissions
	return ab
}

// WithCondition sets the condition
func (ab *ActionBuilder) WithCondition(condition *condition.ConditionGroup) *ActionBuilder {
	ab.action.Condition = condition
	return ab
}

// WithEvaluator sets the evaluator
func (ab *ActionBuilder) WithEvaluator(evaluator *condition.Evaluator) *ActionBuilder {
	ab.evaluator = evaluator
	return ab
}

// Disabled marks action as disabled
func (ab *ActionBuilder) Disabled() *ActionBuilder {
	ab.action.Disabled = true
	return ab
}

// Hidden marks action as hidden
func (ab *ActionBuilder) Hidden() *ActionBuilder {
	ab.action.Hidden = true
	return ab
}

// Build returns the constructed action
func (ab *ActionBuilder) Build() *Action {
	if ab.evaluator != nil {
		ab.action.SetEvaluator(ab.evaluator)
	}
	return ab.action
}

// Common action builders

// NewSubmitAction creates a submit action
func NewSubmitAction(id, text string) *ActionBuilder {
	return NewAction(id, ActionSubmit, text).WithVariant("primary")
}

// NewResetAction creates a reset action
func NewResetAction(id, text string) *ActionBuilder {
	return NewAction(id, ActionReset, text).WithVariant("secondary")
}

// NewCancelAction creates a cancel action
func NewCancelAction(id, text string) *ActionBuilder {
	return NewAction(id, ActionButton, text).WithVariant("outline")
}

// NewDeleteAction creates a delete action with confirmation
func NewDeleteAction(id, text string) *ActionBuilder {
	return NewAction(id, ActionButton, text).
		WithVariant("destructive").
		WithConfirmation("Are you sure you want to delete this item?", "Confirm Deletion")
}


// convertLegacyConditional converts legacy Conditional to condition.ConditionGroup
// This is for backward compatibility with the old conditional format
func convertLegacyConditional(cond *Conditional) *condition.ConditionGroup {
	if cond == nil {
		return nil
	}

	// For actions, we typically check the "show" condition
	if cond.Show != nil {
		return convertConditionGroup(cond.Show)
	}
	
	// If there's a hide condition, invert it for visibility
	if cond.Hide != nil {
		hideGroup := convertConditionGroup(cond.Hide)
		if hideGroup != nil {
			// Wrap in NOT condition
			return &condition.ConditionGroup{
				ID:          "hide_condition_not",
				Conjunction: condition.ConjunctionAnd,
				Not:         true,
				Children:    []any{hideGroup},
			}
		}
	}

	return nil
}

// convertConditionGroup converts schema.ConditionGroup to condition.ConditionGroup
func convertConditionGroup(group *ConditionGroup) *condition.ConditionGroup {
	if group == nil || len(group.Conditions) == 0 {
		return nil
	}

	// Map logic to conjunction
	conjunction := condition.ConjunctionAnd
	if strings.ToLower(group.Logic) == "or" {
		conjunction = condition.ConjunctionOr
	}

	result := &condition.ConditionGroup{
		ID:          "converted_group_" + uuid.New().String()[:8],
		Conjunction: conjunction,
	}

	// Convert each condition to ConditionRule
	for _, cond := range group.Conditions {
		conditionRule := &condition.ConditionRule{
			ID: "converted_rule_" + uuid.New().String()[:8],
			Left: condition.Expression{
				Type:  condition.ValueTypeField,
				Field: cond.Field,
			},
			Op:    condition.OperatorType(convertOperator(cond.Operator)),
			Right: cond.Value,
		}
		result.Children = append(result.Children, conditionRule)
	}

	return result
}

// convertOperator converts legacy operator names to condition package operators
func convertOperator(op string) string {
	// Map common operator names
	switch op {
	case "equal", "equals", "==":
		return "equal"
	case "not_equal", "not_equals", "!=":
		return "not_equal"
	case "greater", "greater_than", ">":
		return "greater"
	case "greater_equal", "greater_or_equal", ">=":
		return "greater_or_equal"
	case "less", "less_than", "<":
		return "less"
	case "less_equal", "less_or_equal", "<=":
		return "less_or_equal"
	case "contains":
		return "contains"
	case "starts_with":
		return "starts_with"
	case "ends_with":
		return "ends_with"
	case "in":
		return "select_any_in"
	case "not_in":
		return "select_not_any_in"
	default:
		// Return as-is if not mapped
		return op
	}
}
