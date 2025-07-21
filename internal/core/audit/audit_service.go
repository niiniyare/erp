package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// AuditService handles all audit logging for the user management system
type AuditService interface {
	// Core audit operations
	LogAuditEvent(ctx context.Context, event *AuditEvent) error
	LogWorkflowEvent(ctx context.Context, workflowEvent *WorkflowAuditEvent) error
	LogPermissionEvaluation(ctx context.Context, evaluation *PermissionEvaluationAudit) error

	// Specific audit events
	LogAccessRequestCreated(ctx context.Context, tenantID, requesterID uuid.UUID, targetUserID *uuid.UUID, entityID uuid.UUID, requestID uuid.UUID, requestType string, justification string, durationHours *int32) error
	LogAccessRequestProcessed(ctx context.Context, tenantID, requestID uuid.UUID, approverID uuid.UUID, targetUserID *uuid.UUID, entityID uuid.UUID, action, comments, requestType, approvalStatus string) error
	LogAccessGranted(ctx context.Context, tenantID, requestID uuid.UUID, grantedBy, userID, entityID uuid.UUID, accessType string, accessID uuid.UUID, grantedAt time.Time, expiresAt *time.Time) error
	LogAccessRevoked(ctx context.Context, tenantID, requestID uuid.UUID, targetUserID *uuid.UUID, entityID uuid.UUID, requestType string, accessID uuid.UUID) error
	LogSecurityViolation(ctx context.Context, userID uuid.UUID, violation string, severity AuditEventSeverity, context map[string]any) error

	// Query operations
	GetAuditEvents(ctx context.Context, req *AuditQueryRequest) ([]*AuditEvent, error)
	GetUserAuditTrail(ctx context.Context, userID uuid.UUID, fromDate, toDate time.Time) ([]*AuditEvent, error)
	GetWorkflowAuditTrail(ctx context.Context, requestID uuid.UUID) ([]*WorkflowAuditEvent, error)

	// Compliance operations
	GenerateComplianceReport(ctx context.Context, req *ComplianceReportRequest) (*ComplianceReport, error)
	ExportAuditData(ctx context.Context, req *AuditExportRequest) ([]byte, error)
}

// auditService implements AuditService
type auditService struct {
	tracing   tracing.TracingService
	metrics   metrics.MetricsProvider
	auditRepo Repository
}

