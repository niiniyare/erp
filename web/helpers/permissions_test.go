package helpers

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/niiniyare/erp/internal/core/iam"
	"github.com/niiniyare/erp/internal/core/iam/authn"
	"github.com/niiniyare/erp/internal/core/iam/authz"
	"github.com/niiniyare/erp/internal/core/iam/model"
	"github.com/niiniyare/erp/internal/core/iam/policy"
)

// Mock IAM Service
type MockIAMService struct {
	mock.Mock
}

func (m *MockIAMService) Authorization() authz.Service {
	args := m.Called()
	return args.Get(0).(authz.Service)
}

func (m *MockIAMService) Authentication() authn.Service {
	args := m.Called()
	return args.Get(0).(authn.Service)
}

func (m *MockIAMService) Policy() policy.Service {
	args := m.Called()
	return args.Get(0).(policy.Service)
}

// Mock Authorization Service
type MockAuthorizationService struct {
	mock.Mock
}

func (m *MockAuthorizationService) EvaluatePermission(ctx context.Context, req *authz.PermissionEvaluationRequest) (*authz.PermissionEvaluationResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authz.PermissionEvaluationResult), args.Error(1)
}

func (m *MockAuthorizationService) BulkEvaluatePermissions(ctx context.Context, req *authz.BulkPermissionEvaluationRequest) (*authz.BulkPermissionEvaluationResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authz.BulkPermissionEvaluationResult), args.Error(1)
}

func (m *MockAuthorizationService) GetUserEffectivePermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (*authz.UserEffectivePermissions, error) {
	args := m.Called(ctx, userID, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authz.UserEffectivePermissions), args.Error(1)
}

func (m *MockAuthorizationService) CalculateRoleHierarchy(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) (*authz.RoleHierarchy, error) {
	args := m.Called(ctx, userID, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authz.RoleHierarchy), args.Error(1)
}

// Stub implementations for other interface methods
func (m *MockAuthorizationService) CreateAccessRequest(ctx context.Context, req *authz.CreateAccessRequestRequest) (*model.AccessRequest, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AccessRequest), args.Error(1)
}

func (m *MockAuthorizationService) GetAccessRequest(ctx context.Context, requestID uuid.UUID) (*model.AccessRequest, error) {
	args := m.Called(ctx, requestID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AccessRequest), args.Error(1)
}

func (m *MockAuthorizationService) ProcessAccessRequest(ctx context.Context, req *authz.ProcessAccessRequestRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthorizationService) ListAccessRequests(ctx context.Context, req *authz.ListAccessRequestsRequest) (*authz.ListAccessRequestsResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authz.ListAccessRequestsResult), args.Error(1)
}

func (m *MockAuthorizationService) CreateApprovalWorkflow(ctx context.Context, req *authz.CreateApprovalWorkflowRequest) (*model.ApprovalWorkflow, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ApprovalWorkflow), args.Error(1)
}

func (m *MockAuthorizationService) GetApprovalWorkflow(ctx context.Context, workflowID uuid.UUID) (*model.ApprovalWorkflow, error) {
	args := m.Called(ctx, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ApprovalWorkflow), args.Error(1)
}

func (m *MockAuthorizationService) UpdateApprovalWorkflow(ctx context.Context, req *authz.UpdateApprovalWorkflowRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthorizationService) EvaluateConditionalAccess(ctx context.Context, req *authz.ConditionalAccessRequest) (*authz.ConditionalAccessResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authz.ConditionalAccessResult), args.Error(1)
}

func (m *MockAuthorizationService) CreateConditionalAccessPolicy(ctx context.Context, req *authz.CreateConditionalAccessPolicyRequest) (*model.ConditionalAccessPolicy, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ConditionalAccessPolicy), args.Error(1)
}

func (m *MockAuthorizationService) UpdateConditionalAccessPolicy(ctx context.Context, req *authz.UpdateConditionalAccessPolicyRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthorizationService) GrantPermission(ctx context.Context, req *authz.GrantPermissionRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthorizationService) RevokePermission(ctx context.Context, req *authz.RevokePermissionRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthorizationService) ListUserPermissions(ctx context.Context, userID uuid.UUID, entityID *uuid.UUID) ([]*model.Permission, error) {
	args := m.Called(ctx, userID, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Permission), args.Error(1)
}

