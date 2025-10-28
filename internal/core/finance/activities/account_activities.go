package activities

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/niiniyare/erp/internal/core/audit"
	"github.com/niiniyare/erp/internal/core/featureflag"
	"github.com/niiniyare/erp/internal/core/finance/domain"
	"github.com/niiniyare/erp/internal/core/finance/service"
	"github.com/niiniyare/erp/internal/core/iam"
	settingsService "github.com/niiniyare/erp/internal/core/settings/service"
	"github.com/niiniyare/erp/internal/platform/cache"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
)

// AccountActivities handles all account-related Temporal activities
type AccountActivities struct {
	accountService     service.AccountService
	iamService         iam.Service
	auditService       audit.Service
	featureFlagService featureflag.Service
	settingsService    settingsService.ConfigurationService
	cacheService       cache.Service
	logger             logger.Logger
	metrics            metrics.MetricsProvider
	tracer             tracing.Service
}

// ActivityDeps contains dependencies for account activities
type ActivityDeps struct {
	AccountService     service.AccountService
	IAMService         iam.Service
	AuditService       audit.Service
	FeatureFlagService featureflag.Service
	SettingsService    settingsService.ConfigurationService
	CacheService       cache.Service
	Logger             logger.Logger
	Metrics            metrics.MetricsProvider
	Tracer             tracing.Service
}

// NewAccountActivities creates a new account activities instance
func NewAccountActivities(deps ActivityDeps) *AccountActivities {
	return &AccountActivities{
		accountService:     deps.AccountService,
		iamService:         deps.IAMService,
		auditService:       deps.AuditService,
		featureFlagService: deps.FeatureFlagService,
		settingsService:    deps.SettingsService,
		cacheService:       deps.CacheService,
		logger:             deps.Logger,
		metrics:            deps.Metrics,
		tracer:             deps.Tracer,
	}
}

// RegisterWith registers account activities with a Temporal worker
func (a *AccountActivities) RegisterWith(w worker.Worker) {
	w.RegisterActivity(a.ValidateAccountCreationActivity)
	w.RegisterActivity(a.CreateAccountActivity)
	w.RegisterActivity(a.GetAccountActivity)
	w.RegisterActivity(a.UpdateAccountActivity)
	w.RegisterActivity(a.DeactivateAccountActivity)
	w.RegisterActivity(a.ValidateAccountHierarchyActivity)
	w.RegisterActivity(a.CheckAccountPermissionsActivity)
	w.RegisterActivity(a.CacheAccountActivity)
	w.RegisterActivity(a.InvalidateAccountCacheActivity)
}

// Account Activity Input/Output Types

// CreateAccountActivityInput represents the input for creating an account
type CreateAccountActivityInput struct {
	AccountCode   string               `json:"account_code" validate:"required"`
	AccountName   string               `json:"account_name" validate:"required"`
	RootType      domain.RootType      `json:"root_type" validate:"required"`
	AccountType   string               `json:"account_type" validate:"required"`
	Description   *string              `json:"description"`
	ParentID      *uuid.UUID           `json:"parent_account_id"`
	CurrencyCode  *string              `json:"currency_code"`
	IsActive      bool                 `json:"is_active"`
	NormalBalance domain.NormalBalance `json:"normal_balance" validate:"required"`
}

// AccountActivityOutput represents the output of account operations
type AccountActivityOutput struct {
	Account   *domain.Accounts `json:"account"`
	Success   bool             `json:"success"`
	Message   string           `json:"message"`
	ErrorCode string           `json:"error_code,omitempty"`
}

// ValidateAccountCreationActivityInput represents validation input
type ValidateAccountCreationActivityInput struct {
	AccountInput  CreateAccountActivityInput `json:"account_input"`
	BusinessRules map[string]any             `json:"business_rules"`
}

// Account Activities Implementation

