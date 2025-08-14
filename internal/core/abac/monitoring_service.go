package abac

//go:generate go run go.uber.org/mock/mockgen -source=monitoring_service.go -destination=mock.go -package=abac

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/niiniyare/erp/internal/core/abac/repository"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"github.com/niiniyare/erp/internal/shared/types"
)

// MonitoringService provides comprehensive ABAC monitoring and observability
type MonitoringService interface {
	// Performance Monitoring
	RecordEvaluationMetrics(ctx context.Context, req *EvaluationMetricsRequest) error
	GetEvaluationMetrics(ctx context.Context, req *MetricsQueryRequest) (*EvaluationMetrics, error)

	// Decision Monitoring
	TrackPolicyDecision(ctx context.Context, req *PolicyDecisionEvent) error
	GetDecisionPatterns(ctx context.Context, req *DecisionPatternRequest) (*DecisionPatternAnalysis, error)

	// Anomaly Detection
	DetectAnomalies(ctx context.Context, req *AnomalyDetectionRequest) (*AnomalyDetectionResult, error)
	ConfigureAnomalyDetection(ctx context.Context, req *AnomalyConfigRequest) (*AnomalyConfiguration, error)

	// Alerting
	CreateAlert(ctx context.Context, req *CreateAlertRequest) (*Alert, error)
	GetActiveAlerts(ctx context.Context, req *AlertQueryRequest) (*AlertQueryResult, error)
	ResolveAlert(ctx context.Context, req *ResolveAlertRequest) error

	// Dashboard Support
	GetDashboardData(ctx context.Context, req *DashboardDataRequest) (*DashboardData, error)
	GenerateReport(ctx context.Context, req *ReportGenerationRequest) (*MonitoringReport, error)

	// Health Monitoring
	GetSystemHealth(ctx context.Context) (*SystemHealthStatus, error)
	RunHealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResult, error)
}

// monitoringService implements MonitoringService
type monitoringService struct {
	policyRepo     repository.PolicyRepository
	evaluationRepo repository.PolicyEvaluationRepository
	attributeRepo  repository.AttributeRepository

	// Monitoring state
	metricsStore    *MetricsStore
	alertManager    *AlertManager
	anomalyDetector *AnomalyDetector
	healthMonitor   *HealthMonitor

	logger  logger.Logger
	metrics metrics.MetricsProvider
	tracer  tracing.TracingService
}

// NewMonitoringService creates a new monitoring service instance
func NewMonitoringService(
	policyRepo repository.PolicyRepository,
	evaluationRepo repository.PolicyEvaluationRepository,
	attributeRepo repository.AttributeRepository,
	logger logger.Logger,
	metrics metrics.MetricsProvider,
	tracer tracing.TracingService,
) MonitoringService {
	return &monitoringService{
		policyRepo:      policyRepo,
		evaluationRepo:  evaluationRepo,
		attributeRepo:   attributeRepo,
		metricsStore:    NewMetricsStore(),
		alertManager:    NewAlertManager(),
		anomalyDetector: NewAnomalyDetector(),
		healthMonitor:   NewHealthMonitor(),
		logger:          logger,
		metrics:         metrics,
		tracer:          tracer,
	}
}

// Performance Monitoring Types

type EvaluationMetricsRequest struct {
	EvaluationID       uuid.UUID                `json:"evaluation_id" validate:"required"`
	UserID             uuid.UUID                `json:"user_id" validate:"required"`
	PolicyID           *uuid.UUID               `json:"policy_id,omitempty"`
	ResourceType       string                   `json:"resource_type" validate:"required"`
	ResourceID         *uuid.UUID               `json:"resource_id,omitempty"`
	Action             string                   `json:"action" validate:"required"`
	Decision           types.PolicyDecisionType `json:"decision" validate:"required"`
	ExecutionTime      time.Duration            `json:"execution_time" validate:"required"`
	CacheHit           bool                     `json:"cache_hit"`
	ErrorOccurred      bool                     `json:"error_occurred"`
	ErrorType          string                   `json:"error_type,omitempty"`
	ContextSize        int32                    `json:"context_size"`
	PoliciesEvaluated  int32                    `json:"policies_evaluated"`
	AttributesAccessed int32                    `json:"attributes_accessed"`
	Timestamp          time.Time                `json:"timestamp"`
	Metadata           map[string]interface{}   `json:"metadata,omitempty"`
}

type MetricsQueryRequest struct {
	TimeRange   TimeRangeFilter    `json:"time_range"`
	Filters     []MetricFilter     `json:"filters,omitempty"`
	Aggregation MetricsAggregation `json:"aggregation"`
	GroupBy     []string           `json:"group_by,omitempty"`
	Limit       int32              `json:"limit"`
}

type TimeRangeFilter struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Interval  string    `json:"interval"` // "1m", "5m", "1h", "1d"
}

type MetricsAggregation struct {
	Function string `json:"function"` // "count", "avg", "min", "max", "sum", "p95", "p99"
	Field    string `json:"field"`
}

