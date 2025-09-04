package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"time"

	db "github.com/niiniyare/erp/db/sqlc"
	"github.com/niiniyare/erp/internal/shared"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"

	"github.com/google/uuid"
)

type repository struct {
	store   db.Store
	logger  logger.Logger
	tracing tracing.TracingService
	metrics metrics.MetricsProvider
}

// NewRepository creates a new audit repository.
func NewRepository(store db.Store, logger logger.Logger, tracing tracing.TracingService, metrics metrics.MetricsProvider) Repository {
	return &repository{
		store:   store,
		logger:  logger,
		tracing: tracing,
		metrics: metrics,
	}
}

func (r *repository) CreateAuditEvent(ctx context.Context, req CreateAuditEventRequest) (*AuditEvent, error) {
	var auditEvent *AuditEvent
	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return nil, fmt.Errorf("tenantID has not been set in the context")
	}
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		var reason string
		if req.Reason != nil {
			reason = *req.Reason
		}

		params := db.CreateAuditEventParams{
			EventType:     req.EventType,
			EventCategory: &req.EventCategory, // Convert string to *string
			Severity:      &req.Severity,      // Convert string to *string
			Context:       req.Context,
			UserID:        req.UserID,
			EntityID:      req.EntityID,
			Decision:      req.Decision,
			Reason:        reason, // Handle nil pointer properly
		}

		dbAuditEvent, err := s.CreateAuditEvent(ctx, params)
		if err != nil {
			return err
		}
		// Convert SQLC model to domain model, handling nullable fields properly
		var userID uuid.UUID
		if dbAuditEvent.UserID != nil {
			userID = *dbAuditEvent.UserID
		}

		var eventCategory, severity string
		if dbAuditEvent.EventCategory != nil {
			eventCategory = *dbAuditEvent.EventCategory
		}
		if dbAuditEvent.Severity != nil {
			severity = *dbAuditEvent.Severity
		}

		var riskScore *int
		if dbAuditEvent.RiskScore != nil {
			riskScoreInt := int(*dbAuditEvent.RiskScore)
			riskScore = &riskScoreInt
		}

		var ipAddress *string
		if dbAuditEvent.IpAddress != nil {
			ipStr := dbAuditEvent.IpAddress.String()
			ipAddress = &ipStr
		}

		var createdAt time.Time
		if dbAuditEvent.CreatedAt.Valid {
			createdAt = dbAuditEvent.CreatedAt.Time
		}

		auditEvent = &AuditEvent{
			ID:              dbAuditEvent.ID,
			UserID:          userID,
			EventType:       dbAuditEvent.EventType,
			EventCategory:   eventCategory,
			Severity:        severity,
			TargetUserID:    dbAuditEvent.TargetUserID,
			EntityID:        dbAuditEvent.EntityID,
			ResourceID:      dbAuditEvent.ResourceID,
			ActionID:        dbAuditEvent.ActionID,
			RoleID:          dbAuditEvent.RoleID,
			PermissionID:    dbAuditEvent.PermissionID,
			Decision:        dbAuditEvent.Decision,
			Reason:          &dbAuditEvent.Reason,
			RiskScore:       riskScore,
			Context:         dbAuditEvent.Context,
			IPAddress:       ipAddress,
			UserAgent:       &dbAuditEvent.UserAgent,
			SessionID:       dbAuditEvent.SessionID,
			ComplianceFlags: dbAuditEvent.ComplianceFlags,
			CreatedAt:       createdAt,
		}
		return nil
	})
	return auditEvent, err
}

func (r *repository) GetAuditEvents(ctx context.Context, tenantID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error) {
	var auditEvents []AuditEvent
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetAuditEvents
		return nil
	})
	return auditEvents, err
}

func (r *repository) GetAuditEventByID(ctx context.Context, tenantID uuid.UUID, eventID uuid.UUID) (*AuditEvent, error) {
	var auditEvent *AuditEvent
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetAuditEventByID
		return nil
	})
	return auditEvent, err
}