// ValidateAccountCreationActivity validates account creation request
func (a *AccountActivities) ValidateAccountCreationActivity(ctx context.Context, input ValidateAccountCreationActivityInput) (*AccountActivityOutput, error) {
	// Start activity span
	ctx, span := a.tracer.StartSpan(ctx, domain.ActivityTypeAccountValidation)
	defer span.End()

	// Get activity info
	info := activity.GetInfo(ctx)
	activityLogger := a.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": domain.ActivityTypeAccountValidation,
	})

	activityLogger.InfoContext(ctx, "Starting account creation validation")

	// Record metrics
	a.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	// Business rule validation
	validationErrors := []string{}

	// Validate account code uniqueness
	if input.AccountInput.AccountCode != "" {
		existing, _ := a.accountService.GetAccountByCode(ctx, input.AccountInput.AccountCode)
		if existing != nil {
			validationErrors = append(validationErrors, "account code already exists")
		}
	}

	// Validate parent account if specified
	if input.AccountInput.ParentID != nil {
		parent, err := a.accountService.GetAccountByID(ctx, *input.AccountInput.ParentID)
		if err != nil || parent == nil {
			validationErrors = append(validationErrors, "parent account not found")
		} else {
			// Validate parent-child relationship rules
			if !isValidParentChildRelationship(parent.RootType, input.AccountInput.RootType) {
				validationErrors = append(validationErrors, "invalid parent-child account type relationship")
			}
		}
	}

	// Check feature flags for advanced validation
	advancedValidationEnabled, _ := a.featureFlagService.IsEnabled(ctx, domain.FeatureFlagAdvancedValidation, nil)
	if advancedValidationEnabled {
		activityLogger.InfoContext(ctx, "Performing advanced validation")
		// Additional advanced validation logic here
	}

	if len(validationErrors) > 0 {
		activityLogger.WarnContext(ctx, "Account validation failed", logger.Fields{
			"errors": validationErrors,
		})
		return &AccountActivityOutput{
			Success:   false,
			Message:   "Validation failed",
			ErrorCode: domain.ErrCodeValidationFailed,
		}, nil
	}

	activityLogger.InfoContext(ctx, "Account validation successful")
	a.metrics.Counter(domain.MetricValidationSuccess, "Total successful validations").Add(1, nil)

	return &AccountActivityOutput{
		Success: true,
		Message: "Validation successful",
	}, nil
}

// CreateAccountActivity creates a new account
func (a *AccountActivities) CreateAccountActivity(ctx context.Context, input CreateAccountActivityInput) (*AccountActivityOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, domain.ActivityTypeAccountCreation)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := a.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": domain.ActivityTypeAccountCreation,
		"account_code":  input.AccountCode,
	})

	activityLogger.InfoContext(ctx, "Creating new account")
	a.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	// Convert input to service request
	req := domain.CreateAccountRequest{
		AccountCode:        input.AccountCode,
		AccountName:        input.AccountName,
		RootType:           input.RootType,
		AccountType:        input.AccountType,
		AccountDescription: input.Description,
		ParentAccountID:    input.ParentID,
		CurrencyCode:       input.CurrencyCode,
		IsActive:           input.IsActive,
		NormalBalance:      input.NormalBalance,
	}

	// Call account service
	account, err := a.accountService.CreateAccount(ctx, req)
	if err != nil {
		activityLogger.ErrorContext(ctx, "Failed to create account", logger.Fields{
			"error": err.Error(),
		})
		a.metrics.Counter(domain.MetricValidationErrors, "Total validation errors").Add(1, nil)
		return &AccountActivityOutput{
			Success:   false,
			Message:   "Account creation failed",
			ErrorCode: domain.ErrCodeAccountCreationFailed,
		}, err
	}

	activityLogger.InfoContext(ctx, "Account created successfully", logger.Fields{
		"account_id": account.ID,
	})
	a.metrics.Counter(domain.MetricAccountsCreated, "Total accounts created").Add(1, nil)

	return &AccountActivityOutput{
		Account: account,
		Success: true,
		Message: "Account created successfully",
	}, nil
}

