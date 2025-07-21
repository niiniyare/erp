package notification

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeAccessRequestCreated  NotificationType = "ACCESS_REQUEST_CREATED"
	NotificationTypeAccessRequestApproved NotificationType = "ACCESS_REQUEST_APPROVED"
	NotificationTypeAccessRequestRejected NotificationType = "ACCESS_REQUEST_REJECTED"
	NotificationTypeAccessRequestExpired  NotificationType = "ACCESS_REQUEST_EXPIRED"
	NotificationTypeAccessRequestRevoked  NotificationType = "ACCESS_REQUEST_REVOKED"
	NotificationTypeGeneric               NotificationType = "GENERIC"
)

// NotificationChannel represents the delivery channel for notifications
type NotificationChannel string

const (
	NotificationChannelEmail   NotificationChannel = "EMAIL"
	NotificationChannelInApp   NotificationChannel = "IN_APP"
	NotificationChannelSlack   NotificationChannel = "SLACK"
	NotificationChannelWebhook NotificationChannel = "WEBHOOK"
)

// Recipient contains the necessary information for a single notification recipient.
type Recipient struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email,omitempty"`
	SlackID     string    `json:"slack_id,omitempty"`
	PhoneNumber string    `json:"phone_number,omitempty"`
}

// Notification represents a notification to be sent.
// It is a simplified version of NotificationRequest, focused on the core content.
type Notification struct {
	TenantID   uuid.UUID             `json:"tenant_id"`
	Recipients []Recipient           `json:"recipients"`
	Subject    string                `json:"subject"`
	Message    string                `json:"message"`
	Data       map[string]any        `json:"data"`               // For templating or rich content
	Type       NotificationType      `json:"type"`               // To respect user preferences
	Channels   []NotificationChannel `json:"channels,omitempty"` // Optional: force specific channels
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
