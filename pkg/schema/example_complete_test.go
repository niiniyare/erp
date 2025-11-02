package schema

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComprehensiveExample(t *testing.T) {
	schema, err := CreateComprehensiveExampleSchema()
	require.NoError(t, err)
	require.NotNil(t, schema)

	// Test basic schema properties
	assert.Equal(t, "comprehensive_example", schema.ID)
	assert.Equal(t, TypeForm, schema.Type)
	assert.Equal(t, "Comprehensive Example Form", schema.Title)
	assert.Equal(t, "1.0.0", schema.Version)

	// Test that mixins were applied
	assert.NotNil(t, schema.Meta)
	appliedMixins, exists := schema.Meta.CustomData["applied_mixins"]
	assert.True(t, exists)
	mixins := appliedMixins.([]string)
	assert.Contains(t, mixins, "audit_fields")
	assert.Contains(t, mixins, "status_fields")

	// Test security configuration
	assert.NotNil(t, schema.Security)
	assert.NotNil(t, schema.Security.CSRF)
	assert.True(t, schema.Security.CSRF.Enabled)
	assert.NotNil(t, schema.Security.RateLimit)
	assert.True(t, schema.Security.RateLimit.Enabled)

	// Test tenant configuration
	assert.NotNil(t, schema.Tenant)
	assert.True(t, schema.Tenant.Enabled)
	assert.Equal(t, "tenant_id", schema.Tenant.Field)
	assert.Equal(t, "strict", schema.Tenant.Isolation)

	// Test HTMX configuration
	assert.NotNil(t, schema.HTMX)
	assert.True(t, schema.HTMX.Enabled)
	assert.Equal(t, "/api/examples", schema.HTMX.Post)
	assert.Equal(t, "#form-result", schema.HTMX.Target)

	// Test Alpine.js configuration
	assert.NotNil(t, schema.Alpine)
	assert.True(t, schema.Alpine.Enabled)
	assert.Contains(t, schema.Alpine.XData, "formData")

	// Test fields from mixins
	auditFields := []string{"created_at", "updated_at", "created_by", "updated_by"}
	for _, fieldName := range auditFields {
		field, exists := schema.GetField(fieldName)
		assert.True(t, exists, "Missing audit field: %s", fieldName)
		assert.True(t, field.Readonly, "Audit field should be readonly: %s", fieldName)
	}

	statusFields := []string{"status", "is_active", "is_deleted"}
	for _, fieldName := range statusFields {
		_, exists := schema.GetField(fieldName)
		assert.True(t, exists, "Missing status field: %s", fieldName)
	}

	// Test custom fields
	emailField, exists := schema.GetField("email")
	assert.True(t, exists)
	assert.Equal(t, FieldEmail, emailField.Type)
	assert.True(t, emailField.Required)

	salaryField, exists := schema.GetField("salary")
	assert.True(t, exists)
	assert.Equal(t, FieldNumber, salaryField.Type)
	assert.NotNil(t, salaryField.Validation)
	assert.NotNil(t, salaryField.Validation.Min)
	assert.NotNil(t, salaryField.Validation.Max)

	// Test repeatable field
	skillsField, exists := schema.GetField("skills")
	assert.True(t, exists)
	assert.Equal(t, FieldRepeatable, skillsField.Type)

	// Test layout configuration
	assert.NotNil(t, schema.Layout)
	assert.Equal(t, LayoutTabs, schema.Layout.Type)
	assert.Len(t, schema.Layout.Tabs, 3)

	tabs := schema.Layout.Tabs
	assert.Equal(t, "personal", tabs[0].ID)
	assert.Equal(t, "skills", tabs[1].ID)
	assert.Equal(t, "system", tabs[2].ID)

	// Test actions
	assert.Len(t, schema.Actions, 4) // submit, preview, validate, export
	
	submitAction := schema.Actions[0]
	assert.Equal(t, "submit", submitAction.ID)
	assert.Equal(t, ActionSubmit, submitAction.Type)
	assert.Equal(t, "primary", submitAction.Variant)

	// Test internationalization
	assert.NotNil(t, schema.I18n)
	assert.True(t, schema.I18n.Enabled)
	assert.Equal(t, "en-US", schema.I18n.DefaultLocale)
	assert.Contains(t, schema.I18n.SupportedLocales, "es-ES")
	assert.Contains(t, schema.I18n.SupportedLocales, "fr-FR")
}

func TestInvoiceFormWithLineItems(t *testing.T) {
	schema, err := CreateInvoiceFormWithLineItems()
	require.NoError(t, err)
	require.NotNil(t, schema)

	// Test workflow configuration
	assert.NotNil(t, schema.Workflow)
	assert.True(t, schema.Workflow.Enabled)
	assert.Len(t, schema.Workflow.Actions, 3) // submit, approve, reject

	// Test approval configuration
	assert.NotNil(t, schema.Workflow.Approvals)
	assert.True(t, schema.Workflow.Approvals.Required)
	assert.Equal(t, 1, schema.Workflow.Approvals.MinApprovals)
	assert.Len(t, schema.Workflow.Approvals.Approvers, 1)

	approver := schema.Workflow.Approvals.Approvers[0]
	assert.Equal(t, "role", approver.Type)
	assert.Equal(t, "finance_manager", approver.Value)

	// Test line items field
	lineItemsField, exists := schema.GetField("line_items")
	assert.True(t, exists)
	assert.Equal(t, FieldRepeatable, lineItemsField.Type)
}

func TestDynamicSchemaWithBusinessRules(t *testing.T) {
	err := DemonstrateDynamicSchemaWithBusinessRules()
	assert.NoError(t, err)
}

