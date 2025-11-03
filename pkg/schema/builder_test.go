package schema

import (
	"context"
	"testing"

	"github.com/niiniyare/erp/pkg/condition"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BuilderFoundationTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *BuilderFoundationTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func TestBuilderFoundationTestSuite(t *testing.T) {
	suite.Run(t, new(BuilderFoundationTestSuite))
}

// Test basic builder creation and Foundation components initialization
func (suite *BuilderFoundationTestSuite) TestNewBuilder() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	require.NotNil(suite.T(), builder)
	require.NotNil(suite.T(), builder.schema)
	require.NotNil(suite.T(), builder.mixinSupport)
	require.NotNil(suite.T(), builder.validator)
	require.NotNil(suite.T(), builder.ruleEngine)

	require.Equal(suite.T(), "test-schema", builder.schema.ID)
	require.Equal(suite.T(), TypeForm, builder.schema.Type)
	require.Equal(suite.T(), "Test Schema", builder.schema.Title)
}

// Test mixin support in builder
func (suite *BuilderFoundationTestSuite) TestBuilderMixinSupport() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	// Test applying built-in audit mixin
	builder.WithMixin("audit_fields")

	schema, err := builder.Build()
	require.NoError(suite.T(), err)

	// Should have audit fields
	hasCreatedAt := false
	hasUpdatedAt := false
	for _, field := range schema.Fields {
		if field.Name == "created_at" {
			hasCreatedAt = true
		}
		if field.Name == "updated_at" {
			hasUpdatedAt = true
		}
	}
	require.True(suite.T(), hasCreatedAt, "Should have created_at field from audit mixin")
	require.True(suite.T(), hasUpdatedAt, "Should have updated_at field from audit mixin")

	// Check metadata
	require.NotNil(suite.T(), schema.Meta)
	require.NotNil(suite.T(), schema.Meta.CustomData)
	appliedMixins, exists := schema.Meta.CustomData["applied_mixins"]
	require.True(suite.T(), exists)
	require.Contains(suite.T(), appliedMixins, "audit_fields")
}

// Test custom mixin support
func (suite *BuilderFoundationTestSuite) TestBuilderCustomMixin() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	customMixin := &Mixin{
		ID:          "custom_test",
		Name:        "Custom Test Mixin",
		Description: "A test mixin",
		Fields: []Field{
			{Name: "custom_field", Type: FieldText, Label: "Custom Field"},
		},
		Actions: []Action{
			{ID: "custom_action", Type: ActionButton, Text: "Custom Action"},
		},
	}

	builder.WithCustomMixin(customMixin)

	schema, err := builder.Build()
	require.NoError(suite.T(), err)

	// Check custom field was added
	hasCustomField := false
	for _, field := range schema.Fields {
		if field.Name == "custom_field" {
			hasCustomField = true
			require.Equal(suite.T(), "Custom Field", field.Label)
		}
	}
	require.True(suite.T(), hasCustomField, "Should have custom field from mixin")

	// Check custom action was added
	hasCustomAction := false
	for _, action := range schema.Actions {
		if action.ID == "custom_action" {
			hasCustomAction = true
			require.Equal(suite.T(), "Custom Action", action.Text)
		}
	}
	require.True(suite.T(), hasCustomAction, "Should have custom action from mixin")
}

// Test repeatable field support
func (suite *BuilderFoundationTestSuite) TestBuilderRepeatableSupport() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	// Create a repeatable field for invoice line items
	repeatableField := &RepeatableField{
		Field: Field{
			Name:  "line_items",
			Type:  FieldRepeatable,
			Label: "Line Items",
		},
		Template: []Field{
			{Name: "product", Type: FieldText, Label: "Product", Required: true},
			{Name: "quantity", Type: FieldNumber, Label: "Quantity", Required: true},
			{Name: "price", Type: FieldNumber, Label: "Unit Price", Required: true},
		},
		MinItems: 1,
		MaxItems: 50,
	}

	builder.WithRepeatable(repeatableField)

	schema, err := builder.Build()
	require.NoError(suite.T(), err)

	// Check that the repeatable field was added
	hasRepeatableField := false
	for _, field := range schema.Fields {
		if field.Name == "line_items" {
			hasRepeatableField = true
			require.Equal(suite.T(), FieldRepeatable, field.Type)
		}
	}
	require.True(suite.T(), hasRepeatableField, "Should have repeatable field")
}

