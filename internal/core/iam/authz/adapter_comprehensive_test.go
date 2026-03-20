package authz

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"awo/internal/core/abac"
	"awo/internal/core/abac/activities"
	"awo/internal/core/abac/models"
	"awo/internal/core/access"
	"awo/internal/core/iam/model"
	"awo/internal/shared/errors"
	"awo/internal/shared/logger"
	"awo/internal/shared/metrics"
	"awo/internal/shared/tracing"
	"awo/internal/shared/types"
)

// AuthzAdapterTestSuite provides test coverage for the authorization adapter
type AuthzAdapterTestSuite struct {
	suite.Suite
	ctrl          *gomock.Controller
	mockABAC      *abac.MockService
	mockAccess    *access.MockService
	mockLogger    *logger.MockLogger
	mockMetrics   *metrics.MockMetricsProvider
	mockTracer    *tracing.MockService
	mockSpan      *tracing.MockSpan
	adapter       Service
	ctx           context.Context
	testUserID    uuid.UUID
	testEntityID  uuid.UUID
	testRequestID string
	testTimeStamp time.Time
}

func TestAuthzAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(AuthzAdapterTestSuite))
}

func (s *AuthzAdapterTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockABAC = abac.NewMockService(s.ctrl)
	s.mockAccess = access.NewMockService(s.ctrl)
	s.mockLogger = logger.NewMockLogger(s.ctrl)
	s.mockMetrics = metrics.NewMockMetricsProvider(s.ctrl)
	s.mockTracer = tracing.NewMockService(s.ctrl)
	s.mockSpan = tracing.NewMockSpan(s.ctrl)

	// Default mock expectations
	s.mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).Return(context.Background(), s.mockSpan).AnyTimes()
	s.mockSpan.EXPECT().End().AnyTimes()
	s.mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	s.mockMetrics.EXPECT().IncrementCounter(gomock.Any(), gomock.Any()).AnyTimes()
	s.mockMetrics.EXPECT().ObserveHistogram(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().InfoContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	s.mockLogger.EXPECT().ErrorContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	s.mockTracer.EXPECT().RecordError(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	s.adapter = NewAdapter(s.mockABAC, s.mockAccess, s.mockLogger, s.mockMetrics, s.mockTracer)
}

func (s *AuthzAdapterTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// ─── PERMISSION EVALUATION TESTS ───────────────────────────────────────────

func (s *AuthzAdapterTestSuite) TestEvaluatePermission() {
	testCases := []struct {
		name           string
		req            *PermissionEvaluationRequest
		abacResult     *abac.PermissionEvaluationResult
		abacError      error
		expectedResult *PermissionEvaluationResult
		expectedError  string
	}{
		{
			name: "successful_evaluation_allow",
			req: &PermissionEvaluationRequest{
				UserID:       s.testUserID,
				ResourceType: "document",
				ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
				Action:       "read",
				EntityID:     &s.testEntityID,
				Context:      map[string]any{"department": "engineering"},
				RequestID:    s.testRequestID,
			},
			abacResult: &abac.PermissionEvaluationResult{
				Decision: types.PolicyDecisionAllow,
				PolicyDecisions: []*models.PolicyDecision{
					{
						PolicyID:      uuid.New(),
						Decision:      types.PolicyDecisionAllow,
						Reason:        "Policy allows access",
						MatchedRule:   "default_rule",
						EvaluationMS:  5,
						TargetMatched: true,
					},
				},
				EvaluationTimeMS: 25,
				CacheHit:         false,
				RequestID:        s.testRequestID,
				Timestamp:        s.testTimeStamp,
			},
			expectedResult: &PermissionEvaluationResult{
				Decision:         model.PolicyDecisionAllow,
				EvaluationTimeMS: 25,
				CacheHit:         false,
				RequestID:        s.testRequestID,
				Timestamp:        s.testTimeStamp,
			},
		},
		{
			name: "successful_evaluation_deny",
			req: &PermissionEvaluationRequest{
				UserID:       s.testUserID,
				ResourceType: "document",
				Action:       "delete",
				RequestID:    s.testRequestID,
			},
			abacResult: &abac.PermissionEvaluationResult{
				Decision:         types.PolicyDecisionDeny,
				EvaluationTimeMS: 15,
				CacheHit:         true,
				RequestID:        s.testRequestID,
				Timestamp:        s.testTimeStamp,
			},
			expectedResult: &PermissionEvaluationResult{
				Decision:         model.PolicyDecisionDeny,
				EvaluationTimeMS: 15,
				CacheHit:         true,
				RequestID:        s.testRequestID,
				Timestamp:        s.testTimeStamp,
			},
		},
		{
			name: "abac_service_error",
			req: &PermissionEvaluationRequest{
				UserID:       s.testUserID,
				ResourceType: "document",
				Action:       "read",
				RequestID:    s.testRequestID,
			},
			abacError:     errors.NewBusinessError("ABAC_ERROR", "ABAC service failed"),
			expectedError: "failed to evaluate permission via ABAC",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.abacError != nil {
				s.mockABAC.EXPECT().EvaluatePermission(gomock.Any(), gomock.Any()).
					Return(nil, tc.abacError)
			} else {
				s.mockABAC.EXPECT().EvaluatePermission(gomock.Any(), gomock.Any()).
					Return(tc.abacResult, nil)
			}

			result, err := s.adapter.EvaluatePermission(s.ctx, tc.req)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Equal(tc.expectedResult.Decision, result.Decision)
				s.Equal(tc.expectedResult.EvaluationTimeMS, result.EvaluationTimeMS)
				s.Equal(tc.expectedResult.CacheHit, result.CacheHit)
				s.Equal(tc.expectedResult.RequestID, result.RequestID)
			}
		})
	}
}

