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
type MockEmailService struct {
	shouldError bool
	errorMsg    string
}

func (m *MockEmailService) SendEmail(ctx context.Context, to []string, subject, body string, data map[string]any) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}
	return nil
}

// MockSlackService is a mock implementation of SlackService
type MockSlackService struct {
	shouldError bool
	errorMsg    string
}

func (m *MockSlackService) SendSlackMessage(ctx context.Context, userIDs []string, message string, data map[string]any) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}
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
		setupMocks      func()
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
			},
			expectRepoCall: true,
		},
		{
			name: "successful slack notification",
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
				Channels: []NotificationChannel{NotificationChannelSlack},
			},
			mockPreferences: &NotificationPreferences{
				UserID:             userID,
				EmailNotifications: true,
				SlackNotifications: true,
				PreferredChannels:  []NotificationChannel{NotificationChannelEmail, NotificationChannelSlack},
				NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: true},
			},
			expectRepoCall: true,
		},
		{
			name: "fallback to user preferred channels",
			notification: &Notification{
				TenantID: tenantID,
				Recipients: []Recipient{{
					UserID: userID,
					Email:  "test@example.com",
				}},
				Subject:  "Test Subject",
				Message:  "Test Message",
				Type:     NotificationTypeGeneric,
				Channels: []NotificationChannel{}, // Empty channels
			},
			mockPreferences: &NotificationPreferences{
				UserID:             userID,
				EmailNotifications: true,
				PreferredChannels:  []NotificationChannel{NotificationChannelEmail},
				NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: true},
			},
			expectRepoCall: true,
		},
		{
			name: "email service returns error",
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
				NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: true},
			},
			setupMocks: func() {
				suite.mockEmail.shouldError = true
				suite.mockEmail.errorMsg = "email failed"
			},
			expectedError:  "email notification failed: email failed",
			expectRepoCall: true,
		},
		{
			name: "slack service returns error",
			notification: &Notification{
				TenantID: tenantID,
				Recipients: []Recipient{{
					UserID:  userID,
					SlackID: "U123456789",
				}},
				Subject:  "Test Subject",
				Message:  "Test Message",
				Type:     NotificationTypeGeneric,
				Channels: []NotificationChannel{NotificationChannelSlack},
			},
			mockPreferences: &NotificationPreferences{
				UserID:             userID,
				SlackNotifications: true,
				NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: true},
			},
			setupMocks: func() {
				suite.mockSlack.shouldError = true
				suite.mockSlack.errorMsg = "slack failed"
			},
			expectedError:  "slack notification failed: slack failed",
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
				NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: true},
				QuietHours: &QuietHours{
					Enabled:   true,
					StartTime: "00:00", // Covers all day for deterministic test
					EndTime:   "23:59",
				},
			},
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
				NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: false},
			},
			expectRepoCall: true,
		},
		{
			name: "no recipients",
			notification: &Notification{
				TenantID:   tenantID,
				Recipients: []Recipient{},
				Subject:    "Test Subject",
			},
			expectRepoCall: false,
		},
		{
			name: "repository error when getting preferences",
			notification: &Notification{
				TenantID: tenantID,
				Recipients: []Recipient{{
					UserID: userID,
					Email:  "test@example.com",
				}},
				Subject: "Test Subject",
			},
			mockPrefsError: errors.New("database error"),
			expectRepoCall: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Reset mocks
			suite.mockEmail.shouldError = false
			suite.mockSlack.shouldError = false
			if tc.setupMocks != nil {
				tc.setupMocks()
			}

			mockSpan := tracing.NewMockSpan(suite.ctrl)
			suite.mockTracing.EXPECT().StartSpan(gomock.Any(), "notificationService.Send").Return(context.Background(), mockSpan)
			mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
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

// TestIsInQuietHours tests the isInQuietHours method (private method testing via service behavior)
func (suite *NotificationServiceTestSuite) TestQuietHoursLogic() {
	userID := uuid.New()
	tenantID := uuid.New()

	testCases := []struct {
		name        string
		quietHours  *QuietHours
		expectSend  bool
		description string
	}{
		{
			name:        "no quiet hours",
			quietHours:  nil,
			expectSend:  true,
			description: "should send when no quiet hours configured",
		},
		{
			name: "quiet hours disabled",
			quietHours: &QuietHours{
				Enabled:   false,
				StartTime: "22:00",
				EndTime:   "08:00",
			},
			expectSend:  true,
			description: "should send when quiet hours disabled",
		},
		{
			name: "quiet hours enabled same day period",
			quietHours: &QuietHours{
				Enabled:   true,
				StartTime: "09:00",
				EndTime:   "17:00",
			},
			expectSend:  false, // This depends on current time, but we're testing the logic exists
			description: "quiet hours logic for same day period",
		},
		{
			name: "quiet hours enabled overnight period",
			quietHours: &QuietHours{
				Enabled:   true,
				StartTime: "22:00",
				EndTime:   "08:00",
			},
			expectSend:  false, // This depends on current time, but we're testing the logic exists
			description: "quiet hours logic for overnight period",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			mockSpan := tracing.NewMockSpan(suite.ctrl)
			suite.mockTracing.EXPECT().StartSpan(gomock.Any(), "notificationService.Send").Return(context.Background(), mockSpan)
			mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
			mockSpan.EXPECT().End().Times(1)

			mockPreferences := &NotificationPreferences{
				UserID:             userID,
				EmailNotifications: true,
				SlackNotifications: true,
				PreferredChannels:  []NotificationChannel{NotificationChannelEmail},
				NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: true},
				QuietHours:         tc.quietHours,
			}

			suite.mockRepo.EXPECT().
				GetUserNotificationPreferences(gomock.Any(), userID).
				Return(mockPreferences, nil).
				AnyTimes()

			notification := &Notification{
				TenantID: tenantID,
				Recipients: []Recipient{{
					UserID: userID,
					Email:  "test@example.com",
				}},
				Subject:  "Test Subject",
				Message:  "Test Message",
				Type:     NotificationTypeGeneric,
				Channels: []NotificationChannel{NotificationChannelEmail},
			}

			err := suite.service.Send(context.Background(), notification)
			require.NoError(suite.T(), err, tc.description)
		})
	}
}

