package authz

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"awo/internal/core/access"
	"awo/internal/core/iam/model"
	"awo/internal/shared/errors"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
)

// ConditionalAccessTestSuite focuses on conditional access adapter methods
type ConditionalAccessTestSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	mockAccess   *access.MockService
	mockLogger   *logger.MockLogger
	mockMetrics  *metrics.MockMetricsProvider
	mockTracer   *tracing.MockService
	mockSpan     *tracing.MockSpan
	adapter      Service
	ctx          context.Context
	testUserID   uuid.UUID
	testEntityID uuid.UUID
}

func TestConditionalAccessTestSuite(t *testing.T) {
	suite.Run(t, new(ConditionalAccessTestSuite))
}

func (s *ConditionalAccessTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockAccess = access.NewMockService(s.ctrl)
	s.mockLogger = logger.NewMockLogger(s.ctrl)
	s.mockMetrics = metrics.NewMockMetricsProvider(s.ctrl)
	s.mockTracer = tracing.NewMockService(s.ctrl)
	s.mockSpan = tracing.NewMockSpan(s.ctrl)

	// Setup common mock expectations
	s.mockSpan.EXPECT().End().AnyTimes()
	s.mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).
		Return(context.Background(), s.mockSpan).AnyTimes()
	s.mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()
	s.mockTracer.EXPECT().RecordError(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	// Create adapter with mock access service (ABAC service not needed for these tests)
	s.adapter = NewAdapter(nil, s.mockAccess, s.mockLogger, s.mockMetrics, s.mockTracer)
	s.ctx = context.Background()
	s.testUserID = uuid.New()
	s.testEntityID = uuid.New()
}