func (s *AuthzAdapterTestSuite) TestBulkEvaluatePermissions() {
	testCases := []struct {
		name           string
		req            *BulkPermissionEvaluationRequest
		abacResult     *abac.BulkPermissionEvaluationResult
		abacError      error
		expectedResult *BulkPermissionEvaluationResult
		expectedError  string
	}{
		{
			name: "successful_bulk_evaluation",
			req: &BulkPermissionEvaluationRequest{
				Requests: []*PermissionEvaluationRequest{
					{UserID: s.testUserID, ResourceType: "document", Action: "read"},
					{UserID: s.testUserID, ResourceType: "file", Action: "write"},
				},
				RequestID: s.testRequestID,
			},
			abacResult: &abac.BulkPermissionEvaluationResult{
				Results: []*abac.PermissionEvaluationResult{
					{
						Decision:         types.PolicyDecisionAllow,
						EvaluationTimeMS: 10,
						CacheHit:         false,
						RequestID:        s.testRequestID + "-1",
					},
					{
						Decision:         types.PolicyDecisionDeny,
						EvaluationTimeMS: 8,
						CacheHit:         true,
						RequestID:        s.testRequestID + "-2",
					},
				},
				TotalRequests:   2,
				SuccessfulCount: 2,
				FailedCount:     0,
				TotalTimeMS:     18,
				AverageTimeMS:   9,
				RequestID:       s.testRequestID,
				Timestamp:       s.testTimeStamp,
			},
			expectedResult: &BulkPermissionEvaluationResult{
				TotalRequests:   2,
				SuccessfulCount: 2,
				FailedCount:     0,
				TotalTimeMS:     18,
				AverageTimeMS:   9,
				RequestID:       s.testRequestID,
				Timestamp:       s.testTimeStamp,
			},
		},
		{
			name: "bulk_evaluation_with_failures",
			req: &BulkPermissionEvaluationRequest{
				Requests:  []*PermissionEvaluationRequest{{UserID: s.testUserID, ResourceType: "document", Action: "read"}},
				RequestID: s.testRequestID,
			},
			abacResult: &abac.BulkPermissionEvaluationResult{
				Results: []*abac.PermissionEvaluationResult{
					{Decision: types.PolicyDecisionAllow, EvaluationTimeMS: 10},
				},
				TotalRequests:   1,
				SuccessfulCount: 1,
				FailedCount:     0,
				TotalTimeMS:     10,
				AverageTimeMS:   10,
				RequestID:       s.testRequestID,
			},
			expectedResult: &BulkPermissionEvaluationResult{
				TotalRequests:   1,
				SuccessfulCount: 1,
				FailedCount:     0,
				TotalTimeMS:     10,
				AverageTimeMS:   10,
				RequestID:       s.testRequestID,
			},
		},
		{
			name: "abac_bulk_error",
			req: &BulkPermissionEvaluationRequest{
				Requests:  []*PermissionEvaluationRequest{{UserID: s.testUserID, ResourceType: "document", Action: "read"}},
				RequestID: s.testRequestID,
			},
			abacError:     errors.NewBusinessError("ABAC_BULK_ERROR", "Bulk evaluation failed"),
			expectedError: "failed to bulk evaluate permissions via ABAC",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.abacError != nil {
				s.mockABAC.EXPECT().BulkEvaluatePermissions(gomock.Any(), gomock.Any()).
					Return(nil, tc.abacError)
			} else {
				s.mockABAC.EXPECT().BulkEvaluatePermissions(gomock.Any(), gomock.Any()).
					Return(tc.abacResult, nil)
			}

			result, err := s.adapter.BulkEvaluatePermissions(s.ctx, tc.req)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Equal(tc.expectedResult.TotalRequests, result.TotalRequests)
				s.Equal(tc.expectedResult.SuccessfulCount, result.SuccessfulCount)
				s.Equal(tc.expectedResult.FailedCount, result.FailedCount)
				s.Len(result.Results, len(tc.abacResult.Results))
			}
		})
	}
}

