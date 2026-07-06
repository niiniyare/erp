package errors

// ─────────────────────────────────────────────────────────────────────────
// SENTINEL ERRORS
//
// These are convenience package-level values for the most commonly
// referenced errors, so call sites can write:
//
//	return nil, errors.ErrUserNotFound
//
// instead of:
//
//	return nil, errors.New(errors.CodeUserNotFound)
//
// SAFETY: because AwoError is immutable (every With* returns a new value,
// see errors.go), it is completely safe to hold onto these as shared
// package-level vars and call With* on them at any call site:
//
//	return nil, errors.ErrUserNotFound.WithDetail("user_id", id)
//
// This does NOT mutate the shared ErrUserNotFound value — WithDetail
// clones it first. That is the crucial difference from the previous
// version of this package, where With* mutated in place and sharing a
// package-level var across concurrent tenant requests was a real data
// race / cross-tenant leak risk.
//
// These sentinels are for the common case only. For anything that needs
// per-call context (an ID, a count, a period), prefer the New(...)/Newf(...)
// constructors or the dedicated helper constructors further down this file
// — reaching for a sentinel and immediately calling three With* methods on
// it is a sign you want one of those instead.
// ─────────────────────────────────────────────────────────────────────────

var (
	// IAM
	ErrUserNotFound        = New(CodeUserNotFound)
	ErrUserExists          = New(CodeUserExists)
	ErrEmailExists         = New(CodeEmailExists)
	ErrUsernameExists      = New(CodeUsernameExists)
	ErrInvalidCredentials  = New(CodeInvalidCredentials)
	ErrAuthenticationFail  = New(CodeAuthenticationFail)
	ErrAccountLocked       = New(CodeAccountLocked)
	ErrMFARequired         = New(CodeMFARequired)
	ErrMFAInvalid          = New(CodeMFAInvalid)
	ErrRoleNotFound        = New(CodeRoleNotFound)
	ErrRoleExists          = New(CodeRoleExists)
	ErrUnauthorized        = New(CodeUnauthorized)
	ErrForbidden           = New(CodeForbidden)
	ErrAPIKeyNotFound      = New(CodeAPIKeyNotFound)
	ErrInvitationNotFound  = New(CodeInvitationNotFound)
	ErrInvitationExpired   = New(CodeInvitationExpired)
	ErrPasswordTooWeak     = New(CodePasswordTooWeak)
	ErrPasswordReused      = New(CodePasswordReused)

	// Tenant
	ErrTenantNotFound         = New(CodeTenantNotFound)
	ErrTenantExists           = New(CodeTenantExists)
	ErrTenantIDNotInContext   = New(CodeTenantContextMissing)
	ErrTenantSuspendedGeneric = New(CodeTenantSuspended)
	ErrSubdomainExists        = New(CodeSubdomainExists)
	ErrSubdomainSoftDeleted   = New(CodeSubdomainSoftDeleted)
	ErrFeatureNotEnabled      = New(CodeFeatureNotEnabled)

	// Entity
	ErrEntityNotFound      = New(CodeEntityNotFound)
	ErrEntityNameExists    = New(CodeEntityNameExists)
	ErrEntityCodeExists    = New(CodeEntityCodeExists)
	ErrInvalidParentEntity = New(CodeInvalidParentEntity)
	ErrCircularReference   = New(CodeCircularReference)
	ErrEntityHasChildren   = New(CodeEntityHasChildren)
	ErrInvalidEntityType   = New(CodeInvalidEntityType)

	// ABAC
	ErrPolicyNotFound         = New(CodePolicyNotFound)
	ErrPolicyInvalid          = New(CodePolicyInvalid)
	ErrPolicyConflict         = New(CodePolicyConflict)
	ErrPolicyExpired          = New(CodePolicyExpired)
	ErrAttributeNotFound      = New(CodeAttributeNotFound)
	ErrAttributeInvalid       = New(CodeAttributeInvalid)
	ErrAttributeExpired       = New(CodeAttributeExpired)
	ErrEvaluationFailed       = New(CodeEvaluationFailed)
	ErrEvaluationTimeout      = New(CodeEvaluationTimeout)
	ErrInsufficientAttributes = New(CodeInsufficientAttributes)

	// General
	ErrInvalidInput = New(CodeInvalidInput)
	ErrNotFound     = New(CodeNotFound)
)

// ─────────────────────────────────────────────────────────────────────────
// ENRICHMENT CONSTRUCTORS
//
// For the common cases that always need at least one piece of per-call
// context (an ID, an email, a count), these save call sites from having
// to remember which detail key name to use, keeping key names consistent
// everywhere they're logged or matched on.
// ─────────────────────────────────────────────────────────────────────────

// NewUserNotFoundError builds ErrUserNotFound enriched with the user ID
// that could not be found.
func NewUserNotFoundError(userID string) *AwoError {
	return ErrUserNotFound.WithDetail("user_id", userID)
}

// NewUserNotFoundByEmailError builds ErrUserNotFound enriched with the
// email address that could not be found.
func NewUserNotFoundByEmailError(email string) *AwoError {
	return ErrUserNotFound.WithDetail("email", email)
}

// NewEntityNotFoundError builds ErrEntityNotFound enriched with the
// entity ID that could not be found.
func NewEntityNotFoundError(entityID string) *AwoError {
	return ErrEntityNotFound.WithDetail("entity_id", entityID)
}

