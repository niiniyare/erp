package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/niiniyare/erp/pkg/schema"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RuntimeTestSuite defines the test suite for Runtime
type RuntimeTestSuite struct {
	suite.Suite
	ctx     context.Context
	schema  *schema.Schema
	runtime *Runtime
}

// SetupTest runs before each test
func (s *RuntimeTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.schema = createTestSchema()
	s.runtime = createTestRuntimeWithValidator()
}

// TearDownTest runs after each test
func (s *RuntimeTestSuite) TearDownTest() {
	// Clean up if needed
	s.runtime = nil
}

// TestNewRuntime tests runtime initialization
func (s *RuntimeTestSuite) TestNewRuntime() {
	schema := createTestSchema()
	runtime := NewRuntime(schema)

	require.NotNil(s.T(), runtime, "NewRuntime() should not return nil")
	require.Equal(s.T(), schema, runtime.schema, "Runtime schema should be set correctly")
	require.NotNil(s.T(), runtime.state, "Runtime state should be initialized")
	require.NotNil(s.T(), runtime.events, "Runtime events should be initialized")
	
	// Test with validator
	runtimeWithValidator := createTestRuntimeWithValidator()
	require.NotNil(s.T(), runtimeWithValidator.validator, "Runtime with validator should have validator initialized")
}

// TestInitialize tests runtime initialization with data
func (s *RuntimeTestSuite) TestInitialize() {
	initialData := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	err := s.runtime.Initialize(s.ctx, initialData)
	require.NoError(s.T(), err, "Initialize() should not return error")

	// Check that initial data was set
	name, exists := s.runtime.GetFieldValue("name")
	require.True(s.T(), exists, "Name field should exist")
	require.Equal(s.T(), "John Doe", name, "Name should be set correctly")

	email, exists := s.runtime.GetFieldValue("email")
	require.True(s.T(), exists, "Email field should exist")
	require.Equal(s.T(), "john@example.com", email, "Email should be set correctly")
}

// TestHandleFieldChange tests field value changes
func (s *RuntimeTestSuite) TestHandleFieldChange() {
	// Initialize first with valid data for required fields
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Test field change
	err = s.runtime.HandleFieldChange(s.ctx, "name", "Jane Doe")
	require.NoError(s.T(), err, "HandleFieldChange() should not return error")

	// Check value was updated
	value, exists := s.runtime.GetFieldValue("name")
	require.True(s.T(), exists, "Field should exist")
	require.Equal(s.T(), "Jane Doe", value, "Field value should be updated")

	// Check field is marked as dirty
	require.True(s.T(), s.runtime.IsFieldDirty("name"), "Field should be marked as dirty after change")
}

// TestHandleFieldBlur tests field blur events
func (s *RuntimeTestSuite) TestHandleFieldBlur() {
	s.runtime.SetValidationTiming(schema.ValidateOnBlur)

	// Initialize first with valid data for required fields
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Test field blur
	err = s.runtime.HandleFieldBlur(s.ctx, "email", "invalid-email")
	require.NoError(s.T(), err, "HandleFieldBlur() should not return error")

	// Check field is marked as touched
	require.True(s.T(), s.runtime.IsFieldTouched("email"), "Field should be marked as touched after blur")
}

// TestValidateField tests field validation
func (s *RuntimeTestSuite) TestValidateField() {
	// Initialize first with valid data for required fields
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	tests := []struct {
		name      string
		fieldName string
		value     any
		wantError bool
	}{
		{
			name:      "valid email",
			fieldName: "email",
			value:     "test@example.com",
			wantError: false,
		},
		{
			name:      "invalid email",
			fieldName: "email",
			value:     "invalid-email",
			wantError: true,
		},
		{
			name:      "valid name",
			fieldName: "name",
			value:     "John Doe",
			wantError: false,
		},
		{
			name:      "empty required name",
			fieldName: "name",
			value:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			errors := s.runtime.ValidateField(s.ctx, tt.fieldName, tt.value)
			_ = len(errors) > 0

			if tt.wantError {
				require.NotEmpty(s.T(), errors, "Expected validation errors for %s", tt.name)
			} else {
				require.Empty(s.T(), errors, "Expected no validation errors for %s", tt.name)
			}
		})
	}
}

