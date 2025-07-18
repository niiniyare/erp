package user

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockRepository implements the Repository interface for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*User, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) DeleteUser(ctx context.Context, id uuid.UUID, permanent bool) error {
	args := m.Called(ctx, id, permanent)
	return args.Error(0)
}

func (m *MockRepository) RestoreUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ListUsers(ctx context.Context, req *ListUsersRequest) ([]*User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]*User), args.Error(1)
}

func (m *MockRepository) GetUserWithDetails(ctx context.Context, id uuid.UUID) (*UserWithDetails, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserWithDetails), args.Error(1)
}

func (m *MockRepository) CreatePerson(ctx context.Context, req *CreatePersonRequest) (*Person, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*Person), args.Error(1)
}

func (m *MockRepository) GetPersonByID(ctx context.Context, id uuid.UUID) (*Person, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Person), args.Error(1)
}

func (m *MockRepository) GetPersonByEmail(ctx context.Context, email string) (*Person, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Person), args.Error(1)
}

func (m *MockRepository) UpdatePerson(ctx context.Context, id uuid.UUID, req *CreatePersonRequest) (*Person, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*Person), args.Error(1)
}

func (m *MockRepository) DeletePerson(ctx context.Context, id uuid.UUID, permanent bool) error {
	args := m.Called(ctx, id, permanent)
	return args.Error(0)
}

func (m *MockRepository) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*Employee, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*Employee), args.Error(1)
}

func (m *MockRepository) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*Employee, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Employee), args.Error(1)
}

func (m *MockRepository) GetEmployeeByPersonID(ctx context.Context, personID uuid.UUID) (*Employee, error) {
	args := m.Called(ctx, personID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Employee), args.Error(1)
}

func (m *MockRepository) GetEmployeeByNumber(ctx context.Context, number string) (*Employee, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Employee), args.Error(1)
}

func (m *MockRepository) UpdateEmployee(ctx context.Context, id uuid.UUID, req *CreateEmployeeRequest) (*Employee, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0).(*Employee), args.Error(1)
}

func (m *MockRepository) DeleteEmployee(ctx context.Context, id uuid.UUID, permanent bool) error {
	args := m.Called(ctx, id, permanent)
	return args.Error(0)
}

func (m *MockRepository) AuthenticateUser(ctx context.Context, identifier, password string) (*User, error) {
	args := m.Called(ctx, identifier, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	args := m.Called(ctx, userID, newPassword)
	return args.Error(0)
}

func (m *MockRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRepository) IncrementFailedLogins(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRepository) UnlockUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRole, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*UserRole), args.Error(1)
}

func (m *MockRepository) AssignUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	args := m.Called(ctx, userID, roleID, entityID)
	return args.Error(0)
}

func (m *MockRepository) RevokeUserRole(ctx context.Context, userID, roleID, entityID uuid.UUID) error {
	args := m.Called(ctx, userID, roleID, entityID)
	return args.Error(0)
}

func (m *MockRepository) ValidateUserEmail(ctx context.Context, email string, excludeID *uuid.UUID) error {
	args := m.Called(ctx, email, excludeID)
	return args.Error(0)
}

func (m *MockRepository) ValidateUsername(ctx context.Context, username string, excludeID *uuid.UUID) error {
	args := m.Called(ctx, username, excludeID)
	return args.Error(0)
}

func (m *MockRepository) ValidateEmployeeNumber(ctx context.Context, number string, excludeID *uuid.UUID) error {
	args := m.Called(ctx, number, excludeID)
	return args.Error(0)
}

func (m *MockRepository) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*User, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]*User), args.Error(1)
}

// Permission evaluation methods
func (m *MockRepository) GetRoleHierarchy(ctx context.Context, roleID uuid.UUID) ([]*Role, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).([]*Role), args.Error(1)
}

func (m *MockRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*Permission, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).([]*Permission), args.Error(1)
}

func (m *MockRepository) GetApplicablePolicies(ctx context.Context, resourceName, actionName string) ([]*Policy, error) {
	args := m.Called(ctx, resourceName, actionName)
	return args.Get(0).([]*Policy), args.Error(1)
}

func (m *MockRepository) GetPolicyByID(ctx context.Context, policyID uuid.UUID) (*Policy, error) {
	args := m.Called(ctx, policyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Policy), args.Error(1)
}

