package errors

import "net/http"

// errorDef is the "default shape" of an error for a given Code: what
// message to show, what HTTP status to map to, which category/severity it
// belongs to, whether it's retryable by default, and what suggestions to
// offer. It exists so that defining a new error is ONE map entry instead
// of a 6-10 line `NewBusinessError(...).With...().With...()` chain
// repeated by hand for every single error in the system.
//
// This is also the natural seam for localisation: because Message here is
// just data, a locale-aware wrapper can swap this whole map (or just the
// Message field) per request locale — e.g. serving the sw-KE or af-SO
// translation of "user not found" — without touching any Go code that
// calls errors.New(CodeUserNotFound). (A LocalizedRegistry type built on
// top of this one is a natural next step once the sw-KE/af-SO message
// catalogues are ready.)
type errorDef struct {
	Message     string   // default human-readable message (English baseline)
	HTTPStatus  int      // HTTP status this error maps to
	Category    Category // routing/metrics category
	Severity    Severity // log-level hint
	Retryable   bool     // whether callers may safely retry by default
	Suggestions []string // default "what to do next" hints
}

// registry is the single source of truth for every Code's default shape.
// Add one entry here per Code you define in codes.go (or in a module's
// own code block) — this is the ONLY place that needs to change to add a
// new error to the framework.
var registry = map[Code]errorDef{

	// ── Core / system-wide ────────────────────────────────────────────
	CodeUnknown: {
		Message: "an unknown error occurred", HTTPStatus: http.StatusInternalServerError,
		Category: CategorySystem, Severity: SeverityError,
	},
	CodeInternal: {
		Message: "an internal error occurred", HTTPStatus: http.StatusInternalServerError,
		Category: CategorySystem, Severity: SeverityCritical,
	},
	CodeValidationFailed: {
		Message: "validation failed", HTTPStatus: http.StatusBadRequest,
		Category: CategoryValidation, Severity: SeverityError,
	},
	CodeNotFound: {
		Message: "resource not found", HTTPStatus: http.StatusNotFound,
		Category: CategoryBusiness, Severity: SeverityError,
		Suggestions: []string{"Verify the resource ID is correct"},
	},
	CodeInvalidInput: {
		Message: "invalid input provided", HTTPStatus: http.StatusBadRequest,
		Category: CategoryValidation, Severity: SeverityError,
		Suggestions: []string{"Check the request format and required fields"},
	},

	// ── IAM ───────────────────────────────────────────────────────────
	CodeUserNotFound: {
		Message: "user not found", HTTPStatus: http.StatusNotFound,
		Category: CategorySecurity, Severity: SeverityError,
		Suggestions: []string{"Verify the user ID or email is correct"},
	},
	CodeUserExists: {
		Message: "user already exists", HTTPStatus: http.StatusConflict,
		Category: CategorySecurity, Severity: SeverityError,
		Suggestions: []string{"Try logging in if you already have an account"},
	},
	CodeEmailExists: {
		Message: "email address already exists", HTTPStatus: http.StatusConflict,
		Category: CategoryValidation, Severity: SeverityError,
		Suggestions: []string{"Use a different email address"},
	},
	CodeUsernameExists: {
		Message: "username already exists", HTTPStatus: http.StatusConflict,
		Category: CategoryValidation, Severity: SeverityError,
		Suggestions: []string{"Choose a different username"},
	},
	CodeInvalidCredentials: {
		Message: "invalid email or password", HTTPStatus: http.StatusUnauthorized,
		Category: CategorySecurity, Severity: SeverityError,
		Suggestions: []string{"Check your email and password", "Use the 'Forgot Password' option if needed"},
	},
	CodeAuthenticationFail: {
		Message: "authentication failed", HTTPStatus: http.StatusUnauthorized,
		Category: CategorySecurity, Severity: SeverityError,
	},
	CodeInvalidUserType: {
		Message: "invalid user type", HTTPStatus: http.StatusBadRequest,
		Category: CategoryValidation, Severity: SeverityError,
	},
	CodeInvalidAcctStatus: {
		Message: "invalid account status", HTTPStatus: http.StatusBadRequest,
		Category: CategoryValidation, Severity: SeverityError,
	},
	CodeAccountLocked: {
		Message: "account is locked", HTTPStatus: http.StatusForbidden,
		Category: CategorySecurity, Severity: SeverityWarning,
		Suggestions: []string{"Contact an administrator to unlock your account"},
	},
	CodeMFARequired: {
		Message: "multi-factor authentication required", HTTPStatus: http.StatusAccepted,
		Category: CategorySecurity, Severity: SeverityInfo,
		Suggestions: []string{"Provide the 6-digit code from your authenticator app"},
	},
	CodeMFAInvalid: {
		Message: "invalid or expired MFA code", HTTPStatus: http.StatusUnauthorized,
		Category: CategorySecurity, Severity: SeverityError,
		Suggestions: []string{"Ensure your device clock is synchronized"},
	},
	CodePasswordResetTokenNotFound: {
		Message: "password reset token not found", HTTPStatus: http.StatusNotFound,
		Category: CategorySecurity, Severity: SeverityError,
		Suggestions: []string{"Request a new password reset link"},
	},
	CodePasswordResetTokenExpired: {
		Message: "password reset token has expired", HTTPStatus: http.StatusGone,
		Category: CategorySecurity, Severity: SeverityError,
		Suggestions: []string{"Request a new password reset link"},
	},
	CodePasswordResetTokenUsed: {
		Message: "password reset token has already been used", HTTPStatus: http.StatusGone,
		Category: CategorySecurity, Severity: SeverityError,
		Suggestions: []string{"Request a new password reset link"},
	},
	CodePasswordTooWeak: {
		Message: "password does not meet strength requirements", HTTPStatus: http.StatusBadRequest,
		Category: CategoryValidation, Severity: SeverityError,
		Suggestions: []string{"Use at least 12 characters including uppercase, lowercase, digit, and special character"},
	},
	CodePasswordReused: {
		Message: "password was recently used", HTTPStatus: http.StatusBadRequest,
		Category: CategorySecurity, Severity: SeverityError,
		Suggestions: []string{"Choose a password you have not used in the last 5 password changes"},
	},
	CodeRoleNotFound: {
		Message: "role not found", HTTPStatus: http.StatusNotFound,
		Category: CategorySecurity, Severity: SeverityError,
	},
	CodeRoleExists: {
		Message: "role already exists", HTTPStatus: http.StatusConflict,
		Category: CategorySecurity, Severity: SeverityError,
	},
	CodeUnauthorized: {
		Message: "authentication required", HTTPStatus: http.StatusUnauthorized,
		Category: CategorySecurity, Severity: SeverityError,
		Suggestions: []string{"Please log in to access this resource"},
	},
	CodeForbidden: {
		Message: "access forbidden", HTTPStatus: http.StatusForbidden,
		Category: CategorySecurity, Severity: SeverityError,
		Suggestions: []string{"Contact your administrator for access"},
	},
	CodeAPIKeyNotFound: {
		Message: "API key not found", HTTPStatus: http.StatusUnauthorized,
		Category: CategorySecurity, Severity: SeverityError,
	},
	CodeInvitationNotFound: {
		Message: "invitation not found", HTTPStatus: http.StatusNotFound,
		Category: CategoryBusiness, Severity: SeverityError,
	},
	CodeInvitationExpired: {
		Message: "invitation has expired", HTTPStatus: http.StatusGone,
		Category: CategoryBusiness, Severity: SeverityError,
		Suggestions: []string{"Request a new invitation"},
	},

	// ── Tenant ────────────────────────────────────────────────────────
	CodeTenantNotFound: {
		Message: "tenant not found", HTTPStatus: http.StatusNotFound,
		Category: CategoryTenant, Severity: SeverityError,
	},
	CodeTenantExists: {
		Message: "tenant already exists", HTTPStatus: http.StatusConflict,
		Category: CategoryTenant, Severity: SeverityError,
	},
	CodeTenantContextMissing: {
		Message: "tenant ID not found in request context", HTTPStatus: http.StatusBadRequest,
		Category: CategoryTenant, Severity: SeverityError,
		Suggestions: []string{"Ensure the request includes proper tenant identification"},
	},
	CodeTenantSuspended: {
		Message: "tenant account is suspended", HTTPStatus: http.StatusForbidden,
		Category: CategoryTenant, Severity: SeverityWarning,
		Suggestions: []string{"Contact support to reactivate your account"},
	},
	CodeTenantLimitExceeded: {
		Message: "tenant limit exceeded", HTTPStatus: http.StatusPaymentRequired,
		Category: CategoryTenant, Severity: SeverityWarning,
		Suggestions: []string{"Upgrade your plan to increase limits"},
	},
	CodeSubdomainExists: {
		Message: "subdomain already exists", HTTPStatus: http.StatusConflict,
		Category: CategoryTenant, Severity: SeverityError,
	},
	CodeSubdomainSoftDeleted: {
		Message: "subdomain is unavailable", HTTPStatus: http.StatusConflict,
		Category: CategoryTenant, Severity: SeverityError,
		Suggestions: []string{"This subdomain was recently used; choose another or contact support"},
	},
	CodeFeatureNotEnabled: {
		Message: "feature not enabled for this tenant", HTTPStatus: http.StatusForbidden,
		Category: CategoryTenant, Severity: SeverityWarning,
		Suggestions: []string{"Upgrade your plan to access this feature"},
	},

	// ── Entity ────────────────────────────────────────────────────────
	CodeEntityNotFound: {
		Message: "entity not found", HTTPStatus: http.StatusNotFound,
		Category: CategoryBusiness, Severity: SeverityError,
	},
	CodeEntityNameExists: {
		Message: "entity name already exists", HTTPStatus: http.StatusConflict,
		Category: CategoryBusiness, Severity: SeverityError,
	},
	CodeEntityCodeExists: {
		Message: "entity code already exists", HTTPStatus: http.StatusConflict,
		Category: CategoryBusiness, Severity: SeverityError,
	},
	CodeInvalidParentEntity: {
		Message: "invalid parent entity", HTTPStatus: http.StatusBadRequest,
		Category: CategoryBusiness, Severity: SeverityError,
	},
	CodeCircularReference: {
		Message: "circular reference detected in entity hierarchy", HTTPStatus: http.StatusBadRequest,
		Category: CategoryBusiness, Severity: SeverityError,
		Suggestions: []string{"An entity cannot be a parent of itself or its ancestors"},
	},
	CodeEntityHasChildren: {
		Message: "entity has child entities and cannot be deleted", HTTPStatus: http.StatusConflict,
		Category: CategoryBusiness, Severity: SeverityError,
		Suggestions: []string{"Remove or reassign child entities first, or deactivate instead of deleting"},
	},
	CodeInvalidEntityType: {
		Message: "invalid entity type", HTTPStatus: http.StatusBadRequest,
		Category: CategoryValidation, Severity: SeverityError,
	},

	// ── ABAC ──────────────────────────────────────────────────────────
	CodePolicyNotFound: {
		Message: "ABAC policy not found", HTTPStatus: http.StatusNotFound,
		Category: CategorySecurity, Severity: SeverityError,
	},
	CodePolicyInvalid: {
		Message: "ABAC policy validation failed", HTTPStatus: http.StatusBadRequest,
		Category: CategoryValidation, Severity: SeverityError,
	},
	CodePolicyConflict: {
		Message: "conflicting ABAC policies detected", HTTPStatus: http.StatusConflict,
		Category: CategorySecurity, Severity: SeverityError,
	},
	CodePolicyExpired: {
		Message: "ABAC policy has expired", HTTPStatus: http.StatusGone,
		Category: CategorySecurity, Severity: SeverityError,
	},
	CodeAttributeNotFound: {
		Message: "required attribute not found", HTTPStatus: http.StatusNotFound,
		Category: CategorySecurity, Severity: SeverityError,
	},
	CodeAttributeInvalid: {
		Message: "attribute value validation failed", HTTPStatus: http.StatusBadRequest,
		Category: CategoryValidation, Severity: SeverityError,
	},
	CodeAttributeExpired: {
		Message: "attribute value has expired", HTTPStatus: http.StatusGone,
		Category: CategorySecurity, Severity: SeverityError,
	},
	CodeAttributeDefinitionInvalid: {
		Message: "attribute definition validation failed", HTTPStatus: http.StatusBadRequest,
		Category: CategoryValidation, Severity: SeverityError,
	},
	CodeEvaluationFailed: {
		Message: "policy evaluation failed", HTTPStatus: http.StatusInternalServerError,
		Category: CategorySecurity, Severity: SeverityError,
	},
	CodeEvaluationTimeout: {
		Message: "policy evaluation timed out", HTTPStatus: http.StatusRequestTimeout,
		Category: CategorySecurity, Severity: SeverityError, Retryable: true,
	},
	CodeCombiningAlgorithmFailed: {
		Message: "policy combining algorithm failed", HTTPStatus: http.StatusInternalServerError,
		Category: CategorySecurity, Severity: SeverityError,
	},
	CodeInsufficientAttributes: {
		Message: "insufficient attributes for policy evaluation", HTTPStatus: http.StatusBadRequest,
		Category: CategorySecurity, Severity: SeverityError,
	},

	// ── Accounting / Finance ──────────────────────────────────────────
	CodeInvalidAccountingPeriod: {
		Message: "invalid accounting period", HTTPStatus: http.StatusBadRequest,
		Category: CategoryBusiness, Severity: SeverityError,
		Suggestions: []string{"Ensure the period start date is before the end date"},
	},
	CodeAccountingPeriodClosed: {
		Message: "accounting period is closed", HTTPStatus: http.StatusConflict,
		Category: CategoryBusiness, Severity: SeverityError,
		Suggestions: []string{"Contact your accountant to reopen the period if necessary"},
	},
	CodeInsufficientBalance: {
		Message: "insufficient account balance", HTTPStatus: http.StatusConflict,
		Category: CategoryBusiness, Severity: SeverityError,
	},
	CodeDuplicateTransaction: {
		Message: "transaction already exists", HTTPStatus: http.StatusConflict,
		Category: CategoryBusiness, Severity: SeverityError,
		Suggestions: []string{"Check if this transaction was already recorded"},
	},

	// ── Integration ───────────────────────────────────────────────────
	CodeThirdPartyAPIFailure: {
		Message: "external service call failed", HTTPStatus: http.StatusBadGateway,
		Category: CategoryIntegration, Severity: SeverityError,
		// Third-party/network failures (M-Pesa Daraja, KRA eTIMS, Africa's
		// Talking) are retryable by default — this is the fix for the bug
		// where the old package set a "retryable" DETAIL key that nothing
		// ever actually read; here it's a first-class field wired straight
		// into IsRetryable().
		Retryable:   true,
		Suggestions: []string{"Please try again later"},
	},

	// ── Infra / Repository ────────────────────────────────────────────
	// Deliberately generic messages: these must never leak raw database
	// detail to an API client (see ToProblem in http.go, which special-
	// cases CategoryRepository to always show a generic message).
	CodeDatabaseError: {
		Message: "a database error occurred", HTTPStatus: http.StatusInternalServerError,
		Category: CategoryRepository, Severity: SeverityCritical,
	},
	CodeConnectionFailed: {
		Message: "a connection error occurred", HTTPStatus: http.StatusServiceUnavailable,
		Category: CategoryRepository, Severity: SeverityCritical, Retryable: true,
	},
	CodeQueryFailed: {
		Message: "a database error occurred", HTTPStatus: http.StatusInternalServerError,
		Category: CategoryRepository, Severity: SeverityError,
	},
	CodeTransactionFailed: {
		Message: "a database error occurred", HTTPStatus: http.StatusInternalServerError,
		Category: CategoryRepository, Severity: SeverityError,
	},

	// ── System ────────────────────────────────────────────────────────
	CodeInternalPanic: {
		Message: "an internal error occurred", HTTPStatus: http.StatusInternalServerError,
		Category: CategorySystem, Severity: SeverityCritical,
	},
	CodeConfigError: {
		Message: "invalid configuration", HTTPStatus: http.StatusInternalServerError,
		Category: CategorySystem, Severity: SeverityCritical,
	},
	CodeStartupError: {
		Message: "startup failed", HTTPStatus: http.StatusInternalServerError,
		Category: CategorySystem, Severity: SeverityCritical,
	},
	CodeShutdownError: {
		Message: "shutdown failed", HTTPStatus: http.StatusInternalServerError,
		Category: CategorySystem, Severity: SeverityWarning,
	},
}

