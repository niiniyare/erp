package audit

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/niiniyare/erp/internal/shared/token"
)

// MockRepository is a mock implementation of the Repository interface
func TestRecord(t *testing.T) {
	// 1. Setup
	mockRepo := new(MockRepository)
	auditService := NewService(mockRepo)

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
		Decision:      "SUCCESS",
		Reason:        "User successfully logged in",
		Context:       json.RawMessage(`{"ip": "127.0.0.1"}`),
	}

	// 2. Expectations
	expectedEvent := event
	expectedEvent.UserID = userID // The service should set this
	mockRepo.On("CreateAuditEvent", ctx, expectedEvent).Return(nil)

	// 3. Execution
	err := auditService.Record(ctx, event)

	// 4. Assertions
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