// ─── USER PERMISSIONS TESTS ────────────────────────────────────────────────

func (s *AuthzAdapterTestSuite) TestGetUserEffectivePermissions() {
	testCases := []struct {
		name           string
		userID         uuid.UUID
		entityID       *uuid.UUID
		abacResult     *abac.UserEffectivePermissions
		abacError      error
		expectedResult *UserEffectivePermissions
		expectedError  string
	}{
		{
			name:     "successful_get_permissions",
			userID:   s.testUserID,
			entityID: &s.testEntityID,
			abacResult: &abac.UserEffectivePermissions{
				UserID:   s.testUserID,
				EntityID: &s.testEntityID,
				Permissions: map[string][]string{
					"document": {"read", "write"},
					"file":     {"read"},
				},
				Roles:     []string{"user", "editor"},
				Timestamp: s.testTimeStamp,
			},
			expectedResult: &UserEffectivePermissions{
				UserID:   s.testUserID,
				EntityID: &s.testEntityID,
				Permissions: map[string][]string{
					"document": {"read", "write"},
					"file":     {"read"},
				},
				Roles:     []string{"user", "editor"},
				Timestamp: s.testTimeStamp,
			},
		},
		{
			name:     "permissions_without_entity",
			userID:   s.testUserID,
			entityID: nil,
			abacResult: &abac.UserEffectivePermissions{
				UserID:      s.testUserID,
				EntityID:    nil,
				Permissions: map[string][]string{"system": {"login"}},
				Roles:       []string{"user"},
				Timestamp:   s.testTimeStamp,
			},
			expectedResult: &UserEffectivePermissions{
				UserID:      s.testUserID,
				EntityID:    nil,
				Permissions: map[string][]string{"system": {"login"}},
				Roles:       []string{"user"},
				Timestamp:   s.testTimeStamp,
			},
		},
		{
			name:          "abac_error",
			userID:        s.testUserID,
			entityID:      &s.testEntityID,
			abacError:     errors.NewBusinessError("ABAC_ERROR", "Failed to get permissions"),
			expectedError: "failed to get user effective permissions via ABAC",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.abacError != nil {
				s.mockABAC.EXPECT().GetUserEffectivePermissions(gomock.Any(), tc.userID, tc.entityID).
					Return(nil, tc.abacError)
			} else {
				s.mockABAC.EXPECT().GetUserEffectivePermissions(gomock.Any(), tc.userID, tc.entityID).
					Return(tc.abacResult, nil)
			}

			result, err := s.adapter.GetUserEffectivePermissions(s.ctx, tc.userID, tc.entityID)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Equal(tc.expectedResult.UserID, result.UserID)
				s.Equal(tc.expectedResult.EntityID, result.EntityID)
				s.Equal(tc.expectedResult.Permissions, result.Permissions)
				s.Equal(tc.expectedResult.Roles, result.Roles)
			}
		})
	}
}

