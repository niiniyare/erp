package errors

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// TestSentinelImmutability proves the core safety claim: enriching a
// shared package-level sentinel with WithDetail never mutates the
// sentinel itself, so concurrent callers can never see each other's
// details.
func TestSentinelImmutability(t *testing.T) {
	before := ErrUserNotFound.ErrorDetails()
	if len(before) != 0 {
		t.Fatalf("expected sentinel to start with no details, got %v", before)
	}

	enriched := ErrUserNotFound.WithDetail("user_id", "u-123")

	if len(ErrUserNotFound.ErrorDetails()) != 0 {
		t.Fatalf("sentinel was mutated! details = %v", ErrUserNotFound.ErrorDetails())
	}
	if enriched.ErrorDetails()["user_id"] != "u-123" {
		t.Fatalf("enriched copy missing expected detail")
	}
}

// TestStdlibErrorsIs proves errors.Is/As see through AwoError to a
// wrapped cause.
func TestStdlibErrorsIs(t *testing.T) {
	cause := fmt.Errorf("connection reset by peer")
	wrapped := Wrap(CodeConnectionFailed, cause)

	if !errors.Is(wrapped, cause) {
		t.Fatalf("expected errors.Is to see through AwoError to its cause")
	}

	var target *AwoError
	if !errors.As(wrapped, &target) {
		t.Fatalf("expected errors.As to extract the *AwoError")
	}
	if target.ErrorCode() != CodeConnectionFailed {
		t.Fatalf("unexpected code: %s", target.ErrorCode())
	}
}

// TestRetryableWiring proves the bug from the old package (retryable set
// as a detail key that nothing read) is fixed: NewThirdPartyAPIFailureError
// is retryable via the real Retryable field/IsRetryable() path.
func TestRetryableWiring(t *testing.T) {
	err := NewThirdPartyAPIFailureError("mpesa-daraja", 502)
	if !IsRetryable(err) {
		t.Fatalf("expected third-party API failure to be retryable")
	}
}

// TestConventionBasedNotFound proves IsNotFound works for ANY code ending
// in ".not_found" without needing to be told about it explicitly —
// including a code this package has never seen before (simulating a
// module-defined code).
func TestConventionBasedNotFound(t *testing.T) {
	const codeFromHypotheticalForecourtModule Code = "forecourt.pump.not_found"
	RegisterCode(codeFromHypotheticalForecourtModule, ErrorDef{
		Message: "pump not found", HTTPStatus: 404, Category: CategoryBusiness, Severity: SeverityError,
	})

	err := New(codeFromHypotheticalForecourtModule)
	if !IsNotFound(err) {
		t.Fatalf("expected a module-defined *.not_found code to be recognised by IsNotFound")
	}
	if IsConflict(err) {
		t.Fatalf("not-found error should not also be a conflict")
	}
}

// TestToProblemHidesRepositoryDetail proves that a repository/infra error
// never leaks its raw message or details into an API response.
func TestToProblemHidesRepositoryDetail(t *testing.T) {
	dbErr := Wrap(CodeDatabaseError, errors.New("pq: duplicate key value violates unique constraint \"users_email_key\"")).
		WithDetail("query", "INSERT INTO users ...")

	p := ToProblem(dbErr)

	if p.Title == dbErr.Error() {
		t.Fatalf("repository error title should be generic, got raw error text")
	}
	if p.Detail != nil {
		t.Fatalf("repository error should not expose details to API clients, got %v", p.Detail)
	}
	if p.Status != 500 {
		t.Fatalf("expected 500 status for database error, got %d", p.Status)
	}
}

// TestToProblemFieldErrors proves FieldErrors serializes as a proper
// 400 validation problem with per-field detail.
func TestToProblemFieldErrors(t *testing.T) {
	var fe FieldErrors
	fe.AddWithCode("email", "email is required", FieldRequired)
	fe.AddWithCode("password", "password too short", FieldTooShort)

	p := ToProblem(fe)
	if p.Status != 400 {
		t.Fatalf("expected 400, got %d", p.Status)
	}
	if len(p.Errors) != 2 {
		t.Fatalf("expected 2 field errors, got %d", len(p.Errors))
	}
}

// TestSanitizeRedactsSecrets proves a detail is redacted whenever EITHER
// the key name (e.g. "daraja_secret") or the value's own content
// (e.g. containing the word "token") indicates it's sensitive — an opaque
// token value under an innocuous key, and a plainly-named secret key with
// an opaque value, must both be caught.
func TestSanitizeRedactsSecrets(t *testing.T) {
	byKeyName := New(CodeThirdPartyAPIFailure).WithDetail("daraja_secret", "sk_live_abc123")
	if byKeyName.ErrorDetails()["daraja_secret"] != "[REDACTED]" {
		t.Fatalf("expected key-name-based redaction, got %v", byKeyName.ErrorDetails()["daraja_secret"])
	}

	byValueContent := New(CodeThirdPartyAPIFailure).WithDetail("note", "using bearer sk_live_abc123")
	if byValueContent.ErrorDetails()["note"] != "[REDACTED]" {
		t.Fatalf("expected value-content-based redaction, got %v", byValueContent.ErrorDetails()["note"])
	}

	safe := New(CodeUserNotFound).WithDetail("user_id", "u-123")
	if safe.ErrorDetails()["user_id"] != "u-123" {
		t.Fatalf("expected non-sensitive detail to pass through unchanged")
	}
}

// TestContextEnrichment proves NewFromContext pulls tenant/user IDs
// automatically.
func TestContextEnrichment(t *testing.T) {
	ctx := context.WithValue(context.Background(), TenantIDKey, "tenant-abc")
	ctx = context.WithValue(ctx, UserIDKey, "user-xyz")

	err := NewFromContext(ctx, CodeForbidden)
	if err.TenantID() != "tenant-abc" {
		t.Fatalf("expected tenant to be attached from context")
	}
	if err.UserID() != "user-xyz" {
		t.Fatalf("expected user to be attached from context")
	}
}
