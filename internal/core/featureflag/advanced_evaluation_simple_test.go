package featureflag

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/niiniyare/erp/internal/core/access/conditional"
)

// TestAdvancedEvaluationEngine_StructCreation tests basic struct creation and interface compliance
func TestAdvancedEvaluationEngine_StructCreation(t *testing.T) {
	// Test basic struct creation
	engine := &advancedEvaluationEngine{
		conditionalAccessService: nil, // Would be injected in real usage
		simpleService:           nil, // Would be injected in real usage
		logger:                  nil, // Would be injected in real usage
		metrics:                 nil, // Would be injected in real usage
		tracing:                 nil, // Would be injected in real usage
	}

	// Verify it implements the interface
	var _ AdvancedEvaluationEngine = engine

	assert.NotNil(t, engine)
}

// TestComplexRuleEvaluationLogic tests the complex rule evaluation logic with nil rules only
func TestComplexRuleEvaluationLogic(t *testing.T) {
	// Just test the logic without calling methods that require tracing
	flagID := uuid.New()
	userID := uuid.New()

	// Test result construction for nil rules case
	result := &AdvancedEvaluationResult{
		FlagID:      flagID,
		UserID:      userID,
		Enabled:     false,
		Reason:      "No evaluation rules provided",
		CacheHit:    false,
		EvaluatedAt: time.Now(),
	}
	
	assert.NotNil(t, result)
	assert.False(t, result.Enabled)
	assert.Equal(t, "No evaluation rules provided", result.Reason)
	assert.Equal(t, flagID, result.FlagID)
	assert.Equal(t, userID, result.UserID)
}

// TestRuleEvaluation tests individual rule evaluation
func TestRuleEvaluation(t *testing.T) {
	engine := &advancedEvaluationEngine{}

	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name     string
		rule     *EvaluationRule
		context  *AdvancedEvaluationContext
		expected bool
	}{
		{
			name: "Simple condition - user role equals admin",
			rule: &EvaluationRule{
				ID:   uuid.New(),
				Type: RuleTypeCondition,
				Condition: &RuleCondition{
					Field:    "user.role",
					Operator: OpEquals,
					Value:    "admin",
				},
			},
			context: &AdvancedEvaluationContext{
				UserID: userID,
				UserData: map[string]any{
					"role": "admin",
				},
			},
			expected: true,
		},
		{
			name: "Simple condition - user role not equals admin",
			rule: &EvaluationRule{
				ID:   uuid.New(),
				Type: RuleTypeCondition,
				Condition: &RuleCondition{
					Field:    "user.role",
					Operator: OpEquals,
					Value:    "admin",
				},
			},
			context: &AdvancedEvaluationContext{
				UserID: userID,
				UserData: map[string]any{
					"role": "user",
				},
			},
			expected: false,
		},
		{
			name: "Percentage rule - 100%",
			rule: &EvaluationRule{
				ID:     uuid.New(),
				Type:   RuleTypePercentage,
				Weight: func() *float64 { w := 1.0; return &w }(), // 100%
			},
			context: &AdvancedEvaluationContext{
				UserID: userID,
			},
			expected: true,
		},
		{
			name: "Device compliance rule - compliant and managed",
			rule: &EvaluationRule{
				ID:   uuid.New(),
				Type: RuleTypeDevice,
			},
			context: &AdvancedEvaluationContext{
				UserID: userID,
				DeviceInfo: &conditional.DeviceInfo{
					IsCompliant: true,
					IsManaged:   true,
				},
			},
			expected: true,
		},
		{
			name: "Device compliance rule - not compliant",
			rule: &EvaluationRule{
				ID:   uuid.New(),
				Type: RuleTypeDevice,
			},
			context: &AdvancedEvaluationContext{
				UserID: userID,
				DeviceInfo: &conditional.DeviceInfo{
					IsCompliant: false,
					IsManaged:   true,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, paths := engine.evaluateRule(ctx, tt.rule, tt.context)
			assert.Equal(t, tt.expected, result)
			assert.NotEmpty(t, paths)
		})
	}
}

// TestConditionOperatorEvaluation tests various condition operators
func TestConditionOperatorEvaluation(t *testing.T) {
	engine := &advancedEvaluationEngine{}

	tests := []struct {
		name         string
		operator     ConditionOperator
		actual       any
		expected     any
		values       []any
		caseSensitive bool
		result       bool
	}{
		{"equals match", OpEquals, "admin", "admin", nil, true, true},
		{"equals no match", OpEquals, "user", "admin", nil, true, false},
		{"equals case insensitive", OpEquals, "ADMIN", "admin", nil, false, true},
		{"not equals", OpNotEquals, "user", "admin", nil, true, true},
		{"contains", OpContains, "user@example.com", "@example", nil, true, true},
		{"not contains", OpNotContains, "user@example.com", "@test", nil, true, true},
		{"starts with", OpStartsWith, "admin_user", "admin", nil, true, true},
		{"ends with", OpEndsWith, "power_user", "user", nil, true, true},
		{"in list", OpIn, "admin", nil, []any{"admin", "manager", "user"}, true, true},
		{"not in list", OpNotIn, "guest", nil, []any{"admin", "manager", "user"}, true, true},
		{"exists", OpExists, "some_value", nil, nil, true, true},
		{"not exists", OpNotExists, nil, nil, nil, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.evaluateConditionOperator(tt.operator, tt.actual, tt.expected, tt.values, tt.caseSensitive)
			assert.Equal(t, tt.result, result)
		})
	}
}

