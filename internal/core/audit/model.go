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
	CreateAuditEvent(ctx context.Context, tenantID uuid.UUID, req CreateAuditEventRequest) (*AuditEvent, error)
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

// AuditEvent represents the data for an audit log entry.
type AuditEvent struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	UserID          uuid.UUID       `json:"user_id" db:"user_id"`
	EventType       string          `json:"event_type" db:"event_type"`
	EventCategory   string          `json:"event_category" db:"event_category"`
	Severity        string          `json:"severity" db:"severity"`
	TargetUserID    *uuid.UUID      `json:"target_user_id" db:"target_user_id"`
	EntityID        *uuid.UUID      `json:"entity_id" db:"entity_id"`
	ResourceID      *uuid.UUID      `json:"resource_id" db:"resource_id"`
	ActionID        *uuid.UUID      `json:"action_id" db:"action_id"`
	RoleID          *uuid.UUID      `json:"role_id" db:"role_id"`
	PermissionID    *uuid.UUID      `json:"permission_id" db:"permission_id"`
	Decision        *string         `json:"decision" db:"decision"`
	Reason          *string         `json:"reason" db:"reason"`
	RiskScore       *int            `json:"risk_score" db:"risk_score"`
	Context         json.RawMessage `json:"context" db:"context"`
	IPAddress       *string         `json:"ip_address" db:"ip_address"`
	UserAgent       *string         `json:"user_agent" db:"user_agent"`
	SessionID       *uuid.UUID      `json:"session_id" db:"session_id"`
	ComplianceFlags json.RawMessage `json:"compliance_flags" db:"compliance_flags"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

type CreateAuditEventRequest struct {
	UserID        *uuid.UUID      `json:"user_id"`
	EventType     string          `json:"event_type" binding:"required"`
	EventCategory string          `json:"event_category" binding:"required"`
	Severity      string          `json:"severity" binding:"required"`
	EntityID      *uuid.UUID      `json:"entity_id"`
	Decision      *string         `json:"decision"`
	Reason        *string         `json:"reason"`
	Context       json.RawMessage `json:"context"`
}

type AuditEventFilters struct {
	UserID        *uuid.UUID `form:"user_id"`
	EventCategory *string    `form:"event_category"`
	Severity      *string    `form:"severity"`
	StartTime     *time.Time `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime       *time.Time `form:"end_time" time_format:"2006-01-02T15:04:05Z07:00"`
	Limit         int        `form:"limit"`
	Offset        int        `form:"offset"`
}

type UserRiskProfile struct {
	UserID             uuid.UUID  `json:"user_id" db:"user_id"`
	TotalEvents        int        `json:"total_events" db:"total_events"`
	AvgRiskScore       *float64   `json:"avg_risk_score" db:"avg_risk_score"`
	MaxRiskScore       *int       `json:"max_risk_score" db:"max_risk_score"`
	FailedAttempts     int        `json:"failed_attempts" db:"failed_attempts"`
	HighSeverityEvents int        `json:"high_severity_events" db:"high_severity_events"`
	UniqueIPs          int        `json:"unique_ips" db:"unique_ips"`
	UniqueCategories   int        `json:"unique_categories" db:"unique_categories"`
	FirstEvent         *time.Time `json:"first_event" db:"first_event"`
	LastEvent          *time.Time `json:"last_event" db:"last_event"`
	EventsLast24h      int        `json:"events_last_24h" db:"events_last_24h"`
	EventsLast7d       int        `json:"events_last_7d" db:"events_last_7d"`
}

type AuditStatsByCategory struct {
	EventCategory string   `json:"event_category" db:"event_category"`
	EventCount    int      `json:"event_count" db:"event_count"`
	DeniedCount   int      `json:"denied_count" db:"denied_count"`
	AvgRiskScore  *float64 `json:"avg_risk_score" db:"avg_risk_score"`
	MaxRiskScore  *int     `json:"max_risk_score" db:"max_risk_score"`
	UniqueUsers   int      `json:"unique_users" db:"unique_users"`
}

type AuditLogHealth struct {
	TotalEvents     int      `json:"total_events" db:"total_events"`
	EventsLastHour  int      `json:"events_last_hour" db:"events_last_hour"`
	EventsLast24h   int      `json:"events_last_24h" db:"events_last_24h"`
	EventsLast7d    int      `json:"events_last_7d" db:"events_last_7d"`
	AvgRiskScore    *float64 `json:"avg_risk_score" db:"avg_risk_score"`
	StddevRiskScore *float64 `json:"stddev_risk_score" db:"stddev_risk_score"`
	P95RiskScore    *float64 `json:"p95_risk_score" db:"p95_risk_score"`
	HighRiskEvents  int      `json:"high_risk_events" db:"high_risk_events"`
	AllowedEvents   int      `json:"allowed_events" db:"allowed_events"`
	DeniedEvents    int      `json:"denied_events" db:"denied_events"`
	DenialRatePct   *float64 `json:"denial_rate_pct" db:"denial_rate_pct"`
}