// NewRoleNotFoundError builds ErrRoleNotFound enriched with the role ID.
func NewRoleNotFoundError(roleID string) *AwoError {
	return ErrRoleNotFound.WithDetail("role_id", roleID)
}

// NewTenantNotFoundError builds ErrTenantNotFound enriched with the
// tenant ID that could not be found.
func NewTenantNotFoundError(tenantID string) *AwoError {
	return ErrTenantNotFound.WithDetail("tenant_id", tenantID).WithTenant(tenantID)
}

// NewInvalidCredentialsError builds ErrInvalidCredentials enriched with
// the login attempt context, escalating severity to Warning once the
// attempt count crosses the lockout-warning threshold.
func NewInvalidCredentialsError(email string, attemptCount int) *AwoError {
	err := ErrInvalidCredentials.
		WithDetail("email", email).
		WithDetail("attempt_count", attemptCount)

	const lockoutWarningThreshold = 3
	if attemptCount >= lockoutWarningThreshold {
		err = err.
			WithSuggestion("Multiple failed attempts detected - account may be locked").
			WithSeverity(SeverityWarning)
	}
	return err
}

// NewEntityNameExistsError builds ErrEntityNameExists enriched with the
// conflicting name.
func NewEntityNameExistsError(name string) *AwoError {
	return ErrEntityNameExists.WithDetail("name", name)
}

// NewEntityCodeExistsError builds ErrEntityCodeExists enriched with the
// conflicting code.
func NewEntityCodeExistsError(code string) *AwoError {
	return ErrEntityCodeExists.WithDetail("code", code)
}

// NewCircularReferenceError builds ErrCircularReference enriched with the
// entity/parent pair that would have formed a cycle.
func NewCircularReferenceError(entityID, parentID string) *AwoError {
	return ErrCircularReference.
		WithDetail("entity_id", entityID).
		WithDetail("parent_id", parentID)
}

// NewInvitationExpiredError builds ErrInvitationExpired enriched with the
// invitation ID and its expiry timestamp (as a string — callers pass
// whatever format they already have, e.g. time.Time.Format(time.RFC3339)).
func NewInvitationExpiredError(invitationID, expiredAt string) *AwoError {
	return ErrInvitationExpired.
		WithDetail("invitation_id", invitationID).
		WithDetail("expired_at", expiredAt)
}

// NewFeatureNotEnabledError builds ErrFeatureNotEnabled enriched with
// which feature was gated and which plan unlocks it.
func NewFeatureNotEnabledError(feature, planRequired string) *AwoError {
	return ErrFeatureNotEnabled.
		WithDetail("feature", feature).
		WithDetail("plan_required", planRequired).
		WithSuggestion("Upgrade to " + planRequired + " plan to access this feature")
}

// NewTenantSuspendedError builds a tenant-suspended error for a specific
// tenant ID.
func NewTenantSuspendedError(tenantID string) *AwoError {
	return ErrTenantSuspendedGeneric.WithDetail("tenant_id", tenantID).WithTenant(tenantID)
}

// NewTenantLimitExceededError reports that a tenant has hit a plan limit
// (users, entities, storage, API calls, ...).
func NewTenantLimitExceededError(tenantID, limitType string, current, max int64) *AwoError {
	return New(CodeTenantLimitExceeded).
		WithDetail("tenant_id", tenantID).
		WithDetail("limit_type", limitType).
		WithDetail("current", current).
		WithDetail("maximum", max).
		WithTenant(tenantID)
}

// ─────────────────────────────────────────────────────────────────────────
// ACCOUNTING / FINANCE CONSTRUCTORS
// (kept here rather than a separate file since they follow the exact same
// "sentinel + enrichment" shape as everything above)
// ─────────────────────────────────────────────────────────────────────────

// NewInvalidAccountingPeriodError reports a period whose start is not
// before its end.
func NewInvalidAccountingPeriodError(periodStart, periodEnd string) *AwoError {
	return New(CodeInvalidAccountingPeriod).
		WithDetail("period_start", periodStart).
		WithDetail("period_end", periodEnd)
}

// NewAccountingPeriodClosedError reports an attempt to post into a closed period.
func NewAccountingPeriodClosedError(period string) *AwoError {
	return New(CodeAccountingPeriodClosed).WithDetail("period", period)
}

// NewInsufficientBalanceError reports that an account does not have
// enough balance for an operation.
func NewInsufficientBalanceError(accountID string, required, available float64) *AwoError {
	return New(CodeInsufficientBalance).
		WithDetail("account_id", accountID).
		WithDetail("required", required).
		WithDetail("available", available)
}

// NewDuplicateTransactionError reports that a transaction with this ID
// has already been recorded — useful for enforcing idempotent-consumer
// guarantees at the ledger boundary.
func NewDuplicateTransactionError(transactionID string) *AwoError {
	return New(CodeDuplicateTransaction).WithDetail("transaction_id", transactionID)
}

// NewThirdPartyAPIFailureError reports a failure calling an external
// system (M-Pesa/Daraja, KRA eTIMS, Africa's Talking, Twilio, ...).
// Retryable defaults to true from the registry; override with
// .WithRetryable(false) for third-party errors you know are permanent
// (e.g. a 4xx validation rejection from the third party).
func NewThirdPartyAPIFailureError(service string, statusCode int) *AwoError {
	return New(CodeThirdPartyAPIFailure).
		WithMessage("external service " + service + " failed").
		WithDetail("service", service).
		WithDetail("status_code", statusCode)
}
