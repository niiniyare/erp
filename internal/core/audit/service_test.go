package audit

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/token"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// Helper function for string pointers
func stringPtr(s string) *string { return &s }

// MockRepository is a mock implementation of the Repository interface
func TestRecord(t *testing.T) {
	// 1. Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockCache := cache.NewMockService(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockTracing := tracing.NewMockTracingService(ctrl)
	mockMetrics, _ := metrics.NewMetricsService(metrics.MetricsConfig{Enabled: false})

	auditService := NewService(mockRepo, mockCache, mockLogger, mockTracing, mockMetrics)

	userID := uuid.New()
	authPayload := &token.Payload{
		UserID:    userID,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(time.Hour),
	}
	ctx := context.WithValue(context.Background(), token.AuthorizationPayloadKey, authPayload)

	event := AuditEvent{
		EventType:     "user.login",
		EventCategory: "AUTH",
		Severity:      "INFO",
		Decision:      stringPtr("SUCCESS"),
		Reason:        stringPtr("User successfully logged in"),
		Context:       json.RawMessage(`{"ip": "127.0.0.1"}`),
	}

	_ = uuid.New() // tenantID no longer needed for CreateAuditEvent

	// 2. Expectations - using gomock expectations
	mockRepo.EXPECT().CreateAuditEvent(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)

	// Mock other dependencies that might be called
	mockSpan := tracing.NewMockSpan(ctrl)
	mockSpan.EXPECT().End().AnyTimes()
	mockSpan.EXPECT().SetAttributes(gomock.Any()).AnyTimes()
	mockSpan.EXPECT().SetStatus(gomock.Any(), gomock.Any()).AnyTimes()
	mockTracing.EXPECT().StartSpan(gomock.Any(), gomock.Any(), gomock.Any()).Return(ctx, mockSpan).AnyTimes()

	// 3. Execution
	auditEvent, err := auditService.CreateAuditEvent(ctx, CreateAuditEventRequest{
		EventType:     event.EventType,
		EventCategory: event.EventCategory,
		Severity:      event.Severity,
		Decision:      event.Decision,
		Reason:        event.Reason,
		Context:       event.Context,
	})

	// 4. Assertions
	assert.NoError(t, err)
	assert.NotNil(t, auditEvent)
}