func TestSchemaValidation(t *testing.T) {
	schema, err := CreateComprehensiveExampleSchema()
	require.NoError(t, err)

	// Test comprehensive validation
	err = schema.Validate()
	assert.NoError(t, err)

	// Test form data validation with registries
	ctx := context.Background()
	
	// Create validation registries
	fieldRegistry := NewValidationRegistry()
	crossFieldRegistry := NewCrossFieldValidationRegistry()

	// Valid data
	validData := map[string]any{
		"email":      "john.doe@company.com",
		"department": "engineering",
		"salary":     75000,
		"status":     "active", // Required field from status mixin
		"skills": []map[string]any{
			{
				"skill_name":        "Go Programming",
				"proficiency":       "advanced",
				"years_experience":  5,
			},
		},
	}

	err = ValidateSchemaWithRegistries(ctx, schema, validData, fieldRegistry, crossFieldRegistry)
	assert.NoError(t, err)

	// Invalid data - missing required field
	invalidData := map[string]any{
		"department": "engineering",
		"salary":     75000,
		// missing email
	}

	err = ValidateSchemaWithRegistries(ctx, schema, invalidData, fieldRegistry, crossFieldRegistry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestSchemaCloning(t *testing.T) {
	original, err := CreateComprehensiveExampleSchema()
	require.NoError(t, err)

	cloned := original.Clone()
	assert.Equal(t, original.ID, cloned.ID)
	assert.Equal(t, original.Title, cloned.Title)
	assert.Len(t, cloned.Fields, len(original.Fields))
	assert.Len(t, cloned.Actions, len(original.Actions))

	// Modify clone to ensure it's independent
	cloned.Title = "Modified Title"
	assert.NotEqual(t, original.Title, cloned.Title)
}

func TestSchemaSerializationRoundTrip(t *testing.T) {
	original, err := CreateComprehensiveExampleSchema()
	require.NoError(t, err)

	// Test JSON marshaling
	data, err := original.MarshalJSON()
	assert.NoError(t, err)
	assert.Greater(t, len(data), 0)

	// Test JSON unmarshaling
	var restored Schema
	err = restored.UnmarshalJSON(data)
	assert.NoError(t, err)
	
	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, original.Title, restored.Title)
}

func TestComprehensiveErrorHandling(t *testing.T) {
	// Test error handling in schema creation
	builder := NewBuilder("", TypeForm, "") // Invalid ID and title
	_, err := builder.Build()
	assert.Error(t, err)
	
	// Test error collection
	collector := NewErrorCollector()
	collector.AddValidationError("field1", "required", "Field is required")
	collector.AddValidationError("field2", "invalid", "Field is invalid")
	
	assert.True(t, collector.HasErrors())
	
	errors := collector.Errors()
	assert.Equal(t, 2, errors.Count())
	
	fieldErrors := errors.ErrorsByField()
	assert.Contains(t, fieldErrors, "field1")
	assert.Contains(t, fieldErrors, "field2")
}

func TestDesignTokensIntegration(t *testing.T) {
	// Test that design tokens are available
	tokens := GetDefaultTokens()
	assert.NotNil(t, tokens)
	
	// Test spacing tokens
	assert.Equal(t, "1rem", tokens.Spacing.MD)
	assert.Equal(t, "0.5rem", tokens.Spacing.SM)
	assert.Equal(t, "2rem", tokens.Spacing.LG)
	
	// Test color tokens
	assert.NotEmpty(t, tokens.Colors.Text.Default)
	assert.NotEmpty(t, tokens.Colors.Background.Default)
	
	// Test token registry
	registry := NewTokenRegistry()
	spacing := registry.GetSpacing("md")
	assert.Equal(t, "1rem", spacing)
	
	color := registry.GetColor("text", "default")
	assert.NotEmpty(t, color)
	
	// Test convenience functions
	assert.Equal(t, "1rem", SpacingMD())
	assert.Equal(t, "0.5rem", SpacingSM())
	assert.Equal(t, "2rem", SpacingLG())
}

func TestCompleteIntegration(t *testing.T) {
	// Create a schema using all major features
	builder := NewBuilder("integration_test", TypeForm, "Integration Test")
	
	// Apply all types of configuration
	builder.WithMixin("audit_fields").
		WithCSRF().
		WithHTMX("/test", "#result").
		WithI18n("en-US")
	
	// Add fields with validation
	builder.AddTextField("name", "Full Name", true).
		AddEmailField("email", "Email", true).
		AddSelectField("role", "Role", true, []Option{
			CreateOption("admin", "Administrator"),
			CreateOption("user", "User"),
		})
	
	// Add business rules
	visibilityRule, _ := CreateFieldVisibilityRule(
		"admin_fields",
		"admin_panel",
		nil, // No condition - always show
		true,
	)
	builder.WithBusinessRule(visibilityRule)
	
	// Add custom validation
	builder.WithCustomValidator("test_validator", func(ctx context.Context, value any, params map[string]any) error {
		return nil // Always pass
	})
	
	// Build and validate
	schema, err := builder.Build()
	require.NoError(t, err)
	
	// Test that all features are present
	assert.NotNil(t, schema.Security.CSRF)
	assert.NotNil(t, schema.HTMX)
	assert.NotNil(t, schema.I18n)
	assert.Greater(t, len(schema.Fields), 3) // audit fields + custom fields
	
	// Test validation
	ctx := context.Background()
	fieldRegistry := builder.GetValidationRegistry()
	err = fieldRegistry.Validate(ctx, "test_validator", "test_value", nil)
	assert.NoError(t, err)
	
	t.Log("✅ Complete integration test passed - all schema features working together")
}