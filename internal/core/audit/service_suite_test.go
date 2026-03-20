//go:build unit
// +build unit

package audit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"awo/internal/shared/token"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockAuditRepository implements the Repository interface for testing
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) CreateAuditEvent(ctx context.Context, event AuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// AuditServiceTestSuite is the main test suite for the audit service
type AuditServiceTestSuite struct {
	suite.Suite
	service      Service
	mockRepo     *MockAuditRepository
	ctx          context.Context
	testUserID   uuid.UUID
	testEntityID uuid.UUID
	testPayload  *token.Payload
}

func (suite *AuditServiceTestSuite) SetupSuite() {
	// Suite-level setup
	suite.testUserID = uuid.New()
	suite.testEntityID = uuid.New()

	suite.testPayload = &token.Payload{
		UserID: suite.testUserID,
	}
}

func (suite *AuditServiceTestSuite) SetupTest() {
	// Test-level setup
	suite.mockRepo = new(MockAuditRepository)
	suite.service = NewService(suite.mockRepo)

	// Create context with auth payload
	suite.ctx = context.WithValue(context.Background(), token.AuthorizationPayloadKey, suite.testPayload)
}

func (suite *AuditServiceTestSuite) TearDownTest() {
	// Test-level cleanup
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test Record - Success Case
func (suite *AuditServiceTestSuite) TestRecord_Success() {
	// Arrange
	contextData := map[string]any{
		"action": "create_user",
		"target": "user123",
	}
	contextJSON, _ := json.Marshal(contextData)

	event := AuditEvent{
		EventType:     "USER_CREATED",
		EventCategory: "USER_MANAGEMENT",
		Severity:      "INFO",
		EntityID:      uuid.NullUUID{UUID: suite.testEntityID, Valid: true},
		Decision:      "ALLOWED",
		Reason:        "User creation successful",
		Context:       contextJSON,
	}

	// The service should set UserID from the context
	expectedEvent := event
	expectedEvent.UserID = suite.testUserID

	suite.mockRepo.On("CreateAuditEvent", suite.ctx, expectedEvent).Return(nil)

	// Act
	err := suite.service.Record(suite.ctx, event)

	// Assert
	suite.NoError(err)
}

// Test Record - Repository Error
func (suite *AuditServiceTestSuite) TestRecord_RepositoryError() {
	// Arrange
	event := AuditEvent{
		EventType:     "USER_LOGIN",
		EventCategory: "AUTHENTICATION",
		Severity:      "INFO",
		Decision:      "ALLOWED",
		Reason:        "Valid credentials",
	}

	expectedEvent := event
	expectedEvent.UserID = suite.testUserID

	suite.mockRepo.On("CreateAuditEvent", suite.ctx, expectedEvent).Return(errors.New("database connection failed"))

	// Act
	err := suite.service.Record(suite.ctx, event)

	// Assert
	suite.Error(err)
	suite.Contains(err.Error(), "database connection failed")
}

// Test Record - User Authentication Event
func (suite *AuditServiceTestSuite) TestRecord_AuthenticationEvent() {
	// Arrange
	contextData := map[string]any{
		"ip_address":   "192.168.1.100",
		"user_agent":   "Mozilla/5.0",
		"login_method": "password",
	}
	contextJSON, _ := json.Marshal(contextData)

	event := AuditEvent{
		EventType:     "USER_LOGIN_SUCCESS",
		EventCategory: "AUTHENTICATION",
		Severity:      "INFO",
		Decision:      "ALLOWED",
		Reason:        "Valid username and password",
		Context:       contextJSON,
	}

	expectedEvent := event
	expectedEvent.UserID = suite.testUserID

	suite.mockRepo.On("CreateAuditEvent", suite.ctx, expectedEvent).Return(nil)

	// Act
	err := suite.service.Record(suite.ctx, event)

	// Assert
	suite.NoError(err)
}

// Test Record - Authorization Event
func (suite *AuditServiceTestSuite) TestRecord_AuthorizationEvent() {
	// Arrange
	contextData := map[string]any{
		"resource":   "/api/users",
		"action":     "READ",
		"permission": "users.read",
	}
	contextJSON, _ := json.Marshal(contextData)

	event := AuditEvent{
		EventType:     "PERMISSION_CHECK",
		EventCategory: "AUTHORIZATION",
		Severity:      "DEBUG",
		EntityID:      uuid.NullUUID{UUID: suite.testEntityID, Valid: true},
		Decision:      "ALLOWED",
		Reason:        "User has required permission",
		Context:       contextJSON,
	}

	expectedEvent := event
	expectedEvent.UserID = suite.testUserID

	suite.mockRepo.On("CreateAuditEvent", suite.ctx, expectedEvent).Return(nil)

	// Act
	err := suite.service.Record(suite.ctx, event)

	// Assert
	suite.NoError(err)
}

// Test Record - Access Request Event
func (suite *AuditServiceTestSuite) TestRecord_AccessRequestEvent() {
	// Arrange
	requestID := uuid.New()
	contextData := map[string]any{
		"request_id":   requestID.String(),
		"request_type": "ROLE_ASSIGNMENT",
		"target_user":  uuid.New().String(),
		"role_name":    "admin",
	}
	contextJSON, _ := json.Marshal(contextData)

	event := AuditEvent{
		EventType:     "ACCESS_REQUEST_APPROVED",
		EventCategory: "ACCESS_MANAGEMENT",
		Severity:      "INFO",
		EntityID:      uuid.NullUUID{UUID: suite.testEntityID, Valid: true},
		Decision:      "APPROVED",
		Reason:        "Request meets approval criteria",
		Context:       contextJSON,
	}

	expectedEvent := event
	expectedEvent.UserID = suite.testUserID

	suite.mockRepo.On("CreateAuditEvent", suite.ctx, expectedEvent).Return(nil)

	// Act
	err := suite.service.Record(suite.ctx, event)

	// Assert
	suite.NoError(err)
}

// Test Record - Security Event
func (suite *AuditServiceTestSuite) TestRecord_SecurityEvent() {
	// Arrange
	contextData := map[string]any{
		"ip_address":     "192.168.1.100",
		"attempt_count":  3,
		"blocked_until":  "2024-01-01T12:00:00Z",
		"trigger_reason": "multiple_failed_logins",
	}
	contextJSON, _ := json.Marshal(contextData)

	event := AuditEvent{
		EventType:     "ACCOUNT_LOCKED",
		EventCategory: "SECURITY",
		Severity:      "WARNING",
		Decision:      "BLOCKED",
		Reason:        "Multiple failed login attempts",
		Context:       contextJSON,
	}

	expectedEvent := event
	expectedEvent.UserID = suite.testUserID

	suite.mockRepo.On("CreateAuditEvent", suite.ctx, expectedEvent).Return(nil)

	// Act
	err := suite.service.Record(suite.ctx, event)

	// Assert
	suite.NoError(err)
}

// Test Record - Data Modification Event
func (suite *AuditServiceTestSuite) TestRecord_DataModificationEvent() {
	// Arrange
	contextData := map[string]any{
		"table_name":     "users",
		"record_id":      suite.testUserID.String(),
		"fields_changed": []string{"email", "last_name"},
		"old_values":     map[string]any{"email": "old@example.com"},
		"new_values":     map[string]any{"email": "new@example.com"},
	}
	contextJSON, _ := json.Marshal(contextData)

	event := AuditEvent{
		EventType:     "USER_DATA_UPDATED",
		EventCategory: "DATA_MODIFICATION",
		Severity:      "INFO",
		EntityID:      uuid.NullUUID{UUID: suite.testEntityID, Valid: true},
		Decision:      "ALLOWED",
		Reason:        "User profile update",
		Context:       contextJSON,
	}

	expectedEvent := event
	expectedEvent.UserID = suite.testUserID

	suite.mockRepo.On("CreateAuditEvent", suite.ctx, expectedEvent).Return(nil)

	// Act
	err := suite.service.Record(suite.ctx, event)

	// Assert
	suite.NoError(err)
}

// Test Record - Error Event
func (suite *AuditServiceTestSuite) TestRecord_ErrorEvent() {
	// Arrange
	contextData := map[string]any{
		"error_message": "Database connection timeout",
		"operation":     "get_user_permissions",
		"duration_ms":   5000,
		"retry_count":   3,
	}
	contextJSON, _ := json.Marshal(contextData)

	event := AuditEvent{
		EventType:     "SYSTEM_ERROR",
		EventCategory: "SYSTEM",
		Severity:      "ERROR",
		Decision:      "DENIED",
		Reason:        "System unavailable",
		Context:       contextJSON,
	}

	expectedEvent := event
	expectedEvent.UserID = suite.testUserID

	suite.mockRepo.On("CreateAuditEvent", suite.ctx, expectedEvent).Return(nil)

	// Act
	err := suite.service.Record(suite.ctx, event)

	// Assert
	suite.NoError(err)
}

// Test Record - Compliance Event
func (suite *AuditServiceTestSuite) TestRecord_ComplianceEvent() {
	// Arrange
	contextData := map[string]any{
		"regulation":         "GDPR",
		"data_subject":       "user123@example.com",
		"request_type":       "data_deletion",
		"processing_time":    "24h",
		"compliance_officer": "officer@company.com",
	}
	contextJSON, _ := json.Marshal(contextData)

	event := AuditEvent{
		EventType:     "DATA_DELETION_REQUEST",
		EventCategory: "COMPLIANCE",
		Severity:      "INFO",
		EntityID:      uuid.NullUUID{UUID: suite.testEntityID, Valid: true},
		Decision:      "APPROVED",
		Reason:        "GDPR Article 17 - Right to erasure",
		Context:       contextJSON,
	}

	expectedEvent := event
	expectedEvent.UserID = suite.testUserID

	suite.mockRepo.On("CreateAuditEvent", suite.ctx, expectedEvent).Return(nil)

	// Act
	err := suite.service.Record(suite.ctx, event)

	// Assert
	suite.NoError(err)
}

// Test Record - With Empty Context
func (suite *AuditServiceTestSuite) TestRecord_EmptyContext() {
	// Arrange
	event := AuditEvent{
		EventType:     "SIMPLE_EVENT",
		EventCategory: "GENERAL",
		Severity:      "INFO",
		Decision:      "ALLOWED",
		Reason:        "Simple event without context",
		Context:       json.RawMessage("{}"),
	}

	expectedEvent := event
	expectedEvent.UserID = suite.testUserID

	suite.mockRepo.On("CreateAuditEvent", suite.ctx, expectedEvent).Return(nil)

	// Act
	err := suite.service.Record(suite.ctx, event)

	// Assert
	suite.NoError(err)
}

// Test Record - With Null Entity ID
func (suite *AuditServiceTestSuite) TestRecord_NullEntityID() {
	// Arrange
	event := AuditEvent{
		EventType:     "GLOBAL_EVENT",
		EventCategory: "SYSTEM",
		Severity:      "INFO",
		EntityID:      uuid.NullUUID{Valid: false}, // Null entity ID
		Decision:      "ALLOWED",
		Reason:        "Global system event",
		Context:       json.RawMessage("{}"),
	}

	expectedEvent := event
	expectedEvent.UserID = suite.testUserID

	suite.mockRepo.On("CreateAuditEvent", suite.ctx, expectedEvent).Return(nil)

	// Act
	err := suite.service.Record(suite.ctx, event)

	// Assert
	suite.NoError(err)
}

// TestAuditService runs the test suite
func TestAuditService(t *testing.T) {
	suite.Run(t, new(AuditServiceTestSuite))
}
