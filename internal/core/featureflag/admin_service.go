package featureflag

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
)

// AdminService provides administrative operations for feature flags
type AdminService interface {
	// Bulk Operations
	BulkEnableFlags(ctx context.Context, request *BulkEnableFlagsRequest) (*BulkOperationResult, error)
	BulkDisableFlags(ctx context.Context, request *BulkDisableFlagsRequest) (*BulkOperationResult, error)
	BulkDeleteFlags(ctx context.Context, request *BulkDeleteFlagsRequest) (*BulkOperationResult, error)
	BulkUpdateRollout(ctx context.Context, request *BulkUpdateRolloutRequest) (*BulkOperationResult, error)

	// Flag Templates and Presets
	CreateFlagTemplate(ctx context.Context, request *CreateFlagTemplateRequest) (*FlagTemplate, error)
	GetFlagTemplate(ctx context.Context, templateID uuid.UUID) (*FlagTemplate, error)
	ListFlagTemplates(ctx context.Context, request *ListTemplatesRequest) (*ListTemplatesResponse, error)
	ApplyTemplate(ctx context.Context, request *ApplyTemplateRequest) (*FeatureFlag, error)

	// System Management
	GetSystemHealth(ctx context.Context) (*SystemHealthResult, error)
	GetSystemMetrics(ctx context.Context, request *SystemMetricsRequest) (*SystemMetricsResult, error)
	GetUsageAnalytics(ctx context.Context, request *UsageAnalyticsRequest) (*UsageAnalyticsResult, error)

	// Emergency Controls
	EmergencyDisableAll(ctx context.Context, reason string) (*EmergencyActionResult, error)
	EmergencyEnableAll(ctx context.Context, reason string) (*EmergencyActionResult, error)
	CreateRolloutStrategy(ctx context.Context, request *CreateRolloutStrategyRequest) (*RolloutStrategy, error)

	// Cache Management
	WarmupCache(ctx context.Context, request *CacheWarmupRequest) (*CacheOperationResult, error)
	ClearCache(ctx context.Context, request *CacheClearRequest) (*CacheOperationResult, error)
	GetCacheStats(ctx context.Context) (*CacheStatsResult, error)
}

// adminServiceImpl implements AdminService
type adminServiceImpl struct {
	baseService  Service
	auditService audit.Service
	logger       logger.Logger
	metrics      *metrics.MetricsService
	tracing      tracing.Service
	cacheWarmup  *CacheWarmer
}

// NewAdminService creates a new admin service for feature flags
func NewAdminService(
	baseService Service,
	auditService audit.Service,
	logger logger.Logger,
	metrics *metrics.MetricsService,
	tracing tracing.Service,
	cacheWarmup *CacheWarmer,
) AdminService {
	return &adminServiceImpl{
		baseService:  baseService,
		auditService: auditService,
		logger:       logger,
		metrics:      metrics,
		tracing:      tracing,
		cacheWarmup:  cacheWarmup,
	}
}

// Request/Response Types

// BulkEnableFlagsRequest enables multiple flags
type BulkEnableFlagsRequest struct {
	FlagNames []string `json:"flag_names" validate:"required,max=100"`
	Reason    string   `json:"reason" validate:"required"`
}

// BulkDisableFlagsRequest disables multiple flags
type BulkDisableFlagsRequest struct {
	FlagNames []string `json:"flag_names" validate:"required,max=100"`
	Reason    string   `json:"reason" validate:"required"`
}

// BulkDeleteFlagsRequest deletes multiple flags
type BulkDeleteFlagsRequest struct {
	FlagIDs []uuid.UUID `json:"flag_ids" validate:"required,max=50"` // Lower limit for destructive ops
	Reason  string      `json:"reason" validate:"required"`
}

// BulkUpdateRolloutRequest updates rollout percentage for multiple flags
type BulkUpdateRolloutRequest struct {
	Updates []RolloutUpdate `json:"updates" validate:"required,max=100"`
	Reason  string          `json:"reason" validate:"required"`
}

type RolloutUpdate struct {
	FlagName   string `json:"flag_name" validate:"required"`
	Percentage int32  `json:"percentage" validate:"min=0,max=100"`
}