type EvaluationMetrics struct {
	TimeRange         TimeRangeFilter            `json:"time_range"`
	TotalEvaluations  int64                      `json:"total_evaluations"`
	PerformanceStats  EvaluationPerformanceStats `json:"performance_stats"`
	DecisionBreakdown DecisionBreakdown          `json:"decision_breakdown"`
	ErrorStats        ErrorStatistics            `json:"error_stats"`
	CacheStats        CacheStatistics            `json:"cache_stats"`
	TrendAnalysis     []MetricTrend              `json:"trend_analysis"`
	TopPolicies       []PolicyUsageStats         `json:"top_policies"`
	TopUsers          []UserActivityStats        `json:"top_users"`
	TopResources      []ResourceAccessStats      `json:"top_resources"`
}

type EvaluationPerformanceStats struct {
	AverageLatency      time.Duration               `json:"average_latency"`
	MedianLatency       time.Duration               `json:"median_latency"`
	P95Latency          time.Duration               `json:"p95_latency"`
	P99Latency          time.Duration               `json:"p99_latency"`
	MinLatency          time.Duration               `json:"min_latency"`
	MaxLatency          time.Duration               `json:"max_latency"`
	TotalExecutionTime  time.Duration               `json:"total_execution_time"`
	ThroughputPerSecond float64                     `json:"throughput_per_second"`
	LatencyDistribution []LatencyDistributionBucket `json:"latency_distribution"`
}

