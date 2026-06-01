---
title: Value Objects
portal: 4 — Backend Engineering
section: 00-module-development-guide/02-ddd-domain-design
audience: [backend-engineer, tech-lead]
related:
  - path: ./02-entity-design.md
    title: Entity Design
  - path: ./05-domain-errors.md
    title: Domain Errors
---

# Value Objects

Value objects are immutable domain types defined by their value, not their identity. Two `ContractValue` instances with the same amount are equal — there is no "which one" distinction. They carry domain validation in their constructors so invalid values cannot be created.

## When to Use a Value Object

Use a value object when:

- The field has domain validation rules (cannot be negative, must match a format, must be within a range)
- The field is referenced in multiple places and you want consistent validation
- The field is conceptually a unit (a monetary amount always has a currency, a reference number has a format)

Do **not** wrap simple strings or UUIDs in value objects unless they have validation. Unnecessary value objects create ceremony without value.

## ContractValue — Monetary Amount

Monetary amounts use `numeric(20,6)` in the database. In Go, represent them as `github.com/shopspring/decimal` for exact arithmetic. Never use `float64` for money.

```go
// internal/core/contracts/domain/value_objects.go
package domain

import (
	"errors"

	"github.com/shopspring/decimal"
)

// ContractValue represents a non-negative monetary amount.
// Zero is valid (e.g., a free contract).
// Negative values are not allowed.
type ContractValue struct {
	amount decimal.Decimal
}

var ErrContractValueNegative = errors.New("contract value cannot be negative")

// NewContractValue creates a ContractValue. Returns error if amount is negative.
func NewContractValue(amount decimal.Decimal) (ContractValue, error) {
	if amount.IsNegative() {
		return ContractValue{}, ErrContractValueNegative
	}
	return ContractValue{amount: amount}, nil
}

// MustContractValue creates a ContractValue and panics if invalid.
// Use only in tests or compile-time constants.
func MustContractValue(amount decimal.Decimal) ContractValue {
	v, err := NewContractValue(amount)
	if err != nil {
		panic(err)
	}
	return v
}

// Amount returns the underlying decimal. Read-only.
func (v ContractValue) Amount() decimal.Decimal { return v.amount }

// IsZero reports whether the value is zero.
func (v ContractValue) IsZero() bool { return v.amount.IsZero() }

// Add returns a new ContractValue with the sum.
func (v ContractValue) Add(other ContractValue) ContractValue {
	return ContractValue{amount: v.amount.Add(other.amount)}
}

// String returns a human-readable representation.
func (v ContractValue) String() string { return v.amount.String() }
```

### Why Not float64?

```go
// float64 arithmetic error — produces 30.000000000000004
total := 10.10 + 9.90 + 10.00

// decimal is exact
total := decimal.NewFromString("10.10").Add(decimal.NewFromString("9.90")).Add(decimal.NewFromString("10.00"))
// total == 30.00
```

The database stores `numeric(20,6)` precisely. float64 round-trip introduces errors. Monetary totals that are off by fractions of a cent cause audit failures and financial reconciliation problems.

## ContractNumber — Reference Number

```go
// ContractNumber is a validated human-readable contract reference.
// Format: CONT-YYYY-NNNN (e.g., CONT-2025-0042)
type ContractNumber struct {
	value string
}

var (
	ErrContractNumberEmpty   = errors.New("contract number cannot be empty")
	ErrContractNumberInvalid = errors.New("contract number must match CONT-YYYY-NNNN")
)

var contractNumberPattern = regexp.MustCompile(`^CONT-\d{4}-\d{4,}$`)

// NewContractNumber creates a ContractNumber. Returns error if format is invalid.
func NewContractNumber(s string) (ContractNumber, error) {
	if s == "" {
		return ContractNumber{}, ErrContractNumberEmpty
	}
	if !contractNumberPattern.MatchString(s) {
		return ContractNumber{}, ErrContractNumberInvalid
	}
	return ContractNumber{value: s}, nil
}

// Value returns the string value.
func (n ContractNumber) Value() string { return n.value }

// String implements fmt.Stringer.
func (n ContractNumber) String() string { return n.value }
```

## Value Object Rules

**Immutable after construction.** All fields are unexported. Methods return new instances, never mutate.

**Validation in constructor.** `New<Type>(...)` validates. If validation fails, return the zero value and an error. Never return a partially-valid value object.

**`Must<Type>` panic variant for tests.** Panicking constructors are only acceptable in tests and init-time constants. Never call them in request-handling code.

**Compare by value, not pointer.** Value objects are compared with `==` or by their accessor methods, not pointer equality.

**Zero value is either valid or clearly invalid.** A zero `ContractValue{}` should be clearly identifiable as the zero state (`IsZero()` returns true). Do not use zero values as sentinels in business logic.

## Storing Value Objects in the Database

Value objects are flattened to columns when stored:

```sql
-- ContractValue stored as a single numeric column
total_value  numeric(20,6) NOT NULL DEFAULT 0
```

The SQLC adapter converts between `decimal.Decimal` (DB row type) and `ContractValue` (domain type):

```go
// in repository/<noun>_sqlc.go
func mapRowToDomain(row db.Contract) *domain.Contract {
	totalValue, _ := domain.NewContractValue(row.TotalValue)  // DB guarantees non-negative
	return &domain.Contract{
		// ...
		TotalValue: totalValue,
		// ...
	}
}
```

If the database guarantees non-negative (CHECK constraint), use `_` to discard the error. If the data can be invalid in legacy rows, handle the error and return a repository error.

## When Not to Use Value Objects

Avoid value objects for:

- Simple IDs (`uuid.UUID` is already a well-typed value)
- Flags and booleans
- Strings that require no validation (`title`, `description`)
- Cross-context references (store as `uuid.UUID` — the remote module owns the validation)

Over-using value objects makes the domain verbose. Apply them only where the validation or behaviour they carry is non-trivial.
