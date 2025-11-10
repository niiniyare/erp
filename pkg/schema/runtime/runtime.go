package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/niiniyare/erp/pkg/schema"
)

// Runtime is the main runtime implementation that orchestrates state and events
// It's UI-agnostic and uses interfaces for maximum flexibility
// Uses types from schema package to avoid import cycles
type Runtime struct {
	// Core components
	schema *schema.Schema
	state  *State
	events *EventHandler

	// Injected dependencies (optional - gracefully degrade if not provided)
	renderer    schema.RuntimeRenderer          // UI rendering implementation
	validator   schema.RuntimeValidator         // Validation implementation
	conditional schema.RuntimeConditionalEngine // Conditional logic implementation
	
	// Internationalization
	currentLocale string // Current locale for the runtime instance

	// Configuration
	config *schema.RuntimeConfig

	// Statistics and tracking
	initializedAt time.Time
	lastActivity  time.Time

	// Concurrency control
	mu sync.RWMutex
}

// ═══════════════════════════════════════════════════════════════════════════
// Builder Pattern for Runtime Creation
// ═══════════════════════════════════════════════════════════════════════════

// RuntimeBuilder provides a fluent builder pattern for runtime creation
type RuntimeBuilder struct {
	schema        *schema.Schema
	config        *schema.RuntimeConfig
	renderer      schema.RuntimeRenderer
	validator     schema.RuntimeValidator
	conditional   schema.RuntimeConditionalEngine
	locale        string // Default locale for this runtime
	initialData   map[string]any
	eventHandlers map[schema.EventType][]schema.EventCallback
}

// NewRuntimeBuilder creates a new builder for runtime
func NewRuntimeBuilder(enrichedSchema *schema.Schema) *RuntimeBuilder {
	return &RuntimeBuilder{
		schema:        enrichedSchema,
		config:        schema.DefaultRuntimeConfig(),
		initialData:   make(map[string]any),
		eventHandlers: make(map[schema.EventType][]schema.EventCallback),
	}
}

// WithRenderer sets the UI renderer implementation
func (b *RuntimeBuilder) WithRenderer(renderer schema.RuntimeRenderer) *RuntimeBuilder {
	b.renderer = renderer
	return b
}

// WithValidator sets the validation implementation
func (b *RuntimeBuilder) WithValidator(validator schema.RuntimeValidator) *RuntimeBuilder {
	b.validator = validator
	return b
}

// WithConditionalEngine sets the conditional logic implementation
func (b *RuntimeBuilder) WithConditionalEngine(conditional schema.RuntimeConditionalEngine) *RuntimeBuilder {
	b.conditional = conditional
	return b
}

// WithLocale sets the default locale for the runtime
func (b *RuntimeBuilder) WithLocale(locale string) *RuntimeBuilder {
	b.locale = locale
	return b
}

// WithConfig sets custom configuration
func (b *RuntimeBuilder) WithConfig(config *schema.RuntimeConfig) *RuntimeBuilder {
	b.config = config
	return b
}

// WithInitialData sets initial form data
func (b *RuntimeBuilder) WithInitialData(data map[string]any) *RuntimeBuilder {
	for k, v := range data {
		b.initialData[k] = v
	}
	return b
}

// WithEventHandler registers an event handler
func (b *RuntimeBuilder) WithEventHandler(eventType schema.EventType, handler schema.EventCallback) *RuntimeBuilder {
	b.eventHandlers[eventType] = append(b.eventHandlers[eventType], handler)
	return b
}

// WithValidationTiming sets validation timing strategy
func (b *RuntimeBuilder) WithValidationTiming(timing schema.ValidationTiming) *RuntimeBuilder {
	b.config.ValidationTiming = timing
	return b
}

// WithDebounce configures debouncing
func (b *RuntimeBuilder) WithDebounce(enabled bool, delay time.Duration) *RuntimeBuilder {
	b.config.EnableDebounce = enabled
	b.config.DebounceDelay = delay
	return b
}

