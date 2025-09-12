package handlers

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"goa.design/goa/v3/pkg"

	"github.com/niiniyare/erp/internal/api/gen/organization"
	"github.com/niiniyare/erp/internal/core/entity"
	sharedErrors "github.com/niiniyare/erp/internal/shared/errors"
	"github.com/niiniyare/erp/internal/shared/logger"
	"github.com/niiniyare/erp/internal/shared/metrics"
	"github.com/niiniyare/erp/internal/shared/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// OrganizationGoaHandler implements the GOA organization service following the data flow pattern
type OrganizationGoaHandler struct {
	entityService entity.Service
	tracing       tracing.TracingService
	metrics       *metrics.MetricsService
}

// NewOrganizationGoaHandler creates a new GOA organization handler following Clean Architecture pattern
func NewOrganizationGoaHandler(entitySvc entity.Service, tracing tracing.TracingService, metrics *metrics.MetricsService) organization.Service {
	return &OrganizationGoaHandler{
		entityService: entitySvc,
		tracing:       tracing,
		metrics:       metrics,
	}
}

// Create creates a new organization following the data flow pattern
func (h *OrganizationGoaHandler) Create(ctx context.Context, p *organization.CreateOrganizationPayload) (*organization.Organization, string, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "organization.create",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("organization.name", p.Name),
			attribute.String("entity.type", string(p.EntityType)),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("organization_create_duration", metrics.Fields{
		"operation": "create",
	})
	defer timer.Stop()

	// Convert GOA payload to entity domain request
	req := entity.CreateEntityRequest{
		Name:          p.Name,
		Code:          generateEntityCode(p.Name), // Generate code from name
		Type:          entity.EntityType(p.EntityType),
		IsActive:      true,
		IsHidden:      false,
		AccrualMethod: true, // Default to accrual accounting
		FYStartMonth:  1,    // Default fiscal year starts in January
		Address:       map[string]any{},
		Settings:      map[string]any{},
		Metadata:      map[string]any{},
	}

	// Debug log to check values
	logger.Info("Creating organization with values", logger.Fields{
		"name":           req.Name,
		"code":           req.Code,
		"type":           string(req.Type),
		"fy_start_month": req.FYStartMonth,
		"accrual_method": req.AccrualMethod,
	})

	// Set parent ID if provided
	if p.ParentID != nil {
		if parentUUID, err := uuid.Parse(*p.ParentID); err == nil {
			req.ParentID = &parentUUID
		}
	}

	// Map organization settings if provided
	if p.Settings != nil {
		// Extract fiscal year start from settings if available
		if p.Settings.FiscalYearStart != "" {
			// Parse fiscal year start (format: "MM-DD")
			if len(p.Settings.FiscalYearStart) >= 2 {
				if month := parseFiscalYearMonth(p.Settings.FiscalYearStart); month > 0 {
					req.FYStartMonth = month
				}
			}
		}

		// Add organization settings to entity settings
		if p.Settings.Timezone != "" {
			req.Settings["timezone"] = p.Settings.Timezone
		}
		if p.Settings.Currency != "" {
			req.Settings["currency"] = p.Settings.Currency
		}
		if p.Settings.BusinessHours != nil {
			req.Settings["business_hours"] = p.Settings.BusinessHours
		}
		if p.Settings.Integrations != nil {
			req.Settings["integrations"] = p.Settings.Integrations
		}
		if p.Settings.Preferences != nil {
			req.Settings["preferences"] = p.Settings.Preferences
		}
	}

	// Add additional fields if provided
	if p.Settings != nil || p.Description != nil {
		req.Metadata = make(map[string]any)
		if p.Description != nil {
			req.Metadata["description"] = *p.Description
		}
		if p.LegalName != nil {
			req.Metadata["legal_name"] = *p.LegalName
		}
		if p.LegalEntityType != nil {
			req.Metadata["legal_entity_type"] = *p.LegalEntityType
		}
		if p.Website != nil {
			req.Metadata["website"] = *p.Website
		}
		if p.Industry != nil {
			req.Metadata["industry"] = *p.Industry
		}
	}

	// Create entity through domain service
	entityResult, err := h.entityService.CreateEntity(ctx, req)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("organization_create_errors_total", metrics.Fields{
			"entity_type": string(p.EntityType),
		})
		return nil, "", mapEntityError(err)
	}

	// Convert entity result to GOA organization
	org := &organization.Organization{
		ID:         entityResult.ID.String(),
		Name:       entityResult.Name,
		EntityType: organization.EntityType(entityResult.Type),
		Status:     organization.OrganizationStatus("ACTIVE"),
		CreatedAt:  entityResult.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  entityResult.UpdatedAt.Format(time.RFC3339),
	}

	// Add optional fields from metadata
	if entityResult.Metadata != nil {
		if desc, ok := entityResult.Metadata["description"].(string); ok {
			org.Description = &desc
		}
		if legalName, ok := entityResult.Metadata["legal_name"].(string); ok {
			org.LegalName = &legalName
		}
		if legalType, ok := entityResult.Metadata["legal_entity_type"].(string); ok {
			org.LegalEntityType = &legalType
		}
		if website, ok := entityResult.Metadata["website"].(string); ok {
			org.Website = &website
		}
		if industry, ok := entityResult.Metadata["industry"].(string); ok {
			org.Industry = &industry
		}
	}

	// Set parent ID if exists
	if entityResult.ParentID != nil {
		parentIDStr := entityResult.ParentID.String()
		org.ParentID = &parentIDStr
	}

	h.metrics.IncrementCounter("organization_create_total", metrics.Fields{
		"entity_type": string(p.EntityType),
	})
	span.SetAttributes(attribute.String("result.organization_id", org.ID))

	return org, "default", nil
}