// RegisterCode lets a module add or override a registry entry at init
// time, WITHOUT needing to modify this file. This is how independently
// developed modules (Forecourt, Payroll, Notifications, Excise) plug
// their own error codes into the shared New()/Newf() machinery:
//
//	// in package payroll, file errors.go
//	const CodePAYEBracketNotFound errors.Code = "payroll.paye.bracket_not_found"
//
//	func init() {
//	    errors.RegisterCode(CodePAYEBracketNotFound, errors.ErrorDef{
//	        Message:    "PAYE bracket not found for the given amount",
//	        HTTPStatus: http.StatusInternalServerError,
//	        Category:   errors.CategoryBusiness,
//	        Severity:   errors.SeverityCritical,
//	    })
//	}
//
// After that, payroll code can simply call:
//
//	errors.New(payroll.CodePAYEBracketNotFound).WithDetail("gross_pay", amount)
//
// and every shared helper (ToProblem, IsRetryable, IsCategory, logging
// middleware, ...) works on it exactly like a built-in code.
type ErrorDef = errorDef

// RegisterCode adds or overwrites a single registry entry. Intended to be
// called from a module's package init() function, before any requests are
// served. It is NOT safe to call concurrently with New()/Newf() calls —
// register all module codes during startup, not at request time.
func RegisterCode(code Code, def ErrorDef) {
	registry[code] = def
}
