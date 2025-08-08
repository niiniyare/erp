package notification

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// MockEmailService is a mock implementation of EmailService
type MockEmailService struct{}

func (m *MockEmailService) SendEmail(ctx context.Context, to []string, subject, body string, data map[string]any) error {
	return nil
}

// MockSlackService is a mock implementation of SlackService
type MockSlackService struct{}

func (m *MockSlackService) SendSlackMessage(ctx context.Context, userIDs []string, message string, data map[string]any) error {
	return nil
}

// NotificationServiceTestSuite defines the test suite
type NotificationServiceTestSuite struct {
	suite.Suite
	ctrl        *gomock.Controller
	mockTracing *tracing.MockTracingService
	mockMetrics metrics.MetricsProvider
	mockEmail   *MockEmailService
	mockSlack   *MockSlackService
	mockRepo    *MockRepository
	service     NotificationService
}

// SetupTest runs before each test
func (suite *NotificationServiceTestSuite) SetupTest() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.mockTracing = tracing.NewMockTracingService(suite.ctrl)

	var err error
	suite.mockMetrics, err = metrics.NewMetricsService(metrics.MetricsConfig{Enabled: false})
	require.NoError(suite.T(), err)

	suite.mockEmail = &MockEmailService{}
	suite.mockSlack = &MockSlackService{}
	suite.mockRepo = NewMockRepository(suite.ctrl)

	suite.service = NewNotificationService(
		suite.mockTracing,
		suite.mockMetrics,
		suite.mockEmail,
		suite.mockSlack,
		suite.mockRepo,
	)
}

// TearDownTest runs after each test
func (suite *NotificationServiceTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

// We'll use the MockSpan from the tracing package

// TestSend tests the Send method with various scenarios
func (suite *NotificationServiceTestSuite) TestSend() {
	userID := uuid.New()
	tenantID := uuid.New()

	testCases := []struct {
		name            string
		notification    *Notification
		mockPreferences *NotificationPreferences
		mockPrefsError  error
		expectedError   string
		expectRepoCall  bool
	}{
		{
			name: "successful email notification",
			notification: &Notification{
				TenantID: tenantID,
				Recipients: []Recipient{{
					UserID:  userID,
					Email:   "test@example.com",
					SlackID: "U123456789",
				}},
				Subject:  "Test Subject",
				Message:  "Test Message",
				Type:     NotificationTypeGeneric,
				Channels: []NotificationChannel{NotificationChannelEmail},
			},
			mockPreferences: &NotificationPreferences{
				UserID:             userID,
				EmailNotifications: true,
				SlackNotifications: true,
				PreferredChannels:  []NotificationChannel{NotificationChannelEmail, NotificationChannelSlack},
				NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: true},
				QuietHours:         nil,
			},
			mockPrefsError: nil,
			expectedError:  "",
			expectRepoCall: true,
		},
		{
			name: "notification with quiet hours enabled",
			notification: &Notification{
				TenantID: tenantID,
				Recipients: []Recipient{{
					UserID: userID,
					Email:  "test@example.com",
				}},
				Subject:  "Test Subject",
				Message:  "Test Message",
				Type:     NotificationTypeGeneric,
				Channels: []NotificationChannel{NotificationChannelEmail},
			},
			mockPreferences: &NotificationPreferences{
				UserID:             userID,
				EmailNotifications: true,
				SlackNotifications: true,
				PreferredChannels:  []NotificationChannel{NotificationChannelEmail},
				NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: true},
				QuietHours: &QuietHours{
					Enabled:   true,
					StartTime: "22:00",
					EndTime:   "08:00",
				},
			},
			mockPrefsError: nil,
			expectedError:  "",
			expectRepoCall: true,
		},
		{
			name: "notification type disabled by user",
			notification: &Notification{
				TenantID: tenantID,
				Recipients: []Recipient{{
					UserID: userID,
					Email:  "test@example.com",
				}},
				Subject:  "Test Subject",
				Message:  "Test Message",
				Type:     NotificationTypeGeneric,
				Channels: []NotificationChannel{NotificationChannelEmail},
			},
			mockPreferences: &NotificationPreferences{
				UserID:             userID,
				EmailNotifications: true,
				SlackNotifications: true,
				PreferredChannels:  []NotificationChannel{NotificationChannelEmail},
				NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: false},
				QuietHours:         nil,
			},
			mockPrefsError: nil,
			expectedError:  "",
			expectRepoCall: true,
		},
		{
			name: "no recipients",
			notification: &Notification{
				TenantID:   tenantID,
				Recipients: []Recipient{},
				Subject:    "Test Subject",
				Message:    "Test Message",
				Type:       NotificationTypeGeneric,
				Channels:   []NotificationChannel{NotificationChannelEmail},
			},
			mockPreferences: nil,
			mockPrefsError:  nil,
			expectedError:   "",
			expectRepoCall:  false,
		},
		{
			name: "repository error when getting preferences",
			notification: &Notification{
				TenantID: tenantID,
				Recipients: []Recipient{{
					UserID: userID,
					Email:  "test@example.com",
				}},
				Subject:  "Test Subject",
				Message:  "Test Message",
				Type:     NotificationTypeGeneric,
				Channels: []NotificationChannel{NotificationChannelEmail},
			},
			mockPreferences: nil,
			mockPrefsError:  errors.New("database error"),
			expectedError:   "",
			expectRepoCall:  true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			mockSpan := tracing.NewMockSpan(suite.ctrl)
			suite.mockTracing.EXPECT().StartSpan(gomock.Any(), "notificationService.Send").Return(context.Background(), mockSpan)
			mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes() // Allow any attributes to be set
			mockSpan.EXPECT().End().Times(1)

			if tc.expectRepoCall {
				suite.mockRepo.EXPECT().
					GetUserNotificationPreferences(gomock.Any(), userID).
					Return(tc.mockPreferences, tc.mockPrefsError).
					AnyTimes()
			}

			err := suite.service.Send(context.Background(), tc.notification)

			if tc.expectedError != "" {
				require.Error(suite.T(), err)
				require.Contains(suite.T(), err.Error(), tc.expectedError)
			} else {
				require.NoError(suite.T(), err)
			}
		})
	}
}

