package editors

import (
	"fmt"

	"github.com/niiniyare/erp/web/builder/core"
)

// EventEditor handles event handlers and actions configuration
type EventEditor struct {
	component *core.ComponentInstance
}

// EventConfiguration represents event handling settings
type EventConfiguration struct {
	Handlers   []EventHandler   `json:"handlers"`
	Actions    []EventAction    `json:"actions"`
	Conditions []EventCondition `json:"conditions,omitempty"`
	Validators []EventValidator `json:"validators,omitempty"`
}

// EventHandler represents an event listener configuration
type EventHandler struct {
	ID         string         `json:"id"`
	Event      string         `json:"event"`                // "click", "change", "submit", etc.
	Target     string         `json:"target"`               // "self", "parent", "child", selector
	Actions    []string       `json:"actions"`              // Action IDs to execute
	Conditions []string       `json:"conditions,omitempty"` // Condition IDs to check
	Debounce   int            `json:"debounce,omitempty"`   // Debounce delay in ms
	Throttle   int            `json:"throttle,omitempty"`   // Throttle delay in ms
	Once       bool           `json:"once,omitempty"`       // Execute only once
	Passive    bool           `json:"passive,omitempty"`    // Passive event listener
	Capture    bool           `json:"capture,omitempty"`    // Capture phase
	Enabled    bool           `json:"enabled"`
	Options    map[string]any `json:"options,omitempty"`
}

// EventAction represents an action to be executed
type EventAction struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"` // "navigate", "api", "update", "custom"
	Target      string         `json:"target,omitempty"`
	Config      map[string]any `json:"config"`
	Async       bool           `json:"async,omitempty"`
	RetryCount  int            `json:"retryCount,omitempty"`
	Timeout     int            `json:"timeout,omitempty"`
	OnSuccess   []string       `json:"onSuccess,omitempty"`  // Action IDs on success
	OnError     []string       `json:"onError,omitempty"`    // Action IDs on error
	OnComplete  []string       `json:"onComplete,omitempty"` // Action IDs on complete
	Enabled     bool           `json:"enabled"`
	Description string         `json:"description,omitempty"`
}

// EventCondition represents a condition for event execution
type EventCondition struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"` // "expression", "permission", "state"
	Expression  string         `json:"expression,omitempty"`
	Config      map[string]any `json:"config"`
	Invert      bool           `json:"invert,omitempty"`
	Description string         `json:"description,omitempty"`
}

