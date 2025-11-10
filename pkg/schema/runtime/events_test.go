package runtime

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/niiniyare/erp/pkg/schema"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// EventHandlerTestSuite defines the test suite for EventHandler
type EventHandlerTestSuite struct {
	suite.Suite
	ctx     context.Context
	schema  *schema.Schema
	runtime *Runtime
	handler *EventHandler
}

// SetupTest runs before each test
func (s *EventHandlerTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.schema = createTestSchema()
	s.runtime = createTestRuntimeWithValidator()
	s.handler = s.runtime.events
}

// TearDownTest runs after each test
func (s *EventHandlerTestSuite) TearDownTest() {
	// Clean up if needed
	s.handler = nil
	s.runtime = nil
}

// TestNewEventHandler tests event handler initialization
func (s *EventHandlerTestSuite) TestNewEventHandler() {
	handler := NewEventHandler()

	require.NotNil(s.T(), handler, "NewEventHandler() should not return nil")
	require.NotNil(s.T(), handler.handlers, "Event handlers map should be initialized")
	require.Equal(s.T(), schema.ValidateOnBlur, handler.validationTiming, "Default validation timing should be ValidateOnBlur")
}

// TestOnChange tests field change event handling
func (s *EventHandlerTestSuite) TestOnChange() {
	// Initialize runtime
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	s.handler.SetValidationTiming(schema.ValidateOnChange)

	// Handle change event through runtime (not directly through handler)
	err = s.runtime.HandleFieldChange(s.ctx, "name", "Jane Doe")
	require.NoError(s.T(), err, "HandleFieldChange() should not return error")

	// Check value was updated in state
	value, exists := s.runtime.state.GetValue("name")
	require.True(s.T(), exists, "Value should exist")
	require.Equal(s.T(), "Jane Doe", value, "Value should be updated")

	// Check field is marked as dirty
	require.True(s.T(), s.runtime.state.IsDirty("name"), "Field should be marked as dirty after change")
}

// TestOnChange_ReadOnlyField tests change event on read-only field
func (s *EventHandlerTestSuite) TestOnChange_ReadOnlyField() {
	// Create schema with read-only field
	readonlySchema := &schema.Schema{
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

	runtime := NewRuntime(readonlySchema)

	// Initialize runtime
	err := runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Handle change event for read-only field (should fail)
	err = runtime.HandleFieldChange(s.ctx, "readonly_field", "new value")
	require.Error(s.T(), err, "HandleFieldChange() should fail for read-only field")

	// Check value was not updated
	value, exists := runtime.state.GetValue("readonly_field")
	if exists {
		require.NotEqual(s.T(), "new value", value, "Read-only field value should not be updated")
	}
}

// TestOnBlur tests field blur event handling
func (s *EventHandlerTestSuite) TestOnBlur() {
	// Initialize runtime
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	s.handler.SetValidationTiming(schema.ValidateOnBlur)

	// Handle blur event through runtime (not directly through handler)
	err = s.runtime.HandleFieldBlur(s.ctx, "email", "invalid-email")
	require.NoError(s.T(), err, "HandleFieldBlur() should not return error")

	// Check field is marked as touched
	require.True(s.T(), s.runtime.state.IsTouched("email"), "Field should be marked as touched after blur")

	// Check validation occurred (should have errors for invalid email)
	errors := s.runtime.state.GetErrors("email")
	require.NotEmpty(s.T(), errors, "Expected validation errors for invalid email on blur")
}

// TestOnSubmit tests submit event with valid data
func (s *EventHandlerTestSuite) TestOnSubmit() {
	// Initialize runtime with valid data
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Handle submit event
	err = s.handler.OnSubmit(s.ctx)
	require.NoError(s.T(), err, "OnSubmit() should not return error")

	// Check state is valid
	require.True(s.T(), s.runtime.state.IsValid(), "State should be valid after successful submit")
}

// TestOnSubmit_WithErrors tests submit event with validation errors
func (s *EventHandlerTestSuite) TestOnSubmit_WithErrors() {
	// Initialize runtime with valid data first
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Now set invalid data through direct state manipulation
	err = s.runtime.state.SetValue("name", "")
	require.NoError(s.T(), err, "Setting empty name should not fail")
	err = s.runtime.state.SetValue("email", "invalid-email")
	require.NoError(s.T(), err, "Setting invalid email should not fail")

	// Handle submit event
	err = s.runtime.HandleSubmit(s.ctx)
	require.Error(s.T(), err, "HandleSubmit() should fail with validation errors")

	// Check state has errors
	require.False(s.T(), s.runtime.state.IsValid(), "State should be invalid with validation errors")

	// Check specific errors exist
	nameErrors := s.runtime.state.GetErrors("name")
	require.NotEmpty(s.T(), nameErrors, "Expected validation errors for empty required name field")

	emailErrors := s.runtime.state.GetErrors("email")
	require.NotEmpty(s.T(), emailErrors, "Expected validation errors for invalid email")
}

// TestValidationTiming tests different validation timing modes
func (s *EventHandlerTestSuite) TestValidationTiming() {
	// Initialize runtime
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	tests := []struct {
		name   string
		timing schema.ValidationTiming
		event  schema.EventType
		hasErr bool
	}{
		{
			name:   "validate on change with change event",
			timing: schema.ValidateOnChange,
			event:  schema.EventChange,
			hasErr: true, // Invalid email should produce errors
		},
		{
			name:   "validate on change with blur event",
			timing: schema.ValidateOnChange,
			event:  schema.EventBlur,
			hasErr: false, // No validation on blur with ValidateOnChange
		},
		{
			name:   "validate on blur with blur event",
			timing: schema.ValidateOnBlur,
			event:  schema.EventBlur,
			hasErr: true, // Invalid email should produce errors
		},
		{
			name:   "validate on blur with change event",
			timing: schema.ValidateOnBlur,
			event:  schema.EventChange,
			hasErr: false, // No validation on change with ValidateOnBlur
		},
		{
			name:   "never validate with change event",
			timing: schema.ValidateNever,
			event:  schema.EventChange,
			hasErr: false, // No validation with ValidateNever
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Reset state
			s.runtime.state.Reset()
			s.runtime.state.ClearErrors("email")

			// Set validation timing
			s.handler.SetValidationTiming(tt.timing)

			// Create event with invalid email
			// Handle event based on type through runtime
			var err error
			switch tt.event {
			case schema.EventChange:
				err = s.runtime.HandleFieldChange(s.ctx, "email", "invalid-email")
			case schema.EventBlur:
				err = s.runtime.HandleFieldBlur(s.ctx, "email", "invalid-email")
			}

			require.NoError(s.T(), err, "Event handler should not return error")

			// Check if errors exist based on expectation
			errors := s.runtime.state.GetErrors("email")
			hasErrors := len(errors) > 0

			if tt.hasErr {
				require.True(s.T(), hasErrors, "Expected validation errors")
			} else {
				require.False(s.T(), hasErrors, "Expected no validation errors")
			}
		})
	}
}