// MockCache implements cache.Service for testing
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

func (m *MockCache) Clear(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCache) Flush(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Test Suite
type UserServiceTestSuite struct {
	suite.Suite
	service      Service
	mockRepo     *MockRepository
	mockCache    *MockCache
	ctx          context.Context
	testUserID   uuid.UUID
	testEntityID uuid.UUID
	testRoleID   uuid.UUID
}

func (suite *UserServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockRepository)
	suite.mockCache = new(MockCache)
	suite.ctx = context.Background()
	suite.testUserID = uuid.New()
	suite.testEntityID = uuid.New()
	suite.testRoleID = uuid.New()

	// Create mock tracing and metrics services
	tracingService := &tracing.TracingService{}
	metricsService := &metrics.MetricsService{}

	suite.service = NewService(suite.mockRepo, suite.mockCache, tracingService, metricsService)
}

func (suite *UserServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
}

// ===== PERMISSION EVALUATION TESTS =====

func (suite *UserServiceTestSuite) TestEvaluatePermission_Success() {
	// Arrange
	req := &PermissionEvaluationRequest{
		UserID:       suite.testUserID,
		ResourceName: "user_management",
		ActionName:   "read",
		EntityID:     &suite.testEntityID,
		Context:      map[string]any{"department": "hr"},
	}

	userRoles := []*UserRole{
		{
			ID:             uuid.New(),
			UserID:         suite.testUserID,
			RoleID:         suite.testRoleID,
			EntityID:       suite.testEntityID,
			AssignmentType: "DIRECT",
			IsActive:       true,
		},
	}

	roles := []*Role{
		{
			ID:       suite.testRoleID,
			Name:     "hr_manager",
			IsActive: true,
		},
	}

	permissions := []*Permission{
		{
			ID:       uuid.New(),
			Name:     "user_management:read",
			Effect:   "ALLOW",
			IsActive: true,
		},
	}

	policies := []*Policy{
		{
			ID:       uuid.New(),
			Name:     "business_hours_policy",
			Effect:   "ALLOW",
			IsActive: true,
		},
	}

	userDetails := &UserWithDetails{
		User: User{
			ID:       suite.testUserID,
			UserType: "INTERNAL",
			Email:    "test@example.com",
		},
	}

	// Mock cache miss
	// suite.mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(cache.ErrCacheMiss)

	// Mock repository calls
	suite.mockRepo.On("GetUserRoles", suite.ctx, suite.testUserID).Return(userRoles, nil)
	suite.mockRepo.On("GetRoleHierarchy", suite.ctx, suite.testRoleID).Return(roles, nil)
	suite.mockRepo.On("GetRolePermissions", suite.ctx, suite.testRoleID).Return(permissions, nil)
	suite.mockRepo.On("GetApplicablePolicies", suite.ctx, "user_management", "read").Return(policies, nil)
	suite.mockRepo.On("GetUserWithDetails", suite.ctx, suite.testUserID).Return(userDetails, nil)

	// Mock cache set
	suite.mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)

	// Act
	result, err := suite.service.EvaluatePermission(suite.ctx, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.Allowed)
	assert.False(suite.T(), result.CacheHit)
	assert.Greater(suite.T(), result.EvaluationTimeMS, 0)
	assert.Contains(suite.T(), result.EffectiveRoles, "hr_manager")
}

func (suite *UserServiceTestSuite) TestEvaluatePermission_CacheHit() {
	// Arrange
	req := &PermissionEvaluationRequest{
		UserID:       suite.testUserID,
		ResourceName: "user_management",
		ActionName:   "read",
	}

	cachedResult := PermissionEvaluationResult{
		Allowed:          true,
		PolicyDecisions:  []string{"cached_decision"},
		EffectiveRoles:   []string{"cached_role"},
		EvaluationTimeMS: 0,
		CacheHit:         false,
	}

	// Mock cache hit
	suite.mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
		result := args.Get(2).(*PermissionEvaluationResult)
		*result = cachedResult
	}).Return(nil)

	// Act
	result, err := suite.service.EvaluatePermission(suite.ctx, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.Allowed)
	assert.True(suite.T(), result.CacheHit)
	assert.Contains(suite.T(), result.PolicyDecisions, "cached_decision")
}

