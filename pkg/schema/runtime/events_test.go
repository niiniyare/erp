package runtime

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/niiniyare/erp/pkg/schema"
)

func TestNewEventHandler(t *testing.T) {
	handler := NewEventHandler()

	if handler == nil {
		t.Fatal("NewEventHandler() returned nil")
	}

	if handler.handlers == nil {
		t.Error("Event handlers map not initialized")
	}

	if handler.validationTiming != ValidateOnBlur {
		t.Error("Default validation timing should be ValidateOnBlur")
	}
}

func TestEventHandler_OnChange(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	// Initialize runtime
	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	handler := runtime.events
	handler.SetValidationTiming(ValidateOnChange)

	// Create change event
	event := &Event{
		Type:      EventChange,
		Field:     "name",
		Value:     "John Doe",
		OldValue:  "",
		Timestamp: time.Now(),
	}

	// Handle change event
	err = handler.OnChange(ctx, event)
	if err != nil {
		t.Errorf("OnChange() error = %v", err)
	}

	// Check value was updated in state
	value, exists := runtime.state.GetValue("name")
	if !exists || value != "John Doe" {
		t.Errorf("Expected value 'John Doe', got %v", value)
	}

	// Check field is marked as dirty
	if !runtime.state.IsDirty("name") {
		t.Error("Field should be marked as dirty after change")
	}
}

func TestEventHandler_OnChange_ReadOnlyField(t *testing.T) {
	// Create schema with read-only field
	schema := &schema.Schema{
		ID:    "readonly_test",
		Title: "Read-only Test",
		Fields: []schema.Field{
			{
				Name:     "readonly_field",
				Type:     schema.FieldText,
				Label:    "Read-only Field",
				Required: false,
				Runtime: &schema.FieldRuntime{
					Visible:  true,
					Editable: false,
					Reason:   "User lacks edit permission",
				},
			},
		},
	}

	runtime := NewRuntime(schema)
	ctx := context.Background()

	// Initialize runtime
	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	handler := runtime.events

	// Create change event for read-only field
	event := &Event{
		Type:      EventChange,
		Field:     "readonly_field",
		Value:     "new value",
		OldValue:  "",
		Timestamp: time.Now(),
	}

	// Handle change event (should fail)
	err = handler.OnChange(ctx, event)
	if err == nil {
		t.Error("OnChange() should fail for read-only field")
	}

	// Check value was not updated
	value, exists := runtime.state.GetValue("readonly_field")
	if exists && value == "new value" {
		t.Error("Read-only field value should not be updated")
	}
}

func TestEventHandler_OnBlur(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	// Initialize runtime
	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	handler := runtime.events
	handler.SetValidationTiming(ValidateOnBlur)

	// Create blur event with invalid email
	event := &Event{
		Type:      EventBlur,
		Field:     "email",
		Value:     "invalid-email",
		Timestamp: time.Now(),
	}

	// Handle blur event
	err = handler.OnBlur(ctx, event)
	if err != nil {
		t.Errorf("OnBlur() error = %v", err)
	}

	// Check field is marked as touched
	if !runtime.state.IsTouched("email") {
		t.Error("Field should be marked as touched after blur")
	}

	// Check validation occurred (should have errors for invalid email)
	errors := runtime.state.GetErrors("email")
	if len(errors) == 0 {
		t.Error("Expected validation errors for invalid email on blur")
	}
}