// BulkOperationResult contains results of bulk operations
type BulkOperationResult struct {
	TotalRequested int                       `json:"total_requested"`
	Successful     int                       `json:"successful"`
	Failed         int                       `json:"failed"`
	Results        []BulkOperationItemResult `json:"results"`
	Summary        BulkOperationSummary      `json:"summary"`
	ExecutedAt     time.Time                 `json:"executed_at"`
	ExecutionTime  time.Duration             `json:"execution_time"`
}

type BulkOperationItemResult struct {
	Identifier string `json:"identifier"` // Flag name or ID
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
}

type BulkOperationSummary struct {
	Operation       string         `json:"operation"`
	SuccessRate     float64        `json:"success_rate"`
	AverageTime     float64        `json:"average_time_ms"`
	ErrorCategories map[string]int `json:"error_categories"`
}

// Flag Templates
type FlagTemplate struct {
	ID              uuid.UUID        `json:"id"`
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	Category        string           `json:"category"`
	FlagType        FlagType         `json:"flag_type"`
	DefaultValue    any              `json:"default_value"`
	RolloutStrategy *RolloutStrategy `json:"rollout_strategy,omitempty"`
	TargetAudience  map[string]any   `json:"target_audience,omitempty"`
	Metadata        map[string]any   `json:"metadata,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

type CreateFlagTemplateRequest struct {
	Name            string           `json:"name" validate:"required"`
	Description     string           `json:"description" validate:"required"`
	Category        string           `json:"category" validate:"required"`
	FlagType        FlagType         `json:"flag_type" validate:"required"`
	DefaultValue    any              `json:"default_value" validate:"required"`
	RolloutStrategy *RolloutStrategy `json:"rollout_strategy,omitempty"`
	TargetAudience  map[string]any   `json:"target_audience,omitempty"`
	Metadata        map[string]any   `json:"metadata,omitempty"`
}

type ApplyTemplateRequest struct {
	TemplateID uuid.UUID      `json:"template_id" validate:"required"`
	FlagName   string         `json:"flag_name" validate:"required"`
	Overrides  map[string]any `json:"overrides,omitempty"`
}

// Rollout Strategies
type RolloutStrategy struct {
	ID            uuid.UUID           `json:"id"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	Type          RolloutStrategyType `json:"type"`
	Configuration map[string]any      `json:"configuration"`
	CreatedAt     time.Time           `json:"created_at"`
}

type RolloutStrategyType string

const (
	RolloutStrategyPercentage    RolloutStrategyType = "percentage"
	RolloutStrategyUserAttribute RolloutStrategyType = "user_attribute"
	RolloutStrategyGradual       RolloutStrategyType = "gradual"
	RolloutStrategyCanary        RolloutStrategyType = "canary"
)

type CreateRolloutStrategyRequest struct {
	Name          string              `json:"name" validate:"required"`
	Description   string              `json:"description" validate:"required"`
	Type          RolloutStrategyType `json:"type" validate:"required"`
	Configuration map[string]any      `json:"configuration" validate:"required"`
}

// System Health and Metrics
type SystemHealthResult struct {
	Status             string                     `json:"status"`
	Timestamp          time.Time                  `json:"timestamp"`
	Version            string                     `json:"version"`
	DatabaseStatus     string                     `json:"database_status"`
	CacheStatus        string                     `json:"cache_status"`
	ComponentHealth    map[string]ComponentHealth `json:"component_health"`
	OverallScore       int                        `json:"overall_score"` // 0-100
	RecommendedActions []string                   `json:"recommended_actions,omitempty"`
}

type ComponentHealth struct {
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	ErrorRate    float64       `json:"error_rate"`
	LastChecked  time.Time     `json:"last_checked"`
	Details      string        `json:"details,omitempty"`
}

type SystemMetricsRequest struct {
	TimeRange TimeRange `json:"time_range"`
	Metrics   []string  `json:"metrics,omitempty"` // If empty, return all
}

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type SystemMetricsResult struct {
	TimeRange          TimeRange                `json:"time_range"`
	GeneratedAt        time.Time                `json:"generated_at"`
	FlagMetrics        FlagSystemMetrics        `json:"flag_metrics"`
	PerformanceMetrics PerformanceSystemMetrics `json:"performance_metrics"`
	UsageMetrics       UsageSystemMetrics       `json:"usage_metrics"`
	TrendAnalysis      TrendAnalysis            `json:"trend_analysis"`
}