func (suite *UserServiceTestSuite) TestEvaluatePermission_Denied() {
	// Arrange
	req := &PermissionEvaluationRequest{
		UserID:       suite.testUserID,
		ResourceName: "admin_panel",
		ActionName:   "access",
	}

	userRoles := []*UserRole{
		{
			ID:             uuid.New(),
			UserID:         suite.testUserID,
			RoleID:         suite.testRoleID,
			EntityID:       suite.testEntityID,
			AssignmentType: "DIRECT",
			IsActive:       true,
		},
	}

	roles := []*Role{
		{
			ID:       suite.testRoleID,
			Name:     "basic_user",
			IsActive: true,
		},
	}

	permissions := []*Permission{} // No permissions for admin panel

	policies := []*Policy{}

	userDetails := &UserWithDetails{
		User: User{
			ID:       suite.testUserID,
			UserType: "INTERNAL",
			Email:    "test@example.com",
		},
	}

	// Mock cache miss
	suite.mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(cache.ErrCacheMiss)

	// Mock repository calls
	suite.mockRepo.On("GetUserRoles", suite.ctx, suite.testUserID).Return(userRoles, nil)
	suite.mockRepo.On("GetRoleHierarchy", suite.ctx, suite.testRoleID).Return(roles, nil)
	suite.mockRepo.On("GetRolePermissions", suite.ctx, suite.testRoleID).Return(permissions, nil)
	suite.mockRepo.On("GetApplicablePolicies", suite.ctx, "admin_panel", "access").Return(policies, nil)
	suite.mockRepo.On("GetUserWithDetails", suite.ctx, suite.testUserID).Return(userDetails, nil)

	// Mock cache set
	suite.mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)

	// Act
	result, err := suite.service.EvaluatePermission(suite.ctx, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.False(suite.T(), result.Allowed)
	assert.Contains(suite.T(), result.PolicyDecisions, "rbac_no_matching_permissions")
}

func (suite *UserServiceTestSuite) TestGetUserEffectivePermissions_Success() {
	// Arrange
	userRoles := []*UserRole{
		{
			ID:             uuid.New(),
			UserID:         suite.testUserID,
			RoleID:         suite.testRoleID,
			EntityID:       suite.testEntityID,
			AssignmentType: "DIRECT",
			IsActive:       true,
		},
	}

	roles := []*Role{
		{
			ID:       suite.testRoleID,
			Name:     "manager",
			IsActive: true,
		},
	}

	permissions := []*Permission{
		{
			ID:         uuid.New(),
			ResourceID: uuid.New(),
			ActionID:   uuid.New(),
			Name:       "users:read",
			Effect:     "ALLOW",
			IsActive:   true,
		},
		{
			ID:         uuid.New(),
			ResourceID: uuid.New(),
			ActionID:   uuid.New(),
			Name:       "users:write",
			Effect:     "ALLOW",
			IsActive:   true,
		},
	}

	// Mock repository calls
	suite.mockRepo.On("GetUserRoles", suite.ctx, suite.testUserID).Return(userRoles, nil)
	suite.mockRepo.On("GetRoleHierarchy", suite.ctx, suite.testRoleID).Return(roles, nil)
	suite.mockRepo.On("GetRolePermissions", suite.ctx, suite.testRoleID).Return(permissions, nil)

	// Act
	result, err := suite.service.GetUserEffectivePermissions(suite.ctx, suite.testUserID, nil)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), "users:read", result[0].Permission.Name)
	assert.Equal(suite.T(), "users:write", result[1].Permission.Name)
	assert.Equal(suite.T(), "manager", result[0].GrantedByRole.Name)
}

