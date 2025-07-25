//go:build unit
// +build unit

package identity

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	cache "github.com/niiniyare/erp/internal/platform/cache"
	metrics "github.com/niiniyare/erp/internal/shared/metrics"
	tracing "github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// IdentityServiceTestSuite is the test suite for the identity service
type IdentityServiceTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	service    Service
	mockRepo   *MockRepository
	mockCache  *cache.MockService
	mockTracer *tracing.MockTracingService
	mockMetric *metrics.MockMetricsService
}

func (s *IdentityServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = NewMockRepository(s.ctrl)
	s.mockCache = cache.NewMockService(s.ctrl)
	s.mockTracer = tracing.NewMockTracingService()
	s.mockMetric = metrics.NewMockMetricsService()

	// No need to set expectations for no-op mocks like tracing and metrics

	s.service = NewService(s.mockRepo, s.mockCache, s.mockTracer, s.mockMetric)
}

func (s *IdentityServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestIdentityServiceTestSuite(t *testing.T) {
	suite.Run(t, new(IdentityServiceTestSuite))
}

func (s *IdentityServiceTestSuite) TestRegisterNewUser_Success() {
	// Arrange
	req := &CreateUserRequest{
		Email:    "test.user@example.com",
		Username: "test.user",
		Password: "a-very-secure-password",
	}
	hashedPasswordMatcher := gomock.Any()
	expectedUser := &User{ID: uuid.New(), Email: req.Email, Username: req.Username}

	s.mockRepo.EXPECT().
		CreateUser(gomock.Any(), req, hashedPasswordMatcher).
		Return(expectedUser, nil)

	// Invalidate cache for all possible keys
	s.mockCache.EXPECT().Delete(gomock.Any(), fmt.Sprintf("user:id:%s", expectedUser.ID)).Return(nil).AnyTimes()
	s.mockCache.EXPECT().Delete(gomock.Any(), fmt.Sprintf("user:email:%s", expectedUser.Email)).Return(nil).AnyTimes()
	s.mockCache.EXPECT().Delete(gomock.Any(), fmt.Sprintf("user:username:%s", expectedUser.Username)).Return(nil).AnyTimes()

	// Act
	user, err := s.service.RegisterNewUser(context.Background(), req)

	// Assert
	require.NoError(s.T(), err)
	require.NotNil(s.T(), user)
	require.Equal(s.T(), expectedUser.Email, user.Email)
}

func (s *IdentityServiceTestSuite) TestGetUserByID_Success_CacheMiss() {
	// Arrange
	userID := uuid.New()
	expectedUser := &User{ID: userID, Email: "cache.miss@example.com"}
	cacheKey := fmt.Sprintf("user:id:%s", userID)

	// Mock cache miss
	s.mockCache.EXPECT().Get(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss"))

	// Mock repository call
	s.mockRepo.EXPECT().GetUserByID(gomock.Any(), userID).Return(expectedUser, nil)

	// Mock cache set
	s.mockCache.EXPECT().Set(gomock.Any(), cacheKey, expectedUser, 30*time.Minute).Return(nil)
	s.mockCache.EXPECT().Set(gomock.Any(), fmt.Sprintf("user:email:%s", expectedUser.Email), expectedUser, 30*time.Minute).Return(nil)

	// Act
	user, err := s.service.GetUserByID(context.Background(), userID)

	// Assert
	require.NoError(s.T(), err)
	require.Equal(s.T(), expectedUser, user)
}

func (s *IdentityServiceTestSuite) TestGetUserByID_Success_CacheHit() {
	// Arrange
	userID := uuid.New()
	expectedUser := &User{ID: userID, Email: "cache.hit@example.com"}
	cacheKey := fmt.Sprintf("user:id:%s", userID)

	// Mock cache hit
	s.mockCache.EXPECT().Get(gomock.Any(), cacheKey, gomock.Any()).DoAndReturn(func(_ context.Context, _ string, dest interface{}) error {
		destPtr := dest.(*User)
		*destPtr = *expectedUser
		return nil
	})

	// Act
	user, err := s.service.GetUserByID(context.Background(), userID)

	// Assert
	require.NoError(s.T(), err)
	require.Equal(s.T(), expectedUser, user)
	// s.mockRepo should have no expectations set, so Finish() in TearDownTest will fail if it's called.
}

func (s *IdentityServiceTestSuite) TestGetUserByEmail_Success_CacheMiss() {
	// Arrange
	email := "email.miss@example.com"
	expectedUser := &User{ID: uuid.New(), Email: email}
	cacheKey := fmt.Sprintf("user:email:%s", email)

	s.mockCache.EXPECT().Get(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss"))
	s.mockRepo.EXPECT().GetUserByEmail(gomock.Any(), email).Return(expectedUser, nil)
	s.mockCache.EXPECT().Set(gomock.Any(), cacheKey, expectedUser, 30*time.Minute).Return(nil)
	s.mockCache.EXPECT().Set(gomock.Any(), fmt.Sprintf("user:id:%s", expectedUser.ID), expectedUser, 30*time.Minute).Return(nil)

	// Act
	user, err := s.service.GetUserByEmail(context.Background(), email)

	// Assert
	require.NoError(s.T(), err)
	require.Equal(s.T(), expectedUser, user)
}

func (s *IdentityServiceTestSuite) TestGetUserByUsername_Success_CacheMiss() {
	// Arrange
	username := "username.miss"
	expectedUser := &User{ID: uuid.New(), Username: username, Email: "user@name.com"}
	cacheKey := fmt.Sprintf("user:username:%s", username)

	s.mockCache.EXPECT().Get(gomock.Any(), cacheKey, gomock.Any()).Return(errors.New("cache miss"))
	s.mockRepo.EXPECT().GetUserByUsername(gomock.Any(), username).Return(expectedUser, nil)
	s.mockCache.EXPECT().Set(gomock.Any(), cacheKey, expectedUser, 30*time.Minute).Return(nil)
	s.mockCache.EXPECT().Set(gomock.Any(), fmt.Sprintf("user:id:%s", expectedUser.ID), expectedUser, 30*time.Minute).Return(nil)
	s.mockCache.EXPECT().Set(gomock.Any(), fmt.Sprintf("user:email:%s", expectedUser.Email), expectedUser, 30*time.Minute).Return(nil)

	// Act
	user, err := s.service.GetUserByUsername(context.Background(), username)

	// Assert
	require.NoError(s.T(), err)
	require.Equal(s.T(), expectedUser, user)
}
