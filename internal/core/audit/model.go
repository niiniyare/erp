package audit

import (
	"time"

	"github.com/google/uuid"
)

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	AuditEventAccessRequestCreated  AuditEventType = "ACCESS_REQUEST_CREATED"
	AuditEventAccessRequestApproved AuditEventType = "ACCESS_REQUEST_APPROVED"
	AuditEventAccessRequestRejected AuditEventType = "ACCESS_REQUEST_REJECTED"
	AuditEventAccessRequestRevoked  AuditEventType = "ACCESS_REQUEST_REVOKED"
	AuditEventAccessRequestExpired  AuditEventType = "ACCESS_REQUEST_EXPIRED"
	AuditEventAccessGranted         AuditEventType = "ACCESS_GRANTED"
	AuditEventAccessRevoked         AuditEventType = "ACCESS_REVOKED"
	AuditEventPermissionEvaluation  AuditEventType = "PERMISSION_EVALUATION"
	AuditEventRoleAssigned          AuditEventType = "ROLE_ASSIGNED"
	AuditEventRoleRevoked           AuditEventType = "ROLE_REVOKED"
	AuditEventPrivilegeElevation    AuditEventType = "PRIVILEGE_ELEVATION"
	AuditEventSecurityViolation     AuditEventType = "SECURITY_VIOLATION"
	AuditEventWorkflowStateChange   AuditEventType = "WORKFLOW_STATE_CHANGE"
)

// AuditEventCategory represents the category of audit event
type AuditEventCategory string

const (
	AuditCategoryAccess     AuditEventCategory = "ACCESS"
	AuditCategoryAdmin      AuditEventCategory = "ADMIN"
	AuditCategoryData       AuditEventCategory = "DATA"
	AuditCategoryAuth       AuditEventCategory = "AUTH"
	AuditCategorySystem     AuditEventCategory = "SYSTEM"
	AuditCategoryCompliance AuditEventCategory = "COMPLIANCE"
	AuditCategoryWorkflow   AuditEventCategory = "WORKFLOW"
)

// AuditEventSeverity represents the severity level of an audit event
type AuditEventSeverity string

const (
	AuditSeverityLow      AuditEventSeverity = "LOW"
	AuditSeverityInfo     AuditEventSeverity = "INFO"
	AuditSeverityWarn     AuditEventSeverity = "WARN"
	AuditSeverityHigh     AuditEventSeverity = "HIGH"
	AuditSeverityCritical AuditEventSeverity = "CRITICAL"
)

// AuditEvent represents a comprehensive audit event
type AuditEvent struct {
	ID              uuid.UUID          `json:"id"`
	TenantID        uuid.UUID          `json:"tenant_id"`
	EventType       AuditEventType     `json:"event_type"`
	EventCategory   AuditEventCategory `json:"event_category"`
	Severity        AuditEventSeverity `json:"severity"`
	UserID          *uuid.UUID         `json:"user_id,omitempty"`
	TargetUserID    *uuid.UUID         `json:"target_user_id,omitempty"`
	EntityID        *uuid.UUID         `json:"entity_id,omitempty"`
	ResourceID      *uuid.UUID         `json:"resource_id,omitempty"`
	ActionID        *uuid.UUID         `json:"action_id,omitempty"`
	RoleID          *uuid.UUID         `json:"role_id,omitempty"`
	PermissionID    *uuid.UUID         `json:"permission_id,omitempty"`
	RequestID       *uuid.UUID         `json:"request_id,omitempty"`
	SessionID       *uuid.UUID         `json:"session_id,omitempty"`
	Decision        *string            `json:"decision,omitempty"` // ALLOW/DENY
	Reason          string             `json:"reason"`
	RiskScore       int                `json:"risk_score"` // 0-100
	Context         map[string]any     `json:"context"`
	IPAddress       string             `json:"ip_address"`
	UserAgent       string             `json:"user_agent"`
	ComplianceFlags []string           `json:"compliance_flags"` // GDPR, SOX, HIPAA, etc.
	Timestamp       time.Time          `json:"timestamp"`
	Metadata        map[string]any     `json:"metadata,omitempty"`
}

