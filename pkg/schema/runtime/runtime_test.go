package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/niiniyare/erp/pkg/schema"
)

func TestNewRuntime(t *testing.T) {
	schema := createTestSchema()
	runtime := NewRuntime(schema)

	if runtime == nil {
		t.Fatal("NewRuntime() returned nil")
	}

	if runtime.schema != schema {
		t.Error("Runtime schema not set correctly")
	}

	if runtime.state == nil {
		t.Error("Runtime state not initialized")
	}

	if runtime.validator == nil {
		t.Error("Runtime validator not initialized")
	}

	if runtime.events == nil {
		t.Error("Runtime events not initialized")
	}
}

func TestRuntime_Initialize(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	initialData := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	err := runtime.Initialize(ctx, initialData)
	if err != nil {
		t.Errorf("Initialize() error = %v", err)
	}

	// Check that initial data was set
	name, exists := runtime.GetFieldValue("name")
	if !exists || name != "John Doe" {
		t.Errorf("Initial data not set correctly for name field")
	}

	email, exists := runtime.GetFieldValue("email")
	if !exists || email != "john@example.com" {
		t.Errorf("Initial data not set correctly for email field")
	}
}

func TestRuntime_HandleFieldChange(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	// Initialize first
	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Test field change
	err = runtime.HandleFieldChange(ctx, "name", "Jane Doe")
	if err != nil {
		t.Errorf("HandleFieldChange() error = %v", err)
	}

	// Check value was updated
	value, exists := runtime.GetFieldValue("name")
	if !exists || value != "Jane Doe" {
		t.Errorf("Field value not updated correctly")
	}

	// Check field is marked as dirty
	if !runtime.IsFieldDirty("name") {
		t.Error("Field should be marked as dirty after change")
	}
}

func TestRuntime_HandleFieldBlur(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	runtime.SetValidationTiming(ValidateOnBlur)
	ctx := context.Background()

	// Initialize first
	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Test field blur
	err = runtime.HandleFieldBlur(ctx, "email", "invalid-email")
	if err != nil {
		t.Errorf("HandleFieldBlur() error = %v", err)
	}

	// Check field is marked as touched
	if !runtime.IsFieldTouched("email") {
		t.Error("Field should be marked as touched after blur")
	}
}

func TestRuntime_ValidateField(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	// Initialize first
	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	tests := []struct {
		name      string
		fieldName string
		value     interface{}
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
		t.Run(tt.name, func(t *testing.T) {
			errors := runtime.ValidateField(ctx, tt.fieldName, tt.value)
			hasError := len(errors) > 0

			if hasError != tt.wantError {
				t.Errorf("ValidateField() hasError = %v, wantError = %v, errors = %v", hasError, tt.wantError, errors)
			}
		})
	}
}

func TestRuntime_ValidationTiming(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	// Initialize first
	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Test different validation timings
	timings := []ValidationTiming{ValidateOnChange, ValidateOnBlur, ValidateOnSubmit, ValidateNever}

	for _, timing := range timings {
		t.Run(string(timing), func(t *testing.T) {
			// Clear previous errors
			runtime.GetState().ClearErrors("email")
			
			runtime.SetValidationTiming(timing)

			// Handle a change with invalid data
			err := runtime.HandleFieldChange(ctx, "email", "invalid-email")
			if err != nil {
				t.Errorf("HandleFieldChange() error = %v", err)
			}

			errors := runtime.GetState().GetErrors("email")
			hasErrors := len(errors) > 0

			// Only ValidateOnChange should have errors at this point
			if timing == ValidateOnChange && !hasErrors {
				t.Error("Expected validation errors with ValidateOnChange timing")
			} else if timing != ValidateOnChange && hasErrors {
				t.Error("Unexpected validation errors with non-change timing")
			}
		})
	}
}