// Get retrieves an organization by ID following the data flow pattern
func (h *OrganizationGoaHandler) Get(ctx context.Context, p *organization.GetPayload) (*organization.Organization, string, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "organization.get",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("organization.id", p.ID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("organization_get_duration", metrics.Fields{
		"operation": "get",
	})
	defer timer.Stop()

	// Parse organization ID
	orgUUID, err := uuid.Parse(p.ID)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("organization_get_errors_total", metrics.Fields{
			"error_type": "invalid_id",
		})
		return nil, "", organization.MakeBadRequest(err)
	}

	// Get entity through domain service
	entityResult, err := h.entityService.GetEntityByID(ctx, orgUUID)
	if err != nil {
		h.tracing.RecordError(ctx, err, tracing.WithErrorStatus())
		h.metrics.IncrementCounter("organization_get_errors_total", metrics.Fields{
			"error_type": "not_found",
		})
		return nil, "", mapEntityError(err)
	}

	// Convert entity to GOA organization
	org := &organization.Organization{
		ID:         entityResult.ID.String(),
		Name:       entityResult.Name,
		EntityType: organization.EntityType(entityResult.Type),
		Status:     organization.OrganizationStatus("ACTIVE"),
		CreatedAt:  entityResult.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  entityResult.UpdatedAt.Format(time.RFC3339),
	}

	// Add optional fields from metadata
	if entityResult.Metadata != nil {
		if desc, ok := entityResult.Metadata["description"].(string); ok {
			org.Description = &desc
		}
		if legalName, ok := entityResult.Metadata["legal_name"].(string); ok {
			org.LegalName = &legalName
		}
		if legalType, ok := entityResult.Metadata["legal_entity_type"].(string); ok {
			org.LegalEntityType = &legalType
		}
		if website, ok := entityResult.Metadata["website"].(string); ok {
			org.Website = &website
		}
		if industry, ok := entityResult.Metadata["industry"].(string); ok {
			org.Industry = &industry
		}
	}

	// Set parent ID if exists
	if entityResult.ParentID != nil {
		parentIDStr := entityResult.ParentID.String()
		org.ParentID = &parentIDStr
	}

	h.metrics.IncrementCounter("organization_get_total", metrics.Fields{})
	return org, "default", nil
}

// List retrieves organizations with pagination following the data flow pattern
func (h *OrganizationGoaHandler) List(ctx context.Context, p *organization.ListPayload) (*organization.ListResult, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "organization.list",
		tracing.WithSpanKind(tracing.SpanKindServer))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("organization_list_duration", metrics.Fields{
		"operation": "list",
	})
	defer timer.Stop()

	// TODO: Integrate with existing entity listing logic using h.entityService.ListEntities
	result := &organization.ListResult{
		Data:       []*organization.Organization{},
		Pagination: &organization.PaginationMeta{CurrentPage: 1, PageSize: 20, TotalItems: 0, TotalPages: 0, HasNext: false, HasPrev: false},
	}

	h.metrics.IncrementCounter("organization_list_total", metrics.Fields{})
	return result, nil
}