// EventValidator represents input validation rules
type EventValidator struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`  // "required", "pattern", "range", "custom"
	Field    string         `json:"field"` // Field to validate
	Config   map[string]any `json:"config"`
	Message  string         `json:"message"`  // Error message
	Severity string         `json:"severity"` // "error", "warning", "info"
}

// NewEventEditor creates a new event editor instance
func NewEventEditor(component *core.ComponentInstance) *EventEditor {
	return &EventEditor{
		component: component,
	}
}

// GetEventConfiguration returns the current event configuration
func (ee *EventEditor) GetEventConfiguration() *EventConfiguration {
	if ee.component.Events == nil {
		return &EventConfiguration{
			Handlers:   []EventHandler{},
			Actions:    []EventAction{},
			Conditions: []EventCondition{},
			Validators: []EventValidator{},
		}
	}

	config := &EventConfiguration{}
	config.Handlers = ee.extractEventHandlers()
	config.Actions = ee.extractEventActions()
	config.Conditions = ee.extractEventConditions()
	config.Validators = ee.extractEventValidators()

	return config
}

// UpdateEventConfiguration applies new event configuration
func (ee *EventEditor) UpdateEventConfiguration(config *EventConfiguration) error {
	if ee.component.Events == nil {
		ee.component.Events = &core.EventHandlers{}
	}

	// Apply event handlers
	if err := ee.applyEventHandlers(config.Handlers); err != nil {
		return fmt.Errorf("failed to apply event handlers: %w", err)
	}

	// Apply actions
	if err := ee.applyEventActions(config.Actions); err != nil {
		return fmt.Errorf("failed to apply event actions: %w", err)
	}

	// Apply conditions
	if err := ee.applyEventConditions(config.Conditions); err != nil {
		return fmt.Errorf("failed to apply event conditions: %w", err)
	}

	// Apply validators
	if err := ee.applyEventValidators(config.Validators); err != nil {
		return fmt.Errorf("failed to apply event validators: %w", err)
	}

	return nil
}

// GetEventPresets returns common event configurations
func (ee *EventEditor) GetEventPresets() []EventPreset {
	return []EventPreset{
		{
			ID:          "button-click",
			Name:        "Button Click",
			Description: "Basic button click handler",
			Events:      []string{"click"},
			Handler: EventHandler{
				Event:   "click",
				Target:  "self",
				Actions: []string{"navigate-action"},
				Enabled: true,
			},
			Actions: []EventAction{
				{
					ID:   "navigate-action",
					Type: "navigate",
					Config: map[string]any{
						"url": "/page",
					},
					Enabled: true,
				},
			},
		},
		{
			ID:          "form-submit",
			Name:        "Form Submit",
			Description: "Form submission with validation",
			Events:      []string{"submit"},
			Handler: EventHandler{
				Event:      "submit",
				Target:     "self",
				Actions:    []string{"validate-action", "submit-action"},
				Conditions: []string{"form-valid"},
				Enabled:    true,
			},
			Actions: []EventAction{
				{
					ID:   "validate-action",
					Type: "validate",
					Config: map[string]any{
						"rules": []string{"required", "email"},
					},
					Enabled: true,
				},
				{
					ID:   "submit-action",
					Type: "api",
					Config: map[string]any{
						"url":    "/api/submit",
						"method": "POST",
					},
					OnSuccess: []string{"success-message"},
					OnError:   []string{"error-message"},
					Enabled:   true,
				},
			},
			Conditions: []EventCondition{
				{
					ID:         "form-valid",
					Type:       "expression",
					Expression: "form.isValid",
				},
			},
		},
		{
			ID:          "input-change",
			Name:        "Input Change",
			Description: "Input field change handler with debouncing",
			Events:      []string{"input", "change"},
			Handler: EventHandler{
				Event:    "input",
				Target:   "self",
				Actions:  []string{"update-state"},
				Debounce: 300,
				Enabled:  true,
			},
			Actions: []EventAction{
				{
					ID:   "update-state",
					Type: "update",
					Config: map[string]any{
						"property": "value",
						"source":   "event.target.value",
					},
					Enabled: true,
				},
			},
		},
		{
			ID:          "api-call",
			Name:        "API Call",
			Description: "Make API request with loading states",
			Events:      []string{"click"},
			Handler: EventHandler{
				Event:   "click",
				Target:  "self",
				Actions: []string{"loading-start", "api-request", "loading-end"},
				Enabled: true,
			},
			Actions: []EventAction{
				{
					ID:   "loading-start",
					Type: "update",
					Config: map[string]any{
						"property": "loading",
						"value":    true,
					},
					Enabled: true,
				},
				{
					ID:   "api-request",
					Type: "api",
					Config: map[string]any{
						"url":    "/api/data",
						"method": "GET",
					},
					OnSuccess: []string{"update-data"},
					OnError:   []string{"show-error"},
					Async:     true,
					Timeout:   5000,
					Enabled:   true,
				},
				{
					ID:   "loading-end",
					Type: "update",
					Config: map[string]any{
						"property": "loading",
						"value":    false,
					},
					Enabled: true,
				},
			},
		},
	}
}

// EventPreset represents a predefined event configuration
type EventPreset struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Events      []string         `json:"events"`
	Handler     EventHandler     `json:"handler"`
	Actions     []EventAction    `json:"actions"`
	Conditions  []EventCondition `json:"conditions,omitempty"`
	Validators  []EventValidator `json:"validators,omitempty"`
	Icon        string           `json:"icon,omitempty"`
}

// GetAvailableEvents returns list of available events for the component type
func (ee *EventEditor) GetAvailableEvents() []EventDefinition {
	componentType := ee.component.Type

	// Common events for all components
	commonEvents := []EventDefinition{
		{
			Name:        "click",
			Description: "Triggered when the element is clicked",
			Parameters:  []string{"event", "target"},
			Bubbles:     true,
		},
		{
			Name:        "dblclick",
			Description: "Triggered when the element is double-clicked",
			Parameters:  []string{"event", "target"},
			Bubbles:     true,
		},
		{
			Name:        "mouseenter",
			Description: "Triggered when mouse enters the element",
			Parameters:  []string{"event", "target"},
			Bubbles:     false,
		},
		{
			Name:        "mouseleave",
			Description: "Triggered when mouse leaves the element",
			Parameters:  []string{"event", "target"},
			Bubbles:     false,
		},
	}

	// Component-specific events
	var specificEvents []EventDefinition

	switch componentType {
	case "atoms.input":
		specificEvents = []EventDefinition{
			{
				Name:        "input",
				Description: "Triggered when input value changes",
				Parameters:  []string{"event", "value", "target"},
				Bubbles:     true,
			},
			{
				Name:        "change",
				Description: "Triggered when input loses focus and value changed",
				Parameters:  []string{"event", "value", "target"},
				Bubbles:     true,
			},
			{
				Name:        "focus",
				Description: "Triggered when input gains focus",
				Parameters:  []string{"event", "target"},
				Bubbles:     false,
			},
			{
				Name:        "blur",
				Description: "Triggered when input loses focus",
				Parameters:  []string{"event", "target"},
				Bubbles:     false,
			},
		}
	case "organisms.form":
		specificEvents = []EventDefinition{
			{
				Name:        "submit",
				Description: "Triggered when form is submitted",
				Parameters:  []string{"event", "data", "target"},
				Bubbles:     true,
			},
			{
				Name:        "reset",
				Description: "Triggered when form is reset",
				Parameters:  []string{"event", "target"},
				Bubbles:     true,
			},
			{
				Name:        "validate",
				Description: "Triggered during form validation",
				Parameters:  []string{"event", "errors", "target"},
				Bubbles:     false,
			},
		}
	case "organisms.table":
		specificEvents = []EventDefinition{
			{
				Name:        "rowclick",
				Description: "Triggered when table row is clicked",
				Parameters:  []string{"event", "row", "index", "target"},
				Bubbles:     true,
			},
			{
				Name:        "sort",
				Description: "Triggered when column is sorted",
				Parameters:  []string{"event", "column", "direction", "target"},
				Bubbles:     false,
			},
			{
				Name:        "filter",
				Description: "Triggered when table is filtered",
				Parameters:  []string{"event", "filters", "target"},
				Bubbles:     false,
			},
		}
	}

	return append(commonEvents, specificEvents...)
}

// EventDefinition represents an available event type
type EventDefinition struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Parameters  []string `json:"parameters"`
	Bubbles     bool     `json:"bubbles"`
	Category    string   `json:"category,omitempty"`
}

// GetAvailableActions returns list of available action types
func (ee *EventEditor) GetAvailableActions() []ActionDefinition {
	return []ActionDefinition{
		{
			Type:        "navigate",
			Name:        "Navigate",
			Description: "Navigate to a different page or URL",
			Parameters: []ActionParameter{
				{Name: "url", Type: "string", Required: true, Description: "Target URL"},
				{Name: "target", Type: "string", Required: false, Description: "Window target (_self, _blank)"},
				{Name: "replace", Type: "boolean", Required: false, Description: "Replace current history entry"},
			},
			Category: "Navigation",
		},
		{
			Type:        "api",
			Name:        "API Call",
			Description: "Make an HTTP request to an API endpoint",
			Parameters: []ActionParameter{
				{Name: "url", Type: "string", Required: true, Description: "API endpoint URL"},
				{Name: "method", Type: "string", Required: true, Description: "HTTP method (GET, POST, PUT, DELETE)"},
				{Name: "headers", Type: "object", Required: false, Description: "Request headers"},
				{Name: "body", Type: "object", Required: false, Description: "Request body"},
				{Name: "timeout", Type: "number", Required: false, Description: "Request timeout in milliseconds"},
			},
			Category: "Data",
		},
		{
			Type:        "update",
			Name:        "Update Property",
			Description: "Update a component property or state",
			Parameters: []ActionParameter{
				{Name: "target", Type: "string", Required: false, Description: "Target component selector"},
				{Name: "property", Type: "string", Required: true, Description: "Property name to update"},
				{Name: "value", Type: "any", Required: false, Description: "New property value"},
				{Name: "source", Type: "string", Required: false, Description: "Source expression for value"},
			},
			Category: "State",
		},
		{
			Type:        "validate",
			Name:        "Validate",
			Description: "Validate form fields or data",
			Parameters: []ActionParameter{
				{Name: "target", Type: "string", Required: false, Description: "Target form or field selector"},
				{Name: "rules", Type: "array", Required: true, Description: "Validation rules to apply"},
				{Name: "showErrors", Type: "boolean", Required: false, Description: "Show validation errors"},
			},
			Category: "Validation",
		},
		{
			Type:        "show",
			Name:        "Show Element",
			Description: "Show a hidden element or modal",
			Parameters: []ActionParameter{
				{Name: "target", Type: "string", Required: true, Description: "Target element selector"},
				{Name: "animation", Type: "string", Required: false, Description: "Show animation type"},
				{Name: "duration", Type: "number", Required: false, Description: "Animation duration"},
			},
			Category: "UI",
		},
		{
			Type:        "hide",
			Name:        "Hide Element",
			Description: "Hide a visible element or modal",
			Parameters: []ActionParameter{
				{Name: "target", Type: "string", Required: true, Description: "Target element selector"},
				{Name: "animation", Type: "string", Required: false, Description: "Hide animation type"},
				{Name: "duration", Type: "number", Required: false, Description: "Animation duration"},
			},
			Category: "UI",
		},
		{
			Type:        "alert",
			Name:        "Show Alert",
			Description: "Display an alert message",
			Parameters: []ActionParameter{
				{Name: "message", Type: "string", Required: true, Description: "Alert message"},
				{Name: "type", Type: "string", Required: false, Description: "Alert type (success, error, warning, info)"},
				{Name: "duration", Type: "number", Required: false, Description: "Auto-dismiss duration"},
			},
			Category: "Feedback",
		},
		{
			Type:        "custom",
			Name:        "Custom Function",
			Description: "Execute a custom JavaScript function",
			Parameters: []ActionParameter{
				{Name: "function", Type: "string", Required: true, Description: "Function name to execute"},
				{Name: "parameters", Type: "array", Required: false, Description: "Function parameters"},
				{Name: "context", Type: "object", Required: false, Description: "Execution context"},
			},
			Category: "Custom",
		},
	}
}

// ActionDefinition represents an available action type
type ActionDefinition struct {
	Type        string            `json:"type"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  []ActionParameter `json:"parameters"`
	Category    string            `json:"category"`
	Icon        string            `json:"icon,omitempty"`
}

