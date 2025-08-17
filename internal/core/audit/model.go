package audit

import (
	"encoding/json"
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// AuditEvent represents the data for an audit log entry.
type AuditEvent struct {
	ID              uuid.UUID       `json:"id"`
	UserID          uuid.UUID       `json:"user_id"`
	EventType       string          `json:"event_type"`
	EventCategory   string          `json:"event_category"`
	Severity        string          `json:"severity"`
	TargetUserID    *uuid.UUID      `json:"target_user_id"`
	EntityID        *uuid.UUID      `json:"entity_id"`
	ResourceID      *uuid.UUID      `json:"resource_id"`
	ActionID        *uuid.UUID      `json:"action_id"`
	RoleID          *uuid.UUID      `json:"role_id"`
	PermissionID    *uuid.UUID      `json:"permission_id"`
	Decision        *string         `json:"decision"`
	Reason          *string         `json:"reason"`
	RiskScore       *int            `json:"risk_score"`
	Context         json.RawMessage `json:"context"`
	IPAddress       *string         `json:"ip_address"`
	UserAgent       *string         `json:"user_agent"`
	SessionID       *uuid.UUID      `json:"session_id"`
	ComplianceFlags json.RawMessage `json:"compliance_flags"`
	CreatedAt       time.Time       `json:"created_at"`
}

// CreateAuditEventRequest defines the structure for creating a new audit event.
type CreateAuditEventRequest struct {
	UserID          *uuid.UUID      `json:"user_id"`
	EventType       string          `json:"event_type" binding:"required"`
	EventCategory   string          `json:"event_category" binding:"required"`
	Severity        string          `json:"severity" binding:"required"`
	TargetUserID    *uuid.UUID      `json:"target_user_id"`
	EntityID        *uuid.UUID      `json:"entity_id"`
	ResourceID      *uuid.UUID      `json:"resource_id"`
	ActionID        *uuid.UUID      `json:"action_id"`
	RoleID          *uuid.UUID      `json:"role_id"`
	PermissionID    *uuid.UUID      `json:"permission_id"`
	Decision        *string         `json:"decision"`
	Reason          *string         `json:"reason"`
	RiskScore       *int            `json:"risk_score"`
	Context         json.RawMessage `json:"context"`
	IPAddress       *string         `json:"ip_address"`
	UserAgent       *string         `json:"user_agent"`
	SessionID       *uuid.UUID      `json:"session_id"`
	ComplianceFlags json.RawMessage `json:"compliance_flags"`
}

// AuditEventFilters defines the available filters for querying audit events.
// Assuming this struct is needed by the service interface.
type AuditEventFilters struct {
	// Define filter fields here, e.g., StartTime, EndTime, EventType, etc.
	Limit  int
	Offset int
}

// --- Analytics & Statistics Structs ---

type UserRiskProfile struct {
	UserID             uuid.UUID  `json:"user_id"`
	TotalEvents        int        `json:"total_events"`
	AvgRiskScore       *float64   `json:"avg_risk_score"`
	MaxRiskScore       *int       `json:"max_risk_score"`
	FailedAttempts     int        `json:"failed_attempts"`
	HighSeverityEvents int        `json:"high_severity_events"`
	UniqueIPs          int        `json:"unique_ips"`
	UniqueCategories   int        `json:"unique_categories"`
	FirstEvent         *time.Time `json:"first_event"`
	LastEvent          *time.Time `json:"last_event"`
	EventsLast24h      int        `json:"events_last_24h"`
	EventsLast7d       int        `json:"events_last_7d"`
}

type AuditStatsByCategory struct {
	EventCategory string   `json:"event_category"`
	EventCount    int      `json:"event_count"`
	DeniedCount   int      `json:"denied_count"`
	AvgRiskScore  *float64 `json:"avg_risk_score"`
	MaxRiskScore  *int     `json:"max_risk_score"`
	UniqueUsers   int      `json:"unique_users"`
}

type AuditLogHealth struct {
	TotalEvents     int      `json:"total_events"`
	EventsLastHour  int      `json:"events_last_hour"`
	EventsLast24h   int      `json:"events_last_24h"`
	EventsLast7d    int      `json:"events_last_7d"`
	AvgRiskScore    *float64 `json:"avg_risk_score"`
	StddevRiskScore *float64 `json:"stddev_risk_score"`
	P95RiskScore    *float64 `json:"p95_risk_score"`
	HighRiskEvents  int      `json:"high_risk_events"`
	AllowedEvents   int      `json:"allowed_events"`
	DeniedEvents    int      `json:"denied_events"`
	DenialRatePct   *float64 `json:"denial_rate_pct"`
}

type HourlyEventRate struct {
	HourBucket   time.Time `json:"hour_bucket"`
	EventCount   int       `json:"event_count"`
	UniqueUsers  int       `json:"unique_users"`
	AvgRiskScore *float64  `json:"avg_risk_score"`
	DeniedCount  int       `json:"denied_count"`
}

type AuditStatsBySeverity struct {
	Severity     string   `json:"severity"`
	EventCount   int      `json:"event_count"`
	UniqueUsers  int      `json:"unique_users"`
	AvgRiskScore *float64 `json:"avg_risk_score"`
}

type AuditStorageStats struct {
	TotalEvents      int        `json:"total_events"`
	UniqueUsers      int        `json:"unique_users"`
	UniqueIPs        int        `json:"unique_ips"`
	UniqueEventTypes int        `json:"unique_event_types"`
	TableSize        string     `json:"table_size"`
	OldestEvent      *time.Time `json:"oldest_event"`
	NewestEvent      *time.Time `json:"newest_event"`
}

type IPUsageStats struct {
	IPAddress        netip.Addr `json:"ip_address"`
	TotalEvents      int        `json:"total_events"`
	DeniedCount      int        `json:"denied_count"`
	UniqueUsersPerIP int        `json:"unique_users_per_ip"`
	FirstSeen        time.Time  `json:"first_seen"`
	LastSeen         time.Time  `json:"last_seen"`
}

type SimilarIncidentPattern struct {
	UserID             *uuid.UUID  `json:"user_id"`
	EventType          string      `json:"event_type"`
	EventCategory      *string     `json:"event_category"`
	Severity           *string     `json:"severity"`
	Decision           *string     `json:"decision"`
	RiskScore          *int        `json:"risk_score"`
	IPAddress          *netip.Addr `json:"ip_address"`
	CreatedAt          time.Time   `json:"created_at"`
	PatternAvgRisk     float64     `json:"pattern_avg_risk"`
	PatternDeniedCount int         `json:"pattern_denied_count"`
	RiskDeviation      int         `json:"risk_deviation"`
}

type AnomalousUserBehavior struct {
	UserID             *uuid.UUID `json:"user_id"`
	EventType          string     `json:"event_type"`
	EventCategory      *string    `json:"event_category"`
	RiskScore          *int       `json:"risk_score"`
	CreatedAt          time.Time  `json:"created_at"`
	UserEmail          *string    `json:"user_email"`
	EntityName         *string    `json:"entity_name"`
	ResourceName       *string    `json:"resource_name"`
	DataClassification string     `json:"data_classification"`
}

type UserAgentAnalysis struct {
	UserAgent     string     `json:"user_agent"`
	EventCount    int        `json:"event_count"`
	UniqueUsers   int        `json:"unique_users"`
	UniqueIPs     int        `json:"unique_ips"`
	AvgRiskScore  *float64   `json:"avg_risk_score"`
	DeniedCount   int        `json:"denied_count"`
	FirstSeen     *time.Time `json:"first_seen"`
	LastSeen      *time.Time `json:"last_seen"`
	AgentCategory string     `json:"agent_category"`
}

type DuplicateEventAnalysis struct {
	UserID          *uuid.UUID  `json:"user_id"`
	EventType       string      `json:"event_type"`
	EventCategory   *string     `json:"event_category"`
	IPAddress       *netip.Addr `json:"ip_address"`
	TimeBucket      time.Time   `json:"time_bucket"`
	DuplicateCount  int         `json:"duplicate_count"`
	EventIDs        []uuid.UUID `json:"event_ids"`
	FirstOccurrence *time.Time  `json:"first_occurrence"`
	LastOccurrence  *time.Time  `json:"last_occurrence"`
}

// --- Parameter Structs for Service/Repository ---

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

// Missing types referenced in interface.go
type SuspiciousActivityByIP struct {
	IPAddress        netip.Addr `json:"ip_address"`
	TotalEvents      int        `json:"total_events"`
	DeniedCount      int        `json:"denied_count"`
	UniqueUsersPerIP int        `json:"unique_users_per_ip"`
	FirstSeen        time.Time  `json:"first_seen"`
	LastSeen         time.Time  `json:"last_seen"`
}

type EventTypeDistribution struct {
	EventType       string    `json:"event_type"`
	EventCount      int       `json:"event_count"`
	UniqueUsers     int       `json:"unique_users"`
	FirstOccurrence time.Time `json:"first_occurrence"`
	LastOccurrence  time.Time `json:"last_occurrence"`
}

type EventTimeline struct {
	EventID       uuid.UUID       `json:"event_id"`
	EventType     string          `json:"event_type"`
	EventCategory string          `json:"event_category"`
	Severity      string          `json:"severity"`
	UserID        uuid.UUID       `json:"user_id"`
	IPAddress     *string         `json:"ip_address"`
	Decision      *string         `json:"decision"`
	Context       json.RawMessage `json:"context"`
	CreatedAt     time.Time       `json:"created_at"`
}

type RelatedEventsByContext struct {
	EventID         uuid.UUID       `json:"event_id"`
	EventType       string          `json:"event_type"`
	UserID          uuid.UUID       `json:"user_id"`
	IPAddress       *string         `json:"ip_address"`
	Context         json.RawMessage `json:"context"`
	SimilarityScore float64         `json:"similarity_score"`
	CreatedAt       time.Time       `json:"created_at"`
}