// NewAuditService creates a new audit service
func NewAuditService(
	tracing tracing.TracingService,
	metrics metrics.MetricsProvider,
	auditRepo Repository,
) AuditService {
	return &auditService{
		tracing:   tracing,
		metrics:   metrics,
		auditRepo: auditRepo,
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

	if err := s.auditRepo.CreateAuditEvent(ctx, event); err != nil {
		// TODO: Add proper error handling, maybe a fallback mechanism
		return fmt.Errorf("failed to log audit event: %w", err)
	}

	s.metrics.IncrementCounter("audit_event_logged", map[string]any{
		"type":     string(event.EventType),
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
func (s *auditService) LogAccessRequestCreated(ctx context.Context, tenantID, requesterID uuid.UUID, targetUserID *uuid.UUID, entityID uuid.UUID, requestID uuid.UUID, requestType string, justification string, durationHours *int32) error {
	auditEvent := &AuditEvent{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EventType:     AuditEventAccessRequestCreated,
		EventCategory: AuditCategoryAccess,
		Severity:      AuditSeverityInfo,
		UserID:        &requesterID,
		TargetUserID:  targetUserID,
		EntityID:      &entityID,
		RequestID:     &requestID,
		Reason:        fmt.Sprintf("Access request created: %s", requestType),
		Context: map[string]any{
			"request_type":   requestType,
			"justification":  justification,
			"duration_hours": durationHours,
		},
		ComplianceFlags: []string{"ACCESS_CONTROL", "SEGREGATION_OF_DUTIES"},
		Timestamp:       time.Now(),
	}

	return s.LogAuditEvent(ctx, auditEvent)
}

// LogAccessRequestProcessed logs when an access request is approved or rejected
func (s *auditService) LogAccessRequestProcessed(ctx context.Context, tenantID, requestID uuid.UUID, approverID uuid.UUID, targetUserID *uuid.UUID, entityID uuid.UUID, action, comments, requestType, approvalStatus string) error {
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
		TenantID:      tenantID,
		EventType:     eventType,
		EventCategory: AuditCategoryAccess,
		Severity:      severity,
		UserID:        &approverID,
		TargetUserID:  targetUserID,
		EntityID:      &entityID,
		RequestID:     &requestID,
		Reason:        fmt.Sprintf("Access request %s: %s", action, comments),
		Context: map[string]any{
			"action":          action,
			"request_type":    requestType,
			"approval_status": approvalStatus,
			"comments":        comments,
		},
		ComplianceFlags: []string{"ACCESS_CONTROL", "APPROVAL_PROCESS"},
		Timestamp:       time.Now(),
	}

	return s.LogAuditEvent(ctx, auditEvent)
}

// LogAccessGranted logs when access is actually granted
func (s *auditService) LogAccessGranted(ctx context.Context, tenantID, requestID uuid.UUID, grantedBy, userID, entityID uuid.UUID, accessType string, accessID uuid.UUID, grantedAt time.Time, expiresAt *time.Time) error {
	auditEvent := &AuditEvent{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EventType:     AuditEventAccessGranted,
		EventCategory: AuditCategoryAccess,
		Severity:      AuditSeverityInfo,
		UserID:        &grantedBy,
		TargetUserID:  &userID,
		EntityID:      &entityID,
		RequestID:     &requestID,
		Reason:        fmt.Sprintf("Access granted: %s", accessType),
		Context: map[string]any{
			"access_type": accessType,
			"access_id":   accessID,
			"granted_at":  grantedAt,
			"expires_at":  expiresAt,
		},
		ComplianceFlags: []string{"ACCESS_CONTROL", "PRIVILEGED_ACCESS"},
		Timestamp:       grantedAt,
	}

	// Higher severity for privilege elevation
	if accessType == "ELEVATION" {
		auditEvent.Severity = AuditSeverityHigh
		auditEvent.ComplianceFlags = append(auditEvent.ComplianceFlags, "PRIVILEGE_ELEVATION")
	}

	return s.LogAuditEvent(ctx, auditEvent)
}

// LogAccessRevoked logs when access is revoked
func (s *auditService) LogAccessRevoked(ctx context.Context, tenantID, requestID uuid.UUID, targetUserID *uuid.UUID, entityID uuid.UUID, requestType string, accessID uuid.UUID) error {
	auditEvent := &AuditEvent{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EventType:     AuditEventAccessRevoked,
		EventCategory: AuditCategoryAccess,
		Severity:      AuditSeverityInfo,
		TargetUserID:  targetUserID,
		EntityID:      &entityID,
		RequestID:     &requestID,
		Reason:        fmt.Sprintf("Access revoked: %s", requestType),
		Context: map[string]any{
			"access_id":    accessID,
			"request_type": requestType,
			"revoked_at":   time.Now(),
		},
		ComplianceFlags: []string{"ACCESS_CONTROL", "ACCESS_REVOCATION"},
		Timestamp:       time.Now(),
	}

	return s.LogAuditEvent(ctx, auditEvent)
}

// LogSecurityViolation logs security violations
func (s *auditService) LogSecurityViolation(ctx context.Context, userID uuid.UUID, violation string, severity AuditEventSeverity, context map[string]any) error {
	auditEvent := &AuditEvent{
		ID:              uuid.New(),
		EventType:       AuditEventSecurityViolation,
		EventCategory:   AuditCategoryAuth,
		Severity:        severity,
		UserID:          &userID,
		Reason:          violation,
		RiskScore:       s.getViolationRiskScore(violation, severity),
		Context:         context,
		ComplianceFlags: []string{"SECURITY_VIOLATION", "INCIDENT_RESPONSE"},
		Timestamp:       time.Now(),
	}

	return s.LogAuditEvent(ctx, auditEvent)
}

// GetAuditEvents retrieves audit events based on query criteria
func (s *auditService) GetAuditEvents(ctx context.Context, req *AuditQueryRequest) ([]*AuditEvent, error) {
	return s.auditRepo.GetAuditEvents(ctx, req)
}

// GetUserAuditTrail gets audit trail for a specific user
func (s *auditService) GetUserAuditTrail(ctx context.Context, userID uuid.UUID, fromDate, toDate time.Time) ([]*AuditEvent, error) {
	return s.auditRepo.GetUserAuditTrail(ctx, userID, fromDate, toDate)
}

// GetWorkflowAuditTrail gets workflow audit trail for a specific request
func (s *auditService) GetWorkflowAuditTrail(ctx context.Context, requestID uuid.UUID) ([]*WorkflowAuditEvent, error) {
	return s.auditRepo.GetWorkflowAuditTrail(ctx, requestID)
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