func (suite *UserServiceTestSuite) TestBulkEvaluatePermissions_Success() {
	// Arrange
	requests := []*PermissionEvaluationRequest{
		{
			UserID:       suite.testUserID,
			ResourceName: "users",
			ActionName:   "read",
		},
		{
			UserID:       suite.testUserID,
			ResourceName: "users",
			ActionName:   "write",
		},
	}

	req := &BulkPermissionEvaluationRequest{
		Requests: requests,
	}

	userRoles := []*UserRole{
		{
			ID:             uuid.New(),
			UserID:         suite.testUserID,
			RoleID:         suite.testRoleID,
			EntityID:       suite.testEntityID,
			AssignmentType: "DIRECT",
			IsActive:       true,
		},
	}

	roles := []*Role{
		{
			ID:       suite.testRoleID,
			Name:     "user_admin",
			IsActive: true,
		},
	}

	permissions := []*Permission{
		{
			ID:       uuid.New(),
			Name:     "users:read",
			Effect:   "ALLOW",
			IsActive: true,
		},
		{
			ID:       uuid.New(),
			Name:     "users:write",
			Effect:   "ALLOW",
			IsActive: true,
		},
	}

	policies := []*Policy{}

	userDetails := &UserWithDetails{
		User: User{
			ID:       suite.testUserID,
			UserType: "INTERNAL",
			Email:    "test@example.com",
		},
	}

	// Mock cache misses for both requests
	suite.mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(cache.ErrCacheMiss).Times(2)

	// Mock repository calls for both requests
	suite.mockRepo.On("GetUserRoles", suite.ctx, suite.testUserID).Return(userRoles, nil).Times(2)
	suite.mockRepo.On("GetRoleHierarchy", suite.ctx, suite.testRoleID).Return(roles, nil).Times(2)
	suite.mockRepo.On("GetRolePermissions", suite.ctx, suite.testRoleID).Return(permissions, nil).Times(2)
	suite.mockRepo.On("GetApplicablePolicies", suite.ctx, "users", "read").Return(policies, nil).Once()
	suite.mockRepo.On("GetApplicablePolicies", suite.ctx, "users", "write").Return(policies, nil).Once()
	suite.mockRepo.On("GetUserWithDetails", suite.ctx, suite.testUserID).Return(userDetails, nil).Times(2)

	// Mock cache sets
	suite.mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil).Times(2)

	// Act
	results, err := suite.service.BulkEvaluatePermissions(suite.ctx, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), results)
	assert.Len(suite.T(), results, 2)
	assert.True(suite.T(), results[0].Allowed)
	assert.True(suite.T(), results[1].Allowed)
}

func (suite *UserServiceTestSuite) TestCalculateRoleHierarchy_Success() {
	// Arrange
	parentRoleID := uuid.New()
	roles := []*Role{
		{
			ID:           suite.testRoleID,
			Name:         "manager",
			ParentRoleID: &parentRoleID,
			IsActive:     true,
		},
		{
			ID:           parentRoleID,
			Name:         "employee",
			ParentRoleID: nil,
			IsActive:     true,
		},
	}

	// Mock cache miss
	suite.mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(cache.ErrCacheMiss)

	// Mock repository calls
	suite.mockRepo.On("GetRoleHierarchy", suite.ctx, suite.testRoleID).Return(roles, nil)

	// Mock cache set
	suite.mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil)

	// Act
	result, err := suite.service.CalculateRoleHierarchy(suite.ctx, suite.testRoleID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), "manager", result[0].Name)
	assert.Equal(suite.T(), "employee", result[1].Name)
}

func (suite *UserServiceTestSuite) TestTestPolicy_Success() {
	// Arrange
	policyID := uuid.New()
	req := &PolicyTestRequest{
		UserID:       suite.testUserID,
		ResourceName: "documents",
		ActionName:   "read",
		EntityID:     &suite.testEntityID,
		Context:      map[string]any{"department": "finance"},
	}

	policy := &Policy{
		ID:     policyID,
		Name:   "finance_documents_policy",
		Effect: "ALLOW",
		Target: map[string]any{
			"department": "finance",
		},
		Rule: map[string]any{
			"and": []map[string]any{
				{"user.department": "finance"},
			},
		},
	}

	userDetails := &UserWithDetails{
		User: User{
			ID:       suite.testUserID,
			UserType: "INTERNAL",
			Email:    "test@example.com",
		},
	}

	// Mock repository calls
	suite.mockRepo.On("GetPolicyByID", suite.ctx, policyID).Return(policy, nil)
	suite.mockRepo.On("GetUserWithDetails", suite.ctx, suite.testUserID).Return(userDetails, nil)

	// Act
	result, err := suite.service.TestPolicy(suite.ctx, policyID, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), policyID, result.PolicyID)
	assert.Equal(suite.T(), "finance_documents_policy", result.PolicyName)
	assert.Equal(suite.T(), "ALLOW", result.Effect)
	// Note: target matching and rule evaluation are simplified in this implementation
}

// ===== USER CRUD TESTS =====