type HourlyEventRate struct {
	HourBucket   time.Time `json:"hour_bucket" db:"hour_bucket"`
	EventCount   int       `json:"event_count" db:"event_count"`
	UniqueUsers  int       `json:"unique_users" db:"unique_users"`
	AvgRiskScore *float64  `json:"avg_risk_score" db:"avg_risk_score"`
	DeniedCount  int       `json:"denied_count" db:"denied_count"`
}

// Service Interface
type Service interface {
	CreateAuditEvent(ctx context.Context, tenantID uuid.UUID, req CreateAuditEventRequest) (*AuditEvent, error)
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

// Additional types for repository responses that extend beyond the basic AuditEvent

type AuditStatsBySeverity struct {
	Severity     string   `json:"severity" db:"severity"`
	EventCount   int      `json:"event_count" db:"event_count"`
	UniqueUsers  int      `json:"unique_users" db:"unique_users"`
	AvgRiskScore *float64 `json:"avg_risk_score" db:"avg_risk_score"`
}

type AuditStorageStats struct {
	TotalEvents      int        `json:"total_events" db:"total_events"`
	UniqueUsers      int        `json:"unique_users" db:"unique_users"`
	UniqueIPs        int        `json:"unique_ips" db:"unique_ips"`
	UniqueEventTypes int        `json:"unique_event_types" db:"unique_event_types"`
	TableSize        string     `json:"table_size" db:"table_size"`
	OldestEvent      *time.Time `json:"oldest_event" db:"oldest_event"`
	NewestEvent      *time.Time `json:"newest_event" db:"newest_event"`
	EventsLast24h    int        `json:"events_last_24h" db:"events_last_24h"`
	EventsLast7d     int        `json:"events_last_7d" db:"events_last_7d"`
	EventsLast30d    int        `json:"events_last_30d" db:"events_last_30d"`
}

type EventTypeDistribution struct {
	EventType         string     `json:"event_type" db:"event_type"`
	EventCategory     *string    `json:"event_category" db:"event_category"`
	EventCount        int        `json:"event_count" db:"event_count"`
	UniqueUsers       int        `json:"unique_users" db:"unique_users"`
	AvgRiskScore      *float64   `json:"avg_risk_score" db:"avg_risk_score"`
	DeniedCount       int        `json:"denied_count" db:"denied_count"`
	HighSeverityCount int        `json:"high_severity_count" db:"high_severity_count"`
	FirstOccurrence   *time.Time `json:"first_occurrence" db:"first_occurrence"`
	LastOccurrence    *time.Time `json:"last_occurrence" db:"last_occurrence"`
}

type EventTimeline struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	EventType        string          `json:"event_type" db:"event_type"`
	EventCategory    *string         `json:"event_category" db:"event_category"`
	Severity         *string         `json:"severity" db:"severity"`
	Decision         *string         `json:"decision" db:"decision"`
	Reason           string          `json:"reason" db:"reason"`
	RiskScore        *int            `json:"risk_score" db:"risk_score"`
	Context          json.RawMessage `json:"context" db:"context"`
	IPAddress        *netip.Addr     `json:"ip_address" db:"ip_address"`
	UserAgent        string          `json:"user_agent" db:"user_agent"`
	SessionID        *uuid.UUID      `json:"session_id" db:"session_id"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	PrevEventTime    *time.Time      `json:"prev_event_time" db:"prev_event_time"`
	NextEventTime    *time.Time      `json:"next_event_time" db:"next_event_time"`
	SecondsSincePrev *float64        `json:"seconds_since_prev" db:"seconds_since_prev"`
}

type RelatedEventsByContext struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	UserID           *uuid.UUID      `json:"user_id" db:"user_id"`
	EventType        string          `json:"event_type" db:"event_type"`
	EventCategory    *string         `json:"event_category" db:"event_category"`
	Severity         *string         `json:"severity" db:"severity"`
	Decision         *string         `json:"decision" db:"decision"`
	RiskScore        *int            `json:"risk_score" db:"risk_score"`
	Context          json.RawMessage `json:"context" db:"context"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
	ContextMatchType string          `json:"context_match_type" db:"context_match_type"`
}

