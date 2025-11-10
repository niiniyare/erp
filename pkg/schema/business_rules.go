package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/niiniyare/erp/pkg/condition"
)

// BusinessRule represents a business logic rule that can be applied to schemas
type BusinessRule struct {
	ID          string                    `json:"id" validate:"required,min=1,max=100"`
	Name        string                    `json:"name" validate:"required,min=1,max=200"`
	Description string                    `json:"description,omitempty"`
	Type        BusinessRuleType          `json:"type" validate:"required"`
	Priority    int                       `json:"priority,omitempty"`
	Enabled     bool                      `json:"enabled"`
	Condition   *condition.ConditionGroup `json:"condition,omitempty"`
	Actions     []BusinessRuleAction      `json:"actions" validate:"dive"`
	Metadata    map[string]any            `json:"metadata,omitempty"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

// BusinessRuleType defines the type of business rule
type BusinessRuleType string

const (
	// Field-level rules
	RuleTypeFieldVisibility BusinessRuleType = "field_visibility"
	RuleTypeFieldRequired   BusinessRuleType = "field_required"
	RuleTypeFieldDefault    BusinessRuleType = "field_default"
	RuleTypeFieldValidation BusinessRuleType = "field_validation"
	RuleTypeFieldOptions    BusinessRuleType = "field_options"

	// Action-level rules
	RuleTypeActionVisibility BusinessRuleType = "action_visibility"
	RuleTypeActionEnabled    BusinessRuleType = "action_enabled"

	// Schema-level rules
	RuleTypeSchemaTransform BusinessRuleType = "schema_transform"
	RuleTypeSchemaLayout    BusinessRuleType = "schema_layout"

	// Data rules
	RuleTypeDataCalculation BusinessRuleType = "data_calculation"
	RuleTypeDataValidation  BusinessRuleType = "data_validation"
	RuleTypeDataFormat      BusinessRuleType = "data_format"

	// Workflow rules
	RuleTypeWorkflowTrigger  BusinessRuleType = "workflow_trigger"
	RuleTypeWorkflowApproval BusinessRuleType = "workflow_approval"
)

// BusinessRuleAction defines what happens when a rule condition is met
type BusinessRuleAction struct {
	Type   BusinessRuleActionType `json:"type" validate:"required"`
	Target string                 `json:"target,omitempty"`
	Value  any                    `json:"value,omitempty"`
	Params map[string]any         `json:"params,omitempty"`
}

// BusinessRuleActionType defines the type of action to take
type BusinessRuleActionType string

const (
	// Field actions
	ActionShowField     BusinessRuleActionType = "show_field"
	ActionHideField     BusinessRuleActionType = "hide_field"
	ActionRequireField  BusinessRuleActionType = "require_field"
	ActionOptionalField BusinessRuleActionType = "optional_field"
	ActionSetDefault    BusinessRuleActionType = "set_default"
	ActionSetOptions    BusinessRuleActionType = "set_options"
	ActionCalculate     BusinessRuleActionType = "calculate"

	// Action actions
	ActionShowButton    BusinessRuleActionType = "show_button"
	ActionHideButton    BusinessRuleActionType = "hide_button"
	ActionEnableButton  BusinessRuleActionType = "enable_button"
	ActionDisableButton BusinessRuleActionType = "disable_button"

	// Schema actions
	ActionSetLayout     BusinessRuleActionType = "set_layout"
	ActionAddField      BusinessRuleActionType = "add_field"
	ActionRemoveField   BusinessRuleActionType = "remove_field"
	ActionTransformData BusinessRuleActionType = "transform_data"

	// Validation actions
	ActionAddValidation    BusinessRuleActionType = "add_validation"
	ActionRemoveValidation BusinessRuleActionType = "remove_validation"

	// Workflow actions
	ActionTriggerWorkflow BusinessRuleActionType = "trigger_workflow"
	ActionRequestApproval BusinessRuleActionType = "request_approval"
	ActionNotify          BusinessRuleActionType = "notify"
)

// BusinessRuleEngine manages and executes business rules
type BusinessRuleEngine struct {
	rules     map[string]*BusinessRule
	evaluator *condition.Evaluator
	mu        sync.RWMutex
}

// NewBusinessRuleEngine creates a new business rule engine
func NewBusinessRuleEngine() *BusinessRuleEngine {
	// Initialize condition evaluator with business rule specific config
	config := &condition.Config{
		MaxLevel: 10,
		Types: map[string]condition.TypeConfig{
			"text": {
				Operators: []condition.OperatorType{
					condition.OpEqual, condition.OpNotEqual, condition.OpContains,
					condition.OpStartsWith, condition.OpEndsWith, condition.OpIsEmpty,
					condition.OpIsNotEmpty, condition.OpMatchRegexp,
				},
				ValueTypes: []condition.ValueType{condition.ValueTypeValue, condition.ValueTypeField},
			},
			"number": {
				Operators: []condition.OperatorType{
					condition.OpEqual, condition.OpNotEqual, condition.OpLess,
					condition.OpLessOrEqual, condition.OpGreater, condition.OpGreaterOrEqual,
					condition.OpBetween, condition.OpNotBetween,
				},
				ValueTypes: []condition.ValueType{condition.ValueTypeValue, condition.ValueTypeField},
			},
			"boolean": {
				Operators: []condition.OperatorType{
					condition.OpEqual, condition.OpNotEqual,
				},
				ValueTypes: []condition.ValueType{condition.ValueTypeValue, condition.ValueTypeField},
			},
			"date": {
				Operators: []condition.OperatorType{
					condition.OpEqual, condition.OpNotEqual, condition.OpLess,
					condition.OpLessOrEqual, condition.OpGreater, condition.OpGreaterOrEqual,
					condition.OpBetween, condition.OpNotBetween,
				},
				ValueTypes: []condition.ValueType{condition.ValueTypeValue, condition.ValueTypeField},
			},
		},
	}

	evaluator := condition.NewEvaluator(config, condition.DefaultEvalOptions())

	return &BusinessRuleEngine{
		rules:     make(map[string]*BusinessRule),
		evaluator: evaluator,
	}
}

// AddRule adds a business rule to the engine
func (bre *BusinessRuleEngine) AddRule(rule *BusinessRule) error {
	if rule == nil {
		return NewValidationError("business_rule", "business rule is required")
	}

	if err := bre.validateRule(rule); err != nil {
		return err
	}

	bre.mu.Lock()
	defer bre.mu.Unlock()

	now := time.Now()
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	rule.UpdatedAt = now

	bre.rules[rule.ID] = rule
	return nil
}

// RemoveRule removes a business rule from the engine
func (bre *BusinessRuleEngine) RemoveRule(ruleID string) error {
	bre.mu.Lock()
	defer bre.mu.Unlock()

	if _, exists := bre.rules[ruleID]; !exists {
		return NewValidationError("rule_not_found", fmt.Sprintf("business rule %s not found", ruleID))
	}

	delete(bre.rules, ruleID)
	return nil
}

// GetRule gets a business rule by ID
func (bre *BusinessRuleEngine) GetRule(ruleID string) (*BusinessRule, bool) {
	bre.mu.RLock()
	defer bre.mu.RUnlock()

	rule, exists := bre.rules[ruleID]
	return rule, exists
}

// ListRules returns all business rules
func (bre *BusinessRuleEngine) ListRules() []*BusinessRule {
	bre.mu.RLock()
	defer bre.mu.RUnlock()

	rules := make([]*BusinessRule, 0, len(bre.rules))
	for _, rule := range bre.rules {
		rules = append(rules, rule)
	}
	return rules
}

// ApplyRules applies business rules to a schema based on provided data with deep copy
func (bre *BusinessRuleEngine) ApplyRules(ctx context.Context, schema *Schema, data map[string]any) (*Schema, error) {
	bre.mu.RLock()
	defer bre.mu.RUnlock()

	// Create a deep copy of the schema to modify
	modifiedSchema := bre.deepCopySchema(schema)

	// Collect applicable rules and sort by priority
	applicableRules := bre.getApplicableRules()

	collector := NewErrorCollector()

	// Apply each rule
	for _, rule := range applicableRules {
		if !rule.Enabled {
			continue
		}

		// Check if rule condition is met
		shouldApply, err := bre.evaluateRuleCondition(ctx, rule, data)
		if err != nil {
			collector.AddValidationError(rule.ID, "condition_evaluation_failed", err.Error())
			continue
		}

		if shouldApply {
			if err := bre.applyRuleActions(ctx, rule, modifiedSchema, data); err != nil {
				if schemaErr, ok := err.(SchemaError); ok {
					collector.AddFieldError(rule.ID, schemaErr)
				} else {
					collector.AddValidationError(rule.ID, "rule_application_failed", err.Error())
				}
			}
		}
	}

	if collector.HasErrors() {
		return schema, collector.Errors()
	}

	return modifiedSchema, nil
}

// deepCopySchema creates a deep copy of a schema preserving evaluators
func (bre *BusinessRuleEngine) deepCopySchema(schema *Schema) *Schema {
	// Use JSON marshal/unmarshal for deep copy
	data, _ := json.Marshal(schema)
	var copy Schema
	json.Unmarshal(data, &copy)

	// Preserve evaluators on fields (they don't serialize)
	for i := range copy.Fields {
		if i < len(schema.Fields) {
			copy.Fields[i].evaluator = schema.Fields[i].evaluator
		}
	}

	return &copy
}

// validateRule validates a business rule configuration
func (bre *BusinessRuleEngine) validateRule(rule *BusinessRule) error {
	collector := NewErrorCollector()

	if rule.ID == "" {
		collector.AddValidationError("id", "required", "rule ID is required")
	}

	if rule.Name == "" {
		collector.AddValidationError("name", "required", "rule name is required")
	}

	if rule.Type == "" {
		collector.AddValidationError("type", "required", "rule type is required")
	}

	if len(rule.Actions) == 0 {
		collector.AddValidationError("actions", "required", "rule must have at least one action")
	}

	// Validate actions
	for i, action := range rule.Actions {
		if action.Type == "" {
			collector.AddValidationError(fmt.Sprintf("actions[%d].type", i), "required", "action type is required")
		}

		// Validate action-specific requirements
		switch action.Type {
		case ActionShowField, ActionHideField, ActionRequireField, ActionOptionalField:
			if action.Target == "" {
				collector.AddValidationError(fmt.Sprintf("actions[%d].target", i), "required", "field actions require target field")
			}
		case ActionSetDefault, ActionCalculate:
			if action.Target == "" {
				collector.AddValidationError(fmt.Sprintf("actions[%d].target", i), "required", "value actions require target field")
			}
		}
	}

	if collector.HasErrors() {
		return collector.Errors()
	}

	return nil
}

// getApplicableRules returns rules sorted by priority
func (bre *BusinessRuleEngine) getApplicableRules() []*BusinessRule {
	rules := make([]*BusinessRule, 0, len(bre.rules))
	for _, rule := range bre.rules {
		rules = append(rules, rule)
	}

	// Sort by priority (higher priority first)
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority > rules[j].Priority
	})

	return rules
}

// evaluateRuleCondition evaluates whether a rule's condition is met
func (bre *BusinessRuleEngine) evaluateRuleCondition(ctx context.Context, rule *BusinessRule, data map[string]any) (bool, error) {
	if rule.Condition == nil {
		return true, nil // No condition means always apply
	}

	evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
	return bre.evaluator.Evaluate(ctx, rule.Condition, evalCtx)
}

// applyRuleActions applies all actions for a rule
func (bre *BusinessRuleEngine) applyRuleActions(ctx context.Context, rule *BusinessRule, schema *Schema, data map[string]any) error {
	collector := NewErrorCollector()

	for i, action := range rule.Actions {
		if err := bre.applyAction(ctx, action, schema, data); err != nil {
			collector.AddValidationError(
				fmt.Sprintf("rule_%s_action_%d", rule.ID, i),
				"action_failed",
				err.Error(),
			)
		}
	}

	if collector.HasErrors() {
		return collector.Errors()
	}

	return nil
}

// applyAction applies a single business rule action
func (bre *BusinessRuleEngine) applyAction(ctx context.Context, action BusinessRuleAction, schema *Schema, data map[string]any) error {
	switch action.Type {
	case ActionShowField:
		return bre.setFieldVisibility(schema, action.Target, true)

	case ActionHideField:
		return bre.setFieldVisibility(schema, action.Target, false)

	case ActionRequireField:
		return bre.setFieldRequired(schema, action.Target, true)

	case ActionOptionalField:
		return bre.setFieldRequired(schema, action.Target, false)

	case ActionSetDefault:
		return bre.setFieldDefault(schema, action.Target, action.Value)

	case ActionSetOptions:
		if options, ok := action.Value.([]Option); ok {
			return bre.setFieldOptions(schema, action.Target, options)
		}
		return NewValidationError("invalid_options", "options must be []Option type")

	case ActionCalculate:
		return bre.calculateFieldValue(ctx, schema, action.Target, action.Params, data)

	case ActionShowButton:
		return bre.setActionVisibility(schema, action.Target, true)

	case ActionHideButton:
		return bre.setActionVisibility(schema, action.Target, false)

	case ActionEnableButton:
		return bre.setActionEnabled(schema, action.Target, true)

	case ActionDisableButton:
		return bre.setActionEnabled(schema, action.Target, false)

	default:
		return NewValidationError("unsupported_action", fmt.Sprintf("action type %s not supported", action.Type))
	}
}

// Field manipulation methods
func (bre *BusinessRuleEngine) setFieldVisibility(schema *Schema, fieldName string, visible bool) error {
	for i := range schema.Fields {
		if schema.Fields[i].Name == fieldName {
			schema.Fields[i].Hidden = !visible
			return nil
		}
	}
	return NewValidationError("field_not_found", fmt.Sprintf("field %s not found", fieldName))
}

func (bre *BusinessRuleEngine) setFieldRequired(schema *Schema, fieldName string, required bool) error {
	for i := range schema.Fields {
		if schema.Fields[i].Name == fieldName {
			schema.Fields[i].Required = required
			return nil
		}
	}
	return NewValidationError("field_not_found", fmt.Sprintf("field %s not found", fieldName))
}

func (bre *BusinessRuleEngine) setFieldDefault(schema *Schema, fieldName string, defaultValue any) error {
	for i := range schema.Fields {
		if schema.Fields[i].Name == fieldName {
			schema.Fields[i].Default = defaultValue
			return nil
		}
	}
	return NewValidationError("field_not_found", fmt.Sprintf("field %s not found", fieldName))
}

func (bre *BusinessRuleEngine) setFieldOptions(schema *Schema, fieldName string, options []Option) error {
	for i := range schema.Fields {
		if schema.Fields[i].Name == fieldName {
			schema.Fields[i].Options = options
			return nil
		}
	}
	return NewValidationError("field_not_found", fmt.Sprintf("field %s not found", fieldName))
}

func (bre *BusinessRuleEngine) calculateFieldValue(ctx context.Context, schema *Schema, fieldName string, params map[string]any, data map[string]any) error {
	formula, exists := params["formula"].(string)
	if !exists {
		return NewValidationError("missing_formula", "calculation requires formula parameter")
	}

	// Find the field
	for i := range schema.Fields {
		if schema.Fields[i].Name == fieldName {
			// Mark as readonly calculated field
			schema.Fields[i].Readonly = true

			// Store formula in config for frontend
			if schema.Fields[i].Config == nil {
				schema.Fields[i].Config = make(map[string]any)
			}
			schema.Fields[i].Config["formula"] = formula
			schema.Fields[i].Config["calculated"] = true

			// Use the condition package's expression evaluator for formula calculation
			evalCtx := condition.NewEvalContext(data, condition.DefaultEvalOptions())
			result, err := bre.evaluator.EvaluateFormula(ctx, formula, evalCtx)
			if err != nil {
				return NewValidationError("formula_evaluation_failed", fmt.Sprintf("failed to evaluate formula %s: %v", formula, err))
			}

			// Set the calculated value as the field default
			schema.Fields[i].Default = result

			return nil
		}
	}

	return NewValidationError("field_not_found", fmt.Sprintf("field %s not found", fieldName))
}

// Action manipulation methods
func (bre *BusinessRuleEngine) setActionVisibility(schema *Schema, actionID string, visible bool) error {
	for i := range schema.Actions {
		if schema.Actions[i].ID == actionID {
			schema.Actions[i].Hidden = !visible
			return nil
		}
	}
	return NewValidationError("action_not_found", fmt.Sprintf("action %s not found", actionID))
}

func (bre *BusinessRuleEngine) setActionEnabled(schema *Schema, actionID string, enabled bool) error {
	for i := range schema.Actions {
		if schema.Actions[i].ID == actionID {
			schema.Actions[i].Disabled = !enabled
			return nil
		}
	}
	return NewValidationError("action_not_found", fmt.Sprintf("action %s not found", actionID))
}

// TestRule tests a rule against multiple data scenarios
func (bre *BusinessRuleEngine) TestRule(ctx context.Context, rule *BusinessRule, schema *Schema, testCases []map[string]any) ([]bool, error) {
	results := make([]bool, len(testCases))

	for i, data := range testCases {
		shouldApply, err := bre.evaluateRuleCondition(ctx, rule, data)
		if err != nil {
			return nil, err
		}
		results[i] = shouldApply
	}

	return results, nil
}

// ExplainRule provides debugging information about a rule
func (bre *BusinessRuleEngine) ExplainRule(ctx context.Context, rule *BusinessRule, data map[string]any) (string, error) {
	explanation := fmt.Sprintf("Rule: %s\n", rule.Name)
	explanation += fmt.Sprintf("Type: %s\n", rule.Type)
	explanation += fmt.Sprintf("Priority: %d\n", rule.Priority)
	explanation += fmt.Sprintf("Enabled: %v\n", rule.Enabled)

	shouldApply, err := bre.evaluateRuleCondition(ctx, rule, data)
	if err != nil {
		return "", err
	}

	explanation += fmt.Sprintf("Condition Met: %v\n", shouldApply)
	explanation += fmt.Sprintf("Actions: %d\n", len(rule.Actions))

	for i, action := range rule.Actions {
		explanation += fmt.Sprintf("  Action %d: %s on %s\n", i+1, action.Type, action.Target)
	}

	return explanation, nil
}

// UpdateRule updates an existing rule
func (bre *BusinessRuleEngine) UpdateRule(rule *BusinessRule) error {
	if rule == nil || rule.ID == "" {
		return NewValidationError("invalid_rule", "rule is required with valid ID")
	}

	if err := bre.validateRule(rule); err != nil {
		return err
	}

	bre.mu.Lock()
	defer bre.mu.Unlock()

	if _, exists := bre.rules[rule.ID]; !exists {
		return NewValidationError("rule_not_found", fmt.Sprintf("business rule %s not found", rule.ID))
	}

	rule.UpdatedAt = time.Now()
	bre.rules[rule.ID] = rule
	return nil
}

// BusinessRuleBuilder provides a fluent interface for building business rules
type BusinessRuleBuilder struct {
	rule *BusinessRule
}

// NewBusinessRule starts building a business rule
func NewBusinessRule(id, name string, ruleType BusinessRuleType) *BusinessRuleBuilder {
	return &BusinessRuleBuilder{
		rule: &BusinessRule{
			ID:       id,
			Name:     name,
			Type:     ruleType,
			Enabled:  true,
			Priority: 0,
			Actions:  []BusinessRuleAction{},
			Metadata: make(map[string]any),
		},
	}
}

// WithDescription sets the rule description
func (brb *BusinessRuleBuilder) WithDescription(desc string) *BusinessRuleBuilder {
	brb.rule.Description = desc
	return brb
}

// WithPriority sets the rule priority
func (brb *BusinessRuleBuilder) WithPriority(priority int) *BusinessRuleBuilder {
	brb.rule.Priority = priority
	return brb
}

// WithCondition sets the rule condition
func (brb *BusinessRuleBuilder) WithCondition(condition *condition.ConditionGroup) *BusinessRuleBuilder {
	brb.rule.Condition = condition
	return brb
}

// WithAction adds an action to the rule
func (brb *BusinessRuleBuilder) WithAction(actionType BusinessRuleActionType, target string, value any) *BusinessRuleBuilder {
	brb.rule.Actions = append(brb.rule.Actions, BusinessRuleAction{
		Type:   actionType,
		Target: target,
		Value:  value,
	})
	return brb
}

// WithActionAndParams adds an action with parameters
func (brb *BusinessRuleBuilder) WithActionAndParams(actionType BusinessRuleActionType, target string, value any, params map[string]any) *BusinessRuleBuilder {
	brb.rule.Actions = append(brb.rule.Actions, BusinessRuleAction{
		Type:   actionType,
		Target: target,
		Value:  value,
		Params: params,
	})
	return brb
}

// Enabled sets whether the rule is enabled
func (brb *BusinessRuleBuilder) Enabled(enabled bool) *BusinessRuleBuilder {
	brb.rule.Enabled = enabled
	return brb
}

// WithMetadata adds metadata to the rule
func (brb *BusinessRuleBuilder) WithMetadata(key string, value any) *BusinessRuleBuilder {
	brb.rule.Metadata[key] = value
	return brb
}

// Build returns the constructed business rule
func (brb *BusinessRuleBuilder) Build() (*BusinessRule, error) {
	// Set timestamps
	now := time.Now()
	if brb.rule.CreatedAt.IsZero() {
		brb.rule.CreatedAt = now
	}
	brb.rule.UpdatedAt = now

	// Validate rule
	engine := &BusinessRuleEngine{}
	if err := engine.validateRule(brb.rule); err != nil {
		return nil, err
	}

	return brb.rule, nil
}

// Helper functions for common business rules

// CreateFieldVisibilityRule creates a rule to show/hide fields based on conditions
func CreateFieldVisibilityRule(id, fieldName string, condition *condition.ConditionGroup, visible bool) (*BusinessRule, error) {
	actionType := ActionShowField
	if !visible {
		actionType = ActionHideField
	}

	return NewBusinessRule(id, fmt.Sprintf("%s visibility rule", fieldName), RuleTypeFieldVisibility).
		WithCondition(condition).
		WithAction(actionType, fieldName, nil).
		Build()
}

// CreateFieldRequiredRule creates a rule to make fields required/optional based on conditions
func CreateFieldRequiredRule(id, fieldName string, condition *condition.ConditionGroup, required bool) (*BusinessRule, error) {
	actionType := ActionRequireField
	if !required {
		actionType = ActionOptionalField
	}

	return NewBusinessRule(id, fmt.Sprintf("%s required rule", fieldName), RuleTypeFieldRequired).
		WithCondition(condition).
		WithAction(actionType, fieldName, nil).
		Build()
}

// CreateCalculationRule creates a rule to calculate field values
func CreateCalculationRule(id, fieldName, formula string, condition *condition.ConditionGroup) (*BusinessRule, error) {
	params := map[string]any{
		"formula": formula,
	}

	return NewBusinessRule(id, fmt.Sprintf("%s calculation rule", fieldName), RuleTypeDataCalculation).
		WithCondition(condition).
		WithActionAndParams(ActionCalculate, fieldName, nil, params).
		Build()
}
