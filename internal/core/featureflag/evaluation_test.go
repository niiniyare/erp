package featureflag

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/niiniyare/erp/internal/core/access/conditional"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

type FeatureFlagTestSuite struct {
	suite.Suite
	engine  *advancedEvaluationEngine
	ctx     context.Context
	userID  uuid.UUID
	flagID  uuid.UUID
	baseCtx *AdvancedEvaluationContext
}

func (s *FeatureFlagTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.engine = &advancedEvaluationEngine{
		conditionalAccessService: conditional.NewMockConditionalAccessService(ctrl),
		simpleService:            nil, //NOTE: Would be injected in real usage
		logger:                   logger.NewMockLogger(ctrl),
		metrics:                  metrics.NewMockMetricsProvider(ctrl),
		tracing:                  tracing.NewMockTracingService(ctrl),
	}
	s.ctx = context.Background()
	s.userID = uuid.New()
	s.flagID = uuid.New()
	s.baseCtx = &AdvancedEvaluationContext{
		UserID: s.userID,
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
}

func (s *FeatureFlagTestSuite) TestStructCreationAndInterfaceCompliance() {
	// Verify it implements the interface
	var _ AdvancedEvaluationEngine = s.engine
	s.Require().NotNil(s.engine)
}

func (s *FeatureFlagTestSuite) TestComplexRuleEvaluationLogic() {
	// Test result construction for nil rules case
	result := &AdvancedEvaluationResult{
		FlagID:      s.flagID,
		UserID:      s.userID,
		Enabled:     false,
		Reason:      "No evaluation rules provided",
		CacheHit:    false,
		EvaluatedAt: time.Now(),
	}

	s.Require().NotNil(result)
	s.Require().False(result.Enabled)
	s.Require().Equal("No evaluation rules provided", result.Reason)
	s.Require().Equal(s.flagID, result.FlagID)
	s.Require().Equal(s.userID, result.UserID)
}

func (s *FeatureFlagTestSuite) TestRuleEvaluation() {
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
				UserID: s.userID,
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
				UserID: s.userID,
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
				UserID: s.userID,
			},
			expected: true,
		},
		{
			name: "Percentage rule - 0%",
			rule: &EvaluationRule{
				ID:     uuid.New(),
				Type:   RuleTypePercentage,
				Weight: func() *float64 { w := 0.0; return &w }(), // 0%
			},
			context: &AdvancedEvaluationContext{
				UserID: s.userID,
			},
			expected: false,
		},
		{
			name: "Device compliance rule - compliant and managed",
			rule: &EvaluationRule{
				ID:   uuid.New(),
				Type: RuleTypeDevice,
			},
			context: &AdvancedEvaluationContext{
				UserID: s.userID,
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
				UserID: s.userID,
				DeviceInfo: &conditional.DeviceInfo{
					IsCompliant: false,
					IsManaged:   true,
				},
			},
			expected: false,
		},
		{
			name: "Device compliance rule - not managed",
			rule: &EvaluationRule{
				ID:   uuid.New(),
				Type: RuleTypeDevice,
			},
			context: &AdvancedEvaluationContext{
				UserID: s.userID,
				DeviceInfo: &conditional.DeviceInfo{
					IsCompliant: true,
					IsManaged:   false,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, paths := s.engine.evaluateRule(s.ctx, tt.rule, tt.context)
			s.Require().Equal(tt.expected, result)
			s.Require().NotEmpty(paths)
		})
	}
}