func TestEventHandler_OnSubmit(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	// Initialize runtime with valid data
	err := runtime.Initialize(ctx, map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	handler := runtime.events

	// Handle submit event
	err = handler.OnSubmit(ctx)
	if err != nil {
		t.Errorf("OnSubmit() error = %v", err)
	}

	// Check state is valid
	if !runtime.state.IsValid() {
		t.Error("State should be valid after successful submit")
	}
}

func TestEventHandler_OnSubmit_WithErrors(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	// Initialize runtime with invalid data
	err := runtime.Initialize(ctx, map[string]interface{}{
		"name":  "",               // Required field empty
		"email": "invalid-email",  // Invalid email
	})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	handler := runtime.events

	// Handle submit event
	err = handler.OnSubmit(ctx)
	if err == nil {
		t.Error("OnSubmit() should fail with validation errors")
	}

	// Check state has errors
	if runtime.state.IsValid() {
		t.Error("State should be invalid with validation errors")
	}

	// Check specific errors exist
	nameErrors := runtime.state.GetErrors("name")
	if len(nameErrors) == 0 {
		t.Error("Expected validation errors for empty required name field")
	}

	emailErrors := runtime.state.GetErrors("email")
	if len(emailErrors) == 0 {
		t.Error("Expected validation errors for invalid email")
	}
}

func TestEventHandler_ValidationTiming(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	// Initialize runtime
	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	handler := runtime.events

	tests := []struct {
		name   string
		timing ValidationTiming
		event  EventType
		hasErr bool
	}{
		{
			name:   "validate on change with change event",
			timing: ValidateOnChange,
			event:  EventChange,
			hasErr: true, // Invalid email should produce errors
		},
		{
			name:   "validate on change with blur event",
			timing: ValidateOnChange,
			event:  EventBlur,
			hasErr: false, // No validation on blur with ValidateOnChange
		},
		{
			name:   "validate on blur with blur event",
			timing: ValidateOnBlur,
			event:  EventBlur,
			hasErr: true, // Invalid email should produce errors
		},
		{
			name:   "validate on blur with change event",
			timing: ValidateOnBlur,
			event:  EventChange,
			hasErr: false, // No validation on change with ValidateOnBlur
		},
		{
			name:   "never validate with change event",
			timing: ValidateNever,
			event:  EventChange,
			hasErr: false, // No validation with ValidateNever
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			runtime.state.Reset()
			runtime.state.ClearErrors("email")

			// Set validation timing
			handler.SetValidationTiming(tt.timing)

			// Create event with invalid email
			event := &Event{
				Type:      tt.event,
				Field:     "email",
				Value:     "invalid-email",
				Timestamp: time.Now(),
			}

			// Handle event based on type
			var err error
			switch tt.event {
			case EventChange:
				err = handler.OnChange(ctx, event)
			case EventBlur:
				err = handler.OnBlur(ctx, event)
			}

			if err != nil {
				t.Errorf("Event handler error: %v", err)
			}

			// Check if errors exist based on expectation
			errors := runtime.state.GetErrors("email")
			hasErrors := len(errors) > 0

			if hasErrors != tt.hasErr {
				t.Errorf("Expected hasErrors = %v, got %v (errors: %v)", tt.hasErr, hasErrors, errors)
			}
		})
	}
}

func TestEventHandler_Register(t *testing.T) {
	handler := NewEventHandler()

	// Track if callback was called
	callbackCalled := false
	var receivedEvent *Event

	// Register callback
	handler.Register(EventChange, func(ctx context.Context, event *Event) error {
		callbackCalled = true
		receivedEvent = event
		return nil
	})

	// Create runtime and set handler
	runtime := NewRuntime(createTestSchema())
	runtime.events = handler
	handler.runtime = runtime

	ctx := context.Background()
	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Create event
	event := &Event{
		Type:      EventChange,
		Field:     "name",
		Value:     "Test Value",
		Timestamp: time.Now(),
	}

	// Trigger event
	err = handler.OnChange(ctx, event)
	if err != nil {
		t.Errorf("OnChange() error = %v", err)
	}

	// Check callback was called
	if !callbackCalled {
		t.Error("Registered callback was not called")
	}

	if receivedEvent == nil {
		t.Error("Callback did not receive event")
	} else {
		if receivedEvent.Field != "name" {
			t.Errorf("Expected field 'name', got '%s'", receivedEvent.Field)
		}
		if receivedEvent.Value != "Test Value" {
			t.Errorf("Expected value 'Test Value', got '%v'", receivedEvent.Value)
		}
	}
}

func TestEventHandler_HandleBatchUpdate(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	// Initialize runtime
	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	handler := runtime.events
	handler.SetValidationTiming(ValidateOnChange)

	// Track callback calls
	changeEvents := 0
	handler.Register(EventChange, func(ctx context.Context, event *Event) error {
		changeEvents++
		return nil
	})

	// Batch update
	updates := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	err = handler.HandleBatchUpdate(ctx, updates)
	if err != nil {
		t.Errorf("HandleBatchUpdate() error = %v", err)
	}

	// Check all values were updated
	for field, expectedValue := range updates {
		value, exists := runtime.state.GetValue(field)
		if !exists {
			t.Errorf("Value not set for field %s", field)
		} else if value != expectedValue {
			t.Errorf("Value for %s = %v, want %v", field, value, expectedValue)
		}
	}

	// Check change events were triggered for each field
	if changeEvents != len(updates) {
		t.Errorf("Expected %d change events, got %d", len(updates), changeEvents)
	}
}

func TestDebouncedEventHandler(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Create debounced handler
	debouncedHandler := NewDebouncedEventHandler(runtime.events)

	// Track if validation occurred
	validationOccurred := false
	runtime.events.Register(EventChange, func(ctx context.Context, event *Event) error {
		validationOccurred = true
		return nil
	})

	// Create event
	event := &Event{
		Type:      EventChange,
		Field:     "name",
		Value:     "Test Value",
		Timestamp: time.Now(),
	}

	// Handle debounced change
	err = debouncedHandler.OnChangeDebounced(ctx, event, 50*time.Millisecond)
	if err != nil {
		t.Errorf("OnChangeDebounced() error = %v", err)
	}

	// Check validation hasn't occurred immediately
	if validationOccurred {
		t.Error("Validation should not occur immediately with debouncing")
	}

	// Wait for debounce period
	time.Sleep(100 * time.Millisecond)

	// Check validation occurred after debounce
	if !validationOccurred {
		t.Error("Validation should occur after debounce period")
	}
}

