package schema

import (
	"context"
	"time"
)

// Runtime Types - Shared across packages to avoid import cycles
// These types are used by the runtime package and other packages that integrate with it

// ═══════════════════════════════════════════════════════════════════════════
// Event Types and Structures
// ═══════════════════════════════════════════════════════════════════════════

// EventType represents runtime event types
type EventType string

const (
	EventChange EventType = "change" // Field value changed
	EventBlur   EventType = "blur"   // Field lost focus
	EventFocus  EventType = "focus"  // Field gained focus
	EventSubmit EventType = "submit" // Form submitted
	EventReset  EventType = "reset"  // Form reset
	EventInit   EventType = "init"   // Form initialized
)

// Event represents a runtime event
type Event struct {
	Type      EventType `json:"type"`
	Field     string    `json:"field,omitempty"`
	Value     any       `json:"value,omitempty"`
	OldValue  any       `json:"old_value,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Context   any       `json:"context,omitempty"`
}

// EventCallback is a function that handles events
type EventCallback func(ctx context.Context, event *Event) error

// ═══════════════════════════════════════════════════════════════════════════
// Validation Types
// ═══════════════════════════════════════════════════════════════════════════

// ValidationTiming defines when validation occurs
type ValidationTiming string

const (
	ValidateOnChange ValidationTiming = "change" // Validate immediately on change
	ValidateOnBlur   ValidationTiming = "blur"   // Validate when field loses focus
	ValidateOnSubmit ValidationTiming = "submit" // Validate only on form submission
	ValidateNever    ValidationTiming = "never"  // Never validate automatically
)

// ═══════════════════════════════════════════════════════════════════════════
// State Types and Structures
// ═══════════════════════════════════════════════════════════════════════════

// StateSnapshot represents a point-in-time snapshot of state
// Used for undo/redo functionality and state history
type StateSnapshot struct {
	Values      map[string]any      `json:"values"`
	Initial     map[string]any      `json:"initial"`
	Touched     map[string]bool     `json:"touched"`
	Dirty       map[string]bool     `json:"dirty"`
	Errors      map[string][]string `json:"errors"`
	Timestamp   time.Time           `json:"timestamp"`
	Initialized time.Time           `json:"initialized"`
}

// StateStats provides statistics about the current state
type StateStats struct {
	FieldCount        int       `json:"field_count"`
	TouchedCount      int       `json:"touched_count"`
	DirtyCount        int       `json:"dirty_count"`
	ErrorCount        int       `json:"error_count"`
	IsValid           bool      `json:"is_valid"`
	InitializedAt     time.Time `json:"initialized_at"`
	LastUpdated       time.Time `json:"last_updated"`
	HasUnsavedChanges bool      `json:"has_unsaved_changes"`
}

// ═══════════════════════════════════════════════════════════════════════════
// Runtime Configuration
// ═══════════════════════════════════════════════════════════════════════════

// RuntimeConfig holds runtime configuration options
type RuntimeConfig struct {
	EnableConditionals  bool             `json:"enableConditionals"`
	EnableValidation    bool             `json:"enableValidation"`
	ValidationTiming    ValidationTiming `json:"validationTiming"`
	EnableDebounce      bool             `json:"enableDebounce"`
	DebounceDelay       time.Duration    `json:"debounceDelay"`
	EnableEventTracking bool             `json:"enableEventTracking"`
	MaxStateSnapshots   int              `json:"maxStateSnapshots"`
}

// DefaultRuntimeConfig returns sensible defaults
func DefaultRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		EnableConditionals:  true,
		EnableValidation:    true,
		ValidationTiming:    ValidateOnBlur,
		EnableDebounce:      true,
		DebounceDelay:       300 * time.Millisecond,
		EnableEventTracking: false,
		MaxStateSnapshots:   10,
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// Runtime Statistics
// ═══════════════════════════════════════════════════════════════════════════

// RuntimeStats provides runtime statistics
type RuntimeStats struct {
	FieldCount    int           `json:"field_count"`
	VisibleFields int           `json:"visible_fields"`
	TouchedFields int           `json:"touched_fields"`
	DirtyFields   int           `json:"dirty_fields"`
	ErrorCount    int           `json:"error_count"`
	IsValid       bool          `json:"is_valid"`
	InitializedAt time.Time     `json:"initialized_at"`
	LastActivity  time.Time     `json:"last_activity"`
	Uptime        time.Duration `json:"uptime"`
	EventStats    *EventStats   `json:"event_stats,omitempty"`
}

// EventStats provides statistics about events
type EventStats struct {
	TotalEvents    int               `json:"total_events"`
	EventsByType   map[EventType]int `json:"events_by_type"`
	EventsByField  map[string]int    `json:"events_by_field"`
	LastEventTime  time.Time         `json:"last_event_time"`
	LastEventType  EventType         `json:"last_event_type"`
	LastEventField string            `json:"last_event_field"`
	FirstEventTime time.Time         `json:"first_event_time"`
	StartTime      time.Time         `json:"start_time"`
}

// ═══════════════════════════════════════════════════════════════════════════
// Event Handler Configuration
// ═══════════════════════════════════════════════════════════════════════════

// EventHandlerConfig holds configuration for event handler
type EventHandlerConfig struct {
	ValidationTiming ValidationTiming `json:"validationTiming"`
	EnableTracking   bool             `json:"enableTracking"`
	Enabled          bool             `json:"enabled"`
}

// DefaultEventHandlerConfig returns sensible defaults
func DefaultEventHandlerConfig() *EventHandlerConfig {
	return &EventHandlerConfig{
		ValidationTiming: ValidateOnBlur,
		EnableTracking:   false,
		Enabled:          true,
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// Debouncing Configuration
// ═══════════════════════════════════════════════════════════════════════════

// DebouncedConfig holds debouncing configuration
type DebouncedConfig struct {
	ChangeDelay time.Duration `json:"changeDelay"`
	BlurDelay   time.Duration `json:"blurDelay"`
	FocusDelay  time.Duration `json:"focusDelay"`
}

// DefaultDebouncedConfig returns sensible debouncing defaults
func DefaultDebouncedConfig() *DebouncedConfig {
	return &DebouncedConfig{
		ChangeDelay: 300 * time.Millisecond,
		BlurDelay:   100 * time.Millisecond,
		FocusDelay:  0, // No debounce for focus
	}
}

// Runtime interfaces - provide abstraction layer for UI implementations
// These interfaces allow runtime to be UI-agnostic while still providing rich functionality

// RuntimeRenderer interface for UI implementations (Templ, Go templates, React SSR, etc.)
// Uses existing Field, Action, Layout, Schema types from schema package
type RuntimeRenderer interface {
	// RenderField renders a single field with schema context and current state values
	RenderField(ctx context.Context, field *Field, value any, errors []string, touched, dirty bool) (string, error)

	// RenderForm renders the complete form using schema and runtime state
	// State contains all values, errors, touched/dirty tracking
	RenderForm(ctx context.Context, schema *Schema, state map[string]any, errors map[string][]string) (string, error)

	// RenderAction renders an action button/element
	RenderAction(ctx context.Context, action *Action, enabled bool) (string, error)

	// RenderLayout renders layout components using existing schema types
	RenderLayout(ctx context.Context, layout *Layout, fields []*Field, state map[string]any) (string, error)
	RenderSection(ctx context.Context, section *Section, fields []*Field, state map[string]any) (string, error)
	RenderTab(ctx context.Context, tab *Tab, fields []*Field, state map[string]any) (string, error)
	RenderStep(ctx context.Context, step *Step, fields []*Field, state map[string]any) (string, error)
	RenderGroup(ctx context.Context, group *Group, fields []*Field, state map[string]any) (string, error)

	// RenderErrors renders validation errors for display
	RenderErrors(ctx context.Context, errors map[string][]string) (string, error)
}

// RuntimeStateManager interface - implemented by runtime.State
// This interface allows runtime package to be used without import cycles
type RuntimeStateManager interface {
	// Core state operations - matches runtime.State methods
	GetValue(fieldName string) (any, bool)
	SetValue(fieldName string, value any) error
	GetAll() map[string]any

	// State tracking - matches runtime.State methods
	IsTouched(fieldName string) bool
	Touch(fieldName string)
	IsDirty(fieldName string) bool
	IsAnyDirty() bool

	// Error management - matches runtime.State methods
	GetErrors(fieldName string) []string
	SetErrors(fieldName string, errors []string)
	GetAllErrors() map[string][]string
	IsValid() bool

	// State lifecycle - matches runtime.State methods
	Initialize(schema *Schema, initialData map[string]any) error
	Reset()
	CreateSnapshot() any          // returns runtime.StateSnapshot
	RestoreSnapshot(snapshot any) // takes runtime.StateSnapshot
}

// RuntimeEventHandler interface - implemented by runtime.EventHandler
// Uses existing runtime.Event types and runtime.EventCallback
type RuntimeEventHandler interface {
	// Core events - uses existing runtime.Event struct
	OnChange(ctx context.Context, event any) error // takes runtime.Event
	OnBlur(ctx context.Context, event any) error   // takes runtime.Event
	OnFocus(ctx context.Context, event any) error  // takes runtime.Event
	OnSubmit(ctx context.Context) error

	// Custom event registration - uses existing runtime types
	Register(eventType string, handler any) // takes runtime.EventCallback
	Unregister(eventType string)

	// Validation timing - matches runtime.EventHandler
	SetValidationTiming(timing string) // uses runtime.ValidationTiming
	GetValidationTiming() string
}

// RuntimeValidator interface for validation implementations
// Uses existing Field, Action, Workflow, Schema types from schema package
type RuntimeValidator interface {
	// Field validation
	ValidateField(ctx context.Context, field *Field, value any, allData map[string]any) []string
	ValidateAllFields(ctx context.Context, schema *Schema, data map[string]any) map[string][]string

	// Action validation
	ValidateAction(ctx context.Context, action *Action, formState map[string]any) error

	// Workflow validation
	ValidateWorkflow(ctx context.Context, workflow *Workflow, currentStage string, formState map[string]any) error

	// Business rules validation using existing ValidationRule from enterprise.go
	ValidateBusinessRules(ctx context.Context, rules []ValidationRule, allValues map[string]any) map[string][]string

	// Async validation
	ValidateFieldAsync(ctx context.Context, field *Field, value any, callback ValidationCallback) error
}

// RuntimeConditionalEngine interface for conditional logic
// Uses existing Field, Action, Layout, Workflow, Schema types from schema package
type RuntimeConditionalEngine interface {
	// Evaluate field visibility and state
	EvaluateFieldVisibility(ctx context.Context, field *Field, data map[string]any) (bool, error)
	EvaluateFieldRequired(ctx context.Context, field *Field, data map[string]any) (bool, error)
	EvaluateFieldEditable(ctx context.Context, field *Field, data map[string]any) (bool, error)

	// Action conditional evaluation
	EvaluateActionConditions(ctx context.Context, action *Action, allValues map[string]any) (bool, error)

	// Layout conditional evaluation
	EvaluateLayoutConditions(ctx context.Context, layout *Layout, allValues map[string]any) (*LayoutConditionalResult, error)

	// Workflow conditional evaluation
	EvaluateWorkflowConditions(ctx context.Context, workflow *Workflow, allValues map[string]any) ([]WorkflowAction, error)

	// Batch evaluations
	EvaluateAllConditions(ctx context.Context, schema *Schema, data map[string]any) (*ConditionalResults, error)

	// Watch for changes
	WatchConditions(ctx context.Context, schema *Schema, onChange ConditionalChangeCallback) error
}

// Runtime data structures - use existing types from runtime package where possible

// ActionRuntimeState represents the runtime state of an action
type ActionRuntimeState struct {
	ID      string `json:"id"`
	Visible bool   `json:"visible"`
	Enabled bool   `json:"enabled"`
	Loading bool   `json:"loading"`
	Reason  string `json:"reason,omitempty"`
}

// ValidationCallback  rllback types for runtime interfaces
type (
	ValidationCallback        func(fieldName string, errors []string)
	ConditionalChangeCallback func(ctx context.Context, updates []ConditionalUpdate)
)

// RuntimeValidationRule Custom validation rule interface
type RuntimeValidationRule interface {
	ID() string
	Validate(ctx context.Context, value any, allData map[string]any) error
}

// ═══════════════════════════════════════════════════════════════════════════
// Runtime Interfaces (already defined in schema.go, kept here for reference)
// ═══════════════════════════════════════════════════════════════════════════

// Note: The following interfaces are already defined in schema.go:
// - RuntimeRenderer
// - RuntimeValidator
// - RuntimeConditionalEngine
// - RuntimeStateManager
// - RuntimeEvent

// These types can be used by any package that needs to integrate with the runtime
// without creating import cycles, as they're all defined in the schema package.