// Build creates the runtime instance
func (b *RuntimeBuilder) Build(ctx context.Context) (*Runtime, error) {
	if b.schema == nil {
		return nil, fmt.Errorf("schema is required")
	}

	// Create core components
	state := NewState()

	// Create event handler with config
	eventConfig := &schema.EventHandlerConfig{
		ValidationTiming: b.config.ValidationTiming,
		EnableTracking:   b.config.EnableEventTracking,
		Enabled:          true,
	}
	events := NewEventHandlerWithConfig(eventConfig)

	// Create runtime instance
	runtime := &Runtime{
		schema:        b.schema,
		state:         state,
		events:        events,
		renderer:      b.renderer,
		validator:     b.validator,
		conditional:   b.conditional,
		currentLocale: b.locale,
		config:        b.config,
		initializedAt: time.Now(),
		lastActivity:  time.Now(),
	}

	// Register custom event handlers
	for eventType, handlers := range b.eventHandlers {
		for _, handler := range handlers {
			events.Register(eventType, handler)
		}
	}

	// Initialize state
	if err := runtime.Initialize(ctx, b.initialData); err != nil {
		return nil, fmt.Errorf("failed to initialize runtime: %w", err)
	}

	return runtime, nil
}

// MustBuild builds the runtime and panics on error (for testing/simple cases)
func (b *RuntimeBuilder) MustBuild(ctx context.Context) *Runtime {
	runtime, err := b.Build(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to build runtime: %v", err))
	}
	return runtime
}

// ═══════════════════════════════════════════════════════════════════════════
// Initialization and Lifecycle
// ═══════════════════════════════════════════════════════════════════════════

// Initialize initializes the runtime with data and performs initial validation
func (r *Runtime) Initialize(ctx context.Context, initialData map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Initialize state
	if err := r.state.Initialize(r.schema, initialData); err != nil {
		return fmt.Errorf("failed to initialize state: %w", err)
	}

	// Trigger init event
	if err := r.events.OnInit(ctx); err != nil {
		return fmt.Errorf("failed to trigger init event: %w", err)
	}

	// Run initial conditional logic if enabled
	if r.config.EnableConditionals && r.conditional != nil {
		if err := r.applyConditionalsUnsafe(ctx); err != nil {
			// Log error but don't fail - graceful degradation
		}
	}

	// Run initial validation if enabled and validator is available
	if r.config.EnableValidation && r.validator != nil {
		errors := r.validator.ValidateAllFields(ctx, r.schema, r.state.GetAll())
		for field, fieldErrors := range errors {
			r.state.SetErrors(field, fieldErrors)
		}
	}

	r.initializedAt = time.Now()
	r.lastActivity = time.Now()

	return nil
}

// ═══════════════════════════════════════════════════════════════════════════
// Event Handling Methods
// ═══════════════════════════════════════════════════════════════════════════

// HandleFieldChange processes a field value change
func (r *Runtime) HandleFieldChange(ctx context.Context, fieldName string, newValue any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastActivity = time.Now()

	// Check if field exists and is editable
	field := r.getField(fieldName)
	if field == nil {
		return fmt.Errorf("field %s not found", fieldName)
	}
	
	// Check if field is editable
	if field.Runtime != nil && !field.Runtime.Editable {
		return fmt.Errorf("field %s is read-only", fieldName)
	}

	// Get old value for event
	oldValue, _ := r.state.GetValue(fieldName)

	// Update state
	if err := r.state.SetValue(fieldName, newValue); err != nil {
		return fmt.Errorf("failed to update field value: %w", err)
	}

	// Create event
	event := &schema.Event{
		Type:      schema.EventChange,
		Field:     fieldName,
		Value:     newValue,
		OldValue:  oldValue,
		Timestamp: time.Now(),
	}

	// Trigger event
	if err := r.events.OnChange(ctx, event); err != nil {
		return fmt.Errorf("failed to handle change event: %w", err)
	}

	// Run conditional logic if enabled
	if r.config.EnableConditionals && r.conditional != nil {
		if err := r.applyConditionalsUnsafe(ctx); err != nil {
			// Log error but continue - graceful degradation
		}
	}

	// Run validation if enabled and timing matches
	if r.config.EnableValidation && r.validator != nil && r.events.ShouldValidateOn(schema.EventChange) {
		field := r.getField(fieldName)
		if field != nil {
			errors := r.validator.ValidateField(ctx, field, newValue, r.state.GetAll())
			// Localize validation messages using embedded translations
			localizedErrors := r.localizeValidationErrors(errors)
			r.state.SetErrors(fieldName, localizedErrors)
		}
	}

	return nil
}

