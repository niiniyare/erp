package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	AuditEventAccessRequestCreated    AuditEventType = "ACCESS_REQUEST_CREATED"
	AuditEventAccessRequestApproved   AuditEventType = "ACCESS_REQUEST_APPROVED"
	AuditEventAccessRequestRejected   AuditEventType = "ACCESS_REQUEST_REJECTED"
	AuditEventAccessRequestRevoked    AuditEventType = "ACCESS_REQUEST_REVOKED"
	AuditEventAccessRequestExpired    AuditEventType = "ACCESS_REQUEST_EXPIRED"
	AuditEventAccessGranted           AuditEventType = "ACCESS_GRANTED"
	AuditEventAccessRevoked           AuditEventType = "ACCESS_REVOKED"
	AuditEventPermissionEvaluation    AuditEventType = "PERMISSION_EVALUATION"
	AuditEventRoleAssigned            AuditEventType = "ROLE_ASSIGNED"
	AuditEventRoleRevoked             AuditEventType = "ROLE_REVOKED"
	AuditEventPrivilegeElevation      AuditEventType = "PRIVILEGE_ELEVATION"
	AuditEventSecurityViolation       AuditEventType = "SECURITY_VIOLATION"
	AuditEventWorkflowStateChange     AuditEventType = "WORKFLOW_STATE_CHANGE"
)

// AuditEventCategory represents the category of audit event
type AuditEventCategory string

const (
	AuditCategoryAccess      AuditEventCategory = "ACCESS"
	AuditCategoryAdmin       AuditEventCategory = "ADMIN"
	AuditCategoryData        AuditEventCategory = "DATA"
	AuditCategoryAuth        AuditEventCategory = "AUTH"
	AuditCategorySystem      AuditEventCategory = "SYSTEM"
	AuditCategoryCompliance  AuditEventCategory = "COMPLIANCE"
	AuditCategoryWorkflow    AuditEventCategory = "WORKFLOW"
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
	ID              uuid.UUID                  `json:"id"`
	TenantID        uuid.UUID                  `json:"tenant_id"`
	EventType       AuditEventType             `json:"event_type"`
	EventCategory   AuditEventCategory         `json:"event_category"`
	Severity        AuditEventSeverity         `json:"severity"`
	UserID          *uuid.UUID                 `json:"user_id,omitempty"`
	TargetUserID    *uuid.UUID                 `json:"target_user_id,omitempty"`
	EntityID        *uuid.UUID                 `json:"entity_id,omitempty"`
	ResourceID      *uuid.UUID                 `json:"resource_id,omitempty"`
	ActionID        *uuid.UUID                 `json:"action_id,omitempty"`
	RoleID          *uuid.UUID                 `json:"role_id,omitempty"`
	PermissionID    *uuid.UUID                 `json:"permission_id,omitempty"`
	RequestID       *uuid.UUID                 `json:"request_id,omitempty"`
	SessionID       *uuid.UUID                 `json:"session_id,omitempty"`
	Decision        *string                    `json:"decision,omitempty"` // ALLOW/DENY
	Reason          string                     `json:"reason"`
	RiskScore       int                        `json:"risk_score"`         // 0-100
	Context         map[string]any     `json:"context"`
	IPAddress       string                     `json:"ip_address"`
	UserAgent       string                     `json:"user_agent"`
	ComplianceFlags []string                   `json:"compliance_flags"`  // GDPR, SOX, HIPAA, etc.
	Timestamp       time.Time                  `json:"timestamp"`
	Metadata        map[string]any     `json:"metadata,omitempty"`
}

// WorkflowAuditEvent represents workflow-specific audit information
type WorkflowAuditEvent struct {
	RequestID      uuid.UUID      `json:"request_id"`
	EventType      string         `json:"event_type"`
	ActorID        *uuid.UUID     `json:"actor_id,omitempty"`
	PreviousState  *ApprovalStatus `json:"previous_state,omitempty"`
	NewState       *ApprovalStatus `json:"new_state,omitempty"`
	Comments       string         `json:"comments"`
	Duration       *time.Duration `json:"duration,omitempty"`
	AutomatedEvent bool           `json:"automated_event"`
	Timestamp      time.Time      `json:"timestamp"`
}

