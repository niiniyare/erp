package errors

// ─── ERROR CODES ─────────────────────────────────────────────

// Tenant Error Codes
const (
	CodeTenantExists         = "TENANT_EXISTS"
	CodeTenantNotFound       = "TENANT_NOT_FOUND"
	CodeTenantContextMissing = "TENANT_CONTEXT_MISSING"
	CodeSubdomainExists      = "SUBDOMAIN_EXISTS"
	CodeSubdomainSoftDeleted = "SUBDOMAIN_SOFT_DELETED"
	CodeFeatureNotEnabled    = "FEATURE_NOT_ENABLED"
	CodeTenantSuspended      = "TENANT_SUSPENDED"
	CodeTenantLimitExceeded  = "TENANT_LIMIT_EXCEEDED"
)

// User Error Codes
const (
	CodeUserNotFound         = "USER_NOT_FOUND"
	CodeUserExists           = "USER_EXISTS"
	CodeEmailExists          = "EMAIL_EXISTS"
	CodeUsernameExists       = "USERNAME_EXISTS"
	CodeInvalidCredentials   = "INVALID_CREDENTIALS"
	CodeAuthenticationFailed = "AUTHENTICATION_FAILED"
	CodeInvalidUserType      = "INVALID_USER_TYPE"
	CodeInvalidAccountStatus = "INVALID_ACCOUNT_STATUS"
	CodeAccountLocked        = "ACCOUNT_LOCKED"
)

// Role & Permission Error Codes
const (
	CodeRoleNotFound = "ROLE_NOT_FOUND"
	CodeRoleExists   = "ROLE_EXISTS"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
)

// Invitation Error Codes
const (
	CodeInvitationNotFound = "INVITATION_NOT_FOUND"
	CodeInvitationExpired  = "INVITATION_EXPIRED"
)

// API Key Error Codes
const (
	CodeAPIKeyNotFound = "API_KEY_NOT_FOUND"
)

// Entity Error Codes
const (
	CodeEntityNotFound      = "ENTITY_NOT_FOUND"
	CodeEntityNameExists    = "ENTITY_NAME_EXISTS"
	CodeEntityCodeExists    = "ENTITY_CODE_EXISTS"
	CodeInvalidParentEntity = "INVALID_PARENT_ENTITY"
	CodeCircularReference   = "CIRCULAR_REFERENCE"
	CodeEntityHasChildren   = "ENTITY_HAS_CHILDREN"
	CodeInvalidEntityType   = "INVALID_ENTITY_TYPE"
)

// ABAC Error Codes
const (
	// Policy Errors
	CodePolicyNotFound = "POLICY_NOT_FOUND"
	CodePolicyInvalid  = "POLICY_INVALID"
	CodePolicyConflict = "POLICY_CONFLICT"
	CodePolicyExpired  = "POLICY_EXPIRED"

	// Attribute Errors
	CodeAttributeNotFound          = "ATTRIBUTE_NOT_FOUND"
	CodeAttributeInvalid           = "ATTRIBUTE_INVALID"
	CodeAttributeExpired           = "ATTRIBUTE_EXPIRED"
	CodeAttributeDefinitionInvalid = "ATTRIBUTE_DEFINITION_INVALID"

	// Evaluation Errors
	CodeEvaluationFailed         = "EVALUATION_FAILED"
	CodeEvaluationTimeout        = "EVALUATION_TIMEOUT"
	CodeCombiningAlgorithmFailed = "COMBINING_ALGORITHM_FAILED"
	CodeInsufficientAttributes   = "INSUFFICIENT_ATTRIBUTES"
)

// General Error Codes
const (
	CodeInvalidInput     = "INVALID_INPUT"
	CodeNotFound         = "NOT_FOUND"
	CodeUnknownError     = "UNKNOWN_ERROR"
	CodeInternalError    = "INTERNAL_ERROR"
	CodeMultipleErrors   = "MULTIPLE_ERRORS"
	CodeValidationFailed = "VALIDATION_FAILED"
)

// Financial Error Codes
const (
	CodeInvalidAccountingPeriod = "INVALID_ACCOUNTING_PERIOD"
	CodeAccountingPeriodClosed  = "ACCOUNTING_PERIOD_CLOSED"
	CodeInsufficientBalance     = "INSUFFICIENT_BALANCE"
	CodeDuplicateTransaction    = "DUPLICATE_TRANSACTION"
)

// Integration Error Codes
const (
	CodeThirdPartyAPIFailure = "THIRD_PARTY_API_FAILURE"
)

// Repository Error Codes
const (
	CodeRejectFailed      = "REJECT_FAILED"
	CodeDatabaseError     = "DATABASE_ERROR"
	CodeConnectionFailed  = "CONNECTION_FAILED"
	CodeQueryFailed       = "QUERY_FAILED"
	CodeTransactionFailed = "TRANSACTION_FAILED"
)

// Validation Error Codes
const (
	CodeRequired      = "REQUIRED"
	CodeInvalidFormat = "INVALID_FORMAT"
	CodeTooShort      = "TOO_SHORT"
	CodeTooLong       = "TOO_LONG"
	CodeInvalidRange  = "INVALID_RANGE"
	CodeInvalidEmail  = "INVALID_EMAIL"
	CodeInvalidURL    = "INVALID_URL"
	CodeInvalidUUID   = "INVALID_UUID"
	CodeInvalidDate   = "INVALID_DATE"
	CodeInvalidEnum   = "INVALID_ENUM"
)

// HTTP Error Codes (for internal mapping)
const (
	CodeHTTPBadRequest          = "BAD_REQUEST"
	CodeHTTPUnauthorized        = "UNAUTHORIZED"
	CodeHTTPForbidden           = "FORBIDDEN"
	CodeHTTPNotFound            = "NOT_FOUND"
	CodeHTTPConflict            = "CONFLICT"
	CodeHTTPInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeHTTPBadGateway          = "BAD_GATEWAY"
	CodeHTTPServiceUnavailable  = "SERVICE_UNAVAILABLE"
)

// System Error Codes
const (
	CodeInternalPanic = "INTERNAL_PANIC"
	CodeConfigError   = "CONFIG_ERROR"
	CodeStartupError  = "STARTUP_ERROR"
	CodeShutdownError = "SHUTDOWN_ERROR"
)
