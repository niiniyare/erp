package schema

import (
	"context"
	"fmt"
	"testing"

	"github.com/niiniyare/erp/pkg/condition"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BusinessRulesTestSuite struct {
	suite.Suite
	engine *BusinessRuleEngine
	ctx    context.Context
}

func (suite *BusinessRulesTestSuite) SetupTest() {
	suite.engine = NewBusinessRuleEngine()
	suite.ctx = context.Background()
}

func TestBusinessRulesTestSuite(t *testing.T) {
	suite.Run(t, new(BusinessRulesTestSuite))
}

// Test BusinessRuleEngine creation
func (suite *BusinessRulesTestSuite) TestNewBusinessRuleEngine() {
	require.NotNil(suite.T(), suite.engine)
	require.NotNil(suite.T(), suite.engine.rules)
	require.NotNil(suite.T(), suite.engine.evaluator)
	require.Equal(suite.T(), 0, len(suite.engine.rules))
}

// Test business rule creation and validation
func (suite *BusinessRulesTestSuite) TestBusinessRuleValidation() {
	// Valid rule
	rule := &BusinessRule{
		ID:      "test_rule",
		Name:    "Test Rule",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{
				Type:   ActionShowField,
				Target: "test_field",
			},
		},
	}

	err := suite.engine.AddRule(rule)
	require.NoError(suite.T(), err)

	// Invalid rule - missing ID
	invalidRule := &BusinessRule{
		Name:    "Invalid Rule",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{
				Type:   ActionShowField,
				Target: "test_field",
			},
		},
	}

	err = suite.engine.AddRule(invalidRule)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "rule ID is required")

	// Invalid rule - missing name
	invalidRule2 := &BusinessRule{
		ID:      "test_rule_2",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{
				Type:   ActionShowField,
				Target: "test_field",
			},
		},
	}

	err = suite.engine.AddRule(invalidRule2)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "rule name is required")

	// Invalid rule - missing actions
	invalidRule3 := &BusinessRule{
		ID:      "test_rule_3",
		Name:    "Test Rule 3",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{},
	}

	err = suite.engine.AddRule(invalidRule3)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "rule must have at least one action")

	// Invalid rule - nil rule
	err = suite.engine.AddRule(nil)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "business rule is required")
}

// Test action validation
func (suite *BusinessRulesTestSuite) TestActionValidation() {
	// Invalid action - missing target for field action
	rule := &BusinessRule{
		ID:      "test_rule",
		Name:    "Test Rule",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{
				Type: ActionShowField,
				// Missing Target
			},
		},
	}

	err := suite.engine.AddRule(rule)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "field actions require target field")

	// Invalid action - missing type
	rule2 := &BusinessRule{
		ID:      "test_rule_2",
		Name:    "Test Rule 2",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{
				// Missing Type
				Target: "test_field",
			},
		},
	}

	err = suite.engine.AddRule(rule2)
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "action type is required")
}

// Test rule CRUD operations
func (suite *BusinessRulesTestSuite) TestRuleCRUD() {
	rule := &BusinessRule{
		ID:      "crud_test",
		Name:    "CRUD Test Rule",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{
				Type:   ActionShowField,
				Target: "test_field",
			},
		},
	}

	// Create
	err := suite.engine.AddRule(rule)
	require.NoError(suite.T(), err)
	require.False(suite.T(), rule.CreatedAt.IsZero())
	require.False(suite.T(), rule.UpdatedAt.IsZero())

	// Read
	retrieved, exists := suite.engine.GetRule("crud_test")
	require.True(suite.T(), exists)
	require.Equal(suite.T(), rule.ID, retrieved.ID)
	require.Equal(suite.T(), rule.Name, retrieved.Name)

	// List
	rules := suite.engine.ListRules()
	require.Len(suite.T(), rules, 1)
	require.Equal(suite.T(), "crud_test", rules[0].ID)

	// Delete
	err = suite.engine.RemoveRule("crud_test")
	require.NoError(suite.T(), err)

	// Verify deletion
	_, exists = suite.engine.GetRule("crud_test")
	require.False(suite.T(), exists)

	rules = suite.engine.ListRules()
	require.Len(suite.T(), rules, 0)

	// Delete non-existent rule
	err = suite.engine.RemoveRule("non_existent")
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "business rule non_existent not found")
}