// PermissionEvaluationAudit represents permission evaluation audit data
type PermissionEvaluationAudit struct {
	UserID           uuid.UUID              `json:"user_id"`
	ResourceName     string                 `json:"resource_name"`
	ActionName       string                 `json:"action_name"`
	Decision         string                 `json:"decision"`
	PolicyDecisions  []string               `json:"policy_decisions"`
	EvaluationTimeMS int                    `json:"evaluation_time_ms"`
	Context          map[string]any `json:"context"`
	RiskFactors      []string               `json:"risk_factors,omitempty"`
}

// AuditService handles all audit logging for the user management system
type AuditService interface {
	// Core audit operations
	LogAuditEvent(ctx context.Context, event *AuditEvent) error
	LogWorkflowEvent(ctx context.Context, workflowEvent *WorkflowAuditEvent) error
	LogPermissionEvaluation(ctx context.Context, evaluation *PermissionEvaluationAudit) error
	
	// Specific audit events
	LogAccessRequestCreated(ctx context.Context, request *AccessRequest, requesterID uuid.UUID) error
	LogAccessRequestProcessed(ctx context.Context, request *AccessRequest, action string, approverID uuid.UUID, comments string) error
	LogAccessGranted(ctx context.Context, request *AccessRequest, grantedAccess []TemporaryAccess) error
	LogAccessRevoked(ctx context.Context, request *AccessRequest, revokedAccess []uuid.UUID) error
	LogSecurityViolation(ctx context.Context, userID uuid.UUID, violation string, severity AuditEventSeverity, context map[string]any) error
	
	// Query operations
	GetAuditEvents(ctx context.Context, req *AuditQueryRequest) ([]*AuditEvent, error)
	GetUserAuditTrail(ctx context.Context, userID uuid.UUID, fromDate, toDate time.Time) ([]*AuditEvent, error)
	GetWorkflowAuditTrail(ctx context.Context, requestID uuid.UUID) ([]*WorkflowAuditEvent, error)
	
	// Compliance operations
	GenerateComplianceReport(ctx context.Context, req *ComplianceReportRequest) (*ComplianceReport, error)
	ExportAuditData(ctx context.Context, req *AuditExportRequest) ([]byte, error)
}

// AuditQueryRequest represents a request to query audit events
type AuditQueryRequest struct {
	TenantID       uuid.UUID             `json:"tenant_id"`
	EventTypes     []AuditEventType      `json:"event_types,omitempty"`
	Categories     []AuditEventCategory  `json:"categories,omitempty"`
	Severities     []AuditEventSeverity  `json:"severities,omitempty"`
	UserID         *uuid.UUID            `json:"user_id,omitempty"`
	FromDate       time.Time             `json:"from_date"`
	ToDate         time.Time             `json:"to_date"`
	Limit          int                   `json:"limit"`
	Offset         int                   `json:"offset"`
	IncludeContext bool                  `json:"include_context"`
}

// ComplianceReportRequest represents a request for compliance reporting
type ComplianceReportRequest struct {
	TenantID         uuid.UUID  `json:"tenant_id"`
	ComplianceFlags  []string   `json:"compliance_flags"`
	FromDate         time.Time  `json:"from_date"`
	ToDate           time.Time  `json:"to_date"`
	IncludeDetails   bool       `json:"include_details"`
	Format           string     `json:"format"` // JSON, CSV, PDF
}

// ComplianceReport represents a compliance audit report
type ComplianceReport struct {
	ID               uuid.UUID                  `json:"id"`
	TenantID         uuid.UUID                  `json:"tenant_id"`
	ComplianceFlags  []string                   `json:"compliance_flags"`
	GeneratedAt      time.Time                  `json:"generated_at"`
	Period           string                     `json:"period"`
	Summary          ComplianceSummary          `json:"summary"`
	EventsByCategory map[string]int             `json:"events_by_category"`
	RiskAssessment   ComplianceRiskAssessment   `json:"risk_assessment"`
	Violations       []ComplianceViolation      `json:"violations,omitempty"`
}