// ActionParameter represents an action parameter definition
type ActionParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     any    `json:"default,omitempty"`
}

// ValidateEventConfiguration validates the event configuration
func (ee *EventEditor) ValidateEventConfiguration(config *EventConfiguration) []ValidationError {
	var errors []ValidationError

	// Validate event handlers
	for i, handler := range config.Handlers {
		if handler.Event == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("handlers[%d].event", i),
				Message: "Event type is required",
			})
		}
		if len(handler.Actions) == 0 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("handlers[%d].actions", i),
				Message: "At least one action is required",
			})
		}
	}

	// Validate actions
	actionIds := make(map[string]bool)
	for i, action := range config.Actions {
		if action.ID == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("actions[%d].id", i),
				Message: "Action ID is required",
			})
		} else if actionIds[action.ID] {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("actions[%d].id", i),
				Message: "Action ID must be unique",
			})
		}
		actionIds[action.ID] = true

		if action.Type == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("actions[%d].type", i),
				Message: "Action type is required",
			})
		}
	}

	// Validate action references in handlers
	for i, handler := range config.Handlers {
		for j, actionId := range handler.Actions {
			if !actionIds[actionId] {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("handlers[%d].actions[%d]", i, j),
					Message: fmt.Sprintf("Referenced action '%s' does not exist", actionId),
				})
			}
		}
	}

	return errors
}

