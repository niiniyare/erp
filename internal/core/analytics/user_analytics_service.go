package analytics

//go:generate go run go.uber.org/mock/mockgen -source=user_analytics_service.go -destination=mock.go -package=analytics


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
	"go.opentelemetry.io/otel/codes"

	"github.com/niiniyare/erp/internal/core/access/conditional"
	"github.com/niiniyare/erp/internal/core/audit"
)

// Type aliases for external dependencies
type AuditService = audit.Service
type AuditEvent = audit.AuditEvent
type DeviceInfo = conditional.DeviceInfo
type LocationInfo = conditional.LocationInfo

const (
	AuditSeverityHigh = "high"
)

// UserBehaviorPattern represents a user's behavioral pattern
type UserBehaviorPattern struct {
	UserID         uuid.UUID `json:"user_id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	AnalysisPeriod DateRange `json:"analysis_period"`
	LastUpdated    time.Time `json:"last_updated"`

	// Login patterns
	LoginPatterns LoginPatterns `json:"login_patterns"`

	// Activity patterns
	ActivityPatterns ActivityPatterns `json:"activity_patterns"`

	// Access patterns
	AccessPatterns AccessPatterns `json:"access_patterns"`

	// Productivity metrics
	ProductivityMetrics ProductivityMetrics `json:"productivity_metrics"`

	// Security profile
	SecurityProfile SecurityProfile `json:"security_profile"`

	// Risk assessment
	RiskAssessment UserRiskAssessment `json:"risk_assessment"`

	// Anomaly detection
	AnomalyIndicators []AnomalyIndicator `json:"anomaly_indicators"`

	// Baseline establishment
	BaselineEstablished bool    `json:"baseline_established"`
	BaselineConfidence  float64 `json:"baseline_confidence"` // 0.0 to 1.0
}

// DateRange represents a date range for analysis
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// LoginPatterns represents user login behavior patterns
type LoginPatterns struct {
	TypicalLoginHours      []int               `json:"typical_login_hours"` // Hours of day (0-23)
	TypicalLoginDays       []time.Weekday      `json:"typical_login_days"`  // Days of week
	AverageSessionDuration time.Duration       `json:"average_session_duration"`
	LoginFrequency         LoginFrequency      `json:"login_frequency"`
	LocationConsistency    LocationConsistency `json:"location_consistency"`
	DeviceConsistency      DeviceConsistency   `json:"device_consistency"`
	LoginTimeVariance      float64             `json:"login_time_variance"`     // Statistical variance
	WeekendLoginFrequency  float64             `json:"weekend_login_frequency"` // Ratio of weekend logins
}

// LoginFrequency represents login frequency patterns
type LoginFrequency struct {
	DailyAverage     float64           `json:"daily_average"`
	WeeklyAverage    float64           `json:"weekly_average"`
	MonthlyAverage   float64           `json:"monthly_average"`
	PeakLoginTimes   []HourlyFrequency `json:"peak_login_times"`
	ConsistencyScore float64           `json:"consistency_score"` // 0.0 to 1.0
}

// HourlyFrequency represents login frequency by hour
type HourlyFrequency struct {
	Hour      int     `json:"hour"`
	Frequency float64 `json:"frequency"`
}

// LocationConsistency represents location pattern consistency
type LocationConsistency struct {
	PrimaryLocations        []LocationFrequency `json:"primary_locations"`
	LocationVariance        float64             `json:"location_variance"`
	UnusualLocationCount    int                 `json:"unusual_location_count"`
	LocationChangeFrequency float64             `json:"location_change_frequency"`
	TrustedLocationRatio    float64             `json:"trusted_location_ratio"`
}

// LocationFrequency represents location usage frequency
type LocationFrequency struct {
	Country   string  `json:"country"`
	Region    string  `json:"region"`
	City      string  `json:"city"`
	Frequency float64 `json:"frequency"`
	IsTrusted bool    `json:"is_trusted"`
}

// DeviceConsistency represents device usage patterns
type DeviceConsistency struct {
	PrimaryDevices     []DeviceFrequency  `json:"primary_devices"`
	DeviceVariance     float64            `json:"device_variance"`
	NewDeviceFrequency float64            `json:"new_device_frequency"`
	TrustedDeviceRatio float64            `json:"trusted_device_ratio"`
	BrowserConsistency BrowserConsistency `json:"browser_consistency"`
}

// DeviceFrequency represents device usage frequency
type DeviceFrequency struct {
	DeviceFingerprint string    `json:"device_fingerprint"`
	DeviceType        string    `json:"device_type"`
	OperatingSystem   string    `json:"operating_system"`
	Browser           string    `json:"browser"`
	Frequency         float64   `json:"frequency"`
	IsTrusted         bool      `json:"is_trusted"`
	LastUsed          time.Time `json:"last_used"`
}

// BrowserConsistency represents browser usage patterns
type BrowserConsistency struct {
	PrimaryBrowser         string   `json:"primary_browser"`
	BrowserSwitchFrequency float64  `json:"browser_switch_frequency"`
	UnusualBrowserUsage    []string `json:"unusual_browser_usage"`
}

// ActivityPatterns represents user activity behavior patterns
type ActivityPatterns struct {
	PeakActivityHours     []int                  `json:"peak_activity_hours"`
	ActivityDistribution  map[string]float64     `json:"activity_distribution"` // activity_type -> frequency
	ResourceUsagePatterns []ResourceUsagePattern `json:"resource_usage_patterns"`
	NavigationPatterns    NavigationPatterns     `json:"navigation_patterns"`
	WorkflowPatterns      []WorkflowPattern      `json:"workflow_patterns"`
	IdleTimePatterns      IdleTimePatterns       `json:"idle_time_patterns"`
}

// ResourceUsagePattern represents how users interact with resources
type ResourceUsagePattern struct {
	ResourceType         string        `json:"resource_type"`
	ResourceName         string        `json:"resource_name"`
	AccessFrequency      float64       `json:"access_frequency"`
	AverageSessionTime   time.Duration `json:"average_session_time"`
	PreferredAccessTimes []int         `json:"preferred_access_times"` // Hours of day
	UsageConsistency     float64       `json:"usage_consistency"`
}

// NavigationPatterns represents user navigation behavior
type NavigationPatterns struct {
	CommonPathways       []NavigationPathway      `json:"common_pathways"`
	PageViewDuration     map[string]time.Duration `json:"page_view_duration"` // page -> average duration
	BounceRate           float64                  `json:"bounce_rate"`
	NavigationEfficiency float64                  `json:"navigation_efficiency"` // 0.0 to 1.0
}

// NavigationPathway represents a common navigation path
type NavigationPathway struct {
	Path      []string      `json:"path"` // sequence of pages/actions
	Frequency float64       `json:"frequency"`
	Duration  time.Duration `json:"duration"`
}

// WorkflowPattern represents user workflow behavior
type WorkflowPattern struct {
	WorkflowType              string        `json:"workflow_type"`
	StepSequence              []string      `json:"step_sequence"`
	AverageCompletionTime     time.Duration `json:"average_completion_time"`
	SuccessRate               float64       `json:"success_rate"`
	CommonErrors              []string      `json:"common_errors"`
	OptimizationOpportunities []string      `json:"optimization_opportunities"`
}

// IdleTimePatterns represents user idle behavior
type IdleTimePatterns struct {
	AverageIdleTime      time.Duration            `json:"average_idle_time"`
	MaxIdleTime          time.Duration            `json:"max_idle_time"`
	IdleFrequency        float64                  `json:"idle_frequency"`
	IdleTimeDistribution map[string]time.Duration `json:"idle_time_distribution"` // time_range -> duration
}

// AccessPatterns represents user access behavior patterns
type AccessPatterns struct {
	PermissionUsagePatterns []PermissionUsagePattern `json:"permission_usage_patterns"`
	ElevatedAccessPatterns  ElevatedAccessPatterns   `json:"elevated_access_patterns"`
	DataAccessPatterns      []DataAccessPattern      `json:"data_access_patterns"`
	APIUsagePatterns        []APIUsagePattern        `json:"api_usage_patterns"`
}

// PermissionUsagePattern represents how users use their permissions
type PermissionUsagePattern struct {
	PermissionName           string   `json:"permission_name"`
	UsageFrequency           float64  `json:"usage_frequency"`
	TypicalUsageTimes        []int    `json:"typical_usage_times"` // Hours of day
	UnusedPermissions        []string `json:"unused_permissions"`
	OverprivilegedIndicators []string `json:"overprivileged_indicators"`
}

// ElevatedAccessPatterns represents elevated access usage
type ElevatedAccessPatterns struct {
	ElevationFrequency       float64        `json:"elevation_frequency"`
	AverageElevationDuration time.Duration  `json:"average_elevation_duration"`
	ElevationReasons         map[string]int `json:"elevation_reasons"` // reason -> count
	ElevationTimes           []int          `json:"elevation_times"`   // Typical hours
	ElevationSuccessRate     float64        `json:"elevation_success_rate"`
}

// DataAccessPattern represents data access behavior
type DataAccessPattern struct {
	DataType            string               `json:"data_type"`
	AccessFrequency     float64              `json:"access_frequency"`
	ReadWriteRatio      float64              `json:"read_write_ratio"`
	DataVolumeAccessed  DataVolumeMetrics    `json:"data_volume_accessed"`
	SensitiveDataAccess SensitiveDataMetrics `json:"sensitive_data_access"`
}

// DataVolumeMetrics represents data volume access metrics
type DataVolumeMetrics struct {
	AverageRecordsPerSession int     `json:"average_records_per_session"`
	MaxRecordsPerSession     int     `json:"max_records_per_session"`
	TotalDataAccessed        int64   `json:"total_data_accessed"` // bytes
	BulkOperationFrequency   float64 `json:"bulk_operation_frequency"`
}

// SensitiveDataMetrics represents sensitive data access metrics
type SensitiveDataMetrics struct {
	SensitiveDataTypes   []string `json:"sensitive_data_types"`
	AccessFrequency      float64  `json:"access_frequency"`
	ComplianceViolations int      `json:"compliance_violations"`
	UnauthorizedAttempts int      `json:"unauthorized_attempts"`
}

// APIUsagePattern represents API usage behavior
type APIUsagePattern struct {
	APIEndpoint         string        `json:"api_endpoint"`
	CallFrequency       float64       `json:"call_frequency"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	ErrorRate           float64       `json:"error_rate"`
	UsageSpikes         []UsageSpike  `json:"usage_spikes"`
}