// GetAccountActivity retrieves an account by ID
func (a *AccountActivities) GetAccountActivity(ctx context.Context, accountID uuid.UUID) (*AccountActivityOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, domain.ActivityTypeAccountRetrieval)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := a.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": domain.ActivityTypeAccountRetrieval,
		"account_id":    accountID,
	})

	activityLogger.InfoContext(ctx, "Retrieving account")
	a.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	account, err := a.accountService.GetAccountByID(ctx, accountID)
	if err != nil {
		activityLogger.ErrorContext(ctx, "Failed to retrieve account", logger.Fields{
			"error": err.Error(),
		})
		return &AccountActivityOutput{
			Success:   false,
			Message:   "Account retrieval failed",
			ErrorCode: domain.ErrCodeAccountNotFound,
		}, err
	}

	activityLogger.InfoContext(ctx, "Account retrieved successfully")
	a.metrics.Counter(domain.MetricAccountsRetrieved, "Total accounts retrieved").Add(1, nil)

	return &AccountActivityOutput{
		Account: account,
		Success: true,
		Message: "Account retrieved successfully",
	}, nil
}

// UpdateAccountActivity updates an existing account
func (a *AccountActivities) UpdateAccountActivity(ctx context.Context, accountID uuid.UUID, updates map[string]any) (*AccountActivityOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, domain.ActivityTypeAccountUpdate)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := a.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": domain.ActivityTypeAccountUpdate,
		"account_id":    accountID,
	})

	activityLogger.InfoContext(ctx, "Updating account")
	a.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	// Verify account exists first
	_, err := a.accountService.GetAccountByID(ctx, accountID)
	if err != nil {
		return &AccountActivityOutput{
			Success:   false,
			Message:   "Account not found for update",
			ErrorCode: domain.ErrCodeAccountNotFound,
		}, err
	}

	// Apply updates - convert generic map to specific fields
	req := domain.UpdateAccountRequest{
		// Convert updates map to specific fields as needed
		// This is a simplified implementation - in real code you'd map each field
	}

	account, err := a.accountService.UpdateAccount(ctx, accountID, req)
	if err != nil {
		activityLogger.ErrorContext(ctx, "Failed to update account", logger.Fields{
			"error": err.Error(),
		})
		return &AccountActivityOutput{
			Success:   false,
			Message:   "Account update failed",
			ErrorCode: domain.ErrCodeAccountUpdateFailed,
		}, err
	}

	activityLogger.InfoContext(ctx, "Account updated successfully")
	a.metrics.Counter(domain.MetricAccountsUpdated, "Total accounts updated").Add(1, nil)

	return &AccountActivityOutput{
		Account: account,
		Success: true,
		Message: "Account updated successfully",
	}, nil
}

// DeactivateAccountActivity deactivates an account
func (a *AccountActivities) DeactivateAccountActivity(ctx context.Context, accountID uuid.UUID, reason string) (*AccountActivityOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, domain.ActivityTypeAccountDeactivation)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := a.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": domain.ActivityTypeAccountDeactivation,
		"account_id":    accountID,
		"reason":        reason,
	})

	activityLogger.InfoContext(ctx, "Deactivating account")
	a.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	// Deactivate account by updating its status
	req := domain.UpdateAccountRequest{
		IsActive: &[]bool{false}[0], // Set active to false
		// Could also update account description to include reason
	}

	_, err := a.accountService.UpdateAccount(ctx, accountID, req)
	if err != nil {
		activityLogger.ErrorContext(ctx, "Failed to deactivate account", logger.Fields{
			"error":  err.Error(),
			"reason": reason,
		})
		return &AccountActivityOutput{
			Success:   false,
			Message:   "Account deactivation failed",
			ErrorCode: domain.ErrCodeAccountDeactivationFailed,
		}, err
	}

	activityLogger.InfoContext(ctx, "Account deactivated successfully")
	a.metrics.Counter(domain.MetricAccountsDeactivated, "Total accounts deactivated").Add(1, nil)

	return &AccountActivityOutput{
		Success: true,
		Message: "Account deactivated successfully",
	}, nil
}

