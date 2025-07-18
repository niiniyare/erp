package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeAccessRequestCreated  NotificationType = "ACCESS_REQUEST_CREATED"
	NotificationTypeAccessRequestApproved NotificationType = "ACCESS_REQUEST_APPROVED"
	NotificationTypeAccessRequestRejected NotificationType = "ACCESS_REQUEST_REJECTED"
	NotificationTypeAccessRequestExpired  NotificationType = "ACCESS_REQUEST_EXPIRED"
	NotificationTypeAccessRequestRevoked  NotificationType = "ACCESS_REQUEST_REVOKED"
)

// NotificationChannel represents the delivery channel for notifications
type NotificationChannel string

const (
	NotificationChannelEmail   NotificationChannel = "EMAIL"
	NotificationChannelInApp   NotificationChannel = "IN_APP"
	NotificationChannelSlack   NotificationChannel = "SLACK"
	NotificationChannelWebhook NotificationChannel = "WEBHOOK"
)

// NotificationRequest represents a notification to be sent
type NotificationRequest struct {
	Type       NotificationType      `json:"type"`
	Channels   []NotificationChannel `json:"channels"`
	Recipients []uuid.UUID           `json:"recipients"`
	Subject    string                `json:"subject"`
	Message    string                `json:"message"`
	Data       map[string]any        `json:"data"`
	Priority   string                `json:"priority"`
	RequestID  uuid.UUID             `json:"request_id"`
}

// NotificationService handles all notification-related operations for access requests
type NotificationService interface {
	// Core notification operations
	SendAccessRequestNotification(ctx context.Context, request *AccessRequest, notificationType NotificationType, recipients []uuid.UUID) error

	// Specific workflow notifications
	NotifyApprovers(ctx context.Context, request *AccessRequest) error
	NotifyRequester(ctx context.Context, request *AccessRequest, action string) error
	NotifyStatusChange(ctx context.Context, request *AccessRequest, previousStatus, newStatus ApprovalStatus) error

	// Administrative notifications
	NotifyExpiredRequests(ctx context.Context, requests []*AccessRequest) error
	SendBulkNotification(ctx context.Context, notificationReq *NotificationRequest) error

	// User preference management
	GetUserNotificationPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error)
	UpdateUserNotificationPreferences(ctx context.Context, userID uuid.UUID, prefs *NotificationPreferences) error
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	UserID             uuid.UUID                 `json:"user_id"`
	EmailNotifications bool                      `json:"email_notifications"`
	InAppNotifications bool                      `json:"in_app_notifications"`
	SlackNotifications bool                      `json:"slack_notifications"`
	NotificationTypes  map[NotificationType]bool `json:"notification_types"`
	PreferredChannels  []NotificationChannel     `json:"preferred_channels"`
	QuietHours         *QuietHours               `json:"quiet_hours,omitempty"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
}

// QuietHours represents hours when notifications should not be sent
type QuietHours struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"start_time"` // Format: "22:00"
	EndTime   string `json:"end_time"`   // Format: "08:00"
	Timezone  string `json:"timezone"`   // Format: "UTC", "America/New_York"
}

// notificationService implements NotificationService
type notificationService struct {
	userRepo        Repository
	tracing         *tracing.TracingService
	metrics         *metrics.MetricsService
	emailService    EmailService
	slackService    SlackService
	approverService ApproverService
}

// EmailService interface for email notifications
type EmailService interface {
	SendEmail(ctx context.Context, to []string, subject, body string, data map[string]any) error
}

// SlackService interface for Slack notifications
type SlackService interface {
	SendSlackMessage(ctx context.Context, userIDs []uuid.UUID, message string, data map[string]any) error
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	userRepo Repository,
	tracing *tracing.TracingService,
	metrics *metrics.MetricsService,
	emailService EmailService,
	slackService SlackService,
	approverService ApproverService,
) NotificationService {
	return &notificationService{
		userRepo:        userRepo,
		tracing:         tracing,
		metrics:         metrics,
		emailService:    emailService,
		slackService:    slackService,
		approverService: approverService,
	}
}