func (m *MockAuthorizationService) GetDecisionHistory(ctx context.Context, userID uuid.UUID, limit int) ([]*authz.DecisionHistoryEntry, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*authz.DecisionHistoryEntry), args.Error(1)
}

func (m *MockAuthorizationService) InvalidateUserCache(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthorizationService) InvalidatePolicyCache(ctx context.Context, policyIDs []uuid.UUID) error {
	args := m.Called(ctx, policyIDs)
	return args.Error(0)
}

func (m *MockAuthorizationService) GetCacheStatistics(ctx context.Context) (*authz.CacheStatistics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authz.CacheStatistics), args.Error(1)
}

// PermissionHelperTestSuite defines the test suite
type PermissionHelperTestSuite struct {
	suite.Suite
	helper     *PermissionHelper
	mockIAM    *MockIAMService
	mockAuthz  *MockAuthorizationService
	mockLogger *MockLogger
}

// SetupTest runs before each test
func (suite *PermissionHelperTestSuite) SetupTest() {
	suite.mockIAM = new(MockIAMService)
	suite.mockAuthz = new(MockAuthorizationService)
	suite.mockLogger = NewMockLogger()

	suite.mockIAM.On("Authorization").Return(suite.mockAuthz).Maybe()

	suite.helper = NewPermissionHelperWithLogger(suite.mockIAM, suite.mockLogger)
}

// TearDownTest runs after each test
func (suite *PermissionHelperTestSuite) TearDownTest() {
	suite.mockIAM.AssertExpectations(suite.T())
	suite.mockAuthz.AssertExpectations(suite.T())
	suite.mockLogger.Reset()
}

// TestPermissionHelperTestSuite runs the test suite
func TestPermissionHelperTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionHelperTestSuite))
}

// Test helper creation
func (suite *PermissionHelperTestSuite) TestNewPermissionHelper() {
	tests := []struct {
		name       string
		iamService iam.Service
		expectNil  bool
	}{
		{
			name:       "valid IAM service",
			iamService: suite.mockIAM,
			expectNil:  false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			helper := NewPermissionHelper(tt.iamService)
			require.NotNil(suite.T(), helper)
			require.NotNil(suite.T(), helper.config)
		})
	}
}

// Test PermissionContext validation
func (suite *PermissionHelperTestSuite) TestPermissionContextValidation() {
	tests := []struct {
		name        string
		context     *PermissionContext
		expectError bool
		errorType   error
	}{
		{
			name: "valid context",
			context: &PermissionContext{
				UserID:       uuid.New(),
				ResourceType: "document",
				Action:       "read",
			},
			expectError: false,
		},
		{
			name: "nil user ID",
			context: &PermissionContext{
				UserID:       uuid.Nil,
				ResourceType: "document",
				Action:       "read",
			},
			expectError: true,
			errorType:   ErrInvalidUserID,
		},
		{
			name: "empty resource type",
			context: &PermissionContext{
				UserID:       uuid.New(),
				ResourceType: "",
				Action:       "read",
			},
			expectError: true,
			errorType:   ErrInvalidResourceType,
		},
		{
			name: "empty action",
			context: &PermissionContext{
				UserID:       uuid.New(),
				ResourceType: "document",
				Action:       "",
			},
			expectError: true,
			errorType:   ErrInvalidAction,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := tt.context.Validate()

			if tt.expectError {
				require.Error(suite.T(), err)
				require.ErrorIs(suite.T(), err, tt.errorType)
			} else {
				require.NoError(suite.T(), err)
			}
		})
	}
}

