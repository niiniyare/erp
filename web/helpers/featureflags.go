package helpers

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/shared"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// FeatureFlagHelper provides template helper functions for feature flag checks
type FeatureFlagHelper struct {
	featureFlagService featureflag.Service
	logger             logger.Logger
	mu                 sync.RWMutex
	config             *FeatureFlagHelperConfig
}

// FeatureFlagHelperConfig holds configuration for the feature flag helper
type FeatureFlagHelperConfig struct {
	Environment      string
	EnableCaching    bool
	CacheTTL         int  // seconds
	DefaultValue     bool // Default value when flag evaluation fails
	StrictValidation bool
}

// DefaultFeatureFlagHelperConfig returns default configuration
func DefaultFeatureFlagHelperConfig() *FeatureFlagHelperConfig {
	return &FeatureFlagHelperConfig{
		Environment:      "production",
		EnableCaching:    false,
		CacheTTL:         60,
		DefaultValue:     false,
		StrictValidation: false,
	}
}

// NewFeatureFlagHelper creates a new feature flag helper instance
func NewFeatureFlagHelper(featureFlagService featureflag.Service) *FeatureFlagHelper {
	return &FeatureFlagHelper{
		featureFlagService: featureFlagService,
		logger:             nil,
		config:             DefaultFeatureFlagHelperConfig(),
	}
}

// NewFeatureFlagHelperWithLogger creates a helper with logger
func NewFeatureFlagHelperWithLogger(featureFlagService featureflag.Service, log logger.Logger) *FeatureFlagHelper {
	return &FeatureFlagHelper{
		featureFlagService: featureFlagService,
		logger:             log,
		config:             DefaultFeatureFlagHelperConfig(),
	}
}

// NewFeatureFlagHelperWithConfig creates a helper with custom config
func NewFeatureFlagHelperWithConfig(featureFlagService featureflag.Service, log logger.Logger, config *FeatureFlagHelperConfig) *FeatureFlagHelper {
	if config == nil {
		config = DefaultFeatureFlagHelperConfig()
	}
	return &FeatureFlagHelper{
		featureFlagService: featureFlagService,
		logger:             log,
		config:             config,
	}
}

// SetLogger sets the logger instance
func (h *FeatureFlagHelper) SetLogger(log logger.Logger) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.logger = log
}

// GetConfig returns a copy of the current configuration
func (h *FeatureFlagHelper) GetConfig() FeatureFlagHelperConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return *h.config
}

// UpdateConfig updates the configuration
func (h *FeatureFlagHelper) UpdateConfig(config *FeatureFlagHelperConfig) {
	if config == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config = config

	h.logDebug("Feature flag helper config updated", logger.Fields{
		"environment":       config.Environment,
		"enable_caching":    config.EnableCaching,
		"default_value":     config.DefaultValue,
		"strict_validation": config.StrictValidation,
	})
}

// SetEnvironment updates the environment setting
func (h *FeatureFlagHelper) SetEnvironment(env string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config.Environment = env

	h.logInfo("Environment updated", logger.Fields{
		"environment": env,
	})
}

// Logging helpers
func (h *FeatureFlagHelper) logDebug(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Debug(msg, fields...)
	}
}

func (h *FeatureFlagHelper) logInfo(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Info(msg, fields...)
	}
}

func (h *FeatureFlagHelper) logWarn(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Warn(msg, fields...)
	}
}

