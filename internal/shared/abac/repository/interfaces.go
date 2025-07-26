package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/niiniyare/erp/internal/core/abac/models"
)

// PolicyRepository defines the interface for policy data access
type PolicyRepository interface {
	// Policy CRUD operations
	CreatePolicy(ctx context.Context, policy *Policy) error
	GetPolicyByID(ctx context.Context, id uuid.UUID) (*Policy, error)
	UpdatePolicy(ctx context.Context, id uuid.UUID, policy *Policy) error
	DeletePolicy(ctx context.Context, id uuid.UUID) error
	ListPolicies(ctx context.Context, filter *PolicyFilter) ([]*Policy, error)

	// Policy queries for ABAC evaluation
	GetApplicablePolicies(ctx context.Context, resourceName, actionName string, entityID uuid.UUID) ([]*Policy, error)
	GetPoliciesByTarget(ctx context.Context, target map[string]interface{}) ([]*Policy, error)
	GetPoliciesByPriority(ctx context.Context, minPriority int) ([]*Policy, error)

	// Policy validation and testing
	ValidatePolicyRule(ctx context.Context, rule map[string]interface{}) error
	TestPolicy(ctx context.Context, policyID uuid.UUID, testContext map[string]interface{}) (*models.PolicyTestResult, error)
}

// AttributeDefinitionRepository defines the interface for attribute definition data access
type AttributeDefinitionRepository interface {
	// Attribute definition CRUD operations
	CreateAttributeDefinition(ctx context.Context, attr *AttributeDefinition) error
	GetAttributeDefinitionByID(ctx context.Context, id uuid.UUID) (*AttributeDefinition, error)
	GetAttributeDefinitionByName(ctx context.Context, name string) (*AttributeDefinition, error)
	UpdateAttributeDefinition(ctx context.Context, id uuid.UUID, attr *AttributeDefinition) error
	DeleteAttributeDefinition(ctx context.Context, id uuid.UUID) error
	ListAttributeDefinitions(ctx context.Context, filter *AttributeFilter) ([]*AttributeDefinition, error)

	// Attribute queries
	GetAttributesByCategory(ctx context.Context, category string) ([]*AttributeDefinition, error)
	ValidateAttributeValue(ctx context.Context, name string, value interface{}) error
}

// PolicyEvaluationCacheRepository defines the interface for policy evaluation caching
type PolicyEvaluationCacheRepository interface {
	// Cache operations
	StorePolicyEvaluation(ctx context.Context, entry *models.PolicyEvaluationCacheEntry) error
	GetPolicyEvaluation(ctx context.Context, cacheKey string) (*models.PolicyEvaluationCacheEntry, error)
	InvalidatePolicyEvaluations(ctx context.Context, userID uuid.UUID, resourcePattern string) error
	DeleteExpiredEntries(ctx context.Context) (int, error)

	// Cache management
	GetCacheStats(ctx context.Context) (*CacheStats, error)
	ClearCache(ctx context.Context) error
}

// UserActivityRepository defines the interface for user activity tracking
type UserActivityRepository interface {
	// Activity logging
	LogActivity(ctx context.Context, activity *UserActivity) error
	LogBulkActivities(ctx context.Context, activities []*UserActivity) error

	// Activity queries
	GetUserActivities(ctx context.Context, userID uuid.UUID, filter *ActivityFilter) ([]*UserActivity, error)
	GetActivitiesByResource(ctx context.Context, resourceName string, filter *ActivityFilter) ([]*UserActivity, error)
	GetHighRiskActivities(ctx context.Context, threshold float64, filter *ActivityFilter) ([]*UserActivity, error)

	// Analytics
	GetUserBehaviorAnalytics(ctx context.Context, userID uuid.UUID, period time.Duration) (*UserBehaviorAnalytics, error)
	GetResourceAccessPattern(ctx context.Context, resourceName string, period time.Duration) (*ResourceAccessPattern, error)
	CalculateAnomalyScore(ctx context.Context, userID uuid.UUID, activity *UserActivity) (float64, error)
}

// Domain models for repository layer

