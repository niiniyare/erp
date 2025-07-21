package notification

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// createTestUser is a helper function to create a tenant, entity, and user for testing.
func createTestUser(t *testing.T) *db.User {
	ctx := context.Background()

	// 1. Create Tenant
	tenant, err := testStore.CreateTenant(ctx, db.CreateTenantParams{
		Name:   "notif-tenant-" + uuid.NewString(),
		Slug:   "notif-tenant-" + uuid.NewString(),
		Email:  "notif-" + uuid.NewString() + "@test.com",
		Status: "active",
	})
	require.NoError(t, err)

	// 2. Create Entity in the tenant's context
	conn, err := testPool.Acquire(ctx)
	require.NoError(t, err)
	defer conn.Release()

	_, err = conn.Exec(ctx, fmt.Sprintf("SET app.tenant_id = '%s'", tenant.ID.String()))
	require.NoError(t, err)

	q := db.New(conn)

	entity, err := q.CreateEntity(ctx, db.CreateEntityParams{
		Uuid:     uuid.New(),
		Name:     "Notif Test Entity",
		Type:     "department",
		IsActive: true,
	})
	require.NoError(t, err)

	// 3. Create User
	username := "notif-user-" + uuid.NewString()
	email := "notif-" + uuid.NewString() + "@test.com"
	user, err := q.CreateUser(ctx, db.CreateUserParams{
		EntityID: entity.Uuid,
		Username: &username,
		Email:    email,
		UserType: "INTERNAL",
	})
	require.NoError(t, err)

	return user
}

func TestNotificationService_Send(t *testing.T) {
	// Initialize mock dependencies
	mockTracing := tracing.NewMockTracingService()
	mockMetrics := metrics.NewMockMetricsService()
	mockEmail := &MockEmailService{}
	mockSlack := &MockSlackService{}
	repo := NewRepository(testStore)

	// Create a new NotificationService instance
	svc := NewNotificationService(
		mockTracing,
		mockMetrics,
		mockEmail,
		mockSlack,
		repo,
	)

	user := createTestUser(t)

	// Create a sample notification
	notification := &Notification{
		TenantID:   user.TenantID,
		Recipients: []Recipient{{UserID: user.ID, Email: user.Email}},
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
	repo := NewRepository(testStore)

	svc := NewNotificationService(
		mockTracing,
		mockMetrics,
		mockEmail,
		mockSlack,
		repo,
	)
	assert.NotNil(t, svc, "NewNotificationService should return a non-nil service")
}
