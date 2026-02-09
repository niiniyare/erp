package tenant

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
)

// List handles listing organizations with pagination, filtering, and sorting.
// @Summary List Organizations
// @Description Retrieves a paginated list of organizations, with support for filtering and sorting.
// @Tags Organizations
// @Produce json
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(20)
// @Param status query string false "Filter by organization status (e.g., ACTIVE, SUSPENDED)"
// @Param search query string false "Search term for organization name or slug"
// @Param sort_by query string false "Field to sort by (e.g., name, created_at)" default(created_at)
// @Success 200 {object} map[string]interface{} "A paginated list of organizations"
// @Failure 500 {object} errors.HTTPError "Internal Server Error"
// @Router /organizations [get]
func (h *TenantHandler) List(c *fiber.Ctx) error {
	ctx, span := h.tracer.StartSpan(c.Context(), "tenant.handler.List")
	defer span.End()

	filter := h.extractListParams(c)

	tenants, total, err := h.service.ListTenants(ctx, filter)
	if err != nil {
		return h.handleError(c, err)
	}

	// Convert domain objects to summary response views
	data := toListResponse(tenants)

	// Construct rich pagination metadata as per the API design
	meta := fiber.Map{
		"pagination": fiber.Map{
			"offset":        filter.Offset,
			"limit":         filter.Limit,
			"count":         len(data),
			"total_records": total,
		},
		"filters": fiber.Map{
			"applied": filter,
		},
		"sorting": fiber.Map{
			"by": filter.SortBy,
		},
	}

	span.SetAttributes(
		attribute.Int("pagination.offset", int(filter.Offset)),
		attribute.Int("pagination.limit", int(filter.Limit)),
		attribute.Int("results.count", len(data)),
		attribute.Int64("results.total", total),
	)

	return h.successWithMeta(c, data, meta)
}
