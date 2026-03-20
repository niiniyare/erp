package repository

import (
	"net/http"
	"strings"

	"awo/internal/core/tenant/domain"
	sharedErrors "awo/internal/shared/errors"
)

// parseTenantDBError inspects a raw pgx/DB error and converts it into a meaningful
// domain or business error. This keeps the repo layer clean and gives the handler
// enough context to return a proper HTTP status without exposing DB internals.
func parseTenantDBError(err error, op string) error {
	if err == nil {
		return nil
	}
	msg := err.Error()

	// ── Not found ────────────────────────────────────────────────────────────
	if strings.Contains(msg, "no rows in result set") {
		return domain.ErrTenantNotFound
	}

	// ── Unique constraint violation (SQLSTATE 23505) ─────────────────────────
	if strings.Contains(msg, "23505") || strings.Contains(msg, "unique constraint") {
		if strings.Contains(msg, "subdomain") {
			return domain.ErrSubdomainTaken
		}
		if strings.Contains(msg, "slug") || strings.Contains(msg, "name") {
			return domain.ErrTenantAlreadyExists
		}
		return domain.ErrTenantAlreadyExists
	}

	// ── Check constraint violation (SQLSTATE 23514) ──────────────────────────
	if strings.Contains(msg, "23514") || strings.Contains(msg, "check constraint") {
		if strings.Contains(msg, "company_size") {
			return sharedErrors.NewBusinessError("INVALID_COMPANY_SIZE", "Invalid company size").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(sharedErrors.CategoryValidation).
				WithSuggestion("Valid values: STARTUP, SMALL, MEDIUM, LARGE, ENTERPRISE")
		}
		if strings.Contains(msg, "status") {
			return sharedErrors.NewBusinessError("INVALID_STATUS", "Invalid tenant status").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(sharedErrors.CategoryValidation).
				WithSuggestion("Valid values: PENDING, ACTIVE, SUSPENDED, ARCHIVED")
		}
		// Generic constraint violation — tell the caller which field failed
		return sharedErrors.NewBusinessError("CONSTRAINT_VIOLATION", "A field value failed database validation").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation).
			WithDetail("operation", op).
			WithDetail("raw", msg)
	}

	// ── Not-null violation (SQLSTATE 23502) ──────────────────────────────────
	if strings.Contains(msg, "23502") || strings.Contains(msg, "null value in column") {
		return sharedErrors.NewBusinessError("MISSING_REQUIRED_FIELD", "A required field is missing").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation).
			WithDetail("operation", op)
	}

	// ── Foreign key violation (SQLSTATE 23503) ────────────────────────────────
	if strings.Contains(msg, "23503") || strings.Contains(msg, "foreign key constraint") {
		return sharedErrors.NewBusinessError("INVALID_REFERENCE", "A referenced record does not exist").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation).
			WithDetail("operation", op)
	}

	// ── Generic DB error — wrap as RepositoryError (→ 500, no internals exposed) ──
	return sharedErrors.NewRepositoryError("DB_ERROR", "Database operation failed", err).
		WithOperation(op).
		WithTable("tenants")
}
