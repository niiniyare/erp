package featureflag

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	db "awo/db/sqlc"
)

// FeatureFlag represents a feature flag in the system (domain model)
type FeatureFlag struct {
	ID                uuid.UUID      `json:"id"`
	TenantID          uuid.UUID      `json:"tenant_id"`
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	FlagType          FlagType       `json:"flag_type"`
	DefaultValue      bool           `json:"default_value"`
	RolloutPercentage *int32         `json:"rollout_percentage,omitempty"`
	TargetAudience    map[string]any `json:"target_audience"`
	Enabled           bool           `json:"enabled"`
	Metadata          map[string]any `json:"metadata"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         *time.Time     `json:"deleted_at,omitempty"`
}

// FlagType represents the type of feature flag
type FlagType string

const (
	FlagTypeBoolean FlagType = "boolean"
	FlagTypeString  FlagType = "string"
	FlagTypeNumber  FlagType = "number"
	FlagTypeJSON    FlagType = "json"
)

// FromSQLCFeatureFlag converts SQLC FeatureFlag to domain FeatureFlag
func FromSQLCFeatureFlag(sqlcFlag *db.FeatureFlag) (*FeatureFlag, error) {
	var targetAudience map[string]any
	if len(sqlcFlag.TargetAudience) > 0 {
		if err := json.Unmarshal(sqlcFlag.TargetAudience, &targetAudience); err != nil {
			return nil, err
		}
	}

	var metadata map[string]any
	if len(sqlcFlag.Metadata) > 0 {
		if err := json.Unmarshal(sqlcFlag.Metadata, &metadata); err != nil {
			return nil, err
		}
	}

	var deletedAt *time.Time
	if sqlcFlag.DeletedAt.Valid {
		deletedAt = &sqlcFlag.DeletedAt.Time
	}

	var description string
	if sqlcFlag.Description != nil {
		description = *sqlcFlag.Description
	}

	return &FeatureFlag{
		ID:                sqlcFlag.ID,
		TenantID:          sqlcFlag.TenantID,
		Name:              sqlcFlag.Name,
		Description:       description,
		FlagType:          FlagType(sqlcFlag.FlagType),
		DefaultValue:      sqlcFlag.DefaultValue,
		RolloutPercentage: sqlcFlag.RolloutPercentage,
		TargetAudience:    targetAudience,
		Metadata:          metadata,
		CreatedAt:         sqlcFlag.CreatedAt,
		UpdatedAt:         sqlcFlag.UpdatedAt,
		DeletedAt:         deletedAt,
	}, nil
}

// ToSQLCCreateParams converts domain CreateFeatureFlagRequest to SQLC params
func (req *CreateFeatureFlagRequest) ToSQLCCreateParams() (db.CreateFeatureFlagParams, error) {
	var targetAudienceJSON []byte
	var err error
	if req.TargetAudience != nil {
		targetAudienceJSON, err = json.Marshal(req.TargetAudience)
		if err != nil {
			return db.CreateFeatureFlagParams{}, err
		}
	}

	var metadataJSON []byte
	if req.Metadata != nil {
		metadataJSON, err = json.Marshal(req.Metadata)
		if err != nil {
			return db.CreateFeatureFlagParams{}, err
		}
	}

	return db.CreateFeatureFlagParams{
		Name:              req.Name,
		Description:       &req.Description,
		FlagType:          string(req.FlagType),
		DefaultValue:      req.DefaultValue,
		RolloutPercentage: req.RolloutPercentage,
		TargetAudience:    targetAudienceJSON,
		Metadata:          metadataJSON,
	}, nil
}

// CreateFeatureFlagRequest represents the request to create a feature flag
type CreateFeatureFlagRequest struct {
	Name              string         `json:"name" validate:"required,min=3,max=100"`
	Description       string         `json:"description" validate:"required,max=500"`
	FlagType          FlagType       `json:"flag_type" validate:"required"`
	DefaultValue      bool           `json:"default_value"`
	RolloutPercentage *int32         `json:"rollout_percentage,omitempty" validate:"omitempty,min=0,max=100"`
	TargetAudience    map[string]any `json:"target_audience,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}

// UpdateFeatureFlagRequest represents the request to update a feature flag
type UpdateFeatureFlagRequest struct {
	Description       *string        `json:"description,omitempty" validate:"omitempty,max=500"`
	DefaultValue      *bool          `json:"default_value,omitempty"`
	RolloutPercentage *int32         `json:"rollout_percentage,omitempty" validate:"omitempty,min=0,max=100"`
	TargetAudience    map[string]any `json:"target_audience,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}

// TenantFeatureOverride represents a tenant-specific feature flag override
type TenantFeatureOverride struct {
	ID              uuid.UUID      `json:"id"`
	TenantID        uuid.UUID      `json:"tenant_id"`
	FeatureFlagID   uuid.UUID      `json:"feature_flag_id"`
	FeatureFlagName string         `json:"feature_flag_name"`
	Enabled         bool           `json:"enabled"`
	Value           map[string]any `json:"value"`
	Reason          *string        `json:"reason,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// FromSQLCTenantFeatureOverride converts SQLC TenantFeatureOverride to domain TenantFeatureOverride
func FromSQLCTenantFeatureOverride(sqlcOverride *db.TenantFeatureOverride) (*TenantFeatureOverride, error) {
	var value map[string]any
	if len(sqlcOverride.Value) > 0 {
		if err := json.Unmarshal(sqlcOverride.Value, &value); err != nil {
			return nil, err
		}
	}

	return &TenantFeatureOverride{
		ID:              sqlcOverride.ID,
		TenantID:        sqlcOverride.TenantID,
		FeatureFlagID:   sqlcOverride.FeatureFlagID,
		FeatureFlagName: sqlcOverride.FeatureFlagName,
		Enabled:         sqlcOverride.Enabled,
		Value:           value,
		Reason:          sqlcOverride.Reason,
		CreatedAt:       sqlcOverride.CreatedAt,
		UpdatedAt:       sqlcOverride.UpdatedAt,
	}, nil
}

// CreateTenantOverrideRequest represents the request to create a tenant override
type CreateTenantOverrideRequest struct {
	FeatureFlagID uuid.UUID      `json:"feature_flag_id" validate:"required"`
	Enabled       bool           `json:"enabled"`
	Value         map[string]any `json:"value,omitempty"`
	Reason        *string        `json:"reason,omitempty" validate:"omitempty,max=500"`
}

// UpdateTenantOverrideRequest represents the request to update a tenant override
type UpdateTenantOverrideRequest struct {
	Enabled *bool          `json:"enabled,omitempty"`
	Value   map[string]any `json:"value,omitempty"`
	Reason  *string        `json:"reason,omitempty" validate:"omitempty,max=500"`
}

// EvaluationContext contains the context for flag evaluation
type EvaluationContext struct {
	TenantID    uuid.UUID         `json:"tenant_id"`
	UserID      *uuid.UUID        `json:"user_id,omitempty"`
	Environment string            `json:"environment"`
	Attributes  map[string]string `json:"attributes"`
	ClientInfo  *ClientInfo       `json:"client_info,omitempty"`
}

// ClientInfo contains client application details
type ClientInfo struct {
	Version   string `json:"version"`
	Platform  string `json:"platform"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

// EvaluationResult contains the flag evaluation outcome
type EvaluationResult struct {
	FlagName    string         `json:"flag_name"`
	Value       any            `json:"value"`
	Enabled     bool           `json:"enabled"`
	Reason      string         `json:"reason"`
	RuleMatched *string        `json:"rule_matched,omitempty"`
	Metadata    ResultMetadata `json:"metadata"`
}

// ResultMetadata provides evaluation context
type ResultMetadata struct {
	EvaluatedAt   time.Time `json:"evaluated_at"`
	CacheHit      bool      `json:"cache_hit"`
	EvaluationMs  float64   `json:"evaluation_ms"`
	ConfigVersion string    `json:"config_version"`
}

// EvaluationReason represents the reason for flag evaluation result
type EvaluationReason string

const (
	ReasonFlagDisabled    EvaluationReason = "flag_disabled"
	ReasonTenantOverride  EvaluationReason = "tenant_override"
	ReasonRuleMatch       EvaluationReason = "rule_match"
	ReasonPercentRollout  EvaluationReason = "percentage_rollout"
	ReasonDefaultValue    EvaluationReason = "default_value"
	ReasonFlagNotFound    EvaluationReason = "flag_not_found"
	ReasonEvaluationError EvaluationReason = "evaluation_error"
)

// TargetRule defines targeting criteria for feature flags
type TargetRule struct {
	Name       string      `json:"name"`
	Percentage int         `json:"percentage"`
	Conditions []Condition `json:"conditions"`
	Value      any         `json:"value"`
	Enabled    bool        `json:"enabled"`
	Priority   int         `json:"priority"`
}

// Condition represents a targeting condition
type Condition struct {
	Attribute string   `json:"attribute"`
	Operator  string   `json:"operator"`
	Values    []string `json:"values"`
}

// BulkEvaluationRequest represents a request to evaluate multiple flags
type BulkEvaluationRequest struct {
	FlagNames         []string          `json:"flag_names,omitempty"`
	EvaluationContext EvaluationContext `json:"evaluation_context"`
}

// BulkEvaluationResponse represents the response for bulk flag evaluation
type BulkEvaluationResponse struct {
	Results map[string]*EvaluationResult `json:"results"`
	Errors  map[string]string            `json:"errors,omitempty"`
}
