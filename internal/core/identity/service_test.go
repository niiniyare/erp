package identity

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
	"golang.org/x/crypto/bcrypt"
)

// MockRepository implements the Repository interface for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, req *CreateUserRequest, hashedPassword string) (*User, error) {
	args := m.Called(ctx, req, hashedPassword)
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

func (m *MockRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) GetUserWithDetails(ctx context.Context, id uuid.UUID) (*UserWithDetails, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserWithDetails), args.Error(1)
}

func (m *MockRepository) GetUserPassword(ctx context.Context, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func (m *MockRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, newPasswordHash string) error {
	args := m.Called(ctx, userID, newPasswordHash)
	return args.Error(0)
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

func (m *MockCache) DeletePattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCache) MGet(ctx context.Context, keys []string, dest interface{}) error {
	args := m.Called(ctx, keys, dest)
	return args.Error(0)
}

func (m *MockCache) MDelete(ctx context.Context, keys []string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockCache) Flush(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCache) MSet(ctx context.Context, pairs map[string]interface{}, expiration time.Duration) error {
	args := m.Called(ctx, pairs, expiration)
	return args.Error(0)
}

func (m *MockCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func (m *MockCache) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCache) Stats() *cache.CacheStats {
	args := m.Called()
	return args.Get(0).(*cache.CacheStats)
}

func (m *MockCache) Close() error {
	args := m.Called()
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
	tracingService := tracing.NewMockTracingService()
	metricsService := metrics.NewMockMetricsService()

	suite.service = NewService(suite.mockRepo, suite.mockCache, tracingService, metricsService)
}

func (suite *UserServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockCache.AssertExpectations(suite.T())
}

// ===== USER CRUD TESTS =====

func (suite *UserServiceTestSuite) TestRegisterNewUser_Success() {
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

	// Mock repository call
	suite.mockRepo.On("CreateUser", suite.ctx, req, mock.AnythingOfType("string")).Return(expectedUser, nil)

	// Mock cache invalidation
	suite.mockCache.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Times(3)

	// Act
	result, err := suite.service.RegisterNewUser(suite.ctx, req)

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

func (suite *UserServiceTestSuite) TestUpdateUser_Success() {
	// Arrange
	updateReq := &UpdateUserRequest{
		Username: stringPtr("newusername"),
	}

	originalUser := &User{
		ID:       suite.testUserID,
		Email:    "test@example.com",
		Username: "testuser",
	}

	updatedUser := &User{
		ID:       suite.testUserID,
		Email:    "test@example.com",
		Username: "newusername",
	}

	// Mock repository calls
	suite.mockRepo.On("GetUserByID", suite.ctx, suite.testUserID).Return(originalUser, nil)
	suite.mockRepo.On("UpdateUser", suite.ctx, suite.testUserID, updateReq).Return(updatedUser, nil)

	// Mock cache operations
	suite.mockCache.On("Delete", suite.ctx, mock.AnythingOfType("string")).Return(nil).Times(3)
	suite.mockCache.On("Set", suite.ctx, mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("time.Duration")).Return(nil).Times(3)

	// Act
	result, err := suite.service.UpdateUser(suite.ctx, suite.testUserID, updateReq)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "newusername", result.Username)
}

func (suite *UserServiceTestSuite) TestAuthenticate_Success() {
	// Arrange
	identifier := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &User{
		ID:    suite.testUserID,
		Email: identifier,
	}

	// Mock repository calls
	suite.mockRepo.On("GetUserByEmail", suite.ctx, identifier).Return(user, nil)
	suite.mockRepo.On("GetUserPassword", suite.ctx, suite.testUserID).Return(string(hashedPassword), nil)

	// Act
	result, err := suite.service.Authenticate(suite.ctx, identifier, password)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), suite.testUserID, result.ID)
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
