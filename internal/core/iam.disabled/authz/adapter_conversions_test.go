package authz

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"awo/internal/core/iam/model"
	"awo/internal/shared/types"
)

// ConversionsTestSuite tests all conversion functions for 100% coverage
type ConversionsTestSuite struct {
	suite.Suite
}

func TestConversionsTestSuite(t *testing.T) {
	suite.Run(t, new(ConversionsTestSuite))
}

// ─── POLICY DECISION TYPE CONVERSIONS ─────────────────────────────────────

func (s *ConversionsTestSuite) TestConvertPolicyDecisionTypeToShared() {
	tests := []struct {
		name     string
		input    model.PolicyDecisionType
		expected types.PolicyDecisionType
	}{
		{
			name:     "Convert Allow Decision",
			input:    model.PolicyDecisionAllow,
			expected: types.PolicyDecisionAllow,
		},
		{
			name:     "Convert Deny Decision",
			input:    model.PolicyDecisionDeny,
			expected: types.PolicyDecisionDeny,
		},
		{
			name:     "Convert NotApplicable Decision",
			input:    model.PolicyDecisionNotApplicable,
			expected: types.PolicyDecisionNotApplicable,
		},
		{
			name:     "Convert Unknown Decision Defaults to Deny",
			input:    model.PolicyDecisionType("unknown"),
			expected: types.PolicyDecisionDeny,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := convertPolicyDecisionTypeToShared(tt.input)
			require.Equal(s.T(), tt.expected, result)
		})
	}
}

func (s *ConversionsTestSuite) TestConvertPolicyDecisionType() {
	tests := []struct {
		name     string
		input    types.PolicyDecisionType
		expected model.PolicyDecisionType
	}{
		{
			name:     "Convert Allow Decision from Types",
			input:    types.PolicyDecisionAllow,
			expected: model.PolicyDecisionAllow,
		},
		{
			name:     "Convert Deny Decision from Types",
			input:    types.PolicyDecisionDeny,
			expected: model.PolicyDecisionDeny,
		},
		{
			name:     "Convert NotApplicable Decision from Types",
			input:    types.PolicyDecisionNotApplicable,
			expected: model.PolicyDecisionNotApplicable,
		},
		{
			name:     "Convert Unknown Decision Defaults to Deny",
			input:    types.PolicyDecisionType("unknown"),
			expected: model.PolicyDecisionDeny,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := convertPolicyDecisionType(tt.input)
			require.Equal(s.T(), tt.expected, result)
		})
	}
}

// ─── POLICY EFFECT TYPE CONVERSIONS ────────────────────────────────────────

func (s *ConversionsTestSuite) TestConvertPolicyEffectType() {
	tests := []struct {
		name     string
		input    types.PolicyEffect
		expected model.PolicyEffect
	}{
		{
			name:     "Convert Allow Effect",
			input:    types.PolicyEffectAllow,
			expected: model.PolicyEffectAllow,
		},
		{
			name:     "Convert Deny Effect",
			input:    types.PolicyEffectDeny,
			expected: model.PolicyEffectDeny,
		},
		{
			name:     "Convert Unknown Effect Defaults to Deny",
			input:    types.PolicyEffect("unknown"),
			expected: model.PolicyEffectDeny,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := convertPolicyEffectType(tt.input)
			require.Equal(s.T(), tt.expected, result)
		})
	}
}

func (s *ConversionsTestSuite) TestConvertPolicyEffectTypeToShared() {
	tests := []struct {
		name     string
		input    model.PolicyEffect
		expected types.PolicyEffect
	}{
		{
			name:     "Convert Allow Effect to Shared",
			input:    model.PolicyEffectAllow,
			expected: types.PolicyEffectAllow,
		},
		{
			name:     "Convert Deny Effect to Shared",
			input:    model.PolicyEffectDeny,
			expected: types.PolicyEffectDeny,
		},
		{
			name:     "Convert Unknown Effect Defaults to Deny",
			input:    model.PolicyEffect("unknown"),
			expected: types.PolicyEffectDeny,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := convertPolicyEffectTypeToShared(tt.input)
			require.Equal(s.T(), tt.expected, result)
		})
	}
}

// ─── SECURITY AND EDGE CASE TESTS ──────────────────────────────────────────

func (s *ConversionsTestSuite) TestConversionSecurityDefaults() {
	// Test that all unknown/invalid values default to secure "Deny" state
	s.Run("Unknown Policy Decision Defaults to Deny", func() {
		// Test various invalid decision types
		invalidDecisions := []types.PolicyDecisionType{
			types.PolicyDecisionType(""),
			types.PolicyDecisionType("invalid"),
			types.PolicyDecisionType("allow-but-typo"),
		}

		for _, invalid := range invalidDecisions {
			result := convertPolicyDecisionType(invalid)
			require.Equal(s.T(), model.PolicyDecisionDeny, result, "Invalid decision should default to Deny for security")
		}
	})

	s.Run("Unknown Policy Effect Defaults to Deny", func() {
		// Test various invalid effect types
		invalidEffects := []types.PolicyEffect{
			types.PolicyEffect(""),
			types.PolicyEffect("invalid"),
			types.PolicyEffect("permit"),
		}

		for _, invalid := range invalidEffects {
			result := convertPolicyEffectType(invalid)
			require.Equal(s.T(), model.PolicyEffectDeny, result, "Invalid effect should default to Deny for security")
		}
	})
}

func (s *ConversionsTestSuite) TestBidirectionalConversionConsistency() {
	// Test that conversions are bidirectional and consistent
	s.Run("Policy Decision Bidirectional Conversion", func() {
		originalDecisions := []model.PolicyDecisionType{
			model.PolicyDecisionAllow,
			model.PolicyDecisionDeny,
			model.PolicyDecisionNotApplicable,
		}

		for _, original := range originalDecisions {
			// Convert to shared and back
			shared := convertPolicyDecisionTypeToShared(original)
			backToModel := convertPolicyDecisionType(shared)
			require.Equal(s.T(), original, backToModel, "Bidirectional conversion should be consistent")
		}
	})

	s.Run("Policy Effect Bidirectional Conversion", func() {
		originalEffects := []model.PolicyEffect{
			model.PolicyEffectAllow,
			model.PolicyEffectDeny,
		}

		for _, original := range originalEffects {
			// Convert to shared and back
			shared := convertPolicyEffectTypeToShared(original)
			backToModel := convertPolicyEffectType(shared)
			require.Equal(s.T(), original, backToModel, "Bidirectional conversion should be consistent")
		}
	})
}

func (s *ConversionsTestSuite) TestConversionPerformance() {
	// Benchmark-style test to ensure conversions are efficient
	s.Run("Policy Decision Conversion Performance", func() {
		// Perform many conversions to ensure no performance issues
		for i := 0; i < 10000; i++ {
			_ = convertPolicyDecisionType(types.PolicyDecisionAllow)
			_ = convertPolicyDecisionTypeToShared(model.PolicyDecisionAllow)
		}
		// If this completes without timeout, performance is acceptable
	})

	s.Run("Policy Effect Conversion Performance", func() {
		// Perform many conversions to ensure no performance issues
		for i := 0; i < 10000; i++ {
			_ = convertPolicyEffectType(types.PolicyEffectAllow)
			_ = convertPolicyEffectTypeToShared(model.PolicyEffectAllow)
		}
		// If this completes without timeout, performance is acceptable
	})
}