func (s *ConditionalAccessTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// ─── CONDITIONAL ACCESS EVALUATION TESTS ───────────────────────────────────

func (s *ConditionalAccessTestSuite) TestEvaluateConditionalAccess() {
	testCases := []struct {
		name           string
		req            *ConditionalAccessRequest
		accessResult   *access.ConditionalAccessResult
		accessError    error
		expectedResult *ConditionalAccessResult
		expectedError  string
	}{
		{
			name: "access_allowed_no_conditions",
			req: &ConditionalAccessRequest{
				UserID:       s.testUserID,
				ResourceType: "document",
				ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
				Action:       "read",
				Context: &model.AccessContext{
					IPAddress: "192.168.1.100",
					UserAgent: "Mozilla/5.0",
					Device: &model.DeviceContext{
						DeviceID:  "device-123",
						Platform:  "Windows",
						Browser:   "Chrome",
						IsTrusted: true,
						IsManaged: true,
					},
					Location: &model.GeolocationContext{
						Country:   "US",
						Region:    "CA",
						City:      "San Francisco",
						Latitude:  37.7749,
						Longitude: -122.4194,
					},
					Time:      time.Now(),
					RiskLevel: "low",
					Attributes: map[string]any{
						"department":     "engineering",
						"security_level": "standard",
					},
				},
			},
			accessResult: &access.ConditionalAccessResult{
				Allowed:    true,
				Conditions: []*access.AccessCondition{},
				Reason:     "Access granted based on trust level and location",
				Metadata: map[string]any{
					"risk_score":     0.1,
					"trust_level":    "high",
					"location_match": true,
				},
			},
			expectedResult: &ConditionalAccessResult{
				Allowed:    true,
				Conditions: []*model.AccessCondition{},
				Reason:     "Access granted based on trust level and location",
				Metadata: map[string]any{
					"risk_score":     0.1,
					"trust_level":    "high",
					"location_match": true,
				},
			},
		},
		{
			name: "access_allowed_with_conditions",
			req: &ConditionalAccessRequest{
				UserID:       s.testUserID,
				ResourceType: "sensitive_document",
				ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
				Action:       "read",
				Context: &model.AccessContext{
					IPAddress: "203.0.113.1", // External IP
					UserAgent: "Mozilla/5.0",
					Device: &model.DeviceContext{
						DeviceID:  "device-456",
						Platform:  "Mobile",
						Browser:   "Safari",
						IsTrusted: false,
						IsManaged: false,
					},
					RiskLevel: "medium",
					Attributes: map[string]any{
						"department": "finance",
						"clearance":  "confidential",
					},
				},
			},
			accessResult: &access.ConditionalAccessResult{
				Allowed: true,
				Conditions: []*access.AccessCondition{
					{
						Type:        "MFA_REQUIRED",
						Description: "Multi-factor authentication required",
						IsMandatory: true,
						Metadata: map[string]any{
							"methods": []string{"totp", "sms"},
							"timeout": 300,
						},
					},
					{
						Type:        "TIME_LIMITED",
						Description: "Access limited to business hours",
						IsMandatory: true,
						Metadata: map[string]any{
							"start_time": "09:00",
							"end_time":   "17:00",
							"timezone":   "UTC",
						},
					},
				},
				Reason: "Conditional access granted with security requirements",
				Metadata: map[string]any{
					"risk_score":       0.7,
					"trust_level":      "medium",
					"device_untrusted": true,
				},
			},
			expectedResult: &ConditionalAccessResult{
				Allowed: true,
				Conditions: []*model.AccessCondition{
					{
						Type:        "MFA_REQUIRED",
						Description: "Multi-factor authentication required",
						Required:    true,
						Parameters: map[string]any{
							"methods": []string{"totp", "sms"},
							"timeout": 300,
						},
					},
					{
						Type:        "TIME_LIMITED",
						Description: "Access limited to business hours",
						Required:    true,
						Parameters: map[string]any{
							"start_time": "09:00",
							"end_time":   "17:00",
							"timezone":   "UTC",
						},
					},
				},
				Reason: "Conditional access granted with security requirements",
				Metadata: map[string]any{
					"risk_score":       0.7,
					"trust_level":      "medium",
					"device_untrusted": true,
				},
			},
		},
		{
			name: "access_denied",
			req: &ConditionalAccessRequest{
				UserID:       s.testUserID,
				ResourceType: "top_secret_document",
				ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
				Action:       "read",
				Context: &model.AccessContext{
					IPAddress: "198.51.100.1", // Suspicious IP
					RiskLevel: "high",
					Device: &model.DeviceContext{
						DeviceID:  "unknown-device",
						Platform:  "Unknown",
						IsTrusted: false,
						IsManaged: false,
					},
					Attributes: map[string]any{
						"department": "external",
						"clearance":  "none",
					},
				},
			},
			accessResult: &access.ConditionalAccessResult{
				Allowed:    false,
				Conditions: []*access.AccessCondition{},
				Reason:     "Access denied due to high risk and insufficient clearance",
				Metadata: map[string]any{
					"risk_score":       0.9,
					"trust_level":      "none",
					"clearance_level":  "insufficient",
					"location_blocked": true,
				},
			},
			expectedResult: &ConditionalAccessResult{
				Allowed:    false,
				Conditions: []*model.AccessCondition{},
				Reason:     "Access denied due to high risk and insufficient clearance",
				Metadata: map[string]any{
					"risk_score":       0.9,
					"trust_level":      "none",
					"clearance_level":  "insufficient",
					"location_blocked": true,
				},
			},
		},
		{
			name: "access_service_error",
			req: &ConditionalAccessRequest{
				UserID:       s.testUserID,
				ResourceType: "document",
				Action:       "read",
			},
			accessError:   errors.NewBusinessError("ACCESS_ERROR", "Failed to evaluate conditional access"),
			expectedError: "failed to evaluate conditional access via access service",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.accessError != nil {
				s.mockAccess.EXPECT().EvaluateConditionalAccess(gomock.Any(), gomock.Any()).
					Return(nil, tc.accessError)
			} else {
				s.mockAccess.EXPECT().EvaluateConditionalAccess(gomock.Any(), gomock.Any()).
					Return(tc.accessResult, nil)
			}

			result, err := s.adapter.EvaluateConditionalAccess(s.ctx, tc.req)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Equal(tc.expectedResult.Allowed, result.Allowed)
				s.Equal(tc.expectedResult.Reason, result.Reason)
				s.Len(result.Conditions, len(tc.expectedResult.Conditions))
				s.Equal(tc.expectedResult.Metadata, result.Metadata)

				// Verify conditions
				for i, expectedCond := range tc.expectedResult.Conditions {
					s.Equal(expectedCond.Type, result.Conditions[i].Type)
					s.Equal(expectedCond.Description, result.Conditions[i].Description)
					s.Equal(expectedCond.Required, result.Conditions[i].Required)
				}
			}
		})
	}
}

// ─── CONDITIONAL ACCESS POLICY TESTS ───────────────────────────────────────