// UsageSpike represents unusual usage spikes
type UsageSpike struct {
	Timestamp time.Time     `json:"timestamp"`
	Magnitude float64       `json:"magnitude"` // multiplier of normal usage
	Duration  time.Duration `json:"duration"`
}

// ProductivityMetrics represents user productivity measurements
type ProductivityMetrics struct {
	TaskCompletionRate    float64                `json:"task_completion_rate"`
	AverageTaskDuration   time.Duration          `json:"average_task_duration"`
	DocumentsCreated      ProductivityCounter    `json:"documents_created"`
	TransactionsProcessed ProductivityCounter    `json:"transactions_processed"`
	ApprovalsCompleted    ProductivityCounter    `json:"approvals_completed"`
	CollaborationMetrics  CollaborationMetrics   `json:"collaboration_metrics"`
	FeatureAdoption       FeatureAdoptionMetrics `json:"feature_adoption"`
	EfficiencyTrends      []EfficiencyDataPoint  `json:"efficiency_trends"`
}

// ProductivityCounter represents productivity counters
type ProductivityCounter struct {
	Daily   float64 `json:"daily"`
	Weekly  float64 `json:"weekly"`
	Monthly float64 `json:"monthly"`
	Total   int64   `json:"total"`
}

// CollaborationMetrics represents collaboration behavior
type CollaborationMetrics struct {
	SharedDocuments        int     `json:"shared_documents"`
	CollaborationSessions  int     `json:"collaboration_sessions"`
	TeamInteractionScore   float64 `json:"team_interaction_score"` // 0.0 to 1.0
	CommunicationFrequency float64 `json:"communication_frequency"`
}

// FeatureAdoptionMetrics represents feature adoption patterns
type FeatureAdoptionMetrics struct {
	NewFeaturesAdopted    []string           `json:"new_features_adopted"`
	FeatureUsageFrequency map[string]float64 `json:"feature_usage_frequency"` // feature -> frequency
	UnderutilizedFeatures []string           `json:"underutilized_features"`
	AdoptionRate          float64            `json:"adoption_rate"` // 0.0 to 1.0
}

// EfficiencyDataPoint represents efficiency measurement over time
type EfficiencyDataPoint struct {
	Timestamp       time.Time     `json:"timestamp"`
	EfficiencyScore float64       `json:"efficiency_score"` // 0.0 to 1.0
	TasksCompleted  int           `json:"tasks_completed"`
	TimeSpent       time.Duration `json:"time_spent"`
}

// SecurityProfile represents user security behavior profile
type SecurityProfile struct {
	PasswordChangeFrequency float64                  `json:"password_change_frequency"`
	MFAUsageConsistency     float64                  `json:"mfa_usage_consistency"`
	SecurityAwareness       SecurityAwarenessMetrics `json:"security_awareness"`
	ComplianceScore         float64                  `json:"compliance_score"` // 0.0 to 1.0
	RiskBehaviors           []RiskBehavior           `json:"risk_behaviors"`
	SecurityIncidents       []SecurityIncident       `json:"security_incidents"`
}