// HandleFieldBlur processes a field blur event
func (r *Runtime) HandleFieldBlur(ctx context.Context, fieldName string, value any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastActivity = time.Now()

	// Mark as touched
	r.state.Touch(fieldName)

	// Create event
	event := &schema.Event{
		Type:      schema.EventBlur,
		Field:     fieldName,
		Value:     value,
		Timestamp: time.Now(),
	}

	// Trigger event
	if err := r.events.OnBlur(ctx, event); err != nil {
		return fmt.Errorf("failed to handle blur event: %w", err)
	}

	// Run validation if enabled and timing matches
	if r.config.EnableValidation && r.validator != nil && r.events.ShouldValidateOn(schema.EventBlur) {
		field := r.getField(fieldName)
		if field != nil {
			errors := r.validator.ValidateField(ctx, field, value, r.state.GetAll())
			// Localize validation messages using embedded translations
			localizedErrors := r.localizeValidationErrors(errors)
			r.state.SetErrors(fieldName, localizedErrors)
		}
	}

	return nil
}

// HandleFieldFocus processes a field focus event
func (r *Runtime) HandleFieldFocus(ctx context.Context, fieldName string, value any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastActivity = time.Now()

	// Create event
	event := &schema.Event{
		Type:      schema.EventFocus,
		Field:     fieldName,
		Value:     value,
		Timestamp: time.Now(),
	}

	// Trigger event
	return r.events.OnFocus(ctx, event)
}

// HandleSubmit processes form submission
func (r *Runtime) HandleSubmit(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastActivity = time.Now()

	// Run final validation if enabled and validator is available
	if r.config.EnableValidation && r.validator != nil {
		allErrors := r.validator.ValidateAllFields(ctx, r.schema, r.state.GetAll())

		// Update state with validation errors
		for field, errors := range allErrors {
			// Localize validation messages using embedded translations
			localizedErrors := r.localizeValidationErrors(errors)
			r.state.SetErrors(field, localizedErrors)
		}

		// Check if form is valid
		if !r.state.IsValid() {
			return fmt.Errorf("form validation failed")
		}
	}

	// Trigger submit event
	return r.events.OnSubmit(ctx)
}

// HandleReset processes form reset
func (r *Runtime) HandleReset(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastActivity = time.Now()

	// Reset state
	r.state.Reset()

	// Trigger reset event
	return r.events.OnReset(ctx)
}

// ═══════════════════════════════════════════════════════════════════════════
// Validation Methods
// ═══════════════════════════════════════════════════════════════════════════

// ValidateField validates a single field value
func (r *Runtime) ValidateField(ctx context.Context, fieldName string, value any) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.validator == nil {
		return nil // No validator configured - skip validation
	}

	field := r.getField(fieldName)
	if field == nil {
		return []string{"field not found"}
	}

	// Check if field is visible
	if field.Runtime != nil && !field.Runtime.Visible {
		return nil // Don't validate invisible fields
	}

	errors := r.validator.ValidateField(ctx, field, value, r.state.GetAll())
	// Localize validation messages using embedded translations
	return r.localizeValidationErrors(errors)
}

// ValidateCurrentState validates all fields using current state
func (r *Runtime) ValidateCurrentState(ctx context.Context) map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.validator == nil {
		return make(map[string][]string) // No validator - return empty errors
	}

	return r.validator.ValidateAllFields(ctx, r.schema, r.state.GetAll())
}

// ValidateWithDebounce validates after a delay for better UX
func (r *Runtime) ValidateWithDebounce(
	ctx context.Context,
	fieldName string,
	value any,
	delay time.Duration,
) <-chan []string {
	result := make(chan []string, 1)

	go func() {
		defer close(result)

		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-timer.C:
			errors := r.ValidateField(ctx, fieldName, value)
			result <- errors
		case <-ctx.Done():
			return
		}
	}()

	return result
}