// Policy represents a policy in the database
type Policy struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	EntityID    *uuid.UUID             `json:"entity_id,omitempty"`
	Name        string                 `json:"name"`
	DisplayName *string                `json:"display_name,omitempty"`
	Description *string                `json:"description,omitempty"`
	PolicyType  string                 `json:"policy_type"`
	Effect      string                 `json:"effect"`
	Priority    int                    `json:"priority"`
	Category    string                 `json:"category"`
	Target      map[string]interface{} `json:"target"`
	Rule        map[string]interface{} `json:"rule"`
	Obligations map[string]interface{} `json:"obligations"`
	Advice      map[string]interface{} `json:"advice"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   *uuid.UUID             `json:"created_by,omitempty"`
	DeletedAt   *time.Time             `json:"deleted_at,omitempty"`
}

// AttributeDefinition represents an attribute definition in the database
type AttributeDefinition struct {
	ID                 uuid.UUID              `json:"id"`
	TenantID           uuid.UUID              `json:"tenant_id"`
	Name               string                 `json:"name"`
	DisplayName        *string                `json:"display_name,omitempty"`
	Description        *string                `json:"description,omitempty"`
	DataType           string                 `json:"data_type"`
	Category           string                 `json:"category"`
	IsRequired         bool                   `json:"is_required"`
	IsSensitive        bool                   `json:"is_sensitive"`
	DefaultValue       *string                `json:"default_value,omitempty"`
	AllowedValues      map[string]interface{} `json:"allowed_values,omitempty"`
	ValidationRules    map[string]interface{} `json:"validation_rules"`
	EncryptionRequired bool                   `json:"encryption_required"`
	IsActive           bool                   `json:"is_active"`
	CreatedAt          time.Time              `json:"created_at"`
}

// UserActivity represents a user activity in the database
type UserActivity struct {
	ID                uuid.UUID              `json:"id"`
	UserID            uuid.UUID              `json:"user_id"`
	TenantID          uuid.UUID              `json:"tenant_id"`
	SessionID         *uuid.UUID             `json:"session_id,omitempty"`
	ActivityType      string                 `json:"activity_type"`
	Module            *string                `json:"module,omitempty"`
	ResourceType      *string                `json:"resource_type,omitempty"`
	ResourceID        *uuid.UUID             `json:"resource_id,omitempty"`
	ActionPerformed   *string                `json:"action_performed,omitempty"`
	IPAddress         *string                `json:"ip_address,omitempty"`
	UserAgent         *string                `json:"user_agent,omitempty"`
	DeviceFingerprint *string                `json:"device_fingerprint,omitempty"`
	LocationData      map[string]interface{} `json:"location_data"`
	RequestMethod     *string                `json:"request_method,omitempty"`
	RequestPath       *string                `json:"request_path,omitempty"`
	RequestParams     map[string]interface{} `json:"request_params"`
	ResponseStatus    *int                   `json:"response_status,omitempty"`
	ResponseTimeMS    *int                   `json:"response_time_ms,omitempty"`
	RiskIndicators    map[string]interface{} `json:"risk_indicators"`
	AnomalyScore      float64                `json:"anomaly_score"`
	Timestamp         time.Time              `json:"timestamp"`
	AdditionalData    map[string]interface{} `json:"additional_data"`
}

// Filter types for repository queries

// PolicyFilter represents filters for policy queries
type PolicyFilter struct {
	TenantID     *uuid.UUID `json:"tenant_id,omitempty"`
	EntityID     *uuid.UUID `json:"entity_id,omitempty"`
	PolicyType   *string    `json:"policy_type,omitempty"`
	Effect       *string    `json:"effect,omitempty"`
	Category     *string    `json:"category,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
	MinPriority  *int       `json:"min_priority,omitempty"`
	MaxPriority  *int       `json:"max_priority,omitempty"`
	CreatedAfter *time.Time `json:"created_after,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// AttributeFilter represents filters for attribute definition queries
type AttributeFilter struct {
	TenantID    *uuid.UUID `json:"tenant_id,omitempty"`
	Category    *string    `json:"category,omitempty"`
	DataType    *string    `json:"data_type,omitempty"`
	IsRequired  *bool      `json:"is_required,omitempty"`
	IsSensitive *bool      `json:"is_sensitive,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

// ActivityFilter represents filters for user activity queries
type ActivityFilter struct {
	TenantID        *uuid.UUID `json:"tenant_id,omitempty"`
	UserID          *uuid.UUID `json:"user_id,omitempty"`
	ActivityType    *string    `json:"activity_type,omitempty"`
	ResourceType    *string    `json:"resource_type,omitempty"`
	MinAnomalyScore *float64   `json:"min_anomaly_score,omitempty"`
	StartTime       *time.Time `json:"start_time,omitempty"`
	EndTime         *time.Time `json:"end_time,omitempty"`
	Limit           int        `json:"limit,omitempty"`
	Offset          int        `json:"offset,omitempty"`
}

// Analytics types

// CacheStats represents cache performance statistics
type CacheStats struct {
	TotalEntries   int64         `json:"total_entries"`
	ExpiredEntries int64         `json:"expired_entries"`
	HitRate        float64       `json:"hit_rate"`
	MissRate       float64       `json:"miss_rate"`
	AverageAge     time.Duration `json:"average_age"`
	LastCleanup    time.Time     `json:"last_cleanup"`
}

// UserBehaviorAnalytics represents user behavior analysis results
type UserBehaviorAnalytics struct {
	UserID              uuid.UUID              `json:"user_id"`
	AnalysisPeriod      time.Duration          `json:"analysis_period"`
	TotalActivities     int64                  `json:"total_activities"`
	UniqueResources     int64                  `json:"unique_resources"`
	AverageAnomalyScore float64                `json:"average_anomaly_score"`
	PeakUsageHours      []int                  `json:"peak_usage_hours"`
	CommonLocations     []string               `json:"common_locations"`
	FrequentResources   []string               `json:"frequent_resources"`
	RiskFactors         []string               `json:"risk_factors"`
	BehaviorPattern     map[string]interface{} `json:"behavior_pattern"`
}

// ResourceAccessPattern represents resource access pattern analysis
type ResourceAccessPattern struct {
	ResourceName        string    `json:"resource_name"`
	TotalAccesses       int64     `json:"total_accesses"`
	UniqueUsers         int64     `json:"unique_users"`
	AverageResponseTime float64   `json:"average_response_time"`
	PeakAccessHours     []int     `json:"peak_access_hours"`
	CommonActions       []string  `json:"common_actions"`
	AccessTrend         string    `json:"access_trend"` // INCREASING, DECREASING, STABLE
	LastAccessed        time.Time `json:"last_accessed"`
}
