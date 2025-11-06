package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/niiniyare/erp/pkg/schema"
)

// EventType represents runtime events
type EventType string

const (
	EventChange EventType = "change" // Field value changed
	EventBlur   EventType = "blur"   // Field lost focus
	EventFocus  EventType = "focus"  // Field gained focus
	EventSubmit EventType = "submit" // Form submitted
)

// ValidationTiming defines when validation occurs
type ValidationTiming string

const (
	ValidateOnChange ValidationTiming = "change" // Validate immediately on change
	ValidateOnBlur   ValidationTiming = "blur"   // Validate when field loses focus
	ValidateOnSubmit ValidationTiming = "submit" // Validate only on form submission
	ValidateNever    ValidationTiming = "never"  // Never validate automatically
)

// Event represents a runtime event
type Event struct {
	Type      EventType   `json:"type"`
	Field     string      `json:"field"`
	Value     any `json:"value"`
	OldValue  any `json:"old_value"`
	Timestamp time.Time   `json:"timestamp"`
}

// EventCallback is a function that handles events
type EventCallback func(ctx context.Context, event *Event) error

// EventHandler handles runtime events
type EventHandler struct {
	runtime          *Runtime                      // Reference to runtime
	handlers         map[EventType][]EventCallback // Event handlers
	validationTiming ValidationTiming              // When to validate
	mu               sync.RWMutex                  // Concurrent access protection
}

// NewEventHandler creates a new event handler
func NewEventHandler() *EventHandler {
	return &EventHandler{
		handlers:         make(map[EventType][]EventCallback),
		validationTiming: ValidateOnBlur, // Default to blur validation
	}
}

// OnChange handles field value changes with enriched schema context
func (h *EventHandler) OnChange(ctx context.Context, event *Event) error {
	// 1. Check if field exists and is editable (from enricher)
	field := h.getField(event.Field)
	if field == nil {
		return fmt.Errorf("field %s not found in schema", event.Field)
	}

	if field.Runtime != nil && !field.Runtime.Editable {
		return fmt.Errorf("field %s is not editable: %s", event.Field, field.Runtime.Reason)
	}

	// 2. Update state
	if err := h.runtime.state.SetValue(event.Field, event.Value); err != nil {
		return fmt.Errorf("failed to update field value: %w", err)
	}

	// 3. Validate based on timing strategy
	if h.validationTiming == ValidateOnChange {
		errors := h.runtime.validateFieldUnlocked(ctx, event.Field, event.Value)
		h.runtime.state.SetErrors(event.Field, errors)
	}

	// 4. Skip conditional logic here to avoid deadlock
	// The runtime will handle this after the event processing

	// 5. Trigger custom event handlers
	return h.triggerCallbacks(ctx, EventChange, event)
}

// OnBlur handles field blur (focus lost)
func (h *EventHandler) OnBlur(ctx context.Context, event *Event) error {
	// 1. Mark field as touched
	h.runtime.state.Touch(event.Field)

	// 2. Validate on blur if using blur strategy
	if h.validationTiming == ValidateOnBlur {
		errors := h.runtime.validateFieldUnlocked(ctx, event.Field, event.Value)
		h.runtime.state.SetErrors(event.Field, errors)
	}

	// 3. Trigger custom event handlers
	return h.triggerCallbacks(ctx, EventBlur, event)
}

// OnFocus handles field focus (focus gained)
func (h *EventHandler) OnFocus(ctx context.Context, event *Event) error {
	// 1. Clear errors if desired (optional behavior)
	// h.runtime.state.ClearErrors(event.Field)

	// 2. Trigger custom event handlers
	return h.triggerCallbacks(ctx, EventFocus, event)
}

// OnSubmit handles form submission
func (h *EventHandler) OnSubmit(ctx context.Context) error {
	// 1. Validate all fields
	allErrors := h.runtime.validateCurrentStateUnlocked(ctx)

	// 2. Update state with all errors
	for field, errors := range allErrors {
		h.runtime.state.SetErrors(field, errors)
	}

	// 3. Check if valid
	if !h.runtime.state.IsValid() {
		return fmt.Errorf("form validation failed")
	}

	// 4. Skip conditional logic here to avoid deadlock
	// The runtime will handle this after the event processing

	// 5. Trigger submit handlers
	event := &Event{
		Type:      EventSubmit,
		Timestamp: time.Now(),
	}

	return h.triggerCallbacks(ctx, EventSubmit, event)
}

// Register adds a custom event handler
func (h *EventHandler) Register(eventType EventType, callback EventCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.handlers[eventType] = append(h.handlers[eventType], callback)
}

// Unregister removes event handlers (removes all handlers for the event type)
func (h *EventHandler) Unregister(eventType EventType) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.handlers, eventType)
}

// SetValidationTiming configures when validation occurs
func (h *EventHandler) SetValidationTiming(timing ValidationTiming) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.validationTiming = timing
}

// GetValidationTiming returns current validation timing strategy
func (h *EventHandler) GetValidationTiming() ValidationTiming {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.validationTiming
}

// triggerCallbacks executes all registered callbacks for an event type
func (h *EventHandler) triggerCallbacks(ctx context.Context, eventType EventType, event *Event) error {
	h.mu.RLock()
	callbacks := h.handlers[eventType]
	h.mu.RUnlock()

	for _, callback := range callbacks {
		if err := callback(ctx, event); err != nil {
			return fmt.Errorf("event callback failed: %w", err)
		}
	}

	return nil
}