func (s *AuthzAdapterTestSuite) TestCalculateRoleHierarchy() {
	testCases := []struct {
		name           string
		userID         uuid.UUID
		entityID       *uuid.UUID
		abacResult     *abac.RoleHierarchy
		abacError      error
		expectedResult *RoleHierarchy
		expectedError  string
	}{
		{
			name:     "successful_role_hierarchy",
			userID:   s.testUserID,
			entityID: &s.testEntityID,
			abacResult: &abac.RoleHierarchy{
				UserID:   s.testUserID,
				EntityID: &s.testEntityID,
				Roles:    []string{"user", "editor", "admin"},
				Hierarchy: map[string][]string{
					"admin":  {"editor", "user"},
					"editor": {"user"},
					"user":   {},
				},
				Timestamp: s.testTimeStamp,
			},
			expectedResult: &RoleHierarchy{
				UserID:   s.testUserID,
				EntityID: &s.testEntityID,
				Roles:    []string{"user", "editor", "admin"},
				Hierarchy: map[string][]string{
					"admin":  {"editor", "user"},
					"editor": {"user"},
					"user":   {},
				},
				Timestamp: s.testTimeStamp,
			},
		},
		{
			name:          "abac_error",
			userID:        s.testUserID,
			entityID:      &s.testEntityID,
			abacError:     errors.NewBusinessError("ABAC_ERROR", "Failed to calculate hierarchy"),
			expectedError: "failed to calculate role hierarchy via ABAC",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.abacError != nil {
				s.mockABAC.EXPECT().CalculateRoleHierarchy(gomock.Any(), tc.userID, tc.entityID).
					Return(nil, tc.abacError)
			} else {
				s.mockABAC.EXPECT().CalculateRoleHierarchy(gomock.Any(), tc.userID, tc.entityID).
					Return(tc.abacResult, nil)
			}

			result, err := s.adapter.CalculateRoleHierarchy(s.ctx, tc.userID, tc.entityID)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Equal(tc.expectedResult.UserID, result.UserID)
				s.Equal(tc.expectedResult.EntityID, result.EntityID)
				s.Equal(tc.expectedResult.Roles, result.Roles)
				s.Equal(tc.expectedResult.Hierarchy, result.Hierarchy)
			}
		})
	}
}

// ─── ACCESS REQUESTS TESTS ─────────────────────────────────────────────────

func (s *AuthzAdapterTestSuite) TestCreateAccessRequest() {
	testCases := []struct {
		name           string
		req            *CreateAccessRequestRequest
		accessResult   *access.AccessRequest
		accessError    error
		expectedResult *model.AccessRequest
		expectedError  string
	}{
		{
			name: "successful_create",
			req: &CreateAccessRequestRequest{
				UserID:        s.testUserID,
				ResourceType:  "document",
				ResourceID:    func() *uuid.UUID { id := uuid.New(); return &id }(),
				Action:        "read",
				Justification: "Need access for review",
				Duration:      func() *time.Duration { d := time.Hour * 24; return &d }(),
				Priority:      "high",
				Metadata:      map[string]any{"department": "engineering"},
			},
			accessResult: &access.AccessRequest{
				ID:            uuid.New(),
				TenantID:      s.testEntityID,
				RequesterID:   s.testUserID,
				TargetUserID:  &s.testUserID,
				EntityID:      s.testEntityID,
				RequestType:   access.RequestType("RESOURCE_ACCESS"),
				ResourceID:    func() *uuid.UUID { id := uuid.New(); return &id }(),
				Justification: "Need access for review",
				ExpiresAt:     &s.testTimeStamp,
				CreatedAt:     s.testTimeStamp,
				UpdatedAt:     s.testTimeStamp,
			},
		},
		{
			name: "access_service_error",
			req: &CreateAccessRequestRequest{
				UserID:       s.testUserID,
				ResourceType: "document",
				Action:       "read",
			},
			accessError:   errors.NewBusinessError("ACCESS_ERROR", "Failed to create request"),
			expectedError: "failed to create access request via access service",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.accessError != nil {
				s.mockAccess.EXPECT().CreateAccessRequest(gomock.Any(), gomock.Any()).
					Return(nil, tc.accessError)
			} else {
				s.mockAccess.EXPECT().CreateAccessRequest(gomock.Any(), gomock.Any()).
					Return(tc.accessResult, nil)
			}

			result, err := s.adapter.CreateAccessRequest(s.ctx, tc.req)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Equal(tc.accessResult.ID, result.ID)
				s.Equal(tc.accessResult.RequesterID, result.RequesterID)
			}
		})
	}
}