// ComplianceSummary represents summary statistics for compliance
type ComplianceSummary struct {
	TotalEvents        int `json:"total_events"`
	AccessRequests     int `json:"access_requests"`
	AccessViolations   int `json:"access_violations"`
	ElevatedAccess     int `json:"elevated_access"`
	PolicyViolations   int `json:"policy_violations"`
}

// ComplianceRiskAssessment represents risk assessment for compliance
type ComplianceRiskAssessment struct {
	OverallRiskScore int                `json:"overall_risk_score"`
	RiskFactors      []string           `json:"risk_factors"`
	Recommendations  []string           `json:"recommendations"`
}

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	ID          uuid.UUID `json:"id"`
	EventID     uuid.UUID `json:"event_id"`
	ViolationType string  `json:"violation_type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	UserID      uuid.UUID `json:"user_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// AuditExportRequest represents a request to export audit data
type AuditExportRequest struct {
	Query    AuditQueryRequest `json:"query"`
	Format   string            `json:"format"` // JSON, CSV, XML
	Compress bool              `json:"compress"`
}

// auditService implements AuditService
type auditService struct {
	tracing       *tracing.TracingService
	metrics       *metrics.MetricsService
	// TODO: Add audit repository when implemented
	// auditRepo     AuditRepository
}

// NewAuditService creates a new audit service
func NewAuditService(
	tracing *tracing.TracingService,
	metrics *metrics.MetricsService,
) AuditService {
	return &auditService{
		tracing: tracing,
		metrics: metrics,
	}
}

// LogAuditEvent logs a comprehensive audit event
func (s *auditService) LogAuditEvent(ctx context.Context, event *AuditEvent) error {
	ctx, span := s.tracing.StartSpan(ctx, "auditService.LogAuditEvent")
	defer span.End()

	span.SetAttributes(
		attribute.String("event_type", string(event.EventType)),
		attribute.String("event_category", string(event.EventCategory)),
		attribute.String("severity", string(event.Severity)),
	)

	// Ensure required fields are set
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Calculate risk score if not provided
	if event.RiskScore == 0 {
		event.RiskScore = s.calculateRiskScore(event)
	}

	// TODO: Store in database when audit repository is available
	// For now, log to structured logger
	auditData, _ := json.Marshal(event)
	
	logger.Info("Audit event logged", logger.Fields{
		"audit_event_id": event.ID,
		"event_type":     event.EventType,
		"category":       event.EventCategory,
		"severity":       event.Severity,
		"user_id":        event.UserID,
		"risk_score":     event.RiskScore,
		"audit_data":     string(auditData),
	})

	s.metrics.IncrementCounter("audit_event_logged", map[string]any{
		"type": string(event.EventType),
		"category": string(event.EventCategory),
		"severity": string(event.Severity),
	})

	return nil
}

// LogWorkflowEvent logs a workflow-specific audit event
func (s *auditService) LogWorkflowEvent(ctx context.Context, workflowEvent *WorkflowAuditEvent) error {
	ctx, span := s.tracing.StartSpan(ctx, "auditService.LogWorkflowEvent")
	defer span.End()

	// Create comprehensive audit event
	auditEvent := &AuditEvent{
		ID:            uuid.New(),
		EventType:     AuditEventWorkflowStateChange,
		EventCategory: AuditCategoryWorkflow,
		Severity:      AuditSeverityInfo,
		UserID:        workflowEvent.ActorID,
		RequestID:     &workflowEvent.RequestID,
		Reason:        fmt.Sprintf("Workflow %s: %s", workflowEvent.EventType, workflowEvent.Comments),
		Context: map[string]any{
			"event_type":     workflowEvent.EventType,
			"previous_state": workflowEvent.PreviousState,
			"new_state":      workflowEvent.NewState,
			"automated":      workflowEvent.AutomatedEvent,
			"duration_ms":    workflowEvent.Duration,
		},
		Timestamp: workflowEvent.Timestamp,
	}

	return s.LogAuditEvent(ctx, auditEvent)
}