func (s *ConditionalAccessTestSuite) TestCreateConditionalAccessPolicy() {
	policyID := uuid.New()
	testCases := []struct {
		name           string
		req            *CreateConditionalAccessPolicyRequest
		accessResult   *access.ConditionalAccessPolicy
		accessError    error
		expectedResult *model.ConditionalAccessPolicy
		expectedError  string
	}{
		{
			name: "successful_policy_creation",
			req: &CreateConditionalAccessPolicyRequest{
				Name:        "High Security Document Access Policy",
				Description: "Requires MFA and trusted device for sensitive documents",
				Conditions: []*model.PolicyCondition{
					{
						Expression: "resource.type == 'sensitive_document'",
						Attributes: map[string]any{
							"resource_type":     "sensitive_document",
							"sensitivity_level": "high",
						},
					},
					{
						Expression: "user.department in ['finance', 'legal', 'executive']",
						Attributes: map[string]any{
							"allowed_departments": []string{"finance", "legal", "executive"},
						},
					},
				},
				Actions: []*model.PolicyAction{
					{
						Type:        "REQUIRE_MFA",
						Description: "Multi-factor authentication required",
						Parameters: map[string]any{
							"methods":  []string{"totp", "hardware_key"},
							"validity": 3600,
						},
					},
					{
						Type:        "REQUIRE_TRUSTED_DEVICE",
						Description: "Device must be managed and trusted",
						Parameters: map[string]any{
							"managed": true,
							"trusted": true,
						},
					},
				},
				Priority: 100,
				Enabled:  true,
				Metadata: map[string]any{
					"category":   "security",
					"compliance": "SOX",
					"created_by": s.testUserID.String(),
				},
			},
			accessResult: &access.ConditionalAccessPolicy{
				ID:       policyID,
				TenantID: s.testEntityID,
				Name:     "High Security Document Access Policy",
				Description: func() *string {
					desc := "Requires MFA and trusted device for sensitive documents"
					return &desc
				}(),
				Conditions: map[string]any{
					"resource_type":       "sensitive_document",
					"sensitivity_level":   "high",
					"allowed_departments": []string{"finance", "legal", "executive"},
				},
				Actions: map[string]any{
					"REQUIRE_MFA":            true,
					"REQUIRE_TRUSTED_DEVICE": true,
				},
				Priority:  100,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedResult: &model.ConditionalAccessPolicy{
				ID:          policyID,
				TenantID:    s.testEntityID,
				Name:        "High Security Document Access Policy",
				Description: "Requires MFA and trusted device for sensitive documents",
				Priority:    100,
				Enabled:     true,
				Metadata:    map[string]any{"source": "conditional_access"},
			},
		},
		{
			name: "minimal_policy_creation",
			req: &CreateConditionalAccessPolicyRequest{
				Name:     "Basic Policy",
				Priority: 50,
				Enabled:  true,
			},
			accessResult: &access.ConditionalAccessPolicy{
				ID:       policyID,
				TenantID: s.testEntityID,
				Name:     "Basic Policy",
				Priority: 50,
				IsActive: true,
			},
			expectedResult: &model.ConditionalAccessPolicy{
				ID:       policyID,
				TenantID: s.testEntityID,
				Name:     "Basic Policy",
				Priority: 50,
				Enabled:  true,
			},
		},
		{
			name: "access_service_error",
			req: &CreateConditionalAccessPolicyRequest{
				Name:     "Failed Policy",
				Priority: 10,
				Enabled:  true,
			},
			accessError:   errors.NewBusinessError("ACCESS_ERROR", "Failed to create policy"),
			expectedError: "failed to create conditional access policy via access service",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.accessError != nil {
				s.mockAccess.EXPECT().CreateConditionalAccessPolicy(gomock.Any(), gomock.Any()).
					Return(nil, tc.accessError)
			} else {
				s.mockAccess.EXPECT().CreateConditionalAccessPolicy(gomock.Any(), gomock.Any()).
					Return(tc.accessResult, nil)
			}

			result, err := s.adapter.CreateConditionalAccessPolicy(s.ctx, tc.req)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Equal(tc.expectedResult.ID, result.ID)
				s.Equal(tc.expectedResult.Name, result.Name)
				s.Equal(tc.expectedResult.Description, result.Description)
				s.Equal(tc.expectedResult.Priority, result.Priority)
				s.Equal(tc.expectedResult.Enabled, result.Enabled)
			}
		})
	}
}