// Test rule priority sorting
func (suite *BusinessRulesTestSuite) TestRulePriority() {
	rules := []*BusinessRule{
		{
			ID:       "low_priority",
			Name:     "Low Priority Rule",
			Type:     RuleTypeFieldVisibility,
			Priority: 1,
			Enabled:  true,
			Actions: []BusinessRuleAction{
				{Type: ActionShowField, Target: "field1"},
			},
		},
		{
			ID:       "high_priority",
			Name:     "High Priority Rule",
			Type:     RuleTypeFieldVisibility,
			Priority: 10,
			Enabled:  true,
			Actions: []BusinessRuleAction{
				{Type: ActionShowField, Target: "field2"},
			},
		},
		{
			ID:       "medium_priority",
			Name:     "Medium Priority Rule",
			Type:     RuleTypeFieldVisibility,
			Priority: 5,
			Enabled:  true,
			Actions: []BusinessRuleAction{
				{Type: ActionShowField, Target: "field3"},
			},
		},
	}

	for _, rule := range rules {
		err := suite.engine.AddRule(rule)
		require.NoError(suite.T(), err)
	}

	sortedRules := suite.engine.getApplicableRules()
	require.Len(suite.T(), sortedRules, 3)

	// Should be sorted by priority (highest first)
	require.Equal(suite.T(), "high_priority", sortedRules[0].ID)
	require.Equal(suite.T(), "medium_priority", sortedRules[1].ID)
	require.Equal(suite.T(), "low_priority", sortedRules[2].ID)
}

// Test field visibility actions
func (suite *BusinessRulesTestSuite) TestFieldVisibilityActions() {
	// Create a test schema
	schema := &Schema{
		ID:    "test_schema",
		Title: "Test Schema",
		Fields: []Field{
			{Name: "field1", Type: FieldText, Hidden: false},
			{Name: "field2", Type: FieldText, Hidden: true},
		},
	}

	// Create rule to show field1 and hide field2
	rule := &BusinessRule{
		ID:      "visibility_rule",
		Name:    "Visibility Rule",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: ActionHideField, Target: "field1"},
			{Type: ActionShowField, Target: "field2"},
		},
	}

	err := suite.engine.AddRule(rule)
	require.NoError(suite.T(), err)

	// Apply rules
	modifiedSchema, err := suite.engine.ApplyRules(suite.ctx, schema, map[string]any{})
	require.NoError(suite.T(), err)

	// Check that field1 is now hidden and field2 is shown
	require.True(suite.T(), modifiedSchema.Fields[0].Hidden)  // field1 should be hidden
	require.False(suite.T(), modifiedSchema.Fields[1].Hidden) // field2 should be shown

	// Test field not found error
	ruleWithInvalidField := &BusinessRule{
		ID:      "invalid_field_rule",
		Name:    "Invalid Field Rule",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: ActionShowField, Target: "non_existent_field"},
		},
	}

	err = suite.engine.AddRule(ruleWithInvalidField)
	require.NoError(suite.T(), err)

	_, err = suite.engine.ApplyRules(suite.ctx, schema, map[string]any{})
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "field non_existent_field not found")
}

// Test field required actions
func (suite *BusinessRulesTestSuite) TestFieldRequiredActions() {
	schema := &Schema{
		ID:    "test_schema",
		Title: "Test Schema",
		Fields: []Field{
			{Name: "field1", Type: FieldText, Required: false},
			{Name: "field2", Type: FieldText, Required: true},
		},
	}

	rule := &BusinessRule{
		ID:      "required_rule",
		Name:    "Required Rule",
		Type:    RuleTypeFieldRequired,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: ActionRequireField, Target: "field1"},
			{Type: ActionOptionalField, Target: "field2"},
		},
	}

	err := suite.engine.AddRule(rule)
	require.NoError(suite.T(), err)

	modifiedSchema, err := suite.engine.ApplyRules(suite.ctx, schema, map[string]any{})
	require.NoError(suite.T(), err)

	require.True(suite.T(), modifiedSchema.Fields[0].Required)   // field1 should be required
	require.False(suite.T(), modifiedSchema.Fields[1].Required) // field2 should be optional
}