// LogPermissionEvaluation logs permission evaluation for audit purposes
func (s *auditService) LogPermissionEvaluation(ctx context.Context, evaluation *PermissionEvaluationAudit) error {
	ctx, span := s.tracing.StartSpan(ctx, "auditService.LogPermissionEvaluation")
	defer span.End()

	// Only log DENY decisions and high-risk evaluations for performance
	shouldLog := evaluation.Decision == "DENY" || 
				 len(evaluation.RiskFactors) > 0 ||
				 evaluation.EvaluationTimeMS > 1000 // Slow evaluations

	if !shouldLog {
		return nil
	}

	severity := AuditSeverityInfo
	if evaluation.Decision == "DENY" {
		severity = AuditSeverityWarn
	}
	if len(evaluation.RiskFactors) > 0 {
		severity = AuditSeverityHigh
	}

	auditEvent := &AuditEvent{
		ID:            uuid.New(),
		EventType:     AuditEventPermissionEvaluation,
		EventCategory: AuditCategoryAuth,
		Severity:      severity,
		UserID:        &evaluation.UserID,
		Decision:      &evaluation.Decision,
		Reason:        fmt.Sprintf("Permission evaluation for %s on %s", evaluation.ActionName, evaluation.ResourceName),
		Context: map[string]any{
			"resource_name":      evaluation.ResourceName,
			"action_name":        evaluation.ActionName,
			"policy_decisions":   evaluation.PolicyDecisions,
			"evaluation_time_ms": evaluation.EvaluationTimeMS,
			"risk_factors":       evaluation.RiskFactors,
			"context":            evaluation.Context,
		},
		Timestamp: time.Now(),
	}

	return s.LogAuditEvent(ctx, auditEvent)
}

// LogAccessRequestCreated logs when an access request is created
func (s *auditService) LogAccessRequestCreated(ctx context.Context, request *AccessRequest, requesterID uuid.UUID) error {
	auditEvent := &AuditEvent{
		ID:            uuid.New(),
		TenantID:      request.TenantID,
		EventType:     AuditEventAccessRequestCreated,
		EventCategory: AuditCategoryAccess,
		Severity:      AuditSeverityInfo,
		UserID:        &requesterID,
		TargetUserID:  request.TargetUserID,
		EntityID:      &request.EntityID,
		ResourceID:    request.ResourceID,
		RoleID:        request.RoleID,
		PermissionID:  request.PermissionID,
		RequestID:     &request.ID,
		Reason:        fmt.Sprintf("Access request created: %s", request.RequestType),
		Context: map[string]any{
			"request_type":     request.RequestType,
			"justification":    request.Justification,
			"business_reason":  request.BusinessReason,
			"duration_hours":   request.DurationHours,
			"auto_revoke":      request.AutoRevoke,
		},
		ComplianceFlags: []string{"ACCESS_CONTROL", "SEGREGATION_OF_DUTIES"},
		Timestamp:       request.CreatedAt,
	}

	return s.LogAuditEvent(ctx, auditEvent)
}