// ValidateAction validates an action before execution
func (r *Runtime) ValidateAction(ctx context.Context, action *schema.Action) error {
	if r.validator == nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.validator.ValidateAction(ctx, action, r.state.GetAll())
}

// ValidateWorkflow validates workflow state
func (r *Runtime) ValidateWorkflow(ctx context.Context, workflow *schema.Workflow) error {
	if r.validator == nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	currentStage := workflow.GetCurrentStage()
	return r.validator.ValidateWorkflow(ctx, workflow, currentStage, r.state.GetAll())
}

// ValidateBusinessRules validates business rules
func (r *Runtime) ValidateBusinessRules(ctx context.Context, rules []schema.ValidationRule) map[string][]string {
	if r.validator == nil {
		return make(map[string][]string)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.validator.ValidateBusinessRules(ctx, rules, r.state.GetAll())
}

// ═══════════════════════════════════════════════════════════════════════════
// Conditional Logic Methods
// ═══════════════════════════════════════════════════════════════════════════

// ApplyConditionalLogic evaluates and applies conditional rules
func (r *Runtime) ApplyConditionalLogic(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.applyConditionalsUnsafe(ctx)
}

// applyConditionalsUnsafe applies conditionals without locking (must be called with lock held)
func (r *Runtime) applyConditionalsUnsafe(ctx context.Context) error {
	if r.conditional == nil {
		return r.applyBasicConditionals(ctx)
	}

	results, err := r.conditional.EvaluateAllConditions(ctx, r.schema, r.state.GetAll())
	if err != nil {
		return fmt.Errorf("failed to evaluate conditions: %w", err)
	}

	if results.Changed {
		for _, update := range results.Updates {
			field := r.getField(update.FieldName)
			if field != nil && field.Runtime != nil {
				field.Runtime.Visible = update.Visible
				field.Runtime.Editable = update.Editable
				if update.Required {
					field.Required = update.Required
				}
				if update.Reason != "" {
					field.Runtime.Reason = update.Reason
				}
			}
		}

		if r.schema.Layout != nil {
			_, _ = r.conditional.EvaluateLayoutConditions(ctx, r.schema.Layout, r.state.GetAll())
		}

		if r.schema.Workflow != nil {
			_, _ = r.conditional.EvaluateWorkflowConditions(ctx, r.schema.Workflow, r.state.GetAll())
		}
	}

	return nil
}

// applyBasicConditionals applies basic field conditionals without full conditional engine
func (r *Runtime) applyBasicConditionals(ctx context.Context) error {
	data := r.state.GetAll()

	for i := range r.schema.Fields {
		field := &r.schema.Fields[i]

		if field.Conditional != nil {
			visible, err := field.IsVisible(ctx, data)
			if err != nil {
				continue
			}

			if field.Runtime != nil {
				field.Runtime.Visible = visible
			}
		}
	}

	return nil
}

// ═══════════════════════════════════════════════════════════════════════════
// Rendering Methods
// ═══════════════════════════════════════════════════════════════════════════

// Render renders the form using the injected renderer
func (r *Runtime) Render(ctx context.Context) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.renderer == nil {
		return "", fmt.Errorf("no renderer configured - use WithRenderer() in builder")
	}

	return r.renderer.RenderForm(ctx, r.schema, r.state.GetAll(), r.state.GetAllErrors())
}

// RenderField renders a single field using the injected renderer
func (r *Runtime) RenderField(ctx context.Context, fieldName string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.renderer == nil {
		return "", fmt.Errorf("no renderer configured")
	}

	field := r.getField(fieldName)
	if field == nil {
		return "", fmt.Errorf("field %s not found", fieldName)
	}

	value, _ := r.state.GetValue(fieldName)
	errors := r.state.GetErrors(fieldName)
	touched := r.state.IsTouched(fieldName)
	dirty := r.state.IsDirty(fieldName)

	return r.renderer.RenderField(ctx, field, value, errors, touched, dirty)
}

// RenderAction renders an action
func (r *Runtime) RenderAction(ctx context.Context, action *schema.Action) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.renderer == nil {
		return "", fmt.Errorf("no renderer configured")
	}

	enabled := !action.Disabled && !action.Loading

	if r.conditional != nil {
		visible, err := r.conditional.EvaluateActionConditions(ctx, action, r.state.GetAll())
		if err == nil && !visible {
			return "", nil
		}
	}

	return r.renderer.RenderAction(ctx, action, enabled)
}

