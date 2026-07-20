> ⚠️ **LEGACY DOCUMENTATION**
>
> This document describes the deprecated Framework A architecture and is retained for historical reference only.
>
> It MUST NOT be used when implementing or extending the Awo Framework (Framework B).
>
> Refer to [`awo/docs/`](../awo/docs/README.md) for the current canonical documentation.

---

---
title: Type Mapping
portal: 4 — Backend Engineering
section: 00-module-development-guide/05-repository-layer
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-sqlc-adapter.md
    title: SQLC Adapter
  - path: ./04-error-mapping.md
    title: Error Mapping
---

# Type Mapping

The repository adapter maps between SQLC-generated Go types and domain types. This page documents the common type mappings and the helper functions used in every adapter.

## SQLC → Domain Mapping Reference

| PostgreSQL type | SQLC Go type | Domain Go type | Conversion |
|----------------|-------------|---------------|------------|
| `uuid` | `pgtype.UUID` or `[16]byte` | `uuid.UUID` | `uuid.UUID(row.ID)` or `row.ID.Bytes` |
| `timestamptz NOT NULL` | `pgtype.Timestamptz` or `time.Time` | `time.Time` | `row.CreatedAt.Time` or direct |
| `timestamptz` (nullable) | `pgtype.Timestamptz` | `*time.Time` | `nullTimestampToPtr(row.DeletedAt)` |
| `numeric(20,6)` | `pgtype.Numeric` | `decimal.Decimal` | `pgNumericToDecimal(row.TotalValue)` |
| `VARCHAR(50)` status | `string` | `domain.ContractStatus` | `domain.ContractStatus(row.Status)` |
| `VARCHAR(50)` type | `string` | `domain.ContractType` | `domain.ContractType(row.ContractType)` |
| `integer` version | `int32` | `int` | `int(row.Version)` |
| `boolean` | `bool` | `bool` | direct |
| `date` | `pgtype.Date` or `time.Time` | `time.Time` | `row.StartDate.Time` or direct |

## Helper Functions

Place these in `repository/helpers.go` — shared across all adapters in the module.

```go
// internal/core/contracts/repository/helpers.go
package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// nullTimestampToPtr converts a nullable pgtype.Timestamptz to *time.Time.
func nullTimestampToPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}

// uuidPtrToNullUUID converts a *uuid.UUID to pgtype.UUID (nullable).
func uuidPtrToNullUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

// stringPtrToNullString converts a *string to pgtype.Text (nullable).
func stringPtrToNullString(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// pgNumericToDecimal converts pgtype.Numeric to shopspring decimal.
// If the value is NaN or invalid, returns decimal.Zero.
func pgNumericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(n.Int.String())
	if err != nil {
		return decimal.Zero
	}
	if n.Exp != 0 {
		scale := decimal.New(1, n.Exp)
		return d.Mul(scale)
	}
	return d
}

// decimalToPgNumeric converts shopspring decimal to pgtype.Numeric.
func decimalToPgNumeric(d decimal.Decimal) pgtype.Numeric {
	// pgx accepts decimal as a string scan
	n := pgtype.Numeric{}
	if err := n.Scan(d.String()); err != nil {
		// fallback to zero — should not happen with valid decimal
		_ = n.Scan("0")
	}
	return n
}
```

## Domain → SQLC Params Mapping

When building SQLC params structs in the adapter:

```go
// domain.ContractType (string alias) → string (SQLC VARCHAR param)
ContractType: string(params.ContractType),

// domain.ContractStatus (string alias) → string (SQLC VARCHAR param)
Status: string(params.Status),

// int (domain version) → int32 (SQLC integer param)
Version: int32(params.Version),

// decimal.Decimal (domain monetary) → pgtype.Numeric (SQLC numeric param)
TotalValue: decimalToPgNumeric(params.TotalValue),

// *uuid.UUID (optional domain filter) → pgtype.UUID (SQLC nullable)
EntityID: uuidPtrToNullUUID(params.EntityID),
```

## Row → Domain Mapping Function

Every adapter has one `map<Noun>RowToDomain` function. It is the single place that translates DB rows to domain structs:

```go
func mapContractRowToDomain(row db.Contract) *domain.Contract {
	totalValue, _ := domain.NewContractValue(pgNumericToDecimal(row.TotalValue))
	return &domain.Contract{
		ID:             row.ID,
		TenantID:       row.TenantID,
		EntityID:       row.EntityID,
		ContractNumber: row.ContractNumber,
		Title:          row.Title,
		Description:    row.Description,
		VendorID:       row.VendorID,
		ContractType:   domain.ContractType(row.ContractType),
		StartDate:      row.StartDate,
		EndDate:        row.EndDate,
		TotalValue:     totalValue,
		Currency:       row.Currency,
		Status:         domain.ContractStatus(row.Status),
		Version:        int(row.Version),
		CreatedBy:      row.CreatedBy,
		UpdatedBy:      row.UpdatedBy,
		DeletedAt:      nullTimestampToPtr(row.DeletedAt),
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}
}
```

Keep this function pure — no side effects, no DB calls. It is called in both the happy path and inside loops (for list results).

## SQLC Type Generation Notes

The exact types SQLC generates depend on the `sqlc.yaml` configuration:

```yaml
# sqlc.yaml (relevant section)
sql:
  - engine: "postgresql"
    gen:
      go:
        emit_pointers_for_null_types: true   # nullable columns become *Type
        emit_empty_slices: true              # :many queries return [] not nil
```

With `emit_pointers_for_null_types: true`, nullable columns like `deleted_at timestamptz` generate `*time.Time` instead of `pgtype.Timestamptz`. This simplifies the mapping function — no `pgtype` unwrapping needed. Check your project's `sqlc.yaml` to confirm.