// SendAccessRequestNotification sends notification for access request events
func (s *notificationService) SendAccessRequestNotification(ctx context.Context, request *AccessRequest, notificationType NotificationType, recipients []uuid.UUID) error {
	ctx, span := s.tracing.StartSpan(ctx, "notificationService.SendAccessRequestNotification")
	defer span.End()

	span.SetAttributes(
		attribute.String("notification_type", string(notificationType)),
		attribute.String("request_id", request.ID.String()),
		attribute.Int("recipient_count", len(recipients)),
	)

	if len(recipients) == 0 {
		logger.Warn("No recipients specified for notification", logger.Fields{
			"request_id":        request.ID,
			"notification_type": notificationType,
		})
		return nil
	}

	// Generate notification content
	subject, message := s.generateNotificationContent(request, notificationType)

	// Prepare notification data
	data := map[string]any{
		"request_id":      request.ID,
		"request_type":    request.RequestType,
		"requester_id":    request.RequesterID,
		"entity_id":       request.EntityID,
		"approval_status": request.ApprovalStatus,
		"justification":   request.Justification,
		"business_reason": request.BusinessReason,
		"created_at":      request.CreatedAt,
		"expires_at":      request.ExpiresAt,
	}

	notificationReq := &NotificationRequest{
		Type:       notificationType,
		Recipients: recipients,
		Subject:    subject,
		Message:    message,
		Data:       data,
		Priority:   s.getPriority(notificationType),
		RequestID:  request.ID,
	}

	return s.SendBulkNotification(ctx, notificationReq)
}

// NotifyApprovers sends notifications to potential approvers
func (s *notificationService) NotifyApprovers(ctx context.Context, request *AccessRequest) error {
	ctx, span := s.tracing.StartSpan(ctx, "notificationService.NotifyApprovers")
	defer span.End()

	// Determine who can approve this request
	approvers, err := s.determineApprovers(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to determine approvers")
		return err
	}

	if len(approvers) == 0 {
		logger.Warn("No approvers found for access request", logger.Fields{
			"request_id":   request.ID,
			"request_type": request.RequestType,
			"entity_id":    request.EntityID,
		})
		return nil
	}

	logger.Info("Sending notifications to approvers", logger.Fields{
		"request_id":     request.ID,
		"approver_count": len(approvers),
	})

	return s.SendAccessRequestNotification(ctx, request, NotificationTypeAccessRequestCreated, approvers)
}

// NotifyRequester sends notification to the requester about status changes
func (s *notificationService) NotifyRequester(ctx context.Context, request *AccessRequest, action string) error {
	ctx, span := s.tracing.StartSpan(ctx, "notificationService.NotifyRequester")
	defer span.End()

	var notificationType NotificationType
	switch action {
	case "approved":
		notificationType = NotificationTypeAccessRequestApproved
	case "rejected":
		notificationType = NotificationTypeAccessRequestRejected
	case "expired":
		notificationType = NotificationTypeAccessRequestExpired
	case "revoked":
		notificationType = NotificationTypeAccessRequestRevoked
	default:
		logger.Warn("Unknown notification action", logger.Fields{
			"action":     action,
			"request_id": request.ID,
		})
		return nil
	}

	recipients := []uuid.UUID{request.RequesterID}

	// If the request is for someone else, notify them too
	if request.TargetUserID != nil && *request.TargetUserID != request.RequesterID {
		recipients = append(recipients, *request.TargetUserID)
	}

	return s.SendAccessRequestNotification(ctx, request, notificationType, recipients)
}

// NotifyStatusChange sends notifications when request status changes
func (s *notificationService) NotifyStatusChange(ctx context.Context, request *AccessRequest, previousStatus, newStatus ApprovalStatus) error {
	ctx, span := s.tracing.StartSpan(ctx, "notificationService.NotifyStatusChange")
	defer span.End()

	logger.Info("Sending status change notification", logger.Fields{
		"request_id":      request.ID,
		"previous_status": previousStatus,
		"new_status":      newStatus,
	})

	var action string
	switch newStatus {
	case ApprovalStatusApproved:
		action = "approved"
	case ApprovalStatusRejected:
		action = "rejected"
	case ApprovalStatusExpired:
		action = "expired"
	case ApprovalStatusRevoked:
		action = "revoked"
	default:
		return nil
	}

	return s.NotifyRequester(ctx, request, action)
}

