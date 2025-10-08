package handlers

import (
	"fmt"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/niiniyare/erp/internal/shared/logger"
)

// APIResponse represents the standard API response structure
type APIResponse struct {
	Data     any       `json:"data,omitempty"`
	Error    *APIError `json:"error,omitempty"`
	Metadata *Metadata `json:"metadata"`
}

// PaginatedResponse extends APIResponse with pagination information
type PaginatedResponse struct {
	Data       any         `json:"data"`
	Pagination *Pagination `json:"pagination"`
	Metadata   *Metadata   `json:"metadata"`
}

// APIError represents error information in API responses
type APIError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// Metadata contains response metadata
type Metadata struct {
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// Pagination contains pagination information
type Pagination struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

// PaginationParams represents pagination query parameters
type PaginationParams struct {
	Page    int `query:"page" validate:"min=1"`
	PerPage int `query:"per_page" validate:"min=1,max=100"`
}

// DefaultPaginationParams returns default pagination parameters
func DefaultPaginationParams() PaginationParams {
	return PaginationParams{
		Page:    1,
		PerPage: 20,
	}
}

// Validate validates pagination parameters
func (p *PaginationParams) Validate() error {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 20
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
	return nil
}

// GetOffset calculates the offset for database queries
func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PerPage
}

// CalculatePagination calculates pagination information
func CalculatePagination(total int64, page, perPage int) *Pagination {
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	return &Pagination{
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}

// Response helpers

// Success sends a successful API response
func Success(c *fiber.Ctx, data any) error {
	return c.JSON(APIResponse{
		Data:     data,
		Metadata: getMetadata(c),
	})
}

// SuccessWithStatus sends a successful API response with custom status
func SuccessWithStatus(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(APIResponse{
		Data:     data,
		Metadata: getMetadata(c),
	})
}

// SuccessPaginated sends a paginated API response
func SuccessPaginated(c *fiber.Ctx, data any, pagination *Pagination) error {
	return c.JSON(PaginatedResponse{
		Data:       data,
		Pagination: pagination,
		Metadata:   getMetadata(c),
	})
}

// Created sends a 201 Created response
func Created(c *fiber.Ctx, data any) error {
	return c.Status(201).JSON(APIResponse{
		Data:     data,
		Metadata: getMetadata(c),
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(204)
}

// Error response helpers

// BadRequest sends a 400 Bad Request error
func BadRequest(c *fiber.Ctx, message string, details ...map[string]any) error {
	return sendError(c, 400, "invalid_input", message, details...)
}

// Unauthorized sends a 401 Unauthorized error
func Unauthorized(c *fiber.Ctx, message string, details ...map[string]any) error {
	return sendError(c, 401, "unauthorized", message, details...)
}

// Forbidden sends a 403 Forbidden error
func Forbidden(c *fiber.Ctx, message string, details ...map[string]any) error {
	return sendError(c, 403, "forbidden", message, details...)
}

// NotFound sends a 404 Not Found error
func NotFound(c *fiber.Ctx, message string, details ...map[string]any) error {
	return sendError(c, 404, "not_found", message, details...)
}

// Conflict sends a 409 Conflict error
func Conflict(c *fiber.Ctx, message string, details ...map[string]any) error {
	return sendError(c, 409, "conflict", message, details...)
}

// ValidationError sends a 400 validation error
func ValidationError(c *fiber.Ctx, message string, details ...map[string]any) error {
	return sendError(c, 400, "validation_error", message, details...)
}

// RateLimitExceeded sends a 429 rate limit error
func RateLimitExceeded(c *fiber.Ctx, message string, details ...map[string]any) error {
	return sendError(c, 429, "rate_limit_exceeded", message, details...)
}

// InternalError sends a 500 Internal Server Error
func InternalError(c *fiber.Ctx, message string, details ...map[string]any) error {
	return sendError(c, 500, "internal_error", message, details...)
}

// CustomError sends a custom error response
func CustomError(c *fiber.Ctx, status int, code, message string, details ...map[string]any) error {
	return sendError(c, status, code, message, details...)
}

// sendError is a helper function to send error responses
func sendError(c *fiber.Ctx, status int, code, message string, details ...map[string]any) error {
	apiError := &APIError{
		Code:    code,
		Message: message,
	}

	if len(details) > 0 && details[0] != nil {
		apiError.Details = details[0]
	}

	// Log error for debugging (excluding client errors like 400, 401, 403, 404)
	if status >= 500 {
		logger.Error("API error response", logger.Fields{
			"status":     status,
			"code":       code,
			"message":    message,
			"path":       c.Path(),
			"method":     c.Method(),
			"request_id": getRequestID(c),
		})
	} else if status >= 400 && status < 500 {
		logger.Debug("Client error response", logger.Fields{
			"status":     status,
			"code":       code,
			"message":    message,
			"path":       c.Path(),
			"method":     c.Method(),
			"request_id": getRequestID(c),
		})
	}

	return c.Status(status).JSON(APIResponse{
		Error:    apiError,
		Metadata: getMetadata(c),
	})
}

// Helper functions

// getMetadata creates metadata for API responses
func getMetadata(c *fiber.Ctx) *Metadata {
	return &Metadata{
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
		Version:   "1.0",
	}
}

// getRequestID extracts request ID from Fiber context
func getRequestID(c *fiber.Ctx) string {
	if requestID := c.Locals("requestid"); requestID != nil {
		return requestID.(string)
	}
	return c.Get("X-Request-ID", "unknown")
}

// ParsePagination parses pagination parameters from query string
func ParsePagination(c *fiber.Ctx) PaginationParams {
	params := DefaultPaginationParams()

	if page := c.QueryInt("page", 1); page > 0 {
		params.Page = page
	}

	if perPage := c.QueryInt("per_page", 20); perPage > 0 {
		params.PerPage = perPage
		if params.PerPage > 100 {
			params.PerPage = 100
		}
	}

	return params
}

// ValidationErrorDetails creates validation error details
func ValidationErrorDetails(field, tag, value string) map[string]any {
	return map[string]any{
		"field":   field,
		"tag":     tag,
		"value":   value,
		"message": fmt.Sprintf("Field '%s' failed validation for tag '%s'", field, tag),
	}
}

// FieldError represents a field validation error
type FieldError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// MultipleValidationErrors creates error details for multiple validation failures
func MultipleValidationErrors(errors []FieldError) map[string]any {
	return map[string]any{
		"validation_errors": errors,
		"error_count":       len(errors),
	}
}

// HeaderBasedResponse determines response type and sends appropriate response
func HeaderBasedResponse(c *fiber.Ctx, data any, htmlTemplate string) error {
	responseType := c.Locals("responseType")

	// Check Accept header as well
	accept := c.Get("Accept")
	if accept == "application/json" || responseType == "json" {
		return Success(c, data)
	}

	// For HTML responses
	if htmlTemplate != "" {
		return c.Render(htmlTemplate, data)
	}

	// Fallback to JSON
	return Success(c, data)
}

// RedirectResponse handles redirects based on request type
func RedirectResponse(c *fiber.Ctx, url string, status ...int) error {
	responseType := c.Locals("responseType")

	if responseType == "json" {
		redirectStatus := 302
		if len(status) > 0 {
			redirectStatus = status[0]
		}
		return c.Status(redirectStatus).JSON(fiber.Map{
			"redirect_url": url,
			"status":       redirectStatus,
		})
	}

	// For HTML responses
	return c.Redirect(url, status...)
}
