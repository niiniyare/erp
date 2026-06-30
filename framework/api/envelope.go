package api

import (
	"math"

	"github.com/gofiber/fiber/v2"
)

// ── Success envelopes ──────────────────────────────────────────────────────────

// Response is the standard success envelope for single-record responses.
//
//	{"data": {...}, "meta": {...}}
type Response struct {
	Data any `json:"data"`
	Meta any `json:"meta,omitempty"`
}

// ListResponse is the standard success envelope for paginated list responses.
//
//	{"data": [...], "meta": {"total": n, "page": p, "per_page": l, "pages": n}}
type ListResponse struct {
	Data []map[string]any `json:"data"`
	Meta ListMeta         `json:"meta"`
}

// ListMeta carries pagination metadata in list responses.
type ListMeta struct {
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Pages   int   `json:"pages"`
}

// ── Error envelopes ────────────────────────────────────────────────────────────

// ErrorResponse is the standard error envelope.
//
//	{"error": {"code": "not_found", "message": "...", "fields": {...}}}
type ErrorResponse struct {
	Err ErrorDetail `json:"error"`
}

// ErrorDetail carries machine-readable and human-readable error information.
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// ── Constructors ───────────────────────────────────────────────────────────────

// OK sends a 200 response with the data wrapped in the standard envelope.
func OK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(Response{Data: data})
}

// Created sends a 201 response with data in the standard envelope.
func Created(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(Response{Data: data})
}

// Accepted sends a 202 response with data and an optional meta payload.
// Used for workflow-triggered operations where processing is asynchronous.
func Accepted(c *fiber.Ctx, data any, meta any) error {
	return c.Status(fiber.StatusAccepted).JSON(Response{Data: data, Meta: meta})
}

// List sends a paginated 200 response.
// total is the total record count; limit and offset come from the query.
func List(c *fiber.Ctx, rows []map[string]any, total int64, limit, offset int) error {
	if limit <= 0 {
		limit = 50
	}
	page := offset/limit + 1
	pages := int(math.Ceil(float64(total) / float64(limit)))
	if pages < 1 {
		pages = 1
	}
	return c.Status(fiber.StatusOK).JSON(ListResponse{
		Data: rows,
		Meta: ListMeta{
			Total:   total,
			Page:    page,
			PerPage: limit,
			Pages:   pages,
		},
	})
}

// Err sends an error response using the standard error envelope.
func Err(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(ErrorResponse{
		Err: ErrorDetail{Code: code, Message: message},
	})
}

// ValidationErr sends a 422 response with field-level error details.
func ValidationErr(c *fiber.Ctx, fields map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(ErrorResponse{
		Err: ErrorDetail{
			Code:    "validation_error",
			Message: "One or more fields failed validation.",
			Fields:  fields,
		},
	})
}