// getField finds a field in the schema by name
func (h *EventHandler) getField(fieldName string) *schema.Field {
	if h.runtime == nil || h.runtime.schema == nil {
		return nil
	}

	for i := range h.runtime.schema.Fields {
		if h.runtime.schema.Fields[i].Name == fieldName {
			return &h.runtime.schema.Fields[i]
		}
	}

	return nil
}

// ValidateFieldWithTiming validates a field based on current timing strategy
func (h *EventHandler) ValidateFieldWithTiming(ctx context.Context, fieldName string, value any, eventType EventType) []string {
	// Check if we should validate based on timing and event type
	shouldValidate := false
	switch h.validationTiming {
	case ValidateOnChange:
		shouldValidate = eventType == EventChange
	case ValidateOnBlur:
		shouldValidate = eventType == EventBlur
	case ValidateOnSubmit:
		shouldValidate = eventType == EventSubmit
	case ValidateNever:
		shouldValidate = false
	}

	if !shouldValidate {
		return nil
	}

	return h.runtime.validateFieldUnlocked(ctx, fieldName, value)
}

// HandleBatchUpdate processes multiple field updates at once
func (h *EventHandler) HandleBatchUpdate(ctx context.Context, updates map[string]any) error {
	// Update all values first
	if err := h.runtime.state.UpdateValues(updates); err != nil {
		return fmt.Errorf("failed to update values: %w", err)
	}

	// Validate based on timing strategy
	if h.validationTiming == ValidateOnChange {
		for field, value := range updates {
			errors := h.runtime.validateFieldUnlocked(ctx, field, value)
			h.runtime.state.SetErrors(field, errors)
		}
	}

	// Skip conditional logic here to avoid deadlock
	// The runtime will handle this after the event processing

	// Trigger change events for each field
	for field, value := range updates {
		oldValue, _ := h.runtime.state.GetValue(field)
		event := &Event{
			Type:      EventChange,
			Field:     field,
			Value:     value,
			OldValue:  oldValue,
			Timestamp: time.Now(),
		}

		if err := h.triggerCallbacks(ctx, EventChange, event); err != nil {
			return fmt.Errorf("failed to trigger callback for field %s: %w", field, err)
		}
	}

	return nil
}

// DebouncedEventHandler wraps events with debouncing
type DebouncedEventHandler struct {
	handler *EventHandler
	timers  map[string]*time.Timer
	mu      sync.Mutex
}

// NewDebouncedEventHandler creates a debounced event handler
func NewDebouncedEventHandler(handler *EventHandler) *DebouncedEventHandler {
	return &DebouncedEventHandler{
		handler: handler,
		timers:  make(map[string]*time.Timer),
	}
}

// OnChangeDebounced handles change events with debouncing
func (d *DebouncedEventHandler) OnChangeDebounced(
	ctx context.Context,
	event *Event,
	delay time.Duration,
) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Cancel existing timer for this field
	if timer, exists := d.timers[event.Field]; exists {
		timer.Stop()
	}

	// Create new timer
	d.timers[event.Field] = time.AfterFunc(delay, func() {
		// Execute the actual handler after delay
		if err := d.handler.OnChange(ctx, event); err != nil {
			// Log error or handle as needed
			// In production, you might want to send this to an error handler
		}

		// Clean up timer
		d.mu.Lock()
		delete(d.timers, event.Field)
		d.mu.Unlock()
	})

	return nil
}

// EventStats provides statistics about events
type EventStats struct {
	TotalEvents    int               `json:"total_events"`
	EventsByType   map[EventType]int `json:"events_by_type"`
	LastEventTime  time.Time         `json:"last_event_time"`
	LastEventType  EventType         `json:"last_event_type"`
	LastEventField string            `json:"last_event_field"`
}

// EventTracker tracks event statistics
type EventTracker struct {
	stats *EventStats
	mu    sync.RWMutex
}

// NewEventTracker creates a new event tracker
func NewEventTracker() *EventTracker {
	return &EventTracker{
		stats: &EventStats{
			EventsByType: make(map[EventType]int),
		},
	}
}

// TrackEvent records an event for statistics
func (t *EventTracker) TrackEvent(event *Event) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.stats.TotalEvents++
	t.stats.EventsByType[event.Type]++
	t.stats.LastEventTime = event.Timestamp
	t.stats.LastEventType = event.Type
	t.stats.LastEventField = event.Field
}

// GetStats returns current event statistics
func (t *EventTracker) GetStats() *EventStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy
	statsCopy := &EventStats{
		TotalEvents:    t.stats.TotalEvents,
		EventsByType:   make(map[EventType]int),
		LastEventTime:  t.stats.LastEventTime,
		LastEventType:  t.stats.LastEventType,
		LastEventField: t.stats.LastEventField,
	}

	for eventType, count := range t.stats.EventsByType {
		statsCopy.EventsByType[eventType] = count
	}

	return statsCopy
}

// Reset clears all statistics
func (t *EventTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.stats = &EventStats{
		EventsByType: make(map[EventType]int),
	}
}
