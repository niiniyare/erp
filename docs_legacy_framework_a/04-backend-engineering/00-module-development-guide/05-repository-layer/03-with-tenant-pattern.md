> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: WithTenant Pattern
portal: 4 — Backend Engineering
section: 00-module-development-guide/05-repository-layer
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-sqlc-adapter.md
    title: SQLC Adapter
  - path: ../03-database-design/03-rls-policies.md
    title: RLS Policies
---

# WithTenant Pattern

`store.WithTenant(ctx, tenantID, func(q *db.Queries) error)` is the mandatory wrapper for every database operation. It sets `app.tenant_id` GUC before executing queries, activating the RLS policy.

## Signature

```go
// db/store.go
type Store interface {
	WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(q *Queries) error) error
	// ... other methods
}
```

## How It Works Internally

```go
// Simplified implementation of WithTenant
func (s *store) WithTenant(ctx context.Context, tenantID uuid.UUID, fn func(q *Queries) error) error {
	return s.execTx(ctx, func(tx pgx.Tx) error {
		// Set the tenant GUC for this transaction scope
		if _, err := tx.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID.String()); err != nil {
			return fmt.Errorf("set tenant GUC: %w", err)
		}
		return fn(New(tx))  // wrap the transaction in a Queries instance
	})
}
```

`SET LOCAL` scopes the GUC to the current transaction. It is automatically reset when the transaction ends — no risk of the GUC leaking to a subsequent pooled connection.

## Required: Every DB Call Through WithTenant

```go
// CORRECT
func (r *contractSQLCRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Contract, error) {
	var row db.Contract
	err := r.store.WithTenant(ctx, tenantID, func(q *db.Queries) error {
		var e error
		row, e = q.GetContractByID(ctx, db.GetContractByIDParams{
			ID:       id,
			TenantID: tenantID,
		})
		return e
	})
	// handle err...
}

// WRONG — bypasses tenant isolation
func (r *contractSQLCRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Contract, error) {
	row, err := r.store.GetContractByID(ctx, ...)  // no WithTenant wrapping
}
```

## Multiple Queries in One Tenant Context

If a service method needs to run multiple queries atomically, pass them all in one `WithTenant` callback:

```go
func (r *contractSQLCRepository) CreateWithLines(
	ctx context.Context,
	params CreateContractParams,
	lines []CreateContractLineParams,
) (*domain.Contract, error) {
	var contract db.Contract

	err := r.store.WithTenant(ctx, params.TenantID, func(q *db.Queries) error {
		// Both inserts share the same transaction and tenant GUC
		var e error
		contract, e = q.CreateContract(ctx, ...)
		if e != nil {
			return e
		}
		for _, lp := range lines {
			if _, e = q.CreateContractLine(ctx, ...); e != nil {
				return e  // transaction rolls back
			}
		}
		return nil
	})
	if err != nil {
		return nil, mapContractDBError(err, "CreateWithLines")
	}
	return mapContractRowToDomain(contract), nil
}
```

Everything inside the `fn` callback executes in the same transaction. A return error from `fn` triggers a rollback.

## What Happens When tenantID Is Zero Value

If `tenantID` is `uuid.Nil` (all zeros), `WithTenant` sets `app.tenant_id = '00000000-0000-0000-0000-000000000000'`. No valid tenant has this ID, so RLS will reject all row accesses — no rows returned, not a security breach.

The service must validate that `tenantID` is not zero before calling repo methods. This validation belongs in the session extraction step in the handler:

```go
// handler
sess, ok := c.Locals(domain.LocalsKeySession).(*iam.ResolvedSession)
if !ok || sess == nil || sess.TenantID == uuid.Nil {
	return fiber.ErrUnauthorized
}
```

## WithTenant vs Raw Transaction

Do not start transactions with `pgx.BeginTx` directly in repository code. Always use `WithTenant`, which wraps a transaction internally. This ensures:
- The tenant GUC is always set before queries run.
- Commit/rollback is handled automatically by the `WithTenant` implementation.
- The same tenant GUC is active for all queries in the callback.

If you need explicit transaction control (e.g., savepoints), contact the platform team.