func (s *ConditionalAccessTestSuite) TestUpdateConditionalAccessPolicy() {
	policyID := uuid.New()
	testCases := []struct {
		name          string
		req           *UpdateConditionalAccessPolicyRequest
		accessError   error
		expectedError string
	}{
		{
			name: "successful_policy_update",
			req: &UpdateConditionalAccessPolicyRequest{
				PolicyID:    policyID,
				Name:        func() *string { s := "Updated Security Policy"; return &s }(),
				Description: func() *string { s := "Updated policy with enhanced security measures"; return &s }(),
				Conditions: []*model.PolicyCondition{
					{
						Expression: "resource.classification >= 'confidential'",
						Attributes: map[string]any{
							"min_classification": "confidential",
						},
					},
				},
				Actions: []*model.PolicyAction{
					{
						Type:        "REQUIRE_APPROVAL",
						Description: "Manager approval required",
						Parameters: map[string]any{
							"approver_level": "manager",
							"timeout":        86400,
						},
					},
				},
				Priority: func() *int { i := 150; return &i }(),
				Enabled:  func() *bool { b := true; return &b }(),
				Metadata: map[string]any{
					"updated_by": s.testUserID.String(),
					"version":    2,
				},
			},
		},
		{
			name: "partial_policy_update",
			req: &UpdateConditionalAccessPolicyRequest{
				PolicyID: policyID,
				Name:     func() *string { s := "Partially Updated Policy"; return &s }(),
				Priority: func() *int { i := 75; return &i }(),
			},
		},
		{
			name: "disable_policy",
			req: &UpdateConditionalAccessPolicyRequest{
				PolicyID: policyID,
				Enabled:  func() *bool { b := false; return &b }(),
			},
		},
		{
			name: "access_service_error",
			req: &UpdateConditionalAccessPolicyRequest{
				PolicyID: policyID,
				Name:     func() *string { s := "Failed Update"; return &s }(),
			},
			accessError:   errors.NewBusinessError("ACCESS_ERROR", "Policy not found"),
			expectedError: "failed to update conditional access policy via access service",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.accessError != nil {
				s.mockAccess.EXPECT().UpdateConditionalAccessPolicy(gomock.Any(), gomock.Any()).
					Return(tc.accessError)
			} else {
				s.mockAccess.EXPECT().UpdateConditionalAccessPolicy(gomock.Any(), gomock.Any()).
					Return(nil)
			}

			err := s.adapter.UpdateConditionalAccessPolicy(s.ctx, tc.req)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
			} else {
				s.NoError(err)
			}
		})
	}
}

// ─── EDGE CASE TESTS ───────────────────────────────────────────────────────

func (s *ConditionalAccessTestSuite) TestEvaluateConditionalAccess_EdgeCases() {
	testCases := []struct {
		name           string
		req            *ConditionalAccessRequest
		accessResult   *access.ConditionalAccessResult
		expectedResult *ConditionalAccessResult
	}{
		{
			name: "nil_context",
			req: &ConditionalAccessRequest{
				UserID:       s.testUserID,
				ResourceType: "document",
				Action:       "read",
				Context:      nil, // nil context
			},
			accessResult: &access.ConditionalAccessResult{
				Allowed:    false,
				Conditions: []*access.AccessCondition{},
				Reason:     "Access denied due to missing context",
				Metadata:   map[string]any{"context_missing": true},
			},
			expectedResult: &ConditionalAccessResult{
				Allowed:    false,
				Conditions: []*model.AccessCondition{},
				Reason:     "Access denied due to missing context",
				Metadata:   map[string]any{"context_missing": true},
			},
		},
		{
			name: "empty_conditions",
			req: &ConditionalAccessRequest{
				UserID:       s.testUserID,
				ResourceType: "public_document",
				Action:       "read",
				Context: &model.AccessContext{
					IPAddress: "192.168.1.1",
					RiskLevel: "low",
				},
			},
			accessResult: &access.ConditionalAccessResult{
				Allowed:    true,
				Conditions: nil, // No conditions
				Reason:     "Public resource access",
				Metadata:   map[string]any{"resource_type": "public"},
			},
			expectedResult: &ConditionalAccessResult{
				Allowed:    true,
				Conditions: []*model.AccessCondition{},
				Reason:     "Public resource access",
				Metadata:   map[string]any{"resource_type": "public"},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.mockAccess.EXPECT().EvaluateConditionalAccess(gomock.Any(), gomock.Any()).
				Return(tc.accessResult, nil)

			result, err := s.adapter.EvaluateConditionalAccess(s.ctx, tc.req)

			s.NoError(err)
			s.NotNil(result)
			s.Equal(tc.expectedResult.Allowed, result.Allowed)
			s.Equal(tc.expectedResult.Reason, result.Reason)
			s.Equal(tc.expectedResult.Metadata, result.Metadata)

			// Handle nil vs empty slice comparison
			if tc.expectedResult.Conditions == nil {
				s.Nil(result.Conditions)
			} else {
				s.Len(result.Conditions, len(tc.expectedResult.Conditions))
			}
		})
	}
}