// Update updates an existing organization following the data flow pattern
func (h *OrganizationGoaHandler) Update(ctx context.Context, p *organization.UpdateOrganizationPayload) (*organization.Organization, string, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "organization.update",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("organization.id", p.ID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("organization_update_duration", metrics.Fields{
		"operation": "update",
	})
	defer timer.Stop()

	// TODO: Integrate with existing entity update logic using h.entityService.UpdateEntity
	org := &organization.Organization{
		ID:         p.ID,
		Name:       *p.Name,
		EntityType: *p.EntityType,
		Status:     *p.Status,
		CreatedAt:  "2024-01-01T00:00:00Z",
		UpdatedAt:  "2024-01-01T00:00:00Z",
	}

	h.metrics.IncrementCounter("organization_update_total", metrics.Fields{})
	return org, "default", nil
}

// Hierarchy retrieves organization hierarchy following the data flow pattern
func (h *OrganizationGoaHandler) Hierarchy(ctx context.Context, p *organization.HierarchyPayload) (*organization.OrganizationHierarchy, error) {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "organization.hierarchy",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("organization.id", p.ID),
			attribute.Int("hierarchy.depth", int(p.Depth)),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("organization_hierarchy_duration", metrics.Fields{
		"operation": "hierarchy",
	})
	defer timer.Stop()

	// TODO: Integrate with existing entity hierarchy logic using h.entityService.GetEntityWithHierarchy
	result := &organization.OrganizationHierarchy{
		Root:      &organization.OrganizationNode{ID: p.ID, Name: "Root Org", EntityType: "COMPANY", Status: "ACTIVE", Children: []string{}, Level: 0},
		Children:  []*organization.OrganizationNode{},
		Ancestors: []*organization.OrganizationNode{},
		Depth:     p.Depth,
	}

	h.metrics.IncrementCounter("organization_hierarchy_total", metrics.Fields{})
	return result, nil
}

// Archive archives an organization following the data flow pattern
func (h *OrganizationGoaHandler) Archive(ctx context.Context, p *organization.ArchivePayload) error {
	// Start tracing span
	ctx, span := h.tracing.StartSpan(ctx, "organization.archive",
		tracing.WithSpanKind(tracing.SpanKindServer),
		tracing.WithAttributes(
			attribute.String("organization.id", p.ID),
		))
	defer span.End()

	// Start metrics timer
	timer := h.metrics.Timer("organization_archive_duration", metrics.Fields{
		"operation": "archive",
	})
	defer timer.Stop()

	logger.Info("Organization archive called", logger.Fields{
		"id": p.ID,
	})

	// TODO: Integrate with existing entity archival logic using h.entityService.DeleteEntity
	h.metrics.IncrementCounter("organization_archive_total", metrics.Fields{})
	return nil
}

// Helper functions for organization GOA handler

// generateEntityCode creates a standardized entity code from the name
func generateEntityCode(name string) string {
	// Convert to uppercase, replace spaces with underscores, and limit length
	code := strings.ReplaceAll(strings.ToUpper(name), " ", "_")
	if len(code) > 20 {
		code = code[:20]
	}
	return code
}

// parseFiscalYearMonth parses fiscal year start string (MM-DD format) to month number
func parseFiscalYearMonth(fyStart string) int {
	// Extract month from "MM-DD" format
	if len(fyStart) >= 2 {
		if month, err := strconv.Atoi(fyStart[:2]); err == nil {
			if month >= 1 && month <= 12 {
				return month
			}
		}
	}
	return 1 // Default to January if parsing fails
}

// mapEntityError converts domain errors to GOA service errors
func mapEntityError(err error) error {
	switch {
	case errors.Is(err, sharedErrors.ErrEntityNotFound):
		return organization.MakeNotFound(err)
	case errors.Is(err, sharedErrors.ErrEntityNameExists):
		return organization.MakeConflict(err)
	case errors.Is(err, sharedErrors.ErrEntityCodeExists):
		return organization.MakeConflict(err)
	case errors.Is(err, sharedErrors.ErrInvalidInput):
		return organization.MakeBadRequest(err)
	case errors.Is(err, sharedErrors.ErrInvalidEntityType):
		return organization.MakeBadRequest(err)
	case sharedErrors.IsValidationError(err):
		return organization.MakeBadRequest(err)
	default:
		return goa.NewServiceError(err, "internal_error", false, false, false)
	}
}