func (r *repository) GetUserAuditHistory(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error) {
	var auditEvents []AuditEvent
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetUserAuditHistory
		return nil
	})
	return auditEvents, err
}

func (r *repository) GetUserRiskProfile(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, startTime *time.Time) (*UserRiskProfile, error) {
	var userRiskProfile *UserRiskProfile
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetUserRiskProfile
		return nil
	})
	return userRiskProfile, err
}

func (r *repository) GetUserSessionEvents(ctx context.Context, tenantID uuid.UUID, sessionID uuid.UUID) ([]AuditEvent, error) {
	var auditEvents []AuditEvent
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetUserSessionEvents
		return nil
	})
	return auditEvents, err
}

func (r *repository) GetAuditEventsByEntity(ctx context.Context, tenantID uuid.UUID, entityID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error) {
	var auditEvents []AuditEvent
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetAuditEventsByEntity
		return nil
	})
	return auditEvents, err
}

func (r *repository) GetAuditEventsByResource(ctx context.Context, tenantID uuid.UUID, resourceID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error) {
	var auditEvents []AuditEvent
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetAuditEventsByResource
		return nil
	})
	return auditEvents, err
}

func (r *repository) GetHighRiskEvents(ctx context.Context, tenantID uuid.UUID, minRiskScore int, filters AuditEventFilters) ([]AuditEvent, error) {
	var auditEvents []AuditEvent
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetHighRiskEvents
		return nil
	})
	return auditEvents, err
}

func (r *repository) GetFailedAccessAttempts(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, userID *uuid.UUID, ipAddress *netip.Addr, limit, offset int) ([]AuditEvent, error) {
	var auditEvents []AuditEvent
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetFailedAccessAttempts
		return nil
	})
	return auditEvents, err
}

func (r *repository) GetRecentSecurityEvents(ctx context.Context, tenantID uuid.UUID, minRiskScore int, startTime time.Time, eventCategory *string, limit int) ([]AuditEvent, error) {
	var auditEvents []AuditEvent
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetRecentSecurityEvents
		return nil
	})
	return auditEvents, err
}

func (r *repository) GetSuspiciousActivityByIP(ctx context.Context, tenantID uuid.UUID, ipAddress *netip.Addr, minRiskScore int, startTime time.Time, limit, offset int) ([]SuspiciousActivityByIP, error) {
	var suspiciousActivity []SuspiciousActivityByIP
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetSuspiciousActivityByIP
		return nil
	})
	return suspiciousActivity, err
}

func (r *repository) GetComplianceEvents(ctx context.Context, tenantID uuid.UUID, complianceFlag string, startTime, endTime *time.Time, limit, offset int) ([]AuditEvent, error) {
	var auditEvents []AuditEvent
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetComplianceEvents
		return nil
	})
	return auditEvents, err
}

func (r *repository) GetAdminActions(ctx context.Context, tenantID uuid.UUID, adminUserID, targetUserID *uuid.UUID, startTime, endTime *time.Time, limit, offset int) ([]AuditEvent, error) {
	var auditEvents []AuditEvent
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetAdminActions
		return nil
	})
	return auditEvents, err
}

func (r *repository) GetAuditStatsByCategory(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, severity *string) ([]AuditStatsByCategory, error) {
	var stats []AuditStatsByCategory
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetAuditStatsByCategory
		return nil
	})
	return stats, err
}

func (r *repository) GetAuditStatsBySeverity(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, eventCategory *string) ([]AuditStatsBySeverity, error) {
	var stats []AuditStatsBySeverity
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetAuditStatsBySeverity
		return nil
	})
	return stats, err
}

func (r *repository) GetAuditLogHealth(ctx context.Context, tenantID uuid.UUID) (*AuditLogHealth, error) {
	var health *AuditLogHealth
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetAuditLogHealth
		return nil
	})
	return health, err
}

func (r *repository) GetAuditStorageStats(ctx context.Context, tenantID uuid.UUID) (*AuditStorageStats, error) {
	var stats *AuditStorageStats
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetAuditStorageStats
		return nil
	})
	return stats, err
}

