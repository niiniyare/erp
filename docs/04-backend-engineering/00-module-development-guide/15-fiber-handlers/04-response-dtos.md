---
title: Response DTOs
portal: 4 — Backend Engineering
section: 00-module-development-guide/15-fiber-handlers
audience: [backend-engineer, tech-lead]
related:
  - path: ./03-request-dtos.md
    title: Request DTOs
  - path: ./05-routes.md
    title: Routes
---

# Response DTOs

`response.go` declares response structs and their mapping functions. Response DTOs are the public API shape — they must be stable across minor service changes.

## response.go

```go
// internal/api/handlers/contracts/response.go
package contracts

import (
	"time"

	"github.com/google/uuid"

	"awo.so/internal/core/contracts/domain"
)

// ContractResponse is the JSON representation of a contract.
type ContractResponse struct {
	ID             string  `json:"id"`
	TenantID       string  `json:"tenant_id"`
	EntityID       string  `json:"entity_id"`
	ContractNumber string  `json:"contract_number"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	VendorID       string  `json:"vendor_id"`
	ContractType   string  `json:"contract_type"`
	StartDate      string  `json:"start_date"`   // ISO 8601 date
	EndDate        string  `json:"end_date"`
	TotalValue     string  `json:"total_value"`  // decimal string — never float
	Currency       string  `json:"currency"`
	Status         string  `json:"status"`
	Version        int     `json:"version"`
	CreatedBy      string  `json:"created_by"`
	UpdatedBy      string  `json:"updated_by"`
	CreatedAt      string  `json:"created_at"`   // ISO 8601 datetime
	UpdatedAt      string  `json:"updated_at"`
}

// ContractListResponse wraps a paginated list of contracts.
type ContractListResponse struct {
	Items      []ContractResponse `json:"items"`
	TotalCount int64              `json:"total_count"`
	PageSize   int                `json:"page_size"`
	PageOffset int                `json:"page_offset"`
	TotalPages int                `json:"total_pages"`
}

// ContractLineResponse is the JSON representation of a contract line.
type ContractLineResponse struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	ContractID  string `json:"contract_id"`
	LineNumber  int    `json:"line_number"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	UnitPrice   string `json:"unit_price"`   // decimal string
	TotalPrice  string `json:"total_price"`  // decimal string
	Status      string `json:"status"`
	Version     int    `json:"version"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ============================================================
// Mapping functions
// ============================================================

// MapContractToResponse converts a domain.Contract to ContractResponse.
func MapContractToResponse(c *domain.Contract) ContractResponse {
	return ContractResponse{
		ID:             c.ID.String(),
		TenantID:       c.TenantID.String(),
		EntityID:       c.EntityID.String(),
		ContractNumber: c.ContractNumber,
		Title:          c.Title,
		Description:    c.Description,
		VendorID:       c.VendorID.String(),
		ContractType:   string(c.ContractType),
		StartDate:      c.StartDate.Format("2006-01-02"),
		EndDate:        c.EndDate.Format("2006-01-02"),
		TotalValue:     c.TotalValue.Amount().String(),
		Currency:       c.Currency,
		Status:         string(c.Status),
		Version:        c.Version,
		CreatedBy:      c.CreatedBy.String(),
		UpdatedBy:      c.UpdatedBy.String(),
		CreatedAt:      c.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      c.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// MapListToResponse converts a service list result to ContractListResponse.
func MapListToResponse(result *service.ListContractsResult) ContractListResponse {
	items := make([]ContractResponse, len(result.Items))
	for i, c := range result.Items {
		items[i] = MapContractToResponse(c)
	}
	return ContractListResponse{
		Items:      items,
		TotalCount: result.TotalCount,
		PageSize:   result.PageSize,
		PageOffset: result.PageOffset,
		TotalPages: result.TotalPages,
	}
}

// MapContractLineToResponse converts a domain.ContractLine to ContractLineResponse.
func MapContractLineToResponse(l *domain.ContractLine) ContractLineResponse {
	return ContractLineResponse{
		ID:          l.ID.String(),
		TenantID:    l.TenantID.String(),
		ContractID:  l.ContractID.String(),
		LineNumber:  l.LineNumber,
		Description: l.Description,
		Quantity:    l.Quantity,
		UnitPrice:   l.UnitPrice.Amount().String(),
		TotalPrice:  l.TotalPrice.Amount().String(),
		Status:      string(l.Status),
		Version:     l.Version,
		CreatedAt:   l.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   l.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
```

## Response DTO Rules

**UUIDs as strings.** All UUIDs in JSON responses are strings, not raw bytes. `uuid.UUID.String()` returns the standard hyphenated hex format.

**Monetary values as strings.** `c.TotalValue.Amount().String()` returns an exact decimal string (e.g., `"50000.000000"`). Never use `float64` for money in JSON.

**Timestamps as UTC ISO 8601.** `time.RFC3339` = `"2006-01-02T15:04:05Z07:00"`. Always call `.UTC()` before formatting to ensure UTC timezone is explicit.

**Dates without time.** `c.StartDate.Format("2006-01-02")` for date-only fields.

**No null fields.** Response DTOs use value types (not pointers) where possible. Empty strings are better than `null` for optional string fields. Exception: `DeletedAt` is omitted with `json:"deleted_at,omitempty"` — it is only relevant in admin/archive views.

**Version is always in the response.** The client needs it for subsequent update/delete operations.
