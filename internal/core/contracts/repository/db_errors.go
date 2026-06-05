package repository

import (
	"net/http"
	"strings"

	"awo.so/internal/core/contracts/domain"
	sharedErrors "awo.so/internal/shared/errors"
)

// parseContractDBError converts raw pgx/DB errors into domain or business errors.
func parseContractDBError(err error, op string) error {
	if err == nil {
		return nil
	}
	msg := err.Error()

	if strings.Contains(msg, "no rows in result set") {
		return domain.ErrContractNotFound
	}

	if strings.Contains(msg, "23505") || strings.Contains(msg, "unique constraint") {
		if strings.Contains(msg, "number") {
			return domain.ErrContractAlreadyExists
		}
		return domain.ErrContractAlreadyExists
	}

	if strings.Contains(msg, "23514") || strings.Contains(msg, "check constraint") {
		if strings.Contains(msg, "status") {
			return sharedErrors.NewBusinessError("INVALID_CONTRACT_STATUS", "Invalid contract status").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(sharedErrors.CategoryValidation)
		}
		if strings.Contains(msg, "contract_type") {
			return sharedErrors.NewBusinessError("INVALID_CONTRACT_TYPE", "Invalid contract type").
				WithHTTPStatus(http.StatusBadRequest).
				WithCategory(sharedErrors.CategoryValidation)
		}
		return sharedErrors.NewBusinessError("CONSTRAINT_VIOLATION", "A field value failed database validation").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation).
			WithDetail("operation", op)
	}

	if strings.Contains(msg, "23502") || strings.Contains(msg, "null value in column") {
		return sharedErrors.NewBusinessError("MISSING_REQUIRED_FIELD", "A required field is missing").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation).
			WithDetail("operation", op)
	}

	if strings.Contains(msg, "23503") || strings.Contains(msg, "foreign key constraint") {
		return sharedErrors.NewBusinessError("INVALID_REFERENCE", "A referenced record does not exist").
			WithHTTPStatus(http.StatusBadRequest).
			WithCategory(sharedErrors.CategoryValidation).
			WithDetail("operation", op)
	}

	return sharedErrors.NewRepositoryError("DB_ERROR", "Database operation failed", err).
		WithOperation(op).
		WithTable("contracts")
}