// Test field default value actions
func (suite *BusinessRulesTestSuite) TestFieldDefaultActions() {
	schema := &Schema{
		ID:    "test_schema",
		Title: "Test Schema",
		Fields: []Field{
			{Name: "field1", Type: FieldText},
			{Name: "field2", Type: FieldNumber},
		},
	}

	rule := &BusinessRule{
		ID:      "default_rule",
		Name:    "Default Rule",
		Type:    RuleTypeFieldDefault,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: ActionSetDefault, Target: "field1", Value: "default_text"},
			{Type: ActionSetDefault, Target: "field2", Value: 42},
		},
	}

	err := suite.engine.AddRule(rule)
	require.NoError(suite.T(), err)

	modifiedSchema, err := suite.engine.ApplyRules(suite.ctx, schema, map[string]any{})
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), "default_text", modifiedSchema.Fields[0].Default)
	require.Equal(suite.T(), 42, modifiedSchema.Fields[1].Default)
}

// Test field options actions
func (suite *BusinessRulesTestSuite) TestFieldOptionsActions() {
	schema := &Schema{
		ID:    "test_schema",
		Title: "Test Schema",
		Fields: []Field{
			{Name: "select_field", Type: FieldSelect},
		},
	}

	options := []Option{
		{Value: "option1", Label: "Option 1"},
		{Value: "option2", Label: "Option 2"},
	}

	rule := &BusinessRule{
		ID:      "options_rule",
		Name:    "Options Rule",
		Type:    RuleTypeFieldOptions,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: ActionSetOptions, Target: "select_field", Value: options},
		},
	}

	err := suite.engine.AddRule(rule)
	require.NoError(suite.T(), err)

	modifiedSchema, err := suite.engine.ApplyRules(suite.ctx, schema, map[string]any{})
	require.NoError(suite.T(), err)

	require.Len(suite.T(), modifiedSchema.Fields[0].Options, 2)
	require.Equal(suite.T(), "option1", modifiedSchema.Fields[0].Options[0].Value)
	require.Equal(suite.T(), "Option 1", modifiedSchema.Fields[0].Options[0].Label)

	// Test invalid options type
	invalidRule := &BusinessRule{
		ID:      "invalid_options_rule",
		Name:    "Invalid Options Rule",
		Type:    RuleTypeFieldOptions,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: ActionSetOptions, Target: "select_field", Value: "invalid_options"},
		},
	}

	err = suite.engine.AddRule(invalidRule)
	require.NoError(suite.T(), err)

	_, err = suite.engine.ApplyRules(suite.ctx, schema, map[string]any{})
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "options must be []Option type")
}

// Test action visibility
func (suite *BusinessRulesTestSuite) TestActionVisibility() {
	schema := &Schema{
		ID:    "test_schema",
		Title: "Test Schema",
		Actions: []Action{
			{ID: "action1", Text: "Action 1", Hidden: false},
			{ID: "action2", Text: "Action 2", Hidden: true},
		},
	}

	rule := &BusinessRule{
		ID:      "action_visibility_rule",
		Name:    "Action Visibility Rule",
		Type:    RuleTypeActionVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: ActionHideButton, Target: "action1"},
			{Type: ActionShowButton, Target: "action2"},
		},
	}

	err := suite.engine.AddRule(rule)
	require.NoError(suite.T(), err)

	modifiedSchema, err := suite.engine.ApplyRules(suite.ctx, schema, map[string]any{})
	require.NoError(suite.T(), err)

	require.True(suite.T(), modifiedSchema.Actions[0].Hidden)   // action1 should be hidden
	require.False(suite.T(), modifiedSchema.Actions[1].Hidden) // action2 should be shown

	// Test action not found
	invalidRule := &BusinessRule{
		ID:      "invalid_action_rule",
		Name:    "Invalid Action Rule",
		Type:    RuleTypeActionVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: ActionShowButton, Target: "non_existent_action"},
		},
	}

	err = suite.engine.AddRule(invalidRule)
	require.NoError(suite.T(), err)

	_, err = suite.engine.ApplyRules(suite.ctx, schema, map[string]any{})
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "action non_existent_action not found")
}