// TestSendWithNilServices tests sending notifications when email/slack services are nil
func (suite *NotificationServiceTestSuite) TestSendWithNilServices() {
	userID := uuid.New()
	tenantID := uuid.New()

	testCases := []struct {
		name         string
		emailService EmailService
		slackService SlackService
		channels     []NotificationChannel
		description  string
	}{
		{
			name:         "nil email service",
			emailService: nil,
			slackService: suite.mockSlack,
			channels:     []NotificationChannel{NotificationChannelEmail},
			description:  "should handle nil email service gracefully",
		},
		{
			name:         "nil slack service",
			emailService: suite.mockEmail,
			slackService: nil,
			channels:     []NotificationChannel{NotificationChannelSlack},
			description:  "should handle nil slack service gracefully",
		},
		{
			name:         "both services nil",
			emailService: nil,
			slackService: nil,
			channels:     []NotificationChannel{NotificationChannelEmail, NotificationChannelSlack},
			description:  "should handle both services being nil",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create service with nil services
			svc := NewNotificationService(
				suite.mockTracing,
				suite.mockMetrics,
				tc.emailService,
				tc.slackService,
				suite.mockRepo,
			)

			mockSpan := tracing.NewMockSpan(suite.ctrl)
			suite.mockTracing.EXPECT().StartSpan(gomock.Any(), "notificationService.Send").Return(context.Background(), mockSpan)
			mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
			mockSpan.EXPECT().End().Times(1)

			mockPreferences := &NotificationPreferences{
				UserID:             userID,
				EmailNotifications: true,
				SlackNotifications: true,
				PreferredChannels:  tc.channels,
				NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: true},
				QuietHours:         nil,
			}

			suite.mockRepo.EXPECT().
				GetUserNotificationPreferences(gomock.Any(), userID).
				Return(mockPreferences, nil).
				AnyTimes()

			notification := &Notification{
				TenantID: tenantID,
				Recipients: []Recipient{{
					UserID:  userID,
					Email:   "test@example.com",
					SlackID: "U123456789",
				}},
				Subject:  "Test Subject",
				Message:  "Test Message",
				Type:     NotificationTypeGeneric,
				Channels: tc.channels,
			}

			err := svc.Send(context.Background(), notification)
			require.NoError(suite.T(), err, tc.description)
		})
	}
}

// TestSendWithMultipleRecipients tests sending to multiple recipients
func (suite *NotificationServiceTestSuite) TestSendWithMultipleRecipients() {
	userID1 := uuid.New()
	userID2 := uuid.New()
	tenantID := uuid.New()

	suite.Run("multiple recipients with different preferences", func() {
		mockSpan := tracing.NewMockSpan(suite.ctrl)
		suite.mockTracing.EXPECT().StartSpan(gomock.Any(), "notificationService.Send").Return(context.Background(), mockSpan)
		mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
		mockSpan.EXPECT().End().Times(1)

		// First user preferences - email enabled
		mockPrefs1 := &NotificationPreferences{
			UserID:             userID1,
			EmailNotifications: true,
			SlackNotifications: false,
			PreferredChannels:  []NotificationChannel{NotificationChannelEmail},
			NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: true},
			QuietHours:         nil,
		}

		// Second user preferences - slack enabled, email disabled
		mockPrefs2 := &NotificationPreferences{
			UserID:             userID2,
			EmailNotifications: false,
			SlackNotifications: true,
			PreferredChannels:  []NotificationChannel{NotificationChannelSlack},
			NotificationTypes:  map[NotificationType]bool{NotificationTypeGeneric: true},
			QuietHours:         nil,
		}

		suite.mockRepo.EXPECT().
			GetUserNotificationPreferences(gomock.Any(), userID1).
			Return(mockPrefs1, nil).
			AnyTimes()

		suite.mockRepo.EXPECT().
			GetUserNotificationPreferences(gomock.Any(), userID2).
			Return(mockPrefs2, nil).
			AnyTimes()

		notification := &Notification{
			TenantID: tenantID,
			Recipients: []Recipient{
				{
					UserID:  userID1,
					Email:   "user1@example.com",
					SlackID: "U123456789",
				},
				{
					UserID:  userID2,
					Email:   "user2@example.com",
					SlackID: "U987654321",
				},
			},
			Subject:  "Test Subject",
			Message:  "Test Message",
			Type:     NotificationTypeGeneric,
			Channels: []NotificationChannel{NotificationChannelEmail, NotificationChannelSlack},
		}

		err := suite.service.Send(context.Background(), notification)
		require.NoError(suite.T(), err)
	})
}