// SecurityAwarenessMetrics represents security awareness indicators
type SecurityAwarenessMetrics struct {
	PhishingTestResults    PhishingTestResults `json:"phishing_test_results"`
	SecurityTrainingScore  float64             `json:"security_training_score"` // 0.0 to 1.0
	PolicyComplianceRate   float64             `json:"policy_compliance_rate"`  // 0.0 to 1.0
	ReportedSecurityIssues int                 `json:"reported_security_issues"`
}

// PhishingTestResults represents phishing simulation results
type PhishingTestResults struct {
	TestsConducted   int    `json:"tests_conducted"`
	TestsPassed      int    `json:"tests_passed"`
	TestsFailed      int    `json:"tests_failed"`
	ImprovementTrend string `json:"improvement_trend"` // IMPROVING, DECLINING, STABLE
}

// RiskBehavior represents risky user behavior
type RiskBehavior struct {
	BehaviorType   string    `json:"behavior_type"`
	Frequency      float64   `json:"frequency"`
	Severity       string    `json:"severity"` // LOW, MEDIUM, HIGH, CRITICAL
	LastOccurrence time.Time `json:"last_occurrence"`
	TrendDirection string    `json:"trend_direction"` // INCREASING, DECREASING, STABLE
}

// SecurityIncident represents security incidents involving the user
type SecurityIncident struct {
	IncidentID         uuid.UUID `json:"incident_id"`
	IncidentType       string    `json:"incident_type"`
	Severity           string    `json:"severity"`
	Timestamp          time.Time `json:"timestamp"`
	Resolution         string    `json:"resolution"`
	UserAccountability string    `json:"user_accountability"` // PRIMARY, SECONDARY, NONE
}

// UserRiskAssessment represents comprehensive user risk assessment
type UserRiskAssessment struct {
	OverallRiskScore      int                      `json:"overall_risk_score"` // 0-100
	RiskLevel             string                   `json:"risk_level"`         // LOW, MEDIUM, HIGH, CRITICAL
	RiskCategories        map[string]int           `json:"risk_categories"`    // category -> score
	BehavioralAnomalies   []BehavioralAnomaly      `json:"behavioral_anomalies"`
	PredictiveRiskFactors []PredictiveRiskFactor   `json:"predictive_risk_factors"`
	RiskTrends            []RiskDataPoint          `json:"risk_trends"`
	Recommendations       []SecurityRecommendation `json:"recommendations"`
	LastAssessment        time.Time                `json:"last_assessment"`
	NextAssessment        time.Time                `json:"next_assessment"`
}

// BehavioralAnomaly represents detected behavioral anomalies
type BehavioralAnomaly struct {
	AnomalyType       string    `json:"anomaly_type"`
	Severity          string    `json:"severity"`
	Confidence        float64   `json:"confidence"` // 0.0 to 1.0
	Description       string    `json:"description"`
	DetectedAt        time.Time `json:"detected_at"`
	BaselineDeviation float64   `json:"baseline_deviation"` // Standard deviations
	AffectedMetrics   []string  `json:"affected_metrics"`
}

// PredictiveRiskFactor represents factors that predict future risk
type PredictiveRiskFactor struct {
	FactorName       string  `json:"factor_name"`
	RiskContribution float64 `json:"risk_contribution"` // 0.0 to 1.0
	Trend            string  `json:"trend"`             // INCREASING, DECREASING, STABLE
	PredictedImpact  string  `json:"predicted_impact"`  // LOW, MEDIUM, HIGH
	Confidence       float64 `json:"confidence"`        // 0.0 to 1.0
}

// RiskDataPoint represents risk score over time
type RiskDataPoint struct {
	Timestamp           time.Time `json:"timestamp"`
	RiskScore           int       `json:"risk_score"`
	RiskLevel           string    `json:"risk_level"`
	ContributingFactors []string  `json:"contributing_factors"`
}

// SecurityRecommendation represents security recommendations for the user
type SecurityRecommendation struct {
	RecommendationType       string `json:"recommendation_type"`
	Priority                 string `json:"priority"` // LOW, MEDIUM, HIGH, URGENT
	Title                    string `json:"title"`
	Description              string `json:"description"`
	ActionRequired           string `json:"action_required"`
	ExpectedImpact           string `json:"expected_impact"`
	ImplementationComplexity string `json:"implementation_complexity"` // LOW, MEDIUM, HIGH
}

// AnomalyIndicator represents detected anomalies
type AnomalyIndicator struct {
	IndicatorType      string    `json:"indicator_type"`
	Severity           string    `json:"severity"`
	Confidence         float64   `json:"confidence"`
	DetectedAt         time.Time `json:"detected_at"`
	Description        string    `json:"description"`
	AffectedBehaviors  []string  `json:"affected_behaviors"`
	RecommendedActions []string  `json:"recommended_actions"`
}

// UserAnalyticsService handles user behavior analysis and risk assessment
type UserAnalyticsService interface {
	// Behavior analysis
	AnalyzeUserBehavior(ctx context.Context, userID uuid.UUID, analysisWindow time.Duration) (*UserBehaviorPattern, error)
	UpdateBehaviorPattern(ctx context.Context, userID uuid.UUID, newActivity *UserActivity) error
	GetBehaviorPattern(ctx context.Context, userID uuid.UUID) (*UserBehaviorPattern, error)

	// Risk assessment
	AssessUserRisk(ctx context.Context, userID uuid.UUID) (*UserRiskAssessment, error)
	UpdateRiskAssessment(ctx context.Context, userID uuid.UUID, assessment *UserRiskAssessment) error
	GetRiskTrends(ctx context.Context, userID uuid.UUID, period time.Duration) ([]RiskDataPoint, error)

	// Anomaly detection
	DetectAnomalies(ctx context.Context, userID uuid.UUID, activity *UserActivity) ([]AnomalyIndicator, error)
	GetUserAnomalies(ctx context.Context, userID uuid.UUID, severity []string) ([]AnomalyIndicator, error)

	// Baseline establishment
	EstablishBaseline(ctx context.Context, userID uuid.UUID, minDataPoints int) error
	IsBaselineEstablished(ctx context.Context, userID uuid.UUID) (bool, float64, error) // established, confidence

	// Predictive analytics
	PredictUserRisk(ctx context.Context, userID uuid.UUID, forecastDays int) (*RiskPrediction, error)
	IdentifyRiskFactors(ctx context.Context, userID uuid.UUID) ([]PredictiveRiskFactor, error)

	// Recommendations
	GenerateSecurityRecommendations(ctx context.Context, userID uuid.UUID) ([]SecurityRecommendation, error)
	GetPersonalizedInsights(ctx context.Context, userID uuid.UUID) (*UserInsights, error)

	// Comparative analysis
	CompareToPeerGroup(ctx context.Context, userID uuid.UUID) (*PeerComparison, error)
	GetBenchmarkMetrics(ctx context.Context, entityID uuid.UUID) (*BenchmarkMetrics, error)
}

