package audit

import (
	"context"
	"testing"
	"time"

	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/stretchr/testify/assert"
)

func TestAuditService_LogAuditEvent(t *testing.T) {
	// Initialize mock dependencies
	mockTracing := tracing.NewMockTracingService()
	mockMetrics := metrics.NewMockMetricsService()
	mockRepo := NewMockRepository() // Using the mock repository from this package

	// Create a new AuditService instance
	svc := NewAuditService(mockTracing, mockMetrics, mockRepo)

	// Create a sample audit event
	event := &AuditEvent{
		EventType:     AuditEventAccessGranted,
		EventCategory: AuditCategoryAccess,
		Severity:      AuditSeverityInfo,
		Reason:        "Test access granted",
		Timestamp:     time.Now(),
	}

	// Log the audit event
	err := svc.LogAuditEvent(context.Background(), event)

	// Assert no error occurred
	assert.NoError(t, err)

	// Optionally, add assertions to check if the mock repository's CreateAuditEvent was called
	// (This would require adding a method to MockRepository to track calls,
	// but for a basic test, just checking for no error is sufficient for now)
}

func TestAuditService_NewAuditService(t *testing.T) {
	mockTracing := tracing.NewMockTracingService()
	mockMetrics := metrics.NewMockMetricsService()
	mockRepo := NewMockRepository()

	svc := NewAuditService(mockTracing, mockMetrics, mockRepo)
	assert.NotNil(t, svc, "NewAuditService should return a non-nil service")
}
