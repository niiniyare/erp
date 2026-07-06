package errors

// Code is a stable, machine-readable error identifier using a dot-namespace
// convention: "{module}.{entity}.{condition}".
//
// This mirrors the framework's existing dot-notation permission convention
// ({module}.{resource}.{action}, e.g. "accounting.journal.post") so anyone
// already familiar with Awo's RBAC permissions instantly understands how
// error codes are organised too.
//
// WHY NAMESPACED CODES INSTEAD OF FLAT ONES ("USER_NOT_FOUND"):
//
//  1. Collision-proof across modules. Both HR and Forecourt might have a
//     concept of an "assignment" — "hr.assignment.not_found" and
//     "forecourt.assignment.not_found" can coexist safely, whereas flat
//     "ASSIGNMENT_NOT_FOUND" would collide.
//
//  2. Greppable and self-documenting. Seeing "payroll.paye.bracket_not_found"
//     in a log line tells you the module and the exact condition without
//     needing to open this file.
//
//  3. Enables convention-based helpers. IsNotFound(err) below matches any
//     code ending in ".not_found" — new modules automatically get correct
//     "is this a not-found error?" behaviour without this package needing
//     to enumerate every module's codes by hand (contrast with the old
//     package's hand-maintained notFoundCodes []string list).
//
// GUIDANCE FOR ADDING NEW MODULE ERRORS:
// When a new module (e.g. Forecourt, Payroll, Notifications) needs errors,
// add its own const block here (or, once the module lives in its own Go
// package, define a local `Code` block there — Code is just a string type,
// so any package can mint its own codes without modifying this file).
// Then add a matching entry to the registry in registry.go.
type Code string

// ─────────────────────────────────────────────────────────────────────────
// CORE / SYSTEM-WIDE
// Used as fallbacks and for truly generic, cross-cutting conditions that
// don't belong to any one business module.
// ─────────────────────────────────────────────────────────────────────────
const (
	CodeUnknown          Code = "core.unknown"           // Fallback when nothing more specific applies
	CodeInternal         Code = "core.internal"          // Unexpected internal/framework failure
	CodeValidationFailed Code = "core.validation.failed" // One or more field-level validation errors
	CodeNotFound         Code = "core.resource.not_found"
	CodeInvalidInput     Code = "core.input.invalid"
)

// ─────────────────────────────────────────────────────────────────────────
// IAM (Identity & Access Management): users, credentials, MFA, roles
// ─────────────────────────────────────────────────────────────────────────
const (
	CodeUserNotFound       Code = "iam.user.not_found"
	CodeUserExists         Code = "iam.user.exists"
	CodeEmailExists        Code = "iam.email.exists"
	CodeUsernameExists     Code = "iam.username.exists"
	CodeInvalidCredentials Code = "iam.credentials.invalid"
	CodeAuthenticationFail Code = "iam.authentication.failed"
	CodeInvalidUserType    Code = "iam.user.invalid_type"
	CodeInvalidAcctStatus  Code = "iam.user.invalid_status"
	CodeAccountLocked      Code = "iam.account.locked"

	CodeMFARequired      Code = "iam.mfa.required"
	CodeMFAInvalid       Code = "iam.mfa.invalid"
	CodeMFANotEnabled    Code = "iam.mfa.not_enabled"
	CodeMFAAlreadyOn     Code = "iam.mfa.already_enabled"
	CodeMFANotConfigured Code = "iam.mfa.not_configured"

	CodePasswordResetTokenNotFound Code = "iam.password_reset.token_not_found"
	CodePasswordResetTokenExpired  Code = "iam.password_reset.token_expired"
	CodePasswordResetTokenUsed     Code = "iam.password_reset.token_used"
	CodePasswordTooWeak            Code = "iam.password.too_weak"
	CodePasswordReused             Code = "iam.password.reused"

	CodeRoleNotFound Code = "iam.role.not_found"
	CodeRoleExists   Code = "iam.role.exists"
	CodeUnauthorized Code = "iam.access.unauthorized" // not authenticated
	CodeForbidden    Code = "iam.access.forbidden"    // authenticated but not permitted

	CodeAPIKeyNotFound Code = "iam.api_key.not_found"

	CodeInvitationNotFound Code = "iam.invitation.not_found"
	CodeInvitationExpired  Code = "iam.invitation.expired"
)

// ─────────────────────────────────────────────────────────────────────────
// TENANT: tenant lifecycle, subdomains, plan/feature gating
// ─────────────────────────────────────────────────────────────────────────
const (
	CodeTenantNotFound       Code = "tenant.record.not_found"
	CodeTenantExists         Code = "tenant.record.exists"
	CodeTenantContextMissing Code = "tenant.context.missing" // request had no resolvable tenant
	CodeTenantSuspended      Code = "tenant.record.suspended"
	CodeTenantLimitExceeded  Code = "tenant.limit.exceeded"

	CodeSubdomainExists      Code = "tenant.subdomain.exists"
	CodeSubdomainSoftDeleted Code = "tenant.subdomain.soft_deleted"

	CodeFeatureNotEnabled Code = "tenant.feature.not_enabled" // plan/billing gated
)