// RenderLayout renders layout
func (r *Runtime) RenderLayout(ctx context.Context, layout *schema.Layout) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.renderer == nil {
		return "", fmt.Errorf("no renderer configured")
	}

	fieldNames := layout.GetAllFields()
	fields := make([]*schema.Field, 0, len(fieldNames))
	for _, fieldName := range fieldNames {
		field := r.getField(fieldName)
		if field != nil {
			fields = append(fields, field)
		}
	}

	return r.renderer.RenderLayout(ctx, layout, fields, r.state.GetAll())
}

// RenderSection renders a section
func (r *Runtime) RenderSection(ctx context.Context, section *schema.Section) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.renderer == nil {
		return "", fmt.Errorf("no renderer configured")
	}

	fields := make([]*schema.Field, 0, len(section.Fields))
	for _, fieldName := range section.Fields {
		field := r.getField(fieldName)
		if field != nil {
			fields = append(fields, field)
		}
	}

	return r.renderer.RenderSection(ctx, section, fields, r.state.GetAll())
}

// RenderTab renders a tab
func (r *Runtime) RenderTab(ctx context.Context, tab *schema.Tab) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.renderer == nil {
		return "", fmt.Errorf("no renderer configured")
	}

	fields := make([]*schema.Field, 0, len(tab.Fields))
	for _, fieldName := range tab.Fields {
		field := r.getField(fieldName)
		if field != nil {
			fields = append(fields, field)
		}
	}

	return r.renderer.RenderTab(ctx, tab, fields, r.state.GetAll())
}

// RenderStep renders a step
func (r *Runtime) RenderStep(ctx context.Context, step *schema.Step) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.renderer == nil {
		return "", fmt.Errorf("no renderer configured")
	}

	fields := make([]*schema.Field, 0, len(step.Fields))
	for _, fieldName := range step.Fields {
		field := r.getField(fieldName)
		if field != nil {
			fields = append(fields, field)
		}
	}

	return r.renderer.RenderStep(ctx, step, fields, r.state.GetAll())
}

// RenderGroup renders a group
func (r *Runtime) RenderGroup(ctx context.Context, group *schema.Group) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.renderer == nil {
		return "", fmt.Errorf("no renderer configured")
	}

	fields := make([]*schema.Field, 0, len(group.Fields))
	for _, fieldName := range group.Fields {
		field := r.getField(fieldName)
		if field != nil {
			fields = append(fields, field)
		}
	}

	return r.renderer.RenderGroup(ctx, group, fields, r.state.GetAll())
}

// RenderErrors renders validation errors
func (r *Runtime) RenderErrors(ctx context.Context) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.renderer == nil {
		return "", fmt.Errorf("no renderer configured")
	}

	return r.renderer.RenderErrors(ctx, r.state.GetAllErrors())
}

// ═══════════════════════════════════════════════════════════════════════════
// Event Registration
// ═══════════════════════════════════════════════════════════════════════════

// RegisterEventHandler registers a custom event handler
func (r *Runtime) RegisterEventHandler(eventType schema.EventType, callback schema.EventCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events.Register(eventType, callback)
}

// SetValidationTiming configures when validation occurs
func (r *Runtime) SetValidationTiming(timing schema.ValidationTiming) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.config.ValidationTiming = timing
	r.events.SetValidationTiming(timing)
}

// ═══════════════════════════════════════════════════════════════════════════
// State Accessors
// ═══════════════════════════════════════════════════════════════════════════

// GetState returns the state manager interface
func (r *Runtime) GetState() schema.RuntimeStateManager {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

// GetRawState returns the concrete state implementation
func (r *Runtime) GetRawState() *State {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

// GetSchema returns the schema
func (r *Runtime) GetSchema() *schema.Schema {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.schema
}

// GetConfig returns the configuration
func (r *Runtime) GetConfig() *schema.RuntimeConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

// IsValid checks if the form is valid
func (r *Runtime) IsValid() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.IsValid()
}

// GetErrors returns all validation errors
func (r *Runtime) GetErrors() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.GetAllErrors()
}

