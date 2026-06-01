---
title: Repository Patterns
portal: 4 — Backend Engineering
section: 04-backend-engineering
audience: [backend-engineer]
related:
  - "[Repository Overview](01-repository-overview.md)"
  - "[SQLC Layer](../04-sqlc-layer/01-sqlc-overview.md)"
  - "[Integration Test Setup](../18-testing/02-integration-test-setup.md)"
---

# Repository Patterns

## Transactional Multi-Write

When a service operation requires multiple DB writes that must be atomic, use `WithTenant` once:

```go
func (r *sqlcContractRepo) CreateWithLines(
    ctx context.Context,
    tenantID uuid.UUID,
    contract CreateParams,
    lines []CreateLineParams,
) (*domain.Contract, error) {
    var result *domain.Contract

    err := r.store.WithTenant(ctx, tenantID, func(tx *pgx.Tx) error {
        q := sqlc.New(tx)

        // Write contract
        row, err := q.CreateContract(ctx, toSQLCCreateParams(contract))
        if err != nil {
            return parseContractDBError(err, "CreateContract")
        }
        result = toDomain(row)

        // Write lines in same transaction
        for i, line := range lines {
            _, err := q.CreateContractLine(ctx, sqlc.CreateContractLineParams{
                ContractID:  result.ID,
                Description: line.Description,
                Quantity:    line.Quantity,
                UnitPrice:   line.UnitPrice,
                LineOrder:   int32(i + 1),
            })
            if err != nil {
                return parseContractDBError(err, "CreateContractLine")
            }
        }
        return nil
    })
    return result, err
}
```

## Optimistic Lock Update

```go
func (r *sqlcContractRepo) UpdateStatus(
    ctx context.Context,
    id, tenantID uuid.UUID,
    params UpdateStatusParams,
) (*domain.Contract, error) {
    var result *domain.Contract
    err := r.store.WithTenant(ctx, tenantID, func(tx *pgx.Tx) error {
        q := sqlc.New(tx)
        row, err := q.UpdateContractStatus(ctx, sqlc.UpdateContractStatusParams{
            ID:      id,
            Status:  string(params.Status),
            Version: int32(params.Version),
        })
        if err != nil {
            if errors.Is(err, pgx.ErrNoRows) {
                // Could be not found OR version mismatch — return version conflict
                // (client should re-fetch to determine which)
                return domain.ErrVersionConflict
            }
            return parseContractDBError(err, "UpdateContractStatus")
        }
        result = toDomain(row)
        return nil
    })
    return result, err
}
```

## Batch Read

```go
func (r *sqlcContractRepo) GetByIDs(
    ctx context.Context,
    tenantID uuid.UUID,
    ids []uuid.UUID,
) ([]domain.Contract, error) {
    var contracts []domain.Contract
    err := r.store.WithTenant(ctx, tenantID, func(tx *pgx.Tx) error {
        q := sqlc.New(tx)
        rows, err := q.GetContractsByIDs(ctx, ids)
        if err != nil {
            return err
        }
        for _, row := range rows {
            contracts = append(contracts, *toDomain(row))
        }
        return nil
    })
    return contracts, err
}
```

## Pagination with Total Count

```go
func (r *sqlcContractRepo) List(
    ctx context.Context,
    tenantID uuid.UUID,
    params ListParams,
) ([]domain.Contract, int64, error) {
    var contracts []domain.Contract
    var total int64

    err := r.store.WithTenant(ctx, tenantID, func(tx *pgx.Tx) error {
        q := sqlc.New(tx)
        rows, err := q.ListContracts(ctx, sqlc.ListContractsParams{
            Status:  nullableStatus(params.Status),
            Limit_: params.Limit,
            Offset: params.Offset,
        })
        if err != nil {
            return err
        }
        contracts = make([]domain.Contract, len(rows))
        for i, row := range rows {
            contracts[i] = *toDomain(row.Contract)
            total = row.TotalCount   // window function result
        }
        return nil
    })
    return contracts, total, err
}
```

## Outbox Write Within Transaction

```go
func (r *sqlcContractRepo) CreateAndEnqueueEvent(
    ctx context.Context,
    tenantID uuid.UUID,
    params CreateParams,
    event domain.ContractCreated,
) (*domain.Contract, error) {
    var result *domain.Contract

    err := r.store.WithTenant(ctx, tenantID, func(tx *pgx.Tx) error {
        q := sqlc.New(tx)

        row, err := q.CreateContract(ctx, toSQLCCreateParams(params))
        if err != nil {
            return parseContractDBError(err, "Create")
        }
        result = toDomain(row)

        payload, _ := json.Marshal(event)
        _, err = q.InsertOutboxEvent(ctx, sqlc.InsertOutboxEventParams{
            TenantID: tenantID,
            Topic:    event.Topic(),
            Payload:  payload,
        })
        return err
    })
    return result, err
}
```

## Nullable Param Helpers

```go
func nullableStatus(s *domain.ContractStatus) pgtype.Text {
    if s == nil {
        return pgtype.Text{Valid: false}
    }
    return pgtype.Text{String: string(*s), Valid: true}
}

func nullableUUID(id *uuid.UUID) pgtype.UUID {
    if id == nil {
        return pgtype.UUID{Valid: false}
    }
    return pgtype.UUID{Bytes: *id, Valid: true}
}

func nullableText(s string) pgtype.Text {
    if s == "" {
        return pgtype.Text{Valid: false}
    }
    return pgtype.Text{String: s, Valid: true}
}
```

## Domain Mapper Completeness

Every column in the SQLC row must be mapped — leaving fields at zero value silently loses data:

```go
func toDomain(row sqlc.Contract) *domain.Contract {
    return &domain.Contract{
        ID:             row.ID,
        TenantID:       row.TenantID,
        ContractNumber: row.ContractNumber,
        Title:          row.Title,
        Description:    row.Description.String,      // pgtype.Text → string
        Status:         domain.ContractStatus(row.Status),
        ContractType:   domain.ContractType(row.ContractType),
        TotalValue:     row.TotalValue,              // decimal.Decimal (configured in sqlc.yaml)
        Currency:       row.Currency,
        StartDate:      row.StartDate.Time,          // pgtype.Date → time.Time
        EndDate:        row.EndDate.Time,
        VendorID:       row.VendorID.Bytes,          // pgtype.UUID → uuid.UUID
        EntityID:       row.EntityID.Bytes,
        Version:        int(row.Version),
        CreatedAt:      row.CreatedAt.Time,
        UpdatedAt:      row.UpdatedAt.Time,
        DeletedAt:      nullTimePtr(row.DeletedAt),
    }
}

func nullTimePtr(t pgtype.Timestamptz) *time.Time {
    if !t.Valid {
        return nil
    }
    return &t.Time
}
```