// TestValidationTiming tests different validation timing modes
func (s *RuntimeTestSuite) TestValidationTiming() {
	// Initialize first with valid data for required fields
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Test different validation timings
	timings := []schema.ValidationTiming{schema.ValidateOnChange, schema.ValidateOnBlur, schema.ValidateOnSubmit, schema.ValidateNever}

	for _, timing := range timings {
		s.Run(string(timing), func() {
			// Clear previous errors by setting empty errors
			s.runtime.GetState().SetErrors("email", []string{})

			s.runtime.SetValidationTiming(timing)

			// Handle a change with invalid data
			err := s.runtime.HandleFieldChange(s.ctx, "email", "invalid-email")
			require.NoError(s.T(), err, "HandleFieldChange should not return error")

			errors := s.runtime.GetState().GetErrors("email")
			hasErrors := len(errors) > 0

			// Only ValidateOnChange should have errors at this point
			if timing == schema.ValidateOnChange {
				require.True(s.T(), hasErrors, "Expected validation errors with ValidateOnChange timing")
			} else {
				require.False(s.T(), hasErrors, "Expected no validation errors with %s timing", timing)
			}
		})
	}
}

// TestGetStats tests runtime statistics
func (s *RuntimeTestSuite) TestGetStats() {
	// Initialize with data
	initialData := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	err := s.runtime.Initialize(s.ctx, initialData)
	require.NoError(s.T(), err, "Initialize should not fail")

	// Make some changes
	err = s.runtime.HandleFieldChange(s.ctx, "name", "Jane Doe")
	require.NoError(s.T(), err)

	err = s.runtime.HandleFieldChange(s.ctx, "email", "jane@example.com")
	require.NoError(s.T(), err)

	err = s.runtime.HandleFieldBlur(s.ctx, "email", "jane@example.com")
	require.NoError(s.T(), err)

	stats := s.runtime.GetStats()

	require.NotNil(s.T(), stats, "GetStats() should not return nil")
	require.Equal(s.T(), 3, stats.FieldCount, "Expected 3 fields (name, email, age)")
	require.Equal(s.T(), 1, stats.TouchedFields, "Expected 1 touched field (email)")
	require.Equal(s.T(), 2, stats.DirtyFields, "Expected 2 dirty fields (name, email)")
}

// TestReset tests runtime reset functionality
func (s *RuntimeTestSuite) TestReset() {
	initialData := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	}

	// Initialize with data
	err := s.runtime.Initialize(s.ctx, initialData)
	require.NoError(s.T(), err, "Initialize should not fail")

	// Make changes
	err = s.runtime.HandleFieldChange(s.ctx, "name", "Jane Doe")
	require.NoError(s.T(), err)

	err = s.runtime.HandleFieldBlur(s.ctx, "email", "jane@example.com")
	require.NoError(s.T(), err)

	// Verify changes
	require.True(s.T(), s.runtime.IsDirty(), "Runtime should be dirty after changes")

	// Reset
	s.runtime.Reset()

	// Verify reset
	require.False(s.T(), s.runtime.IsDirty(), "Runtime should not be dirty after reset")

	// Check values are back to initial
	name, exists := s.runtime.GetFieldValue("name")
	require.True(s.T(), exists)
	require.Equal(s.T(), "John Doe", name, "Name should be reset to initial value")
}