// UserActivity represents a user activity event for analysis
type UserActivity struct {
	UserID           uuid.UUID      `json:"user_id"`
	SessionID        uuid.UUID      `json:"session_id"`
	ActivityType     string         `json:"activity_type"`
	ResourceAccessed string         `json:"resource_accessed"`
	ActionPerformed  string         `json:"action_performed"`
	Timestamp        time.Time      `json:"timestamp"`
	Duration         time.Duration  `json:"duration"`
	IPAddress        string         `json:"ip_address"`
	UserAgent        string         `json:"user_agent"`
	DeviceInfo       *DeviceInfo    `json:"device_info,omitempty"`
	LocationInfo     *LocationInfo  `json:"location_info,omitempty"`
	Success          bool           `json:"success"`
	ErrorCode        string         `json:"error_code,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

// RiskPrediction represents predicted risk assessment
type RiskPrediction struct {
	UserID                 uuid.UUID              `json:"user_id"`
	PredictionDate         time.Time              `json:"prediction_date"`
	ForecastPeriod         time.Duration          `json:"forecast_period"`
	PredictedRiskScore     int                    `json:"predicted_risk_score"`
	PredictedRiskLevel     string                 `json:"predicted_risk_level"`
	Confidence             float64                `json:"confidence"`
	KeyRiskFactors         []PredictiveRiskFactor `json:"key_risk_factors"`
	RecommendedActions     []string               `json:"recommended_actions"`
	EarlyWarningIndicators []string               `json:"early_warning_indicators"`
}

// UserInsights represents personalized user insights
type UserInsights struct {
	UserID                  uuid.UUID `json:"user_id"`
	GeneratedAt             time.Time `json:"generated_at"`
	ProductivityInsights    []string  `json:"productivity_insights"`
	SecurityInsights        []string  `json:"security_insights"`
	UsageOptimizations      []string  `json:"usage_optimizations"`
	LearningRecommendations []string  `json:"learning_recommendations"`
	EfficiencyTips          []string  `json:"efficiency_tips"`
}

// PeerComparison represents comparison with peer group
type PeerComparison struct {
	UserID                 uuid.UUID          `json:"user_id"`
	PeerGroupSize          int                `json:"peer_group_size"`
	PerformancePercentile  float64            `json:"performance_percentile"`  // 0.0 to 1.0
	SecurityPercentile     float64            `json:"security_percentile"`     // 0.0 to 1.0
	ProductivityPercentile float64            `json:"productivity_percentile"` // 0.0 to 1.0
	ComparisonMetrics      map[string]float64 `json:"comparison_metrics"`      // metric -> percentile
	StrengthAreas          []string           `json:"strength_areas"`
	ImprovementAreas       []string           `json:"improvement_areas"`
}

// BenchmarkMetrics represents benchmark metrics for an entity
type BenchmarkMetrics struct {
	EntityID          uuid.UUID                     `json:"entity_id"`
	GeneratedAt       time.Time                     `json:"generated_at"`
	UserCount         int                           `json:"user_count"`
	AverageMetrics    map[string]float64            `json:"average_metrics"`
	MedianMetrics     map[string]float64            `json:"median_metrics"`
	PercentileMetrics map[string]map[string]float64 `json:"percentile_metrics"` // metric -> percentile -> value
	TopPerformers     []uuid.UUID                   `json:"top_performers"`
	RiskDistribution  map[string]int                `json:"risk_distribution"` // risk_level -> count
}

// userAnalyticsService implements UserAnalyticsService
type userAnalyticsService struct {
	tracing      tracing.TracingService
	metrics      metrics.MetricsProvider
	auditService AuditService
	// TODO: Add analytics repository when implemented
	// analyticsRepo   UserAnalyticsRepository
}

// NewUserAnalyticsService creates a new user analytics service
func NewUserAnalyticsService(
	tracing tracing.TracingService,
	metrics metrics.MetricsProvider,
	auditService AuditService,
) UserAnalyticsService {
	return &userAnalyticsService{
		tracing:      tracing,
		metrics:      metrics,
		auditService: auditService,
	}
}

// AnalyzeUserBehavior performs comprehensive user behavior analysis
func (s *userAnalyticsService) AnalyzeUserBehavior(ctx context.Context, userID uuid.UUID, analysisWindow time.Duration) (*UserBehaviorPattern, error) {
	ctx, span := s.tracing.StartSpan(ctx, "userAnalyticsService.AnalyzeUserBehavior")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("analysis_window", analysisWindow.String()),
	)

	logger.Info("Analyzing user behavior", logger.Fields{
		"user_id":         userID,
		"analysis_window": analysisWindow,
	})

	// Get historical user activities (placeholder - would query database)
	activities := s.getUserActivities(ctx, userID, analysisWindow)

	// Analyze different aspects of behavior
	loginPatterns := s.analyzeLoginPatterns(activities)
	activityPatterns := s.analyzeActivityPatterns(activities)
	accessPatterns := s.analyzeAccessPatterns(activities)
	productivityMetrics := s.analyzeProductivityMetrics(activities)
	securityProfile := s.analyzeSecurityProfile(activities)
	riskAssessment := s.assessRisk(activities)
	anomalyIndicators := s.detectBehaviorAnomalies(activities)

	// Determine if baseline is established
	baselineEstablished := len(activities) >= 100 // Minimum data points
	baselineConfidence := s.calculateBaselineConfidence(activities)

	pattern := &UserBehaviorPattern{
		UserID: userID,
		// TenantID would be retrieved from user context
		AnalysisPeriod:      DateRange{Start: time.Now().Add(-analysisWindow), End: time.Now()},
		LastUpdated:         time.Now(),
		LoginPatterns:       loginPatterns,
		ActivityPatterns:    activityPatterns,
		AccessPatterns:      accessPatterns,
		ProductivityMetrics: productivityMetrics,
		SecurityProfile:     securityProfile,
		RiskAssessment:      riskAssessment,
		AnomalyIndicators:   anomalyIndicators,
		BaselineEstablished: baselineEstablished,
		BaselineConfidence:  baselineConfidence,
	}

	// Store the pattern (placeholder - would save to database)
	s.storeBehaviorPattern(ctx, pattern)

	s.metrics.ObserveHistogram("user_behavior_analysis_duration", float64(time.Since(pattern.LastUpdated).Milliseconds()), map[string]any{
		"user_id": userID.String(),
	})

	return pattern, nil
}

// AssessUserRisk performs comprehensive user risk assessment
func (s *userAnalyticsService) AssessUserRisk(ctx context.Context, userID uuid.UUID) (*UserRiskAssessment, error) {
	ctx, span := s.tracing.StartSpan(ctx, "userAnalyticsService.AssessUserRisk")
	defer span.End()

	logger.Info("Assessing user risk", logger.Fields{
		"user_id": userID,
	})

	// Get current behavior pattern
	pattern, err := s.GetBehaviorPattern(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get behavior pattern")
		return nil, err
	}

	// Calculate risk scores for different categories
	riskCategories := map[string]int{
		"login_behavior":    s.calculateLoginRisk(pattern.LoginPatterns),
		"access_patterns":   s.calculateAccessRisk(pattern.AccessPatterns),
		"security_behavior": s.calculateSecurityRisk(pattern.SecurityProfile),
		"anomaly_risk":      s.calculateAnomalyRisk(pattern.AnomalyIndicators),
		"device_risk":       s.calculateDeviceRisk(pattern.LoginPatterns.DeviceConsistency),
		"location_risk":     s.calculateLocationRisk(pattern.LoginPatterns.LocationConsistency),
	}

	// Calculate overall risk score (weighted average)
	overallRiskScore := s.calculateOverallRiskScore(riskCategories)
	riskLevel := s.getRiskLevel(overallRiskScore)

	// Detect behavioral anomalies
	behavioralAnomalies := s.identifyBehavioralAnomalies(pattern)

	// Identify predictive risk factors
	predictiveFactors := s.identifyPredictiveRiskFactors(pattern, riskCategories)

	// Generate risk trends (placeholder - would use historical data)
	riskTrends := s.generateRiskTrends(userID, 30) // Last 30 days

	// Generate security recommendations
	recommendations := s.generateRiskBasedRecommendations(riskCategories, behavioralAnomalies)

	assessment := &UserRiskAssessment{
		OverallRiskScore:      overallRiskScore,
		RiskLevel:             riskLevel,
		RiskCategories:        riskCategories,
		BehavioralAnomalies:   behavioralAnomalies,
		PredictiveRiskFactors: predictiveFactors,
		RiskTrends:            riskTrends,
		Recommendations:       recommendations,
		LastAssessment:        time.Now(),
		NextAssessment:        time.Now().Add(24 * time.Hour), // Daily assessment
	}

	// Store the assessment
	if err := s.UpdateRiskAssessment(ctx, userID, assessment); err != nil {
		logger.Error("Failed to store risk assessment", logger.Fields{
			"user_id": userID,
			"error":   err.Error(),
		})
	}

	// Log security event if high risk detected
	if overallRiskScore >= 70 {
		contextData, _ := json.Marshal(map[string]any{
			"risk_score":      overallRiskScore,
			"risk_level":      riskLevel,
			"risk_categories": riskCategories,
			"anomalies":       len(behavioralAnomalies),
		})

		s.auditService.Record(ctx, AuditEvent{
			UserID:        userID,
			EventType:     "security_violation",
			EventCategory: "risk_assessment",
			Severity:      AuditSeverityHigh,
			Decision:      "high_risk_detected",
			Reason:        fmt.Sprintf("High risk user detected: score %d", overallRiskScore),
			Context:       contextData,
		})
	}

	s.metrics.ObserveHistogram("user_risk_score", float64(overallRiskScore), map[string]any{
		"risk_level": riskLevel,
	})

	return assessment, nil
}

// DetectAnomalies detects anomalies in user activity
func (s *userAnalyticsService) DetectAnomalies(ctx context.Context, userID uuid.UUID, activity *UserActivity) ([]AnomalyIndicator, error) {
	ctx, span := s.tracing.StartSpan(ctx, "userAnalyticsService.DetectAnomalies")
	defer span.End()

	var anomalies []AnomalyIndicator

	// Get baseline behavior pattern
	pattern, err := s.GetBehaviorPattern(ctx, userID)
	if err != nil || !pattern.BaselineEstablished {
		// Cannot detect anomalies without established baseline
		return anomalies, nil
	}

	// Time-based anomaly detection
	if timeAnomaly := s.detectTimeAnomaly(activity, pattern.LoginPatterns); timeAnomaly != nil {
		anomalies = append(anomalies, *timeAnomaly)
	}

	// Location-based anomaly detection
	if locationAnomaly := s.detectLocationAnomaly(activity, pattern.LoginPatterns.LocationConsistency); locationAnomaly != nil {
		anomalies = append(anomalies, *locationAnomaly)
	}

	// Device-based anomaly detection
	if deviceAnomaly := s.detectDeviceAnomaly(activity, pattern.LoginPatterns.DeviceConsistency); deviceAnomaly != nil {
		anomalies = append(anomalies, *deviceAnomaly)
	}

	// Activity volume anomaly detection
	if volumeAnomaly := s.detectActivityVolumeAnomaly(activity, pattern.ActivityPatterns); volumeAnomaly != nil {
		anomalies = append(anomalies, *volumeAnomaly)
	}

	// Resource access anomaly detection
	if accessAnomaly := s.detectResourceAccessAnomaly(activity, pattern.AccessPatterns); accessAnomaly != nil {
		anomalies = append(anomalies, *accessAnomaly)
	}

	// Log detected anomalies
	if len(anomalies) > 0 {
		logger.Warn("User behavior anomalies detected", logger.Fields{
			"user_id":       userID,
			"anomaly_count": len(anomalies),
			"activity_type": activity.ActivityType,
		})

		s.metrics.IncrementCounter("user_anomalies_detected", map[string]any{
			"user_id":       userID.String(),
			"anomaly_count": len(anomalies),
		})
	}

	return anomalies, nil
}

// EstablishBaseline establishes behavioral baseline for a user
func (s *userAnalyticsService) EstablishBaseline(ctx context.Context, userID uuid.UUID, minDataPoints int) error {
	ctx, span := s.tracing.StartSpan(ctx, "userAnalyticsService.EstablishBaseline")
	defer span.End()

	logger.Info("Establishing user behavioral baseline", logger.Fields{
		"user_id":         userID,
		"min_data_points": minDataPoints,
	})

	// Get sufficient historical data
	analysisWindow := 30 * 24 * time.Hour // 30 days
	activities := s.getUserActivities(ctx, userID, analysisWindow)

	if len(activities) < minDataPoints {
		return fmt.Errorf("insufficient data points for baseline: have %d, need %d", len(activities), minDataPoints)
	}

	// Perform comprehensive analysis to establish baseline
	pattern, err := s.AnalyzeUserBehavior(ctx, userID, analysisWindow)
	if err != nil {
		return fmt.Errorf("failed to analyze behavior for baseline: %w", err)
	}

	// Mark baseline as established
	pattern.BaselineEstablished = true
	pattern.BaselineConfidence = s.calculateBaselineConfidence(activities)

	// Store updated pattern
	s.storeBehaviorPattern(ctx, pattern)

	s.metrics.IncrementCounter("user_baseline_established", map[string]any{
		"user_id": userID.String(),
	})

	return nil
}

// UpdateBehaviorPattern updates behavior pattern with new activity
func (s *userAnalyticsService) UpdateBehaviorPattern(ctx context.Context, userID uuid.UUID, newActivity *UserActivity) error {
	// TODO: Implement incremental pattern updates
	// This would update the existing pattern with new activity data
	// rather than recalculating everything
	return nil
}

// GetBehaviorPattern gets current behavior pattern for a user
func (s *userAnalyticsService) GetBehaviorPattern(ctx context.Context, userID uuid.UUID) (*UserBehaviorPattern, error) {
	// TODO: Implement database retrieval
	// For now, return a sample pattern
	return &UserBehaviorPattern{
		UserID:              userID,
		BaselineEstablished: true,
		BaselineConfidence:  0.85,
		LastUpdated:         time.Now().Add(-time.Hour),
	}, nil
}

// UpdateRiskAssessment updates risk assessment for a user
func (s *userAnalyticsService) UpdateRiskAssessment(ctx context.Context, userID uuid.UUID, assessment *UserRiskAssessment) error {
	// TODO: Implement database storage
	logger.Info("Updated user risk assessment", logger.Fields{
		"user_id":    userID,
		"risk_score": assessment.OverallRiskScore,
		"risk_level": assessment.RiskLevel,
	})
	return nil
}

// GetRiskTrends gets risk trends for a user over a period
func (s *userAnalyticsService) GetRiskTrends(ctx context.Context, userID uuid.UUID, period time.Duration) ([]RiskDataPoint, error) {
	// TODO: Implement database query for historical risk data
	return []RiskDataPoint{}, nil
}

// GetUserAnomalies gets anomalies for a user filtered by severity
func (s *userAnalyticsService) GetUserAnomalies(ctx context.Context, userID uuid.UUID, severity []string) ([]AnomalyIndicator, error) {
	// TODO: Implement database query for anomalies
	return []AnomalyIndicator{}, nil
}

// IsBaselineEstablished checks if baseline is established for a user
func (s *userAnalyticsService) IsBaselineEstablished(ctx context.Context, userID uuid.UUID) (bool, float64, error) {
	pattern, err := s.GetBehaviorPattern(ctx, userID)
	if err != nil {
		return false, 0.0, err
	}
	return pattern.BaselineEstablished, pattern.BaselineConfidence, nil
}

// PredictUserRisk predicts future risk for a user
func (s *userAnalyticsService) PredictUserRisk(ctx context.Context, userID uuid.UUID, forecastDays int) (*RiskPrediction, error) {
	// TODO: Implement machine learning-based risk prediction
	// This would use historical patterns and trends to predict future risk

	assessment, err := s.AssessUserRisk(ctx, userID)
	if err != nil {
		return nil, err
	}

	prediction := &RiskPrediction{
		UserID:                 userID,
		PredictionDate:         time.Now(),
		ForecastPeriod:         time.Duration(forecastDays) * 24 * time.Hour,
		PredictedRiskScore:     assessment.OverallRiskScore, // Simplified - would use ML models
		PredictedRiskLevel:     assessment.RiskLevel,
		Confidence:             0.75, // Placeholder confidence
		KeyRiskFactors:         assessment.PredictiveRiskFactors,
		RecommendedActions:     s.extractActionRecommendations(assessment.Recommendations),
		EarlyWarningIndicators: s.generateEarlyWarningIndicators(assessment),
	}

	return prediction, nil
}

// IdentifyRiskFactors identifies key risk factors for a user
func (s *userAnalyticsService) IdentifyRiskFactors(ctx context.Context, userID uuid.UUID) ([]PredictiveRiskFactor, error) {
	assessment, err := s.AssessUserRisk(ctx, userID)
	if err != nil {
		return nil, err
	}
	return assessment.PredictiveRiskFactors, nil
}

// GenerateSecurityRecommendations generates security recommendations for a user
func (s *userAnalyticsService) GenerateSecurityRecommendations(ctx context.Context, userID uuid.UUID) ([]SecurityRecommendation, error) {
	assessment, err := s.AssessUserRisk(ctx, userID)
	if err != nil {
		return nil, err
	}
	return assessment.Recommendations, nil
}

// GetPersonalizedInsights generates personalized insights for a user
func (s *userAnalyticsService) GetPersonalizedInsights(ctx context.Context, userID uuid.UUID) (*UserInsights, error) {
	pattern, err := s.GetBehaviorPattern(ctx, userID)
	if err != nil {
		return nil, err
	}

	insights := &UserInsights{
		UserID:                  userID,
		GeneratedAt:             time.Now(),
		ProductivityInsights:    s.generateProductivityInsights(pattern.ProductivityMetrics),
		SecurityInsights:        s.generateSecurityInsights(pattern.SecurityProfile),
		UsageOptimizations:      s.generateUsageOptimizations(pattern.ActivityPatterns),
		LearningRecommendations: s.generateLearningRecommendations(pattern),
		EfficiencyTips:          s.generateEfficiencyTips(pattern),
	}

	return insights, nil
}

// CompareToPeerGroup compares user to peer group
func (s *userAnalyticsService) CompareToPeerGroup(ctx context.Context, userID uuid.UUID) (*PeerComparison, error) {
	// TODO: Implement peer group analysis
	// This would compare the user's metrics to similar users in their organization

	comparison := &PeerComparison{
		UserID:                 userID,
		PeerGroupSize:          50, // Placeholder
		PerformancePercentile:  0.75,
		SecurityPercentile:     0.80,
		ProductivityPercentile: 0.70,
		ComparisonMetrics:      make(map[string]float64),
		StrengthAreas:          []string{"Security Awareness", "Compliance"},
		ImprovementAreas:       []string{"Productivity", "Feature Adoption"},
	}

	return comparison, nil
}

// GetBenchmarkMetrics gets benchmark metrics for an entity
func (s *userAnalyticsService) GetBenchmarkMetrics(ctx context.Context, entityID uuid.UUID) (*BenchmarkMetrics, error) {
	// TODO: Implement entity-wide benchmark calculation
	return &BenchmarkMetrics{
		EntityID:    entityID,
		GeneratedAt: time.Now(),
		UserCount:   100, // Placeholder
	}, nil
}

// Helper methods for analysis

// getUserActivities gets user activities for analysis period (placeholder)
func (s *userAnalyticsService) getUserActivities(ctx context.Context, userID uuid.UUID, window time.Duration) []*UserActivity {
	// TODO: Implement database query for user activities
	// This would retrieve activities from audit logs, session data, etc.
	return []*UserActivity{}
}

// analyzeLoginPatterns analyzes login behavior patterns
func (s *userAnalyticsService) analyzeLoginPatterns(activities []*UserActivity) LoginPatterns {
	// TODO: Implement comprehensive login pattern analysis
	return LoginPatterns{
		TypicalLoginHours:      []int{9, 10, 13, 14}, // Sample data
		AverageSessionDuration: 4 * time.Hour,
		LoginTimeVariance:      2.5,
		WeekendLoginFrequency:  0.15,
	}
}

// analyzeActivityPatterns analyzes activity behavior patterns
func (s *userAnalyticsService) analyzeActivityPatterns(activities []*UserActivity) ActivityPatterns {
	// TODO: Implement comprehensive activity pattern analysis
	return ActivityPatterns{
		PeakActivityHours: []int{10, 11, 14, 15},
		ActivityDistribution: map[string]float64{
			"document_access": 0.4,
			"data_entry":      0.3,
			"administration":  0.2,
			"reporting":       0.1,
		},
	}
}

// analyzeAccessPatterns analyzes access behavior patterns
func (s *userAnalyticsService) analyzeAccessPatterns(activities []*UserActivity) AccessPatterns {
	// TODO: Implement comprehensive access pattern analysis
	return AccessPatterns{
		PermissionUsagePatterns: []PermissionUsagePattern{},
		ElevatedAccessPatterns: ElevatedAccessPatterns{
			ElevationFrequency:       0.1,
			AverageElevationDuration: 30 * time.Minute,
			ElevationSuccessRate:     0.95,
		},
	}
}

// analyzeProductivityMetrics analyzes productivity patterns
func (s *userAnalyticsService) analyzeProductivityMetrics(activities []*UserActivity) ProductivityMetrics {
	// TODO: Implement comprehensive productivity analysis
	return ProductivityMetrics{
		TaskCompletionRate:  0.85,
		AverageTaskDuration: 45 * time.Minute,
	}
}

// analyzeSecurityProfile analyzes security behavior
func (s *userAnalyticsService) analyzeSecurityProfile(activities []*UserActivity) SecurityProfile {
	// TODO: Implement comprehensive security profile analysis
	return SecurityProfile{
		PasswordChangeFrequency: 0.25, // Every 4 months
		MFAUsageConsistency:     0.95,
		ComplianceScore:         0.88,
	}
}

// assessRisk performs risk assessment based on patterns
func (s *userAnalyticsService) assessRisk(activities []*UserActivity) UserRiskAssessment {
	// TODO: Implement comprehensive risk assessment
	return UserRiskAssessment{
		OverallRiskScore: 25,
		RiskLevel:        "LOW",
		RiskCategories: map[string]int{
			"login_behavior":    20,
			"access_patterns":   15,
			"security_behavior": 30,
		},
		LastAssessment: time.Now(),
		NextAssessment: time.Now().Add(24 * time.Hour),
	}
}

// detectBehaviorAnomalies detects anomalies in behavior patterns
func (s *userAnalyticsService) detectBehaviorAnomalies(activities []*UserActivity) []AnomalyIndicator {
	// TODO: Implement anomaly detection algorithms
	return []AnomalyIndicator{}
}

// calculateBaselineConfidence calculates confidence in established baseline
func (s *userAnalyticsService) calculateBaselineConfidence(activities []*UserActivity) float64 {
	// Simple confidence calculation based on data volume and consistency
	dataPoints := float64(len(activities))
	if dataPoints < 50 {
		return 0.3
	}
	if dataPoints < 100 {
		return 0.6
	}
	if dataPoints < 200 {
		return 0.8
	}
	return 0.95
}

// storeBehaviorPattern stores behavior pattern (placeholder)
func (s *userAnalyticsService) storeBehaviorPattern(ctx context.Context, pattern *UserBehaviorPattern) {
	// TODO: Implement database storage
	logger.Debug("Storing user behavior pattern", logger.Fields{
		"user_id":    pattern.UserID,
		"confidence": pattern.BaselineConfidence,
	})
}

// Risk calculation methods

func (s *userAnalyticsService) calculateLoginRisk(patterns LoginPatterns) int {
	risk := 0
	if patterns.LoginTimeVariance > 5.0 {
		risk += 20
	}
	if patterns.WeekendLoginFrequency > 0.3 {
		risk += 15
	}
	return min(risk, 100)
}

func (s *userAnalyticsService) calculateAccessRisk(patterns AccessPatterns) int {
	risk := 0
	if patterns.ElevatedAccessPatterns.ElevationFrequency > 0.2 {
		risk += 25
	}
	if patterns.ElevatedAccessPatterns.ElevationSuccessRate < 0.8 {
		risk += 20
	}
	return min(risk, 100)
}

func (s *userAnalyticsService) calculateSecurityRisk(profile SecurityProfile) int {
	risk := 0
	if profile.ComplianceScore < 0.7 {
		risk += 30
	}
	if profile.MFAUsageConsistency < 0.8 {
		risk += 25
	}
	return min(risk, 100)
}

func (s *userAnalyticsService) calculateAnomalyRisk(anomalies []AnomalyIndicator) int {
	risk := len(anomalies) * 10
	for _, anomaly := range anomalies {
		switch anomaly.Severity {
		case "HIGH":
			risk += 15
		case "CRITICAL":
			risk += 25
		}
	}
	return min(risk, 100)
}

func (s *userAnalyticsService) calculateDeviceRisk(consistency DeviceConsistency) int {
	risk := 0
	if consistency.NewDeviceFrequency > 0.1 {
		risk += 20
	}
	if consistency.TrustedDeviceRatio < 0.5 {
		risk += 15
	}
	return min(risk, 100)
}

func (s *userAnalyticsService) calculateLocationRisk(consistency LocationConsistency) int {
	risk := 0
	if consistency.UnusualLocationCount > 5 {
		risk += 25
	}
	if consistency.TrustedLocationRatio < 0.7 {
		risk += 15
	}
	return min(risk, 100)
}

func (s *userAnalyticsService) calculateOverallRiskScore(categories map[string]int) int {
	// Weighted average of risk categories
	weights := map[string]float64{
		"login_behavior":    0.2,
		"access_patterns":   0.25,
		"security_behavior": 0.3,
		"anomaly_risk":      0.15,
		"device_risk":       0.05,
		"location_risk":     0.05,
	}

	weightedSum := 0.0
	totalWeight := 0.0

	for category, score := range categories {
		if weight, exists := weights[category]; exists {
			weightedSum += float64(score) * weight
			totalWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	return int(weightedSum / totalWeight)
}

func (s *userAnalyticsService) getRiskLevel(score int) string {
	switch {
	case score >= 80:
		return "CRITICAL"
	case score >= 60:
		return "HIGH"
	case score >= 30:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// Anomaly detection methods

func (s *userAnalyticsService) detectTimeAnomaly(activity *UserActivity, patterns LoginPatterns) *AnomalyIndicator {
	hour := activity.Timestamp.Hour()
	isTypicalHour := false

	for _, typicalHour := range patterns.TypicalLoginHours {
		if hour == typicalHour {
			isTypicalHour = true
			break
		}
	}

	if !isTypicalHour {
		return &AnomalyIndicator{
			IndicatorType:      "UNUSUAL_TIME_ACCESS",
			Severity:           "MEDIUM",
			Confidence:         0.8,
			DetectedAt:         time.Now(),
			Description:        fmt.Sprintf("Access at unusual hour: %d", hour),
			AffectedBehaviors:  []string{"login_timing"},
			RecommendedActions: []string{"verify_user_identity", "audit_session"},
		}
	}

	return nil
}

func (s *userAnalyticsService) detectLocationAnomaly(activity *UserActivity, consistency LocationConsistency) *AnomalyIndicator {
	// TODO: Implement sophisticated location anomaly detection
	// This would check against known locations and calculate geographic distance
	return nil
}

func (s *userAnalyticsService) detectDeviceAnomaly(activity *UserActivity, consistency DeviceConsistency) *AnomalyIndicator {
	// TODO: Implement device fingerprint comparison
	// This would check device characteristics against known devices
	return nil
}

func (s *userAnalyticsService) detectActivityVolumeAnomaly(activity *UserActivity, patterns ActivityPatterns) *AnomalyIndicator {
	// TODO: Implement activity volume anomaly detection
	// This would compare current activity volume to historical patterns
	return nil
}

func (s *userAnalyticsService) detectResourceAccessAnomaly(activity *UserActivity, patterns AccessPatterns) *AnomalyIndicator {
	// TODO: Implement resource access anomaly detection
	// This would check for unusual resource access patterns
	return nil
}

// Additional helper methods

func (s *userAnalyticsService) identifyBehavioralAnomalies(pattern *UserBehaviorPattern) []BehavioralAnomaly {
	// TODO: Implement behavioral anomaly identification
	return []BehavioralAnomaly{}
}

func (s *userAnalyticsService) identifyPredictiveRiskFactors(pattern *UserBehaviorPattern, riskCategories map[string]int) []PredictiveRiskFactor {
	// TODO: Implement predictive risk factor identification
	return []PredictiveRiskFactor{}
}

func (s *userAnalyticsService) generateRiskTrends(userID uuid.UUID, days int) []RiskDataPoint {
	// TODO: Implement risk trend generation from historical data
	return []RiskDataPoint{}
}

func (s *userAnalyticsService) generateRiskBasedRecommendations(riskCategories map[string]int, anomalies []BehavioralAnomaly) []SecurityRecommendation {
	recommendations := []SecurityRecommendation{}

	// Generate recommendations based on risk categories
	for category, score := range riskCategories {
		if score >= 50 {
			switch category {
			case "security_behavior":
				recommendations = append(recommendations, SecurityRecommendation{
					RecommendationType: "SECURITY_TRAINING",
					Priority:           "HIGH",
					Title:              "Enhanced Security Training",
					Description:        "User shows elevated security risk behaviors",
					ActionRequired:     "Enroll in advanced security awareness training",
					ExpectedImpact:     "Reduce security risk by 20-30%",
				})
			case "access_patterns":
				recommendations = append(recommendations, SecurityRecommendation{
					RecommendationType: "ACCESS_REVIEW",
					Priority:           "MEDIUM",
					Title:              "Access Rights Review",
					Description:        "Unusual access patterns detected",
					ActionRequired:     "Review and validate current access permissions",
					ExpectedImpact:     "Ensure principle of least privilege",
				})
			}
		}
	}

	return recommendations
}

func (s *userAnalyticsService) extractActionRecommendations(recommendations []SecurityRecommendation) []string {
	actions := make([]string, len(recommendations))
	for i, rec := range recommendations {
		actions[i] = rec.ActionRequired
	}
	return actions
}

func (s *userAnalyticsService) generateEarlyWarningIndicators(assessment *UserRiskAssessment) []string {
	indicators := []string{}

	if assessment.OverallRiskScore > 60 {
		indicators = append(indicators, "elevated_risk_score")
	}

	if len(assessment.BehavioralAnomalies) > 3 {
		indicators = append(indicators, "multiple_anomalies")
	}

	return indicators
}

func (s *userAnalyticsService) generateProductivityInsights(metrics ProductivityMetrics) []string {
	insights := []string{}

	if metrics.TaskCompletionRate < 0.8 {
		insights = append(insights, "Consider time management training to improve task completion")
	}

	if metrics.FeatureAdoption.AdoptionRate < 0.5 {
		insights = append(insights, "Explore new features to boost productivity")
	}

	return insights
}

func (s *userAnalyticsService) generateSecurityInsights(profile SecurityProfile) []string {
	insights := []string{}

	if profile.ComplianceScore < 0.8 {
		insights = append(insights, "Review security policies to improve compliance")
	}

	if profile.MFAUsageConsistency < 0.9 {
		insights = append(insights, "Enable MFA for all accounts to enhance security")
	}

	return insights
}

func (s *userAnalyticsService) generateUsageOptimizations(patterns ActivityPatterns) []string {
	optimizations := []string{}

	if len(patterns.PeakActivityHours) < 3 {
		optimizations = append(optimizations, "Consider spreading work across more hours for better work-life balance")
	}

	return optimizations
}

func (s *userAnalyticsService) generateLearningRecommendations(pattern *UserBehaviorPattern) []string {
	recommendations := []string{}

	if pattern.ProductivityMetrics.FeatureAdoption.AdoptionRate < 0.6 {
		recommendations = append(recommendations, "Take a tutorial on advanced features")
	}

	return recommendations
}

func (s *userAnalyticsService) generateEfficiencyTips(pattern *UserBehaviorPattern) []string {
	tips := []string{}

	if len(pattern.ActivityPatterns.NavigationPatterns.CommonPathways) > 0 {
		tips = append(tips, "Use keyboard shortcuts for frequently accessed features")
	}

	return tips
}

// Utility function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