// Reset resets the form to initial state
func (r *Runtime) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.state.Reset()
	r.events.OnReset(context.Background())
}

// GetFieldValue gets a field value
func (r *Runtime) GetFieldValue(fieldName string) (any, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.GetValue(fieldName)
}

// SetFieldValue sets a field value
func (r *Runtime) SetFieldValue(fieldName string, value any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.state.SetValue(fieldName, value)
}

// IsDirty checks if any field has been modified
func (r *Runtime) IsDirty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.IsAnyDirty()
}

// IsFieldDirty checks if a specific field has been modified
func (r *Runtime) IsFieldDirty(fieldName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.IsDirty(fieldName)
}

// IsFieldTouched checks if a field has been touched
func (r *Runtime) IsFieldTouched(fieldName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.IsTouched(fieldName)
}

// GetAllData returns all form data
func (r *Runtime) GetAllData() map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state.GetAll()
}

// UpdateSchema updates the schema
func (r *Runtime) UpdateSchema(newSchema *schema.Schema) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.schema = newSchema
	return nil
}

// ═══════════════════════════════════════════════════════════════════════════
// Statistics
// ═══════════════════════════════════════════════════════════════════════════

// GetStats returns runtime statistics
func (r *Runtime) GetStats() *schema.RuntimeStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := &schema.RuntimeStats{
		FieldCount:    len(r.schema.Fields),
		TouchedFields: r.state.GetTouchedCount(),
		DirtyFields:   r.state.GetDirtyCount(),
		ErrorCount:    r.state.GetErrorCount(),
		IsValid:       r.state.IsValid(),
		InitializedAt: r.initializedAt,
		LastActivity:  r.lastActivity,
		Uptime:        time.Since(r.initializedAt),
	}

	// Count visible fields
	for i := range r.schema.Fields {
		field := &r.schema.Fields[i]
		if field.Runtime == nil || field.Runtime.Visible {
			stats.VisibleFields++
		}
	}

	// Add event stats if tracking enabled
	if tracker := r.events.GetTracker(); tracker != nil {
		stats.EventStats = tracker.GetStats()
	}

	return stats
}

// ═══════════════════════════════════════════════════════════════════════════
// Private Helper Methods
// ═══════════════════════════════════════════════════════════════════════════

// getField finds a field in the schema by name
func (r *Runtime) getField(fieldName string) *schema.Field {
	for i := range r.schema.Fields {
		if r.schema.Fields[i].Name == fieldName {
			return &r.schema.Fields[i]
		}
	}
	return nil
}

// ═══════════════════════════════════════════════════════════════════════════
// Convenience Constructors
// ═══════════════════════════════════════════════════════════════════════════

// NewRuntime creates a runtime with default configuration
func NewRuntime(enrichedSchema *schema.Schema) *Runtime {
	runtime, err := NewRuntimeBuilder(enrichedSchema).
		Build(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to create runtime: %v", err))
	}
	return runtime
}

// NewRuntimeWithValidator creates a runtime with a validator
func NewRuntimeWithValidator(enrichedSchema *schema.Schema, validator schema.RuntimeValidator) *Runtime {
	runtime, err := NewRuntimeBuilder(enrichedSchema).
		WithValidator(validator).
		Build(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to create runtime: %v", err))
	}
	return runtime
}

// NewRuntimeWithRenderer creates a runtime with a renderer
func NewRuntimeWithRenderer(enrichedSchema *schema.Schema, renderer schema.RuntimeRenderer) *Runtime {
	runtime, err := NewRuntimeBuilder(enrichedSchema).
		WithRenderer(renderer).
		Build(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to create runtime: %v", err))
	}
	return runtime
}

// NewRuntimeWithConfig creates a runtime with custom config
func NewRuntimeWithConfig(enrichedSchema *schema.Schema, config *schema.RuntimeConfig) *Runtime {
	runtime, err := NewRuntimeBuilder(enrichedSchema).
		WithConfig(config).
		Build(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to create runtime: %v", err))
	}
	return runtime
}

