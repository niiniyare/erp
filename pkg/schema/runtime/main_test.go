package runtime

import (
	"context"
	"strings"

	"github.com/niiniyare/erp/pkg/schema"
)

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

// MockValidator implements schema.RuntimeValidator for testing
type MockValidator struct{}

func (mv *MockValidator) ValidateField(ctx context.Context, field *schema.Field, value any, allData map[string]any) []string {
	var errors []string
	
	// Required field validation
	if field.Required && (value == nil || value == "") {
		errors = append(errors, field.Label+" is required")
	}
	
	// Email validation
	if field.Type == schema.FieldEmail && value != nil {
		email := value.(string)
		if email != "" && !strings.Contains(email, "@") {
			errors = append(errors, "Invalid email format")
		}
	}
	
	return errors
}

func (mv *MockValidator) ValidateAllFields(ctx context.Context, schema *schema.Schema, data map[string]any) map[string][]string {
	errors := make(map[string][]string)
	
	for _, field := range schema.Fields {
		value, _ := data[field.Name]
		fieldErrors := mv.ValidateField(ctx, &field, value, data)
		if len(fieldErrors) > 0 {
			errors[field.Name] = fieldErrors
		}
	}
	
	return errors
}

func (mv *MockValidator) ValidateAction(ctx context.Context, action *schema.Action, formState map[string]any) error {
	return nil
}

func (mv *MockValidator) ValidateWorkflow(ctx context.Context, workflow *schema.Workflow, currentStage string, formState map[string]any) error {
	return nil
}

func (mv *MockValidator) ValidateBusinessRules(ctx context.Context, rules []schema.ValidationRule, allValues map[string]any) map[string][]string {
	return make(map[string][]string)
}

func (mv *MockValidator) ValidateFieldAsync(ctx context.Context, field *schema.Field, value any, callback schema.ValidationCallback) error {
	errors := mv.ValidateField(ctx, field, value, nil)
	callback(field.Name, errors)
	return nil
}

// createTestRuntimeWithValidator creates a runtime with a mock validator for testing
func createTestRuntimeWithValidator() *Runtime {
	return NewRuntimeWithValidator(createTestSchema(), &MockValidator{})
}
