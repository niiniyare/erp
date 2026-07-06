// Package errors is the Awo Framework's shared error package.
//
// It is used by every module in the ERP (IAM, Tenant, Entity/Org hierarchy,
// ABAC, Finance/Accounting, Inventory, Forecourt, Payroll, Notifications,
// and any future module) so that:
//
//   - Every error carries a stable, machine-readable Code (dot-namespaced,
//     e.g. "iam.user.not_found") that can be safely returned to API
//     clients, logged, matched in tests, and translated (sw-KE / af-SO).
//
//   - Every error works with the standard library: errors.Is, errors.As,
//     errors.Unwrap, and fmt.Errorf("%w", ...) all behave exactly as a Go
//     developer would expect, with no custom idioms to learn.
//
//   - Modules can define their OWN lightweight error types (they do not
//     have to import or embed AwoError) and still plug into shared
//     tooling (HTTP mapping, logging, retry logic) by implementing small
//     interfaces such as Coded and HTTPStatuser. This is what makes the
//     package scale across many independently-owned ERP modules instead
//     of becoming a single giant switch statement that everyone has to
//     keep extending.
//
//   - Error VALUES ARE IMMUTABLE. Calling a With* method never mutates the
//     receiver — it returns a new *AwoError. This means package-level
//     sentinel errors (see sentinels.go) are safe to share and enrich
//     concurrently across goroutines and tenants, which matters a lot in
//     a multi-tenant RLS system where a data race on a shared error value
//     could otherwise leak details across tenant boundaries.
package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────
// SEVERITY & CATEGORY
//
// These are deliberately kept separate from Code:
//   - Code identifies WHAT went wrong ("iam.user.not_found") and is stable
//     forever — API clients and tests key off it.
//   - Category groups errors for ROUTING/METRICS/DASHBOARDS ("security",
//     "tenant", "business", ...) and can be shared by many unrelated codes.
//   - Severity says HOW BAD it is, for log-level selection and alerting.
// Keeping these three orthogonal means you never have to encode severity
// or category information inside the Code string itself.
// ─────────────────────────────────────────────────────────────────────────

// Severity is a syslog-style severity level, used to pick a log level and
// to decide whether an on-call alert should fire.
type Severity string

const (
	SeverityInfo     Severity = "info"     // Expected/benign, e.g. "not found"
	SeverityWarning  Severity = "warning"  // Notable but not urgent, e.g. account locked
	SeverityError    Severity = "error"    // Something failed, needs attention
	SeverityCritical Severity = "critical" // Data integrity / system-down class issue
)

// Category groups errors for dashboards, metrics, and coarse-grained
// handling (e.g. "retry every Integration error", "page on-call for every
// Critical System error"). Category is independent of which module raised
// the error — Finance, HR, and Forecourt can all raise CategorySecurity
// errors, for example.
type Category string

const (
	CategoryValidation  Category = "validation"  // Bad input from the caller
	CategoryRepository  Category = "repository"  // Persistence-layer failure
	CategoryBusiness    Category = "business"    // Domain rule violation
	CategorySecurity    Category = "security"    // AuthN/AuthZ/IAM related
	CategoryIntegration Category = "integration" // Third-party API / external system
	CategorySystem      Category = "system"      // Infra/framework internal failure
	CategoryTenant      Category = "tenant"      // Multi-tenancy specific
)

// ─────────────────────────────────────────────────────────────────────────
// STANDARD-LIBRARY-COMPATIBLE INTERFACES
//
// Any type — not just *AwoError — can implement these. This is the key
// mechanism that lets dozens of independently developed ERP modules share
// one HTTP mapper, one logger, one retry-checker, without importing a
// concrete struct from this package. A module can define:
//
//	type PumpOfflineError struct{ PumpID string }
//	func (e *PumpOfflineError) Error() string { return "pump offline: " + e.PumpID }
//	func (e *PumpOfflineError) ErrorCode() errors.Code { return "forecourt.pump.offline" }
//	func (e *PumpOfflineError) HTTPStatus() int { return http.StatusServiceUnavailable }
//
// ...and it will automatically work with errors.Is, ToProblem, IsRetryable,
// logging middleware, etc. — with zero changes to this package.
// ─────────────────────────────────────────────────────────────────────────

