package helpers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/shared"
)

// FeatureFlagHelper provides template helper functions for feature flag checks
type FeatureFlagHelper struct {
	featureFlagService featureflag.Service
}

// NewFeatureFlagHelper creates a new feature flag helper instance
func NewFeatureFlagHelper(featureFlagService featureflag.Service) *FeatureFlagHelper {
	return &FeatureFlagHelper{
		featureFlagService: featureFlagService,
	}
}

// FeatureFlagContext holds the context needed for feature flag evaluation
type FeatureFlagContext struct {
	UserID     uuid.UUID
	TenantID   *uuid.UUID
	EntityID   *uuid.UUID
	FlagName   string
	Attributes map[string]any
}

// IsFeatureEnabled checks if a feature flag is enabled for the current context
func (h *FeatureFlagHelper) IsFeatureEnabled(ctx context.Context, flagName string) bool {
	if h.featureFlagService == nil {
		return false
	}

	userID, ok := shared.GetUserID(ctx)
	if !ok {
		return false
	}

	tenantID, _ := shared.GetTenantID(ctx)

	// Get request context for additional attributes
	reqCtx, _ := shared.GetRequestContext(ctx)
	attributes := make(map[string]string)
	
	if reqCtx != nil {
		attributes["user_agent"] = reqCtx.UserAgent
		attributes["ip_address"] = reqCtx.IPAddress
		attributes["session_id"] = reqCtx.SessionID
	}

	// Create evaluation context
	evalCtx := &featureflag.EvaluationContext{
		TenantID:   tenantID,
		UserID:     &userID,
		Environment: "production", // TODO: Make this configurable
		Attributes: attributes,
	}

	if reqCtx != nil {
		evalCtx.ClientInfo = &featureflag.ClientInfo{
			IPAddress: reqCtx.IPAddress,
			UserAgent: reqCtx.UserAgent,
		}
	}

	// Evaluate feature flag
	enabled, err := h.featureFlagService.IsEnabled(ctx, flagName, evalCtx)
	if err != nil {
		return false
	}

	return enabled
}

// IsFeatureEnabledForUser checks if a feature flag is enabled for a specific user
func (h *FeatureFlagHelper) IsFeatureEnabledForUser(ctx context.Context, flagName string, userID uuid.UUID, tenantID *uuid.UUID) bool {
	if h.featureFlagService == nil {
		return false
	}

	var tid uuid.UUID
	if tenantID != nil {
		tid = *tenantID
	}

	evalCtx := &featureflag.EvaluationContext{
		TenantID:   tid,
		UserID:     &userID,
		Environment: "production",
		Attributes: make(map[string]string),
	}

	enabled, err := h.featureFlagService.IsEnabled(ctx, flagName, evalCtx)
	if err != nil {
		return false
	}

	return enabled
}

// GetFeatureFlagValue retrieves the value of a feature flag
func (h *FeatureFlagHelper) GetFeatureFlagValue(ctx context.Context, flagName string) (any, error) {
	if h.featureFlagService == nil {
		return nil, ErrFeatureFlagServiceNotAvailable
	}

	userID, ok := shared.GetUserID(ctx)
	if !ok {
		return nil, ErrUserNotFound
	}

	tenantID, _ := shared.GetTenantID(ctx)

	evalCtx := &featureflag.EvaluationContext{
		TenantID:   tenantID,
		UserID:     &userID,
		Environment: "production",
		Attributes: make(map[string]string),
	}

	result, err := h.featureFlagService.EvaluateFlag(ctx, flagName, evalCtx)
	if err != nil {
		return nil, err
	}

	return result.Value, nil
}

// GetFeatureFlagValueWithDefault retrieves the value of a feature flag with a default
func (h *FeatureFlagHelper) GetFeatureFlagValueWithDefault(ctx context.Context, flagName string, defaultValue any) any {
	value, err := h.GetFeatureFlagValue(ctx, flagName)
	if err != nil {
		return defaultValue
	}
	return value
}

