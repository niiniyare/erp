> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Read Operations
portal: 4 — Backend Engineering
section: 00-module-development-guide/06-service-layer
audience: [backend-engineer, tech-lead]
related:
  - path: ./03-write-operations.md
    title: Write Operations
  - path: ./04-authorization.md
    title: Authorization
---

# Read Operations

Read operations are simpler than writes — they enforce permission, call the repository, and return. No events, no audit trail, no notifications.

## GetByID

```go
func (s *contractService) GetByID(
	ctx context.Context,
	id, tenantID uuid.UUID,
	principal iam.Principal,
) (*domain.Contract, error) {
	ctx, span := s.tracer.Start(ctx, "ContractService.GetByID")
	defer span.End()

	// 1. Enforce permission
	allowed, err := s.authzSvc.Enforce(ctx, iam.Request{
		Subject: iam.TenantSubject(principal.UserID()),
		Domain:  iam.TenantDomain(tenantID),
		Object:  "contracts/contract/*",
		Action:  "read",
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, iam.ErrForbidden
	}

	// 2. Fetch
	contract, err := s.repo.GetByID(ctx, id, tenantID)
	if err != nil {
		span.RecordError(err)
		return nil, err  // ErrContractNotFound propagates to handler
	}

	return contract, nil
}
```

Read operations do not record audit events by default. If specific read operations need to be audited (e.g., accessing a sensitive financial record), add the audit call with `Action: "contract.read"`. Do not audit every list or detail view — it creates excessive noise.

## List with EntityScope

The `List` method applies the caller's `EntityScope` from the session. This restricts results to contracts the caller's organisational scope can see:

```go
func (s *contractService) List(
	ctx context.Context,
	params repository.ListContractsParams,
	principal iam.Principal,
) (*ListContractsResult, error) {
	ctx, span := s.tracer.Start(ctx, "ContractService.List")
	defer span.End()

	// 1. Enforce permission
	allowed, err := s.authzSvc.Enforce(ctx, iam.Request{
		Subject: iam.TenantSubject(principal.UserID()),
		Domain:  iam.TenantDomain(params.TenantID),
		Object:  "contracts/contract/*",
		Action:  "read",
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, iam.ErrForbidden
	}

	// 2. Apply default sort if not specified
	if params.SortBy == "" {
		params.SortBy = "created_at"
	}
	if params.SortDir == "" {
		params.SortDir = "desc"
	}

	// 3. Apply default page size
	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = 20
	}

	// 4. Run list + count in parallel
	var (
		items      []*domain.Contract
		totalCount int64
		listErr    error
		countErr   error
		wg         sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		items, listErr = s.repo.List(ctx, params)
	}()
	go func() {
		defer wg.Done()
		totalCount, countErr = s.repo.Count(ctx, params)
	}()
	wg.Wait()

	if listErr != nil {
		span.RecordError(listErr)
		return nil, listErr
	}
	if countErr != nil {
		span.RecordError(countErr)
		return nil, countErr
	}

	// 5. Compute pagination metadata
	totalPages := int(math.Ceil(float64(totalCount) / float64(params.PageSize)))

	return &ListContractsResult{
		Items:      items,
		TotalCount: totalCount,
		PageSize:   params.PageSize,
		PageOffset: params.PageOffset,
		TotalPages: totalPages,
	}, nil
}
```

### Parallel List + Count

Running `List` and `Count` concurrently halves the latency of paginated list endpoints. Both queries use the same filter, so their results are consistent within the same snapshot (PostgreSQL's default READ COMMITTED isolation).

### EntityScope Filtering

The handler passes `EntityID` in the `params` based on the session's `EntityScope`:

```go
// handler extracts entity scope from session
entityID := &sess.EntityScope.EntityID
if sess.EntityScope.Type == iam.EntityScopeAll {
	entityID = nil  // no entity filter — can see all entities in tenant
}

params := repository.ListContractsParams{
	TenantID: sess.TenantID,
	EntityID: entityID,
	// ...
}
```

The service does not read the session directly — it receives pre-computed params from the handler.

## List Defaults

Always enforce sensible defaults in the service, not the handler:

| Parameter | Default |
|-----------|---------|
| `SortBy` | `"created_at"` |
| `SortDir` | `"desc"` |
| `PageSize` | `20` |
| `PageSize` max | `100` |

The handler should validate that the client-provided page size is within bounds before calling the service, but the service also enforces the cap as a safety measure.