func (r *repository) GetHourlyEventRates(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, limit int) ([]HourlyEventRate, error) {
	var rates []HourlyEventRate
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetHourlyEventRates
		return nil
	})
	return rates, err
}

func (r *repository) GetEventTypeDistribution(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, eventCategory *string, minOccurrences int, limit int) ([]EventTypeDistribution, error) {
	var distribution []EventTypeDistribution
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetEventTypeDistribution
		return nil
	})
	return distribution, err
}

func (r *repository) GetEventTimelineForUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, startTime, endTime time.Time) ([]EventTimeline, error) {
	var timeline []EventTimeline
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetEventTimelineForUser
		return nil
	})
	return timeline, err
}

func (r *repository) GetRelatedEventsByContext(ctx context.Context, tenantID uuid.UUID, contextKey string, contextSearch string, startTime, endTime time.Time, excludeEventID *uuid.UUID, limit int) ([]RelatedEventsByContext, error) {
	var events []RelatedEventsByContext
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetRelatedEventsByContext
		return nil
	})
	return events, err
}

func (r *repository) GetSimilarIncidentPatterns(ctx context.Context, tenantID uuid.UUID, params SimilarIncidentPatternsParams) ([]SimilarIncidentPattern, error) {
	var patterns []SimilarIncidentPattern
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetSimilarIncidentPatterns
		return nil
	})
	return patterns, err
}

func (r *repository) GetAnomalousUserBehavior(ctx context.Context, tenantID uuid.UUID, params AnomalousUserBehaviorParams) ([]AnomalousUserBehavior, error) {
	var behaviors []AnomalousUserBehavior
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetAnomalousUserBehavior
		return nil
	})
	return behaviors, err
}

func (r *repository) GetUserAgentAnalysis(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, minEvents int, limit int) ([]UserAgentAnalysis, error) {
	var analysis []UserAgentAnalysis
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetUserAgentAnalysis
		return nil
	})
	return analysis, err
}

func (r *repository) GetDuplicateEventAnalysis(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, minDuplicateCount int, limit int) ([]DuplicateEventAnalysis, error) {
	var analysis []DuplicateEventAnalysis
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for GetDuplicateEventAnalysis
		return nil
	})
	return analysis, err
}

func (r *repository) BulkUpdateEventRiskScores(ctx context.Context, tenantID uuid.UUID, params BulkUpdateRiskScoresParams) error {
	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for BulkUpdateEventRiskScores
		return nil
	})
}

func (r *repository) BulkAddComplianceFlags(ctx context.Context, tenantID uuid.UUID, params BulkAddComplianceFlagsParams) error {
	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for BulkAddComplianceFlags
		return nil
	})
}

func (r *repository) UpdateEventRiskScore(ctx context.Context, tenantID uuid.UUID, eventID uuid.UUID, riskScore int) error {
	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for UpdateEventRiskScore
		return nil
	})
}

func (r *repository) UpdateEventComplianceFlags(ctx context.Context, tenantID uuid.UUID, eventID uuid.UUID, complianceFlags json.RawMessage) error {
	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for UpdateEventComplianceFlags
		return nil
	})
}

func (r *repository) DeleteOldAuditEvents(ctx context.Context, tenantID uuid.UUID, cutoffDate time.Time) error {
	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for DeleteOldAuditEvents
		return nil
	})
}

func (r *repository) ArchiveOldAuditEvents(ctx context.Context, tenantID uuid.UUID, cutoffDate time.Time, eventCategory *string, limit int) ([]AuditEvent, error) {
	var auditEvents []AuditEvent
	err := r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for ArchiveOldAuditEvents
		return nil
	})
	return auditEvents, err
}

func (r *repository) CleanupDuplicateEvents(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time) error {
	return r.store.WithTenant(ctx, tenantID, func(ctx context.Context, s db.Store) error {
		// Implementation for CleanupDuplicateEvents
		return nil
	})
}