// Coded is implemented by any error that carries a stable machine-readable
// Code. This is the minimum contract the framework relies on.
type Coded interface {
	ErrorCode() Code
}

// Categorized is implemented by any error that can report its Category
// for routing/metrics purposes.
type Categorized interface {
	ErrorCategory() Category
}

// HTTPStatuser is implemented by any error that knows which HTTP status
// code it should map to when serialized as an API response.
type HTTPStatuser interface {
	HTTPStatus() int
}

// RetryableErr is implemented by any error that can say whether the
// operation that produced it is safe to retry (e.g. a transient network
// failure calling M-Pesa/Daraja vs. a permanent validation failure).
type RetryableErr interface {
	IsRetryable() bool
}

// Detailer is implemented by any error that carries structured,
// machine-readable context (already sanitized of secrets) for logging or
// API responses.
type Detailer interface {
	ErrorDetails() map[string]any
}

// ─────────────────────────────────────────────────────────────────────────
// AwoError — the framework's default, general-purpose structured error.
//
// Modules are free to define their own error types for highly specific
// cases (see the PumpOfflineError example above), but for the vast
// majority of "not found / already exists / invalid / forbidden" style
// errors across every ERP module, AwoError is the one type you need.
// ─────────────────────────────────────────────────────────────────────────

// AwoError is the framework's canonical structured error value.
//
// IMPORTANT: AwoError is immutable. Every With* method below returns a
// NEW *AwoError built via clone() — it never modifies the receiver. This
// is what makes it safe to keep package-level "sentinel" errors (see
// sentinels.go, e.g. ErrUserNotFound) and enrich them per-call-site
// without any risk of one goroutine's `.WithDetail(...)` leaking into
// another goroutine's (or another tenant's) copy of "the same" error.
type AwoError struct {
	code        Code           // stable machine-readable identifier, e.g. "iam.user.not_found"
	message     string         // human-readable message (safe to show to end users)
	httpStatus  int            // HTTP status this error should map to
	severity    Severity       // log-level hint
	category    Category       // routing/metrics grouping
	retryable   bool           // whether the caller may safely retry
	details     map[string]any // structured, sanitized context (e.g. {"user_id": "..."})
	suggestions []string       // human-readable "what to do next" hints
	tenantID    string         // tenant context, if known, for multi-tenant tracing
	userID      string         // user context, if known
	cause       error          // wrapped underlying error, if any (supports errors.Unwrap)
}

// New creates an AwoError from a registered Code, filling in the default
// message, HTTP status, category, severity, and suggestions from the
// central registry (see registry.go). This is the primary, everyday
// constructor:
//
//	return nil, errors.New(errors.CodeUserNotFound)
//
// If the code has no registry entry (e.g. a typo, or a module that hasn't
// registered its codes yet), New falls back to a safe generic 500-class
// error rather than panicking — this package must never itself be the
// cause of a production incident due to a missing map entry.
func New(code Code) *AwoError {
	def, ok := registry[code]
	if !ok {
		def = errorDef{
			Message:    "an unexpected error occurred",
			HTTPStatus: http.StatusInternalServerError,
			Category:   CategorySystem,
			Severity:   SeverityError,
		}
	}
	return &AwoError{
		code:       code,
		message:    def.Message,
		httpStatus: def.HTTPStatus,
		category:   def.Category,
		severity:   def.Severity,
		retryable:  def.Retryable,
		// Copy the suggestions slice so that callers mutating their own
		// instance (e.g. via WithSuggestion) can never corrupt the shared
		// registry's slice backing array.
		suggestions: append([]string(nil), def.Suggestions...),
	}
}

