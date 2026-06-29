package tenant

import (
	"fmt"
	"time"

	coreTenant "awo.so/internal/core/tenant"
	"github.com/gofiber/fiber/v2"
)

// tenantLinks generates the HATEOAS links section for a tenant resource.
func tenantLinks(id string) fiber.Map {
	base := fmt.Sprintf("/api/v1/tenants/%s", id)
	return fiber.Map{
		"self":      base,
		"dashboard": base + "/dashboard",
		"settings":  base + "/settings",
		"members":   base + "/members",
		"activate":  base + "/activate",
		"suspend":   base + "/suspend",
		"archive":   base + "/archive",
	}
}

// toResponse converts a tenant domain object to a structured map for API responses.
// It adapts the output based on the requested 'view' (summary, default, detailed).
func toResponse(t *coreTenant.Tenant, view string) fiber.Map {
	if t == nil {
		return nil
	}

	id := t.ID.String()

	// The 'summary' view is the base for all other views.
	response := fiber.Map{
		"id":     id,
		"object": "tenant",
		"slug":   t.Slug,
		"name":   t.Name,
		"status": string(t.Status),
		"links":  tenantLinks(id),
	}

	if view == "summary" {
		return response
	}

	// The 'default' view adds essential operational details.
	response["email"] = t.Email
	response["preferences"] = fiber.Map{
		"timezone":      t.Timezone,
		"currency_code": t.CurrencyCode,
	}
	response["timestamps"] = fiber.Map{
		"created_at": t.CreatedAt.Format(time.RFC3339),
		"updated_at": t.UpdatedAt.Format(time.RFC3339),
	}

	if view == "default" {
		return response
	}

	// The 'detailed' view includes everything, reflecting the rich response pattern.
	if t.Subdomain != nil {
		response["subdomain"] = *t.Subdomain
	}
	if t.Industry != nil {
		response["industry"] = *t.Industry
	}
	if t.CompanySize != nil {
		response["company_size"] = *t.CompanySize
	}
	if t.TaxID != nil {
		response["tax_id"] = *t.TaxID
	}
	if t.RegistrationNumber != nil {
		response["registration_number"] = *t.RegistrationNumber
	}
	if t.LegalEntityType != nil {
		response["legal_entity_type"] = *t.LegalEntityType
	}
	if t.Metadata != nil {
		response["metadata"] = t.Metadata
	}
	if t.Settings != nil {
		response["settings"] = t.Settings
	}
	if t.DeletedAt != nil {
		response["timestamps"].(fiber.Map)["deleted_at"] = t.DeletedAt.Format(time.RFC3339)
	}
	if t.LastActivityAt != nil {
		response["timestamps"].(fiber.Map)["last_activity_at"] = t.LastActivityAt.Format(time.RFC3339)
	}

	return response
}

// toActionResponse builds a response for lifecycle action endpoints (activate, suspend, archive).
func toActionResponse(t *coreTenant.Tenant, action string) fiber.Map {
	if t == nil {
		return fiber.Map{"status": action}
	}
	id := t.ID.String()
	return fiber.Map{
		"id":     id,
		"object": "tenant",
		"name":   t.Name,
		"status": string(t.Status),
		"action": action,
		"links":  tenantLinks(id),
	}
}

// toListResponse converts a slice of tenant domain objects to a slice of default views.
func toListResponse(tenants []*coreTenant.Tenant) []fiber.Map {
	list := make([]fiber.Map, len(tenants))
	for i, t := range tenants {
		list[i] = toResponse(t, "default")
	}
	return list
}