type SuspiciousActivityByIP struct {
	IPAddress        netip.Addr `json:"ip_address" db:"ip_address"`
	UserID           *uuid.UUID `json:"user_id" db:"user_id"`
	EventType        string     `json:"event_type" db:"event_type"`
	Decision         *string    `json:"decision" db:"decision"`
	RiskScore        *int       `json:"risk_score" db:"risk_score"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	IPEventCount     int        `json:"ip_event_count" db:"ip_event_count"`
	IPDeniedCount    int        `json:"ip_denied_count" db:"ip_denied_count"`
	UniqueUsersPerIP int        `json:"unique_users_per_ip" db:"unique_users_per_ip"`
}

type SimilarIncidentPattern struct {
	UserID             *uuid.UUID  `json:"user_id" db:"user_id"`
	EventType          string      `json:"event_type" db:"event_type"`
	EventCategory      *string     `json:"event_category" db:"event_category"`
	Severity           *string     `json:"severity" db:"severity"`
	Decision           *string     `json:"decision" db:"decision"`
	RiskScore          *int        `json:"risk_score" db:"risk_score"`
	IPAddress          *netip.Addr `json:"ip_address" db:"ip_address"`
	CreatedAt          time.Time   `json:"created_at" db:"created_at"`
	PatternAvgRisk     float64     `json:"pattern_avg_risk" db:"pattern_avg_risk"`
	PatternDeniedCount int         `json:"pattern_denied_count" db:"pattern_denied_count"`
	RiskDeviation      int         `json:"risk_deviation" db:"risk_deviation"`
}

type AnomalousUserBehavior struct {
	UserID             *uuid.UUID `json:"user_id" db:"user_id"`
	EventType          string     `json:"event_type" db:"event_type"`
	EventCategory      *string    `json:"event_category" db:"event_category"`
	RiskScore          *int       `json:"risk_score" db:"risk_score"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UserEmail          *string    `json:"user_email" db:"user_email"`
	EntityName         *string    `json:"entity_name" db:"entity_name"`
	ResourceName       *string    `json:"resource_name" db:"resource_name"`
	DataClassification string     `json:"data_classification" db:"data_classification"`
}

type UserAgentAnalysis struct {
	UserAgent     string     `json:"user_agent" db:"user_agent"`
	EventCount    int        `json:"event_count" db:"event_count"`
	UniqueUsers   int        `json:"unique_users" db:"unique_users"`
	UniqueIPs     int        `json:"unique_ips" db:"unique_ips"`
	AvgRiskScore  *float64   `json:"avg_risk_score" db:"avg_risk_score"`
	DeniedCount   int        `json:"denied_count" db:"denied_count"`
	FirstSeen     *time.Time `json:"first_seen" db:"first_seen"`
	LastSeen      *time.Time `json:"last_seen" db:"last_seen"`
	AgentCategory string     `json:"agent_category" db:"agent_category"`
}

type DuplicateEventAnalysis struct {
	UserID          *uuid.UUID  `json:"user_id" db:"user_id"`
	EventType       string      `json:"event_type" db:"event_type"`
	EventCategory   *string     `json:"event_category" db:"event_category"`
	IPAddress       *netip.Addr `json:"ip_address" db:"ip_address"`
	TimeBucket      time.Time   `json:"time_bucket" db:"time_bucket"`
	DuplicateCount  int         `json:"duplicate_count" db:"duplicate_count"`
	EventIDs        []uuid.UUID `json:"event_ids" db:"event_ids"`
	FirstOccurrence *time.Time  `json:"first_occurrence" db:"first_occurrence"`
	LastOccurrence  *time.Time  `json:"last_occurrence" db:"last_occurrence"`
}

// Parameter types for complex queries

type SimilarIncidentPatternsParams struct {
	SearchStart           time.Time `json:"search_start"`
	SearchEnd             time.Time `json:"search_end"`
	MaxRiskDeviation      int       `json:"max_risk_deviation"`
	Limit                 int       `json:"limit"`
	IncidentEventType     string    `json:"incident_event_type"`
	IncidentEventCategory *string   `json:"incident_event_category"`
	PatternStart          time.Time `json:"pattern_start"`
	PatternEnd            time.Time `json:"pattern_end"`
}

type AnomalousUserBehaviorParams struct {
	StartTime          time.Time `json:"start_time"`
	EndTime            time.Time `json:"end_time"`
	DataClassification *string   `json:"data_classification"`
	Limit              int       `json:"limit"`
	BaselineStart      time.Time `json:"baseline_start"`
	BaselineEnd        time.Time `json:"baseline_end"`
	MinBaselineEvents  int       `json:"min_baseline_events"`
}

type BulkUpdateRiskScoresParams struct {
	NewRiskScore     int       `json:"new_risk_score"`
	EventType        string    `json:"event_type"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	CurrentRiskScore *int      `json:"current_risk_score"`
	EventCategory    *string   `json:"event_category"`
}

type BulkAddComplianceFlagsParams struct {
	NewComplianceFlags json.RawMessage `json:"new_compliance_flags"`
	EventType          string          `json:"event_type"`
	StartTime          time.Time       `json:"start_time"`
	EndTime            time.Time       `json:"end_time"`
	EventCategory      *string         `json:"event_category"`
}