// Test action enabled/disabled
func (suite *BusinessRulesTestSuite) TestActionEnabled() {
	schema := &Schema{
		ID:    "test_schema",
		Title: "Test Schema",
		Actions: []Action{
			{ID: "action1", Text: "Action 1", Disabled: false},
			{ID: "action2", Text: "Action 2", Disabled: true},
		},
	}

	rule := &BusinessRule{
		ID:      "action_enabled_rule",
		Name:    "Action Enabled Rule",
		Type:    RuleTypeActionEnabled,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: ActionDisableButton, Target: "action1"},
			{Type: ActionEnableButton, Target: "action2"},
		},
	}

	err := suite.engine.AddRule(rule)
	require.NoError(suite.T(), err)

	modifiedSchema, err := suite.engine.ApplyRules(suite.ctx, schema, map[string]any{})
	require.NoError(suite.T(), err)

	require.True(suite.T(), modifiedSchema.Actions[0].Disabled)   // action1 should be disabled
	require.False(suite.T(), modifiedSchema.Actions[1].Disabled) // action2 should be enabled
}

// Test calculation actions
func (suite *BusinessRulesTestSuite) TestCalculationActions() {
	schema := &Schema{
		ID:    "test_schema",
		Title: "Test Schema",
		Fields: []Field{
			{Name: "calculated_field", Type: FieldNumber, Readonly: false},
		},
	}

	rule := &BusinessRule{
		ID:      "calculation_rule",
		Name:    "Calculation Rule",
		Type:    RuleTypeDataCalculation,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{
				Type:   ActionCalculate,
				Target: "calculated_field",
				Params: map[string]any{"formula": "quantity * unit_price"},
			},
		},
	}

	err := suite.engine.AddRule(rule)
	require.NoError(suite.T(), err)

	modifiedSchema, err := suite.engine.ApplyRules(suite.ctx, schema, map[string]any{})
	require.NoError(suite.T(), err)

	require.True(suite.T(), modifiedSchema.Fields[0].Readonly) // Calculated fields should be readonly
}

// Test disabled rules
func (suite *BusinessRulesTestSuite) TestDisabledRules() {
	schema := &Schema{
		ID:    "test_schema",
		Title: "Test Schema",
		Fields: []Field{
			{Name: "field1", Type: FieldText, Hidden: false},
		},
	}

	rule := &BusinessRule{
		ID:      "disabled_rule",
		Name:    "Disabled Rule",
		Type:    RuleTypeFieldVisibility,
		Enabled: false, // Rule is disabled
		Actions: []BusinessRuleAction{
			{Type: ActionHideField, Target: "field1"},
		},
	}

	err := suite.engine.AddRule(rule)
	require.NoError(suite.T(), err)

	modifiedSchema, err := suite.engine.ApplyRules(suite.ctx, schema, map[string]any{})
	require.NoError(suite.T(), err)

	// Field should not be affected since rule is disabled
	require.False(suite.T(), modifiedSchema.Fields[0].Hidden)
}

// Test unsupported action types
func (suite *BusinessRulesTestSuite) TestUnsupportedActionTypes() {
	schema := &Schema{
		ID:    "test_schema",
		Title: "Test Schema",
		Fields: []Field{
			{Name: "field1", Type: FieldText},
		},
	}

	rule := &BusinessRule{
		ID:      "unsupported_rule",
		Name:    "Unsupported Rule",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: BusinessRuleActionType("unsupported_action"), Target: "field1"},
		},
	}

	err := suite.engine.AddRule(rule)
	require.NoError(suite.T(), err)

	_, err = suite.engine.ApplyRules(suite.ctx, schema, map[string]any{})
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "action type unsupported_action not supported")
}