func (s *AuthzAdapterTestSuite) TestGetAccessRequest() {
	requestID := uuid.New()
	testCases := []struct {
		name           string
		requestID      uuid.UUID
		accessResult   *access.AccessRequest
		accessError    error
		expectedResult *model.AccessRequest
		expectedError  string
	}{
		{
			name:      "successful_get",
			requestID: requestID,
			accessResult: &access.AccessRequest{
				ID:          requestID,
				TenantID:    s.testEntityID,
				RequesterID: s.testUserID,
				EntityID:    s.testEntityID,
				RequestType: access.RequestType("RESOURCE_ACCESS"),
				CreatedAt:   s.testTimeStamp,
			},
		},
		{
			name:          "access_service_error",
			requestID:     requestID,
			accessError:   errors.NewBusinessError("ACCESS_ERROR", "Request not found"),
			expectedError: "failed to get access request via access service",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.accessError != nil {
				s.mockAccess.EXPECT().GetAccessRequest(gomock.Any(), tc.requestID).
					Return(nil, tc.accessError)
			} else {
				s.mockAccess.EXPECT().GetAccessRequest(gomock.Any(), tc.requestID).
					Return(tc.accessResult, nil)
			}

			result, err := s.adapter.GetAccessRequest(s.ctx, tc.requestID)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Equal(tc.accessResult.ID, result.ID)
			}
		})
	}
}