func (suite *UserServiceTestSuite) TestCreateUser_Success() {
	// Arrange
	req := &CreateUserRequest{
		EntityID: suite.testEntityID,
		Username: "testuser",
		Email:    "test@example.com",
		Password: "testpassword123",
		UserType: "INTERNAL",
	}

	expectedUser := &User{
		ID:       suite.testUserID,
		EntityID: suite.testEntityID,
		Username: "testuser",
		Email:    "test@example.com",
		UserType: "INTERNAL",
		IsActive: true,
	}

	// Mock validation
	suite.mockRepo.On("ValidateUserEmail", suite.ctx, "test@example.com", mock.AnythingOfType("*uuid.UUID")).Return(nil)
	suite.mockRepo.On("ValidateUsername", suite.ctx, "testuser", mock.AnythingOfType("*uuid.UUID")).Return(nil)

	// Mock repository call
	suite.mockRepo.On("CreateUser", suite.ctx, req).Return(expectedUser, nil)

	// Mock cache set
	suite.mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil).Times(3) // ID, email, username

	// Act
	result, err := suite.service.CreateUser(suite.ctx, req)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedUser.ID, result.ID)
	assert.Equal(suite.T(), expectedUser.Email, result.Email)
	assert.Equal(suite.T(), expectedUser.Username, result.Username)
}

func (suite *UserServiceTestSuite) TestGetUserByID_CacheHit() {
	// Arrange
	expectedUser := &User{
		ID:       suite.testUserID,
		Email:    "test@example.com",
		Username: "testuser",
		UserType: "INTERNAL",
	}

	// Mock cache hit
	suite.mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Run(func(args mock.Arguments) {
		user := args.Get(2).(*User)
		*user = *expectedUser
	}).Return(nil)

	// Act
	result, err := suite.service.GetUserByID(suite.ctx, suite.testUserID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedUser.ID, result.ID)
	assert.Equal(suite.T(), expectedUser.Email, result.Email)
}

func (suite *UserServiceTestSuite) TestGetUserByID_CacheMiss() {
	// Arrange
	expectedUser := &User{
		ID:       suite.testUserID,
		Email:    "test@example.com",
		Username: "testuser",
		UserType: "INTERNAL",
	}

	// Mock cache miss
	suite.mockCache.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(cache.ErrCacheMiss)

	// Mock repository call
	suite.mockRepo.On("GetUserByID", suite.ctx, suite.testUserID).Return(expectedUser, nil)

	// Mock cache set
	suite.mockCache.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil).Times(3) // ID, email, username

	// Act
	result, err := suite.service.GetUserByID(suite.ctx, suite.testUserID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedUser.ID, result.ID)
	assert.Equal(suite.T(), expectedUser.Email, result.Email)
}

// Run the test suite
func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

// ===== HELPER FUNCTION TESTS =====

func TestIsValidUserType(t *testing.T) {
	tests := []struct {
		name     string
		userType string
		expected bool
	}{
		{"Valid ADMIN", "ADMIN", true},
		{"Valid INTERNAL", "INTERNAL", true},
		{"Valid CUSTOMER", "CUSTOMER", true},
		{"Valid VENDOR", "VENDOR", true},
		{"Invalid type", "INVALID", false},
		{"Empty string", "", false},
		{"Lowercase", "admin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidUserType(tt.userType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidAccountStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"Valid ACTIVE", "ACTIVE", true},
		{"Valid INACTIVE", "INACTIVE", true},
		{"Valid LOCKED", "LOCKED", true},
		{"Valid SUSPENDED", "SUSPENDED", true},
		{"Invalid status", "INVALID", false},
		{"Empty string", "", false},
		{"Lowercase", "active", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidAccountStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPersonGetFullName(t *testing.T) {
	tests := []struct {
		name     string
		person   Person
		expected string
	}{
		{
			name: "With middle name",
			person: Person{
				FirstName:  "John",
				MiddleName: stringPtr("William"),
				LastName:   "Doe",
			},
			expected: "John William Doe",
		},
		{
			name: "Without middle name",
			person: Person{
				FirstName: "John",
				LastName:  "Doe",
			},
			expected: "John Doe",
		},
		{
			name: "With empty middle name",
			person: Person{
				FirstName:  "John",
				MiddleName: stringPtr(""),
				LastName:   "Doe",
			},
			expected: "John Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.person.GetFullName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