type FlagSystemMetrics struct {
	TotalFlags       int            `json:"total_flags"`
	ActiveFlags      int            `json:"active_flags"`
	FlagsByType      map[string]int `json:"flags_by_type"`
	FlagsByTenant    map[string]int `json:"flags_by_tenant"`
	AverageRollout   float64        `json:"average_rollout"`
	EvaluationVolume int64          `json:"evaluation_volume"`
	CreatedToday     int            `json:"created_today"`
	ModifiedToday    int            `json:"modified_today"`
}

type PerformanceSystemMetrics struct {
	AverageEvaluationTime time.Duration `json:"average_evaluation_time"`
	CacheHitRate          float64       `json:"cache_hit_rate"`
	ErrorRate             float64       `json:"error_rate"`
	ThroughputQPS         float64       `json:"throughput_qps"`
	P95ResponseTime       time.Duration `json:"p95_response_time"`
	P99ResponseTime       time.Duration `json:"p99_response_time"`
}

type UsageSystemMetrics struct {
	UniqueUsers     int             `json:"unique_users"`
	UniqueTenants   int             `json:"unique_tenants"`
	TopFlags        []FlagUsageStat `json:"top_flags"`
	UsageByHour     []HourlyUsage   `json:"usage_by_hour"`
	GeographicUsage map[string]int  `json:"geographic_usage"`
}

type FlagUsageStat struct {
	FlagName        string  `json:"flag_name"`
	EvaluationCount int64   `json:"evaluation_count"`
	UniqueUsers     int     `json:"unique_users"`
	SuccessRate     float64 `json:"success_rate"`
}

type HourlyUsage struct {
	Hour            time.Time `json:"hour"`
	EvaluationCount int64     `json:"evaluation_count"`
	UniqueUsers     int       `json:"unique_users"`
	ErrorCount      int       `json:"error_count"`
}

type TrendAnalysis struct {
	GrowthRate      float64               `json:"growth_rate"` // Percentage growth
	Seasonality     map[string]float64    `json:"seasonality"` // Day/hour patterns
	Anomalies       []AnomalyDetection    `json:"anomalies"`
	Forecasting     ForecastData          `json:"forecasting"`
	Recommendations []TrendRecommendation `json:"recommendations"`
}

