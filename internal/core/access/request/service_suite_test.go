//go:build unit
// +build unit

package request

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/access/approval"
	"github.com/niiniyare/erp/internal/core/access/execution"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/notification"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockAccessRequestRepository implements the AccessRequestRepository interface for testing
type MockAccessRequestRepository struct {
	mock.Mock
}

func (m *MockAccessRequestRepository) CreateAccessRequest(ctx context.Context, req *CreateAccessRequestRequest, requesterID uuid.UUID) (*AccessRequest, error) {
	args := m.Called(ctx, req, requesterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccessRequest), args.Error(1)
}

func (m *MockAccessRequestRepository) GetAccessRequestByID(ctx context.Context, id uuid.UUID) (*AccessRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccessRequest), args.Error(1)
}

func (m *MockAccessRequestRepository) ListAccessRequestsWithDetails(ctx context.Context, req *ListAccessRequestsRequest) ([]*AccessRequestWithDetails, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*AccessRequestWithDetails), args.Error(1)
}

func (m *MockAccessRequestRepository) UpdateAccessRequestStatus(ctx context.Context, id uuid.UUID, req *UpdateAccessRequestRequest, approverID uuid.UUID) (*AccessRequest, error) {
	args := m.Called(ctx, id, req, approverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccessRequest), args.Error(1)
}

func (m *MockAccessRequestRepository) RevokeAccessRequest(ctx context.Context, id uuid.UUID) (*AccessRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccessRequest), args.Error(1)
}

func (m *MockAccessRequestRepository) GetUserAccessRequestHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AccessRequest, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*AccessRequest), args.Error(1)
}

func (m *MockAccessRequestRepository) GetPendingRequestsForApprover(ctx context.Context, limit, offset int) ([]*AccessRequest, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*AccessRequest), args.Error(1)
}

func (m *MockAccessRequestRepository) GetAccessRequestStats(ctx context.Context, fromDate, toDate *time.Time) (*AccessRequestStats, error) {
	args := m.Called(ctx, fromDate, toDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AccessRequestStats), args.Error(1)
}

func (m *MockAccessRequestRepository) GetExpiredAccessRequests(ctx context.Context) ([]*AccessRequest, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*AccessRequest), args.Error(1)
}

func (m *MockAccessRequestRepository) ExpireAccessRequest(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockUserService implements the UserService interface for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

// Mock implementations for other services
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

type MockTracingService struct {
	mock.Mock
}

func (m *MockTracingService) StartSpan(ctx context.Context, name string, opts ...tracing.SpanOption) (context.Context, tracing.Span) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(context.Context), args.Get(1).(tracing.Span)
}

type MockSpan struct {
	mock.Mock
}

func (m *MockSpan) End() {
	m.Called()
}

func (m *MockSpan) SetAttributes(attrs ...interface{}) {
	m.Called(attrs)
}

func (m *MockSpan) RecordError(err error, opts ...tracing.SpanOption) {
	m.Called(err, opts)
}

func (m *MockSpan) SetStatus(code tracing.SpanStatusCode, description string) {
	m.Called(code, description)
}

type MockMetricsProvider struct {
	mock.Mock
}

func (m *MockMetricsProvider) IncrementCounter(name string, tags map[string]any) {
	m.Called(name, tags)
}

func (m *MockMetricsProvider) ObserveHistogram(name string, value float64, tags map[string]any) {
	m.Called(name, value, tags)
}

func (m *MockMetricsProvider) RecordGauge(name string, value float64, tags map[string]any) {
	m.Called(name, value, tags)
}

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) NotifyApprovers(ctx context.Context, request *AccessRequest) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockNotificationService) NotifyRequester(ctx context.Context, request *AccessRequest, action string) error {
	args := m.Called(ctx, request, action)
	return args.Error(0)
}

type MockApproverService struct {
	mock.Mock
}

func (m *MockApproverService) ValidateApprover(ctx context.Context, approverID uuid.UUID, request *AccessRequest) (*approval.ApprovalValidationResult, error) {
	args := m.Called(ctx, approverID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*approval.ApprovalValidationResult), args.Error(1)
}

type MockExecutionService struct {
	mock.Mock
}

func (m *MockExecutionService) ExecuteAccessRequest(ctx context.Context, req *execution.AccessRequest) (*execution.ExecutionResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*execution.ExecutionResult), args.Error(1)
}

func (m *MockExecutionService) RevokeAccessRequest(ctx context.Context, req *execution.AccessRequest) (*execution.ExecutionResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*execution.ExecutionResult), args.Error(1)
}

