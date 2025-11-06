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
	s.schema = s.createTestSchema()
	s.runtime = NewRuntime(s.schema)
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
	require.Equal(s.T(), ValidateOnBlur, handler.validationTiming, "Default validation timing should be ValidateOnBlur")
}

// TestOnChange tests field change event handling
func (s *EventHandlerTestSuite) TestOnChange() {
	// Initialize runtime
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	s.handler.SetValidationTiming(ValidateOnChange)

	// Create change event
	event := &Event{
		Type:      EventChange,
		Field:     "name",
		Value:     "Jane Doe",
		OldValue:  "John Doe",
		Timestamp: time.Now(),
	}

	// Handle change event
	err = s.handler.OnChange(s.ctx, event)
	require.NoError(s.T(), err, "OnChange() should not return error")

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
	err = handler.OnChange(s.ctx, event)
	require.Error(s.T(), err, "OnChange() should fail for read-only field")

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

	s.handler.SetValidationTiming(ValidateOnBlur)

	// Create blur event with invalid email
	event := &Event{
		Type:      EventBlur,
		Field:     "email",
		Value:     "invalid-email",
		Timestamp: time.Now(),
	}

	// Handle blur event
	err = s.handler.OnBlur(s.ctx, event)
	require.NoError(s.T(), err, "OnBlur() should not return error")

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
	err = s.handler.OnSubmit(s.ctx)
	require.Error(s.T(), err, "OnSubmit() should fail with validation errors")

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
		s.Run(tt.name, func() {
			// Reset state
			s.runtime.state.Reset()
			s.runtime.state.ClearErrors("email")

			// Set validation timing
			s.handler.SetValidationTiming(tt.timing)

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
				err = s.handler.OnChange(s.ctx, event)
			case EventBlur:
				err = s.handler.OnBlur(s.ctx, event)
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
	var receivedEvent *Event

	// Register callback
	handler.Register(EventChange, func(ctx context.Context, event *Event) error {
		callbackCalled = true
		receivedEvent = event
		return nil
	})

	// Create runtime and set handler
	runtime := NewRuntime(s.createTestSchema())
	runtime.events = handler
	handler.runtime = runtime

	err := runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Create event
	event := &Event{
		Type:      EventChange,
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

	s.handler.SetValidationTiming(ValidateOnChange)

	// Track callback calls
	changeEvents := 0
	s.handler.Register(EventChange, func(ctx context.Context, event *Event) error {
		changeEvents++
		return nil
	})

	// Batch update
	updates := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	err = s.handler.HandleBatchUpdate(s.ctx, updates)
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
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Create debounced handler
	debouncedHandler := NewDebouncedEventHandler(s.runtime.events)

	// Track if validation occurred
	validationOccurred := false
	s.runtime.events.Register(EventChange, func(ctx context.Context, event *Event) error {
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
	err = debouncedHandler.OnChangeDebounced(s.ctx, event, 50*time.Millisecond)
	require.NoError(s.T(), err, "OnChangeDebounced() should not return error")

	// Check validation hasn't occurred immediately
	require.False(s.T(), validationOccurred, "Validation should not occur immediately with debouncing")

	// Wait for debounce period
	time.Sleep(100 * time.Millisecond)

	// Check validation occurred after debounce
	require.True(s.T(), validationOccurred, "Validation should occur after debounce period")
}

// TestEventTracker tests event tracking functionality
func (s *EventHandlerTestSuite) TestEventTracker() {
	tracker := NewEventTracker()

	// Initially no events
	stats := tracker.GetStats()
	require.Equal(s.T(), 0, stats.TotalEvents, "Should start with 0 events")

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
	require.Equal(s.T(), 4, stats.TotalEvents, "Should have 4 total events")
	require.Equal(s.T(), 2, stats.EventsByType[EventChange], "Should have 2 change events")
	require.Equal(s.T(), 1, stats.EventsByType[EventBlur], "Should have 1 blur event")
	require.Equal(s.T(), 1, stats.EventsByType[EventSubmit], "Should have 1 submit event")
	require.Equal(s.T(), EventSubmit, stats.LastEventType, "Last event type should be submit")

	// Reset and check
	tracker.Reset()
	stats = tracker.GetStats()
	require.Equal(s.T(), 0, stats.TotalEvents, "Should have 0 events after reset")
}

// TestGetValidationTiming tests getting validation timing
func (s *EventHandlerTestSuite) TestGetValidationTiming() {
	handler := NewEventHandler()

	// Check default timing
	require.Equal(s.T(), ValidateOnBlur, handler.GetValidationTiming(), "Default validation timing should be ValidateOnBlur")

	// Set and check different timings
	timings := []ValidationTiming{
		ValidateOnChange,
		ValidateOnBlur,
		ValidateOnSubmit,
		ValidateNever,
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

	require.Equal(s.T(), 2, changeHandlers, "Should have 2 change handlers")

	// Unregister
	handler.Unregister(EventChange)

	// Verify callbacks are removed
	handler.mu.RLock()
	changeHandlers = len(handler.handlers[EventChange])
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
	s.handler.Register(EventChange, func(ctx context.Context, event *Event) error {
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

	// Create event for non-existent field
	event := &Event{
		Type:      EventChange,
		Field:     "non_existent_field",
		Value:     "test value",
		Timestamp: time.Now(),
	}

	// Handle event (should fail)
	err = s.handler.OnChange(s.ctx, event)
	require.Error(s.T(), err, "OnChange() should fail for non-existent field")
}

// Helper method to create a test schema
func (s *EventHandlerTestSuite) createTestSchema() *schema.Schema {
	return &schema.Schema{
		ID:    "test_schema",
		Title: "Test Schema",
		Fields: []schema.Field{
			{
				Name:     "name",
				Type:     schema.FieldText,
				Label:    "Full Name",
				Required: true,
				Runtime: &schema.FieldRuntime{
					Visible:  true,
					Editable: true,
					Reason:   "",
				},
			},
			{
				Name:     "email",
				Type:     schema.FieldEmail,
				Label:    "Email Address",
				Required: true,
				Runtime: &schema.FieldRuntime{
					Visible:  true,
					Editable: true,
					Reason:   "",
				},
			},
			{
				Name:     "age",
				Type:     schema.FieldNumber,
				Label:    "Age",
				Required: false,
				Runtime: &schema.FieldRuntime{
					Visible:  true,
					Editable: true,
					Reason:   "",
				},
			},
		},
	}
}

// TestEventHandlerTestSuite runs the test suite
func TestEventHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(EventHandlerTestSuite))
}