type LatencyDistributionBucket struct {
	Range      string  `json:"range"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

type DecisionBreakdown struct {
	AllowCount              int64   `json:"allow_count"`
	DenyCount               int64   `json:"deny_count"`
	NotApplicableCount      int64   `json:"not_applicable_count"`
	AllowPercentage         float64 `json:"allow_percentage"`
	DenyPercentage          float64 `json:"deny_percentage"`
	NotApplicablePercentage float64 `json:"not_applicable_percentage"`
}

type ErrorStatistics struct {
	TotalErrors       int64            `json:"total_errors"`
	ErrorRate         float64          `json:"error_rate"`
	ErrorsByType      map[string]int64 `json:"errors_by_type"`
	ErrorsByPolicy    map[string]int64 `json:"errors_by_policy"`
	CriticalErrors    int64            `json:"critical_errors"`
	RecoverableErrors int64            `json:"recoverable_errors"`
	ErrorTrends       []ErrorTrend     `json:"error_trends"`
}

type ErrorTrend struct {
	Timestamp  time.Time `json:"timestamp"`
	ErrorCount int64     `json:"error_count"`
	ErrorRate  float64   `json:"error_rate"`
	ErrorType  string    `json:"error_type"`
}

type CacheStatistics struct {
	TotalRequests    int64   `json:"total_requests"`
	CacheHits        int64   `json:"cache_hits"`
	CacheMisses      int64   `json:"cache_misses"`
	HitRate          float64 `json:"hit_rate"`
	MissRate         float64 `json:"miss_rate"`
	EvictionCount    int64   `json:"eviction_count"`
	CacheSize        int64   `json:"cache_size"`
	CacheUtilization float64 `json:"cache_utilization"`
}

type MetricTrend struct {
	MetricName  string           `json:"metric_name"`
	DataPoints  []TrendDataPoint `json:"data_points"`
	TrendType   string           `json:"trend_type"` // "increasing", "decreasing", "stable", "volatile"
	ChangeRate  float64          `json:"change_rate"`
	Correlation float64          `json:"correlation"`
	Seasonality string           `json:"seasonality"`
}

type TrendDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type PolicyUsageStats struct {
	PolicyID        uuid.UUID     `json:"policy_id"`
	PolicyName      string        `json:"policy_name"`
	EvaluationCount int64         `json:"evaluation_count"`
	AllowCount      int64         `json:"allow_count"`
	DenyCount       int64         `json:"deny_count"`
	AverageLatency  time.Duration `json:"average_latency"`
	ErrorCount      int64         `json:"error_count"`
	UsagePercentage float64       `json:"usage_percentage"`
}

type UserActivityStats struct {
	UserID          uuid.UUID     `json:"user_id"`
	RequestCount    int64         `json:"request_count"`
	AllowedRequests int64         `json:"allowed_requests"`
	DeniedRequests  int64         `json:"denied_requests"`
	UniqueResources int32         `json:"unique_resources"`
	AverageLatency  time.Duration `json:"average_latency"`
	LastActivity    time.Time     `json:"last_activity"`
}

type ResourceAccessStats struct {
	ResourceType   string        `json:"resource_type"`
	ResourceID     *uuid.UUID    `json:"resource_id,omitempty"`
	AccessCount    int64         `json:"access_count"`
	UniqueUsers    int32         `json:"unique_users"`
	AllowedAccess  int64         `json:"allowed_access"`
	DeniedAccess   int64         `json:"denied_access"`
	AverageLatency time.Duration `json:"average_latency"`
	LastAccessed   time.Time     `json:"last_accessed"`
}

func (ms *monitoringService) RecordEvaluationMetrics(ctx context.Context, req *EvaluationMetricsRequest) error {
	ctx, span := ms.tracer.StartSpan(ctx, "abac.monitoring_service.RecordEvaluationMetrics",
		tracing.WithAttributes(
			attribute.String("evaluation_id", req.EvaluationID.String()),
			attribute.String("user_id", req.UserID.String()),
			attribute.String("decision", string(req.Decision)),
			attribute.Int64("execution_time_ms", req.ExecutionTime.Milliseconds()),
		))
	defer span.End()

	// Store metrics in metrics store
	ms.metricsStore.AddEvaluationMetric(req)

	// Update real-time metrics
	ms.metrics.ObserveHistogram("abac_evaluation_duration_seconds", req.ExecutionTime.Seconds(),
		metrics.Fields{
			"resource_type": req.ResourceType,
			"action":        req.Action,
			"decision":      string(req.Decision),
			"cache_hit":     fmt.Sprintf("%t", req.CacheHit),
		})

	ms.metrics.IncrementCounter("abac_evaluations_total",
		metrics.Fields{
			"resource_type": req.ResourceType,
			"action":        req.Action,
			"decision":      string(req.Decision),
		})

	if req.ErrorOccurred {
		ms.metrics.IncrementCounter("abac_evaluation_errors_total",
			metrics.Fields{
				"error_type":    req.ErrorType,
				"resource_type": req.ResourceType,
			})
	}

	// Check for anomalies
	anomalyCtx := context.Background()
	_, err := ms.anomalyDetector.CheckForAnomalies(anomalyCtx, req)
	if err != nil {
		ms.logger.WarnContext(ctx, "Failed to check for anomalies", logger.Fields{"error": err.Error()})
	}

	ms.logger.DebugContext(ctx, "Recorded evaluation metrics",
		logger.Fields{
			"evaluation_id":  req.EvaluationID,
			"execution_time": req.ExecutionTime.Milliseconds(),
			"decision":       req.Decision,
			"cache_hit":      req.CacheHit,
			"error_occurred": req.ErrorOccurred,
		})

	return nil
}

// Decision Monitoring Types

type PolicyDecisionEvent struct {
	EventID        uuid.UUID                `json:"event_id" validate:"required"`
	EvaluationID   uuid.UUID                `json:"evaluation_id" validate:"required"`
	UserID         uuid.UUID                `json:"user_id" validate:"required"`
	PolicyID       *uuid.UUID               `json:"policy_id,omitempty"`
	ResourceType   string                   `json:"resource_type" validate:"required"`
	ResourceID     *uuid.UUID               `json:"resource_id,omitempty"`
	Action         string                   `json:"action" validate:"required"`
	Decision       types.PolicyDecisionType `json:"decision" validate:"required"`
	DecisionReason string                   `json:"decision_reason"`
	Context        map[string]interface{}   `json:"context"`
	IPAddress      string                   `json:"ip_address,omitempty"`
	UserAgent      string                   `json:"user_agent,omitempty"`
	SessionID      *uuid.UUID               `json:"session_id,omitempty"`
	Timestamp      time.Time                `json:"timestamp"`
	Severity       EventSeverity            `json:"severity"`
	Tags           []string                 `json:"tags,omitempty"`
}

type EventSeverity string

const (
	EventSeverityInfo     EventSeverity = "info"
	EventSeverityWarning  EventSeverity = "warning"
	EventSeverityError    EventSeverity = "error"
	EventSeverityCritical EventSeverity = "critical"
)

type DecisionPatternRequest struct {
	TimeRange       TimeRangeFilter `json:"time_range"`
	Filters         []PatternFilter `json:"filters,omitempty"`
	PatternTypes    []PatternType   `json:"pattern_types"`
	MinOccurrences  int32           `json:"min_occurrences"`
	ConfidenceLevel float64         `json:"confidence_level"`
}

type PatternFilter struct {
	FilterType  string      `json:"filter_type"`
	FilterKey   string      `json:"filter_key"`
	FilterValue interface{} `json:"filter_value"`
}

type PatternType string

const (
	PatternTypeAccess   PatternType = "access_pattern"
	PatternTypeTemporal PatternType = "temporal_pattern"
	PatternTypeUser     PatternType = "user_pattern"
	PatternTypeResource PatternType = "resource_pattern"
	PatternTypeAnomaly  PatternType = "anomaly_pattern"
)

type DecisionPatternAnalysis struct {
	TimeRange        TimeRangeFilter         `json:"time_range"`
	DetectedPatterns []DecisionPattern       `json:"detected_patterns"`
	PatternSummary   PatternSummary          `json:"pattern_summary"`
	Recommendations  []PatternRecommendation `json:"recommendations"`
	AnalysisMetadata PatternAnalysisMetadata `json:"analysis_metadata"`
}

type DecisionPattern struct {
	PatternID       uuid.UUID     `json:"pattern_id"`
	PatternType     PatternType   `json:"pattern_type"`
	PatternName     string        `json:"pattern_name"`
	Description     string        `json:"description"`
	Occurrences     int64         `json:"occurrences"`
	Confidence      float64       `json:"confidence"`
	Significance    string        `json:"significance"`
	PatternData     PatternData   `json:"pattern_data"`
	DetectedAt      time.Time     `json:"detected_at"`
	FirstOccurrence time.Time     `json:"first_occurrence"`
	LastOccurrence  time.Time     `json:"last_occurrence"`
	Trend           string        `json:"trend"`
	Impact          PatternImpact `json:"impact"`
}

type PatternData struct {
	UserPatterns     []UserPattern     `json:"user_patterns,omitempty"`
	ResourcePatterns []ResourcePattern `json:"resource_patterns,omitempty"`
	TemporalPatterns []TemporalPattern `json:"temporal_patterns,omitempty"`
	AccessPatterns   []AccessPattern   `json:"access_patterns,omitempty"`
}

type UserPattern struct {
	UserID           uuid.UUID `json:"user_id"`
	BehaviorType     string    `json:"behavior_type"`
	Frequency        float64   `json:"frequency"`
	Regularity       float64   `json:"regularity"`
	DeviationScore   float64   `json:"deviation_score"`
	TypicalTimeSlots []string  `json:"typical_time_slots"`
}

type ResourcePattern struct {
	ResourceType     string             `json:"resource_type"`
	AccessPattern    string             `json:"access_pattern"`
	PeakTimes        []string           `json:"peak_times"`
	UsageIntensity   float64            `json:"usage_intensity"`
	UserDistribution map[string]float64 `json:"user_distribution"`
}

type TemporalPattern struct {
	PatternName    string             `json:"pattern_name"`
	TimeSlots      []TemporalTimeSlot `json:"time_slots"`
	Seasonality    string             `json:"seasonality"`
	CyclicBehavior bool               `json:"cyclic_behavior"`
	PeakPeriods    []TimePeriod       `json:"peak_periods"`
}

type TemporalTimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Activity  float64   `json:"activity"`
	DayOfWeek string    `json:"day_of_week"`
}

type TimePeriod struct {
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Intensity float64   `json:"intensity"`
}

type AccessPattern struct {
	PatternName      string             `json:"pattern_name"`
	ResourceSequence []string           `json:"resource_sequence"`
	ActionSequence   []string           `json:"action_sequence"`
	TransitionMatrix map[string]float64 `json:"transition_matrix"`
	Probability      float64            `json:"probability"`
}

type PatternImpact struct {
	SecurityImpact    string   `json:"security_impact"`
	PerformanceImpact string   `json:"performance_impact"`
	BusinessImpact    string   `json:"business_impact"`
	RiskLevel         string   `json:"risk_level"`
	AffectedUsers     int32    `json:"affected_users"`
	AffectedResources int32    `json:"affected_resources"`
	Implications      []string `json:"implications"`
}

type PatternSummary struct {
	TotalPatterns            int32                 `json:"total_patterns"`
	PatternsByType           map[PatternType]int32 `json:"patterns_by_type"`
	HighConfidencePatterns   int32                 `json:"high_confidence_patterns"`
	SecurityRelevantPatterns int32                 `json:"security_relevant_patterns"`
	AverageConfidence        float64               `json:"average_confidence"`
	TrendDirection           string                `json:"trend_direction"`
}

type PatternRecommendation struct {
	RecommendationID     uuid.UUID   `json:"recommendation_id"`
	RecommendationType   string      `json:"recommendation_type"`
	Priority             string      `json:"priority"`
	Description          string      `json:"description"`
	ActionRequired       bool        `json:"action_required"`
	ExpectedBenefit      string      `json:"expected_benefit"`
	ImplementationEffort string      `json:"implementation_effort"`
	RelatedPatterns      []uuid.UUID `json:"related_patterns"`
}

type PatternAnalysisMetadata struct {
	AnalysisAlgorithm    string        `json:"analysis_algorithm"`
	AnalysisVersion      string        `json:"analysis_version"`
	ProcessingTime       time.Duration `json:"processing_time"`
	DataPointsAnalyzed   int64         `json:"data_points_analyzed"`
	ConfidenceThreshold  float64       `json:"confidence_threshold"`
	AnalysisCompleteness float64       `json:"analysis_completeness"`
}

func (ms *monitoringService) TrackPolicyDecision(ctx context.Context, req *PolicyDecisionEvent) error {
	ctx, span := ms.tracer.StartSpan(ctx, "abac.monitoring_service.TrackPolicyDecision",
		tracing.WithAttributes(
			attribute.String("event_id", req.EventID.String()),
			attribute.String("user_id", req.UserID.String()),
			attribute.String("decision", string(req.Decision)),
			attribute.String("severity", string(req.Severity)),
		))
	defer span.End()

	// Store decision event
	ms.metricsStore.AddDecisionEvent(req)

	// Check for security-relevant patterns
	if req.Decision == types.PolicyDecisionDeny || req.Severity >= EventSeverityWarning {
		go func() {
			securityCtx := context.Background()
			ms.checkSecurityImplications(securityCtx, req)
		}()
	}

	// Update decision metrics
	ms.metrics.IncrementCounter("abac_policy_decisions_total",
		metrics.Fields{
			"decision":      string(req.Decision),
			"resource_type": req.ResourceType,
			"action":        req.Action,
			"severity":      string(req.Severity),
		})

	ms.logger.InfoContext(ctx, "Tracked policy decision event",
		logger.Fields{
			"event_id":      req.EventID,
			"user_id":       req.UserID,
			"decision":      req.Decision,
			"resource_type": req.ResourceType,
			"action":        req.Action,
			"severity":      req.Severity,
		})

	return nil
}

// Anomaly Detection Types

type AnomalyDetectionRequest struct {
	TimeRange          TimeRangeFilter    `json:"time_range"`
	DetectionMethods   []DetectionMethod  `json:"detection_methods"`
	SensitivityLevel   SensitivityLevel   `json:"sensitivity_level"`
	Scope              AnomalyScope       `json:"scope"`
	ThresholdOverrides map[string]float64 `json:"threshold_overrides,omitempty"`
}

type DetectionMethod string

const (
	DetectionMethodStatistical DetectionMethod = "statistical"
	DetectionMethodMachine     DetectionMethod = "machine_learning"
	DetectionMethodRule        DetectionMethod = "rule_based"
	DetectionMethodTemporal    DetectionMethod = "temporal"
	DetectionMethodBehavioral  DetectionMethod = "behavioral"
)

type SensitivityLevel string

const (
	SensitivityLevelLow    SensitivityLevel = "low"
	SensitivityLevelMedium SensitivityLevel = "medium"
	SensitivityLevelHigh   SensitivityLevel = "high"
)

type AnomalyScope struct {
	UserIDs       []uuid.UUID `json:"user_ids,omitempty"`
	ResourceTypes []string    `json:"resource_types,omitempty"`
	PolicyIDs     []uuid.UUID `json:"policy_ids,omitempty"`
	Actions       []string    `json:"actions,omitempty"`
	IPRanges      []string    `json:"ip_ranges,omitempty"`
}

type AnomalyDetectionResult struct {
	DetectionID       uuid.UUID                `json:"detection_id"`
	TimeRange         TimeRangeFilter          `json:"time_range"`
	DetectedAnomalies []DetectedAnomaly        `json:"detected_anomalies"`
	AnomalySummary    AnomalySummary           `json:"anomaly_summary"`
	DetectionMetadata AnomalyDetectionMetadata `json:"detection_metadata"`
	Recommendations   []AnomalyRecommendation  `json:"recommendations"`
	ExecutionTime     time.Duration            `json:"execution_time"`
	Timestamp         time.Time                `json:"timestamp"`
}

type DetectedAnomaly struct {
	AnomalyID        uuid.UUID        `json:"anomaly_id"`
	AnomalyType      AnomalyType      `json:"anomaly_type"`
	Severity         AnomalySeverity  `json:"severity"`
	Confidence       float64          `json:"confidence"`
	Description      string           `json:"description"`
	AffectedEntities []AffectedEntity `json:"affected_entities"`
	DetectionMethod  DetectionMethod  `json:"detection_method"`
	AnomalyScore     float64          `json:"anomaly_score"`
	BaselineValue    float64          `json:"baseline_value"`
	ObservedValue    float64          `json:"observed_value"`
	Deviation        float64          `json:"deviation"`
	DetectedAt       time.Time        `json:"detected_at"`
	FirstOccurrence  time.Time        `json:"first_occurrence"`
	Duration         time.Duration    `json:"duration"`
	Context          AnomalyContext   `json:"context"`
	RelatedEvents    []RelatedEvent   `json:"related_events,omitempty"`
}

type AnomalyType string

const (
	AnomalyTypeVolumeSpike    AnomalyType = "volume_spike"
	AnomalyTypeVolumeDrop     AnomalyType = "volume_drop"
	AnomalyTypeLatencySpike   AnomalyType = "latency_spike"
	AnomalyTypeErrorSpike     AnomalyType = "error_spike"
	AnomalyTypeUnusualAccess  AnomalyType = "unusual_access"
	AnomalyTypeSuspiciousUser AnomalyType = "suspicious_user"
	AnomalyTypePatternChange  AnomalyType = "pattern_change"
)

type AnomalySeverity string

const (
	AnomalySeverityLow      AnomalySeverity = "low"
	AnomalySeverityMedium   AnomalySeverity = "medium"
	AnomalySeverityHigh     AnomalySeverity = "high"
	AnomalySeverityCritical AnomalySeverity = "critical"
)

type AffectedEntity struct {
	EntityType string    `json:"entity_type"` // "user", "resource", "policy", "system"
	EntityID   uuid.UUID `json:"entity_id"`
	EntityName string    `json:"entity_name"`
	Impact     string    `json:"impact"`
}

type AnomalyContext struct {
	TriggerEvents        []string               `json:"trigger_events"`
	EnvironmentalFactors []string               `json:"environmental_factors"`
	SystemState          map[string]interface{} `json:"system_state"`
	TemporalContext      TemporalContext        `json:"temporal_context"`
	GeographicContext    *GeographicContext     `json:"geographic_context,omitempty"`
}

type TemporalContext struct {
	TimeOfDay      string `json:"time_of_day"`
	DayOfWeek      string `json:"day_of_week"`
	IsBusinessHour bool   `json:"is_business_hour"`
	IsHoliday      bool   `json:"is_holiday"`
	Timezone       string `json:"timezone"`
}

type GeographicContext struct {
	Country     string    `json:"country"`
	Region      string    `json:"region"`
	City        string    `json:"city"`
	Coordinates []float64 `json:"coordinates,omitempty"`
	IPRange     string    `json:"ip_range"`
}

type RelatedEvent struct {
	EventID     uuid.UUID     `json:"event_id"`
	EventType   string        `json:"event_type"`
	Correlation float64       `json:"correlation"`
	TimeDelta   time.Duration `json:"time_delta"`
	Description string        `json:"description"`
}

type AnomalySummary struct {
	TotalAnomalies      int32                     `json:"total_anomalies"`
	AnomaliesBySeverity map[AnomalySeverity]int32 `json:"anomalies_by_severity"`
	AnomaliesByType     map[AnomalyType]int32     `json:"anomalies_by_type"`
	CriticalAnomalies   int32                     `json:"critical_anomalies"`
	NewAnomalies        int32                     `json:"new_anomalies"`
	ResolvedAnomalies   int32                     `json:"resolved_anomalies"`
	AverageConfidence   float64                   `json:"average_confidence"`
	TrendDirection      string                    `json:"trend_direction"`
}

type AnomalyDetectionMetadata struct {
	DetectionAlgorithms []string           `json:"detection_algorithms"`
	ModelVersions       map[string]string  `json:"model_versions"`
	TrainingDataPeriod  TimeRangeFilter    `json:"training_data_period"`
	DetectionThresholds map[string]float64 `json:"detection_thresholds"`
	ProcessingTime      time.Duration      `json:"processing_time"`
	DataQualityScore    float64            `json:"data_quality_score"`
}

type AnomalyRecommendation struct {
	RecommendationID   uuid.UUID   `json:"recommendation_id"`
	RecommendationType string      `json:"recommendation_type"`
	Priority           string      `json:"priority"`
	Description        string      `json:"description"`
	ActionRequired     bool        `json:"action_required"`
	AutoRemediable     bool        `json:"auto_remediable"`
	RelatedAnomalies   []uuid.UUID `json:"related_anomalies"`
	EstimatedImpact    string      `json:"estimated_impact"`
}

// Metrics Store Implementation

type MetricsStore struct {
	evaluationMetrics []EvaluationMetricsRequest
	decisionEvents    []PolicyDecisionEvent
	mutex             sync.RWMutex
	retentionPeriod   time.Duration
}

func NewMetricsStore() *MetricsStore {
	return &MetricsStore{
		evaluationMetrics: make([]EvaluationMetricsRequest, 0),
		decisionEvents:    make([]PolicyDecisionEvent, 0),
		retentionPeriod:   7 * 24 * time.Hour, // 7 days default
	}
}

func (ms *MetricsStore) AddEvaluationMetric(metric *EvaluationMetricsRequest) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	ms.evaluationMetrics = append(ms.evaluationMetrics, *metric)
	ms.cleanupOldMetrics()
}

func (ms *MetricsStore) AddDecisionEvent(event *PolicyDecisionEvent) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	ms.decisionEvents = append(ms.decisionEvents, *event)
	ms.cleanupOldEvents()
}

func (ms *MetricsStore) QueryMetrics(query *MetricsQueryRequest) (*EvaluationMetrics, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	// Filter metrics based on query
	filteredMetrics := ms.filterEvaluationMetrics(query)

	// Calculate aggregated metrics
	return ms.calculateAggregatedMetrics(filteredMetrics, query), nil
}

func (ms *MetricsStore) filterEvaluationMetrics(query *MetricsQueryRequest) []EvaluationMetricsRequest {
	var filtered []EvaluationMetricsRequest

	for _, metric := range ms.evaluationMetrics {
		// Time range filter
		if metric.Timestamp.Before(query.TimeRange.StartTime) || metric.Timestamp.After(query.TimeRange.EndTime) {
			continue
		}

		// Apply additional filters
		matchesFilters := true
		for _, filter := range query.Filters {
			if !ms.matchesFilter(metric, filter) {
				matchesFilters = false
				break
			}
		}

		if matchesFilters {
			filtered = append(filtered, metric)
		}
	}

	return filtered
}

func (ms *MetricsStore) matchesFilter(metric EvaluationMetricsRequest, filter MetricFilter) bool {
	// Simplified filter matching
	switch filter.FilterKey {
	case "resource_type":
		return metric.ResourceType == filter.FilterValue
	case "action":
		return metric.Action == filter.FilterValue
	case "decision":
		return string(metric.Decision) == filter.FilterValue
	case "cache_hit":
		return metric.CacheHit == filter.FilterValue
	default:
		return true
	}
}

func (ms *MetricsStore) calculateAggregatedMetrics(metrics []EvaluationMetricsRequest, query *MetricsQueryRequest) *EvaluationMetrics {
	if len(metrics) == 0 {
		return &EvaluationMetrics{
			TimeRange:        query.TimeRange,
			TotalEvaluations: 0,
		}
	}

	// Calculate performance stats
	var totalLatency time.Duration
	var latencies []time.Duration
	var allowCount, denyCount, notApplicableCount int64
	var errorCount int64
	var cacheHits, cacheMisses int64

	for _, metric := range metrics {
		totalLatency += metric.ExecutionTime
		latencies = append(latencies, metric.ExecutionTime)

		switch metric.Decision {
		case types.PolicyDecisionAllow:
			allowCount++
		case types.PolicyDecisionDeny:
			denyCount++
		default:
			notApplicableCount++
		}

		if metric.ErrorOccurred {
			errorCount++
		}

		if metric.CacheHit {
			cacheHits++
		} else {
			cacheMisses++
		}
	}

	// Sort latencies for percentile calculations
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	total := int64(len(metrics))
	avgLatency := totalLatency / time.Duration(total)

	// Calculate percentiles
	p95Index := int(float64(len(latencies)) * 0.95)
	p99Index := int(float64(len(latencies)) * 0.99)
	medianIndex := len(latencies) / 2

	performanceStats := EvaluationPerformanceStats{
		AverageLatency:      avgLatency,
		MedianLatency:       latencies[medianIndex],
		P95Latency:          latencies[p95Index],
		P99Latency:          latencies[p99Index],
		MinLatency:          latencies[0],
		MaxLatency:          latencies[len(latencies)-1],
		TotalExecutionTime:  totalLatency,
		ThroughputPerSecond: float64(total) / query.TimeRange.EndTime.Sub(query.TimeRange.StartTime).Seconds(),
	}

	// Calculate decision breakdown
	decisionBreakdown := DecisionBreakdown{
		AllowCount:              allowCount,
		DenyCount:               denyCount,
		NotApplicableCount:      notApplicableCount,
		AllowPercentage:         float64(allowCount) / float64(total) * 100,
		DenyPercentage:          float64(denyCount) / float64(total) * 100,
		NotApplicablePercentage: float64(notApplicableCount) / float64(total) * 100,
	}

	// Calculate error stats
	errorStats := ErrorStatistics{
		TotalErrors:       errorCount,
		ErrorRate:         float64(errorCount) / float64(total) * 100,
		ErrorsByType:      make(map[string]int64),
		CriticalErrors:    0, // Simplified
		RecoverableErrors: errorCount,
	}

	// Calculate cache stats
	totalCacheRequests := cacheHits + cacheMisses
	cacheStats := CacheStatistics{
		TotalRequests:    totalCacheRequests,
		CacheHits:        cacheHits,
		CacheMisses:      cacheMisses,
		HitRate:          float64(cacheHits) / float64(totalCacheRequests) * 100,
		MissRate:         float64(cacheMisses) / float64(totalCacheRequests) * 100,
		CacheUtilization: 75.0, // Simplified
	}

	return &EvaluationMetrics{
		TimeRange:         query.TimeRange,
		TotalEvaluations:  total,
		PerformanceStats:  performanceStats,
		DecisionBreakdown: decisionBreakdown,
		ErrorStats:        errorStats,
		CacheStats:        cacheStats,
		TrendAnalysis:     []MetricTrend{},         // Simplified
		TopPolicies:       []PolicyUsageStats{},    // Simplified
		TopUsers:          []UserActivityStats{},   // Simplified
		TopResources:      []ResourceAccessStats{}, // Simplified
	}
}

func (ms *MetricsStore) cleanupOldMetrics() {
	cutoff := time.Now().Add(-ms.retentionPeriod)

	var filtered []EvaluationMetricsRequest
	for _, metric := range ms.evaluationMetrics {
		if metric.Timestamp.After(cutoff) {
			filtered = append(filtered, metric)
		}
	}
	ms.evaluationMetrics = filtered
}

func (ms *MetricsStore) cleanupOldEvents() {
	cutoff := time.Now().Add(-ms.retentionPeriod)

	var filtered []PolicyDecisionEvent
	for _, event := range ms.decisionEvents {
		if event.Timestamp.After(cutoff) {
			filtered = append(filtered, event)
		}
	}
	ms.decisionEvents = filtered
}

// Alert Manager Implementation

type AlertManager struct {
	alerts []Alert
	mutex  sync.RWMutex
}

type Alert struct {
	AlertID     uuid.UUID              `json:"alert_id"`
	AlertType   AlertType              `json:"alert_type"`
	Severity    AlertSeverity          `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Status      AlertStatus            `json:"status"`
	Source      AlertSource            `json:"source"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Labels      map[string]string      `json:"labels,omitempty"`
}

type AlertType string

const (
	AlertTypePerformance AlertType = "performance"
	AlertTypeSecurity    AlertType = "security"
	AlertTypeAnomaly     AlertType = "anomaly"
	AlertTypeSystem      AlertType = "system"
	AlertTypePolicy      AlertType = "policy"
)

type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
	AlertStatusSuppressed   AlertStatus = "suppressed"
)

type AlertSource struct {
	Component string `json:"component"`
	Instance  string `json:"instance"`
	Version   string `json:"version"`
}

func NewAlertManager() *AlertManager {
	return &AlertManager{
		alerts: make([]Alert, 0),
	}
}

// Anomaly Detector Implementation

type AnomalyDetector struct {
	baselines map[string]Baseline
	mutex     sync.RWMutex
}

type Baseline struct {
	MetricName  string             `json:"metric_name"`
	Mean        float64            `json:"mean"`
	StdDev      float64            `json:"std_dev"`
	LastUpdated time.Time          `json:"last_updated"`
	SampleCount int64              `json:"sample_count"`
	Percentiles map[string]float64 `json:"percentiles"`
}

func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{
		baselines: make(map[string]Baseline),
	}
}

func (ad *AnomalyDetector) CheckForAnomalies(ctx context.Context, metric *EvaluationMetricsRequest) (*DetectedAnomaly, error) {
	ad.mutex.RLock()
	defer ad.mutex.RUnlock()

	// Simplified anomaly detection based on execution time
	metricKey := fmt.Sprintf("%s_%s", metric.ResourceType, metric.Action)
	baseline, exists := ad.baselines[metricKey]

	if !exists {
		// No baseline yet, create one
		return nil, nil
	}

	// Check if execution time is anomalous
	executionTimeMs := float64(metric.ExecutionTime.Milliseconds())
	threshold := baseline.Mean + (3 * baseline.StdDev) // 3-sigma rule

	if executionTimeMs > threshold {
		anomaly := &DetectedAnomaly{
			AnomalyID:       uuid.New(),
			AnomalyType:     AnomalyTypeLatencySpike,
			Severity:        AnomalySeverityMedium,
			Confidence:      0.8,
			Description:     fmt.Sprintf("Execution time anomaly detected for %s %s", metric.ResourceType, metric.Action),
			DetectionMethod: DetectionMethodStatistical,
			AnomalyScore:    (executionTimeMs - baseline.Mean) / baseline.StdDev,
			BaselineValue:   baseline.Mean,
			ObservedValue:   executionTimeMs,
			Deviation:       (executionTimeMs - baseline.Mean) / baseline.Mean * 100,
			DetectedAt:      time.Now(),
			FirstOccurrence: metric.Timestamp,
			Duration:        0,
		}

		return anomaly, nil
	}

	return nil, nil
}

// Health Monitor Implementation

type HealthMonitor struct {
	healthChecks map[string]HealthCheck
	mutex        sync.RWMutex
}

type HealthCheck struct {
	CheckName    string                 `json:"check_name"`
	CheckType    string                 `json:"check_type"`
	Status       HealthStatus           `json:"status"`
	LastChecked  time.Time              `json:"last_checked"`
	ResponseTime time.Duration          `json:"response_time"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

func NewHealthMonitor() *HealthMonitor {
	return &HealthMonitor{
		healthChecks: make(map[string]HealthCheck),
	}
}

// Implementation of remaining interface methods (simplified for brevity)

func (ms *monitoringService) GetEvaluationMetrics(ctx context.Context, req *MetricsQueryRequest) (*EvaluationMetrics, error) {
	return ms.metricsStore.QueryMetrics(req)
}

func (ms *monitoringService) GetDecisionPatterns(ctx context.Context, req *DecisionPatternRequest) (*DecisionPatternAnalysis, error) {
	// Simplified implementation
	return &DecisionPatternAnalysis{
		TimeRange:        req.TimeRange,
		DetectedPatterns: []DecisionPattern{},
		PatternSummary: PatternSummary{
			TotalPatterns: 0,
		},
	}, nil
}

func (ms *monitoringService) DetectAnomalies(ctx context.Context, req *AnomalyDetectionRequest) (*AnomalyDetectionResult, error) {
	// Simplified implementation
	return &AnomalyDetectionResult{
		DetectionID:       uuid.New(),
		TimeRange:         req.TimeRange,
		DetectedAnomalies: []DetectedAnomaly{},
		AnomalySummary: AnomalySummary{
			TotalAnomalies: 0,
		},
		ExecutionTime: 100 * time.Millisecond,
		Timestamp:     time.Now(),
	}, nil
}

func (ms *monitoringService) checkSecurityImplications(ctx context.Context, event *PolicyDecisionEvent) {
	// Simplified security analysis
	if event.Decision == types.PolicyDecisionDeny {
		ms.logger.WarnContext(ctx, "Security event: Access denied",
			logger.Fields{
				"user_id":       event.UserID,
				"resource_type": event.ResourceType,
				"action":        event.Action,
				"ip_address":    event.IPAddress,
			})
	}
}

// Placeholder implementations for remaining interface methods

type AnomalyConfigRequest struct{}
type AnomalyConfiguration struct{}
type CreateAlertRequest struct{}
type AlertQueryRequest struct{}
type AlertQueryResult struct{}
type ResolveAlertRequest struct{}
type DashboardDataRequest struct{}
type DashboardData struct{}
type ReportGenerationRequest struct{}
type MonitoringReport struct{}
type SystemHealthStatus struct{}
type HealthCheckRequest struct{}
type HealthCheckResult struct{}

func (ms *monitoringService) ConfigureAnomalyDetection(ctx context.Context, req *AnomalyConfigRequest) (*AnomalyConfiguration, error) {
	return &AnomalyConfiguration{}, nil
}

func (ms *monitoringService) CreateAlert(ctx context.Context, req *CreateAlertRequest) (*Alert, error) {
	return &Alert{}, nil
}

func (ms *monitoringService) GetActiveAlerts(ctx context.Context, req *AlertQueryRequest) (*AlertQueryResult, error) {
	return &AlertQueryResult{}, nil
}

func (ms *monitoringService) ResolveAlert(ctx context.Context, req *ResolveAlertRequest) error {
	return nil
}

func (ms *monitoringService) GetDashboardData(ctx context.Context, req *DashboardDataRequest) (*DashboardData, error) {
	return &DashboardData{}, nil
}

func (ms *monitoringService) GenerateReport(ctx context.Context, req *ReportGenerationRequest) (*MonitoringReport, error) {
	return &MonitoringReport{}, nil
}

func (ms *monitoringService) GetSystemHealth(ctx context.Context) (*SystemHealthStatus, error) {
	return &SystemHealthStatus{}, nil
}

func (ms *monitoringService) RunHealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResult, error) {
	return &HealthCheckResult{}, nil
}
