package runtime

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/niiniyare/erp/pkg/schema"
	"github.com/niiniyare/erp/pkg/schema/validate"
)

// Runtime manages the execution of an enriched schema
type Runtime struct {
	schema    *schema.Schema      // Enriched schema with runtime context
	state     *State              // Current form state
	validator *validate.Validator // Runtime validator
	events    *EventHandler       // Event handling
	mu        sync.RWMutex        // Concurrent access protection
}

// NewRuntime creates a new runtime instance for an enriched schema
func NewRuntime(enrichedSchema *schema.Schema) *Runtime {
	return &Runtime{
		schema:    enrichedSchema,
		state:     NewState(),
		validator: validate.NewValidator(nil), // No database for basic runtime
		events:    NewEventHandler(),
	}
}

// Initialize prepares the runtime with initial data and validates the schema
func (r *Runtime) Initialize(ctx context.Context, initialData map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Set initial values from enriched schema defaults
	if err := r.state.Initialize(r.schema, initialData); err != nil {
		return fmt.Errorf("failed to initialize state: %w", err)
	}

	// Set up event handler with runtime reference
	r.events.runtime = r

	// Run initial validation using schema validation rules
	err := r.validator.ValidateData(ctx, r.schema, r.state.GetAll())
	if err != nil {
		return fmt.Errorf("failed to validate initial data: %w", err)
	}

	return nil
}

// GetState returns the current runtime state
func (r *Runtime) GetState() *State {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}

// GetSchema returns the enriched schema
func (r *Runtime) GetSchema() *schema.Schema {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.schema
}

// HandleFieldChange processes a field value change
func (r *Runtime) HandleFieldChange(ctx context.Context, fieldName string, value any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get old value for event
	oldValue, _ := r.state.GetValue(fieldName)

	// Create change event
	event := &Event{
		Type:      EventChange,
		Field:     fieldName,
		Value:     value,
		OldValue:  oldValue,
		Timestamp: time.Now(),
	}

	// Handle the event
	if err := r.events.OnChange(ctx, event); err != nil {
		return err
	}

	// Apply conditional logic after event handling (without additional locking)
	return r.applyConditionalLogicUnlocked(ctx)
}

// HandleFieldBlur processes a field blur event
func (r *Runtime) HandleFieldBlur(ctx context.Context, fieldName string, value any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create blur event
	event := &Event{
		Type:      EventBlur,
		Field:     fieldName,
		Value:     value,
		Timestamp: time.Now(),
	}

	// Handle the event
	return r.events.OnBlur(ctx, event)
}

// HandleFieldFocus processes a field focus event
func (r *Runtime) HandleFieldFocus(ctx context.Context, fieldName string, value any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create focus event
	event := &Event{
		Type:      EventFocus,
		Field:     fieldName,
		Value:     value,
		Timestamp: time.Now(),
	}

	// Handle the event
	return r.events.OnFocus(ctx, event)
}

// HandleSubmit processes form submission
func (r *Runtime) HandleSubmit(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Handle submit event
	return r.events.OnSubmit(ctx)
}

// ValidateField validates a single field value
func (r *Runtime) ValidateField(ctx context.Context, fieldName string, value any) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.validateFieldUnlocked(ctx, fieldName, value)
}

// validateFieldUnlocked validates a single field value without locking
func (r *Runtime) validateFieldUnlocked(ctx context.Context, fieldName string, value any) []string {
	// Find the field in our enriched schema
	var field *schema.Field
	for i := range r.schema.Fields {
		if r.schema.Fields[i].Name == fieldName {
			field = &r.schema.Fields[i]
			break
		}
	}

	if field == nil {
		return []string{"field not found"}
	}

	// Check if field is visible (enricher sets this)
	if field.Runtime != nil && !field.Runtime.Visible {
		return nil // Don't validate invisible fields
	}

	// Use schema validator to validate the field
	return r.validator.ValidateField(ctx, &fieldAdapter{field}, value, true)
}

// ValidateCurrentState validates all fields using current state
func (r *Runtime) ValidateCurrentState(ctx context.Context) map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.validateCurrentStateUnlocked(ctx)
}

// validateCurrentStateUnlocked validates all fields using current state without locking
func (r *Runtime) validateCurrentStateUnlocked(ctx context.Context) map[string][]string {
	allErrors := make(map[string][]string)
	data := r.state.GetAll()

	// Validate each field in the schema
	for i := range r.schema.Fields {
		field := &r.schema.Fields[i]

		// Skip invisible fields
		if field.Runtime != nil && !field.Runtime.Visible {
			continue
		}

		value, exists := data[field.Name]
		if errors := r.validator.ValidateField(ctx, &fieldAdapter{field}, value, exists); len(errors) > 0 {
			allErrors[field.Name] = errors
		}
	}

	return allErrors
}

// ApplyConditionalLogic evaluates and applies conditional rules
func (r *Runtime) ApplyConditionalLogic(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.applyConditionalLogicUnlocked(ctx)
}