// Test validation registry support
func (suite *BuilderFoundationTestSuite) TestBuilderValidationSupport() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	// Add a custom validator
	builder.WithCustomValidator("custom_test", func(ctx context.Context, value any, params map[string]any) error {
		if value == "invalid" {
			return NewValidationError("custom_error", "custom validation failed")
		}
		return nil
	})

	// Test async validator
	asyncValidator := &AsyncValidator{
		Name: "async_test",
		Validate: func(ctx context.Context, value any, params map[string]any) error {
			return nil
		},
	}
	builder.WithAsyncValidator(asyncValidator)

	registry := builder.GetValidationRegistry()
	require.NotNil(suite.T(), registry)

	// Test custom validator
	err := registry.Validate(suite.ctx, "custom_test", "valid", nil)
	require.NoError(suite.T(), err)

	err = registry.Validate(suite.ctx, "custom_test", "invalid", nil)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "custom validation failed")

	// Test async validator
	err = registry.ValidateAsync(suite.ctx, "async_test", "test", nil)
	require.NoError(suite.T(), err)
}

// Test custom validation registry
func (suite *BuilderFoundationTestSuite) TestBuilderCustomValidationRegistry() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	customRegistry := NewValidationRegistry()
	customRegistry.Register("special_validator", func(ctx context.Context, value any, params map[string]any) error {
		if value == "special" {
			return nil
		}
		return NewValidationError("not_special", "value must be special")
	})

	builder.WithValidationRegistry(customRegistry)

	registry := builder.GetValidationRegistry()
	require.Equal(suite.T(), customRegistry, registry)

	err := registry.Validate(suite.ctx, "special_validator", "special", nil)
	require.NoError(suite.T(), err)

	err = registry.Validate(suite.ctx, "special_validator", "not_special", nil)
	require.Error(suite.T(), err)
}

// Test business rule support
func (suite *BuilderFoundationTestSuite) TestBuilderBusinessRuleSupport() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	// Add a field to work with
	builder.AddTextField("status", "Status", true)
	builder.AddTextField("conditional_field", "Conditional Field", false)

	// Create a business rule
	rule := &BusinessRule{
		ID:      "visibility_rule",
		Name:    "Visibility Rule",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: ActionShowField, Target: "conditional_field"},
		},
	}

	builder.WithBusinessRule(rule)

	ruleEngine := builder.GetBusinessRuleEngine()
	require.NotNil(suite.T(), ruleEngine)

	retrievedRule, exists := ruleEngine.GetRule("visibility_rule")
	require.True(suite.T(), exists)
	require.Equal(suite.T(), "Visibility Rule", retrievedRule.Name)
}

// Test business rule builder support
func (suite *BuilderFoundationTestSuite) TestBuilderBusinessRuleBuilderSupport() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	// Add fields to work with
	builder.AddTextField("amount", "Amount", true)
	builder.AddTextField("tax", "Tax", false)

	// Create a business rule using the builder
	ruleBuilder := NewBusinessRule("calculation_rule", "Tax Calculation", RuleTypeDataCalculation).
		WithDescription("Calculate tax based on amount").
		WithAction(ActionCalculate, "tax", nil)

	builder.WithBusinessRuleBuilder(ruleBuilder)

	ruleEngine := builder.GetBusinessRuleEngine()
	rule, exists := ruleEngine.GetRule("calculation_rule")
	require.True(suite.T(), exists)
	require.Equal(suite.T(), "Tax Calculation", rule.Name)
	require.Equal(suite.T(), "Calculate tax based on amount", rule.Description)
}

// Test applying business rules through builder
func (suite *BuilderFoundationTestSuite) TestBuilderApplyBusinessRules() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	// Add fields
	builder.AddTextField("field1", "Field 1", false)
	builder.AddTextField("field2", "Field 2", false)

	// Add a rule that hides field2
	rule := &BusinessRule{
		ID:      "hide_rule",
		Name:    "Hide Rule",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: ActionHideField, Target: "field2"},
		},
	}

	builder.WithBusinessRule(rule)

	// Apply business rules
	modifiedSchema, err := builder.ApplyBusinessRules(suite.ctx, map[string]any{})
	require.NoError(suite.T(), err)

	// Field2 should be hidden
	for _, field := range modifiedSchema.Fields {
		if field.Name == "field2" {
			require.True(suite.T(), field.Hidden, "Field2 should be hidden by business rule")
		}
	}
}