func (s *AuthzAdapterTestSuite) TestProcessAccessRequest() {
	requestID := uuid.New()
	approverID := uuid.New()

	testCases := []struct {
		name          string
		req           *ProcessAccessRequestRequest
		accessError   error
		expectedError string
	}{
		{
			name: "successful_approval",
			req: &ProcessAccessRequestRequest{
				RequestID:  requestID,
				Action:     "approve",
				ApproverID: approverID,
				Comments:   "Approved for business needs",
				Conditions: []string{"time_limited"},
			},
		},
		{
			name: "successful_rejection",
			req: &ProcessAccessRequestRequest{
				RequestID:  requestID,
				Action:     "reject",
				ApproverID: approverID,
				Comments:   "Insufficient justification",
			},
		},
		{
			name: "access_service_error",
			req: &ProcessAccessRequestRequest{
				RequestID:  requestID,
				Action:     "approve",
				ApproverID: approverID,
			},
			accessError:   errors.NewBusinessError("ACCESS_ERROR", "Failed to process"),
			expectedError: "failed to process access request via access service",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.accessError != nil {
				s.mockAccess.EXPECT().ProcessAccessRequest(gomock.Any(), gomock.Any()).
					Return(tc.accessError)
			} else {
				s.mockAccess.EXPECT().ProcessAccessRequest(gomock.Any(), gomock.Any()).
					Return(nil)
			}

			err := s.adapter.ProcessAccessRequest(s.ctx, tc.req)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *AuthzAdapterTestSuite) TestListAccessRequests() {
	testCases := []struct {
		name           string
		req            *ListAccessRequestsRequest
		accessResult   *access.ListAccessRequestsResult
		accessError    error
		expectedResult *ListAccessRequestsResult
		expectedError  string
	}{
		{
			name: "successful_list",
			req: &ListAccessRequestsRequest{
				UserID:   &s.testUserID,
				Status:   func() *string { s := "pending"; return &s }(),
				EntityID: &s.testEntityID,
				Limit:    10,
				Offset:   0,
			},
			accessResult: &access.ListAccessRequestsResult{
				Requests: []*access.AccessRequest{
					{
						ID:          uuid.New(),
						RequesterID: s.testUserID,
						EntityID:    s.testEntityID,
						RequestType: access.RequestType("RESOURCE_ACCESS"),
					},
				},
				Total:   1,
				Limit:   10,
				Offset:  0,
				HasMore: false,
			},
			expectedResult: &ListAccessRequestsResult{
				Total:   1,
				Limit:   10,
				Offset:  0,
				HasMore: false,
			},
		},
		{
			name: "empty_list",
			req: &ListAccessRequestsRequest{
				UserID: &s.testUserID,
				Limit:  10,
				Offset: 0,
			},
			accessResult: &access.ListAccessRequestsResult{
				Requests: []*access.AccessRequest{},
				Total:    0,
				Limit:    10,
				Offset:   0,
				HasMore:  false,
			},
			expectedResult: &ListAccessRequestsResult{
				Total:   0,
				Limit:   10,
				Offset:  0,
				HasMore: false,
			},
		},
		{
			name: "access_service_error",
			req: &ListAccessRequestsRequest{
				UserID: &s.testUserID,
			},
			accessError:   errors.NewBusinessError("ACCESS_ERROR", "Failed to list requests"),
			expectedError: "failed to list access requests via access service",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.accessError != nil {
				s.mockAccess.EXPECT().ListAccessRequests(gomock.Any(), gomock.Any()).
					Return(nil, tc.accessError)
			} else {
				s.mockAccess.EXPECT().ListAccessRequests(gomock.Any(), gomock.Any()).
					Return(tc.accessResult, nil)
			}

			result, err := s.adapter.ListAccessRequests(s.ctx, tc.req)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Equal(tc.expectedResult.Total, result.Total)
				s.Equal(tc.expectedResult.Limit, result.Limit)
				s.Equal(tc.expectedResult.Offset, result.Offset)
				s.Equal(tc.expectedResult.HasMore, result.HasMore)
				s.Len(result.Requests, len(tc.accessResult.Requests))
			}
		})
	}
}

// ─── CACHE MANAGEMENT TESTS ────────────────────────────────────────────────