// ─────────────────────────────────────────────────────────────────────────
// ENTITY: organisational hierarchy (companies, departments, cost centers, ...)
// ─────────────────────────────────────────────────────────────────────────
const (
	CodeEntityNotFound      Code = "entity.record.not_found"
	CodeEntityNameExists    Code = "entity.name.exists"
	CodeEntityCodeExists    Code = "entity.code.exists"
	CodeInvalidParentEntity Code = "entity.parent.invalid"
	CodeCircularReference   Code = "entity.hierarchy.circular"
	CodeEntityHasChildren   Code = "entity.children.exist" // blocks deletion
	CodeInvalidEntityType   Code = "entity.type.invalid"
)

// ─────────────────────────────────────────────────────────────────────────
// ABAC: attribute-based policy evaluation (layered on top of Casbin RBAC)
// ─────────────────────────────────────────────────────────────────────────
const (
	CodePolicyNotFound Code = "abac.policy.not_found"
	CodePolicyInvalid  Code = "abac.policy.invalid"
	CodePolicyConflict Code = "abac.policy.conflict"
	CodePolicyExpired  Code = "abac.policy.expired"

	CodeAttributeNotFound          Code = "abac.attribute.not_found"
	CodeAttributeInvalid           Code = "abac.attribute.invalid"
	CodeAttributeExpired           Code = "abac.attribute.expired"
	CodeAttributeDefinitionInvalid Code = "abac.attribute.definition_invalid"

	CodeEvaluationFailed         Code = "abac.evaluation.failed"
	CodeEvaluationTimeout        Code = "abac.evaluation.timeout"
	CodeCombiningAlgorithmFailed Code = "abac.evaluation.combining_failed"
	CodeInsufficientAttributes   Code = "abac.evaluation.insufficient_attributes"
)

// ─────────────────────────────────────────────────────────────────────────
// ACCOUNTING / FINANCE
// ─────────────────────────────────────────────────────────────────────────
const (
	CodeInvalidAccountingPeriod Code = "accounting.period.invalid"
	CodeAccountingPeriodClosed  Code = "accounting.period.closed"
	CodeInsufficientBalance     Code = "accounting.balance.insufficient"
	CodeDuplicateTransaction    Code = "accounting.transaction.duplicate"
)

// ─────────────────────────────────────────────────────────────────────────
// INTEGRATION: third-party systems (M-Pesa/Daraja, KRA eTIMS, Africa's Talking, ...)
// ─────────────────────────────────────────────────────────────────────────
const (
	CodeThirdPartyAPIFailure Code = "integration.thirdparty.failed"
)

// ─────────────────────────────────────────────────────────────────────────
// INFRA / REPOSITORY
// These should never be surfaced verbatim to an external API client (see
// http.go — repository-category errors get a generic message on purpose).
// ─────────────────────────────────────────────────────────────────────────
const (
	CodeDatabaseError     Code = "infra.database.error"
	CodeConnectionFailed  Code = "infra.connection.failed"
	CodeQueryFailed       Code = "infra.query.failed"
	CodeTransactionFailed Code = "infra.transaction.failed"
)

// ─────────────────────────────────────────────────────────────────────────
// SYSTEM / STARTUP
// ─────────────────────────────────────────────────────────────────────────
const (
	CodeInternalPanic Code = "system.panic"
	CodeConfigError   Code = "system.config.invalid"
	CodeStartupError  Code = "system.startup.failed"
	CodeShutdownError Code = "system.shutdown.failed"
)

// ─────────────────────────────────────────────────────────────────────────
// FIELD-LEVEL VALIDATION CODES
//
// These are NOT AwoError Codes — they identify the *kind* of validation
// failure on a single field inside a FieldError (see validation.go), which
// is a different, more granular concern than "the request failed with
// code X". Kept as plain strings (not the Code type) to make that
// distinction explicit at the type level.
// ─────────────────────────────────────────────────────────────────────────
const (
	FieldRequired      = "required"
	FieldInvalidFormat = "invalid_format"
	FieldTooShort      = "too_short"
	FieldTooLong       = "too_long"
	FieldInvalidRange  = "invalid_range"
	FieldInvalidEmail  = "invalid_email"
	FieldInvalidURL    = "invalid_url"
	FieldInvalidUUID   = "invalid_uuid"
	FieldInvalidDate   = "invalid_date"
	FieldInvalidEnum   = "invalid_enum"
)
