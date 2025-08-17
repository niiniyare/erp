package featureflag

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
)

// Repository defines the data access interface for feature flags
type Repository interface {
	// Feature Flag CRUD Operations
	CreateFeatureFlag(ctx context.Context, flag *CreateFeatureFlagRequest) (*FeatureFlag, error)
	GetFeatureFlagByName(ctx context.Context, name string) (*FeatureFlag, error)
	GetFeatureFlagByID(ctx context.Context, id uuid.UUID) (*FeatureFlag, error)
	UpdateFeatureFlag(ctx context.Context, id uuid.UUID, flag *UpdateFeatureFlagRequest) (*FeatureFlag, error)
	DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error
	ListFeatureFlags(ctx context.Context, params ListFeatureFlagsParams) ([]*FeatureFlag, error)

	// Evaluation Operations
	GetActiveFlags(ctx context.Context) ([]*FeatureFlag, error)
	EvaluateFlags(ctx context.Context, userID *string, flagNames []string) ([]*SimpleEvaluationResult, error)

	// Statistics
	GetFlagStats(ctx context.Context) (*FlagStats, error)
	SearchFlags(ctx context.Context, query string, limit, offset int32) ([]*FeatureFlag, error)
	GetFlagsByType(ctx context.Context, flagType string) ([]*FeatureFlag, error)
}

// RepositoryImpl implements the Repository interface
type repository struct {
	store db.Store
}

// NewRepository creates a new simple feature flag repository
func NewRepository(store db.Store) Repository {
	return &repository{
		store: store,
	}
}

// ListFeatureFlagsParams defines parameters for listing feature flags
type ListFeatureFlagsParams struct {
	FlagType *string
	Limit    int32
	Offset   int32
}

// FlagStats represents feature flag statistics
type FlagStats struct {
	TotalFlags           int64    `json:"total_flags"`
	EnabledFlags         int64    `json:"enabled_flags"`
	RolloutFlags         int64    `json:"rollout_flags"`
	AvgRolloutPercentage *float64 `json:"avg_rollout_percentage,omitempty"`
}

// SimpleEvaluationResult represents the result of flag evaluation
type SimpleEvaluationResult struct {
	FlagKey     string    `json:"flag_key"`
	Value       bool      `json:"value"`
	Enabled     bool      `json:"enabled"`
	Source      string    `json:"source"`
	Reason      string    `json:"reason"`
	Variation   string    `json:"variation"`
	EvaluatedAt time.Time `json:"evaluated_at"`
}

// CreateFeatureFlag creates a new feature flag
func (r *repository) CreateFeatureFlag(ctx context.Context, req *CreateFeatureFlagRequest) (*FeatureFlag, error) {
	var result *FeatureFlag

	err := r.store.WithTx(ctx, func(ctx context.Context, store db.Store) error {
		// Convert to SQLC params
		params := db.CreateFeatureFlagParams{
			Name:              req.Name,
			Description:       req.Description,
			FlagType:          string(req.FlagType),
			DefaultValue:      req.DefaultValue,
			RolloutPercentage: req.RolloutPercentage,
			TargetAudience:    mustMarshalJSON(req.TargetAudience),
			Metadata:          mustMarshalJSON(req.Metadata),
		}

		dbFlag, err := store.CreateFeatureFlag(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to create feature flag: %w", err)
		}

		result = fromSQLCFeatureFlag(*dbFlag)
		return nil
	})

	return result, err
}

// GetFeatureFlagByName retrieves a feature flag by name
func (r *repository) GetFeatureFlagByName(ctx context.Context, name string) (*FeatureFlag, error) {
	dbFlag, err := r.store.GetFeatureFlagByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrFeatureFlagNotFound
		}
		return nil, fmt.Errorf("failed to get feature flag by name: %w", err)
	}

	return fromSQLCFeatureFlag(*dbFlag), nil
}

// GetFeatureFlagByID retrieves a feature flag by ID
func (r *repository) GetFeatureFlagByID(ctx context.Context, id uuid.UUID) (*FeatureFlag, error) {
	dbFlag, err := r.store.GetFeatureFlagByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrFeatureFlagNotFound
		}
		return nil, fmt.Errorf("failed to get feature flag by ID: %w", err)
	}

	return fromSQLCFeatureFlag(*dbFlag), nil
}

// UpdateFeatureFlag updates an existing feature flag
func (r *repository) UpdateFeatureFlag(ctx context.Context, id uuid.UUID, req *UpdateFeatureFlagRequest) (*FeatureFlag, error) {
	var result *FeatureFlag

	err := r.store.WithTx(ctx, func(ctx context.Context, store db.Store) error {
		// Get existing flag first
		existing, err := store.GetFeatureFlagByID(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get existing flag: %w", err)
		}

		// Apply updates
		params := db.UpdateFeatureFlagParams{
			ID:                id,
			Name:              existing.Name,
			Description:       existing.Description,
			FlagType:          existing.FlagType,
			DefaultValue:      existing.DefaultValue,
			RolloutPercentage: existing.RolloutPercentage,
			TargetAudience:    existing.TargetAudience,
			Metadata:          existing.Metadata,
		}

		if req.Description != nil {
			params.Description = *req.Description
		}
		if req.DefaultValue != nil {
			params.DefaultValue = *req.DefaultValue
		}
		if req.RolloutPercentage != nil {
			params.RolloutPercentage = req.RolloutPercentage
		}
		if req.TargetAudience != nil {
			params.TargetAudience = mustMarshalJSON(req.TargetAudience)
		}
		if req.Metadata != nil {
			params.Metadata = mustMarshalJSON(req.Metadata)
		}

		dbFlag, err := store.UpdateFeatureFlag(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to update feature flag: %w", err)
		}

		result = fromSQLCFeatureFlag(*dbFlag)
		return nil
	})

	return result, err
}

