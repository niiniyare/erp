package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// Repository defines the interface for audit data persistence.
type Repository interface {
	CreateAuditEvent(ctx context.Context, event *AuditEvent) error
	GetAuditEvents(ctx context.Context, req *AuditQueryRequest) ([]*AuditEvent, error)
	GetUserAuditTrail(ctx context.Context, userID uuid.UUID, fromDate, toDate time.Time) ([]*AuditEvent, error)
	GetWorkflowAuditTrail(ctx context.Context, requestID uuid.UUID) ([]*WorkflowAuditEvent, error)
}

// MockRepository is a mock implementation of the Repository interface for testing purposes.
type MockRepository struct{}

// NewMockRepository creates a new mock repository.
func NewMockRepository() Repository {
	return &MockRepository{}
}

// CreateAuditEvent logs the audit event to the console instead of a database.
func (m *MockRepository) CreateAuditEvent(ctx context.Context, event *AuditEvent) error {
	logger.Info("Mock CreateAuditEvent called", logger.Fields{
		"event_id":   event.ID,
		"event_type": event.EventType,
	})
	return nil
}

// GetAuditEvents returns a mock list of audit events.
func (m *MockRepository) GetAuditEvents(ctx context.Context, req *AuditQueryRequest) ([]*AuditEvent, error) {
	logger.Info("Mock GetAuditEvents called", nil)
	return []*AuditEvent{}, nil
}

// GetUserAuditTrail returns a mock user audit trail.
func (m *MockRepository) GetUserAuditTrail(ctx context.Context, userID uuid.UUID, fromDate, toDate time.Time) ([]*AuditEvent, error) {
	logger.Info("Mock GetUserAuditTrail called", logger.Fields{"user_id": userID})
	return []*AuditEvent{}, nil
}

// GetWorkflowAuditTrail returns a mock workflow audit trail.
func (m *MockRepository) GetWorkflowAuditTrail(ctx context.Context, requestID uuid.UUID) ([]*WorkflowAuditEvent, error) {
	logger.Info("Mock GetWorkflowAuditTrail called", logger.Fields{"request_id": requestID})
	return []*WorkflowAuditEvent{}, nil
}