// TestGetUserNotificationPreferences tests getting user preferences
func (suite *NotificationServiceTestSuite) TestGetUserNotificationPreferences() {
	userID := uuid.New()

	testCases := []struct {
		name            string
		userID          uuid.UUID
		mockPreferences *NotificationPreferences
		mockError       error
		expectedError   string
	}{
		{
			name:   "successful retrieval",
			userID: userID,
			mockPreferences: &NotificationPreferences{
				UserID:             userID,
				EmailNotifications: true,
				SlackNotifications: false,
				PreferredChannels:  []NotificationChannel{NotificationChannelEmail},
				NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: true},
				QuietHours:         nil,
			},
			mockError:     nil,
			expectedError: "",
		},
		{
			name:            "repository error",
			userID:          userID,
			mockPreferences: nil,
			mockError:       errors.New("database connection failed"),
			expectedError:   "database connection failed",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.mockRepo.EXPECT().
				GetUserNotificationPreferences(gomock.Any(), tc.userID).
				Return(tc.mockPreferences, tc.mockError).
				Times(1)

			prefs, err := suite.service.GetUserNotificationPreferences(context.Background(), tc.userID)

			if tc.expectedError != "" {
				require.Error(suite.T(), err)
				require.Contains(suite.T(), err.Error(), tc.expectedError)
				require.Nil(suite.T(), prefs)
			} else {
				require.NoError(suite.T(), err)
				require.Equal(suite.T(), tc.mockPreferences, prefs)
			}
		})
	}
}

// TestUpdateUserNotificationPreferences tests updating user preferences
func (suite *NotificationServiceTestSuite) TestUpdateUserNotificationPreferences() {
	userID := uuid.New()

	testCases := []struct {
		name          string
		userID        uuid.UUID
		preferences   *NotificationPreferences
		mockError     error
		expectedError string
	}{
		{
			name:   "successful update",
			userID: userID,
			preferences: &NotificationPreferences{
				UserID:             userID,
				EmailNotifications: false,
				SlackNotifications: true,
				PreferredChannels:  []NotificationChannel{NotificationChannelSlack},
				NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: false},
				QuietHours:         nil,
			},
			mockError:     nil,
			expectedError: "",
		},
		{
			name:   "repository error during update",
			userID: userID,
			preferences: &NotificationPreferences{
				UserID:             userID,
				EmailNotifications: true,
				SlackNotifications: true,
				PreferredChannels:  []NotificationChannel{NotificationChannelEmail},
				NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: true},
				QuietHours:         nil,
			},
			mockError:     errors.New("update failed"),
			expectedError: "update failed",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.mockRepo.EXPECT().
				UpdateUserNotificationPreferences(gomock.Any(), tc.userID, tc.preferences).
				Return(tc.mockError).
				Times(1)

			err := suite.service.UpdateUserNotificationPreferences(context.Background(), tc.userID, tc.preferences)

			if tc.expectedError != "" {
				require.Error(suite.T(), err)
				require.Contains(suite.T(), err.Error(), tc.expectedError)
			} else {
				require.NoError(suite.T(), err)
			}
		})
	}
}

// TestNewNotificationService tests service instantiation
func (suite *NotificationServiceTestSuite) TestNewNotificationService() {
	testCases := []struct {
		name         string
		tracing      tracing.TracingService
		metrics      metrics.MetricsProvider
		emailService EmailService
		slackService SlackService
		repo         Repository
		expectNil    bool
	}{
		{
			name:         "valid dependencies",
			tracing:      suite.mockTracing,
			metrics:      suite.mockMetrics,
			emailService: suite.mockEmail,
			slackService: suite.mockSlack,
			repo:         suite.mockRepo,
			expectNil:    false,
		},
		{
			name:         "nil email service (optional)",
			tracing:      suite.mockTracing,
			metrics:      suite.mockMetrics,
			emailService: nil,
			slackService: suite.mockSlack,
			repo:         suite.mockRepo,
			expectNil:    false,
		},
		{
			name:         "nil slack service (optional)",
			tracing:      suite.mockTracing,
			metrics:      suite.mockMetrics,
			emailService: suite.mockEmail,
			slackService: nil,
			repo:         suite.mockRepo,
			expectNil:    false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			svc := NewNotificationService(
				tc.tracing,
				tc.metrics,
				tc.emailService,
				tc.slackService,
				tc.repo,
			)

			if tc.expectNil {
				require.Nil(suite.T(), svc)
			} else {
				require.NotNil(suite.T(), svc)
			}
		})
	}
}

// TestRunner runs the test suite
func TestNotificationServiceSuite(t *testing.T) {
	suite.Run(t, new(NotificationServiceTestSuite))
}
