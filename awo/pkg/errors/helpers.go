package errors

import (
	"errors"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────
// GENERIC ERROR-MATCHING HELPERS
//
// The previous version of this package had one hand-written function per
// error code — IsUserNotFound, IsEntityNotFound, IsRoleNotFound, ... —
// which meant every new module error required a new copy-pasted function
// here. That doesn't scale to "huge framework, many ERP modules".
//
// Instead, everything below is generic: it dispatches on the Coded /
// Categorized / RetryableErr interfaces, so it works identically for
// *AwoError values AND for any module's own hand-written error types,
// without this file ever needing to change again.
// ─────────────────────────────────────────────────────────────────────────

// Is reports whether err's Code (found anywhere in its Unwrap chain)
// equals code. This is the direct replacement for the old
// IsUserNotFound(err)-style, one-function-per-code helpers:
//
//	if errors.Is(err, errors.CodeUserNotFound) { ... }
//
// Note this is Awo's own errors.Is, distinct from (but consistent with)
// the standard library's errors.Is(err, target error) — it exists because
// comparing by Code is almost always what you want for an AwoError
// (whose values are never singletons, since New() always builds a fresh
// instance), whereas stdlib errors.Is compares by identity/Is-method.
func Is(err error, code Code) bool {
	var coded Coded
	if errors.As(err, &coded) {
		return coded.ErrorCode() == code
	}
	return false
}

// IsAny reports whether err's Code matches any of the given codes. Useful
// for "is this any kind of conflict" checks without maintaining a
// separate named function per grouping.
func IsAny(err error, codes ...Code) bool {
	var coded Coded
	if !errors.As(err, &coded) {
		return false
	}
	actual := coded.ErrorCode()
	for _, c := range codes {
		if actual == c {
			return true
		}
	}
	return false
}

// IsCategory reports whether err belongs to the given Category. Replaces
// the old IsSecurityError/IsTenantError/... one-per-category functions:
//
//	if errors.IsCategory(err, errors.CategorySecurity) { ... }
func IsCategory(err error, category Category) bool {
	var c Categorized
	if errors.As(err, &c) {
		return c.ErrorCategory() == category
	}
	return false
}

// IsRetryable reports whether the operation that produced err may safely
// be retried. Replaces IsTemporary/IsTemporaryError.
//
// Unlike the old implementation, this does NOT fall back to sniffing the
// error message for substrings like "timeout" or "rate limit" — that
// approach is fragile (an unrelated error whose text happens to contain
// "rate limit" would be misclassified) and dangerous when it drives real
// retry logic against Postgres/Temporal/M-Pesa. If a lower-level error
// needs to be retryable, wrap it explicitly:
//
//	return errors.Wrap(errors.CodeConnectionFailed, err) // Retryable: true via registry
//
// or check well-known stdlib/driver sentinels (context.DeadlineExceeded,
// a net.Error with Timeout() == true, etc.) at the call site before
// deciding to retry — don't rely on string matching here.
func IsRetryable(err error) bool {
	var r RetryableErr
	if errors.As(err, &r) {
		return r.IsRetryable()
	}
	return false
}

// IsNotFound reports whether err represents a "not found" condition, by
// checking whether its Code ends in ".not_found". This convention-based
// check means every new module's "X not found" errors are automatically
// recognised here as long as the module follows the naming convention
// documented in codes.go — no enumerated list to maintain.
func IsNotFound(err error) bool {
	return codeHasSuffix(err, ".not_found")
}

// IsConflict reports whether err represents an "already exists" /
// conflict condition, by checking whether its Code ends in ".exists".
func IsConflict(err error) bool {
	return codeHasSuffix(err, ".exists")
}

// IsExpired reports whether err represents an expiry condition (expired
// token, expired invitation, expired policy, ...), by Code suffix.
func IsExpired(err error) bool {
	return codeHasSuffix(err, ".expired")
}

func codeHasSuffix(err error, suffix string) bool {
	var coded Coded
	if !errors.As(err, &coded) {
		return false
	}
	return strings.HasSuffix(string(coded.ErrorCode()), suffix)
}

// IsAuthenticationError reports whether err is one of the well-known
// authentication/authorization failure conditions. Kept as an explicit
// list (rather than a suffix convention) because "unauthorized" and
// "forbidden" don't share a common code suffix the way not_found/exists/
// expired do.
func IsAuthenticationError(err error) bool {
	return IsAny(err,
		CodeInvalidCredentials,
		CodeAuthenticationFail,
		CodeUnauthorized,
		CodeForbidden,
		CodeAccountLocked,
	)
}

// GetCode extracts the Code from err, or CodeUnknown if err doesn't carry
// one. Prefer this over a type assertion when logging or building metrics
// labels from arbitrary errors.
func GetCode(err error) Code {
	var coded Coded
	if errors.As(err, &coded) {
		return coded.ErrorCode()
	}
	return CodeUnknown
}

// GetHTTPStatus extracts the HTTP status err should map to, defaulting to
// 500 for anything that doesn't implement HTTPStatuser. Most call sites
// should prefer ToProblem(err).Status, which also handles validation and
// repository-error special-casing — this is exposed for callers that only
// need the status code in isolation.
func GetHTTPStatus(err error) int {
	var s HTTPStatuser
	if errors.As(err, &s) {
		return s.HTTPStatus()
	}
	return 500
}

// GetDetails extracts structured details from err, or nil if it doesn't
// carry any.
func GetDetails(err error) map[string]any {
	var d Detailer
	if errors.As(err, &d) {
		return d.ErrorDetails()
	}
	return nil
}