// Newf behaves like New but overrides the display message with a
// formatted string. The registered Code, HTTP status, category, etc. are
// unchanged — only the human-readable text differs for this instance.
//
//	errors.Newf(errors.CodeInsufficientBalance,
//	    "account %s needs %.2f KES but only has %.2f KES", acctID, need, have)
func Newf(code Code, format string, args ...any) *AwoError {
	return New(code).WithMessage(fmt.Sprintf(format, args...))
}

// Wrap creates a new AwoError for the given code and attaches cause as
// the wrapped underlying error, so errors.Unwrap/errors.Is/errors.As can
// still reach it. Use this whenever you're translating a lower-level
// error (a pgx error, an HTTP client error, a Temporal activity error)
// into a domain-meaningful AwoError without losing the original cause.
//
//	if err != nil {
//	    return errors.Wrap(errors.CodeDatabaseError, err)
//	}
func Wrap(code Code, cause error) *AwoError {
	return New(code).WithCause(cause)
}

// clone returns a shallow copy of e with its own independent details map
// and suggestions slice, so that mutating the copy can never affect e.
// Every With* method below is built on top of clone — this single
// function is what guarantees the whole type's immutability contract.
func (e *AwoError) clone() *AwoError {
	c := *e // copy all value fields (code, message, httpStatus, ...)

	if e.details != nil {
		c.details = make(map[string]any, len(e.details))
		for k, v := range e.details {
			c.details[k] = v
		}
	}
	c.suggestions = append([]string(nil), e.suggestions...)

	return &c
}

// ─────────────────────────────────────────────────────────────────────────
// IMMUTABLE FLUENT BUILDERS
//
// Every method here follows the same pattern: clone, mutate the copy,
// return the copy. Chaining works exactly like the old mutable API did
// (`New(code).WithDetail(...).WithSuggestion(...)`), but is now free of
// the shared-mutable-sentinel hazard.
// ─────────────────────────────────────────────────────────────────────────

// WithDetail attaches a single structured detail (e.g. "user_id", "tenant_id").
// Values are passed through sanitizeValue so that anything that looks like
// a secret (password, token, api key, ...) is redacted before it can ever
// reach a log line or an API response.
func (e *AwoError) WithDetail(key string, value any) *AwoError {
	c := e.clone()
	if c.details == nil {
		c.details = make(map[string]any)
	}
	c.details[key] = sanitizeDetail(key, value)
	return c
}

// WithDetails attaches several structured details at once. Prefer this
// over multiple WithDetail calls when you already have a map — it avoids
// re-cloning the error once per key.
func (e *AwoError) WithDetails(kv map[string]any) *AwoError {
	c := e.clone()
	if c.details == nil {
		c.details = make(map[string]any, len(kv))
	}
	for k, v := range kv {
		c.details[k] = sanitizeDetail(k, v)
	}
	return c
}

// WithMessage overrides the human-readable message for this instance only
// (the registry default for the Code is untouched).
func (e *AwoError) WithMessage(msg string) *AwoError {
	c := e.clone()
	c.message = msg
	return c
}

// WithCause attaches an underlying error so that errors.Unwrap(e) reaches
// it, enabling errors.Is/errors.As to see through AwoError to whatever
// caused it (a pgx.ErrNoRows, a context.DeadlineExceeded, etc.).
func (e *AwoError) WithCause(cause error) *AwoError {
	c := e.clone()
	c.cause = cause
	return c
}

// WithTenant records which tenant this error occurred for. Always set
// this at the point an error crosses a tenant-context boundary so that
// logs and traces stay tenant-attributable, matching the RLS model.
func (e *AwoError) WithTenant(tenantID string) *AwoError {
	c := e.clone()
	c.tenantID = tenantID
	return c
}

// WithUser records which user was involved when the error occurred.
func (e *AwoError) WithUser(userID string) *AwoError {
	c := e.clone()
	c.userID = userID
	return c
}

// WithSuggestion appends one more "what to do next" hint aimed at the end
// user or API consumer (e.g. "Use the 'Forgot Password' option").
func (e *AwoError) WithSuggestion(s string) *AwoError {
	c := e.clone()
	c.suggestions = append(c.suggestions, s)
	return c
}

