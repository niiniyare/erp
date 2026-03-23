package notification

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"awo.so/internal/shared/logger"
	"awo.so/internal/shared/metrics"
	"awo.so/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// NotificationService handles sending notifications via various channels.
type NotificationService interface {
	// Send sends a notification to the specified recipients.
	Send(ctx context.Context, notification *Notification) error

	// GetUserNotificationPreferences retrieves a user's notification settings.
	GetUserNotificationPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error)

	// UpdateUserNotificationPreferences updates a user's notification settings.
	UpdateUserNotificationPreferences(ctx context.Context, userID uuid.UUID, prefs *NotificationPreferences) error
}

// EmailService defines the interface for an external email sending service.
type EmailService interface {
	SendEmail(ctx context.Context, to []string, subject, body string, data map[string]any) error
}

// SlackService defines the interface for an external Slack messaging service.
type SlackService interface {
	SendSlackMessage(ctx context.Context, userIDs []string, message string, data map[string]any) error
}

// notificationService implements the NotificationService.
type notificationService struct {
	tracing      tracing.Service
	metrics      metrics.MetricsProvider
	emailService EmailService // Can be nil
	slackService SlackService // Can be nil
	repo         Repository
}

// NewNotificationService creates a new notification service.
// External services like email or slack can be nil if not configured.
func NewNotificationService(
	tracing tracing.Service,
	metrics metrics.MetricsProvider,
	emailService EmailService,
	slackService SlackService,
	repo Repository,
) NotificationService {
	return &notificationService{
		tracing:      tracing,
		metrics:      metrics,
		emailService: emailService,
		slackService: slackService,
		repo:         repo,
	}
}

// Send dispatches a notification based on recipient preferences and available channels.
func (s *notificationService) Send(ctx context.Context, notification *Notification) error {
	ctx, span := s.tracing.StartSpan(ctx, "notificationService.Send")
	defer span.End()

	span.SetAttributes(
		attribute.String("notification.type", string(notification.Type)),
		attribute.Int("notification.recipient_count", len(notification.Recipients)),
	)

	if len(notification.Recipients) == 0 {
		logger.Warn("Notification has no recipients, skipping.", logger.Fields{"subject": notification.Subject})
		return nil
	}

	var errors []error

	// Group recipients by their preferred channels
	emailRecipients := make([]string, 0)
	slackRecipientIDs := make([]string, 0)

	for _, recipient := range notification.Recipients {
		prefs, err := s.GetUserNotificationPreferences(ctx, recipient.UserID)
		if err != nil {
			logger.Error("Failed to get user notification preferences", logger.Fields{"user_id": recipient.UserID, "error": err})
			continue // Skip this recipient
		}

		// Skip if user has disabled this notification type
		if enabled, exists := prefs.NotificationTypes[notification.Type]; exists && !enabled {
			continue
		}

		// Skip if in quiet hours
		if s.isInQuietHours(prefs.QuietHours) {
			logger.Debug("Skipping notification due to quiet hours", logger.Fields{"user_id": recipient.UserID})
			continue
		}

		// Determine target channels
		targetChannels := notification.Channels
		if len(targetChannels) == 0 {
			targetChannels = prefs.PreferredChannels
		}

		for _, channel := range targetChannels {
			switch channel {
			case NotificationChannelEmail:
				if prefs.EmailNotifications && recipient.Email != "" {
					emailRecipients = append(emailRecipients, recipient.Email)
				}
			case NotificationChannelSlack:
				if prefs.SlackNotifications && recipient.SlackID != "" {
					slackRecipientIDs = append(slackRecipientIDs, recipient.SlackID)
				}
			}
		}
	}

	// Send via Email
	if len(emailRecipients) > 0 && s.emailService != nil {
		if err := s.emailService.SendEmail(ctx, emailRecipients, notification.Subject, notification.Message, notification.Data); err != nil {
			errors = append(errors, fmt.Errorf("email notification failed: %w", err))
			s.metrics.IncrementCounter("notifications_failed", map[string]any{"channel": "email", "type": string(notification.Type)})
		} else {
			s.metrics.IncrementCounter("notifications_sent", map[string]any{"channel": "email", "type": string(notification.Type)})
		}
	}

	// Send via Slack
	if len(slackRecipientIDs) > 0 && s.slackService != nil {
		if err := s.slackService.SendSlackMessage(ctx, slackRecipientIDs, notification.Message, notification.Data); err != nil {
			errors = append(errors, fmt.Errorf("slack notification failed: %w", err))
			s.metrics.IncrementCounter("notifications_failed", map[string]any{"channel": "slack", "type": string(notification.Type)})
		} else {
			s.metrics.IncrementCounter("notifications_sent", map[string]any{"channel": "slack", "type": string(notification.Type)})
		}
	}

	if len(errors) > 0 {
		// For simplicity, return the first error. In a real-world scenario, you might want to return a multi-error.
		return errors[0]
	}

	return nil
}

// GetUserNotificationPreferences retrieves a user's notification preferences.
func (s *notificationService) GetUserNotificationPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error) {
	return s.repo.GetUserNotificationPreferences(ctx, userID)
}

// UpdateUserNotificationPreferences updates a user's notification preferences.
func (s *notificationService) UpdateUserNotificationPreferences(ctx context.Context, userID uuid.UUID, prefs *NotificationPreferences) error {
	return s.repo.UpdateUserNotificationPreferences(ctx, userID, prefs)
}

// isInQuietHours checks if the current time is within the user's quiet hours.
// Note: This is a simplified implementation. A robust solution would require timezone handling.
func (s *notificationService) isInQuietHours(quietHours *QuietHours) bool {
	if quietHours == nil || !quietHours.Enabled {
		return false
	}
	// Simplified logic, does not handle timezones correctly.
	now := time.Now().UTC()
	currentHour := now.Hour()
	start, _ := time.Parse("15:04", quietHours.StartTime)
	end, _ := time.Parse("15:04", quietHours.EndTime)

	if end.Before(start) { // Overnight period (e.g., 22:00 - 08:00)
		return currentHour >= start.Hour() || currentHour < end.Hour()
	}
	// Same-day period (end hour is inclusive)
	return currentHour >= start.Hour() && currentHour <= end.Hour()
}