// LogAccessRequestProcessed logs when an access request is approved or rejected
func (s *auditService) LogAccessRequestProcessed(ctx context.Context, request *AccessRequest, action string, approverID uuid.UUID, comments string) error {
	var eventType AuditEventType
	var severity AuditEventSeverity

	switch action {
	case "approve", "approved":
		eventType = AuditEventAccessRequestApproved
		severity = AuditSeverityInfo
	case "reject", "rejected":
		eventType = AuditEventAccessRequestRejected
		severity = AuditSeverityWarn
	case "revoke", "revoked":
		eventType = AuditEventAccessRequestRevoked
		severity = AuditSeverityHigh
	case "expire", "expired":
		eventType = AuditEventAccessRequestExpired
		severity = AuditSeverityInfo
	default:
		eventType = AuditEventWorkflowStateChange
		severity = AuditSeverityInfo
	}

	auditEvent := &AuditEvent{
		ID:            uuid.New(),
		TenantID:      request.TenantID,
		EventType:     eventType,
		EventCategory: AuditCategoryAccess,
		Severity:      severity,
		UserID:        &approverID,
		TargetUserID:  request.TargetUserID,
		EntityID:      &request.EntityID,
		RequestID:     &request.ID,
		Reason:        fmt.Sprintf("Access request %s: %s", action, comments),
		Context: map[string]any{
			"action":           action,
			"request_type":     request.RequestType,
			"approval_status":  request.ApprovalStatus,
			"comments":         comments,
			"requester_id":     request.RequesterID,
		},
		ComplianceFlags: []string{"ACCESS_CONTROL", "APPROVAL_PROCESS"},
		Timestamp:       time.Now(),
	}

	return s.LogAuditEvent(ctx, auditEvent)
}

// LogAccessGranted logs when access is actually granted
func (s *auditService) LogAccessGranted(ctx context.Context, request *AccessRequest, grantedAccess []TemporaryAccess) error {
	for _, access := range grantedAccess {
		auditEvent := &AuditEvent{
			ID:            uuid.New(),
			TenantID:      request.TenantID,
			EventType:     AuditEventAccessGranted,
			EventCategory: AuditCategoryAccess,
			Severity:      AuditSeverityInfo,
			UserID:        &access.GrantedBy,
			TargetUserID:  &access.UserID,
			EntityID:      &access.EntityID,
			ResourceID:    access.ResourceID,
			RoleID:        access.RoleID,
			PermissionID:  access.PermissionID,
			RequestID:     &request.ID,
			Reason:        fmt.Sprintf("Access granted: %s", access.AccessType),
			Context: map[string]any{
				"access_type":    access.AccessType,
				"access_id":      access.ID,
				"granted_at":     access.GrantedAt,
				"expires_at":     access.ExpiresAt,
				"metadata":       access.Metadata,
			},
			ComplianceFlags: []string{"ACCESS_CONTROL", "PRIVILEGED_ACCESS"},
			Timestamp:       access.GrantedAt,
		}

		// Higher severity for privilege elevation
		if access.AccessType == RequestTypeElevation {
			auditEvent.Severity = AuditSeverityHigh
			auditEvent.ComplianceFlags = append(auditEvent.ComplianceFlags, "PRIVILEGE_ELEVATION")
		}

		if err := s.LogAuditEvent(ctx, auditEvent); err != nil {
			logger.Error("Failed to log access granted event", logger.Fields{
				"access_id": access.ID,
				"error":     err.Error(),
			})
		}
	}

	return nil
}

// LogAccessRevoked logs when access is revoked
func (s *auditService) LogAccessRevoked(ctx context.Context, request *AccessRequest, revokedAccess []uuid.UUID) error {
	for _, accessID := range revokedAccess {
		auditEvent := &AuditEvent{
			ID:            uuid.New(),
			TenantID:      request.TenantID,
			EventType:     AuditEventAccessRevoked,
			EventCategory: AuditCategoryAccess,
			Severity:      AuditSeverityInfo,
			TargetUserID:  request.TargetUserID,
			EntityID:      &request.EntityID,
			RequestID:     &request.ID,
			Reason:        fmt.Sprintf("Access revoked: %s", request.RequestType),
			Context: map[string]any{
				"access_id":      accessID,
				"request_type":   request.RequestType,
				"revoked_at":     time.Now(),
			},
			ComplianceFlags: []string{"ACCESS_CONTROL", "ACCESS_REVOCATION"},
			Timestamp:       time.Now(),
		}

		if err := s.LogAuditEvent(ctx, auditEvent); err != nil {
			logger.Error("Failed to log access revoked event", logger.Fields{
				"access_id": accessID,
				"error":     err.Error(),
			})
		}
	}

	return nil
}