// WithRetryable overrides whether this specific instance should be
// treated as retryable, regardless of the registry default. Useful for
// cases like a third-party API failure where retryability actually
// depends on the specific HTTP status returned by the third party.
func (e *AwoError) WithRetryable(retryable bool) *AwoError {
	c := e.clone()
	c.retryable = retryable
	return c
}

// WithSeverity overrides the log-level hint for this instance.
func (e *AwoError) WithSeverity(s Severity) *AwoError {
	c := e.clone()
	c.severity = s
	return c
}

// WithCategory overrides the routing category for this instance. Rarely
// needed since the registry sets a sensible default per Code, but
// available for edge cases.
func (e *AwoError) WithCategory(cat Category) *AwoError {
	c := e.clone()
	c.category = cat
	return c
}

// WithHTTPStatus overrides the HTTP status this instance maps to.
func (e *AwoError) WithHTTPStatus(status int) *AwoError {
	c := e.clone()
	c.httpStatus = status
	return c
}

// ─────────────────────────────────────────────────────────────────────────
// STANDARD LIBRARY ERROR INTERFACE
// ─────────────────────────────────────────────────────────────────────────

// Error implements the built-in error interface. The format is stable and
// intended for LOGS, not for end users — API responses should be built
// from ToProblem(err), which exposes Code/Message/Details separately.
func (e *AwoError) Error() string {
	var b strings.Builder
	b.WriteByte('[')
	b.WriteString(string(e.code))
	b.WriteByte(']')
	b.WriteByte(' ')
	b.WriteString(e.message)
	if e.tenantID != "" {
		fmt.Fprintf(&b, " (tenant=%s)", e.tenantID)
	}
	if e.cause != nil {
		b.WriteString(": ")
		b.WriteString(e.cause.Error())
	}
	return b.String()
}

// Unwrap exposes the wrapped cause (if any) to the standard library, so
// errors.Is/errors.As/errors.Unwrap all work transparently through an
// AwoError to whatever it wraps — for example:
//
//	err := errors.Wrap(errors.CodeDatabaseError, pgx.ErrNoRows)
//	errors.Is(err, pgx.ErrNoRows) // true
func (e *AwoError) Unwrap() error {
	return e.cause
}

// Is enables errors.Is(err, errors.New(SomeCode)) to compare AwoErrors by
// Code rather than by pointer identity. Because New(code) constructs a
// brand new value every time, pointer equality would never match two
// "logically equal" errors — Is fixes that by comparing the Code field,
// which is the actual identity of an AwoError.
//
//	if errors.Is(err, errors.New(errors.CodeUserNotFound)) { ... }
//
// (Prefer the more concise errors.Is(err, errors.CodeUserNotFound) helper
// in helpers.go for everyday call sites — this method exists so the
// stdlib idiom above also works if that's what a developer reaches for.)
func (e *AwoError) Is(target error) bool {
	t, ok := target.(*AwoError)
	if !ok {
		return false
	}
	return e.code == t.code
}

// ─────────────────────────────────────────────────────────────────────────
// ACCESSOR INTERFACE IMPLEMENTATIONS
//
// These make *AwoError satisfy Coded / Categorized / HTTPStatuser /
// RetryableErr / Detailer, so the generic helpers in helpers.go and the
// HTTP mapping in http.go work identically whether they're handed an
// *AwoError or some completely unrelated module-defined error type that
// also implements these small interfaces.
// ─────────────────────────────────────────────────────────────────────────

func (e *AwoError) ErrorCode() Code              { return e.code }
func (e *AwoError) ErrorCategory() Category      { return e.category }
func (e *AwoError) HTTPStatus() int              { return e.httpStatus }
func (e *AwoError) IsRetryable() bool            { return e.retryable }
func (e *AwoError) ErrorDetails() map[string]any { return e.details }
func (e *AwoError) Suggestions() []string        { return e.suggestions }
func (e *AwoError) GetSeverity() Severity        { return e.severity }
func (e *AwoError) TenantID() string             { return e.tenantID }
func (e *AwoError) UserID() string               { return e.userID }
func (e *AwoError) Message() string              { return e.message }

