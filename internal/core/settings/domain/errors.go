package domain

import "errors"

// Domain errors for the Settings module
var (
	// Configuration errors
	ErrConfigurationNotFound           = errors.New("configuration not found")
	ErrConfigurationNotOverridable     = errors.New("configuration cannot be overridden")
	ErrCannotResetSystemConfiguration  = errors.New("cannot reset system configuration")
	ErrCannotOverrideWithLowerPriority = errors.New("cannot override with lower priority source")
	ErrInheritedValueMustHaveParent    = errors.New("inherited value must have parent source")
	ErrEntityConfigRequiresEntityID    = errors.New("entity configuration requires entity ID")

	// Module errors
	ErrModuleRequired = errors.New("module name is required")
	ErrInvalidModule  = errors.New("invalid module name")

	// Configuration key errors
	ErrKeyRequired          = errors.New("configuration key is required")
	ErrConfigKeyRequired    = errors.New("config key is required")
	ErrInvalidConfigKey     = errors.New("invalid configuration key")
	ErrInvalidKeyFormat     = errors.New("invalid key format: must be lowercase letters, numbers, and underscores only")
	ErrKeyTooLong           = errors.New("key too long: maximum 100 characters")
	ErrInvalidFullKeyFormat = errors.New("invalid full key format: must be 'module.key'")

	// Configuration value errors
	ErrInvalidValueForType = errors.New("invalid value for data type")
	ErrInvalidDataType     = errors.New("invalid data type")
	ErrWrongDataType       = errors.New("wrong data type for operation")
	ErrValueNotString      = errors.New("value is not a string")
	ErrValueNotInteger     = errors.New("value is not an integer")
	ErrValueNotBoolean     = errors.New("value is not a boolean")
	ErrValueNotDecimal     = errors.New("value is not a decimal")
	ErrValueNotJSON        = errors.New("value is not a JSON object")

	// Configuration source errors
	ErrInvalidConfigSource = errors.New("invalid configuration source")

	// Template errors
	ErrTemplateNotFound             = errors.New("template not found")
	ErrTemplateNotActive            = errors.New("template is not active")
	ErrTemplateInvalidCategory      = errors.New("invalid template category")
	ErrTemplateInvalidVersion       = errors.New("invalid template version")
	ErrTemplateConfigurationInvalid = errors.New("template configuration is invalid")
	ErrTemplateConflictResolution   = errors.New("invalid conflict resolution strategy")
	ErrTemplateDependencyNotMet     = errors.New("template dependency not met")
	ErrTemplateNotApplicable        = errors.New("template not applicable to target")

	// Template application errors
	ErrTemplateApplicationFailed = errors.New("template application failed")
	ErrTemplateConflict          = errors.New("template configuration conflict")

	// Validation errors
	ErrValidationFailed      = errors.New("validation failed")
	ErrInvalidValidationRule = errors.New("invalid validation rule")

	// Permission errors
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrFeatureFlagNotEnabled   = errors.New("required feature flag not enabled")

	// Audit errors
	ErrAuditTrailRequired = errors.New("audit trail is required")
	ErrInvalidOperation   = errors.New("invalid audit operation")

	// Repository errors
	ErrRepositoryOperation = errors.New("repository operation failed")
	ErrConcurrentUpdate    = errors.New("concurrent update detected")
	ErrOptimisticLocking   = errors.New("optimistic locking failure")

	// Cache errors
	ErrCacheOperation = errors.New("cache operation failed")

	// Bulk operation errors
	ErrBulkOperationFailed      = errors.New("bulk operation failed")
	ErrBulkOperationPartialFail = errors.New("bulk operation partially failed")
	ErrInvalidBulkTarget        = errors.New("invalid bulk operation target")

	// Configuration definition errors
	ErrConfigDefinitionNotFound = errors.New("configuration definition not found")
	ErrConfigDefinitionExists   = errors.New("configuration definition already exists")
	ErrInvalidDefaultValue      = errors.New("invalid default value for data type")

	// Search errors
	ErrInvalidSearchCriteria = errors.New("invalid search criteria")
	ErrSearchFailed          = errors.New("configuration search failed")

	// Event errors
	ErrEventPublishFailed = errors.New("failed to publish domain event")

	// Context errors
	ErrInvalidTenantContext = errors.New("invalid tenant context")
	ErrMissingTenantContext = errors.New("missing tenant context")
	ErrInvalidEntityContext = errors.New("invalid entity context")
)

// IsNotFoundError returns true if the error indicates a resource was not found
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrConfigurationNotFound) ||
		errors.Is(err, ErrTemplateNotFound) ||
		errors.Is(err, ErrConfigDefinitionNotFound)
}

// IsValidationError returns true if the error is a validation error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrValidationFailed) ||
		errors.Is(err, ErrInvalidValueForType) ||
		errors.Is(err, ErrInvalidKeyFormat) ||
		errors.Is(err, ErrInvalidModule) ||
		errors.Is(err, ErrInvalidDataType) ||
		errors.Is(err, ErrInvalidValidationRule)
}

// IsPermissionError returns true if the error is a permission-related error
func IsPermissionError(err error) bool {
	return errors.Is(err, ErrInsufficientPermissions) ||
		errors.Is(err, ErrFeatureFlagNotEnabled) ||
		errors.Is(err, ErrConfigurationNotOverridable)
}

// IsConcurrencyError returns true if the error is related to concurrency
func IsConcurrencyError(err error) bool {
	return errors.Is(err, ErrConcurrentUpdate) ||
		errors.Is(err, ErrOptimisticLocking)
}

// IsBusinessRuleError returns true if the error represents a business rule violation
func IsBusinessRuleError(err error) bool {
	return errors.Is(err, ErrCannotResetSystemConfiguration) ||
		errors.Is(err, ErrCannotOverrideWithLowerPriority) ||
		errors.Is(err, ErrInheritedValueMustHaveParent) ||
		errors.Is(err, ErrEntityConfigRequiresEntityID) ||
		errors.Is(err, ErrTemplateNotApplicable) ||
		errors.Is(err, ErrTemplateDependencyNotMet)
}