// LogSecurityViolation logs security violations
func (s *auditService) LogSecurityViolation(ctx context.Context, userID uuid.UUID, violation string, severity AuditEventSeverity, context map[string]any) error {
	auditEvent := &AuditEvent{
		ID:            uuid.New(),
		EventType:     AuditEventSecurityViolation,
		EventCategory: AuditCategoryAuth,
		Severity:      severity,
		UserID:        &userID,
		Reason:        violation,
		RiskScore:     s.getViolationRiskScore(violation, severity),
		Context:       context,
		ComplianceFlags: []string{"SECURITY_VIOLATION", "INCIDENT_RESPONSE"},
		Timestamp:     time.Now(),
	}

	return s.LogAuditEvent(ctx, auditEvent)
}

// GetAuditEvents retrieves audit events based on query criteria
func (s *auditService) GetAuditEvents(ctx context.Context, req *AuditQueryRequest) ([]*AuditEvent, error) {
	// TODO: Implement database query when audit repository is available
	return []*AuditEvent{}, nil
}

// GetUserAuditTrail gets audit trail for a specific user
func (s *auditService) GetUserAuditTrail(ctx context.Context, userID uuid.UUID, fromDate, toDate time.Time) ([]*AuditEvent, error) {
	// TODO: Implement user-specific audit trail query
	return []*AuditEvent{}, nil
}

// GetWorkflowAuditTrail gets workflow audit trail for a specific request
func (s *auditService) GetWorkflowAuditTrail(ctx context.Context, requestID uuid.UUID) ([]*WorkflowAuditEvent, error) {
	// TODO: Implement workflow audit trail query
	return []*WorkflowAuditEvent{}, nil
}

// GenerateComplianceReport generates compliance audit reports
func (s *auditService) GenerateComplianceReport(ctx context.Context, req *ComplianceReportRequest) (*ComplianceReport, error) {
	// TODO: Implement compliance report generation
	report := &ComplianceReport{
		ID:              uuid.New(),
		TenantID:        req.TenantID,
		ComplianceFlags: req.ComplianceFlags,
		GeneratedAt:     time.Now(),
		Period:          fmt.Sprintf("%s to %s", req.FromDate.Format("2006-01-02"), req.ToDate.Format("2006-01-02")),
	}

	return report, nil
}

// ExportAuditData exports audit data in specified format
func (s *auditService) ExportAuditData(ctx context.Context, req *AuditExportRequest) ([]byte, error) {
	// TODO: Implement audit data export
	return []byte{}, nil
}

// Helper methods

// calculateRiskScore calculates risk score based on audit event properties
func (s *auditService) calculateRiskScore(event *AuditEvent) int {
	score := 0

	// Base score by event type
	switch event.EventType {
	case AuditEventPrivilegeElevation:
		score += 40
	case AuditEventSecurityViolation:
		score += 60
	case AuditEventAccessGranted:
		score += 20
	case AuditEventAccessRevoked:
		score += 10
	default:
		score += 5
	}

	// Severity modifier
	switch event.Severity {
	case AuditSeverityCritical:
		score += 40
	case AuditSeverityHigh:
		score += 30
	case AuditSeverityWarn:
		score += 20
	case AuditSeverityInfo:
		score += 10
	default:
		score += 5
	}

	// Context-based risk factors
	if event.Context != nil {
		if _, hasRiskFactors := event.Context["risk_factors"]; hasRiskFactors {
			score += 20
		}
		if decision, hasDecision := event.Context["decision"]; hasDecision && decision == "DENY" {
			score += 15
		}
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// getViolationRiskScore gets risk score for security violations
func (s *auditService) getViolationRiskScore(violation string, severity AuditEventSeverity) int {
	baseScore := map[AuditEventSeverity]int{
		AuditSeverityCritical: 90,
		AuditSeverityHigh:     70,
		AuditSeverityWarn:     50,
		AuditSeverityInfo:     30,
		AuditSeverityLow:      10,
	}

	return baseScore[severity]
}