// Test BuildWithRules method
func (suite *BuilderFoundationTestSuite) TestBuilderBuildWithRules() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	// Add fields
	builder.AddTextField("name", "Name", true)
	builder.AddTextField("optional_field", "Optional Field", false)

	// Add a rule with condition
	conditionRule := &condition.ConditionRule{
		ID: "name_check",
		Left: condition.Expression{
			Type:  condition.ValueTypeField,
			Field: "name",
		},
		Op:    condition.OpEqual,
		Right: "admin",
	}

	ruleCondition := &condition.ConditionGroup{
		ID:          "admin_condition",
		Conjunction: condition.ConjunctionAnd,
		Children:    []any{conditionRule},
	}

	rule := &BusinessRule{
		ID:        "admin_rule",
		Name:      "Admin Rule",
		Type:      RuleTypeFieldVisibility,
		Enabled:   true,
		Condition: ruleCondition,
		Actions: []BusinessRuleAction{
			{Type: ActionShowField, Target: "optional_field"},
		},
	}

	builder.WithBusinessRule(rule)

	// Build with rules for admin user
	adminSchema, err := builder.BuildWithRules(suite.ctx, map[string]any{"name": "admin"})
	require.NoError(suite.T(), err)

	// Optional field should be visible for admin
	for _, field := range adminSchema.Fields {
		if field.Name == "optional_field" {
			require.False(suite.T(), field.Hidden, "Optional field should be visible for admin")
		}
	}

	// Build with rules for regular user
	userSchema, err := builder.BuildWithRules(suite.ctx, map[string]any{"name": "user"})
	require.NoError(suite.T(), err)

	// Since condition doesn't match, field should remain in its default state
	for _, field := range userSchema.Fields {
		if field.Name == "optional_field" {
			require.False(suite.T(), field.Hidden, "Optional field should remain visible in default state")
		}
	}
}

// Test multiple mixin application
func (suite *BuilderFoundationTestSuite) TestBuilderMultipleMixins() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	// Apply audit mixin twice to test multiple mixin application
	builder.WithMixin("audit_fields")

	schema, err := builder.Build()
	require.NoError(suite.T(), err)

	// Check that we have fields from audit mixin
	fieldNames := make(map[string]bool)
	for _, field := range schema.Fields {
		fieldNames[field.Name] = true
	}

	// Audit fields
	require.True(suite.T(), fieldNames["created_at"], "Should have created_at from audit mixin")
	require.True(suite.T(), fieldNames["updated_at"], "Should have updated_at from audit mixin")
	require.True(suite.T(), fieldNames["created_by"], "Should have created_by from audit mixin")
	require.True(suite.T(), fieldNames["updated_by"], "Should have updated_by from audit mixin")

	// Check applied mixins metadata
	appliedMixins := schema.Meta.CustomData["applied_mixins"].([]string)
	require.Contains(suite.T(), appliedMixins, "audit_fields")
}

// Test Foundation feature integration
func (suite *BuilderFoundationTestSuite) TestFoundationIntegration() {
	builder := NewBuilder("invoice-form", TypeForm, "Invoice Form")

	// Apply audit mixin for tracking
	builder.WithMixin("audit_fields")

	// Add invoice fields
	builder.AddTextField("invoice_number", "Invoice Number", true).
		AddTextField("customer_name", "Customer Name", true).
		AddNumberField("total", "Total Amount", true, nil, nil)

	// Add line items as repeatable field
	lineItemField := &RepeatableField{
		Field: Field{
			Name:  "line_items",
			Type:  FieldRepeatable,
			Label: "Line Items",
		},
		Template: []Field{
			{Name: "product", Type: FieldText, Label: "Product", Required: true},
			{Name: "quantity", Type: FieldNumber, Label: "Quantity", Required: true},
			{Name: "unit_price", Type: FieldNumber, Label: "Unit Price", Required: true},
			{Name: "line_total", Type: FieldNumber, Label: "Line Total", Readonly: true},
		},
		MinItems: 1,
		MaxItems: 100,
	}

	builder.WithRepeatable(lineItemField)

	// Add custom validator for invoice number format
	builder.WithCustomValidator("invoice_format", func(ctx context.Context, value any, params map[string]any) error {
		str, ok := value.(string)
		if !ok {
			return NewValidationError("invalid_type", "invoice number must be string")
		}
		if len(str) < 3 || str[:3] != "INV" {
			return NewValidationError("invalid_format", "invoice number must start with INV")
		}
		return nil
	})

	// Add business rule for automatic total calculation
	calcRule, buildErr := NewBusinessRule("total_calculation", "Total Calculation", RuleTypeDataCalculation).
		WithDescription("Calculate total from line items").
		WithAction(ActionCalculate, "total", nil).
		Build()
	require.NoError(suite.T(), buildErr)

	builder.WithBusinessRule(calcRule)

	// Add CSRF protection
	builder.WithCSRF()

	// Add rate limiting
	builder.WithRateLimit(10, 60)

	// Add submit button
	builder.AddSubmitButton("Create Invoice")

	// Build the complete schema
	schema, err := builder.Build()
	require.NoError(suite.T(), err)

	// Verify the schema has all expected components
	require.Equal(suite.T(), "invoice-form", schema.ID)
	require.Equal(suite.T(), "Invoice Form", schema.Title)

	// Check fields (audit + custom + repeatable)
	require.GreaterOrEqual(suite.T(), len(schema.Fields), 7) // at least audit(2) + custom(3) + repeatable(1) + line items template(1)

	// Check security configuration
	require.NotNil(suite.T(), schema.Security)
	require.NotNil(suite.T(), schema.Security.CSRF)
	require.True(suite.T(), schema.Security.CSRF.Enabled)
	require.NotNil(suite.T(), schema.Security.RateLimit)
	require.True(suite.T(), schema.Security.RateLimit.Enabled)

	// Check actions
	require.Len(suite.T(), schema.Actions, 1)
	require.Equal(suite.T(), "submit", schema.Actions[0].ID)

	// Check validation registry has custom validator
	registry := builder.GetValidationRegistry()
	validationErr := registry.Validate(suite.ctx, "invoice_format", "INV001", nil)
	require.NoError(suite.T(), validationErr)

	validationErr = registry.Validate(suite.ctx, "invoice_format", "ABC001", nil)
	require.Error(suite.T(), validationErr)

	// Check business rule engine has the calculation rule
	ruleEngine := builder.GetBusinessRuleEngine()
	rule, exists := ruleEngine.GetRule("total_calculation")
	require.True(suite.T(), exists)
	require.Equal(suite.T(), "Total Calculation", rule.Name)
}