// TestRegister tests event handler registration
func (s *EventHandlerTestSuite) TestRegister() {
	handler := NewEventHandler()

	// Track if callback was called
	callbackCalled := false
	var receivedEvent *schema.Event

	// Register callback
	handler.Register(schema.EventChange, func(ctx context.Context, event *schema.Event) error {
		callbackCalled = true
		receivedEvent = event
		return nil
	})

	// Create runtime and set handler
	runtime := NewRuntime(createTestSchema())
	runtime.events = handler

	err := runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Create event
	event := &schema.Event{
		Type:      schema.EventChange,
		Field:     "name",
		Value:     "Test Value",
		Timestamp: time.Now(),
	}

	// Trigger event
	err = handler.OnChange(s.ctx, event)
	require.NoError(s.T(), err, "OnChange() should not return error")

	// Check callback was called
	require.True(s.T(), callbackCalled, "Registered callback should be called")
	require.NotNil(s.T(), receivedEvent, "Callback should receive event")

	if receivedEvent != nil {
		require.Equal(s.T(), "name", receivedEvent.Field, "Event field should match")
		require.Equal(s.T(), "Test Value", receivedEvent.Value, "Event value should match")
	}
}

// TestHandleBatchUpdate tests batch field updates
func (s *EventHandlerTestSuite) TestHandleBatchUpdate() {
	// Initialize runtime
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	s.handler.SetValidationTiming(schema.ValidateOnChange)

	// Track callback calls
	changeEvents := 0
	s.handler.Register(schema.EventChange, func(ctx context.Context, event *schema.Event) error {
		changeEvents++
		return nil
	})

	// Batch update
	updates := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	// HandleBatchUpdate was removed - update fields individually
	for field, value := range updates {
		err := s.runtime.HandleFieldChange(s.ctx, field, value)
		require.NoError(s.T(), err, "HandleFieldChange() should not return error for field %s", field)
	}
	require.NoError(s.T(), err, "HandleBatchUpdate() should not return error")

	// Check all values were updated
	for field, expectedValue := range updates {
		value, exists := s.runtime.state.GetValue(field)
		require.True(s.T(), exists, "Value should exist for field %s", field)
		require.Equal(s.T(), expectedValue, value, "Value for %s should match", field)
	}

	// Check change events were triggered for each field
	require.Equal(s.T(), len(updates), changeEvents, "Expected change event for each field")
}