func (s *FeatureFlagTestSuite) TestConditionOperatorEvaluation() {
	tests := []struct {
		name          string
		operator      ConditionOperator
		actual        any
		expected      any
		values        []any
		caseSensitive bool
		result        bool
	}{
		// Equality operators
		{"equals match", OpEquals, "admin", "admin", nil, true, true},
		{"equals no match", OpEquals, "user", "admin", nil, true, false},
		{"equals case insensitive match", OpEquals, "ADMIN", "admin", nil, false, true},
		{"equals case insensitive no match", OpEquals, "USER", "admin", nil, false, false},
		{"not equals match", OpNotEquals, "user", "admin", nil, true, true},
		{"not equals no match", OpNotEquals, "admin", "admin", nil, true, false},

		// String operators
		{"contains match", OpContains, "user@example.com", "@example", nil, true, true},
		{"contains no match", OpContains, "user@example.com", "@test", nil, true, false},
		{"not contains match", OpNotContains, "user@example.com", "@test", nil, true, true},
		{"not contains no match", OpNotContains, "user@example.com", "@example", nil, true, false},
		{"starts with match", OpStartsWith, "admin_user", "admin", nil, true, true},
		{"starts with no match", OpStartsWith, "user_admin", "admin", nil, true, false},
		{"ends with match", OpEndsWith, "power_user", "user", nil, true, true},
		{"ends with no match", OpEndsWith, "user_power", "user", nil, true, false},

		// List operators
		{"in list match", OpIn, "admin", nil, []any{"admin", "manager", "user"}, true, true},
		{"in list no match", OpIn, "guest", nil, []any{"admin", "manager", "user"}, true, false},
		{"not in list match", OpNotIn, "guest", nil, []any{"admin", "manager", "user"}, true, true},
		{"not in list no match", OpNotIn, "admin", nil, []any{"admin", "manager", "user"}, true, false},

		// Existence operators
		{"exists with value", OpExists, "some_value", nil, nil, true, true},
		{"exists with empty string", OpExists, "", nil, nil, true, true},
		{"exists with nil", OpExists, nil, nil, nil, true, false},
		{"not exists with nil", OpNotExists, nil, nil, nil, true, true},
		{"not exists with value", OpNotExists, "some_value", nil, nil, true, false},

		// Numeric operators (if supported)
		{"numeric equals", OpEquals, 42, 42, nil, true, true},
		{"numeric not equals", OpNotEquals, 42, 24, nil, true, true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := s.engine.evaluateConditionOperator(tt.operator, tt.actual, tt.expected, tt.values, tt.caseSensitive)
			s.Require().Equal(tt.result, result)
		})
	}
}

func (s *FeatureFlagTestSuite) TestContextFieldValueExtraction() {
	tests := []struct {
		field    string
		expected any
	}{
		// User data
		{"user.role", "admin"},
		{"user.department", "engineering"},

		// Session data
		{"session.duration", "2h"},

		// Device info
		{"device.type", "desktop"},
		{"device.os", "Windows"},
		{"device.browser", "Chrome"},
		{"device.is_managed", true},
		{"device.is_compliant", true},

		// Location info
		{"location.country", "US"},
		{"location.region", "CA"},
		{"location.city", "San Francisco"},
		{"location.is_trusted", true},

		// Custom data
		{"custom.segment", "premium"},

		// Non-existent fields
		{"unknown.field", nil},
		{"user.nonexistent", nil},
		{"device.invalid", nil},
	}

	for _, tt := range tests {
		s.Run(tt.field, func() {
			result := s.engine.extractFieldValue(tt.field, s.baseCtx)
			s.Require().Equal(tt.expected, result)
		})
	}
}

func (s *FeatureFlagTestSuite) TestUserIDHashingConsistency() {
	userID1 := uuid.New()
	userID2 := uuid.New()

	// Same user ID should always produce the same hash
	hash1a := s.engine.hashUserID(userID1)
	hash1b := s.engine.hashUserID(userID1)
	s.Require().Equal(hash1a, hash1b)

	// Different user IDs should (very likely) produce different hashes
	hash2 := s.engine.hashUserID(userID2)
	s.Require().NotEqual(hash1a, hash2)

	// Hash should be within expected range for percentage calculations
	s.Require().True(int(hash1a) >= 0)
	s.Require().True(int(hash2) >= 0)

	// Test multiple iterations for consistency
	for i := 0; i < 100; i++ {
		testUserID := uuid.New()
		hash1 := s.engine.hashUserID(testUserID)
		hash2 := s.engine.hashUserID(testUserID)
		s.Require().Equal(hash1, hash2, "Hash should be consistent for user %s", testUserID)
	}
}