// WorkflowAuditEvent represents workflow-specific audit information
type WorkflowAuditEvent struct {
	RequestID      uuid.UUID      `json:"request_id"`
	EventType      string         `json:"event_type"`
	ActorID        *uuid.UUID     `json:"actor_id,omitempty"`
	PreviousState  *string        `json:"previous_state,omitempty"` // Changed from ApprovalStatus
	NewState       *string        `json:"new_state,omitempty"`      // Changed from ApprovalStatus
	Comments       string         `json:"comments"`
	Duration       *time.Duration `json:"duration,omitempty"`
	AutomatedEvent bool           `json:"automated_event"`
	Timestamp      time.Time      `json:"timestamp"`
}

// PermissionEvaluationAudit represents permission evaluation audit data
type PermissionEvaluationAudit struct {
	UserID           uuid.UUID      `json:"user_id"`
	ResourceName     string         `json:"resource_name"`
	ActionName       string         `json:"action_name"`
	Decision         string         `json:"decision"`
	PolicyDecisions  []string       `json:"policy_decisions"`
	EvaluationTimeMS int            `json:"evaluation_time_ms"`
	Context          map[string]any `json:"context"`
	RiskFactors      []string       `json:"risk_factors,omitempty"`
}

// AuditQueryRequest represents a request to query audit events
type AuditQueryRequest struct {
	TenantID       uuid.UUID            `json:"tenant_id"`
	EventTypes     []AuditEventType     `json:"event_types,omitempty"`
	Categories     []AuditEventCategory `json:"categories,omitempty"`
	Severities     []AuditEventSeverity `json:"severities,omitempty"`
	UserID         *uuid.UUID           `json:"user_id,omitempty"`
	FromDate       time.Time            `json:"from_date"`
	ToDate         time.Time            `json:"to_date"`
	Limit          int                  `json:"limit"`
	Offset         int                  `json:"offset"`
	IncludeContext bool                 `json:"include_context"`
}

// ComplianceReportRequest represents a request for compliance reporting
type ComplianceReportRequest struct {
	TenantID        uuid.UUID `json:"tenant_id"`
	ComplianceFlags []string  `json:"compliance_flags"`
	FromDate        time.Time `json:"from_date"`
	ToDate          time.Time `json:"to_date"`
	IncludeDetails  bool      `json:"include_details"`
	Format          string    `json:"format"` // JSON, CSV, PDF
}

// ComplianceReport represents a compliance audit report
type ComplianceReport struct {
	ID               uuid.UUID                `json:"id"`
	TenantID         uuid.UUID                `json:"tenant_id"`
	ComplianceFlags  []string                 `json:"compliance_flags"`
	GeneratedAt      time.Time                `json:"generated_at"`
	Period           string                   `json:"period"`
	Summary          ComplianceSummary        `json:"summary"`
	EventsByCategory map[string]int           `json:"events_by_category"`
	RiskAssessment   ComplianceRiskAssessment `json:"risk_assessment"`
	Violations       []ComplianceViolation    `json:"violations,omitempty"`
}

// ComplianceSummary represents summary statistics for compliance
type ComplianceSummary struct {
	TotalEvents      int `json:"total_events"`
	AccessRequests   int `json:"access_requests"`
	AccessViolations int `json:"access_violations"`
	ElevatedAccess   int `json:"elevated_access"`
	PolicyViolations int `json:"policy_violations"`
}

// ComplianceRiskAssessment represents risk assessment for compliance
type ComplianceRiskAssessment struct {
	OverallRiskScore int      `json:"overall_risk_score"`
	RiskFactors      []string `json:"risk_factors"`
	Recommendations  []string `json:"recommendations"`
}

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	ID            uuid.UUID `json:"id"`
	EventID       uuid.UUID `json:"event_id"`
	ViolationType string    `json:"violation_type"`
	Severity      string    `json:"severity"`
	Description   string    `json:"description"`
	UserID        uuid.UUID `json:"user_id"`
	Timestamp     time.Time `json:"timestamp"`
}

// AuditExportRequest represents a request to export audit data
type AuditExportRequest struct {
	Query    AuditQueryRequest `json:"query"`
	Format   string            `json:"format"` // JSON, CSV, XML
	Compress bool              `json:"compress"`
}