// TestDebouncedEventHandler tests debounced event handling
func (s *EventHandlerTestSuite) TestDebouncedEventHandler() {
	// Debounced handler functionality was intentionally removed
	s.T().Skip("DebouncedEventHandler functionality was removed from design")
}

// TestEventTracker tests event tracking functionality
func (s *EventHandlerTestSuite) TestEventTracker() {
	tracker := NewEventTracker()

	// Initially no events
	stats := tracker.GetStats()
	require.Equal(s.T(), 0, stats.TotalEvents, "Should start with 0 events")

	// Track some events
	events := []*schema.Event{
		{Type: schema.EventChange, Field: "name", Timestamp: time.Now()},
		{Type: schema.EventChange, Field: "email", Timestamp: time.Now()},
		{Type: schema.EventBlur, Field: "name", Timestamp: time.Now()},
		{Type: schema.EventSubmit, Timestamp: time.Now()},
	}

	for _, event := range events {
		tracker.TrackEvent(event)
	}

	// Check stats
	stats = tracker.GetStats()
	require.Equal(s.T(), 4, stats.TotalEvents, "Should have 4 total events")
	require.Equal(s.T(), 2, stats.EventsByType[schema.EventChange], "Should have 2 change events")
	require.Equal(s.T(), 1, stats.EventsByType[schema.EventBlur], "Should have 1 blur event")
	require.Equal(s.T(), 1, stats.EventsByType[schema.EventSubmit], "Should have 1 submit event")
	require.Equal(s.T(), schema.EventSubmit, stats.LastEventType, "Last event type should be submit")

	// Reset and check
	tracker.Reset()
	stats = tracker.GetStats()
	require.Equal(s.T(), 0, stats.TotalEvents, "Should have 0 events after reset")
}

// TestGetValidationTiming tests getting validation timing
func (s *EventHandlerTestSuite) TestGetValidationTiming() {
	handler := NewEventHandler()

	// Check default timing
	require.Equal(s.T(), schema.ValidateOnBlur, handler.GetValidationTiming(), "Default validation timing should be ValidateOnBlur")

	// Set and check different timings
	timings := []schema.ValidationTiming{
		schema.ValidateOnChange,
		schema.ValidateOnBlur,
		schema.ValidateOnSubmit,
		schema.ValidateNever,
	}

	for _, timing := range timings {
		handler.SetValidationTiming(timing)
		require.Equal(s.T(), timing, handler.GetValidationTiming(), "Timing should match set value")
	}
}

// TestUnregister tests event handler unregistration
func (s *EventHandlerTestSuite) TestUnregister() {
	handler := NewEventHandler()

	// Register some callbacks
	handler.Register(schema.EventChange, func(ctx context.Context, event *schema.Event) error {
		return nil
	})
	handler.Register(schema.EventChange, func(ctx context.Context, event *schema.Event) error {
		return nil
	})

	// Verify callbacks exist
	handler.mu.RLock()
	changeHandlers := len(handler.handlers[schema.EventChange])
	handler.mu.RUnlock()

	require.Equal(s.T(), 2, changeHandlers, "Should have 2 change handlers")

	// Unregister
	handler.Unregister(schema.EventChange)

	// Verify callbacks are removed
	handler.mu.RLock()
	changeHandlers = len(handler.handlers[schema.EventChange])
	handler.mu.RUnlock()

	require.Equal(s.T(), 0, changeHandlers, "Should have 0 change handlers after unregister")
}

// TestErrorHandling tests error propagation from callbacks
func (s *EventHandlerTestSuite) TestErrorHandling() {
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Register callback that returns error
	s.handler.Register(schema.EventChange, func(ctx context.Context, event *schema.Event) error {
		return fmt.Errorf("callback error")
	})

	// Create event
	event := &schema.Event{
		Type:      schema.EventChange,
		Field:     "name",
		Value:     "Test Value",
		Timestamp: time.Now(),
	}

	// Handle event (should propagate callback error)
	err = s.handler.OnChange(s.ctx, event)
	require.Error(s.T(), err, "OnChange() should return error from callback")
}

// TestNonExistentField tests handling of non-existent field
func (s *EventHandlerTestSuite) TestNonExistentField() {
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Handle change for non-existent field (should fail)
	err = s.runtime.HandleFieldChange(s.ctx, "non_existent_field", "test value")
	require.Error(s.T(), err, "HandleFieldChange() should fail for non-existent field")
}

// Helper method to create a test schema

// TestEventHandlerTestSuite runs the test suite
func TestEventHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(EventHandlerTestSuite))
}