// DeleteFeatureFlag deletes a feature flag (soft delete)
func (r *repository) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error {
	return r.store.WithTx(ctx, func(ctx context.Context, store db.Store) error {
		err := store.DeleteFeatureFlag(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to delete feature flag: %w", err)
		}
		return nil
	})
}

// ListFeatureFlags lists feature flags with optional filters
func (r *repository) ListFeatureFlags(ctx context.Context, params ListFeatureFlagsParams) ([]*FeatureFlag, error) {
	var flagType string
	if params.FlagType != nil {
		flagType = *params.FlagType
	}

	dbFlags, err := r.store.ListFeatureFlags(ctx, db.ListFeatureFlagsParams{
		Column1: flagType,
		Limit:   params.Limit,
		Offset:  params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list feature flags: %w", err)
	}

	flags := make([]*FeatureFlag, len(dbFlags))
	for i, dbFlag := range dbFlags {
		flags[i] = fromSQLCFeatureFlag(*dbFlag)
	}

	return flags, nil
}

// GetActiveFlags retrieves all active feature flags
func (r *repository) GetActiveFlags(ctx context.Context) ([]*FeatureFlag, error) {
	dbFlags, err := r.store.GetActiveFeatureFlags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active flags: %w", err)
	}

	flags := make([]*FeatureFlag, len(dbFlags))
	for i, dbFlag := range dbFlags {
		flags[i] = fromSQLCFeatureFlag(*dbFlag)
	}

	return flags, nil
}

// EvaluateFlags performs bulk flag evaluation
func (r *repository) EvaluateFlags(ctx context.Context, userID *string, flagNames []string) ([]*SimpleEvaluationResult, error) {
	var userParam string
	if userID != nil {
		userParam = *userID
	}

	dbResults, err := r.store.BulkEvaluateFeatureFlags(ctx, db.BulkEvaluateFeatureFlagsParams{
		Column1: userParam,
		Column2: flagNames,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to bulk evaluate flags: %w", err)
	}

	results := make([]*SimpleEvaluationResult, len(dbResults))
	for i, dbResult := range dbResults {
		results[i] = &SimpleEvaluationResult{
			FlagKey:     dbResult.FlagKey,
			Value:       dbResult.IsEnabled, // Use IsEnabled for boolean value
			Enabled:     dbResult.IsEnabled,
			Source:      dbResult.Source,
			Reason:      dbResult.Reason,
			Variation:   dbResult.Variation,
			EvaluatedAt: time.Now(), // Use current time since dbResult.EvaluatedAt is interface{}
		}
	}

	return results, nil
}

// GetFlagStats retrieves feature flag statistics
func (r *repository) GetFlagStats(ctx context.Context) (*FlagStats, error) {
	dbStats, err := r.store.GetFeatureFlagStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get flag stats: %w", err)
	}

	stats := &FlagStats{
		TotalFlags:           dbStats.TotalFlags,
		EnabledFlags:         dbStats.EnabledFlags,
		RolloutFlags:         dbStats.RolloutFlags,
		AvgRolloutPercentage: &dbStats.AvgRolloutPercentage,
	}

	return stats, nil
}

// SearchFlags searches feature flags by name/description
func (r *repository) SearchFlags(ctx context.Context, query string, limit, offset int32) ([]*FeatureFlag, error) {
	dbFlags, err := r.store.SearchFeatureFlags(ctx, db.SearchFeatureFlagsParams{
		Column1: query,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search flags: %w", err)
	}

	flags := make([]*FeatureFlag, len(dbFlags))
	for i, dbFlag := range dbFlags {
		flags[i] = fromSQLCFeatureFlag(*dbFlag)
	}

	return flags, nil
}

// GetFlagsByType retrieves flags by type
func (r *repository) GetFlagsByType(ctx context.Context, flagType string) ([]*FeatureFlag, error) {
	dbFlags, err := r.store.GetFeatureFlagsByType(ctx, flagType)
	if err != nil {
		return nil, fmt.Errorf("failed to get flags by type: %w", err)
	}

	flags := make([]*FeatureFlag, len(dbFlags))
	for i, dbFlag := range dbFlags {
		flags[i] = fromSQLCFeatureFlag(*dbFlag)
	}

	return flags, nil
}

// Helper functions

// mustMarshalJSON marshals data to JSON, returning empty object on error
func mustMarshalJSON(v interface{}) []byte {
	if v == nil {
		return []byte("{}")
	}
	data, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return data
}

// fromSQLCFeatureFlag converts SQLC FeatureFlag to domain FeatureFlag
func fromSQLCFeatureFlag(dbFlag db.FeatureFlag) *FeatureFlag {
	flag := &FeatureFlag{
		ID:                dbFlag.ID,
		TenantID:          dbFlag.TenantID,
		Name:              dbFlag.Name,
		Description:       dbFlag.Description,
		FlagType:          FlagType(dbFlag.FlagType),
		DefaultValue:      dbFlag.DefaultValue,
		RolloutPercentage: dbFlag.RolloutPercentage,
		CreatedAt:         dbFlag.CreatedAt,
		UpdatedAt:         dbFlag.UpdatedAt,
	}

	// Handle nullable deleted_at
	if dbFlag.DeletedAt.Valid {
		flag.DeletedAt = &dbFlag.DeletedAt.Time
	}

	// Unmarshal JSON fields
	if len(dbFlag.TargetAudience) > 0 {
		json.Unmarshal(dbFlag.TargetAudience, &flag.TargetAudience)
	}
	if len(dbFlag.Metadata) > 0 {
		json.Unmarshal(dbFlag.Metadata, &flag.Metadata)
	}

	return flag
}