type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogAccessRequestCreated(ctx context.Context, request *AccessRequest, requesterID uuid.UUID) error {
	args := m.Called(ctx, request, requesterID)
	return args.Error(0)
}

// AccessRequestServiceTestSuite is the main test suite for the access request service
type AccessRequestServiceTestSuite struct {
	suite.Suite
	service           AccessRequestService
	mockRepo          *MockAccessRequestRepository
	mockCache         *MockCache
	mockTracing       *MockTracingService
	mockSpan          *MockSpan
	mockMetrics       *MockMetricsProvider
	mockUserService   *MockUserService
	mockNotification  *MockNotificationService
	mockApprover      *MockApproverService
	mockExecution     *MockExecutionService
	mockAudit         *MockAuditService
	ctx               context.Context
	testAccessRequest *AccessRequest
	testUser          *User
	testUserID        uuid.UUID
	testTenantID      uuid.UUID
	testEntityID      uuid.UUID
	testRequestID     uuid.UUID
}

func (suite *AccessRequestServiceTestSuite) SetupSuite() {
	// Suite-level setup
	suite.testUserID = uuid.New()
	suite.testTenantID = uuid.New()
	suite.testEntityID = uuid.New()
	suite.testRequestID = uuid.New()

	suite.testUser = &User{
		ID:       suite.testUserID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	suite.testAccessRequest = &AccessRequest{
		ID:             suite.testRequestID,
		TenantID:       suite.testTenantID,
		RequestType:    RequestTypeRoleAssignment,
		RequesterID:    suite.testUserID,
		EntityID:       suite.testEntityID,
		ApprovalStatus: ApprovalStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func (suite *AccessRequestServiceTestSuite) SetupTest() {
	// Test-level setup
	suite.mockRepo = new(MockAccessRequestRepository)
	suite.mockCache = new(MockCache)
	suite.mockTracing = new(MockTracingService)
	suite.mockSpan = new(MockSpan)
	suite.mockMetrics = new(MockMetricsProvider)
	suite.mockUserService = new(MockUserService)
	suite.mockNotification = new(MockNotificationService)
	suite.mockApprover = new(MockApproverService)
	suite.mockExecution = new(MockExecutionService)
	suite.mockAudit = new(MockAuditService)
	suite.ctx = context.Background()

	// Setup default tracing mock behavior
	suite.mockTracing.On("StartSpan", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(suite.ctx, suite.mockSpan)
	suite.mockSpan.On("End").Return()
	suite.mockSpan.On("SetAttributes", mock.Anything).Return()

	// Create service with mocked dependencies
	suite.service = NewAccessRequestService(
		suite.mockRepo,
		suite.mockCache,
		suite.mockTracing,
		suite.mockMetrics,
		suite.mockUserService,
		suite.mockNotification,
		suite.mockApprover,
		suite.mockExecution,
		suite.mockAudit,
	)
}

func (suite *AccessRequestServiceTestSuite) TearDownTest() {
	// Test-level cleanup
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
	suite.mockTracing.AssertExpectations(suite.T())
	suite.mockSpan.AssertExpectations(suite.T())
	suite.mockMetrics.AssertExpectations(suite.T())
	suite.mockUserService.AssertExpectations(suite.T())
}

// Test CreateAccessRequest - Success Case
func (suite *AccessRequestServiceTestSuite) TestCreateAccessRequest_Success() {
	// Arrange
	roleID := uuid.New()
	req := &CreateAccessRequestRequest{
		EntityID:      suite.testEntityID,
		RequestType:   RequestTypeRoleAssignment,
		RoleID:        &roleID,
		Justification: "Need role for project access",
		DurationHours: int32Ptr(24),
	}

	expectedRequest := &AccessRequest{
		ID:             uuid.New(),
		TenantID:       suite.testTenantID,
		RequestType:    req.RequestType,
		RequesterID:    suite.testUserID,
		EntityID:       req.EntityID,
		RoleID:         req.RoleID,
		ApprovalStatus: ApprovalStatusPending,
		Justification:  req.Justification,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Mock user service to validate requester
	suite.mockUserService.On("GetUserByID", suite.ctx, suite.testUserID).Return(suite.testUser, nil)

	// Mock repository call for duplicate check
	suite.mockRepo.On("ListAccessRequestsWithDetails", suite.ctx, mock.AnythingOfType("*request.ListAccessRequestsRequest")).Return([]*AccessRequestWithDetails{}, nil)

	// Mock repository call for creation
	suite.mockRepo.On("CreateAccessRequest", suite.ctx, req, suite.testUserID).Return(expectedRequest, nil)

	// Mock metrics
	suite.mockMetrics.On("IncrementCounter", "access_request_created", mock.AnythingOfType("map[string]interface {}")).Return()

	// Act
	result, err := suite.service.CreateAccessRequest(suite.ctx, req, suite.testUserID)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(expectedRequest.RequestType, result.RequestType)
	suite.Equal(expectedRequest.RequesterID, result.RequesterID)
	suite.Equal(expectedRequest.EntityID, result.EntityID)
	suite.Equal(ApprovalStatusPending, result.ApprovalStatus)
}

// Test CreateAccessRequest - Validation Error (Missing Role ID)
func (suite *AccessRequestServiceTestSuite) TestCreateAccessRequest_ValidationError() {
	// Arrange
	req := &CreateAccessRequestRequest{
		EntityID:      suite.testEntityID,
		RequestType:   RequestTypeRoleAssignment,
		RoleID:        nil, // Missing required role ID
		Justification: "Need role for project access",
	}

	// Mock user service to validate requester
	suite.mockUserService.On("GetUserByID", suite.ctx, suite.testUserID).Return(suite.testUser, nil)

	// Mock metrics for validation failure
	suite.mockMetrics.On("IncrementCounter", "access_request_validation_failed", mock.AnythingOfType("map[string]interface {}")).Return()

	// Act
	result, err := suite.service.CreateAccessRequest(suite.ctx, req, suite.testUserID)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "ROLE_ID_REQUIRED")
}

// Test CreateAccessRequest - Duplicate Request
func (suite *AccessRequestServiceTestSuite) TestCreateAccessRequest_DuplicateRequest() {
	// Arrange
	roleID := uuid.New()
	req := &CreateAccessRequestRequest{
		EntityID:      suite.testEntityID,
		RequestType:   RequestTypeRoleAssignment,
		RoleID:        &roleID,
		Justification: "Need role for project access",
	}

	existingRequest := &AccessRequestWithDetails{
		AccessRequest: &AccessRequest{
			ID:             uuid.New(),
			RequestType:    RequestTypeRoleAssignment,
			RequesterID:    suite.testUserID,
			RoleID:         &roleID,
			ApprovalStatus: ApprovalStatusPending,
		},
	}

	// Mock user service to validate requester
	suite.mockUserService.On("GetUserByID", suite.ctx, suite.testUserID).Return(suite.testUser, nil)

	// Mock repository call for duplicate check - return existing request
	suite.mockRepo.On("ListAccessRequestsWithDetails", suite.ctx, mock.AnythingOfType("*request.ListAccessRequestsRequest")).Return([]*AccessRequestWithDetails{existingRequest}, nil)

	// Mock metrics for duplicate detection
	suite.mockMetrics.On("IncrementCounter", "access_request_duplicate", mock.AnythingOfType("map[string]interface {}")).Return()

	// Act
	result, err := suite.service.CreateAccessRequest(suite.ctx, req, suite.testUserID)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "DUPLICATE_REQUEST")
}

// Test ProcessAccessRequest - Approve Success
func (suite *AccessRequestServiceTestSuite) TestProcessAccessRequest_ApproveSuccess() {
	// Arrange
	approverID := uuid.New()
	action := &AccessRequestApprovalRequest{
		Action:   "approve",
		Comments: stringPtr("Approved for business need"),
	}

	approvedRequest := &AccessRequest{
		ID:             suite.testRequestID,
		TenantID:       suite.testTenantID,
		RequestType:    RequestTypeRoleAssignment,
		RequesterID:    suite.testUserID,
		EntityID:       suite.testEntityID,
		ApprovalStatus: ApprovalStatusApproved,
		ApprovedBy:     &approverID,
		CreatedAt:      suite.testAccessRequest.CreatedAt,
		UpdatedAt:      time.Now(),
	}

	executionResult := &execution.ExecutionResult{
		Success:       true,
		ExecutedAt:    time.Now(),
		AccessGranted: []string{"role:admin"},
		ErrorMessage:  "",
	}

	// Mock repository calls
	suite.mockRepo.On("GetAccessRequestByID", suite.ctx, suite.testRequestID).Return(suite.testAccessRequest, nil)
	suite.mockRepo.On("UpdateAccessRequestStatus", suite.ctx, suite.testRequestID, mock.AnythingOfType("*request.UpdateAccessRequestRequest"), approverID).Return(approvedRequest, nil)

	// Mock execution service
	suite.mockExecution.On("ExecuteAccessRequest", suite.ctx, mock.AnythingOfType("*execution.AccessRequest")).Return(executionResult, nil)

	// Mock metrics
	suite.mockMetrics.On("IncrementCounter", "access_request_approved", mock.AnythingOfType("map[string]interface {}")).Return()
	suite.mockMetrics.On("IncrementCounter", "access_request_executed", mock.AnythingOfType("map[string]interface {}")).Return()

	// Act
	result, err := suite.service.ProcessAccessRequest(suite.ctx, suite.testRequestID, action, approverID)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(ApprovalStatusApproved, result.ApprovalStatus)
	suite.Equal(approverID, *result.ApprovedBy)
}

// Test ProcessAccessRequest - Reject Success
func (suite *AccessRequestServiceTestSuite) TestProcessAccessRequest_RejectSuccess() {
	// Arrange
	approverID := uuid.New()
	action := &AccessRequestApprovalRequest{
		Action:   "reject",
		Comments: stringPtr("Insufficient business justification"),
	}

	rejectedRequest := &AccessRequest{
		ID:             suite.testRequestID,
		TenantID:       suite.testTenantID,
		RequestType:    RequestTypeRoleAssignment,
		RequesterID:    suite.testUserID,
		EntityID:       suite.testEntityID,
		ApprovalStatus: ApprovalStatusRejected,
		ApprovedBy:     &approverID,
		CreatedAt:      suite.testAccessRequest.CreatedAt,
		UpdatedAt:      time.Now(),
	}

	// Mock repository calls
	suite.mockRepo.On("GetAccessRequestByID", suite.ctx, suite.testRequestID).Return(suite.testAccessRequest, nil)
	suite.mockRepo.On("UpdateAccessRequestStatus", suite.ctx, suite.testRequestID, mock.AnythingOfType("*request.UpdateAccessRequestRequest"), approverID).Return(rejectedRequest, nil)

	// Mock metrics
	suite.mockMetrics.On("IncrementCounter", "access_request_rejected", mock.AnythingOfType("map[string]interface {}")).Return()

	// Act
	result, err := suite.service.ProcessAccessRequest(suite.ctx, suite.testRequestID, action, approverID)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(ApprovalStatusRejected, result.ApprovalStatus)
}

// Test ProcessAccessRequest - Invalid Action
func (suite *AccessRequestServiceTestSuite) TestProcessAccessRequest_InvalidAction() {
	// Arrange
	approverID := uuid.New()
	action := &AccessRequestApprovalRequest{
		Action:   "invalid_action",
		Comments: stringPtr("Some comment"),
	}

	// Mock repository call
	suite.mockRepo.On("GetAccessRequestByID", suite.ctx, suite.testRequestID).Return(suite.testAccessRequest, nil)

	// Act
	result, err := suite.service.ProcessAccessRequest(suite.ctx, suite.testRequestID, action, approverID)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "INVALID_ACTION")
}

// Test GetAccessRequest - Cache Hit
func (suite *AccessRequestServiceTestSuite) TestGetAccessRequest_CacheHit() {
	// Arrange
	cachedRequest := &AccessRequestWithDetails{
		AccessRequest: suite.testAccessRequest,
		RequesterName: "Test User",
		ApproverName:  stringPtr("Approver User"),
	}

	suite.mockCache.On("Get", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("**request.AccessRequestWithDetails")).Return(nil).Run(func(args mock.Arguments) {
		dest := args[2].(**AccessRequestWithDetails)
		*dest = cachedRequest
	})
	suite.mockMetrics.On("IncrementCounter", "access_request_cache_hit", mock.AnythingOfType("map[string]interface {}")).Return()

	// Act
	result, err := suite.service.GetAccessRequest(suite.ctx, suite.testRequestID)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(suite.testRequestID, result.AccessRequest.ID)
	suite.Equal("Test User", result.RequesterName)
}

// Test GetAccessRequest - Cache Miss
func (suite *AccessRequestServiceTestSuite) TestGetAccessRequest_CacheMiss() {
	// Arrange
	requestWithDetails := &AccessRequestWithDetails{
		AccessRequest: suite.testAccessRequest,
		RequesterName: "Test User",
		ApproverName:  nil,
	}

	suite.mockCache.On("Get", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("**request.AccessRequestWithDetails")).Return(errors.New("cache miss"))
	suite.mockRepo.On("ListAccessRequestsWithDetails", suite.ctx, mock.AnythingOfType("*request.ListAccessRequestsRequest")).Return([]*AccessRequestWithDetails{requestWithDetails}, nil)
	suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"), requestWithDetails, mock.AnythingOfType("time.Duration")).Return(nil)
	suite.mockMetrics.On("IncrementCounter", "access_request_cache_miss", mock.AnythingOfType("map[string]interface {}")).Return()

	// Act
	result, err := suite.service.GetAccessRequest(suite.ctx, suite.testRequestID)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(suite.testRequestID, result.AccessRequest.ID)
}

// Test RevokeAccessRequest - Success
func (suite *AccessRequestServiceTestSuite) TestRevokeAccessRequest_Success() {
	// Arrange
	approvedRequest := &AccessRequest{
		ID:             suite.testRequestID,
		TenantID:       suite.testTenantID,
		RequestType:    RequestTypeRoleAssignment,
		RequesterID:    suite.testUserID,
		EntityID:       suite.testEntityID,
		ApprovalStatus: ApprovalStatusApproved,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	revokedRequest := &AccessRequest{
		ID:             suite.testRequestID,
		ApprovalStatus: ApprovalStatusRevoked,
		CreatedAt:      approvedRequest.CreatedAt,
		UpdatedAt:      time.Now(),
	}

	executionResult := &execution.ExecutionResult{
		Success:       true,
		AccessRevoked: []string{"role:admin"},
		ErrorMessage:  "",
	}

	// Mock repository calls
	suite.mockRepo.On("GetAccessRequestByID", suite.ctx, suite.testRequestID).Return(approvedRequest, nil)
	suite.mockExecution.On("RevokeAccessRequest", suite.ctx, mock.AnythingOfType("*execution.AccessRequest")).Return(executionResult, nil)
	suite.mockRepo.On("RevokeAccessRequest", suite.ctx, suite.testRequestID).Return(revokedRequest, nil)

	// Mock metrics
	suite.mockMetrics.On("IncrementCounter", "access_request_revoked", mock.AnythingOfType("map[string]interface {}")).Return()

	// Act
	result, err := suite.service.RevokeAccessRequest(suite.ctx, suite.testRequestID)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(ApprovalStatusRevoked, result.ApprovalStatus)
}

// Test GetAccessRequestStats - Success
func (suite *AccessRequestServiceTestSuite) TestGetAccessRequestStats_Success() {
	// Arrange
	fromDate := time.Now().AddDate(0, -1, 0) // 1 month ago
	toDate := time.Now()

	expectedStats := &AccessRequestStats{
		TotalRequests:    100,
		PendingRequests:  10,
		ApprovedRequests: 70,
		RejectedRequests: 20,
		Period:           "30d",
	}

	suite.mockRepo.On("GetAccessRequestStats", suite.ctx, &fromDate, &toDate).Return(expectedStats, nil)

	// Act
	result, err := suite.service.GetAccessRequestStats(suite.ctx, &fromDate, &toDate)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(expectedStats.TotalRequests, result.TotalRequests)
	suite.Equal(expectedStats.PendingRequests, result.PendingRequests)
	suite.Equal(expectedStats.ApprovedRequests, result.ApprovedRequests)
	suite.Equal(expectedStats.RejectedRequests, result.RejectedRequests)
}

// Test ValidateAccessRequest - Success
func (suite *AccessRequestServiceTestSuite) TestValidateAccessRequest_Success() {
	// Arrange
	roleID := uuid.New()
	req := &CreateAccessRequestRequest{
		EntityID:      suite.testEntityID,
		RequestType:   RequestTypeRoleAssignment,
		RoleID:        &roleID,
		Justification: "Need role for project access",
		DurationHours: int32Ptr(24),
	}

	// Mock user service calls
	suite.mockUserService.On("GetUserByID", suite.ctx, suite.testUserID).Return(suite.testUser, nil)

	// Act
	err := suite.service.ValidateAccessRequest(suite.ctx, req, suite.testUserID)

	// Assert
	suite.NoError(err)
}

// Test ValidateAccessRequest - User Not Found
func (suite *AccessRequestServiceTestSuite) TestValidateAccessRequest_UserNotFound() {
	// Arrange
	roleID := uuid.New()
	req := &CreateAccessRequestRequest{
		EntityID:      suite.testEntityID,
		RequestType:   RequestTypeRoleAssignment,
		RoleID:        &roleID,
		Justification: "Need role for project access",
	}

	// Mock user service to return error
	suite.mockUserService.On("GetUserByID", suite.ctx, suite.testUserID).Return((*User)(nil), errors.New("user not found"))

	// Act
	err := suite.service.ValidateAccessRequest(suite.ctx, req, suite.testUserID)

	// Assert
	suite.Error(err)
	suite.Contains(err.Error(), "REQUESTER_NOT_FOUND")
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

// TestAccessRequestService runs the test suite
func TestAccessRequestService(t *testing.T) {
	suite.Run(t, new(AccessRequestServiceTestSuite))
}