// Test BusinessRuleBuilder
func (suite *BusinessRulesTestSuite) TestBusinessRuleBuilder() {
	conditionRule := &condition.ConditionRule{
		ID: "status_check",
		Left: condition.Expression{
			Type:  condition.ValueTypeField,
			Field: "status",
		},
		Op:    condition.OpEqual,
		Right: "active",
	}

	condition := &condition.ConditionGroup{
		ID:          "status_condition",
		Conjunction: condition.ConjunctionAnd,
		Children:    []any{conditionRule},
	}

	rule, err := NewBusinessRule("builder_test", "Builder Test Rule", RuleTypeFieldVisibility).
		WithDescription("A test rule created with builder").
		WithPriority(5).
		WithCondition(condition).
		WithAction(ActionShowField, "test_field", nil).
		WithMetadata("category", "test").
		Enabled(true).
		Build()

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "builder_test", rule.ID)
	require.Equal(suite.T(), "Builder Test Rule", rule.Name)
	require.Equal(suite.T(), "A test rule created with builder", rule.Description)
	require.Equal(suite.T(), 5, rule.Priority)
	require.True(suite.T(), rule.Enabled)
	require.NotNil(suite.T(), rule.Condition)
	require.Len(suite.T(), rule.Actions, 1)
	require.Equal(suite.T(), ActionShowField, rule.Actions[0].Type)
	require.Equal(suite.T(), "test_field", rule.Actions[0].Target)
	require.Equal(suite.T(), "test", rule.Metadata["category"])
	require.False(suite.T(), rule.CreatedAt.IsZero())
	require.False(suite.T(), rule.UpdatedAt.IsZero())

	// Test WithActionAndParams
	rule2, err := NewBusinessRule("builder_test_2", "Builder Test Rule 2", RuleTypeDataCalculation).
		WithActionAndParams(ActionCalculate, "calc_field", nil, map[string]any{"formula": "a + b"}).
		Build()

	require.NoError(suite.T(), err)
	require.Len(suite.T(), rule2.Actions, 1)
	require.Equal(suite.T(), "a + b", rule2.Actions[0].Params["formula"])

	// Test builder validation
	_, err = NewBusinessRule("", "", RuleTypeFieldVisibility).Build()
	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "rule ID is required")
}

// Test helper functions
func (suite *BusinessRulesTestSuite) TestHelperFunctions() {
	conditionRule := &condition.ConditionRule{
		ID: "status_check",
		Left: condition.Expression{
			Type:  condition.ValueTypeField,
			Field: "status",
		},
		Op:    condition.OpEqual,
		Right: "active",
	}

	condition := &condition.ConditionGroup{
		ID:          "status_condition",
		Conjunction: condition.ConjunctionAnd,
		Children:    []any{conditionRule},
	}

	// Test CreateFieldVisibilityRule
	visibilityRule, err := CreateFieldVisibilityRule("vis_rule", "test_field", condition, true)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "vis_rule", visibilityRule.ID)
	require.Equal(suite.T(), RuleTypeFieldVisibility, visibilityRule.Type)
	require.Equal(suite.T(), ActionShowField, visibilityRule.Actions[0].Type)

	visibilityRule2, err := CreateFieldVisibilityRule("vis_rule_2", "test_field", condition, false)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), ActionHideField, visibilityRule2.Actions[0].Type)

	// Test CreateFieldRequiredRule
	requiredRule, err := CreateFieldRequiredRule("req_rule", "test_field", condition, true)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "req_rule", requiredRule.ID)
	require.Equal(suite.T(), RuleTypeFieldRequired, requiredRule.Type)
	require.Equal(suite.T(), ActionRequireField, requiredRule.Actions[0].Type)

	requiredRule2, err := CreateFieldRequiredRule("req_rule_2", "test_field", condition, false)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), ActionOptionalField, requiredRule2.Actions[0].Type)

	// Test CreateCalculationRule
	calcRule, err := CreateCalculationRule("calc_rule", "total", "quantity * price", condition)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "calc_rule", calcRule.ID)
	require.Equal(suite.T(), RuleTypeDataCalculation, calcRule.Type)
	require.Equal(suite.T(), ActionCalculate, calcRule.Actions[0].Type)
	require.Equal(suite.T(), "quantity * price", calcRule.Actions[0].Params["formula"])
}

