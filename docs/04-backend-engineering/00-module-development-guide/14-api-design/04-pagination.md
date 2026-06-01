---
title: Pagination
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer, tech-lead]
related:
  - "[API Design Overview](01-api-design-overview.md)"
  - "[URL Conventions](02-url-conventions.md)"
---

# Pagination

All list endpoints paginate. Unbounded results are never returned.

## Page-Based Pagination

AwoERP uses page-based (offset) pagination — suitable for UI with page numbers:

| Parameter | Default | Max | Description |
|-----------|---------|-----|-------------|
| `page` | 1 | — | 1-based page number |
| `page_size` | 20 | 100 | Items per page |

## Handler Pattern

```go
func (h *contractHandler) List(c *fiber.Ctx) error {
    page     := c.QueryInt("page", 1)
    pageSize := c.QueryInt("page_size", 20)

    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }

    result, err := h.svc.List(c.Context(), service.ListContractsRequest{
        TenantID:  session.TenantID,
        Principal: session.ToPrincipal(),
        Page:      page,
        PageSize:  pageSize,
        // ... filters ...
    })
    if err != nil {
        return h.mapError(err)
    }

    totalPages := int(math.Ceil(float64(result.Total) / float64(pageSize)))
    return c.JSON(dto.ListContractsResponse{
        Data: dto.ContractsToResponse(result.Contracts),
        Pagination: dto.PaginationMeta{
            Page:       page,
            PageSize:   pageSize,
            Total:      result.Total,
            TotalPages: totalPages,
        },
    })
}
```

## Service Pattern

```go
func (s *contractService) List(ctx context.Context, req ListContractsRequest) (*ListContractsResult, error) {
    // Enforce default and max page size
    if req.PageSize <= 0 {
        req.PageSize = 20
    }
    if req.PageSize > 100 {
        req.PageSize = 100
    }

    offset := (req.Page - 1) * req.PageSize

    var contracts []*domain.Contract
    var total int
    var wg sync.WaitGroup
    var listErr, countErr error

    wg.Add(2)
    go func() {
        defer wg.Done()
        contracts, listErr = s.repo.List(ctx, repository.ListContractsParams{
            TenantID: req.TenantID,
            Limit:    req.PageSize,
            Offset:   offset,
        })
    }()
    go func() {
        defer wg.Done()
        total, countErr = s.repo.Count(ctx, repository.CountContractsParams{
            TenantID: req.TenantID,
        })
    }()
    wg.Wait()

    if listErr != nil {
        return nil, listErr
    }
    if countErr != nil {
        return nil, countErr
    }

    return &ListContractsResult{
        Contracts: contracts,
        Total:     total,
    }, nil
}
```

## SQLC Query

```sql
-- name: ListContracts :many
SELECT * FROM contracts
WHERE tenant_id = @tenant_id
  AND deleted_at IS NULL
  AND (@status::text IS NULL OR status = @status)
  AND (@entity_id::uuid IS NULL OR entity_id = @entity_id)
ORDER BY created_at DESC
LIMIT @limit_count
OFFSET @offset_count;
```

## Empty Results

List endpoints always return an empty array, never null:

```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 0,
    "total_pages": 0
  }
}
```

Guarantee this in the repository by returning `make([]*domain.Contract, 0)` instead of nil when no rows found.