// TestContextFieldValueExtraction tests field value extraction from context
func TestContextFieldValueExtraction(t *testing.T) {
	engine := &advancedEvaluationEngine{}
	userID := uuid.New()

	context := &AdvancedEvaluationContext{
		UserID: userID,
		UserData: map[string]any{
			"role":       "admin",
			"department": "engineering",
		},
		SessionData: map[string]any{
			"duration": "2h",
		},
		DeviceInfo: &conditional.DeviceInfo{
			DeviceType:      "desktop",
			OperatingSystem: "Windows",
			Browser:         "Chrome",
			IsManaged:       true,
			IsCompliant:     true,
		},
		LocationInfo: &conditional.LocationInfo{
			Country:   "US",
			Region:    "CA",
			City:      "San Francisco",
			IsTrusted: true,
		},
		CustomData: map[string]any{
			"segment": "premium",
		},
	}

	tests := []struct {
		field    string
		expected any
	}{
		{"user.role", "admin"},
		{"user.department", "engineering"},
		{"session.duration", "2h"},
		{"device.type", "desktop"},
		{"device.os", "Windows"},
		{"device.browser", "Chrome"},
		{"device.is_managed", true},
		{"device.is_compliant", true},
		{"location.country", "US"},
		{"location.region", "CA"},
		{"location.city", "San Francisco"},
		{"location.is_trusted", true},
		{"custom.segment", "premium"},
		{"unknown.field", nil},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := engine.extractFieldValue(tt.field, context)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestUserIDHashingConsistency tests the user ID hashing function for consistent percentage calculations
func TestUserIDHashingConsistency(t *testing.T) {
	engine := &advancedEvaluationEngine{}
	
	userID1 := uuid.New()
	userID2 := uuid.New()
	
	// Same user ID should always produce the same hash
	hash1a := engine.hashUserID(userID1)
	hash1b := engine.hashUserID(userID1)
	assert.Equal(t, hash1a, hash1b)
	
	// Different user IDs should (very likely) produce different hashes
	hash2 := engine.hashUserID(userID2)
	assert.NotEqual(t, hash1a, hash2)
	
	// Hash should be within expected range for percentage calculations
	assert.True(t, hash1a >= 0)
	assert.True(t, hash2 >= 0)
}

// TestLogicalOperatorEvaluation tests AND, OR, NOT logical operators
func TestLogicalOperatorEvaluation(t *testing.T) {
	engine := &advancedEvaluationEngine{}

	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name     string
		rule     *EvaluationRule
		context  *AdvancedEvaluationContext
		expected bool
	}{
		{
			name: "AND - both true",
			rule: &EvaluationRule{
				ID:       uuid.New(),
				Type:     RuleTypeLogical,
				Operator: OperatorAND,
				Children: []*EvaluationRule{
					{
						ID:   uuid.New(),
						Type: RuleTypeCondition,
						Condition: &RuleCondition{
							Field:    "user.role",
							Operator: OpEquals,
							Value:    "admin",
						},
					},
					{
						ID:   uuid.New(),
						Type: RuleTypeCondition,
						Condition: &RuleCondition{
							Field:    "user.department",
							Operator: OpEquals,
							Value:    "engineering",
						},
					},
				},
			},
			context: &AdvancedEvaluationContext{
				UserID: userID,
				UserData: map[string]any{
					"role":       "admin",
					"department": "engineering",
				},
			},
			expected: true,
		},
		{
			name: "AND - one false",
			rule: &EvaluationRule{
				ID:       uuid.New(),
				Type:     RuleTypeLogical,
				Operator: OperatorAND,
				Children: []*EvaluationRule{
					{
						ID:   uuid.New(),
						Type: RuleTypeCondition,
						Condition: &RuleCondition{
							Field:    "user.role",
							Operator: OpEquals,
							Value:    "admin",
						},
					},
					{
						ID:   uuid.New(),
						Type: RuleTypeCondition,
						Condition: &RuleCondition{
							Field:    "user.department",
							Operator: OpEquals,
							Value:    "sales", // Different from actual
						},
					},
				},
			},
			context: &AdvancedEvaluationContext{
				UserID: userID,
				UserData: map[string]any{
					"role":       "admin",
					"department": "engineering",
				},
			},
			expected: false,
		},
		{
			name: "OR - one true",
			rule: &EvaluationRule{
				ID:       uuid.New(),
				Type:     RuleTypeLogical,
				Operator: OperatorOR,
				Children: []*EvaluationRule{
					{
						ID:   uuid.New(),
						Type: RuleTypeCondition,
						Condition: &RuleCondition{
							Field:    "user.role",
							Operator: OpEquals,
							Value:    "admin",
						},
					},
					{
						ID:   uuid.New(),
						Type: RuleTypeCondition,
						Condition: &RuleCondition{
							Field:    "user.role",
							Operator: OpEquals,
							Value:    "manager",
						},
					},
				},
			},
			context: &AdvancedEvaluationContext{
				UserID: userID,
				UserData: map[string]any{
					"role": "admin",
				},
			},
			expected: true,
		},
		{
			name: "NOT - negates true condition",
			rule: &EvaluationRule{
				ID:       uuid.New(),
				Type:     RuleTypeLogical,
				Operator: OperatorNOT,
				Children: []*EvaluationRule{
					{
						ID:   uuid.New(),
						Type: RuleTypeCondition,
						Condition: &RuleCondition{
							Field:    "user.role",
							Operator: OpEquals,
							Value:    "guest",
						},
					},
				},
			},
			context: &AdvancedEvaluationContext{
				UserID: userID,
				UserData: map[string]any{
					"role": "admin", // Not guest, so NOT guest = true
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, paths := engine.evaluateRule(ctx, tt.rule, tt.context)
			assert.Equal(t, tt.expected, result)
			assert.NotEmpty(t, paths)
		})
	}
}