// HasAnyFeatureEnabled checks if any of the specified feature flags are enabled
func (h *FeatureFlagHelper) HasAnyFeatureEnabled(ctx context.Context, flagNames []string) bool {
	for _, flagName := range flagNames {
		if h.IsFeatureEnabled(ctx, flagName) {
			return true
		}
	}
	return false
}

// HasAllFeaturesEnabled checks if all of the specified feature flags are enabled
func (h *FeatureFlagHelper) HasAllFeaturesEnabled(ctx context.Context, flagNames []string) bool {
	for _, flagName := range flagNames {
		if !h.IsFeatureEnabled(ctx, flagName) {
			return false
		}
	}
	return true
}

// GetUserFromContext extracts user information from request context
func (h *FeatureFlagHelper) GetUserFromContext(ctx context.Context) (uuid.UUID, *uuid.UUID, error) {
	userID, ok := shared.GetUserID(ctx)
	if !ok {
		return uuid.Nil, nil, ErrUserNotFound
	}

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		return userID, nil, nil
	}

	return userID, &tenantID, nil
}

// Utility functions for common feature flag patterns

// IsBetaFeatureEnabled checks if beta features are enabled for the user
func (h *FeatureFlagHelper) IsBetaFeatureEnabled(ctx context.Context) bool {
	return h.IsFeatureEnabled(ctx, "beta_features")
}

// IsExperimentalFeatureEnabled checks if experimental features are enabled
func (h *FeatureFlagHelper) IsExperimentalFeatureEnabled(ctx context.Context, experimentName string) bool {
	return h.IsFeatureEnabled(ctx, "experimental_"+experimentName)
}

// IsUIFeatureEnabled checks if a specific UI feature is enabled
func (h *FeatureFlagHelper) IsUIFeatureEnabled(ctx context.Context, featureName string) bool {
	return h.IsFeatureEnabled(ctx, "ui_"+featureName)
}

// IsIntegrationEnabled checks if a specific integration is enabled
func (h *FeatureFlagHelper) IsIntegrationEnabled(ctx context.Context, integrationName string) bool {
	return h.IsFeatureEnabled(ctx, "integration_"+integrationName)
}

// IsModuleEnabled checks if a specific module is enabled
func (h *FeatureFlagHelper) IsModuleEnabled(ctx context.Context, moduleName string) bool {
	return h.IsFeatureEnabled(ctx, "module_"+moduleName)
}

// Error definitions
var (
	ErrFeatureFlagServiceNotAvailable = errors.New("feature flag service not available")
)

// Common feature flag constants
const (
	// UI Features
	FeatureFlagNewDashboard     = "ui_new_dashboard"
	FeatureFlagAdvancedReports  = "ui_advanced_reports"
	FeatureFlagBulkOperations   = "ui_bulk_operations"
	FeatureFlagDarkMode         = "ui_dark_mode"
	FeatureFlagMobileLayout     = "ui_mobile_layout"

	// Module Features
	FeatureFlagFinanceModule    = "module_finance"
	FeatureFlagInventoryModule  = "module_inventory"
	FeatureFlagHRModule         = "module_hr"
	FeatureFlagCRMModule        = "module_crm"
	FeatureFlagProjectModule    = "module_project"

	// Integration Features
	FeatureFlagSlackIntegration = "integration_slack"
	FeatureFlagEmailIntegration = "integration_email"
	FeatureFlagSSOIntegration   = "integration_sso"
	FeatureFlagAPIv2            = "integration_api_v2"

	// Experimental Features
	FeatureFlagAIAssistant      = "experimental_ai_assistant"
	FeatureFlagAdvancedAnalytics = "experimental_advanced_analytics"
	FeatureFlagVoiceCommands    = "experimental_voice_commands"

	// Beta Features
	FeatureFlagBetaFeatures     = "beta_features"
	FeatureFlagEarlyAccess      = "early_access"
	FeatureFlagDeveloperMode    = "developer_mode"
)