// Helper methods for extracting configuration

func (ee *EventEditor) extractEventHandlers() []EventHandler {
	var handlers []EventHandler

	if ee.component.Events.Handlers != nil {
		for _, handler := range ee.component.Events.Handlers {
			eventHandler := EventHandler{
				ID:         handler.ID,
				Event:      handler.Event,
				Target:     handler.Target,
				Actions:    handler.Actions,
				Conditions: handler.Conditions,
				Debounce:   handler.Debounce,
				Throttle:   handler.Throttle,
				Once:       handler.Once,
				Passive:    handler.Passive,
				Capture:    handler.Capture,
				Enabled:    handler.Enabled,
				Options:    handler.Options,
			}
			handlers = append(handlers, eventHandler)
		}
	}

	return handlers
}

func (ee *EventEditor) extractEventActions() []EventAction {
	var actions []EventAction

	if ee.component.Events.Actions != nil {
		for _, action := range ee.component.Events.Actions {
			eventAction := EventAction{
				ID:          action.ID,
				Type:        action.Type,
				Target:      action.Target,
				Config:      action.Config,
				Async:       action.Async,
				RetryCount:  action.RetryCount,
				Timeout:     action.Timeout,
				OnSuccess:   action.OnSuccess,
				OnError:     action.OnError,
				OnComplete:  action.OnComplete,
				Enabled:     action.Enabled,
				Description: action.Description,
			}
			actions = append(actions, eventAction)
		}
	}

	return actions
}