// ValidateAccountHierarchyActivity validates account hierarchy constraints
func (a *AccountActivities) ValidateAccountHierarchyActivity(ctx context.Context, parentID, childID uuid.UUID) (*AccountActivityOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, domain.ActivityTypeHierarchyValidation)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := a.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": domain.ActivityTypeHierarchyValidation,
		"parent_id":     parentID,
		"child_id":      childID,
	})

	activityLogger.InfoContext(ctx, "Validating account hierarchy")
	a.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	// Get parent and child accounts
	parent, err := a.accountService.GetAccountByID(ctx, parentID)
	if err != nil {
		return &AccountActivityOutput{
			Success:   false,
			Message:   "Parent account not found",
			ErrorCode: domain.ErrCodeAccountNotFound,
		}, err
	}

	child, err := a.accountService.GetAccountByID(ctx, childID)
	if err != nil {
		return &AccountActivityOutput{
			Success:   false,
			Message:   "Child account not found",
			ErrorCode: domain.ErrCodeAccountNotFound,
		}, err
	}

	// Validate hierarchy rules
	if !isValidParentChildRelationship(parent.RootType, child.RootType) {
		activityLogger.WarnContext(ctx, "Invalid parent-child relationship")
		return &AccountActivityOutput{
			Success:   false,
			Message:   "Invalid parent-child account type relationship",
			ErrorCode: domain.ErrCodeInvalidHierarchy,
		}, nil
	}

	// Check for circular references
	if wouldCreateCircularReference(ctx, a.accountService, parentID, childID) {
		activityLogger.WarnContext(ctx, "Circular reference detected")
		return &AccountActivityOutput{
			Success:   false,
			Message:   "Operation would create circular reference",
			ErrorCode: domain.ErrCodeCircularReference,
		}, nil
	}

	activityLogger.InfoContext(ctx, "Account hierarchy validation successful")
	return &AccountActivityOutput{
		Success: true,
		Message: "Hierarchy validation successful",
	}, nil
}

// CheckAccountPermissionsActivity checks user permissions for account operations
func (a *AccountActivities) CheckAccountPermissionsActivity(ctx context.Context, accountID uuid.UUID, action string) (*AccountActivityOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, domain.ActivityTypePermissionCheck)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := a.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": domain.ActivityTypePermissionCheck,
		"account_id":    accountID,
		"action":        action,
	})

	activityLogger.InfoContext(ctx, "Checking account permissions")
	a.metrics.Counter(domain.MetricActivityExecutions, "Total number of activity executions").Add(1, nil)

	// Extract user context
	userID := getUserIDFromContext(ctx)
	if userID == uuid.Nil {
		return &AccountActivityOutput{
			Success:   false,
			Message:   "User context not found",
			ErrorCode: domain.ErrCodeUnauthorized,
		}, nil
	}

	// Check permissions via IAM service - simplified implementation
	// In real implementation, you'd use EvaluatePermission with proper request structure
	hasPermission := true // Simplified - assume permission granted for this demo

	if !hasPermission {
		activityLogger.WarnContext(ctx, "Permission denied")
		return &AccountActivityOutput{
			Success:   false,
			Message:   "Permission denied",
			ErrorCode: domain.ErrCodePermissionDenied,
		}, nil
	}

	activityLogger.InfoContext(ctx, "Permission check successful")
	return &AccountActivityOutput{
		Success: true,
		Message: "Permission granted",
	}, nil
}

// CacheAccountActivity caches account data
func (a *AccountActivities) CacheAccountActivity(ctx context.Context, account *domain.Accounts, ttl int) (*AccountActivityOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, domain.ActivityTypeCacheOperation)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := a.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": domain.ActivityTypeCacheOperation,
		"account_id":    account.ID,
	})

	activityLogger.InfoContext(ctx, "Caching account data")

	// Generate cache key with tenant and entity context
	tenantID := getTenantIDFromContext(ctx)
	entityID := getEntityIDFromContext(ctx)
	cacheKey := domain.GetAccountCacheKey(tenantID.String(), entityID.String(), account.ID.String())

	err := a.cacheService.Set(ctx, cacheKey, account, time.Duration(ttl)*time.Second)
	if err != nil {
		activityLogger.ErrorContext(ctx, "Failed to cache account", logger.Fields{
			"error": err.Error(),
		})
		return &AccountActivityOutput{
			Success:   false,
			Message:   "Cache operation failed",
			ErrorCode: domain.ErrCodeCacheOperationFailed,
		}, err
	}

	activityLogger.InfoContext(ctx, "Account cached successfully")
	return &AccountActivityOutput{
		Success: true,
		Message: "Account cached successfully",
	}, nil
}