func (h *FeatureFlagHelper) logError(msg string, fields ...logger.Fields) {
	if h.logger != nil {
		h.logger.Error(msg, fields...)
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

// Validate checks if the feature flag context is valid
func (fc *FeatureFlagContext) Validate() error {
	if fc.FlagName == "" {
		return ErrInvalidFlagName
	}
	if fc.UserID == uuid.Nil {
		return ErrInvalidUserID
	}
	return nil
}

// FeatureFlagContextBuilder provides a fluent API for building feature flag contexts
type FeatureFlagContextBuilder struct {
	context FeatureFlagContext
}

// NewFeatureFlagContext creates a new feature flag context builder
func NewFeatureFlagContext() *FeatureFlagContextBuilder {
	return &FeatureFlagContextBuilder{
		context: FeatureFlagContext{
			Attributes: make(map[string]any),
		},
	}
}

func (b *FeatureFlagContextBuilder) WithUserID(userID uuid.UUID) *FeatureFlagContextBuilder {
	b.context.UserID = userID
	return b
}

func (b *FeatureFlagContextBuilder) WithTenantID(tenantID uuid.UUID) *FeatureFlagContextBuilder {
	b.context.TenantID = &tenantID
	return b
}

func (b *FeatureFlagContextBuilder) WithEntityID(entityID uuid.UUID) *FeatureFlagContextBuilder {
	b.context.EntityID = &entityID
	return b
}

func (b *FeatureFlagContextBuilder) WithFlagName(flagName string) *FeatureFlagContextBuilder {
	b.context.FlagName = flagName
	return b
}

func (b *FeatureFlagContextBuilder) WithAttribute(key string, value any) *FeatureFlagContextBuilder {
	if b.context.Attributes == nil {
		b.context.Attributes = make(map[string]any)
	}
	b.context.Attributes[key] = value
	return b
}

func (b *FeatureFlagContextBuilder) WithAttributes(attributes map[string]any) *FeatureFlagContextBuilder {
	b.context.Attributes = attributes
	return b
}

func (b *FeatureFlagContextBuilder) Build() *FeatureFlagContext {
	ctx := b.context
	return &ctx
}

// IsFeatureEnabled checks if a feature flag is enabled for the current context
func (h *FeatureFlagHelper) IsFeatureEnabled(ctx context.Context, flagName string) bool {
	if h.featureFlagService == nil {
		h.logWarn("Feature flag service not available", logger.Fields{
			"flag_name": flagName,
		})
		h.mu.RLock()
		defaultValue := h.config.DefaultValue
		h.mu.RUnlock()
		return defaultValue
	}

	if flagName == "" {
		h.logWarn("IsFeatureEnabled called with empty flag name")
		return false
	}

	userID, ok := shared.GetUserID(ctx)
	if !ok {
		h.logWarn("No user ID in context", logger.Fields{
			"flag_name": flagName,
		})
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

	// Get environment from config
	h.mu.RLock()
	environment := h.config.Environment
	h.mu.RUnlock()

	// Create evaluation context
	evalCtx := &featureflag.EvaluationContext{
		TenantID:    tenantID,
		UserID:      &userID,
		Environment: environment,
		Attributes:  attributes,
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
		h.logError("Feature flag evaluation failed", logger.Fields{
			"error":     err.Error(),
			"flag_name": flagName,
			"user_id":   userID.String(),
		})
		return false
	}

	h.logDebug("Feature flag evaluated", logger.Fields{
		"flag_name": flagName,
		"user_id":   userID.String(),
		"enabled":   enabled,
	})

	return enabled
}

// IsFeatureEnabledWithValidation checks feature flag with validation
func (h *FeatureFlagHelper) IsFeatureEnabledWithValidation(ctx context.Context, flagName string) (bool, error) {
	if flagName == "" {
		return false, ErrInvalidFlagName
	}

	if h.featureFlagService == nil {
		return false, ErrFeatureFlagServiceNotAvailable
	}

	_, ok := shared.GetUserID(ctx)
	if !ok {
		return false, ErrUserNotFound
	}

	// Use IsFeatureEnabled for actual evaluation
	enabled := h.IsFeatureEnabled(ctx, flagName)
	return enabled, nil
}

// IsFeatureEnabledForUser checks if a feature flag is enabled for a specific user
func (h *FeatureFlagHelper) IsFeatureEnabledForUser(ctx context.Context, flagName string, userID uuid.UUID, tenantID *uuid.UUID) bool {
	if h.featureFlagService == nil {
		h.logWarn("Feature flag service not available")
		return false
	}

	if flagName == "" {
		h.logWarn("IsFeatureEnabledForUser called with empty flag name")
		return false
	}

	if userID == uuid.Nil {
		h.logWarn("IsFeatureEnabledForUser called with nil user ID")
		return false
	}

	var tid uuid.UUID
	if tenantID != nil {
		tid = *tenantID
	}

	h.mu.RLock()
	environment := h.config.Environment
	h.mu.RUnlock()

	evalCtx := &featureflag.EvaluationContext{
		TenantID:    tid,
		UserID:      &userID,
		Environment: environment,
		Attributes:  make(map[string]string),
	}

	enabled, err := h.featureFlagService.IsEnabled(ctx, flagName, evalCtx)
	if err != nil {
		h.logError("Feature flag evaluation failed", logger.Fields{
			"error":     err.Error(),
			"flag_name": flagName,
			"user_id":   userID.String(),
		})
		return false
	}

	h.logDebug("Feature flag evaluated for user", logger.Fields{
		"flag_name": flagName,
		"user_id":   userID.String(),
		"enabled":   enabled,
	})

	return enabled
}

// GetFeatureFlagValue retrieves the value of a feature flag
func (h *FeatureFlagHelper) GetFeatureFlagValue(ctx context.Context, flagName string) (any, error) {
	if h.featureFlagService == nil {
		h.logError("Feature flag service not available", logger.Fields{
			"flag_name": flagName,
		})
		return nil, ErrFeatureFlagServiceNotAvailable
	}

	if flagName == "" {
		return nil, ErrInvalidFlagName
	}

	userID, ok := shared.GetUserID(ctx)
	if !ok {
		return nil, ErrUserNotFound
	}

	tenantID, _ := shared.GetTenantID(ctx)

	h.mu.RLock()
	environment := h.config.Environment
	h.mu.RUnlock()

	evalCtx := &featureflag.EvaluationContext{
		TenantID:    tenantID,
		UserID:      &userID,
		Environment: environment,
		Attributes:  make(map[string]string),
	}

	result, err := h.featureFlagService.EvaluateFlag(ctx, flagName, evalCtx)
	if err != nil {
		h.logError("Failed to get feature flag value", logger.Fields{
			"error":     err.Error(),
			"flag_name": flagName,
			"user_id":   userID.String(),
		})
		return nil, fmt.Errorf("failed to evaluate flag: %w", err)
	}

	h.logDebug("Feature flag value retrieved", logger.Fields{
		"flag_name": flagName,
		"user_id":   userID.String(),
		"value":     result.Value,
	})

	return result.Value, nil
}

// GetFeatureFlagValueWithDefault retrieves the value of a feature flag with a default
func (h *FeatureFlagHelper) GetFeatureFlagValueWithDefault(ctx context.Context, flagName string, defaultValue any) any {
	value, err := h.GetFeatureFlagValue(ctx, flagName)
	if err != nil {
		h.logDebug("Using default value for feature flag", logger.Fields{
			"flag_name":     flagName,
			"default_value": defaultValue,
			"error":         err.Error(),
		})
		return defaultValue
	}
	return value
}

// Batch Flag Checks

// HasAnyFeatureEnabled checks if any of the specified feature flags are enabled
func (h *FeatureFlagHelper) HasAnyFeatureEnabled(ctx context.Context, flagNames []string) bool {
	if len(flagNames) == 0 {
		h.logWarn("HasAnyFeatureEnabled called with empty flag names")
		return false
	}

	for i, flagName := range flagNames {
		if h.IsFeatureEnabled(ctx, flagName) {
			h.logDebug("HasAnyFeatureEnabled matched", logger.Fields{
				"matched_index": i,
				"flag_name":     flagName,
			})
			return true
		}
	}

	h.logDebug("HasAnyFeatureEnabled no match", logger.Fields{
		"flag_count": len(flagNames),
	})
	return false
}

// HasAllFeaturesEnabled checks if all of the specified feature flags are enabled
func (h *FeatureFlagHelper) HasAllFeaturesEnabled(ctx context.Context, flagNames []string) bool {
	if len(flagNames) == 0 {
		h.logWarn("HasAllFeaturesEnabled called with empty flag names")
		return true // Vacuous truth
	}

	for i, flagName := range flagNames {
		if !h.IsFeatureEnabled(ctx, flagName) {
			h.logDebug("HasAllFeaturesEnabled failed", logger.Fields{
				"failed_index": i,
				"flag_name":    flagName,
			})
			return false
		}
	}

	h.logDebug("HasAllFeaturesEnabled success", logger.Fields{
		"flag_count": len(flagNames),
	})
	return true
}

// GetEnabledFeatures returns all enabled features from the provided list
func (h *FeatureFlagHelper) GetEnabledFeatures(ctx context.Context, flagNames []string) []string {
	var enabledFlags []string

	for _, flagName := range flagNames {
		if h.IsFeatureEnabled(ctx, flagName) {
			enabledFlags = append(enabledFlags, flagName)
		}
	}

	h.logDebug("Retrieved enabled features", logger.Fields{
		"total_flags":   len(flagNames),
		"enabled_count": len(enabledFlags),
	})

	return enabledFlags
}

// CheckMultipleFlags checks multiple flags and returns their states
func (h *FeatureFlagHelper) CheckMultipleFlags(ctx context.Context, flagNames []string) map[string]bool {
	results := make(map[string]bool, len(flagNames))

	for _, flagName := range flagNames {
		results[flagName] = h.IsFeatureEnabled(ctx, flagName)
	}

	h.logDebug("Batch flag check completed", logger.Fields{
		"flag_count": len(flagNames),
	})

	return results
}

// Context Extraction

// GetUserFromContext extracts user information from request context
func (h *FeatureFlagHelper) GetUserFromContext(ctx context.Context) (uuid.UUID, *uuid.UUID, error) {
	userID, ok := shared.GetUserID(ctx)
	if !ok {
		h.logWarn("Failed to get user ID from context")
		return uuid.Nil, nil, ErrUserNotFound
	}

	tenantID, ok := shared.GetTenantID(ctx)
	if !ok {
		h.logDebug("No tenant ID in context", logger.Fields{
			"user_id": userID.String(),
		})
		return userID, nil, nil
	}

	return userID, &tenantID, nil
}

// Utility functions for common feature flag patterns

// IsBetaFeatureEnabled checks if beta features are enabled for the user
func (h *FeatureFlagHelper) IsBetaFeatureEnabled(ctx context.Context) bool {
	return h.IsFeatureEnabled(ctx, FeatureFlagBetaFeatures)
}

// IsExperimentalFeatureEnabled checks if experimental features are enabled
func (h *FeatureFlagHelper) IsExperimentalFeatureEnabled(ctx context.Context, experimentName string) bool {
	if experimentName == "" {
		h.logWarn("IsExperimentalFeatureEnabled called with empty experiment name")
		return false
	}
	return h.IsFeatureEnabled(ctx, "experimental_"+experimentName)
}

// IsUIFeatureEnabled checks if a specific UI feature is enabled
func (h *FeatureFlagHelper) IsUIFeatureEnabled(ctx context.Context, featureName string) bool {
	if featureName == "" {
		h.logWarn("IsUIFeatureEnabled called with empty feature name")
		return false
	}
	return h.IsFeatureEnabled(ctx, "ui_"+featureName)
}

// IsIntegrationEnabled checks if a specific integration is enabled
func (h *FeatureFlagHelper) IsIntegrationEnabled(ctx context.Context, integrationName string) bool {
	if integrationName == "" {
		h.logWarn("IsIntegrationEnabled called with empty integration name")
		return false
	}
	return h.IsFeatureEnabled(ctx, "integration_"+integrationName)
}

// IsModuleEnabled checks if a specific module is enabled
func (h *FeatureFlagHelper) IsModuleEnabled(ctx context.Context, moduleName string) bool {
	if moduleName == "" {
		h.logWarn("IsModuleEnabled called with empty module name")
		return false
	}
	return h.IsFeatureEnabled(ctx, "module_"+moduleName)
}

// Error definitions
var (
	ErrFeatureFlagServiceNotAvailable = errors.New("feature flag service not available")
	ErrInvalidFlagName                = errors.New("invalid flag name")
)

// Common feature flag constants
const (
	// UI Features
	FeatureFlagNewDashboard    = "ui_new_dashboard"
	FeatureFlagAdvancedReports = "ui_advanced_reports"
	FeatureFlagBulkOperations  = "ui_bulk_operations"
	FeatureFlagDarkMode        = "ui_dark_mode"
	FeatureFlagMobileLayout    = "ui_mobile_layout"

	// Module Features
	FeatureFlagFinanceModule   = "module_finance"
	FeatureFlagInventoryModule = "module_inventory"
	FeatureFlagHRModule        = "module_hr"
	FeatureFlagCRMModule       = "module_crm"
	FeatureFlagProjectModule   = "module_project"

	// Integration Features
	FeatureFlagSlackIntegration = "integration_slack"
	FeatureFlagEmailIntegration = "integration_email"
	FeatureFlagSSOIntegration   = "integration_sso"
	FeatureFlagAPIv2            = "integration_api_v2"

	// Experimental Features
	FeatureFlagAIAssistant       = "experimental_ai_assistant"
	FeatureFlagAdvancedAnalytics = "experimental_advanced_analytics"
	FeatureFlagVoiceCommands     = "experimental_voice_commands"

	// Beta Features
	FeatureFlagBetaFeatures  = "beta_features"
	FeatureFlagEarlyAccess   = "early_access"
	FeatureFlagDeveloperMode = "developer_mode"
)