func (ee *EventEditor) extractEventConditions() []EventCondition {
	var conditions []EventCondition

	if ee.component.Events.Conditions != nil {
		for _, condition := range ee.component.Events.Conditions {
			eventCondition := EventCondition{
				ID:          condition.ID,
				Type:        condition.Type,
				Expression:  condition.Expression,
				Config:      condition.Config,
				Invert:      condition.Invert,
				Description: condition.Description,
			}
			conditions = append(conditions, eventCondition)
		}
	}

	return conditions
}

func (ee *EventEditor) extractEventValidators() []EventValidator {
	var validators []EventValidator

	if ee.component.Events.Validators != nil {
		for _, validator := range ee.component.Events.Validators {
			eventValidator := EventValidator{
				ID:       validator.ID,
				Type:     validator.Type,
				Field:    validator.Field,
				Config:   validator.Config,
				Message:  validator.Message,
				Severity: validator.Severity,
			}
			validators = append(validators, eventValidator)
		}
	}

	return validators
}

// Helper methods for applying configuration

func (ee *EventEditor) applyEventHandlers(handlers []EventHandler) error {
	ee.component.Events.Handlers = make([]core.EventHandler, len(handlers))

	for i, handler := range handlers {
		ee.component.Events.Handlers[i] = core.EventHandler{
			ID:         handler.ID,
			Event:      handler.Event,
			Target:     handler.Target,
			Actions:    handler.Actions,
			Conditions: handler.Conditions,
			Debounce:   handler.Debounce,
			Throttle:   handler.Throttle,
			Once:       handler.Once,
			Passive:    handler.Passive,
			Capture:    handler.Capture,
			Enabled:    handler.Enabled,
			Options:    handler.Options,
		}
	}

	return nil
}

func (ee *EventEditor) applyEventActions(actions []EventAction) error {
	ee.component.Events.Actions = make([]core.EventAction, len(actions))

	for i, action := range actions {
		ee.component.Events.Actions[i] = core.EventAction{
			ID:          action.ID,
			Type:        action.Type,
			Target:      action.Target,
			Config:      action.Config,
			Async:       action.Async,
			RetryCount:  action.RetryCount,
			Timeout:     action.Timeout,
			OnSuccess:   action.OnSuccess,
			OnError:     action.OnError,
			OnComplete:  action.OnComplete,
			Enabled:     action.Enabled,
			Description: action.Description,
		}
	}

	return nil
}

func (ee *EventEditor) applyEventConditions(conditions []EventCondition) error {
	if len(conditions) == 0 {
		ee.component.Events.Conditions = nil
		return nil
	}

	ee.component.Events.Conditions = make([]core.EventCondition, len(conditions))

	for i, condition := range conditions {
		ee.component.Events.Conditions[i] = core.EventCondition{
			ID:          condition.ID,
			Type:        condition.Type,
			Expression:  condition.Expression,
			Config:      condition.Config,
			Invert:      condition.Invert,
			Description: condition.Description,
		}
	}

	return nil
}

func (ee *EventEditor) applyEventValidators(validators []EventValidator) error {
	if len(validators) == 0 {
		ee.component.Events.Validators = nil
		return nil
	}

	ee.component.Events.Validators = make([]core.EventValidator, len(validators))

	for i, validator := range validators {
		ee.component.Events.Validators[i] = core.EventValidator{
			ID:       validator.ID,
			Type:     validator.Type,
			Field:    validator.Field,
			Config:   validator.Config,
			Message:  validator.Message,
			Severity: validator.Severity,
		}
	}

	return nil
}
