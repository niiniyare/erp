package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// Repository defines the interface for notification data persistence,
// primarily for managing user notification preferences.
type Repository interface {
	GetUserNotificationPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error)
	UpdateUserNotificationPreferences(ctx context.Context, userID uuid.UUID, prefs *NotificationPreferences) error
}

// MockRepository is a mock implementation of the Repository interface for testing purposes.
type MockRepository struct{}

// NewMockRepository creates a new mock repository.
func NewMockRepository() Repository {
	return &MockRepository{}
}

// GetUserNotificationPreferences returns default preferences for now.
func (m *MockRepository) GetUserNotificationPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error) {
	logger.Info("Mock GetUserNotificationPreferences called", logger.Fields{"user_id": userID})
	// Return default preferences for now
	return &NotificationPreferences{
		UserID:             userID,
		EmailNotifications: true,
		InAppNotifications: true,
		SlackNotifications: true,
		NotificationTypes: map[NotificationType]bool{
			NotificationTypeAccessRequestCreated:  true,
			NotificationTypeAccessRequestApproved: true,
			NotificationTypeAccessRequestRejected: true,
			NotificationTypeAccessRequestExpired:  true,
			NotificationTypeAccessRequestRevoked:  true,
			NotificationTypeGeneric:               true,
		},
		PreferredChannels: []NotificationChannel{NotificationChannelEmail, NotificationChannelInApp},
		QuietHours:        &QuietHours{Enabled: false},
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}, nil
}

// UpdateUserNotificationPreferences logs the update and returns no error.
func (m *MockRepository) UpdateUserNotificationPreferences(ctx context.Context, userID uuid.UUID, prefs *NotificationPreferences) error {
	logger.Info("Mock UpdateUserNotificationPreferences called", logger.Fields{"user_id": userID})
	return nil
}