// NotifyExpiredRequests sends notifications about expired requests
func (s *notificationService) NotifyExpiredRequests(ctx context.Context, requests []*AccessRequest) error {
	ctx, span := s.tracing.StartSpan(ctx, "notificationService.NotifyExpiredRequests")
	defer span.End()

	for _, request := range requests {
		if err := s.NotifyRequester(ctx, request, "expired"); err != nil {
			logger.Error("Failed to send expiration notification", logger.Fields{
				"request_id": request.ID,
				"error":      err.Error(),
			})
			// Continue with other notifications
		}
	}

	return nil
}

// SendBulkNotification sends notifications through appropriate channels
func (s *notificationService) SendBulkNotification(ctx context.Context, notificationReq *NotificationRequest) error {
	ctx, span := s.tracing.StartSpan(ctx, "notificationService.SendBulkNotification")
	defer span.End()

	var errors []error

	// Get user preferences and filter recipients by channel
	emailRecipients := make([]string, 0)
	slackRecipients := make([]uuid.UUID, 0)

	for _, userID := range notificationReq.Recipients {
		prefs, err := s.GetUserNotificationPreferences(ctx, userID)
		if err != nil {
			logger.Error("Failed to get user notification preferences", logger.Fields{
				"user_id": userID,
				"error":   err.Error(),
			})
			continue
		}

		// Check if user wants this type of notification
		if enabled, exists := prefs.NotificationTypes[notificationReq.Type]; exists && !enabled {
			continue
		}

		// Check quiet hours
		if s.isInQuietHours(prefs.QuietHours) {
			logger.Debug("Skipping notification due to quiet hours", logger.Fields{
				"user_id": userID,
			})
			continue
		}

		// Add to appropriate channels based on preferences
		for _, channel := range prefs.PreferredChannels {
			switch channel {
			case NotificationChannelEmail:
				if prefs.EmailNotifications {
					// Get user email
					user, err := s.userRepo.GetUserByID(ctx, userID)
					if err == nil && user.Email != "" {
						emailRecipients = append(emailRecipients, user.Email)
					}
				}
			case NotificationChannelSlack:
				if prefs.SlackNotifications {
					slackRecipients = append(slackRecipients, userID)
				}
			}
		}
	}

	// Send email notifications
	if len(emailRecipients) > 0 && s.emailService != nil {
		if err := s.emailService.SendEmail(ctx, emailRecipients, notificationReq.Subject, notificationReq.Message, notificationReq.Data); err != nil {
			errors = append(errors, fmt.Errorf("email notification failed: %w", err))
			s.metrics.IncrementCounter("notification_email_failed", map[string]any{"type": string(notificationReq.Type)})
		} else {
			s.metrics.IncrementCounter("notification_email_sent", map[string]any{"type": string(notificationReq.Type)})
		}
	}

	// Send Slack notifications
	if len(slackRecipients) > 0 && s.slackService != nil {
		if err := s.slackService.SendSlackMessage(ctx, slackRecipients, notificationReq.Message, notificationReq.Data); err != nil {
			errors = append(errors, fmt.Errorf("slack notification failed: %w", err))
			s.metrics.IncrementCounter("notification_slack_failed", map[string]any{"type": string(notificationReq.Type)})
		} else {
			s.metrics.IncrementCounter("notification_slack_sent", map[string]any{"type": string(notificationReq.Type)})
		}
	}

	// TODO: Implement in-app notifications (store in database for UI)
	// TODO: Implement webhook notifications for external systems

	if len(errors) > 0 {
		logger.Error("Some notifications failed to send", logger.Fields{
			"error_count": len(errors),
			"request_id":  notificationReq.RequestID,
		})
		// Return first error for now
		return errors[0]
	}

	s.metrics.IncrementCounter("notification_sent", map[string]any{"type": string(notificationReq.Type)})
	return nil
}