func TestEventTracker(t *testing.T) {
	tracker := NewEventTracker()

	// Initially no events
	stats := tracker.GetStats()
	if stats.TotalEvents != 0 {
		t.Errorf("Expected 0 total events, got %d", stats.TotalEvents)
	}

	// Track some events
	events := []*Event{
		{Type: EventChange, Field: "name", Timestamp: time.Now()},
		{Type: EventChange, Field: "email", Timestamp: time.Now()},
		{Type: EventBlur, Field: "name", Timestamp: time.Now()},
		{Type: EventSubmit, Timestamp: time.Now()},
	}

	for _, event := range events {
		tracker.TrackEvent(event)
	}

	// Check stats
	stats = tracker.GetStats()
	if stats.TotalEvents != 4 {
		t.Errorf("Expected 4 total events, got %d", stats.TotalEvents)
	}

	if stats.EventsByType[EventChange] != 2 {
		t.Errorf("Expected 2 change events, got %d", stats.EventsByType[EventChange])
	}

	if stats.EventsByType[EventBlur] != 1 {
		t.Errorf("Expected 1 blur event, got %d", stats.EventsByType[EventBlur])
	}

	if stats.EventsByType[EventSubmit] != 1 {
		t.Errorf("Expected 1 submit event, got %d", stats.EventsByType[EventSubmit])
	}

	if stats.LastEventType != EventSubmit {
		t.Errorf("Expected last event type to be submit, got %s", stats.LastEventType)
	}

	// Reset and check
	tracker.Reset()
	stats = tracker.GetStats()
	if stats.TotalEvents != 0 {
		t.Errorf("Expected 0 total events after reset, got %d", stats.TotalEvents)
	}
}

func TestEventHandler_GetValidationTiming(t *testing.T) {
	handler := NewEventHandler()

	// Check default timing
	if handler.GetValidationTiming() != ValidateOnBlur {
		t.Error("Default validation timing should be ValidateOnBlur")
	}

	// Set and check different timings
	timings := []ValidationTiming{
		ValidateOnChange,
		ValidateOnBlur,
		ValidateOnSubmit,
		ValidateNever,
	}

	for _, timing := range timings {
		handler.SetValidationTiming(timing)
		if handler.GetValidationTiming() != timing {
			t.Errorf("Expected timing %s, got %s", timing, handler.GetValidationTiming())
		}
	}
}

func TestEventHandler_Unregister(t *testing.T) {
	handler := NewEventHandler()

	// Register some callbacks
	handler.Register(EventChange, func(ctx context.Context, event *Event) error {
		return nil
	})
	handler.Register(EventChange, func(ctx context.Context, event *Event) error {
		return nil
	})

	// Verify callbacks exist
	handler.mu.RLock()
	changeHandlers := len(handler.handlers[EventChange])
	handler.mu.RUnlock()

	if changeHandlers != 2 {
		t.Errorf("Expected 2 change handlers, got %d", changeHandlers)
	}

	// Unregister
	handler.Unregister(EventChange)

	// Verify callbacks are removed
	handler.mu.RLock()
	changeHandlers = len(handler.handlers[EventChange])
	handler.mu.RUnlock()

	if changeHandlers != 0 {
		t.Errorf("Expected 0 change handlers after unregister, got %d", changeHandlers)
	}
}

func TestEventHandler_ErrorHandling(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	handler := runtime.events

	// Register callback that returns error
	handler.Register(EventChange, func(ctx context.Context, event *Event) error {
		return fmt.Errorf("callback error")
	})

	// Create event
	event := &Event{
		Type:      EventChange,
		Field:     "name",
		Value:     "Test Value",
		Timestamp: time.Now(),
	}

	// Handle event (should propagate callback error)
	err = handler.OnChange(ctx, event)
	if err == nil {
		t.Error("OnChange() should return error from callback")
	}
}

func TestEventHandler_NonExistentField(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	handler := runtime.events

	// Create event for non-existent field
	event := &Event{
		Type:      EventChange,
		Field:     "non_existent_field",
		Value:     "test value",
		Timestamp: time.Now(),
	}

	// Handle event (should fail)
	err = handler.OnChange(ctx, event)
	if err == nil {
		t.Error("OnChange() should fail for non-existent field")
	}
}