// Test PermissionContextBuilder
func (suite *PermissionHelperTestSuite) TestPermissionContextBuilder() {
	userID := uuid.New()
	entityID := uuid.New()
	resourceID := uuid.New()

	tests := []struct {
		name     string
		build    func() *PermissionContext
		validate func(*testing.T, *PermissionContext)
	}{
		{
			name: "build simple context",
			build: func() *PermissionContext {
				return NewPermissionContext().
					WithUserID(userID).
					WithResourceType("document").
					WithAction("read").
					Build()
			},
			validate: func(t *testing.T, ctx *PermissionContext) {
				require.Equal(t, userID, ctx.UserID)
				require.Equal(t, "document", ctx.ResourceType)
				require.Equal(t, "read", ctx.Action)
			},
		},
		{
			name: "build complete context",
			build: func() *PermissionContext {
				return NewPermissionContext().
					WithUserID(userID).
					WithEntityID(entityID).
					WithResourceType("document").
					WithResourceID(resourceID).
					WithAction("write").
					WithContext("ip_address", "192.168.1.1").
					Build()
			},
			validate: func(t *testing.T, ctx *PermissionContext) {
				require.Equal(t, userID, ctx.UserID)
				require.NotNil(t, ctx.EntityID)
				require.Equal(t, entityID, *ctx.EntityID)
				require.NotNil(t, ctx.ResourceID)
				require.Equal(t, resourceID, *ctx.ResourceID)
				require.Equal(t, "write", ctx.Action)
				require.Equal(t, "192.168.1.1", ctx.Context["ip_address"])
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx := tt.build()
			require.NotNil(suite.T(), ctx)
			tt.validate(suite.T(), ctx)
		})
	}
}

// Test HasPermission
func (suite *PermissionHelperTestSuite) TestHasPermission() {
	userID := uuid.New()
	resourceID := uuid.New()

	tests := []struct {
		name          string
		permCtx       *PermissionContext
		mockSetup     func()
		expected      bool
		expectWarning bool
	}{
		{
			name:    "nil permission context",
			permCtx: nil,
			mockSetup: func() {
				// No mock setup needed
			},
			expected:      false,
			expectWarning: true,
		},
		{
			name: "nil user ID",
			permCtx: &PermissionContext{
				UserID:       uuid.Nil,
				ResourceType: "document",
				Action:       "read",
			},
			mockSetup: func() {
				// No mock setup needed
			},
			expected:      false,
			expectWarning: true,
		},
		{
			name: "permission allowed",
			permCtx: &PermissionContext{
				UserID:       userID,
				ResourceType: "document",
				ResourceID:   &resourceID,
				Action:       "read",
			},
			mockSetup: func() {
				suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.MatchedBy(func(req *authz.PermissionEvaluationRequest) bool {
					return req.UserID == userID && req.Action == "read"
				})).Return(&authz.PermissionEvaluationResult{
					Decision: "allow",
				}, nil).Once()
			},
			expected: true,
		},
		{
			name: "permission denied",
			permCtx: &PermissionContext{
				UserID:       userID,
				ResourceType: "document",
				ResourceID:   &resourceID,
				Action:       "write",
			},
			mockSetup: func() {
				suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.Anything).Return(&authz.PermissionEvaluationResult{
					Decision: "deny",
				}, nil).Once()
			},
			expected: false,
		},
		{
			name: "evaluation error",
			permCtx: &PermissionContext{
				UserID:       userID,
				ResourceType: "document",
				Action:       "delete",
			},
			mockSetup: func() {
				suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.Anything).Return(nil, errors.New("evaluation failed")).Once()
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.mockLogger.Reset()
			tt.mockSetup()

			result := suite.helper.HasPermission(context.Background(), tt.permCtx)

			require.Equal(suite.T(), tt.expected, result)

			if tt.expectWarning {
				require.NotEmpty(suite.T(), suite.mockLogger.WarnCalls)
			}
		})
	}
}

// Test HasAnyPermission
func (suite *PermissionHelperTestSuite) TestHasAnyPermission() {
	userID := uuid.New()

	tests := []struct {
		name        string
		permissions []*PermissionContext
		mockSetup   func()
		expected    bool
	}{
		{
			name:        "empty permissions list",
			permissions: []*PermissionContext{},
			mockSetup:   func() {},
			expected:    false,
		},
		{
			name: "first permission allowed",
			permissions: []*PermissionContext{
				{
					UserID:       userID,
					ResourceType: "document",
					Action:       "read",
				},
				{
					UserID:       userID,
					ResourceType: "document",
					Action:       "write",
				},
			},
			mockSetup: func() {
				suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.MatchedBy(func(req *authz.PermissionEvaluationRequest) bool {
					return req.Action == "read"
				})).Return(&authz.PermissionEvaluationResult{
					Decision: "allow",
				}, nil).Once()
			},
			expected: true,
		},
		{
			name: "second permission allowed",
			permissions: []*PermissionContext{
				{
					UserID:       userID,
					ResourceType: "document",
					Action:       "read",
				},
				{
					UserID:       userID,
					ResourceType: "document",
					Action:       "write",
				},
			},
			mockSetup: func() {
				suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.MatchedBy(func(req *authz.PermissionEvaluationRequest) bool {
					return req.Action == "read"
				})).Return(&authz.PermissionEvaluationResult{
					Decision: "deny",
				}, nil).Once()

				suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.MatchedBy(func(req *authz.PermissionEvaluationRequest) bool {
					return req.Action == "write"
				})).Return(&authz.PermissionEvaluationResult{
					Decision: "allow",
				}, nil).Once()
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.mockLogger.Reset()
			tt.mockSetup()

			result := suite.helper.HasAnyPermission(context.Background(), tt.permissions)

			require.Equal(suite.T(), tt.expected, result)
		})
	}
}

