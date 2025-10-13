package components

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Schema Validator Creation
func TestSchemaValidatorCreation(t *testing.T) {
	validator := NewSchemaValidator()
	assert.NotNil(t, validator, "Validator should not be nil")
	assert.NotNil(t, validator.context, "Validation context should be initialized")
	assert.Equal(t, ValidationModeStrict, validator.context.ValidationMode, "Default mode should be strict")
}

// Test Validation Mode Setting
func TestValidationModeSettings(t *testing.T) {
	validator := NewSchemaValidator()
	
	// Test setting different modes
	validator.SetValidationMode(ValidationModeRelaxed)
	assert.Equal(t, ValidationModeRelaxed, validator.context.ValidationMode)
	
	validator.SetValidationMode(ValidationModeProduction)
	assert.Equal(t, ValidationModeProduction, validator.context.ValidationMode)
	
	validator.SetValidationMode(ValidationModeStrict)
	assert.Equal(t, ValidationModeStrict, validator.context.ValidationMode)
}

// Test Basic Component Validation
func TestBasicComponentValidation(t *testing.T) {
	validator := NewSchemaValidator()

	t.Run("Valid CRUD Component", func(t *testing.T) {
		crud := &CRUDSchema{
			BaseComponentProps: BaseComponentProps{
				ID:        "valid-crud",
				ClassName: "custom-crud-class",
			},
			Type:  "crud",
			Title: "User Management",
			Mode:  "table",
		}

		result := validator.ValidateComponent(crud)
		assert.NotNil(t, result, "Result should not be nil")
		assert.True(t, result.Valid, "CRUD should be valid")
		assert.Equal(t, 0, len(result.Errors), "Should have no errors")
		assert.Equal(t, "crud", result.Summary.ComponentType, "Component type should be crud")
	})

	t.Run("Invalid CRUD Component", func(t *testing.T) {
		crud := &CRUDSchema{
			Type: "crud",
			Mode: "invalid-mode",
		}

		result := validator.ValidateComponent(crud)
		assert.NotNil(t, result, "Result should not be nil")
		assert.False(t, result.Valid, "CRUD with invalid mode should not be valid")
		assert.Greater(t, len(result.Errors), 0, "Should have at least one error")
		
		// Check for specific error about invalid mode
		found := false
		for _, err := range result.Errors {
			if strings.Contains(err.Message, "mode must be one of") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should have error about invalid mode")
	})
}

// Test Form Component Validation
func TestFormComponentValidation(t *testing.T) {
	validator := NewSchemaValidator()

	t.Run("Valid Form", func(t *testing.T) {
		form := &FormSchema{
			BaseComponentProps: BaseComponentProps{ID: "test-form"},
			Type:  "form",
			Title: "Test Form",
			Mode:  "horizontal",
			Body:  []any{"input1", "input2"}, // Non-empty body
		}

		result := validator.ValidateComponent(form)
		assert.True(t, result.Valid, "Form should be valid")
		assert.Equal(t, 0, len(result.Errors), "Should have no errors")
	})

	t.Run("Form with Invalid Mode", func(t *testing.T) {
		form := &FormSchema{
			Type: "form",
			Mode: "invalid-mode",
		}

		result := validator.ValidateComponent(form)
		assert.False(t, result.Valid, "Form with invalid mode should not be valid")
		assert.Greater(t, len(result.Errors), 0, "Should have errors")
	})

	t.Run("Form with Empty Body Warning", func(t *testing.T) {
		form := &FormSchema{
			Type:  "form",
			Title: "Empty Form",
			Mode:  "normal",
			Body:  []any{}, // Empty body
		}

		result := validator.ValidateComponent(form)
		// This should generate a warning, not an error
		assert.Greater(t, len(result.Warnings), 0, "Should have warnings about empty body")
	})
}

// Test Text Input Advanced Validation
func TestTextInputAdvancedValidation(t *testing.T) {
	validator := NewSchemaValidator()

	t.Run("Valid Text Input", func(t *testing.T) {
		textInput := &TextControlSchema{
			BaseComponentProps: BaseComponentProps{ID: "username-input"},
			Type:  "input-text",
			Name:  "username",
			Label: "Username",
		}

		result := validator.ValidateComponent(textInput)
		assert.True(t, result.Valid, "Text input should be valid")
		assert.Equal(t, 0, len(result.Errors), "Should have no errors")
	})

	t.Run("Text Input Missing Name", func(t *testing.T) {
		textInput := &TextControlSchema{
			Type:  "input-text",
			Label: "Username",
			// Missing Name field
		}

		result := validator.ValidateComponent(textInput)
		assert.False(t, result.Valid, "Text input without name should not be valid")
		assert.Greater(t, len(result.Errors), 0, "Should have errors about missing name")
		
		// Check for name required error
		found := false
		for _, err := range result.Errors {
			if err.Field == "name" && strings.Contains(err.Message, "required") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should have error about required name")
	})

	t.Run("Text Input Missing Label Warning", func(t *testing.T) {
		textInput := &TextControlSchema{
			Type: "input-text",
			Name: "username",
			// Missing Label field
		}

		result := validator.ValidateComponent(textInput)
		// Should have warnings about missing label for accessibility
		assert.Greater(t, len(result.Warnings), 0, "Should have warnings about missing label")
	})
}

// Test Select Control Validation
func TestSelectControlValidation(t *testing.T) {
	validator := NewSchemaValidator()

	t.Run("Select with Options", func(t *testing.T) {
		selectControl := &SelectControlSchema{
			Type: "select",
			Name: "category",
			Options: []SelectOption{
				{Label: "Option 1", Value: "opt1"},
				{Label: "Option 2", Value: "opt2"},
			},
		}

		result := validator.ValidateComponent(selectControl)
		assert.True(t, result.Valid, "Select with options should be valid")
		assert.Equal(t, 0, len(result.Errors), "Should have no errors")
	})

	t.Run("Select with Source", func(t *testing.T) {
		selectControl := &SelectControlSchema{
			Type:   "select",
			Name:   "category",
			Source: "/api/categories",
		}

		result := validator.ValidateComponent(selectControl)
		assert.True(t, result.Valid, "Select with source should be valid")
		assert.Equal(t, 0, len(result.Errors), "Should have no errors")
	})

	t.Run("Select with API", func(t *testing.T) {
		selectControl := &SelectControlSchema{
			Type: "select",
			Name: "category",
			API: &APIConfig{
				URL:    "/api/categories",
				Method: "GET",
			},
		}

		result := validator.ValidateComponent(selectControl)
		assert.True(t, result.Valid, "Select with API should be valid")
		assert.Equal(t, 0, len(result.Errors), "Should have no errors")
	})

	t.Run("Select without Options or Source", func(t *testing.T) {
		selectControl := &SelectControlSchema{
			Type: "select",
			Name: "category",
			// No Options, Source, or API
		}

		result := validator.ValidateComponent(selectControl)
		assert.False(t, result.Valid, "Select without options/source should not be valid")
		assert.Greater(t, len(result.Errors), 0, "Should have errors")
		
		// Check for specific error about missing options/source
		found := false
		for _, err := range result.Errors {
			if strings.Contains(err.Message, "options") || strings.Contains(err.Message, "source") {
				found = true
				assert.NotEmpty(t, err.Suggestions, "Should provide suggestions")
				break
			}
		}
		assert.True(t, found, "Should have error about missing options or source")
	})
}

// Test Action Component Validation
func TestActionComponentValidation(t *testing.T) {
	validator := NewSchemaValidator()

	t.Run("Action with Label", func(t *testing.T) {
		action := &ActionSchema{
			Type:       "action",
			ActionType: "ajax",
			Label:      "Save",
		}

		result := validator.ValidateComponent(action)
		assert.True(t, result.Valid, "Action with label should be valid")
		assert.Equal(t, 0, len(result.Errors), "Should have no errors")
	})

	t.Run("Action with Icon", func(t *testing.T) {
		action := &ActionSchema{
			Type:       "action",
			ActionType: "ajax",
			Icon:       "save-icon",
		}

		result := validator.ValidateComponent(action)
		assert.True(t, result.Valid, "Action with icon should be valid")
		assert.Equal(t, 0, len(result.Errors), "Should have no errors")
	})

	t.Run("Action without Label or Icon", func(t *testing.T) {
		action := &ActionSchema{
			Type:       "action",
			ActionType: "ajax",
			// No Label or Icon
		}

		result := validator.ValidateComponent(action)
		// This should generate a warning, not an error (usability issue)
		assert.Greater(t, len(result.Warnings), 0, "Should have warnings about missing label/icon")
		
		// Check for suggestions
		found := false
		for _, warn := range result.Warnings {
			if len(warn.Suggestions) > 0 {
				found = true
				break
			}
		}
		assert.True(t, found, "Should provide suggestions for improvement")
	})

	t.Run("Action with Invalid Type", func(t *testing.T) {
		action := &ActionSchema{
			Type:       "action",
			ActionType: "invalid-type",
			Label:      "Save",
		}

		result := validator.ValidateComponent(action)
		assert.False(t, result.Valid, "Action with invalid type should not be valid")
		assert.Greater(t, len(result.Errors), 0, "Should have errors about invalid action type")
	})
}

// Test Dialog Component Validation
func TestDialogComponentValidation(t *testing.T) {
	validator := NewSchemaValidator()

	t.Run("Dialog with Title", func(t *testing.T) {
		dialog := &DialogSchema{
			Type:  "dialog",
			Title: "Confirmation",
		}

		result := validator.ValidateComponent(dialog)
		assert.True(t, result.Valid, "Dialog with title should be valid")
		assert.Equal(t, 0, len(result.Errors), "Should have no errors")
	})

	t.Run("Dialog with Body", func(t *testing.T) {
		dialog := &DialogSchema{
			Type: "dialog",
			Body: []any{"content"},
		}

		result := validator.ValidateComponent(dialog)
		assert.True(t, result.Valid, "Dialog with body should be valid")
		assert.Equal(t, 0, len(result.Errors), "Should have no errors")
	})

	t.Run("Dialog without Title or Body", func(t *testing.T) {
		dialog := &DialogSchema{
			Type: "dialog",
			// No Title or Body
		}

		result := validator.ValidateComponent(dialog)
		// Should have warnings about empty dialog
		assert.Greater(t, len(result.Warnings), 0, "Should have warnings about empty dialog")
	})
}

// Test Custom Validation Rules
func TestCustomValidationRules(t *testing.T) {
	validator := NewSchemaValidator()

	// Add a custom rule
	validator.AddCustomRule("custom_test", func(value any, context *ValidationContext) []ComponentValidationError {
		crud, ok := value.(*CRUDSchema)
		if !ok {
			return []ComponentValidationError{}
		}

		var errors []ComponentValidationError
		if crud.Title == "FORBIDDEN" {
			errors = append(errors, ComponentValidationError{
				Field:    "title",
				Rule:     "custom_test",
				Message:  "Title 'FORBIDDEN' is not allowed",
				Severity: SeverityError,
			})
		}
		return errors
	})

	t.Run("Custom Rule Pass", func(t *testing.T) {
		crud := &CRUDSchema{
			Type:  "crud",
			Title: "Allowed Title",
			Mode:  "table",
		}

		result := validator.ValidateComponent(crud)
		assert.True(t, result.Valid, "Should pass custom validation")
	})

	t.Run("Custom Rule Fail", func(t *testing.T) {
		crud := &CRUDSchema{
			Type:  "crud",
			Title: "FORBIDDEN",
			Mode:  "table",
		}

		result := validator.ValidateComponent(crud)
		assert.False(t, result.Valid, "Should fail custom validation")
		
		// Check for custom rule error
		found := false
		for _, err := range result.Errors {
			if err.Rule == "custom_test" {
				found = true
				break
			}
		}
		assert.True(t, found, "Should have custom rule error")
	})
}

// Test Batch Validation
func TestBatchValidation(t *testing.T) {
	validator := NewSchemaValidator()

	components := map[string]any{
		"valid_crud": &CRUDSchema{
			Type:  "crud",
			Title: "Valid CRUD",
			Mode:  "table",
		},
		"invalid_crud": &CRUDSchema{
			Type: "crud",
			Mode: "invalid",
		},
		"valid_form": &FormSchema{
			Type:  "form",
			Title: "Valid Form",
			Mode:  "normal",
		},
	}

	results := validator.ValidateComponentBatch(components)
	assert.Len(t, results, 3, "Should have 3 results")
	
	assert.True(t, results["valid_crud"].Valid, "Valid CRUD should pass")
	assert.False(t, results["invalid_crud"].Valid, "Invalid CRUD should fail")
	assert.True(t, results["valid_form"].Valid, "Valid form should pass")
}

// Test Validation Result Methods
func TestValidationResultMethods(t *testing.T) {
	// Create a result with mixed errors and warnings
	result := &ValidationResult{
		Valid: false,
		Errors: []ComponentValidationError{
			{Field: "type", Message: "Type is required", Severity: SeverityError},
		},
		Warnings: []ComponentValidationError{
			{Field: "label", Message: "Label recommended", Severity: SeverityWarning},
		},
		Summary: ValidationSummary{
			TotalIssues:    2,
			ErrorCount:     1,
			WarningCount:   1,
			ComponentType:  "test",
			ValidationMode: "strict",
		},
	}

	// Test helper methods
	assert.True(t, result.HasErrors(), "Should have errors")
	assert.True(t, result.HasWarnings(), "Should have warnings")

	// Test report generation
	report := result.GetValidationReport()
	assert.NotEmpty(t, report, "Report should not be empty")
	assert.Contains(t, report, "=== Validation Report ===", "Should have header")
	assert.Contains(t, report, "ERRORS:", "Should have errors section")
	assert.Contains(t, report, "WARNINGS:", "Should have warnings section")

	// Test JSON serialization
	jsonStr, err := result.ToJSON()
	require.NoError(t, err, "Should serialize to JSON")
	assert.NotEmpty(t, jsonStr, "JSON should not be empty")
	
	// Verify it's valid JSON
	var parsed map[string]any
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	assert.NoError(t, err, "Should be valid JSON")
}

// Test Validation Severity Levels
func TestValidationSeverityLevels(t *testing.T) {
	assert.Equal(t, "error", SeverityError.String())
	assert.Equal(t, "warning", SeverityWarning.String())
	assert.Equal(t, "info", SeverityInfo.String())
}

// Test Validation Mode Strings
func TestValidationModeStrings(t *testing.T) {
	assert.Equal(t, "strict", ValidationModeStrict.String())
	assert.Equal(t, "relaxed", ValidationModeRelaxed.String())
	assert.Equal(t, "production", ValidationModeProduction.String())
}

// Test Pattern Validation
func TestPatternValidation(t *testing.T) {
	validator := NewSchemaValidator()

	t.Run("Valid ID Pattern", func(t *testing.T) {
		component := &CRUDSchema{
			BaseComponentProps: BaseComponentProps{
				ID: "validId123",
			},
			Type: "crud",
			Mode: "table",
		}

		result := validator.ValidateComponent(component)
		// Should not have warnings about ID format
		idWarnings := 0
		for _, warn := range result.Warnings {
			if warn.Field == "id" {
				idWarnings++
			}
		}
		assert.Equal(t, 0, idWarnings, "Valid ID should not generate warnings")
	})

	t.Run("Invalid ID Pattern", func(t *testing.T) {
		component := &CRUDSchema{
			BaseComponentProps: BaseComponentProps{
				ID: "123invalid", // Starts with number
			},
			Type: "crud",
			Mode: "table",
		}

		result := validator.ValidateComponent(component)
		// Should have warning about invalid ID format
		idWarnings := 0
		for _, warn := range result.Warnings {
			if warn.Field == "id" {
				idWarnings++
				break
			}
		}
		assert.Greater(t, idWarnings, 0, "Invalid ID should generate warning")
	})
}

// Test Date Control Validation
func TestDateControlValidation(t *testing.T) {
	validator := NewSchemaValidator()

	t.Run("Valid Date Control", func(t *testing.T) {
		dateControl := &DateControlSchema{
			Type:   "input-date",
			Name:   "birthdate",
			Format: "YYYY-MM-DD",
		}

		result := validator.ValidateComponent(dateControl)
		assert.True(t, result.Valid, "Date control should be valid")
	})

	t.Run("Date Control Missing Name", func(t *testing.T) {
		dateControl := &DateControlSchema{
			Type:   "input-date",
			Format: "YYYY-MM-DD",
			// Missing Name
		}

		result := validator.ValidateComponent(dateControl)
		assert.False(t, result.Valid, "Date control without name should not be valid")
	})

	t.Run("Date Control Invalid Format Warning", func(t *testing.T) {
		dateControl := &DateControlSchema{
			Type:   "input-date",
			Name:   "date",
			Format: "INVALID-FORMAT!!!", // Invalid format
		}

		result := validator.ValidateComponent(dateControl)
		// Should generate a warning about format
		formatWarnings := 0
		for _, warn := range result.Warnings {
			if warn.Field == "format" {
				formatWarnings++
				break
			}
		}
		assert.Greater(t, formatWarnings, 0, "Invalid format should generate warning")
	})
}

// Test Performance Metrics
func TestValidationPerformanceMetrics(t *testing.T) {
	validator := NewSchemaValidator()

	component := &CRUDSchema{
		Type:  "crud",
		Title: "Performance Test",
		Mode:  "table",
		Columns: []TableColumn{
			{Name: "col1", Label: "Column 1"},
			{Name: "col2", Label: "Column 2"},
		},
	}

	result := validator.ValidateComponent(component)
	
	// Check that performance metrics are captured
	assert.NotZero(t, result.Performance.RulesExecuted, "Should track rules executed")
	assert.NotZero(t, result.Performance.FieldsValidated, "Should track fields validated")
	assert.NotNil(t, result.Performance.ValidationTimeMs, "Should track validation time")
}

// Test Validation Context
func TestValidationContext(t *testing.T) {
	validator := NewSchemaValidator()

	// Customize validation context
	validator.context.ComponentName = "test-component"

	component := &FormSchema{
		Type:  "form",
		Title: "Context Test",
	}

	result := validator.ValidateComponent(component)
	assert.NotNil(t, result, "Should validate with custom context")
	assert.Equal(t, "form", result.Summary.ComponentType, "Should track component type")
}

// Benchmark validation performance
func BenchmarkValidationPerformance(b *testing.B) {
	validator := NewSchemaValidator()
	component := &CRUDSchema{
		BaseComponentProps: BaseComponentProps{
			ID:        "benchmark-crud",
			ClassName: "benchmark-class",
		},
		Type:  "crud",
		Title: "Benchmark CRUD",
		Mode:  "table",
		Columns: []TableColumn{
			{Name: "id", Label: "ID", Type: "text"},
			{Name: "name", Label: "Name", Type: "text"},
			{Name: "email", Label: "Email", Type: "email"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := validator.ValidateComponent(component)
		if !result.Valid {
			b.Fatal("Component should be valid")
		}
	}
}

// Test error edge cases
func TestValidationEdgeCases(t *testing.T) {
	validator := NewSchemaValidator()

	t.Run("Nil Component", func(t *testing.T) {
		result := validator.ValidateComponent(nil)
		assert.NotNil(t, result, "Should handle nil component")
		assert.False(t, result.Valid, "Nil component should not be valid")
	})

	t.Run("Unknown Component Type", func(t *testing.T) {
		type UnknownComponent struct {
			Type string `json:"type"`
		}
		unknown := &UnknownComponent{Type: "unknown"}

		result := validator.ValidateComponent(unknown)
		assert.NotNil(t, result, "Should handle unknown component")
		assert.Equal(t, "unknown", result.Summary.ComponentType, "Should track unknown type")
	})

	t.Run("Empty Validation Result", func(t *testing.T) {
		result := &ValidationResult{}
		
		assert.False(t, result.HasErrors(), "Empty result should not have errors")
		assert.False(t, result.HasWarnings(), "Empty result should not have warnings")
		
		report := result.GetValidationReport()
		assert.NotEmpty(t, report, "Should generate report for empty result")
	})
}