type AnomalyDetection struct {
	Type        string    `json:"type"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Value       float64   `json:"value"`
	Expected    float64   `json:"expected"`
	Confidence  float64   `json:"confidence"`
}

type ForecastData struct {
	NextHourPrediction int64   `json:"next_hour_prediction"`
	NextDayPrediction  int64   `json:"next_day_prediction"`
	NextWeekPrediction int64   `json:"next_week_prediction"`
	ConfidenceInterval float64 `json:"confidence_interval"`
	SeasonalityFactor  float64 `json:"seasonality_factor"`
}

type TrendRecommendation struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Effort      string `json:"effort"`
}

// Usage Analytics
type UsageAnalyticsRequest struct {
	TimeRange    TimeRange `json:"time_range"`
	TenantFilter []string  `json:"tenant_filter,omitempty"`
	FlagFilter   []string  `json:"flag_filter,omitempty"`
	Granularity  string    `json:"granularity"` // hour, day, week, month
}

type UsageAnalyticsResult struct {
	TimeRange        TimeRange         `json:"time_range"`
	GeneratedAt      time.Time         `json:"generated_at"`
	Summary          UsageSummary      `json:"summary"`
	TimeSeries       []TimeSeriesPoint `json:"time_series"`
	TopPerformers    []PerformerStat   `json:"top_performers"`
	BehaviorAnalysis BehaviorAnalysis  `json:"behavior_analysis"`
}

type UsageSummary struct {
	TotalEvaluations    int64   `json:"total_evaluations"`
	UniqueFlags         int     `json:"unique_flags"`
	UniqueUsers         int     `json:"unique_users"`
	SuccessRate         float64 `json:"success_rate"`
	AverageResponseTime float64 `json:"average_response_time_ms"`
}

type TimeSeriesPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	Evaluations int64     `json:"evaluations"`
	UniqueUsers int       `json:"unique_users"`
	Errors      int       `json:"errors"`
}

type PerformerStat struct {
	Name   string  `json:"name"`
	Value  int64   `json:"value"`
	Change float64 `json:"change"` // Percentage change from previous period
}

type BehaviorAnalysis struct {
	PeakUsageHours   []int                  `json:"peak_usage_hours"`
	UserSegments     map[string]UserSegment `json:"user_segments"`
	FlagCorrelations []FlagCorrelation      `json:"flag_correlations"`
	AdoptionMetrics  AdoptionMetrics        `json:"adoption_metrics"`
}

type UserSegment struct {
	Name            string   `json:"name"`
	UserCount       int      `json:"user_count"`
	AvgEvaluations  float64  `json:"avg_evaluations"`
	PreferredFlags  []string `json:"preferred_flags"`
	BehaviorPattern string   `json:"behavior_pattern"`
}

type FlagCorrelation struct {
	Flag1       string  `json:"flag1"`
	Flag2       string  `json:"flag2"`
	Correlation float64 `json:"correlation"`
	Confidence  float64 `json:"confidence"`
}

type AdoptionMetrics struct {
	NewFlagsThisPeriod    int     `json:"new_flags_this_period"`
	AdoptionRate          float64 `json:"adoption_rate"`
	TimeToFirstEvaluation float64 `json:"time_to_first_evaluation_hours"`
	StickinessRate        float64 `json:"stickiness_rate"`
}

// Emergency Controls
type EmergencyActionResult struct {
	ActionType      string    `json:"action_type"`
	AffectedFlags   int       `json:"affected_flags"`
	ExecutedAt      time.Time `json:"executed_at"`
	Reason          string    `json:"reason"`
	RollbackToken   string    `json:"rollback_token"`
	EstimatedImpact string    `json:"estimated_impact"`
}

// Cache Management
type CacheWarmupRequest struct {
	TenantIDs  []uuid.UUID `json:"tenant_ids,omitempty"`
	FlagNames  []string    `json:"flag_names,omitempty"`
	Priority   string      `json:"priority"`    // high, normal, low
	WarmupType string      `json:"warmup_type"` // full, partial, smart
}

type CacheClearRequest struct {
	TenantIDs []uuid.UUID `json:"tenant_ids,omitempty"`
	FlagNames []string    `json:"flag_names,omitempty"`
	CacheType string      `json:"cache_type"` // flags, evaluations, stats, all
}

type CacheOperationResult struct {
	OperationType  string        `json:"operation_type"`
	ExecutedAt     time.Time     `json:"executed_at"`
	ExecutionTime  time.Duration `json:"execution_time"`
	ItemsProcessed int           `json:"items_processed"`
	Success        bool          `json:"success"`
	Details        string        `json:"details,omitempty"`
}

type CacheStatsResult struct {
	GeneratedAt     time.Time                   `json:"generated_at"`
	OverallStats    CacheOverallStats           `json:"overall_stats"`
	CacheTypeStats  map[string]CacheTypeStats   `json:"cache_type_stats"`
	TenantStats     map[string]CacheTenantStats `json:"tenant_stats"`
	Recommendations []CacheRecommendation       `json:"recommendations"`
}

type CacheOverallStats struct {
	TotalKeys      int64   `json:"total_keys"`
	TotalMemoryMB  float64 `json:"total_memory_mb"`
	HitRate        float64 `json:"hit_rate"`
	MissRate       float64 `json:"miss_rate"`
	EvictionRate   float64 `json:"eviction_rate"`
	AverageKeySize float64 `json:"average_key_size_bytes"`
}

type CacheTypeStats struct {
	KeyCount   int64   `json:"key_count"`
	MemoryMB   float64 `json:"memory_mb"`
	HitRate    float64 `json:"hit_rate"`
	AverageTTL int     `json:"average_ttl_seconds"`
}

type CacheTenantStats struct {
	TenantID string  `json:"tenant_id"`
	KeyCount int64   `json:"key_count"`
	MemoryMB float64 `json:"memory_mb"`
	HitRate  float64 `json:"hit_rate"`
}

type CacheRecommendation struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Action      string `json:"action"`
}

// List requests
type ListTemplatesRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Category string `json:"category,omitempty"`
	FlagType string `json:"flag_type,omitempty"`
}

type ListTemplatesResponse struct {
	Templates []*FlagTemplate `json:"templates"`
	Total     int64           `json:"total"`
	Page      int             `json:"page"`
	PageSize  int             `json:"page_size"`
}