// applyConditionalLogicUnlocked evaluates and applies conditional rules without locking
func (r *Runtime) applyConditionalLogicUnlocked(ctx context.Context) error {
	data := r.state.GetAll()

	for i := range r.schema.Fields {
		field := &r.schema.Fields[i]

		// Check visibility conditions only if field has conditional logic
		// Don't override explicit runtime visibility settings from enricher
		if field.Conditional != nil {
			visible, err := field.IsVisible(ctx, data)
			if err != nil {
				// Log error but don't fail - graceful degradation
				continue
			}

			// Update runtime state if field has runtime info
			if field.Runtime != nil {
				field.Runtime.Visible = visible
			}
		}
	}

	return nil
}

// RegisterEventHandler registers a custom event handler
func (r *Runtime) RegisterEventHandler(eventType EventType, callback EventCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events.Register(eventType, callback)
}

// SetValidationTiming configures when validation occurs
func (r *Runtime) SetValidationTiming(timing ValidationTiming) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events.validationTiming = timing
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

		// Wait for the debounce period
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

// IsValid checks if the entire form is currently valid
func (r *Runtime) IsValid() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.state.IsValid()
}

// GetErrors returns all current validation errors
func (r *Runtime) GetErrors() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.state.GetAllErrors()
}

// Reset clears all state back to initial values
func (r *Runtime) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.state.Reset()
}

// GetFieldValue gets the current value of a specific field
func (r *Runtime) GetFieldValue(fieldName string) (any, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.state.GetValue(fieldName)
}

// SetFieldValue sets the value of a specific field
func (r *Runtime) SetFieldValue(fieldName string, value any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.state.SetValue(fieldName, value)
}

// IsDirty checks if any field has been modified from initial values
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

// IsFieldTouched checks if a specific field has been interacted with
func (r *Runtime) IsFieldTouched(fieldName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.state.IsTouched(fieldName)
}

// GetAllData returns all current form data
func (r *Runtime) GetAllData() map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.state.GetAll()
}

// UpdateSchema updates the runtime schema (useful for dynamic schema changes)
func (r *Runtime) UpdateSchema(newSchema *schema.Schema) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.schema = newSchema
	return nil
}

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
}

// GetStats returns runtime statistics
func (r *Runtime) GetStats() *RuntimeStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := &RuntimeStats{
		FieldCount:    len(r.schema.Fields),
		TouchedFields: r.state.GetTouchedCount(),
		DirtyFields:   r.state.GetDirtyCount(),
		ErrorCount:    r.state.GetErrorCount(),
		IsValid:       r.state.IsValid(),
		LastActivity:  time.Now(),
	}

	// Count visible fields
	for i := range r.schema.Fields {
		field := &r.schema.Fields[i]
		if field.Runtime == nil || field.Runtime.Visible {
			stats.VisibleFields++
		}
	}

	return stats
}

// Adapter types to make schema types compatible with validator interfaces

// schemaAdapter adapts schema.Schema to validate.SchemaInterface
type schemaAdapter struct {
	*schema.Schema
}

func (s *schemaAdapter) GetID() string {
	return s.ID
}

func (s *schemaAdapter) GetType() string {
	return string(s.Type)
}

func (s *schemaAdapter) GetTitle() string {
	return s.Title
}

func (s *schemaAdapter) GetFields() []validate.FieldInterface {
	fields := make([]validate.FieldInterface, len(s.Fields))
	for i := range s.Fields {
		fields[i] = &fieldAdapter{&s.Fields[i]}
	}
	return fields
}

func (s *schemaAdapter) GetValidation() any {
	return s.Validation
}

// fieldAdapter adapts schema.Field to validate.FieldInterface
type fieldAdapter struct {
	*schema.Field
}

func (f *fieldAdapter) GetName() string {
	return f.Name
}

func (f *fieldAdapter) GetType() validate.FieldType {
	return validate.FieldType(f.Type)
}

func (f *fieldAdapter) GetRequired() bool {
	return f.Required
}

func (f *fieldAdapter) GetValidation() *validate.FieldValidation {
	if f.Validation == nil {
		return nil
	}

	// Convert schema validation to validator validation
	return &validate.FieldValidation{
		MinLength: f.Validation.MinLength,
		MaxLength: f.Validation.MaxLength,
		Pattern:   f.Validation.Pattern,
		Format:    f.Validation.Format,
		Min:       f.Validation.Min,
		Max:       f.Validation.Max,
		Step:      f.Validation.Step,
		Custom:    f.Validation.Custom,
	}
}

func (f *fieldAdapter) GetOptions() []validate.Option {
	if f.Options == nil {
		return nil
	}

	options := make([]validate.Option, len(f.Options))
	for i, opt := range f.Options {
		options[i] = validate.Option{
			Value: opt.Value,
			Label: opt.Label,
		}
	}
	return options
}

func (f *fieldAdapter) GetConfig() map[string]any {
	// Build config from field properties
	config := make(map[string]any)

	if f.Config != nil {
		// Copy existing config
		maps.Copy(config, f.Config)
	}

	// Add common field properties as config
	config["label"] = f.Label
	config["placeholder"] = f.Placeholder
	config["help"] = f.Help
	config["hidden"] = f.Hidden

	return config
}