// GetUserNotificationPreferences gets user notification preferences
func (s *notificationService) GetUserNotificationPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error) {
	// TODO: Implement database storage for notification preferences
	// For now, return default preferences
	defaultPrefs := &NotificationPreferences{
		UserID:             userID,
		EmailNotifications: true,
		InAppNotifications: true,
		SlackNotifications: false,
		NotificationTypes: map[NotificationType]bool{
			NotificationTypeAccessRequestCreated:  true,
			NotificationTypeAccessRequestApproved: true,
			NotificationTypeAccessRequestRejected: true,
			NotificationTypeAccessRequestExpired:  true,
			NotificationTypeAccessRequestRevoked:  true,
		},
		PreferredChannels: []NotificationChannel{
			NotificationChannelEmail,
			NotificationChannelInApp,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return defaultPrefs, nil
}

// UpdateUserNotificationPreferences updates user notification preferences
func (s *notificationService) UpdateUserNotificationPreferences(ctx context.Context, userID uuid.UUID, prefs *NotificationPreferences) error {
	// TODO: Implement database storage for notification preferences
	logger.Info("Updated user notification preferences", logger.Fields{
		"user_id": userID,
	})
	return nil
}

// Helper methods

// determineApprovers determines who can approve an access request
func (s *notificationService) determineApprovers(ctx context.Context, request *AccessRequest) ([]uuid.UUID, error) {
	return s.approverService.GetEligibleApprovers(ctx, request)
}

// generateNotificationContent generates subject and message for notifications
func (s *notificationService) generateNotificationContent(request *AccessRequest, notificationType NotificationType) (string, string) {
	switch notificationType {
	case NotificationTypeAccessRequestCreated:
		subject := fmt.Sprintf("New Access Request: %s", request.RequestType)
		message := fmt.Sprintf(
			"A new access request has been submitted and requires your approval.\n\n"+
				"Request Type: %s\n"+
				"Justification: %s\n"+
				"Business Reason: %s\n"+
				"Created: %s\n\n"+
				"Please review and approve/reject this request.",
			request.RequestType,
			request.Justification,
			request.BusinessReason,
			request.CreatedAt.Format(time.RFC3339),
		)
		return subject, message

	case NotificationTypeAccessRequestApproved:
		subject := "Access Request Approved"
		message := fmt.Sprintf(
			"Your access request has been approved.\n\n"+
				"Request Type: %s\n"+
				"Approved: %s\n\n"+
				"The requested access has been granted.",
			request.RequestType,
			request.ApprovedAt.Format(time.RFC3339),
		)
		return subject, message

	case NotificationTypeAccessRequestRejected:
		subject := "Access Request Rejected"
		message := fmt.Sprintf(
			"Your access request has been rejected.\n\n"+
				"Request Type: %s\n"+
				"Comments: %s\n\n"+
				"Please contact your administrator if you need further clarification.",
			request.RequestType,
			request.ApprovalComments,
		)
		return subject, message

	case NotificationTypeAccessRequestExpired:
		subject := "Access Request Expired"
		message := fmt.Sprintf(
			"Your access request has expired.\n\n"+
				"Request Type: %s\n"+
				"Expired: %s\n\n"+
				"Please submit a new request if you still need access.",
			request.RequestType,
			time.Now().Format(time.RFC3339),
		)
		return subject, message

	case NotificationTypeAccessRequestRevoked:
		subject := "Access Request Revoked"
		message := fmt.Sprintf(
			"Your access request has been revoked.\n\n"+
				"Request Type: %s\n"+
				"Revoked: %s\n\n"+
				"The granted access has been removed.",
			request.RequestType,
			time.Now().Format(time.RFC3339),
		)
		return subject, message

	default:
		return "Access Request Notification", "An access request event has occurred."
	}
}

// getPriority determines notification priority based on type
func (s *notificationService) getPriority(notificationType NotificationType) string {
	switch notificationType {
	case NotificationTypeAccessRequestCreated:
		return "high"
	case NotificationTypeAccessRequestApproved, NotificationTypeAccessRequestRejected:
		return "medium"
	case NotificationTypeAccessRequestExpired, NotificationTypeAccessRequestRevoked:
		return "low"
	default:
		return "medium"
	}
}

// isInQuietHours checks if current time is within user's quiet hours
func (s *notificationService) isInQuietHours(quietHours *QuietHours) bool {
	if quietHours == nil || !quietHours.Enabled {
		return false
	}

	// TODO: Implement proper timezone-aware quiet hours checking
	// This is a simplified implementation
	now := time.Now()
	currentHour := now.Hour()

	// Parse start and end times (simplified)
	// In a real implementation, this would properly handle timezones and time parsing
	return currentHour >= 22 || currentHour <= 8
}