func (s *FeatureFlagTestSuite) TestLogicalOperatorEvaluation() {
	tests := []struct {
		name     string
		rule     *EvaluationRule
		context  *AdvancedEvaluationContext
		expected bool
	}{
		{
			name: "AND - both conditions true",
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
				UserID: s.userID,
				UserData: map[string]any{
					"role":       "admin",
					"department": "engineering",
				},
			},
			expected: true,
		},
		{
			name: "AND - first condition false",
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
							Value:    "user", // Different from actual
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
				UserID: s.userID,
				UserData: map[string]any{
					"role":       "admin",
					"department": "engineering",
				},
			},
			expected: false,
		},
		{
			name: "AND - second condition false",
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
				UserID: s.userID,
				UserData: map[string]any{
					"role":       "admin",
					"department": "engineering",
				},
			},
			expected: false,
		},
		{
			name: "OR - first condition true",
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
				UserID: s.userID,
				UserData: map[string]any{
					"role": "admin",
				},
			},
			expected: true,
		},
		{
			name: "OR - second condition true",
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
							Value:    "manager",
						},
					},
					{
						ID:   uuid.New(),
						Type: RuleTypeCondition,
						Condition: &RuleCondition{
							Field:    "user.role",
							Operator: OpEquals,
							Value:    "admin",
						},
					},
				},
			},
			context: &AdvancedEvaluationContext{
				UserID: s.userID,
				UserData: map[string]any{
					"role": "admin",
				},
			},
			expected: true,
		},
		{
			name: "OR - both conditions false",
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
							Value:    "manager",
						},
					},
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
				UserID: s.userID,
				UserData: map[string]any{
					"role": "admin",
				},
			},
			expected: false,
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
				UserID: s.userID,
				UserData: map[string]any{
					"role": "admin", // Not guest, so NOT guest = true
				},
			},
			expected: true,
		},
		{
			name: "NOT - negates false condition",
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
							Value:    "admin",
						},
					},
				},
			},
			context: &AdvancedEvaluationContext{
				UserID: s.userID,
				UserData: map[string]any{
					"role": "admin", // Is admin, so NOT admin = false
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, paths := s.engine.evaluateRule(s.ctx, tt.rule, tt.context)
			s.Require().Equal(tt.expected, result)
			s.Require().NotEmpty(paths)
		})
	}
}

func (s *FeatureFlagTestSuite) TestComplexNestedLogicalRules() {
	tests := []struct {
		name     string
		rule     *EvaluationRule
		context  *AdvancedEvaluationContext
		expected bool
	}{
		{
			name: "Complex nested - (admin AND engineering) OR (manager AND sales)",
			rule: &EvaluationRule{
				ID:       uuid.New(),
				Type:     RuleTypeLogical,
				Operator: OperatorOR,
				Children: []*EvaluationRule{
					{
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
					{
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
									Value:    "manager",
								},
							},
							{
								ID:   uuid.New(),
								Type: RuleTypeCondition,
								Condition: &RuleCondition{
									Field:    "user.department",
									Operator: OpEquals,
									Value:    "sales",
								},
							},
						},
					},
				},
			},
			context: &AdvancedEvaluationContext{
				UserID: s.userID,
				UserData: map[string]any{
					"role":       "admin",
					"department": "engineering",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, paths := s.engine.evaluateRule(s.ctx, tt.rule, tt.context)
			s.Require().Equal(tt.expected, result)
			s.Require().NotEmpty(paths)
		})
	}
}

// TestFeatureFlagTestSuite runs the test suite
func TestFeatureFlagTestSuite(t *testing.T) {
	suite.Run(t, new(FeatureFlagTestSuite))
}