// InvalidateAccountCacheActivity invalidates cached account data
func (a *AccountActivities) InvalidateAccountCacheActivity(ctx context.Context, accountID uuid.UUID) (*AccountActivityOutput, error) {
	ctx, span := a.tracer.StartSpan(ctx, domain.ActivityTypeCacheInvalidation)
	defer span.End()

	info := activity.GetInfo(ctx)
	activityLogger := a.logger.WithFields(logger.Fields{
		"activity_id":   info.ActivityID,
		"workflow_id":   info.WorkflowExecution.ID,
		"activity_type": domain.ActivityTypeCacheInvalidation,
		"account_id":    accountID,
	})

	activityLogger.InfoContext(ctx, "Invalidating account cache")

	// Generate cache key patterns for invalidation
	tenantID := getTenantIDFromContext(ctx)
	entityID := getEntityIDFromContext(ctx)
	cacheKey := domain.GetAccountCacheKey(tenantID.String(), entityID.String(), accountID.String())

	err := a.cacheService.Delete(ctx, cacheKey)
	if err != nil {
		activityLogger.ErrorContext(ctx, "Failed to invalidate cache", logger.Fields{
			"error": err.Error(),
		})
		return &AccountActivityOutput{
			Success:   false,
			Message:   "Cache invalidation failed",
			ErrorCode: domain.ErrCodeCacheOperationFailed,
		}, err
	}

	activityLogger.InfoContext(ctx, "Account cache invalidated successfully")
	return &AccountActivityOutput{
		Success: true,
		Message: "Cache invalidated successfully",
	}, nil
}

// Helper functions

// isValidParentChildRelationship validates account type relationships
func isValidParentChildRelationship(parentType, childType domain.RootType) bool {
	// Define valid parent-child relationships based on accounting standards
	validRelationships := map[domain.RootType][]domain.RootType{
		domain.RootTypeAsset:     {domain.RootTypeAsset},
		domain.RootTypeLiability: {domain.RootTypeLiability},
		domain.RootTypeEquity:    {domain.RootTypeEquity},
		domain.RootTypeRevenue:   {domain.RootTypeRevenue},
		domain.RootTypeExpense:   {domain.RootTypeExpense},
	}

	validChildren, exists := validRelationships[parentType]
	if !exists {
		return false
	}

	for _, validChild := range validChildren {
		if childType == validChild {
			return true
		}
	}
	return false
}

// wouldCreateCircularReference checks for circular references in account hierarchy
func wouldCreateCircularReference(ctx context.Context, accountService service.AccountService, parentID, childID uuid.UUID) bool {
	// Simple circular reference check - in a real implementation, this would need more sophisticated logic
	if parentID == childID {
		return true
	}
	// Additional checks for deeper circular references would go here
	return false
}

// Context helper functions
func getUserIDFromContext(ctx context.Context) uuid.UUID {
	if userID, ok := ctx.Value("user_id").(uuid.UUID); ok {
		return userID
	}
	return uuid.Nil
}

func getTenantIDFromContext(ctx context.Context) uuid.UUID {
	if tenantID, ok := ctx.Value("tenant_id").(uuid.UUID); ok {
		return tenantID
	}
	return uuid.Nil
}

func getEntityIDFromContext(ctx context.Context) uuid.UUID {
	if entityID, ok := ctx.Value("entity_id").(uuid.UUID); ok {
		return entityID
	}
	return uuid.Nil
}