// Test CRUD operations
func (suite *PermissionHelperTestSuite) TestCRUDOperations() {
	userID := uuid.New()
	resourceID := uuid.New()
	entityID := uuid.New()

	tests := []struct {
		name      string
		operation func() bool
		action    string
		mockSetup func(string)
		expected  bool
	}{
		{
			name: "CanRead",
			operation: func() bool {
				return suite.helper.CanRead(context.Background(), userID, "document", &resourceID, &entityID)
			},
			action: ActionRead,
			mockSetup: func(action string) {
				suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.MatchedBy(func(req *authz.PermissionEvaluationRequest) bool {
					return req.Action == action
				})).Return(&authz.PermissionEvaluationResult{
					Decision: "allow",
				}, nil).Once()
			},
			expected: true,
		},
		{
			name: "CanWrite",
			operation: func() bool {
				return suite.helper.CanWrite(context.Background(), userID, "document", &resourceID, &entityID)
			},
			action: ActionWrite,
			mockSetup: func(action string) {
				suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.MatchedBy(func(req *authz.PermissionEvaluationRequest) bool {
					return req.Action == action
				})).Return(&authz.PermissionEvaluationResult{
					Decision: "allow",
				}, nil).Once()
			},
			expected: true,
		},
		{
			name: "CanDelete",
			operation: func() bool {
				return suite.helper.CanDelete(context.Background(), userID, "document", &resourceID, &entityID)
			},
			action: ActionDelete,
			mockSetup: func(action string) {
				suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.MatchedBy(func(req *authz.PermissionEvaluationRequest) bool {
					return req.Action == action
				})).Return(&authz.PermissionEvaluationResult{
					Decision: "allow",
				}, nil).Once()
			},
			expected: true,
		},
		{
			name: "CanCreate",
			operation: func() bool {
				return suite.helper.CanCreate(context.Background(), userID, "document", &entityID)
			},
			action: ActionCreate,
			mockSetup: func(action string) {
				suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.MatchedBy(func(req *authz.PermissionEvaluationRequest) bool {
					return req.Action == action
				})).Return(&authz.PermissionEvaluationResult{
					Decision: "allow",
				}, nil).Once()
			},
			expected: true,
		},
		{
			name: "CanManage",
			operation: func() bool {
				return suite.helper.CanManage(context.Background(), userID, "document", &resourceID, &entityID)
			},
			action: ActionManage,
			mockSetup: func(action string) {
				suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.MatchedBy(func(req *authz.PermissionEvaluationRequest) bool {
					return req.Action == action
				})).Return(&authz.PermissionEvaluationResult{
					Decision: "allow",
				}, nil).Once()
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.mockLogger.Reset()
			tt.mockSetup(tt.action)

			result := tt.operation()

			require.Equal(suite.T(), tt.expected, result)
		})
	}
}