// NewRuntimeFull creates a runtime with all dependencies
func NewRuntimeFull(
	enrichedSchema *schema.Schema,
	renderer schema.RuntimeRenderer,
	validator schema.RuntimeValidator,
	conditional schema.RuntimeConditionalEngine,
) *Runtime {
	runtime, err := NewRuntimeBuilder(enrichedSchema).
		WithRenderer(renderer).
		WithValidator(validator).
		WithConditionalEngine(conditional).
		Build(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to create runtime: %v", err))
	}
	return runtime
}

// NewRuntimeWithLocale creates a runtime with a specific locale
func NewRuntimeWithLocale(enrichedSchema *schema.Schema, locale string) *Runtime {
	runtime, err := NewRuntimeBuilder(enrichedSchema).
		WithLocale(locale).
		Build(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to create runtime: %v", err))
	}
	return runtime
}

// GetCurrentLocale returns the current locale for the runtime
func (r *Runtime) GetCurrentLocale() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.currentLocale == "" {
		return "en" // Default fallback
	}
	return r.currentLocale
}

// SetLocale changes the current locale for the runtime
func (r *Runtime) SetLocale(ctx context.Context, locale string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate locale is available
	if !schema.T_HasLocale(locale) {
		return fmt.Errorf("locale '%s' is not available", locale)
	}

	r.currentLocale = locale
	return nil
}

// DetectUserLocale automatically detects the best locale from user preferences
func (r *Runtime) DetectUserLocale(userPreferences []string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	bestLocale := schema.T_DetectLocale(userPreferences)
	r.currentLocale = bestLocale
	return bestLocale
}

// localizeValidationErrors translates validation error messages to current locale
func (r *Runtime) localizeValidationErrors(errors []string) []string {
	if len(errors) == 0 {
		return errors
	}

	locale := r.GetCurrentLocale()
	localizedErrors := make([]string, len(errors))
	
	for i, errorMsg := range errors {
		// Try to map common validation error patterns to translation keys
		translationKey := r.mapErrorToTranslationKey(errorMsg)
		if translationKey != "" {
			localizedErrors[i] = schema.T_Validation(locale, translationKey, nil)
		} else {
			// Fallback to original message if no mapping found
			localizedErrors[i] = errorMsg
		}
	}
	
	return localizedErrors
}

// mapErrorToTranslationKey maps validation error messages to translation keys
func (r *Runtime) mapErrorToTranslationKey(errorMsg string) string {
	// Map common English validation messages to translation keys
	errorMap := map[string]string{
		"This field is required":                    "required",
		"field is required":                        "required", 
		"required":                                 "required",
		"Must be a valid email address":           "invalidEmail",
		"Must be a valid email":                   "invalidEmail",
		"invalid email":                           "invalidEmail",
		"Passwords do not match":                  "passwordMismatch",
		"passwords must match":                    "passwordMismatch",
		"Must be a valid phone number":            "invalidPhone",
		"invalid phone":                           "invalidPhone",
		"Must be a valid URL":                     "invalidUrl",
		"invalid url":                             "invalidUrl",
		"Must be a valid date":                    "invalidDate",
		"invalid date":                            "invalidDate",
		"Must be a valid number":                  "invalidNumber",
		"invalid number":                          "invalidNumber",
		"Invalid format":                          "patternMismatch",
		"This value is already taken":             "uniqueViolation",
		"value already exists":                    "uniqueViolation",
	}
	
	// Check exact matches first
	if key, exists := errorMap[errorMsg]; exists {
		return key
	}
	
	// Check partial matches for parameterized messages
	lowerMsg := strings.ToLower(errorMsg)
	if strings.Contains(lowerMsg, "must be at least") && strings.Contains(lowerMsg, "character") {
		return "minLength"
	}
	if strings.Contains(lowerMsg, "must be no more than") && strings.Contains(lowerMsg, "character") {
		return "maxLength"
	}
	if strings.Contains(lowerMsg, "must be at least") && !strings.Contains(lowerMsg, "character") {
		return "numberMin"
	}
	if strings.Contains(lowerMsg, "must be no more than") && !strings.Contains(lowerMsg, "character") {
		return "numberMax"
	}
	
	// No mapping found
	return ""
}