// TestConditionalLogic tests conditional field logic
func (s *RuntimeTestSuite) TestConditionalLogic() {
	// Create schema with conditional field
	conditionalSchema := &schema.Schema{
		ID:    "test_conditional",
		Title: "Test Conditional Schema",
		Fields: []schema.Field{
			{
				Name:     "show_field",
				Type:     schema.FieldCheckbox,
				Label:    "Show Field",
				Required: false,
			},
			{
				Name:     "conditional_field",
				Type:     schema.FieldText,
				Label:    "Conditional Field",
				Required: false,
			},
		},
	}

	runtime := NewRuntime(conditionalSchema)

	// Initialize
	err := runtime.Initialize(s.ctx, map[string]any{
		"show_field": false,
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Test applying conditional logic
	err = runtime.ApplyConditionalLogic(s.ctx)
	require.NoError(s.T(), err, "ApplyConditionalLogic() should not return error")
}

// TestEventHandling tests custom event handler registration
func (s *RuntimeTestSuite) TestEventHandling() {
	// Initialize
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Register custom event handler
	eventTriggered := false
	expectedField := "name"
	expectedValue := "Test Name"

	s.runtime.RegisterEventHandler(schema.EventChange, func(ctx context.Context, event *schema.Event) error {
		eventTriggered = true
		require.Equal(s.T(), expectedField, event.Field, "Event field should match")
		require.Equal(s.T(), expectedValue, event.Value, "Event value should match")
		return nil
	})

	// Trigger change
	err = s.runtime.HandleFieldChange(s.ctx, expectedField, expectedValue)
	require.NoError(s.T(), err, "HandleFieldChange should not return error")

	require.True(s.T(), eventTriggered, "Custom event handler should be triggered")
}

// TestValidateWithDebounce tests debounced validation
func (s *RuntimeTestSuite) TestValidateWithDebounce() {
	// Initialize
	err := s.runtime.Initialize(s.ctx, map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Test debounced validation
	resultChan := s.runtime.ValidateWithDebounce(s.ctx, "email", "invalid-email", 100*time.Millisecond)

	// Wait for result
	select {
	case errors := <-resultChan:
		require.NotEmpty(s.T(), errors, "Expected validation errors for invalid email")
	case <-time.After(200 * time.Millisecond):
		s.T().Error("Debounced validation timed out")
	}
}

// TestEnrichedSchemaIntegration tests runtime with enriched schema context
func (s *RuntimeTestSuite) TestEnrichedSchemaIntegration() {
	// Create schema with runtime permissions set by enricher
	enrichedSchema := &schema.Schema{
		ID:    "enriched_test",
		Title: "Enriched Test Schema",
		Fields: []schema.Field{
			{
				Name:     "public_field",
				Type:     schema.FieldText,
				Label:    "Public Field",
				Required: false,
				Runtime: &schema.FieldRuntime{
					Visible:  true,
					Editable: true,
					Reason:   "",
				},
			},
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
			{
				Name:     "hidden_field",
				Type:     schema.FieldText,
				Label:    "Hidden Field",
				Required: false,
				Runtime: &schema.FieldRuntime{
					Visible:  false,
					Editable: false,
					Reason:   "User lacks view permission",
				},
			},
		},
	}

	runtime := NewRuntime(enrichedSchema)

	// Initialize
	err := runtime.Initialize(s.ctx, map[string]any{})
	require.NoError(s.T(), err, "Initialize should not fail")

	// Test editing public field (should work)
	err = runtime.HandleFieldChange(s.ctx, "public_field", "test value")
	require.NoError(s.T(), err, "Should be able to edit public field")

	// Test editing readonly field (should fail)
	err = runtime.HandleFieldChange(s.ctx, "readonly_field", "test value")
	require.Error(s.T(), err, "Should not be able to edit readonly field")

	// Test validation only processes visible fields
	allErrors := runtime.ValidateCurrentState(s.ctx)
	_, hasHidden := allErrors["hidden_field"]
	require.False(s.T(), hasHidden, "Hidden field should not be validated")

	// Test stats reflect visible fields correctly
	stats := runtime.GetStats()
	require.Equal(s.T(), 2, stats.VisibleFields, "Expected 2 visible fields (public and readonly)")
}

// Helper method to create a test schema

// TestRuntimeTestSuite runs the test suite
func TestRuntimeTestSuite(t *testing.T) {
	suite.Run(t, new(RuntimeTestSuite))
}