// Test admin checks
func (suite *PermissionHelperTestSuite) TestAdminChecks() {
	userID := uuid.New()
	tenantID := uuid.New()

	tests := []struct {
		name      string
		operation func() bool
		mockSetup func()
		expected  bool
	}{
		{
			name: "IsSystemAdmin - true",
			operation: func() bool {
				return suite.helper.IsSystemAdmin(context.Background(), userID)
			},
			mockSetup: func() {
				suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.MatchedBy(func(req *authz.PermissionEvaluationRequest) bool {
					return req.ResourceType == ResourceTypeSystem && req.Action == ActionAdmin
				})).Return(&authz.PermissionEvaluationResult{
					Decision: "allow",
				}, nil).Once()
			},
			expected: true,
		},
		{
			name: "IsTenantAdmin - true",
			operation: func() bool {
				return suite.helper.IsTenantAdmin(context.Background(), userID, tenantID)
			},
			mockSetup: func() {
				suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.MatchedBy(func(req *authz.PermissionEvaluationRequest) bool {
					return req.ResourceType == ResourceTypeTenant && req.Action == ActionAdmin
				})).Return(&authz.PermissionEvaluationResult{
					Decision: "allow",
				}, nil).Once()
			},
			expected: true,
		},
		{
			name: "IsSystemAdmin with nil user ID",
			operation: func() bool {
				return suite.helper.IsSystemAdmin(context.Background(), uuid.Nil)
			},
			mockSetup: func() {},
			expected:  false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.mockLogger.Reset()
			tt.mockSetup()

			result := tt.operation()

			require.Equal(suite.T(), tt.expected, result)
		})
	}
}

// Test batch operations
func (suite *PermissionHelperTestSuite) TestCheckMultipleResources() {
	userID := uuid.New()
	resourceIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	suite.Run("check multiple resources", func() {
		suite.mockLogger.Reset()

		// Setup mock to allow first and third resources
		suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.MatchedBy(func(req *authz.PermissionEvaluationRequest) bool {
			return *req.ResourceID == resourceIDs[0]
		})).Return(&authz.PermissionEvaluationResult{
			Decision: "allow",
		}, nil).Once()

		suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.MatchedBy(func(req *authz.PermissionEvaluationRequest) bool {
			return *req.ResourceID == resourceIDs[1]
		})).Return(&authz.PermissionEvaluationResult{
			Decision: "deny",
		}, nil).Once()

		suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.MatchedBy(func(req *authz.PermissionEvaluationRequest) bool {
			return *req.ResourceID == resourceIDs[2]
		})).Return(&authz.PermissionEvaluationResult{
			Decision: "allow",
		}, nil).Once()

		results := suite.helper.CheckMultipleResources(context.Background(), userID, "document", resourceIDs, "read", nil)

		require.Len(suite.T(), results, 3)
		require.True(suite.T(), results[resourceIDs[0]])
		require.False(suite.T(), results[resourceIDs[1]])
		require.True(suite.T(), results[resourceIDs[2]])
	})
}

// Test thread safety
func (suite *PermissionHelperTestSuite) TestThreadSafety() {
	suite.Run("concurrent config updates", func() {
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				config := DefaultPermissionHelperConfig()
				suite.helper.UpdateConfig(config)
			}()
		}

		wg.Wait()
	})

	suite.Run("concurrent config reads", func() {
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				config := suite.helper.GetConfig()
				require.NotNil(suite.T(), &config)
			}()
		}

		wg.Wait()
	})
}

// Test logger integration
func (suite *PermissionHelperTestSuite) TestLoggerIntegration() {
	userID := uuid.New()

	suite.Run("logs warning on nil context", func() {
		suite.mockLogger.Reset()

		suite.helper.HasPermission(context.Background(), nil)

		require.NotEmpty(suite.T(), suite.mockLogger.WarnCalls)
	})

	suite.Run("logs debug on success", func() {
		suite.mockLogger.Reset()

		suite.mockAuthz.On("EvaluatePermission", mock.Anything, mock.Anything).Return(&authz.PermissionEvaluationResult{
			Decision: "allow",
		}, nil).Once()

		suite.helper.IsSystemAdmin(context.Background(), userID)

		require.NotEmpty(suite.T(), suite.mockLogger.DebugCalls)
	})
}

// Benchmark tests
func BenchmarkPermissionHelper(b *testing.B) {
	mockIAM := new(MockIAMService)
	mockAuthz := new(MockAuthorizationService)
	mockIAM.On("Authorization").Return(mockAuthz)

	helper := NewPermissionHelper(mockIAM)
	userID := uuid.New()
	ctx := context.Background()

	b.Run("HasPermission", func(b *testing.B) {
		mockAuthz.On("EvaluatePermission", mock.Anything, mock.Anything).Return(&authz.PermissionEvaluationResult{
			Decision: "allow",
		}, nil)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = helper.HasPermission(ctx, &PermissionContext{
				UserID:       userID,
				ResourceType: "document",
				Action:       "read",
			})
		}
	})
}
