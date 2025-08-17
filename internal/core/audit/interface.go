package audit

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"encoding/json"
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// Repository defines the interface for the audit repository.
type Repository interface {
	// Core CRUD Operations
	// CreateAuditEvent(ctx context.Context, arg AuditEvent) error
	CreateAuditEvent(ctx context.Context, req CreateAuditEventRequest) (*AuditEvent, error)
	GetAuditEvents(ctx context.Context, tenantID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error)
	GetAuditEventByID(ctx context.Context, tenantID uuid.UUID, eventID uuid.UUID) (*AuditEvent, error)

	// User-specific queries
	GetUserAuditHistory(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error)
	GetUserRiskProfile(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, startTime *time.Time) (*UserRiskProfile, error)
	GetUserSessionEvents(ctx context.Context, tenantID uuid.UUID, sessionID uuid.UUID) ([]AuditEvent, error)

	// Entity and Resource queries
	GetAuditEventsByEntity(ctx context.Context, tenantID uuid.UUID, entityID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error)
	GetAuditEventsByResource(ctx context.Context, tenantID uuid.UUID, resourceID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error)

	// Security and Risk Analysis
	GetHighRiskEvents(ctx context.Context, tenantID uuid.UUID, minRiskScore int, filters AuditEventFilters) ([]AuditEvent, error)
	GetFailedAccessAttempts(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, userID *uuid.UUID, ipAddress *netip.Addr, limit, offset int) ([]AuditEvent, error)
	GetRecentSecurityEvents(ctx context.Context, tenantID uuid.UUID, minRiskScore int, startTime time.Time, eventCategory *string, limit int) ([]AuditEvent, error)
	GetSuspiciousActivityByIP(ctx context.Context, tenantID uuid.UUID, ipAddress *netip.Addr, minRiskScore int, startTime time.Time, limit, offset int) ([]SuspiciousActivityByIP, error)

	// Compliance and Admin queries
	GetComplianceEvents(ctx context.Context, tenantID uuid.UUID, complianceFlag string, startTime, endTime *time.Time, limit, offset int) ([]AuditEvent, error)
	GetAdminActions(ctx context.Context, tenantID uuid.UUID, adminUserID, targetUserID *uuid.UUID, startTime, endTime *time.Time, limit, offset int) ([]AuditEvent, error)

	// Analytics and Statistics
	GetAuditStatsByCategory(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, severity *string) ([]AuditStatsByCategory, error)
	GetAuditStatsBySeverity(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, eventCategory *string) ([]AuditStatsBySeverity, error)
	GetAuditLogHealth(ctx context.Context, tenantID uuid.UUID) (*AuditLogHealth, error)
	GetAuditStorageStats(ctx context.Context, tenantID uuid.UUID) (*AuditStorageStats, error)
	GetHourlyEventRates(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, limit int) ([]HourlyEventRate, error)
	GetEventTypeDistribution(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, eventCategory *string, minOccurrences int, limit int) ([]EventTypeDistribution, error)

	// Forensic Investigation
	GetEventTimelineForUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, startTime, endTime time.Time) ([]EventTimeline, error)
	GetRelatedEventsByContext(ctx context.Context, tenantID uuid.UUID, contextKey string, contextSearch string, startTime, endTime time.Time, excludeEventID *uuid.UUID, limit int) ([]RelatedEventsByContext, error)
	GetSimilarIncidentPatterns(ctx context.Context, tenantID uuid.UUID, params SimilarIncidentPatternsParams) ([]SimilarIncidentPattern, error)

	// Advanced Analysis
	GetAnomalousUserBehavior(ctx context.Context, tenantID uuid.UUID, params AnomalousUserBehaviorParams) ([]AnomalousUserBehavior, error)
	GetUserAgentAnalysis(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, minEvents int, limit int) ([]UserAgentAnalysis, error)
	GetDuplicateEventAnalysis(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, minDuplicateCount int, limit int) ([]DuplicateEventAnalysis, error)

	// Bulk Operations
	BulkUpdateEventRiskScores(ctx context.Context, tenantID uuid.UUID, params BulkUpdateRiskScoresParams) error
	BulkAddComplianceFlags(ctx context.Context, tenantID uuid.UUID, params BulkAddComplianceFlagsParams) error

	// Maintenance Operations
	UpdateEventRiskScore(ctx context.Context, tenantID uuid.UUID, eventID uuid.UUID, riskScore int) error
	UpdateEventComplianceFlags(ctx context.Context, tenantID uuid.UUID, eventID uuid.UUID, complianceFlags json.RawMessage) error
	DeleteOldAuditEvents(ctx context.Context, tenantID uuid.UUID, cutoffDate time.Time) error
	ArchiveOldAuditEvents(ctx context.Context, tenantID uuid.UUID, cutoffDate time.Time, eventCategory *string, limit int) ([]AuditEvent, error)
	CleanupDuplicateEvents(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time) error
}

// Service Interface
type Service interface {
	CreateAuditEvent(ctx context.Context, req CreateAuditEventRequest) (*AuditEvent, error)
	GetAuditEvents(ctx context.Context, tenantID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error)
	GetAuditEventByID(ctx context.Context, tenantID uuid.UUID, eventID uuid.UUID) (*AuditEvent, error)
	GetUserAuditHistory(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error)
	GetHighRiskEvents(ctx context.Context, tenantID uuid.UUID, minRiskScore int, filters AuditEventFilters) ([]AuditEvent, error)
	GetFailedAccessAttempts(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, userID *uuid.UUID, ipAddress *string, limit, offset int) ([]AuditEvent, error)
	GetComplianceEvents(ctx context.Context, tenantID uuid.UUID, complianceFlag string, startTime, endTime *time.Time, limit, offset int) ([]AuditEvent, error)
	GetAuditEventsByEntity(ctx context.Context, tenantID uuid.UUID, entityID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error)
	GetAuditEventsByResource(ctx context.Context, tenantID uuid.UUID, resourceID uuid.UUID, filters AuditEventFilters) ([]AuditEvent, error)
	GetAdminActions(ctx context.Context, tenantID uuid.UUID, adminUserID, targetUserID *uuid.UUID, startTime, endTime *time.Time, limit, offset int) ([]AuditEvent, error)
	GetUserSessionEvents(ctx context.Context, tenantID uuid.UUID, sessionID uuid.UUID) ([]AuditEvent, error)

	// Analytics
	GetAuditStatsByCategory(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, severity *string) ([]AuditStatsByCategory, error)
	GetUserRiskProfile(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, startTime *time.Time) (*UserRiskProfile, error)
	GetAuditLogHealth(ctx context.Context, tenantID uuid.UUID) (*AuditLogHealth, error)
	GetHourlyEventRates(ctx context.Context, tenantID uuid.UUID, startTime, endTime time.Time, limit int) ([]HourlyEventRate, error)

	// Forensics
	GetEventTimelineForUser(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, startTime, endTime time.Time) ([]AuditEvent, error)
	GetRelatedEventsByContext(ctx context.Context, tenantID uuid.UUID, contextKey, contextSearch string, startTime, endTime time.Time, excludeEventID *uuid.UUID, limit int) ([]AuditEvent, error)

	// Maintenance
	UpdateEventRiskScore(ctx context.Context, tenantID uuid.UUID, eventID uuid.UUID, riskScore int) error
	UpdateEventComplianceFlags(ctx context.Context, tenantID uuid.UUID, eventID uuid.UUID, complianceFlags json.RawMessage) error
	DeleteOldAuditEvents(ctx context.Context, tenantID uuid.UUID, cutoffDate time.Time) error
}
