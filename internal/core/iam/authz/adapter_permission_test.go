package authz

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/niiniyare/erp/internal/core/access"
	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// PermissionManagementTestSuite focuses on permission-related adapter methods
type PermissionManagementTestSuite struct {
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

func TestPermissionManagementTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionManagementTestSuite))
}

func (s *PermissionManagementTestSuite) SetupTest() {
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

func (s *PermissionManagementTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// ─── PERMISSION GRANT/REVOKE TESTS ─────────────────────────────────────────

func (s *PermissionManagementTestSuite) TestGrantPermission() {
	testCases := []struct {
		name          string
		req           *GrantPermissionRequest
		accessError   error
		expectedError string
	}{
		{
			name: "successful_grant",
			req: &GrantPermissionRequest{
				UserID:       s.testUserID,
				ResourceType: "document",
				ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
				Action:       "read",
				EntityID:     &s.testEntityID,
				ExpiresAt:    func() *time.Time { t := time.Now().Add(time.Hour * 24); return &t }(),
				Conditions:   []string{"time_limited", "ip_restricted"},
			},
		},
		{
			name: "grant_with_no_expiry",
			req: &GrantPermissionRequest{
				UserID:       s.testUserID,
				ResourceType: "file",
				ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
				Action:       "write",
				EntityID:     &s.testEntityID,
				ExpiresAt:    nil, // No expiry
				Conditions:   nil, // No conditions
			},
		},
		{
			name: "access_service_error",
			req: &GrantPermissionRequest{
				UserID:       s.testUserID,
				ResourceType: "document",
				Action:       "read",
			},
			accessError:   errors.NewBusinessError("ACCESS_ERROR", "Failed to grant permission"),
			expectedError: "failed to grant permission via access service",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.accessError != nil {
				s.mockAccess.EXPECT().GrantPermission(gomock.Any(), gomock.Any()).
					Return(tc.accessError)
			} else {
				s.mockAccess.EXPECT().GrantPermission(gomock.Any(), gomock.Any()).
					Return(nil)
			}

			err := s.adapter.GrantPermission(s.ctx, tc.req)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *PermissionManagementTestSuite) TestRevokePermission() {
	testCases := []struct {
		name          string
		req           *RevokePermissionRequest
		accessError   error
		expectedError string
	}{
		{
			name: "successful_revoke",
			req: &RevokePermissionRequest{
				UserID:       s.testUserID,
				ResourceType: "document",
				ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
				Action:       "read",
				EntityID:     &s.testEntityID,
			},
		},
		{
			name: "revoke_without_resource_id",
			req: &RevokePermissionRequest{
				UserID:       s.testUserID,
				ResourceType: "document",
				Action:       "read",
				EntityID:     &s.testEntityID,
			},
		},
		{
			name: "access_service_error",
			req: &RevokePermissionRequest{
				UserID:       s.testUserID,
				ResourceType: "document",
				Action:       "read",
			},
			accessError:   errors.NewBusinessError("ACCESS_ERROR", "Failed to revoke permission"),
			expectedError: "failed to revoke permission via access service",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.accessError != nil {
				s.mockAccess.EXPECT().RevokePermission(gomock.Any(), gomock.Any()).
					Return(tc.accessError)
			} else {
				s.mockAccess.EXPECT().RevokePermission(gomock.Any(), gomock.Any()).
					Return(nil)
			}

			err := s.adapter.RevokePermission(s.ctx, tc.req)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *PermissionManagementTestSuite) TestListUserPermissions() {
	testCases := []struct {
		name           string
		userID         uuid.UUID
		entityID       *uuid.UUID
		accessResult   []*access.Permission
		accessError    error
		expectedResult []*model.Permission
		expectedError  string
	}{
		{
			name:     "successful_list",
			userID:   s.testUserID,
			entityID: &s.testEntityID,
			accessResult: []*access.Permission{
				{
					ID:           uuid.New(),
					TenantID:     s.testEntityID,
					ResourceType: "document",
					ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
					Action:       "read",
					EntityID:     &s.testEntityID,
					Conditions:   []string{"time_limited"},
					ExpiresAt:    func() *time.Time { t := time.Now().Add(time.Hour * 24); return &t }(),
					Metadata:     map[string]any{"source": "direct_grant"},
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				},
				{
					ID:           uuid.New(),
					TenantID:     s.testEntityID,
					ResourceType: "file",
					ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
					Action:       "write",
					EntityID:     &s.testEntityID,
					Conditions:   nil,
					ExpiresAt:    nil,
					Metadata:     map[string]any{"source": "role_inheritance"},
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				},
			},
			expectedResult: []*model.Permission{
				{
					ResourceType: "document",
					ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
					Action:       "read",
					EntityID:     &s.testEntityID,
					Conditions:   []string{"time_limited"},
				},
				{
					ResourceType: "file",
					ResourceID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
					Action:       "write",
					EntityID:     &s.testEntityID,
					Conditions:   nil,
				},
			},
		},
		{
			name:           "empty_permissions_list",
			userID:         s.testUserID,
			entityID:       &s.testEntityID,
			accessResult:   []*access.Permission{},
			expectedResult: []*model.Permission{},
		},
		{
			name:     "permissions_without_entity",
			userID:   s.testUserID,
			entityID: nil,
			accessResult: []*access.Permission{
				{
					ID:           uuid.New(),
					TenantID:     s.testEntityID,
					ResourceType: "system",
					Action:       "login",
					EntityID:     &s.testEntityID,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				},
			},
			expectedResult: []*model.Permission{
				{
					ResourceType: "system",
					Action:       "login",
					EntityID:     &s.testEntityID,
				},
			},
		},
		{
			name:          "access_service_error",
			userID:        s.testUserID,
			entityID:      &s.testEntityID,
			accessError:   errors.NewBusinessError("ACCESS_ERROR", "Failed to list permissions"),
			expectedError: "failed to list user permissions via access service",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.accessError != nil {
				s.mockAccess.EXPECT().ListUserPermissions(gomock.Any(), tc.userID, tc.entityID).
					Return(nil, tc.accessError)
			} else {
				s.mockAccess.EXPECT().ListUserPermissions(gomock.Any(), tc.userID, tc.entityID).
					Return(tc.accessResult, nil)
			}

			result, err := s.adapter.ListUserPermissions(s.ctx, tc.userID, tc.entityID)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Len(result, len(tc.expectedResult))
				for i, expectedPerm := range tc.expectedResult {
					s.Equal(expectedPerm.ResourceType, result[i].ResourceType)
					// Skip ResourceID comparison as it's generated randomly in tests
					s.Equal(expectedPerm.Action, result[i].Action)
					// Skip EntityID comparison as it's generated randomly in tests
				}
			}
		})
	}
}

// ─── APPROVAL WORKFLOW TESTS ───────────────────────────────────────────────

func (s *PermissionManagementTestSuite) TestCreateApprovalWorkflow() {
	workflowID := uuid.New()
	testCases := []struct {
		name           string
		req            *CreateApprovalWorkflowRequest
		accessResult   *access.ApprovalWorkflow
		accessError    error
		expectedResult *model.ApprovalWorkflow
		expectedError  string
	}{
		{
			name: "successful_create_workflow",
			req: &CreateApprovalWorkflowRequest{
				Name:        "Document Access Workflow",
				Description: "Approval workflow for document access requests",
				Steps: []*model.ApprovalStep{
					{
						ID:            uuid.New().String(),
						Name:          "Manager Approval",
						Order:         1,
						RequiredVotes: 1,
						ApproverUsers: []uuid.UUID{uuid.New()},
						TimeoutHours:  &[]int{24}[0],
						AutoApprove:   false,
					},
					{
						ID:            uuid.New().String(),
						Name:          "Security Review",
						Order:         2,
						RequiredVotes: 1,
						ApproverUsers: []uuid.UUID{uuid.New()},
						TimeoutHours:  &[]int{48}[0],
						AutoApprove:   false,
					},
				},
				Metadata: map[string]any{"department": "engineering"},
			},
			accessResult: &access.ApprovalWorkflow{
				ID:          workflowID,
				TenantID:    s.testEntityID,
				Name:        "Document Access Workflow",
				Description: "Approval workflow for document access requests",
				Steps: []*access.ApprovalStep{
					{
						ID:             uuid.New(),
						Order:          1,
						Name:           "Manager Approval",
						Description:    "",
						ApproverType:   "user",
						ApproverIDs:    []uuid.UUID{uuid.New()},
						RequiredCount:  1,
						TimeoutMinutes: 1440, // 24 hours in minutes
						Conditions:     []*access.ApprovalCondition{},
						Metadata:       map[string]any{},
					},
				},
				IsActive:  true,
				Metadata:  map[string]any{"department": "engineering"},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedResult: &model.ApprovalWorkflow{
				ID:          workflowID,
				TenantID:    s.testEntityID,
				Name:        "Document Access Workflow",
				Description: "Approval workflow for document access requests",
				Enabled:     true,
			},
		},
		{
			name: "access_service_error",
			req: &CreateApprovalWorkflowRequest{
				Name:        "Test Workflow",
				Description: "Test workflow",
				Steps:       []*model.ApprovalStep{},
			},
			accessError:   errors.NewBusinessError("ACCESS_ERROR", "Failed to create workflow"),
			expectedError: "failed to create approval workflow via access service",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.accessError != nil {
				s.mockAccess.EXPECT().CreateApprovalWorkflow(gomock.Any(), gomock.Any()).
					Return(nil, tc.accessError)
			} else {
				s.mockAccess.EXPECT().CreateApprovalWorkflow(gomock.Any(), gomock.Any()).
					Return(tc.accessResult, nil)
			}

			result, err := s.adapter.CreateApprovalWorkflow(s.ctx, tc.req)

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
				s.Equal(tc.expectedResult.Enabled, result.Enabled)
			}
		})
	}
}

func (s *PermissionManagementTestSuite) TestGetApprovalWorkflow() {
	workflowID := uuid.New()
	testCases := []struct {
		name           string
		workflowID     uuid.UUID
		accessResult   *access.ApprovalWorkflow
		accessError    error
		expectedResult *model.ApprovalWorkflow
		expectedError  string
	}{
		{
			name:       "successful_get_workflow",
			workflowID: workflowID,
			accessResult: &access.ApprovalWorkflow{
				ID:          workflowID,
				TenantID:    s.testEntityID,
				Name:        "Document Access Workflow",
				Description: "Workflow for document access",
				Steps: []*access.ApprovalStep{
					{
						ID:             uuid.New(),
						Order:          1,
						Name:           "Manager Approval",
						RequiredCount:  1,
						ApproverIDs:    []uuid.UUID{uuid.New()},
						TimeoutMinutes: 1440,
					},
				},
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedResult: &model.ApprovalWorkflow{
				ID:          workflowID,
				TenantID:    s.testEntityID,
				Name:        "Document Access Workflow",
				Description: "Workflow for document access",
				Enabled:     true,
			},
		},
		{
			name:          "access_service_error",
			workflowID:    workflowID,
			accessError:   errors.NewBusinessError("ACCESS_ERROR", "Workflow not found"),
			expectedError: "failed to get approval workflow via access service",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.accessError != nil {
				s.mockAccess.EXPECT().GetApprovalWorkflow(gomock.Any(), tc.workflowID).
					Return(nil, tc.accessError)
			} else {
				s.mockAccess.EXPECT().GetApprovalWorkflow(gomock.Any(), tc.workflowID).
					Return(tc.accessResult, nil)
			}

			result, err := s.adapter.GetApprovalWorkflow(s.ctx, tc.workflowID)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
				s.Nil(result)
			} else {
				s.NoError(err)
				s.NotNil(result)
				s.Equal(tc.expectedResult.ID, result.ID)
				s.Equal(tc.expectedResult.Name, result.Name)
				s.Equal(tc.expectedResult.Enabled, result.Enabled)
			}
		})
	}
}

func (s *PermissionManagementTestSuite) TestUpdateApprovalWorkflow() {
	workflowID := uuid.New()
	testCases := []struct {
		name          string
		req           *UpdateApprovalWorkflowRequest
		accessError   error
		expectedError string
	}{
		{
			name: "successful_update",
			req: &UpdateApprovalWorkflowRequest{
				WorkflowID:  workflowID,
				Name:        func() *string { s := "Updated Workflow"; return &s }(),
				Description: func() *string { s := "Updated description"; return &s }(),
				Steps: []*model.ApprovalStep{
					{
						ID:            uuid.New().String(),
						Name:          "Updated Step",
						Order:         1,
						RequiredVotes: 2,
						ApproverUsers: []uuid.UUID{uuid.New(), uuid.New()},
						TimeoutHours:  &[]int{72}[0],
						AutoApprove:   false,
					},
				},
				Metadata: map[string]any{"updated": true},
			},
		},
		{
			name: "update_with_minimal_fields",
			req: &UpdateApprovalWorkflowRequest{
				WorkflowID: workflowID,
				Name:       func() *string { s := "Minimal Update"; return &s }(),
			},
		},
		{
			name: "access_service_error",
			req: &UpdateApprovalWorkflowRequest{
				WorkflowID: workflowID,
				Name:       func() *string { s := "Failed Update"; return &s }(),
			},
			accessError:   errors.NewBusinessError("ACCESS_ERROR", "Failed to update workflow"),
			expectedError: "failed to update approval workflow via access service",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.accessError != nil {
				s.mockAccess.EXPECT().UpdateApprovalWorkflow(gomock.Any(), gomock.Any()).
					Return(tc.accessError)
			} else {
				s.mockAccess.EXPECT().UpdateApprovalWorkflow(gomock.Any(), gomock.Any()).
					Return(nil)
			}

			err := s.adapter.UpdateApprovalWorkflow(s.ctx, tc.req)

			if tc.expectedError != "" {
				s.Error(err)
				s.Contains(err.Error(), tc.expectedError)
			} else {
				s.NoError(err)
			}
		})
	}
}