func (s *AuthzAdapterTestSuite) TestInvalidateUserCache() {
	testCases := []struct {
		name          string
		userID        uuid.UUID
		abacError     error
		expectedError string
	}{
		{
			name:   "successful_invalidation",
			userID: s.testUserID,
		},
		{
			name:          "abac_error",
			userID:        s.testUserID,
			abacError:     errors.NewBusinessError("CACHE_ERROR", "Failed to invalidate"),
			expectedError: "failed to invalidate user cache via ABAC",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.abacError != nil {
				s.mockABAC.EXPECT().InvalidateUserCache(gomock.Any(), tc.userID).
					Return(tc.abacError)
			} else {
				s.mockABAC.EXPECT().InvalidateUserCache(gomock.Any(), tc.userID).
					Return(nil)
			}

			err := s.adapter.InvalidateUserCache(s.ctx, tc.userID)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *AuthzAdapterTestSuite) TestInvalidatePolicyCache() {
	policyIDs := []uuid.UUID{uuid.New(), uuid.New()}

	testCases := []struct {
		name          string
		policyIDs     []uuid.UUID
		abacError     error
		expectedError string
	}{
		{
			name:      "successful_invalidation",
			policyIDs: policyIDs,
		},
		{
			name:          "abac_error",
			policyIDs:     policyIDs,
			abacError:     errors.NewBusinessError("CACHE_ERROR", "Failed to invalidate policies"),
			expectedError: "failed to invalidate policy cache via ABAC",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.abacError != nil {
				s.mockABAC.EXPECT().InvalidatePolicyCache(gomock.Any(), tc.policyIDs).
					Return(tc.abacError)
			} else {
				s.mockABAC.EXPECT().InvalidatePolicyCache(gomock.Any(), tc.policyIDs).
					Return(nil)
			}

			err := s.adapter.InvalidatePolicyCache(s.ctx, tc.policyIDs)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *AuthzAdapterTestSuite) TestGetCacheStatistics() {
	testCases := []struct {
		name           string
		abacResult     *activities.GetCacheStatsOutput
		abacError      error
		expectedResult *CacheStatistics
		expectedError  string
	}{
		{
			name: "successful_get_stats",
			abacResult: &activities.GetCacheStatsOutput{
				PolicyEvaluationStats: nil, // Simplified for test
				AttributeStats:        nil, // Simplified for test
				GeneratedAt:           time.Now(),
			},
			expectedResult: &CacheStatistics{
				HitRate:        0.8,
				MissRate:       0.2,
				TotalRequests:  100,
				CacheHits:      80,
				CacheMisses:    20,
				EvictionCount:  0,
				AverageLatency: 0,
			},
		},
		{
			name:          "abac_error",
			abacError:     errors.NewBusinessError("CACHE_ERROR", "Failed to get statistics"),
			expectedError: "failed to get cache statistics via ABAC",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.abacError != nil {
				s.mockABAC.EXPECT().GetCacheStatistics(gomock.Any()).
					Return(nil, tc.abacError)
			} else {
				s.mockABAC.EXPECT().GetCacheStatistics(gomock.Any()).
					Return(tc.abacResult, nil)
			}

			result, err := s.adapter.GetCacheStatistics(s.ctx)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				// Note: The actual implementation uses placeholder values
				// In a real implementation, these would be calculated from abac stats
			}
		})
	}
}

// ─── DECISION HISTORY TESTS ────────────────────────────────────────────────

func (s *AuthzAdapterTestSuite) TestGetDecisionHistory() {
	testCases := []struct {
		name           string
		userID         uuid.UUID
		limit          int
		abacResult     []*abac.DecisionHistoryEntry
		abacError      error
		expectedResult []*DecisionHistoryEntry
		expectedError  string
	}{
		{
			name:   "successful_get_history",
			userID: s.testUserID,
			limit:  10,
			abacResult: []*abac.DecisionHistoryEntry{
				{
					ID:             uuid.New(),
					UserID:         s.testUserID,
					ResourceType:   "document",
					ResourceID:     uuid.New(),
					Action:         "read",
					Decision:       types.PolicyDecisionAllow,
					Allowed:        true,
					EvaluationTime: 25,
					EvaluatedAt:    s.testTimeStamp,
					PolicyCount:    3,
					CacheHit:       false,
					RequestID:      s.testRequestID,
				},
			},
			expectedResult: []*DecisionHistoryEntry{
				{
					UserID:         s.testUserID,
					ResourceType:   "document",
					Action:         "read",
					Decision:       model.PolicyDecisionAllow,
					Allowed:        true,
					EvaluationTime: 25,
					EvaluatedAt:    s.testTimeStamp,
					PolicyCount:    3,
					CacheHit:       false,
					RequestID:      s.testRequestID,
				},
			},
		},
		{
			name:          "abac_error",
			userID:        s.testUserID,
			limit:         10,
			abacError:     errors.NewBusinessError("HISTORY_ERROR", "Failed to get history"),
			expectedError: "failed to get decision history via ABAC",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.abacError != nil {
				s.mockABAC.EXPECT().GetDecisionHistory(gomock.Any(), tc.userID, tc.limit).
					Return(nil, tc.abacError)
			} else {
				s.mockABAC.EXPECT().GetDecisionHistory(gomock.Any(), tc.userID, tc.limit).
					Return(tc.abacResult, nil)
			}

			result, err := s.adapter.GetDecisionHistory(s.ctx, tc.userID, tc.limit)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Len(result, len(tc.expectedResult))
				if len(result) > 0 {
					s.Equal(tc.expectedResult[0].UserID, result[0].UserID)
					s.Equal(tc.expectedResult[0].ResourceType, result[0].ResourceType)
					s.Equal(tc.expectedResult[0].Decision, result[0].Decision)
				}
			}
		})
	}
}