// Test error handling in Foundation features
func (suite *BuilderFoundationTestSuite) TestFoundationErrorHandling() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	// Test non-existent mixin (should not crash)
	builder.WithMixin("non_existent_mixin")
	schema, err := builder.Build()
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), schema)

	// Test invalid business rule (builder should handle gracefully)
	invalidRule, err := NewBusinessRule("", "", RuleTypeFieldVisibility).Build()
	require.Error(suite.T(), err) // Should fail at build time
	require.Nil(suite.T(), invalidRule)

	// Test builder with invalid business rule builder (should not add rule)
	invalidRuleBuilder := NewBusinessRule("", "Invalid Rule", RuleTypeFieldVisibility)
	builder.WithBusinessRuleBuilder(invalidRuleBuilder) // Should not add due to validation failure

	ruleEngine := builder.GetBusinessRuleEngine()
	rules := ruleEngine.ListRules()
	require.Len(suite.T(), rules, 0, "Invalid rule should not be added")
}

// Test accessor methods
func (suite *BuilderFoundationTestSuite) TestBuilderAccessors() {
	builder := NewBuilder("test-schema", TypeForm, "Test Schema")

	// Test mixin registry accessor
	mixinRegistry := builder.GetMixinRegistry()
	require.NotNil(suite.T(), mixinRegistry)
	require.IsType(suite.T(), &MixinRegistry{}, mixinRegistry)

	// Test validation registry accessor
	validationRegistry := builder.GetValidationRegistry()
	require.NotNil(suite.T(), validationRegistry)
	require.IsType(suite.T(), &ValidationRegistry{}, validationRegistry)

	// Test business rule engine accessor
	ruleEngine := builder.GetBusinessRuleEngine()
	require.NotNil(suite.T(), ruleEngine)
	require.IsType(suite.T(), &BusinessRuleEngine{}, ruleEngine)
}

// Test builder fluent interface with Foundation features
func (suite *BuilderFoundationTestSuite) TestBuilderFluentInterface() {
	// Test that all Foundation methods return *Builder for chaining
	schema, err := NewBuilder("fluent-test", TypeForm, "Fluent Test").
		WithDescription("Testing fluent interface").
		WithMixin("audit_fields").
		WithCustomValidator("test", func(ctx context.Context, value any, params map[string]any) error {
			return nil
		}).
		WithBusinessRule(&BusinessRule{
			ID:      "test_rule",
			Name:    "Test Rule",
			Type:    RuleTypeFieldVisibility,
			Enabled: true,
			Actions: []BusinessRuleAction{
				{Type: ActionShowField, Target: "test_field"},
			},
		}).
		AddTextField("test_field", "Test Field", false).
		AddSubmitButton("Submit").
		Build()

	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), schema)
	require.Equal(suite.T(), "fluent-test", schema.ID)
	require.Equal(suite.T(), "Testing fluent interface", schema.Description)
	require.Len(suite.T(), schema.Fields, 5) // audit fields (4) + test field (1)
	require.Len(suite.T(), schema.Actions, 1)
}