// MarshalJSON gives AwoError a stable, intentional JSON shape — this
// controls exactly what an AwoError looks like if it's ever serialized
// directly (e.g. accidentally logged as JSON), separate from the
// intentionally-curated shape that Problem (http.go) produces for actual
// API responses.
func (e *AwoError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Code        Code           `json:"code"`
		Message     string         `json:"message"`
		Details     map[string]any `json:"details,omitempty"`
		Suggestions []string       `json:"suggestions,omitempty"`
		Severity    Severity       `json:"severity"`
		Category    Category       `json:"category"`
		Retryable   bool           `json:"retryable"`
	}{
		Code:        e.code,
		Message:     e.message,
		Details:     e.details,
		Suggestions: e.suggestions,
		Severity:    e.severity,
		Category:    e.category,
		Retryable:   e.retryable,
	})
}

// ─────────────────────────────────────────────────────────────────────────
// SANITIZATION
//
// Any detail value passed through WithDetail/WithDetails/AddWithValue is
// routed through here first. This is the ONE place in the package that
// decides what is safe to keep in logs/API responses — so it only ever
// needs to be reviewed/audited in one spot as new secret-like field names
// come up across new ERP modules (e.g. "mpesa_secret", "eTIMS_token").
// ─────────────────────────────────────────────────────────────────────────

// sensitiveSubstrings lists lowercase substrings that, if found anywhere
// in a string value, cause the whole value to be redacted rather than
// partially masked — a partial mask on secret-shaped data is not safe.
var sensitiveSubstrings = []string{
	"password", "passwd", "secret", "token", "apikey", "api_key",
	"private_key", "authorization", "bearer",
}

// sanitizeValue redacts values that look like secrets and truncates
// long strings so that oversized payloads (e.g. an entire request body
// accidentally passed as a "detail") don't blow up logs or API payloads.
//
// This checks the VALUE's own content — it catches cases where a secret
// ends up embedded inside an otherwise innocuous-looking string (e.g. a
// URL with a token query param, or a raw error message that happens to
// quote a password). It is intentionally kept separate from
// sanitizeDetail below, which ALSO considers the key name — the two
// checks catch different mistakes and are both needed:
//   - sanitizeValue: "the value itself looks secret-shaped"
//   - sanitizeDetail: "the key name says this value IS a secret, even if
//     the opaque value itself doesn't look like anything in particular"
func sanitizeValue(value any) any {
	s, ok := value.(string)
	if !ok {
		// Non-string values (ints, structs, etc.) are assumed safe. If a
		// module ever needs to pass a struct that might itself contain
		// secrets, it should sanitize before calling WithDetail.
		return value
	}

	lower := strings.ToLower(s)
	for _, needle := range sensitiveSubstrings {
		if strings.Contains(lower, needle) {
			return "[REDACTED]"
		}
	}

	const maxLen = 100
	if len(s) > maxLen {
		return s[:maxLen] + "...(truncated)"
	}
	return s
}

// sanitizeDetail is the entry point used by WithDetail/WithDetails. It
// redacts unconditionally whenever the KEY name itself indicates the
// value is sensitive (e.g. "daraja_secret", "mpesa_token",
// "eTIMS_api_key") — this matters because a secret value is often an
// opaque token/hash that doesn't textually contain any of the
// sensitiveSubstrings itself, so key-name-based detection is the only
// reliable signal in that case. If the key looks safe, it falls through
// to sanitizeValue's content-based check as a second line of defence.
func sanitizeDetail(key string, value any) any {
	lowerKey := strings.ToLower(key)
	for _, needle := range sensitiveSubstrings {
		if strings.Contains(lowerKey, needle) {
			return "[REDACTED]"
		}
	}
	return sanitizeValue(value)
}