// Test rule condition evaluation
func (suite *BusinessRulesTestSuite) TestRuleConditionEvaluation() {
	// Create a condition that checks if status equals "active"
	conditionRule := &condition.ConditionRule{
		ID: "status_check",
		Left: condition.Expression{
			Type:  condition.ValueTypeField,
			Field: "status",
		},
		Op:    condition.OpEqual,
		Right: "active",
	}

	condition := &condition.ConditionGroup{
		ID:          "status_condition",
		Conjunction: condition.ConjunctionAnd,
		Children:    []any{conditionRule},
	}

	rule := &BusinessRule{
		ID:        "conditional_rule",
		Name:      "Conditional Rule",
		Type:      RuleTypeFieldVisibility,
		Enabled:   true,
		Condition: condition,
		Actions: []BusinessRuleAction{
			{Type: ActionShowField, Target: "conditional_field"},
		},
	}

	err := suite.engine.AddRule(rule)
	require.NoError(suite.T(), err)

	schema := &Schema{
		ID:    "test_schema",
		Title: "Test Schema",
		Fields: []Field{
			{Name: "conditional_field", Type: FieldText, Hidden: true},
		},
	}

	// Test with condition met (status = "active")
	data := map[string]any{"status": "active"}
	modifiedSchema, err := suite.engine.ApplyRules(suite.ctx, schema, data)
	require.NoError(suite.T(), err)
	require.False(suite.T(), modifiedSchema.Fields[0].Hidden) // Field should be shown

	// Test with condition not met (status = "inactive")
	data = map[string]any{"status": "inactive"}
	modifiedSchema, err = suite.engine.ApplyRules(suite.ctx, schema, data)
	require.NoError(suite.T(), err)
	require.True(suite.T(), modifiedSchema.Fields[0].Hidden) // Field should remain hidden

	// Test rule without condition (should always apply)
	// Clean the engine first
	suite.engine.RemoveRule("conditional_rule")
	
	ruleWithoutCondition := &BusinessRule{
		ID:      "always_apply_rule",
		Name:    "Always Apply Rule",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: ActionShowField, Target: "conditional_field"},
		},
	}

	err = suite.engine.AddRule(ruleWithoutCondition)
	require.NoError(suite.T(), err)

	modifiedSchema, err = suite.engine.ApplyRules(suite.ctx, schema, map[string]any{})
	require.NoError(suite.T(), err)
	require.False(suite.T(), modifiedSchema.Fields[0].Hidden) // Field should be shown
}

// Test schema modification doesn't affect original
func (suite *BusinessRulesTestSuite) TestSchemaImmutability() {
	originalSchema := &Schema{
		ID:    "test_schema",
		Title: "Test Schema",
		Fields: []Field{
			{Name: "field1", Type: FieldText, Hidden: false},
		},
	}

	rule := &BusinessRule{
		ID:      "modification_rule",
		Name:    "Modification Rule",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: ActionHideField, Target: "field1"},
		},
	}

	err := suite.engine.AddRule(rule)
	require.NoError(suite.T(), err)

	modifiedSchema, err := suite.engine.ApplyRules(suite.ctx, originalSchema, map[string]any{})
	require.NoError(suite.T(), err)

	// Original schema should be unchanged
	require.False(suite.T(), originalSchema.Fields[0].Hidden)
	// Modified schema should have the changes
	require.True(suite.T(), modifiedSchema.Fields[0].Hidden)
}

// Test concurrent access to rules
func (suite *BusinessRulesTestSuite) TestConcurrentAccess() {
	rule := &BusinessRule{
		ID:      "concurrent_rule",
		Name:    "Concurrent Rule",
		Type:    RuleTypeFieldVisibility,
		Enabled: true,
		Actions: []BusinessRuleAction{
			{Type: ActionShowField, Target: "test_field"},
		},
	}

	// Add rule
	err := suite.engine.AddRule(rule)
	require.NoError(suite.T(), err)

	// Test concurrent reads
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			_, exists := suite.engine.GetRule("concurrent_rule")
			require.True(suite.T(), exists)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			rules := suite.engine.ListRules()
			require.Len(suite.T(), rules, 1)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Test concurrent writes
	go func() {
		for i := 0; i < 10; i++ {
			newRule := &BusinessRule{
				ID:      fmt.Sprintf("rule_%d", i),
				Name:    fmt.Sprintf("Rule %d", i),
				Type:    RuleTypeFieldVisibility,
				Enabled: true,
				Actions: []BusinessRuleAction{
					{Type: ActionShowField, Target: "test_field"},
				},
			}
			suite.engine.AddRule(newRule)
		}
		done <- true
	}()

	<-done

	// Should have at least the original rule
	rules := suite.engine.ListRules()
	require.GreaterOrEqual(suite.T(), len(rules), 1)
}