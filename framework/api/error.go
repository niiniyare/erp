package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"awo.so/framework/persistence"
	"awo.so/framework/validate"
)

// ErrorHandler is a Fiber error handler that maps errors to the standard
// ErrorResponse envelope. Register it as fiber.Config.ErrorHandler.
//
// Mapping:
//   - *fiber.Error        → HTTP status from error, code from status text
//   - ErrNotFound         → 404 not_found
//   - ErrConflict         → 409 conflict
//   - errForbidden        → 403 permission_denied
//   - ValidationErrors    → 422 validation_error  (with field map)
//   - everything else     → 500 internal_error (message never exposed)
func ErrorHandler(c *fiber.Ctx, err error) error {
	code, detail := mapError(c, err)
	return c.Status(code).JSON(ErrorResponse{Err: detail})
}

// mapError converts err into an HTTP status code and ErrorDetail.
// Side-effect: logs 5xx errors with structured context.
func mapError(c *fiber.Ctx, err error) (int, ErrorDetail) {
	// fiber.Error carries an explicit status + message (e.g. from fiber.ErrNotFound).
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return fe.Code, ErrorDetail{
			Code:    httpStatusCode(fe.Code),
			Message: fe.Message,
		}
	}

	// Domain sentinel errors from the persistence layer.
	if errors.Is(err, persistence.ErrNotFound) {
		return fiber.StatusNotFound, ErrorDetail{
			Code:    "not_found",
			Message: "Record not found.",
		}
	}
	if errors.Is(err, persistence.ErrConflict) {
		return fiber.StatusConflict, ErrorDetail{
			Code:    "conflict",
			Message: "Record already exists.",
		}
	}
	if errors.Is(err, persistence.ErrInvalidTenant) {
		return fiber.StatusUnauthorized, ErrorDetail{
			Code:    "invalid_tenant",
			Message: "Tenant not found or inactive.",
		}
	}
	if errors.Is(err, errForbidden) {
		return fiber.StatusForbidden, ErrorDetail{
			Code:    "permission_denied",
			Message: "Access denied.",
		}
	}

	// Validation errors carry a field → message map.
	var verrs validate.ValidationErrors
	if errors.As(err, &verrs) {
		fields := make(map[string]string, len(verrs))
		for _, ve := range verrs {
			fields[ve.Field] = ve.Message
		}
		return fiber.StatusUnprocessableEntity, ErrorDetail{
			Code:    "validation_error",
			Message: "One or more fields failed validation.",
			Fields:  fields,
		}
	}

	// Unhandled — log with request context, never expose internals.
	log.Error().
		Err(err).
		Str("method", c.Method()).
		Str("path", c.Path()).
		Str("request_id", c.GetRespHeader("X-Request-ID")).
		Msg("unhandled error")

	return fiber.StatusInternalServerError, ErrorDetail{
		Code:    "internal_error",
		Message: "An unexpected error occurred.",
	}
}

// httpStatusCode maps an HTTP status integer to a snake_case error code string.
func httpStatusCode(status int) string {
	switch status {
	case fiber.StatusBadRequest:
		return "bad_request"
	case fiber.StatusUnauthorized:
		return "unauthorized"
	case fiber.StatusForbidden:
		return "permission_denied"
	case fiber.StatusNotFound:
		return "not_found"
	case fiber.StatusMethodNotAllowed:
		return "method_not_allowed"
	case fiber.StatusConflict:
		return "conflict"
	case fiber.StatusUnprocessableEntity:
		return "validation_error"
	case fiber.StatusTooManyRequests:
		return "rate_limited"
	case fiber.StatusServiceUnavailable:
		return "service_unavailable"
	default:
		return "internal_error"
	}
}