func TestRuntime_GetStats(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	// Initialize first
	err := runtime.Initialize(ctx, map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Make some changes
	runtime.HandleFieldChange(ctx, "name", "Jane Doe")
	runtime.HandleFieldChange(ctx, "email", "jane@example.com")
	runtime.HandleFieldBlur(ctx, "email", "jane@example.com")

	stats := runtime.GetStats()

	if stats == nil {
		t.Fatal("GetStats() returned nil")
	}

	if stats.FieldCount != 3 { // name, email, age from test schema
		t.Errorf("Expected 3 fields, got %d", stats.FieldCount)
	}

	if stats.TouchedFields != 1 { // only email was blurred (touched)
		t.Errorf("Expected 1 touched field, got %d", stats.TouchedFields)
	}

	if stats.DirtyFields != 2 { // both name and email were changed
		t.Errorf("Expected 2 dirty fields, got %d", stats.DirtyFields)
	}
}

func TestRuntime_Reset(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	initialData := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}

	// Initialize with data
	err := runtime.Initialize(ctx, initialData)
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Make changes
	runtime.HandleFieldChange(ctx, "name", "Jane Doe")
	runtime.HandleFieldBlur(ctx, "email", "jane@example.com")

	// Verify changes
	if !runtime.IsDirty() {
		t.Error("Runtime should be dirty after changes")
	}

	// Reset
	runtime.Reset()

	// Verify reset
	if runtime.IsDirty() {
		t.Error("Runtime should not be dirty after reset")
	}

	// Check values are back to initial
	name, _ := runtime.GetFieldValue("name")
	if name != "John Doe" {
		t.Errorf("Name should be reset to initial value, got %v", name)
	}
}

func TestRuntime_ConditionalLogic(t *testing.T) {
	// Create schema with conditional field
	schema := &schema.Schema{
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
				// Note: Conditional logic would be handled by field.IsVisible()
				// The runtime applies this through ApplyConditionalLogic()
			},
		},
	}

	runtime := NewRuntime(schema)
	ctx := context.Background()

	// Initialize
	err := runtime.Initialize(ctx, map[string]interface{}{
		"show_field": false,
	})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Test applying conditional logic
	err = runtime.ApplyConditionalLogic(ctx)
	if err != nil {
		t.Errorf("ApplyConditionalLogic() error = %v", err)
	}

	// The actual conditional logic evaluation depends on the condition evaluator
	// For this test, we just verify the method doesn't error
}

func TestRuntime_EventHandling(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	// Initialize
	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Register custom event handler
	eventTriggered := false
	runtime.RegisterEventHandler(EventChange, func(ctx context.Context, event *Event) error {
		eventTriggered = true
		if event.Field != "name" {
			t.Errorf("Expected field 'name', got '%s'", event.Field)
		}
		if event.Value != "Test Name" {
			t.Errorf("Expected value 'Test Name', got '%v'", event.Value)
		}
		return nil
	})

	// Trigger change
	err = runtime.HandleFieldChange(ctx, "name", "Test Name")
	if err != nil {
		t.Errorf("HandleFieldChange() error = %v", err)
	}

	if !eventTriggered {
		t.Error("Custom event handler was not triggered")
	}
}

func TestRuntime_ValidateWithDebounce(t *testing.T) {
	runtime := NewRuntime(createTestSchema())
	ctx := context.Background()

	// Initialize
	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Test debounced validation
	resultChan := runtime.ValidateWithDebounce(ctx, "email", "invalid-email", 100*time.Millisecond)

	// Wait for result
	select {
	case errors := <-resultChan:
		if len(errors) == 0 {
			t.Error("Expected validation errors for invalid email")
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Debounced validation timed out")
	}
}

// Helper function to create a test schema
func createTestSchema() *schema.Schema {
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

// Test runtime with enriched schema context
func TestRuntime_EnrichedSchemaIntegration(t *testing.T) {
	// Create schema with runtime permissions set by enricher
	schema := &schema.Schema{
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

	runtime := NewRuntime(schema)
	ctx := context.Background()

	// Initialize
	err := runtime.Initialize(ctx, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to initialize runtime: %v", err)
	}

	// Test editing public field (should work)
	err = runtime.HandleFieldChange(ctx, "public_field", "test value")
	if err != nil {
		t.Errorf("Should be able to edit public field: %v", err)
	}

	// Test editing readonly field (should fail)
	err = runtime.HandleFieldChange(ctx, "readonly_field", "test value")
	if err == nil {
		t.Error("Should not be able to edit readonly field")
	}

	// Test validation only processes visible fields
	allErrors := runtime.ValidateCurrentState(ctx)
	if _, hasHidden := allErrors["hidden_field"]; hasHidden {
		t.Error("Hidden field should not be validated")
	}

	// Test stats reflect visible fields correctly
	stats := runtime.GetStats()
	if stats.VisibleFields != 2 { // only public_field and readonly_field
		t.Errorf("Expected 2 visible fields, got %d", stats.VisibleFields)
	}
}