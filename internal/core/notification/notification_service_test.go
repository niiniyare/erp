package notification

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/assert"
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

func TestNotificationService_Send(t *testing.T) {
	// Initialize mock dependencies
	mockTracing := tracing.NewMockTracingService()
	mockMetrics := metrics.NewMockMetricsService()
	mockEmail := &MockEmailService{}
	mockSlack := &MockSlackService{}
	mockRepo := NewMockRepository()

	// Create a new NotificationService instance
	svc := NewNotificationService(
		mockTracing,
		mockMetrics,
		mockEmail,
		mockSlack,
		mockRepo,
	)

	// Create a sample notification
	notification := &Notification{
		TenantID:   uuid.New(),
		Recipients: []Recipient{{UserID: uuid.New(), Email: "test@example.com"}},
		Subject:    "Test Subject",
		Message:    "Test Message",
		Type:       NotificationTypeGeneric,
		Channels:   []NotificationChannel{NotificationChannelEmail},
	}

	// Send the notification
	err := svc.Send(context.Background(), notification)

	// Assert no error occurred
	assert.NoError(t, err)
}

func TestNotificationService_NewNotificationService(t *testing.T) {
	mockTracing := tracing.NewMockTracingService()
	mockMetrics := metrics.NewMockMetricsService()
	mockEmail := &MockEmailService{}
	mockSlack := &MockSlackService{}
	mockRepo := NewMockRepository()

	svc := NewNotificationService(
		mockTracing,
		mockMetrics,
		mockEmail,
		mockSlack,
		mockRepo,
	)
	assert.NotNil(t, svc, "NewNotificationService should return a non-nil service")
}
